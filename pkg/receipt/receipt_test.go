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
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
	"sigs.k8s.io/krew/pkg/index/indexscanner"
	"sigs.k8s.io/krew/pkg/testutil"
)

const testPluginName = "some"

var (
	testPlugin = index.Plugin{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.CurrentAPIVersion,
			Kind:       constants.PluginKind,
		},
		ObjectMeta: metav1.ObjectMeta{Name: testPluginName},
		Spec: index.PluginSpec{
			Version:          "v1.0.0",
			ShortDescription: "short",
			Platforms: []index.Platform{{
				URI:      "http://example.com",
				Sha256:   "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
				Selector: nil,
				Files:    []index.FileOperation{{From: "", To: ""}},
				Bin:      "foo",
			}},
		},
	}
)

func TestStore(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	dest := tmpDir.Path("some.yaml")

	if err := Store(testPlugin, dest); err != nil {
		t.Error(err)
	}

	actual, err := indexscanner.LoadPluginFileFromFS(tmpDir.Root(), testPluginName)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(&testPlugin, &actual); diff != "" {
		t.Fatal(diff)
	}
}

func TestLoad(t *testing.T) {
	// TODO(ahmetb): Avoid reading test data from other packages. It would be
	// good to have an in-memory Plugin object (issue#270) that we can Store()
	// first then load here.
	_, err := Load(filepath.Join("..", "..", "integration_test", "testdata", "foo.yaml"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoad_preservesNonExistsError(t *testing.T) {
	_, err := Load(filepath.Join("foo", "non-existing.yaml"))
	if !os.IsNotExist(err) {
		t.Fatalf("returned error is not ENOENT: %+v", err)
	}
}
