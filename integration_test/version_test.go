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
	"fmt"
	"regexp"
	"testing"
)

func TestKrewVersion(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	output := test.Krew("version").RunOrFailOutput()

	requiredSubstrings := []string{
		fmt.Sprintf(`BasePath\s+%s`, test.Root()),
		"GitTag",
		"GitCommit",
		`IndexURI\s+https://github.com/kubernetes-sigs/krew-index.git`,
		"IndexPath",
		"InstallPath",
		"BinPath",
	}

	for _, p := range requiredSubstrings {
		if regexp.MustCompile(p).FindSubmatchIndex(output) == nil {
			t.Errorf("Expected to find %q in output of `krew version`", p)
		}
	}
}
