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
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"sigs.k8s.io/krew/internal/gitutil"
	"sigs.k8s.io/krew/internal/index/indexoperations"
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
	RunE: ensureIndexes,
}

func showFormattedPluginsInfo(out io.Writer, header string, plugins []string) {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("  %s:\n", header))

	for _, p := range plugins {
		b.WriteString(fmt.Sprintf("    * %s\n", p))
	}

	fmt.Fprintf(out, "%s", b.String())
}

func showUpdatedPlugins(out io.Writer, preUpdate, postUpdate []pluginEntry, installedPlugins map[string]string) {
	var newPlugins []pluginEntry
	var updatedPlugins []pluginEntry

	oldIndexMap := make(map[string]pluginEntry)
	for _, p := range preUpdate {
		oldIndexMap[canonicalName(p.p, p.indexName)] = p
	}

	for _, p := range postUpdate {
		cName := canonicalName(p.p, p.indexName)
		old, ok := oldIndexMap[cName]
		if !ok {
			newPlugins = append(newPlugins, p)
			continue
		}

		if _, ok := installedPlugins[cName]; !ok {
			continue
		}

		if old.p.Spec.Version != p.p.Spec.Version {
			updatedPlugins = append(updatedPlugins, p)
		}
	}

	if len(newPlugins) > 0 {
		var s []string
		for _, p := range newPlugins {
			s = append(s, displayName(p.p, p.indexName))
		}
		showFormattedPluginsInfo(out, "New plugins available", s)
	}

	if len(updatedPlugins) > 0 {
		var s []string
		for _, p := range updatedPlugins {
			old := oldIndexMap[canonicalName(p.p, p.indexName)]
			name := displayName(p.p, p.indexName)
			s = append(s, fmt.Sprintf("%s %s -> %s", name, old.p.Spec.Version, p.p.Spec.Version))
		}
		showFormattedPluginsInfo(out, "Upgrades available for installed plugins", s)
	}
}

// loadPlugins loads plugin entries from specified indexes. Parse errors
// are ignored and logged.
func loadPlugins(indexes []indexoperations.Index) []pluginEntry {
	var out []pluginEntry
	for _, idx := range indexes {
		list, err := indexscanner.LoadPluginListFromFS(paths.IndexPluginsPath(idx.Name))
		if err != nil {
			klog.V(1).Infof("WARNING: failed to load plugin list from %q: %v", idx.Name, err)
			continue
		}
		for _, v := range list {
			out = append(out, pluginEntry{indexName: idx.Name, p: v})
		}
	}
	return out
}

func ensureIndexes(_ *cobra.Command, _ []string) error {
	klog.V(3).Infof("Will check if there are any indexes added.")
	if err := ensureDefaultIndexIfNoneExist(); err != nil {
		return err
	}
	return ensureIndexesUpdated()
}

// ensureDefaultIndexIfNoneExist adds the default index automatically
// (and informs the user about it) if no plugin index exists for krew.
func ensureDefaultIndexIfNoneExist() error {
	idx, err := indexoperations.ListIndexes(paths)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve plugin indexes")
	}
	if len(idx) > 0 {
		klog.V(3).Infof("Found %d indexes, skipping adding default index.", len(idx))
		return nil
	}

	klog.V(3).Infof("No index found, add default index.")
	defaultIndex := index.DefaultIndex()
	fmt.Fprintf(os.Stderr, "Adding \"default\" plugin index from %s.\n", defaultIndex)
	return errors.Wrap(indexoperations.AddIndex(paths, constants.DefaultIndexName, defaultIndex),
		"failed to add default plugin index in absence of no indexes")
}

// ensureIndexesUpdated iterates over all indexes and updates them
// and prints new plugins and upgrades available for installed plugins.
func ensureIndexesUpdated() error {
	indexes, err := indexoperations.ListIndexes(paths)
	if err != nil {
		return errors.Wrap(err, "failed to list indexes")
	}

	// collect list of existing plugins
	preUpdatePlugins := loadPlugins(indexes)

	var failed []string
	var returnErr error
	for _, idx := range indexes {
		indexPath := paths.IndexPath(idx.Name)
		klog.V(1).Infof("Updating the local copy of plugin index (%s)", indexPath)
		if err := gitutil.EnsureUpdated(idx.URL, indexPath); err != nil {
			klog.Warningf("failed to update index %q: %v", idx.Name, err)
			failed = append(failed, idx.Name)
			if returnErr == nil {
				returnErr = err
			}
			continue
		}

		if isDefaultIndex(idx.Name) {
			fmt.Fprintln(os.Stderr, "Updated the local copy of plugin index.")
		} else {
			fmt.Fprintf(os.Stderr, "Updated the local copy of plugin index %q.\n", idx.Name)
		}
	}

	if len(preUpdatePlugins) != 0 {
		postUpdatePlugins := loadPlugins(indexes)

		receipts, err := installation.GetInstalledPluginReceipts(paths.InstallReceiptsPath())
		if err != nil {
			return errors.Wrap(err, "failed to load installed plugins list after update")
		}
		installedPlugins := make(map[string]string)
		for _, receipt := range receipts {
			installedPlugins[canonicalName(receipt.Plugin, indexOf(receipt))] = receipt.Spec.Version
		}
		showUpdatedPlugins(os.Stderr, preUpdatePlugins, postUpdatePlugins, installedPlugins)
	}
	return errors.Wrapf(returnErr, "failed to update the following indexes: %s", strings.Join(failed, ", "))
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
