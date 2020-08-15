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

package indexmigration

import (
	"os"
	"testing"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/testutil"
)

func TestIsMigrated(t *testing.T) {
	tests := []struct {
		name     string
		dirPath  string
		expected bool
	}{
		{
			name:     "Already migrated",
			dirPath:  "index/default/.git",
			expected: true,
		},
		{
			name:     "Not migrated",
			dirPath:  "index/.git",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpDir := testutil.NewTempDir(t)

			err := os.MkdirAll(tmpDir.Path(test.dirPath), os.ModePerm)
			if err != nil {
				t.Fatal(err)
			}

			newPaths := environment.NewPaths(tmpDir.Root())
			actual, err := Done(newPaths)
			if err != nil {
				t.Fatal(err)
			}
			if actual != test.expected {
				t.Errorf("Expected %v but found %v", test.expected, actual)
			}
		})
	}
}

func TestMigrate(t *testing.T) {
	tmpDir := testutil.NewTempDir(t)

	tmpDir.Write("index/.git", nil)

	newPaths := environment.NewPaths(tmpDir.Root())
	err := Migrate(newPaths)
	if err != nil {
		t.Fatal(err)
	}
	done, err := Done(newPaths)
	if err != nil || !done {
		t.Errorf("expected migration to be done: %s", err)
	}
}
