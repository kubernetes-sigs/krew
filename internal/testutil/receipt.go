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

	"sigs.k8s.io/krew/pkg/index"
)

type Receipt struct{ v index.Receipt }

// NewPlugin builds a index.Plugin that is valid.
func NewReceipt() *Receipt {
	return &Receipt{v: index.Receipt{Plugin: NewPlugin().V()}}
}

func (r *Receipt) WithName(s string) *Receipt                 { r.v.ObjectMeta.Name = s; return r }
func (r *Receipt) WithShortDescription(v string) *Receipt     { r.v.Spec.ShortDescription = v; return r }
func (r *Receipt) WithTypeMeta(v metav1.TypeMeta) *Receipt    { r.v.TypeMeta = v; return r }
func (r *Receipt) WithPlatforms(v ...index.Platform) *Receipt { r.v.Spec.Platforms = v; return r }
func (r *Receipt) WithVersion(v string) *Receipt              { r.v.Spec.Version = v; return r }
func (r *Receipt) V() index.Receipt                           { return r.v }
