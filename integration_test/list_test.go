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
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestKrewList(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	initialList := test.WithIndex().Krew("list").RunOrFailOutput()
	initialOut := []byte{'\n'}

	if diff := cmp.Diff(initialList, initialOut); diff != "" {
		t.Fatalf("expected empty output from 'list':\n%s", diff)
	}

	test.Krew("install", validPlugin).RunOrFail()
	expected := []byte(validPlugin + "\n")

	eventualList := test.Krew("list").RunOrFailOutput()
	if diff := cmp.Diff(eventualList, expected); diff != "" {
		t.Fatalf("'list' output doesn't match:\n%s", diff)
	}

	test.Krew("uninstall", validPlugin).RunOrFail()
	expected = []byte{'\n'}

	uninstallList := test.Krew("list").RunOrFailOutput()
	if diff := cmp.Diff(uninstallList, expected); diff != "" {
		t.Fatalf("'list' output doesn't match:\n%s", diff)
	}

	test.Krew("install", validPlugin5, validPlugin4, validPlugin3, validPlugin2, validPlugin).RunOrFail()
	expected = []byte(validPlugin + "\n")

	sortedList := test.Krew("list").RunOrFailOutput()
	if diff := cmp.Diff(sortedList, expected); diff != "" {
		t.Fatalf("'list' output doesn't match:\n%s", diff)
	}

	expected = []byte("PLUGIN\tVERSION\n" + "v")
	overrideList := test.Krew("list", "-o").RunOrFailOutput()
	if diff := cmp.Diff(overrideList, expected); diff != "" {
		t.Fatalf("'list' output doesn't match:\n%s", diff)
	}
	overrideList = test.Krew("list", "--override").RunOrFailOutput()
	if diff := cmp.Diff(overrideList, expected); diff != "" {
		t.Fatalf("'list' output doesn't match:\n%s", diff)
	}
}
