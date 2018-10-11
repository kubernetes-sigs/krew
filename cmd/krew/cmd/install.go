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
	"bufio"
	"fmt"
	"os"

	"github.com/GoogleContainerTools/krew/pkg/index"
	"github.com/GoogleContainerTools/krew/pkg/index/indexscanner"
	"github.com/GoogleContainerTools/krew/pkg/installation"

	"github.com/golang/glog"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	var forceHEAD *bool
	var manifest, forceDownloadFile *string

	// installCmd represents the install command
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install a new plugin",
		Long: `Install a new plugin.
All plugins will be downloaded and made available to: "kubectl plugin <name>"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var pluginNames = make([]string, len(args))
			copy(pluginNames, args)

			if (len(pluginNames) != 0 || *manifest != "") && !(isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())) {
				fmt.Fprintln(os.Stderr, "Detected Stdin, but discarding it because of --manifest or args")
			}

			if len(pluginNames) == 0 && *manifest == "" && !(isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())) {
				fmt.Fprintln(os.Stderr, "Read from standard input")
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
				plugin, err := indexscanner.LoadPluginFileFromFS(paths.IndexPath(), name)
				if err != nil {
					return errors.Wrapf(err, "failed to load plugin %q from the index", name)
				}
				install = append(install, plugin)
			}

			if *manifest != "" {
				// TODO(ahmetb) do not clone index (ensureIndex) in PreRunE when
				// custom manifest is specified (krew initial installation
				// scenario).
				plugin, err := indexscanner.ReadPluginFile(*manifest)
				if err != nil {
					return errors.Wrap(err, "failed to load custom manifest file")
				}
				if err := plugin.Validate(plugin.Name); err != nil {
					return errors.Wrap(err, "plugin manifest validation error")
				}
				install = append(install, plugin)
			}

			if len(install) > 1 && *forceHEAD {
				return errors.New("can't use --HEAD option with multiple plugins")
			}
			if len(install) > 1 && *manifest != "" {
				return errors.New("can't use --manifest option with multiple plugins")
			}

			if len(install) == 0 {
				return cmd.Help()
			}

			// Print plugin namesFromFile
			for _, plugin := range install {
				fmt.Fprintf(os.Stderr, "Will install plugin: %s\n", plugin.Name)
			}

			var failed []string
			// Do install
			for _, plugin := range install {
				glog.V(2).Infof("Installing plugin: %s\n", plugin.Name)
				err := installation.Install(paths, plugin, *forceHEAD, *forceDownloadFile)
				if err == installation.ErrIsAlreadyInstalled {
					glog.Warningf("Skipping plugin %s, it is already installed", plugin.Name)
					continue
				}
				if err != nil {
					glog.Warningf("failed to install plugin %q", plugin.Name)
					failed = append(failed, plugin.Name)
					continue
				}
				fmt.Fprintf(os.Stderr, "Installed plugin: %s\n", plugin.Name)
				if plugin.Spec.Caveats != "" {
					fmt.Fprintf(os.Stderr, "CAVEATS: %s\n", plugin.Spec.Caveats)
				}
			}
			if len(failed) > 0 {
				return errors.Errorf("failed to install some plugins: %+v", failed)
			}
			return nil
		},
		PreRunE: ensureUpdated,
	}

	forceHEAD = installCmd.Flags().Bool("HEAD", false, "Force HEAD if versioned and HEAD installs are possible.")
	manifest = installCmd.Flags().String("manifest", "", "(Development-only) specify plugin manifest directly.")
	forceDownloadFile = installCmd.Flags().String("archive", "", "(Development-only) force all downloads to use the specified file")

	rootCmd.AddCommand(installCmd)
}
