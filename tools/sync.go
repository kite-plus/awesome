package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"testing/fstest"
	"time"
)

// syncOptions are what a sync reads with, and what it checks.
type syncOptions struct {
	GitHub *GitHub
	Kite   string   // the kite to check releases with
	DryRun bool     // report, and write nothing
	Only   []string // entry files; every entry when empty
	Out    io.Writer
}

func runSync(root string, args []string) error {
	flags := flag.NewFlagSet("sync", flag.ContinueOnError)
	kite := flags.String("kite", "kite", "the `kite` to check releases with, which has to be Kite's latest release")
	dryRun := flags.Bool("dry-run", false, "check the entries and report, writing nothing")
	if err := flags.Parse(args); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return syncList(ctx, root, syncOptions{
		GitHub: newGitHub(),
		Kite:   *kite,
		DryRun: *dryRun,
		Only:   flags.Args(),
		Out:    os.Stdout,
	})
}

// syncList checks the latest release of each entry chosen, and writes the
// index and the READMEs from what passed. An entry whose release does not
// pass is left out of the list until one does; its file stays. Anything
// else that goes wrong, such as the network, stops the sync before it
// writes a thing.
func syncList(ctx context.Context, root string, o syncOptions) error {
	cats, err := loadCategories(root)
	if err != nil {
		return err
	}
	entries, err := loadEntries(root, cats)
	if err != nil {
		return err
	}
	chosen, err := choose(entries, o.Only)
	if err != nil {
		return err
	}
	old, err := readIndex(root)
	if err != nil {
		return err
	}

	kite, err := kiteVersion(ctx, o.Kite)
	if err != nil {
		return err
	}
	latest, err := o.GitHub.LatestRelease(ctx, org, "kite")
	if err != nil {
		return fmt.Errorf("reading Kite's latest release: %w", err)
	}
	if want := strings.TrimPrefix(latest.TagName, "v"); kite != want {
		return fmt.Errorf("%s is Kite %s, but the latest release is %s: check with that one, or name it with -kite", o.Kite, kite, want)
	}
	// Every release in the index is checked with one Kite.
	if len(o.Only) > 0 && !o.DryRun && old.Kite != "" && old.Kite != kite {
		return fmt.Errorf("%s was checked with Kite %s; to check with Kite %s, sync every entry", indexFile, old.Kite, kite)
	}

	synced := make(map[string]Record)
	var refused []string
	for _, e := range chosen {
		rec, said, err := syncEntry(ctx, o, e)
		var r *refusal
		switch {
		case errors.As(err, &r):
			refused = append(refused, e.Path()+": "+r.reason)
			fmt.Fprintf(o.Out, "%s\n  left out: %s\n\n", e.Path(), r.reason)
		case err != nil:
			return fmt.Errorf("%s: %w", e.Path(), err)
		default:
			synced[e.Path()] = *rec
			report(o.Out, e, rec, said)
		}
	}

	if !o.DryRun {
		ix := merge(old, entries, chosen, synced)
		ix.Kite = kite
		if err := write(root, ix, cats); err != nil {
			return err
		}
	}
	if len(refused) > 0 {
		return fmt.Errorf("%d of %d entries did not pass with Kite %s:\n  %s",
			len(refused), len(chosen), kite, strings.Join(refused, "\n  "))
	}
	return nil
}

// choose finds the entries named by their files, or all of them.
func choose(entries []Entry, only []string) ([]Entry, error) {
	if len(only) == 0 {
		return entries, nil
	}
	var out []Entry
	for _, arg := range only {
		file := strings.TrimSuffix(filepath.ToSlash(arg), ".yaml") + ".yaml"
		i := slices.IndexFunc(entries, func(e Entry) bool { return e.Path() == file })
		if i < 0 {
			return nil, fmt.Errorf("%s is not an entry; name one by its file, as plugins/search.yaml", arg)
		}
		if !slices.Contains(out, entries[i]) {
			out = append(out, entries[i])
		}
	}
	return out, nil
}

// merge makes the next index: what was synced, and for the entries not
// chosen what the index already said, as long as it still describes the
// same repository.
func merge(old *Index, entries, chosen []Entry, synced map[string]Record) *Index {
	prev := make(map[string]Record)
	for _, k := range kinds {
		for _, r := range old.Of(k) {
			prev[Entry{Kind: k, Name: r.Name}.Path()] = r
		}
	}
	ix := &Index{}
	for _, e := range entries {
		rec, ok := synced[e.Path()]
		if !slices.Contains(chosen, e) {
			rec, ok = prev[e.Path()]
			ok = ok && rec.Repo == e.Repo
			rec.Category = e.Category
		}
		if ok {
			ix.add(e.Kind, rec)
		}
	}
	return ix
}

// syncEntry reads an entry's latest release and checks it the way a site
// would install it. It returns what the list will say, and what Kite said.
func syncEntry(ctx context.Context, o syncOptions, e Entry) (*Record, string, error) {
	owner, name := e.Owner(), e.Repository()
	repo, err := o.GitHub.Repository(ctx, owner, name)
	if errors.Is(err, errNotFound) {
		return nil, "", refuse("%s does not exist, or is not public", e.Repo)
	}
	if err != nil {
		return nil, "", err
	}
	if repo.Archived {
		return nil, "", refuse("%s is archived", e.Repo)
	}

	rel, err := o.GitHub.LatestRelease(ctx, owner, name)
	if errors.Is(err, errNotFound) {
		return nil, "", refuse("%s has no release; its latest may be a draft or a pre-release", e.Repo)
	}
	if err != nil {
		return nil, "", err
	}
	asset, err := archiveOf(rel, e.Name)
	if err != nil {
		return nil, "", err
	}
	data, err := o.GitHub.Download(ctx, asset, maxArchive)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(data)
	digest := hex.EncodeToString(sum[:])
	if asset.Digest != "" && asset.Digest != "sha256:"+digest {
		return nil, "", fmt.Errorf("%s came down as sha256:%s, but GitHub has it as %s", asset.Name, digest, asset.Digest)
	}

	files, err := unpack(data, e.Kind)
	if err != nil {
		return nil, "", err
	}
	pkg, err := readPackage(files, e.Kind)
	if err != nil {
		return nil, "", err
	}
	manifest := e.Kind.Manifest()
	switch {
	case pkg.Name != e.Name:
		return nil, "", refuse("%s in the release names the %s %q, but the entry's file names it %q", manifest, e.Kind, pkg.Name, e.Name)
	case strings.TrimPrefix(rel.TagName, "v") != pkg.Version:
		return nil, "", refuse("the release is tagged %s, but %s says version %s", rel.TagName, manifest, pkg.Version)
	case strings.TrimSpace(pkg.Description) == "":
		return nil, "", refuse("%s has no description", manifest)
	case strings.TrimSpace(pkg.License) == "":
		return nil, "", refuse("%s has no license", manifest)
	case e.Kind == Theme && pkg.Screenshot == "":
		return nil, "", refuse("the theme has no screenshot: screenshot.png, .jpg, .jpeg or .webp, or the file screenshot: names in %s", manifest)
	}

	rec := &Record{
		Name:     e.Name,
		Repo:     e.Repo,
		Category: e.Category,
		Official: owner == org,

		Version:  pkg.Version,
		Release:  rel.HTMLURL,
		Download: Download{URL: asset.URL, SHA256: digest, Size: int64(len(data))},

		APIVersion: pkg.APIVersion,
		Requires:   pkg.Requires,

		Title:       pkg.Title,
		Description: pkg.Description,
		I18n:        pkg.I18n,

		License:  pkg.License,
		Homepage: pkg.Homepage,
		Hosts:    pkg.Hosts,
		Hooks:    pkg.Hooks,
		Tags:     pkg.Tags,
	}
	if rec.Title == "" {
		rec.Title = pkg.Name
	}
	if pkg.Author.Name != "" {
		rec.Author = &pkg.Author
	}
	if e.Kind == Theme {
		// The README shows the screenshot from the repository, where it has
		// to be at the same path as in the zip.
		rec.Screenshot = o.GitHub.RawURL(owner, name, rel.TagName, pkg.Screenshot)
		found, err := o.GitHub.Exists(ctx, rec.Screenshot)
		if err != nil {
			return nil, "", err
		}
		if !found {
			return nil, "", refuse("%s is in the release's zip, but not in the repository at %s, where the README would show it from", pkg.Screenshot, rel.TagName)
		}
	}

	said, err := kiteVerify(ctx, o.Kite, e, files)
	if err != nil {
		return nil, "", err
	}
	return rec, said, nil
}

// archiveOf finds the zip a release carries for the studio: its only one,
// or the one named <name>-<version>.zip.
func archiveOf(rel *Release, name string) (Asset, error) {
	var zips []Asset
	for _, a := range rel.Assets {
		if strings.HasSuffix(strings.ToLower(a.Name), ".zip") {
			zips = append(zips, a)
		}
	}
	switch len(zips) {
	case 0:
		return Asset{}, refuse("the release %s carries no zip; attach the one the studio installs", rel.TagName)
	case 1:
		return zips[0], nil
	}
	want := name + "-" + strings.TrimPrefix(rel.TagName, "v") + ".zip"
	for _, a := range zips {
		if a.Name == want {
			return a, nil
		}
	}
	return Asset{}, refuse("the release %s carries several zips and none named %s", rel.TagName, want)
}

// kiteVersion asks a kite which version it is.
func kiteVersion(ctx context.Context, kite string) (string, error) {
	out, err := exec.CommandContext(ctx, kite, "version").Output()
	if err != nil {
		return "", fmt.Errorf("running %s version: %w; install Kite's latest release, or name it with -kite", kite, err)
	}
	first, _, _ := strings.Cut(string(out), "\n")
	fields := strings.Fields(first)
	if len(fields) < 2 || fields[0] != "kite" {
		return "", fmt.Errorf("%s version said %q, which is not what Kite says", kite, first)
	}
	return strings.TrimPrefix(fields[1], "v"), nil
}

// kiteVerify has Kite check a theme or a plugin the way a site loads it,
// from a folder named after it, as installed ones are.
func kiteVerify(ctx context.Context, kite string, e Entry, files fstest.MapFS) (string, error) {
	tmp, err := os.MkdirTemp("", "awesome-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)
	dir := filepath.Join(tmp, e.Name)
	for name, f := range files {
		file := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(file, f.Data, 0o644); err != nil {
			return "", err
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	out, err := exec.CommandContext(ctx, kite, string(e.Kind), "verify", dir).CombinedOutput()
	said := strings.TrimSpace(strings.ReplaceAll(string(out), dir, e.Name))
	var exit *exec.ExitError
	switch {
	case errors.As(err, &exit):
		return "", refuse("kite %s verify: %s", e.Kind, said)
	case err != nil:
		return "", fmt.Errorf("running kite %s verify: %w", e.Kind, err)
	}
	return said, nil
}

// report says what the list will show of an entry, for whoever reviews it.
func report(w io.Writer, e Entry, r *Record, said string) {
	fmt.Fprintf(w, "%s\n  %s %s, %s\n", e.Path(), r.Name, r.Version, r.Release)
	for _, lang := range []string{"en", "zh-CN"} {
		words := r.Words(lang)
		fmt.Fprintf(w, "  %s: %s: %s\n", lang, words.Title, words.Description)
	}
	if e.Kind == Plugin {
		if len(r.Hosts) == 0 {
			fmt.Fprintln(w, "  loads nothing from other sites")
		} else {
			fmt.Fprintf(w, "  loads from %s\n", strings.Join(r.Hosts, ", "))
		}
	}
	for line := range strings.SplitSeq(said, "\n") {
		fmt.Fprintf(w, "  kite: %s\n", line)
	}
	fmt.Fprintln(w)
}

// write writes the index and the READMEs made from it.
func write(root string, ix *Index, cats Categories) error {
	data, err := ix.encode()
	if err != nil {
		return err
	}
	if err := writeFile(root, indexFile, data); err != nil {
		return err
	}
	return writeReadmes(root, ix, cats)
}
