// Copyright © 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/GoogleContainerTools/krew/pkg/index/indexscanner"
	"github.com/GoogleContainerTools/krew/pkg/installation"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade installed plugins to a newer version",
	Long: `Upgrade installed plugins to a newer version.
This will reinstall all plugins that have a newer version in the local index.
Use "kubectl plugin update" to renew the index. All plugins that rely on HEAD
will always be installed.
To only upgrade single plugins provide them as arguments:
kubectl plugin upgrade foo bar"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var ignoreUpgraded bool
		var pluginNames []string
		// Upgrade all plugins.
		if len(args) == 0 {
			installed, err := installation.ListInstalledPlugins(paths.Install)
			if err != nil {
				return fmt.Errorf("failed to find all installed versions, err: %v", err)
			}
			for name := range installed {
				pluginNames = append(pluginNames, name)
			}
			ignoreUpgraded = true
		} else {
			pluginNames = args
		}

		for _, name := range pluginNames {
			plugin, err := indexscanner.LoadPluginFileFromFS(paths.Index, name)
			if err != nil {
				return fmt.Errorf("failed to load the index file for plugin %s, err: %v", plugin.Name, err)
			}

			glog.V(2).Infof("Upgrading plugin: %s\n", plugin.Name)
			err = installation.Upgrade(paths, plugin, krewExecutedVersion)
			if ignoreUpgraded && err == installation.IsAlreadyUpgradedErr {
				fmt.Fprintf(os.Stderr, "Skipping plugin %s, it is already on the newest version\n", plugin.Name)
				continue
			}
			if err != nil {
				return fmt.Errorf("failed to upgrade plugin %q, err: %v", plugin.Name, err)
			}
			fmt.Fprintf(os.Stderr, "Upgraded plugin: %s\n", plugin.Name)
		}
		return nil
	},
	PreRunE: ensureUpdated,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
