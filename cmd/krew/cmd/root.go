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
	"flag"
	"fmt"
	"os"

	"github.com/google/krew/pkg/environment"
	"github.com/google/krew/pkg/gitutil"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var paths environment.KrewPaths

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "krew",
	Short: "krew is the kubectl plugin manager",
	Long: `krew is the kubectl plugin manager.
You can invoke krew through kubectl with: "kubectl plugin [krew] option..."`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		glog.Fatal(err)
	}
}

// InjectCommand InjectCommand adds a cobra command to the main tree
func InjectCommand(c *cobra.Command) {
	rootCmd.AddCommand(c)
}

func init() {
	cobra.OnInitialize(initConfig)
	// Add glog flags
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	// Set glog default to stderr
	flag.Set("logtostderr", "true")
	// Required by glog
	flag.Parse()

	paths = environment.MustGetKrewPathsFromEnvs(os.Environ())
	if err := ensureDirs(paths.Base, paths.Download, paths.Install); err != nil {
		glog.Fatal(err)
	}
}

func checkIndex(cmd *cobra.Command, _ []string) error {
	if ok, err := gitutil.IsGitCloned(paths.Index); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("krew local plugin index is not initialized (run \"krew update\")")
	}
	return nil
}
func ensureDirs(paths ...string) error {
	for _, p := range paths {
		glog.V(4).Infof("Ensure creating dir: %q", p)
		if err := os.MkdirAll(p, 0755); err != nil {
			return fmt.Errorf("failed to ensure create directory, err: %v", err)
		}
	}
	return nil
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv()
	// Make environment vars kubectl-plugin compatible.
	bindEnvironmentVariables(viper.GetViper(), rootCmd)
}
