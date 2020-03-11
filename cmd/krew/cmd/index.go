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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog"

	"sigs.k8s.io/krew/internal/index/indexoperations"
	"sigs.k8s.io/krew/pkg/constants"
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:    "index",
	Short:  "Manage custom plugin indexes",
	Long:   "Manage which repositories are used to discover and install plugins from.",
	Args:   cobra.NoArgs,
	Hidden: true, // TODO(chriskim06) remove this once multi-index is enabled
}

var indexListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured indexes",
	Long: `Print a list of configured indexes.

This command prints a list of indexes. It shows the name and the remote URL for
each configured index in table format.`,
	Args: cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		indexes, err := indexoperations.ListIndexes(paths.IndexBase())
		if err != nil {
			return errors.Wrap(err, "failed to list indexes")
		}
		var rows [][]string
		for _, index := range indexes {
			rows = append(rows, []string{index.Name, index.URL})
		}
		return printTable(os.Stdout, []string{"INDEX", "URL"}, rows)
	},
}

var indexAddCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add a new index",
	Long:    "Configure a new index to install plugins from.",
	Example: "kubectl krew index add default " + constants.IndexURI,
	Args:    cobra.ExactArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		err := indexoperations.AddIndex(paths.IndexBase(), args[0], args[1])
		if err != nil {
			return errors.Wrap(err, "failed to add index")
		}
		klog.Infoln("Successfully added index")
		return nil
	},
}

func init() {
	if _, ok := os.LookupEnv(constants.EnableMultiIndexSwitch); ok {
		indexCmd.AddCommand(indexListCmd)
		indexCmd.AddCommand(indexAddCmd)
		rootCmd.AddCommand(indexCmd)
	}
}
