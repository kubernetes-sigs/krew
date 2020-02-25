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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/klog"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/krew/internal/index/validation"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

func findPluginManifestFiles(indexDir string) ([]string, error) {
	var out []string
	files, err := ioutil.ReadDir(indexDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open index dir")
	}
	for _, file := range files {
		if file.Mode().IsRegular() && filepath.Ext(file.Name()) == constants.ManifestExtension {
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
	if !validation.IsSafePluginName(pluginName) {
		return index.Plugin{}, errors.Errorf("plugin name %q not allowed", pluginName)
	}

	klog.V(4).Infof("Reading plugin %q", pluginName)
	return ReadPluginFromFile(filepath.Join(pluginsDir, pluginName+constants.ManifestExtension))
}

// ReadPluginFromFile loads a file from the FS. When plugin file not found, it
// returns an error that can be checked with os.IsNotExist.
func ReadPluginFromFile(path string) (index.Plugin, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		// TODO(ahmetb): we should use go1.13+ errors.Is construct at call sites to evaluate if an error is os.IsNotExist
		return index.Plugin{}, err
	} else if err != nil {
		return index.Plugin{}, errors.Wrap(err, "failed to open index file")
	}
	return ReadPlugin(f)
}

func ReadPlugin(f io.ReadCloser) (index.Plugin, error) {
	defer f.Close()
	plugin := &index.Plugin{}
	err := decodeFile(f, plugin)
	p := *plugin
	if err != nil {
		return p, errors.Wrap(err, "failed to decode plugin manifest")
	}
	return p, errors.Wrap(validation.ValidatePlugin(p.Name, p), "plugin manifest validation error")
}

// ReadReceiptFromFile loads a file from the FS. When receipt file not found, it
// returns an error that can be checked with os.IsNotExist.
func ReadReceiptFromFile(path string) (index.Receipt, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		// TODO(ahmetb): we should use go1.13+ errors.Is construct at call sites to evaluate if an error is os.IsNotExist
		return index.Receipt{}, err
	} else if err != nil {
		return index.Receipt{}, errors.Wrap(err, "failed to open index file")
	}
	return ReadReceipt(f)
}

func ReadReceipt(f io.ReadCloser) (index.Receipt, error) {
	defer f.Close()
	receipt := &index.Receipt{}
	err := decodeFile(f, receipt)
	r := *receipt
	if err != nil {
		return r, errors.Wrap(err, "failed to decode plugin manifest")
	}
	return r, errors.Wrap(validation.ValidateReceipt(r.Name, r), "receipt manifest validation error")
}

// decodeFile tries to decode
func decodeFile(r io.Reader, as interface{}) error {
	b, err := ioutil.ReadAll(r)
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
