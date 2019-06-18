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
	"runtime"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"sigs.k8s.io/krew/pkg/download"
	"sigs.k8s.io/krew/pkg/environment"
	"sigs.k8s.io/krew/pkg/index"
	"sigs.k8s.io/krew/pkg/pathutil"
	"sigs.k8s.io/krew/pkg/receipt"
)

// Plugin Lifecycle Errors
var (
	ErrIsAlreadyInstalled = errors.New("can't install, the newest version is already installed")
	ErrIsNotInstalled     = errors.New("plugin is not installed")
	ErrIsAlreadyUpgraded  = errors.New("can't upgrade, the newest version is already installed")
)

const (
	krewPluginName = "krew"
)

func downloadAndMove(version, uri string, fos []index.FileOperation, downloadPath, installPath, forceDownloadFile string) (dst string, err error) {
	glog.V(3).Infof("Creating download dir %q", downloadPath)
	if err = os.MkdirAll(downloadPath, 0755); err != nil {
		return "", errors.Wrapf(err, "could not create download path %q", downloadPath)
	}
	defer os.RemoveAll(downloadPath)

	var fetcher download.Fetcher = download.HTTPFetcher{}
	if forceDownloadFile != "" {
		fetcher = download.NewFileFetcher(forceDownloadFile)
	}

	verifier := download.NewSha256Verifier(version)
	if err := download.NewDownloader(verifier, fetcher).Get(uri, downloadPath); err != nil {
		return "", errors.Wrap(err, "failed to download and verify file")
	}
	return moveToInstallDir(downloadPath, installPath, version, fos)
}

// Install will download and install a plugin. The operation tries
// to not get the plugin dir in a bad state if it fails during the process.
func Install(p environment.Paths, plugin index.Plugin, forceDownloadFile string) error {
	glog.V(2).Infof("Looking for installed versions")
	_, ok, err := findInstalledPluginVersion(p.InstallPath(), p.BinPath(), plugin.Name)
	if err != nil {
		return err
	}
	if ok {
		return ErrIsAlreadyInstalled
	}

	glog.V(1).Infof("Finding download target for plugin %s", plugin.Name)
	version, uri, fos, bin, err := getDownloadTarget(plugin)
	if err != nil {
		return err
	}

	glog.V(2).Infof("Store install receipt for plugin %s", plugin.Name)
	if err = receipt.Store(plugin, p.PluginReceiptPath(plugin.Name)); err != nil {
		return errors.Wrap(err, "installation receipt could not be stored, uninstall may fail")
	}

	// The actual install should be the last action so that a failure during receipt
	// saving does not result in an installed plugin without receipt.
	glog.V(3).Infof("Install plugin %s", plugin.Name)
	err = install(plugin.Name, version, uri, bin, p, fos, forceDownloadFile)
	return errors.Wrap(err, "install failed")
}

func install(plugin, version, uri, bin string, p environment.Paths, fos []index.FileOperation, forceDownloadFile string) error {
	dst, err := downloadAndMove(version, uri, fos, filepath.Join(p.DownloadPath(), plugin), p.PluginInstallPath(plugin), forceDownloadFile)
	if err != nil {
		return errors.Wrap(err, "failed to download and move during installation")
	}

	subPathAbs, err := filepath.Abs(dst)
	if err != nil {
		return errors.Wrapf(err, "failed to get the absolute fullPath of %q", dst)
	}
	fullPath := filepath.Join(dst, filepath.FromSlash(bin))
	pathAbs, err := filepath.Abs(fullPath)
	if err != nil {
		return errors.Wrapf(err, "failed to get the absolute fullPath of %q", fullPath)
	}
	if _, ok := pathutil.IsSubPath(subPathAbs, pathAbs); !ok {
		return errors.Wrapf(err, "the fullPath %q does not extend the sub-fullPath %q", fullPath, dst)
	}
	return createOrUpdateLink(p.BinPath(), filepath.Join(dst, filepath.FromSlash(bin)), plugin)
}

// Uninstall will uninstall a plugin.
func Uninstall(p environment.Paths, name string) error {
	if name == krewPluginName {
		return errors.Errorf("removing krew is not allowed through krew. Please run:\n\t rm -r %s", p.BasePath())
	}
	glog.V(3).Infof("Finding installed version to delete")
	version, installed, err := findInstalledPluginVersion(p.InstallPath(), p.BinPath(), name)
	if err != nil {
		return errors.Wrap(err, "can't uninstall plugin")
	}
	if !installed {
		return ErrIsNotInstalled
	}
	glog.V(1).Infof("Deleting plugin version %s", version)
	glog.V(3).Infof("Deleting path %q", p.PluginInstallPath(name))

	symlinkPath := filepath.Join(p.BinPath(), pluginNameToBin(name, isWindows()))
	if err := removeLink(symlinkPath); err != nil {
		return errors.Wrap(err, "could not uninstall symlink of plugin")
	}
	return os.RemoveAll(p.PluginInstallPath(name))
}

func createOrUpdateLink(binDir string, binary string, plugin string) error {
	dst := filepath.Join(binDir, pluginNameToBin(plugin, isWindows()))

	if err := removeLink(dst); err != nil {
		return errors.Wrap(err, "failed to remove old symlink")
	}
	if _, err := os.Stat(binary); os.IsNotExist(err) {
		return errors.Wrapf(err, "can't create symbolic link, source binary (%q) cannot be found in extracted archive", binary)
	}

	// Create new
	glog.V(2).Infof("Creating symlink from %q to %q", binary, dst)
	if err := os.Symlink(binary, dst); err != nil {
		return errors.Wrapf(err, "failed to create a symlink form %q to %q", binDir, dst)
	}
	glog.V(2).Infof("Created symlink at %q", dst)

	return nil
}

// removeLink removes a symlink reference if exists.
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

func isWindows() bool {
	goos := runtime.GOOS
	if env := os.Getenv("KREW_OS"); env != "" {
		goos = env
	}
	return goos == "windows"
}

// pluginNameToBin creates the name of the symlink file for the plugin name.
// It converts dashes to underscores.
func pluginNameToBin(name string, isWindows bool) string {
	name = strings.Replace(name, "-", "_", -1)
	name = "kubectl-" + name
	if isWindows {
		name += ".exe"
	}
	return name
}
