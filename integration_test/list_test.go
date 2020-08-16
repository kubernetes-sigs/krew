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
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/index/indexscanner"
	"sigs.k8s.io/krew/internal/installation/receipt"
	"sigs.k8s.io/krew/internal/testutil"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

func TestKrewList(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test = test.WithDefaultIndex().WithCustomIndexFromDefault("foo")
	initialList := test.Krew("list").RunOrFailOutput()
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

	test.Krew("install", "foo/"+validPlugin2).RunOrFail()

	want := []string{validPlugin, "foo/" + validPlugin2}
	actual := lines(test.Krew("list").RunOrFailOutput())
	if diff := cmp.Diff(actual, want); diff != "" {
		t.Fatalf("'list' output doesn't match:\n%s", diff)
	}
}

func TestKrewListSorted(t *testing.T) {
	skipShort(t)
	test := NewTest(t)

	test = test.WithDefaultIndex()

	paths := environment.NewPaths(test.Root())
	ps, err := indexscanner.LoadPluginListFromFS(paths.IndexPluginsPath(constants.DefaultIndexName))
	if err != nil {
		t.Fatal(err)
	}

	indexes := []string{"", "default", "bar"}
	for i, p := range ps {
		src := indexes[i%len(indexes)]
		r := testutil.NewReceipt().WithPlugin(p).WithStatus(index.ReceiptStatus{Source: index.SourceIndex{Name: src}}).V()
		if err := receipt.Store(r, paths.PluginInstallReceiptPath(p.Name)); err != nil {
			t.Fatal(err)
		}
	}
	out := lines(test.Krew("list").RunOrFailOutput())
	if !sort.StringsAreSorted(out) {
		t.Fatalf("list output is not sorted: [%s]", strings.Join(out, ", "))
	}
}
