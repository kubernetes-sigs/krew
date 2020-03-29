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

package cmd

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"sigs.k8s.io/krew/internal/testutil"
	"sigs.k8s.io/krew/pkg/index"
)

func Test_displayName(t *testing.T) {
	tests := []struct {
		name     string
		receipt  index.Receipt
		expected string
	}{
		{
			name:     "explicit default index",
			receipt:  testutil.NewReceipt().WithPlugin(testutil.NewPlugin().WithName("foo").V()).V(),
			expected: "foo",
		},
		{
			name:     "no index",
			receipt:  testutil.NewReceipt().WithPlugin(testutil.NewPlugin().WithName("foo").V()).WithStatus(index.ReceiptStatus{}).V(),
			expected: "foo",
		},
		{
			name: "custom index",
			receipt: testutil.NewReceipt().WithPlugin(testutil.NewPlugin().WithName("bar").V()).WithStatus(index.ReceiptStatus{
				Source: index.SourceIndex{
					Name: "foo",
				},
			}).V(),
			expected: "foo/bar",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := displayName(test.receipt)
			if diff := cmp.Diff(test.expected, actual); diff != "" {
				t.Fatalf("expected name to match: %s", diff)
			}
		})
	}
}
