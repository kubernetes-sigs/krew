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

// todo(corneliusweig) remove migration code with v0.4
package receiptsmigration

import (
	"os"

	"sigs.k8s.io/krew/internal/environment"
)

// Done checks if the krew installation requires a migration.
// It considers a migration necessary when plugins are installed, but no receipts are present.
func Done(newPaths environment.Paths) (bool, error) {
	receipts, err := os.ReadDir(newPaths.InstallReceiptsPath())
	if err != nil {
		return false, err
	}
	plugins, err := os.ReadDir(newPaths.BinPath())
	if err != nil {
		return false, err
	}

	hasInstalledPlugins := len(plugins) > 0
	hasNoReceipts := len(receipts) == 0

	return !(hasInstalledPlugins && hasNoReceipts), nil
}
