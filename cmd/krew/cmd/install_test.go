// Copyright 2019 The Kubernetes Authors.
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

package cmd

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func Test_readPluginFromURL(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir(filepath.FromSlash("../../../integration_test/testdata"))))
	defer server.Close()

	tests := []struct {
		name     string
		url      string
		wantName string
		wantErr  bool
	}{
		{
			name:     "successful request",
			url:      server.URL + "/ctx.yaml",
			wantName: "ctx",
		},
		{
			name:    "invalid server",
			url:     "http://example.invalid:80/foo.yaml",
			wantErr: true,
		},
		{
			name:    "bad response",
			url:     server.URL + "/404.yaml",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readPluginFromURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Fatalf("readPluginFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Name != tt.wantName {
				t.Fatalf("readPluginFromURL() returned name=%v; want=%v", got.Name, tt.wantName)
			}
		})
	}
}
