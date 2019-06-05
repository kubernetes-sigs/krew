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
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	krewTest, cleanup := krew.NewKrewTest(t)
	defer cleanup()

	krewTest.WithIndex().Cmd("install", validPlugin).RunOrFail()
}

func TestUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	krewTest, cleanup := krew.NewKrewTest(t)
	defer cleanup()

	krewTest.WithIndex().Cmd("update").RunOrFail()

	indexFiles, err := krewTest.TempDir().List("index")
	if err != nil {
		t.Error(err)
	}

	if len(indexFiles) == 0 {
		t.Error("expected some index files but found none")
	}
}

func TestUninstall(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	krewTest, cleanup := krew.NewKrewTest(t)
	defer cleanup()

	krewTest.WithIndex().Cmd("install", validPlugin).RunOrFailOutput()
	krewTest.Cmd("remove", validPlugin).RunOrFailOutput()

	indexFiles, err := krewTest.TempDir().List("store")
	if err != nil {
		t.Error(err)
	}

	if len(indexFiles) != 0 {
		t.Error("expected the store to be empty")
	}
}
