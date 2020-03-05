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
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"sigs.k8s.io/krew/internal/gitutil"
	"sigs.k8s.io/krew/pkg/constants"
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:    "index",
	Short:  "Perform krew index commands",
	Long:   "Show a list of installed kubectl plugins and their versions.",
	Args:   cobra.NoArgs,
	Hidden: true, // remove this once multi-index is enabled
}

var indexListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured indexes",
	Long: `Print a list of configured indexes.

This command prints a list of indexes. It shows the name and the remote URL for
each configured index.`,
	Args: cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		dirs, err := ioutil.ReadDir(paths.IndexBase())
		if err != nil {
			return errors.Wrapf(err, "failed to read directory %s", paths.IndexBase())
		}
		var rows [][]string
		for _, dir := range dirs {
			indexName := dir.Name()
			remote, err := gitutil.GetRemoteURL(paths.IndexPath(indexName))
			if err != nil {
				return errors.Wrapf(err, "failed to list the remote URL for index %s", indexName)
			}
			rows = append(rows, []string{indexName, remote})
		}
		rows = sortByFirstColumn(rows)
		return printTable(os.Stdout, []string{"INDEX", "URL"}, rows)
	},
}

func init() {
	if _, ok := os.LookupEnv(constants.EnableMultiIndexSwitch); ok {
		rootCmd.AddCommand(indexCmd)
		indexCmd.AddCommand(indexListCmd)
	}
}
