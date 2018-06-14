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
	"io"
	"os"

	"github.com/google/krew/pkg/index"
	"github.com/google/krew/pkg/index/indexscanner"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Info shows plugin details",
	Long: `Info shows plugin details.
Use this command to find out about plugin requirements and caveats.`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, arg := range args {
			plugin, err := indexscanner.LoadPluginFileFromFS(paths.Index, arg)
			if err != nil {
				glog.Fatal(err)
			}
			printPluginInfo(os.Stdout, plugin)
		}
	},
	PreRunE: checkIndex,
	Args:    cobra.MinimumNArgs(1),
}

// TODO(lbb): Make output fancy (i.e emojis?)
func printPluginInfo(out io.Writer, i index.Plugin) {
	fmt.Fprintf(out, "NAME: %s\n", i.Name)
	if i.Spec.Description != "" {
		fmt.Fprintf(out, "DESCRIPTION: \n%s\n", i.Spec.Description)
	}
	if i.Spec.Version != "" {
		fmt.Fprintf(out, "VERSION: %s\n", i.Spec.Version)
	}
	if i.Spec.Caveats != "" {
		fmt.Fprintf(out, "CAVEATS: \n%s\n", i.Spec.Caveats)
	}
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
