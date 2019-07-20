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
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/krew/pkg/index"
)

func Test_getDownloadTarget(t *testing.T) {
	matchingPlatform := index.Platform{
		URI:    "https://uri.git",
		Sha256: "deadbeef",
		Selector: &v1.LabelSelector{
			MatchLabels: map[string]string{
				"os": runtime.GOOS,
			},
		},
		Bin:   "kubectl-foo",
		Files: nil,
	}
	type args struct {
		index index.Plugin
	}
	tests := []struct {
		name          string
		args          args
		wantVersion   string
		wantSHA256Sum string
		wantURI       string
		wantFos       []index.FileOperation
		wantBin       string
		wantErr       bool
	}{
		{
			name: "Find Matching Platform",
			args: args{
				index: index.Plugin{
					Spec: index.PluginSpec{
						Version: "v1.0.1",
						Platforms: []index.Platform{
							matchingPlatform,
							{
								URI: "https://wrong.com",
								Selector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"os": "None",
									},
								},
							},
						},
					},
				},
			},
			wantVersion:   "v1.0.1",
			wantSHA256Sum: "deadbeef",
			wantURI:       "https://uri.git",
			wantFos:       nil,
			wantBin:       "kubectl-foo",
			wantErr:       false,
		}, {
			name: "No Matching Platform",
			args: args{
				index: index.Plugin{
					Spec: index.PluginSpec{
						Version: "v1.0.2",
						Platforms: []index.Platform{
							{
								URI: "https://wrong.com",
								Selector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"os": "None",
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVersion, gotSHA256Sum, gotURI, gotFos, bin, err := getDownloadTarget(tt.args.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDownloadTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("getDownloadTarget() gotVersion = %v, want %v", gotVersion, tt.wantVersion)
			}
			if gotSHA256Sum != tt.wantSHA256Sum {
				t.Errorf("getDownloadTarget() gotSHA256Sum = %v, want %v", gotSHA256Sum, tt.wantSHA256Sum)
			}
			if bin != tt.wantBin {
				t.Errorf("getDownloadTarget() bin = %v, want %v", bin, tt.wantBin)
			}
			if gotURI != tt.wantURI {
				t.Errorf("getDownloadTarget() gotURI = %v, want %v", gotURI, tt.wantURI)
			}
			if !reflect.DeepEqual(gotFos, tt.wantFos) {
				t.Errorf("getDownloadTarget() gotFos = %v, want %v", gotFos, tt.wantFos)
			}
		})
	}
}

func testdataPath(t *testing.T) string {
	pwd, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(pwd, "testdata")
}

func Test_pluginVersionFromPath(t *testing.T) {
	type args struct {
		installPath string
		pluginPath  string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "normal version",
			args: args{
				installPath: filepath.FromSlash("install/"),
				pluginPath:  filepath.FromSlash("install/foo/HEAD/kubectl-foo"),
			},
			want:    "HEAD",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pluginVersionFromPath(tt.args.installPath, tt.args.pluginPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("pluginVersionFromPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("pluginVersionFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
