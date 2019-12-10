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
	errs "errors"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog"

	"sigs.k8s.io/krew/cmd/krew/cmd/internal"
	"sigs.k8s.io/krew/internal/index/indexscanner"
	"sigs.k8s.io/krew/internal/index/validation"
	"sigs.k8s.io/krew/internal/installation"
	"sigs.k8s.io/krew/pkg/index"
)

func init() {
	var (
		manifest, archiveFileOverride *string
		noUpdateIndex                 *bool
	)

	// installCmd represents the install command
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install kubectl plugins",
		Long: `Install one or multiple kubectl plugins.

Examples:
  To install one or multiple plugins, run:
    kubectl krew install NAME [NAME...]

  To install plugins from a newline-delimited file, run:
    kubectl krew install < file.txt

  (For developers) To provide a custom plugin manifest, use the --manifest
  argument. Similarly, instead of downloading files from a URL, you can specify a
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
				return errors.New("must specify either specify either plugin names (via positional arguments or STDIN), or --manifest; not both")
			}

			if *archiveFileOverride != "" && *manifest == "" {
				return errors.New("--archive can be specified only with --manifest")
			}

			installed, err := installation.ListInstalledPlugins(paths.InstallReceiptsPath())
			if err != nil {
				return errors.Wrap(err, "failed to find all installed versions")
			}
			skipped := false
			var install []index.Plugin
			for _, name := range pluginNames {
				if _, ok := installed[name]; ok {
					fmt.Fprintf(os.Stderr, "plugin \"%s\" is already installed\n", name)
					skipped = true
					continue
				}
				plugin, err := indexscanner.LoadPluginFileFromFS(paths.IndexPluginsPath(), name)
				if err != nil {
					if errs.Is(err, os.ErrNotExist) {
						return fmt.Errorf("plugin \"%s\" does not exist in the plugin index", name)
					}
					return errors.Wrapf(err, "failed to load plugin %q from the index", name)
				}
				install = append(install, plugin)
			}

			if *manifest != "" {
				plugin, err := indexscanner.ReadPluginFile(*manifest)
				if err != nil {
					return errors.Wrap(err, "failed to load custom manifest file")
				}
				if err := validation.ValidatePlugin(plugin.Name, plugin); err != nil {
					return errors.Wrap(err, "plugin manifest validation error")
				}
				install = append(install, plugin)
			}

			if len(install) == 0 {
				if skipped {
					return nil
				}
				return cmd.Help()
			}

			for _, plugin := range install {
				klog.V(2).Infof("Will install plugin: %s\n", plugin.Name)
			}

			var failed []string
			var returnErr error
			for _, plugin := range install {
				fmt.Fprintf(os.Stderr, "Installing plugin: %s\n", plugin.Name)
				err := installation.Install(paths, plugin, installation.InstallOpts{
					ArchiveFileOverride: *archiveFileOverride,
				})
				if err == installation.ErrIsAlreadyInstalled {
					klog.Warningf("Skipping plugin %q, it is already installed", plugin.Name)
					continue
				}
				if err != nil {
					klog.Warningf("failed to install plugin %q: %v", plugin.Name, err)
					if returnErr == nil {
						returnErr = err
					}
					failed = append(failed, plugin.Name)
					continue
				}
				fmt.Fprintf(os.Stderr, "Installed plugin: %s\n", plugin.Name)
				output := fmt.Sprintf("Use this plugin:\n\tkubectl %s\n", plugin.Name)
				if plugin.Spec.Homepage != "" {
					output += fmt.Sprintf("Documentation:\n\t%s\n", plugin.Spec.Homepage)
				}
				if plugin.Spec.Caveats != "" {
					output += fmt.Sprintf("Caveats:\n%s\n", indent(plugin.Spec.Caveats))
				}
				fmt.Fprintln(os.Stderr, indent(output))
				internal.PrintSecurityNotice()
			}
			if len(failed) > 0 {
				return errors.Wrapf(returnErr, "failed to install some plugins: %+v", failed)
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if *manifest != "" {
				klog.V(4).Infof("--manifest specified, not ensuring plugin index")
				return nil
			}
			if *noUpdateIndex {
				klog.V(4).Infof("--no-update-index specified, skipping updating local copy of plugin index")
				return nil
			}
			return ensureIndexUpdated(cmd, args)
		},
	}

	manifest = installCmd.Flags().String("manifest", "", "(Development-only) specify plugin manifest directly.")
	archiveFileOverride = installCmd.Flags().String("archive", "", "(Development-only) force all downloads to use the specified file")
	noUpdateIndex = installCmd.Flags().Bool("no-update-index", false, "(Experimental) do not update local copy of plugin index before installing")

	rootCmd.AddCommand(installCmd)
}
