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
)

func TestKrewSystem(t *testing.T) {
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

func prepareOldKrewRoot(test *ITest) {
	// receipts are not present in old krew home
	if err := os.RemoveAll(test.tempDir.Path("receipts")); err != nil {
		test.t.Fatal(err)
	}
}
