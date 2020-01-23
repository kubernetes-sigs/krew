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
	"os"

	"github.com/spf13/cobra"

	"sigs.k8s.io/krew/internal/index/indexoperations"
)

var indexConfig *indexoperations.IndexConfig

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Perform krew index commands",
	Long:  "Perform krew index commands such as adding and removing indices.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%+v\n", indexConfig.Indices)
		return nil
	},
}

// indexAddCmd represents the index add command
var indexAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a custom index to download plugins from",
	Long:  "Add a custom index to download plugins from",
	RunE: func(cmd *cobra.Command, args []string) error {
		return indexConfig.AddIndex(args[0], args[1])
	},
}

var indexRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a custom index that you've added",
	Long:  "Remove a custom index that you've added",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Printf("Must provide an index name that you want to remove")
			os.Exit(1)
		}
		return indexConfig.RemoveIndex(args[0])
	},
}

func init() {
	ic, err := indexoperations.GetIndexConfig()
	if err != nil {
		fmt.Println("AHHHHHHHHH!")
		os.Exit(1)
	}
	indexConfig = ic
	indexCmd.AddCommand(indexAddCmd)
	indexCmd.AddCommand(indexRemoveCmd)
	rootCmd.AddCommand(indexCmd)
}
