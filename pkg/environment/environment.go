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

package environment

import (
	"fmt"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"

	"github.com/GoogleContainerTools/krew/pkg/pathutil"
)

// KrewPaths contains all important environment paths
type KrewPaths struct {
	// Base is the path of the krew root.
	// The default path is ~/.kube/plugins/krew
	Base string

	// Index is a git(1) repository containing all local plugin manifests files.
	// ${Index}/<plugin-name>.yaml
	Index string

	// Install is the dir where all plugins will be installed to.
	// ${Install}/<version>/<plugin-content>
	Install string

	// Download is a directory where plugins will be temporarily downloaded to.
	// This is currently a path in os.TempDir()
	Download string

	// Bin is the path krew links all plugin binaries to.
	// This path has to be added to the $PATH.
	Bin string
}

// MustGetKrewPaths returns ensured index paths for krew.
func MustGetKrewPaths() KrewPaths {
	base := filepath.Join(homedir.HomeDir(), ".kube", "plugins", "krew")
	base, err := filepath.Abs(base)
	if err != nil {
		panic(fmt.Errorf("cannot get current pwd err: %v", err))
	}

	return KrewPaths{
		Base:     base,
		Index:    filepath.Join(base, "index"),
		Install:  filepath.Join(base, "store"),
		Download: filepath.Join(os.TempDir(), "krew"),
		Bin:      filepath.Join(base, "bin"),
	}
}

// GetExecutedVersion returns the currently executed version. If krew is
// not executed as an plugin it will return a nil error and an empty string.
func GetExecutedVersion(paths KrewPaths, cmdArgs []string, resolver func(string) (string, error)) (string, bool, error) {
	path, err := resolver(cmdArgs[0])
	if err != nil {
		return "", false, fmt.Errorf("failed to resolve path, err: %v", err)
	}

	currentBinaryPath, err := filepath.Abs(path)
	if err != nil {
		return "", false, err
	}

	pluginsPath, err := filepath.Abs(filepath.Join(paths.Install, "krew"))
	if err != nil {
		return "", false, err
	}

	elems, ok := pathutil.IsSubPath(pluginsPath, currentBinaryPath)
	if !ok || len(elems) < 2 {
		return "", false, nil
	}

	return elems[0], true, nil
}

// DefaultSymlinkResolver resolves symlinks paths
func DefaultSymlinkResolver(path string) (string, error) {
	s, err := os.Lstat(path)
	if err != nil {
		return "", fmt.Errorf("failed to stat the currently executed path")
	}

	if s.Mode()&os.ModeSymlink != 0 {
		if path, err = os.Readlink(path); err != nil {
			return "", fmt.Errorf("failed to resolve the symnlink of the currently executed version")
		}
	}
	return path, nil
}
