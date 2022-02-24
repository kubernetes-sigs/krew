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

package internal

import (
	"os"
	"strings"
	"testing"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/testutil"
)

const environmentPath = "/home/../home/user//./.krew"

func TestIsBinDirInPATH_firstRun(t *testing.T) {
	tempDir := testutil.NewTempDir(t)

	paths := environment.NewPaths(tempDir.Path("does-not-exist"))
	res := IsBinDirInPATH(paths)
	if res == false {
		t.Errorf("expected positive result on first run")
	}
}

func TestIsBinDirInPATH_secondRun(t *testing.T) {
	tempDir := testutil.NewTempDir(t)

	paths := environment.NewPaths(tempDir.Root())
	res := IsBinDirInPATH(paths)
	if res == true {
		t.Errorf("expected negative result on second run")
	}
}

// TestIsBinDirInPATH_NonNormalized case when PATH content is not well normalized
func TestIsBinDirInPATH_NonNormalized(t *testing.T) {
	tempDir := testutil.NewTempDir(t)

	// set PATH with non-normalized folder path
	t.Setenv("PATH", tempDir.Path(environmentPath+"/bin"))

	paths := environment.NewPaths(tempDir.Path(environmentPath))
	err := os.MkdirAll(paths.BasePath(), os.ModePerm)
	if err != nil {
		t.Fatalf("os.MkdirAll(%s) failed with %v", paths.BasePath(), err)
	}

	got := IsBinDirInPATH(paths)
	if got == false {
		t.Errorf("IsBinDirPATH(%v) = %t, want true", paths, got)
	}
}

func TestSetupInstructions_windows(t *testing.T) {
	const instructionsContain = `USERPROFILE`
	t.Setenv("KREW_OS", "windows")
	instructions := SetupInstructions()
	if !strings.Contains(instructions, instructionsContain) {
		t.Errorf("expected %q\nto contain %q", instructions, instructionsContain)
	}
}

func TestSetupInstructions_unix(t *testing.T) {
	tests := []struct {
		name                string
		shell               string
		instructionsContain string
	}{
		{
			name:                "When the shell is zsh",
			shell:               "/bin/zsh",
			instructionsContain: "~/.zshrc",
		},
		{
			name:                "When the shell is bash",
			shell:               "/bin/bash",
			instructionsContain: "~/.bash_profile or ~/.bashrc",
		},
		{
			name:                "When the shell is fish",
			shell:               "/bin/fish",
			instructionsContain: "config.fish",
		},
		{
			name:                "When the shell is unknown",
			shell:               "other",
			instructionsContain: "~/.bash_profile, ~/.bashrc, or ~/.zshrc",
		},
	}

	// always set KREW_OS, so that tests succeed on windows
	t.Setenv("KREW_OS", "linux")
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			tt.Setenv("SHELL", test.shell)
			instructions := SetupInstructions()
			if !strings.Contains(instructions, test.instructionsContain) {
				tt.Errorf("expected %q\nto contain %q", instructions, test.instructionsContain)
			}
		})
	}
}
