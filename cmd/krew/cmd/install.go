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

	"github.com/golang/glog"
	"github.com/google/krew/pkg/index"
	"github.com/google/krew/pkg/index/indexscanner"
	"github.com/google/krew/pkg/installation"
	"github.com/spf13/cobra"
)

func init() {
	var (
		forceHEAD *bool
		file      *string
	)

	// installCmd represents the install command
	var installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install a new plugin",
		Long: `Install a new plugin.
All plugins will be downloaded and made available to: "kubectl plugin <name>"`,
		Run: func(cmd *cobra.Command, args []string) {
			var install []index.Plugin
			// From args
			for _, arg := range args {
				plugin, err := indexscanner.LoadPluginFileFromFS(paths.Index, arg)
				if err != nil {
					glog.Fatal(err)
				}
				install = append(install, plugin)
			}

			// Check if is stdin convention
			var f *os.File
			if *file == "-" {
				f = os.Stdin
			} else if *file != "" {
				fi, err := os.Open(*file)
				if err != nil {
					glog.Fatal(fmt.Errorf("failed to open provided file, err %v", err))
				}
				defer fi.Close()
				f = fi
			}
			if f != nil {
				plugin, err := indexscanner.DecodePluginFile(f)
				if err != nil {
					glog.Fatal(fmt.Errorf("failed to decode provided file, err %v", err))
				}
				install = append(install, plugin)
			}

			if len(install) > 1 && *forceHEAD {
				glog.Fatalln(fmt.Errorf("Cannot use HEAD option with multiple plugins"))
			}
			// Print plugin names
			for _, plugin := range install {
				glog.Infof("Will install plugin: %s", plugin.Name)
			}
			// Do install
			for _, plugin := range install {
				if err := installation.Install(paths, plugin, *forceHEAD); err != nil {
					glog.Fatalln(err)
				}
			}
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if (len(args) == 0 && *file == "") || (len(args) > 0 && *file != "") {
				return fmt.Errorf("must specify either names or files")
			}
			return nil
		},
		PreRun: ensureUpdated,
	}
	forceHEAD = installCmd.Flags().Bool("HEAD", false, "Force HEAD if versioned and HEAD installs are possible.")
	file = installCmd.Flags().StringP("file", "f", "", "File with a list of plugin names to install.")

	rootCmd.AddCommand(installCmd)
}
