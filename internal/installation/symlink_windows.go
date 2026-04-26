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

//go:build windows

package installation

import (
	"encoding/binary"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

const (
	ioReparseTagMountPoint = 0xA0000003
	fsctlSetReparsePoint   = 0x000900A4
)

// createSymlink creates a directory junction from newname pointing to oldname.
// Unlike os.Symlink, junctions do not require elevated privileges or Developer
// Mode on Windows.
func createSymlink(oldname, newname string) error {
	target, err := filepath.Abs(oldname)
	if err != nil {
		return errors.Wrapf(err, "failed to resolve absolute path for %q", oldname)
	}

	ntTarget := `\??\` + target
	ntTargetUTF16, err := windows.UTF16FromString(ntTarget)
	if err != nil {
		return errors.Wrapf(err, "failed to encode target path to UTF-16")
	}

	if err := os.Mkdir(newname, 0o750); err != nil {
		return errors.Wrapf(err, "failed to create junction directory %q", newname)
	}

	dirPtr, err := windows.UTF16PtrFromString(newname)
	if err != nil {
		os.Remove(newname)
		return errors.Wrapf(err, "failed to encode junction path to UTF-16")
	}

	handle, err := windows.CreateFile(
		dirPtr,
		windows.GENERIC_WRITE,
		0,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		os.Remove(newname)
		return errors.Wrapf(err, "failed to open directory %q", newname)
	}
	defer windows.CloseHandle(handle)

	buf := buildJunctionReparseBuffer(ntTargetUTF16)

	var bytesReturned uint32
	if err := windows.DeviceIoControl(
		handle,
		fsctlSetReparsePoint,
		&buf[0],
		uint32(len(buf)),
		nil,
		0,
		&bytesReturned,
		nil,
	); err != nil {
		os.Remove(newname)
		return errors.Wrapf(err, "failed to create junction %q -> %q", newname, target)
	}

	return nil
}

// isLink reports whether fi describes a symlink or a directory junction.
// Go 1.22+ no longer sets ModeSymlink for junctions; they get ModeIrregular instead.
func isLink(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeSymlink != 0 || fi.Mode()&os.ModeIrregular != 0
}

// buildJunctionReparseBuffer constructs a REPARSE_DATA_BUFFER for a mount point
// (directory junction). The buffer layout is:
//
//	Offset  Size  Field
//	0       4     ReparseTag
//	4       2     ReparseDataLength
//	6       2     Reserved
//	8       2     SubstituteNameOffset
//	10      2     SubstituteNameLength
//	12      2     PrintNameOffset
//	14      2     PrintNameLength
//	16      var   PathBuffer (SubstituteName + null + PrintName + null)
func buildJunctionReparseBuffer(ntTargetUTF16 []uint16) []byte {
	substNameLen := (len(ntTargetUTF16) - 1) * 2
	printNameOffset := substNameLen + 2
	pathBufferLen := substNameLen + 2 + 2
	reparseDataLen := 8 + pathBufferLen
	bufLen := 8 + reparseDataLen

	buf := make([]byte, bufLen)
	binary.LittleEndian.PutUint32(buf[0:], ioReparseTagMountPoint)
	binary.LittleEndian.PutUint16(buf[4:], uint16(reparseDataLen))
	binary.LittleEndian.PutUint16(buf[8:], 0)
	binary.LittleEndian.PutUint16(buf[10:], uint16(substNameLen))
	binary.LittleEndian.PutUint16(buf[12:], uint16(printNameOffset))
	binary.LittleEndian.PutUint16(buf[14:], 0)

	for i, c := range ntTargetUTF16 {
		binary.LittleEndian.PutUint16(buf[16+i*2:], c)
	}

	return buf
}
