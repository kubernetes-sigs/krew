package integration

import (
	"testing"

	"sigs.k8s.io/krew/integration/krew"
)

const (
	// validPlugin is a valid plugin with a small download size
	validPlugin = "konfig"
)

func TestKrewInstall(t *testing.T) {
	skipShort(t)

	test, cleanup := krew.NewKrewTest(t)
	defer cleanup()

	test.WithIndex().Krew("install", validPlugin).RunOrFailOutput()
	test.Call(validPlugin, "--help").RunOrFail()
}

func TestKrewHelp(t *testing.T) {
	skipShort(t)

	test, cleanup := krew.NewKrewTest(t)
	defer cleanup()

	test.Krew("help").RunOrFail()
}

func skipShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
}
