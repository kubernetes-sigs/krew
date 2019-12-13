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
	"os"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"

	"sigs.k8s.io/krew/internal/testutil"
	"sigs.k8s.io/krew/pkg/index"
)

func Test_osArch(t *testing.T) {
	inOS, inArch := runtime.GOOS, runtime.GOARCH
	out := osArch()
	if inOS != out.os {
		t.Errorf("returned OS=%q; expected=%q", out.os, inOS)
	}
	if inArch != out.arch {
		t.Errorf("returned Arch=%q; expected=%q", out.arch, inArch)
	}
}

func Test_osArch_override(t *testing.T) {
	customOS, customArch := "dragons", "metav1"
	os.Setenv("KREW_OS", customOS)
	os.Setenv("KREW_ARCH", customArch)
	defer func() {
		os.Unsetenv("KREW_ARCH")
		os.Unsetenv("KREW_OS")
	}()

	out := osArch()
	if customOS != out.os {
		t.Errorf("returned OS=%q; expected=%q", out.os, customOS)
	}
	if customArch != out.arch {
		t.Errorf("returned Arch=%q; expected=%q", out.arch, customArch)
	}
}

func Test_matchPlatform(t *testing.T) {
	target := goOSArch{os: "foo", arch: "amd64"}
	matchingPlatform := testutil.NewPlatform().WithOSArch(target.os, target.arch).V()
	differentOS := testutil.NewPlatform().WithOSArch("other", target.arch).V()
	differentArch := testutil.NewPlatform().WithOSArch(target.os, "other").V()

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
