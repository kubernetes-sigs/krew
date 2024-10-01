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
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Test_searchByNameAndDesc tests fuzzy search
// name matches are shown first, then the description matches
func Test_searchByNameAndDesc(t *testing.T) {
	testPlugins := []struct {
		keyword  string
		names    []string
		descs    []string
		expected []string
	}{
		{
			keyword: "foo",
			names:   []string{"foo", "bar", "foobar"}, // names match first
			descs: []string{
				"This is the description for the first plugin, not contain keyword",
				"This is the description to the second plugin, not contain keyword",
				"This is the description for the third plugin, not contain keyword",
			},
			expected: []string{"foo", "foobar"},
		},
		{
			keyword: "bar",
			names:   []string{"baz", "qux", "fred"}, // names not match
			descs: []string{
				"This is the description for the first plugin, contain keyword bar", // description match, but score < 0
				"This is the description for the second plugin, not contain keyword",
				"This is the description for the third plugin, contain ba fuzzy keyword", // fuzzy match, but score < 0
			},
			expected: []string{},
		},
		{
			keyword: "baz",
			names:   []string{"baz", "foo", "bar"}, // both name and description match
			descs: []string{
				"This is the description for the first plugin, contain keyword baz", // both name and description match
				"This is the description for the second plugin, not contain keyword",
				"This is the description for the third plugin, contain bar keyword",
			},
			expected: []string{"baz"},
		},
		{
			keyword: "",
			names:   []string{"plugin1", "plugin2", "plugin3"}, // empty keyword, only names match
			descs: []string{
				"Description for plugin1",
				"Description for plugin2",
				"Description for plugin3",
			},
			expected: []string{"plugin1", "plugin2", "plugin3"},
		},
	}

	for _, tp := range testPlugins {
		t.Run(tp.keyword, func(t *testing.T) {
			searchTarget := make([]searchItem, len(tp.names))
			for i, name := range tp.names {
				searchTarget[i] = searchItem{
					name:        name,
					description: tp.descs[i],
				}
			}
			result := searchByNameAndDesc(tp.keyword, searchTarget)
			if diff := cmp.Diff(tp.expected, result); diff != "" {
				t.Fatalf("expected %v does not match got %v", tp.expected, result)
			}
		})
	}
}
