package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

const testCategories = `themes:
  - {id: blog, en: Blog, zh-CN: 博客}
  - {id: docs, en: Documentation, zh-CN: 文档}
plugins:
  - {id: comments, en: Comments, zh-CN: 评论}
  - {id: search, en: Search, zh-CN: 搜索}
`

const testReadme = `# The list

<p>
<!-- BEGIN awesome:badges -->
<!-- END awesome:badges -->
</p>

<!-- BEGIN awesome:status -->
<!-- END awesome:status -->

<!-- BEGIN awesome:themes -->
<!-- END awesome:themes -->

<!-- BEGIN awesome:plugins -->
<!-- END awesome:plugins -->
`

// newList makes a list in a temporary folder with the files given, by
// their path in it.
func newList(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	all := map[string]string{categoriesFile: testCategories}
	for _, r := range readmes {
		all[r.File] = testReadme
	}
	for name, body := range files {
		all[name] = body
	}
	for name, body := range all {
		file := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func readFile(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func zipOf(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	var names []string
	for name := range files {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(files[name])); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// fakeGitHub stands in for GitHub's API, its release downloads and its raw
// files.
type fakeGitHub struct {
	srv      *httptest.Server
	repos    map[string]Repository
	releases map[string]Release
	files    map[string][]byte
	broken   map[string]bool // repositories GitHub fails to answer for
}

func newFakeGitHub(t *testing.T) *fakeGitHub {
	f := &fakeGitHub{
		repos:    make(map[string]Repository),
		releases: make(map[string]Release),
		files:    make(map[string][]byte),
		broken:   make(map[string]bool),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/{owner}/{name}", func(w http.ResponseWriter, r *http.Request) {
		repo := r.PathValue("owner") + "/" + r.PathValue("name")
		if f.broken[repo] {
			http.Error(w, "unavailable", http.StatusBadGateway)
			return
		}
		answer(w, f.repos, repo)
	})
	mux.HandleFunc("GET /repos/{owner}/{name}/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		answer(w, f.releases, r.PathValue("owner")+"/"+r.PathValue("name"))
	})
	mux.HandleFunc("/files/{path...}", func(w http.ResponseWriter, r *http.Request) {
		data, ok := f.files[r.PathValue("path")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(data)
	})
	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	return f
}

func answer[T any](w http.ResponseWriter, m map[string]T, key string) {
	v, ok := m[key]
	if !ok {
		http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

func (f *fakeGitHub) client() *GitHub {
	return &GitHub{API: f.srv.URL, Raw: f.srv.URL + "/files/raw", HTTP: f.srv.Client()}
}

// release publishes a release of a repository with one zip, and the files
// of the repository at its tag.
func (f *fakeGitHub) release(repo, tag, zipName string, archive []byte, tree map[string]string) {
	f.repos[repo] = Repository{}
	sum := sha256.Sum256(archive)
	download := "releases/" + repo + "/" + tag + "/" + zipName
	f.files[download] = archive
	for name, body := range tree {
		f.files["raw/"+repo+"/"+tag+"/"+name] = []byte(body)
	}
	f.releases[repo] = Release{
		TagName: tag,
		HTMLURL: "https://github.com/" + repo + "/releases/tag/" + tag,
		Assets: []Asset{{
			Name:   zipName,
			Size:   int64(len(archive)),
			Digest: "sha256:" + hex.EncodeToString(sum[:]),
			URL:    f.srv.URL + "/files/" + download,
		}},
	}
}

// fakeKite writes a kite that says it is version, and passes every theme
// and plugin except one holding a file named refuse.
func fakeKite(t *testing.T, version string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake kite is a shell script")
	}
	file := filepath.Join(t.TempDir(), "kite")
	script := `#!/bin/sh
if [ "$1" = version ]; then
  echo "kite ` + version + ` (abc1234, 2026-09-26T00:00:00Z) go1.26.8 linux/amd64"
  echo "theme api kite/v1, plugin abi 1"
  exit 0
fi
if [ -f "$3/refuse" ]; then
  echo "$1 $(basename "$3") does not load" >&2
  exit 1
fi
echo "$(basename "$3") loads"
`
	if err := os.WriteFile(file, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return file
}
