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

	"sigs.k8s.io/krew/internal/index/indexscanner"
	"sigs.k8s.io/krew/internal/installation"
	"sigs.k8s.io/krew/internal/pathutil"
	"sigs.k8s.io/krew/pkg/index"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show information about an available plugin",
	Long:  `Show detailed information about an available plugin.`,
	Example: `  kubectl krew info PLUGIN
  kubectl krew info INDEX/PLUGIN`,
	RunE: func(_ *cobra.Command, args []string) error {
		index, plugin := pathutil.CanonicalPluginName(args[0])

		p, err := indexscanner.LoadPluginByName(paths.IndexPluginsPath(index), plugin)
		if os.IsNotExist(err) {
			return errors.Errorf("plugin %q not found in index %q", args[0], index)
		} else if err != nil {
			return errors.Wrap(err, "failed to load plugin manifest")
		}
		printPluginInfo(os.Stdout, index, p)
		return nil
	},
	PreRunE: checkIndex,
	Args:    cobra.ExactArgs(1),
}

func printPluginInfo(out io.Writer, indexName string, plugin index.Plugin) {
	fmt.Fprintf(out, "NAME: %s\n", plugin.Name)
	fmt.Fprintf(out, "INDEX: %s\n", indexName)
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
//	\
//	 | This plugin is great, use it with great care.
//	 | Also, plugin will require the following programs to run:
//	 |  * jq
//	 |  * base64
//	/
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
