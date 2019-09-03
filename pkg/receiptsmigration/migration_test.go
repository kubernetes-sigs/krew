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

// todo(corneliusweig) remove migration code with v0.4
package receiptsmigration

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"sigs.k8s.io/krew/pkg/environment"
	"sigs.k8s.io/krew/pkg/receiptsmigration/oldenvironment"
	"sigs.k8s.io/krew/pkg/testutil"
)

func TestIsMigrated(t *testing.T) {
	tests := []struct {
		name         string
		filesPresent []string
		expected     bool
	}{
		{
			name:         "One plugin and receipts",
			filesPresent: []string{"store/konfig/konfig.sh", "receipts/present"},
			expected:     true,
		},
		{
			name:     "When nothing is installed",
			expected: true,
		},
		{
			name:         "When a plugin is installed but no receipts",
			filesPresent: []string{"store/konfig/konfig.sh"},
			expected:     false,
		},
		{
			name:         "When no plugin is installed but a receipt exists",
			filesPresent: []string{"receipts/present"},
			expected:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.NewTempDir(t)
			defer cleanup()

			os.Setenv("KREW_ROOT", tmpDir.Root())
			defer os.Unsetenv("KREW_ROOT")

			newPaths := environment.MustGetKrewPaths()

			_ = os.MkdirAll(tmpDir.Path("receipts"), os.ModePerm)
			_ = os.MkdirAll(tmpDir.Path("store"), os.ModePerm)
			for _, name := range test.filesPresent {
				touch(tmpDir, name)
			}

			actual, err := Done(newPaths)
			if err != nil {
				t.Fatal(err)
			}
			if actual != test.expected {
				t.Errorf("Expected %v but found %v", test.expected, actual)
			}
		})
	}
}

func Test_getPluginsToReinstall(t *testing.T) {
	tests := []struct {
		name         string
		filesPresent []string
		expected     []string
	}{
		{
			name:         "single plugin",
			filesPresent: []string{"store/konfig/konfig.sh", "index/plugins/konfig.yaml"},
			expected:     []string{"konfig"},
		},
		{
			name: "multiple plugins",
			filesPresent: []string{
				"store/konfig/foo", "index/plugins/konfig.yaml",
				"store/plug-in/foo", "index/plugins/plug-in.yaml",
			},
			expected: []string{"konfig", "plug-in"},
		},
		{
			name: "skip unsafe name",
			filesPresent: []string{
				"store/LPT6/foo", "index/plugins/LPT6.yaml",
			},
			expected: []string{},
		},
		{
			name: "skip krew",
			filesPresent: []string{
				"store/krew/foo", "index/plugins/krew.yaml",
			},
			expected: []string{},
		},
		{
			name: "skip if missing in index",
			filesPresent: []string{
				"store/missing/foo",
			},
			expected: []string{},
		},
		{
			name:     "check multiple conditions",
			expected: []string{"konfig", "plug-in"},
			filesPresent: []string{
				"store/konfig/foo", "index/plugins/konfig.yaml",
				"store/plug-in/foo", "index/plugins/plug-in.yaml",
				"store/LPT6/foo", "index/plugins/LPT6.yaml",
				"store/krew/foo", "index/plugins/krew.yaml",
				"store/missing/foo",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.NewTempDir(t)
			defer cleanup()

			os.Setenv("KREW_ROOT", tmpDir.Root())
			defer os.Unsetenv("KREW_ROOT")

			oldPaths := oldenvironment.MustGetKrewPaths()
			newPaths := environment.MustGetKrewPaths()

			for _, name := range test.filesPresent {
				touch(tmpDir, name)
			}

			actual, err := getPluginsToReinstall(oldPaths, newPaths)
			if err != nil {
				t.Fatal(err)
			}
			sort.Strings(actual)
			sort.Strings(test.expected)

			if diff := cmp.Diff(actual, test.expected); diff != "" {
				t.Errorf(diff)
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
	envPath := oldenvironment.MustGetKrewPaths()
	expectedErrorMessagePart := "not allowed"
	if err := uninstall(envPath, "krew"); !strings.Contains(err.Error(), expectedErrorMessagePart) {
		t.Fatalf("wrong error message for 'uninstall krew' action, expected message contains %q; got %q",
			expectedErrorMessagePart, err.Error())
	}
}

func Test_removeLink_linkExists(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	link := filepath.Join(tmpDir.Root(), "some-symlink")
	if err := os.Symlink(os.TempDir(), link); err != nil {
		t.Fatal(err)
	}

	if err := removeLink(link); err != nil {
		t.Fatalf("removeLink(%s) failed: %+v", link, err)
	}
}

func Test_removeLink_fails(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

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

// touch creates a file without content in the temporary directory.
func touch(td *testutil.TempDir, file string) {
	td.Write(file, nil)
}
