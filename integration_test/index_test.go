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
	"os"
	"strings"
	"testing"

	"sigs.k8s.io/krew/pkg/constants"
)

func TestKrewIndexAdd(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.WithEnv(constants.EnableMultiIndexSwitch, 1).WithIndex()
	if _, err := test.Krew("index", "add").Run(); err == nil {
		t.Fatal("expected index add with no args to fail")
	}
	if _, err := test.Krew("index", "add", "foo", "https://invalid").Run(); err == nil {
		t.Fatal("expected index add with invalid URL to fail")
	}
	if err := test.Krew("index", "add", "../../usr/bin", constants.IndexURI); err == nil {
		t.Fatal("expected index add with path characters to fail")
	}
	if _, err := test.Krew("index", "add", "foo", test.TempDir().Path("index/"+constants.DefaultIndexName)).Run(); err != nil {
		t.Fatalf("error adding new index: %v", err)
	}
	if _, err := test.Krew("index", "add", "foo", test.TempDir().Path("index/"+constants.DefaultIndexName)).Run(); err == nil {
		t.Fatal("expected adding same index to fail")
	}
}

func TestKrewIndexAddUnsafe(t *testing.T) {
	skipShort(t)
	test, cleanup := NewTest(t)
	defer cleanup()
	test = test.WithEnv(constants.EnableMultiIndexSwitch, 1).WithIndex()

	cases := []string{"a/b", `a\b`, "../a", `..\a`}
	expected := "invalid index name"

	for _, c := range cases {
		b, err := test.Krew("index", "add", c, constants.IndexURI).Run()
		if err == nil {
			t.Fatalf("%q: expected error", c)
		} else if !strings.Contains(string(b), expected) {
			t.Fatalf("%q: output doesn't contain %q: %q", c, expected, string(b))
		}
	}
}

func TestKrewIndexList(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.WithEnv(constants.EnableMultiIndexSwitch, 1).WithIndex()
	out := test.Krew("index", "list").RunOrFailOutput()
	if indexes := lines(out); len(indexes) < 2 {
		// the first line is the header
		t.Fatal("expected at least 1 index in output")
	}

	localIndex := test.TempDir().Path("index/" + constants.DefaultIndexName)
	test.Krew("index", "add", "foo", localIndex).RunOrFail()
	out = test.Krew("index", "list").RunOrFailOutput()
	if indexes := lines(out); len(indexes) < 3 {
		// the first line is the header
		t.Fatal("expected 2 indexes in output")
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
	if indexes := lines(out); len(indexes) > 1 {
		// the first line is the header
		t.Fatalf("expected index list to be empty:\n%s", string(out))
	}
}

func TestKrewIndexRemove_nonExisting(t *testing.T) {
	skipShort(t)
	test, cleanup := NewTest(t)
	test = test.WithEnv(constants.EnableMultiIndexSwitch, 1)
	defer cleanup()

	_, err := test.Krew("index", "remove", "non-existing").Run()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestKrewIndexRemove_ok(t *testing.T) {
	skipShort(t)
	test, cleanup := NewTest(t)
	test = test.WithEnv(constants.EnableMultiIndexSwitch, 1).WithIndex()
	defer cleanup()

	localIndex := test.TempDir().Path("index/" + constants.DefaultIndexName)
	test.Krew("index", "add", "foo", localIndex).RunOrFail()
	test.Krew("index", "remove", "foo").RunOrFail()
}

func TestKrewIndexRemove_unsafe(t *testing.T) {
	skipShort(t)
	test, cleanup := NewTest(t)
	test = test.WithEnv(constants.EnableMultiIndexSwitch, 1).WithIndex()
	defer cleanup()

	expected := "invalid index name"
	cases := []string{"a/b", `a\b`, "../a", `..\a`}
	for _, c := range cases {
		b, err := test.Krew("index", "remove", c).Run()
		if err == nil {
			t.Fatalf("%q: expected error", c)
		} else if !strings.Contains(string(b), expected) {
			t.Fatalf("%q: output doesn't contain %q: %q", c, expected, string(b))
		}
	}
}

func TestKrewIndexRemoveFailsWhenPluginsInstalled(t *testing.T) {
	skipShort(t)
	test, cleanup := NewTest(t)
	test = test.WithEnv(constants.EnableMultiIndexSwitch, 1)
	defer cleanup()

	test.Krew("install", validPlugin).RunOrFailOutput()
	if _, err := test.Krew("index", "remove", "default").Run(); err == nil {
		t.Fatal("expected error while removing index when there are installed plugins")
	}

	// using --force skips the check
	test.Krew("index", "remove", "--force", "default").RunOrFail()
}

func TestKrewIndexRemoveForce_nonExisting(t *testing.T) {
	skipShort(t)
	test, cleanup := NewTest(t)
	test = test.WithEnv(constants.EnableMultiIndexSwitch, 1)
	defer cleanup()

	// --force returns success for non-existing indexes
	test.Krew("index", "remove", "--force", "non-existing").RunOrFail()
}
