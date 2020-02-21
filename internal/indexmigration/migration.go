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

package indexmigration

import (
	"os"

	"k8s.io/klog"
	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/gitutil"
	"sigs.k8s.io/krew/pkg/constants"
)

// Done checks if the krew installation requires a migration to support multiple indexes.
// A migration is necessary when the index directory doesn't contain a "default" directory.
func Done(paths environment.Paths) (bool, error) {
	f, err := os.Stat(paths.DefaultIndexPath())
	if err == nil {
		return f.IsDir(), nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Migrate removes the index directory and then clones krew-index to the new default index path.
func Migrate(paths environment.Paths) error {
	isMigrated, err := Done(paths)
	if err != nil {
		return err
	}
	if isMigrated {
		klog.Infoln("Already migrated")
		return nil
	}

	err := os.RemoveAll(paths.IndexPath())
	if err != nil {
		return err
	}
	err = gitutil.EnsureCloned(constants.IndexURI, paths.DefaultIndexPath())
	if err != nil {
		return err
	}
	return nil
}
