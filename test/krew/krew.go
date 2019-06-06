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
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"sigs.k8s.io/krew/pkg/testutil"
)

const krewBinaryEnv = "KREW_BINARY"

// KrewTest is used to set up `krew` integration tests.
type KrewTest struct {
	t       *testing.T
	plugin  string
	args    []string
	env     []string
	tempDir *testutil.TempDir
}

// NewKrewTest creates a fluent krew KrewTest.
func NewKrewTest(t *testing.T) (*KrewTest, func()) {
	tempDir, cleanup := testutil.NewTempDir(t)
	pathEnv := setupPathEnv(t, tempDir)
	return &KrewTest{
		t:       t,
		env:     []string{pathEnv, fmt.Sprintf("KREW_ROOT=%s", tempDir.Root())},
		tempDir: tempDir,
	}, cleanup
}

func setupPathEnv(t *testing.T, tempDir *testutil.TempDir) string {
	krewBinPath := tempDir.Path("bin")
	if err := os.MkdirAll(krewBinPath, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	if krewBinary, found := os.LookupEnv(krewBinaryEnv); found {
		if err := os.Symlink(krewBinary, tempDir.Path("bin/kubectl-krew")); err != nil {
			t.Fatalf("Cannot link to krew: %s", err)
		}
	} else {
		t.Logf("Environment variable %q was not found, using krew installation from host", krewBinaryEnv)
	}

	path, found := os.LookupEnv("PATH")
	if !found {
		t.Fatalf("PATH variable is not set up")
	}

	return fmt.Sprintf("PATH=%s:%s", krewBinPath, path)
}

// Call configures the runner to call plugin with arguments args.
func (k *KrewTest) Call(plugin string, args ...string) *KrewTest {
	k.plugin = plugin
	k.args = args
	return k
}

// Krew configures the runner to call krew with arguments args.
func (k *KrewTest) Krew(args ...string) *KrewTest {
	k.plugin = "krew"
	k.args = args
	return k
}

// Root returns the krew root directory for this test.
func (k *KrewTest) Root() string {
	return k.tempDir.Root()
}

// WithIndex initializes the index with the actual krew-index from github/kubernetes-sigs/krew-index.
func (k *KrewTest) WithIndex() *KrewTest {
	k.initializeIndex()
	return k
}

// WithEnv sets an environment variable for the krew run.
func (k *KrewTest) WithEnv(key string, value interface{}) *KrewTest {
	if key == "KREW_ROOT" {
		glog.V(1).Infoln("Overriding KREW_ROOT in tests is forbidden")
		return k
	}
	k.env = append(k.env, fmt.Sprintf("%s=%v", key, value))
	return k
}

// RunOrFail runs the krew command and fails the test if the command returns an error.
func (k *KrewTest) RunOrFail() {
	k.t.Helper()
	if err := k.Run(); err != nil {
		k.t.Fatal(err)
	}
}

// Run runs the krew command.
func (k *KrewTest) Run() error {
	k.t.Helper()

	cmd := k.cmd(context.Background())
	glog.V(1).Infoln(cmd.Args)

	start := time.Now()
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "krew %v", k.args)
	}

	glog.V(1).Infoln("Ran in", time.Since(start))
	return nil
}

// RunOrFailOutput runs the krew command and fails the test if the command
// returns an error. It only returns the standard output.
func (k *KrewTest) RunOrFailOutput() []byte {
	k.t.Helper()

	cmd := k.cmd(context.Background())
	glog.V(1).Infoln(cmd.Args)

	start := time.Now()
	out, err := cmd.CombinedOutput()
	if err != nil {
		k.t.Fatalf("krew %v: %v, %s", k.args, err, out)
	}

	glog.V(1).Infoln("Ran in", time.Since(start))
	return out
}

func (k *KrewTest) cmd(ctx context.Context) *exec.Cmd {
	args := make([]string, 0, len(k.args)+1)
	args = append(args, k.plugin)
	args = append(args, k.args...)

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Env = append(os.Environ(), k.env...)

	return cmd
}

func (k *KrewTest) TempDir() *testutil.TempDir {
	return k.tempDir
}
