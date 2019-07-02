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
	"strings"
	"testing"
)

func TestKrewSearchAll(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	output := test.WithIndex().Krew("search").RunOrFailOutput()
	if plugins := lines(output); len(plugins) < 10 {
		// the first line is the header
		t.Errorf("Expected at least %d plugins", len(plugins)-1)
	}
}

func TestKrewSearchOne(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	plugins := lines(test.WithIndex().Krew("search", "krew").RunOrFailOutput())
	if len(plugins) < 2 {
		t.Errorf("Expected krew to be a valid plugin")
	}
	if !strings.HasPrefix(plugins[1], "krew ") {
		t.Errorf("The first match should be krew")
	}
}
