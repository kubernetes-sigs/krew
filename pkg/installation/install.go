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
	"os"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/google/krew/pkg/download"
	"github.com/google/krew/pkg/environment"
	"github.com/google/krew/pkg/index"
)

func downloadAndMove(version, uri string, fos []index.FileOperation, downloadPath, installPath string) (err error) {
	if err = os.MkdirAll(downloadPath, os.ModePerm); err != nil {
		return fmt.Errorf("could not create download path %q, err: %v", downloadPath, err)
	}
	defer os.RemoveAll(downloadPath)

	if version == "HEAD" {
		err = download.GetInsecure(uri, downloadPath, download.HTTPFetcher{})
	} else {
		err = download.GetWithSha256(uri, downloadPath, version, download.HTTPFetcher{})
	}
	if err != nil {
		return err
	}

	return moveToInstallAtomic(downloadPath, installPath, version, fos)
}

// Install will download and install a plugin. The operation tries
// to not get the plugin dir in a bad satate if it failes during the process.
func Install(p environment.KrewPaths, index index.Plugin, forceHEAD bool) error {
	_, ok, err := findInstalledPluginVersion(p.Install, index.Name)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("can't install plugin %q, it is already installed", index.Name)
	}
	version, uri, fos, err := getDownloadTarget(index, forceHEAD)
	if err != nil {
		return err
	}
	return downloadAndMove(version, uri, fos, filepath.Join(p.Download, index.Name), filepath.Join(p.Install, index.Name))
}

// Upgrade will reinstall and delete the old plugin. The operation tries
// to not get the plugin dir in a bad satate if it failes during the process.
func Upgrade(p environment.KrewPaths, index index.Plugin) error {
	version, ok, err := findInstalledPluginVersion(p.Install, index.Name)
	if err != nil {
		return fmt.Errorf("could not detect installed plugin version, err: %v", err)
	}
	if !ok {
		return fmt.Errorf("can't upgrade plugin %q, it is not installed", index.Name)
	}
	if version == "HEAD" {
		oldHEADPath, newHEADPath := filepath.Join(p.Install, index.Name, "HEAD"), filepath.Join(p.Install, index.Name, "HEAD-OLD")
		glog.V(2).Infof("Move old HEAD from: %q -> %q", oldHEADPath, newHEADPath)
		if err = os.Rename(oldHEADPath, newHEADPath); err != nil {
			return fmt.Errorf("failed to rename HEAD -> HEAD-OLD, from %q to %q, err: %v", oldHEADPath, newHEADPath, err)
		}
		version = "HEAD-OLD"
	}

	// Re-Install
	newVersion, uri, fos, err := getDownloadTarget(index, version == "HEAD-OLD")
	if version == newVersion {
		return fmt.Errorf("can't upgrade to version %q as it is the current version", newVersion)
	}
	if err != nil {
		return err
	}
	if err = downloadAndMove(newVersion, uri, fos, filepath.Join(p.Download, index.Name), filepath.Join(p.Install, index.Name)); err != nil {
		return err
	}

	// Cleanup
	return os.RemoveAll(filepath.Join(p.Install, index.Name, version))
}

// Remove will remove a plugin without bringing the plugin dir in a
// bad state.
func Remove(p environment.KrewPaths, name string) error {
	_, installed, err := findInstalledPluginVersion(p.Install, name)
	if err != nil {
		return fmt.Errorf("can't remove plugin, err: %v", err)
	}
	if !installed {
		return fmt.Errorf("can't remove plugin %q, it is not installed", name)
	}
	return os.RemoveAll(filepath.Join(p.Install, name))
}
