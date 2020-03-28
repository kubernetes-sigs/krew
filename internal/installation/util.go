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

package installation

import (
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/klog"

	"sigs.k8s.io/krew/internal/installation/receipt"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

// InstalledPluginsFromIndex returns a list of all install plugins from a particular index.
func InstalledPluginsFromIndex(receiptsDir, indexName string) ([]index.Receipt, error) {
	var out []index.Receipt
	receipts, err := GetInstalledPluginReceipts(receiptsDir)
	if err != nil {
		return nil, err
	}
	for _, r := range receipts {
		if r.Status.Source.Name == indexName {
			out = append(out, r)
		}
	}
	return out, nil
}

// GetInstalledPluginReceipts returns a list of receipts.
func GetInstalledPluginReceipts(receiptsDir string) ([]index.Receipt, error) {
	files, err := filepath.Glob(filepath.Join(receiptsDir, "*"+constants.ManifestExtension))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to glob receipts directory (%s) for manifests", receiptsDir)
	}
	out := make([]index.Receipt, 0, len(files))
	for _, f := range files {
		r, err := receipt.Load(f)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse plugin install receipt %s", f)
		}
		out = append(out, r)
		klog.V(4).Infof("parsed receipt for %s: version=%s", r.GetObjectMeta().GetName(), r.Spec.Version)

	}
	return out, nil
}
