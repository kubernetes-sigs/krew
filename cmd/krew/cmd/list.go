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
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"sigs.k8s.io/krew/internal/installation"
)

func init() {
	// listCmd represents the list command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed kubectl plugins",
		Long: `Show a list of installed kubectl plugins and their versions.

Remarks:
  Redirecting the output of this command to a program or file will only print
  the names of the plugins installed. This output can be piped back to the
  "install" command.`,
		Aliases: []string{"ls"},
		RunE: func(_ *cobra.Command, _ []string) error {
			receipts, err := installation.GetInstalledPluginReceipts(paths.InstallReceiptsPath())
			if err != nil {
				return errors.Wrap(err, "failed to find all installed versions")
			}

			// return sorted list of plugin names when piped to other commands or file
			if !isTerminal(os.Stdout) {
				var names []string
				for _, r := range receipts {
					names = append(names, displayName(r.Plugin, indexOf(r)))
				}
				sort.Strings(names)
				fmt.Fprintln(os.Stdout, strings.Join(names, "\n"))
				return nil
			}

			// print table
			var rows [][]string
			for _, r := range receipts {
				rows = append(rows, []string{displayName(r.Plugin, indexOf(r)), r.Spec.Version})
			}
			rows = sortByFirstColumn(rows)
			return printTable(os.Stdout, []string{"PLUGIN", "VERSION"}, rows)
		},
		PreRunE: checkIndex,
	}

	rootCmd.AddCommand(listCmd)
}

func printTable(out io.Writer, columns []string, rows [][]string) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprint(w, strings.Join(columns, "\t"))
	fmt.Fprintln(w)
	for _, values := range rows {
		fmt.Fprint(w, strings.Join(values, "\t"))
		fmt.Fprintln(w)
	}
	return w.Flush()
}

func sortByFirstColumn(rows [][]string) [][]string {
	sort.Slice(rows, func(a, b int) bool {
		return rows[a][0] < rows[b][0]
	})
	return rows
}
