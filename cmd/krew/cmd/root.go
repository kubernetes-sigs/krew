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
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/environment"
	"sigs.k8s.io/krew/pkg/gitutil"
	"sigs.k8s.io/krew/pkg/installation"
	"sigs.k8s.io/krew/pkg/installation/receipt"
	"sigs.k8s.io/krew/pkg/receiptsmigration"
)

var (
	paths environment.Paths // krew paths used by the process
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "krew",
	Short: "krew is the kubectl plugin manager",
	Long: `krew is the kubectl plugin manager.
You can invoke krew through kubectl: "kubectl krew [command]..."`,
	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRunE: preRun,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if glog.V(1) {
			glog.Fatalf("%+v", err) // with stack trace
		} else {
			glog.Fatal(err) // just error message
		}
	}
}

func init() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	flag.CommandLine.Parse([]string{}) // convince pkg/flag we parsed the flags
	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		if f.Name != "v" { // hide all glog flags except for -v
			pflag.Lookup(f.Name).Hidden = true
		}
	})
	flag.Set("logtostderr", "true") // Set glog default to stderr

	paths = environment.MustGetKrewPaths()
	if err := ensureDirs(paths.BasePath(),
		paths.DownloadPath(),
		paths.InstallPath(),
		paths.BinPath(),
		paths.InstallReceiptsPath()); err != nil {
		glog.Fatal(err)
	}
}

func preRun(cmd *cobra.Command, _ []string) error {
	// detect if receipts migration (v0.2.x->v0.3.x) is complete
	isMigrated, err := receiptsmigration.Done(paths)
	if err != nil {
		return err
	}
	if !isMigrated && cmd.Use != "receipts-upgrade" {
		fmt.Fprintln(os.Stderr, "You need to perform a migration to continue using krew.\nPlease run `kubectl krew system receipts-upgrade`")
		return fmt.Errorf("krew home outdated")
	}

	if installation.IsWindows() {
		glog.V(4).Infof("detected windows, will check for old krew installations to clean up")
		err := cleanupStaleKrewInstallations()
		if err != nil {
			glog.Warningf("Failed to clean up old installations of krew (on windows).")
			glog.Warningf("You may need to clean them up manually. Error: %v", err)
		}
	}
	return nil
}

func cleanupStaleKrewInstallations() error {
	r, err := receipt.Load(paths.PluginInstallReceiptPath(constants.KrewPluginName))
	if os.IsNotExist(err) {
		glog.V(1).Infof("could not find krew's own plugin receipt, skipping cleanup of stale krew installations")
		return nil
	} else if err != nil {
		return errors.Wrap(err, "cannot load krew's own plugin receipt")
	}
	v := r.Spec.Version

	glog.V(1).Infof("Clean up krew stale installations, current=%s", v)
	return installation.CleanupStaleKrewInstallations(paths.PluginInstallPath(constants.KrewPluginName), v)
}

func checkIndex(_ *cobra.Command, _ []string) error {
	if ok, err := gitutil.IsGitCloned(paths.IndexPath()); err != nil {
		return errors.Wrap(err, "failed to check local index git repository")
	} else if !ok {
		return errors.New(`krew local plugin index is not initialized (run "kubectl krew update")`)
	}
	return nil
}

func ensureDirs(paths ...string) error {
	for _, p := range paths {
		glog.V(4).Infof("Ensure creating dir: %q", p)
		if err := os.MkdirAll(p, 0755); err != nil {
			return errors.Wrapf(err, "failed to ensure create directory %q", p)
		}
	}
	return nil
}

func isTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
