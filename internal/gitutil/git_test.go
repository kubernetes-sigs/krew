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

package gitutil

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testUnknownRepo = "./testdata/unknown-repo"
)

// createEmptyDir creates a temporary empty dir
func createEmptyDir(t *testing.T) string {
	t.Helper()
	testDir, err := ioutil.TempDir("", "test-empty-dir")
	if err != nil {
		t.Fatalf("failed to create test-empty-dir: %s", err.Error())
	}
	t.Cleanup(func() {
		os.RemoveAll(testDir)
	})
	return testDir
}

// setupGitRepo is a helper function used to setup the test git repo
func setupRemoteRepos(t *testing.T) (string, string) {
	const (
		testRemoteRepo = "git-remote-repo"
		testClonedRepo = "git-cloned-repo"
	)
	t.Helper()
	// create the git repo
	pathToLocalRepo, err := ioutil.TempDir("", testRemoteRepo)
	if err != nil {
		t.Fatalf("failed to create remote repo %s: %s", testRemoteRepo, err.Error())
	}
	createGitRepo(t, pathToLocalRepo)
	// clone the existing repo
	pathToNewCloneRepo, err := ioutil.TempDir("", testClonedRepo)
	if err != nil {
		t.Fatalf("failed to create clone repo %s: %s", testClonedRepo, err.Error())
	}
	cloneGitRepo(t, pathToLocalRepo, pathToNewCloneRepo)
	t.Cleanup(func() {
		os.RemoveAll(pathToLocalRepo)
		os.RemoveAll(pathToNewCloneRepo)
	})
	return pathToLocalRepo, pathToNewCloneRepo
}

// execute the git commands
func execute(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	_, err := cmd.Output()
	if err != nil {
		command := []string{"git"}
		command = append(command, args...)
		t.Fatalf("error while executing command %s: %s", strings.Join(command, " "), err.Error())
	}
}

// createGitRepo is a helper function used to create a git repo for local testing
func createGitRepo(t *testing.T, localRepo string) {
	const (
		testNewFile = "example-git-file"
	)
	t.Helper()
	// create a git local repo
	execute(t, localRepo, "init")
	// add a file to the repo
	filename := filepath.Join(localRepo, testNewFile)
	if err := os.WriteFile(filename, []byte("this is a test repo for gitutil package"), 0o644); err != nil {
		t.Fatalf("error while creating test file: %s", err.Error())
	}
	// git add command
	execute(t, localRepo, "add", ".")
	// git commit command
	execute(t, localRepo, "-c", "user.name='test'", "-c", "user.email='test@example.com'", "commit", "-m", "\"init\"")
}

// cloneGitRepo is a helper function used to clone the local repo to a new path
func cloneGitRepo(t *testing.T, existingRemoteRepo, newCloneRepo string) {
	t.Helper()
	// git clone command
	execute(t, "", "clone", existingRemoteRepo, newCloneRepo)
	// git remote add upstream command
	execute(t, newCloneRepo, "remote", "add", "upstream", newCloneRepo)
}

// TestEnsureCloned tests the EnsureCloned function
func TestEnsureCloned(t *testing.T) {
	// create a local git repository in testdata
	testRemoteRepo, testClonedRepo := setupRemoteRepos(t)
	testEmptyRepo := createEmptyDir(t)

	type args struct {
		uri             string
		destinationPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "clone existing repo",
			args: args{
				uri:             testRemoteRepo,
				destinationPath: testClonedRepo,
			},
			wantErr: false,
		},
		{
			name: "uri is invalid",
			args: args{
				uri:             testUnknownRepo,
				destinationPath: testEmptyRepo,
			},
			wantErr: true,
		},
		{
			name: "destination path is invalid",
			args: args{
				uri:             testUnknownRepo,
				destinationPath: testEmptyRepo,
			},
			wantErr: true,
		},
		{
			name: "repo is not cloned and not a git repo",
			args: args{
				uri:             testEmptyRepo,
				destinationPath: testUnknownRepo,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := EnsureCloned(tt.args.uri, tt.args.destinationPath); (err != nil) != tt.wantErr {
				t.Errorf("EnsureCloned() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

// TestIsGitCloned tests IsGitCloned function
func TestIsGitCloned(t *testing.T) {
	// setup local git repositories in testdata
	_, testClonedRepo := setupRemoteRepos(t)
	testEmptyRepo := createEmptyDir(t)

	type args struct {
		gitPath string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "check if the test-cloned-repo is cloned",
			args: args{
				gitPath: testClonedRepo,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "unknown repo",
			args: args{
				gitPath: testUnknownRepo,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "empty git repo",
			args: args{
				gitPath: testEmptyRepo,
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsGitCloned(tt.args.gitPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsGitCloned() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsGitCloned() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_updateAndCleanUntracked tests the updateAndCleanUntracked function
func Test_updateAndCleanUntracked(t *testing.T) {
	// setup local git repositories in testdata
	testRemoteRepo, testClonedRepo := setupRemoteRepos(t)
	testEmptyRepo := createEmptyDir(t)

	type args struct {
		destinationPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "destination path is invalid",
			args: args{
				destinationPath: testEmptyRepo,
			},
			wantErr: true,
		},
		{
			name: "destination path is invalid - unknown repo",
			args: args{
				destinationPath: testUnknownRepo,
			},
			wantErr: true,
		},
		{
			name: "update and clean untracked",
			args: args{
				destinationPath: testClonedRepo,
			},
			wantErr: false,
		},
		{
			name: "upstream is not configured",
			args: args{
				destinationPath: testRemoteRepo,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := updateAndCleanUntracked(tt.args.destinationPath); (err != nil) != tt.wantErr {
				t.Errorf("updateAndCleanUntracked() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestEnsureUpdated tests the EnsureUpdated function
func TestEnsureUpdated(t *testing.T) {
	// setup local git repositories in testdata
	testRemoteRepo, testClonedRepo := setupRemoteRepos(t)
	testEmptyRepo := createEmptyDir(t)

	type args struct {
		uri             string
		destinationPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "update and clean untracked - no clone",
			args: args{
				uri:             testEmptyRepo,
				destinationPath: testRemoteRepo,
			},
			wantErr: true,
		},
		{
			name: "update and clean untracked - unknown repo",
			args: args{
				uri:             testUnknownRepo,
				destinationPath: testRemoteRepo,
			},
			wantErr: true,
		},
		{
			name: "update and clean untracked - empty repo",
			args: args{
				uri:             testUnknownRepo,
				destinationPath: testEmptyRepo,
			},
			wantErr: true,
		},
		{
			name: "ensure updated",
			args: args{
				uri:             testRemoteRepo,
				destinationPath: testClonedRepo,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := EnsureUpdated(tt.args.uri, tt.args.destinationPath); (err != nil) != tt.wantErr {
				t.Errorf("EnsureUpdated() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGetRemoteURL tests the GetRemoteURL function
func TestGetRemoteURL(t *testing.T) {
	// setup the git repositories for the testing in testdata
	testRemoteRepo, testClonedRepo := setupRemoteRepos(t)

	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "get remote url for existing repo",
			args: args{
				dir: testClonedRepo,
			},
			want:    testRemoteRepo,
			wantErr: false,
		},
		{
			name: "remote not set",
			args: args{
				dir: testRemoteRepo,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "repo does not exists",
			args: args{
				dir: testUnknownRepo,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRemoteURL(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRemoteURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetRemoteURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestExec tests the Exec function
func TestExec(t *testing.T) {
	// create a local git repository in testdata
	testRemoteRepo, _ := setupRemoteRepos(t)

	type args struct {
		pwd  string
		args []string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "exec command",
			args: args{
				pwd:  testRemoteRepo,
				args: []string{"status"},
			},
			want:    "On branch master\nnothing to commit, working tree clean",
			wantErr: false,
		},
		{
			name: "exec bad command",
			args: args{
				pwd:  testRemoteRepo,
				args: []string{"unknown", "command"},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Exec(tt.args.pwd, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Exec() = %v, want %v", got, tt.want)
			}
		})
	}
}
