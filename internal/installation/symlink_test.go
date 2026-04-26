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

//go:build !windows

package installation

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_createSymlink_newLinkToDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	if err := os.Mkdir(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(tmpDir, "link")
	if err := createSymlink(targetDir, linkPath); err != nil {
		t.Fatalf("createSymlink() error = %v", err)
	}

	fi, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("os.Lstat() error = %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error("expected created link to have ModeSymlink set")
	}
}

func Test_createSymlink_pointsToCorrectTarget(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	if err := os.Mkdir(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(tmpDir, "link")
	if err := createSymlink(targetDir, linkPath); err != nil {
		t.Fatalf("createSymlink() error = %v", err)
	}

	got, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("os.Readlink() error = %v", err)
	}
	wantResolved, err := filepath.EvalSymlinks(targetDir)
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks(targetDir) error = %v", err)
	}
	if got != wantResolved {
		t.Errorf("Readlink() = %q; want %q", got, wantResolved)
	}

	resolved, err := filepath.EvalSymlinks(linkPath)
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks() error = %v", err)
	}
	if resolved != wantResolved {
		t.Errorf("EvalSymlinks() = %q; want %q", resolved, wantResolved)
	}
}

func Test_createSymlink_pathMismatchWithPlatformSymlinks(t *testing.T) {
	tmpDir := t.TempDir()

	resolvedTmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if tmpDir == resolvedTmpDir {
		t.Skip("no platform symlinks detected (e.g., macOS /var -> /private/var)")
	}

	targetDir := filepath.Join(tmpDir, "target")
	if err := os.Mkdir(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(tmpDir, "link")
	if err := createSymlink(targetDir, linkPath); err != nil {
		t.Fatalf("createSymlink() error = %v", err)
	}

	got, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("os.Readlink() error = %v", err)
	}

	resolvedTarget, err := filepath.EvalSymlinks(targetDir)
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks() error = %v", err)
	}

	if got != resolvedTarget {
		t.Errorf("Readlink() = %q; want resolved %q", got, resolvedTarget)
	}
}

func Test_createSymlink_newLinkToFile(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "kubectl-foo")
	if err := os.WriteFile(targetFile, []byte("#!/bin/sh\necho foo"), 0o755); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(tmpDir, "link")
	if err := createSymlink(targetFile, linkPath); err != nil {
		t.Fatalf("createSymlink() error = %v", err)
	}

	fi, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("os.Lstat() error = %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error("expected created link to have ModeSymlink set")
	}

	got, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("os.Readlink() error = %v", err)
	}
	wantResolved, err := filepath.EvalSymlinks(targetFile)
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks() error = %v", err)
	}
	if got != wantResolved {
		t.Errorf("Readlink() = %q; want %q", got, wantResolved)
	}
}

func Test_createSymlink_errorOnNonexistentTarget(t *testing.T) {
	tmpDir := t.TempDir()
	nonexistent := filepath.Join(tmpDir, "does-not-exist")
	linkPath := filepath.Join(tmpDir, "link")

	err := createSymlink(nonexistent, linkPath)

	if err != nil {
		t.Fatalf("createSymlink() error = %v; unix symlinks should succeed for nonexistent targets", err)
	}

	fi, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("os.Lstat() error = %v; dangling symlink should still exist", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error("expected dangling link to have ModeSymlink set")
	}

	if _, err := os.Stat(linkPath); !os.IsNotExist(err) {
		t.Errorf("expected Stat to report not-exist for dangling symlink, got err = %v", err)
	}
}
