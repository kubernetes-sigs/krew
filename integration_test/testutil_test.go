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

package integrationtest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/indexmigration"
	"sigs.k8s.io/krew/internal/testutil"
	"sigs.k8s.io/krew/pkg/constants"
)

const (
	persistentIndexCache = "krew-persistent-index-cache"
	krewBinaryEnv        = "KREW_BINARY"
	validPlugin          = "ctx"   // a plugin in central index with small size
	validPlugin2         = "mtail" // a plugin in central index with small size
)

var (
	initIndexOnce sync.Once
	indexTar      []byte
)

// ITest is used to set up `krew` integration tests.
type ITest struct {
	t             *testing.T
	plugin        string
	pluginsBinDir string
	args          []string
	env           []string
	stdin         io.Reader
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
			"KREW_OS=linux",
			"KREW_ARCH=amd64",
			"KREW_NO_UPGRADE_CHECK=1",
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

	// todo(corneliusweig): when the receipts migration logic is removed, this becomes obsolete
	tempDir.Write("receipts/krew.notyaml", []byte("must be present for receipts migration check"))

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

// skipShort is a test helper for skipping tests in -test.short runs.
func skipShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
}

// lines parses command outputs into separate lines while trimming the trailing
// newline.
func lines(in []byte) []string {
	trimmed := strings.TrimRight(string(in), " \t\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

func (it *ITest) LookupExecutable(file string) (string, error) {
	orig := os.Getenv("PATH")
	defer func() { os.Setenv("PATH", orig) }()

	binPath := filepath.Join(it.Root(), "bin")
	os.Setenv("PATH", binPath)

	return exec.LookPath(file)
}

// AssertExecutableInPATH asserts that the executable file is in bin path.
func (it *ITest) AssertExecutableInPATH(file string) {
	it.t.Helper()
	if _, err := it.LookupExecutable(file); err != nil {
		it.t.Fatalf("executable %s not in PATH: %+v", file, err)
	}
}

// AssertExecutableNotInPATH asserts that the executable file is not in bin
// path.
func (it *ITest) AssertExecutableNotInPATH(file string) {
	it.t.Helper()
	if _, err := it.LookupExecutable(file); err == nil {
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
		it.t.Fatal("Overriding KREW_ROOT in tests is forbidden")
		return it
	}
	it.env = append(it.env, fmt.Sprintf("%s=%v", key, value))
	return it
}

// WithStdin sets standard input for the krew command execution.
func (it *ITest) WithStdin(r io.Reader) *ITest {
	it.stdin = r
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
	it.t.Log(cmd.Args)

	start := time.Now()
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "krew %v", it.args)
	}

	it.t.Log("Ran in", time.Since(start))
	return nil
}

// RunOrFailOutput runs the krew command and fails the test if the command
// returns an error. It only returns the standard output.
func (it *ITest) RunOrFailOutput() []byte {
	it.t.Helper()

	cmd := it.cmd(context.Background())
	if it.stdin != nil {
		cmd.Stdin = it.stdin
	}
	it.t.Log(cmd.Args)

	start := time.Now()
	out, err := cmd.CombinedOutput()
	if err != nil {
		it.t.Fatalf("krew %v: %v, %s", it.args, err, out)
	}

	it.t.Log("Ran in", time.Since(start))
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

// InitializeIndex initializes the krew index in `$root/index` with the actual krew-index.
// It caches the index tree as in-memory tar after the first run.
func (it *ITest) initializeIndex() {
	initIndexOnce.Do(func() {
		persistentCacheFile := filepath.Join(os.TempDir(), persistentIndexCache)
		fileInfo, err := os.Stat(persistentCacheFile)

		if err == nil && fileInfo.Mode().IsRegular() {
			it.t.Logf("Using persistent index cache from file %q", persistentCacheFile)
			if indexTar, err = ioutil.ReadFile(persistentCacheFile); err == nil {
				return
			}
		}

		if indexTar, err = initFromGitClone(it.t); err != nil {
			it.t.Fatalf("cannot clone repository: %s", err)
		}

		if err = ioutil.WriteFile(persistentCacheFile, indexTar, 0600); err != nil {
			it.t.Fatalf("cannot write persistent cache file: %s", err)
		}
	})

	indexDir := filepath.Join(it.Root(), "index")
	if err := os.Mkdir(indexDir, 0777); err != nil {
		if os.IsExist(err) {
			it.t.Log("initializeIndex should only be called once")
			return
		}
		it.t.Fatal(err)
	}

	cmd := exec.Command("tar", "xzf", "-", "-C", indexDir)
	cmd.Stdin = bytes.NewReader(indexTar)
	if err := cmd.Run(); err != nil {
		it.t.Fatalf("cannot restore index from cache: %s", err)
	}

	// TODO(chriskim06) simplify once multi-index is enabled
	for _, e := range it.env {
		if strings.Contains(e, constants.EnableMultiIndexSwitch) {
			if err := indexmigration.Migrate(environment.NewPaths(it.Root())); err != nil {
				it.t.Fatalf("error migrating index: %s", err)
			}
		}
	}
}

func initFromGitClone(t *testing.T) ([]byte, error) {
	const tarName = "index.tar"
	indexDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()
	indexRoot := indexDir.Root()

	cmd := exec.Command("git", "clone", "--depth=1", "--single-branch", "--no-tags", constants.IndexURI)
	cmd.Dir = indexRoot
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	cmd = exec.Command("tar", "czf", tarName, "-C", "krew-index", ".")
	cmd.Dir = indexRoot
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return ioutil.ReadFile(filepath.Join(indexRoot, tarName))
}
