package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// The files of the list, relative to its root.
const (
	categoriesFile = "categories.yaml"
	indexFile      = "index.json"
)

// org publishes Kite; what it publishes is marked official.
const org = "kite-plus"

// Kind is what an entry is.
type Kind string

const (
	Theme  Kind = "theme"
	Plugin Kind = "plugin"
)

// kinds is the order the list keeps them in.
var kinds = []Kind{Theme, Plugin}

// Dir is the folder a kind's entries are kept in.
func (k Kind) Dir() string { return string(k) + "s" }

// Manifest is the file that makes a folder a theme or a plugin.
func (k Kind) Manifest() string { return string(k) + ".yaml" }

var (
	// The names Kite installs under: a plugin's id, a theme's folder.
	pluginID  = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	themeName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

	githubRepo = regexp.MustCompile(`^https://github\.com/([A-Za-z0-9](?:[A-Za-z0-9-]*[A-Za-z0-9])?)/([A-Za-z0-9._-]+)$`)
)

// ValidName reports whether Kite can install a theme or a plugin as name.
func (k Kind) ValidName(name string) bool {
	if k == Plugin {
		return pluginID.MatchString(name)
	}
	return themeName.MatchString(name) && name != "default"
}

// Category is a heading entries are filed under, named in the language of
// each README.
type Category struct {
	ID   string `yaml:"id"`
	En   string `yaml:"en"`
	ZhCN string `yaml:"zh-CN"`
}

// Categories are each kind's headings, in the order the READMEs list them.
type Categories struct {
	Themes  []Category `yaml:"themes"`
	Plugins []Category `yaml:"plugins"`
}

func (c Categories) Of(k Kind) []Category {
	if k == Theme {
		return c.Themes
	}
	return c.Plugins
}

// Rank is where a category comes in its kind's order, or -1.
func (c Categories) Rank(k Kind, id string) int {
	return slices.IndexFunc(c.Of(k), func(c Category) bool { return c.ID == id })
}

func (c Categories) ids(k Kind) string {
	var ids []string
	for _, cat := range c.Of(k) {
		ids = append(ids, cat.ID)
	}
	return strings.Join(ids, ", ")
}

// Entry is one file of the list: where a theme or a plugin lives, and what
// it is filed under. Its name is the file's, which is the theme's name or
// the plugin's id, so no two entries can claim the same one.
type Entry struct {
	Kind     Kind   `yaml:"-"`
	Name     string `yaml:"-"`
	Repo     string `yaml:"repo"`
	Category string `yaml:"category"`
}

// Path is the entry's file, relative to the list.
func (e Entry) Path() string { return path.Join(e.Kind.Dir(), e.Name+".yaml") }

// Owner and Repository split Repo, which loading has checked.
func (e Entry) Owner() string      { return githubRepo.FindStringSubmatch(e.Repo)[1] }
func (e Entry) Repository() string { return githubRepo.FindStringSubmatch(e.Repo)[2] }

func loadCategories(root string) (Categories, error) {
	var c Categories
	if err := decodeStrict(root, categoriesFile, &c); err != nil {
		return c, err
	}
	var errs []error
	for _, k := range kinds {
		seen := make(map[string]bool)
		for _, cat := range c.Of(k) {
			switch {
			case !pluginID.MatchString(cat.ID):
				errs = append(errs, fmt.Errorf("%s: the %s category %q is not lowercase words joined by -", categoriesFile, k, cat.ID))
			case seen[cat.ID]:
				errs = append(errs, fmt.Errorf("%s: the %s category %s is there twice", categoriesFile, k, cat.ID))
			case cat.En == "" || cat.ZhCN == "":
				errs = append(errs, fmt.Errorf("%s: the %s category %s needs a name in en and in zh-CN", categoriesFile, k, cat.ID))
			}
			seen[cat.ID] = true
		}
	}
	return c, errors.Join(errs...)
}

// loadEntries reads every entry, reporting all that is wrong at once.
func loadEntries(root string, cats Categories) ([]Entry, error) {
	var entries []Entry
	var errs []error
	repos := make(map[string]string)
	for _, k := range kinds {
		files, err := os.ReadDir(filepath.Join(root, k.Dir()))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		// Names that differ only in case would be one file on macOS.
		names := make(map[string]string)
		for _, f := range files {
			if strings.HasPrefix(f.Name(), ".") {
				continue
			}
			rel := path.Join(k.Dir(), f.Name())
			name, ok := strings.CutSuffix(f.Name(), ".yaml")
			if f.IsDir() || !ok {
				errs = append(errs, fmt.Errorf("%s: only <name>.yaml files belong in %s/", rel, k.Dir()))
				continue
			}
			if !k.ValidName(name) {
				errs = append(errs, fmt.Errorf("%s: Kite cannot install a %s named %q", rel, k, name))
				continue
			}
			e := Entry{Kind: k, Name: name}
			if err := decodeStrict(root, rel, &e); err != nil {
				errs = append(errs, err)
				continue
			}
			if !githubRepo.MatchString(e.Repo) {
				errs = append(errs, fmt.Errorf("%s: repo is %q; want a repository on GitHub, as https://github.com/kite-plus/plugin-search", rel, e.Repo))
				continue
			}
			if cats.Rank(k, e.Category) < 0 {
				errs = append(errs, fmt.Errorf("%s: category is %q; want one of the %s categories in %s: %s", rel, e.Category, k, categoriesFile, cats.ids(k)))
				continue
			}
			if other, ok := names[strings.ToLower(name)]; ok {
				errs = append(errs, fmt.Errorf("%s: %s has the same name", rel, other))
				continue
			}
			if other, ok := repos[strings.ToLower(e.Repo)]; ok {
				errs = append(errs, fmt.Errorf("%s: %s lists the same repository", rel, other))
				continue
			}
			names[strings.ToLower(name)] = rel
			repos[strings.ToLower(e.Repo)] = rel
			entries = append(entries, e)
		}
	}
	return entries, errors.Join(errs...)
}

// decodeStrict reads a YAML file of the list, refusing keys it does not
// know, so that a misspelled one is not silently ignored.
func decodeStrict(root, rel string, v any) error {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return err
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(v); err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("%s is empty", rel)
		}
		return fmt.Errorf("%s: %w", rel, err)
	}
	return nil
}
