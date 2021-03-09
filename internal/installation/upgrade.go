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

	"github.com/pkg/errors"
	"k8s.io/klog"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/installation/receipt"
	"sigs.k8s.io/krew/internal/installation/semver"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

// Upgrade will reinstall and delete the old plugin. The operation tries
// to not get the plugin dir in a bad state if it fails during the process.
func Upgrade(p environment.Paths, plugin index.Plugin, indexName string) error {
	installReceipt, err := receipt.Load(p.PluginInstallReceiptPath(plugin.Name))
	if err != nil {
		return errors.Wrapf(err, "failed to load install receipt for plugin %q", plugin.Name)
	}

	curVersion := installReceipt.Spec.Version
	curv, err := semver.Parse(curVersion)
	if err != nil {
		return errors.Wrapf(err, "failed to parse installed plugin version (%q) as a semver value", curVersion)
	}

	// Find available installation candidate
	candidate, ok, err := GetMatchingPlatform(plugin.Spec.Platforms)
	if err != nil {
		return errors.Wrap(err, "failed trying to find a matching platform in plugin spec")
	}
	if !ok {
		return errors.Errorf("plugin %q does not offer installation for this platform (%s)",
			plugin.Name, OSArch())
	}

	newVersion := plugin.Spec.Version
	newv, err := semver.Parse(newVersion)
	if err != nil {
		return errors.Wrapf(err, "failed to parse candidate version spec (%q)", newVersion)
	}
	klog.V(2).Infof("Comparing versions: current=%s target=%s", curv, newv)

	// See if it's a newer version
	if !semver.Less(curv, newv) {
		klog.V(3).Infof("Plugin does not need upgrade (%s ≥ %s)", curv, newv)
		return ErrIsAlreadyUpgraded
	}
	klog.V(1).Infof("Plugin needs upgrade (%s < %s)", curv, newv)

	// Re-Install
	klog.V(1).Infof("Installing new version %s", newVersion)
	if err := install(installOperation{
		pluginName: plugin.Name,
		platform:   candidate,

		installDir: p.PluginVersionInstallPath(plugin.Name, newVersion),
		binDir:     p.BinPath(),
	}, InstallOpts{}); err != nil {
		return errors.Wrap(err, "failed to install new version")
	}

	klog.V(2).Infof("Upgrading install receipt for plugin %s", plugin.Name)
	if err = receipt.Store(receipt.New(plugin, indexName, installReceipt.CreationTimestamp), p.PluginInstallReceiptPath(plugin.Name)); err != nil {
		return errors.Wrap(err, "installation receipt could not be stored, uninstall may fail")
	}

	// Clean old installations
	klog.V(2).Infof("Starting old version cleanup")
	return cleanupInstallation(p, plugin, curVersion)
}

// cleanupInstallation will remove a plugin directly if it not krew.
//
// Krew on Windows needs special care because active directories can't be
// deleted. This method will mark old krew versions and during next run clean
// the directory.
func cleanupInstallation(p environment.Paths, plugin index.Plugin, oldVersion string) error {
	if plugin.Name == constants.KrewPluginName && IsWindows() {
		klog.V(1).Infof("not removing old version of krew during upgrade on windows (should be cleaned up on the next run)")
		return nil
	}

	klog.V(1).Infof("Remove old plugin installation under %q", p.PluginVersionInstallPath(plugin.Name, oldVersion))
	return os.RemoveAll(p.PluginVersionInstallPath(plugin.Name, oldVersion))
}
