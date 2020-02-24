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
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog"

	"sigs.k8s.io/krew/internal/gitutil"
	"sigs.k8s.io/krew/internal/index/indexscanner"
	"sigs.k8s.io/krew/internal/installation"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the local copy of the plugin index",
	Long: `Update the local copy of the plugin index.

This command synchronizes the local copy of the plugin manifests with the
plugin index from the internet.

Remarks:
  You don't need to run this command: Running "krew update" or "krew upgrade"
  will silently run this command.`,
	RunE: ensureIndexUpdated,
}

func showFormattedPluginsInfo(out io.Writer, header string, plugins []string) {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("  %s:\n", header))

	for _, p := range plugins {
		b.WriteString(fmt.Sprintf("    * %s\n", p))
	}

	fmt.Fprintf(out, "%s", b.String())
}

func showUpdatedPlugins(out io.Writer, preUpdate, posUpdate []index.Plugin, installedPlugins map[string]string) {
	var newPlugins []index.Plugin
	var updatedPlugins []index.Plugin

	oldIndex := make(map[string]index.Plugin)
	for _, p := range preUpdate {
		oldIndex[p.Name] = p
	}

	for _, p := range posUpdate {
		old, ok := oldIndex[p.Name]
		if !ok {
			newPlugins = append(newPlugins, p)
			continue
		}

		if _, ok := installedPlugins[p.Name]; !ok {
			continue
		}

		if old.Spec.Version != p.Spec.Version {
			updatedPlugins = append(updatedPlugins, p)
		}
	}

	if len(newPlugins) > 0 {
		var s []string
		for _, p := range newPlugins {
			s = append(s, p.Name)
		}

		showFormattedPluginsInfo(out, "New plugins available", s)
	}

	if len(updatedPlugins) > 0 {
		var s []string
		for _, p := range updatedPlugins {
			old := oldIndex[p.Name]
			s = append(s, fmt.Sprintf("%s %s -> %s", p.Name, old.Spec.Version, p.Spec.Version))
		}

		showFormattedPluginsInfo(out, "Upgrades available for installed plugins", s)
	}
}

func ensureIndexUpdated(_ *cobra.Command, _ []string) error {
	preUpdateIndex, _ := indexscanner.LoadPluginListFromFS(paths.IndexPluginsPath(constants.DefaultIndexName))

	klog.V(1).Infof("Updating the local copy of plugin index (%s)", paths.IndexPath(constants.DefaultIndexName))
	if err := gitutil.EnsureUpdated(constants.IndexURI, paths.IndexPath(constants.DefaultIndexName)); err != nil {
		return errors.Wrap(err, "failed to update the local index")
	}
	fmt.Fprintln(os.Stderr, "Updated the local copy of plugin index.")

	if len(preUpdateIndex) == 0 {
		return nil
	}

	posUpdateIndex, err := indexscanner.LoadPluginListFromFS(paths.IndexPluginsPath(constants.DefaultIndexName))
	if err != nil {
		return errors.Wrap(err, "failed to load plugin index after update")
	}

	installedPlugins, err := installation.ListInstalledPlugins(paths.InstallReceiptsPath())
	if err != nil {
		return errors.Wrap(err, "failed to load installed plugins list after update")
	}

	showUpdatedPlugins(os.Stderr, preUpdateIndex, posUpdateIndex, installedPlugins)

	return nil
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
