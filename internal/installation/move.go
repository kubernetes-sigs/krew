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

package installation

import (
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	"sigs.k8s.io/krew/internal/pathutil"
	"sigs.k8s.io/krew/pkg/index"
)

type move struct {
	from, to string
}

func findMoveTargets(fromDir, toDir string, fo index.FileOperation) ([]move, error) {
	if fo.To != filepath.Clean(fo.To) {
		return nil, errors.Errorf("the provided path is not clean, %q should be %q", fo.To, filepath.Clean(fo.To))
	}
	fromDir, err := filepath.Abs(fromDir)
	if err != nil {
		return nil, errors.Wrap(err, "could not get the relative path for the move src")
	}

	klog.V(4).Infof("Trying to move single file directly from=%q to=%q with file operation=%#v", fromDir, toDir, fo)
	if m, ok, err := getDirectMove(fromDir, toDir, fo); err != nil {
		return nil, errors.Wrap(err, "failed to detect single move operation")
	} else if ok {
		klog.V(3).Infof("Detected single move from file operation=%#v", fo)
		return []move{m}, nil
	}

	klog.V(4).Infoln("Wasn't a single file, proceeding with Glob move")
	newDir, err := filepath.Abs(filepath.Join(filepath.FromSlash(toDir), filepath.FromSlash(fo.To)))
	if err != nil {
		return nil, errors.Wrap(err, "could not get the relative path for the move dst")
	}

	gl, err := filepath.Glob(filepath.Join(filepath.FromSlash(fromDir), filepath.FromSlash(fo.From)))
	if err != nil {
		return nil, errors.Wrap(err, "could not get files using a glob string")
	}
	if len(gl) == 0 {
		return nil, errors.Errorf("no files in the plugin archive matched the glob pattern=%s", fo.From)
	}

	moves := make([]move, 0, len(gl))
	for _, v := range gl {
		newPath := filepath.Join(newDir, filepath.Base(filepath.FromSlash(v)))
		// Check secure path
		m := move{from: v, to: newPath}
		if !isMoveAllowed(fromDir, toDir, m) {
			return nil, errors.Errorf("can't move, move target %v is not a subpath from=%q, to=%q", m, fromDir, toDir)
		}
		moves = append(moves, m)
	}
	return moves, nil
}

func getDirectMove(fromDir, toDir string, fo index.FileOperation) (move, bool, error) {
	var m move
	fromDir, err := filepath.Abs(fromDir)
	if err != nil {
		return m, false, errors.Wrap(err, "could not get the relative path for the move src")
	}

	toDir, err = filepath.Abs(toDir)
	if err != nil {
		return m, false, errors.Wrap(err, "could not get the relative path for the move src")
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
		return m, false, errors.Wrap(err, "could not get the relative path for the move dst")
	}

	// Check sane path
	m = move{from: fromFilePath, to: toFilePath}
	if !isMoveAllowed(fromDir, toDir, m) {
		return move{}, false, errors.Errorf("can't move, move target %v is out of bounds from=%q, to=%q", m, fromDir, toDir)
	}

	return m, true, nil
}

func isMoveAllowed(fromBase, toBase string, m move) bool {
	_, okFrom := pathutil.IsSubPath(fromBase, m.from)
	_, okTo := pathutil.IsSubPath(toBase, m.to)
	return okFrom && okTo
}

func moveFiles(fromDir, toDir string, fo index.FileOperation) error {
	klog.V(4).Infof("Finding move targets from %q to %q with file operation=%#v", fromDir, toDir, fo)
	moves, err := findMoveTargets(fromDir, toDir, fo)
	if err != nil {
		return errors.Wrap(err, "could not find move targets")
	}

	for _, m := range moves {
		klog.V(2).Infof("Move file from %q to %q", m.from, m.to)
		if err := os.MkdirAll(filepath.Dir(m.to), 0o755); err != nil {
			return errors.Wrapf(err, "failed to create move path %q", filepath.Dir(m.to))
		}

		if err = renameOrCopy(m.from, m.to); err != nil {
			return errors.Wrapf(err, "could not rename/copy file from %q to %q", m.from, m.to)
		}
	}
	klog.V(4).Infoln("Move operations are complete")
	return nil
}

func moveAllFiles(fromDir, toDir string, fos []index.FileOperation) error {
	for _, fo := range fos {
		if err := moveFiles(fromDir, toDir, fo); err != nil {
			return errors.Wrap(err, "failed moving files")
		}
	}
	return nil
}

// moveToInstallDir moves plugins from srcDir to dstDir (created in this method) with given FileOperation.
func moveToInstallDir(srcDir, installDir string, fos []index.FileOperation) error {
	installationDir := filepath.Dir(installDir)
	klog.V(4).Infof("Creating directory %q", installationDir)
	if err := os.MkdirAll(installationDir, 0o755); err != nil {
		return errors.Wrapf(err, "error creating directory at %q", installationDir)
	}

	tmp, err := os.MkdirTemp("", "krew-temp-move")
	klog.V(4).Infof("Creating temp plugin move operations dir %q", tmp)
	if err != nil {
		return errors.Wrap(err, "failed to find a temporary directory")
	}
	klog.V(4).Infof("Chmoding tmpdir %q to 0755", tmp)
	if err := os.Chmod(tmp, 0o755); err != nil {
		// mktemp gives a 0700 directory but since we move this to KREW_ROOT, we need to make it 0755
		return errors.Wrap(err, "failed to chmod temp directory")
	}
	defer os.RemoveAll(tmp)

	if err = moveAllFiles(srcDir, tmp, fos); err != nil {
		return errors.Wrap(err, "failed to move files")
	}

	klog.V(2).Infof("Move directory %q to %q", tmp, installDir)
	if err = renameOrCopy(tmp, installDir); err != nil {
		defer func() {
			klog.V(3).Info("Cleaning up installation directory due to error during copying files")
			os.Remove(installDir)
		}()
		return errors.Wrapf(err, "could not rename/copy directory %q to %q", tmp, installDir)
	}
	return nil
}

// renameOrCopy will try to rename a dir or file. If rename is not supported, a manual copy will be performed.
// Existing files at "to" will be deleted.
func renameOrCopy(from, to string) error {
	// Try atomic rename (does not work cross partition).
	fi, err := os.Stat(to)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "error checking move target dir %q", to)
	}
	if fi != nil && fi.IsDir() {
		klog.V(4).Infof("There's already a directory at move target %q. deleting.", to)
		if err := os.RemoveAll(to); err != nil {
			return errors.Wrapf(err, "error cleaning up dir %q", to)
		}
		klog.V(4).Infof("Move target directory %q cleaned up", to)
	}

	err = os.Rename(from, to)
	// Fallback for invalid cross-device link (errno:18).
	if isCrossDeviceRenameErr(err) {
		klog.V(2).Infof("Cross-device link error while copying, fallback to manual copy")
		return errors.Wrap(copyTree(from, to), "failed to copy directory tree as a fallback")
	}
	return err
}

// copyTree copies files or directories, recursively.
func copyTree(from, to string) (err error) {
	return filepath.Walk(from, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		newPath, _ := pathutil.ReplaceBase(path, from, to)
		if info.IsDir() {
			klog.V(4).Infof("Creating new dir %q", newPath)
			err = os.MkdirAll(newPath, info.Mode())
		} else {
			klog.V(4).Infof("Copying file %q", newPath)
			err = copyFile(path, newPath, info.Mode())
		}
		return err
	})
}

func copyFile(source, dst string, mode os.FileMode) (err error) {
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

// isCrossDeviceRenameErr determines if a os.Rename error is due to cross-fs/drive/volume copying.
func isCrossDeviceRenameErr(err error) bool {
	le, ok := err.(*os.LinkError)
	if !ok {
		return false
	}
	errno, ok := le.Err.(syscall.Errno)
	if !ok {
		return false
	}
	return (IsWindows() && errno == 17) || // syscall.ERROR_NOT_SAME_DEVICE
		(!IsWindows() && errno == 18) // syscall.EXDEV
}
