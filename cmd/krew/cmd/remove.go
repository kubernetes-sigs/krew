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

	"github.com/GoogleContainerTools/krew/pkg/installation"
	"github.com/pkg/errors"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a plugin from the system",
	Long: `Remove a plugin from the system.
This will delete all plugin related files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, name := range args {
			glog.V(4).Infof("Going to remove plugin %s\n", name)
			if err := installation.Remove(paths, name); err != nil {
				return errors.Wrapf(err, "failed to remove plugin %s", name)
			}
			fmt.Fprintf(os.Stderr, "Removed plugin %s\n", name)
		}
		return nil
	},
	PreRunE: checkIndex,
	Args:    cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
