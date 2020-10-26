package dogstests

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

var apiURL string
var systemDomain string
var cfAdminPassword string

func TestBase(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DOGS")
}

var _ = BeforeSuite(func() {
	var err error
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
