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
	"sigs.k8s.io/krew/pkg/constants"
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
	t.Setenv("KREW_ROOT", custom)

	p := MustGetKrewPaths()
	if expected, got := custom, p.BasePath(); got != expected {
		t.Fatalf("MustGetKrewPaths()=%s; expected=%s", got, expected)
	}
}

func TestPaths(t *testing.T) {
	base := filepath.FromSlash("/foo")
	p := NewPaths(base)
	if got := p.BasePath(); got != base {
		t.Errorf("BasePath()=%s; expected=%s", got, base)
	}
	if got, expected := p.BinPath(), filepath.FromSlash("/foo/bin"); got != expected {
		t.Errorf("BinPath()=%s; expected=%s", got, expected)
	}

	if got, expected := p.IndexPath(constants.DefaultIndexName), filepath.FromSlash("/foo/index/default"); got != expected {
		t.Errorf("IndexPath(\"%s\")=%s; expected=%s", constants.DefaultIndexName, got, expected)
	}
	if got, expected := p.IndexPluginsPath(constants.DefaultIndexName), filepath.FromSlash("/foo/index/default/plugins"); got != expected {
		t.Errorf("IndexPluginsPath(\"%s\")=%s; expected=%s", constants.DefaultIndexName, got, expected)
	}

	if got, expected := p.InstallPath(), filepath.FromSlash("/foo/store"); got != expected {
		t.Errorf("InstallPath()=%s; expected=%s", got, expected)
	}
	if got, expected := p.PluginInstallPath("my-plugin"), filepath.FromSlash("/foo/store/my-plugin"); got != expected {
		t.Errorf("PluginInstallPath()=%s; expected=%s", got, expected)
	}
	if got, expected := p.PluginVersionInstallPath("my-plugin", "v1"), filepath.FromSlash("/foo/store/my-plugin/v1"); got != expected {
		t.Errorf("PluginVersionInstallPath()=%s; expected=%s", got, expected)
	}
	if got := p.InstallReceiptsPath(); !strings.HasSuffix(got, filepath.FromSlash("receipts")) {
		t.Errorf("InstallReceiptsPath()=%s; expected suffix 'receipts'", got)
	}
	if got := p.PluginInstallReceiptPath("my-plugin"); !strings.HasSuffix(got, filepath.FromSlash("receipts/my-plugin.yaml")) {
		t.Errorf("PluginInstallReceiptPath()=%s; expected suffix 'receipts/my-plugin.yaml'", got)
	}
}

func TestRealpath(t *testing.T) {
	tmpDir := testutil.NewTempDir(t)

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
