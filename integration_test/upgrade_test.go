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

package integrationtest

import (
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestKrewUpgrade(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	test.WithIndex().
		Krew("install", "--manifest", filepath.Join(cwd, "testdata", "konfig.yaml")).
		RunOrFail()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
	initialHash := hashFile(t, location)

	test.Krew("upgrade").RunOrFail()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
	eventualHash := hashFile(t, location)

	if string(initialHash) == string(eventualHash) {
		t.Errorf("Expecting the plugin file to change but was the same.")
	}
}

func hashFile(t *testing.T, path string) []byte {
	t.Helper()
	hasher := sha256.New()
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(hasher, file); err != nil {
		t.Fatal(err)
	}
	return hasher.Sum(nil)
}
