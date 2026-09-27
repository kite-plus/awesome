package main

import (
	"cmp"
	"fmt"
	"html"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// readme is one of the READMEs, and the words its generated parts are
// written in.
type readme struct {
	File string
	Lang string // the language packs are picked for

	Checked   string // with the version of Kite
	Unchecked string
	NoThemes  string
	NoPlugins string
	Official  string
	By        string // with the author
	Plugins   string // the table's header
	None      string // for a plugin that loads nothing from other sites
	Shot      string // with the theme's title

	// The badges' links to the lists, and what they say to a screen reader.
	ThemesAnchor  string
	PluginsAnchor string
	CheckedAlt    string // with the version of Kite
	ThemesAlt     string // with the number listed
	PluginsAlt    string
}

var readmes = []readme{
	{
		File:      "README.md",
		Lang:      "en",
		Checked:   "The latest release of every entry has been checked with Kite %s, the latest version: it installs, and it loads.",
		Unchecked: "Nothing has been checked yet.",
		NoThemes:  "No theme is listed yet.",
		NoPlugins: "No plugin is listed yet.",
		Official:  "Official",
		By:        "by %s",
		Plugins:   "| Plugin | What it does | Other sites it loads from |",
		None:      "None",
		Shot:      "Screenshot of %s",

		ThemesAnchor:  "#themes",
		PluginsAnchor: "#plugins",
		CheckedAlt:    "Checked with Kite %s",
		ThemesAlt:     "Themes listed: %d",
		PluginsAlt:    "Plugins listed: %d",
	},
	{
		File:      "README.zh-CN.md",
		Lang:      "zh-CN",
		Checked:   "每一项的最新版本都用 Kite %s（目前最新的版本）检查过：能安装，能加载。",
		Unchecked: "还没有检查过任何一项。",
		NoThemes:  "还没有收录主题。",
		NoPlugins: "还没有收录插件。",
		Official:  "官方",
		By:        "作者 %s",
		Plugins:   "| 插件 | 作用 | 会连接的其他网站 |",
		None:      "无",
		Shot:      "%s的截图",

		ThemesAnchor:  "#主题",
		PluginsAnchor: "#插件",
		CheckedAlt:    "用 Kite %s 检查过",
		ThemesAlt:     "收录的主题：%d",
		PluginsAlt:    "收录的插件：%d",
	},
}

func (r readme) category(c Category) string {
	if r.Lang == "en" {
		return c.En
	}
	return c.ZhCN
}

// blocks are the parts of the README written from the index, by the name
// their markers carry.
func (r readme) blocks(ix *Index, cats Categories) map[string]string {
	status := r.Unchecked
	if ix.Kite != "" {
		status = fmt.Sprintf(r.Checked, ix.Kite)
	}
	return map[string]string{
		"badges":  r.badges(ix),
		"status":  status,
		"themes":  r.themes(ix, cats),
		"plugins": r.plugins(ix, cats),
	}
}

// blockNames are the blocks, in the order they come in a README.
var blockNames = []string{"badges", "status", "themes", "plugins"}

// kiteLogo is Kite's logo in white, as the official plugins' badges carry
// it, and brand is Kite's blue.
const (
	kiteLogo = "data:image/svg%2bxml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCA2NCA2NCI+PGcgZmlsbD0iI2ZmZiIgc3Ryb2tlPSIjZmZmIiBzdHJva2Utd2lkdGg9IjUiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxwYXRoIGQ9Ik0xMCAxNC41IEwyNyAyMSBMMjcgMzAgTDEwIDIzLjUgWiIvPjxwYXRoIGQ9Ik0xMCAzMiBMMjcgMzguNSBMMjcgNDkgTDEwIDQyLjUgWiIvPjxwYXRoIGQ9Ik0zNyAyMSBMNTQgMTQuNSBMNTQgNDIuNSBMMzcgNDkgWiIvPjwvZz48L3N2Zz4="
	brand    = "4A77D6"
)

// badges say which Kite checked the list and how much it holds, so that
// the numbers change with the list rather than by hand.
func (r readme) badges(ix *Index) string {
	var lines []string
	if ix.Kite != "" {
		lines = append(lines, badge("https://github.com/kite-plus/kite/releases/tag/v"+ix.Kite,
			shield("checked with", "Kite "+ix.Kite)+"?logo="+kiteLogo, fmt.Sprintf(r.CheckedAlt, ix.Kite)))
	}
	lines = append(lines,
		badge(r.ThemesAnchor, shield("themes", strconv.Itoa(len(ix.Themes))), fmt.Sprintf(r.ThemesAlt, len(ix.Themes))),
		badge(r.PluginsAnchor, shield("plugins", strconv.Itoa(len(ix.Plugins))), fmt.Sprintf(r.PluginsAlt, len(ix.Plugins))),
	)
	return strings.Join(lines, "\n")
}

// shield is the address of a badge from shields.io, where a dash or an
// underscore in the text is written twice.
func shield(label, message string) string {
	esc := func(s string) string {
		return url.PathEscape(strings.NewReplacer("-", "--", "_", "__").Replace(s))
	}
	return "https://img.shields.io/badge/" + esc(label) + "-" + esc(message) + "-" + brand
}

func badge(href, src, alt string) string {
	return `  <a href="` + html.EscapeString(href) + `"><img src="` + html.EscapeString(src) +
		`" alt="` + html.EscapeString(alt) + `"></a>`
}

// listed orders a kind's records by category, then by name.
func listed(ix *Index, cats Categories, k Kind) []Record {
	out := slices.Clone(ix.Of(k))
	slices.SortFunc(out, func(a, b Record) int {
		return cmp.Or(
			cmp.Compare(cats.Rank(k, a.Category), cats.Rank(k, b.Category)),
			strings.Compare(a.Name, b.Name),
		)
	})
	return out
}

// markup writes words and links for the part of a README they go in: HTML
// for the themes, which sit in an HTML table, and Markdown for the plugins.
type markup struct {
	text func(string) string
	link func(text, href string) string
}

var (
	asHTML     = markup{text: htmlText, link: htmlLink}
	asMarkdown = markup{text: text, link: mdLink}
)

// label is the line under a title: its category, who made it, and the
// version, linked to its release.
func (r readme) label(k Kind, rec Record, cats Categories, m markup) string {
	parts := []string{m.text(r.category(cats.Of(k)[cats.Rank(k, rec.Category)]))}
	switch {
	case rec.Official:
		parts = append(parts, r.Official)
	case rec.Author != nil && webAddress(rec.Author.URL):
		parts = append(parts, fmt.Sprintf(r.By, m.link(rec.Author.Name, rec.Author.URL)))
	case rec.Author != nil:
		parts = append(parts, fmt.Sprintf(r.By, m.text(rec.Author.Name)))
	}
	parts = append(parts, m.link(rec.Version, rec.Release))
	return strings.Join(parts, " · ")
}

func webAddress(s string) bool {
	return strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "http://")
}

// themes is a table of screenshots, each beside what the theme says.
func (r readme) themes(ix *Index, cats Categories) string {
	records := listed(ix, cats, Theme)
	if len(records) == 0 {
		return r.NoThemes
	}
	var b strings.Builder
	b.WriteString("<table>\n")
	for _, rec := range records {
		words := r.Words(rec)
		repo := html.EscapeString(rec.Repo)
		fmt.Fprintf(&b, "<tr>\n<td width=\"320\" valign=\"top\"><a href=\"%s\"><img src=\"%s\" width=\"320\" alt=\"%s\"></a></td>\n",
			repo, html.EscapeString(rec.Screenshot), html.EscapeString(fmt.Sprintf(r.Shot, words.Title)))
		fmt.Fprintf(&b, "<td valign=\"top\"><strong><a href=\"%s\">%s</a></strong><br><sub>%s</sub><br><br>%s</td>\n</tr>\n",
			repo, htmlText(words.Title), r.label(Theme, rec, cats, asHTML), htmlText(words.Description))
	}
	b.WriteString("</table>")
	return b.String()
}

// plugins is a table of what each plugin does and which other sites its
// code loads from.
func (r readme) plugins(ix *Index, cats Categories) string {
	records := listed(ix, cats, Plugin)
	if len(records) == 0 {
		return r.NoPlugins
	}
	var b strings.Builder
	b.WriteString(r.Plugins + "\n| --- | --- | --- |\n")
	for _, rec := range records {
		words := r.Words(rec)
		sites := r.None
		if len(rec.Hosts) > 0 {
			sites = "`" + strings.Join(rec.Hosts, "`<br>`") + "`"
		}
		fmt.Fprintf(&b, "| **%s**<br><sub>%s</sub> | %s | %s |\n",
			mdLink(words.Title, rec.Repo), r.label(Plugin, rec, cats, asMarkdown), text(words.Description), sites)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func (r readme) Words(rec Record) Words { return rec.Words(r.Lang) }

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

var markdownSpecial = strings.NewReplacer(
	`&`, "&amp;", `<`, "&lt;", `>`, "&gt;",
	`\`, `\\`, "`", "\\`", `*`, `\*`, `_`, `\_`,
	`[`, `\[`, `]`, `\]`, `|`, `\|`, `~`, `\~`,
)

// text makes what a manifest says safe in a Markdown table: on one line,
// and read as words rather than as markup.
func text(s string) string { return markdownSpecial.Replace(oneLine(s)) }

func mdLink(s, href string) string {
	return "[" + text(s) + "](" + strings.NewReplacer("(", "%28", ")", "%29", " ", "%20").Replace(href) + ")"
}

func htmlText(s string) string { return html.EscapeString(oneLine(s)) }

func htmlLink(s, href string) string {
	return `<a href="` + html.EscapeString(href) + `">` + htmlText(s) + "</a>"
}

// replaceBlock puts body between a block's markers.
func replaceBlock(doc, name, body string) (string, error) {
	begin := "<!-- BEGIN awesome:" + name + " -->"
	end := "<!-- END awesome:" + name + " -->"
	i := strings.Index(doc, begin)
	j := strings.Index(doc, end)
	if i < 0 || j < i {
		return "", fmt.Errorf("it has no %s and %s around the part the list writes", begin, end)
	}
	return doc[:i+len(begin)] + "\n" + body + "\n" + doc[j:], nil
}

// rendered is a README with its blocks written from the index.
func (r readme) rendered(root string, ix *Index, cats Categories) (current, want string, err error) {
	data, err := os.ReadFile(filepath.Join(root, r.File))
	if err != nil {
		return "", "", err
	}
	current, want = string(data), string(data)
	blocks := r.blocks(ix, cats)
	for _, name := range blockNames {
		if want, err = replaceBlock(want, name, blocks[name]); err != nil {
			return "", "", fmt.Errorf("%s: %w", r.File, err)
		}
	}
	return current, want, nil
}

func writeReadmes(root string, ix *Index, cats Categories) error {
	for _, r := range readmes {
		_, want, err := r.rendered(root, ix, cats)
		if err != nil {
			return err
		}
		if err := writeFile(root, r.File, []byte(want)); err != nil {
			return err
		}
	}
	return nil
}

func runReadme(root string) error {
	cats, err := loadCategories(root)
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
	return writeReadmes(root, ix, cats)
}
