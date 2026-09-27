package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// GitHub reads repositories and releases. The addresses are fields so that
// tests can stand in for GitHub.
type GitHub struct {
	API   string
	Raw   string
	Token string
	HTTP  *http.Client
}

func newGitHub() *GitHub {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	return &GitHub{
		API:   "https://api.github.com",
		Raw:   "https://raw.githubusercontent.com",
		Token: token,
		HTTP:  &http.Client{Timeout: 2 * time.Minute},
	}
}

const userAgent = "kite-plus-awesome"

var errNotFound = errors.New("not found")

type Repository struct {
	Archived bool `json:"archived"`
}

type Release struct {
	TagName string  `json:"tag_name"`
	HTMLURL string  `json:"html_url"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	// Digest is "sha256:<hex>" for assets uploaded since GitHub began
	// recording one, and empty for older ones.
	Digest string `json:"digest"`
	URL    string `json:"browser_download_url"`
}

func (g *GitHub) Repository(ctx context.Context, owner, name string) (*Repository, error) {
	var r Repository
	return &r, g.getJSON(ctx, "/repos/"+owner+"/"+name, &r)
}

// LatestRelease is the newest release that is neither a draft nor a
// pre-release.
func (g *GitHub) LatestRelease(ctx context.Context, owner, name string) (*Release, error) {
	var r Release
	return &r, g.getJSON(ctx, "/repos/"+owner+"/"+name+"/releases/latest", &r)
}

func (g *GitHub) getJSON(ctx context.Context, p string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.API+p, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", userAgent)
	if g.Token != "" {
		req.Header.Set("Authorization", "Bearer "+g.Token)
	}
	resp, err := g.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return json.NewDecoder(resp.Body).Decode(v)
	case http.StatusNotFound:
		return errNotFound
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return fmt.Errorf("GET %s: %s: %s", p, resp.Status, bytes.TrimSpace(body))
}

// Download fetches a release asset, refusing one larger than limit. The
// token is not sent: assets of public repositories need none, and the
// download is redirected to another host.
func (g *GitHub) Download(ctx context.Context, a Asset, limit int64) ([]byte, error) {
	if a.Size > limit {
		return nil, refuse("%s is %d MB, more than the %d MB the studio accepts", a.Name, a.Size>>20, limit>>20)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", userAgent)
	resp, err := g.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading %s: %s", a.URL, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %w", a.URL, err)
	}
	if int64(len(data)) > limit {
		return nil, refuse("%s is more than the %d MB the studio accepts", a.Name, limit>>20)
	}
	return data, nil
}

// RawURL is the address of a file of a repository at a tag.
func (g *GitHub) RawURL(owner, name, tag, file string) string {
	parts := strings.Split(file, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return g.Raw + "/" + owner + "/" + name + "/" + url.PathEscape(tag) + "/" + strings.Join(parts, "/")
}

// Exists reports whether an address answers with a file.
func (g *GitHub) Exists(ctx context.Context, address string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, address, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := g.HTTP.Do(req)
	if err != nil {
		return false, err
	}
	resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	}
	return false, fmt.Errorf("HEAD %s: %s", address, resp.Status)
}
