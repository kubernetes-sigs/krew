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
	"os"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version prints the current executable variables",
	Long: `Version prints the current executable variables.
ExecutedVersion is the version of the currently executed binary. This is detected through the path.
IsPlugin is true if the binary is executed as a plugin.
BasePath is the root path for all krew related binaries.
IndexPath is the path to the index repo see git(1).
InstallPath is the base path for all plugin installations.
DownloadPath is the path used to store download binaries.`,
	Run: func(cmd *cobra.Command, args []string) {
		conf := map[string]string{
			"IsPlugin":        fmt.Sprintf("%v", krewExecutedVersion != ""),
			"ExecutedVersion": krewExecutedVersion,
			"BasePath":        paths.Base,
			"IndexPath":       paths.Index,
			"InstallPath":     paths.Install,
			"DownloadPath":    paths.Download,
		}
		printAlignedColums(os.Stdout, "OPTION", "VALUE", conf)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
