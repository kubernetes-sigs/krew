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

// Package validation implements functions to validate Krew plugin types.
package validation

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
	"sigs.k8s.io/krew/pkg/installation/semver"
)

const (
	sha256Pattern = `^[a-f0-9]{64}$`
)

var (
	safePluginRegexp = regexp.MustCompile(`^[\w-]+$`)
	validSHA256      = regexp.MustCompile(sha256Pattern)

	// windowsForbidden is taken from  https://docs.microsoft.com/en-us/windows/desktop/FileIO/naming-a-file
	windowsForbidden = []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2",
		"COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2",
		"LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
)

// IsSafePluginName checks if the plugin Name is safe to use.
func IsSafePluginName(name string) bool {
	if !safePluginRegexp.MatchString(name) {
		return false
	}
	for _, forbidden := range windowsForbidden {
		if strings.EqualFold(forbidden, name) {
			return false
		}
	}
	return true
}

func isSupportedAPIVersion(apiVersion string) bool {
	return apiVersion == constants.CurrentAPIVersion
}

func isValidSHA256(s string) bool { return validSHA256.MatchString(s) }

// ValidatePlugin checks for structural validity of the Plugin object with given
// name.
func ValidatePlugin(name string, p index.Plugin) error {
	if !isSupportedAPIVersion(p.APIVersion) {
		return errors.Errorf("plugin manifest has apiVersion=%q, not supported in this version of krew (try updating plugin index or install a newer version of krew)", p.APIVersion)
	}

	if p.Kind != constants.PluginKind {
		return errors.Errorf("plugin manifest has kind=%q, but only %q is supported", p.Kind, constants.PluginKind)
	}
	if !IsSafePluginName(name) {
		return errors.Errorf("the plugin name %q is not allowed, must match %q", name, safePluginRegexp.String())
	}
	if p.Name != name {
		return errors.Errorf("plugin should be named %q, not %q", name, p.Name)
	}
	if p.Spec.ShortDescription == "" {
		return errors.New("should have a short description")
	}
	if len(p.Spec.Platforms) == 0 {
		return errors.New("should have a platform specified")
	}
	if p.Spec.Version == "" {
		return errors.New("should have a version specified")
	}
	if _, err := semver.Parse(p.Spec.Version); err != nil {
		return errors.Wrap(err, "failed to parse plugin version")
	}
	for _, pl := range p.Spec.Platforms {
		if err := validatePlatform(pl); err != nil {
			return errors.Wrapf(err, "platform (%+v) is badly constructed", pl)
		}
	}
	return nil
}

// validatePlatform checks Platform for structural validity.
func validatePlatform(p index.Platform) error {
	if p.URI == "" {
		return errors.New("`uri` has to be set")
	}
	if p.Sha256 == "" {
		return errors.New("`sha256` sum has to be set")
	}
	if !isValidSHA256(p.Sha256) {
		return errors.Errorf("`sha256` value %s is not valid, must match pattern %s", p.Sha256, sha256Pattern)
	}
	if p.Bin == "" {
		return errors.New("`bin` has to be set")
	}
	if err := validateFiles(p.Files); err != nil {
		return errors.Wrap(err, "`files` is invalid")
	}
	if err := validateSelector(p.Selector); err != nil {
		return errors.Wrap(err, "invalid platform selector")
	}
	return nil
}

func validateFiles(fops []index.FileOperation) error {
	if fops == nil {
		return nil
	}
	if len(fops) == 0 {
		return errors.New("`files` has to be unspecified or non-empty")
	}
	for _, op := range fops {
		if op.From == "" {
			return errors.New("`from` field has to be set")
		} else if op.To == "" {
			return errors.New("`to` field has to be set")
		}
	}
	return nil
}

// validateSelector checks if the platform selector uses supported keys and is not empty or nil.
func validateSelector(sel *metav1.LabelSelector) error {
	if sel == nil {
		return errors.New("nil selector is not supported")
	}
	if sel.MatchLabels == nil && len(sel.MatchExpressions) == 0 {
		return errors.New("empty selector is not supported")
	}

	// check for unsupported keys
	keys := []string{}
	for k := range sel.MatchLabels {
		keys = append(keys, k)
	}
	for _, expr := range sel.MatchExpressions {
		keys = append(keys, expr.Key)
	}
	for _, key := range keys {
		if key != "os" && key != "arch" {
			return errors.Errorf("key %q not supported", key)
		}
	}

	if sel.MatchLabels != nil && len(sel.MatchLabels) == 0 {
		return errors.New("`matchLabels` specified but empty")
	}
	if sel.MatchExpressions != nil && len(sel.MatchExpressions) == 0 {
		return errors.New("`matchExpressions` specified but empty")
	}

	return nil
}
