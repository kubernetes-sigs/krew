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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/pkg/constants"
)

func TestKrewUpgrade_WithoutIndexInitialized(t *testing.T) {
	skipShort(t)

	test := NewTest(t)
	test.Krew("upgrade").RunOrFailOutput()
}

func TestKrewUpgrade(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex().
		Krew("install", "--manifest", filepath.Join("testdata", validPlugin+constants.ManifestExtension)).
		RunOrFail()

	// plugins installed via manifest get the special "detached" index so this needs to
	// be changed to default in order for it to be upgraded here
	receiptPath := environment.NewPaths(test.Root()).PluginInstallReceiptPath(validPlugin)
	modifyReceiptIndex(t, receiptPath, "default")
	initialCreationTimestamp := test.loadReceipt(receiptPath).CreationTimestamp

	initialLocation := resolvePluginSymlink(test, validPlugin)
	test.Krew("upgrade").RunOrFail()
	eventualLocation := resolvePluginSymlink(test, validPlugin)
	if initialLocation == eventualLocation {
		t.Errorf("Expecting the plugin path to change but was the same.")
	}

	eventualCreationTimestamp := test.loadReceipt(receiptPath).CreationTimestamp
	if initialCreationTimestamp != eventualCreationTimestamp {
		t.Errorf("expected the receipt creationTimestamp to remain unchanged after upgrade")
	}
}

func TestKrewUpgradePluginsFromCustomIndex(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex().WithCustomIndexFromDefault("foo")
	test.Krew("install", "foo/"+validPlugin).RunOrFail()

	receiptPath := environment.NewPaths(test.Root()).PluginInstallReceiptPath(validPlugin)
	modifyManifestVersion(t, receiptPath, "v0.0.0")
	out := string(test.Krew("upgrade").RunOrFailOutput())
	if !strings.Contains(out, "Upgrading plugin: foo/"+validPlugin) {
		t.Errorf("expected plugin foo/%s to be upgraded", validPlugin)
	}

	modifyManifestVersion(t, receiptPath, "v0.0.0")
	out = string(test.Krew("upgrade", validPlugin).RunOrFailOutput())
	if !strings.Contains(out, "Upgrading plugin: foo/"+validPlugin) {
		t.Errorf("expected plugin foo/%s to be upgraded", validPlugin)
	}
}

func TestKrewUpgradeSkipsManifestPlugin(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex().
		Krew("install", "--manifest", filepath.Join("testdata", validPlugin+constants.ManifestExtension)).
		RunOrFail()

	out := string(test.Krew("upgrade").RunOrFailOutput())
	if !strings.Contains(out, "Skipping upgrade") {
		t.Errorf("expected plugin %q to be skipped during upgrade", validPlugin)
	}
}

func TestKrewUpgradeNoSecurityWarningForCustomIndex(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex().WithCustomIndexFromDefault("foo")
	test.Krew("install", "foo/"+validPlugin).RunOrFail()

	pluginReceipt := environment.NewPaths(test.Root()).PluginInstallReceiptPath(validPlugin)
	modifyManifestVersion(t, pluginReceipt, "v0.0.1")
	out := string(test.Krew("upgrade").RunOrFailOutput())
	if strings.Contains(out, "Run them at your own risk") {
		t.Errorf("expected install of custom plugin to not show security warning: %v", out)
	}
}

func TestKrewUpgrade_CannotUseIndexSyntax(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	b, err := test.Krew("upgrade", "foo/"+validPlugin).Run()
	if err == nil {
		t.Error("expected error when using canonical name with upgrade")
	}
	if !strings.Contains(string(b), "INDEX/PLUGIN") {
		t.Error("expected warning about using canonical name to be in output")
	}
}

func TestKrewUpgradeUnsafe(t *testing.T) {
	skipShort(t)
	test := NewTest(t)
	test.WithDefaultIndex()

	cases := []string{
		`../index/` + validPlugin,
		`..\index\` + validPlugin,
		`../default/` + validPlugin,
		`..\default\` + validPlugin,
		`index-name/sub-directory/plugin-name`,
	}

	expectedErr := `not allowed`
	for _, c := range cases {
		b, err := test.Krew("upgrade", c).Run()
		if err == nil {
			t.Fatalf("%q expected failure", c)
		} else if !strings.Contains(string(b), expectedErr) {
			t.Fatalf("%q does not contain err %q: %q", c, expectedErr, string(b))
		}
	}
}

func TestKrewUpgradeWhenPlatformNoLongerMatches(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex().
		Krew("install", validPlugin).
		RunOrFail()

	test.WithEnv("KREW_OS", "somethingelse")

	// if upgrading 'all' plugins, must succeed
	out := string(test.Krew("upgrade", "--no-update-index").RunOrFailOutput())
	if !strings.Contains(out, "WARNING: Some plugins failed to upgrade") {
		t.Fatalf("upgrade all plugins output doesn't contain warnings about failed plugins:\n%s", out)
	}

	// if upgrading a specific plugin, it must fail, because no longer matching to a platform
	_, err := test.Krew("upgrade", validPlugin, "--no-update-index").Run()
	if err == nil {
		t.Fatal("expected failure when upgraded a specific plugin that no longer has a matching platform")
	}
}

func TestKrewUpgrade_ValidPluginInstalledFromManifest(t *testing.T) {
	skipShort(t)

	test := NewTest(t)

	test.WithDefaultIndex().
		Krew("install", validPlugin).
		RunOrFail()

	pluginPath := environment.NewPaths(test.Root()).IndexPluginsPath(constants.DefaultIndexName)
	validPluginPath := filepath.Join(pluginPath, validPlugin+constants.ManifestExtension)
	if err := os.Remove(validPluginPath); err != nil {
		t.Fatalf("can't remove valid plugin from index: %q", validPluginPath)
	}

	// if upgrading 'all' plugins, must succeed
	out := string(test.Krew("upgrade", "--no-update-index").RunOrFailOutput())
	if !strings.Contains(out, "WARNING: Some plugins failed to upgrade") {
		t.Fatalf("upgrade all plugins output doesn't contain warnings about failed plugins:\n%s", out)
	}

	// if upgrading a specific plugin, it must fail, because it's not included into index
	_, err := test.Krew("upgrade", validPlugin, "--no-update-index").Run()
	if err == nil {
		t.Fatal("expected failure when upgraded a specific plugin that is not included in index")
	}
}

func resolvePluginSymlink(test *ITest, plugin string) string {
	test.t.Helper()
	linkToPlugin, err := test.LookupExecutable("kubectl-" + plugin)
	if err != nil {
		test.t.Fatal(err)
	}

	realLocation, err := os.Readlink(linkToPlugin)
	if err != nil {
		test.t.Fatal(err)
	}

	return realLocation
}

func modifyReceiptIndex(t *testing.T, file, index string) {
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	r := regexp.MustCompile(`(?m)(\bstatus:\n\s+source:\n\s+name:\s)(.*)$`) // patch index name
	b = r.ReplaceAll(b, []byte(fmt.Sprintf("${1}%s", index)))
	if err := os.WriteFile(file, b, 0); err != nil {
		t.Fatal(err)
	}
}
