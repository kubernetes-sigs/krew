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

package indexscanner

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/krew/internal/index/validation"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

func findPluginManifestFiles(indexDir string) ([]string, error) {
	var out []string
	files, err := os.ReadDir(indexDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open index dir")
	}
	for _, file := range files {
		if file.Type().IsRegular() && filepath.Ext(file.Name()) == constants.ManifestExtension {
			out = append(out, file.Name())
		}
	}
	return out, nil
}

// LoadPluginListFromFS will parse and retrieve all plugin files.
func LoadPluginListFromFS(indexDir string) ([]index.Plugin, error) {
	indexDir, err := filepath.EvalSymlinks(indexDir)
	if err != nil {
		return nil, err
	}

	files, err := findPluginManifestFiles(indexDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to scan plugins in index directory")
	}
	klog.V(4).Infof("found %d plugins in dir %s", len(files), indexDir)

	list := make([]index.Plugin, 0, len(files))
	for _, file := range files {
		pluginName := strings.TrimSuffix(file, filepath.Ext(file))
		p, err := LoadPluginByName(indexDir, pluginName)
		if err != nil {
			// Index loading shouldn't fail because of one plugin.
			// Show error instead.
			klog.Errorf("failed to read or parse plugin manifest %q: %v", pluginName, err)
			continue
		}
		list = append(list, p)
	}
	return list, nil
}

// LoadPluginByName loads a plugins index file by its name. When plugin
// file not found, it returns an error that can be checked with os.IsNotExist.
func LoadPluginByName(pluginsDir, pluginName string) (index.Plugin, error) {
	klog.V(4).Infof("Reading plugin %q from %s", pluginName, pluginsDir)
	return ReadPluginFromFile(filepath.Join(pluginsDir, pluginName+constants.ManifestExtension))
}

// ReadPluginFromFile loads a file from the FS. When plugin file not found, it
// returns an error that can be checked with os.IsNotExist.
func ReadPluginFromFile(path string) (index.Plugin, error) {
	var plugin index.Plugin
	err := readFromFile(path, &plugin)
	if err != nil {
		return plugin, err
	}
	return plugin, errors.Wrap(validation.ValidatePlugin(plugin.Name, plugin), "plugin manifest validation error")
}

func ReadPlugin(f io.ReadCloser) (index.Plugin, error) {
	var plugin index.Plugin
	err := decodeFile(f, &plugin)
	if err != nil {
		return plugin, errors.Wrap(err, "failed to decode plugin manifest")
	}
	return plugin, errors.Wrap(validation.ValidatePlugin(plugin.Name, plugin), "plugin manifest validation error")
}

// ReadReceiptFromFile loads a file from the FS. When receipt file not found, it
// returns an error that can be checked with os.IsNotExist.
func ReadReceiptFromFile(path string) (index.Receipt, error) {
	var receipt index.Receipt
	err := readFromFile(path, &receipt)
	if receipt.Status.Source.Name == "" {
		receipt.Status.Source.Name = constants.DefaultIndexName
	}
	return receipt, err
}

// readFromFile loads a file from the FS into the provided object.
func readFromFile(path string, as interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	err = decodeFile(f, &as)
	return errors.Wrapf(err, "failed to parse yaml file %q", path)
}

// decodeFile tries to decode a plugin/receipt
func decodeFile(r io.ReadCloser, as interface{}) error {
	defer r.Close()
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// TODO(ahmetb): when we have a stable API that won't add new fields,
	// we can consider failing on unknown fields. Currently, disabling due to
	// incremental field additions to plugin manifests independently from the
	// installed version of krew.
	// yaml.UnmarshalStrict()
	return yaml.Unmarshal(b, &as)
}
