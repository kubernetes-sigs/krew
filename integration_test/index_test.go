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

func TestKrewIndex(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	initialIndices := test.Krew("index").RunOrFailOutput()
	initialOut := []byte{}
	if diff := cmp.Diff(initialIndices, initialOut); diff != "" {
		t.Fatalf("expected empty output from 'index':\n%s", diff)
	}

	test.Krew("index", "add", "test", "https://github.com/kubernetes-sigs/krew-index.git").RunOrFail()
	expected := []byte("map[test:https://github.com/kubernetes-sigs/krew-index.git]\n")
	eventual := test.Krew("index").RunOrFailOutput()
	if diff := cmp.Diff(eventual, expected); diff != "" {
		t.Fatalf("'index' output doesn't match:\n%s", diff)
	}

	test.Krew("index", "add", "test2", "https://github.com/kubernetes-sigs/krew-index.git").RunOrFail()
	expected = []byte("map[test:https://github.com/kubernetes-sigs/krew-index.git test2:https://github.com/kubernetes-sigs/krew-index.git]\n")
	eventual = test.Krew("index").RunOrFailOutput()
	if diff := cmp.Diff(eventual, expected); diff != "" {
		t.Fatalf("'index' output doesn't match:\n%s", diff)
	}

	test.Krew("index", "remove", "test").RunOrFail()
	expected = []byte("map[test2:https://github.com/kubernetes-sigs/krew-index.git]\n")
	eventual = test.Krew("index").RunOrFailOutput()
	if diff := cmp.Diff(eventual, expected); diff != "" {
		t.Fatalf("'index' output doesn't match:\n%s", diff)
	}

	test.Krew("index", "remove", "test2").RunOrFail()
	expected = []byte{}
	eventual = test.Krew("index").RunOrFailOutput()
	if diff := cmp.Diff(eventual, expected); diff != "" {
		t.Fatalf("'index' output doesn't match:\n%s", diff)
	}
}
