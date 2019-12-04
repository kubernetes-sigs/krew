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
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/krew/pkg/constants"
)

const (
	fooPlugin = "foo"
)

func TestKrewInstall(t *testing.T) {
	skipShort(t)
	test, cleanup := NewTest(t)
	defer cleanup()

	if err := test.Krew("install", validPlugin); err == nil {
		t.Fatal("expected to fail without initializing the index")
	}

	test = test.WithIndex()
	if err := test.Krew("install"); err == nil {
		t.Fatal("expected failure without any args or stdin")
	}

	test.Krew("install", validPlugin).RunOrFailOutput()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
}

func TestKrewInstallReRun(t *testing.T) {
	skipShort(t)
	test, cleanup := NewTest(t)
	defer cleanup()

	test = test.WithIndex()
	test.Krew("install", validPlugin).RunOrFail()
	test.Krew("install", validPlugin).RunOrFail()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
}

func TestKrewInstall_MultiplePositionalArgs(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.WithIndex().Krew("install", validPlugin, validPlugin2).RunOrFailOutput()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
	test.AssertExecutableInPATH("kubectl-" + validPlugin2)
}

func TestKrewInstall_Stdin(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.WithIndex().WithStdin(strings.NewReader(validPlugin + "\n" + validPlugin2)).
		Krew("install").RunOrFailOutput()

	test.AssertExecutableInPATH("kubectl-" + validPlugin)
	test.AssertExecutableInPATH("kubectl-" + validPlugin2)
}

func TestKrewInstall_StdinAndPositionalArguments(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	// when stdin is detected, it's ignored in favor of positional arguments
	test.WithIndex().
		WithStdin(strings.NewReader(validPlugin2)).
		Krew("install", validPlugin).RunOrFail()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
	test.AssertExecutableNotInPATH("kubectl-" + validPlugin2)
}

func TestKrewInstall_Manifest(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.Krew("install",
		"--manifest", filepath.Join("testdata", validPlugin+constants.ManifestExtension)).
		RunOrFail()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
}

func startServer() (*httptest.Server, error) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/konfig_localhost.yaml":
			file, _ := os.Open("./testdata/konfig_localhost.yaml")
			defer file.Close()
			if _, err := io.Copy(w, file); err != nil {
				http.Error(w, err.Error()+" - "+http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		case "/bundle.tar.gz":
			file, _ := os.Open("./testdata/bundle.tar.gz")
			defer file.Close()
			if _, err := io.Copy(w, file); err != nil {
				http.Error(w, err.Error()+" - "+http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		default:
			if _, err := io.WriteString(w, http.StatusText(http.StatusNotFound)); err != nil {
				http.Error(w, err.Error()+" - "+http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			w.WriteHeader(404)
		}
	}))

	listener, err := net.Listen("tcp", LocalhostURL)
	if err != nil {
		return nil, errors.New("Trouble starting local server")
	}
	server.Listener = listener
	server.Start()
	return server, nil
}

func TestKrewInstall_ManifestURL(t *testing.T) {
	skipShort(t)

	server, err := startServer()
	if err != nil {
		t.Errorf("Trouble starting local server")
	}
	defer server.Close()

	test, cleanup := NewTest(t)
	defer cleanup()

	test.Krew("install",
		"--manifest-url", LocalhostManifestURL).RunOrFail()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
}

func TestKrewInstall_ManifestURLAndArchive(t *testing.T) {
	skipShort(t)

	server, err := startServer()
	if err != nil {
		t.Errorf("Trouble starting local server")
	}
	defer server.Close()

	test, cleanup := NewTest(t)
	defer cleanup()

	err = test.Krew("install",
		"--manifest-url", LocalhostManifestURL,
		"--archive", filepath.Join("testdata", fooPlugin+".tar.gz")).Run()
	if err == nil {
		t.Errorf("Expected install to fail but was successful")
	}
}

func TestKrewInstall_ManifestURLAndManifest(t *testing.T) {
	skipShort(t)

	server, err := startServer()
	if err != nil {
		t.Errorf("Trouble starting local server")
	}
	defer server.Close()

	test, cleanup := NewTest(t)
	defer cleanup()

	err = test.Krew("install",
		"--manifest-url", LocalhostManifestURL,
		"--manifest", filepath.Join("testdata", fooPlugin+constants.ManifestExtension)).Run()
	if err == nil {
		t.Errorf("Expected install to fail but was successful")
	}
}

func TestKrewInstall_ManifestAndArchive(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.Krew("install",
		"--manifest", filepath.Join("testdata", fooPlugin+constants.ManifestExtension),
		"--archive", filepath.Join("testdata", fooPlugin+".tar.gz")).
		RunOrFail()
	test.AssertExecutableInPATH("kubectl-" + fooPlugin)
}

func TestKrewInstall_OnlyArchive(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	err := test.Krew("install",
		"--archive", filepath.Join("testdata", fooPlugin+".tar.gz")).
		Run()
	if err == nil {
		t.Errorf("Expected install to fail but was successful")
	}
}

func TestKrewInstall_PositionalArgumentsAndManifest(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	err := test.Krew("install", validPlugin,
		"--manifest", filepath.Join("testdata", fooPlugin+constants.ManifestExtension),
		"--archive", filepath.Join("testdata", fooPlugin+".tar.gz")).
		Run()
	if err == nil {
		t.Fatal("expected failure")
	}
}
