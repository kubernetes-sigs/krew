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
	"sigs.k8s.io/krew/pkg/constants"
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
		indexes := []indexoperations.Index{
			{
				Name: constants.DefaultIndexName,
				URL:  constants.IndexURI, // unused here but providing for completeness
			},
		}
		if os.Getenv(constants.EnableMultiIndexSwitch) != "" {
			out, err := indexoperations.ListIndexes(paths)
			if err != nil {
				return errors.Wrapf(err, "failed to list plugin indexes available")
			}
			indexes = out
		}

		klog.V(3).Infof("found %d indexes", len(indexes))

		var plugins []pluginEntry
		for _, idx := range indexes {
			ps, err := indexscanner.LoadPluginListFromFS(paths.IndexPluginsPath(idx.Name))
			if err != nil {
				return errors.Wrap(err, "failed to load the list of plugins from the index")
			}
			for _, p := range ps {
				plugins = append(plugins, pluginEntry{p, idx.Name})
			}
		}

		keys := func(v map[string]pluginEntry) []string {
			out := make([]string, 0, len(v))
			for k := range v {
				out = append(out, k)
			}
			return out
		}

		pluginMap := make(map[string]pluginEntry, len(plugins))
		for _, p := range plugins {
			pluginMap[canonicalName(p.p, p.indexName)] = p
		}

		installed := make(map[string]string)
		receipts, err := installation.GetInstalledPluginReceipts(paths.InstallReceiptsPath())
		if err != nil {
			return errors.Wrap(err, "failed to load installed plugins")
		}
		for _, receipt := range receipts {
			index := receipt.Status.Source.Name
			if index == "" {
				index = constants.DefaultIndexName
			}
			installed[receipt.Name] = index
		}

		corpus := keys(pluginMap)
		var searchResults []string
		if len(args) > 0 {
			matches := fuzzy.Find(strings.Join(args, ""), corpus)
			for _, m := range matches {
				searchResults = append(searchResults, m.Str)
			}
		} else {
			searchResults = corpus
		}

		// No plugins found
		if len(searchResults) == 0 {
			return nil
		}

		var rows [][]string
		cols := []string{"NAME", "DESCRIPTION", "INSTALLED"}
		for _, name := range searchResults {
			v := pluginMap[name]
			var status string
			if index := installed[v.p.Name]; index == v.indexName {
				status = "yes"
			} else if _, ok, err := installation.GetMatchingPlatform(v.p.Spec.Platforms); err != nil {
				return errors.Wrapf(err, "failed to get the matching platform for plugin %s", name)
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
