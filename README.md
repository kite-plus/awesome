<!-- What sits between BEGIN and END is written by `make sync` from
index.json. To change a list, edit themes/ or plugins/ instead. -->

<p align="center">
  <img src="https://raw.githubusercontent.com/kite-plus/.github/main/assets/readme/logo.svg" width="64" alt="Kite logo">
</p>

<h1 align="center">Awesome Kite</h1>

<p align="center">Themes and plugins for Kite, from us and from everyone else.</p>

<p align="center">
  <a href="https://awesome.re"><img src="https://awesome.re/badge-flat.svg" alt="Awesome"></a>
  <a href="https://github.com/kite-plus/awesome/actions/workflows/check.yml"><img src="https://github.com/kite-plus/awesome/actions/workflows/check.yml/badge.svg" alt="Check"></a>
<!-- BEGIN awesome:badges -->
  <a href="https://github.com/kite-plus/kite/releases/tag/v0.1.1"><img src="https://img.shields.io/badge/checked%20with-Kite%200.1.1-4A77D6?logo=data:image/svg%2bxml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCA2NCA2NCI+PGcgZmlsbD0iI2ZmZiIgc3Ryb2tlPSIjZmZmIiBzdHJva2Utd2lkdGg9IjUiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxwYXRoIGQ9Ik0xMCAxNC41IEwyNyAyMSBMMjcgMzAgTDEwIDIzLjUgWiIvPjxwYXRoIGQ9Ik0xMCAzMiBMMjcgMzguNSBMMjcgNDkgTDEwIDQyLjUgWiIvPjxwYXRoIGQ9Ik0zNyAyMSBMNTQgMTQuNSBMNTQgNDIuNSBMMzcgNDkgWiIvPjwvZz48L3N2Zz4=" alt="Checked with Kite 0.1.1"></a>
  <a href="#themes"><img src="https://img.shields.io/badge/themes-1-4A77D6" alt="Themes listed: 1"></a>
  <a href="#plugins"><img src="https://img.shields.io/badge/plugins-4-4A77D6" alt="Plugins listed: 4"></a>
<!-- END awesome:badges -->
  <a href="CONTRIBUTING.md"><img src="https://img.shields.io/badge/PRs-welcome-brightgreen" alt="Pull requests welcome"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-CC0%201.0-blue" alt="CC0 1.0"></a>
</p>

<p align="center">
  <strong>English</strong> · <a href="README.zh-CN.md">简体中文</a>
</p>

A theme decides how a [Kite](https://github.com/kite-plus/kite) site looks,
and a plugin adds what the theme does not: comments, analytics, search, math.
Each one here lives in its own repository and has a release whose zip
installs straight from Kite's studio.

<!-- BEGIN awesome:status -->
The latest release of every entry has been checked with Kite 0.1.1, the latest version: it installs, and it loads.
<!-- END awesome:status -->

- [Themes](#themes)
- [Plugins](#plugins)
- [Installing](#installing)
- [Adding yours](#adding-yours)

## Themes

Kite comes with a theme built in, so it needs no installing: serif type for
personal writing, in light and dark. These are the others.

<!-- BEGIN awesome:themes -->
<table>
<tr>
<td width="320" valign="top"><a href="https://github.com/kite-plus/theme-vane"><img src="https://raw.githubusercontent.com/kite-plus/theme-vane/v0.2.0/screenshot.webp" width="320" alt="Screenshot of Vane"></a></td>
<td valign="top"><strong><a href="https://github.com/kite-plus/theme-vane">Vane</a></strong><br><sub>Documentation · Official · <a href="https://github.com/kite-plus/theme-vane/releases/tag/v0.2.0">0.2.0</a></sub><br><br>A documentation theme for product docs, project sites and knowledge bases, drawn on paper in ink. A home page with a kite in its sky, docs in an ordered tree with a table of contents, and news, by day and by night, with nothing loaded from third parties.</td>
</tr>
</table>
<!-- END awesome:themes -->

## Plugins

<!-- BEGIN awesome:plugins -->
| Plugin | What it does | Other sites it loads from |
| --- | --- | --- |
| **[Comments](https://github.com/kite-plus/plugin-comments)**<br><sub>Comments · Official · [0.1.0](https://github.com/kite-plus/plugin-comments/releases/tag/v0.1.0)</sub> | A comment thread under every post, with Giscus, Waline or Twikoo. | `cdn.jsdelivr.net`<br>`giscus.app`<br>`registry.npmmirror.com`<br>`unpkg.com` |
| **[Analytics](https://github.com/kite-plus/plugin-analytics)**<br><sub>Analytics · Official · [0.1.0](https://github.com/kite-plus/plugin-analytics/releases/tag/v0.1.0)</sub> | Counts visits with Baidu Tongji, Google Analytics, Umami or Plausible. | `cloud.umami.is`<br>`hm.baidu.com`<br>`plausible.io`<br>`www.googletagmanager.com` |
| **[Search](https://github.com/kite-plus/plugin-search)**<br><sub>Search · Official · [0.1.0](https://github.com/kite-plus/plugin-search/releases/tag/v0.1.0)</sub> | Lets readers search every post and page, in their browser and without a server. | None |
| **[Math and Diagrams](https://github.com/kite-plus/plugin-math)**<br><sub>Writing · Official · [0.1.0](https://github.com/kite-plus/plugin-math/releases/tag/v0.1.0)</sub> | Typesets TeX math with KaTeX and draws mermaid code blocks as diagrams. | `cdn.jsdelivr.net`<br>`registry.npmmirror.com`<br>`unpkg.com` |
<!-- END awesome:plugins -->

The last column names every other site a plugin's code may load from. Which
of them it does depends on its settings: the analytics plugin reaches only
the service you choose. The studio names the same sites before a plugin is
turned on.

## Installing

**A theme:** download the zip from its release and drop it on the upload
tile under **Settings → Theme** in the studio. You can try it on the whole
site before switching to it. Or unzip it into your site's `themes/` folder
and set `theme.name` to its name in `kite.yaml`.

**A plugin:** download the zip from its release, drop it on **Plugins** in
the studio, and turn it on. Or, in your site's folder:

```bash
kite plugin add search-0.1.0.zip
kite plugin enable search
```

Kite's reference says more about [themes](https://github.com/kite-plus/kite/blob/main/docs/reference.md#themes)
and [plugins](https://github.com/kite-plus/kite/blob/main/docs/reference.md#plugins).

## Adding yours

Made a theme or a plugin for Kite? Add a file named after it to `themes/` or
`plugins/`, give it the repository and one of the [categories](categories.yaml),
and open a pull request:

```yaml
# plugins/search.yaml
repo: https://github.com/kite-plus/plugin-search
category: search
```

The title, the description and the version come from its latest release, so
you do not write them twice. [CONTRIBUTING.md](CONTRIBUTING.md) says what a
release needs and what is checked. Starting one? Kite's reference covers
[writing a theme](https://github.com/kite-plus/kite/blob/main/docs/reference.md#themes)
and [writing a plugin](https://github.com/kite-plus/kite/blob/main/docs/reference.md#writing-a-plugin).

## License

[CC0 1.0](LICENSE): the list and `index.json` are free to use for anything,
without asking and without credit. Each theme and plugin has its own
license, in its repository.
