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

// validate-krew-manifest makes sure a manifest file is valid.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/index/indexscanner"
	"sigs.k8s.io/krew/internal/index/validation"
	"sigs.k8s.io/krew/internal/installation"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

var flManifest string
var flSkipInstall bool

func init() {
	flag.StringVar(&flManifest, "manifest", "", "path to plugin manifest file")
	flag.BoolVar(&flSkipInstall, "skip-install", false, "skips installing the plugin as part of the validation")
}

func main() {
	klog.InitFlags(nil)
	if err := flag.Set("logtostderr", "true"); err != nil {
		fmt.Fprintf(os.Stderr, "can't set log to stderr: %v\n", err)
		os.Exit(1)
	}
	flag.Parse()

	err := run()
	klog.Flush()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if flManifest == "" {
		return errors.New("-manifest must be specified")
	}
	return validateManifestFile(flManifest)
}

func validateManifestFile(path string) error {
	klog.Infof("reading file %q", path)
	p, err := indexscanner.ReadPluginFromFile(path)
	if err != nil {
		return fmt.Errorf("failed to read plugin file: %w", err)
	}
	filename := filepath.Base(path)
	manifestExtension := filepath.Ext(filename)
	if manifestExtension != constants.ManifestExtension {
		return fmt.Errorf("expected manifest extension %q but found %q", constants.ManifestExtension, manifestExtension)
	}
	pluginNameFromFileName := strings.TrimSuffix(filename, manifestExtension)
	klog.V(4).Infof("inferred plugin name as %s", pluginNameFromFileName)

	// validate plugin manifest
	if err := validation.ValidatePlugin(pluginNameFromFileName, p); err != nil {
		return fmt.Errorf("plugin validation error: %w", err)
	}
	klog.Infof("structural validation OK")

	// make sure each platform matches a supported platform
	for i, p := range p.Spec.Platforms {
		if env := findAnyMatchingPlatform(p.Selector); env.OS == "" || env.Arch == "" {
			return fmt.Errorf("spec.platform[%d]'s selector (%v) doesn't match any supported platforms", i, p.Selector)
		}
	}
	klog.Infof("all spec.platform[] items are used")

	// validate no supported <os,arch> is matching multiple platform specs
	if err := isOverlappingPlatformSelectors(p.Spec.Platforms); err != nil {
		return fmt.Errorf("overlapping platform selectors found: %w", err)
	}
	klog.Infof("no overlapping spec.platform[].selector")

	if !flSkipInstall {
		// exercise "install" for all platforms
		for i, p := range p.Spec.Platforms {
			klog.Infof("installing spec.platform[%d]", i)
			if err := installPlatformSpec(path, p); err != nil {
				return fmt.Errorf("spec.platforms[%d] failed to install: %w", i, err)
			}
			klog.Infof("installed  spec.platforms[%d]", i)
		}
		log.Printf("all %d spec.platforms installed fine", len(p.Spec.Platforms))
	}
	return nil
}

// isOverlappingPlatformSelectors validates if multiple platforms have selectors
// that match to a supported <os,arch> pair.
func isOverlappingPlatformSelectors(platforms []index.Platform) error {
	for _, env := range allPlatforms() {
		var matchIndex []int
		for i, p := range platforms {
			if selectorMatchesOSArch(p.Selector, env) {
				matchIndex = append(matchIndex, i)
			}
		}

		if len(matchIndex) > 1 {
			return fmt.Errorf("multiple spec.platforms (at indexes %v) have overlapping selectors that select %s", matchIndex, env)
		}
	}
	return nil
}

// installPlatformSpec installs the p to a temporary location on disk to verify
// by shelling out to external command.
func installPlatformSpec(manifestFile string, p index.Platform) error {
	env := findAnyMatchingPlatform(p.Selector)
	if env.OS == "" || env.Arch == "" {
		return fmt.Errorf("no supported platform matched platform selector: %+v", p.Selector)
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), "krew-test")
	if err != nil {
		return fmt.Errorf("failed to create temp dir for plugin install: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			klog.Warningf("failed to remove temp dir: %s", tmpDir)
		}
	}()

	cmd := exec.Command("kubectl", "krew", "install", "--manifest", manifestFile, "-v=4")
	cmd.Stdin = nil
	cmd.Env = []string{
		"KREW_ROOT=" + tmpDir,
		"KREW_OS=" + env.OS,
		"KREW_ARCH=" + env.Arch,
	}
	klog.V(2).Infof("installing plugin with: %+v", cmd.Env)
	cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH"))

	b, err := cmd.CombinedOutput()
	if err != nil {
		output := strings.ReplaceAll(string(b), "\n", "\n\t")
		return fmt.Errorf("plugin install command failed: %s: %w", output, err)
	}

	if err := validateLicenseFileExists(tmpDir); err != nil {
		return fmt.Errorf("LICENSE (or alike) file is not extracted from the archive as part of installation: %w", err)
	}
	return nil
}

var licenseFiles = map[string]struct{}{
	"license":      {},
	"license.txt":  {},
	"license.md":   {},
	"licenses":     {},
	"licenses.txt": {},
	"licenses.md":  {},
	"copying":      {},
	"copying.txt":  {},
}

func validateLicenseFileExists(krewRoot string) error {
	dir := environment.NewPaths(krewRoot).InstallPath()
	var files []string
	err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			files = append(files, info.Name())
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk installation directory: %w", err)
	}

	for _, f := range files {
		klog.V(8).Infof("found installed file: %s", f)
		if _, ok := licenseFiles[strings.ToLower(filepath.Base(f))]; ok {
			klog.V(8).Infof("found license file %q", f)
			return nil
		}
	}
	return fmt.Errorf("could not find license file among [%s]", strings.Join(files, ", "))
}

// findAnyMatchingPlatform finds an <os,arch> pair matches to given selector
func findAnyMatchingPlatform(selector *metav1.LabelSelector) installation.OSArchPair {
	for _, p := range allPlatforms() {
		if selectorMatchesOSArch(selector, p) {
			klog.V(4).Infof("%s MATCHED <%s>", selector, p)
			return p
		}
		klog.V(4).Infof("%s didn't match <%s>", selector, p)
	}
	return installation.OSArchPair{}
}

func selectorMatchesOSArch(selector *metav1.LabelSelector, env installation.OSArchPair) bool {
	sel, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		// this should've been caught by validation.ValidatePlatform() earlier
		klog.Warningf("Failed to convert label selector: %+v", selector)
		return false
	}
	return sel.Matches(labels.Set{
		"os":   env.OS,
		"arch": env.Arch,
	})
}

// allPlatforms returns all <os,arch> pairs krew is supported on.
func allPlatforms() []installation.OSArchPair {
	// TODO(ahmetb) find a more authoritative source for this list
	return []installation.OSArchPair{
		{OS: "windows", Arch: "386"},
		{OS: "windows", Arch: "amd64"},
		{OS: "windows", Arch: "arm64"},
		{OS: "linux", Arch: "386"},
		{OS: "linux", Arch: "amd64"},
		{OS: "linux", Arch: "arm"},
		{OS: "linux", Arch: "arm64"},
		{OS: "linux", Arch: "ppc64le"},
		{OS: "darwin", Arch: "386"},
		{OS: "darwin", Arch: "amd64"},
		{OS: "darwin", Arch: "arm64"},
	}
}
