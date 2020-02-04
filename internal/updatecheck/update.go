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
	"math/rand"
	"net/http"

	"github.com/pkg/errors"
	"k8s.io/klog"

	"sigs.k8s.io/krew/internal/installation/semver"
	"sigs.k8s.io/krew/internal/version"
)

const (
	githubVersionURL = "https://api.github.com/repos/kubernetes-sigs/krew/releases/latest"

	// showRate is the percentage of krew runs for which the upgrade check is performed.
	showRate = 0.4
)

// for testing
var versionURL = githubVersionURL

// LatestTag returns the tag name of the latest release on GitHub or "" at random.
// For development builds, it always returns "".
func LatestTag() (string, error) {
	ourTag := version.GitTag()
	if !isSemver(ourTag) || // no upgrade check for dev builds
		showRate < rand.Float64() { // only do the upgrade check randomly
		return "", nil
	}
	return fetchLatestTag()
}

// fetchLatestTag fetches the tag name of the latest release from GitHub.
func fetchLatestTag() (string, error) {
	klog.V(4).Infof("Fetching latest tag from GitHub")
	response, err := http.Get(versionURL)
	if err != nil {
		return "", errors.Wrapf(err, "could not GET the latest release")
	}
	defer response.Body.Close()

	var res struct {
		Tag string `json:"tag_name"`
	}
	klog.V(4).Infof("Parsing response from GitHub")
	if err := json.NewDecoder(response.Body).Decode(&res); err != nil {
		return "", errors.Wrapf(err, "could not parse the response from GitHub")
	}
	klog.V(4).Infof("Fetched latest tag name (%s) from GitHub", res.Tag)
	return res.Tag, nil
}

// isSemver tries to parse s as semver. If it fails, it assumes that s is not a semver.
// For development builds, this usually returns false.
func isSemver(s string) bool {
	_, err := semver.Parse(s)
	return err == nil
}
