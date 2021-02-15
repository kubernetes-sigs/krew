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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/krew/internal/index/indexscanner"
	"sigs.k8s.io/krew/internal/testutil"
	"sigs.k8s.io/krew/pkg/constants"
)

func TestStore(t *testing.T) {
	tmpDir := testutil.NewTempDir(t)

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
	tmpDir := testutil.NewTempDir(t)

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

func TestNew(t *testing.T) {
	timestamp := metav1.Now()
	testPlugin := testutil.NewPlugin().WithName("foo").WithPlatforms(testutil.NewPlatform().V()).V()
	wantReceipt := testutil.NewReceipt().WithPlugin(testPlugin).V()
	wantReceipt.CreationTimestamp = timestamp

	gotReceipt := New(testPlugin, constants.DefaultIndexName, timestamp)
	if diff := cmp.Diff(gotReceipt, wantReceipt); diff != "" {
		t.Fatalf("expected receipts to match: %s", diff)
	}
}
