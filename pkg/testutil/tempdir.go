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

package testutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type TempDir struct {
	t    *testing.T
	root string
}

// NewTempDir creates a temporary directory and a cleanup function.
// It is the responsibility of calling code to call cleanup when done.
func NewTempDir(t *testing.T) (tmpDir *TempDir, cleanup func()) {
	t.Helper()
	root, err := ioutil.TempDir("", "krew-test")
	if err != nil {
		t.Fatal(err)
	}

	tmpDir = &TempDir{t: t, root: root}

	return tmpDir, func() {
		if err := os.RemoveAll(root); err != nil {
			t.Logf("warning: failed to remove tempdir %s: %+v", root, err)
		}
	}
}

// Root returns the root of the temporary directory.
func (td *TempDir) Root() string {
	return td.root
}

// Path returns the path to a file in the temp directory.
// The input file is expected to use '/' as directory separator regardless of the host OS.
func (td *TempDir) Path(file string) string {
	pathElems := []string{td.root}
	pathElems = append(pathElems, strings.Split(file, "/")...)
	return filepath.Join(pathElems...)
}

// Write creates a file containing content in the temporary directory.
func (td *TempDir) Write(file string, content []byte) *TempDir {
	td.t.Helper()
	path := td.Path(file)
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		td.t.Fatalf("cannot create directory %q: %s", filepath.Dir(path), err)
	}
	if err := ioutil.WriteFile(path, content, os.ModePerm); err != nil {
		td.t.Fatalf("cannot write to file %q: %s", path, err)
	}
	return td
}
