package main

import (
	"strings"
	"testing"
)

func TestLoadEntriesReportsEverythingAtOnce(t *testing.T) {
	root := newList(t, map[string]string{
		"plugins/search.yaml":  "repo: https://github.com/kite-plus/plugin-search\ncategory: search\n",
		"plugins/find.yaml":    "repo: https://github.com/kite-plus/plugin-search\ncategory: search\n",
		"plugins/Talk.yaml":    "repo: https://github.com/a/b\ncategory: search\n",
		"plugins/typo.yaml":    "repo: https://github.com/a/c\ncategry: search\n",
		"plugins/other.yaml":   "repo: https://github.com/a/d\ncategory: seo\n",
		"plugins/gitlab.yaml":  "repo: https://gitlab.com/a/e\ncategory: search\n",
		"plugins/short.yml":    "repo: https://github.com/a/f\ncategory: search\n",
		"themes/default.yaml":  "repo: https://github.com/a/g\ncategory: blog\n",
		"themes/Paper_2.yaml":  "repo: https://github.com/a/h\ncategory: blog\n",
		"themes/.gitkeep":      "",
		"plugins/empty.yaml":   "",
		"plugins/archive.yaml": "repo: https://github.com/a/i\ncategory: comments\n",
	})
	entries, err := loadEntries(root, mustCategories(t, root))
	if err == nil {
		t.Fatal("loadEntries passed a list with mistakes in it")
	}
	for _, want := range []string{
		"plugins/search.yaml: plugins/find.yaml lists the same repository",
		`plugins/Talk.yaml: Kite cannot install a plugin named "Talk"`,
		"plugins/typo.yaml: yaml: unmarshal errors:\n  line 2: field categry not found",
		`plugins/other.yaml: category is "seo"; want one of the plugin categories in categories.yaml: comments, search`,
		`plugins/gitlab.yaml: repo is "https://gitlab.com/a/e"`,
		"plugins/short.yml: only <name>.yaml files belong in plugins/",
		`themes/default.yaml: Kite cannot install a theme named "default"`,
		"plugins/empty.yaml is empty",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the problems lack %q:\n%v", want, err)
		}
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Path())
	}
	if got := strings.Join(names, " "); got != "themes/Paper_2.yaml plugins/archive.yaml plugins/find.yaml" {
		t.Errorf("the entries that are fine are %s", got)
	}
}

func TestLoadCategoriesWantsBothLanguages(t *testing.T) {
	root := newList(t, map[string]string{
		categoriesFile: "plugins:\n  - {id: seo, en: SEO}\n  - {id: Seo, en: SEO, zh-CN: SEO}\n",
	})
	_, err := loadCategories(root)
	if err == nil || !strings.Contains(err.Error(), "seo needs a name in en and in zh-CN") || !strings.Contains(err.Error(), `"Seo" is not lowercase`) {
		t.Fatalf("loadCategories said %v", err)
	}
}

func mustCategories(t *testing.T, root string) Categories {
	t.Helper()
	cats, err := loadCategories(root)
	if err != nil {
		t.Fatal(err)
	}
	return cats
}
