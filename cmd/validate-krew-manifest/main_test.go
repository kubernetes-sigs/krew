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

// validate-krew-manifest makes sure a manifest file is valid.s
package main

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/krew/pkg/index"
	"sigs.k8s.io/krew/pkg/testutil"
)

func TestValidateManifestFile(t *testing.T) {
	tests := []struct {
		name          string
		manifestFile  string
		writeManifest bool
		shouldErr     bool
		errMsg        string
		plugin        index.Plugin
	}{
		{
			name:          "manifest does not exist",
			manifestFile:  "test.yaml",
			writeManifest: false,
			shouldErr:     true,
			errMsg:        "failed to read plugin file",
		},
		{
			name:          "manifest has wrong file ending",
			manifestFile:  "test.yml",
			writeManifest: true,
			plugin:        testutil.NewPlugin().WithName("test").V(),
			shouldErr:     true,
			errMsg:        "expected manifest extension \".yaml\"",
		},
		{
			name:          "manifest validation fails (name mismatch)",
			manifestFile:  "foo.yaml",
			writeManifest: true,
			plugin:        testutil.NewPlugin().WithName("not-foo").V(),
			shouldErr:     true,
			errMsg:        "plugin validation error",
		},
		{
			name:          "architecture selector not supported",
			manifestFile:  "test.yaml",
			writeManifest: true,
			plugin: testutil.NewPlugin().WithName("test").WithPlatforms(
				testutil.NewPlatform().WithOSArch("darwin", "arm").V()).V(),
			shouldErr: true,
			errMsg:    "doesn't match any supported platforms",
		},
		{
			name:          "overlapping platform selectors",
			manifestFile:  "test.yaml",
			writeManifest: true,
			plugin: testutil.NewPlugin().WithName("test").WithPlatforms(
				testutil.NewPlatform().WithOS("linux").V(),
				testutil.NewPlatform().WithOSArch("linux", "amd64").V()).V(),
			shouldErr: true,
			errMsg:    "overlapping platform selectors found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmp, cleanup := testutil.NewTempDir(t)
			defer cleanup()

			if test.writeManifest {
				content, err := yaml.Marshal(test.plugin)
				if err != nil {
					t.Fatal(err)
				}
				tmp.Write(test.manifestFile, content)
			}

			err := validateManifestFile(tmp.Path(test.manifestFile))
			if test.shouldErr {
				if err == nil {
					t.Errorf("Expected an error '%s' but found none", test.errMsg)
				} else if !strings.Contains(err.Error(), test.errMsg) {
					t.Errorf("Error '%s' should contain error message '%s'", err, test.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but found '%s'", err)
				}
			}
		})
	}
}

func Test_selectorMatchesOSArch(t *testing.T) {
	type args struct {
		selector *metav1.LabelSelector
		os       string
		arch     string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"label - no match", args{&metav1.LabelSelector{MatchLabels: map[string]string{"os": "darwin"}}, "windows", "amd64"}, false},
		{"label - match", args{&metav1.LabelSelector{MatchLabels: map[string]string{"os": "darwin"}}, "darwin", "amd64"}, true},
		{"expression - no match", args{&metav1.LabelSelector{MatchExpressions: []v1.LabelSelectorRequirement{{
			Key:      "os",
			Operator: v1.LabelSelectorOpIn,
			Values:   []string{"darwin", "linux"},
		}}}, "windows", "amd64"}, false},
		{"expression - match", args{&metav1.LabelSelector{MatchExpressions: []v1.LabelSelectorRequirement{{
			Key:      "os",
			Operator: v1.LabelSelectorOpIn,
			Values:   []string{"darwin", "linux"},
		}}}, "darwin", "amd64"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectorMatchesOSArch(tt.args.selector, tt.args.os, tt.args.arch); got != tt.want {
				t.Errorf("selectorMatchesOSArch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findAnyMatchingPlatform(t *testing.T) {
	s1 := &v1.LabelSelector{MatchLabels: map[string]string{"os": "darwin"}}
	o1, a1 := findAnyMatchingPlatform(s1)
	if o1 == "" || a1 == "" {
		t.Fatalf("with selector %v, expected os/arch", s1)
	}

	s2 := &v1.LabelSelector{MatchLabels: map[string]string{"os": "non-existing"}}
	o2, a2 := findAnyMatchingPlatform(s1)
	if o2 == "" && a2 == "" {
		t.Fatalf("with selector %v, expected os/arch", s2)
	}

	s3 := &v1.LabelSelector{MatchExpressions: []v1.LabelSelectorRequirement{{
		Key:      "os",
		Operator: v1.LabelSelectorOpIn,
		Values:   []string{"darwin", "linux"}}}}
	o3, a3 := findAnyMatchingPlatform(s3)
	if o3 == "" || a3 == "" {
		t.Fatalf("with selector %v, expected os/arch", s2)
	}
}

func Test_isOverlappingPlatformSelectors_noOverlap(t *testing.T) {
	p1 := testutil.NewPlatform().WithOSes("darwin", "linux").V()
	p2 := testutil.NewPlatform().WithOSes("windows").V()

	err := isOverlappingPlatformSelectors([]index.Platform{p1, p2})
	if err != nil {
		t.Fatalf("expected no overlap: %+v", err)
	}
}

func Test_isOverlappingPlatformSelectors_overlap(t *testing.T) {
	p1 := testutil.NewPlatform().WithOS("darwin").V()
	p2 := testutil.NewPlatform().WithOSes("darwin", "linux").V()
	err := isOverlappingPlatformSelectors([]index.Platform{p1, p2})
	if err == nil {
		t.Fatal("expected overlap")
	}
}
