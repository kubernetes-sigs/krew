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

package internal

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// DownloadFile function takes URL as input, stores the contents in temporary file and returns filename, cleanup function to delete temporary files and error
func DownloadFile(url string) (string, func(), error) {

	tmpFile, err := ioutil.TempFile(os.TempDir(), "krew-")

	cleanup := func() { os.Remove(tmpFile.Name()) }
	fileName := tmpFile.Name()

	if err != nil {
		return fileName, cleanup, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return fileName, cleanup, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fileName, cleanup, fmt.Errorf("bad status: %s", resp.Status)
	}

	if _, err = io.Copy(tmpFile, resp.Body); err != nil {
		return fileName, cleanup, err
	}

	// Close the file
	if err := tmpFile.Close(); err != nil {
		return fileName, cleanup, err
	}

	return fileName, cleanup, nil
}
