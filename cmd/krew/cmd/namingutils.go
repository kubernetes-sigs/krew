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
	"regexp"

	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
)

var canonicalNameRegex = regexp.MustCompile(`^[\w-]+/[\w-]+$`)

// indexOf returns the index name of a receipt.
func indexOf(r index.Receipt) string {
	if r.Status.Source.Name == "" {
		return constants.DefaultIndexName
	}
	return r.Status.Source.Name
}

// displayName returns the display name of a Plugin.
// The index name is omitted if it is the default index.
func displayName(p index.Plugin, indexName string) string {
	if isDefaultIndex(indexName) {
		return p.Name
	}
	return indexName + "/" + p.Name
}

func isDefaultIndex(name string) bool {
	return name == "" || name == constants.DefaultIndexName
}

// canonicalName returns INDEX/NAME value for a plugin, even if
// it is in the default index.
func canonicalName(p index.Plugin, indexName string) string {
	if isDefaultIndex(indexName) {
		indexName = constants.DefaultIndexName
	}
	return indexName + "/" + p.Name
}

func isCanonicalName(s string) bool {
	return canonicalNameRegex.MatchString(s)
}
