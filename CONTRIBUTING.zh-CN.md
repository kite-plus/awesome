# 参与收录

[English](CONTRIBUTING.md) · **简体中文**

任何别人能装上的 Kite 主题或插件都可以收录：开源、放在 GitHub 上、有发布版本。每一项只需要一个两行的文件，其余内容都从它最新的发布版本里读取，并用最新的 Kite 检查。

## 发布版本要满足的条件

- **GitHub 上的公开仓库**，没有归档，`theme.yaml` 或 `plugin.yaml` 在仓库根目录。
- **发布版本附带能在后台安装的 zip。** 读取的是最新的正式版本（不含预发布）。主题或插件打成 zip 附在版本上，清单文件在 zip 的最外层或唯一的文件夹里，和[风标](https://github.com/kite-plus/theme-vane)的 `scripts/package.sh`、[官方插件](https://github.com/kite-plus/plugin-search)的 `make zip` 打出来的一样。附带多个 zip 时，读取名为 `<名称>-<版本>.zip` 的那个。
- **tag 就是版本号**：`version: 1.2.0` 对应 `v1.2.0` 或 `1.2.0`。
- **能在最新的 Kite 上加载**：用 Kite 最新的发布版本运行 `kite theme verify` 或 `kite plugin verify` 能通过，也就是 `requires` 包含这个版本。
- **清单里写了说明和许可证**，许可证是开源许可证。
- **主题要有截图**，放在 Kite 查找的位置：`screenshot.png`、`.jpg`、`.jpeg` 或 `.webp`，或者 `screenshot:` 指定的文件。README 从仓库里对应 tag 的位置显示截图，所以截图既要打进 zip，也要提交在仓库里。Kite 建议的尺寸是 1280×800。
- **插件在 `hosts` 里写明自带脚本会访问的其他网站。** 列表会把它们和注入代码里写到的网站一起，列在插件的作用旁边。

欢迎提供、但不强制：中文 README 用到的语言包 `i18n/zh-CN.yaml`；清单不是用英文写的，再加一份 `i18n/en.yaml`。列表按后台的规则读取它们。给仓库加上 GitHub topic `kite-theme` 或 `kite-plugin`，别人也更容易找到。

## 提交

1. 加一个以主题或插件命名的文件：主题是 `themes/<名称>.yaml`，用主题的 `name`；插件是 `plugins/<id>.yaml`，用插件的 `id`。这个名字就是安装时的目录名，已经被收录的名字不能重复。

   ```yaml
   repo: https://github.com/you/kite-theme-paper
   category: blog
   ```

   类别见 [categories.yaml](categories.yaml)。都不合适的话，在同一个 pull request 里加一个，写上英文和中文名称。
2. 提交 pull request。检查会读取最新的发布版本，按后台的方式解压，用最新的 Kite 运行 `kite theme verify` 或 `kite plugin verify`，并打印出列表里会怎么介绍它。

不用改 README 和 `index.json`：pull request 合并后，维护者运行 `make sync` 生成它们。

不方便提 pull request？可以[在 issue 里推荐](https://github.com/kite-plus/awesome/issues/new?template=suggest.yml)。

## 更新与下架

发布新版本即可：每次同步都会读取最新的版本。每周会用 Kite 最新的发布版本检查一遍所有条目。最新版本不再通过检查的条目，会在下次同步时移出列表，等有版本重新通过后再回来，它的文件一直保留。想永久下架，提一个删除这个文件的 pull request。

## 维护

```bash
make check                          # 离线检查条目、index.json 和 README
make sync                           # 读取所有最新版本，生成 index.json 和 README
make sync ARGS=plugins/search.yaml  # 只同步一项
make sync ARGS=-dry-run             # 只检查和报告，不写任何文件
make readme                         # 根据 index.json 重新生成 README
make test                           # 测试工具本身
```

`make sync` 需要 Go 和 Kite 最新的发布版本：放在 `PATH` 里，或者用 `ARGS=-kite=/path/to/kite` 指定。从源码构建的 Kite 不是网站实际安装的那个版本，不能用来检查。它用 `GH_TOKEN` 访问 GitHub，GitHub CLI 登录后会自动从 `gh auth token` 取得，避免匿名请求的次数限制。
