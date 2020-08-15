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

// todo(corneliusweig) remove this test file with v0.4
package integrationtest

import (
	"os"
	"strings"
	"testing"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/pkg/constants"
)

func TestKrewIndexAutoMigration(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex()
	prepareOldIndexLayout(test)

	// any command here should cause the index migration to occur
	test.Krew("index", "list").RunOrFail()
	if !isIndexMigrated(test) {
		t.Error("index should have been auto-migrated")
	}
}

func TestKrewUnsupportedVersion(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex().Krew("install", validPlugin).RunOrFail()

	// needs to be after initial installation
	prepareOldKrewRoot(test)

	// any command should fail here
	out, err := test.Krew("list").Run()
	if err == nil {
		t.Error("krew should fail when old receipts structure is detected")
	}
	if !strings.Contains(string(out), "Uninstall Krew") {
		t.Errorf("output should contain instructions on upgrading: %s", string(out))
	}
}

func isIndexMigrated(it *ITest) bool {
	indexPath := environment.NewPaths(it.Root()).IndexPath(constants.DefaultIndexName)
	_, err := os.Stat(indexPath)
	return err == nil
}

// TODO remove when testing indexmigration is no longer necessary
func prepareOldIndexLayout(it *ITest) {
	paths := environment.NewPaths(it.Root())
	indexPath := paths.IndexPath(constants.DefaultIndexName)
	tmpPath := it.TempDir().Path("tmp_index")
	newPath := paths.IndexBase()
	if err := os.Rename(indexPath, tmpPath); err != nil {
		it.t.Fatal(err)
	}
	if err := os.Remove(newPath); err != nil {
		it.t.Fatal(err)
	}
	if err := os.Rename(tmpPath, newPath); err != nil {
		it.t.Fatal(err)
	}
}

func prepareOldKrewRoot(test *ITest) {
	// receipts are not present in old krew home
	if err := os.RemoveAll(test.tempDir.Path("receipts")); err != nil {
		test.t.Fatal(err)
	}
}
