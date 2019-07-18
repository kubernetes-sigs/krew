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
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/installation/semver"
)

var (
	safePluginRegexp = regexp.MustCompile(`^[\w-]+$`)

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
		if strings.ToLower(forbidden) == strings.ToLower(name) {
			return false
		}
	}
	return true
}

func isSupportedAPIVersion(apiVersion string) bool {
	return apiVersion == constants.CurrentAPIVersion
}

// Validate checks for structural validity of the Plugin object.
func (p Plugin) Validate(name string) error {
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
		if err := pl.Validate(); err != nil {
			return errors.Wrapf(err, "platform (%+v) is badly constructed", pl)
		}
	}
	return nil
}

// Validate checks Platform for structural validity.
func (p Platform) Validate() error {
	if p.URI == "" {
		return errors.New("URI has to be set")
	}
	if p.Sha256 == "" {
		return errors.New("sha256 sum has to be set")
	}
	if p.Bin == "" {
		return errors.New("bin has to be set")
	}
	if len(p.Files) == 0 {
		return errors.New("can't have a plugin without specifying file operations")
	}
	return nil
}
