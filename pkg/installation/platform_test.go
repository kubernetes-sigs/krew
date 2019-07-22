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

	"sigs.k8s.io/krew/pkg/index"
	"sigs.k8s.io/krew/pkg/testutil"
)

func Test_osArch(t *testing.T) {
	inOS, inArch := runtime.GOOS, runtime.GOARCH
	outOS, outArch := osArch()
	if inOS != outOS {
		t.Errorf("returned OS=%q; expected=%q", outOS, inOS)
	}
	if inArch != outArch {
		t.Errorf("returned Arch=%q; expected=%q", outArch, inArch)
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

	outOS, outArch := osArch()
	if customOS != outOS {
		t.Errorf("returned OS=%q; expected=%q", outOS, customOS)
	}
	if customArch != outArch {
		t.Errorf("returned Arch=%q; expected=%q", outArch, customArch)
	}
}

func Test_matchPlatform(t *testing.T) {
	const targetOS, targetArch = "foo", "amd64"
	matchingPlatform := testutil.NewPlatform().WithOSArch(targetOS, targetArch).V()
	differentOS := testutil.NewPlatform().WithOSArch("other", targetArch).V()
	differentArch := testutil.NewPlatform().WithOSArch(targetOS, "other").V()

	p, ok, err := matchPlatform([]index.Platform{differentOS, differentArch, matchingPlatform}, targetOS, targetArch)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("failed to find a match")
	}
	if diff := cmp.Diff(p, matchingPlatform); diff != "" {
		t.Fatalf("got a different object from the matching platform:\n%s", diff)
	}

	_, ok, err = matchPlatform([]index.Platform{differentOS, differentArch}, targetOS, targetArch)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("got a matching platform, but was not expecting")
	}
}
