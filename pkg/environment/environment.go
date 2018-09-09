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
	"os"
	"path/filepath"
	"strings"

	"github.com/GoogleContainerTools/krew/pkg/pathutil"
	"k8s.io/client-go/util/homedir"
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

func parseEnvs(environ []string) map[string]string {
	flags := make(map[string]string, len(environ))
	for _, pair := range environ {
		kv := strings.SplitN(pair, "=", 2)
		flags[string(kv[0])] = kv[1]
	}
	return flags
}

// MustGetKrewPathsFromEnvs returns ensured index paths for krew.
func MustGetKrewPathsFromEnvs(envs []string) KrewPaths {
	base := filepath.Join(getKubectlPluginsPath(envs), "krew")
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

func getKubectlPluginsPath(_ []string) string {
	// Golang does no '~' expansion
	return filepath.Join(homedir.HomeDir(), ".kube", "plugins")
}

// GetExecutedVersion returns the currently executed version. If krew is
// not executed as an plugin it will return a nil error and an empty string.
// TODO(lbb): Breaks, refactor.
func GetExecutedVersion(paths KrewPaths, cmdArgs []string) (string, bool, error) {
	currentBinaryPath, err := filepath.Abs(cmdArgs[0])
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

// IsPlugin checks if the currently executed binary is a plugin.
func IsPlugin(environ []string) bool {
	_, ok := parseEnvs(environ)["KUBECTL_PLUGINS_DESCRIPTOR_NAME"]
	return ok
}
