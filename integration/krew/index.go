package krew

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/golang/glog"
	"sigs.k8s.io/krew/cmd/krew/cmd"
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
func (k *KrewTest) initializeIndex() {
	once.Do(func() {
		persistentCacheFile := filepath.Join(os.TempDir(), persistentIndexCache)
		fileInfo, err := os.Stat(persistentCacheFile)

		if err == nil && fileInfo.Mode().IsRegular() {
			k.t.Logf("Using persistent index cache from file %q", persistentCacheFile)
			if indexTar, err = ioutil.ReadFile(persistentCacheFile); err == nil {
				return
			}
		}

		if indexTar, err = initFromGitClone(); err != nil {
			k.t.Fatalf("cannot clone repository: %s", err)
		}

		ioutil.WriteFile(persistentCacheFile, indexTar, 0600)
	})

	indexDir := filepath.Join(k.Root(), "index")
	if err := os.Mkdir(indexDir, 0777); err != nil {
		k.t.Fatal(err)
	}

	cmd := exec.Command("tar", "xzf", "-", "-C", indexDir)
	cmd.Stdin = bytes.NewReader(indexTar)
	if err := cmd.Run(); err != nil {
		k.t.Fatalf("cannot restore index from cache: %s", err)
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

	cmd := exec.Command("git", "clone", "--depth=1", "--single-branch", "--no-tags", cmd.IndexURI)
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
