package cuts_tests

import (
	"os/exec"
	"path"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var testOrg = "cutsTestOrg"
var testSpace = "cutsTestSpace"
var testUser = "cutsTestUser"
var testUserPassword = "changeme"
var testApp = "testDora"

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
	//It("runs the update process", func() {
	//	//TODO: fork into two processes, run the upgrade in one and the specs in the other
	//	//      pass a semaphore from the upgrade process to the tests
	//	cmd := exec.Cmd("KUBECF_CHART="+os.Getenv("KUBECF_TARGET_CHART"), "make", "kubecf-upgrade")
	//	_, err := cmd.Run()
	//
	//	Expect(err).ToNot(HaveOccurred())
	//})
})

var _ = Describe("PostUpgrade", func() {

})
