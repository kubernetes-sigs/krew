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
	"encoding/json"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

// FIXME: still necessary?
func TestKrewList(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	initialList := test.WithIndex().Krew("list").RunOrFailOutput()
	initialOut := []byte{'\n'}

	if diff := cmp.Diff(initialList, initialOut); diff != "" {
		t.Fatalf("expected empty output from 'list':\n%s", diff)
	}

	test.Krew("install", validPlugin).RunOrFail()
	expected := []byte(validPlugin + "\n")

	eventualList := test.Krew("list").RunOrFailOutput()
	if diff := cmp.Diff(eventualList, expected); diff != "" {
		t.Fatalf("'list' output doesn't match:\n%s", diff)
	}
	// TODO(ahmetb): install multiple plugins and see if the output is sorted
}

func TestKrewListJSON(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.Krew("install",
		"--manifest", filepath.Join("testdata", "foo.yaml"),
		"--archive", filepath.Join("testdata", "foo.tar.gz")).
		RunOrFail()

	expected := []byte("[\n    {\n        \"Name\": \"foo\",\n        \"Version\": \"v0.1.0\"\n    }\n]\n")

	eventualList := test.WithIndex().Krew("list", "-o", "json").RunOrFailOutput()
	if diff := cmp.Diff(eventualList, expected); diff != "" {
		t.Fatalf("'list' output doesn't match:\n%s", diff)
	}

	var Plugin []struct {
		Name    string
		Version string
	}

	err := json.Unmarshal(eventualList, &Plugin)
	if err != nil || Plugin[0].Name != "foo" || Plugin[0].Version != "v0.1.0" {
		t.Fatalf("Error unmarshaling: %s. Plugin: \"%s\". Version: \"%s\".", err, Plugin[0].Name, Plugin[0].Version)
	}
}

func TestKrewListJSONMulti(t *testing.T) {
	skipShort(t)
	test, cleanup := NewTest(t)
	defer cleanup()

	test.Krew("install", validPlugin).RunOrFail()
	test.Krew("install",
		"--manifest", filepath.Join("testdata", "foo.yaml"),
		"--archive", filepath.Join("testdata", "foo.tar.gz")).
		RunOrFail()

	eventualList := test.Krew("list", "-o", "json").RunOrFailOutput()

	var Plugin []struct {
		Name    string
		Version string
	}
	err := json.Unmarshal(eventualList, &Plugin)
	if err != nil || Plugin[0].Name != "foo" || Plugin[0].Version != "v0.1.0" {
		t.Fatalf("Error unmarshaling: %s. Plugin[0]: \"%s\". Version: \"%s\". Plugin[1]: \"%s\"", err, Plugin[0].Name, Plugin[0].Version, Plugin[1].Name)
	}
}

func TestKrewListJSONEmpty(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	initialList := test.WithIndex().Krew("list", "-o", "json").RunOrFailOutput()
	initialOut := []byte("[]\n")

	if diff := cmp.Diff(initialList, initialOut); diff != "" {
		t.Fatalf("expected empty output from 'list':\n%s", diff)
	}

	var Plugin []struct {
		Name    string
		Version string
	}
	err := json.Unmarshal(initialList, &Plugin)
	if err != nil || len(Plugin) != 0 {
		t.Fatalf("Error unmarshaling empty list: %s. Length: %d", err, len(Plugin))
	}
}

func TestKrewListYAML(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.Krew("install",
		"--manifest", filepath.Join("testdata", "foo.yaml"),
		"--archive", filepath.Join("testdata", "foo.tar.gz")).
		RunOrFail()

	expected := []byte("- Name: foo\n  Version: v0.1.0\n")

	eventualList := test.WithIndex().Krew("list", "-o", "yaml").RunOrFailOutput()
	if diff := cmp.Diff(eventualList, expected); diff != "" {
		t.Fatalf("'list' output doesn't match:\n%s", diff)
	}

	var Plugin []struct {
		Name    string `yaml:"Name"`
		Version string `yaml:"Version"`
	}

	err := yaml.Unmarshal(eventualList, &Plugin)
	if err != nil || Plugin[0].Name != "foo" || Plugin[0].Version != "v0.1.0" {
		t.Fatalf("Error unmarshaling: %s.\nPlugin: \"%s\". Version: \"%s\".", err, Plugin[0].Name, Plugin[0].Version)
	}
}

func TestKrewListYAMLMulti(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.Krew("install", validPlugin).RunOrFail()
	test.Krew("install",
		"--manifest", filepath.Join("testdata", "foo.yaml"),
		"--archive", filepath.Join("testdata", "foo.tar.gz")).
		RunOrFail()

	eventualList := test.WithIndex().Krew("list", "-o", "yaml").RunOrFailOutput()

	var Plugin []struct {
		Name    string `yaml:"Name"`
		Version string `yaml:"Version"`
	}

	err := yaml.Unmarshal(eventualList, &Plugin)
	if err != nil || Plugin[0].Name != "foo" || Plugin[0].Version != "v0.1.0" {
		t.Fatalf("Error unmarshaling: %s.\nPlugin: \"%s\". Version: \"%s\".", err, Plugin[0].Name, Plugin[0].Version)
	}
}

func TestKrewListYAMLEmpty(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	initialList := test.WithIndex().Krew("list", "-o", "yaml").RunOrFailOutput()
	initialOut := []byte("[]\n")

	if diff := cmp.Diff(initialList, initialOut); diff != "" {
		t.Fatalf("expected empty output from 'list':\n%s", diff)
	}

	var Plugin []struct {
		Name    string `yaml:"Name"`
		Version string `yaml:"Version"`
	}

	err := yaml.Unmarshal(initialList, &Plugin)
	if err != nil || len(Plugin) != 0 {
		t.Fatalf("Error unmarshaling empty list: %s. Length: %d.", err, len(Plugin))
	}
}

func TestKrewListWide(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.Krew("install",
		"--manifest", filepath.Join("testdata", "foo.yaml"),
		"--archive", filepath.Join("testdata", "foo.tar.gz")).
		RunOrFail()

	expected := []byte("PLUGIN  VERSION\nfoo     v0.1.0\n")

	eventualList := test.WithIndex().Krew("list", "-o", "wide").RunOrFailOutput()
	if diff := cmp.Diff(eventualList, expected); diff != "" {
		t.Fatalf("'list' output doesn't match:\n%s", diff)
	}
}

func TestKrewListWideMulti(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.Krew("install", validPlugin).RunOrFail()
	test.Krew("install",
		"--manifest", filepath.Join("testdata", "foo.yaml"),
		"--archive", filepath.Join("testdata", "foo.tar.gz")).
		RunOrFail()

	eventualList := test.WithIndex().Krew("list", "-o", "wide").RunOrFailOutput()
	ok, err := regexp.Match("PLUGIN  [ ]*VERSION\nfoo     [ ]*v0.1.0\nkonfig  [ ]*v[0-9]*[.][0-9]*[.][0-9]*\n", eventualList)
	if !ok || err != nil {
		t.Fatalf("'list' output doesn't match; err:\n%s", err)
	}
}

func TestKrewListWideEmpty(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	initialList := test.WithIndex().Krew("list", "-o", "wide").RunOrFailOutput()
	initialOut := []byte("PLUGIN  VERSION\n")

	if diff := cmp.Diff(initialList, initialOut); diff != "" {
		t.Fatalf("expected empty output from 'list':\n%s", diff)
	}
}

func TestKrewListName(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.Krew("install",
		"--manifest", filepath.Join("testdata", "foo.yaml"),
		"--archive", filepath.Join("testdata", "foo.tar.gz")).
		RunOrFail()

	expected := []byte("foo\n")

	eventualList := test.WithIndex().Krew("list", "-o", "name").RunOrFailOutput()
	if diff := cmp.Diff(eventualList, expected); diff != "" {
		t.Fatalf("'list' output doesn't match:\n%s", diff)
	}
}

func TestKrewListNameMulti(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.Krew("install", validPlugin).RunOrFail()
	test.Krew("install",
		"--manifest", filepath.Join("testdata", "foo.yaml"),
		"--archive", filepath.Join("testdata", "foo.tar.gz")).
		RunOrFail()

	eventualList := test.WithIndex().Krew("list", "-o", "name").RunOrFailOutput()
	ok, err := regexp.Match("foo\nkonfig\n", eventualList)
	if !ok || err != nil {
		t.Fatalf("'list' output doesn't match; err:\n%s", err)
	}
}

func TestKrewListNameEmpty(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	initialList := test.WithIndex().Krew("list", "-o", "name").RunOrFailOutput()
	initialOut := []byte("\n")

	if diff := cmp.Diff(initialList, initialOut); diff != "" {
		t.Fatalf("expected empty output from 'list':\n%s", diff)
	}
}
