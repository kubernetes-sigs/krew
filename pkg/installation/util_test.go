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

func Test_getPluginVersion(t *testing.T) {
	wantVersion := "deadbeef"
	wantURI := "https://uri.git"
	platform := index.Platform{
		URI:    "https://uri.git",
		Sha256: "dEaDbEeF",
	}

	gotVersion, gotURI := getPluginVersion(platform)
	if gotVersion != wantVersion {
		t.Errorf("getPluginVersion() gotVersion = %v, want %v", gotVersion, wantVersion)
	}
	if gotURI != wantURI {
		t.Errorf("getPluginVersion() gotURI = %v, want %v", gotURI, wantURI)
	}
}

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
		name        string
		args        args
		wantVersion string
		wantURI     string
		wantFos     []index.FileOperation
		wantBin     string
		wantErr     bool
	}{
		{
			name: "Find Matching Platform",
			args: args{
				index: index.Plugin{
					Spec: index.PluginSpec{
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
			wantVersion: "deadbeef",
			wantURI:     "https://uri.git",
			wantFos:     nil,
			wantBin:     "kubectl-foo",
			wantErr:     false,
		}, {
			name: "No Matching Platform",
			args: args{
				index: index.Plugin{
					Spec: index.PluginSpec{
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
			wantVersion: "",
			wantURI:     "",
			wantFos:     nil,
			wantBin:     "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVersion, gotURI, gotFos, bin, err := getDownloadTarget(tt.args.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDownloadTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("getDownloadTarget() gotVersion = %v, want %v", gotVersion, tt.wantVersion)
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

func Test_findInstalledPluginVersion(t *testing.T) {
	type args struct {
		installPath string
		binDir      string
		pluginName  string
	}
	tests := []struct {
		name          string
		args          args
		wantName      string
		wantInstalled bool
		wantErr       bool
	}{
		{
			name: "Find version",
			args: args{
				installPath: filepath.Join(testdataPath(t), "index"),
				binDir:      filepath.Join(testdataPath(t), "bin"),
				pluginName:  "foo",
			},
			wantName:      "deadbeef",
			wantInstalled: true,
			wantErr:       false,
		}, {
			name: "No installed version",
			args: args{
				installPath: filepath.Join(testdataPath(t), "index"),
				binDir:      filepath.Join(testdataPath(t), "bin"),
				pluginName:  "not-found",
			},
			wantName:      "",
			wantInstalled: false,
			wantErr:       false,
		}, {
			name: "Insecure name",
			args: args{
				installPath: filepath.Join(testdataPath(t), "index"),
				binDir:      filepath.Join(testdataPath(t), "bin"),
				pluginName:  "../foo",
			},
			wantName:      "",
			wantInstalled: false,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotInstalled, err := findInstalledPluginVersion(tt.args.installPath, tt.args.binDir, tt.args.pluginName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getOtherInstalledVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("getOtherInstalledVersion() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotInstalled != tt.wantInstalled {
				t.Errorf("getOtherInstalledVersion() gotInstalled = %v, want %v", gotInstalled, tt.wantInstalled)
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
