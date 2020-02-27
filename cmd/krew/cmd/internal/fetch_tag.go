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

package internal

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"k8s.io/klog"
)

const (
	githubVersionURL = "https://api.github.com/repos/kubernetes-sigs/krew/releases/latest"
)

// for testing
var versionURL = githubVersionURL

// FetchLatestTag fetches the tag name of the latest release from GitHub.
func FetchLatestTag() (string, error) {
	klog.V(4).Infof("Fetching latest tag from GitHub")
	response, err := http.Get(versionURL)
	if err != nil {
		return "", errors.Wrapf(err, "could not GET the latest release")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", errors.Errorf("expected HTTP status 200 OK, got %s", response.Status)
	}

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
