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
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
	"sigs.k8s.io/krew/pkg/index/indexscanner"
	"sigs.k8s.io/krew/pkg/testutil"
)

const pluginName = "some"

var (
	plugin = index.Plugin{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.CurrentAPIVersion,
			Kind:       constants.PluginKind,
		},
		ObjectMeta: metav1.ObjectMeta{Name: pluginName},
		Spec: index.PluginSpec{
			Version:          "",
			ShortDescription: "short",
			Description:      "",
			Caveats:          "",
			Homepage:         "",
			Platforms: []index.Platform{{
				URI:      "http://example.com",
				Sha256:   "deadbeef",
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

	if err := Store(plugin, dest); err != nil {
		t.Error(err)
	}

	actual, err := indexscanner.LoadPluginFileFromFS(tmpDir.Root(), pluginName)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(&plugin, &actual); diff != "" {
		t.Error(diff)
	}
}
