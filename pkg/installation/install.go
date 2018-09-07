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

	"github.com/GoogleContainerTools/krew/pkg/download"
	"github.com/GoogleContainerTools/krew/pkg/environment"
	"github.com/GoogleContainerTools/krew/pkg/index"
	"github.com/golang/glog"
)

// Plugin Lifecycle Errors
var (
	ErrIsAlreadyInstalled = fmt.Errorf("can't install, the newest version is already installed")
	ErrIsNotInstalled     = fmt.Errorf("plugin is not installed")
	ErrIsAlreadyUpgraded  = fmt.Errorf("can't upgrade, the newest version is already installed")
)

const (
	headVersion    = "HEAD"
	headOldVersion = "HEAD-OLD"
	krewPluginName = "krew"
)

func downloadAndMove(version, uri string, fos []index.FileOperation, downloadPath, installPath string) (err error) {
	glog.V(3).Infof("Creating download dir %q", downloadPath)
	if err = os.MkdirAll(downloadPath, 0755); err != nil {
		return fmt.Errorf("could not create download path %q, err: %v", downloadPath, err)
	}
	defer os.RemoveAll(downloadPath)

	if version == headVersion {
		glog.V(1).Infof("Getting latest version from HEAD")
		err = download.GetInsecure(uri, downloadPath, download.HTTPFetcher{})
	} else {
		glog.V(1).Infof("Getting sha256 (%s) signed version", version)
		err = download.GetWithSha256(uri, downloadPath, version, download.HTTPFetcher{})
	}
	if err != nil {
		return err
	}

	return moveToInstallDir(downloadPath, installPath, version, fos)
}

// Install will download and install a plugin. The operation tries
// to not get the plugin dir in a bad state if it fails during the process.
func Install(p environment.KrewPaths, plugin index.Plugin, forceHEAD bool) error {
	glog.V(2).Infof("Looking for installed versions")
	_, ok, err := findInstalledPluginVersion(p.Install, p.Bin, plugin.Name)
	if err != nil {
		return err
	}
	if ok {
		return ErrIsAlreadyInstalled
	}

	glog.V(1).Infof("Finding download target for plugin %s", plugin.Name)
	version, uri, fos, err := getDownloadTarget(plugin, forceHEAD)
	if err != nil {
		return err
	}
	return downloadAndMove(version, uri, fos, filepath.Join(p.Download, plugin.Name), filepath.Join(p.Install, plugin.Name))
}

// Remove will remove a plugin.
func Remove(p environment.KrewPaths, name string) error {
	if name == krewPluginName {
		return fmt.Errorf("removing krew is not allowed through krew, see docs for help")
	}
	glog.V(3).Infof("Finding installed version to delete")
	version, installed, err := findInstalledPluginVersion(p.Install, p.Bin, name)
	if err != nil {
		return fmt.Errorf("can't remove plugin, err: %v", err)
	}
	if !installed {
		return ErrIsNotInstalled
	}
	glog.V(1).Infof("Deleting plugin version %s", version)
	glog.V(3).Infof("Deleting path %q", filepath.Join(p.Install, name))
	return os.RemoveAll(filepath.Join(p.Install, name))
}

func createOrUpdateLink(bindir string, binary string) error {
	name := filepath.Base(binary)
	dst := filepath.Join(bindir, name)

	if err := removeLink(bindir, name); err != nil {
		return fmt.Errorf("failed to remove old symlink, err: %v", err)
	}
	// Create new
	if err := os.Symlink(dst, binary); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create a symlink form %q to %q, err: %v", bindir, dst, err)
	}
	glog.V(2).Infof("Created symlink in %q\n", dst)

	return nil
}

func removeLink(bindir string, plugin string) error {
	dst := filepath.Join(bindir, plugin)
	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove the symlink in %q, err: %v", dst, err)
	} else if err == nil {
		glog.V(3).Infof("Removed symlink from %q\n", dst)
	}
	return nil
}
