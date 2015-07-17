package v3_helpers

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/runner"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

func StartApp(appGuid string) {
	startURL := fmt.Sprintf("/v3/apps/%s/start", appGuid)
	Expect(cf.Cf("curl", startURL, "-X", "PUT").Wait(DEFAULT_TIMEOUT)).To(Exit(0))
}

func StopApp(appGuid string) {
	stopURL := fmt.Sprintf("/v3/apps/%s/stop", appGuid)
	Expect(cf.Cf("curl", stopURL, "-X", "PUT").Wait(DEFAULT_TIMEOUT)).To(Exit(0))
}

func CreateApp(appName, spaceGuid, environmentVariables string) string {
	session := cf.Cf("curl", "/v3/apps", "-X", "POST", "-d", fmt.Sprintf(`{"name":"%s", "space_guid":"%s", "environment_variables":%s}`, appName, spaceGuid, environmentVariables))
	bytes := session.Wait(DEFAULT_TIMEOUT).Out.Contents()
	var app struct {
		Guid string `json:"guid"`
	}
	json.Unmarshal(bytes, &app)
	return app.Guid
}

func WaitForPackageToBeReady(packageGuid string) {
	pkgUrl := fmt.Sprintf("/v3/packages/%s", packageGuid)
	Eventually(func() *Session {
		session := cf.Cf("curl", pkgUrl)
		Expect(session.Wait(DEFAULT_TIMEOUT)).To(Exit(0))
		return session
	}, LONG_CURL_TIMEOUT).Should(Say("READY"))
}

func CreatePackage(appGuid string) string {
	packageCreateUrl := fmt.Sprintf("/v3/apps/%s/packages", appGuid)
	session := cf.Cf("curl", packageCreateUrl, "-X", "POST", "-d", fmt.Sprintf(`{"type":"bits"}`))
	bytes := session.Wait(DEFAULT_TIMEOUT).Out.Contents()
	var pac struct {
		Guid string `json:"guid"`
	}
	json.Unmarshal(bytes, &pac)
	return pac.Guid
}

func GetSpaceGuidFromName(spaceName string) string {
	session := cf.Cf("space", spaceName, "--guid")
	bytes := session.Wait(DEFAULT_TIMEOUT).Out.Contents()
	return strings.TrimSpace(string(bytes))
}

func GetAuthToken() string {
	bytes := runner.Run("bash", "-c", "cf oauth-token | tail -n +4").Wait(DEFAULT_TIMEOUT).Out.Contents()
	return strings.TrimSpace(string(bytes))
}

func UploadPackage(uploadUrl, packageZipPath, token string) {
	bits := fmt.Sprintf(`bits=@"%s"`, packageZipPath)
	_, err := exec.Command("curl", "-v", "-s", uploadUrl, "-F", bits, "-H", fmt.Sprintf("Authorization: %s", token)).CombinedOutput()
	Expect(err).NotTo(HaveOccurred())
}

func StagePackage(packageGuid, stageBody string) string {
	stageUrl := fmt.Sprintf("/v3/packages/%s/droplets", packageGuid)
	session := cf.Cf("curl", stageUrl, "-X", "POST", "-d", stageBody)
	bytes := session.Wait(DEFAULT_TIMEOUT).Out.Contents()
	var droplet struct {
		Guid string `json:"guid"`
	}
	json.Unmarshal(bytes, &droplet)
	return droplet.Guid
}