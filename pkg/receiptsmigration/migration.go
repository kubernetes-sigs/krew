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

package receiptsmigration

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/environment"
	"sigs.k8s.io/krew/pkg/index"
	"sigs.k8s.io/krew/pkg/receiptsmigration/oldenvironment"
)

const (
	krewPluginName = "krew"
)

// Done checks if the krew installation requires a migration.
// It considers a migration necessary when plugins are installed, but no receipts are present.
func Done(newPaths environment.Paths) (bool, error) {
	receipts, err := ioutil.ReadDir(newPaths.InstallReceiptPath())
	if err != nil {
		return false, err
	}
	store, err := ioutil.ReadDir(newPaths.InstallPath())
	if err != nil {
		return false, err
	}

	var hasInstalledPlugins bool
	for _, entry := range store {
		if entry.IsDir() {
			hasInstalledPlugins = true
			break
		}
	}

	hasNoReceipts := len(receipts) == 0

	return !(hasInstalledPlugins && hasNoReceipts), nil
}

// Migrate searches for installed plugins, removes each plugin and reinstalls afterwards.
// Once started, it keeps going even if there are errors.
func Migrate(newPaths environment.Paths) error {
	isMigrated, err := Done(newPaths)
	if err != nil {
		return err
	}
	if isMigrated {
		glog.Infoln("Already migrated")
		return nil
	}

	oldPaths := oldenvironment.MustGetKrewPaths()
	installed, err := getPluginsToReinstall(oldPaths, newPaths)
	if err != nil {
		return errors.Wrapf(err, "failed to build list of plugins")
	}

	glog.Infoln("These plugins will be reinstalled: ", installed)
	if err := os.MkdirAll(newPaths.InstallReceiptPath(), 0755); err != nil {
		return errors.Wrapf(err, "failed to create directory %q", newPaths.InstallReceiptPath())
	}

	// krew must be skipped by the normal migration logic
	if err := copyKrewManifest(newPaths.IndexPluginsPath(), newPaths.InstallReceiptPath()); err != nil {
		return errors.Wrapf(err, "failed to copy krew manifest")
	}

	for _, plugin := range installed {
		if consistent := isInstallationConsistent(oldPaths.BinPath(), plugin); !consistent {
			glog.Infof("Skipping inconsistent plugin installation %s.", plugin)
			continue
		}

		if err := uninstall(oldPaths, plugin); err != nil {
			glog.Infof("Uninstalling of %s failed, skipping reinstall", plugin)
			continue
		}

		if err := reinstall(plugin); err != nil {
			glog.Infof("Reinstalling %s failed", plugin)
		}
	}

	return nil
}

func copyKrewManifest(srcFolder, dstFolder string) error {
	manifestName := "krew" + constants.ManifestExtension
	src, err := os.Open(filepath.Join(srcFolder, manifestName))
	defer src.Close()
	if err != nil {
		return err
	}

	dst, err := os.Create(filepath.Join(dstFolder, manifestName))
	defer dst.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	return err
}

// getPluginsToReinstall collects a list of installed plugins which can be reinstalled.
func getPluginsToReinstall(oldPaths oldenvironment.Paths, newPaths environment.Paths) ([]string, error) {
	store := oldPaths.InstallPath()
	fileInfos, err := ioutil.ReadDir(store)
	if err != nil {
		return nil, err
	}

	renewable := []string{}
	for _, fileInfo := range fileInfos {
		plugin := fileInfo.Name()
		if !fileInfo.IsDir() || !index.IsSafePluginName(plugin) || plugin == krewPluginName {
			continue
		}
		if !isAvailableInIndex(newPaths, plugin) {
			glog.Infof("Skipping plugin %s, because it is missing in the index", plugin)
			continue
		}
		renewable = append(renewable, plugin)
	}
	return renewable, nil
}

// isAvailableInIndex checks that the given plugin is available in the index
func isAvailableInIndex(paths environment.Paths, plugin string) bool {
	pluginYaml := filepath.Join(paths.IndexPluginsPath(), plugin+constants.ManifestExtension)
	_, err := os.Lstat(pluginYaml)
	return err == nil
}

// uninstall will uninstall a plugin in the old krew home layout.
func uninstall(p oldenvironment.Paths, name string) error {
	if name == krewPluginName {
		return errors.Errorf("removing krew is not allowed through krew. Please run:\n\t rm -r %s", p.BasePath())
	}
	glog.Infof("Uninstalling %s", name)

	symlinkPath := filepath.Join(p.BinPath(), pluginNameToBin(name, isWindows()))
	glog.V(3).Infof("Unlink %q", symlinkPath)
	if err := removeLink(symlinkPath); err != nil {
		return errors.Wrap(err, "could not uninstall symlink of plugin")
	}

	pluginInstallPath := p.PluginInstallPath(name)
	glog.V(3).Infof("Deleting path %q", pluginInstallPath)
	return errors.Wrapf(os.RemoveAll(pluginInstallPath), "could not remove plugin directory %q", pluginInstallPath)
}

// reinstall shells out to `krew` to install the given plugin.
func reinstall(plugin string) error {
	glog.Infoln("Re-installing", plugin)
	cmd := exec.Command("kubectl", "krew", "install", plugin)
	output, err := cmd.CombinedOutput()
	if err != nil {
		glog.Info(string(output))
	}
	return err
}

// isInstallationConsistent checks if a plugin is linked in the bin directory.
func isInstallationConsistent(binDir, pluginName string) bool {
	if _, err := os.Readlink(filepath.Join(binDir, pluginNameToBin(pluginName, isWindows()))); err != nil {
		return false
	}
	return true
}

// removeLink removes a symlink reference if exists.
// same as pkg/installation/install.go:167
func removeLink(path string) error {
	fi, err := os.Lstat(path)
	if os.IsNotExist(err) {
		glog.V(3).Infof("No file found at %q", path)
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "failed to read the symlink in %q", path)
	}

	if fi.Mode()&os.ModeSymlink == 0 {
		return errors.Errorf("file %q is not a symlink (mode=%s)", path, fi.Mode())
	}
	if err := os.Remove(path); err != nil {
		return errors.Wrapf(err, "failed to remove the symlink in %q", path)
	}
	glog.V(3).Infof("Removed symlink from %q", path)
	return nil
}

/// same as pkg/installation/install.go:186
func isWindows() bool {
	goos := runtime.GOOS
	if env := os.Getenv("KREW_OS"); env != "" {
		goos = env
	}
	return goos == "windows"
}

// pluginNameToBin creates the name of the symlink file for the plugin name.
// It converts dashes to underscores.
// same as pkg/installation/install.go:196
func pluginNameToBin(name string, isWindows bool) string {
	name = strings.Replace(name, "-", "_", -1)
	name = "kubectl-" + name
	if isWindows {
		name += ".exe"
	}
	return name
}
