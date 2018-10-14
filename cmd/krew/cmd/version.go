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

	"github.com/GoogleContainerTools/krew/pkg/environment"
	"github.com/GoogleContainerTools/krew/pkg/version"
	"github.com/golang/glog"
	"github.com/pkg/errors"
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
IndexURI is the URI where the index is updated from.
InstallPath is the base path for all plugin installations.
DownloadPath is the path used to store download binaries.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		selfPath, err := os.Executable()
		if err != nil {
			glog.Fatalf("failed to get the own executable path")
		}

		executedVersion, runningAsPlugin, err := environment.GetExecutedVersion(paths.InstallPath(), selfPath, environment.Realpath)
		if err != nil {
			return errors.Wrap(err, "failed to find current krew version")
		}

		conf := [][]string{
			{"IsPlugin", fmt.Sprintf("%v", runningAsPlugin)},
			{"ExecutedVersion", executedVersion},
			{"GitTag", version.GitTag()},
			{"GitCommit", version.GitCommit()},
			{"IndexURI", IndexURI},
			{"BasePath", paths.BasePath()},
			{"IndexPath", paths.IndexPath()},
			{"InstallPath", paths.InstallPath()},
			{"DownloadPath", paths.DownloadPath()},
			{"BinPath", paths.BinPath()},
		}
		return printTable(os.Stdout, []string{"OPTION", "VALUE"}, conf)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
