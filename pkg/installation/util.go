// Copyright © 2018 Google Inc.
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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/golang/glog"
	"github.com/google/krew/pkg/index"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func getMatchingPlatform(i index.Plugin) (index.Platform, bool, error) {
	return matchPlatformToSystemEnvs(i, runtime.GOOS, runtime.GOARCH)
}

func matchPlatformToSystemEnvs(i index.Plugin, os, arch string) (index.Platform, bool, error) {
	envLabels := labels.Set{
		"os":   os,
		"arch": arch,
	}
	glog.V(3).Infof("Matching for labels(%v)", envLabels)
	for _, platform := range i.Spec.Platforms {
		sel, err := metav1.LabelSelectorAsSelector(platform.Selector)
		if err != nil {
			return index.Platform{}, false, fmt.Errorf("failed to compile label selector, err: %v", err)
		}
		if sel.Matches(envLabels) {
			return platform, true, nil
		}
	}
	return index.Platform{}, false, nil
}

func findInstalledPluginVersion(installPath, pluginname string) (name string, installed bool, err error) {
	if !indexscanner.IsSafePluginName(pluginname) {
		return "", false, fmt.Errorf("the plugin name %q is not allowed", pluginname)
	}
	fis, err := ioutil.ReadDir(filepath.Join(installPath, pluginname))
	if os.IsNotExist(err) {
		return "", false, nil
	} else if err != nil {
		return "", false, fmt.Errorf("could not read direcory, err: %v", err)
	}
	for _, fi := range fis {
		if fi.IsDir() {
			return fi.Name(), true, nil
		}
	}
	return "", false, nil
}

}
func getPluginVersion(p index.Platform, forceHEAD bool) (version, uri string, err error) {
	if (forceHEAD && p.Head != "") || (p.Head != "" && p.Sha256 == "" && p.URI == "") {
		return "HEAD", p.Head, nil
	}
	if forceHEAD && p.Head == "" {
		return "", "", fmt.Errorf("can't force HEAD, with no HEAD specified")
	}
	return strings.ToLower(p.Sha256), p.URI, nil
}

func getDownloadTarget(index index.Plugin, forceHEAD bool) (version, uri string, fos []index.FileOperation, err error) {
	p, ok, err := getMatchingPlatform(index)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to get matching platforms, err: %v", err)
	}
	if !ok {
		return "", "", nil, fmt.Errorf("no matching platform found")
	}
	version, uri, err = getPluginVersion(p, forceHEAD)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to get the plugin version, err: %v", err)
	}

	return version, uri, p.Files, nil
}
