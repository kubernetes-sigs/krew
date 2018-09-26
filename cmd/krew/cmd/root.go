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

	"github.com/GoogleContainerTools/krew/pkg/environment"
	"github.com/GoogleContainerTools/krew/pkg/gitutil"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var paths environment.KrewPaths
var krewExecutedVersion string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "krew",
	Short: "krew is the kubectl plugin manager",
	Long: `krew is the kubectl plugin manager.
You can invoke krew through kubectl with: "kubectl plugin [krew] option..."`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		glog.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	// Set glog default to stderr
	flag.Set("logtostderr", "true")
	// Required by glog
	flag.Parse()

	paths = environment.MustGetKrewPaths()
	if err := ensureDirs(paths.Base, paths.Download, paths.Install, paths.Bin); err != nil {
		glog.Fatal(err)
	}

	selfPath, err := os.Executable()
	if err != nil {
		glog.Errorf("failed to get the own executable path")
	}
	if krewVersion, ok, err := environment.GetExecutedVersion(paths.Install, selfPath, environment.Realpath); err != nil {
		glog.Fatal(fmt.Errorf("failed to find current krew version, err: %v", err))
	} else if ok {
		krewExecutedVersion = krewVersion
	}

	SetGlogFlags(krewExecutedVersion != "")
}

// SetGlogFlags will add glog flags to the CLI. This command can be executed multiple times.
func SetGlogFlags(hidden bool) {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	// Add glog flags if not run as a plugin.
	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		pflag.Lookup(f.Name).Hidden = hidden
	})
}

func checkIndex(_ *cobra.Command, _ []string) error {
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
}
