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
	"k8s.io/klog"

	"sigs.k8s.io/krew/internal/index/indexoperations"
	"sigs.k8s.io/krew/internal/index/indexscanner"
	"sigs.k8s.io/krew/internal/installation"
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
		indexes, err := indexoperations.ListIndexes(paths)
		if err != nil {
			return errors.Wrap(err, "failed to list indexes")
		}

		klog.V(3).Infof("found %d indexes", len(indexes))

		var plugins []pluginEntry
		for _, idx := range indexes {
			ps, err := indexscanner.LoadPluginListFromFS(paths.IndexPluginsPath(idx.Name))
			if err != nil {
				return errors.Wrapf(err, "failed to load the list of plugins from the index %q", idx.Name)
			}
			for _, p := range ps {
				plugins = append(plugins, pluginEntry{p, idx.Name})
			}
		}

		pluginCanonicalNames := make([]string, len(plugins))
		pluginCanonicalNameMap := make(map[string]pluginEntry, len(plugins))
		for i, p := range plugins {
			cn := canonicalName(p.p, p.indexName)
			pluginCanonicalNames[i] = cn
			pluginCanonicalNameMap[cn] = p
		}

		installed := make(map[string]bool)
		receipts, err := installation.GetInstalledPluginReceipts(paths.InstallReceiptsPath())
		if err != nil {
			return errors.Wrap(err, "failed to load installed plugins")
		}
		for _, receipt := range receipts {
			cn := canonicalName(receipt.Plugin, indexOf(receipt))
			installed[cn] = true
		}

		var searchResults []string
		if len(args) > 0 {
			matches := fuzzy.Find(strings.Join(args, ""), pluginCanonicalNames)
			for _, m := range matches {
				searchResults = append(searchResults, m.Str)
			}
		} else {
			searchResults = pluginCanonicalNames
		}

		// No plugins found
		if len(searchResults) == 0 {
			return nil
		}

		var rows [][]string
		cols := []string{"NAME", "DESCRIPTION", "INSTALLED"}
		for _, canonicalName := range searchResults {
			v := pluginCanonicalNameMap[canonicalName]
			var status string
			if installed[canonicalName] {
				status = "yes"
			} else if _, ok, err := installation.GetMatchingPlatform(v.p.Spec.Platforms); err != nil {
				return errors.Wrapf(err, "failed to get the matching platform for plugin %s", canonicalName)
			} else if ok {
				status = "no"
			} else {
				status = "unavailable on " + runtime.GOOS
			}

			rows = append(rows, []string{displayName(v.p, v.indexName), limitString(v.p.Spec.ShortDescription, 50), status})
		}
		rows = sortByFirstColumn(rows)
		return printTable(os.Stdout, cols, rows)
	},
	PreRunE: checkIndex,
}

func limitString(s string, length int) string {
	if len(s) > length && length > 3 {
		s = s[:length-3] + "..."
	}
	return s
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
