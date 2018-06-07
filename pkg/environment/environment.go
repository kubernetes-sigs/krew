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

	"k8s.io/client-go/util/homedir"
)

// KrewPaths contains all important enviroment paths
type KrewPaths struct {
	// Base is the path of the krew root.
	// The default path is ~/.kube/plugins/krew
	Base string

	// Index is a git(1) repository containing all local plugin manifests files.
	// ${Index}/<plugin-name>.yaml
	Index string

	// Install is the folder where all plugins will be installed to.
	// ${Install}/<version>/<plugin-content>
	Install string

	// Download is a directory where plugins will be temporarily downloaded to.
	// This is currently a path in os.TempDir()
	Download string
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
	}
}

func getKubectlPluginsPath(envs []string) string {
	envvars := parseEnvs(envs)

	// Look for "${KUBECTL_PLUGINS_PATH}"
	if path, ok := envvars["KUBECTL_PLUGINS_PATH"]; ok && path != "" {
		return path
	}

	// TODO(lbb): make os call testable! (dep injection)
	// Look for "${XDG_DATA_DIRS}/kubectl/plugins"
	if path := envvars["XDG_DATA_DIRS"]; path != "" {
		fullPath := filepath.Join(path, "kubectl", "plugins")
		stat, err := os.Stat(fullPath)
		if err == nil && stat.IsDir() {
			return fullPath
		}
	}

	// Golang does no '~' expansion
	return filepath.Join(homedir.HomeDir(), ".kube", "plugins")
}
