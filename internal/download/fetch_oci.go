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
	"io"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

var _ Fetcher = OCIFetcher{}

// OCIFetcher fetches a plugin archive stored as an OCI artifact
// (uri scheme "oci://", e.g. oci://ghcr.io/org/plugins/foo:v1.0.0). The
// archive is the artifact's single layer, pushed as-is (e.g. `oras push
// repo:tag plugin.tar.gz`), so the manifest's sha256 keeps verifying the
// exact same bytes as an http download would. Credentials come from the
// standard docker keychain (~/.docker/config.json including credential
// helpers) — private registries work for anyone who can `docker login`.
type OCIFetcher struct{}

// IsOCI reports whether the uri uses the oci:// scheme.
func IsOCI(uri string) bool { return strings.HasPrefix(uri, "oci://") }

// Get gets the artifact's first layer and returns a stream of its raw bytes.
func (OCIFetcher) Get(uri string) (io.ReadCloser, error) {
	klog.V(2).Infof("Fetching OCI artifact %q", uri)
	ref, err := name.ParseReference(strings.TrimPrefix(uri, "oci://"))
	if err != nil {
		return nil, errors.Wrapf(err, "invalid OCI reference %q", uri)
	}
	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to resolve OCI artifact %q", uri)
	}
	layers, err := img.Layers()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read layers of %q", uri)
	}
	if len(layers) != 1 {
		return nil, errors.Errorf("OCI artifact %q must have exactly one layer (the plugin archive), found %d", uri, len(layers))
	}
	rc, err := layers[0].Compressed()
	return rc, errors.Wrapf(err, "failed to fetch archive layer of %q", uri)
}
