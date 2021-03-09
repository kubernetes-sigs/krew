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
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/pkg/constants"
)

const (
	fooPlugin = "foo"
)

func TestKrewInstall(t *testing.T) {
	skipShort(t)
	test := NewTest(t)

	if err := test.Krew("install", validPlugin); err == nil {
		t.Fatal("expected to fail without initializing the index")
	}

	test = test.WithDefaultIndex()
	if err := test.Krew("install"); err == nil {
		t.Fatal("expected failure without any args or stdin")
	}

	test.Krew("install", validPlugin).RunOrFailOutput()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
	test.AssertPluginFromIndex(validPlugin, "default")

	receiptPath := environment.NewPaths(test.Root()).PluginInstallReceiptPath(validPlugin)
	if r := test.loadReceipt(receiptPath); r.CreationTimestamp.Time.IsZero() {
		t.Fatal("expected receipt to have a valid creationTimestamp")
	}
}

func TestKrewInstallReRun(t *testing.T) {
	skipShort(t)
	test := NewTest(t)

	test = test.WithDefaultIndex()
	test.Krew("install", validPlugin).RunOrFail()
	test.Krew("install", validPlugin).RunOrFail()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
}

func TestKrewInstallUnsafe(t *testing.T) {
	skipShort(t)
	test := NewTest(t)
	test.WithDefaultIndex()

	cases := []string{
		`../index/` + validPlugin,
		`..\index\` + validPlugin,
		`../default/` + validPlugin,
		`..\default\` + validPlugin,
		`index-name/sub-directory/plugin-name`,
	}

	expectedErr := `not allowed`
	for _, c := range cases {
		b, err := test.Krew("install", c).Run()
		if err == nil {
			t.Fatalf("%q expected failure", c)
		} else if !strings.Contains(string(b), expectedErr) {
			t.Fatalf("%q does not contain err %q: %q", c, expectedErr, string(b))
		}
	}
}

func TestKrewInstall_MultiplePositionalArgs(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex().Krew("install", validPlugin, validPlugin2).RunOrFailOutput()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
	test.AssertExecutableInPATH("kubectl-" + validPlugin2)
}

func TestKrewInstall_Stdin(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex().WithStdin(strings.NewReader(validPlugin + "\n" + validPlugin2)).
		Krew("install").RunOrFailOutput()

	test.AssertExecutableInPATH("kubectl-" + validPlugin)
	test.AssertExecutableInPATH("kubectl-" + validPlugin2)
}

func TestKrewInstall_StdinAndPositionalArguments(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	// when stdin is detected, it's ignored in favor of positional arguments
	test.WithDefaultIndex().
		WithStdin(strings.NewReader(validPlugin2)).
		Krew("install", validPlugin).RunOrFail()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
	test.AssertExecutableNotInPATH("kubectl-" + validPlugin2)
}

func TestKrewInstall_ExplicitDefaultIndex(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.Krew("install", "default/"+validPlugin).RunOrFail()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
	test.AssertPluginFromIndex(validPlugin, "default")
}

func TestKrewInstall_CustomIndex(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex().WithCustomIndexFromDefault("foo")
	test.Krew("install", "foo/"+validPlugin).RunOrFail()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
	test.AssertPluginFromIndex(validPlugin, "foo")

	if _, err := test.Krew("install", "invalid/"+validPlugin2).Run(); err == nil {
		t.Fatal("expected install from invalid index to fail")
	}
	test.AssertExecutableNotInPATH("kubectl-" + validPlugin2)
}

func TestKrewInstallNoSecurityWarningForCustomIndex(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex().WithCustomIndexFromDefault("foo")
	out := string(test.Krew("install", "foo/"+validPlugin).RunOrFailOutput())
	if strings.Contains(out, "Run them at your own risk") {
		t.Errorf("expected install of custom plugin to not show security warning: %v", out)
	}
}

func TestKrewInstall_Manifest(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.Krew("install",
		"--manifest", filepath.Join("testdata", validPlugin+constants.ManifestExtension)).
		RunOrFail()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
	test.AssertPluginFromIndex(validPlugin, "detached")
}

func TestKrewInstall_ManifestURL(t *testing.T) {
	skipShort(t)

	test := NewTest(t)
	srv, shutdown := localTestServer()
	defer shutdown()

	test.Krew("install",
		"--manifest-url", srv+"/"+validPlugin+constants.ManifestExtension).
		RunOrFail()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
	test.AssertPluginFromIndex(validPlugin, "detached")
}

func TestKrewInstall_ManifestAndArchive(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.Krew("install",
		"--manifest", filepath.Join("testdata", fooPlugin+constants.ManifestExtension),
		"--archive", filepath.Join("testdata", fooPlugin+".tar.gz")).
		RunOrFail()
	test.AssertExecutableInPATH("kubectl-" + fooPlugin)
	test.AssertPluginFromIndex(fooPlugin, "detached")
}

func TestKrewInstall_OnlyArchive(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	_, err := test.Krew("install",
		"--archive", filepath.Join("testdata", fooPlugin+".tar.gz")).
		Run()
	if err == nil {
		t.Errorf("Expected install to fail but was successful")
	}
}

func TestKrewInstall_ManifestArgsAreMutuallyExclusive(t *testing.T) {
	skipShort(t)

	test := NewTest(t)
	srv, shutdown := localTestServer()
	defer shutdown()

	if _, err := test.Krew("install",
		"--manifest", filepath.Join("testdata", fooPlugin+constants.ManifestExtension),
		"--manifest-url", srv+"/"+validPlugin+constants.ManifestExtension).
		Run(); err == nil {
		t.Fatal("expected mutually exclusive arguments to cause failure")
	}
}

func TestKrewInstall_NoManifestArgsWhenPositionalArgsSpecified(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	_, err := test.Krew("install", validPlugin,
		"--manifest", filepath.Join("testdata", fooPlugin+constants.ManifestExtension),
		"--archive", filepath.Join("testdata", fooPlugin+".tar.gz")).
		Run()
	if err == nil {
		t.Fatal("expected failure when positional args and --manifest specified")
	}

	_, err = test.Krew("install", validPlugin,
		"--manifest-url", filepath.Join("testdata", fooPlugin+constants.ManifestExtension)).Run()
	if err == nil {
		t.Fatal("expected failure when positional args and --manifest-url specified")
	}
}

func localTestServer() (string, func()) {
	s := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	return s.URL, s.Close
}
