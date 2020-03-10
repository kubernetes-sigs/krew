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

package indexoperations

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"sigs.k8s.io/krew/internal/gitutil"
)

// Index describes the name and URL of a configured index.
type Index struct {
	Name string
	URL  string
}

// ListIndexes returns a slice of Index objects. The path argument is used as
// the base path of the index.
func ListIndexes(path string) ([]Index, error) {
	dirs, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read directory %s", path)
	}

	indexes := []Index{}
	for _, dir := range dirs {
		indexName := dir.Name()
		remote, err := gitutil.GetRemoteURL(filepath.Join(path, indexName))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list the remote URL for index %s", indexName)
		}

		indexes = append(indexes, Index{
			Name: indexName,
			URL:  remote,
		})
	}
	return indexes, nil
}

// AddIndex initializes a new index to install plugins from.
func AddIndex(path, name, url string) error {
	dir := filepath.Join(path, name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return gitutil.EnsureCloned(url, dir)
	} else if err != nil {
		return errors.Wrapf(err, "failed to describe %s", dir)
	}
	return errors.New("index already exists")
}
