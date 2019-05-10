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

// Package version contains the version information of the krew binary.
package version

var (
	// gitCommit contains the git commit identifier.
	gitCommit string

	// gitTag contains the git tag or describe output.
	gitTag string
)

// GitCommit returns the value stamped into the binary at compile-time or a
// default "unknown" value.
func GitCommit() string {
	if gitCommit == "" {
		return "unknown"
	}
	return gitCommit
}

// GitTag returns the value stamped into the binary at compile-time or a
// default "unknown" value.
func GitTag() string {
	if gitTag == "" {
		return "unknown"
	}
	return gitTag
}
