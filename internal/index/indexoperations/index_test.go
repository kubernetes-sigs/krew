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
	"sigs.k8s.io/krew/internal/testutil"
)

func TestListIndexes(t *testing.T) {
	tmpDir := testutil.NewTempDir(t)

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

	paths := environment.NewPaths(tmpDir.Root())
	for _, index := range wantIndexes {
		path := paths.IndexPath(index.Name)
		tmpDir.InitEmptyGitRepo(path, index.URL)
	}

	gotIndexes, err := ListIndexes(paths)
	if err != nil {
		t.Errorf("error listing indexes: %v", err)
	}
	if diff := cmp.Diff(wantIndexes, gotIndexes); diff != "" {
		t.Errorf("output does not match: %s", diff)
	}
}

func TestAddIndexSuccess(t *testing.T) {
	tmpDir := testutil.NewTempDir(t)

	indexName := "foo"
	localRepo := tmpDir.Path("local/" + indexName)
	tmpDir.InitEmptyGitRepo(localRepo, "")

	paths := environment.NewPaths(tmpDir.Root())
	if err := AddIndex(paths, indexName, localRepo); err != nil {
		t.Errorf("error adding index: %v", err)
	}
	gotIndexes, err := ListIndexes(paths)
	if err != nil {
		t.Errorf("error listing indexes: %s", err)
	}
	wantIndexes := []Index{
		{
			Name: indexName,
			URL:  localRepo,
		},
	}
	if diff := cmp.Diff(wantIndexes, gotIndexes); diff != "" {
		t.Errorf("expected index %s in list: %s", indexName, diff)
	}
}

func TestAddIndexFailure(t *testing.T) {
	tmpDir := testutil.NewTempDir(t)

	indexName := "foo"
	paths := environment.NewPaths(tmpDir.Root())
	if err := AddIndex(paths, indexName, tmpDir.Path("invalid/repo")); err == nil {
		t.Error("expected error when adding index with invalid URL")
	}

	localRepo := tmpDir.Path("local/" + indexName)
	tmpDir.InitEmptyGitRepo(tmpDir.Path("index/"+indexName), "")
	tmpDir.InitEmptyGitRepo(localRepo, "")

	if err := AddIndex(paths, indexName, localRepo); err == nil {
		t.Error("expected error when adding an index that already exists")
	}

	if err := AddIndex(paths, "foo/bar", ""); err == nil {
		t.Error("expected error with invalid index name")
	}
}

func TestDeleteIndex(t *testing.T) {
	// root directory does not exist
	if err := DeleteIndex(environment.NewPaths(filepath.FromSlash("/tmp/does-not-exist/foo")), "bar"); err == nil {
		t.Fatal("expected error")
	} else if !os.IsNotExist(err) {
		t.Fatalf("not ENOENT error: %v", err)
	}

	tmpDir := testutil.NewTempDir(t)
	p := environment.NewPaths(tmpDir.Root())

	// index does not exist
	if err := DeleteIndex(p, "unknown-index"); err == nil {
		t.Fatal("expected error")
	} else if !os.IsNotExist(err) {
		t.Fatalf("not ENOENT error: %v", err)
	}

	if err := os.MkdirAll(p.IndexPath("some-index"), 0o755); err != nil {
		t.Fatalf("err creating test index: %v", err)
	}

	if err := DeleteIndex(p, "some-index"); err != nil {
		t.Fatalf("got error while deleting index: %v", err)
	}
}

func TestIsValidIndexName(t *testing.T) {
	tests := []struct {
		name  string
		index string
		want  bool
	}{
		{
			name:  "with space",
			index: "foo bar",
			want:  false,
		},
		{
			name:  "with forward slash",
			index: "foo/bar",
			want:  false,
		},
		{
			name:  "relative path",
			index: "../foo",
			want:  false,
		},
		{
			name:  "with back slash",
			index: "foo\\bar",
			want:  false,
		},
		{
			name:  "with period",
			index: "foo.bar",
			want:  false,
		},
		{
			name:  "valid name",
			index: "foo",
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidIndexName(tt.index); tt.want != got {
				t.Errorf("IsValidIndexName(%s), got = %t, want = %t", tt.index, got, tt.want)
			}
		})
	}
}
