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

package receipt

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"sigs.k8s.io/krew/internal/index/indexscanner"
	"sigs.k8s.io/krew/internal/testutil"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

func TestStore(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	testPlugin := testutil.NewPlugin().WithName("some-plugin").WithPlatforms(testutil.NewPlatform().V()).V()
	testReceipt := testutil.NewReceipt().WithPlugin(testPlugin).V()
	dest := tmpDir.Path("some-plugin.yaml")

	if err := Store(testReceipt, dest); err != nil {
		t.Fatal(err)
	}

	actual, err := indexscanner.ReadReceiptFromFile(dest)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(&testReceipt, &actual); diff != "" {
		t.Fatal(diff)
	}
}

func TestLoad(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	testPlugin := testutil.NewPlugin().WithName("foo").WithPlatforms(testutil.NewPlatform().V()).V()
	testPluginReceipt := testutil.NewReceipt().WithPlugin(testPlugin).V()
	if err := Store(testPluginReceipt, tmpDir.Path("foo.yaml")); err != nil {
		t.Fatal(err)
	}

	gotPlugin, err := Load(tmpDir.Path("foo.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(&gotPlugin, &testPluginReceipt); diff != "" {
		t.Fatal(diff)
	}
}

func TestLoad_preservesNonExistsError(t *testing.T) {
	_, err := Load("non-existing.yaml")
	if !os.IsNotExist(err) {
		t.Fatalf("returned error is not ENOENT: %+v", err)
	}
}

func TestCanonicalName(t *testing.T) {
	tests := []struct {
		name     string
		receipt  index.Receipt
		expected string
	}{
		{
			name:     "from default index",
			receipt:  testutil.NewReceipt().WithPlugin(testutil.NewPlugin().WithName("foo").V()).V(),
			expected: "foo",
		},
		{
			name: "from custom index",
			receipt: testutil.NewReceipt().WithPlugin(testutil.NewPlugin().WithName("bar").V()).WithStatus(index.ReceiptStatus{
				Source: index.SourceIndex{
					Name: "foo",
				},
			}).V(),
			expected: "foo/bar",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := CanonicalName(test.receipt)
			if diff := cmp.Diff(test.expected, actual); diff != "" {
				t.Fatalf("expected name to match: %s", diff)
			}
		})
	}
}

func TestNew(t *testing.T) {
	testPlugin := testutil.NewPlugin().WithName("foo").WithPlatforms(testutil.NewPlatform().V()).V()
	wantReceipt := testutil.NewReceipt().WithPlugin(testPlugin).V()

	gotReceipt := New(testPlugin, constants.DefaultIndexName)
	if diff := cmp.Diff(gotReceipt, wantReceipt); diff != "" {
		t.Fatalf("expected receipts to match: %s", diff)
	}
}
