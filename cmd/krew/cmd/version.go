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

	"github.com/spf13/cobra"

	"sigs.k8s.io/krew/internal/installation"
	"sigs.k8s.io/krew/internal/version"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show krew version and diagnostics",
	Long: `Show version information and diagnostics about krew itself.

Remarks:
  - GitTag describes the release name krew is built from.
  - GitCommit describes the git revision ID which krew is built from.
  - DefaultIndexURI is the URI where the index is updated from.
  - BasePath is the root directory for krew installation.
  - IndexPath is the directory that stores the local copy of the index git repository.
  - InstallPath is the directory for plugin installations.
  - BinPath is the directory for the symbolic links to the installed plugin executables.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		conf := [][]string{
			{"GitTag", version.GitTag()},
			{"GitCommit", version.GitCommit()},
			{"IndexURI", index.DefaultIndex()},
			{"BasePath", paths.BasePath()},
			{"IndexPath", paths.IndexPath(constants.DefaultIndexName)},
			{"InstallPath", paths.InstallPath()},
			{"BinPath", paths.BinPath()},
			{"DetectedPlatform", installation.OSArch().String()},
		}
		return printTable(os.Stdout, []string{"OPTION", "VALUE"}, conf)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
