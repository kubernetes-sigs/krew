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

	"k8s.io/client-go/util/homedir"
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

func Test_getKubectlPluginsPath(t *testing.T) {
	type args struct {
		envs []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default path",
			args: args{
				envs: []string{"HOME=~"},
			},
			want: filepath.Join(homedir.HomeDir(), ".kube", "plugins"),
		},
		{
			name: "manual plugin path",
			args: args{
				envs: []string{"KUBECTL_PLUGINS_PATH=/foobar", "HOME=~"},
			},
			want: "/foobar",
		},
		{
			name: "no further xdg",
			args: args{
				envs: []string{"XDG_DATA_DIRS=/", "HOME=~"},
			},
			want: filepath.Join(homedir.HomeDir(), ".kube", "plugins"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getKubectlPluginsPath(tt.args.envs); got != tt.want {
				t.Errorf("getKubectlPluginsPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
