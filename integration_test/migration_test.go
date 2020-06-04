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

	"sigs.k8s.io/krew/pkg/constants"
)

func TestKrewIndexAutoMigration(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.WithEnv(constants.EnableMultiIndexSwitch, 1).WithDefaultIndex()
	prepareOldIndex(test)

	// any command here should cause the index migration to occur
	out, err := test.Krew("index", "list").Run()
	if err != nil {
		t.Errorf("command failed: %v", err)
	}
	if !strings.Contains(string(out), "krew-index.git") {
		t.Error("output should include the default index after migration")
	}
}

func TestKrewMigrationSkippedWithNoCommand(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.WithEnv(constants.EnableMultiIndexSwitch, 1).WithDefaultIndex()
	prepareOldIndex(test)

	out, err := test.Krew().Run()
	if err != nil {
		t.Errorf("command failed: %v", err)
	}
	if strings.Contains(string(out), "Migration completed") {
		t.Error("output should not include the migration message")
	}
}

func TestKrewUnsupportedVersion(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

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

func prepareOldIndex(it *ITest) {
	indexPath := it.TempDir().Path("index/default")
	tmpPath := it.TempDir().Path("tmp_index")
	newPath := it.TempDir().Path("index")
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
