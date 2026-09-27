package main

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestWhatAManifestSaysIsReadAsWords(t *testing.T) {
	ix := &Index{Kite: "0.1.0", Plugins: []Record{{
		Name: "x", Repo: "https://github.com/a/x", Category: "search",
		Version: "1.0.0", Release: "https://github.com/a/x/releases/tag/v1.0.0",
		Title:       "X | <b>Y</b>",
		Description: "Finds [things](https://evil.example) *fast* <!-- END awesome:plugins -->\nwith a | pipe.",
		Author:      &Author{Name: "Mallory", URL: "javascript:alert(1)"},
	}}, Themes: []Record{{
		Name: "t", Repo: "https://github.com/a/t", Category: "blog",
		Version: "1.0.0", Release: "https://github.com/a/t/releases/tag/v1.0.0",
		Screenshot: `https://raw.example/"onerror="x`, Title: `T"><script>`, Description: "<img src=x>",
	}}}
	cats := Categories{
		Themes:  []Category{{ID: "blog", En: "Blog", ZhCN: "博客"}},
		Plugins: []Category{{ID: "search", En: "Search", ZhCN: "搜索"}},
	}
	blocks := readmes[0].blocks(ix, cats)

	plugins := blocks["plugins"]
	for _, want := range []string{
		`[X \| &lt;b&gt;Y&lt;/b&gt;](https://github.com/a/x)`,
		`Finds \[things\](https://evil.example) \*fast\* &lt;!-- END awesome:plugins --&gt; with a \| pipe.`,
		"by Mallory ·",
	} {
		if !strings.Contains(plugins, want) {
			t.Errorf("the plugins lack %q:\n%s", want, plugins)
		}
	}
	if strings.Count(plugins, "\n") != 2 {
		t.Errorf("a description spilled onto another line:\n%s", plugins)
	}

	themes := blocks["themes"]
	for _, bad := range []string{"<script>", "<img src=x>", `"onerror="`} {
		if strings.Contains(themes, bad) {
			t.Errorf("the themes let %q through:\n%s", bad, themes)
		}
	}
}

func TestShieldsWritesDashesTwice(t *testing.T) {
	got := shield("checked with", "Kite 0.2.0-rc.1_x")
	if want := "https://img.shields.io/badge/checked%20with-Kite%200.2.0--rc.1__x-4A77D6"; got != want {
		t.Errorf("shield = %s, want %s", got, want)
	}
}

func TestReplaceBlockNeedsItsMarkers(t *testing.T) {
	doc := "a\n<!-- BEGIN awesome:status -->\nold\n<!-- END awesome:status -->\nb\n"
	got, err := replaceBlock(doc, "status", "new")
	if err != nil || got != "a\n<!-- BEGIN awesome:status -->\nnew\n<!-- END awesome:status -->\nb\n" {
		t.Errorf("replaceBlock = %q, %v", got, err)
	}
	if _, err := replaceBlock("no markers", "status", "new"); err == nil {
		t.Error("replaceBlock wrote a block into a README without its markers")
	}
}

func TestCheckFindsAnEditedList(t *testing.T) {
	root := newList(t, nil)
	var notes strings.Builder
	if err := runCheck(root, &notes); err == nil {
		t.Fatal("check passed READMEs whose lists were never written")
	}
	if err := runReadme(root); err != nil {
		t.Fatal(err)
	}
	if err := runCheck(root, &notes); err != nil {
		t.Fatalf("check after readme: %v", err)
	}

	edited := strings.Replace(readFile(t, root, "README.md"), "No plugin is listed yet.", "My plugin.", 1)
	root2 := newList(t, map[string]string{"README.md": edited, "README.zh-CN.md": readFile(t, root, "README.zh-CN.md")})
	err := runCheck(root2, &notes)
	if err == nil || !strings.Contains(err.Error(), "README.md: the lists in it are not what index.json makes") || strings.Contains(err.Error(), "README.zh-CN.md") {
		t.Errorf("check of a README edited by hand said %v", err)
	}
}

func TestCheckNotesWhatTheNextSyncSettles(t *testing.T) {
	root := newList(t, map[string]string{
		"plugins/search.yaml": "repo: https://github.com/kite-plus/plugin-search\ncategory: search\n",
		indexFile: `{"kite": "0.1.0", "themes": [], "plugins": [{"name": "gone", "repo": "https://github.com/a/gone",
			"category": "search", "official": false, "version": "1.0.0", "release": "r",
			"download": {"url": "u", "sha256": "s", "size": 1}, "apiVersion": "kite/plugin/v1",
			"title": "Gone", "description": "d", "license": "MIT"}]}`,
	})
	if err := runReadme(root); err != nil {
		t.Fatal(err)
	}
	var notes strings.Builder
	if err := runCheck(root, &notes); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"plugins/search.yaml: not listed yet", "plugins/gone.yaml: gone, but still listed"} {
		if !strings.Contains(notes.String(), want) {
			t.Errorf("the notes lack %q:\n%s", want, notes.String())
		}
	}
}

func TestWordsArePickedAsTheStudioPicksThem(t *testing.T) {
	cases := []struct {
		packs map[string]Words
		want  string
	}{
		{map[string]Words{"zh-cn": {Title: "简"}, "zh": {Title: "中"}}, "简"},
		{map[string]Words{"zh": {Title: "中"}, "zh-tw": {Title: "繁"}}, "中"},
		{map[string]Words{"zh-tw": {Title: "繁"}, "ja": {Title: "日"}}, "繁"},
		// The pack for the language is the one used, even without a title.
		{map[string]Words{"zh-cn": {Description: "只有说明"}, "zh": {Title: "中"}}, "Manifest"},
		{nil, "Manifest"},
	}
	for _, c := range cases {
		r := Record{Title: "Manifest", I18n: c.packs}
		if got := r.Words("zh-CN").Title; got != c.want {
			t.Errorf("with %v, zh-CN reads %q, want %q", c.packs, got, c.want)
		}
	}
}

func TestReadPackagesTheWayKiteDoes(t *testing.T) {
	theme := fstest.MapFS{
		"theme.yaml":         {Data: []byte("name: paper\ntitle: Paper\nscreenshot: shots/home.webp\n")},
		"shots/home.webp":    {Data: []byte("webp")},
		"screenshot.png":     {Data: []byte("png")},
		"i18n/zh-CN.yaml":    {Data: []byte("theme:\n  title: 纸\n  settings: {accent: {label: 强调色}}\n")},
		"i18n/README.md":     {Data: []byte("not a pack")},
		"i18n/draft/x.yaml":  {Data: []byte("a folder is not a pack")},
		"layouts/index.html": {Data: []byte("")},
	}
	p, err := readPackage(theme, Theme)
	if err != nil {
		t.Fatal(err)
	}
	if p.Screenshot != "shots/home.webp" || p.I18n["zh-cn"].Title != "纸" || len(p.I18n) != 1 {
		t.Errorf("package = %+v", p)
	}

	theme["i18n/chinese.yaml"] = &fstest.MapFile{Data: []byte("theme: {title: x}\n")}
	if _, err := readPackage(theme, Theme); err == nil || !strings.Contains(err.Error(), "i18n/chinese.yaml is not named after a language") {
		t.Errorf("a pack not named after a language: %v", err)
	}
}
