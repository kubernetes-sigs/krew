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

package semver

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"empty", "", "", true},
		{"zero is valid", "v0.0.0", "v0.0.0", false},
		{"no v prefix", "1.0.0", "", true},
		{"spaces between v and value", "v 1.0.0", "", true},
		{"major", "v1", "", true},
		{"major.minor", "v1.2", "", true},
		{"major.minor.patch", "v1.2.3", "v1.2.3", false},
		{"major.minor.patch-suffix", "v1.2.3-beta.2+foo.bar", "v1.2.3-beta.2+foo.bar", false},

		{"empty pre-release identifier", "v1.0.1-", "", true},
		{"empty meta identifier", "v1.0.1+", "", true},
		{"negative value in major", "v-1.2.3", "", true},
		{"negative value in minor", "v1.-2.3", "", true},
		{"negative value in patch", "v1.2.-3", "", true},
		{"leading zero in major", "v01.2.3", "", true},
		{"leading zero in minor", "v1.02.3", "", true},
		{"leading zero in patch", "v1.2.03", "", true},
		{"major with alpha chars", "v0a.0.0", "", true},
		{"minor with alpha chars", "v0.0a.0", "", true},
		{"patch with alpha chars", "v0.0.0a", "", true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%s) error = %v, wantErr %v", tt.in, err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.String() != tt.want {
				t.Errorf("Parse(%s) = %q, want %q", tt.in, got.String(), tt.want)
			}
		})
	}
}

func TestLess(t *testing.T) {
	tests := []struct {
		a      string
		b      string // b≥a
		equals bool
	}{
		{"v0.1.2", "v0.1.2", true},                 // equals
		{"v1.0.0-alpha", "v1.0.0-alpha+foo", true}, // equals
		{"v1.0.0-0.3.7", "v1.0.0-alpha", false},
		{"v1.0.0-alpha", "v1.0.0-alpha.1", false},
		{"v1.0.0-alpha.1", "v1.0.0-alpha.2", false},
		{"v1.0.0-alpha.2", "v1.0.0-alpha.a", false},
		{"v1.0.0-alpha", "v1.0.0-beta", false},
		{"v1.0.1", "v1.0.2", false},
		{"v1.0.1", "v1.2.0", false},
		{"v1.0.1", "v2.1.0", false},
		{"v1.0.0-alpha.2", "v1.0.1-alpha.1", false},
		{"v1.0.1-rc1", "v1.0.1", false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s<%s", tt.a, tt.b), func(t *testing.T) {
			va, err := Parse(tt.a)
			if err != nil {
				t.Fatalf("error parsing version %q", tt.a)
			}
			vb, err := Parse(tt.b)
			if err != nil {
				t.Fatalf("error parsing version %q", tt.b)
			}

			if tt.equals {
				if o1, o2 := Less(va, vb), Less(vb, va); o1 || o2 {
					t.Errorf("Less(%s,%s)=%v and Less(%s,%s)=%v; but they both should be false since the values are equal", tt.a, tt.b, o1, tt.b, tt.a, o2)
				}
			} else {
				if !Less(va, vb) {
					t.Errorf("Less(%s,%s) returned false; expected true", tt.a, tt.b)
					return
				}
				if Less(vb, va) {
					t.Errorf("(flipped) Less(%s,%s) returned true; expected false", tt.b, tt.a)
				}
			}
		})
	}
}
