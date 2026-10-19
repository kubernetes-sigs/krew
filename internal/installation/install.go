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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"sigs.k8s.io/krew/internal/download"
	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/installation/receipt"
	"sigs.k8s.io/krew/internal/pathutil"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

// InstallOpts specifies options for plugin installation operation.
type InstallOpts struct {
	ArchiveFileOverride string
	EnableNetrc         bool
	NetrcFile           string
}

type installOperation struct {
	pluginName string
	platform   index.Platform

	installDir string
	binDir     string
}

// Plugin lifecycle errors
var (
	ErrIsAlreadyInstalled = errors.New("can't install, the newest version is already installed")
	ErrIsNotInstalled     = errors.New("plugin is not installed")
	ErrIsAlreadyUpgraded  = errors.New("can't upgrade, the newest version is already installed")
)

// Install will download and install a plugin. The operation tries
// to not get the plugin dir in a bad state if it fails during the process.
func Install(p environment.Paths, plugin index.Plugin, indexName string, opts InstallOpts) error {
	klog.V(2).Infof("Looking for installed versions")
	_, err := receipt.Load(p.PluginInstallReceiptPath(plugin.Name))
	if err == nil {
		return ErrIsAlreadyInstalled
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to look up plugin receipt: %w", err)
	}

	// Find available installation candidate
	candidate, ok, err := GetMatchingPlatform(plugin.Spec.Platforms)
	if err != nil {
		return fmt.Errorf("failed trying to find a matching platform in plugin spec: %w", err)
	}
	if !ok {
		return fmt.Errorf("plugin %q does not offer installation for this platform", plugin.Name)
	}

	// The actual install should be the last action so that a failure during receipt
	// saving does not result in an installed plugin without receipt.
	klog.V(3).Infof("Install plugin %s at version=%s", plugin.Name, plugin.Spec.Version)
	if err := install(installOperation{
		pluginName: plugin.Name,
		platform:   candidate,

		binDir:     p.BinPath(),
		installDir: p.PluginVersionInstallPath(plugin.Name, plugin.Spec.Version),
	}, opts); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	klog.V(3).Infof("Storing install receipt for plugin %s", plugin.Name)
	if err := receipt.Store(receipt.New(plugin, indexName, metav1.Now()), p.PluginInstallReceiptPath(plugin.Name)); err != nil {
		return fmt.Errorf("installation receipt could not be stored, uninstall may fail: %w", err)
	}
	return nil
}

func install(op installOperation, opts InstallOpts) error {
	// Download and extract
	klog.V(3).Infof("Creating download staging directory")
	downloadStagingDir, err := os.MkdirTemp("", "krew-downloads")
	if err != nil {
		return fmt.Errorf("could not create staging dir %q: %w", downloadStagingDir, err)
	}
	klog.V(3).Infof("Successfully created download staging directory %q", downloadStagingDir)
	defer func() {
		klog.V(3).Infof("Deleting the download staging directory %s", downloadStagingDir)
		if err := os.RemoveAll(downloadStagingDir); err != nil {
			klog.Warningf("failed to clean up download staging directory: %s", err)
		}
	}()
	if err := downloadAndExtract(downloadStagingDir, op.platform.URI, op.platform.Sha256, opts.ArchiveFileOverride, opts.EnableNetrc, opts.NetrcFile); err != nil {
		return fmt.Errorf("failed to unpack into staging dir: %w", err)
	}

	applyDefaults(&op.platform)
	if err := moveToInstallDir(downloadStagingDir, op.installDir, op.platform.Files); err != nil {
		return fmt.Errorf("failed while moving files to the installation directory: %w", err)
	}

	subPathAbs, err := filepath.Abs(op.installDir)
	if err != nil {
		return fmt.Errorf("failed to get the absolute fullPath of %q: %w", op.installDir, err)
	}
	fullPath := filepath.Join(op.installDir, filepath.FromSlash(op.platform.Bin))
	pathAbs, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("failed to get the absolute fullPath of %q: %w", fullPath, err)
	}
	if _, ok := pathutil.IsSubPath(subPathAbs, pathAbs); !ok {
		return fmt.Errorf("the fullPath %q does not extend the sub-fullPath %q", fullPath, op.installDir)
	}
	if err := createOrUpdateLink(op.binDir, fullPath, op.pluginName); err != nil {
		return fmt.Errorf("failed to link installed plugin: %w", err)
	}
	return nil
}

func applyDefaults(platform *index.Platform) {
	if platform.Files == nil {
		platform.Files = []index.FileOperation{{From: "*", To: "."}}
		klog.V(4).Infof("file operation not specified, assuming %v", platform.Files)
	}
}

// downloadAndExtract downloads the specified archive uri (or uses the provided overrideFile, if a non-empty value)
// while validating its checksum with the provided sha256sum, and extracts its contents to extractDir that must be.
// created.
func downloadAndExtract(extractDir, uri, sha256sum, overrideFile string, enableNetrc bool, netrcFile string) error {
	var fetcher download.Fetcher = download.HTTPFetcher{
		EnableNetrc: enableNetrc,
		NetrcFile:   netrcFile,
	}
	if overrideFile != "" {
		fetcher = download.NewFileFetcher(overrideFile)
	}

	verifier, err := download.NewSha256Verifier(sha256sum)
	if err != nil {
		return fmt.Errorf("failed to create checksum verifier: %w", err)
	}
	if err := download.NewDownloader(verifier, fetcher).Get(uri, extractDir); err != nil {
		return fmt.Errorf("failed to unpack the plugin archive: %w", err)
	}
	return nil
}

// Uninstall will uninstall a plugin.
func Uninstall(p environment.Paths, name string) error {
	if name == constants.KrewPluginName {
		klog.Errorf("Removing krew through krew is not supported.")
		if !IsWindows() { // assume POSIX-like
			klog.Errorf("If you’d like to uninstall krew altogether, run:\n\trm -rf -- %q", p.BasePath())
		}
		return errors.New("self-uninstall not allowed")
	}
	klog.V(3).Infof("Finding installed version to delete")

	if _, err := receipt.Load(p.PluginInstallReceiptPath(name)); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrIsNotInstalled
		}
		return fmt.Errorf("failed to look up install receipt for plugin %q: %w", name, err)
	}

	klog.V(1).Infof("Deleting plugin %s", name)

	symlinkPath := filepath.Join(p.BinPath(), pluginNameToBin(name, IsWindows()))
	klog.V(3).Infof("Unlink %q", symlinkPath)
	if err := removeLink(symlinkPath); err != nil {
		return fmt.Errorf("could not uninstall symlink of plugin: %w", err)
	}

	pluginInstallPath := p.PluginInstallPath(name)
	klog.V(3).Infof("Deleting path %q", pluginInstallPath)
	if err := os.RemoveAll(pluginInstallPath); err != nil {
		return fmt.Errorf("could not remove plugin directory %q: %w", pluginInstallPath, err)
	}
	pluginReceiptPath := p.PluginInstallReceiptPath(name)
	klog.V(3).Infof("Deleting plugin receipt %q", pluginReceiptPath)
	if err := os.Remove(pluginReceiptPath); err != nil {
		return fmt.Errorf("could not remove plugin receipt %q: %w", pluginReceiptPath, err)
	}
	return nil
}

func createOrUpdateLink(binDir, binary, plugin string) error {
	dst := filepath.Join(binDir, pluginNameToBin(plugin, IsWindows()))

	if err := removeLink(dst); err != nil {
		return fmt.Errorf("failed to remove old symlink: %w", err)
	}
	if _, err := os.Stat(binary); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("can't create symbolic link, source binary (%q) cannot be found in extracted archive: %w", binary, err)
	}

	// Compute relative path from binDir to binary for portable symlinks
	relPath, err := filepath.Rel(binDir, binary)
	if err != nil {
		return fmt.Errorf("failed to compute relative path for symlink: %w", err)
	}

	// Create new symlink with relative path
	klog.V(2).Infof("Creating relative symlink to %q at %q", relPath, dst)
	if err := os.Symlink(relPath, dst); err != nil {
		return fmt.Errorf("failed to create a symlink from %q to %q: %w", relPath, dst, err)
	}
	klog.V(2).Infof("Created symlink at %q", dst)

	return nil
}

// removeLink removes a symlink reference if exists.
func removeLink(path string) error {
	fi, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		klog.V(3).Infof("No file found at %q", path)
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to read the symlink in %q: %w", path, err)
	}

	if fi.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("file %q is not a symlink (mode=%s)", path, fi.Mode())
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove the symlink in %q: %w", path, err)
	}
	klog.V(3).Infof("Removed symlink from %q", path)
	return nil
}

// IsWindows sees if KREW_OS or runtime.GOOS to find out if current execution mode is win32.
func IsWindows() bool {
	goos := runtime.GOOS
	if env := os.Getenv("KREW_OS"); env != "" {
		goos = env
	}
	return goos == "windows"
}

// pluginNameToBin creates the name of the symlink file for the plugin name.
// It converts dashes to underscores.
func pluginNameToBin(name string, isWindows bool) string {
	name = strings.ReplaceAll(name, "-", "_")
	name = "kubectl-" + name
	if isWindows {
		name += ".exe"
	}
	return name
}

// CleanupStaleKrewInstallations removes the versions that aren't the current version.
func CleanupStaleKrewInstallations(dir, currentVersion string) error {
	ls, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read krew store directory: %w", err)
	}
	klog.V(2).Infof("Found %d entries in krew store directory", len(ls))
	for _, d := range ls {
		klog.V(2).Infof("Found a krew installation: %s (%s)", d.Name(), d.Type())
		if d.IsDir() && d.Name() != currentVersion {
			klog.V(1).Infof("Deleting stale krew install directory: %s", d.Name())
			p := filepath.Join(dir, d.Name())
			if err := os.RemoveAll(p); err != nil {
				return fmt.Errorf("failed to remove stale krew version at path '%s': %w", p, err)
			}
			klog.V(1).Infof("Stale installation directory removed")
		}
	}
	return nil
}
