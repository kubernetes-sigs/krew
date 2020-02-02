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

package updatecheck

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"k8s.io/klog"

	"sigs.k8s.io/krew/internal/version"
)

const (
	githubVersionURL = "https://api.github.com/repos/kubernetes-sigs/krew/releases/latest"

	upgradeNotification = `A newer version of krew is available (%s -> %s). 
Run "kubectl krew upgrade" to get the newest version!`
)

// for testing
var versionURL = githubVersionURL

// CheckVersion returns a notification message to inform the user
// about a new version of krew, or an empty string. The notification
// will only be emitted once a day.
func CheckVersion(basePath string) (string, error) {
	f := filepath.Join(basePath, "last_update_check")
	if lastCheck := loadTimestamp(f); time.Since(lastCheck).Hours() <= 24 {
		klog.V(3).Info("Last check was recently, skipping update check")
		return "", nil
	}

	latestTag, err := fetchLatestTag()
	if err != nil {
		return "", errors.Wrapf(err, "could not determine latest tag")
	}
	saveTimestamp(f)

	if version.GitTag() == latestTag {
		klog.V(3).Info("Latest tag is same as ours.")
		return "", nil
	}
	return color.New(color.Bold).Sprintf(upgradeNotification, version.GitTag(), latestTag), nil
}

// fetchLatestTag fetches the tag name of the latest release from GitHub.
func fetchLatestTag() (string, error) {
	response, err := http.Get(versionURL)
	if err != nil {
		return "", errors.Wrapf(err, "could not GET the latest release")
	}
	defer response.Body.Close()

	var res struct {
		Tag string `json:"tag_name"`
	}
	if err := json.NewDecoder(response.Body).Decode(&res); err != nil {
		return "", errors.Wrapf(err, "could not parse the response from GitHub")
	}
	klog.V(4).Infof("Fetched latest tag name (%s) from GitHub", res.Tag)
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

	klog.V(4).Infof("Last version check was on %s", timestamp)
	return timestamp
}
