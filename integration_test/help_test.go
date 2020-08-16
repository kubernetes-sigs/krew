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
	"testing"
)

func TestKrewHelp(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.Krew().RunOrFail() // no args
	test.Krew("help").RunOrFail()
	test.Krew("-h").RunOrFail()
	test.Krew("--help").RunOrFail()
}

func TestRootHelpShowsKubectlPrefix(t *testing.T) {
	skipShort(t)
	test := NewTest(t)

	out := string(test.Krew("help").RunOrFailOutput())

	expect := []*regexp.Regexp{
		regexp.MustCompile(`(?m)Usage:\s+kubectl krew`),
		regexp.MustCompile(`(?m)Use "kubectl krew`),
	}

	for _, e := range expect {
		if !e.MatchString(out) {
			t.Errorf("output does not have matching string to pattern %s ; output=%s", e, out)
		}
	}
}
