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
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/pkg/constants"
)

func TestKrewUninstall(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test = test.WithDefaultIndex()

	if _, err := test.Krew("uninstall").Run(); err == nil {
		t.Fatal("expected failure without no arguments")
	}
	if _, err := test.Krew("uninstall", validPlugin).Run(); err == nil {
		t.Fatal("expected failure deleting non-installed plugin")
	}
	test.Krew("install", validPlugin).RunOrFailOutput()
	test.Krew("uninstall", validPlugin).RunOrFailOutput()
	test.AssertExecutableNotInPATH("kubectl-" + validPlugin)

	if _, err := test.Krew("uninstall", validPlugin).Run(); err == nil {
		t.Fatal("expected failure for uninstalled plugin")
	}
}

func TestKrewUninstallPluginsFromCustomIndex(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex().WithCustomIndexFromDefault("foo")
	test.Krew("install", "foo/"+validPlugin).RunOrFail()

	test.Krew("uninstall", validPlugin).RunOrFail()
	test.AssertExecutableNotInPATH("kubectl-" + validPlugin)
}

func TestKrewUninstall_CannotUseIndexSyntax(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex()
	out, err := test.Krew("uninstall", "foo/"+validPlugin).Run()
	if err == nil {
		t.Error("expected error when uninstalling by canonical name")
	}
	if !strings.Contains(string(out), "INDEX/PLUGIN") {
		t.Error("expected warning about using canonical name to be in output")
	}
}

func TestKrewRemove_AliasSupported(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex().Krew("install", validPlugin).RunOrFailOutput()
	test.Krew("remove", validPlugin).RunOrFailOutput()
	test.AssertExecutableNotInPATH("kubectl-" + validPlugin)
}

func TestKrewRemove_ManifestRemovedFromIndex(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test = test.WithDefaultIndex()
	manifestDir := environment.NewPaths(test.Root()).IndexPluginsPath(constants.DefaultIndexName)
	localManifest := filepath.Join(manifestDir, validPlugin+constants.ManifestExtension)
	if _, err := os.Stat(localManifest); err != nil {
		t.Fatalf("could not read local manifest file at %s: %v", localManifest, err)
	}
	test.Krew("install", validPlugin).RunOrFail()
	if err := os.Remove(localManifest); err != nil {
		t.Fatalf("failed to remove local manifest file: %v", err)
	}
	test.Krew("remove", validPlugin).RunOrFail()
}

func TestKrewRemove_Unsafe(t *testing.T) {
	skipShort(t)
	test := NewTest(t)
	test.WithDefaultIndex()
	test.Krew("install", validPlugin).RunOrFailOutput()

	cases := []string{
		`../index/` + validPlugin,
		`..\index\` + validPlugin,
		`../default/` + validPlugin,
		`..\default\` + validPlugin,
		`../receipts/` + validPlugin,
		`..\receipts\` + validPlugin,
		`default/subdir/plugin-name`,
	}
	expectedErr := `not allowed`
	for _, c := range cases {
		b, err := test.Krew("uninstall", c).Run()
		if err == nil {
			t.Fatalf("%q expected failure", c)
		} else if !strings.Contains(string(b), expectedErr) {
			t.Fatalf("%q does not contain err %q: %q", c, expectedErr, string(b))
		}
	}
}
