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

	"github.com/spf13/cobra"

	"sigs.k8s.io/krew/internal/index/indexoperations"
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:    "index",
	Short:  "Perform krew index commands",
	Long:   "Perform krew index commands such as adding and removing indices.",
	Args:   cobra.NoArgs,
	Hidden: true,
}

// indexAddCmd represents the index add command
var indexAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a custom index to download plugins from",
	Long:  "Add a custom index to download plugins from",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("index add: %+v\n", args)
		indexConfig, err := indexoperations.GetIndexConfig()
		if err != nil {
			return err
		}
		return indexConfig.WriteIndexConfig(args[0], args[1])
	},
}

var indexRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a custom index that you've added",
	Long:  "Remove a custom index that you've added",
	RunE: func(cmd *cobra.Command, args []string) error {
		indexConfig, err := indexoperations.GetIndexConfig()
		if err != nil {
			return err
		}
		return indexConfig.RemoveIndex(args[0])
	},
}

var indexListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured indices",
	Long:  "List configured indices",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		indexConfig, err := indexoperations.GetIndexConfig()
		if err != nil {
			return err
		}
		fmt.Printf("%+v\n", indexConfig.Indices)
		return nil
	},
}

func init() {
	indexCmd.AddCommand(indexAddCmd)
	indexCmd.AddCommand(indexRemoveCmd)
	indexCmd.AddCommand(indexListCmd)
	rootCmd.AddCommand(indexCmd)
}
