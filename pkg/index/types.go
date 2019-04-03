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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Plugin is a top-level type.
// TODO(lbb): Add deepcopy code generation.
type Plugin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PluginSpec `json:"spec"`
}

// PluginSpec TODO(lbb)
type PluginSpec struct {
	Version          string `json:"version,omitempty"`
	ShortDescription string `json:"shortDescription,omitempty"`
	Description      string `json:"description,omitempty"`
	Caveats          string `json:"caveats,omitempty"`
	Homepage         string `json:"homepage,omitempty"`

	Platforms []Platform `json:"platforms,omitempty"`
}

// Platform TODO(lbb)
type Platform struct {
	Head   string `json:"head,omitempty"`
	URI    string `json:"uri,omitempty"`
	Sha256 string `json:"sha256,omitempty"`

	Selector *metav1.LabelSelector `json:"selector,omitempty"`
	Files    []FileOperation       `json:"files"`

	// Bin specifies the path to the plugin executable.
	// The path is relative to the root of the installation folder.
	// The binary will be linked after all FileOperations are executed.
	Bin string `json:"bin"`
}

// FileOperation TODO(lbb)
type FileOperation struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

// PluginList TODO(lbb)
type PluginList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Plugin `json:"items"`
}
