package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const searchManifest = `id: search
name: Search
version: 0.1.0
apiVersion: kite/plugin/v1
requires: ">=0.1.0 <2.0.0"
description: Lets readers search every post and page.
author: {name: Kite, url: https://github.com/kite-plus}
license: Apache-2.0
inject:
  - at: body
    html: <script src="{{ asset "search.js" }}" defer></script>
`

const commentsManifest = `id: comments
name: Comments
version: 0.3.0
apiVersion: kite/plugin/v1
description: A thread under every post.
author: {name: Someone, url: https://example.com}
license: MIT
hosts: [giscus.app]
inject:
  - at: body
    html: <script src="https://cdn.jsdelivr.net/npm/x/x.js"></script>
`

const paperManifest = `name: paper
title: Paper
version: 1.2.0
apiVersion: kite/v1
description: A quiet theme for writing.
author: {name: Someone}
license: MIT
tags: [blog]
`

// publishAll publishes a plugin of Kite's and a theme and a plugin of
// someone else's, the theme's zip holding it in a folder, as a
// repository's archive does.
func publishAll(t *testing.T, gh *fakeGitHub) {
	gh.release("kite-plus/kite", "v0.1.0", "kite.zip", nil, nil)
	gh.release("kite-plus/plugin-search", "v0.1.0", "search-0.1.0.zip", zipOf(t, map[string]string{
		"plugin.yaml":       searchManifest,
		"assets/search.js":  "search()",
		"i18n/zh-CN.yaml":   "plugin:\n  title: 站内搜索\n  description: 让读者搜索全部文章和页面。\n",
		"__MACOSX/._plugin": "junk",
	}), nil)
	gh.release("someone/kite-comments", "0.3.0", "comments-0.3.0.zip", zipOf(t, map[string]string{
		"plugin.yaml": commentsManifest,
	}), nil)
	gh.release("someone/kite-theme-paper", "v1.2.0", "paper-1.2.0.zip", zipOf(t, map[string]string{
		"paper/theme.yaml":             paperManifest,
		"paper/layouts/index.html":     "home",
		"paper/screenshot.png":         "png",
		"paper/i18n/zh.yaml":           "theme:\n  title: 纸\n",
		"paper/i18n/ja.yaml":           "theme:\n  title: 紙\n",
		"paper/layouts/_default/x.css": "",
	}), map[string]string{"screenshot.png": "png"})
}

var allEntries = map[string]string{
	"plugins/search.yaml":   "repo: https://github.com/kite-plus/plugin-search\ncategory: search\n",
	"plugins/comments.yaml": "repo: https://github.com/someone/kite-comments\ncategory: comments\n",
	"themes/paper.yaml":     "repo: https://github.com/someone/kite-theme-paper\ncategory: blog\n",
}

func runList(t *testing.T, root string, gh *fakeGitHub, kite string, only ...string) (string, error) {
	t.Helper()
	var out strings.Builder
	err := syncList(context.Background(), root, syncOptions{GitHub: gh.client(), Kite: kite, Only: only, Out: &out})
	return out.String(), err
}

func TestSyncListsWhatPasses(t *testing.T) {
	gh := newFakeGitHub(t)
	publishAll(t, gh)
	root := newList(t, allEntries)

	out, err := runList(t, root, gh, fakeKite(t, "0.1.0"))
	if err != nil {
		t.Fatalf("sync: %v\n%s", err, out)
	}

	var ix Index
	if err := json.Unmarshal([]byte(readFile(t, root, indexFile)), &ix); err != nil {
		t.Fatal(err)
	}
	if ix.Kite != "0.1.0" || len(ix.Plugins) != 2 || len(ix.Themes) != 1 {
		t.Fatalf("index = %+v", ix)
	}
	comments, search, paper := ix.Plugins[0], ix.Plugins[1], ix.Themes[0]
	if !search.Official || comments.Official {
		t.Errorf("official: search %v, comments %v", search.Official, comments.Official)
	}
	if search.Hosts != nil {
		t.Errorf("search loads from %v, want nothing: its own script is not another site", search.Hosts)
	}
	if want := []string{"cdn.jsdelivr.net", "giscus.app"}; strings.Join(comments.Hosts, " ") != strings.Join(want, " ") {
		t.Errorf("comments loads from %v, want %v, declared and injected", comments.Hosts, want)
	}
	if search.I18n["zh-cn"].Title != "站内搜索" {
		t.Errorf("search's packs = %+v", search.I18n)
	}
	if search.Download.SHA256 == "" || !strings.HasSuffix(search.Download.URL, "/search-0.1.0.zip") {
		t.Errorf("search's download = %+v", search.Download)
	}
	if !strings.HasSuffix(paper.Screenshot, "/someone/kite-theme-paper/v1.2.0/screenshot.png") {
		t.Errorf("paper's screenshot = %q", paper.Screenshot)
	}

	en, zh := readFile(t, root, "README.md"), readFile(t, root, "README.zh-CN.md")
	for _, want := range []string{
		"checked with Kite 0.1.0",
		`<a href="https://github.com/kite-plus/kite/releases/tag/v0.1.0"><img src="https://img.shields.io/badge/checked%20with-Kite%200.1.0-4A77D6?logo=data:image/svg%2bxml;base64,`,
		`<a href="#themes"><img src="https://img.shields.io/badge/themes-1-4A77D6" alt="Themes listed: 1"></a>`,
		`<a href="#plugins"><img src="https://img.shields.io/badge/plugins-2-4A77D6" alt="Plugins listed: 2"></a>`,
		"| **[Search](https://github.com/kite-plus/plugin-search)**<br><sub>Search · Official · [0.1.0](https://github.com/kite-plus/plugin-search/releases/tag/v0.1.0)</sub> | Lets readers search every post and page. | None |",
		"<sub>Comments · by [Someone](https://example.com) · [0.3.0]",
		"`cdn.jsdelivr.net`<br>`giscus.app`",
		`<strong><a href="https://github.com/someone/kite-theme-paper">Paper</a></strong><br><sub>Blog · by Someone · <a href=`,
	} {
		if !strings.Contains(en, want) {
			t.Errorf("README.md lacks %q:\n%s", want, en)
		}
	}
	// zh-CN takes the zh pack, the studio's choice, over ja.
	for _, want := range []string{"Kite 0.1.0（目前最新的版本）", "[站内搜索]", "| 无 |", ">纸</a>", "作者 Someone", `href="#插件"`, `alt="用 Kite 0.1.0 检查过"`} {
		if !strings.Contains(zh, want) {
			t.Errorf("README.zh-CN.md lacks %q:\n%s", want, zh)
		}
	}
	// Comments comes before search, as categories.yaml has them.
	if strings.Index(en, "[Comments]") > strings.Index(en, "[Search]") {
		t.Error("the plugins are not in the order of their categories")
	}

	var notes strings.Builder
	if err := runCheck(root, &notes); err != nil || notes.Len() > 0 {
		t.Errorf("check after a sync: %v\n%s", err, notes.String())
	}

	// A second sync with nothing new changes nothing.
	before := readFile(t, root, indexFile) + en + zh
	if out, err := runList(t, root, gh, fakeKite(t, "0.1.0")); err != nil {
		t.Fatalf("sync again: %v\n%s", err, out)
	}
	if after := readFile(t, root, indexFile) + readFile(t, root, "README.md") + readFile(t, root, "README.zh-CN.md"); after != before {
		t.Error("a sync that found nothing new changed the list")
	}
}

func TestSyncLeavesOutWhatDoesNotPass(t *testing.T) {
	cases := map[string]struct {
		files map[string]string // the release's zip
		tree  map[string]string // the repository at the tag
		want  string
	}{
		"another id": {
			files: map[string]string{"plugin.yaml": strings.Replace(commentsManifest, "id: comments", "id: talk", 1)},
			want:  `names the plugin "talk"`,
		},
		"another version": {
			files: map[string]string{"plugin.yaml": strings.Replace(commentsManifest, "0.3.0", "0.2.9", 1)},
			want:  "says version 0.2.9",
		},
		"no license": {
			files: map[string]string{"plugin.yaml": strings.Replace(commentsManifest, "license: MIT\n", "", 1)},
			want:  "has no license",
		},
		"refused by kite": {
			files: map[string]string{"plugin.yaml": commentsManifest, "refuse": ""},
			want:  "kite plugin verify: plugin comments does not load",
		},
		"no manifest": {
			files: map[string]string{"a/plugin.yaml": commentsManifest, "b/x": ""},
			want:  "no plugin.yaml at its top or in a single folder",
		},
		"a path out of it": {
			files: map[string]string{"plugin.yaml": commentsManifest, "../x": ""},
			want:  "leads out of the plugin",
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			gh := newFakeGitHub(t)
			publishAll(t, gh)
			gh.release("someone/kite-comments", "0.3.0", "comments-0.3.0.zip", zipOf(t, c.files), c.tree)
			root := newList(t, allEntries)

			out, err := runList(t, root, gh, fakeKite(t, "0.1.0"))
			if err == nil || !strings.Contains(err.Error(), "plugins/comments.yaml: ") || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("sync said %v, want it to leave out comments for %q\n%s", err, c.want, out)
			}
			ix, err := readIndex(root)
			if err != nil {
				t.Fatal(err)
			}
			if len(ix.Plugins) != 1 || ix.Plugins[0].Name != "search" || len(ix.Themes) != 1 {
				t.Errorf("index = %+v, want search and paper only", ix)
			}
			if readme := readFile(t, root, "README.md"); strings.Contains(readme, "[Comments]") {
				t.Error("the README lists what did not pass")
			}
		})
	}
}

func TestSyncWantsAScreenshotInTheRepository(t *testing.T) {
	gh := newFakeGitHub(t)
	publishAll(t, gh)
	gh.release("someone/kite-theme-paper", "v1.3.0", "paper-1.3.0.zip", zipOf(t, map[string]string{
		"theme.yaml":     strings.Replace(paperManifest, "1.2.0", "1.3.0", 1),
		"screenshot.png": "png",
	}), nil)
	root := newList(t, allEntries)

	_, err := runList(t, root, gh, fakeKite(t, "0.1.0"))
	if err == nil || !strings.Contains(err.Error(), "screenshot.png is in the release's zip, but not in the repository at v1.3.0") {
		t.Fatalf("sync said %v", err)
	}
}

func TestSyncWritesNothingWhenGitHubFails(t *testing.T) {
	gh := newFakeGitHub(t)
	publishAll(t, gh)
	gh.broken["someone/kite-comments"] = true
	root := newList(t, allEntries)

	_, err := runList(t, root, gh, fakeKite(t, "0.1.0"))
	if err == nil || !strings.Contains(err.Error(), "502") {
		t.Fatalf("sync said %v, want GitHub's failure", err)
	}
	if _, err := os.Stat(filepath.Join(root, indexFile)); !os.IsNotExist(err) {
		t.Error("a sync that could not read an entry wrote the index")
	}
}

func TestSyncChecksWithTheLatestKite(t *testing.T) {
	gh := newFakeGitHub(t)
	publishAll(t, gh)
	root := newList(t, allEntries)

	_, err := runList(t, root, gh, fakeKite(t, "v0.1.0-18-g7396dac-dirty"))
	if err == nil || !strings.Contains(err.Error(), "but the latest release is 0.1.0") {
		t.Fatalf("sync with a build from source said %v", err)
	}
}

func TestSyncOfOneEntryKeepsTheOthers(t *testing.T) {
	gh := newFakeGitHub(t)
	publishAll(t, gh)
	root := newList(t, allEntries)
	kite := fakeKite(t, "0.1.0")
	if out, err := runList(t, root, gh, kite); err != nil {
		t.Fatalf("sync: %v\n%s", err, out)
	}

	gh.release("someone/kite-comments", "0.4.0", "comments-0.4.0.zip", zipOf(t, map[string]string{
		"plugin.yaml": strings.Replace(commentsManifest, "0.3.0", "0.4.0", 1),
	}), nil)
	delete(gh.releases, "kite-plus/plugin-search") // not read when not chosen
	if out, err := runList(t, root, gh, kite, "plugins/comments.yaml"); err != nil {
		t.Fatalf("sync of one: %v\n%s", err, out)
	}
	ix, err := readIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(ix.Plugins) != 2 || ix.Plugins[0].Version != "0.4.0" || ix.Plugins[1].Version != "0.1.0" {
		t.Errorf("plugins = %+v", ix.Plugins)
	}
}

func TestDryRunWritesNothing(t *testing.T) {
	gh := newFakeGitHub(t)
	publishAll(t, gh)
	root := newList(t, allEntries)

	var out strings.Builder
	err := syncList(context.Background(), root, syncOptions{
		GitHub: gh.client(), Kite: fakeKite(t, "0.1.0"), DryRun: true,
		Only: []string{"plugins/search"}, Out: &out,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, indexFile)); !os.IsNotExist(err) {
		t.Error("a dry run wrote the index")
	}
	for _, want := range []string{"search 0.1.0", "zh-CN: 站内搜索: 让读者搜索全部文章和页面。", "loads nothing from other sites", "kite: search loads"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the report lacks %q:\n%s", want, out.String())
		}
	}
}

func TestSyncNamesEntriesByTheirFiles(t *testing.T) {
	root := newList(t, allEntries)
	err := syncList(context.Background(), root, syncOptions{Only: []string{"search"}, Out: io.Discard})
	if err == nil || !strings.Contains(err.Error(), "name one by its file") {
		t.Fatalf("sync said %v", err)
	}
}
