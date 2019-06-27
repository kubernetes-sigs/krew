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

package krew

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"sigs.k8s.io/krew/pkg/testutil"
)

const krewBinaryEnv = "KREW_BINARY"

// ITest is used to set up `krew` integration tests.
type ITest struct {
	t             *testing.T
	plugin        string
	pluginsBinDir string
	args          []string
	env           []string
	tempDir       *testutil.TempDir
}

// NewTest creates a fluent krew ITest.
func NewTest(t *testing.T) (*ITest, func()) {
	tempDir, cleanup := testutil.NewTempDir(t)
	binDir := setupKrewBin(t, tempDir)

	return &ITest{
		t:             t,
		pluginsBinDir: binDir,
		env: []string{
			fmt.Sprintf("KREW_ROOT=%s", tempDir.Root()),
			fmt.Sprintf("PATH=%s", augmentPATH(t, binDir)),
		},
		tempDir: tempDir,
	}, cleanup
}

// setupKrewBin symlinks the $KREW_BINARY to $tempDir/bin and returns the path
// to this directory.
func setupKrewBin(t *testing.T, tempDir *testutil.TempDir) string {
	krewBinary, found := os.LookupEnv(krewBinaryEnv)
	if !found {
		t.Fatalf("%s environment variable pointing to krew binary not set", krewBinaryEnv)
	}
	binPath := tempDir.Path("bin")
	if err := os.MkdirAll(binPath, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(krewBinary, filepath.Join(binPath, "kubectl-krew")); err != nil {
		t.Fatalf("cannot link krew binary: %s", err)
	}
	return binPath
}

// augmentPATH apprends the value to the current $PATH and returns the new
// value.
func augmentPATH(t *testing.T, v string) string {
	curPath, found := os.LookupEnv("PATH")
	if !found {
		t.Fatalf("$PATH variable is not set up, required to run tests")
	}

	return v + string(os.PathListSeparator) + curPath
}

func (it *ITest) lookupExecutable(file string) error {
	orig := os.Getenv("PATH")
	defer func() { os.Setenv("PATH", orig) }()

	binPath := filepath.Join(it.Root(), "bin")
	os.Setenv("PATH", binPath)

	_, err := exec.LookPath(file)
	return err
}

// AssertExecutableInPATH asserts that the executable file is in bin path.
func (it *ITest) AssertExecutableInPATH(file string) {
	it.t.Helper()
	if err := it.lookupExecutable(file); err != nil {
		it.t.Fatalf("executable %s not in PATH: %+v", file, err)
	}
}

// AssertExecutableNotInPATH asserts that the executable file is not in bin
// path.
func (it *ITest) AssertExecutableNotInPATH(file string) {
	it.t.Helper()
	if err := it.lookupExecutable(file); err == nil {
		it.t.Fatalf("executable %s still exists in PATH", file)
	}
}

// Krew configures the runner to call krew with arguments args.
func (it *ITest) Krew(args ...string) *ITest {
	it.plugin = "krew"
	it.args = args
	return it
}

// Root returns the krew root directory for this test.
func (it *ITest) Root() string {
	return it.tempDir.Root()
}

// WithIndex initializes the index with the actual krew-index from github/kubernetes-sigs/krew-index.
func (it *ITest) WithIndex() *ITest {
	it.initializeIndex()
	return it
}

// WithEnv sets an environment variable for the krew run.
func (it *ITest) WithEnv(key string, value interface{}) *ITest {
	if key == "KREW_ROOT" {
		glog.V(1).Infoln("Overriding KREW_ROOT in tests is forbidden")
		return it
	}
	it.env = append(it.env, fmt.Sprintf("%s=%v", key, value))
	return it
}

// RunOrFail runs the krew command and fails the test if the command returns an error.
func (it *ITest) RunOrFail() {
	it.t.Helper()
	if err := it.Run(); err != nil {
		it.t.Fatal(err)
	}
}

// Run runs the krew command.
func (it *ITest) Run() error {
	it.t.Helper()

	cmd := it.cmd(context.Background())
	glog.V(1).Infoln(cmd.Args)

	start := time.Now()
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "krew %v", it.args)
	}

	glog.V(1).Infoln("Ran in", time.Since(start))
	return nil
}

// RunOrFailOutput runs the krew command and fails the test if the command
// returns an error. It only returns the standard output.
func (it *ITest) RunOrFailOutput() []byte {
	it.t.Helper()

	cmd := it.cmd(context.Background())
	glog.V(1).Infoln(cmd.Args)

	start := time.Now()
	out, err := cmd.CombinedOutput()
	if err != nil {
		it.t.Fatalf("krew %v: %v, %s", it.args, err, out)
	}

	glog.V(1).Infoln("Ran in", time.Since(start))
	return out
}

func (it *ITest) cmd(ctx context.Context) *exec.Cmd {
	args := make([]string, 0, len(it.args)+1)
	args = append(args, it.plugin)
	args = append(args, it.args...)

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Env = it.env // clear env, do not inherit from system
	return cmd
}

func (it *ITest) TempDir() *testutil.TempDir {
	return it.tempDir
}
