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
	"os/exec"
	"strings"

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
type HTTPFetcher struct{}

// Get gets the file and returns an stream to read the file.
func (HTTPFetcher) Get(uri string) (io.ReadCloser, error) {
	klog.V(2).Infof("Fetching %q", uri)
	resp, err := http.Get(uri)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to download %q", uri)
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

var _ Fetcher = CommandFetcher{}

// CommandFetcher is used to run a command to receive a file as stdout
type CommandFetcher struct{}

func (CommandFetcher) Get(cmd string) (io.ReadCloser, error) {
	var stream io.ReadCloser

	// Create tempFile for loading the plugin artifact
	tempFile, err := os.CreateTemp("", "")
	if err != nil {
		return stream, err
	}
	defer klog.V(2).Infof("Removed temp file at %q", tempFile.Name())
	defer os.Remove(tempFile.Name())

	klog.V(2).Infof("Created temp file for command output at %q", tempFile.Name())

	// Intentionally not closing the stream object in this function!
	// The open file to tempFile will remain readable until the last process
	// releases it. Another function is responsible for closing the
	// io.ReaderCloser.
	stream, err = os.Open(tempFile.Name())
	if err != nil {
		return stream, err
	}

	// HACK REMOVEME TESTING ONLY
	// This is a workaround to accept command within the `uri` key of the krew
	// plugin spec. An alternative solution, which is likely more optimal, is
	// adding a new field to the spec for `downloadCommand` or similarly named
	// key which would contain a command to run. This hack was put in place to
	// test the rough implementation of loading a plugin from the stdout of a
	// command.
	cmd = strings.Replace(cmd, "cmd://", "", 1)

	// TODO Improve splitting, this implementation has issues with newlines and
	// qoutes in the cmd string.
	c := strings.Split(cmd, " ")
	runner := exec.Command(c[0], c[1:]...)

	// NOTE Attempted to pass runner.Stdout() as an io.ReaderCloser but the io
	// closes as soon as the application finishes running, which cannot extend
	// reading beyond this function. Instead opted to push this into a tempFile
	// on the local filesystem.
	// Send stdout to tempFile for later ingestion
	runner.Stdout = tempFile

	klog.V(2).Infof("Running command %q", cmd)
	if err := runner.Run(); err != nil {
		// TODO It would be helpful to have more diagnostic information like the
		// stdout and stderr of the command if Run() fails to exit 0. Right now the
		// Command err just says things like `exited code 1` or similarly vague
		// output. Could use klog.V(N)... to log output with verbosity.
		return stream, errors.Wrapf(err, "failed to run command: %s", cmd)
	}

	return stream, err
}
