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
	"net/url"
	"os"

	"github.com/git-lfs/go-netrc/netrc"
	"github.com/pkg/errors"
)

// NetrcEntry represents a single entry in .netrc
type NetrcEntry struct {
	Machine  string
	Login    string
	Password string
}

// FindNetrcEntry finds the netrc entry for a given URL
func FindNetrcEntry(uri, netrcFile string) (*NetrcEntry, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse URL %q", uri)
	}

	file, err := os.Open(netrcFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open netrc file %q", netrcFile)
	}
	defer file.Close()

	n, err := netrc.Parse(file)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse netrc file %q", netrcFile)
	}

	// Use FindMachine which handles host:port matching automatically
	// Pass empty string for login name since we don't have a specific login to match
	if machine := n.FindMachine(u.Hostname(), ""); machine != nil {
		// Ensure both login and password are present
		if machine.Login != "" && machine.Password != "" {
			return &NetrcEntry{
				Machine:  machine.Name,
				Login:    machine.Login,
				Password: machine.Password,
			}, nil
		}
	}

	return nil, nil
}
