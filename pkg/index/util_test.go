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

package index

import (
	"testing"

	"sigs.k8s.io/krew/pkg/constants"
)

func TestDefaultIndex(t *testing.T) {
	if got := DefaultIndex(); got != constants.DefaultIndexURI {
		t.Errorf("DefaultIndex() = %q, want %q", got, constants.DefaultIndexURI)
	}

	want := "foo"
	t.Setenv("KREW_DEFAULT_INDEX_URI", want)

	if got := DefaultIndex(); got != want {
		t.Errorf("DefaultIndex() = %q, want %q", got, want)
	}
}
