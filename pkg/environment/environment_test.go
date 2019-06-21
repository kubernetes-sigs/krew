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

package environment

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/client-go/util/homedir"

	"sigs.k8s.io/krew/pkg/testutil"
)

func TestMustGetKrewPaths_resolvesToHomeDir(t *testing.T) {
	home := homedir.HomeDir()
	expectedBase := filepath.Join(home, ".krew")
	p := MustGetKrewPaths()
	if got := p.BasePath(); got != expectedBase {
		t.Fatalf("MustGetKrewPaths()=%s; expected=%s", got, expectedBase)
	}
}

func TestMustGetKrewPaths_envOverride(t *testing.T) {
	custom := filepath.FromSlash("/custom/krew/path")
	os.Setenv("KREW_ROOT", custom)
	defer os.Unsetenv("KREW_ROOT")

	p := MustGetKrewPaths()
	if expected, got := custom, p.BasePath(); got != expected {
		t.Fatalf("MustGetKrewPaths()=%s; expected=%s", got, expected)
	}
}

func TestPaths(t *testing.T) {
	base := filepath.FromSlash("/foo")
	p := newPaths(base)
	if got := p.BasePath(); got != base {
		t.Fatalf("BasePath()=%s; expected=%s", got, base)
	}
	if got, expected := p.BinPath(), filepath.FromSlash("/foo/bin"); got != expected {
		t.Fatalf("BinPath()=%s; expected=%s", got, expected)
	}
	if got, expected := p.IndexPath(), filepath.FromSlash("/foo/index"); got != expected {
		t.Fatalf("IndexPath()=%s; expected=%s", got, expected)
	}
	if got, expected := p.IndexPluginsPath(), filepath.FromSlash("/foo/index/plugins"); got != expected {
		t.Fatalf("IndexPluginsPath()=%s; expected=%s", got, expected)
	}
	if got, expected := p.InstallPath(), filepath.FromSlash("/foo/store"); got != expected {
		t.Fatalf("InstallPath()=%s; expected=%s", got, expected)
	}
	if got, expected := p.PluginInstallPath("my-plugin"), filepath.FromSlash("/foo/store/my-plugin"); got != expected {
		t.Fatalf("PluginInstallPath()=%s; expected=%s", got, expected)
	}
	if got, expected := p.PluginVersionInstallPath("my-plugin", "v1"), filepath.FromSlash("/foo/store/my-plugin/v1"); got != expected {
		t.Fatalf("PluginVersionInstallPath()=%s; expected=%s", got, expected)
	}
	if got := p.DownloadPath(); !strings.HasSuffix(got, "krew-downloads") {
		t.Fatalf("DownloadPath()=%s; expected suffix 'krew-downloads'", got)
	}
	if got := p.ReceiptsPath(); !strings.HasSuffix(got, filepath.FromSlash("receipts/krew-index")) {
		t.Fatalf("ReceiptsPath()=%s; expected suffix 'receipts/krew-index'", got)
	}
	if got := p.PluginReceiptPath("my-plugin"); !strings.HasSuffix(got, filepath.FromSlash("receipts/krew-index/my-plugin.yaml")) {
		t.Fatalf("PluginReceiptPath()=%s; expected suffix 'receipts/krew-index/my-plugin.yaml'", got)
	}
}

func TestGetExecutedVersion(t *testing.T) {
	type args struct {
		paths         Paths
		executionPath string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		inPath  bool
		wantErr bool
	}{
		{
			name: "is in krew path",
			args: args{
				paths:         newPaths(filepath.FromSlash("/plugins")),
				executionPath: filepath.FromSlash("/plugins/store/krew/deadbeef/krew.exe"),
			},
			want:    "deadbeef",
			inPath:  true,
			wantErr: false,
		},
		{
			name: "is not in krew path",
			args: args{
				paths:         newPaths(filepath.FromSlash("/plugins")),
				executionPath: filepath.FromSlash("/plugins/store/NOTKREW/deadbeef/krew.exe"),
			},
			want:    "",
			inPath:  false,
			wantErr: false,
		},
		{
			name: "is in longer krew path",
			args: args{
				paths:         newPaths(filepath.FromSlash("/plugins")),
				executionPath: filepath.FromSlash("/plugins/store/krew/deadbeef/foo/krew.exe"),
			},
			want:    "deadbeef",
			inPath:  true,
			wantErr: false,
		},
		{
			name: "is in smaller krew path",
			args: args{
				paths:         newPaths(filepath.FromSlash("/plugins")),
				executionPath: filepath.FromSlash("/krew.exe"),
			},
			want:    "",
			inPath:  false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("installpath=%s", tt.args.paths.InstallPath())
			got, isVersion, err := GetExecutedVersion(tt.args.paths.InstallPath(), tt.args.executionPath, func(s string) (string, error) {
				return s, nil
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetExecutedVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetExecutedVersion() got = %v, want %v", got, tt.want)
			}
			if isVersion != tt.inPath {
				t.Errorf("GetExecutedVersion() isVersion = %v, want %v", isVersion, tt.inPath)
			}
		})
	}
}

func TestRealpath(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	// create regular file
	tmpDir.Write("regular-file", nil)

	// create absolute symlink
	orig := filepath.Clean(os.TempDir())
	if err := os.Symlink(orig, filepath.Join(tmpDir.Root(), "symbolic-link-abs")); err != nil {
		t.Fatal(err)
	}

	// create relative symlink
	if err := os.Symlink("./another-file", filepath.Join(tmpDir.Root(), "symbolic-link-rel")); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"file not exists", tmpDir.Path("/not/exists"), "", true},
		{"directory", tmpDir.Root(), tmpDir.Root(), false},
		{"regular file", tmpDir.Path("regular-file"), tmpDir.Path("regular-file"), false},
		{"directory unclean", tmpDir.Path("foo/.."), tmpDir.Root(), false},
		{"regular file unclean", tmpDir.Path("regular-file/foo/.."), tmpDir.Path("regular-file"), false},
		{"relative symbolic link", tmpDir.Path("symbolic-link-rel"), "", true},
		{"absolute symbolic link", tmpDir.Path("symbolic-link-abs"), orig, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Realpath(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("Realpath(%s) error = %v, wantErr %v", tt.in, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Realpath(%s) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
