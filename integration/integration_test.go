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

	krewTest, cleanup := krew.NewKrewTest(t)
	defer cleanup()

	krewTest.WithIndex().Cmd("install", validPlugin).RunOrFailOutput()
	// todo(corneliusweig): make sure that the plugin can be executed as `kubectl konfig --help`
}

func TestKrewHelp(t *testing.T) {
	skipShort(t)

	krewTest, cleanup := krew.NewKrewTest(t)
	defer cleanup()

	krewTest.Cmd("help").RunOrFail()
}

func skipShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
}
