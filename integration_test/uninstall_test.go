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

import "testing"

func TestKrewUninstall(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test = test.WithIndex()

	if err := test.Krew("uninstall").Run(); err == nil {
		t.Fatal("expected failure without no arguments")
	}
	if err := test.Krew("uninstall", validPlugin).Run(); err == nil {
		t.Fatal("expected failure deleting non-installed plugin")
	}
	test.Krew("install", validPlugin).RunOrFailOutput()
	test.Krew("uninstall", validPlugin).RunOrFailOutput()
	test.AssertExecutableNotInPATH("kubectl-" + validPlugin)

	if err := test.Krew("uninstall", validPlugin).Run(); err == nil {
		t.Fatal("expected failure for uninstalled plugin")
	}
}

func TestKrewRemove_AliasSupported(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.WithIndex().Krew("install", validPlugin).RunOrFailOutput()
	test.Krew("remove", validPlugin).RunOrFailOutput()
	test.AssertExecutableNotInPATH("kubectl-" + validPlugin)
}
