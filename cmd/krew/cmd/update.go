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

	"github.com/GoogleContainerTools/krew/pkg/gitutil"
	"github.com/golang/glog"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// IndexURI points to the upstream index.
const IndexURI = "https://github.com/GoogleContainerTools/krew-index.git"

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update local plugin index",
	Long: `Update local plugin index.
Fetch the newest version of Krew and all formulae from GitHub using git(1) and
perform any necessary migrations.`,
	RunE: ensureIndexUpdated,
}

func ensureIndexUpdated(_ *cobra.Command, _ []string) error {
	glog.V(1).Infof("Updating the local copy of plugin index (%s)", paths.IndexPath())
	if err := gitutil.EnsureUpdated(IndexURI, paths.IndexPath()); err != nil {
		return errors.Wrap(err, "failed to update the local index")
	}
	fmt.Fprintln(os.Stderr, "Updated the local copy of plugin index.")
	return nil
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
