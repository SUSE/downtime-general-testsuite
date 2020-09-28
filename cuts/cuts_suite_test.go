package cuts_tests

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	. "github.com/cloudfoundry/cf-acceptance-tests/helpers/cli_version_check"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

const minCliVersion = "6.33.1"

var apiURL string
var systemDomain string
var cfAdminPassword string

func TestBase(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CUTS")
}

var _ = BeforeSuite(func() {
	installedVersion, err := GetInstalledCliVersionString()
	Expect(err).ToNot(HaveOccurred(), "Error trying to determine CF CLI version")
	fmt.Println("Running CUTs with CF CLI version ", installedVersion)

	apiURL, systemDomain, err = getCfDomains()
	Expect(err).ToNot(HaveOccurred())
	Expect(systemDomain).ToNot(BeEmpty())
	Expect(apiURL).ToNot(BeEmpty())

	cfAdminPassword, err = getCfAdminPassword()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfAdminPassword).ToNot(BeEmpty())
})

func getCfAdminPassword() (string, error) {
	cmd := exec.Command("kubectl", "get", "secret", "-nscf", "var-cf-admin-password", "-ojson")
	rawJSON, err := cmd.Output()
	if err != nil {
		return "", err
	}

	jsonResult := map[string]interface{}{}
	err = json.Unmarshal(rawJSON, &jsonResult)
	if err != nil {
		return "", err
	}
	secretData := jsonResult["data"].(map[string]interface{})

	rawCfAdminPassword, err := base64.StdEncoding.DecodeString(secretData["password"].(string))
	if err != nil {
		return "", err
	}

	return string(rawCfAdminPassword), nil
}

func getCfDomains() (string, string, error) {
	configValuesYaml := os.Getenv("CATAPULT_DIR") + "/buildkind/scf-config-values.yaml"
	data, err := ioutil.ReadFile((configValuesYaml))
	if err != nil {
		return "", "", err
	}

	result := map[string]interface{}{}
	err = yaml.Unmarshal(data, &result)
	if err != nil {
		return "", "", err
	}

	domain := result["system_domain"].(string)
	apiurl := "api." + domain

	return apiurl, domain, nil
}
