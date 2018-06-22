// Copyright © 2018 Google Inc.
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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/krew/pkg/index"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// LoadIndexListFromFS will parse and retrieve all index files.
func LoadIndexListFromFS(indexdir string) (index.IndexList, error) {
	var indexList index.IndexList
	indexdir, err := filepath.EvalSymlinks(indexdir)
	if err != nil {
		return indexList, err
	}

	files, err := ioutil.ReadDir(indexdir)
	if err != nil {
		return indexList, fmt.Errorf("failed to open index dir, err: %v", err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		pluginName := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		// TODO(lbb): Use go routines to speed up slow FS operations.
		p, err := LoadPluginFileFromFS(indexdir, pluginName)
		if err != nil {
			// Index loading shouldn't fail because of one plugin.
			// Show error instead.
			glog.Errorf("failed to load file %q, err: %v", pluginName, err)
			continue
		}

		indexList.Items = append(indexList.Items, p)
	}

	return indexList, nil
}

// LoadPluginFileFromFS loads a plugins index file by its name.
func LoadPluginFileFromFS(indexdir, pluginName string) (index.Plugin, error) {
	if !IsSafepluginName(pluginName) {
		return index.Plugin{}, fmt.Errorf("plugin name %q not allowed", pluginName)
	}
	indexdir, err := filepath.EvalSymlinks(indexdir)
	if err != nil {
		return index.Plugin{}, err
	}
	p, err := ReadPluginFile(filepath.Join(indexdir, pluginName+".yaml"))
	if err != nil {
		return index.Plugin{}, fmt.Errorf("failed to read the plugin file, err: %v", err)
	}
	if p.Name != pluginName {
		return index.Plugin{}, fmt.Errorf("can't accept plugin with different plugin name, requested name=%q, loaded name=%q", pluginName, p.Name)
	}
	return p, nil
}

// ReadPluginFile loads a file from the FS
// TODO(lbb): Add object verification
func ReadPluginFile(indexFilePath string) (index.Plugin, error) {
	f, err := os.Open(indexFilePath)
	if err != nil {
		return index.Plugin{}, fmt.Errorf("failed to open index file, err: %v", err)
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
	decoder.DisallowUnknownFields()
	return plugin, decoder.Decode(&plugin)
}

// IsSafepluginName checks if the plugin Name is save to use.
func IsSafepluginName(name string) bool {
	return !strings.ContainsAny(name, string([]rune{filepath.Separator, filepath.ListSeparator, '.'}))
}
