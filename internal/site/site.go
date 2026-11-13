package site

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"image"
	"image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"cairn/internal/theme"
)

//go:embed assets
var assets embed.FS

var layout = template.Must(template.ParseFS(assets, "assets/layout.html"))

// Said is what the site is built from: the version it documents, its changelog, the repository its pages are read from, and the two boards as cairn draws them.
type Said struct {
	Version   string
	Changelog []byte
	Repo      fs.FS
	Terminal  string
	Window    func(slug string) (image.Image, error)
}

type page struct {
	Name, Title, Group, Top, Parent, Source string
	Body                                    string
	Toc                                     []heading
	Wide                                    bool
}

type link struct{ Name, Title string }

type shown struct {
	Page    page
	Body    template.HTML
	Version string
	Groups  []struct {
		Name  string
		Pages []link
	}
	Schemes []pick
	Known   []string
	Dark    string
	Light   string
	Source  string
}

type pick struct {
	Slug, Name string
	Dots       []string
}

type found struct {
	T string   `json:"t"`
	U string   `json:"u"`
	H []string `json:"h"`
	X string   `json:"x"`
}

// ErrNoFolder is a build not told where to write.
var ErrNoFolder = errors.New("--out names the folder the site is written to")

func docPages(said Said) ([]page, error) {
	var pages []page
	for _, s := range published {
		src, err := fs.ReadFile(said.Repo, s.from)
		if err != nil {
			return nil, err
		}
		got, err := write(s.from, src, said.Version)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", s.from, err)
		}
		p := page{Name: nameOf(s.from), Title: got.title, Group: s.group, Top: s.top, Source: s.from, Body: got.body, Toc: got.toc}
		if s.from == "docs/config.md" {
			p.Body = configurationIntro() + p.Body
			p.Toc = append([]heading{{ID: "the-whole-file", Title: "The whole file"}}, p.Toc...)
		}
		pages = append(pages, p)
	}
	return pages, nil
}

func changelogOf(said Said) (page, error) {
	p := page{Name: changelogPage, Title: "Changelog", Group: "Project", Top: "changelog", Body: "<p>This build was given no changelog.</p>"}
	if len(said.Changelog) == 0 {
		return p, nil
	}
	got, err := write("CHANGELOG.md", said.Changelog, said.Version)
	if err != nil {
		return page{}, err
	}
	p.Body, p.Toc = got.body, got.toc
	return p, nil
}

func ordered(docs []page, generated ...page) []page {
	by := map[string]page{}
	for _, p := range append(docs, generated...) {
		by[p.Name] = p
	}
	var pages []page
	for _, name := range []string{"index.html", "install.html", "usage.html"} {
		pages = append(pages, by[name])
	}
	for _, p := range docs {
		if p.Group == "Guide" {
			pages = append(pages, p)
		}
	}
	for _, name := range []string{commandsPage, toolsPage, "mcp.html", "config.html", themesPage, "decisions.html", changelogPage} {
		pages = append(pages, by[name])
	}
	for _, p := range generated {
		if p.Parent != "" {
			pages = append(pages, p)
		}
	}
	return pages
}

func pictures(to string, said Said, schemes []theme.Scheme) error {
	if said.Window == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Join(to, "shots"), 0o755); err != nil {
		return err
	}
	for _, s := range schemes {
		drawn, err := said.Window(s.Slug)
		if err != nil {
			return fmt.Errorf("the window in %s: %w", s.Slug, err)
		}
		var out bytes.Buffer
		if err := png.Encode(&out, drawn); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(to, windowPicture(s.Slug)), out.Bytes(), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func styled(schemes []theme.Scheme) ([]byte, error) {
	base, err := assets.ReadFile("assets/site.css")
	if err != nil {
		return nil, err
	}
	var css bytes.Buffer
	css.Write(base)
	for i := range 16 {
		fmt.Fprintf(&css, ".term .t%d{color:var(--t%d)}.term .u%d{background:var(--t%d)}\n", i, i, i, i)
	}
	for _, s := range schemes {
		css.WriteString(schemeRules(s))
	}
	return css.Bytes(), nil
}

func indexed(pages []page) ([]byte, error) {
	var all []found
	for _, p := range pages {
		f := found{T: p.Title, U: p.Name, X: searchable(p.Body)}
		for _, h := range p.Toc {
			f.H = append(f.H, h.Title)
		}
		all = append(all, f)
	}
	said, err := json.Marshal(all)
	return append(append([]byte("window.cairnSearch="), said...), ";\n"...), err
}

func (s shown) nav(pages []page) shown {
	for _, group := range groups {
		var in []link
		for _, p := range pages {
			if p.Group == group && p.Parent == "" {
				in = append(in, link{Name: p.Name, Title: p.Title})
			}
		}
		s.Groups = append(s.Groups, struct {
			Name  string
			Pages []link
		}{group, in})
	}
	return s
}

// Build writes the site into the folder: every page, its style, its search, and the pictures of the window.
func Build(to string, said Said) error {
	var schemes []theme.Scheme
	common := shown{Version: said.Version, Dark: darkFirst, Light: lightFirst, Source: repository}
	for _, slug := range theme.Shipped() {
		s, err := theme.Find("", slug)
		if err != nil {
			return err
		}
		schemes = append(schemes, s)
		slots := s.Slots()
		common.Schemes = append(common.Schemes, pick{Slug: slug, Name: s.Name, Dots: []string{hex(slots[0]), hex(slots[2]), hex(slots[0xD]), hex(slots[0xB])}})
		common.Known = append(common.Known, slug)
	}
	docs, err := docPages(said)
	if err != nil {
		return err
	}
	changelog, err := changelogOf(said)
	if err != nil {
		return err
	}
	generated := append(commandPages(), toolsPageOf(), themesPageOf(schemes, said.Window != nil), changelog, home(said, schemes))
	pages := ordered(docs, generated...)
	common = common.nav(pages)
	if err := os.MkdirAll(to, 0o755); err != nil {
		return err
	}
	if err := pictures(to, said, schemes); err != nil {
		return err
	}
	logo, err := fs.ReadFile(said.Repo, "internal/logo/cairn.svg")
	if err != nil {
		return err
	}
	css, err := styled(schemes)
	if err != nil {
		return err
	}
	js, err := assets.ReadFile("assets/site.js")
	if err != nil {
		return err
	}
	search, err := indexed(pages)
	if err != nil {
		return err
	}
	for name, data := range map[string][]byte{"logo.svg": logo, styleSheet: css, script: js, searchIndex: search} {
		if err := os.WriteFile(filepath.Join(to, name), data, 0o644); err != nil {
			return err
		}
	}
	for _, p := range pages {
		view := common
		view.Page, view.Body = p, template.HTML(p.Body)
		var out bytes.Buffer
		if err := layout.Execute(&out, view); err != nil {
			return fmt.Errorf("%s: %w", p.Name, err)
		}
		if err := os.WriteFile(filepath.Join(to, p.Name), out.Bytes(), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// Main builds the site from the repository it is run in, over cairn's fixtures, and answers the code the process exits with.
func Main(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("site", flag.ContinueOnError)
	flags.SetOutput(stderr)
	out := flags.String("out", "", "the folder the site is written to")
	version := flags.String("version", "dev", "the version the site documents")
	changelog := flags.String("changelog", "", "a file holding the changelog, as scripts/changelog.sh writes it")
	repo := flags.String("repo", ".", "the repository the pages are read from")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *out == "" {
		_, _ = fmt.Fprintln(stderr, ErrNoFolder)
		return 2
	}
	if err := built(*out, *version, *changelog, *repo); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "the site of %s is in %s\n", *version, *out)
	return 0
}

func built(out, version, changelog, repo string) error {
	said := Said{Version: version, Repo: os.DirFS(repo)}
	if changelog != "" {
		data, err := os.ReadFile(changelog)
		if err != nil {
			return err
		}
		said.Changelog = data
	}
	folder, err := os.MkdirTemp("", "cairn-site-fixtures-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(folder) }()
	boards, err := OpenFixtures(folder)
	if err != nil {
		return err
	}
	defer func() { _ = boards.Close() }()
	if said.Terminal, err = boards.Terminal(); err != nil {
		return err
	}
	said.Window = boards.Window
	return Build(out, said)
}

func home(said Said, schemes []theme.Scheme) page {
	var body strings.Builder
	body.WriteString(`<section class="hero"><div>
<h1>Your objectives and tasks on one board, wherever you work.</h1>
<p class="lead">A command line, a board in the terminal, the same board in a window, and an MCP server for your agent. Four ways in to one store that stays on your machine.</p>
<pre><code>brew install vonglasow/tap/cairn</code></pre>
<div class="cta"><a class="btn" href="install.html">Install</a><a class="btn ghost" href="usage.html">Read the guide</a></div>
</div>`)
	if said.Window != nil {
		fmt.Fprintf(&body, `<img class="shot" data-shot src="%s" alt="The board in a window">`, windowPicture(darkFirst))
	}
	body.WriteString(`</section>
<section class="band"><h2 id="four-surfaces">Four surfaces, one engine</h2><div class="grid3">
<a class="tile" href="` + commandsPage + `"><span class="k">cairn task new</span><b>Command line</b><span>Every change, scriptable, with <code>--json</code> for machines.</span></a>
<a class="tile" href="tui.html"><span class="k">cairn tui</span><b>Terminal board</b><span>The board in any terminal, a key for everything.</span></a>
<a class="tile" href="gui.html"><span class="k">cairn gui</span><b>Window</b><span>The same board on macOS, Linux and Windows.</span></a>
<a class="tile" href="claude-code.html"><span class="k">cairn mcp serve</span><b>Agent</b><span>An MCP server and a Claude Code plugin.</span></a>
<a class="tile" href="ordering.html"><span class="k">⊳3</span><b>Choose what to do</b><span>Facts on every task. Your agent proposes, you decide.</span></a>
<a class="tile" href="trace.html"><span class="k">cairn trace list</span><b>What happened</b><span>What was done to every task, and the time it took.</span></a>
</div></section>
<section class="band"><h2 id="in-the-terminal">In the terminal</h2><pre class="term">`)
	body.WriteString(terminalHTML(said.Terminal))
	body.WriteString(`</pre></section>
<section class="band"><h2 id="wear-it">Wear it your way</h2><p class="faint">Every scheme cairn carries, the same in the window and on this site. Pick one at the top: the site keeps it for your next visit. <a href="` + themesPage + `">All of them</a>.</p><div class="schemes">`)
	for _, s := range schemes {
		fmt.Fprintf(&body, `<button class="scheme" data-pick="%s"><span class="sw">`, s.Slug)
		for _, slot := range []int{0x0, 0x2, 0x5, 0xD, 0xB, 0xA, 0x8} {
			fmt.Fprintf(&body, `<i style="background:%s"></i>`, hex(s.Slots()[slot]))
		}
		fmt.Fprintf(&body, `</span><b>%s</b></button>`, e(s.Name))
	}
	body.WriteString(`</div></section>
<section class="band"><div class="grid2">
<div class="tile"><b>Local first</b><span>One SQLite file. No account, no server, nothing sent anywhere.</span></div>
<div class="tile"><b>Says why</b><span>Every refusal names what was in the way, and every decision cairn does not take is written down in <a href="decisions.html">Decisions</a>.</span></div>
</div></section>`)
	return page{Name: "index.html", Title: "cairn", Wide: true, Body: body.String()}
}
