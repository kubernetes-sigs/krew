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

package validation

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/krew/internal/testutil"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

func Test_IsSafePluginName(t *testing.T) {
	tests := []struct {
		name       string
		pluginName string
		want       bool
	}{
		{
			name:       "secure name",
			pluginName: "foo-bar",
			want:       true,
		},
		{
			name:       "insecure path name",
			pluginName: "/foo-bar",
			want:       false,
		},
		{
			name:       "relative name",
			pluginName: "..foo-bar",
			want:       false,
		},
		{
			name:       "bad windows name",
			pluginName: "nul",
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSafePluginName(tt.pluginName); got != tt.want {
				t.Errorf("IsSafePluginName(%s) = %v, want %v", tt.pluginName, got, tt.want)
			}
		})
	}
}

func Test_isSupportedAPIVersion(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"wrong group", "networking.k8s.io/v1", false},
		{"just api group", "krew.googlecontainertools.github.com", false},
		{"old version", "krew.googlecontainertools.github.com/v1alpha1", false},
		{"equal version", "krew.googlecontainertools.github.com/v1alpha2", true},
		{"newer 1", "krew.googlecontainertools.github.com/v1alpha3", false},
		{"newer 2", "krew.googlecontainertools.github.com/v1", false},
		{"newer 2", "krew.googlecontainertools.github.com/v2alpha1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSupportedAPIVersion(tt.in); got != tt.want {
				t.Errorf("isSupportedAPIVersion(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestValidatePlugin(t *testing.T) {
	tests := []struct {
		name       string
		plugin     index.Plugin
		pluginName string
		wantErr    bool
	}{
		{
			name:       "success",
			pluginName: "foo",
			plugin:     testutil.NewPlugin().WithName("foo").V(),
			wantErr:    false,
		},
		{
			name:       "file name mismatch",
			pluginName: "orange",
			plugin:     testutil.NewPlugin().WithName("apple").V(),
			wantErr:    true,
		},
		{
			name:       "bad api version",
			pluginName: "foo",
			plugin: testutil.NewPlugin().WithName("foo").WithTypeMeta(metav1.TypeMeta{
				APIVersion: "core/v1",
				Kind:       constants.PluginKind,
			}).V(),
			wantErr: true,
		},
		{
			name:       "wrong kind",
			pluginName: "foo",
			plugin: testutil.NewPlugin().WithName("foo").WithTypeMeta(
				metav1.TypeMeta{
					APIVersion: constants.CurrentAPIVersion,
					Kind:       "Not" + constants.PluginKind,
				}).V(),
			wantErr: true,
		},
		{
			name:       "shortDescription unspecified",
			pluginName: "foo",
			plugin:     testutil.NewPlugin().WithName("foo").WithShortDescription("").V(),
			wantErr:    true,
		},
		{
			name:       "version unspecified",
			pluginName: "foo",
			plugin:     testutil.NewPlugin().WithName("foo").WithVersion("").V(),
			wantErr:    true,
		},
		{
			name:       "version malformed",
			pluginName: "foo",
			plugin:     testutil.NewPlugin().WithName("foo").WithVersion("v01.02.3-a").V(),
			wantErr:    true,
		},
		{
			name:       "no platform specified",
			pluginName: "foo",
			plugin:     testutil.NewPlugin().WithName("foo").WithPlatforms().V(),
			wantErr:    true,
		},
		{
			name:       "empty file operations",
			pluginName: "foo",
			plugin: testutil.NewPlugin().WithName("foo").WithPlatforms(
				testutil.NewPlatform().WithFiles([]index.FileOperation{}).V()).V(),
			wantErr: true,
		},
		{
			name:       "unsafe plugin name",
			pluginName: "../foo",
			plugin:     testutil.NewPlugin().WithName("../foo").V(),
			wantErr:    true,
		},
		{
			name:       "short description with line break",
			pluginName: "foo",
			plugin:     testutil.NewPlugin().WithShortDescription("just foo\n").V(),
			wantErr:    true,
		},
		{
			name:       "short description with carriage return",
			pluginName: "foo",
			plugin:     testutil.NewPlugin().WithShortDescription("just foo\r").V(),
			wantErr:    true,
		},
		{
			name:       "short description with CRLF",
			pluginName: "foo",
			plugin:     testutil.NewPlugin().WithShortDescription("just\r\nfoo").V(),
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidatePlugin(tt.pluginName, tt.plugin); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePlugin(%s) error = %v, wantErr %v", tt.pluginName, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePlatform(t *testing.T) {
	tests := []struct {
		name     string
		platform index.Platform
		wantErr  bool
	}{
		{
			name:     "success",
			platform: testutil.NewPlatform().V(),
			wantErr:  false,
		},
		{
			name:     "missing url",
			platform: testutil.NewPlatform().WithURI("").V(),
			wantErr:  true,
		},
		{
			name:     "missing sha256",
			platform: testutil.NewPlatform().WithSHA256("").V(),
			wantErr:  true,
		},
		{
			name:     "empty file operations",
			platform: testutil.NewPlatform().WithFiles([]index.FileOperation{}).V(),
			wantErr:  true,
		},
		{
			name:     "no bin field",
			platform: testutil.NewPlatform().WithBin("").V(),
			wantErr:  true,
		},
		{
			name: "invalid platform selector",
			platform: testutil.NewPlatform().WithSelector(&metav1.LabelSelector{
				MatchLabels: map[string]string{"unsupported-field": "orange"}}).V(),
			wantErr: true,
		},
		// TODO(ahmetb): add test case "bin field outside the plugin installation directory"
		// by testing .WithBin("foo/../../../malicious-file").
		// It appears like currently we're allowing this.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePlatform(tt.platform); (err != nil) != tt.wantErr {
				t.Errorf("validatePlatform() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateSelector(t *testing.T) {
	var tests = []struct {
		name    string
		sel     *metav1.LabelSelector
		wantErr bool
	}{
		{
			name:    "nil selector",
			sel:     nil,
			wantErr: true,
		},
		{
			name:    "empty (wildcard) selector",
			sel:     &metav1.LabelSelector{},
			wantErr: true,
		},
		{
			name:    "valid matchLabels",
			sel:     &metav1.LabelSelector{MatchLabels: map[string]string{"os": "foo", "arch": "bar"}},
			wantErr: false,
		},
		{
			name: "valid matchExpressions",
			sel: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{Key: "os",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"apple", "orange"},
					}}},
			wantErr: false,
		},
		{
			name:    "empty matchLabels",
			sel:     &metav1.LabelSelector{MatchLabels: map[string]string{}},
			wantErr: true,
		},
		{
			name:    "empty matchExpressions",
			sel:     &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{}},
			wantErr: true,
		},
		{
			name:    "unsupported key in matchLabels",
			sel:     &metav1.LabelSelector{MatchLabels: map[string]string{"unsupported-key": "value"}},
			wantErr: true,
		},
		{
			name: "unsupported key in matchExpressions",
			sel: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{Key: "unsupported-key",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"apple", "orange"}}}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateSelector(tt.sel); (err != nil) != tt.wantErr {
				t.Errorf("validateSelector() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateFiles(t *testing.T) {
	tests := []struct {
		name    string
		files   []index.FileOperation
		wantErr bool
	}{
		{
			name:    "success",
			files:   []index.FileOperation{{From: "here", To: "there"}},
			wantErr: false,
		},
		{
			name:    "unspecified file operations",
			files:   nil,
			wantErr: false,
		},
		{
			name:    "empty file operations",
			files:   []index.FileOperation{},
			wantErr: true,
		},
		{
			name:    "empty `to` field in file operations",
			files:   []index.FileOperation{{From: "present", To: ""}},
			wantErr: true,
		},
		{
			name:    "empty `from` field in file operations",
			files:   []index.FileOperation{{From: "", To: "present"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateFiles(tt.files); (err != nil) != tt.wantErr {
				t.Errorf("validateFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
