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

// todo(corneliusweig) remove migration code with v0.4
package receiptsmigration

import (
	"os"
	"testing"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/testutil"
)

func TestIsMigrated(t *testing.T) {
	tests := []struct {
		name         string
		filesPresent []string
		expected     bool
	}{
		{
			name:         "One plugin and receipts",
			filesPresent: []string{"bin/foo", "receipts/present"},
			expected:     true,
		},
		{
			name:     "When nothing is installed",
			expected: true,
		},
		{
			name:         "When a plugin is installed but no receipts",
			filesPresent: []string{"bin/foo"},
			expected:     false,
		},
		{
			name:         "When no plugin is installed but a receipt exists",
			filesPresent: []string{"receipts/present"},
			expected:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpDir := testutil.NewTempDir(t)

			newPaths := environment.NewPaths(tmpDir.Root())

			_ = os.MkdirAll(tmpDir.Path("receipts"), os.ModePerm)
			_ = os.MkdirAll(tmpDir.Path("bin"), os.ModePerm)
			for _, name := range test.filesPresent {
				touch(tmpDir, name)
			}

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

// touch creates a file without content in the temporary directory.
func touch(td *testutil.TempDir, file string) {
	td.Write(file, nil)
}
