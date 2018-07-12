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
	"strings"
	"text/tabwriter"

	"github.com/GoogleContainerTools/krew/pkg/index/indexscanner"

	"github.com/GoogleContainerTools/krew/pkg/index"
	"github.com/GoogleContainerTools/krew/pkg/installation"
	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Discover plugins in your local index using fuzzy search",
	Long: `Discover plugins in your local index using fuzzy search.
Search accepts a list of words as options.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		plugins, err := indexscanner.LoadPluginListFromFS(paths.Index)
		if err != nil {
			return fmt.Errorf("failed to load the index, err %v", err)
		}
		names := make([]string, len(plugins.Items))
		pluginMap := make(map[string]index.Plugin, len(plugins.Items))
		for i, p := range plugins.Items {
			names[i] = p.Name
			pluginMap[p.Name] = p
		}

		installed, err := installation.ListInstalledPlugins(paths.Install)
		if err != nil {
			return fmt.Errorf("failed to load installed plugins, err: %v", err)
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

		rowPattern := "%s\t%s\t%s\n"
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintf(w, rowPattern, "NAME", "DESCRIPTION", "STATUS")
		for _, name := range matchNames {
			plugin := pluginMap[name]
			var status string
			if _, ok := installed[name]; ok {
				status = "installed"
			} else if _, ok, err := installation.GetMatchingPlatform(plugin); err != nil {
				return fmt.Errorf("failed to get the matching platform for plugin %s, err: %v", name, err)
			} else if ok {
				status = "available"
			} else {
				status = "unavailable"
			}
			fmt.Fprintf(w, rowPattern, name, limitString(plugin.Spec.ShortDescription, 50), status)
		}
		w.Flush()
		return nil
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
