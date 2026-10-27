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

package testutil

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// NewOCIRegistry starts an in-memory OCI registry, stopped when the test
// exits, and returns its host.
func NewOCIRegistry(t *testing.T) string {
	t.Helper()
	server := httptest.NewServer(registry.New())
	t.Cleanup(server.Close)
	return strings.TrimPrefix(server.URL, "http://")
}

// PushOCIArtifact pushes an artifact made of the given layers to host, the way
// `oras push <ref> <file>` stores a file, and returns its oci:// uri.
func PushOCIArtifact(t *testing.T, host, repoTag string, layers ...[]byte) string {
	t.Helper()
	img := mutate.MediaType(empty.Image, types.OCIManifestSchema1)
	for _, l := range layers {
		var err error
		img, err = mutate.AppendLayers(img, static.NewLayer(l, types.MediaType("application/vnd.oci.image.layer.v1.tar")))
		if err != nil {
			t.Fatal(err)
		}
	}
	ref, err := name.ParseReference(host + "/" + repoTag)
	if err != nil {
		t.Fatal(err)
	}
	if err := remote.Write(ref, img); err != nil {
		t.Fatal(err)
	}
	return "oci://" + host + "/" + repoTag
}
