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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sigs.k8s.io/krew/internal/testutil"
	"sigs.k8s.io/krew/internal/version"
)

type ConstantHandler string

func (c ConstantHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(c))
}

func TestCheckVersion(t *testing.T) {
	tempDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	server := httptest.NewServer(ConstantHandler(`{"tag_name": "some_tag"}`))
	defer server.Close()

	firstMessage := CheckVersion(tempDir.Root())
	if !strings.Contains(firstMessage, "You are using an old version of krew") {
		t.Error("The initial version check should notify about the new version")
	}

	secondMessage := CheckVersion(tempDir.Root())
	if secondMessage != "" {
		t.Error("An immediately following check should return an empty message")
	}
}

func Test_fetchTag(t *testing.T) {
	var dawnOfTime = time.Unix(0, 0)
	tests := []struct {
		name      string
		expected  string
		response  string
		lastCheck time.Time
	}{
		{
			name:      "broken json",
			response:  `{"tag_name"::]`,
			expected:  version.GitTag(),
			lastCheck: dawnOfTime,
		},
		{
			name:      "field missing",
			response:  `{}`,
			expected:  version.GitTag(),
			lastCheck: dawnOfTime,
		},
		{
			name:      "should get the correct tag",
			response:  `{"tag_name": "some_tag"}`,
			expected:  "some_tag",
			lastCheck: dawnOfTime,
		},
		{
			name:      "should not fetch the tag when done so recently",
			response:  `{"tag_name": "some_tag"}`,
			expected:  version.GitTag(),
			lastCheck: time.Now(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(ConstantHandler(test.response))
			defer server.Close()

			versionURL = server.URL
			defer func() { versionURL = githubVersionURL }()

			tag := fetchTag(test.lastCheck)
			if tag != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, tag)
			}
		})
	}
}
