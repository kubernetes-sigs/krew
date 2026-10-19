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
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"os"
	"strings"
	"sync/atomic"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"sigs.k8s.io/krew/cmd/krew/cmd/internal"
	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/gitutil"
	"sigs.k8s.io/krew/internal/indexmigration"
	"sigs.k8s.io/krew/internal/installation"
	"sigs.k8s.io/krew/internal/installation/receipt"
	"sigs.k8s.io/krew/internal/installation/semver"
	"sigs.k8s.io/krew/internal/receiptsmigration"
	"sigs.k8s.io/krew/internal/version"
	"sigs.k8s.io/krew/pkg/constants"
)

const (
	upgradeNotification = "A newer version of krew is available (%s -> %s).\nRun \"kubectl krew upgrade\" to get the newest version!\n"

	// upgradeCheckRate is the percentage of krew runs for which the upgrade check is performed.
	upgradeCheckRate = 0.4
)

var (
	paths environment.Paths // krew paths used by the process

	// latestTag is updated by a goroutine with the latest tag from GitHub.
	// A nil value indicates that the API request was skipped or has not
	// completed.
	latestTag atomic.Pointer[string]
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "krew", // This is prefixed by kubectl in the custom usage template
	Short: "krew is the kubectl plugin manager",
	Long: `krew is the kubectl plugin manager.
You can invoke krew through kubectl: "kubectl krew [command]..."`,
	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRunE: preRun,
	PersistentPostRun: showUpgradeNotification,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// It returns the process exit code.
func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func init() {
	klog.InitFlags(nil)
	// Opt into fixed stderrthreshold behavior (kubernetes/klog#212).
	_ = flag.Set("legacy_stderr_threshold_behavior", "false")
	_ = flag.Set("stderrthreshold", "INFO")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	_ = flag.CommandLine.Parse([]string{}) // convince pkg/flag we parsed the flags

	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		if f.Name != "v" { // hide all glog flags except for -v
			pflag.Lookup(f.Name).Hidden = true
		}
	})
	if err := flag.Set("logtostderr", "true"); err != nil {
		fmt.Fprintf(os.Stderr, "can't set log to stderr: %v\n", err)
		os.Exit(1)
	}

	// Cobra doesn't have a way to specify a two word command (ie. "kubectl krew"), so set a custom usage template
	// with kubectl in it. Cobra will use this template for the root and all child commands.
	rootCmd.SetUsageTemplate(strings.NewReplacer(
		"{{.UseLine}}", "kubectl {{.UseLine}}",
		"{{.CommandPath}}", "kubectl {{.CommandPath}}").Replace(rootCmd.UsageTemplate()))
}

func preRun(_ *cobra.Command, _ []string) error {
	paths = environment.MustGetKrewPaths()

	// check must be done before ensureDirs, to detect krew's self-installation
	if !internal.IsBinDirInPATH(paths) {
		internal.PrintWarning(os.Stderr, "%s", internal.SetupInstructions()+"\n\n")
	}

	if err := ensureDirs(paths.BasePath(),
		paths.InstallPath(),
		paths.BinPath(),
		paths.IndexBase(),
		paths.InstallReceiptsPath()); err != nil {
		return err
	}

	go func() {
		if _, disabled := os.LookupEnv("KREW_NO_UPGRADE_CHECK"); disabled ||
			isDevelopmentBuild() || // no upgrade check for dev builds
			upgradeCheckRate < rand.Float64() { // only do the upgrade check randomly
			klog.V(1).Infof("skipping upgrade check")
			return
		}
		klog.V(1).Infof("starting upgrade check")
		tag, err := internal.FetchLatestTag()
		if err != nil {
			klog.V(1).Infoln("WARNING:", err)
			return
		}
		latestTag.Store(&tag)
	}()

	// Detect whether the receipts migration (v0.2.x->v0.3.x) is complete.
	// Keep this check for users who still migrate from v0.2; remove it in a
	// future release once that upgrade path is no longer supported.
	isMigrated, err := receiptsmigration.Done(paths)
	if err != nil {
		return err
	}
	if !isMigrated {
		fmt.Fprintln(os.Stderr, `This version of Krew is not supported anymore. Please manually migrate:
1. Uninstall Krew: https://krew.sigs.k8s.io/docs/user-guide/setup/uninstall/
2. Install latest Krew: https://krew.sigs.k8s.io/docs/user-guide/setup/install/
3. Install the plugins you used`)
		return errors.New("krew home outdated")
	}

	isMigrated, err = indexmigration.Done(paths)
	if err != nil {
		return fmt.Errorf("failed to check if index migration is complete: %w", err)
	}
	if !isMigrated {
		if err := indexmigration.Migrate(paths); err != nil {
			return fmt.Errorf("index migration failed: %w", err)
		}
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
	tag := latestTag.Load()
	if tag == nil {
		klog.V(4).Infof("Upgrade check was skipped or has not finished")
		return
	}
	latestVer, err := semver.Parse(*tag)
	if err != nil {
		klog.V(4).Infof("Could not parse remote tag as semver: %s", err)
		return
	}
	currentVer, err := semver.Parse(version.GitTag())
	if err != nil {
		klog.V(4).Infof("Could not parse current tag as semver: %s", err)
		return
	}
	if semver.Less(currentVer, latestVer) {
		color.New(color.Bold).Fprintf(os.Stderr, upgradeNotification, version.GitTag(), *tag)
	} else {
		klog.V(4).Infof("upgrade check found no new versions (%s>=%s)", currentVer, latestVer)
	}
}

func cleanupStaleKrewInstallations() error {
	r, err := receipt.Load(paths.PluginInstallReceiptPath(constants.KrewPluginName))
	if errors.Is(err, fs.ErrNotExist) {
		klog.V(1).Infof("could not find krew's own plugin receipt, skipping cleanup of stale krew installations")
		return nil
	} else if err != nil {
		return fmt.Errorf("cannot load krew's own plugin receipt: %w", err)
	}
	v := r.Spec.Version

	klog.V(1).Infof("Clean up krew stale installations, current=%s", v)
	return installation.CleanupStaleKrewInstallations(paths.PluginInstallPath(constants.KrewPluginName), v)
}

func checkIndex(_ *cobra.Command, _ []string) error {
	entries, err := os.ReadDir(paths.IndexBase())
	if err != nil {
		return fmt.Errorf("failed to list directory: %w", err)
	}
	if len(entries) == 0 {
		return errors.New(`krew local plugin index is not initialized (run "kubectl krew update")`)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		indexPath := paths.IndexPath(e.Name())
		if ok, err := gitutil.IsGitCloned(indexPath); err != nil {
			return fmt.Errorf("failed to check local index git repository: %w", err)
		} else if !ok {
			return fmt.Errorf("invalid index %q, non git directory found in index folder", e.Name())
		}
	}
	return nil
}

func ensureDirs(paths ...string) error {
	for _, p := range paths {
		klog.V(4).Infof("Ensure creating dir: %q", p)
		if err := os.MkdirAll(p, 0o755); err != nil {
			return fmt.Errorf("failed to ensure create directory %q: %w", p, err)
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
