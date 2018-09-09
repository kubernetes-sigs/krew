// Copyright © 2018 Google Inc.
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

func TestPlugin_Validate(t *testing.T) {
	type fields struct {
		TypeMeta   metav1.TypeMeta
		ObjectMeta metav1.ObjectMeta
		Spec       PluginSpec
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "validate success",
			fields: fields{
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: PluginSpec{
					Version:          "",
					ShortDescription: "short",
					Description:      "",
					Caveats:          "",
					Platforms: []Platform{{
						Head:     "http://example.com",
						URI:      "",
						Sha256:   "",
						Selector: nil,
						Files:    []FileOperation{{"", ""}},
						Bin:      "kubectl-foo",
					}},
				},
			},
			args: args{
				name: "foo",
			},
			wantErr: false,
		},
		{
			name: "no short description",
			fields: fields{
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: PluginSpec{
					Version:          "",
					ShortDescription: "",
					Description:      "",
					Caveats:          "",
					Platforms: []Platform{{
						Head:     "http://example.com",
						URI:      "",
						Sha256:   "",
						Selector: nil,
						Files:    []FileOperation{{"", ""}},
						Bin:      "kubectl-foo",
					}},
				},
			},
			args: args{
				name: "foo",
			},
			wantErr: true,
		},
		{
			name: "no file operations",
			fields: fields{
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: PluginSpec{
					Version:          "",
					ShortDescription: "short",
					Description:      "",
					Caveats:          "",
					Platforms: []Platform{{
						Head:     "http://example.com",
						URI:      "",
						Sha256:   "",
						Selector: nil,
						Files:    []FileOperation{},
						Bin:      "kubectl-foo",
					}},
				},
			},
			args: args{
				name: "foo",
			},
			wantErr: true,
		},
		{
			name: "wrong plugin name",
			fields: fields{
				ObjectMeta: metav1.ObjectMeta{Name: "wrong-name"},
				Spec: PluginSpec{
					Version:          "",
					ShortDescription: "short",
					Description:      "",
					Caveats:          "",
					Platforms: []Platform{{
						Head:     "http://example.com",
						URI:      "",
						Sha256:   "",
						Selector: nil,
						Files:    []FileOperation{{"", ""}},
						Bin:      "kubectl-foo",
					}},
				},
			},
			args: args{
				name: "foo",
			},
			wantErr: true,
		},
		{
			name: "unsafe plugin name",
			fields: fields{
				ObjectMeta: metav1.ObjectMeta{Name: "../foo"},
				Spec: PluginSpec{
					Version:          "",
					ShortDescription: "short",
					Description:      "",
					Caveats:          "",
					Platforms: []Platform{{
						Head:     "http://example.com",
						URI:      "",
						Sha256:   "",
						Selector: nil,
						Files:    []FileOperation{{"", ""}},
						Bin:      "kubectl-foo",
					}},
				},
			},
			args: args{
				name: "../foo",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Plugin{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
			}
			if err := p.Validate(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("Plugin.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlatform_Validate(t *testing.T) {
	type fields struct {
		Head     string
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
				Head:     "http://example.com",
				URI:      "",
				Sha256:   "",
				Selector: nil,
				Files:    []FileOperation{{"", ""}},
				Bin:      "kubectl-foo",
			},
			wantErr: false,
		},
		{
			name: "no url",
			fields: fields{
				Head:     "",
				URI:      "",
				Sha256:   "",
				Selector: nil,
				Files:    []FileOperation{{"", ""}},
				Bin:      "kubectl-foo",
			},
			wantErr: true,
		},
		{
			name: "no only hash",
			fields: fields{
				Head:     "",
				URI:      "",
				Sha256:   "deadbeef",
				Selector: nil,
				Files:    []FileOperation{{"", ""}},
				Bin:      "kubectl-foo",
			},
			wantErr: true,
		},
		{
			name: "no hash but uri",
			fields: fields{
				Head:     "",
				URI:      "http://example.com",
				Sha256:   "",
				Selector: nil,
				Files:    []FileOperation{{"", ""}},
				Bin:      "kubectl-foo",
			},
			wantErr: true,
		},
		{
			name: "no files",
			fields: fields{
				Head:     "http://example.com",
				URI:      "",
				Sha256:   "",
				Selector: nil,
				Files:    []FileOperation{},
				Bin:      "kubectl-foo",
			},
			wantErr: true,
		},
		{
			name: "no bin",
			fields: fields{
				Head:     "http://example.com",
				URI:      "",
				Sha256:   "",
				Selector: nil,
				Files:    []FileOperation{{"", ""}},
				Bin:      "",
			},
			wantErr: true,
		},
		{
			name: "wrong bin prefix",
			fields: fields{
				Head:     "http://example.com",
				URI:      "",
				Sha256:   "",
				Selector: nil,
				Files:    []FileOperation{{"", ""}},
				Bin:      "foo",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Platform{
				Head:     tt.fields.Head,
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
