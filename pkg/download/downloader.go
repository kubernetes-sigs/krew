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

package download

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golang/glog"
)

// download gets a file from the internet in memory and writes it content
// to a verifier.
func download(url string, verifier Verifier, fetcher Fetcher) (io.ReaderAt, int64, error) {
	glog.V(2).Infof("Fetching %q", url)
	body, err := fetcher.Get(url)
	if err != nil {
		return nil, 0, fmt.Errorf("could not download %q, err %v: ", url, err)
	}
	defer body.Close()

	glog.V(3).Infof("Reading download data into memory")
	data, err := ioutil.ReadAll(io.TeeReader(body, verifier))
	if err != nil {
		return nil, 0, fmt.Errorf("could not read download content, err %v: ", err)
	}
	glog.V(2).Infof("Read %d bytes of download data into memory", len(data))

	return bytes.NewReader(data), int64(len(data)), verifier.Verify()
}

// extractZIP currently only supports the ZIP format. It will extractZIP
// files into the target directory.
func extractZIP(targetDir string, read io.ReaderAt, size int64) error {
	glog.V(4).Infof("Extracting download zip to %q", targetDir)
	zipReader, err := zip.NewReader(read, size)
	if err != nil {
		return err
	}

	var basepath string
	for _, f := range zipReader.File {
		if basepath == "" {
			basepath = f.Name
		}

		path := filepath.Join(targetDir, filepath.FromSlash(f.Name[len(basepath):]))
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
			continue
		}

		src, err := f.Open()
		if err != nil {
			return fmt.Errorf("could not open inflating zip file, err: %v", err)
		}

		dst, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("can't create file in zip destination dir, err: %v", err)
		}

		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("can't copy content to zip destination file, err: %v", err)
		}

		// Cleanup the open fd. Don't use defer in case of many files.
		// Don't be blocking
		src.Close()
		dst.Close()
	}

	return nil
}

// GetWithSha256 downloads a zip, verifies it and extracts it to the dir.
func GetWithSha256(uri, dir, sha string, fetcher Fetcher) error {
	body, size, err := download(uri, NewSha256Verifier(sha), fetcher)
	if err != nil {
		return err
	}
	return extractZIP(dir, body, size)
}

// GetInsecure downloads a zip and extracts it to the dir.
func GetInsecure(uri, dir string, fetcher Fetcher) error {
	body, size, err := download(uri, NewTrueVerifier(), fetcher)
	if err != nil {
		return err
	}
	return extractZIP(dir, body, size)
}
