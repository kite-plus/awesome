package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"
	"testing/fstest"

	"gopkg.in/yaml.v3"
)

// The limits of Kite's studio: the largest upload it takes, and what a
// theme or a plugin may unpack to.
const (
	maxArchive  = 64 << 20
	maxUnpacked = 128 << 20
	maxFiles    = 5000
)

// unpack reads a theme or a plugin out of its release archive as the studio
// installs one (internal/archive in Kite): from the archive's top, or from
// the one folder everything in it sits in.
func unpack(archive []byte, kind Kind) (fstest.MapFS, error) {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, refuse("the release's zip is not a zip archive")
	}
	manifest := kind.Manifest()

	var names []string
	for _, f := range zr.File {
		if name := archived(f.Name); name != "" && !strings.HasSuffix(name, "/") {
			names = append(names, name)
		}
	}
	root := ""
	if !slices.Contains(names, manifest) {
		missing := refuse("the release's zip has no %s at its top or in a single folder", manifest)
		if len(names) == 0 {
			return nil, refuse("the release's zip is empty")
		}
		top, _, _ := strings.Cut(names[0], "/")
		for _, name := range names {
			if !strings.HasPrefix(name, top+"/") {
				return nil, missing
			}
		}
		if !slices.Contains(names, top+"/"+manifest) {
			return nil, missing
		}
		root = top + "/"
	}

	files := make(fstest.MapFS)
	var total int64
	for _, f := range zr.File {
		name := archived(f.Name)
		if name == "" || strings.HasSuffix(name, "/") {
			continue
		}
		rel := strings.TrimPrefix(name, root)
		if !fs.ValidPath(rel) {
			return nil, refuse("the release's zip holds a path that leads out of the %s: %s", kind, f.Name)
		}
		if !f.Mode().IsRegular() {
			return nil, refuse("the release's zip holds something other than a plain file: %s", f.Name)
		}
		if len(files) == maxFiles {
			return nil, refuse("a %s may hold at most %d files", kind, maxFiles)
		}
		rc, err := f.Open()
		if err != nil {
			return nil, refuse("the release's zip cannot be read: %v", err)
		}
		// The sizes an archive declares are not trusted: the bytes are
		// counted as they come out.
		data, err := io.ReadAll(io.LimitReader(rc, maxUnpacked-total+1))
		_ = rc.Close()
		if err != nil {
			return nil, refuse("the release's zip cannot be read: %v", err)
		}
		total += int64(len(data))
		if total > maxUnpacked {
			return nil, refuse("a %s may take at most %d MB unpacked", kind, maxUnpacked>>20)
		}
		files[rel] = &fstest.MapFile{Data: data, Mode: 0o644}
	}
	return files, nil
}

// archived is the name of a file in an archive with forward slashes, or ""
// for what a computer adds to an archive on its own.
func archived(name string) string {
	name = strings.ReplaceAll(name, `\`, "/")
	for part := range strings.SplitSeq(name, "/") {
		switch part {
		case "__MACOSX", ".git", ".DS_Store", "Thumbs.db", "desktop.ini":
			return ""
		}
	}
	return name
}

// Package is what a theme or a plugin says about itself in its manifest and
// its language packs.
type Package struct {
	Name        string // a theme's name, a plugin's id
	Title       string
	Version     string
	APIVersion  string
	Requires    string
	Description string
	Author      Author
	License     string
	Homepage    string
	I18n        map[string]Words

	Hosts []string
	Hooks []string

	Tags       []string
	Screenshot string // its path in the theme, when it has one
}

type themeManifest struct {
	Name        string   `yaml:"name"`
	Title       string   `yaml:"title"`
	Version     string   `yaml:"version"`
	APIVersion  string   `yaml:"apiVersion"`
	Requires    string   `yaml:"requires"`
	Description string   `yaml:"description"`
	Author      Author   `yaml:"author"`
	License     string   `yaml:"license"`
	Homepage    string   `yaml:"homepage"`
	Tags        []string `yaml:"tags"`
	Screenshot  string   `yaml:"screenshot"`
}

type pluginManifest struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	APIVersion  string   `yaml:"apiVersion"`
	Requires    string   `yaml:"requires"`
	Description string   `yaml:"description"`
	Author      Author   `yaml:"author"`
	License     string   `yaml:"license"`
	Homepage    string   `yaml:"homepage"`
	Hosts       []string `yaml:"hosts"`
	Hooks       []string `yaml:"hooks"`
	Inject      []struct {
		HTML string `yaml:"html"`
	} `yaml:"inject"`
}

// readPackage reads what the list shows of a theme or a plugin. Whether it
// loads is for Kite to say.
func readPackage(fsys fs.FS, kind Kind) (*Package, error) {
	data, err := fs.ReadFile(fsys, kind.Manifest())
	if err != nil {
		return nil, refuse("%s cannot be read: %v", kind.Manifest(), err)
	}
	var p Package
	switch kind {
	case Theme:
		var m themeManifest
		if err := yaml.Unmarshal(data, &m); err != nil {
			return nil, refuse("%s: %v", kind.Manifest(), err)
		}
		p = Package{
			Name: m.Name, Title: m.Title, Version: m.Version,
			APIVersion: m.APIVersion, Requires: m.Requires,
			Description: m.Description, Author: m.Author,
			License: m.License, Homepage: m.Homepage, Tags: m.Tags,
		}
		p.Screenshot = screenshot(fsys, m.Screenshot)
	case Plugin:
		var m pluginManifest
		if err := yaml.Unmarshal(data, &m); err != nil {
			return nil, refuse("%s: %v", kind.Manifest(), err)
		}
		p = Package{
			Name: m.ID, Title: m.Name, Version: m.Version,
			APIVersion: m.APIVersion, Requires: m.Requires,
			Description: m.Description, Author: m.Author,
			License: m.License, Homepage: m.Homepage, Hooks: m.Hooks,
		}
		var html []string
		for _, in := range m.Inject {
			html = append(html, in.HTML)
		}
		p.Hosts = hosts(m.Hosts, html)
	}
	if p.I18n, err = readPacks(fsys, kind); err != nil {
		return nil, err
	}
	return &p, nil
}

// hostsIn finds the sites injected code loads from, as Kite does.
var hostsIn = regexp.MustCompile(`(?i)(?:src|href)\s*=\s*["']?(?:https?:)?//([a-z0-9.-]+\.[a-z]{2,})`)

// hosts lists the other sites a plugin's code loads from, as the studio
// does before the plugin is turned on: the ones it declares and the ones
// written into what it injects.
func hosts(declared, injected []string) []string {
	out := slices.Clone(declared)
	for _, html := range injected {
		for _, m := range hostsIn.FindAllStringSubmatch(html, -1) {
			out = append(out, strings.ToLower(m[1]))
		}
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// screenshot finds a theme's screenshot where Kite looks for it.
func screenshot(fsys fs.FS, named string) string {
	candidates := []string{"screenshot.png", "screenshot.jpg", "screenshot.jpeg", "screenshot.webp"}
	if named != "" {
		candidates = []string{path.Clean(named)}
	}
	for _, name := range candidates {
		if !fs.ValidPath(name) {
			continue
		}
		if info, err := fs.Stat(fsys, name); err == nil && !info.IsDir() {
			return name
		}
	}
	return ""
}

var packName = regexp.MustCompile(`^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)

// readPacks reads the title and description each language pack gives, by
// lowercased language tag, as Kite keys them.
func readPacks(fsys fs.FS, kind Kind) (map[string]Words, error) {
	entries, err := fs.ReadDir(fsys, "i18n")
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, refuse("i18n cannot be read: %v", err)
	}
	packs := make(map[string]Words)
	for _, entry := range entries {
		ext := path.Ext(entry.Name())
		if entry.IsDir() || (ext != ".yaml" && ext != ".yml") {
			continue
		}
		lang := strings.TrimSuffix(entry.Name(), ext)
		where := path.Join("i18n", entry.Name())
		if !packName.MatchString(lang) {
			return nil, refuse("%s is not named after a language, as zh-CN.yaml is", where)
		}
		data, err := fs.ReadFile(fsys, where)
		if err != nil {
			return nil, refuse("%s cannot be read: %v", where, err)
		}
		var tree map[string]any
		if err := yaml.Unmarshal(data, &tree); err != nil {
			return nil, refuse("%s: %v", where, err)
		}
		var w Words
		if words, ok := tree[string(kind)].(map[string]any); ok {
			w.Title, _ = words["title"].(string)
			w.Description, _ = words["description"].(string)
		}
		packs[strings.ToLower(lang)] = w
	}
	if len(packs) == 0 {
		return nil, nil
	}
	return packs, nil
}

// refusal is why an entry cannot be listed: something about the entry or
// its release, rather than the network or this tool.
type refusal struct{ reason string }

func (r *refusal) Error() string { return r.reason }

func refuse(format string, args ...any) error {
	return &refusal{fmt.Sprintf(format, args...)}
}
