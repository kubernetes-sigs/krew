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
	"regexp"

	"github.com/pkg/errors"

	"sigs.k8s.io/krew/internal/gitutil"
	"sigs.k8s.io/krew/pkg/constants"
)

var (
	NoIndexMessage   = "No index configured, run kubectl krew index add default " + constants.IndexURI + " to get started"
	validNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
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
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
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
	if !IsValidIndexName(name) {
		return errors.New("index name cannot contain path characters")
	}

	dir := filepath.Join(path, name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return gitutil.EnsureCloned(url, dir)
	} else if err != nil {
		return err
	}
	return errors.New("index already exists")
}

// IsValidIndexName validates if an index name contains invalid characters
func IsValidIndexName(name string) bool {
	return validNamePattern.MatchString(name)
}
