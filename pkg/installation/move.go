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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/GoogleContainerTools/krew/pkg/index"
	"github.com/GoogleContainerTools/krew/pkg/pathutil"

	"github.com/golang/glog"
)

type move struct {
	from, to string
}

func findMoveTargets(fromDir, toDir string, fo index.FileOperation) ([]move, error) {
	if fo.To != filepath.Clean(fo.To) {
		return nil, fmt.Errorf("the provided path is not clean, %q should be %q", fo.To, filepath.Clean(fo.To))
	}
	fromDir, err := filepath.Abs(fromDir)
	if err != nil {
		return nil, fmt.Errorf("could not get the relative path for the move src, err: %v", err)
	}

	glog.V(4).Infof("Trying to move single file directly from=%q to=%q with file operation=%+v", fromDir, toDir, fo)
	if m, ok, err := getDirectMove(fromDir, toDir, fo); err != nil {
		return nil, fmt.Errorf("failed to detect single move operation, err: %v", err)
	} else if ok {
		glog.V(3).Infof("Detected single move from file operation=%+v", fo)
		return []move{m}, nil
	}

	glog.V(4).Infoln("Wasn't a single file, proceeding with Glob move")
	newDir, err := filepath.Abs(filepath.Join(filepath.FromSlash(toDir), filepath.FromSlash(fo.To)))
	if err != nil {
		return nil, fmt.Errorf("could not get the relative path for the move dst, err: %v", err)
	}

	gl, err := filepath.Glob(filepath.Join(filepath.FromSlash(fromDir), filepath.FromSlash(fo.From)))
	if err != nil {
		return nil, fmt.Errorf("could not get files using a glob string, err: %v", err)
	}

	var moves []move
	for _, v := range gl {
		newPath := filepath.Join(newDir, filepath.Base(filepath.FromSlash(v)))
		// Check secure path
		m := move{from: v, to: newPath}
		if !isMoveAllowed(fromDir, toDir, m) {
			return nil, fmt.Errorf("can't move, move target %v is not a subpath from=%q, to=%q", m, fromDir, toDir)
		}
		moves = append(moves, m)
	}
	return moves, nil
}

func getDirectMove(fromDir, toDir string, fo index.FileOperation) (move, bool, error) {
	var m move
	fromDir, err := filepath.Abs(fromDir)
	if err != nil {
		return m, false, fmt.Errorf("could not get the relative path for the move src, err: %v", err)
	}

	toDir, err = filepath.Abs(toDir)
	if err != nil {
		return m, false, fmt.Errorf("could not get the relative path for the move src, err: %v", err)
	}

	// Check is direct file (not a Glob)
	fromFilePath := filepath.Clean(filepath.Join(fromDir, fo.From))
	_, err = os.Stat(fromFilePath)
	if err != nil {
		return m, false, nil
	}

	// If target is empty use old file name.
	if filepath.Clean(fo.To) == "." {
		fo.To = filepath.Base(fromFilePath)
	}

	// Build new file name
	toFilePath, err := filepath.Abs(filepath.Join(filepath.FromSlash(toDir), filepath.FromSlash(fo.To)))
	if err != nil {
		return m, false, fmt.Errorf("could not get the relative path for the move dst, err: %v", err)
	}

	// Check sane path
	m = move{from: fromFilePath, to: toFilePath}
	if !isMoveAllowed(fromDir, toDir, m) {
		return move{}, false, fmt.Errorf("can't move, move target %v is out of bounds from=%q, to=%q", m, fromDir, toDir)
	}

	return m, true, nil
}

func isMoveAllowed(fromBase, toBase string, m move) bool {
	_, okFrom := pathutil.IsSubPath(fromBase, m.from)
	_, okTo := pathutil.IsSubPath(toBase, m.to)
	return okFrom && okTo
}

func moveFiles(fromDir, toDir string, fo index.FileOperation) error {
	glog.V(4).Infof("Finding move targets from %q to %q with file operation=%v", fromDir, toDir, fo)
	moves, err := findMoveTargets(fromDir, toDir, fo)
	if err != nil {
		return fmt.Errorf("could not find move targets, err: %v", err)
	}

	for _, m := range moves {
		glog.V(2).Infof("Move file from %q to %q", m.from, m.to)
		if err := os.MkdirAll(filepath.Dir(m.to), 0755); err != nil {
			return fmt.Errorf("failed to create move path %q, err: %v", filepath.Dir(m.to), err)
		}

		if err = os.Rename(m.from, m.to); err != nil {
			return fmt.Errorf("could not rename file from %q to %q, err: %v", m.from, m.to, err)
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

func moveToInstallAtomic(download, pluginDir, version string, fos []index.FileOperation) error {
	glog.V(4).Infof("Creating plugin dir %q", pluginDir)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("error creating path to %q, err: %v", pluginDir, err)
	}

	tempdir, err := ioutil.TempDir("", "krew-temp-move")
	glog.V(4).Infof("Creating temp plugin move operations dir %q", tempdir)
	if err != nil {
		return fmt.Errorf("failed to find a temporary director, err: %v", err)
	}
	defer os.RemoveAll(tempdir)

	if err = moveAllFiles(download, tempdir, fos); err != nil {
		return fmt.Errorf("failed to move files, err: %v", err)
	}

	installPath := filepath.Join(pluginDir, version)
	glog.V(2).Infof("Move %q to %q", tempdir, installPath)
	if err = moveOrCopy(tempdir, installPath); err != nil {
		defer os.Remove(installPath)
		return fmt.Errorf("could not rename file from %q to %q, err: %v", tempdir, installPath, err)
	}

	return nil
}

// moveOrCopy will try to rename a dir or file. If rename is not supported a manual copy will be performed.
func moveOrCopy(from, to string) error {
	// Try atomic rename (does not work cross partition).
	err := os.Rename(from, to)
	// Fallback for invalid cross-device link (errno:18).
	if le, ok := err.(*os.LinkError); err != nil && ok {
		if errno, ok := le.Err.(syscall.Errno); ok && errno == 18 {
			glog.V(4).Infof("Cross-device link error (ERRNO=18), fallback to manual copy")
			return copyDir(from, to)
		}
	}
	return err
}

func copyDir(from string, to string) (err error) {
	return filepath.Walk(from, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		newPath, _ := pathutil.ReplaceBase(path, from, to)
		if info.IsDir() {
			glog.V(4).Infof("Creating new dir %q", newPath)
			err = os.MkdirAll(newPath, info.Mode())
		} else {
			glog.V(4).Infof("Copying file %q", newPath)
			err = copyFile(path, newPath, info.Mode())
		}
		return err
	})
}

func copyFile(source string, dst string, mode os.FileMode) (err error) {
	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()

	df, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer df.Close()

	_, err = io.Copy(df, sf)
	if err != nil {
		return err
	}
	return os.Chmod(dst, mode)
}
