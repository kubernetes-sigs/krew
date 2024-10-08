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
	"regexp"
	"sort"
	"strings"
	"testing"

	"sigs.k8s.io/krew/pkg/constants"
)

func TestKrewSearchAll(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	output := test.WithDefaultIndex().Krew("search").RunOrFailOutput()
	availablePlugins := test.IndexPluginCount(constants.DefaultIndexName)
	if plugins := lines(output); len(plugins)-1 != availablePlugins {
		// the first line is the header
		t.Errorf("Expected %d plugins, got %d", availablePlugins, len(plugins)-1)
	}
}

func TestKrewSearchOne(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	plugins := lines(test.WithDefaultIndex().Krew("search", "krew").RunOrFailOutput())
	if len(plugins) < 2 {
		t.Errorf("Expected krew to be a valid plugin")
	}
	if !strings.HasPrefix(plugins[1], "krew ") {
		t.Errorf("The first match should be krew")
	}
}

func TestKrewSearchMultiIndex(t *testing.T) {
	skipShort(t)
	test := NewTest(t)
	test = test.WithDefaultIndex().WithCustomIndexFromDefault("foo")

	test.Krew("install", validPlugin).RunOrFail()
	test.Krew("install", "foo/"+validPlugin2).RunOrFail()

	output := string(test.Krew("search").RunOrFailOutput())
	wantPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?m)^` + validPlugin + `\b.*\byes`),
		regexp.MustCompile(`(?m)^` + validPlugin2 + `\b.*\bno`),
		regexp.MustCompile(`(?m)^foo/` + validPlugin + `\b.*\bno$`),
		regexp.MustCompile(`(?m)^foo/` + validPlugin2 + `\b.*\byes$`),
	}
	for _, p := range wantPatterns {
		if !p.MatchString(output) {
			t.Fatalf("pattern %s not found in search output=%s", p, output)
		}
	}
}

func TestKrewSearchMultiIndexSortedByDisplayName(t *testing.T) {
	skipShort(t)
	test := NewTest(t)
	test = test.WithDefaultIndex().WithCustomIndexFromDefault("foo")

	output := string(test.Krew("search").RunOrFailOutput())

	// match first column that is not NAME by matching everything up until a space
	names := regexp.MustCompile(`(?m)^[^\s|NAME]+\b`).FindAllString(output, -1)
	if len(names) < 10 {
		t.Fatalf("could not capture names")
	}
	if !sort.StringsAreSorted(names) {
		t.Fatalf("names are not sorted: [%s]", strings.Join(names, ", "))
	}
}
