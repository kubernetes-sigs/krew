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
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"sigs.k8s.io/krew/cmd/krew/cmd/internal"
	"sigs.k8s.io/krew/internal/index/indexoperations"
	"sigs.k8s.io/krew/internal/installation"
	"sigs.k8s.io/krew/pkg/constants"
)

var (
	forceIndexDelete    *bool
	errInvalidIndexName = errors.New("invalid index name")
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Manage custom plugin indexes",
	Long:  "Manage which repositories are used to discover and install plugins from.",
	Args:  cobra.NoArgs,
}

var indexListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured indexes",
	Long: `Print a list of configured indexes.

This command prints a list of indexes. It shows the name and the remote URL for
each configured index in table format.`,
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		indexes, err := indexoperations.ListIndexes(paths)
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
	Example: "kubectl krew index add default " + constants.DefaultIndexURI,
	Args:    cobra.ExactArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		name := args[0]
		if !indexoperations.IsValidIndexName(name) {
			return errInvalidIndexName
		}
		err := indexoperations.AddIndex(paths, name, args[1])
		if err != nil {
			return err
		}
		internal.PrintWarning(os.Stderr, `You have added a new index from %q
The plugins in this index are not audited for security by the Krew maintainers.
Install them at your own risk.
`, args[1])
		return nil
	},
}

var indexDeleteCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a configured index",
	Long: `Remove a configured plugin index.

It is only safe to remove indexes without installed plugins. Removing an index
while there are plugins installed will result in an error, unless the --force
option is used (not recommended).`,

	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	RunE:    indexDelete,
}

func indexDelete(_ *cobra.Command, args []string) error {
	name := args[0]
	if !indexoperations.IsValidIndexName(name) {
		return errInvalidIndexName
	}

	ps, err := installation.InstalledPluginsFromIndex(paths.InstallReceiptsPath(), name)
	if err != nil {
		return errors.Wrap(err, "failed querying plugins installed from the index")
	}
	klog.V(4).Infof("Found %d plugins from index", len(ps))

	if len(ps) > 0 && !*forceIndexDelete {
		var names []string
		for _, pl := range ps {
			names = append(names, pl.Name)
		}

		internal.PrintWarning(os.Stderr, `Plugins [%s] are still installed from index %q!
Removing indexes while there are plugins installed from is not recommended
(you can use --force to ignore this check).`+"\n", strings.Join(names, ", "), name)
		return errors.Errorf("there are still plugins installed from this index")
	}

	err = indexoperations.DeleteIndex(paths, name)
	if os.IsNotExist(err) {
		if *forceIndexDelete {
			klog.V(4).Infof("Index not found, but --force is used, so not returning an error")
			return nil // success if --force specified and index does not exist.
		}
		return errors.Errorf("index %q does not exist", name)
	}
	return errors.Wrap(err, "error while removing the plugin index")
}

func init() {
	forceIndexDelete = indexDeleteCmd.Flags().Bool("force", false,
		"Remove index even if it has plugins currently installed (may result in unsupported behavior)")

	indexCmd.AddCommand(indexAddCmd)
	indexCmd.AddCommand(indexListCmd)
	indexCmd.AddCommand(indexDeleteCmd)
	rootCmd.AddCommand(indexCmd)
}
