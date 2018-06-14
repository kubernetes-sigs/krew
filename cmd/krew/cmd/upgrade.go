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
	"github.com/google/krew/pkg/index/indexscanner"
	"github.com/google/krew/pkg/installation"
	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade an installed plugin to a newer version",
	Long: `Upgrade an installed plugin to a newer version.
This will reinstall all plugins that have a newer version in the local index.
Use "kubectl plugin update" to renew the index. All plugins that rely on HEAD
will always be installed.
To only upgrade single plugins provide them as arguments:
kubectl plugin upgrade foo bar"`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, arg := range args {
			index, err := indexscanner.LoadPluginFileFromFS(paths.Index, arg)
			if err != nil {
				glog.Fatal(err)
			}
			if err = installation.Upgrade(paths, index); err != nil {
				glog.Fatalln(err)
			}
		}
	},
	PreRunE: checkIndex,
	Args:    cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
