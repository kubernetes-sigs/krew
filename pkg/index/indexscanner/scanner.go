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
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/yaml"

	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

// LoadPluginListFromFS will parse and retrieve all plugin files.
func LoadPluginListFromFS(indexDir string) ([]index.Plugin, error) {
	indexDir, err := filepath.EvalSymlinks(indexDir)
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(filepath.Join(indexDir, "plugins"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to open index dir")
	}

	list := make([]index.Plugin, 0, len(files))
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		pluginName := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		p, err := LoadPluginFileFromFS(indexDir, pluginName)
		if err != nil {
			// Index loading shouldn't fail because of one plugin.
			// Show error instead.
			glog.Errorf("failed to load file %q, err: %v", pluginName, err)
			continue
		}
		list = append(list, p)
	}
	glog.V(4).Infof("Found %d plugins in dir %s", len(list), indexDir)
	return list, nil
}

// LoadPluginFileFromFS loads a plugins index file by its name. When plugin
// file not found, it returns an error that can be checked with os.IsNotExist.
func LoadPluginFileFromFS(indexDir, pluginName string) (index.Plugin, error) {
	if !index.IsSafePluginName(pluginName) {
		return index.Plugin{}, errors.Errorf("plugin name %q not allowed", pluginName)
	}

	glog.V(4).Infof("Reading plugin %q", pluginName)
	indexDir, err := filepath.EvalSymlinks(filepath.Join(indexDir, "plugins"))
	if err != nil {
		return index.Plugin{}, err
	}
	p, err := ReadPluginFile(filepath.Join(indexDir, pluginName+constants.ManifestExtension))
	if os.IsNotExist(err) {
		return index.Plugin{}, err
	} else if err != nil {
		return index.Plugin{}, errors.Wrap(err, "failed to read the plugin manifest")
	}
	return p, p.Validate(pluginName)
}

// ReadPluginFile loads a file from the FS. When plugin file not found, it
// returns an error that can be checked with os.IsNotExist.
// TODO(lbb): Add object verification
func ReadPluginFile(indexFilePath string) (index.Plugin, error) {
	f, err := os.Open(indexFilePath)
	if os.IsNotExist(err) {
		return index.Plugin{}, err
	} else if err != nil {
		return index.Plugin{}, errors.Wrap(err, "failed to open index file")
	}
	return DecodePluginFile(f)
}

// DecodePluginFile tries to decodes a plugin manifest from r.
func DecodePluginFile(r io.Reader) (index.Plugin, error) {
	var plugin index.Plugin
	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return plugin, err
	}
	jsonRaw, err := yaml.ToJSON(raw)
	if err != nil {
		return plugin, err
	}
	decoder := json.NewDecoder(bytes.NewReader(jsonRaw))

	// TODO(ahmetb): when we have a stable API that won't add new fields,
	// we can consider failing on unknown fields. Currently disabling due to
	// incremental field additions to plugin manifests independently from the
	// installed version of krew.
	// decoder.DisallowUnknownFields()
	return plugin, decoder.Decode(&plugin)
}
