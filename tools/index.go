package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Index is what sync read from each entry's latest release. The READMEs are
// written from it, and it is kept as index.json for anything else that
// wants the list.
type Index struct {
	// Kite is the version every release in the index was checked with.
	Kite    string   `json:"kite"`
	Themes  []Record `json:"themes"`
	Plugins []Record `json:"plugins"`
}

// Record is an entry as its latest release describes it.
type Record struct {
	Name     string `json:"name"`
	Repo     string `json:"repo"`
	Category string `json:"category"`
	Official bool   `json:"official"`

	Version  string   `json:"version"`
	Release  string   `json:"release"`
	Download Download `json:"download"`

	APIVersion string `json:"apiVersion"`
	Requires   string `json:"requires,omitempty"`

	// Title and Description are as the manifest writes them; I18n holds
	// what each language pack says instead, by lowercased language tag.
	Title       string           `json:"title"`
	Description string           `json:"description"`
	I18n        map[string]Words `json:"i18n,omitempty"`

	Author   *Author `json:"author,omitempty"`
	License  string  `json:"license"`
	Homepage string  `json:"homepage,omitempty"`

	// A plugin's other sites, which it loads nothing from when there are
	// none, and the build hooks its module exports.
	Hosts []string `json:"hosts,omitempty"`
	Hooks []string `json:"hooks,omitempty"`

	// A theme's tags, and the address of its screenshot.
	Tags       []string `json:"tags,omitempty"`
	Screenshot string   `json:"screenshot,omitempty"`
}

// Download is the release's zip, as the studio installs it.
type Download struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// Words are what a language pack says about a theme or a plugin.
type Words struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

type Author struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url,omitempty" yaml:"url"`
}

func (ix *Index) Of(k Kind) []Record {
	if k == Theme {
		return ix.Themes
	}
	return ix.Plugins
}

func (ix *Index) add(k Kind, r Record) {
	if k == Theme {
		ix.Themes = append(ix.Themes, r)
	} else {
		ix.Plugins = append(ix.Plugins, r)
	}
}

// Words picks the title and description for a language the way the studio
// does: from the pack named after it, then the one for its language alone,
// then any other for that language, falling back to the manifest's own.
func (r Record) Words(lang string) Words {
	w := Words{Title: r.Title, Description: r.Description}
	lang = strings.ToLower(lang)
	base, _, _ := strings.Cut(lang, "-")
	pack, ok := r.I18n[lang]
	if !ok {
		pack, ok = r.I18n[base]
	}
	if !ok {
		names := make([]string, 0, len(r.I18n))
		for name := range r.I18n {
			names = append(names, name)
		}
		slices.Sort(names)
		for _, name := range names {
			if strings.HasPrefix(name, base+"-") {
				pack = r.I18n[name]
				break
			}
		}
	}
	if pack.Title != "" {
		w.Title = pack.Title
	}
	if pack.Description != "" {
		w.Description = pack.Description
	}
	return w
}

func readIndex(root string) (*Index, error) {
	ix := &Index{}
	data, err := os.ReadFile(filepath.Join(root, indexFile))
	if errors.Is(err, fs.ErrNotExist) {
		return ix, nil
	}
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(ix); err != nil {
		return nil, errors.New(indexFile + ": " + err.Error())
	}
	return ix, nil
}

// encode writes the index the same way every time, so that a sync that
// finds nothing new changes nothing.
func (ix *Index) encode() ([]byte, error) {
	out := Index{Kite: ix.Kite, Themes: sorted(ix.Themes), Plugins: sorted(ix.Plugins)}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sorted(records []Record) []Record {
	out := slices.Clone(records)
	if out == nil {
		out = []Record{}
	}
	slices.SortFunc(out, func(a, b Record) int { return strings.Compare(a.Name, b.Name) })
	return out
}

// writeFile replaces a file of the list only when its bytes change.
func writeFile(root, rel string, data []byte) error {
	file := filepath.Join(root, filepath.FromSlash(rel))
	if old, err := os.ReadFile(file); err == nil && bytes.Equal(old, data) {
		return nil
	}
	tmp := file + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, file)
}
