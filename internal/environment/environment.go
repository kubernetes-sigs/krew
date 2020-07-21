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

package environment

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog"

	"sigs.k8s.io/krew/pkg/constants"
)

// Paths contains all important environment paths
type Paths struct {
	base string
	tmp  string
}

// MustGetKrewPaths returns the inferred paths for krew. By default, it assumes
// $HOME/.krew as the base path, but can be overridden via KREW_ROOT environment
// variable.
func MustGetKrewPaths() Paths {
	base := filepath.Join(homedir.HomeDir(), ".krew")
	if fromEnv := os.Getenv("KREW_ROOT"); fromEnv != "" {
		base = fromEnv
		klog.V(4).Infof("using environment override KREW_ROOT=%s", fromEnv)
	}
	base, err := filepath.Abs(base)
	if err != nil {
		panic(errors.Wrap(err, "cannot get absolute path"))
	}
	return NewPaths(base)
}

func NewPaths(base string) Paths {
	return Paths{base: base, tmp: os.TempDir()}
}

// BasePath returns krew base directory.
func (p Paths) BasePath() string { return p.base }

// IndexBase returns the krew index directory. This directory contains the default
// index and custom ones.
func (p Paths) IndexBase() string {
	return filepath.Join(p.base, "index")
}

// IndexPath returns the directory where a plugin index repository is cloned.
// e.g. {BasePath}/index/default or {BasePath}/index
func (p Paths) IndexPath(name string) string {
	return filepath.Join(p.base, "index", name)
}

// IndexPluginsPath returns the plugins directory of an index repository.
// e.g. {BasePath}/index/default/plugins/ or {BasePath}/index/plugins/
func (p Paths) IndexPluginsPath(name string) string {
	return filepath.Join(p.IndexPath(name), "plugins")
}

// InstallReceiptsPath returns the base directory where plugin receipts are stored.
//
// e.g. {BasePath}/receipts
func (p Paths) InstallReceiptsPath() string { return filepath.Join(p.base, "receipts") }

// BinPath returns the path where plugin executable symbolic links are found.
// This path should be added to $PATH in client machine.
//
// e.g. {BasePath}/bin
func (p Paths) BinPath() string { return filepath.Join(p.base, "bin") }

// InstallPath returns the base directory for plugin installations.
//
// e.g. {BasePath}/store
func (p Paths) InstallPath() string { return filepath.Join(p.base, "store") }

// PluginInstallPath returns the path to install the plugin.
//
// e.g. {InstallPath}/{version}/{..files..}
func (p Paths) PluginInstallPath(plugin string) string {
	return filepath.Join(p.InstallPath(), plugin)
}

// PluginInstallReceiptPath returns the path to the install receipt for plugin.
//
// e.g. {InstallReceiptsPath}/{plugin}.yaml
func (p Paths) PluginInstallReceiptPath(plugin string) string {
	return filepath.Join(p.InstallReceiptsPath(), plugin+constants.ManifestExtension)
}

// PluginVersionInstallPath returns the path to the specified version of specified
// plugin.
//
// e.g. {PluginInstallPath}/{plugin}/{version}
func (p Paths) PluginVersionInstallPath(plugin, version string) string {
	return filepath.Join(p.InstallPath(), plugin, version)
}

// Realpath evaluates symbolic links. If the path is not a symbolic link, it
// returns the cleaned path. Symbolic links with relative paths return error.
func Realpath(path string) (string, error) {
	s, err := os.Lstat(path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to stat the currently executed path (%q)", path)
	}

	if s.Mode()&os.ModeSymlink != 0 {
		if path, err = os.Readlink(path); err != nil {
			return "", errors.Wrap(err, "failed to resolve the symlink of the currently executed version")
		}
		if !filepath.IsAbs(path) {
			return "", errors.Errorf("symbolic link is relative (%s)", path)
		}
	}
	return filepath.Clean(path), nil
}
