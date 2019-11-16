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
	"os"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"

	"sigs.k8s.io/krew/pkg/index"
	"sigs.k8s.io/krew/pkg/index/indexscanner"
	"sigs.k8s.io/krew/pkg/installation"
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
		plugins, err := indexscanner.LoadPluginListFromFS(paths.IndexPluginsPath())
		if err != nil {
			return errors.Wrap(err, "failed to load the list of plugins from the index")
		}
		names := make([]string, len(plugins))
		pluginMap := make(map[string]index.Plugin, len(plugins))
		for i, p := range plugins {
			names[i] = p.Name
			pluginMap[p.Name] = p
		}

		installed, err := installation.ListInstalledPlugins(paths.InstallReceiptsPath())
		if err != nil {
			return errors.Wrap(err, "failed to load installed plugins")
		}

		var matchNames []string
		if len(args) > 0 {
			matches := fuzzy.Find(strings.Join(args, ""), names)
			for _, m := range matches {
				matchNames = append(matchNames, m.Str)
			}
		} else {
			matchNames = names
		}

		// No plugins found
		if len(matchNames) == 0 {
			return nil
		}

		var rows [][]string
		cols := []string{"NAME", "DESCRIPTION", "INSTALLED"}
		for _, name := range matchNames {
			plugin := pluginMap[name]
			var status string
			if _, ok := installed[name]; ok {
				status = "yes"
			} else if _, ok, err := installation.GetMatchingPlatform(plugin.Spec.Platforms); err != nil {
				return errors.Wrapf(err, "failed to get the matching platform for plugin %s", name)
			} else if ok {
				status = "no"
			} else {
				status = "unavailable on " + runtime.GOOS
			}
			rows = append(rows, []string{name, limitString(plugin.Spec.ShortDescription, 50), status})
		}
		rows = sortByFirstColumn(rows)
		return printTable(os.Stdout, cols, rows)
	},
	PreRunE: checkIndex,
}

func limitString(s string, length int) string {
	s = strings.TrimSpace(s)
	if len(s) > length && length > 3 {
		s = s[:length-3] + "..."
	}
	return s
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
