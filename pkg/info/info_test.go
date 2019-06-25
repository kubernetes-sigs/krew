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

package info

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/environment"
	"sigs.k8s.io/krew/pkg/index"
	"sigs.k8s.io/krew/pkg/testutil"
)

const pluginName = "some"

var (
	plugin = index.Plugin{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.CurrentAPIVersion,
			Kind:       constants.PluginKind,
		},
		ObjectMeta: metav1.ObjectMeta{Name: pluginName},
		Spec: index.PluginSpec{
			Version:          "",
			ShortDescription: "short",
			Description:      "",
			Caveats:          "",
			Homepage:         "",
			Platforms: []index.Platform{{
				URI:      "http://example.com",
				Sha256:   "deadbeef",
				Selector: nil,
				Files:    []index.FileOperation{{"", ""}},
				Bin:      "foo",
			}},
		},
	}
)

func TestLoadManifestFromReceiptOrIndex(t *testing.T) {
	yamlBytes, err := yaml.Marshal(plugin)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		prepare   func(environment.Paths, *testutil.TempDir)
		shouldErr bool
	}{
		{
			name: "manifest in receipts",
			prepare: func(paths environment.Paths, tmpDir *testutil.TempDir) {
				path := paths.PluginReceiptPath(pluginName)
				tmpDir.Write(path, yamlBytes)
			},
		},
		{
			name: "manifest in index",
			prepare: func(paths environment.Paths, tmpDir *testutil.TempDir) {
				path := filepath.Join(paths.IndexPluginsPath(), pluginName+".yaml")
				tmpDir.Write(path, yamlBytes)
			},
		},
		{
			name: "invalid manifest in receipts",
			prepare: func(paths environment.Paths, tmpDir *testutil.TempDir) {
				path := paths.PluginReceiptPath(pluginName)
				tmpDir.Write(path, []byte("invalid yaml file"))
			},
			shouldErr: true,
		},
		{
			name: "invalid manifest in index",
			prepare: func(paths environment.Paths, tmpDir *testutil.TempDir) {
				path := filepath.Join(paths.IndexPluginsPath(), pluginName+".yaml")
				tmpDir.Write(path, []byte("invalid yaml file"))
			},
			shouldErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.NewTempDir(t)
			defer cleanupEnvAndTempdir(t, os.Getenv("KREW_ROOT"), cleanup)
			if err := os.Setenv("KREW_ROOT", tmpDir.Root()); err != nil {
				t.Fatal(err)
			}

			paths := environment.MustGetKrewPaths()
			test.prepare(paths, tmpDir)
			actual, err := LoadManifestFromReceiptOrIndex(paths, pluginName)

			if test.shouldErr {
				if err == nil {
					t.Error("LoadManifestFromReceiptOrIndex expected an error but found none")
				}
			} else {
				if err != nil {
					t.Error(err)
				}
				if diff := cmp.Diff(&plugin, &actual); diff != "" {
					t.Error(diff)
				}
			}
		})
	}
}

func TestLoadManifestFromReceiptOrIndexReturnsIsNotExist(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanupEnvAndTempdir(t, os.Getenv("KREW_ROOT"), cleanup)
	if err := os.Setenv("KREW_ROOT", tmpDir.Root()); err != nil {
		t.Fatal(err)
	}

	paths := environment.MustGetKrewPaths()
	_, err := LoadManifestFromReceiptOrIndex(paths, pluginName)

	if err == nil {
		t.Fatalf("Expected LoadManifestFromReceiptOrIndex to fail")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Expected error to be ENOENT but was %q", err)
	}
}

func cleanupEnvAndTempdir(t *testing.T, originalKrewRoot string, cleanupTempdir func()) {
	cleanupTempdir()
	var err error
	if originalKrewRoot == "" {
		err = os.Unsetenv("KREW_ROOT")
	} else {
		err = os.Setenv("KREW_ROOT", originalKrewRoot)
	}
	if err != nil {
		t.Log(err)
	}
}
