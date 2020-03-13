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
	"bytes"
	"os"
	"testing"

	"sigs.k8s.io/krew/pkg/constants"
)

func TestKrewIndexAdd(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.WithEnv(constants.EnableMultiIndexSwitch, 1).WithIndex()
	if err := test.Krew("index", "add").Run(); err == nil {
		t.Fatal("expected index add with no args to fail")
	}
	if err := test.Krew("index", "add", "foo", "https://invalid").Run(); err == nil {
		t.Fatal("expected index add with invalid URL to fail")
	}
	if err := test.Krew("index", "add", "../../usr/bin", constants.IndexURI); err == nil {
		t.Fatal("expected index add with path characters to fail")
	}
	if err := test.Krew("index", "add", "foo", test.TempDir().Path("index/"+constants.DefaultIndexName)).Run(); err != nil {
		t.Fatalf("error adding new index: %v", err)
	}
	if err := test.Krew("index", "add", "foo", test.TempDir().Path("index/"+constants.DefaultIndexName)).Run(); err == nil {
		t.Fatal("expected adding same index to fail")
	}
}

func TestKrewIndexList(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.WithEnv(constants.EnableMultiIndexSwitch, 1).WithIndex()
	out := test.Krew("index", "list").RunOrFailOutput()
	if !bytes.Contains(out, []byte(constants.DefaultIndexName)) {
		t.Fatalf("expected index 'default' in output:\n%s", string(out))
	}

	test.Krew("index", "add", "foo", test.TempDir().Path("index/"+constants.DefaultIndexName)).RunOrFail()
	out = test.Krew("index", "list").RunOrFailOutput()
	if !bytes.Contains(out, []byte("foo")) {
		t.Fatalf("expected index 'foo' in output:\n%s", string(out))
	}
}

func TestKrewIndexList_NoIndexes(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.WithEnv(constants.EnableMultiIndexSwitch, 1).WithIndex()
	index := test.TempDir().Path("index")
	if err := os.RemoveAll(index); err != nil {
		t.Fatalf("error removing default index: %v", err)
	}

	out := test.Krew("index", "list").RunOrFailOutput()
	if !bytes.Equal(out, []byte("INDEX  URL\n")) {
		t.Fatalf("expected index list to be empty:\n%s", string(out))
	}
}
