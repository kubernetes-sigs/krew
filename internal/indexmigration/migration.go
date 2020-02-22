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
)

// Done checks if the krew installation requires a migration to support multiple indexes.
// A migration is necessary when the index directory contains a ".git" directory.
func Done(paths environment.Paths) (bool, error) {
	_, err := os.Stat(filepath.Join(paths.IndexPath(), ".git"))
	if err != nil && os.IsNotExist(err) {
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
	indexPath := paths.IndexPath()
	tmpPath := filepath.Join(paths.BasePath(), "tmp_index_migration")
	newPath := filepath.Join(paths.IndexPath(), "default")

	if err := os.Rename(indexPath, tmpPath); err != nil {
		return errors.Wrapf(err, "could not move index directory %q to temporary location %q", indexPath, tmpPath)
	}

	if err := os.Mkdir(indexPath, os.ModePerm); err != nil {
		return errors.Wrapf(err, "could create index directory %q", indexPath)
	}

	if err := os.Rename(tmpPath, newPath); err != nil {
		return errors.Wrapf(err, "could not move temporary index directory %q to new location %q", tmpPath, newPath)
	}

	klog.Infof("Migration completed successfully.")
	return nil
}
