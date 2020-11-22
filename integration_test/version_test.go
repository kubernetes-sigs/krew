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
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

func TestKrewVersion(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	stdOut := string(test.Krew("version").RunOrFailOutput())
	err := checkRequiredSubstrings(test, "https://github.com/kubernetes-sigs/krew-index.git", stdOut)
	if err != nil {
		t.Error(err)
	}
}

func TestKrewVersion_CustomDefaultIndexURI(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	stdOut := string(test.WithEnv("KREW_DEFAULT_INDEX_URI", "foo").Krew("version").RunOrFailOutput())
	err := checkRequiredSubstrings(test, "foo", stdOut)
	if err != nil {
		t.Error(err)
	}
}

func checkRequiredSubstrings(test *ITest, index, stdOut string) error {
	lineSplit := regexp.MustCompile(`\s+`)
	actual := make(map[string]string)
	for _, line := range strings.Split(stdOut, "\n") {
		if line == "" {
			continue
		}
		optionValue := lineSplit.Split(line, 2)
		if len(optionValue) < 2 {
			return errors.Errorf("`%v` is not an `OPTION VALUE` pair separated by spaces", optionValue)
		}
		actual[optionValue[0]] = optionValue[1]
	}

	requiredSubstrings := map[string]string{
		"OPTION":           "VALUE",
		"BasePath":         test.Root(),
		"GitTag":           "",
		"GitCommit":        "",
		"IndexURI":         index,
		"IndexPath":        path.Join(test.Root(), "index"),
		"InstallPath":      path.Join(test.Root(), "store"),
		"BinPath":          path.Join(test.Root(), "bin"),
		"DetectedPlatform": "/",
	}

	for k, v := range requiredSubstrings {
		got, ok := actual[k]
		if !ok {
			return errors.Errorf("`krew version` output doesn't contain field %q", k)
		} else if !strings.Contains(got, v) {
			return errors.Errorf("`krew version` %q field doesn't contain string %q (got: %q)", k, v, got)
		}
	}

	return nil
}
