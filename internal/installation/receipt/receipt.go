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

package receipt

import (
	"os"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/krew/internal/index/indexscanner"
	"sigs.k8s.io/krew/pkg/index"
)

// Store saves the given receipt at the destination.
// The caller has to ensure that the destination directory exists.
func Store(receipt index.Receipt, dest string) error {
	yamlBytes, err := yaml.Marshal(receipt)
	if err != nil {
		return errors.Wrapf(err, "convert to yaml")
	}

	err = os.WriteFile(dest, yamlBytes, 0o644)
	return errors.Wrapf(err, "write plugin receipt %q", dest)
}

// Load reads the plugin receipt at the specified destination.
// If not found, it returns os.IsNotExist error.
func Load(path string) (index.Receipt, error) {
	return indexscanner.ReadReceiptFromFile(path)
}

// New returns a new receipt with the given plugin and index name.
func New(plugin index.Plugin, indexName string, timestamp metav1.Time) index.Receipt {
	plugin.CreationTimestamp = timestamp
	return index.Receipt{
		Plugin: plugin,
		Status: index.ReceiptStatus{
			Source: index.SourceIndex{
				Name: indexName,
			},
		},
	}
}
