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

package download

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

// download gets a file from the internet in memory and writes it content
// to a Verifier.
func download(url string, verifier Verifier, fetcher Fetcher) (io.ReaderAt, int64, error) {
	body, err := fetcher.Get(url)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "failed to obtain plugin archive")
	}
	defer body.Close()

	klog.V(3).Infof("Reading archive file into memory")
	data, err := io.ReadAll(io.TeeReader(body, verifier))
	if err != nil {
		return nil, 0, errors.Wrap(err, "could not read archive")
	}
	klog.V(2).Infof("Read %d bytes from archive into memory", len(data))

	return bytes.NewReader(data), int64(len(data)), verifier.Verify()
}

// extractZIP extracts a zip file into the target directory.
func extractZIP(targetDir string, read io.ReaderAt, size int64) error {
	klog.V(4).Infof("Extracting zip archive to %q", targetDir)
	zipReader, err := zip.NewReader(read, size)
	if err != nil {
		return err
	}

	for _, f := range zipReader.File {
		if err := suspiciousPath(f.Name); err != nil {
			return err
		}

		path := filepath.Join(targetDir, filepath.FromSlash(f.Name))
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, f.Mode()); err != nil {
				return errors.Wrap(err, "can't create directory tree")
			}
			continue
		}

		dir := filepath.Dir(path)
		klog.V(4).Infof("zip: ensuring parent dirs exist for regular file, dir=%s", dir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return errors.Wrap(err, "failed to create directory for zip entry")
		}
		src, err := f.Open()
		if err != nil {
			return errors.Wrap(err, "could not open inflating zip file")
		}

		dst, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			src.Close()
			return errors.Wrap(err, "can't create file in zip destination dir")
		}
		closeAll := func() {
			src.Close()
			dst.Close()
		}

		if _, err := io.Copy(dst, src); err != nil {
			closeAll()
			return errors.Wrap(err, "can't copy content to zip destination file")
		}
		closeAll()
	}

	return nil
}

// extractTARGZ extracts a gzipped tar file into the target directory.
func extractTARGZ(targetDir string, at io.ReaderAt, size int64) error {
	klog.V(4).Infof("tar: extracting to %q", targetDir)
	in := io.NewSectionReader(at, 0, size)

	gzr, err := gzip.NewReader(in)
	if err != nil {
		return errors.Wrap(err, "failed to create gzip reader")
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "tar extraction error")
		}
		klog.V(4).Infof("tar: processing %q (type=%d, mode=%s)", hdr.Name, hdr.Typeflag, os.FileMode(hdr.Mode))
		// see https://golang.org/cl/78355 for handling pax_global_header
		if hdr.Name == "pax_global_header" {
			klog.V(4).Infof("tar: skipping pax_global_header file")
			continue
		}

		if err := suspiciousPath(hdr.Name); err != nil {
			return err
		}

		path := filepath.Join(targetDir, filepath.FromSlash(hdr.Name))
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, os.FileMode(hdr.Mode)); err != nil {
				return errors.Wrap(err, "failed to create directory from tar")
			}
		case tar.TypeReg:
			dir := filepath.Dir(path)
			klog.V(4).Infof("tar: ensuring parent dirs exist for regular file, dir=%s", dir)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return errors.Wrap(err, "failed to create directory for tar")
			}
			f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.FileMode(hdr.Mode))
			if err != nil {
				return errors.Wrapf(err, "failed to create file %q", path)
			}

			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return errors.Wrapf(err, "failed to copy %q from tar into file", hdr.Name)
			}
			f.Close()
		default:
			return errors.Errorf("unable to handle file type %d for %q in tar", hdr.Typeflag, hdr.Name)
		}
		klog.V(4).Infof("tar: processed %q", hdr.Name)
	}
	klog.V(4).Infof("tar extraction to %s complete", targetDir)
	return nil
}

func suspiciousPath(path string) error {
	if strings.Contains(path, "..") {
		return errors.Errorf("refusing to unpack archive with suspicious entry %q", path)
	}

	if strings.HasPrefix(path, `/`) || strings.HasPrefix(path, `\`) {
		return errors.Errorf("refusing to unpack archive with absolute entry %q", path)
	}

	return nil
}

func detectMIMEType(at io.ReaderAt) (string, error) {
	buf := make([]byte, 512)
	n, err := at.ReadAt(buf, 0)
	if err != nil && err != io.EOF {
		return "", errors.Wrap(err, "failed to read first 512 bytes")
	}
	if n < 512 {
		klog.V(5).Infof("Did only read %d of 512 bytes to determine the file type", n)
	}

	// Cut off mime extra info beginning with ';' i.e:
	// "text/plain; charset=utf-8" should result in "text/plain".
	return strings.Split(http.DetectContentType(buf[:n]), ";")[0], nil
}

type extractor func(targetDir string, read io.ReaderAt, size int64) error

var defaultExtractors = map[string]extractor{
	"application/zip":    extractZIP,
	"application/x-gzip": extractTARGZ,
}

func extractArchive(dst string, at io.ReaderAt, size int64) error {
	t, err := detectMIMEType(at)
	if err != nil {
		return errors.Wrap(err, "failed to determine content type")
	}
	klog.V(4).Infof("detected %q file type", t)
	exf, ok := defaultExtractors[t]
	if !ok {
		return errors.Errorf("mime type %q for archive file is not a supported archive format", t)
	}
	return errors.Wrap(exf(dst, at, size), "failed to extract file")
}

// Downloader is responsible for fetching, verifying and extracting a binary.
type Downloader struct {
	verifier Verifier
	fetcher  Fetcher
}

// NewDownloader builds a new Downloader.
func NewDownloader(v Verifier, f Fetcher) Downloader {
	return Downloader{
		verifier: v,
		fetcher:  f,
	}
}

// Get pulls the uri and verifies it. On success, the download gets extracted
// into dst.
func (d Downloader) Get(uri, dst string) error {
	body, size, err := download(uri, d.verifier, d.fetcher)
	if err != nil {
		return err
	}
	return extractArchive(dst, body, size)
}
