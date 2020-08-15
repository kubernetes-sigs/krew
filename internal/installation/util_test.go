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

func TestGetInstalledPluginReceipts(t *testing.T) {
	tests := []struct {
		name     string
		receipts []index.Receipt
	}{
		{
			name: "single plugin",
			receipts: []index.Receipt{
				testutil.NewReceipt().WithPlugin(testutil.NewPlugin().WithName("test").WithVersion("v0.0.1").V()).V(),
			},
		},
		{
			name: "multiple plugins",
			receipts: []index.Receipt{
				testutil.NewReceipt().WithPlugin(testutil.NewPlugin().WithName("plugin-a").WithVersion("v0.0.1").V()).V(),
				testutil.NewReceipt().WithPlugin(testutil.NewPlugin().WithName("plugin-b").WithVersion("v0.1.0").V()).V(),
				testutil.NewReceipt().WithPlugin(testutil.NewPlugin().WithName("plugin-c").WithVersion("v1.0.0").V()).V(),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tempDir := testutil.NewTempDir(t)

			for _, plugin := range test.receipts {
				tempDir.WriteYAML(plugin.Name+constants.ManifestExtension, plugin)
			}

			actual, err := GetInstalledPluginReceipts(tempDir.Root())
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test.receipts, actual); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestInstalledPluginsFromIndex(t *testing.T) {
	tempDir := testutil.NewTempDir(t)

	indexA := index.ReceiptStatus{Source: index.SourceIndex{Name: "a"}}
	var indexNone index.ReceiptStatus

	for _, testReceipt := range []index.Receipt{
		testutil.NewReceipt().WithPlugin(testutil.NewPlugin().WithName("a1").V()).WithStatus(indexA).V(),
		testutil.NewReceipt().WithPlugin(testutil.NewPlugin().WithName("a2").V()).WithStatus(indexA).V(),
		testutil.NewReceipt().WithPlugin(testutil.NewPlugin().WithName("default1").V()).WithStatus(indexNone).V(),
	} {
		tempDir.WriteYAML(testReceipt.Name+constants.ManifestExtension, testReceipt)
	}

	v, err := InstalledPluginsFromIndex(tempDir.Root(), "a")
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := 2, len(v); expected != got {
		t.Fatalf("expected %d, got: %d for index 'a'", expected, got)
	}

	v, err = InstalledPluginsFromIndex(tempDir.Root(), constants.DefaultIndexName)
	if err != nil {
		t.Fatal(err)
	}
	if expected, got := 1, len(v); expected != got {
		t.Fatalf("expected %d, got: %d for index 'a'", expected, got)
	}
}
