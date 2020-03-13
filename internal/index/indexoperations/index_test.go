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

package indexoperations

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/gitutil"
	"sigs.k8s.io/krew/internal/testutil"
)

func TestListIndexes(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	wantIndexes := []Index{
		{
			Name: "custom",
			URL:  "https://github.com/custom/index.git",
		},
		{
			Name: "default",
			URL:  "https://github.com/default/index.git",
		},
	}

	for _, index := range wantIndexes {
		path := tmpDir.Path("index/" + index.Name)
		initEmptyGitRepo(t, path, index.URL)
	}

	gotIndexes, err := ListIndexes(environment.NewPaths(tmpDir.Root()).IndexBase())
	if err != nil {
		t.Errorf("error listing indexes: %v", err)
	}
	if diff := cmp.Diff(wantIndexes, gotIndexes); diff != "" {
		t.Errorf("output does not match: %s", diff)
	}
}

func TestAddIndexSuccess(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	indexName := "foo"
	localRepo := tmpDir.Path("local/" + indexName)
	initEmptyGitRepo(t, localRepo, "")

	if err := AddIndex(tmpDir.Path("index"), indexName, localRepo); err != nil {
		t.Errorf("error adding index: %v", err)
	}
	if _, err := os.Stat(tmpDir.Path("index/" + indexName)); err != nil {
		t.Errorf("error reading newly added index: %s", err)
	}
}

func TestAddIndexFailure(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	indexName := "foo"
	indexRoot := tmpDir.Path("index")
	if err := AddIndex(indexRoot, indexName, tmpDir.Path("invalid/repo")); err == nil {
		t.Error("expected error when adding index with invalid URL")
	}

	localRepo := tmpDir.Path("local/" + indexName)
	initEmptyGitRepo(t, tmpDir.Path("index/"+indexName), "")
	initEmptyGitRepo(t, localRepo, "")

	if err := AddIndex(indexRoot, indexName, localRepo); err == nil {
		t.Error("expected error when adding an index that already exists")
	}

	if err := AddIndex(indexRoot, "foo/bar", ""); err == nil {
		t.Error("expected error with invalid index name")
	}
}

func TestIsValidIndexName(t *testing.T) {
	if IsValidIndexName(" ") {
		t.Error("name cannot contain spaces")
	}
	if IsValidIndexName("foo/bar") {
		t.Error("name cannot contain forward slashes")
	}
	if IsValidIndexName("foo\\bar") {
		t.Error("name cannot contain back slashes")
	}
	if IsValidIndexName("foo.bar") {
		t.Error("name cannot contain forward periods")
	}
	if !IsValidIndexName("foo") {
		t.Error("expected valid index name: foo")
	}
}

func initEmptyGitRepo(t *testing.T, path, url string) {
	t.Helper()

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		t.Fatalf("cannot create directory %q: %s", filepath.Dir(path), err)
	}
	if _, err := gitutil.Exec(path, "init"); err != nil {
		t.Fatalf("error initializing git repo: %s", err)
	}
	if _, err := gitutil.Exec(path, "remote", "add", "origin", url); err != nil {
		t.Fatalf("error setting remote origin: %s", err)
	}
}
