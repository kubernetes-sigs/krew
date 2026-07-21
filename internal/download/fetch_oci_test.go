// Copyright 2026 The Kubernetes Authors.
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
	"bytes"
	"io"
	"strings"
	"testing"

	"sigs.k8s.io/krew/internal/testutil"
)

func TestIsOCI(t *testing.T) {
	tests := map[string]bool{
		"oci://ghcr.io/org/plugins/foo:v1.0.0": true,
		"https://example.com/foo.tar.gz":       false,
		"oci:ghcr.io/org/plugins/foo:v1.0.0":   false,
		"":                                     false,
	}
	for uri, want := range tests {
		if got := IsOCI(uri); got != want {
			t.Errorf("IsOCI(%q) = %v, want %v", uri, got, want)
		}
	}
}

func TestOCIFetcher_Get(t *testing.T) {
	archive := []byte("the plugin archive, byte for byte")
	uri := testutil.PushOCIArtifact(t, testutil.NewOCIRegistry(t), "plugins/foo:v1.0.0", archive)

	rc, err := OCIFetcher{}.Get(uri)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, archive) {
		t.Errorf("Get(%q) = %q, want the layer's exact bytes %q", uri, got, archive)
	}
}

func TestOCIFetcher_Get_errors(t *testing.T) {
	host := testutil.NewOCIRegistry(t)
	tests := []struct {
		name    string
		uri     string
		wantErr string
	}{
		{
			name:    "invalid reference",
			uri:     "oci://" + host + "/Plugins/Foo:v1.0.0",
			wantErr: "invalid OCI reference",
		},
		{
			name:    "unknown tag",
			uri:     "oci://" + host + "/plugins/foo:v9.9.9",
			wantErr: "failed to resolve OCI artifact",
		},
		{
			name:    "no layer",
			uri:     testutil.PushOCIArtifact(t, host, "plugins/empty:v1.0.0"),
			wantErr: "must have exactly one layer",
		},
		{
			name:    "two layers",
			uri:     testutil.PushOCIArtifact(t, host, "plugins/two:v1.0.0", []byte("a"), []byte("b")),
			wantErr: "must have exactly one layer",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc, err := OCIFetcher{}.Get(tt.uri)
			if err == nil {
				rc.Close()
				t.Fatalf("Get(%q) returned no error, want one containing %q", tt.uri, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Get(%q) error = %q, want one containing %q", tt.uri, err, tt.wantErr)
			}
		})
	}
}
