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

package installation

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/testutil"
	"sigs.k8s.io/krew/pkg/index"
)

func Test_moveTargets(t *testing.T) {
	type args struct {
		fromDir string
		toDir   string
		fo      index.FileOperation
	}
	tests := []struct {
		name    string
		args    args
		want    []move
		wantErr bool
	}{
		{
			name: "read testdir",
			args: args{
				fromDir: filepath.Join(testdataPath(t), "testdir_A"),
				toDir:   filepath.Join(testdataPath(t), "testdir_B"),
				fo: index.FileOperation{
					From: "*",
					To:   ".",
				},
			},
			want: []move{{
				from: filepath.Join(testdataPath(t), "testdir_A", ".secret"),
				to:   filepath.Join(testdataPath(t), "testdir_B", ".secret"),
			}, {
				from: filepath.Join(testdataPath(t), "testdir_A", "notsecret"),
				to:   filepath.Join(testdataPath(t), "testdir_B", "notsecret"),
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findMoveTargets(tt.args.fromDir, tt.args.toDir, tt.args.fo)
			if (err != nil) != tt.wantErr {
				t.Errorf("moveTargets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("moveTargets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createOrUpdateLink(t *testing.T) {
	tests := []struct {
		name       string
		pluginName string
		binary     string
		wantErr    bool
	}{
		{
			name:       "normal link",
			pluginName: "foo",
			binary:     filepath.Join(testdataPath(t), "plugin-foo", "kubectl-foo"),
			wantErr:    false,
		},
		{
			name:       "update link",
			pluginName: "foo",
			binary:     filepath.Join(testdataPath(t), "plugin-foo", "kubectl-foo"),
			wantErr:    false,
		},
		{
			name:       "wrong path link",
			pluginName: "foo",
			binary:     filepath.Join(testdataPath(t), "plugin-foo", "foo", "not-exist"),
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := testutil.NewTempDir(t)

			if err := createOrUpdateLink(tmpDir.Root(), tt.binary, tt.pluginName); (err != nil) != tt.wantErr {
				t.Errorf("createOrUpdateLink() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_pluginNameToBin(t *testing.T) {
	tests := []struct {
		name      string
		isWindows bool
		want      string
	}{
		{"foo", false, "kubectl-foo"},
		{"foo-bar", false, "kubectl-foo_bar"},
		{"foo", true, "kubectl-foo.exe"},
		{"foo-bar", true, "kubectl-foo_bar.exe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pluginNameToBin(tt.name, tt.isWindows); got != tt.want {
				t.Errorf("pluginNameToBin(%v, %v) = %v; want %v", tt.name, tt.isWindows, got, tt.want)
			}
		})
	}
}

func Test_removeLink_notExists(t *testing.T) {
	if err := removeLink("/non/existing/path"); err != nil {
		t.Fatalf("removeLink failed with non-existing path: %+v", err)
	}
}

func TestUninstall_cantUninstallItself(t *testing.T) {
	tempDir := testutil.NewTempDir(t)
	envPath := environment.NewPaths(tempDir.Root())
	expectedErrorMessagePart := "not allowed"
	if err := Uninstall(envPath, "krew"); !strings.Contains(err.Error(), expectedErrorMessagePart) {
		t.Fatalf("wrong error message for 'uninstall krew' action, expected message contains %q; got %q",
			expectedErrorMessagePart, err.Error())
	}
}

func Test_removeLink_linkExists(t *testing.T) {
	tmpDir := testutil.NewTempDir(t)

	link := filepath.Join(tmpDir.Root(), "some-symlink")
	if err := os.Symlink(os.TempDir(), link); err != nil {
		t.Fatal(err)
	}

	if err := removeLink(link); err != nil {
		t.Fatalf("removeLink(%s) failed: %+v", link, err)
	}
}

func Test_removeLink_fails(t *testing.T) {
	tmpDir := testutil.NewTempDir(t)

	// create an unreadable directory and trigger "permission denied" error
	unreadableDir := filepath.Join(tmpDir.Root(), "unreadable")
	if err := os.MkdirAll(unreadableDir, 0); err != nil {
		t.Fatal(err)
	}
	unreadableFile := tmpDir.Path("unreadable/mysterious-file")

	if err := removeLink(unreadableFile); err == nil {
		t.Fatalf("removeLink(%s) with unreadable file returned err==nil", unreadableFile)
	}
}

func Test_removeLink_regularFileExists(t *testing.T) {
	f, err := os.CreateTemp("", "some-regular-file")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()
	defer os.Remove(path)

	if err := removeLink(path); err == nil {
		t.Fatalf("removeLink(%s) with regular file was expected to fail; got: err=nil", path)
	}
}

func TestIsWindows(t *testing.T) {
	expected := runtime.GOOS == "windows"
	got := IsWindows()
	if expected != got {
		t.Fatalf("IsWindows()=%v; expected=%v (on %s)", got, expected, runtime.GOOS)
	}
}

func TestIsWindows_envOverride(t *testing.T) {
	t.Setenv("KREW_OS", "windows")
	if !IsWindows() {
		t.Fatalf("IsWindows()=false when KREW_OS=windows")
	}

	t.Setenv("KREW_OS", "not-windows")
	if IsWindows() {
		t.Fatalf("IsWindows()=true when KREW_OS != windows")
	}
}

func Test_downloadAndExtract(t *testing.T) {
	tmpDir := testutil.NewTempDir(t)

	// start a local http server to serve the test archive from pkg/download/testdata
	testdataDir := filepath.Join(testdataPath(t), "..", "..", "download", "testdata")
	server := httptest.NewServer(http.FileServer(http.Dir(testdataDir)))
	defer server.Close()

	url := server.URL + "/test-flat-hierarchy.tar.gz"
	checksum := "433b9e0b6cb9f064548f451150799daadcc70a3496953490c5148c8e550d2f4e"

	if err := downloadAndExtract(tmpDir.Root(), url, checksum, "", false, ""); err != nil {
		t.Fatal(err)
	}
	files, err := os.ReadDir(tmpDir.Root())
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no files found in the extract output directory")
	}
}

func Test_downloadAndExtract_fileOverride(t *testing.T) {
	tmpDir := testutil.NewTempDir(t)

	testFile := filepath.Join(testdataPath(t), "..", "..", "download", "testdata", "test-flat-hierarchy.tar.gz")
	checksum := "433b9e0b6cb9f064548f451150799daadcc70a3496953490c5148c8e550d2f4e"

	if err := downloadAndExtract(tmpDir.Root(), "", checksum, testFile, false, ""); err != nil {
		t.Fatal(err)
	}
	files, err := os.ReadDir(tmpDir.Root())
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no files found in the extract output directory")
	}
}

func Test_applyDefaults(t *testing.T) {
	tests := []struct {
		name     string
		platform index.Platform
		expected index.Platform
	}{
		{
			name:     "with files given",
			platform: testutil.NewPlatform().WithFiles([]index.FileOperation{{From: "here", To: "there"}}).V(),
			expected: testutil.NewPlatform().WithFiles([]index.FileOperation{{From: "here", To: "there"}}).V(),
		},
		{
			name:     "with empty files",
			platform: testutil.NewPlatform().WithFiles([]index.FileOperation{}).V(),
			expected: testutil.NewPlatform().WithFiles([]index.FileOperation{}).V(),
		},
		{
			name:     "with unspecified files",
			platform: testutil.NewPlatform().WithFiles(nil).V(),
			expected: testutil.NewPlatform().WithFiles([]index.FileOperation{{From: "*", To: "."}}).V(),
		},
		{
			name:     "no defaults for other fields",
			platform: testutil.NewPlatform().WithBin("").WithOS("").WithSelector(nil).WithSHA256("").WithURI("").V(),
			expected: testutil.NewPlatform().WithBin("").WithOS("").WithSelector(nil).WithSHA256("").WithURI("").V(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			applyDefaults(&test.platform)
			if diff := cmp.Diff(test.platform, test.expected); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_bug_osSymlink_doesNotResolvePlatformSymlinks(t *testing.T) {
	tmpDir := t.TempDir()

	resolvedTmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if tmpDir == resolvedTmpDir {
		t.Skip("no platform symlinks detected (e.g., macOS /var -> /private/var)")
	}

	binaryPath := filepath.Join(tmpDir, "kubectl-foo")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\necho foo"), 0o755); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(resolvedTmpDir, "kubectl-foo-link")
	if err := os.Symlink(binaryPath, linkPath); err != nil {
		t.Fatalf("os.Symlink() error = %v", err)
	}
	defer os.Remove(linkPath)

	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("os.Readlink() error = %v", err)
	}

	resolvedBinary, err := filepath.EvalSymlinks(binaryPath)
	if err != nil {
		t.Fatal(err)
	}

	if target == resolvedBinary {
		t.Fatal("UNEXPECTED: os.Symlink resolved the path; bug may be fixed at the OS level")
	}
	if target != binaryPath {
		t.Fatalf("expected os.Symlink to store unresolved path %q, got %q", binaryPath, target)
	}
}

func Test_createSymlink_exists(t *testing.T) {
	tmpDir := testutil.NewTempDir(t)
	src := filepath.Join(testdataPath(t), "plugin-foo", "kubectl-foo")
	dst := filepath.Join(tmpDir.Root(), "kubectl-test_plugin")

	if err := createSymlink(src, dst); err != nil {
		t.Fatalf("createSymlink() error = %v", err)
	}

	fi, err := os.Lstat(dst)
	if err != nil {
		t.Fatalf("os.Lstat() error = %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error("expected link to have ModeSymlink set")
	}
}

func Test_createOrUpdateLink_usesCreateSymlink(t *testing.T) {
	tmpDir := testutil.NewTempDir(t)
	binary := filepath.Join(testdataPath(t), "plugin-foo", "kubectl-foo")

	if err := createOrUpdateLink(tmpDir.Root(), binary, "test-plugin"); err != nil {
		t.Fatalf("createOrUpdateLink() error = %v", err)
	}

	linkPath := filepath.Join(tmpDir.Root(), pluginNameToBin("test-plugin", IsWindows()))
	fi, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("failed to lstat link: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatal("expected link to have ModeSymlink set")
	}

	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("failed to readlink: %v", err)
	}
	wantResolved, err := filepath.EvalSymlinks(binary)
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks() error = %v", err)
	}
	if target != wantResolved {
		t.Errorf("link target = %q; want %q", target, wantResolved)
	}
}

func Test_createOrUpdateLink_resolvesSymlinks(t *testing.T) {
	tmpDir := testutil.NewTempDir(t)
	resolvedRoot, err := filepath.EvalSymlinks(tmpDir.Root())
	if err != nil {
		t.Fatal(err)
	}
	if tmpDir.Root() == resolvedRoot {
		t.Skip("no platform symlinks detected (e.g., macOS /var -> /private/var)")
	}

	binDir := filepath.Join(resolvedRoot, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	binaryDir := filepath.Join(tmpDir.Root(), "store", "plugin-foo", "v1.0.0")
	if err := os.MkdirAll(binaryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	binaryPath := filepath.Join(binaryDir, "kubectl-foo")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\necho foo"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := createOrUpdateLink(binDir, binaryPath, "foo"); err != nil {
		t.Fatalf("createOrUpdateLink() error = %v", err)
	}

	linkPath := filepath.Join(binDir, pluginNameToBin("foo", IsWindows()))
	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("os.Readlink() error = %v", err)
	}

	resolvedBinary, err := filepath.EvalSymlinks(binaryPath)
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks() error = %v", err)
	}

	if target != resolvedBinary {
		t.Errorf("link target = %q; want resolved %q (unresolved was %q)", target, resolvedBinary, binaryPath)
	}
}

func Test_reproduce_osSymlink_vs_createSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	resolved, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if tmpDir == resolved {
		t.Skip("no platform symlinks detected (e.g., macOS /var -> /private/var)")
	}

	binary := filepath.Join(tmpDir, "kubectl-repro")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\necho repro"), 0o755); err != nil {
		t.Fatal(err)
	}

	resolvedBinary, err := filepath.EvalSymlinks(binary)
	if err != nil {
		t.Fatal(err)
	}

	oldLink := filepath.Join(resolved, "old-link")
	if err := os.Symlink(binary, oldLink); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(oldLink)

	oldTarget, _ := os.Readlink(oldLink)
	if oldTarget == resolvedBinary {
		t.Fatal("UNEXPECTED: os.Symlink resolved the path; cannot reproduce bug")
	}
	if oldTarget != binary {
		t.Fatalf("os.Symlink target = %q; want unresolved %q", oldTarget, binary)
	}

	newLink := filepath.Join(resolved, "new-link")
	if err := createSymlink(binary, newLink); err != nil {
		t.Fatal(err)
	}
	newTarget, _ := os.Readlink(newLink)
	if newTarget != resolvedBinary {
		t.Fatalf("createSymlink target = %q; want resolved %q", newTarget, resolvedBinary)
	}
}

func TestCleanupStaleKrewInstallations(t *testing.T) {
	dir := testutil.NewTempDir(t)

	testFiles := []string{
		"dir1/f1.txt",
		"dir2/f2.txt",
		"dir3/subdir/f3.txt",
		"file1.txt",
		"file2.txt",
	}
	for _, tf := range testFiles {
		dir.Write(filepath.FromSlash(tf), nil)
	}

	err := CleanupStaleKrewInstallations(dir.Root(), "dir2")
	if err != nil {
		t.Fatal(err)
	}

	ls, err := os.ReadDir(dir.Root())
	if err != nil {
		t.Fatal(err)
	}

	got := make([]string, 0, len(ls))
	for _, l := range ls {
		got = append(got, l.Name())
	}

	expected := []string{"dir2", "file1.txt", "file2.txt"}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Fatal(diff)
	}
}
