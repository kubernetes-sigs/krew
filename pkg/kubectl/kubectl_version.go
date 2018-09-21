package kubectl

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/coreos/go-semver/semver"
)

var (
	minKubectl      = semver.New("1.12.0-alpha.0")
	versionProvider = execKubectlVersion
)

func execKubectlVersion() (string, error) {
	cmd := exec.Command("kubectl", "version", "--client", "--output=json")
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute kubectl version. error=%v, out=%v", err, string(b))
	}
	return string(b), nil
}

func parseKubectlVersion(s string) (*semver.Version, error) {
	v := struct {
		ClientVersion struct {
			GitVersion string `json:"gitVersion"`
		} `json:"clientVersion"`
	}{}

	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil, fmt.Errorf("error parsing kubectl version: %+v", err)
	}
	if v.ClientVersion.GitVersion == "" {
		return nil, fmt.Errorf("could not locate kubectl client version in: %s", s)
	}
	ver, err := semver.NewVersion(strings.TrimPrefix(v.ClientVersion.GitVersion, "v"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse semver string=%q error=%q", v.ClientVersion.GitVersion, err)
	}
	return ver, nil
}

func getKubectlVersion() (*semver.Version, error) {
	out, err := versionProvider()
	if err != nil {
		return nil, fmt.Errorf("could not detect kubectl version: %+v", err)
	}
	v, err := parseKubectlVersion(out)
	if err != nil {
		return nil, fmt.Errorf("error parsing kubectl version: %+v", err)
	}
	return v, nil
}

// IsSupportedVersion determines if kubectl version satisfies the minimum
// verison requirements.
func IsSupportedVersion() (bool, error) {
	v, err := getKubectlVersion()
	if err != nil {
		return false, fmt.Errorf("version check error: %+v", err)
	}
	return !v.LessThan(*minKubectl), nil
}
