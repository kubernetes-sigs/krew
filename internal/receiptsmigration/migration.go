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

// todo(corneliusweig) remove migration code with v0.4
package receiptsmigration

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/klog"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/index/validation"
	"sigs.k8s.io/krew/internal/receiptsmigration/oldenvironment"
	"sigs.k8s.io/krew/pkg/constants"
)

const (
	krewPluginName = "krew"
)

// Done checks if the krew installation requires a migration.
// It considers a migration necessary when plugins are installed, but no receipts are present.
func Done(newPaths environment.Paths) (bool, error) {
	receipts, err := ioutil.ReadDir(newPaths.InstallReceiptsPath())
	if err != nil {
		return false, err
	}
	plugins, err := ioutil.ReadDir(newPaths.BinPath())
	if err != nil {
		return false, err
	}

	hasInstalledPlugins := len(plugins) > 0
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
		klog.Infoln("Already migrated")
		return nil
	}

	oldPaths := oldenvironment.MustGetKrewPaths()
	installed, err := getPluginsToReinstall(oldPaths, newPaths)
	if err != nil {
		return errors.Wrapf(err, "failed to build list of plugins")
	}

	klog.Infoln("These plugins will be reinstalled: ", installed)

	// krew must be skipped by the normal migration logic
	if err := copyKrewManifest(newPaths.IndexPluginsPath(""), newPaths.InstallReceiptsPath()); err != nil {
		return errors.Wrapf(err, "failed to copy krew manifest")
	}

	// point of no return: keep on going when encountering errors
	for _, plugin := range installed {
		if err := uninstall(oldPaths, plugin); err != nil {
			klog.Infof("Uninstalling of %s failed, skipping reinstall", plugin)
			continue
		}

		if err := reinstall(plugin); err != nil {
			klog.Infof("Reinstalling %s failed", plugin)
		}
	}

	return nil
}

func copyKrewManifest(srcFolder, dstFolder string) error {
	manifestName := "krew" + constants.ManifestExtension
	src, err := os.Open(filepath.Join(srcFolder, manifestName))
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(filepath.Join(dstFolder, manifestName))
	if err != nil {
		return err
	}
	defer dst.Close()

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
		if !fileInfo.IsDir() || !validation.IsSafePluginName(plugin) || plugin == krewPluginName {
			continue
		}
		if !isAvailableInIndex(newPaths, plugin) {
			klog.Infof("Skipping plugin %s, because it is missing in the index", plugin)
			continue
		}
		renewable = append(renewable, plugin)
	}
	return renewable, nil
}

// isAvailableInIndex checks that the given plugin is available in the index
func isAvailableInIndex(paths environment.Paths, plugin string) bool {
	pluginYaml := filepath.Join(paths.IndexPluginsPath(""), plugin+constants.ManifestExtension)
	_, err := os.Lstat(pluginYaml)
	return err == nil
}

// uninstall will uninstall a plugin in the old krew home layout.
func uninstall(p oldenvironment.Paths, name string) error {
	if name == krewPluginName {
		return errors.Errorf("removing krew is not allowed through krew. Please run:\n\t rm -r %s", p.BasePath())
	}
	klog.Infof("Uninstalling %s", name)

	symlinkPath := filepath.Join(p.BinPath(), pluginNameToBin(name, isWindows()))
	klog.V(3).Infof("Unlink %q", symlinkPath)
	if err := removeLink(symlinkPath); err != nil {
		return errors.Wrap(err, "could not uninstall symlink of plugin")
	}

	pluginInstallPath := p.PluginInstallPath(name)
	klog.V(3).Infof("Deleting path %q", pluginInstallPath)
	return errors.Wrapf(os.RemoveAll(pluginInstallPath), "could not remove plugin directory %q", pluginInstallPath)
}

// reinstall shells out to `krew` to install the given plugin.
func reinstall(plugin string) error {
	klog.Infoln("Re-installing", plugin)
	cmd := exec.Command("kubectl", "krew", "install", plugin)
	output, err := cmd.CombinedOutput()
	if err != nil {
		klog.Info(string(output))
	}
	return err
}

// removeLink removes a symlink reference if exists.
// same as pkg/installation/install.go:167
func removeLink(path string) error {
	fi, err := os.Lstat(path)
	if os.IsNotExist(err) {
		klog.V(3).Infof("No file found at %q", path)
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
	klog.V(3).Infof("Removed symlink from %q", path)
	return nil
}

// same as pkg/installation/install.go:186
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
	name = strings.ReplaceAll(name, "-", "_")
	name = "kubectl-" + name
	if isWindows {
		name += ".exe"
	}
	return name
}
