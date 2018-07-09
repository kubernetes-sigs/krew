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

package installation

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/GoogleContainerTools/krew/pkg/index"
)

func Test_moveTargets(t *testing.T) {
	type args struct {
		fromDir string
		toDir   string
		fo      index.FileOperation
	}
	tests := []struct {
		name    string
		args    args
		want    []move
		wantErr bool
	}{
		{
			name: "read testdir",
			args: args{
				fromDir: filepath.Join(testdataPath(t), "testdir_A"),
				toDir:   filepath.Join(testdataPath(t), "testdir_B"),
				fo: index.FileOperation{
					From: "*",
					To:   ".",
				},
			},
			want: []move{{
				from: filepath.Join(testdataPath(t), "testdir_A", ".secret"),
				to:   filepath.Join(testdataPath(t), "testdir_B", ".secret"),
			}, {
				from: filepath.Join(testdataPath(t), "testdir_A", "notsecret"),
				to:   filepath.Join(testdataPath(t), "testdir_B", "notsecret"),
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findMoveTargets(tt.args.fromDir, tt.args.toDir, tt.args.fo)
			if (err != nil) != tt.wantErr {
				t.Errorf("moveTargets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("moveTargets() = %v, want %v", got, tt.want)
			}
		})
	}
}
