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
	"os"
	"reflect"
	"runtime"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_osArch_default(t *testing.T) {
	inOS, inArch := runtime.GOOS, runtime.GOARCH
	outOS, outArch := osArch()
	if inOS != outOS {
		t.Fatalf("returned OS=%q; expected=%q", outOS, inOS)
	}
	if inArch != outArch {
		t.Fatalf("returned Arch=%q; expected=%q", outArch, inArch)
	}
}
func Test_osArch_override(t *testing.T) {
	customOS, customArch := "dragons", "metav1"
	os.Setenv("KREW_OS", customOS)
	os.Setenv("KREW_ARCH", customArch)
	defer func() {
		os.Unsetenv("KREW_ARCH")
		os.Unsetenv("KREW_OS")
	}()

	outOS, outArch := osArch()
	if customOS != outOS {
		t.Fatalf("returned OS=%q; expected=%q", outOS, customOS)
	}
	if customArch != outArch {
		t.Fatalf("returned Arch=%q; expected=%q", outArch, customArch)
	}
}

func Test_matchPlatformToSystemEnvs(t *testing.T) {
	matchingPlatform := Platform{
		URI: "A",
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"os": "foo",
			},
		},
		Files: nil,
	}

	type args struct {
		i Plugin
	}
	tests := []struct {
		name         string
		args         args
		wantPlatform Platform
		wantFound    bool
		wantErr      bool
	}{
		{
			name: "Test Matching Index",
			args: args{
				i: Plugin{
					Spec: PluginSpec{
						Platforms: []Platform{
							matchingPlatform, {
								URI: "B",
								Selector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"os": "None",
									},
								},
							},
						},
					},
				},
			},
			wantPlatform: matchingPlatform,
			wantFound:    true,
			wantErr:      false,
		}, {
			name: "Test Matching Index Not Found",
			args: args{
				i: Plugin{
					Spec: PluginSpec{
						Platforms: []Platform{
							{
								URI: "B",
								Selector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"os": "None",
									},
								},
							},
						},
					},
				},
			},
			wantPlatform: Platform{},
			wantFound:    false,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPlatform, gotFound, err := tt.args.i.Spec.matchPlatformToSystemEnvs("foo", "amdBar")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMatchingPlatform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPlatform, tt.wantPlatform) {
				t.Errorf("GetMatchingPlatform() gotPlatform = %v, want %v", gotPlatform, tt.wantPlatform)
			}
			if gotFound != tt.wantFound {
				t.Errorf("GetMatchingPlatform() gotFound = %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}
