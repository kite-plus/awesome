<!-- BEGIN 和 END 之间的内容由 `make sync` 根据 index.json 生成。
要修改列表，请改 themes/ 或 plugins/ 下的文件。 -->

<p align="center">
  <img src="https://raw.githubusercontent.com/kite-plus/.github/main/assets/readme/logo.svg" width="64" alt="Kite 标志">
</p>

<h1 align="center">Awesome Kite</h1>

<p align="center">Kite 的主题和插件，官方的和大家写的都在这里。</p>

<p align="center">
  <a href="https://awesome.re"><img src="https://awesome.re/badge-flat.svg" alt="Awesome"></a>
  <a href="https://github.com/kite-plus/awesome/actions/workflows/check.yml"><img src="https://github.com/kite-plus/awesome/actions/workflows/check.yml/badge.svg" alt="检查"></a>
<!-- BEGIN awesome:badges -->
  <a href="https://github.com/kite-plus/kite/releases/tag/v0.1.9"><img src="https://img.shields.io/badge/checked%20with-Kite%200.1.9-4A77D6?logo=data:image/svg%2bxml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCA2NCA2NCI+PGcgZmlsbD0iI2ZmZiIgc3Ryb2tlPSIjZmZmIiBzdHJva2Utd2lkdGg9IjUiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxwYXRoIGQ9Ik0xMCAxNC41IEwyNyAyMSBMMjcgMzAgTDEwIDIzLjUgWiIvPjxwYXRoIGQ9Ik0xMCAzMiBMMjcgMzguNSBMMjcgNDkgTDEwIDQyLjUgWiIvPjxwYXRoIGQ9Ik0zNyAyMSBMNTQgMTQuNSBMNTQgNDIuNSBMMzcgNDkgWiIvPjwvZz48L3N2Zz4=" alt="用 Kite 0.1.9 检查过"></a>
  <a href="#主题"><img src="https://img.shields.io/badge/themes-1-4A77D6" alt="收录的主题：1"></a>
  <a href="#插件"><img src="https://img.shields.io/badge/plugins-4-4A77D6" alt="收录的插件：4"></a>
<!-- END awesome:badges -->
  <a href="CONTRIBUTING.zh-CN.md"><img src="https://img.shields.io/badge/PRs-welcome-brightgreen" alt="欢迎提交 pull request"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-CC0%201.0-blue" alt="CC0 1.0"></a>
</p>

<p align="center">
  <a href="README.md">English</a> · <strong>简体中文</strong>
</p>

主题决定 [Kite](https://github.com/kite-plus/kite) 网站的样子，插件加上主题之外的功能：评论、统计、站内搜索、公式。这里收录的每一项都放在自己的仓库里，都有发布版本，附带的 zip 可以直接在后台安装。

<!-- BEGIN awesome:status -->
每一项的最新版本都用 Kite 0.1.9（目前最新的版本）检查过：能安装，能加载。
<!-- END awesome:status -->

- [主题](#主题)
- [插件](#插件)
- [安装](#安装)
- [收录你的作品](#收录你的作品)

## 主题

Kite 自带一套主题，编译在程序里，不需要安装：为个人写作准备的衬线排版，深浅两色。下面是其他主题。

<!-- BEGIN awesome:themes -->
<table>
<tr>
<td width="320" valign="top"><a href="https://github.com/kite-plus/theme-vane"><img src="https://raw.githubusercontent.com/kite-plus/theme-vane/v1.0.2/screenshot.webp" width="320" alt="风标的截图"></a></td>
<td valign="top"><strong><a href="https://github.com/kite-plus/theme-vane">风标</a></strong><br><sub>文档 · 官方 · <a href="https://github.com/kite-plus/theme-vane/releases/tag/v1.0.2">1.0.2</a></sub><br><br>为产品文档、项目官网和知识库设计的文档主题，纸上落墨。天空里飘着风筝的首页、按顺序排列的文档目录、页内目录和新闻，白天与夜晚两种样子，不从第三方加载任何资源。</td>
</tr>
</table>
<!-- END awesome:themes -->

## 插件

<!-- BEGIN awesome:plugins -->
| 插件 | 作用 | 会连接的其他网站 |
| --- | --- | --- |
| **[评论](https://github.com/kite-plus/plugin-comments)**<br><sub>评论 · 官方 · [0.1.0](https://github.com/kite-plus/plugin-comments/releases/tag/v0.1.0)</sub> | 在每篇文章下放一个评论区，支持 Giscus、Waline 和 Twikoo。 | `cdn.jsdelivr.net`<br>`giscus.app`<br>`registry.npmmirror.com`<br>`unpkg.com` |
| **[访问统计](https://github.com/kite-plus/plugin-analytics)**<br><sub>统计 · 官方 · [0.1.0](https://github.com/kite-plus/plugin-analytics/releases/tag/v0.1.0)</sub> | 用百度统计、Google Analytics、Umami 或 Plausible 统计访问量。 | `cloud.umami.is`<br>`hm.baidu.com`<br>`plausible.io`<br>`www.googletagmanager.com` |
| **[站内搜索](https://github.com/kite-plus/plugin-search)**<br><sub>搜索 · 官方 · [0.1.0](https://github.com/kite-plus/plugin-search/releases/tag/v0.1.0)</sub> | 让读者在浏览器里搜索全部文章和页面，不需要服务端。 | 无 |
| **[公式与图表](https://github.com/kite-plus/plugin-math)**<br><sub>写作 · 官方 · [0.1.0](https://github.com/kite-plus/plugin-math/releases/tag/v0.1.0)</sub> | 用 KaTeX 排版 TeX 公式，把 mermaid 代码块画成图表。 | `cdn.jsdelivr.net`<br>`registry.npmmirror.com`<br>`unpkg.com` |
<!-- END awesome:plugins -->

最后一列是插件的代码可能从哪些其他网站加载内容。实际连接哪些取决于设置，比如统计插件只连接你选的那一家服务。后台在打开插件之前也会列出同样的网站。

## 安装

**主题**：从主题的发布页下载 zip，在后台「设置 → 主题」上传。启用之前可以先在整站上预览。也可以解压到站点的 `themes/` 目录，在 `kite.yaml` 里把 `theme.name` 设成主题的名字。

**插件**：从插件的发布页下载 zip，在后台「插件」页上传，然后打开。也可以在站点目录里用命令行：

```bash
kite plugin add search-0.1.0.zip
kite plugin enable search
```

详细说明见 Kite 官网的[使用主题](https://www.kite.plus/using-themes/)和[使用插件](https://www.kite.plus/using-plugins/)。

## 收录你的作品

给 Kite 写了主题或插件？在 `themes/` 或 `plugins/` 下加一个以它的名字命名的文件，写上仓库地址和[类别](categories.yaml)，提一个 pull request 就行：

```yaml
# plugins/search.yaml
repo: https://github.com/kite-plus/plugin-search
category: search
```

名称、说明和版本都从最新的发布版本里读取，不用再写一遍。发布版本要满足什么、会检查哪些内容，见[参与收录](CONTRIBUTING.zh-CN.md)。还没开始写？从官网的[编写主题](https://www.kite.plus/writing-themes/)和[编写插件](https://www.kite.plus/writing-plugins/)开始。

## 许可

[CC0 1.0](LICENSE)：列表和 `index.json` 可以任意使用，不需要征得同意，也不需要署名。各个主题和插件的许可证以它们自己的仓库为准。
