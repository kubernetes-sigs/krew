// Copyright © 2018 Google Inc.
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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGetExecutedVersion(t *testing.T) {
	type args struct {
		paths         KrewPaths
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
				paths: KrewPaths{
					Base:     filepath.FromSlash("/plugins/"),
					Index:    filepath.FromSlash("/plugins/index"),
					Install:  filepath.FromSlash("/plugins/store"),
					Download: filepath.FromSlash("/plugins/download"),
				},
				executionPath: filepath.FromSlash("/plugins/store/krew/deadbeef/krew.exe"),
			},
			want:    "deadbeef",
			inPath:  true,
			wantErr: false,
		},
		{
			name: "is not in krew path",
			args: args{
				paths: KrewPaths{
					Base:     filepath.FromSlash("/plugins/"),
					Index:    filepath.FromSlash("/plugins/index"),
					Install:  filepath.FromSlash("/plugins/store"),
					Download: filepath.FromSlash("/plugins/download"),
				},
				executionPath: filepath.FromSlash("/plugins/store/NOTKREW/deadbeef/krew.exe"),
			},
			want:    "",
			inPath:  false,
			wantErr: false,
		},
		{
			name: "is in longer krew path",
			args: args{
				paths: KrewPaths{
					Base:     filepath.FromSlash("/plugins/"),
					Index:    filepath.FromSlash("/plugins/index"),
					Install:  filepath.FromSlash("/plugins/store"),
					Download: filepath.FromSlash("/plugins/download"),
				},
				executionPath: filepath.FromSlash("/plugins/store/krew/deadbeef/foo/krew.exe"),
			},
			want:    "deadbeef",
			inPath:  true,
			wantErr: false,
		},
		{
			name: "is in smaller krew path",
			args: args{
				paths: KrewPaths{
					Base:     filepath.FromSlash("/plugins/"),
					Index:    filepath.FromSlash("/plugins/index"),
					Install:  filepath.FromSlash("/plugins/store"),
					Download: filepath.FromSlash("/plugins/download"),
				},
				executionPath: filepath.FromSlash("/krew.exe"),
			},
			want:    "",
			inPath:  false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, isVersion, err := GetExecutedVersion(tt.args.paths.Install, tt.args.executionPath, func(s string) (string, error) {
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
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	// create regular file
	if err := ioutil.WriteFile(filepath.Join(tmp, "regular-file"), nil, 0644); err != nil {
		t.Fatal(err)
	}

	// create absolute symlink
	orig := filepath.Clean(os.TempDir())
	if err := os.Symlink(orig, filepath.Join(tmp, "symbolic-link-abs")); err != nil {
		t.Fatal(err)
	}

	// create relative symlink
	if err := os.Symlink("./another-file", filepath.Join(tmp, "symbolic-link-rel")); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"file not exists", filepath.FromSlash("/not/exists"), "", true},
		{"directory", tmp, tmp, false},
		{"regular file", filepath.Join(tmp, "regular-file"), filepath.Join(tmp, "regular-file"), false},
		{"directory unclean", filepath.Join(tmp, "foo", ".."), tmp, false},
		{"regular file unclean", filepath.Join(tmp, "regular-file", "foo", ".."), filepath.Join(tmp, "regular-file"), false},
		{"relative symbolic link", filepath.Join(tmp, "symbolic-link-rel"), "", true},
		{"absolute symbolic link", filepath.Join(tmp, "symbolic-link-abs"), orig, false},
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
