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
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_fetchLatestTag_GitHubAPI(t *testing.T) {
	tag, err := FetchLatestTag()
	if err != nil {
		t.Error(err)
	}
	if tag == "" {
		t.Errorf("Expected a latest tag in the response")
	}
}

func Test_fetchLatestTag(t *testing.T) {
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
			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, _ *http.Request) {
					_, _ = w.Write([]byte(test.response))
				},
			))

			defer server.Close()

			versionURL = server.URL
			defer func() { versionURL = githubVersionURL }()

			tag, err := FetchLatestTag()
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

func Test_fetchLatestTagFailure(t *testing.T) {
	versionURL = "http://localhost/nirvana"
	defer func() { versionURL = githubVersionURL }()

	_, err := FetchLatestTag()
	if err == nil {
		t.Error("Expected an error but found none")
	}
}
