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
	"time"

	"github.com/apex/gateway"
	"github.com/google/go-github/v80/github"
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
	httpClient        = &http.Client{Timeout: 10 * time.Second}
	knownHomePages    = map[string]string{
		`https://krew.sigs.k8s.io/`:                                  "kubernetes-sigs/krew",
		`https://sigs.k8s.io/krew`:                                   "kubernetes-sigs/krew",
		`https://kubernetes.github.io/ingress-nginx/kubectl-plugin/`: "kubernetes/ingress-nginx",
		`https://kudo.dev/`:                                          "kudobuilder/kudo",
		`https://kubevirt.io`:                                        "kubevirt/kubectl-virt-plugin",
		`https://popeyecli.io`:                                       "derailed/popeye",
		`https://soluble-ai.github.io/kubetap/`:                      "soluble-ai/kubetap",
	}
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
		writeJSON(w, http.StatusInternalServerError, PluginCountResponse{Error: ErrorResponse{Message: fmt.Sprintf("error retrieving repo contents: %v", err)}})
		return
	}
	yamls := filterYAMLs(dir)
	count := len(yamls)
	log.Printf("github response=%s count=%d rate: limit=%d remaining=%d",
		resp.Status, count, resp.Rate.Limit, resp.Rate.Remaining)

	var out PluginCountResponse
	out.Data.Count = count

	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", cacheSeconds))
	writeJSON(w, http.StatusOK, out)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
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
		writeJSON(w, http.StatusInternalServerError, PluginsResponse{Error: ErrorResponse{Message: fmt.Sprintf("error retrieving repo contents: %v", err)}})
		return
	}
	log.Printf("github response=%s rate: limit=%d remaining=%d",
		resp.Status, resp.Rate.Limit, resp.Rate.Remaining)
	var out PluginsResponse

	plugins, err := fetchPlugins(ctx, filterYAMLs(dir))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, PluginsResponse{Error: ErrorResponse{Message: fmt.Sprintf("failed to fetch plugins: %v", err)}})
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
	writeJSON(w, http.StatusOK, out)
}

func fetchPlugins(ctx context.Context, entries []*github.RepositoryContent) ([]*krew.Plugin, error) {
	out := make([]*krew.Plugin, len(entries))
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(pluginFetchWorkers)

	for i, entry := range entries {
		g.Go(func() error {
			p, err := readPlugin(ctx, entry.GetDownloadURL())
			if err != nil {
				return err
			}
			out[i] = p
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}

func readPlugin(ctx context.Context, url string) (*krew.Plugin, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", url, err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get %s: status %s", url, resp.Status)
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", url, err)
	}

	var v krew.Plugin
	if err = yaml.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("failed to parse plugin manifest for %s: %w", url, err)
	}
	return &v, nil
}

func findRepo(homePage string) string {
	if matches := githubRepoPattern.FindStringSubmatch(homePage); matches != nil {
		return matches[1]
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
