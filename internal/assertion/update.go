// Copyright 2020 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package assertion

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"k8s.io/klog"

	"sigs.k8s.io/krew/internal/version"
)

const githubVersionURL = "https://api.github.com/repos/kubernetes-sigs/krew/releases/latest"

// for testing
var versionURL = githubVersionURL

type githubResponse struct {
	Tag string `json:"tag_name"`
}

// CheckVersion returns a notification message to inform the user
// about a new version of krew, or an empty string. The notification
// will only be emitted once a day.
func CheckVersion(basePath string) string {
	f := filepath.Join(basePath, "last_update_check")
	lastCheck := loadTimestamp(f)
	tag := fetchTag(lastCheck)
	if version.GitTag() == tag {
		return ""
	}
	saveTimestamp(f)

	bold := color.New(color.Bold)
	return bold.Sprintf("You are using an old version of krew (%s). Please upgrade to the latest version (%s)", version.GitTag(), tag)
}

// fetchTag tries to return the tag name of the most recent krew
// release. If the last check happened recently, or an error occurs
// the hardcoded tag of the executing krew binary will be returned.
// This effectively disables the update notification message.
func fetchTag(lastCheck time.Time) string {
	if time.Since(lastCheck).Hours() <= 24 {
		klog.V(3).Info("Last check was recently, skipping update check")
		return version.GitTag()
	}

	currentVersion, err := fetchLatestTagFromGithub()
	if err != nil || currentVersion == "" {
		klog.V(3).Infof("Could not fetch most recent tag name: %s", err)
		return version.GitTag()
	}

	return currentVersion
}

// fetchLatestTagFromGithub fetches the tag name of the latest release from github.
func fetchLatestTagFromGithub() (string, error) {
	response, err := http.Get(versionURL)
	if err != nil {
		klog.V(4).Infof("Could not GET the latest release: %s", err)
		return "", err
	}
	defer response.Body.Close()
	var res githubResponse
	if err := json.NewDecoder(response.Body).Decode(&res); err != nil {
		klog.V(4).Infof("Could not parse the response from GitHub: %s", err)
		return "", err
	}
	return res.Tag, nil
}

func saveTimestamp(file string) {
	klog.V(4).Info("Saving timestamp for last version check")
	timestamp := time.Now().Format(time.RFC1123)
	if err := ioutil.WriteFile(file, []byte(timestamp), 0666); err != nil {
		klog.V(4).Info("Could not write version information")
	}
}

// loadTimestamp tries to load the timestamp of the last version check.
// In case of error the returned timestamp will force a version check.
func loadTimestamp(file string) time.Time {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		klog.V(4).Infof("Could not read timestamp file %q", file)
		return time.Unix(0, 0)
	}

	timestamp, err := time.Parse(time.RFC1123, string(content))
	if err != nil {
		klog.V(4).Infof("Could not parse timestamp %q", string(content))
		return time.Unix(0, 0)
	}

	klog.V(4).Infof("Last version check on %s", timestamp)
	return timestamp
}
