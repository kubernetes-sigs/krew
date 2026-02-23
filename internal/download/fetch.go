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
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

// Fetcher is used to get files from a URI.
type Fetcher interface {
	// Get gets the file and returns an stream to read the file.
	Get(uri string) (io.ReadCloser, error)
}

var _ Fetcher = HTTPFetcher{}

// HTTPFetcher is used to get a file from a http:// or https:// schema path.
type HTTPFetcher struct {
	EnableNetrc bool
	NetrcFile   string
}

// Get gets the file and returns an stream to read the file.
func (f HTTPFetcher) Get(uri string) (io.ReadCloser, error) {
	klog.V(2).Infof("Fetching %q", uri)

	req, err := http.NewRequest("GET", uri, http.NoBody)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create request for %q", uri)
	}

	// Check for netrc credentials
	if f.EnableNetrc {
		entry, err := FindNetrcEntry(uri, f.NetrcFile)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load netrc credentials")
		}
		if entry != nil {
			klog.V(3).Infof("Using netrc credentials for %s", entry.Machine)
			req.SetBasicAuth(entry.Login, entry.Password)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to download %q", uri)
	}
	if resp.StatusCode > 200 {
		resp.Body.Close()
		return nil, errors.Errorf("failed to download %q, status code %d", uri, resp.StatusCode)
	}
	return resp.Body, nil
}

var _ Fetcher = fileFetcher{}

type fileFetcher struct{ f string }

func (f fileFetcher) Get(_ string) (io.ReadCloser, error) {
	klog.V(2).Infof("Reading %q", f.f)
	file, err := os.Open(f.f)
	return file, errors.Wrapf(err, "failed to open archive file %q for reading", f.f)
}

// NewFileFetcher returns a local file reader.
func NewFileFetcher(path string) Fetcher { return fileFetcher{f: path} }
