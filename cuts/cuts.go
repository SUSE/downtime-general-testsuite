package cuts_tests

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var testOrg = "cutsTestOrg"
var testSpace = "cutsTestSpace"
var testUser = "cutsTestUser"
var testUserPassword = "changeme"

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
