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
	"os"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/gitutil"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the local copy of the plugin index",
	Long: `Update the local copy of the plugin index.

This command synchronizes the local copy of the plugin manifests with the
plugin index from the internet.

Remarks:
  You don't need to run this command: Running "krew update" or "krew upgrade"
  will silently run this command.`,
	RunE: ensureIndexUpdated,
}

func ensureIndexUpdated(_ *cobra.Command, _ []string) error {
	glog.V(1).Infof("Updating the local copy of plugin index (%s)", paths.IndexPath())
	if err := gitutil.EnsureUpdated(constants.IndexURI, paths.IndexPath()); err != nil {
		return errors.Wrap(err, "failed to update the local index")
	}
	fmt.Fprintln(os.Stderr, "Updated the local copy of plugin index.")
	return nil
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
