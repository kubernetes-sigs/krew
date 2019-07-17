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

package installation

import (
	"path/filepath"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
	"sigs.k8s.io/krew/pkg/pathutil"
	"sigs.k8s.io/krew/pkg/receipt"
)

func pluginVersionFromPath(installPath, pluginPath string) (string, error) {
	// plugin path: {install_path}/{plugin_name}/{version}/...
	elems, ok := pathutil.IsSubPath(installPath, pluginPath)
	if !ok || len(elems) < 2 {
		return "", errors.Errorf("failed to get the version from execution path=%q, with install path=%q", pluginPath, installPath)
	}
	return elems[1], nil
}

func getDownloadTarget(index index.Plugin) (version, sha256sum, uri string, fos []index.FileOperation, bin string, err error) {
	// TODO(ahmetb): We have many return values from this method, indicating
	// code smell. More specifically we return all-or-nothing, so ideally this
	// should be converted into a struct, like InstallOperation{} contains all
	// the data needed to install a plugin.
	p, ok, err := index.Spec.GetMatchingPlatform()
	if err != nil {
		return "", "", "", nil, p.Bin, errors.Wrap(err, "failed to get matching platforms")
	}
	if !ok {
		return "", "", "", nil, p.Bin, errors.New("no matching platform found")
	}
	version = index.Spec.Version
	uri = p.URI
	sha256sum = p.Sha256
	glog.V(4).Infof("found a matching platform, version=%s checksum=%s", version, sha256sum)
	return version, sha256sum, uri, p.Files, p.Bin, nil
}

// ListInstalledPlugins returns a list of all install plugins in a
// name:version format based on the install receipts at the specified dir.
func ListInstalledPlugins(receiptsDir string) (map[string]string, error) {
	// TODO(ahmetb): Write unit tests for this method. Currently blocked by
	// lack of an in-memory recipt object (issue#270) that we can use to save
	// receipts to a tempdir that can be read from unit tests.

	matches, err := filepath.Glob(filepath.Join(receiptsDir, "*"+constants.ManifestExtension))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to grab receipts directory (%s) for manifests", receiptsDir)
	}
	glog.V(4).Infof("Found %d install receipts in %s", len(matches), receiptsDir)
	installed := make(map[string]string)
	for _, m := range matches {
		r, err := receipt.Load(m)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse plugin install receipt %s", m)
		}
		glog.V(4).Infof("parsed receipt for %s: version=%s", r.GetObjectMeta().GetName(), r.Spec.Version)
		installed[r.GetObjectMeta().GetName()] = r.Spec.Version
	}
	return installed, nil
}
