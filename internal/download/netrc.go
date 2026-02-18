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
	"net"
	"net/url"
	"os"
	"path/filepath"

	"github.com/jdx/go-netrc"
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
		return nil, err
	}

	var netrcPath string
	if netrcFile != "" {
		netrcPath = netrcFile
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		netrcPath = filepath.Join(homeDir, ".netrc")
	}

	n, err := netrc.Parse(netrcPath)
	if err != nil {
		return nil, err
	}

	// Try exact match first
	if machine := n.Machine(u.Host); machine != nil {
		return &NetrcEntry{
			Machine:  u.Host,
			Login:    machine.Get("login"),
			Password: machine.Get("password"),
		}, nil
	}

	// Try without port
	if host, _, err := net.SplitHostPort(u.Host); err == nil {
		if machine := n.Machine(host); machine != nil {
			return &NetrcEntry{
				Machine:  host,
				Login:    machine.Get("login"),
				Password: machine.Get("password"),
			}, nil
		}
	}

	return nil, nil
}
