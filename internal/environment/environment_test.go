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

	"sigs.k8s.io/krew/internal/testutil"
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
	if got := p.InstallReceiptsPath(); !strings.HasSuffix(got, filepath.FromSlash("receipts")) {
		t.Fatalf("InstallReceiptsPath()=%s; expected suffix 'receipts'", got)
	}
	if got := p.PluginInstallReceiptPath("my-plugin"); !strings.HasSuffix(got, filepath.FromSlash("receipts/my-plugin.yaml")) {
		t.Fatalf("PluginInstallReceiptPath()=%s; expected suffix 'receipts/my-plugin.yaml'", got)
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
