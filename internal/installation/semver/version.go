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

// Package semver is a wrapper for handling of semantic version
// (https://semver.org) values.
package semver

import (
	"strings"

	"github.com/pkg/errors"
	k8sver "k8s.io/apimachinery/pkg/util/version"
)

// Version is in-memory representation of a semantic version
// (https://semver.org) value.
type Version k8sver.Version

// String returns string representation of a semantic version value with a
// leading 'v' character.
func (v Version) String() string {
	vv := k8sver.Version(v)
	s := (&vv).String()
	if !strings.HasPrefix(s, "v") {
		s = "v" + s
	}
	return s
}

// Parse parses a semantic version value with a leading 'v' character.
func Parse(s string) (Version, error) {
	var vv Version
	if !strings.HasPrefix(s, "v") {
		return vv, errors.Errorf("version string %q not starting with 'v'", s)
	}
	v, err := k8sver.ParseSemantic(s)
	if err != nil {
		return vv, err
	}
	return Version(*v), nil
}

// Less checks if a is strictly less than b (a<b).
func Less(a, b Version) bool {
	aa := k8sver.Version(a)
	bb := k8sver.Version(b)
	return aa.LessThan(&bb)
}
