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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sigs.k8s.io/krew/internal/testutil"
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

	versionURL = server.URL
	defer func() { versionURL = githubVersionURL }()

	firstMessage, err := CheckVersion(tempDir.Root())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(firstMessage, `Run "kubectl krew upgrade" to get the newest version!`) {
		t.Error("The initial version check should notify about the new version")
	}

	secondMessage, err := CheckVersion(tempDir.Root())
	if err != nil {
		t.Fatal(err)
	}
	if secondMessage != "" {
		t.Error("An immediately following check should return an empty message")
	}
}

func Test_fetchTag(t *testing.T) {
	tests := []struct {
		name      string
		expected  string
		response  string
		shouldErr bool
	}{
		{
			name:      "broken json",
			response:  `{"tag_name"::]`,
			shouldErr: true,
		},
		{
			name:     "field missing",
			response: `{}`,
		},
		{
			name:     "should get the correct tag",
			response: `{"tag_name": "some_tag"}`,
			expected: "some_tag",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			server := httptest.NewServer(ConstantHandler(test.response))
			defer server.Close()

			versionURL = server.URL
			defer func() { versionURL = githubVersionURL }()

			tag, err := fetchLatestTag()
			if test.shouldErr && err == nil {
				tt.Error("Expected an error but found none")
			}
			if !test.shouldErr && err != nil {
				tt.Errorf("Expected no error but found: %s", err)
			}
			if tag != test.expected {
				tt.Errorf("Expected %s, got %s", test.expected, tag)
			}
		})
	}
}
