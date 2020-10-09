package cutstests

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"time"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/dragonchaser/cluster-acceptance-tests/cuts/uaa"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Upgrade process", func() {
	var (
		client *cfclient.Client

		testDomain string
		testOrg    = "cutsTestOrg"
		testSpace  = "cutsTestSpace"
		testUser   = "cutsTestUser"
		// testUserPassword = "changeme"
		testApp = "testDora"
	)

	BeforeEach(func() {
		testDomain = "cutstests." + systemDomain
		c := &cfclient.Config{
			ApiAddress:        "https://" + apiURL,
			Username:          "admin",
			Password:          cfAdminPassword,
			SkipSslValidation: true,
		}
		var err error
		client, err = cfclient.NewClient(c)
		Expect(err).ToNot(HaveOccurred())

		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	})

	var _ = Describe("Before the upgrade", func() {
		It("sets up an org and space", func() {
			orgReq := cfclient.OrgRequest{
				Name: testOrg,
			}
			org, err := client.CreateOrg(orgReq)
			Expect(err).ToNot(HaveOccurred())

			spaceReq := cfclient.SpaceRequest{
				Name:             testSpace,
				OrganizationGuid: org.Guid,
			}
			_, err = client.CreateSpace(spaceReq)
			Expect(err).ToNot(HaveOccurred())
		})

		It("sets up an user", func() {
			uaaClient := uaa.NewClient("admin", cfAdminPassword, "https://uaa."+systemDomain)

			guid, err := uaaClient.CreateUser(testUser)
			Expect(err).ToNot(HaveOccurred())

			userReq := cfclient.UserRequest{
				Guid: guid,
			}
			_, err = client.CreateUser(userReq)
			Expect(err).ToNot(HaveOccurred())
		})

		It("pushes a test app", func() {
			org, err := client.GetOrgByName(testOrg)
			Expect(err).ToNot(HaveOccurred())
			space, err := client.GetSpaceByName(testSpace, org.Guid)
			Expect(err).ToNot(HaveOccurred())

			// Create app
			appReq := cfclient.AppCreateRequest{
				Name:      testApp,
				SpaceGuid: space.Guid,
				Instances: 3,
			}
			app, err := client.CreateApp(appReq)
			Expect(err).ToNot(HaveOccurred())

			_, filename, _, ok := runtime.Caller(0)
			Expect(ok).To(BeTrue())

			reader, err := os.Open(path.Join(path.Dir(filename), "../assets/dora.zip"))
			Expect(err).ToNot(HaveOccurred())

			err = client.UploadAppBits(reader, app.Guid)
			Expect(err).ToNot(HaveOccurred())

			// Start app
			err = client.StartApp(app.Guid)
			Expect(err).ToNot(HaveOccurred())

			// Create and bind route
			domain, err := client.CreateDomain(testDomain, org.Guid)
			Expect(err).ToNot(HaveOccurred())
			routeReq := cfclient.RouteRequest{
				DomainGuid: domain.Guid,
				SpaceGuid:  space.Guid,
			}
			route, err := client.CreateRoute(routeReq)
			Expect(err).ToNot(HaveOccurred())
			err = client.BindRoute(route.Guid, app.Guid)
			Expect(err).ToNot(HaveOccurred())

			// Wait for app to come up
			Eventually(func() bool {
				res, err := http.Get("https://" + testDomain)
				return err == nil && res.StatusCode == 200
			}, "180s", "1s").Should(BeTrue())
		})
	})

	var _ = Describe("During the upgrade", func() {
		It("runs the update process", func() {
			type PollResult struct {
				Requests       int
				Failures       int
				Downtime       time.Duration
				UpdateDuration time.Duration
			}
			pollFunction := func(result chan PollResult, abort chan bool) {
				fails := 0
				requests := 0
				var firstFailure *time.Time
				var lastFailure *time.Time
				startTime := time.Now()

				for {
					select {
					case <-abort:
						if lastFailure == nil {
							t := time.Now()
							lastFailure = &t
						}
						var downtime time.Duration
						if firstFailure != nil {
							downtime = lastFailure.Sub(*firstFailure)
						} else {
							downtime = 0
						}
						result <- PollResult{
							Requests:       requests,
							Failures:       fails,
							Downtime:       downtime,
							UpdateDuration: time.Now().Sub(startTime),
						}
						break
					default:
						time.Sleep(50 * time.Millisecond)
						_, err := http.Get("https://" + testApp + "." + systemDomain)
						requests++
						if err != nil {
							t := time.Now()
							if firstFailure == nil {
								firstFailure = &t
							} else {
								lastFailure = &t
							}
							fails++
						}
					}
				}
			}

			result := make(chan PollResult, 2)
			abort := make(chan bool)
			go pollFunction(result, abort)

			_, filename, _, ok := runtime.Caller(0)
			Expect(ok).To(BeTrue())

			targetChart := os.Getenv("KUBECF_TARGET_CHART")
			cmd := exec.Command("./kubecf_helper.sh", "upgrade")
			cmd.Dir = path.Join(path.Dir(filename), "../helpers")
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, "KUBECF_CHART="+targetChart)
			cmd.Stdout = GinkgoWriter
			cmd.Stderr = GinkgoWriter
			err := cmd.Run()
			Expect(err).ToNot(HaveOccurred())

			abort <- true
			pollResult := <-result
			fmt.Printf("Requests: %d  Failed: %d Downtime %.2f s Updateduration: %.2f s", pollResult.Requests,
				pollResult.Failures, pollResult.Downtime.Seconds(), pollResult.UpdateDuration.Seconds())
		})
	})

	var _ = Describe("After the upgrade", func() {
		var _ = Context("with a test org", func() {
			var (
				org cfclient.Org
			)

			BeforeEach(func() {
				var err error
				org, err = client.GetOrgByName(testOrg)
				Expect(err).ToNot(HaveOccurred())
				_, err = client.GetSpaceByName(testSpace, org.Guid)
				Expect(err).ToNot(HaveOccurred())
			})

			It("deletes the org", func() {
				err := client.DeleteOrg(org.Guid, true, false)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		var _ = Context("with a test user", func() {
			var (
				user      cfclient.User
				uaaClient *uaa.Client
			)

			BeforeEach(func() {
				uaaClient = uaa.NewClient("admin", cfAdminPassword, "https://uaa."+systemDomain)

				guid, err := uaaClient.GetUserGUID(testUser)
				Expect(err).ToNot(HaveOccurred())

				user, err = client.GetUserByGUID(guid)
				Expect(err).ToNot(HaveOccurred())
			})

			It("preserves the user", func() {
				Expect(user.Username).To(Equal(testUser))
			})

			It("deletes the user", func() {
				err := client.DeleteUser(user.Guid)
				Expect(err).ToNot(HaveOccurred())

				err = uaaClient.DeleteUser(user.Guid)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
