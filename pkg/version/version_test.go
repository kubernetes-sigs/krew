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
