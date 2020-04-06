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

package cmd

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/index/indexoperations"
	"sigs.k8s.io/krew/internal/testutil"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

func Test_isDefaultIndex(t *testing.T) {
	if !isDefaultIndex("") {
		t.Error("empty string must indicate default index")
	}
	if !isDefaultIndex("default") { // nb: intentionally not using the const to ensure compatibility
		t.Error("default index must indicate default index")
	}
	if isDefaultIndex("foo") {
		t.Error("name=foo must not indicate default index")
	}
}

func TestIndexOf(t *testing.T) {
	noIndex := testutil.NewReceipt().WithPlugin(testutil.NewPlugin().V()).WithStatus(index.ReceiptStatus{}).V()
	if got := indexOf(noIndex); got != constants.DefaultIndexName {
		t.Errorf("expected default index for no index in status; got=%q", got)
	}
	defaultIndex := testutil.NewReceipt().WithPlugin(testutil.NewPlugin().V()).V()
	if got := indexOf(defaultIndex); got != constants.DefaultIndexName {
		t.Errorf("expected 'default' for default index; got=%q", got)
	}
	customIndex := testutil.NewReceipt().WithPlugin(testutil.NewPlugin().V()).WithStatus(
		index.ReceiptStatus{Source: index.SourceIndex{Name: "foo"}}).V()
	if got := indexOf(customIndex); got != "foo" {
		t.Errorf("expected custom index name; got=%q", got)
	}
}

func Test_displayName(t *testing.T) {
	type args struct {
		p     index.Plugin
		index string
	}
	tests := []struct {
		name     string
		in       args
		expected string
	}{
		{
			name: "explicit default index",
			in: args{
				p:     testutil.NewPlugin().WithName("foo").V(),
				index: constants.DefaultIndexName,
			},
			expected: "foo",
		},
		{
			name: "no index",
			in: args{
				p:     testutil.NewPlugin().WithName("foo").V(),
				index: "",
			},
			expected: "foo",
		},
		{
			name: "custom index",
			in: args{
				p:     testutil.NewPlugin().WithName("bar").V(),
				index: "foo",
			},
			expected: "foo/bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := displayName(tt.in.p, tt.in.index)
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Fatalf("expected name to match: %s", diff)
			}
		})
	}
}

func Test_canonicalName(t *testing.T) {
	p1 := testutil.NewPlugin().WithName("foo").V()
	if expected, got := "default/foo", canonicalName(p1, ""); got != expected {
		t.Errorf("expected=%q; got=%q", expected, got)
	}
	p2 := testutil.NewPlugin().WithName("bar").V()
	if expected, got := "default/bar", canonicalName(p2, "default"); got != expected {
		t.Errorf("expected=%q; got=%q", expected, got)
	}
	p3 := testutil.NewPlugin().WithName("quux").V()
	if expected, got := "custom/quux", canonicalName(p3, "custom"); got != expected {
		t.Errorf("expected=%q; got=%q", expected, got)
	}
}

func Test_allIndexes(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()
	paths := environment.NewPaths(tmpDir.Root())

	expected := []indexoperations.Index{
		{
			Name: constants.DefaultIndexName,
			URL:  constants.DefaultIndexURI,
		},
	}
	actual, err := allIndexes(paths)
	if err != nil {
		t.Errorf("unexpected error getting indexes: %s", err)
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("got diffent output: %s", diff)
	}

	os.Setenv(constants.EnableMultiIndexSwitch, "1")
	defer os.Unsetenv(constants.EnableMultiIndexSwitch)

	expected = []indexoperations.Index{
		{
			Name: "custom",
			URL:  "https://github.com/custom/index.git",
		},
		{
			Name: "foo",
			URL:  "https://github.com/foo/index.git",
		},
	}

	for _, index := range expected {
		path := paths.IndexPath(index.Name)
		tmpDir.InitEmptyGitRepo(path, index.URL)
	}

	actual, err = allIndexes(paths)
	if err != nil {
		t.Errorf("unexpected error getting indexes: %s", err)
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("got diffent output: %s", diff)
	}
}
