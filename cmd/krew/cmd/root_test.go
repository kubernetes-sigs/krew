// Copyright 2020 The Kubernetes Authors.
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
	"strings"
	"testing"
)

func TestUsageTemplateContainsReplacements(t *testing.T) {
	usageTemplate := rootCmd.UsageTemplate()

	expectedStrings := []string{
		"kubectl {{.CommandPath}}",
		"kubectl {{.UseLine}}",
	}

	for _, expectedString := range expectedStrings {
		if !strings.Contains(usageTemplate, expectedString) {
			t.Errorf("expected usage template to contain '%s' but it did not:\n%v", expectedString, usageTemplate)
		}
	}
}
