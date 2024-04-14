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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"sigs.k8s.io/krew/cmd/krew/cmd/internal"
	"sigs.k8s.io/krew/internal/index/indexscanner"
	"sigs.k8s.io/krew/internal/index/validation"
	"sigs.k8s.io/krew/internal/installation"
	"sigs.k8s.io/krew/internal/installation/receipt"
	"sigs.k8s.io/krew/internal/pathutil"
	"sigs.k8s.io/krew/pkg/constants"
)

func init() {
	var noUpdateIndex *bool

	// upgradeCmd represents the upgrade command
	var upgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade installed plugins to newer versions",
		Long: `Upgrade installed plugins to a newer version.
This will reinstall all plugins that have a newer version in the local index.
Use "kubectl krew update" to renew the index.
To only upgrade single plugins provide them as arguments:
kubectl krew upgrade foo bar"`,
		RunE: func(_ *cobra.Command, args []string) error {
			var ignoreUpgraded bool
			var skipErrors bool

			var pluginNames []string
			if len(args) == 0 {
				// Upgrade all plugins.
				installed, err := installation.GetInstalledPluginReceipts(paths.InstallReceiptsPath())
				if err != nil {
					return errors.Wrap(err, "failed to find all installed versions")
				}
				for _, receipt := range installed {
					pluginNames = append(pluginNames, receipt.Status.Source.Name+"/"+receipt.Name)
				}
				ignoreUpgraded = true
				skipErrors = true
			} else {
				// Upgrade certain plugins
				for _, arg := range args {
					if isCanonicalName(arg) {
						return errors.New("upgrade command does not support INDEX/PLUGIN syntax; just specify PLUGIN")
					} else if !validation.IsSafePluginName(arg) {
						return unsafePluginNameErr(arg)
					}
					r, err := receipt.Load(paths.PluginInstallReceiptPath(arg))
					if err != nil {
						return errors.Wrapf(err, "read receipt %q", arg)
					}
					pluginNames = append(pluginNames, r.Status.Source.Name+"/"+r.Name)
				}
			}

			var nErrors int
			for _, name := range pluginNames {
				indexName, pluginName := pathutil.CanonicalPluginName(name)
				if indexName == "detached" {
					klog.Warningf("Skipping upgrade for %q because it was installed via manifest\n", pluginName)
					continue
				}

				plugin, err := indexscanner.LoadPluginByName(paths.IndexPluginsPath(indexName), pluginName)
				if err != nil {
					if !os.IsNotExist(err) {
						return errors.Wrapf(err, "failed to load the plugin manifest for plugin %s", name)
					} else if !skipErrors {
						return errors.Errorf("plugin %q does not exist in the plugin index", name)
					}
				}

				pluginDisplayName := displayName(plugin, indexName)
				if err == nil {
					fmt.Fprintf(os.Stderr, "Upgrading plugin: %s\n", pluginDisplayName)
					err = installation.Upgrade(paths, plugin, indexName)
					if ignoreUpgraded && err == installation.ErrIsAlreadyUpgraded {
						fmt.Fprintf(os.Stderr, "Skipping plugin %s, it is already on the newest version\n", pluginDisplayName)
						continue
					}
				}
				if err != nil {
					nErrors++
					if skipErrors {
						fmt.Fprintf(os.Stderr, "WARNING: failed to upgrade plugin %q, skipping (error: %v)\n", pluginDisplayName, err)
						continue
					}
					return errors.Wrapf(err, "failed to upgrade plugin %q", pluginDisplayName)
				}
				fmt.Fprintf(os.Stderr, "Upgraded plugin: %s\n", pluginDisplayName)
				if indexName == constants.DefaultIndexName {
					internal.PrintSecurityNotice(plugin.Name)
				}
			}
			if nErrors > 0 {
				fmt.Fprintf(os.Stderr, "WARNING: Some plugins failed to upgrade, check logs above.\n")
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if *noUpdateIndex {
				klog.V(4).Infof("--no-update-index specified, skipping updating local copy of plugin index")
				return nil
			}
			return ensureIndexes(cmd, args)
		},
	}

	noUpdateIndex = upgradeCmd.Flags().Bool("no-update-index", false, "(Experimental) do not update local copy of plugin index before upgrading")
	rootCmd.AddCommand(upgradeCmd)
}
