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
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

type Receipt struct{ v index.Receipt }

// NewReceipt builds an index.Receipt that is valid.
func NewReceipt() *Receipt {
	return &Receipt{v: index.Receipt{
		Status: index.ReceiptStatus{
			Source: index.SourceIndex{
				Name: constants.DefaultIndexName,
			},
		},
	}}
}

func (r *Receipt) WithPlugin(p index.Plugin) *Receipt        { r.v.Plugin = p; return r }
func (r *Receipt) WithStatus(s index.ReceiptStatus) *Receipt { r.v.Status = s; return r }
func (r *Receipt) V() index.Receipt                          { return r.v }
