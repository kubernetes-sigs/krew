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

package index

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/krew/pkg/constants"
)

var (
	defaultMeta = metav1.TypeMeta{
		APIVersion: constants.CurrentAPIVersion,
		Kind:       constants.PluginKind,
	}
)

func Test_IsSafePluginName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "secure name",
			args: args{
				name: "foo-bar",
			},
			want: true,
		},
		{
			name: "insecure path name",
			args: args{
				name: "/foo-bar",
			},
			want: false,
		},
		{
			name: "relative name",
			args: args{
				name: "..foo-bar",
			},
			want: false,
		},
		{
			name: "bad windows name",
			args: args{
				name: "nul",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSafePluginName(tt.args.name); got != tt.want {
				t.Errorf("IsSafePluginName() = %v, want %v", got, tt.want)
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

func TestPlugin_Validate(t *testing.T) {
	type fields struct {
		TypeMeta   metav1.TypeMeta
		ObjectMeta metav1.ObjectMeta
		Spec       PluginSpec
	}
	tests := []struct {
		name       string
		fields     fields
		pluginName string
		wantErr    bool
	}{
		{
			name: "success",
			fields: fields{
				TypeMeta:   defaultMeta,
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: PluginSpec{
					Version:          "v1.0.0",
					ShortDescription: "short",
					Platforms: []Platform{{
						URI:      "http://example.com",
						Sha256:   "deadbeef",
						Selector: nil,
						Files:    []FileOperation{{"", ""}},
						Bin:      "foo",
					}},
				},
			},
			pluginName: "foo",
			wantErr:    false,
		},
		{
			name: "bad api version",
			fields: fields{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "core/v1",
					Kind:       constants.PluginKind,
				},
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: PluginSpec{
					Version:          "v1.0.0",
					ShortDescription: "short",
					Platforms: []Platform{{
						URI:      "http://example.com",
						Sha256:   "deadbeef",
						Selector: nil,
						Files:    []FileOperation{{"", ""}},
						Bin:      "foo",
					}},
				},
			},
			pluginName: "foo",
			wantErr:    true,
		},
		{
			name: "wrong kind",
			fields: fields{
				TypeMeta: metav1.TypeMeta{
					APIVersion: constants.CurrentAPIVersion,
					Kind:       "not-Plugin",
				},
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: PluginSpec{
					Version:          "v1.0.0",
					ShortDescription: "short",
					Platforms: []Platform{{
						URI:      "http://example.com",
						Sha256:   "deadbeef",
						Selector: nil,
						Files:    []FileOperation{{"", ""}},
						Bin:      "foo",
					}},
				},
			},
			pluginName: "foo",
			wantErr:    true,
		},
		{
			name: "shortDescription unspecified",
			fields: fields{
				TypeMeta:   defaultMeta,
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: PluginSpec{
					Version:          "v1.0.0",
					ShortDescription: "",
					Platforms: []Platform{{
						URI:      "http://example.com",
						Sha256:   "deadbeef",
						Selector: nil,
						Files:    []FileOperation{{"", ""}},
						Bin:      "foo",
					}},
				},
			},
			pluginName: "foo",
			wantErr:    true,
		},
		{
			name: "version unspecified",
			fields: fields{
				TypeMeta:   defaultMeta,
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: PluginSpec{
					Version:          "",
					ShortDescription: "short",
					Platforms: []Platform{{
						URI:      "http://example.com",
						Sha256:   "deadbeef",
						Selector: nil,
						Files:    []FileOperation{{"", ""}},
						Bin:      "foo",
					}},
				},
			},
			pluginName: "foo",
			wantErr:    true,
		},
		{
			name: "version malformed",
			fields: fields{
				TypeMeta:   defaultMeta,
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: PluginSpec{
					Version:          "v01.02-a",
					ShortDescription: "short",
					Platforms: []Platform{{
						URI:    "http://example.com",
						Sha256: "deadbeef",
						Files:  []FileOperation{{"", ""}},
						Bin:    "foo",
					}},
				},
			},
			pluginName: "foo",
			wantErr:    true,
		},
		{
			name: "no file operations",
			fields: fields{
				TypeMeta:   defaultMeta,
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: PluginSpec{
					Version:          "v1.0.0",
					ShortDescription: "short",
					Platforms: []Platform{{
						URI:      "http://example.com",
						Sha256:   "deadbeef",
						Selector: nil,
						Files:    []FileOperation{},
						Bin:      "foo",
					}},
				},
			},
			pluginName: "foo",
			wantErr:    true,
		},
		{
			name: "wrong plugin name",
			fields: fields{
				TypeMeta:   defaultMeta,
				ObjectMeta: metav1.ObjectMeta{Name: "wrong-name"},
				Spec: PluginSpec{
					Version:          "v1.0.0",
					ShortDescription: "short",
					Platforms: []Platform{{
						URI:      "http://example.com",
						Sha256:   "deadbeef",
						Selector: nil,
						Files:    []FileOperation{{"", ""}},
						Bin:      "foo",
					}},
				},
			},
			pluginName: "foo",
			wantErr:    true,
		},
		{
			name: "unsafe plugin name",
			fields: fields{
				TypeMeta:   defaultMeta,
				ObjectMeta: metav1.ObjectMeta{Name: "../foo"},
				Spec: PluginSpec{
					Version:          "v1.0.0",
					ShortDescription: "short",
					Platforms: []Platform{{
						URI:      "http://example.com",
						Sha256:   "deadbeef",
						Selector: nil,
						Files:    []FileOperation{{"", ""}},
						Bin:      "foo",
					}},
				},
			},
			pluginName: "../foo",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Plugin{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
			}
			if err := p.Validate(tt.pluginName); (err != nil) != tt.wantErr {
				t.Errorf("Plugin.Validate(%s) error = %v, wantErr %v", tt.pluginName, err, tt.wantErr)
			}
		})
	}
}

func TestPlatform_Validate(t *testing.T) {
	type fields struct {
		URI      string
		Sha256   string
		Selector *metav1.LabelSelector
		Files    []FileOperation
		Bin      string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "no error validation",
			fields: fields{
				URI:      "http://example.com",
				Sha256:   "deadbeef",
				Selector: nil,
				Files:    []FileOperation{{"", ""}},
				Bin:      "foo",
			},
			wantErr: false,
		},
		{
			name: "only hash",
			fields: fields{
				URI:      "",
				Sha256:   "deadbeef",
				Selector: nil,
				Files:    []FileOperation{{"", ""}},
				Bin:      "foo",
			},
			wantErr: true,
		},
		{
			name: "only uri",
			fields: fields{
				URI:      "http://example.com",
				Sha256:   "",
				Selector: nil,
				Files:    []FileOperation{{"", ""}},
				Bin:      "foo",
			},
			wantErr: true,
		},
		{
			name: "no file operations",
			fields: fields{
				URI:      "http://example.com",
				Sha256:   "deadbeef",
				Selector: nil,
				Files:    []FileOperation{},
				Bin:      "foo",
			},
			wantErr: true,
		},
		{
			name: "no bin field",
			fields: fields{
				URI:      "http://example.com",
				Sha256:   "deadbeef",
				Selector: nil,
				Files:    []FileOperation{{"", ""}},
				Bin:      "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Platform{
				URI:      tt.fields.URI,
				Sha256:   tt.fields.Sha256,
				Selector: tt.fields.Selector,
				Files:    tt.fields.Files,
				Bin:      tt.fields.Bin,
			}
			if err := p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Platform.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
