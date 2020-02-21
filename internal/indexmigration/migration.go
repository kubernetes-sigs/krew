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
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/klog"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/gitutil"
	"sigs.k8s.io/krew/pkg/constants"
)

// Done checks if the krew installation requires a migration to support multiple indexes.
// A migration is necessary when the index directory contains a ".git" directory.
func Done(paths environment.Paths) (bool, error) {
	_, err := os.Stat(filepath.Join(paths.IndexPath(), ".git"))
	if err == nil {
		return false, nil
	} else if os.IsNotExist(err) {
		return true, nil
	}
	return false, err
}

// Migrate removes the index directory and then clones krew-index to the new default index path.
func Migrate(paths environment.Paths) error {
	isMigrated, err := Done(paths)
	if err != nil {
		return errors.Wrap(err, "failed to check if index migration is complete")
	}
	if isMigrated {
		klog.V(2).Infoln("Already migrated")
		return nil
	}

	err = os.RemoveAll(paths.IndexPath())
	if err != nil {
		return errors.Wrapf(err, "could not remove index directory %q", paths.IndexPath())
	}
	err = gitutil.EnsureCloned(constants.IndexURI, filepath.Join(paths.IndexPath(), "default"))
	if err != nil {
		return errors.Wrap(err, "failed to clone index")
	}
	return nil
}
