package site

import (
	"context"
	"fmt"
	"html"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"cairn/internal/cli"
	"cairn/internal/config"
	"cairn/internal/mcp"
	"cairn/internal/theme"
	"cairn/test"
)

const windowless = "CAIRN_WINDOWLESS"

func TestAcceptance_Site(t *testing.T) {
	t.Parallel()
	if test.Suite(t, "site.feature", initSite, "~@later").Run() != 0 {
		t.Fatal("acceptance site.feature failed")
	}
}

type siteWorld struct {
	changelog string
	out       string
	version   string
	cleanup   []func()
}

func initSite(sc *godog.ScenarioContext) {
	w := &siteWorld{}
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		for _, done := range w.cleanup {
			done()
		}
		return ctx, err
	})
	sc.Step(`^a changelog that says:$`, w.aChangelogThatSays)
	sc.Step(`^the site is built for version "([^"]*)"$`, w.theSiteIsBuiltForVersion)
	sc.Step(`^the site has a page for "([^"]*)"$`, w.theSiteHasAPageFor)
	sc.Step(`^the site has no page for "([^"]*)"$`, w.theSiteHasNoPageFor)
	sc.Step(`^every page of docs/ is published or named as a developer's$`, w.everyPageOfDocsIsPlaced)
	sc.Step(`^every link from one page of the site to another lands on a page that exists$`, w.everyLinkLandsOnAPage)
	sc.Step(`^every link to a heading lands on a heading of that page$`, w.everyLinkLandsOnAHeading)
	sc.Step(`^no link on the site points at a Markdown file of its own$`, w.noLinkPointsAtMarkdown)
	sc.Step(`^the link from "([^"]*)" to "([^"]*)" goes to it on GitHub$`, w.theLinkGoesToGitHub)
	sc.Step("^the reference lists every command cairn has, with the first line of its help$", w.theReferenceListsEveryCommand)
	sc.Step("^the page of `cairn ([^`]*)` lists each of its flags with what its help says$", w.thePageOfACommandListsItsFlags)
	sc.Step(`^the reference lists every MCP tool cairn serves, with what it does$`, w.theReferenceListsEveryTool)
	sc.Step("^the configuration reference shows the file `cairn config init` writes$", w.theConfigurationShowsTheFile)
	sc.Step(`^the site offers every scheme cairn carries$`, w.theSiteOffersEveryScheme)
	sc.Step(`^a first visit wears "([^"]*)" when the system is dark, and "([^"]*)" when it is light$`, w.aFirstVisitWears)
	sc.Step(`^the scheme chosen is read back from the browser before the page is drawn$`, w.theSchemeIsReadBeforeDrawing)
	sc.Step(`^every page says "([^"]*)"$`, w.everyPageSays)
	sc.Step(`^the changelog page says "([^"]*)"$`, w.theChangelogPageSays)
	sc.Step(`^every page is in what the search reads$`, w.everyPageIsSearched)
	sc.Step(`^the site carries a picture of the window in every scheme cairn carries$`, w.aPictureOfTheWindowInEveryScheme)
	sc.Step(`^each picture is drawn on the back of its own scheme$`, w.eachPictureOnItsOwnBack)
	sc.Step(`^the home page shows the terminal board drawn over the fixtures$`, w.theHomePageShowsTheTerminal)
	sc.Step(`^each scheme gives the terminal its sixteen colours$`, w.eachSchemeGivesTheTerminalItsColours)
}

func (w *siteWorld) aChangelogThatSays(said *godog.DocString) error {
	w.changelog = said.Content
	return nil
}

func (w *siteWorld) theSiteIsBuiltForVersion(version string) error {
	dir, err := os.MkdirTemp("", "cairn-site-")
	if err != nil {
		return err
	}
	w.cleanup = append(w.cleanup, func() { _ = os.RemoveAll(dir) })
	boards, err := OpenFixtures(filepath.Join(dir, "fixtures"))
	if err != nil {
		return err
	}
	w.cleanup = append([]func(){func() { _ = boards.Close() }}, w.cleanup...)
	terminal, err := boards.Terminal()
	if err != nil {
		return err
	}
	said := Said{Version: version, Changelog: []byte(w.changelog), Repo: os.DirFS(filepath.Join("..", "..")), Terminal: terminal}
	if os.Getenv(windowless) == "" {
		said.Window = boards.Window
	}
	w.out, w.version = filepath.Join(dir, "site"), version
	return Build(w.out, said)
}

func (w *siteWorld) read(name string) (string, error) {
	data, err := os.ReadFile(filepath.Join(w.out, name))
	return string(data), err
}

func (w *siteWorld) pages() (map[string]string, error) {
	found := map[string]string{}
	err := filepath.WalkDir(w.out, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".html" {
			return err
		}
		data, err := os.ReadFile(path)
		rel, _ := filepath.Rel(w.out, path)
		found[filepath.ToSlash(rel)] = string(data)
		return err
	})
	if len(found) == 0 && err == nil {
		err = fmt.Errorf("the site in %s has no page at all", w.out)
	}
	return found, err
}

func (w *siteWorld) theSiteHasAPageFor(source string) error {
	name, ok := pageFor(source)
	if !ok {
		return fmt.Errorf("%s is not published", source)
	}
	page, err := w.read(name)
	if err != nil {
		return err
	}
	if !strings.Contains(page, "<article") {
		return fmt.Errorf("%s has no article in it", name)
	}
	return nil
}

func (w *siteWorld) theSiteHasNoPageFor(source string) error {
	if name, ok := pageFor(source); ok {
		return fmt.Errorf("%s is published as %s", source, name)
	}
	plain := strings.TrimSuffix(filepath.Base(source), ".md") + ".html"
	if _, err := os.Stat(filepath.Join(w.out, plain)); err == nil {
		return fmt.Errorf("the site carries %s", plain)
	}
	return nil
}

func (w *siteWorld) everyPageOfDocsIsPlaced() error {
	docs, err := filepath.Glob(filepath.Join("..", "..", "docs", "*.md"))
	if err != nil {
		return err
	}
	for _, doc := range docs {
		source := "docs/" + filepath.Base(doc)
		if _, ok := pageFor(source); !ok && !slices.Contains(forDevelopers(), source) {
			return fmt.Errorf("%s is neither published nor named as a developer's", source)
		}
	}
	return nil
}

var (
	hrefs = regexp.MustCompile(`href="([^"]*)"`)
	idsOf = regexp.MustCompile(` id="([^"]*)"`)
)

func local(href string) bool {
	return !strings.Contains(href, "://") && !strings.HasPrefix(href, "mailto:")
}

func (w *siteWorld) eachLocalLink(do func(from, page, fragment string) error) error {
	pages, err := w.pages()
	if err != nil {
		return err
	}
	for from, body := range pages {
		for _, found := range hrefs.FindAllStringSubmatch(body, -1) {
			href := html.UnescapeString(found[1])
			if !local(href) || strings.HasSuffix(href, ".css") || strings.HasSuffix(href, ".svg") || strings.HasSuffix(href, ".png") {
				continue
			}
			page, fragment, _ := strings.Cut(href, "#")
			if page == "" {
				page = from
			}
			if err := do(from, page, fragment); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *siteWorld) everyLinkLandsOnAPage() error {
	return w.eachLocalLink(func(from, page, _ string) error {
		if _, err := os.Stat(filepath.Join(w.out, page)); err != nil {
			return fmt.Errorf("%s links to %s, which is not on the site", from, page)
		}
		return nil
	})
}

func (w *siteWorld) everyLinkLandsOnAHeading() error {
	pages, err := w.pages()
	if err != nil {
		return err
	}
	return w.eachLocalLink(func(from, page, fragment string) error {
		if fragment == "" {
			return nil
		}
		for _, found := range idsOf.FindAllStringSubmatch(pages[page], -1) {
			if html.UnescapeString(found[1]) == fragment {
				return nil
			}
		}
		return fmt.Errorf("%s links to %s#%s, and %s has no such heading", from, page, fragment, page)
	})
}

func (w *siteWorld) noLinkPointsAtMarkdown() error {
	return w.eachLocalLink(func(from, page, _ string) error {
		if strings.HasSuffix(page, ".md") {
			return fmt.Errorf("%s links to %s", from, page)
		}
		return nil
	})
}

func (w *siteWorld) theLinkGoesToGitHub(from, to string) error {
	name, ok := pageFor(from)
	if !ok {
		return fmt.Errorf("%s is not published", from)
	}
	page, err := w.read(name)
	if err != nil {
		return err
	}
	want := fmt.Sprintf(`href="https://github.com/vonglasow/cairn/blob/%s/%s`, w.version, to)
	if !strings.Contains(page, want) {
		return fmt.Errorf("%s has no link starting %s", name, want)
	}
	return nil
}

func commands(c *cobra.Command) []*cobra.Command {
	var found []*cobra.Command
	for _, sub := range c.Commands() {
		if sub.Hidden || !sub.IsAvailableCommand() {
			continue
		}
		if sub.HasAvailableSubCommands() {
			found = append(found, commands(sub)...)
			continue
		}
		found = append(found, sub)
	}
	return found
}

func (w *siteWorld) theReferenceListsEveryCommand() error {
	index, err := w.read(commandsPage)
	if err != nil {
		return err
	}
	for _, c := range commands(cli.Reference()) {
		if !strings.Contains(index, html.EscapeString(c.CommandPath())) || !strings.Contains(index, html.EscapeString(c.Short)) {
			return fmt.Errorf("the commands page does not list %q with %q", c.CommandPath(), c.Short)
		}
		if _, err := w.read(commandPage(c)); err != nil {
			return fmt.Errorf("%s has no page of its own: %w", c.CommandPath(), err)
		}
	}
	return nil
}

func (w *siteWorld) thePageOfACommandListsItsFlags(line string) error {
	c, _, err := cli.Reference().Find(strings.Fields(line))
	if err != nil {
		return err
	}
	page, err := w.read(commandPage(c))
	if err != nil {
		return err
	}
	var missing []string
	c.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if !strings.Contains(page, "--"+f.Name) || !strings.Contains(page, html.EscapeString(f.Usage)) {
			missing = append(missing, f.Name)
		}
	})
	if len(missing) > 0 {
		return fmt.Errorf("the page of %s does not say %v", line, missing)
	}
	return nil
}

func (w *siteWorld) theReferenceListsEveryTool() error {
	page, err := w.read(toolsPage)
	if err != nil {
		return err
	}
	for _, tool := range mcp.Tools() {
		if !strings.Contains(page, `id="`+tool.Name+`"`) || !strings.Contains(page, html.EscapeString(tool.Description)) {
			return fmt.Errorf("the tools page does not list %s with what it does", tool.Name)
		}
	}
	return nil
}

func (w *siteWorld) theConfigurationShowsTheFile() error {
	name, _ := pageFor("docs/config.md")
	page, err := w.read(name)
	if err != nil {
		return err
	}
	if !strings.Contains(page, html.EscapeString(config.Example)) {
		return fmt.Errorf("%s does not show the file config init writes", name)
	}
	return nil
}

func (w *siteWorld) theSiteOffersEveryScheme() error {
	pages, err := w.pages()
	if err != nil {
		return err
	}
	style, err := w.read(styleSheet)
	if err != nil {
		return err
	}
	for _, slug := range theme.Shipped() {
		if !strings.Contains(style, `[data-scheme="`+slug+`"]`) {
			return fmt.Errorf("the style sheet has no colours for %s", slug)
		}
		for name, page := range pages {
			if !strings.Contains(page, `data-pick="`+slug+`"`) {
				return fmt.Errorf("%s offers no %s", name, slug)
			}
		}
	}
	return nil
}

func (w *siteWorld) head(name string) (string, error) {
	page, err := w.read(name)
	if err != nil {
		return "", err
	}
	head, _, found := strings.Cut(page, "<body")
	if !found {
		return "", fmt.Errorf("%s has no body", name)
	}
	return head, nil
}

func (w *siteWorld) aFirstVisitWears(dark, light string) error {
	head, err := w.head("index.html")
	if err != nil {
		return err
	}
	want := fmt.Sprintf(`matchMedia("(prefers-color-scheme: dark)").matches ? %q : %q`, dark, light)
	if !strings.Contains(head, want) {
		return fmt.Errorf("the head of the home page does not choose %s", want)
	}
	return nil
}

func (w *siteWorld) theSchemeIsReadBeforeDrawing() error {
	head, err := w.head("index.html")
	if err != nil {
		return err
	}
	if !strings.Contains(head, `localStorage.getItem("cairn-theme")`) {
		return fmt.Errorf("the head of the home page does not read the scheme kept")
	}
	return nil
}

func (w *siteWorld) everyPageSays(said string) error {
	pages, err := w.pages()
	if err != nil {
		return err
	}
	for name, page := range pages {
		if !strings.Contains(page, said) {
			return fmt.Errorf("%s does not say %s", name, said)
		}
	}
	return nil
}

func (w *siteWorld) theChangelogPageSays(said string) error {
	page, err := w.read(changelogPage)
	if err != nil {
		return err
	}
	if !strings.Contains(page, said) {
		return fmt.Errorf("the changelog page does not say %q", said)
	}
	return nil
}

func (w *siteWorld) everyPageIsSearched() error {
	pages, err := w.pages()
	if err != nil {
		return err
	}
	index, err := w.read(searchIndex)
	if err != nil {
		return err
	}
	for name := range pages {
		if !strings.Contains(index, `"u":"`+name+`"`) {
			return fmt.Errorf("the search does not read %s", name)
		}
	}
	return nil
}

func (w *siteWorld) aPictureOfTheWindowInEveryScheme() error {
	if os.Getenv(windowless) != "" {
		return godog.ErrSkip
	}
	pages, err := w.pages()
	if err != nil {
		return err
	}
	for _, slug := range theme.Shipped() {
		if _, err := os.Stat(filepath.Join(w.out, windowPicture(slug))); err != nil {
			return err
		}
	}
	if !strings.Contains(pages["index.html"], `src="`+windowPicture("default")+`"`) {
		return fmt.Errorf("the home page shows no picture of the window")
	}
	return nil
}

func (w *siteWorld) eachPictureOnItsOwnBack() error {
	if os.Getenv(windowless) != "" {
		return godog.ErrSkip
	}
	for _, slug := range theme.Shipped() {
		file, err := os.Open(filepath.Join(w.out, windowPicture(slug)))
		if err != nil {
			return err
		}
		picture, err := png.Decode(file)
		_ = file.Close()
		if err != nil {
			return err
		}
		scheme, err := theme.Find("", slug)
		if err != nil {
			return err
		}
		r, g, b, _ := picture.At(0, 0).RGBA()
		back := scheme.Colours().Surface.Back
		if uint8(r>>8) != back.R || uint8(g>>8) != back.G || uint8(b>>8) != back.B {
			return fmt.Errorf("the picture in %s is not drawn on its back %v", slug, back)
		}
	}
	return nil
}

func (w *siteWorld) theHomePageShowsTheTerminal() error {
	page, err := w.read("index.html")
	if err != nil {
		return err
	}
	_, term, found := strings.Cut(page, `<pre class="term"`)
	if !found || !strings.Contains(term, "BACKLOG") || !strings.Contains(term, `class="t4`) {
		return fmt.Errorf("the home page shows no terminal board in colour")
	}
	return nil
}

func (w *siteWorld) eachSchemeGivesTheTerminalItsColours() error {
	style, err := w.read(styleSheet)
	if err != nil {
		return err
	}
	for _, slug := range theme.Shipped() {
		scheme, err := theme.Find("", slug)
		if err != nil {
			return err
		}
		_, rules, _ := strings.Cut(style, `[data-scheme="`+slug+`"]`)
		rules, _, _ = strings.Cut(rules, "}")
		for i, slot := range terminal16(scheme) {
			want := fmt.Sprintf("--t%d:#%02x%02x%02x", i, slot.R, slot.G, slot.B)
			if !strings.Contains(rules, want) {
				return fmt.Errorf("%s does not give the terminal %s", slug, want)
			}
		}
	}
	return nil
}
