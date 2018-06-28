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
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/google/krew/pkg/installation"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func init() {
	// listCmd represents the list command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all installed plugin names",
		Long: `List all installed plugin names.
Plugins will be shown as "PLUGIN,VERSION"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			plugins, err := installation.ListInstalledPlugins(paths.Install)
			if err != nil {
				return fmt.Errorf("failed to find all installed versions, err %v", err)
			}
			if !(isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())) {
				fmt.Fprintf(os.Stdout, "%s\n", strings.Join(sortedKeys(plugins), "\n"))
				return nil
			}
			printAlignedColumns(os.Stdout, "PLUGIN", "VERSION", plugins)
			return nil
		},
		PreRunE: checkIndex,
	}

	rootCmd.AddCommand(listCmd)
}

func printAlignedColumns(out io.Writer, keyHeader, valueHeader string, columns map[string]string) error {
	w := tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\n", keyHeader, valueHeader)
	// TODO(lbb): print sorted map, to allow unix parsing or allow -o json flag
	keys := sortedKeys(columns)
	for _, name := range keys {
		fmt.Fprintf(w, "%s\t%s\n", name, columns[name])
	}
	return w.Flush()
}

func sortedKeys(m map[string]string) []string {
	keys := stringKeys(m)
	sort.Strings(keys)
	return keys
}

func stringKeys(m map[string]string) []string {
	var keys = make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}
