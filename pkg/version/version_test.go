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

package version

import "testing"

func TestGitCommit(t *testing.T) {
	orig := gitCommit
	defer func() { gitCommit = orig }()

	gitCommit = ""
	if v := GitCommit(); v != "unknown" {
		t.Errorf("empty gitCommit, expected=\"unknown\" got=%q", v)
	}

	gitCommit = "abcdef"
	if v := GitCommit(); v != "abcdef" {
		t.Errorf("empty gitCommit, expected=\"abcdef\" got=%q", v)
	}
}

func TestGitTag(t *testing.T) {
	orig := gitTag
	defer func() { gitTag = orig }()

	gitTag = ""
	if v := GitTag(); v != "unknown" {
		t.Errorf("empty gitTag, expected=\"unknown\" got=%q", v)
	}

	gitTag = "abcdef"
	if v := GitTag(); v != "abcdef" {
		t.Errorf("empty gitTag, expected=\"abcdef\" got=%q", v)
	}
}
