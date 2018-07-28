// Copyright © 2018 Google Inc.
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
	"fmt"
	"regexp"
	"strings"
)

var (
	safePluginRegexp = regexp.MustCompile(`^[\w-]+$`)
	// windowsForbidden is taken from  https://docs.microsoft.com/en-us/windows/desktop/FileIO/naming-a-file
	windowsForbidden = []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
)

// IsSafePluginName checks if the plugin Name is save to use.
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

// Validate TODO(lbb)
func (p Plugin) Validate(name string) error {
	if !IsSafePluginName(name) {
		return fmt.Errorf("the plugin name %q is not allowed, must match %q", name, safePluginRegexp.String())
	}
	if p.Name != name {
		return fmt.Errorf("plugin should be named %q, not %q", name, p.Name)
	}
	if p.Spec.ShortDescription == "" {
		return fmt.Errorf("should have a short description")
	}
	if len(p.Spec.Platforms) == 0 {
		return fmt.Errorf("should have a platform specified")
	}
	for _, pl := range p.Spec.Platforms {
		if err := pl.Validate(); err != nil {
			return fmt.Errorf("platform (%+v) is badly constructed, err: %v", pl, err)
		}
	}
	return nil
}

// Validate TODO(lbb)
func (p Platform) Validate() error {
	if (p.Sha256 != "") != (p.URI != "") {
		return fmt.Errorf("can't get version URI and sha have both to be set or unset")
	}
	if p.Head == "" && p.URI == "" {
		return fmt.Errorf("head or URI have to be set")
	}
	if len(p.Files) == 0 {
		return fmt.Errorf("can't have a plugin without specifying file operations")
	}
	return nil
}
