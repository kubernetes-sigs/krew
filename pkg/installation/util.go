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
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
	"sigs.k8s.io/krew/pkg/pathutil"
	"sigs.k8s.io/krew/pkg/receipt"
)

func findInstalledPluginVersion(installPath, binDir, pluginName string) (name string, installed bool, err error) {
	if !index.IsSafePluginName(pluginName) {
		return "", false, errors.Errorf("the plugin name %q is not allowed", pluginName)
	}
	glog.V(3).Infof("Searching for installed versions of %s in %q", pluginName, binDir)
	link, err := os.Readlink(filepath.Join(binDir, pluginNameToBin(pluginName, isWindows())))
	if os.IsNotExist(err) {
		return "", false, nil
	} else if err != nil {
		return "", false, errors.Wrap(err, "could not read plugin link")
	}

	if !filepath.IsAbs(link) {
		if link, err = filepath.Abs(filepath.Join(binDir, link)); err != nil {
			return "", true, errors.Wrapf(err, "failed to get the absolute path for the link of %q", link)
		}
	}

	name, err = pluginVersionFromPath(installPath, link)
	if err != nil {
		return "", true, errors.Wrap(err, "cloud not parse plugin version")
	}
	return name, true, nil
}

func pluginVersionFromPath(installPath, pluginPath string) (string, error) {
	// plugin path: {install_path}/{plugin_name}/{version}/...
	elems, ok := pathutil.IsSubPath(installPath, pluginPath)
	if !ok || len(elems) < 2 {
		return "", errors.Errorf("failed to get the version from execution path=%q, with install path=%q", pluginPath, installPath)
	}
	return elems[1], nil
}

func getPluginVersion(p index.Platform) (version, uri string) {
	return strings.ToLower(p.Sha256), p.URI
}

func getDownloadTarget(index index.Plugin) (version, uri string, fos []index.FileOperation, bin string, err error) {
	p, ok, err := index.Spec.GetMatchingPlatform()
	if err != nil {
		return "", "", nil, p.Bin, errors.Wrap(err, "failed to get matching platforms")
	}
	if !ok {
		return "", "", nil, p.Bin, errors.New("no matching platform found")
	}
	version, uri = getPluginVersion(p)
	glog.V(4).Infof("Matching plugin version is %s", version)

	return version, uri, p.Files, p.Bin, nil
}

// ListInstalledPlugins returns a list of all install plugins in a
// name:version format based on the install receipts at the specified dir.
func ListInstalledPlugins(receiptsDir string) (map[string]string, error) {
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
