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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"k8s.io/klog/v2"

	"sigs.k8s.io/krew/internal/pathutil"
	"sigs.k8s.io/krew/pkg/index"
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
		return nil, fmt.Errorf("could not get the relative path for the move src: %w", err)
	}

	klog.V(4).Infof("Trying to move single file directly from=%q to=%q with file operation=%#v", fromDir, toDir, fo)
	if m, ok, err := getDirectMove(fromDir, toDir, fo); err != nil {
		return nil, fmt.Errorf("failed to detect single move operation: %w", err)
	} else if ok {
		klog.V(3).Infof("Detected single move from file operation=%#v", fo)
		return []move{m}, nil
	}

	klog.V(4).Infoln("Wasn't a single file, proceeding with Glob move")
	newDir, err := filepath.Abs(filepath.Join(filepath.FromSlash(toDir), filepath.FromSlash(fo.To)))
	if err != nil {
		return nil, fmt.Errorf("could not get the relative path for the move dst: %w", err)
	}

	gl, err := filepath.Glob(filepath.Join(filepath.FromSlash(fromDir), filepath.FromSlash(fo.From)))
	if err != nil {
		return nil, fmt.Errorf("could not get files using a glob string: %w", err)
	}
	if len(gl) == 0 {
		return nil, fmt.Errorf("no files in the plugin archive matched the glob pattern=%s", fo.From)
	}

	moves := make([]move, 0, len(gl))
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
		return m, false, fmt.Errorf("could not get the relative path for the move src: %w", err)
	}

	toDir, err = filepath.Abs(toDir)
	if err != nil {
		return m, false, fmt.Errorf("could not get the relative path for the move src: %w", err)
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
		return m, false, fmt.Errorf("could not get the relative path for the move dst: %w", err)
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
	klog.V(4).Infof("Finding move targets from %q to %q with file operation=%#v", fromDir, toDir, fo)
	moves, err := findMoveTargets(fromDir, toDir, fo)
	if err != nil {
		return fmt.Errorf("could not find move targets: %w", err)
	}

	for _, m := range moves {
		klog.V(2).Infof("Move file from %q to %q", m.from, m.to)
		if err := os.MkdirAll(filepath.Dir(m.to), 0o755); err != nil {
			return fmt.Errorf("failed to create move path %q: %w", filepath.Dir(m.to), err)
		}

		if err = renameOrCopy(m.from, m.to); err != nil {
			return fmt.Errorf("could not rename/copy file from %q to %q: %w", m.from, m.to, err)
		}
	}
	klog.V(4).Infoln("Move operations are complete")
	return nil
}

func moveAllFiles(fromDir, toDir string, fos []index.FileOperation) error {
	for _, fo := range fos {
		if err := moveFiles(fromDir, toDir, fo); err != nil {
			return fmt.Errorf("failed moving files: %w", err)
		}
	}
	return nil
}

// moveToInstallDir moves plugins from srcDir to dstDir (created in this method) with given FileOperation.
func moveToInstallDir(srcDir, installDir string, fos []index.FileOperation) error {
	installationDir := filepath.Dir(installDir)
	klog.V(4).Infof("Creating directory %q", installationDir)
	if err := os.MkdirAll(installationDir, 0o755); err != nil {
		return fmt.Errorf("error creating directory at %q: %w", installationDir, err)
	}

	tmp, err := os.MkdirTemp("", "krew-temp-move")
	klog.V(4).Infof("Creating temp plugin move operations dir %q", tmp)
	if err != nil {
		return fmt.Errorf("failed to find a temporary directory: %w", err)
	}
	klog.V(4).Infof("Chmoding tmpdir %q to 0755", tmp)
	if err := os.Chmod(tmp, 0o755); err != nil {
		// mktemp gives a 0700 directory but since we move this to KREW_ROOT, we need to make it 0755
		return fmt.Errorf("failed to chmod temp directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tmp); err != nil {
			klog.Warningf("failed to clean up temporary directory %q: %v", tmp, err)
		}
	}()

	if err = moveAllFiles(srcDir, tmp, fos); err != nil {
		return fmt.Errorf("failed to move files: %w", err)
	}

	klog.V(2).Infof("Move directory %q to %q", tmp, installDir)
	if err = renameOrCopy(tmp, installDir); err != nil {
		defer func() {
			klog.V(3).Info("Cleaning up installation directory due to error during copying files")
			os.Remove(installDir)
		}()
		return fmt.Errorf("could not rename/copy directory %q to %q: %w", tmp, installDir, err)
	}
	return nil
}

// renameOrCopy will try to rename a dir or file. If rename is not supported, a manual copy will be performed.
// Existing files at "to" will be deleted.
func renameOrCopy(from, to string) error {
	// Try atomic rename (does not work cross partition).
	fi, err := os.Stat(to)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("error checking move target dir %q: %w", to, err)
	}
	if fi != nil && fi.IsDir() {
		klog.V(4).Infof("There's already a directory at move target %q. deleting.", to)
		if err := os.RemoveAll(to); err != nil {
			return fmt.Errorf("error cleaning up dir %q: %w", to, err)
		}
		klog.V(4).Infof("Move target directory %q cleaned up", to)
	}

	err = os.Rename(from, to)
	// Fallback for invalid cross-device link (errno:18).
	if isCrossDeviceRenameErr(err) {
		klog.V(2).Infof("Cross-device link error while copying, fallback to manual copy")
		if err := copyTree(from, to); err != nil {
			return fmt.Errorf("failed to copy directory tree as a fallback: %w", err)
		}
		return nil
	}
	return err
}

// copyTree copies files or directories, recursively.
func copyTree(from, to string) error {
	return filepath.WalkDir(from, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		newPath, err := pathutil.ReplaceBase(path, from, to)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if d.IsDir() {
			klog.V(4).Infof("Creating new dir %q", newPath)
			return os.MkdirAll(newPath, info.Mode().Perm())
		}
		klog.V(4).Infof("Copying file %q", newPath)
		return copyFile(path, newPath, info.Mode().Perm())
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
	var le *os.LinkError
	if !errors.As(err, &le) {
		return false
	}
	var errno syscall.Errno
	if !errors.As(le.Err, &errno) {
		return false
	}
	return (IsWindows() && errno == 17) || // syscall.ERROR_NOT_SAME_DEVICE
		(!IsWindows() && errno == 18) // syscall.EXDEV
}
