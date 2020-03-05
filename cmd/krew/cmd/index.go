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
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/krew/pkg/constants"
)

// listCmd represents the list command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Perform krew index commands",
	Long: `Show a list of installed kubectl plugins and their versions.

Remarks:
  Redirecting the output of this command to a program or file will only print
  the names of the plugins installed. This output can be piped back to the
  "install" command.`,
	Args:   cobra.NoArgs,
	Hidden: true,
}

var indexListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured indexes",
	Long:  `Print a list of configured indexes.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dirs, err := ioutil.ReadDir(paths.IndexBase())
		if err != nil {
			return errors.Wrapf(err, "failed to read directory %s", paths.IndexBase())
		}
		for _, dir := range dirs {
			fmt.Fprintln(os.Stdout, dir.Name())
		}
		return nil
	},
}

func init() {
	if _, ok := os.LookupEnv(constants.EnableMultiIndexSwitch); ok {
		indexCmd.AddCommand(indexListCmd)
		rootCmd.AddCommand(indexCmd)
	}
}
