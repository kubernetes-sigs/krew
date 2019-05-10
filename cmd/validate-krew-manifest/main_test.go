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

// validate-krew-manifest makes sure a manifest file is valid.s
package main

import (
	"testing"

	"sigs.k8s.io/krew/pkg/index"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_selectorMatchesOSArch(t *testing.T) {
	type args struct {
		selector *metav1.LabelSelector
		os       string
		arch     string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"label - no match", args{&metav1.LabelSelector{MatchLabels: map[string]string{"os": "darwin"}}, "windows", "amd64"}, false},
		{"label - match", args{&metav1.LabelSelector{MatchLabels: map[string]string{"os": "darwin"}}, "darwin", "amd64"}, true},
		{"expression - no match", args{&metav1.LabelSelector{MatchExpressions: []v1.LabelSelectorRequirement{{
			Key:      "os",
			Operator: v1.LabelSelectorOpIn,
			Values:   []string{"darwin", "linux"},
		}}}, "windows", "amd64"}, false},
		{"expression - match", args{&metav1.LabelSelector{MatchExpressions: []v1.LabelSelectorRequirement{{
			Key:      "os",
			Operator: v1.LabelSelectorOpIn,
			Values:   []string{"darwin", "linux"},
		}}}, "darwin", "amd64"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectorMatchesOSArch(tt.args.selector, tt.args.os, tt.args.arch); got != tt.want {
				t.Errorf("selectorMatchesOSArch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findAnyMatchingPlatform(t *testing.T) {
	s1 := &v1.LabelSelector{MatchLabels: map[string]string{"os": "darwin"}}
	o1, a1 := findAnyMatchingPlatform(s1)
	if o1 == "" || a1 == "" {
		t.Fatalf("with selector %v, expected os/arch", s1)
	}

	s2 := &v1.LabelSelector{MatchLabels: map[string]string{"os": "non-existing"}}
	o2, a2 := findAnyMatchingPlatform(s1)
	if o2 == "" && a2 == "" {
		t.Fatalf("with selector %v, expected os/arch", s2)
	}

	s3 := &v1.LabelSelector{MatchExpressions: []v1.LabelSelectorRequirement{{
		Key:      "os",
		Operator: v1.LabelSelectorOpIn,
		Values:   []string{"darwin", "linux"}}}}
	o3, a3 := findAnyMatchingPlatform(s3)
	if o3 == "" || a3 == "" {
		t.Fatalf("with selector %v, expected os/arch", s2)
	}
}

func Test_isOverlappingPlatformSelectors_noOverlap(t *testing.T) {
	p1 := index.Platform{
		Selector: &metav1.LabelSelector{MatchExpressions: []v1.LabelSelectorRequirement{{
			Key:      "os",
			Operator: v1.LabelSelectorOpIn,
			Values:   []string{"darwin", "linux"}}}}}
	p2 := index.Platform{
		Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"os": "windows"}}}

	err := isOverlappingPlatformSelectors([]index.Platform{p1, p2})
	if err != nil {
		t.Fatalf("expected no overlap: %+v", err)
	}
}

func Test_isOverlappingPlatformSelectors_overlap(t *testing.T) {
	p1 := index.Platform{
		Selector: &metav1.LabelSelector{MatchExpressions: []v1.LabelSelectorRequirement{{
			Key:      "os",
			Operator: v1.LabelSelectorOpIn,
			Values:   []string{"darwin", "linux"}}}}}
	p2 := index.Platform{
		Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"os": "darwin"}}}

	err := isOverlappingPlatformSelectors([]index.Platform{p1, p2})
	if err == nil {
		t.Fatal("expected overlap")
	}
}
