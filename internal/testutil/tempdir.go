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

// Package testutil contains test utilities for krew.
package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"

	"sigs.k8s.io/krew/internal/gitutil"
)

type TempDir struct {
	t    *testing.T
	root string
}

// NewTempDir creates a temporary directory which is automatically cleaned up
// when the test exits.
func NewTempDir(t *testing.T) *TempDir {
	t.Helper()
	root := t.TempDir()

	return &TempDir{t: t, root: root}
}

// Root returns the root of the temporary directory.
func (td *TempDir) Root() string {
	return td.root
}

// Path returns the path to a file in the temp directory.
// The input file is expected to use '/' as directory separator regardless of the host OS.
func (td *TempDir) Path(file string) string {
	if strings.HasPrefix(file, td.root) {
		return filepath.FromSlash(file)
	}
	return filepath.Join(td.root, filepath.FromSlash(file))
}

// Write creates a file containing content in the temporary directory.
func (td *TempDir) Write(file string, content []byte) *TempDir {
	td.t.Helper()
	path := td.Path(file)
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		td.t.Fatalf("cannot create directory %q: %s", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, content, os.ModePerm); err != nil {
		td.t.Fatalf("cannot write to file %q: %s", path, err)
	}
	return td
}

func (td *TempDir) WriteYAML(file string, obj interface{}) *TempDir {
	td.t.Helper()
	content, err := yaml.Marshal(obj)
	if err != nil {
		td.t.Fatalf("cannot marshal obj: %s", err)
	}
	return td.Write(file, content)
}

func (td *TempDir) InitEmptyGitRepo(path, url string) {
	td.t.Helper()

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		td.t.Fatalf("cannot create directory %q: %s", filepath.Dir(path), err)
	}
	if _, err := gitutil.Exec(path, "init"); err != nil {
		td.t.Fatalf("error initializing git repo: %s", err)
	}
	if _, err := gitutil.Exec(path, "remote", "add", "origin", url); err != nil {
		td.t.Fatalf("error setting remote origin: %s", err)
	}
}
