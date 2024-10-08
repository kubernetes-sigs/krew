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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/installation/receipt"
	"sigs.k8s.io/krew/internal/testutil"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

const (
	persistentIndexCache = "krew-persistent-index-cache"
	krewBinaryEnv        = "KREW_BINARY"
	validPlugin          = "ctx" // a plugin in central index with small size
	validPlugin2         = "ns"  // a plugin in central index with small size
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
func NewTest(t *testing.T) *ITest {
	tempDir := testutil.NewTempDir(t)
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
	}
}

// setupKrewBin symlinks the $KREW_BINARY to $tempDir/bin and returns the path
// to this directory.
func setupKrewBin(t *testing.T, tempDir *testutil.TempDir) string {
	krewBinary, found := os.LookupEnv(krewBinaryEnv)
	if !found {
		t.Fatalf("%s environment variable pointing to krew binary not set", krewBinaryEnv)
	}
	binPath := tempDir.Path("bin")
	if err := os.MkdirAll(binPath, 0o755); err != nil {
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
func skipShort(t *testing.T) { //nolint:gocritic
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

// AssertPluginFromIndex asserts that a receipt exists for the given plugin and
// that it is from the specified index.
func (it *ITest) AssertPluginFromIndex(plugin, indexName string) {
	it.t.Helper()

	receiptPath := environment.NewPaths(it.Root()).PluginInstallReceiptPath(plugin)
	r := it.loadReceipt(receiptPath)
	if r.Status.Source.Name != indexName {
		it.t.Errorf("wanted index '%s', got: '%s'", indexName, r.Status.Source.Name)
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

// WithDefaultIndex initializes the index with the actual krew-index from github/kubernetes-sigs/krew-index.
func (it *ITest) WithDefaultIndex() *ITest {
	it.initializeIndex()
	return it
}

// WithCustomIndexFromDefault initializes a new index by cloning the default index. WithDefaultIndex needs
// to be called before this function. This is a helper function for working with custom indexes in the
// integration tests so that developers don't need to alias the cloned default index each time.
func (it *ITest) WithCustomIndexFromDefault(name string) *ITest {
	indexPath := environment.NewPaths(it.Root()).IndexPath(constants.DefaultIndexName)
	it.Krew("index", "add", name, indexPath).RunOrFail()
	return it
}

// IndexPluginCount returns the number of plugins available in a given index.
func (it *ITest) IndexPluginCount(name string) int {
	indexPath := environment.NewPaths(it.Root()).IndexPluginsPath(name)
	indexDir, err := os.Open(indexPath)
	if err != nil {
		it.t.Fatal(err)
	}
	defer indexDir.Close()
	plugins, err := indexDir.Readdirnames(-1)
	if err != nil {
		it.t.Fatal(err)
	}
	return len(plugins)
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

// Run runs the krew command, and returns its combined output even when
// it fails.
func (it *ITest) Run() ([]byte, error) {
	it.t.Helper()

	cmd := it.cmd(context.Background())
	if it.stdin != nil {
		cmd.Stdin = it.stdin
	}
	it.t.Log(cmd.Args)

	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b

	start := time.Now()
	err := cmd.Run()
	out := b.Bytes()
	it.t.Log("Ran in", time.Since(start))
	return out, errors.Wrapf(err, "krew %v: %v, %s", it.args, err, string(out))
}

// RunOrFail runs the krew command and fails the test if the command returns an error.
func (it *ITest) RunOrFail() {
	it.t.Helper()
	if _, err := it.Run(); err != nil {
		it.t.Fatal(err)
	}
}

// RunOrFailOutput runs the krew command and fails the test if the command
// returns an error. It only returns the standard output.
func (it *ITest) RunOrFailOutput() []byte {
	it.t.Helper()
	out, err := it.Run()
	if err != nil {
		it.t.Fatal(err)
	}
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

func (it *ITest) loadReceipt(path string) index.Receipt {
	pluginReceipt, err := receipt.Load(path)
	if err != nil {
		it.t.Fatalf("error loading receipt: %v", err)
	}
	return pluginReceipt
}

// InitializeIndex initializes the krew index in `$root/index` with the actual krew-index.
// It caches the index tree as in-memory tar after the first run.
func (it *ITest) initializeIndex() {
	initIndexOnce.Do(func() {
		persistentCacheFile := filepath.Join(os.TempDir(), persistentIndexCache)
		fileInfo, err := os.Stat(persistentCacheFile)

		if err == nil && fileInfo.Mode().IsRegular() {
			it.t.Logf("Using persistent index cache from file %q", persistentCacheFile)
			if indexTar, err = os.ReadFile(persistentCacheFile); err == nil {
				return
			}
		}

		if indexTar, err = initFromGitClone(it.t); err != nil {
			it.t.Fatalf("cannot clone repository: %s", err)
		}

		if err = os.WriteFile(persistentCacheFile, indexTar, 0o600); err != nil {
			it.t.Fatalf("cannot write persistent cache file: %s", err)
		}
	})

	indexDir := filepath.Join(it.Root(), "index", "default")
	if err := os.MkdirAll(indexDir, 0o777); err != nil {
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
}

func initFromGitClone(t *testing.T) ([]byte, error) {
	const tarName = "index.tar"
	indexDir := testutil.NewTempDir(t)
	indexRoot := indexDir.Root()

	cmd := exec.Command("git", "clone", "--depth=1", "--single-branch", "--no-tags", constants.DefaultIndexURI)
	cmd.Dir = indexRoot
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	cmd = exec.Command("tar", "czf", tarName, "-C", "krew-index", ".")
	cmd.Dir = indexRoot
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return os.ReadFile(filepath.Join(indexRoot, tarName))
}
