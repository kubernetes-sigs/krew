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

package download

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindNetrcEntry(t *testing.T) {
	// Create a temporary .netrc file
	content := `machine example.com login user password pass
machine api.example.com login apiuser password apipass
`
	tmpDir := t.TempDir()
	netrcPath := filepath.Join(tmpDir, ".netrc")
	err := os.WriteFile(netrcPath, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test .netrc: %v", err)
	}

	tests := []struct {
		uri      string
		expected *NetrcEntry
	}{
		{"https://example.com/path", &NetrcEntry{Machine: "example.com", Login: "user", Password: "pass"}},
		{"https://api.example.com/path", &NetrcEntry{Machine: "api.example.com", Login: "apiuser", Password: "apipass"}},
		{"https://unknown.com/path", nil},
		{"https://example.com:8080/path", &NetrcEntry{Machine: "example.com", Login: "user", Password: "pass"}},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			entry, err := FindNetrcEntry(tt.uri, netrcPath)
			if err != nil {
				t.Errorf("FindNetrcEntry failed: %v", err)
				return
			}
			switch {
			case tt.expected == nil && entry != nil:
				t.Errorf("Expected nil, got %+v", entry)
			case tt.expected != nil && entry == nil:
				t.Errorf("Expected %+v, got nil", tt.expected)
			case tt.expected != nil && entry != nil && *entry != *tt.expected:
				t.Errorf("Got %+v, want %+v", entry, tt.expected)
			}
		})
	}
}

func TestFindNetrcEntry_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "does-not-exist")

	_, err := FindNetrcEntry("https://example.com/path", nonExistentPath)
	if err == nil {
		t.Error("Expected error when netrc file does not exist, got nil")
	}
}
