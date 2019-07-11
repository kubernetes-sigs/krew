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
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/krew/pkg/testutil"
)

func TestKrewSystem(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.WithIndex().Krew("install", validPlugin).RunOrFail()

	// needs to be after initial installation
	prepareOldKrewRoot(test)

	test.Krew("system", "receipts-upgrade").RunOrFailOutput()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)

	assertReceiptExistsFor(test, validPlugin)
}

func TestKrewSystem_ReceiptForKrew(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	prepareOldKrewRoot(test)
	touch(test.tempDir, "store/krew/ensure-folder-exists")

	test.WithIndex().Krew("system", "receipts-upgrade").RunOrFailOutput()

	assertReceiptExistsFor(test, "krew")
}

func TestKrewSystem_IgnoreAdditionalFolders(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	prepareOldKrewRoot(test)

	touch(test.tempDir, "store/not-a-plugin/ensure-folder-exists")
	out := test.WithIndex().Krew("system", "receipts-upgrade").RunOrFailOutput()

	if !bytes.Contains(out, []byte("Skipping plugin not-a-plugin")) {
		t.Errorf("Expected a message that 'not-a-plugin' is skipped, but output was:")
		t.Log(string(out))
	}
}

func TestKrewSystem_IgnoreUnknownPlugins(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.Krew("install",
		"--manifest", filepath.Join("testdata", "foo.yaml"),
		"--archive", filepath.Join("testdata", "foo.tar.gz")).
		RunOrFail()

	prepareOldKrewRoot(test)

	out := test.WithIndex().Krew("system", "receipts-upgrade").RunOrFailOutput()

	if !bytes.Contains(out, []byte("Skipping plugin foo")) {
		t.Errorf("Expected a message that 'foo' is skipped, but output was:")
		t.Log(string(out))
	}
}

func prepareOldKrewRoot(test *ITest) {
	// receipts are not present in old krew home
	if err := os.RemoveAll(test.tempDir.Path("receipts")); err != nil {
		test.t.Fatal(err)
	}
}

func assertReceiptExistsFor(it *ITest, plugin string) {
	receipt := "receipts/" + plugin + ".yaml"
	_, err := os.Lstat(it.tempDir.Path(receipt))
	if err != nil {
		it.t.Errorf("Expected plugin receipt %q but found none.", receipt)
	}
}

// touch creates a file without content in the temporary directory.
func touch(td *testutil.TempDir, file string) {
	td.Write(file, nil)
}
