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
	"os"
	"path/filepath"
	"runtime"
)

// resolveNetrcFile returns the netrc file path, defaulting to ~/.netrc (~/_netrc on Windows) if empty
func resolveNetrcFile(netrcFile string) (string, error) {
	if netrcFile != "" {
		return netrcFile, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Use _netrc on Windows, .netrc on other systems
	var netrcFilename string
	if runtime.GOOS == "windows" {
		netrcFilename = "_netrc"
	} else {
		netrcFilename = ".netrc"
	}

	return filepath.Join(homeDir, netrcFilename), nil
}
