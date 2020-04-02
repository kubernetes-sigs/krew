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

package gitutil

import (
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/krew/internal/testutil"
)

func assertRepoIsCloned(t *testing.T, repoPath string) {
	t.Helper()

	ok, err := IsGitCloned(repoPath)

	if err != nil {
		t.Errorf("expected to finish correctly: %s", err)
	}

	if ok != true {
		t.Errorf("expected to return true after cloning the repo")
	}
}

func TestEnsureClonedClonesARepository(t *testing.T) {
	tempDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	httpClonePath := tempDir.Path("krew-from-https")
	localClonePath := tempDir.Path("krew-from-local-path")

	if err := EnsureCloned("https://github.com/kubernetes-sigs/krew.git", httpClonePath); err != nil {
		t.Errorf("http clone expected to finish correctly: %s", err)
	}

	if err := EnsureCloned(httpClonePath, localClonePath); err != nil {
		t.Errorf("folder clone expected to finish correctly: %s", err)
	}

	assertRepoIsCloned(t, httpClonePath)
	assertRepoIsCloned(t, localClonePath)
}

func TestEnsureClonedDoesntCloneAnExistingRepo(t *testing.T) {
	tempDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	if err := EnsureCloned("https://github.com/kubernetes-sigs/krew.git", tempDir.Root()); err != nil {
		t.Errorf("clone expected to finish correctly: %s", err)
	}

	tempDir.Write("file", []byte("create a file to ensure is not overwriten in the next clone"))

	if err := EnsureCloned("https://github.com/kubernetes-sigs/krew.git", tempDir.Root()); err != nil {
		t.Errorf("second clone expected to finish correctly: %s", err)
	}

	_, err := os.Stat(tempDir.Path("file"))

	if err != nil {
		t.Error("expected to dont perform any operation in the second clone")
	}
}

func TestIsGitClonedWhenPathDoesntExists(t *testing.T) {
	tempDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	ok, err := IsGitCloned(tempDir.Path("does-not-exist"))

	if err != nil {
		t.Errorf("expected to finish correctly: %s", err)
	}

	if ok != false {
		t.Errorf("expected to return false on a non existing folder")
	}
}

func TestIsGitClonedWhenIsFalse(t *testing.T) {
	tempDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	ok, err := IsGitCloned(tempDir.Root())

	if err != nil {
		t.Errorf("expected to finish correctly: %s", err)
	}

	if ok != false {
		t.Errorf("expected to return false on a folder that's not a git repo")
	}
}

func TestIsGitClonedWhenIsTrue(t *testing.T) {
	tempDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	if err := os.MkdirAll(tempDir.Path(".git"), os.ModePerm); err != nil {
		t.Fatalf("cannot create directory %q: %s", filepath.Dir(tempDir.Path(".git")), err)
	}

	ok, err := IsGitCloned(tempDir.Root())

	if err != nil {
		t.Errorf("expected to finish correctly: %s", err)
	}

	if ok != true {
		t.Errorf("expected to return true on a git repo")
	}
}

func TestEnsureUpdatedClonesTheRepoIfDoesntExists(t *testing.T) {
	tempDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	if err := EnsureUpdated("https://github.com/kubernetes-sigs/krew.git", tempDir.Root()); err != nil {
		t.Errorf("expected to finish correctly: %s", err)
	}

	assertRepoIsCloned(t, tempDir.Root())
}

func TestEnsureUpdatedUpdatesTheRepo(t *testing.T) {
	tempDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	if err := EnsureCloned("https://github.com/kubernetes-sigs/krew.git", tempDir.Root()); err != nil {
		t.Errorf("first clone expected to finish correctly: %s", err)
	}

	lastCommitID, err := Exec(tempDir.Root(), "rev-parse", "HEAD")

	if err != nil {
		t.Errorf("get last commit expected to finish correctly: %s", err)
	}

	if _, err := Exec(tempDir.Root(), "reset", "--hard", "HEAD~1"); err != nil {
		t.Errorf("commit reset expected to finish correctly: %s", err)
	}

	err = EnsureUpdated("https://github.com/kubernetes-sigs/krew.git", tempDir.Root())

	if err != nil {
		t.Errorf("update expected to finish correctly: %s", err)
	}

	commitIDAfterUpdate, _ := Exec(tempDir.Root(), "rev-parse", "HEAD")

	if lastCommitID != commitIDAfterUpdate {
		t.Errorf("expected to update to latest commit id %s but instead got %s", lastCommitID, commitIDAfterUpdate)
	}
}

func TestEnsureUpdatedRemovesUntrackedFiles(t *testing.T) {
	tempDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	if err := EnsureCloned("https://github.com/kubernetes-sigs/krew.git", tempDir.Root()); err != nil {
		t.Errorf("clone expected to finish correctly: %s", err)
	}

	tempDir.Write("file", []byte("create a file to ensure is removed after EnsureUpdated"))

	if err := EnsureUpdated("https://github.com/kubernetes-sigs/krew.git", tempDir.Root()); err != nil {
		t.Errorf("update expected to finish correctly: %s", err)
	}

	_, err := os.Stat(tempDir.Path("file"))

	if !os.IsNotExist(err) {
		t.Errorf("expected to get the file removed")
	}
}

func TestGetRemoteURL(t *testing.T) {
	tempDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	expectedURL := "https://github.com/kubernetes-sigs/krew.git"

	if _, err := Exec(tempDir.Root(), "init"); err != nil {
		t.Fatalf("error initializing git repo: %s", err)
	}
	if _, err := Exec(tempDir.Root(), "remote", "add", "origin", expectedURL); err != nil {
		t.Fatalf("error setting remote origin: %s", err)
	}

	url, err := GetRemoteURL(tempDir.Root())

	if err != nil {
		t.Errorf("expected to finish correctly: %s", err)
	}

	if url != expectedURL {
		t.Errorf("expected to get %s instead got %s", expectedURL, url)
	}
}
