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
	"io/ioutil"
	"os"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"sigs.k8s.io/krew/pkg/environment"
	"sigs.k8s.io/krew/pkg/index"
	"sigs.k8s.io/krew/pkg/installation/receipt"
	"sigs.k8s.io/krew/pkg/installation/semver"
)

// Upgrade will reinstall and delete the old plugin. The operation tries
// to not get the plugin dir in a bad state if it fails during the process.
func Upgrade(p environment.Paths, plugin index.Plugin) error {
	installReceipt, err := receipt.Load(p.PluginInstallReceiptPath(plugin.Name))
	if err != nil {
		return errors.Wrapf(err, "failed to load install receipt for plugin %q", plugin.Name)
	}

	curVersion := installReceipt.Spec.Version
	curv, err := semver.Parse(curVersion)
	if err != nil {
		return errors.Wrapf(err, "failed to parse installed version (%q) plugin %q as semantic version", curVersion, plugin.Name)
	}

	// Find available installation candidate
	newVersion, sha256sum, uri, fos, binName, err := getDownloadTarget(plugin)
	if err != nil {
		return errors.Wrap(err, "failed to get the current download target")
	}
	newv, err := semver.Parse(newVersion)
	if err != nil {
		return errors.Wrapf(err, "failed to parse installation candidate version spec (%q) for plugin %q", newVersion, plugin.Name)
	}
	glog.V(2).Infof("Comparing versions: current=%s target=%s", curv, newv)

	// See if it's a newer version
	if !semver.Less(curv, newv) {
		glog.V(3).Infof("Plugin does not need upgrade (%s ≥ %s)", curv, newv)
		return ErrIsAlreadyUpgraded
	}
	glog.V(1).Infof("Plugin needs upgrade (%s < %s)", curv, newv)

	glog.V(2).Infof("Upgrading install receipt for plugin %s", plugin.Name)
	if err = receipt.Store(plugin, p.PluginInstallReceiptPath(plugin.Name)); err != nil {
		return errors.Wrap(err, "installation receipt could not be stored, uninstall may fail")
	}

	// Re-Install
	glog.V(1).Infof("Installing new version %s", newVersion)
	if err := install(plugin.Name, newVersion, sha256sum, uri, binName, p, fos, ""); err != nil {
		return errors.Wrap(err, "failed to install new version")
	}

	// Clean old installations
	glog.V(4).Infof("Starting old version cleanup")
	return removePluginVersionFromFS(p, plugin, newVersion, curVersion)
}

// removePluginVersionFromFS will remove a plugin directly if it not krew.

// Krew on Windows needs special care because active directories can't be
// deleted. This method will unlink old krew versions and during next run clean
// the directory.
func removePluginVersionFromFS(p environment.Paths, plugin index.Plugin, newVersion, oldVersion string) error {
	// Cleanup if we haven't updated krew during this execution.
	if plugin.Name == krewPluginName {
		glog.V(1).Infof("Handling removal for older version of krew")
		execPath, err := os.Executable()
		if err != nil {
			return errors.Wrap(err, "could not get krew's own executable path")
		}
		executedKrewVersion, _, err := environment.GetExecutedVersion(p.InstallPath(), execPath, environment.Realpath)
		if err != nil {
			return errors.Wrap(err, "failed to find current krew version")
		}
		glog.V(1).Infof("Detected running krew version=%s", executedKrewVersion)
		return handleKrewRemove(p, plugin, newVersion, executedKrewVersion)
	}

	glog.V(1).Infof("Remove old plugin installation under %q", p.PluginVersionInstallPath(plugin.Name, oldVersion))
	return os.RemoveAll(p.PluginVersionInstallPath(plugin.Name, oldVersion))
}

// handleKrewRemove will remove and unlink old krew versions.
func handleKrewRemove(p environment.Paths, plugin index.Plugin, newVersion, currentKrewVersion string) error {
	dir, err := ioutil.ReadDir(p.PluginInstallPath(plugin.Name))
	if err != nil {
		return errors.Wrap(err, "can't read plugin dir")
	}
	for _, f := range dir {
		pluginVersionPath := p.PluginVersionInstallPath(plugin.Name, f.Name())
		if !f.IsDir() {
			continue
		}
		// Delete old dir
		if f.Name() != newVersion && f.Name() != currentKrewVersion {
			glog.V(1).Infof("Remove old krew installation under %q", pluginVersionPath)
			if err = os.RemoveAll(pluginVersionPath); err != nil {
				return errors.Wrapf(err, "can't remove plugin oldVersion=%q, path=%q", f.Name(), pluginVersionPath)
			}
		} else if f.Name() != newVersion {
			glog.V(1).Infof("Unlink krew installation under %q", pluginVersionPath)
			// TODO(ahmetb,lbb) is this part implemented???
		}
	}
	return nil
}
