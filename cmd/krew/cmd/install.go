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
	"bufio"
	"fmt"
	"os"

	"sigs.k8s.io/krew/pkg/index"
	"sigs.k8s.io/krew/pkg/index/indexscanner"
	"sigs.k8s.io/krew/pkg/installation"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	var manifest, forceDownloadFile *string

	// installCmd represents the install command
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install kubectl plugins",
		Long: `Install a plugin.

This command can be used to install one or multiple plugins.

Examples:
  To install one or multiple plugins, run:
    kubectl krew install NAME [NAME...]

  To install plugins from a file, run:
    kubectl krew install < file.txt

  (For developers) To provide a custom plugin manifest, use the --manifest
  argument Similarly, instead of downloading files from a URL, you can specify a
  local --archive file:
	kubectl krew install --manifest=FILE [--archive=FILE]

Remarks:
  If a plugin is already installed, it will be skipped.
  Failure to install a plugin will not stop the installation of other plugins.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var pluginNames = make([]string, len(args))
			copy(pluginNames, args)

			if !isTerminal(os.Stdin) && (len(pluginNames) != 0 || *manifest != "") {
				fmt.Fprintln(os.Stderr, "WARNING: Detected stdin, but discarding it because of --manifest or args")
			}

			if !isTerminal(os.Stdin) && (len(pluginNames) == 0 && *manifest == "") {
				fmt.Fprintln(os.Stderr, "Reading plugin names via stdin")
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Split(bufio.ScanLines)
				for scanner.Scan() {
					if name := scanner.Text(); name != "" {
						pluginNames = append(pluginNames, name)
					}
				}
			}

			if len(pluginNames) != 0 && *manifest != "" {
				return errors.New("must specify either specify stdin or --manifest or args")
			}

			if *forceDownloadFile != "" && *manifest == "" {
				return errors.New("--archive can be specified only with --manifest")
			}

			var install []index.Plugin
			for _, name := range pluginNames {
				plugin, err := indexscanner.LoadPluginFileFromFS(paths.IndexPluginsPath(), name)
				if err != nil {
					return errors.Wrapf(err, "failed to load plugin %q from the index", name)
				}
				install = append(install, plugin)
			}

			if *manifest != "" {
				plugin, err := indexscanner.ReadPluginFile(*manifest)
				if err != nil {
					return errors.Wrap(err, "failed to load custom manifest file")
				}
				if err := plugin.Validate(plugin.Name); err != nil {
					return errors.Wrap(err, "plugin manifest validation error")
				}
				install = append(install, plugin)
			}

			if len(install) > 1 && *manifest != "" {
				return errors.New("can't use --manifest option with multiple plugins")
			}

			if len(install) == 0 {
				return cmd.Help()
			}

			// Print plugin namesFromFile
			for _, plugin := range install {
				glog.V(2).Infof("Will install plugin: %s\n", plugin.Name)
			}

			var failed []string
			// Do install
			for _, plugin := range install {
				fmt.Fprintf(os.Stderr, "Installing plugin: %s\n", plugin.Name)
				err := installation.Install(paths, plugin, *forceDownloadFile)
				if err == installation.ErrIsAlreadyInstalled {
					glog.Warningf("Skipping plugin %q, it is already installed", plugin.Name)
					continue
				}
				if err != nil {
					glog.Warningf("failed to install plugin %q: %v", plugin.Name, err)
					failed = append(failed, plugin.Name)
					continue
				}
				if plugin.Spec.Caveats != "" {
					fmt.Fprintln(os.Stderr, prepCaveats(plugin.Spec.Caveats))
				}
				fmt.Fprintf(os.Stderr, "Installed plugin: %s\n", plugin.Name)
			}
			if len(failed) > 0 {
				return errors.Errorf("failed to install some plugins: %+v", failed)
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if *manifest == "" {
				return ensureIndexUpdated(cmd, args)
			}
			glog.V(4).Infof("--manifest specified, not ensuring plugin index")
			return nil
		},
	}

	manifest = installCmd.Flags().String("manifest", "", "(Development-only) specify plugin manifest directly.")
	forceDownloadFile = installCmd.Flags().String("archive", "", "(Development-only) force all downloads to use the specified file")

	rootCmd.AddCommand(installCmd)
}
