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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// bindEnvironmentVariables will allow the cobra command to accept kubectl
// styled flags. https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/
func bindEnvironmentVariables(v *viper.Viper, cmd *cobra.Command) {
	// set logging option
	v.BindEnv("v", "KUBECTL_PLUGINS_GLOBAL_FLAG_V")
	bindPrefixedEnvironmentVariables(v, cmd)
}

func bindPrefixedEnvironmentVariables(v *viper.Viper, cmd *cobra.Command) {
	v.SetEnvPrefix("KUBECTL_PLUGINS_LOCAL_FLAG")
	v.BindPFlags(cmd.Flags())
	v.AutomaticEnv()
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed && v.IsSet(f.Name) {
			cmd.Flags().Set(f.Name, v.GetString(f.Name))
		}
	})
}
