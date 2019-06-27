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

package krew

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/golang/glog"

	"sigs.k8s.io/krew/pkg/constants"
)

const (
	persistentIndexCache = "krew-persistent-index-cache"
)

var (
	once     sync.Once
	indexTar []byte
)

// InitializeIndex initializes the krew index in `$root/index` with the actual krew-index.
// It caches the index tree as in-memory tar after the first run.
func (it *ITest) initializeIndex() {
	once.Do(func() {
		persistentCacheFile := filepath.Join(os.TempDir(), persistentIndexCache)
		fileInfo, err := os.Stat(persistentCacheFile)

		if err == nil && fileInfo.Mode().IsRegular() {
			it.t.Logf("Using persistent index cache from file %q", persistentCacheFile)
			if indexTar, err = ioutil.ReadFile(persistentCacheFile); err == nil {
				return
			}
		}

		if indexTar, err = initFromGitClone(); err != nil {
			it.t.Fatalf("cannot clone repository: %s", err)
		}

		ioutil.WriteFile(persistentCacheFile, indexTar, 0600)
	})

	indexDir := filepath.Join(it.Root(), "index")
	if err := os.Mkdir(indexDir, 0777); err != nil {
		if os.IsExist(err) {
			it.t.Log("initializeIndex should only be called once")
			return
		}
		it.t.Fatal(err)
	}

	cmd := exec.Command("tar", "xzf", "-", "-C", indexDir)
	cmd.Stdin = bytes.NewReader(indexTar)
	if err := cmd.Run(); err != nil {
		it.t.Fatalf("cannot restore index from cache: %s", err)
	}
}

func initFromGitClone() ([]byte, error) {
	const tarName = "index.tar"
	indexRoot, err := ioutil.TempDir("", "krew-index-cache")
	if err != nil {
		return nil, err
	}
	defer func() {
		err := os.RemoveAll(indexRoot)
		glog.V(1).Infoln("cannot remove temporary directory:", err)
	}()

	cmd := exec.Command("git", "clone", "--depth=1", "--single-branch", "--no-tags", constants.IndexURI)
	cmd.Dir = indexRoot
	if err = cmd.Run(); err != nil {
		return nil, err
	}

	cmd = exec.Command("tar", "czf", tarName, "-C", "krew-index", ".")
	cmd.Dir = indexRoot
	if err = cmd.Run(); err != nil {
		return nil, err
	}

	return ioutil.ReadFile(filepath.Join(indexRoot, tarName))
}
