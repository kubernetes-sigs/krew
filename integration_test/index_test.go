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
	"regexp"
	"strings"
	"testing"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/pkg/constants"
)

func TestKrewIndexAdd(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex()
	if _, err := test.Krew("index", "add").Run(); err == nil {
		t.Fatal("expected index add with no args to fail")
	}
	if _, err := test.Krew("index", "add", "foo", "https://invalid").Run(); err == nil {
		t.Fatal("expected index add with invalid URL to fail")
	}
	if err := test.Krew("index", "add", "../../usr/bin", constants.DefaultIndexURI); err == nil {
		t.Fatal("expected index add with path characters to fail")
	}
	index := environment.NewPaths(test.Root()).IndexPath(constants.DefaultIndexName)
	if _, err := test.Krew("index", "add", "foo", index).Run(); err != nil {
		t.Fatalf("error adding new index: %v", err)
	}
	if _, err := test.Krew("index", "add", "foo", index).Run(); err == nil {
		t.Fatal("expected adding same index to fail")
	}
}

func TestKrewIndexAddUnsafe(t *testing.T) {
	skipShort(t)
	test := NewTest(t)
	test.WithDefaultIndex()

	cases := []string{"a/b", `a\b`, "../a", `..\a`}
	expected := "invalid index name"

	for _, c := range cases {
		b, err := test.Krew("index", "add", c, constants.DefaultIndexURI).Run()
		if err == nil {
			t.Fatalf("%q: expected error", c)
		} else if !strings.Contains(string(b), expected) {
			t.Fatalf("%q: output doesn't contain %q: %q", c, expected, string(b))
		}
	}
}

func TestKrewIndexAddShowsSecurityWarning(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex()
	index := environment.NewPaths(test.Root()).IndexPath(constants.DefaultIndexName)
	out := string(test.Krew("index", "add", "foo", index).RunOrFailOutput())
	if !strings.Contains(out, "WARNING: You have added a new index") {
		t.Errorf("expected output to contain warning when adding custom index: %v", out)
	}
}

func TestKrewIndexList(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex()
	out := test.Krew("index", "list").RunOrFailOutput()
	if indexes := lines(out); len(indexes) < 2 {
		// the first line is the header
		t.Fatal("expected at least 1 index in output")
	}

	test.WithCustomIndexFromDefault("foo")
	out = test.Krew("index", "list").RunOrFailOutput()
	if indexes := lines(out); len(indexes) < 3 {
		// the first line is the header
		t.Fatal("expected 2 indexes in output")
	}
}

func TestKrewIndexList_NoIndexes(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex()
	index := environment.NewPaths(test.Root()).IndexBase()
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
	test := NewTest(t)

	_, err := test.Krew("index", "remove", "non-existing").Run()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestKrewIndexRemove_ok(t *testing.T) {
	skipShort(t)
	test := NewTest(t)
	test.WithDefaultIndex().WithCustomIndexFromDefault("foo")

	test.Krew("index", "remove", "foo").RunOrFail()
}

func TestKrewIndexRemove_unsafe(t *testing.T) {
	skipShort(t)
	test := NewTest(t)
	test.WithDefaultIndex()

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
	test := NewTest(t)
	test.WithDefaultIndex()

	test.Krew("install", validPlugin).RunOrFailOutput()
	if _, err := test.Krew("index", "remove", "default").Run(); err == nil {
		t.Fatal("expected error while removing index when there are installed plugins")
	}

	// using --force skips the check
	test.Krew("index", "remove", "--force", "default").RunOrFail()
}

func TestKrewIndexRemoveForce_nonExisting(t *testing.T) {
	skipShort(t)
	test := NewTest(t)

	// --force returns success for non-existing indexes
	test.Krew("index", "remove", "--force", "non-existing").RunOrFail()
}

func TestKrewDefaultIndex_notAutomaticallyAdded(t *testing.T) {
	skipShort(t)
	test := NewTest(t)

	test.Krew("help").RunOrFail()
	out, err := test.Krew("search").Run()
	if err == nil {
		t.Fatalf("search must've failed without any indexes. output=%s", string(out))
	}
	out = test.Krew("index", "list").RunOrFailOutput()
	if len(lines(out)) > 1 {
		t.Fatalf("expected no indexes; got output=%q", string(out))
	}
}

func TestKrewDefaultIndex_AutoAddedOnInstall(t *testing.T) {
	skipShort(t)
	test := NewTest(t)

	test.Krew("install", validPlugin).RunOrFail()
	ensureIndexListHasDefaultIndex(t, string(test.Krew("index", "list").RunOrFailOutput()))
}

func TestKrewDefaultIndex_AutoAddedOnUpdate(t *testing.T) {
	skipShort(t)
	test := NewTest(t)

	test.Krew("update").RunOrFail()
	ensureIndexListHasDefaultIndex(t, string(test.Krew("index", "list").RunOrFailOutput()))
}

func TestKrewDefaultIndex_AutoAddedOnUpgrade(t *testing.T) {
	skipShort(t)
	test := NewTest(t)

	test.Krew("upgrade").RunOrFail()
	ensureIndexListHasDefaultIndex(t, string(test.Krew("index", "list").RunOrFailOutput()))
}

func TestKrewOnlyCustomIndex(t *testing.T) {
	skipShort(t)
	test := NewTest(t)
	out, err := test.Krew("list").Run()
	if err == nil {
		t.Fatalf("list should've failed without default index output=%s", string(out))
	}
	test.Krew("index", "add", "custom-index", constants.DefaultIndexURI).RunOrFail()
	test.Krew("list").RunOrFail()
}

func ensureIndexListHasDefaultIndex(t *testing.T, output string) {
	t.Helper()
	if !regexp.MustCompile(`(?m)^default\b`).MatchString(output) {
		t.Fatalf("index list did not return default index:\n%s", output)
	}
}
