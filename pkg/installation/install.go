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

func downloadAndMove(version, sha256sum, uri string, fos []index.FileOperation, downloadPath, installPath, forceDownloadFile string) (dst string, err error) {
	glog.V(3).Infof("Creating download dir %q", downloadPath)
	if err = os.MkdirAll(downloadPath, 0755); err != nil {
		return "", errors.Wrapf(err, "could not create download path %q", downloadPath)
	}
	defer os.RemoveAll(downloadPath)

	var fetcher download.Fetcher = download.HTTPFetcher{}
	if forceDownloadFile != "" {
		fetcher = download.NewFileFetcher(forceDownloadFile)
	}

	verifier := download.NewSha256Verifier(sha256sum)
	if err := download.NewDownloader(verifier, fetcher).Get(uri, downloadPath); err != nil {
		return "", errors.Wrap(err, "failed to download and verify file")
	}
	return moveToInstallDir(downloadPath, installPath, version, fos)
}

// Install will download and install a plugin. The operation tries
// to not get the plugin dir in a bad state if it fails during the process.
func Install(p environment.Paths, plugin index.Plugin, forceDownloadFile string) error {
	glog.V(2).Infof("Looking for installed versions")
	_, err := receipt.Load(p.PluginInstallReceiptPath(plugin.Name))
	if err == nil {
		return ErrIsAlreadyInstalled
	} else if !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to look up plugin receipt")
	}

	glog.V(1).Infof("Finding download target for plugin %s", plugin.Name)
	version, sha256, uri, fos, bin, err := getDownloadTarget(plugin)
	if err != nil {
		return err
	}

	// The actual install should be the last action so that a failure during receipt
	// saving does not result in an installed plugin without receipt.
	glog.V(3).Infof("Install plugin %s", plugin.Name)
	if err := install(plugin.Name, version, sha256, uri, bin, p, fos, forceDownloadFile); err != nil {
		return errors.Wrap(err, "install failed")
	}
	glog.V(3).Infof("Storing install receipt for plugin %s", plugin.Name)
	err = receipt.Store(plugin, p.PluginInstallReceiptPath(plugin.Name))
	return errors.Wrap(err, "installation receipt could not be stored, uninstall may fail")
}

func install(plugin, version, sha256sum, uri, bin string, p environment.Paths, fos []index.FileOperation, forceDownloadFile string) error {
	dst, err := downloadAndMove(version, sha256sum, uri, fos, filepath.Join(p.DownloadPath(), plugin), p.PluginInstallPath(plugin), forceDownloadFile)
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
		glog.Errorf("Removing krew through krew is not supported.")
		if !isWindows() { // assume POSIX-like
			glog.Errorf("If you’d like to uninstall krew altogether, run:\n\trm -rf -- %q", p.BasePath())
		}
		return errors.New("self-uninstall not allowed")
	}
	glog.V(3).Infof("Finding installed version to delete")

	if _, err := receipt.Load(p.PluginInstallReceiptPath(name)); err != nil {
		if os.IsNotExist(err) {
			return ErrIsNotInstalled
		}
		return errors.Wrapf(err, "failed to look up install receipt for plugin %q", name)
	}

	glog.V(1).Infof("Deleting plugin %s", name)

	symlinkPath := filepath.Join(p.BinPath(), pluginNameToBin(name, isWindows()))
	glog.V(3).Infof("Unlink %q", symlinkPath)
	if err := removeLink(symlinkPath); err != nil {
		return errors.Wrap(err, "could not uninstall symlink of plugin")
	}

	pluginInstallPath := p.PluginInstallPath(name)
	glog.V(3).Infof("Deleting path %q", pluginInstallPath)
	if err := os.RemoveAll(pluginInstallPath); err != nil {
		return errors.Wrapf(err, "could not remove plugin directory %q", pluginInstallPath)
	}
	pluginReceiptPath := p.PluginInstallReceiptPath(name)
	glog.V(3).Infof("Deleting plugin receipt %q", pluginReceiptPath)
	err := os.Remove(pluginReceiptPath)
	return errors.Wrapf(err, "could not remove plugin receipt %q", pluginReceiptPath)
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
