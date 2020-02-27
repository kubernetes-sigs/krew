// Copyright 2020 The Kubernetes Authors.
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
	"testing"

	"github.com/google/go-cmp/cmp"

	"sigs.k8s.io/krew/internal/testutil"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

func testdataPath(t *testing.T) string {
	pwd, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(pwd, "testdata")
}

func TestListInstalledPlugins(t *testing.T) {
	tests := []struct {
		name     string
		plugins  []index.Plugin
		expected map[string]string
	}{
		{
			name:     "single plugin",
			plugins:  []index.Plugin{testutil.NewPlugin().WithName("test").WithVersion("v0.0.1").V()},
			expected: map[string]string{"test": "v0.0.1"},
		},
		{
			name: "multiple plugins",
			plugins: []index.Plugin{
				testutil.NewPlugin().WithName("plugin-a").WithVersion("v0.0.1").V(),
				testutil.NewPlugin().WithName("plugin-b").WithVersion("v0.1.0").V(),
				testutil.NewPlugin().WithName("plugin-c").WithVersion("v1.0.0").V(),
			},
			expected: map[string]string{
				"plugin-a": "v0.0.1",
				"plugin-b": "v0.1.0",
				"plugin-c": "v1.0.0",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tempDir, cleanup := testutil.NewTempDir(t)
			defer cleanup()

			for _, plugin := range test.plugins {
				tempDir.WriteYAML(plugin.Name+constants.ManifestExtension, plugin)
			}

			actual, err := ListInstalledPlugins(tempDir.Root())
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test.expected, actual); diff != "" {
				t.Error(diff)
			}
		})
	}
}
