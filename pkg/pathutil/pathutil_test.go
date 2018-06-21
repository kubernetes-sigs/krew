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

package pathutil

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestIsSubPathExtending(t *testing.T) {
	type args struct {
		basePath      string
		extendingPath string
	}
	tests := []struct {
		name            string
		args            args
		wantExtending   []string
		wantIsExtending bool
	}{
		{
			name: "is extending",
			args: args{
				basePath:      filepath.Join("a", "b"),
				extendingPath: filepath.Join("a", "b", "c"),
			},
			wantExtending:   []string{"c"},
			wantIsExtending: true,
		},
		{
			name: "is extending same length",
			args: args{
				basePath:      filepath.Join("a", "b", "c"),
				extendingPath: filepath.Join("a", "b", "c"),
			},
			wantExtending:   []string{},
			wantIsExtending: true,
		},
		{
			name: "is not extending",
			args: args{
				basePath:      filepath.Join("a", "b"),
				extendingPath: filepath.Join("a", "a", "c"),
			},
			wantExtending:   nil,
			wantIsExtending: false,
		},
		{
			name: "is not smaller extending",
			args: args{
				basePath:      filepath.Join("a", "b"),
				extendingPath: filepath.Join("a"),
			},
			wantExtending:   nil,
			wantIsExtending: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExtending, gotIsExtending := IsSubPath(tt.args.basePath, tt.args.extendingPath)
			if !reflect.DeepEqual(gotExtending, tt.wantExtending) {
				t.Errorf("IsSubPath() gotExtending = %v, want %v", gotExtending, tt.wantExtending)
			}
			if gotIsExtending != tt.wantIsExtending {
				t.Errorf("IsSubPath() gotIsExtending = %v, want %v", gotIsExtending, tt.wantIsExtending)
			}
		})
	}
}
