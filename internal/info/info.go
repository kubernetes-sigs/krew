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

package info

import (
	"os"

	"github.com/pkg/errors"
	"k8s.io/klog"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/index/indexscanner"
	"sigs.k8s.io/krew/pkg/index"
)

// LoadManifestFromReceiptOrIndex tries to load a plugin manifest from the
// receipts directory or from the index directory if the former fails.
func LoadManifestFromReceiptOrIndex(p environment.Paths, name string) (index.Plugin, error) {
	receipt, err := indexscanner.LoadPluginByName(p.InstallReceiptsPath(), name)

	if err == nil {
		klog.V(3).Infof("Found plugin manifest for %q in the receipts dir", name)
		return receipt, nil
	}

	if !os.IsNotExist(err) {
		return index.Plugin{}, errors.Wrapf(err, "loading plugin %q from receipts dir", name)
	}

	klog.V(3).Infof("Plugin manifest for %q not found in the receipts dir", name)
	return indexscanner.LoadPluginByName(p.IndexPluginsPath(), name)
}
