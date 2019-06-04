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

// KrewTest is used to set up `krew` integration tests.
type KrewTest struct {
	t       *testing.T
	args    []string
	env     []string
	tempDir *testutil.TempDir
}

// NewKrewTest creates a fluent krew KrewTest
func NewKrewTest(t *testing.T) (*KrewTest, func()) {
	tempDir, cleanup := testutil.NewTempDir(t)
	return &KrewTest{
		t:       t,
		env:     []string{fmt.Sprintf("KREW_ROOT=%s", tempDir.Root())},
		tempDir: tempDir,
	}, cleanup
}

func (k *KrewTest) Root() string {
	return k.tempDir.Root()
}

// Cmd sets the arguments to krew
func (k *KrewTest) Cmd(args ...string) *KrewTest {
	k.args = args
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
	cmd.Stdout, cmd.Stderr = nil, nil
	glog.V(1).Infoln(cmd.Args)

	start := time.Now()
	out, err := cmd.Output()
	if err != nil {
		k.t.Fatalf("krew %v: %v, %s", k.args, err, out)
	}

	glog.V(1).Infoln("Ran in", time.Since(start))
	return out
}

func (k *KrewTest) cmd(ctx context.Context) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "krew", k.args...)
	cmd.Env = append(os.Environ(), k.env...)

	return cmd
}
