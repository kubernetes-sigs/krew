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

	"github.com/golang/glog"
	"github.com/google/krew/pkg/index/indexscanner"

	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Discover plugins in your local index using fuzzy search",
	Long: `Discover plugins in your local index using fuzzy search.
Search accepts a list of words as options. Search will weight fuzzy matches
in (Name,Intro, Description) in descending order.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(lbb): Implement real search, don't just list plugin names.
		index, err := indexscanner.LoadIndexListFromFS(paths.Index)
		if err != nil {
			glog.Fatal(err)
		}
		for _, i := range index.Items {
			fmt.Println(i.Name)
		}
	},
	PreRunE: checkIndex,
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
