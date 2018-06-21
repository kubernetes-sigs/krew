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

package plugin

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var reservedKubectlFlags = map[string]struct{}{
	"alsologtostderr":       {},
	"as":                    {},
	"as-group":              {},
	"cache-dir":             {},
	"certificate-authority": {},
	"client-certificate":    {},
	"client-key":            {},
	"cluster":               {},
	"context":               {},
	"help":                  {},
	"insecure-skip-tls-verify": {},
	"kubeconfig":               {},
	"log-backtrace-at":         {},
	"log_backtrace_at":         {},
	"log-dir":                  {},
	"log_dir":                  {},
	"log-flush-frequency":      {},
	"logtostderr":              {},
	"match-server-version":     {},
	"n":               {},
	"namespace":       {},
	"password":        {},
	"request-timeout": {},
	"s":               {},
	"server":          {},
	"stderrthreshold": {},
	"token":           {},
	"user":            {},
	"username":        {},
	"v":               {},
	"vmodule":         {},
}

// Plugin holds everything needed to register a
// plugin as a command. Usually comes from a descriptor file.
// From https://github.com/kubernetes/kubernetes/blob/23cd1434e69c1a984e8c24c875c19ccbdd0ba2fe/pkg/kubectl/plugins/plugins.go#L47
type Plugin struct {
	Name      string   `yaml:"name"`
	Use       string   `yaml:"use"`
	ShortDesc string   `yaml:"shortDesc"`
	LongDesc  string   `yaml:"longDesc,omitempty"`
	Example   string   `yaml:"example,omitempty"`
	Command   string   `yaml:"command"`
	Flags     []Flag   `yaml:"flags,omitempty"`
	Tree      []Plugin `yaml:"tree,omitempty"`
}

// Flag describes a single flag supported by a given plugin.
// From https://github.com/kubernetes/kubernetes/blob/23cd1434e69c1a984e8c24c875c19ccbdd0ba2fe/pkg/kubectl/plugins/plugins.go#L93
type Flag struct {
	Name      string `yaml:"name"`
	Shorthand string `yaml:"shorthand,omitempty"`
	Desc      string `yaml:"desc"`
	DefValue  string `yaml:"defValue,omitempty"`
}

// traversPluginEntryPoints traverses a cobra command from it's root to generate
// plugin.yaml file as entry points for kubectl. It is assumed that the
// injected command "krew generate" is.
func traversPluginEntryPoints(cmd *cobra.Command) (Plugin, []Plugin) {
	plugin := convertToPlugin(cmd)
	return plugin, convertPluginToPlugins(plugin)
}

func convertPluginToPlugins(root Plugin) []Plugin {
	plugins := make([]Plugin, len(root.Tree))
	for i, p := range root.Tree {
		p.Flags = append(p.Flags, root.Flags...)
		p.Command = filepath.Join("..", "..", p.Command)
		p.ShortDesc = "[krew] " + p.ShortDesc
		plugins[i] = p
	}
	return plugins
}

func convertToPlugin(cmd *cobra.Command) Plugin {
	p := Plugin{
		Name:      strings.Split(cmd.Use, " ")[0],
		Use:       cmd.Use,
		ShortDesc: cmd.Short,
		LongDesc:  cmd.Long,
		Command:   strings.Join([]string{".", cmd.CommandPath()}, string(filepath.Separator)),
		Example:   cmd.Example,
	}
	// The plugin won't validate if empty
	if p.ShortDesc == "" {
		p.ShortDesc = " "
	}

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if f, ok := convertToFlag(flag); ok {
			p.Flags = append(p.Flags, f)
		}
	})

	for _, subCmd := range cmd.Commands() {
		// Skip the injected generator command
		if subCmd.CommandPath() != "krew generate" {
			p.Tree = append(p.Tree, convertToPlugin(subCmd))
		}
	}

	return p
}

func convertToFlag(src *pflag.Flag) (Flag, bool) {
	if _, reserved := reservedKubectlFlags[src.Name]; reserved {
		return Flag{}, false
	}

	dest := Flag{Name: src.Name, Desc: src.Usage}
	if _, reserved := reservedKubectlFlags[src.Shorthand]; !reserved {
		dest.Shorthand = src.Shorthand
	}

	return dest, true
}
