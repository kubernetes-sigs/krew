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

package indexscanner

import (
	"os"
	"path/filepath"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func Test_readIndexFile(t *testing.T) {
	type args struct {
		indexFilePath string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		matchFirst labels.Set
	}{
		{
			name: "read index file",
			args: args{
				indexFilePath: filepath.Join(testdataPath(t), "testindex", "plugins", "foo.yaml"),
			},
			wantErr: false,
			matchFirst: labels.Set{
				"os": "macos",
			},
		},
		/*
			{
				name: "read index file with unknown keys",
				args: args{
					indexFilePath: filepath.Join(testdataPath(t), "testindex", "plugins", "badplugin.yaml"),
				},
				wantErr: true,
			},
			{
				name: "read index file with unknown keys",
				args: args{
					indexFilePath: filepath.Join(testdataPath(t), "testindex", "plugins", "badplugin2.yaml"),
				},
				wantErr: true,
			},
		*/
	}
	neverMatch := labels.Set{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadPluginFile(tt.args.indexFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("readIndexFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.Name != "foo" && got.Kind != "Plugin" {
				t.Errorf("readIndexFile() has not parsed the metainformations %v", got)
				return
			}

			sel, err := metav1.LabelSelectorAsSelector(got.Spec.Platforms[0].Selector)
			if err != nil {
				t.Errorf("readIndexFile() error parsing label err: %v", err)
				return
			}
			if !sel.Matches(tt.matchFirst) || sel.Matches(neverMatch) {
				t.Errorf("readIndexFile() didn't parse label selector properly: %##v", sel)
				return
			}
		})
	}
}

func TestLoadIndexListFromFS(t *testing.T) {
	type args struct {
		indexDir string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "load index dir",
			args: args{
				indexDir: filepath.Join(testdataPath(t), "testindex"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadPluginListFromFS(tt.args.indexDir)
			if err != nil {
				t.Errorf("LoadPluginListFromFS() error = %v)", err)
				return
			}
			if len(got) != 2 {
				t.Errorf("LoadPluginListFromFS() didn't read enough index files, got %d)", len(got))
				return
			}
		})
	}
}

func TestLoadIndexFileFromFS(t *testing.T) {
	type args struct {
		indexDir   string
		pluginName string
	}
	tests := []struct {
		name              string
		args              args
		wantErr           bool
		wantIsNotExistErr bool
	}{
		{
			name: "load single index file",
			args: args{
				indexDir:   filepath.Join(testdataPath(t), "testindex"),
				pluginName: "foo",
			},
			wantErr:           false,
			wantIsNotExistErr: false,
		},
		{
			name: "plugin file not found",
			args: args{
				indexDir:   filepath.FromSlash("./testdata"),
				pluginName: "not",
			},
			wantErr:           true,
			wantIsNotExistErr: true,
		},
		{
			name: "plugin file bad name",
			args: args{
				indexDir:   filepath.FromSlash("./testdata"),
				pluginName: "wrongname",
			},
			wantErr:           true,
			wantIsNotExistErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadPluginFileFromFS(tt.args.indexDir, tt.args.pluginName)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadIndexFileFromFS() got = %##v,error = %v, wantErr %v", got, err, tt.wantErr)
				return
			}
			if os.IsNotExist(err) != tt.wantIsNotExistErr {
				t.Errorf("LoadIndexFileFromFS() got = %##v,error = %v, wantIsNotExistErr %v", got, err, tt.wantErr)
				return
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
