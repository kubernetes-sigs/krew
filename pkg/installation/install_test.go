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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"sigs.k8s.io/krew/pkg/environment"
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
	tempDir, err := ioutil.TempDir(os.TempDir(), "krew-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	type args struct {
		binDir string
		binary string
	}
	tests := []struct {
		name       string
		pluginName string
		args       args
		wantErr    bool
	}{
		{
			name:       "normal link",
			pluginName: "foo",
			args: args{
				binDir: tempDir,
				binary: filepath.Join(testdataPath(t), "plugin-foo", "kubectl-foo"),
			},
			wantErr: false,
		},
		{
			name:       "update link",
			pluginName: "foo",
			args: args{
				binDir: tempDir,
				binary: filepath.Join(testdataPath(t), "plugin-foo", "kubectl-foo"),
			},
			wantErr: false,
		},
		{
			name:       "wrong path link",
			pluginName: "foo",
			args: args{
				binDir: tempDir,
				binary: filepath.Join(testdataPath(t), "plugin-foo", "foo", "not-exist"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createOrUpdateLink(tt.args.binDir, tt.args.binary, tt.pluginName); (err != nil) != tt.wantErr {
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

func TestRemove_cantUninstallItself(t *testing.T) {
	envPath := environment.MustGetKrewPaths()
	expectedErrorMessagePart := "not allowed"
	if err := Remove(envPath, "krew"); !strings.Contains(err.Error(), expectedErrorMessagePart) {
		t.Fatalf("wrong error message for 'remove krew' action, expected message contains %q; got %q",
			expectedErrorMessagePart, err.Error())
	}
}

func Test_removeLink_linkExists(t *testing.T) {
	dir, err := ioutil.TempDir("", "removelink-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	link := filepath.Join(dir, "some-symlink")
	if err := os.Symlink(os.TempDir(), link); err != nil {
		t.Fatal(err)
	}

	if err := removeLink(link); err != nil {
		t.Fatalf("removeLink(%s) failed: %+v", link, err)
	}
}

func Test_removeLink_fails(t *testing.T) {
	// create an unreadable directory and trigger "permission denied" error
	dir, err := ioutil.TempDir("", "removelink-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	unreadableDir := filepath.Join(dir, "unreadable")
	if err := os.MkdirAll(unreadableDir, 0); err != nil {
		t.Fatal(err)
	}
	unreadableFile := filepath.Join(unreadableDir, "mysterious-file")

	if err := removeLink(unreadableFile); err == nil {
		t.Fatalf("removeLink(%s) with unreadable file returned err==nil", unreadableFile)
	}
}

func Test_removeLink_regularFileExists(t *testing.T) {
	f, err := ioutil.TempFile("", "some-regular-file")
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

func Test_isWindows(t *testing.T) {
	expected := runtime.GOOS == "windows"
	got := isWindows()
	if expected != got {
		t.Fatalf("isWindows()=%v; expected=%v (on %s)", got, expected, runtime.GOOS)
	}
}

func Test_isWindows_envOverride(t *testing.T) {
	defer os.Unsetenv("KREW_OS")

	os.Setenv("KREW_OS", "windows")
	if !isWindows() {
		t.Fatalf("isWindows()=false when KREW_OS=windows")
	}

	os.Setenv("KREW_OS", "not-windows")
	if isWindows() {
		t.Fatalf("isWindows()=true when KREW_OS != windows")
	}
}
