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
	"math/rand"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog"

	"sigs.k8s.io/krew/cmd/krew/cmd/internal"
	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/gitutil"
	"sigs.k8s.io/krew/internal/installation"
	"sigs.k8s.io/krew/internal/installation/receipt"
	"sigs.k8s.io/krew/internal/installation/semver"
	"sigs.k8s.io/krew/internal/receiptsmigration"
	"sigs.k8s.io/krew/internal/updatecheck"
	"sigs.k8s.io/krew/internal/version"
	"sigs.k8s.io/krew/pkg/constants"
)

const (
	upgradeNotification = "A newer version of krew is available (%s -> %s).\nRun \"kubectl krew upgrade\" to get the newest version!\n"

	// showRate is the percentage of krew runs for which the upgrade check is performed.
	showRate = 0.4
)

var (
	paths environment.Paths // krew paths used by the process

	// latestTag is updated by a go-routine with the latest tag from GitHub.
	// An empty string indicates that the API request was skipped or
	// has not completed.
	latestTag = ""
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
	PersistentPostRun: showUpgradeNotification,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if klog.V(1) {
			klog.Fatalf("%+v", err) // with stack trace
		} else {
			klog.Fatal(err) // just error message
		}
	}
}

func init() {
	klog.InitFlags(nil)
	rand.Seed(time.Now().UnixNano())

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	_ = flag.CommandLine.Parse([]string{}) // convince pkg/flag we parsed the flags

	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		if f.Name != "v" { // hide all glog flags except for -v
			pflag.Lookup(f.Name).Hidden = true
		}
	})
	if err := flag.Set("logtostderr", "true"); err != nil {
		fmt.Printf("can't set log to stderr %+v", err)
		os.Exit(1)
	}

	paths = environment.MustGetKrewPaths()
}

func preRun(cmd *cobra.Command, _ []string) error {
	// check must be done before ensureDirs, to detect krew's self-installation
	if !internal.IsBinDirInPATH(paths) {
		boldRed := color.New(color.FgRed, color.Bold).SprintfFunc()
		fmt.Fprintf(os.Stderr, "%s: %s\n\n", boldRed("WARNING"), internal.SetupInstructions())
	}

	if err := ensureDirs(paths.BasePath(),
		paths.InstallPath(),
		paths.BinPath(),
		paths.InstallReceiptsPath()); err != nil {
		klog.Fatal(err)
	}

	go func() {
		if _, disabled := os.LookupEnv("KREW_NO_UPGRADE_CHECK"); disabled ||
			isDevelopmentBuild() || // no upgrade check for dev builds
			showRate < rand.Float64() { // only do the upgrade check randomly
			return
		}
		var err error
		latestTag, err = updatecheck.FetchLatestTag()
		if err != nil {
			klog.V(1).Infoln("WARNING:", err)
		}
	}()

	// detect if receipts migration (v0.2.x->v0.3.x) is complete
	isMigrated, err := receiptsmigration.Done(paths)
	if err != nil {
		return err
	}
	if !isMigrated && cmd.Use != "receipts-upgrade" {
		fmt.Fprintln(os.Stderr, "You need to perform a migration to continue using krew.\nPlease run `kubectl krew system receipts-upgrade`")
		return errors.New("krew home outdated")
	}

	if installation.IsWindows() {
		klog.V(4).Infof("detected windows, will check for old krew installations to clean up")
		err := cleanupStaleKrewInstallations()
		if err != nil {
			klog.Warningf("Failed to clean up old installations of krew (on windows).")
			klog.Warningf("You may need to clean them up manually. Error: %v", err)
		}
	}

	return nil
}

func showUpgradeNotification(*cobra.Command, []string) {
	if latestTag == "" || latestTag == version.GitTag() {
		klog.V(4).Infof("Skipping upgrade notification (latest=%q, current=%q)", latestTag, version.GitTag())
		return
	}
	color.New(color.Bold).Fprintf(os.Stderr, upgradeNotification, version.GitTag(), latestTag)
}

func cleanupStaleKrewInstallations() error {
	r, err := receipt.Load(paths.PluginInstallReceiptPath(constants.KrewPluginName))
	if os.IsNotExist(err) {
		klog.V(1).Infof("could not find krew's own plugin receipt, skipping cleanup of stale krew installations")
		return nil
	} else if err != nil {
		return errors.Wrap(err, "cannot load krew's own plugin receipt")
	}
	v := r.Spec.Version

	klog.V(1).Infof("Clean up krew stale installations, current=%s", v)
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
		klog.V(4).Infof("Ensure creating dir: %q", p)
		if err := os.MkdirAll(p, 0755); err != nil {
			return errors.Wrapf(err, "failed to ensure create directory %q", p)
		}
	}
	return nil
}

func isTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

// isDevelopmentBuild tries to parse this builds tag as semver.
// If it fails, this usually means that this is a development build.
func isDevelopmentBuild() bool {
	_, err := semver.Parse(version.GitTag())
	return err != nil
}
