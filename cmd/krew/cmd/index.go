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

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog"

	"sigs.k8s.io/krew/internal/index/indexoperations"
	"sigs.k8s.io/krew/internal/installation"
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
	Example: "kubectl krew index add default " + constants.IndexURI,
	Args:    cobra.ExactArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		return indexoperations.AddIndex(paths, args[0], args[1])
	},
}

var indexDeleteCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a configured index",
	Long: `Remove a configured index repository which is used to discover and
install plugins from.

If there are plugins installed from the specified index, the index cannot be
removed (unless --force) is used. Removing indexes while there are plugins
installed from them is not supported behavior.`,
	Args: cobra.ExactArgs(1),
	RunE: indexDelete,
}

func indexDelete(_ *cobra.Command, args []string) error {
	name := args[0]

	pl, err := installation.InstalledPluginsFromIndex(paths.InstallReceiptsPath(), name)
	if err != nil {
		return errors.Wrap(err, "failed querying plugins installed from the index")
	}
	klog.V(4).Infof("Found %d plugins from index", len(pl))

	if len(pl) > 0 && !*forceIndexDelete {
		var names []string
		for _, v := range pl {
			names = append(names, v.Name)
		}

		warning := color.New(color.FgRed, color.Bold).Sprint("WARNING:")
		klog.Warningf("%s Plugins [%s] are still installed from index %q.", warning, strings.Join(names, ", "), name)
		klog.Warning("Uninstall them first before removing this index (or, use --force, which may result in unsupported behavior).")
		return errors.Errorf("refusing to remove due to installed plugins from this index")
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

var (
	forceIndexDelete *bool
)

func init() {
	forceIndexDelete = indexDeleteCmd.Flags().Bool("force", false, "Remove index even if it has plugins installed (may yield in unsupported behavior)")

	indexCmd.AddCommand(indexAddCmd)
	indexCmd.AddCommand(indexListCmd)
	indexCmd.AddCommand(indexDeleteCmd)

	if _, ok := os.LookupEnv(constants.EnableMultiIndexSwitch); ok {
		rootCmd.AddCommand(indexCmd)
	}
}
