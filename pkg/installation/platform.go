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
	"runtime"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"sigs.k8s.io/krew/pkg/index"
)

// GetMatchingPlatform finds the platform spec in the specified plugin that
// matches the OS/arch of the current machine (can be overridden via KREW_OS
// and/or KREW_ARCH).
func GetMatchingPlatform(platforms []index.Platform) (index.Platform, bool, error) {
	os, arch := osArch()
	return matchPlatform(platforms, os, arch)
}

// matchPlatform returns the first matching platform to given os/arch.
func matchPlatform(platforms []index.Platform, os, arch string) (index.Platform, bool, error) {
	envLabels := labels.Set{
		"os":   os,
		"arch": arch,
	}
	glog.V(2).Infof("Matching platform for labels(%v)", envLabels)

	for i, platform := range platforms {
		sel, err := metav1.LabelSelectorAsSelector(platform.Selector)
		if err != nil {
			return index.Platform{}, false, errors.Wrap(err, "failed to compile label selector")
		}
		if sel.Matches(envLabels) {
			glog.V(2).Infof("Found matching platform with index (%d)", i)
			return platform, true, nil
		}
	}
	return index.Platform{}, false, nil
}

// osArch returns the OS/arch combination to be used on the current system. It
// can be overridden by setting KREW_OS and/or KREW_ARCH environment variables.
func osArch() (string, string) {
	goos, goarch := runtime.GOOS, runtime.GOARCH
	envOS, envArch := os.Getenv("KREW_OS"), os.Getenv("KREW_ARCH")
	if envOS != "" {
		goos = envOS
	}
	if envArch != "" {
		goarch = envArch
	}
	return goos, goarch
}
