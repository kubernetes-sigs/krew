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
	"strconv"
	"strings"
	"testing"
)

func initRemoteRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if _, err := Exec(dir, "init", "-b", "master"); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	if _, err := Exec(dir, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config email failed: %v", err)
	}
	if _, err := Exec(dir, "config", "user.name", "test"); err != nil {
		t.Fatalf("git config name failed: %v", err)
	}
	commitFile(t, dir, "first.txt", "first")
	return dir
}

func commitFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("write file failed: %v", err)
	}
	if _, err := Exec(dir, "add", name); err != nil {
		t.Fatalf("git add failed: %v", err)
	}
	if _, err := Exec(dir, "commit", "-m", "add "+name); err != nil {
		t.Fatalf("git commit failed: %v", err)
	}
}

func remoteURL(dir string) string {
	return "file://" + dir
}

func commitCount(t *testing.T, dir string) int {
	t.Helper()
	out, err := Exec(dir, "rev-list", "--count", "HEAD")
	if err != nil {
		t.Fatalf("git rev-list failed: %v", err)
	}
	n, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		t.Fatalf("could not parse commit count %q: %v", out, err)
	}
	return n
}

func TestEnsureClonedIsShallow(t *testing.T) {
	remote := initRemoteRepo(t)
	commitFile(t, remote, "second.txt", "second")
	if got := commitCount(t, remote); got != 2 {
		t.Fatalf("remote should have 2 commits, got %d", got)
	}

	dest := filepath.Join(t.TempDir(), "clone")
	if err := EnsureCloned(remoteURL(remote), dest); err != nil {
		t.Fatalf("EnsureCloned failed: %v", err)
	}

	ok, err := IsGitCloned(dest)
	if err != nil || !ok {
		t.Fatalf("expected destination to be a git clone, ok=%v err=%v", ok, err)
	}
	if got := commitCount(t, dest); got != 1 {
		t.Fatalf("expected a shallow clone with 1 commit, got %d", got)
	}
}

func TestEnsureClonedNoopWhenAlreadyCloned(t *testing.T) {
	remote := initRemoteRepo(t)
	dest := filepath.Join(t.TempDir(), "clone")
	if err := EnsureCloned(remoteURL(remote), dest); err != nil {
		t.Fatalf("first EnsureCloned failed: %v", err)
	}
	if err := EnsureCloned(remoteURL(remote), dest); err != nil {
		t.Fatalf("second EnsureCloned should be a no-op, got error: %v", err)
	}
}

func TestEnsureUpdatedFetchesNewCommitsShallow(t *testing.T) {
	remote := initRemoteRepo(t)
	dest := filepath.Join(t.TempDir(), "clone")
	if err := EnsureCloned(remoteURL(remote), dest); err != nil {
		t.Fatalf("EnsureCloned failed: %v", err)
	}

	commitFile(t, remote, "second.txt", "second")
	commitFile(t, remote, "third.txt", "third")

	if err := EnsureUpdated(remoteURL(remote), dest); err != nil {
		t.Fatalf("EnsureUpdated failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dest, "third.txt")); err != nil {
		t.Fatalf("expected updated clone to contain third.txt: %v", err)
	}
	if got := commitCount(t, dest); got != 1 {
		t.Fatalf("expected shallow clone to still have 1 reachable commit after update, got %d", got)
	}
}

func TestEnsureUpdatedRemovesUntrackedFiles(t *testing.T) {
	remote := initRemoteRepo(t)
	dest := filepath.Join(t.TempDir(), "clone")
	if err := EnsureCloned(remoteURL(remote), dest); err != nil {
		t.Fatalf("EnsureCloned failed: %v", err)
	}

	stray := filepath.Join(dest, "untracked.txt")
	if err := os.WriteFile(stray, []byte("stray"), 0o600); err != nil {
		t.Fatalf("write untracked file failed: %v", err)
	}

	if err := EnsureUpdated(remoteURL(remote), dest); err != nil {
		t.Fatalf("EnsureUpdated failed: %v", err)
	}

	if _, err := os.Stat(stray); !os.IsNotExist(err) {
		t.Fatalf("expected untracked file to be cleaned, stat err=%v", err)
	}
}

func TestEnsureClonedIgnoresGlobalDefaultRemoteName(t *testing.T) {
	globalConfig := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(globalConfig, []byte("[clone]\n\tdefaultRemoteName = notorigin\n"), 0o600); err != nil {
		t.Fatalf("write global gitconfig failed: %v", err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", globalConfig)

	remote := initRemoteRepo(t)
	dest := filepath.Join(t.TempDir(), "clone")
	if err := EnsureCloned(remoteURL(remote), dest); err != nil {
		t.Fatalf("EnsureCloned failed: %v", err)
	}

	if _, err := GetRemoteURL(dest); err != nil {
		t.Fatalf("expected the clone to have a remote named origin, GetRemoteURL failed: %v", err)
	}
}

func TestGetRemoteURL(t *testing.T) {
	remote := initRemoteRepo(t)
	dest := filepath.Join(t.TempDir(), "clone")
	if err := EnsureCloned(remoteURL(remote), dest); err != nil {
		t.Fatalf("EnsureCloned failed: %v", err)
	}

	url, err := GetRemoteURL(dest)
	if err != nil {
		t.Fatalf("GetRemoteURL failed: %v", err)
	}
	if want := remoteURL(remote); url != want {
		t.Fatalf("expected remote url %q, got %q", want, url)
	}
}
