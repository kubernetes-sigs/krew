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
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/krew/pkg/constants"
)

const (
	fooPlugin = "test-foo"
)

func TestKrewInstall(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.WithIndex().Krew("install", validPlugin).RunOrFailOutput()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
}

func TestKrewInstall_Manifest(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.
		Krew("install",
			"--manifest", filepath.Join("testdata", validPlugin+constants.ManifestExtension)).
		RunOrFail()
	test.AssertExecutableInPATH("kubectl-" + validPlugin)
}

func TestKrewInstall_ManifestAndArchive(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.
		Krew("install",
			"--manifest", filepath.Join("testdata", fooPlugin+constants.ManifestExtension),
			"--archive", filepath.Join("testdata", fooPlugin+".tar.gz")).
		RunOrFail()
	test.AssertExecutableInPATH("kubectl-" + strings.ReplaceAll(fooPlugin, "-", "_"))
}

func TestKrewInstall_OnlyArchiveFails(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	err := test.
		Krew("install",
			"--archive", filepath.Join("testdata", fooPlugin+".tar.gz")).
		Run()
	if err == nil {
		t.Errorf("Expected install to fail but was successful")
	}
}
