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
	"github.com/golang/glog"
	"github.com/google/krew/pkg/gitutil"
	"github.com/spf13/cobra"
)

// TODO(lbb): Replace with real index later.
const IndexURI = "file:///Users/lbb/git/krew-index"

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update local plugin index",
	Long: `Update local plugin index.
Fetch the newest version of Krew and all formulae from GitHub using git(1) and
perform any necessary migrations.`,
	Run: ensureUpdated,
}

func ensureUpdated(cmd *cobra.Command, args []string) {
	if err := gitutil.EnsureUpdated(IndexURI, paths.Index); err != nil {
		glog.Fatal(err)
	}
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
