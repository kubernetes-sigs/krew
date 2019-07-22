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

	"sigs.k8s.io/krew/pkg/index"
	"sigs.k8s.io/krew/pkg/testutil"
)

func Test_getDownloadTarget(t *testing.T) {
	tests := []struct {
		name          string
		plugin        index.Plugin
		wantVersion   string
		wantSHA256Sum string
		wantURI       string
		wantFos       []index.FileOperation
		wantBin       string
		wantErr       bool
	}{
		{
			name: "matches to a platform in the list",
			plugin: testutil.NewPlugin().
				WithVersion("v1.0.1").WithPlatforms(
				testutil.NewPlatform().WithOS("none").V(),
				testutil.NewPlatform().WithBin("kubectl-foo").
					WithOS(runtime.GOOS).
					WithFiles([]index.FileOperation{{From: "a", To: "b"}}).
					WithSHA256("f0f0f0").
					WithURI("http://localhost").V()).V(),
			wantVersion:   "v1.0.1",
			wantSHA256Sum: "f0f0f0",
			wantURI:       "http://localhost",
			wantFos:       []index.FileOperation{{From: "a", To: "b"}},
			wantBin:       "kubectl-foo",
			wantErr:       false,
		},
		{
			name: "does not match to a platform",
			plugin: testutil.NewPlugin().
				WithVersion("v1.0.1").
				WithPlatforms(
					testutil.NewPlatform().WithOS("foo").V(),
					testutil.NewPlatform().WithOS("bar").V(),
				).V(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVersion, gotSHA256Sum, gotURI, gotFos, bin, err := getDownloadTarget(tt.plugin)
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
