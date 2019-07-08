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

package integrationtest

import (
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/krew/pkg/constants"
)

func TestKrewUpgrade(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.WithIndex().
		Krew("install", "--manifest", filepath.Join("testdata", validPlugin+constants.ManifestExtension)).
		RunOrFail()
	initialLocation := resolvePluginSymlink(test, validPlugin)

	test.Krew("upgrade").RunOrFail()
	eventualLocation := resolvePluginSymlink(test, validPlugin)

	if initialLocation == eventualLocation {
		t.Errorf("Expecting the plugin path to change but was the same.")
	}
}

func resolvePluginSymlink(test *ITest, plugin string) string {
	test.t.Helper()
	linkToPlugin, err := test.LookupExecutable("kubectl-" + plugin)
	if err != nil {
		test.t.Fatal(err)
	}

	realLocation, err := os.Readlink(linkToPlugin)
	if err != nil {
		test.t.Fatal(err)
	}

	return realLocation
}
