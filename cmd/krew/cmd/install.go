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
	"path/filepath"

	"github.com/google/krew/pkg/index/indexscanner"
	"github.com/google/krew/pkg/installation"

	"github.com/golang/glog"
	"github.com/google/krew/pkg/index"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func init() {
	var forceHEAD *bool
	var manifest *string

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
				fmt.Fprintln(os.Stderr, "Detected Stdin, but discarding it because of --source or args")
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
				return fmt.Errorf("must specify either specify stdin or source or args")
			}

			var install []index.Plugin
			for _, name := range pluginNames {
				plugin, err := indexscanner.LoadPluginFileFromFS(paths.Index, name)
				if err != nil {
					return fmt.Errorf("failed to load plugin %s from index, err: %v", name, err)
				}
				install = append(install, plugin)
			}

			if *manifest != "" {
				file, err := getFileFromArg(*manifest)
				if err != nil {
					return fmt.Errorf("failed to get the file %q, err: %v", *manifest, err)
				}
				plugin, err := indexscanner.ReadPluginFile(file)
				if err != nil {
					return err
				}
				if err := plugin.Validate(plugin.Name); err != nil {
					return fmt.Errorf("failed to validate the plugin file, err %v", err)
				}
				install = append(install, plugin)
			}

			if len(install) > 1 && *forceHEAD {
				return fmt.Errorf("can't use HEAD option with multiple plugins")
			}

			if len(install) == 0 {
				return cmd.Help()
			}

			// Print plugin namesFromFile
			for _, plugin := range install {
				fmt.Printf("Will install plugin: %s\n", plugin.Name)
			}

			var failed []string
			// Do install
			for _, plugin := range install {
				fmt.Printf("Installing plugin: %s\n", plugin.Name)
				err := installation.Install(paths, plugin, *forceHEAD)
				if err == installation.IsAlreadyInstalledErr {
					glog.Warningf("Skipping plugin %s, it is already installed", plugin.Name)
					continue
				}
				if err != nil {
					glog.Warningf("failed to install plugin %q, err: %v", plugin.Name, err)
					failed = append(failed, plugin.Name)
					continue
				}
				if plugin.Spec.Caveats != "" {
					fmt.Printf("CAVEATS: %s\n", plugin.Spec.Caveats)
				}
			}
			if len(failed) > 0 {
				return fmt.Errorf("failed to instlal some plugins: %+v", failed)
			}
			return nil
		},
		PreRunE: ensureUpdated,
	}

	forceHEAD = installCmd.Flags().Bool("HEAD", false, "Force HEAD if versioned and HEAD installs are possible.")
	manifest = installCmd.Flags().String("source", "", "(Development-only) specify plugin manifest directly.")

	rootCmd.AddCommand(installCmd)
}

func getFileFromArg(file string) (string, error) {
	if filepath.IsAbs(file) {
		return file, nil
	}
	abs, err := filepath.Abs(filepath.Join(os.Getenv("PWD"), file))
	if err != nil {
		return "", fmt.Errorf("failed to find absolute filepath, err %v", err)
	}
	return abs, nil
}
