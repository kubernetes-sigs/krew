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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"sigs.k8s.io/krew/internal/index/indexscanner"
	"sigs.k8s.io/krew/internal/installation"
	"sigs.k8s.io/krew/internal/installation/semver"
)

func init() {
	outdatedCmd := &cobra.Command{
		Use:   "outdated",
		Short: "List installed plugins with newer versions available",
		Long: `List all installed kubectl plugins that have newer versions available
in the local index. This command does not perform any upgrades.

Use "kubectl krew update" to refresh the index before checking for
outdated plugins.

To upgrade all outdated plugins, use:
  kubectl krew upgrade`,
		RunE: func(_ *cobra.Command, _ []string) error {
			receipts, err := installation.GetInstalledPluginReceipts(paths.InstallReceiptsPath())
			if err != nil {
				return fmt.Errorf("failed to find installed plugins: %w", err)
			}

			var rows [][]string
			for _, r := range receipts {
				indexName := indexOf(r)
				pluginName := r.Name

				// Skip plugins installed from a manifest (detached)
				if indexName == "detached" {
					klog.V(2).Infof("Skipping %q: installed via manifest", pluginName)
					continue
				}

				// Load latest version from the index
				indexPlugin, err := indexscanner.LoadPluginByName(paths.IndexPluginsPath(indexName), pluginName)
				if err != nil {
					if errors.Is(err, fs.ErrNotExist) {
						klog.V(1).Infof("Skipping %q: plugin no longer exists in index %q", pluginName, indexName)
						continue
					}
					return fmt.Errorf("failed to load index entry for plugin %q: %w", pluginName, err)
				}

				curVersion := r.Spec.Version
				newVersion := indexPlugin.Spec.Version

				curv, err := semver.Parse(curVersion)
				if err != nil {
					klog.V(1).Infof("Skipping %q: cannot parse installed version %q", pluginName, curVersion)
					continue
				}
				newv, err := semver.Parse(newVersion)
				if err != nil {
					klog.V(1).Infof("Skipping %q: cannot parse index version %q", pluginName, newVersion)
					continue
				}

				if semver.Less(curv, newv) {
					rows = append(rows, []string{displayName(indexPlugin, indexName), curVersion, newVersion})
				}
			}

			if len(rows) == 0 {
				fmt.Fprintln(os.Stderr, "All plugins are up to date.")
				return nil
			}

			rows = sortByFirstColumn(rows)

			// Return only names when piped
			if !isTerminal(os.Stdout) {
				var names []string
				for _, row := range rows {
					names = append(names, row[0])
				}
				fmt.Fprintln(os.Stdout, strings.Join(names, "\n"))
				return nil
			}

			return printTable(os.Stdout, []string{"PLUGIN", "INSTALLED", "AVAILABLE"}, rows)
		},
		PreRunE: checkIndex,
	}

	rootCmd.AddCommand(outdatedCmd)
}
