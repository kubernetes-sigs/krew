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

package installation

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/google/krew/pkg/index"
)

type move struct {
	from, to string
}

func findMoveTargets(fromDir, toDir string, fo index.FileOperation) ([]move, error) {
	fromDir, err := filepath.Abs(fromDir)
	if err != nil {
		return nil, fmt.Errorf("could not get the realtive path for the move src, err: %v", err)
	}

	newDir, err := filepath.Abs(filepath.Join(filepath.FromSlash(toDir), filepath.FromSlash(fo.To)))
	if err != nil {
		return nil, fmt.Errorf("could not get the realtive path for the move dst, err: %v", err)
	}

	gl, err := filepath.Glob(filepath.Join(filepath.FromSlash(fromDir), filepath.FromSlash(fo.From)))
	if err != nil {
		return nil, fmt.Errorf("could not get files using a glob string, err: %v", err)
	}

	moves := []move{}
	for _, v := range gl {
		// total path is in a subdir
		if len(fromDir) > len(v) {
			return nil, fmt.Errorf("cannot move a file that is not in the from dir (%s -> %s)", fromDir, v)
		}

		newpath := filepath.Join(newDir, filepath.Base(filepath.FromSlash(v)))
		moves = append(moves, move{from: v, to: newpath})
	}
	return moves, nil
}
func moveFiles(fromDir, toDir string, fo index.FileOperation) error {
	moves, err := findMoveTargets(fromDir, toDir, fo)
	if err != nil {
		return fmt.Errorf("could not find move targets, err: %v", err)
	}

	for _, m := range moves {
		glog.V(2).Infof("Move file from %q -> %q\n", m.from, m.to)
		if err = os.Rename(m.from, m.to); err != nil {
			return fmt.Errorf("could not renmame file from %q to %q, err: %v", m.from, m.to, err)
		}
	}
	return nil
}

func moveAllFiles(fromDir, toDir string, fos []index.FileOperation) error {
	for _, fo := range fos {
		if err := moveFiles(fromDir, toDir, fo); err != nil {
			return fmt.Errorf("failed moving files, err: %v", err)
		}
	}
	return nil
}

func moveToInstallAtomic(download, plugindir, version string, fos []index.FileOperation) error {
	if err := os.MkdirAll(plugindir, os.ModePerm); err != nil {
		return fmt.Errorf("Error creating path to %q, err: %v", plugindir, err)
	}
	tempdir, err := ioutil.TempDir("", "krew-temp-move")
	if err != nil {
		return fmt.Errorf("failed to find a temporary director, err: %v", err)
	}
	defer os.RemoveAll(tempdir)
	if err = moveAllFiles(download, tempdir, fos); err != nil {
		return fmt.Errorf("failed to move files, err: %v", err)
	}
	installPath := filepath.Join(plugindir, version)
	glog.V(2).Infof("Move %q -> %q", tempdir, installPath)
	if err = os.Rename(tempdir, installPath); err != nil {
		defer os.RemoveAll(installPath)
		return fmt.Errorf("could not renmame file from %q to %q, err: %v", tempdir, installPath, err)
	}

	return nil
}
