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

package installation

import (
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"

	"sigs.k8s.io/krew/internal/testutil"
	"sigs.k8s.io/krew/pkg/index"
)

func Test_osArch(t *testing.T) {
	in := OSArchPair{OS: runtime.GOOS, Arch: runtime.GOARCH}

	if diff := cmp.Diff(in, OSArch()); diff != "" {
		t.Errorf("os/arch got a different result:\n%s", diff)
	}
}

func Test_osArch_override(t *testing.T) {
	customGoOSArch := OSArchPair{OS: "dragons", Arch: "metav1"}
	t.Setenv("KREW_OS", customGoOSArch.OS)
	t.Setenv("KREW_ARCH", customGoOSArch.Arch)

	if diff := cmp.Diff(customGoOSArch, OSArch()); diff != "" {
		t.Errorf("os/arch override got a different result:\n%s", diff)
	}
}

func Test_matchPlatform(t *testing.T) {
	target := OSArchPair{OS: "foo", Arch: "amd64"}
	matchingPlatform := testutil.NewPlatform().WithOSArch(target.OS, target.Arch).V()
	differentOS := testutil.NewPlatform().WithOSArch("other", target.Arch).V()
	differentArch := testutil.NewPlatform().WithOSArch(target.OS, "other").V()

	p, ok, err := matchPlatform([]index.Platform{differentOS, differentArch, matchingPlatform}, target)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("failed to find a match")
	}
	if diff := cmp.Diff(p, matchingPlatform); diff != "" {
		t.Fatalf("got a different object from the matching platform:\n%s", diff)
	}

	_, ok, err = matchPlatform([]index.Platform{differentOS, differentArch}, target)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("got a matching platform, but was not expecting")
	}
}
