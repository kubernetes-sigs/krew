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
	"os"

	"github.com/pkg/errors"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/index/indexoperations"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

// indexOf returns the index name of a receipt.
func indexOf(r index.Receipt) string {
	if r.Status.Source.Name == "" {
		return constants.DefaultIndexName
	}
	return r.Status.Source.Name
}

// displayName returns the display name of a Plugin.
// The index name is omitted if it is the default index.
func displayName(p index.Plugin, indexName string) string {
	if isDefaultIndex(indexName) {
		return p.Name
	}
	return indexName + "/" + p.Name
}

func isDefaultIndex(name string) bool {
	return name == "" || name == constants.DefaultIndexName
}

// canonicalName returns INDEX/NAME value for a plugin, even if
// it is in the default index.
func canonicalName(p index.Plugin, indexName string) string {
	if isDefaultIndex(indexName) {
		indexName = constants.DefaultIndexName
	}
	return indexName + "/" + p.Name
}

// allIndexes returns a slice of index name and URL pairs
func allIndexes(p environment.Paths) ([]indexoperations.Index, error) {
	indexes := []indexoperations.Index{
		{
			Name: constants.DefaultIndexName,
			URL:  constants.DefaultIndexURI,
		},
	}
	if os.Getenv(constants.EnableMultiIndexSwitch) != "" {
		out, err := indexoperations.ListIndexes(p)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list plugin indexes available")
		}
		if len(out) != 0 {
			indexes = out
		}
	}
	return indexes, nil
}
