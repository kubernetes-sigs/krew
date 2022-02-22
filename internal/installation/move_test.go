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
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"sigs.k8s.io/krew/internal/testutil"
	"sigs.k8s.io/krew/pkg/index"
)

func Test_findMoveTargets(t *testing.T) {
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
		}, {
			name: "rename move",
			args: args{
				fromDir: filepath.Join(testdataPath(t), "testdir_A"),
				toDir:   filepath.Join(testdataPath(t), "testdir_B"),
				fo: index.FileOperation{
					From: ".secret",
					To:   "foo",
				},
			},
			want: []move{{
				from: filepath.Join(testdataPath(t), "testdir_A", ".secret"),
				to:   filepath.Join(testdataPath(t), "testdir_B", "foo"),
			}},
			wantErr: false,
		},
		{
			name: "glob not matching any files",
			args: args{
				fromDir: filepath.Join(testdataPath(t), "testdir_A"),
				toDir:   filepath.Join(testdataPath(t), "testdir_B"),
				fo: index.FileOperation{
					From: "./nonexisting-*",
					To:   "unused",
				},
			},
			wantErr: true,
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

func Test_getDirectMove(t *testing.T) {
	type args struct {
		fromDir string
		toDir   string
		fo      index.FileOperation
	}
	tests := []struct {
		name      string
		args      args
		want      move
		wantFound bool
		wantErr   bool
	}{
		{
			name: "do detect single move",
			args: args{
				fromDir: filepath.Join(testdataPath(t), "testdir_A"),
				toDir:   filepath.Join(testdataPath(t), "testdir_B"),
				fo: index.FileOperation{
					From: ".secret",
					To:   "foo",
				},
			},
			want: move{
				from: filepath.Join(testdataPath(t), "testdir_A", ".secret"),
				to:   filepath.Join(testdataPath(t), "testdir_B", "foo"),
			},
			wantFound: true,
		}, {
			name: "don't detect single move",
			args: args{
				fromDir: filepath.Join(testdataPath(t), "testdir_A"),
				toDir:   filepath.Join(testdataPath(t), "testdir_B"),
				fo: index.FileOperation{
					From: "*",
					To:   "foo",
				},
			},
			want:      move{},
			wantFound: false,
		}, {
			name: "do move and default",
			args: args{
				fromDir: filepath.Join(testdataPath(t), "testdir_A"),
				toDir:   filepath.Join(testdataPath(t), "testdir_B"),
				fo: index.FileOperation{
					From: ".secret",
					To:   "",
				},
			},
			want: move{
				from: filepath.Join(testdataPath(t), "testdir_A", ".secret"),
				to:   filepath.Join(testdataPath(t), "testdir_B", ".secret"),
			},
			wantFound: true,
		}, {
			name: "don't move bad path",
			args: args{
				fromDir: filepath.Join(testdataPath(t), "testdir_A"),
				toDir:   filepath.Join(testdataPath(t), "testdir_B"),
				fo: index.FileOperation{
					From: "../../",
					To:   ".",
				},
			},
			want:      move{},
			wantFound: false,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getDirectMove(tt.args.fromDir, tt.args.toDir, tt.args.fo)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDirectMove() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDirectMove() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.wantFound {
				t.Errorf("getDirectMove() got1 = %v, want %v", got1, tt.wantFound)
			}
		})
	}
}

func Test_moveOrCopyDir_canMoveToNonExistingDir(t *testing.T) {
	srcDir := testutil.NewTempDir(t)

	srcDir.Write("some-file", nil)

	dstDir := testutil.NewTempDir(t)

	dst := dstDir.Path("non-existing-dir")

	if err := renameOrCopy(srcDir.Root(), dst); err != nil {
		t.Fatalf("move failed: %+v", err)
	}

	items, err := os.ReadDir(dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		for _, fi := range items {
			t.Logf("found: %s", fi.Name())
		}
		t.Fatalf("expected to find 1 file in target directory, found: %d", len(items))
	}
}

func Test_moveOrCopyDir_removesExistingTarget(t *testing.T) {
	srcDir := testutil.NewTempDir(t)

	srcDir.Write("some-file", nil)

	dstDir := testutil.NewTempDir(t)

	for i := 0; i < 3; i++ { // write some files
		dstDir.Write(fmt.Sprintf("file-%d", i), nil)
	}

	if err := renameOrCopy(srcDir.Root(), dstDir.Root()); err != nil {
		t.Fatalf("move failed: %+v", err)
	}

	items, err := os.ReadDir(dstDir.Root())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		for _, fi := range items {
			t.Logf("found: %s", fi.Name())
		}
		t.Fatalf("expected to find 1 file in target directory, found: %d", len(items))
	}

}
