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

	"io/ioutil"

	"github.com/GoogleContainerTools/krew/pkg/environment"
	"github.com/GoogleContainerTools/krew/pkg/index"
	"github.com/golang/glog"
)

// Upgrade will reinstall and delete the old plugin. The operation tries
// to not get the plugin dir in a bad state if it fails during the process.
func Upgrade(p environment.KrewPaths, plugin index.Plugin, currentKrewVersion string) error {
	oldVersion, ok, err := findInstalledPluginVersion(p.Install, p.Bin, plugin.Name)
	if err != nil {
		return fmt.Errorf("could not detect installed plugin oldVersion, err: %v", err)
	}
	if !ok {
		return fmt.Errorf("can't upgrade plugin %q, it is not installed", plugin.Name)
	}

	// Check allowed installation
	newVersion, uri, fos, binName, err := getDownloadTarget(plugin, oldVersion == headVersion)
	if oldVersion == newVersion && oldVersion != headVersion {
		return ErrIsAlreadyUpgraded
	}
	if err != nil {
		return fmt.Errorf("failed to get the current download target, err: %v", err)
	}

	// Move head to save location
	if oldVersion == headVersion {
		oldHEADPath, newHEADPath := filepath.Join(p.Install, plugin.Name, headVersion), filepath.Join(p.Install, plugin.Name, headOldVersion)
		glog.V(2).Infof("Move old HEAD from: %q to %q", oldHEADPath, newHEADPath)
		if err = os.Rename(oldHEADPath, newHEADPath); err != nil {
			return fmt.Errorf("failed to rename HEAD to HEAD-OLD, from %q to %q, err: %v", oldHEADPath, newHEADPath, err)
		}
		oldVersion = headOldVersion
	}

	// Re-Install
	glog.V(1).Infof("Installing new version %s", newVersion)
	if err := install(plugin.Name, newVersion, uri, binName, p, fos); err != nil {
		return fmt.Errorf("failed to install new version, err: %v", err)
	}

	// Clean old installations
	glog.V(4).Infof("Starting old version cleanup")
	return removePluginVersionFromFS(p, plugin, newVersion, oldVersion, currentKrewVersion)
}

// removePluginVersionFromFS will remove a plugin directly if it not krew. Krew on Windows needs special care
// because active directories can't be deleted. This method will unlink old krew versions and during next run clean
// the directory.
func removePluginVersionFromFS(p environment.KrewPaths, plugin index.Plugin, newVersion, oldVersion, currentKrewVersion string) error {
	// Cleanup if we haven't updated krew during this execution.
	if plugin.Name == krewPluginName {
		return handleKrewRemove(p, plugin, newVersion, oldVersion, currentKrewVersion)
	}
	glog.V(1).Infof("Remove old plugin installation under %q", filepath.Join(p.Install, plugin.Name, oldVersion))
	return os.RemoveAll(filepath.Join(p.Install, plugin.Name, oldVersion))
}

// handleKrewRemove will remove and unlink old krew versions.
func handleKrewRemove(p environment.KrewPaths, plugin index.Plugin, newVersion, oldVersion, currentKrewVersion string) error {
	dir, err := ioutil.ReadDir(filepath.Join(p.Install, plugin.Name))
	if err != nil {
		return fmt.Errorf("can't read plugin dir, err: %v", err)
	}
	for _, f := range dir {
		pluginVersionPath := filepath.Join(p.Install, plugin.Name, f.Name())
		if !f.IsDir() {
			continue
		}
		// Delete old dir
		if f.Name() != newVersion && f.Name() != currentKrewVersion {
			glog.V(1).Infof("Remove old krew installation under %q", pluginVersionPath)
			if err = os.RemoveAll(pluginVersionPath); err != nil {
				return fmt.Errorf("can't remove plugin oldVersion=%q, path=%q, err: %v", f.Name(), pluginVersionPath, err)
			}
		} else if f.Name() != newVersion {
			glog.V(1).Infof("Unlink krew installation under %q", pluginVersionPath)
		}
	}
	return nil
}
