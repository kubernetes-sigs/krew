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
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
)

const (
	testRemoteRepo  = "./testdata/test-remote-repo"
	testClonedRepo  = "./testdata/test-cloned-repo"
	testUnknownRepo = "./testdata/unknown-repo"
	testEmptyRepo   = "./testdata/test-folder"
	testNewFile     = "example-git-file"
)

// setupGitRepo is a helper function used to setup the test git repo
func setupRemoteRepos(t *testing.T, localRepo, newCloneRepo string) {
	t.Helper()
	// get the pwd
	pwd := os.Getenv("PWD")
	pathToLocalRepo := filepath.Join(pwd, localRepo)
	pathToNewCloneRepo := filepath.Join(pwd, newCloneRepo)
	// create the git repo
	if err := createGitRepo(t, pathToLocalRepo); err != nil {
		t.Fatalf("failed to create test repo: %v", err)
	}
	// clone the existing repo
	if err := cloneGitRepo(t, pathToLocalRepo, pathToNewCloneRepo); err != nil {
		t.Fatalf("failed to clone test repo: %v", err)
	}
}

// createGitRepo is a helper function used to create a git repo for local testing
func createGitRepo(t *testing.T, localRepo string) error {
	t.Helper()
	// create a git local repo
	r, err := git.PlainInit(localRepo, false)
	if err != nil {
		t.Error("error while initializing git repo")
		return err
	}
	// add a file to the repo
	filename := filepath.Join(localRepo, testNewFile)
	if err := ioutil.WriteFile(filename, []byte("this is a test repo for gitutil package"), 0644); err != nil {
		t.Error("error while creating test file")
		return err
	}
	// get the reference of the local repo via Worktree
	w, err := r.Worktree()
	if err != nil {
		t.Error("error while getting worktree")
		return err
	}
	// git add the file
	if err := w.AddGlob("."); err != nil {
		t.Error("error while adding file to git repo")
		return err
	}
	// git commit the file
	if _, err := w.Commit("init commit", &git.CommitOptions{
		Author: &object.Signature{},
	}); err != nil {
		t.Error("error while committing file to git repo")
		return err
	}

	return nil
}

// cloneGitRepo is a helper function used to clone the local repo to a new path
func cloneGitRepo(t *testing.T, existingRemoteRepo, newCloneRepo string) error {
	t.Helper()
	// git clone using the local repo
	r, err := git.PlainClone(newCloneRepo, false,
		&git.CloneOptions{URL: existingRemoteRepo})
	if err != nil {
		return err
	}
	// add upstream remote
	if _, err := r.CreateRemote(&config.RemoteConfig{
		Name: "upstream",
		URLs: []string{newCloneRepo},
	}); err != nil {
		return err
	}
	return nil
}

// initLogging initializes the logging with klog.
func initLogging(t *testing.T, enable bool) {
	t.Helper()
	if enable {
		klog.InitFlags(nil)
		flag.Set("v", "3")
		flag.Parse()
		log := klogr.New().WithName("test").WithValues("gitutil", "pkg")
		log.Info("testexec", "withlogging", 1, "withinfo", map[string]int{"k": 1})
		log.V(3).Info("testing gitutil package")
		klog.Flush()
	}
}

// TestEnsureCloned tests the EnsureCloned function
func TestEnsureCloned(t *testing.T) {
	// create a local git repository in testdata
	if err := createGitRepo(t, testRemoteRepo); err != nil {
		t.Fatalf("failed to create test repo: %v", err)
	}
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
			name: "test clone existing repo",
			args: args{
				uri:             testRemoteRepo,
				destinationPath: testClonedRepo,
			},
			wantErr: false,
		},
		{
			name: "test where uri is invalid",
			args: args{
				uri:             testUnknownRepo,
				destinationPath: testEmptyRepo,
			},
			wantErr: true,
		},
		{
			name: "test where destination path is invalid",
			args: args{
				uri:             testUnknownRepo,
				destinationPath: testEmptyRepo,
			},
			wantErr: true,
		},
		{
			name: "test where repo is not cloned and not a git repo",
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
	// cleanup the git testdata repositories used for the testing in this package.
	t.Run("cleanup test git repos", func(t *testing.T) {
		t.Cleanup(func() {
			os.RemoveAll(testRemoteRepo)
			os.RemoveAll(testClonedRepo)
		})
	})
}

// TestIsGitCloned tests IsGitCloned function
func TestIsGitCloned(t *testing.T) {
	// setup local git repositories in testdata
	setupRemoteRepos(t, testRemoteRepo, testClonedRepo)
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
			name: "test if the test-cloned-repo is cloned",
			args: args{
				gitPath: testClonedRepo,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "test check for unknown repo",
			args: args{
				gitPath: testUnknownRepo,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "test check for git repo",
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
	// cleanup the git testdata repositories used for the testing in this package.
	t.Run("cleanup test git repos", func(t *testing.T) {
		t.Cleanup(func() {
			os.RemoveAll(testRemoteRepo)
			os.RemoveAll(testClonedRepo)
		})
	})
}

// Test_updateAndCleanUntracked tests the updateAndCleanUntracked function
func Test_updateAndCleanUntracked(t *testing.T) {
	// setup local git repositories in testdata
	setupRemoteRepos(t, testRemoteRepo, testClonedRepo)
	type args struct {
		destinationPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test where destination path is invalid",
			args: args{
				destinationPath: testEmptyRepo,
			},
			wantErr: true,
		},
		{
			name: "test where destination path is invalid - unknown repo",
			args: args{
				destinationPath: testUnknownRepo,
			},
			wantErr: true,
		},
		{
			name: "test update and clean untracked",
			args: args{
				destinationPath: testClonedRepo,
			},
			wantErr: false,
		},
		{
			name: "test upstream is not configured",
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
	// cleanup the git testdata repositories used for the testing in this package.
	t.Run("cleanup test git repos", func(t *testing.T) {
		t.Cleanup(func() {
			os.RemoveAll(testRemoteRepo)
			os.RemoveAll(testClonedRepo)
		})
	})
}

// TestEnsureUpdated tests the EnsureUpdated function
func TestEnsureUpdated(t *testing.T) {
	// setup local git repositories in testdata
	setupRemoteRepos(t, testRemoteRepo, testClonedRepo)
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
			name: "test update and clean untracked - no clone",
			args: args{
				uri:             testEmptyRepo,
				destinationPath: testRemoteRepo,
			},
			wantErr: true,
		},
		{
			name: "test update and clean untracked - unknown repo",
			args: args{
				uri:             testUnknownRepo,
				destinationPath: testRemoteRepo,
			},
			wantErr: true,
		},
		{
			name: "test update and clean untracked - empty repo",
			args: args{
				uri:             testUnknownRepo,
				destinationPath: testEmptyRepo,
			},
			wantErr: true,
		},
		{
			name: "test update and clean untracked",
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
	// cleanup the git testdata repositories used for the testing in this package.
	t.Run("cleanup test git repos", func(t *testing.T) {
		t.Cleanup(func() {
			os.RemoveAll(testRemoteRepo)
			os.RemoveAll(testClonedRepo)
		})
	})
}

// TestGetRemoteURL tests the GetRemoteURL function
func TestGetRemoteURL(t *testing.T) {
	// setup the git repositories for the testing in testdata
	setupRemoteRepos(t, testRemoteRepo, testClonedRepo)
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
			name: "test get remote url for existing repo",
			args: args{
				dir: testClonedRepo,
			},
			want:    filepath.Join(os.Getenv("PWD"), "/"+testRemoteRepo),
			wantErr: false,
		},
		{
			name: "test get remote url for existing repo - remote not set",
			args: args{
				dir: testRemoteRepo,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "test get remote url for existing repo - repo does not exists",
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
	// cleanup the git testdata repositories used for the testing in this package.
	t.Run("cleanup test git repos", func(t *testing.T) {
		t.Cleanup(func() {
			os.RemoveAll(testRemoteRepo)
			os.RemoveAll(testClonedRepo)
		})
	})
}

// TestExec tests the Exec function
func TestExec(t *testing.T) {
	// create a local git repository in testdata
	if err := createGitRepo(t, testRemoteRepo); err != nil {
		t.Fatalf("failed to create git repo: %v", err)
	}
	initLogging(t, true)

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
			name: "test exec command",
			args: args{
				pwd:  testRemoteRepo,
				args: []string{"status"},
			},
			want:    "On branch master\nnothing to commit, working tree clean",
			wantErr: false,
		},
		{
			name: "test exec command - bad command",
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
	// cleanup the git testdata repositories used for the testing in this package.
	t.Run("cleanup test git repos", func(t *testing.T) {
		t.Cleanup(func() {
			os.RemoveAll(testRemoteRepo)
			os.RemoveAll(testClonedRepo)
		})
	})
}
