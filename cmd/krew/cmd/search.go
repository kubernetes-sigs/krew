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
	"os"
	"strings"

	"github.com/GoogleContainerTools/krew/pkg/index/indexscanner"
	"github.com/pkg/errors"

	"github.com/GoogleContainerTools/krew/pkg/index"
	"github.com/GoogleContainerTools/krew/pkg/installation"
	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Discover kubectl plugins",
	Long: `List kubectl plugins available on krew and search among them.
If no arguments are provided, all plugins will be listed.

Examples:
  To list all plugins:
    kubectl krew search

  To fuzzy search plugins with a keyword:
    kubectl krew search KEYWORD`,
	RunE: func(cmd *cobra.Command, args []string) error {
		plugins, err := indexscanner.LoadPluginListFromFS(paths.IndexPath())
		if err != nil {
			return errors.Wrap(err, "failed to load the index")
		}
		names := make([]string, len(plugins.Items))
		pluginDescriptions := make([]string, len(plugins.Items))
		pluginDescriptionsNameMap := make(map[string]string, len(plugins.Items))
		pluginMap := make(map[string]index.Plugin, len(plugins.Items))
		for i, p := range plugins.Items {
			names[i] = p.Name
			pluginDescriptions[i] = p.Spec.Description
			pluginMap[p.Name] = p
			pluginDescriptionsNameMap[p.Spec.Description] = p.Name
		}

		installed, err := installation.ListInstalledPlugins(paths.InstallPath(), paths.BinPath())
		if err != nil {
			return errors.Wrap(err, "failed to load installed plugins")
		}

		var matchNames []string
		if len(args) > 0 {
			matches := fuzzy.Find(strings.Join(args, ""), names)
			pluginMatches := fuzzy.Find(strings.Join(args, ""), pluginDescriptions)
			for _, m := range matches {
				matchNames = append(matchNames, m.Str)
			}
			for _, pm := range pluginMatches {
				matchName := pluginDescriptionsNameMap[pm.Str]
				matchNames = append(matchNames, matchName)
			}
		} else {
			matchNames = names
		}

		// No plugins found
		if len(matchNames) == 0 {
			return nil
		}

		var rows [][]string
		cols := []string{"NAME", "DESCRIPTION", "STATUS"}
		for _, name := range matchNames {
			plugin := pluginMap[name]
			var status string
			if _, ok := installed[name]; ok {
				status = "installed"
			} else if _, ok, err := installation.GetMatchingPlatform(plugin); err != nil {
				return errors.Wrapf(err, "failed to get the matching platform for plugin %s", name)
			} else if ok {
				status = "available"
			} else {
				status = "unavailable"
			}
			rows = append(rows, []string{name, limitString(plugin.Spec.ShortDescription, 50), status})
		}
		rows = sortByFirstColumn(rows)
		return printTable(os.Stdout, cols, rows)
	},
	PreRunE: checkIndex,
}

func limitString(s string, length int) string {
	if len(s) > length && length > 3 {
		s = string(s[:length-3]) + "..."
	}
	return s
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
