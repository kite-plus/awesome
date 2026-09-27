# Adding to the list

**English** · [简体中文](CONTRIBUTING.zh-CN.md)

The list takes any theme or plugin for Kite that someone else can install:
open source, on GitHub, with a release. A file of two lines names each one;
everything else is read from its latest release, and checked with the
latest Kite.

## What a release needs

- **A public repository on GitHub** that is not archived, with `theme.yaml`
  or `plugin.yaml` at its root.
- **A release with a zip the studio installs.** The latest release that is
  not a pre-release is the one read. It carries the theme or plugin as a zip,
  with the manifest at the zip's top or in its one folder, the way
  `scripts/package.sh` of [Vane](https://github.com/kite-plus/theme-vane)
  and `make zip` of the [official plugins](https://github.com/kite-plus/plugin-search)
  pack them. When it carries more than one zip, the one named
  `<name>-<version>.zip` is read.
- **Its tag is its version:** `v1.2.0` or `1.2.0` for `version: 1.2.0`.
- **It loads with the latest Kite:** `kite theme verify` or
  `kite plugin verify` passes with Kite's latest release, which means that
  release is inside `requires`.
- **A description and a license** in the manifest. The license is an open
  source one.
- **A theme shows itself.** It has a screenshot where Kite looks for one:
  `screenshot.png`, `.jpg`, `.jpeg` or `.webp`, or the file `screenshot:`
  names. The README shows it from the repository at the release's tag, so it
  is committed there as well as packed in the zip. Kite suggests 1280×800.
- **A plugin names the other sites its own scripts load from** under
  `hosts`. The list shows them beside what the plugin does, together with
  the ones written into what it injects.

Welcome but not required: a language pack for the Chinese README,
`i18n/zh-CN.yaml`, and `i18n/en.yaml` when the manifest is not written in
English. The list reads them the way the studio does. The GitHub topic
`kite-theme` or `kite-plugin` helps people find the repository.

## Adding an entry

1. Add a file named after the theme or the plugin: `themes/<name>.yaml`,
   with the theme's `name`, or `plugins/<id>.yaml`, with the plugin's `id`.
   The name is what the theme or plugin is installed as, so one taken by an
   entry already cannot be listed twice.

   ```yaml
   repo: https://github.com/you/kite-theme-paper
   category: blog
   ```

   The categories are in [categories.yaml](categories.yaml). If none fits,
   add one there in the same pull request, named in English and in
   Chinese.
2. Open a pull request. A check reads the latest release, unpacks it as the
   studio does, runs `kite theme verify` or `kite plugin verify` with the
   latest Kite, and prints what the list will say about it.

Leave the READMEs and `index.json` alone: once the pull request is merged, a
maintainer runs `make sync`, which writes them.

Rather not open a pull request?
[Suggest it in an issue](https://github.com/kite-plus/awesome/issues/new?template=suggest.yml).

## Updates, and leaving the list

Publish a release: the list reads the newest one each time it is synced.
Once a week every entry is checked against Kite's latest release. An entry
whose latest release no longer passes leaves the list at the next sync, and
comes back once a release passes again; its file stays. To leave for good,
open a pull request that deletes it.

## Maintaining

```bash
make check                          # the entries, index.json and the READMEs, offline
make sync                           # read every latest release; write index.json and the READMEs
make sync ARGS=plugins/search.yaml  # just one entry
make sync ARGS=-dry-run             # check and report, and write nothing
make readme                         # write the READMEs from index.json again
make test                           # test the tools
```

`make sync` needs Go and the latest release of Kite, on `PATH` or named with
`ARGS=-kite=/path/to/kite`: a build from source is not the Kite that sites
install. It reads GitHub with `GH_TOKEN`, taken from `gh auth token` when
the GitHub CLI is signed in, so that GitHub's limit on anonymous requests
does not get in the way.
