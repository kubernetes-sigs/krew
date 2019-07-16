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

// plugin-overview reads the manifests in a directory and creates a markdown overview page
package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	"sigs.k8s.io/krew/pkg/index"
	"sigs.k8s.io/krew/pkg/index/indexscanner"
)

const (
	separator  = " | "
	pageHeader = `# Plugin overview

Below is a table of plugins available in the central krew-index.

`
)

var (
	githubRepoPattern = regexp.MustCompile(`.*github.com/([^/]+/[^/#]+)`)
)

func main() {
	pluginsDir := pflag.String("plugins-dir", "", "The directory containing the plugin manifests")
	pflag.Parse()

	if *pluginsDir == "" {
		pflag.Usage()
		return
	}

	plugins, err := indexscanner.LoadPluginListFromFS(*pluginsDir)
	if err != nil {
		glog.Fatal(err)
	}

	out := os.Stdout

	_, _ = fmt.Fprintln(out, pageHeader)

	printTableHeader(out)
	for _, p := range plugins {
		printTableRowForPlugin(out, &p)
	}
}

func printTableHeader(out io.Writer) {
	printRow(out, "Name", "Description", "GitHub Stars")
	printRow(out, "----", "-----------", "------------")
}

func printTableRowForPlugin(out io.Writer, p *index.Plugin) {
	// 1st column
	name := p.Name
	if homepage := p.Spec.Homepage; homepage != "" {
		name = fmt.Sprintf("[%s](%s)", strings.TrimSpace(name), homepage)
	}

	// 2nd column
	description := strings.TrimSpace(p.Spec.ShortDescription)

	// 3rd column
	shield := makeGithubShield(p.Spec.Homepage)

	printRow(out, name, description, shield)
}

func makeGithubShield(homepage string) string {
	repo := ""

	if matches := githubRepoPattern.FindStringSubmatch(homepage); matches != nil {
		repo = matches[1]
	} else if homepage == `https://sigs.k8s.io/krew` {
		repo = "kubernetes-sigs/krew"
	} else if homepage == `https://kubernetes.github.io/ingress-nginx/kubectl-plugin/` {
		repo = "kubernetes/ingress-nginx"
	}

	if repo == "" {
		return ""
	}
	return "![GitHub stars](https://img.shields.io/github/stars/" + repo + ".svg?label=github%20stars&logo=github)"
}

func printRow(w io.Writer, cols ...string) {
	_, _ = fmt.Fprintln(w, strings.Join(cols, separator))
}
