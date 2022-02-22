// Copyright 2021 The Kubernetes Authors.
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

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/apex/gateway"
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/krew/pkg/constants"
	krew "sigs.k8s.io/krew/pkg/index"
)

const (
	orgName    = "kubernetes-sigs"
	repoName   = "krew-index"
	pluginsDir = "plugins"

	pluginFetchWorkers = 40
	cacheSeconds       = 60 * 60
)

var (
	githubRepoPattern = regexp.MustCompile(`.*github\.com/([^/]+/[^/#]+)`)
)

type PluginCountResponse struct {
	Data struct {
		Count int `json:"count"`
	} `json:"data"`
	Error ErrorResponse `json:"error,omitempty"`
}

type pluginInfo struct {
	Name             string `json:"name,omitempty"`
	Homepage         string `json:"homepage,omitempty"`
	ShortDescription string `json:"short_description,omitempty"`
	GithubRepo       string `json:"github_repo,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message,omitempty"`
}

type PluginsResponse struct {
	Data struct {
		Plugins []pluginInfo `json:"plugins,omitempty"`
	} `json:"data,omitempty"`
	Error ErrorResponse `json:"error"`
}

func githubClient(ctx context.Context) *github.Client {
	var hc *http.Client

	// if not configured, you should configure a GITHUB_ACCESS_TOKEN
	// variable on Netlify dashboard for the site. You can create a
	// permission-less "personal access token" on GitHub account settings.
	if v := os.Getenv("GITHUB_ACCESS_TOKEN"); v != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: v})
		hc = oauth2.NewClient(ctx, ts)
	}
	return github.NewClient(hc)
}

func pluginCountHandler(w http.ResponseWriter, req *http.Request) {
	_, dir, resp, err := githubClient(req.Context()).
		Repositories.GetContents(req.Context(), orgName, repoName, pluginsDir, &github.RepositoryContentGetOptions{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, PluginCountResponse{Error: ErrorResponse{Message: fmt.Sprintf("error retrieving repo contents: %v", err)}})
		return
	}
	yamls := filterYAMLs(dir)
	count := len(yamls)
	log.Printf("github response=%s count=%d rate: limit=%d remaining=%d",
		resp.Status, count, resp.Rate.Limit, resp.Rate.Remaining)

	var out PluginCountResponse
	out.Data.Count = count

	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", cacheSeconds))
	writeJSON(w, out)
}

func writeJSON(w io.Writer, v interface{}) {
	e := json.NewEncoder(w)
	if err := e.Encode(v); err != nil {
		log.Printf("json write error: %v", err)
	}
}

func filterYAMLs(entries []*github.RepositoryContent) []*github.RepositoryContent {
	var out []*github.RepositoryContent
	for _, v := range entries {
		if v == nil {
			continue
		}
		if v.GetType() == "file" && strings.HasSuffix(v.GetName(), constants.ManifestExtension) {
			out = append(out, v)
		}
	}
	return out
}

func loggingHandler(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		log.Printf("[req]  > method=%s path=%s", req.Method, req.URL)
		defer func() {
			log.Printf("[resp] < method=%s path=%s took=%v", req.Method, req.URL, time.Since(start))
		}()
		f.ServeHTTP(w, req)
	})
}

func pluginsHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	_, dir, resp, err := githubClient(ctx).
		Repositories.GetContents(ctx, orgName, repoName, pluginsDir, &github.RepositoryContentGetOptions{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, PluginsResponse{Error: ErrorResponse{Message: fmt.Sprintf("error retrieving repo contents: %v", err)}})
		return
	}
	log.Printf("github response=%s rate: limit=%d remaining=%d",
		resp.Status, resp.Rate.Limit, resp.Rate.Remaining)
	var out PluginsResponse

	plugins, err := fetchPlugins(ctx, filterYAMLs(dir))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, PluginsResponse{Error: ErrorResponse{Message: fmt.Sprintf("failed to fetch plugins: %v", err)}})
		return
	}

	for _, v := range plugins {
		pi := pluginInfo{
			Name:             v.Name,
			Homepage:         v.Spec.Homepage,
			ShortDescription: v.Spec.ShortDescription,
			GithubRepo:       findRepo(v.Spec.Homepage),
		}
		out.Data.Plugins = append(out.Data.Plugins, pi)
	}

	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", cacheSeconds))
	writeJSON(w, out)
}

func fetchPlugins(ctx context.Context, entries []*github.RepositoryContent) ([]*krew.Plugin, error) {
	var (
		mu  sync.Mutex
		out []*krew.Plugin
	)

	queue := make(chan string)
	g, ctx := errgroup.WithContext(ctx)

	for i := 0; i < pluginFetchWorkers; i++ {
		g.Go(func() error {
			for url := range queue {
				p, err := readPlugin(url)
				if err != nil {
					return err
				}
				mu.Lock()
				out = append(out, p)
				mu.Unlock()
			}
			return nil
		})
	}

	for _, v := range entries {
		url := v.GetDownloadURL()
		select {
		case <-ctx.Done():
			break
		case queue <- url:
		}
	}

	close(queue)

	return out, g.Wait()
}

func readPlugin(url string) (*krew.Plugin, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get %s", url)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read file %s: %w", url)
	}

	var v krew.Plugin
	if err = yaml.Unmarshal(b, &v); err != nil {
		return nil, errors.Wrapf(err, "failed to parse plugin manifest for %s", url)
	}
	return &v, nil
}

func findRepo(homePage string) string {
	if matches := githubRepoPattern.FindStringSubmatch(homePage); matches != nil {
		return matches[1]
	}

	knownHomePages := map[string]string{
		`https://krew.sigs.k8s.io/`:                                  "kubernetes-sigs/krew",
		`https://sigs.k8s.io/krew`:                                   "kubernetes-sigs/krew",
		`https://kubernetes.github.io/ingress-nginx/kubectl-plugin/`: "kubernetes/ingress-nginx",
		`https://kudo.dev/`:                                          "kudobuilder/kudo",
		`https://kubevirt.io`:                                        "kubevirt/kubectl-virt-plugin",
		`https://popeyecli.io`:                                       "derailed/popeye",
		`https://soluble-ai.github.io/kubetap/`:                      "soluble-ai/kubetap",
	}
	return knownHomePages[homePage]
}

func main() {
	port := flag.Int("port", -1, `"to debug locally, set a port number"`)
	flag.Parse()
	local := *port != -1

	mux := http.NewServeMux()
	mux.HandleFunc("/.netlify/functions/api/pluginCount", pluginCountHandler)
	mux.HandleFunc("/.netlify/functions/api/plugins", pluginsHandler)
	if local {
		mux.Handle("/", httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: "localhost:1313"}))
	}

	handler := loggingHandler(mux)
	if local {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), handler))
	}
	log.Fatal(gateway.ListenAndServe("n/a", handler))
}
