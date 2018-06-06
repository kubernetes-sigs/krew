//  Copyright © 2018 Google Inc.
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

package index

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/google/krew/pkg/index"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// LoadIndexListFromFS will parse and retrieve all index files.
func LoadIndexListFromFS(indexdir string) (*index.IndexList, error) {
	indexdir, err := filepath.EvalSymlinks(indexdir)
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(indexdir)
	if err != nil {
		return nil, fmt.Errorf("failed to open index dir, err: %v", err)
	}

	indexList := &index.IndexList{}
	for _, f := range files {
		if f.IsDir() || !(strings.HasSuffix(f.Name(), ".yaml") || strings.HasSuffix(f.Name(), ".json")) {
			continue
		}

		fpath, err := filepath.EvalSymlinks(filepath.Join(indexdir, f.Name()))
		if err != nil {
			return nil, err
		}
		index, err := readIndexFile(fpath)
		if err != nil {
			glog.Errorf("skip index file %s err: %v", fpath, err)
			continue
		}

		indexList.Items = append(indexList.Items, *index)
	}

	return indexList, nil
}

// LoadIndexFileFromFS loads a plugins index file by its name.
func LoadIndexFileFromFS(indexdir, pluginName string) (*index.Index, error) {
	indexdir, err := filepath.EvalSymlinks(indexdir)
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(indexdir)
	if err != nil {
		return nil, fmt.Errorf("failed to open dir %s, err: %v", indexdir, err)
	}

	for _, f := range files {
		if f.IsDir() || !(f.Name() == pluginName+".yaml" || f.Name() == pluginName+".json") {
			continue
		}
		fpath := filepath.Join(indexdir, f.Name())
		return readIndexFile(fpath)
	}
	return nil, fmt.Errorf("could not find the plugin %q", pluginName)
}

// readIndexFile loads a file from the FS
// TODO(lbb): Add object verification
func readIndexFile(indexFilePath string) (*index.Index, error) {
	f, err := os.Open(indexFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open index file, err: %v", err)
	}
	index := &index.Index{}
	return index, yaml.NewYAMLOrJSONDecoder(f, 1024).Decode(index)
}
