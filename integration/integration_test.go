package integration

import (
	"testing"

	"sigs.k8s.io/krew/integration/krew"
)

const (
	// validPlugin is a valid plugin with a small download size
	validPlugin = "konfig"
)

func TestInstall(t *testing.T) {
	skipShort(t)

	krewTest, cleanup := krew.NewKrewTest(t)
	defer cleanup()

	krewTest.WithIndex().Cmd("install", validPlugin).RunOrFail()
	// todo(corneliusweig): make sure that the plugin can be executed as `kubectl konfig --help`
}

func skipShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
}
