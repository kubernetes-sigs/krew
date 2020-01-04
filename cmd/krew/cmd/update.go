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
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog"

	"sigs.k8s.io/krew/internal/gitutil"
	"sigs.k8s.io/krew/internal/index/indexscanner"
	"sigs.k8s.io/krew/internal/installation"
	"sigs.k8s.io/krew/pkg/constants"
)

type newPlugin struct {
	name    string
	version string
}

type newVersionPlugin struct {
	name       string
	installed  bool
	oldVersion string
	newVersion string
}

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

func retrieveUpdatedPluginList() ([]string, error) {
	modifiedFiles, err := gitutil.ListModifiedFiles(constants.IndexURI, paths.IndexPluginsPath())
	if err != nil {
		return []string{}, err
	}

	var plugins []string
	for _, f := range modifiedFiles {
		filename := filepath.Base(f)
		extension := filepath.Ext(filename)
		name := filename[0 : len(filename)-len(extension)]
		plugins = append(plugins, name)
	}

	return plugins, nil
}

func retrievePluginNameVersionMap(names []string) map[string]string {
	m := make(map[string]string, len(names))
	for _, n := range names {
		plugin, err := indexscanner.LoadPluginByName(paths.IndexPluginsPath(), n)
		if err != nil {
			continue
		}

		m[n] = plugin.Spec.Version
	}

	return m
}

func retrieveInstalledPluginMap() (map[string]bool, error) {
	plugins, err := installation.ListInstalledPlugins(paths.InstallReceiptsPath())
	if err != nil {
		return map[string]bool{}, err
	}
	m := make(map[string]bool, len(plugins))
	for name := range plugins {
		m[name] = true
	}

	return m, nil
}

func filterAndSortUpdatedPlugins(old, updated map[string]string, installed map[string]bool) ([]newPlugin, []newVersionPlugin) {
	var newPluginList []newPlugin
	var newVersionList []newVersionPlugin

	for name, version := range updated {
		oldVersion, ok := old[name]
		if !ok {
			newPluginList = append(newPluginList, newPlugin{
				name:    name,
				version: version,
			})
			continue
		}

		if version != oldVersion {
			_, installed := installed[name]
			newVersionList = append(newVersionList, newVersionPlugin{
				name:       name,
				installed:  installed,
				oldVersion: oldVersion,
				newVersion: version,
			})
			continue
		}
	}

	sort.Slice(newPluginList, func(i, j int) bool {
		return newPluginList[i].name < newPluginList[j].name
	})

	sort.Slice(newVersionList, func(i, j int) bool {
		if newVersionList[i].installed && !newVersionList[j].installed {
			return true
		}

		if !newVersionList[i].installed && newVersionList[j].installed {
			return false
		}

		return newVersionList[i].name < newVersionList[j].name
	})

	return newPluginList, newVersionList
}

func showUpdatedPlugins(newPluginList []newPlugin, newVersionList []newVersionPlugin) {
	if len(newPluginList) > 0 {
		fmt.Fprintln(os.Stderr, "  New plugins available:")
		for _, np := range newPluginList {
			fmt.Fprintf(os.Stderr, "    * %s %s\n", np.name, np.version)
		}
	}

	if len(newVersionList) > 0 {
		fmt.Fprintln(os.Stderr, "  The following plugins have new version:")
		for _, np := range newVersionList {
			if np.installed {
				fmt.Fprintf(os.Stderr, "    * %s %s -> %s (!)\n", np.name, np.oldVersion, np.newVersion)
				continue
			}
			fmt.Fprintf(os.Stderr, "    * %s %s -> %s\n", np.name, np.oldVersion, np.newVersion)
		}
	}
}

func ensureIndexUpdated(_ *cobra.Command, _ []string) error {
	updatedPlugins, err := retrieveUpdatedPluginList()
	if err != nil {
		return errors.Wrap(err, "failed to load the list of updated plugins from the index")
	}

	var oldMap map[string]string
	if len(updatedPlugins) > 0 {
		oldMap = retrievePluginNameVersionMap(updatedPlugins)
	}

	klog.V(1).Infof("Updating the local copy of plugin index (%s)", paths.IndexPath())
	if err := gitutil.EnsureUpdated(constants.IndexURI, paths.IndexPath()); err != nil {
		return errors.Wrap(err, "failed to update the local index")
	}
	fmt.Fprintln(os.Stderr, "Updated the local copy of plugin index.")

	if len(updatedPlugins) < 0 {
		return nil
	}

	updatedMap := retrievePluginNameVersionMap(updatedPlugins)

	installedMap, err := retrieveInstalledPluginMap()
	if err != nil {
		return errors.Wrap(err, "failed to find all installed versions")
	}

	newPluginList, newVersionList := filterAndSortUpdatedPlugins(oldMap, updatedMap, installedMap)

	showUpdatedPlugins(newPluginList, newVersionList)

	return nil
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
