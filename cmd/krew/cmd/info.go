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
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"sigs.k8s.io/krew/internal/info"
	"sigs.k8s.io/krew/internal/installation"
	"sigs.k8s.io/krew/pkg/index"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show information about a kubectl plugin",
	Long: `Show information about a kubectl plugin.

This command can be used to print information such as its download URL, last
available version, platform availability and the caveats.

Example:
  kubectl krew info PLUGIN`,
	RunE: func(cmd *cobra.Command, args []string) error {
		plugin, err := info.LoadManifestFromReceiptOrIndex(paths, args[0])
		if os.IsNotExist(err) {
			return errors.Errorf("plugin %q not found", args[0])
		} else if err != nil {
			return errors.Wrap(err, "failed to load plugin manifest")
		}
		printPluginInfo(os.Stdout, plugin)
		return nil
	},
	PreRunE: checkIndex,
	Args:    cobra.ExactArgs(1),
}

func printPluginInfo(out io.Writer, plugin index.Plugin) {
	fmt.Fprintf(out, "NAME: %s\n", plugin.Name)
	if platform, ok, err := installation.GetMatchingPlatform(plugin.Spec.Platforms); err == nil && ok {
		if platform.URI != "" {
			fmt.Fprintf(out, "URI: %s\n", platform.URI)
			fmt.Fprintf(out, "SHA256: %s\n", platform.Sha256)
		}
	}
	if plugin.Spec.Version != "" {
		fmt.Fprintf(out, "VERSION: %s\n", plugin.Spec.Version)
	}
	if plugin.Spec.Homepage != "" {
		fmt.Fprintf(out, "HOMEPAGE: %s\n", plugin.Spec.Homepage)
	}
	if plugin.Spec.Description != "" {
		fmt.Fprintf(out, "DESCRIPTION: \n%s\n", plugin.Spec.Description)
	}
	if plugin.Spec.Caveats != "" {
		fmt.Fprintf(out, "CAVEATS:\n%s\n", indent(plugin.Spec.Caveats))
	}
}

// indent converts strings to an indented format ready for printing.
// Example:
//
//     \
//      | This plugin is great, use it with great care.
//      | Also, plugin will require the following programs to run:
//      |  * jq
//      |  * base64
//     /
func indent(s string) string {
	out := "\\\n"
	s = strings.TrimRightFunc(s, unicode.IsSpace)
	out += regexp.MustCompile("(?m)^").ReplaceAllString(s, " | ")
	out += "\n/"
	return out
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
