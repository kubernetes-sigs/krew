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
	"path/filepath"
	"reflect"
	"testing"
)

func Test_parseEnvs(t *testing.T) {
	type args struct {
		environ []string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "normalParseEnvs",
			args: args{
				environ: []string{"TERM=A", "CC=en"},
			},
			want: map[string]string{
				"TERM": "A",
				"CC":   "en",
			},
		}, {
			name: "normalParseEnvs",
			args: args{
				environ: []string{"TERM="},
			},
			want: map[string]string{
				"TERM": "",
			},
		}, {
			name: "normalParseEnvs",
			args: args{
				environ: []string{"FOO=A=B"},
			},
			want: map[string]string{
				"FOO": "A=B",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseEnvs(tt.args.environ); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseEnvs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetExecutedVersion(t *testing.T) {
	type args struct {
		paths   KrewPaths
		cmdArgs []string
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
				cmdArgs: []string{filepath.FromSlash("/plugins/store/krew/deadbeef/krew.exe")},
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
				cmdArgs: []string{filepath.FromSlash("/plugins/store/NOTKREW/deadbeef/krew.exe")},
			},
			want:    "",
			inPath:  false,
			wantErr: false,
		},
		{
			name: "is in longer kreR path",
			args: args{
				paths: KrewPaths{
					Base:     filepath.FromSlash("/plugins/"),
					Index:    filepath.FromSlash("/plugins/index"),
					Install:  filepath.FromSlash("/plugins/store"),
					Download: filepath.FromSlash("/plugins/download"),
				},
				cmdArgs: []string{filepath.FromSlash("/plugins/store/krew/deadbeef/foo/krew.exe")},
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
				cmdArgs: []string{filepath.FromSlash("/krew.exe")},
			},
			want:    "",
			inPath:  false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := GetExecutedVersion(tt.args.paths, tt.args.cmdArgs)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetExecutedVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetExecutedVersion() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.inPath {
				t.Errorf("GetExecutedVersion() got1 = %v, want %v", got1, tt.inPath)
			}
		})
	}
}
