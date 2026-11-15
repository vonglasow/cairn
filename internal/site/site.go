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
	"slices"
	"strings"

	"cairn/internal/theme"

	"maps"
)

//go:embed assets
var assets embed.FS

var layout = template.Must(template.ParseFS(assets, "assets/layout.html"))

// Said is what the site is built from: the version it documents, the repository its pages are read from, and the two boards as cairn draws them.
type Said struct {
	Version  string
	Repo     fs.FS
	Terminal string
	Window   func(slug, page string) (image.Image, error)
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
	Home    string
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
		p := page{Name: nameOf(s.from), Title: got.title, Group: s.group, Top: "docs", Source: s.from, Body: got.body, Toc: got.toc}
		if p.Name == downloadPage {
			p.Top = "download"
		}
		if at := slices.IndexFunc(byTask, func(g guide) bool { return g.from == s.from }); at >= 0 && said.Window != nil {
			p.Body = illustrated(p.Body, byTask[at].pictures)
		}
		pages = append(pages, p)
	}
	return pages, nil
}

func illustrated(body string, pictures []picture) string {
	for _, p := range pictures {
		if p.at != "" {
			body = strings.Replace(body, `<h2 id="`+p.at+`">`, drawing(p.of, p.shows)+`<h2 id="`+p.at+`">`, 1)
			continue
		}
		lead, rest, found := strings.Cut(body, "</p>\n")
		if !found || !strings.HasPrefix(lead, "<p>") {
			body = drawing(p.of, p.shows) + body
			continue
		}
		body = lead + "</p>\n" + drawing(p.of, p.shows) + rest
	}
	return body
}

func ordered(docs []page, generated ...page) []page {
	by := map[string]page{}
	for _, p := range generated {
		by[p.Name] = p
	}
	pages := []page{by[homePage], by[docsPage], by[agentsPage]}
	for _, group := range groups {
		for _, p := range docs {
			if p.Group == group {
				pages = append(pages, p)
			}
		}
	}
	for _, name := range []string{commandsPage, toolsPage, keysPage, themesPage} {
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
		drawn, err := said.Window(s.Slug, "tasks")
		if err != nil {
			return fmt.Errorf("the window in %s: %w", s.Slug, err)
		}
		if err := saved(filepath.Join(to, windowPicture(s.Slug)), drawn); err != nil {
			return err
		}
	}
	for _, of := range shots {
		drawn, err := said.Window(darkFirst, of)
		if err != nil {
			return fmt.Errorf("the window on %s: %w", of, err)
		}
		if err := saved(filepath.Join(to, shotPicture(of)), drawn); err != nil {
			return err
		}
	}
	return nil
}

func saved(file string, drawn image.Image) error {
	var out bytes.Buffer
	if err := png.Encode(&out, drawn); err != nil {
		return err
	}
	return os.WriteFile(file, out.Bytes(), 0o644)
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
	for _, role := range slices.Sorted(maps.Values(drawnAs)) {
		fmt.Fprintf(&css, ".term .tp-%s{color:var(--p-%s)}.term .up-%s{background:var(--p-%s)}\n", role, role, role, role)
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
	common := shown{Version: said.Version, Dark: darkFirst, Light: lightFirst, Source: repository, Home: address}
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
	generated := append(commandPages(), toolsPageOf(), keysPageOf(), themesPageOf(schemes, said.Window != nil), home(said, schemes), agents(said.Version))
	pages := ordered(docs, append(generated, docsHome(append(docs, generated...)))...)
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
	repo := flags.String("repo", ".", "the repository the pages are read from")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *out == "" {
		_, _ = fmt.Fprintln(stderr, ErrNoFolder)
		return 2
	}
	if err := built(*out, *version, *repo); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "the site of %s is in %s\n", *version, *out)
	return 0
}

func built(out, version, repo string) error {
	said := Said{Version: version, Repo: os.DirFS(repo)}
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
