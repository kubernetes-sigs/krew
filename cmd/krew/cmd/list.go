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
	"io/ioutil"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all installed plugin names",
	Long: `List all installed plugin names.
Plugins will be shown as "PLUGIN,VERSION"`,
	Run: func(cmd *cobra.Command, args []string) {
		glog.V(4).Infof("Reading installation path %q", paths.Install)
		plugins, err := ioutil.ReadDir(paths.Install)
		if err != nil {
			glog.Fatal(err)
		}
		columns := make(map[string]string)
		for _, p := range plugins {
			if !p.IsDir() {
				continue
			}
			versions, err := ioutil.ReadDir(filepath.Join(paths.Install, p.Name()))
			if err != nil {
				glog.Fatal(err)
			}
			for _, v := range versions {
				if !v.IsDir() {
					continue
				}
				columns[p.Name()] = v.Name()
			}
		}
		printAlignedColums(os.Stdout, columns)
	},
	PreRunE: checkIndex,
}

func printAlignedColums(out io.Writer, columns map[string]string) error {
	w := tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w, "PLUGIN\tVERSION")
	for name, version := range columns {
		fmt.Fprintf(w, "%s\t%s\n", name, version)
	}
	return w.Flush()
}

func init() {
	rootCmd.AddCommand(listCmd)
}
