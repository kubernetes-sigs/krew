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

package testutil

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

var (
	defaultName             = "test-plugin"
	defaultShortDescription = `in-memory test plugin object`
	defaultVersion          = `v1.0.0-test.1`
	defaultMeta             = metav1.TypeMeta{
		APIVersion: constants.CurrentAPIVersion,
		Kind:       constants.PluginKind,
	}

	defaultPlatformURI    = "http://example.com/"
	defaultPlatformSHA256 = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	defaultPlatformOSArch = map[string]string{
		"os":   "linux",
		"arch": "amd64"}
	defaultPlatformBin            = "kubectl-" + defaultName
	defaultPlatformFileOperations = []index.FileOperation{{From: "./*.sh", To: "."}}
)

type P struct{ v index.Plugin }
type R struct{ v index.Platform }

// NewPlugin builds a index.Plugin that is valid.
func NewPlugin() *P {
	return &P{v: index.Plugin{
		TypeMeta: defaultMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultName,
		},
		Spec: index.PluginSpec{
			ShortDescription: defaultShortDescription,
			Version:          defaultVersion,
			Platforms:        []index.Platform{NewPlatform().V()},
		},
	}}
}

func (p *P) WithName(s string) *P                 { p.v.ObjectMeta.Name = s; return p }
func (p *P) WithShortDescription(v string) *P     { p.v.Spec.ShortDescription = v; return p }
func (p *P) WithTypeMeta(v metav1.TypeMeta) *P    { p.v.TypeMeta = v; return p }
func (p *P) WithPlatforms(v ...index.Platform) *P { p.v.Spec.Platforms = v; return p }
func (p *P) WithVersion(v string) *P              { p.v.Spec.Version = v; return p }
func (p *P) V() index.Plugin                      { return p.v }

func NewPlatform() *R {
	return &R{
		v: index.Platform{
			URI:    defaultPlatformURI,
			Sha256: defaultPlatformSHA256,
			Selector: &metav1.LabelSelector{
				MatchLabels: defaultPlatformOSArch,
			},
			Bin:   defaultPlatformBin,
			Files: defaultPlatformFileOperations,
		},
	}
}

func (p *R) WithOS(os string) *R {
	p.v.Selector = &metav1.LabelSelector{MatchLabels: map[string]string{"os": os}}
	return p
}

func (p *R) WithOSes(os ...string) *R {
	p.v.Selector = &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{
		Key:      "os",
		Operator: metav1.LabelSelectorOpIn,
		Values:   os}}}
	return p
}

func (p *R) WithOSArch(os, arch string) *R {
	p.v.Selector.MatchLabels = map[string]string{"os": os, "arch": arch}
	return p
}

func (p *R) WithSelector(v *metav1.LabelSelector) *R { p.v.Selector = v; return p }
func (p *R) WithFiles(v []index.FileOperation) *R    { p.v.Files = v; return p }
func (p *R) WithBin(v string) *R                     { p.v.Bin = v; return p }
func (p *R) WithURI(v string) *R                     { p.v.URI = v; return p }
func (p *R) WithSHA256(v string) *R                  { p.v.Sha256 = v; return p }
func (p *R) V() index.Platform                       { return p.v }
