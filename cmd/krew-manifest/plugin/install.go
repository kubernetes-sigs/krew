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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// NewGenerateCmd builds a command that generates plugin.yaml files.
func NewGenerateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate a plugin manifest for krew",
		RunE: func(cmd *cobra.Command, args []string) error {
			krewCommand, topCommands := traversPluginEntryPoints(cmd.Root())
			b, err := yaml.Marshal(krewCommand)
			if err != nil {
				return fmt.Errorf("failed to marshal root plugin, err: %v", err)
			}
			glog.Infof("Creating command \"kubectl plugin krew\" under path \"plugin.yaml\"")
			if err = ioutil.WriteFile("plugin.yaml", b, 0644); err != nil {
				return fmt.Errorf("failed to write \"plugin.yaml\", err: %v", err)
			}
			for _, command := range topCommands {
				if err = os.MkdirAll(filepath.Join("commands", command.Name), 0755); err != nil {
					return fmt.Errorf("failed to create commands dir, err: %v", err)
				}
				b, err := yaml.Marshal(command)
				if err != nil {
					return fmt.Errorf("failed to marshal root plugin with name %q, err: %v", command.Name, err)
				}
				pluginFilePath := filepath.Join("commands", command.Name, "plugin.yaml")
				glog.Infof("Creating command \"kubectl plugin %s\" under path %q", command.Name, pluginFilePath)
				if ioutil.WriteFile(pluginFilePath, b, 0644); err != nil {
					return fmt.Errorf("failed to write plugin %q, err: %v", pluginFilePath, err)
				}
			}
			return nil
		},
	}
}
