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
// oldenvironment is a copy of the relevant function in environment before the index migration.
package oldenvironment

import (
	"os"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/homedir"
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
		glog.V(4).Infof("using environment override KREW_ROOT=%s", fromEnv)
	}
	base, err := filepath.Abs(base)
	if err != nil {
		panic(errors.Wrap(err, "cannot get absolute path"))
	}
	return newPaths(base)
}

func newPaths(base string) Paths {
	return Paths{base: base, tmp: os.TempDir()}
}

// BasePath returns krew base directory.
func (p Paths) BasePath() string { return p.base }

// IndexPath returns the base directory where plugin index repository is cloned.
//
// e.g. {BasePath}/index/
func (p Paths) IndexPath() string { return filepath.Join(p.base, "index") }

// IndexPluginsPath returns the plugins directory of the index repository.
//
// e.g. {BasePath}/index/plugins/
func (p Paths) IndexPluginsPath() string { return filepath.Join(p.base, "index", "plugins") }

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

// PluginVersionInstallPath returns the path to the specified version of specified
// plugin.
//
// e.g. {PluginInstallPath}/{plugin}/{version}
func (p Paths) PluginVersionInstallPath(plugin, version string) string {
	return filepath.Join(p.InstallPath(), plugin, version)
}
