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

	"sigs.k8s.io/krew/pkg/constants"
)

func TestKrewInfo(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	out := string(test.WithIndex().Krew("info", validPlugin).RunOrFailOutput())
	expected := `INDEX: default`
	if !strings.Contains(out, expected) {
		t.Fatalf("info output doesn't have %q. output=%q", expected, out)
	}
}

func TestKrewInfoInvalidPlugin(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	plugin := "invalid-plugin"
	_, err := test.WithIndex().Krew("info", plugin).Run()
	if err == nil {
		t.Errorf("Expected `krew info %s` to fail", plugin)
	}
}

func TestKrewInfoCustomIndex(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test = test.WithEnv(constants.EnableMultiIndexSwitch, 1).WithIndex()
	defaultIndexRoot := test.TempDir().Path("index/" + constants.DefaultIndexName)

	test.Krew("index", "add", "foo", defaultIndexRoot).RunOrFail()
	test.Krew("install", "foo/"+validPlugin).RunOrFail()

	out := string(test.Krew("info", "foo/"+validPlugin).RunOrFailOutput())
	expected := `INDEX: foo`
	if !strings.Contains(out, expected) {
		t.Fatalf("info output doesn't have %q. output=%q", expected, out)
	}
}
