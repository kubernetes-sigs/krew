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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"sigs.k8s.io/krew/internal/index/validation"
	"sigs.k8s.io/krew/internal/installation"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall plugins",
	Long: `Uninstall one or more plugins.

Example:
  kubectl krew uninstall NAME [NAME...]

Remarks:
  Failure to uninstall a plugin will result in an error and exit immediately.`,
	RunE: func(_ *cobra.Command, args []string) error {
		for _, name := range args {
			if isCanonicalName(name) {
				return errors.New("uninstall command does not support INDEX/PLUGIN syntax; just specify PLUGIN")
			} else if !validation.IsSafePluginName(name) {
				return unsafePluginNameErr(name)
			}
			klog.V(4).Infof("Going to uninstall plugin %s\n", name)
			if err := installation.Uninstall(paths, name); err != nil {
				return errors.Wrapf(err, "failed to uninstall plugin %s", name)
			}
			fmt.Fprintf(os.Stderr, "Uninstalled plugin: %s\n", name)
		}
		return nil
	},
	PreRunE: checkIndex,
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"remove", "rm"},
}

func unsafePluginNameErr(n string) error { return errors.Errorf("plugin name %q not allowed", n) }

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
