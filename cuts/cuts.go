package cuts_tests

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var testOrg = "cutsTestOrg"
var testSpace = "cutsTestSpace"
var testUser = "cutsTestUser"
var testUserPassword = "changeme"
var testApp = "testDora"

var _ = Describe("Upgrade process", func() {
	var _ = Describe("PreUpgrade", func() {
		It("sets the api to the kubecf cluster", func() {
			cmd := exec.Command("cf", "login", "--skip-ssl-validation", "-p"+cfAdminPassword, "-uadmin", "-a"+apiURL)
			err := cmd.Run()
			Expect(err).ToNot(HaveOccurred())
		})

		It("sets up an org", func() {
			cmd := exec.Command("cf", "create-org", testOrg)
			err := cmd.Run()
			Expect(err).ToNot(HaveOccurred())
		})

		It("sets up a space", func() {
			cmd := exec.Command("cf", "create-space", testSpace, "-o"+testOrg)
			err := cmd.Run()
			Expect(err).ToNot(HaveOccurred())
		})

		It("sets up an user", func() {
			cmd := exec.Command("cf", "create-user", testUser, testUserPassword)
			err := cmd.Run()
			Expect(err).ToNot(HaveOccurred())
		})
		It("pushes a test app", func() {
			cmd := exec.Command("cf", "target", "-o"+testOrg, "-s"+testSpace)
			err := cmd.Run()
			Expect(err).ToNot(HaveOccurred())

			_, filename, _, ok := runtime.Caller(0)
			Expect(ok).To(BeTrue())

			cmd = exec.Command("cf", "push", testApp)
			cmd.Dir = path.Join(path.Dir(filename), "../assets/go-dora")
			cmd.Stderr = GinkgoWriter
			cmd.Stdout = GinkgoWriter
			err = cmd.Run()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	var _ = Describe("DuringUpgrade", func() {
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

				http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
				for {
					select {
					case <-abort:
						fmt.Printf("%#v\n", lastFailure)
						if lastFailure == nil {
							t := time.Now()
							lastFailure = &t
						}
						result <- PollResult{
							Requests:       requests,
							Failures:       fails,
							Downtime:       lastFailure.Sub(*firstFailure),
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
								fmt.Printf("%#v\n", firstFailure)
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
			err := cmd.Run()
			Expect(err).ToNot(HaveOccurred())

			abort <- true
			pollResult := <-result
			fmt.Printf("Requests: %d  Failed: %d Downtime %.2f s Updateduration: %.2f s", pollResult.Requests,
				pollResult.Failures, pollResult.Downtime.Seconds(), pollResult.UpdateDuration.Seconds())
		})
	})

	var _ = Describe("PostUpgrade", func() {

	})
})
