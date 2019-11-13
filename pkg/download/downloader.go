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
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/mholt/archiver"
	"github.com/pkg/errors"
)

// download gets a file from the internet in memory and writes it content
// to a Verifier.
func download(url string, verifier Verifier, fetcher Fetcher) ([]byte, error) {
	glog.V(2).Infof("Fetching %q", url)
	body, err := fetcher.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "could not download %q", url)
	}
	defer body.Close()

	glog.V(3).Infof("Reading download data into memory")
	data, err := ioutil.ReadAll(io.TeeReader(body, verifier))
	if err != nil {
		return nil, errors.Wrap(err, "could not read download content")
	}
	glog.V(2).Infof("Read %d bytes of download data into memory", len(data))

	return data, verifier.Verify()
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

func isSuspiciousArchive(path string) error {
	return archiver.Walk(path, func(f archiver.File) error {
		switch h := f.Header.(type) {
		case *tar.Header:
			return suspiciousPath(h.Name)
		case zip.FileHeader:
			return suspiciousPath(h.Name)
		default:
			return errors.Errorf("Unknow header type: %T", h)
		}
	})
}

func detectMIMEType(data []byte) string {
	n := 512
	if l := len(data); l < n {
		n = l
	}
	// Cut off mime extra info beginning with ';' i.e:
	// "text/plain; charset=utf-8" should result in "text/plain".
	return strings.Split(http.DetectContentType(data[:n]), ";")[0]
}

func extensionFromMIME(mime string) (string, error) {
	switch mime {
	case "application/zip":
		return "zip", nil
	case "application/x-gzip":
		return "tar.gz", nil
	default:
		return "", errors.Errorf("unknown mime type to extract: %q", mime)
	}
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
	data, err := download(uri, d.verifier, d.fetcher)
	if err != nil {
		return err
	}
	extension, err := extensionFromMIME(detectMIMEType(data))
	if err != nil {
		return err
	}

	f, err := ioutil.TempFile("", "plugin.*."+extension)
	if err != nil {
		return errors.Wrap(err, "failed to create temp file to write")
	}
	defer os.Remove(f.Name())
	if n, err := f.Write(data); err != nil {
		return errors.Wrap(err, "failed to write temp download file")
	} else if n != len(data) {
		return errors.Errorf("failed to write whole download archive")
	}

	if err := isSuspiciousArchive(f.Name()); err != nil {
		return err
	}
	return archiver.Unarchive(f.Name(), dst)
}
