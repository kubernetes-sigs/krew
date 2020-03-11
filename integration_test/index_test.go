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
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/krew/pkg/constants"
)

func TestKrewIndex(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	initialList := test.WithMigratedIndex().WithEnv(constants.EnableMultiIndexSwitch, 1).Krew("index", "list").RunOrFailOutput()
	if !bytes.Contains(initialList, []byte(constants.IndexURI)) {
		t.Fatalf("expected krew-index in output:\n%s", string(initialList))
	}

	indexName := "foo"
	test.Krew("index", "add", indexName, constants.IndexURI).RunOrFail()
	if _, err := os.Stat(filepath.Join(test.Root(), "index", indexName, ".git")); err != nil {
		t.Fatalf("error adding index: %s", err)
	}

	eventualList := test.Krew("index", "list").RunOrFailOutput()
	if !bytes.Contains(eventualList, []byte(indexName)) {
		t.Fatalf("expected index 'foo' in output:\n%s", string(eventualList))
	}
}
