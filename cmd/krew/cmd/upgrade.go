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

package cmd

import (
	"fmt"
	"os"

	"sigs.k8s.io/krew/pkg/index/indexscanner"
	"sigs.k8s.io/krew/pkg/installation"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade installed plugins to newer versions",
	Long: `Upgrade installed plugins to a newer version.
This will reinstall all plugins that have a newer version in the local index.
Use "kubectl krew update" to renew the index.
To only upgrade single plugins provide them as arguments:
kubectl krew upgrade foo bar"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var ignoreUpgraded bool
		var pluginNames []string
		// Upgrade all plugins.
		if len(args) == 0 {
			installed, err := installation.ListInstalledPlugins(paths.InstallPath(), paths.BinPath())
			if err != nil {
				return errors.Wrap(err, "failed to find all installed versions")
			}
			for name := range installed {
				pluginNames = append(pluginNames, name)
			}
			ignoreUpgraded = true
		} else {
			pluginNames = args
		}

		for _, name := range pluginNames {
			plugin, err := indexscanner.LoadPluginFileFromFS(paths.IndexPath(), name)
			if err != nil {
				return errors.Wrapf(err, "failed to load the index file for plugin %s", plugin.Name)
			}

			glog.V(2).Infof("Upgrading plugin: %s\n", plugin.Name)
			err = installation.Upgrade(paths, plugin)
			if ignoreUpgraded && err == installation.ErrIsAlreadyUpgraded {
				fmt.Fprintf(os.Stderr, "Skipping plugin %s, it is already on the newest version\n", plugin.Name)
				continue
			}
			if err != nil {
				return errors.Wrapf(err, "failed to upgrade plugin %q", plugin.Name)
			}
			fmt.Fprintf(os.Stderr, "Upgraded plugin: %s\n", plugin.Name)
		}
		return nil
	},
	PreRunE: ensureIndexUpdated,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
