package main

import (
	"errors"
	"fmt"
	"io"
)

// runCheck checks what can be checked without the network: the entries and
// the categories, index.json, and that the READMEs are what index.json
// makes. An entry that index.json does not have yet, or no longer should,
// is reported but passes: the next sync settles it.
func runCheck(root string, out io.Writer) error {
	cats, err := loadCategories(root)
	if err != nil {
		return err
	}
	entries, err := loadEntries(root, cats)
	if err != nil {
		return err
	}
	ix, err := readIndex(root)
	if err != nil {
		return err
	}
	if err := checkIndex(ix, cats); err != nil {
		return err
	}

	var errs []error
	for _, r := range readmes {
		current, want, err := r.rendered(root, ix, cats)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if current != want {
			errs = append(errs, fmt.Errorf("%s: the lists in it are not what %s makes; run make readme, and edit the entries rather than the lists", r.File, indexFile))
		}
	}
	if err := errors.Join(errs...); err != nil {
		return err
	}

	for _, note := range pending(ix, entries) {
		fmt.Fprintln(out, note)
	}
	return nil
}

// checkIndex makes sure the READMEs can be written from the index.
func checkIndex(ix *Index, cats Categories) error {
	var errs []error
	for _, k := range kinds {
		seen := make(map[string]bool)
		for _, r := range ix.Of(k) {
			where := fmt.Sprintf("%s: %s %q", indexFile, k, r.Name)
			switch {
			case !k.ValidName(r.Name):
				errs = append(errs, fmt.Errorf("%s: the name is not one Kite installs", where))
			case seen[r.Name]:
				errs = append(errs, fmt.Errorf("%s: it is there twice", where))
			case cats.Rank(k, r.Category) < 0:
				errs = append(errs, fmt.Errorf("%s: its category %q is not in %s; run make sync", where, r.Category, categoriesFile))
			case r.Version == "" || r.Release == "" || r.Repo == "":
				errs = append(errs, fmt.Errorf("%s: it has no version, release or repository; run make sync", where))
			case k == Theme && r.Screenshot == "":
				errs = append(errs, fmt.Errorf("%s: it has no screenshot; run make sync", where))
			}
			seen[r.Name] = true
		}
	}
	if len(errs) == 0 && ix.Kite == "" && (len(ix.Themes) > 0 || len(ix.Plugins) > 0) {
		errs = append(errs, fmt.Errorf("%s: it does not say which Kite checked it; run make sync", indexFile))
	}
	return errors.Join(errs...)
}

// pending says where the entries and the index differ, which the next sync
// settles.
func pending(ix *Index, entries []Entry) []string {
	var notes []string
	files := make(map[string]Entry)
	for _, e := range entries {
		files[e.Path()] = e
	}
	records := make(map[string]Record)
	for _, k := range kinds {
		for _, r := range ix.Of(k) {
			records[Entry{Kind: k, Name: r.Name}.Path()] = r
		}
	}
	for _, e := range entries {
		r, ok := records[e.Path()]
		switch {
		case !ok:
			notes = append(notes, e.Path()+": not listed yet; make sync reads its release and lists it if it passes")
		case r.Repo != e.Repo || r.Category != e.Category:
			notes = append(notes, e.Path()+": changed since it was synced; make sync lists it as it is now")
		}
	}
	for _, k := range kinds {
		for _, r := range ix.Of(k) {
			if path := (Entry{Kind: k, Name: r.Name}).Path(); files[path] == (Entry{}) {
				notes = append(notes, path+": gone, but still listed; make sync takes it off the list")
			}
		}
	}
	return notes
}
