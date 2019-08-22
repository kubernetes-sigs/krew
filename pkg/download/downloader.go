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
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// download gets a file from the internet in memory and writes it content
// to a Verifier.
func download(url string, verifier Verifier, fetcher Fetcher) (io.ReaderAt, int64, error) {
	glog.V(2).Infof("Fetching %q", url)
	body, err := fetcher.Get(url)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "could not download %q", url)
	}
	defer body.Close()

	glog.V(3).Infof("Reading download data into memory")
	data, err := ioutil.ReadAll(io.TeeReader(body, verifier))
	if err != nil {
		return nil, 0, errors.Wrap(err, "could not read download content")
	}
	glog.V(2).Infof("Read %d bytes of download data into memory", len(data))

	return bytes.NewReader(data), int64(len(data)), verifier.Verify()
}

// extractZIP extracts a zip file into the target directory.
func extractZIP(targetDir string, read io.ReaderAt, size int64) error {
	glog.V(4).Infof("Extracting download zip to %q", targetDir)
	zipReader, err := zip.NewReader(read, size)
	if err != nil {
		return err
	}

	for _, f := range zipReader.File {
		path := filepath.Join(targetDir, filepath.FromSlash(f.Name))
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
			continue
		}

		src, err := f.Open()
		if err != nil {
			return errors.Wrap(err, "could not open inflating zip file")
		}

		dst, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			return errors.Wrap(err, "can't create file in zip destination dir")
		}
		close := func() {
			src.Close()
			dst.Close()
		}

		if _, err := io.Copy(dst, src); err != nil {
			close()
			return errors.Wrap(err, "can't copy content to zip destination file")
		}
		close()
	}

	return nil
}

// extractTARGZ extracts a gzipped tar file into the target directory.
func extractTARGZ(targetDir string, at io.ReaderAt, size int64) error {
	glog.V(4).Infof("tar: extracting to %q", targetDir)
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
		glog.V(4).Infof("tar: processing %q (type=%d, mode=%s)", hdr.Name, hdr.Typeflag, os.FileMode(hdr.Mode))
		// see https://golang.org/cl/78355 for handling pax_global_header
		if hdr.Name == "pax_global_header" {
			glog.V(4).Infof("tar: skipping pax_global_header file")
			continue
		}

		path := filepath.Join(targetDir, filepath.FromSlash(hdr.Name))
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, os.FileMode(hdr.Mode)); err != nil {
				return errors.Wrap(err, "failed to create directory from tar")
			}
		case tar.TypeReg:
			dir := filepath.Dir(path)
			glog.V(4).Infof("tar: ensuring parent dirs exist for regular file, dir=%s", dir)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return errors.Wrap(err, "failed to create directory for tar")
			}
			f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.FileMode(hdr.Mode))
			if err != nil {
				return errors.Wrapf(err, "failed to create file %q", path)
			}
			close := func() { f.Close() }
			if _, err := io.Copy(f, tr); err != nil {
				close()
				return errors.Wrapf(err, "failed to copy %q from tar into file", hdr.Name)
			}
			close()
		default:
			return errors.Errorf("unable to handle file type %d for %q in tar", hdr.Typeflag, hdr.Name)
		}
		glog.V(4).Infof("tar: processed %q", hdr.Name)
	}
	glog.V(4).Infof("tar extraction to %s complete", targetDir)
	return nil
}

func detectMIMEType(at io.ReaderAt) (string, error) {
	buf := make([]byte, 512)
	n, err := at.ReadAt(buf, 0)
	if err != nil && err != io.EOF {
		return "", errors.Wrap(err, "failed to read first 512 bytes")
	}
	if n < 512 {
		glog.V(5).Infof("Did only read %d of 512 bytes to determine the file type", n)
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
	// TODO(ahmetb) This package is not architected well, this method should not
	// be receiving this many args. Primary problem is at GetInsecure and
	// GetWithSha256 methods that embed extraction in them, which is orthogonal.

	t, err := detectMIMEType(at)
	if err != nil {
		return errors.Wrap(err, "failed to determine content type")
	}
	glog.V(4).Infof("detected %q file type", t)
	exf, ok := defaultExtractors[t]
	if !ok {
		return errors.Errorf("mime type %q for downloaded file is not a supported archive format", t)
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
		return errors.Wrapf(err, "failed to get the uri %q", uri)
	}
	return extractArchive(dst, body, size)
}
