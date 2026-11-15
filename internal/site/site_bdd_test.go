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
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"cairn/internal/cli"
	"cairn/internal/store/files"
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
	sc.Step("^the reference lists every command cairn has, with the first line of its help$", w.theReferenceListsEveryCommand)
	sc.Step("^the page of `cairn ([^`]*)` lists each of its flags with what its help says$", w.thePageOfACommandListsItsFlags)
	sc.Step(`^the reference lists every key of the configuration, with what it takes and what it says$`, w.theReferenceListsEveryKey)
	sc.Step(`^every page says "([^"]*)"$`, w.everyPageSays)
	sc.Step(`^the site carries a picture of the window in every scheme cairn carries$`, w.aPictureOfTheWindowInEveryScheme)
	sc.Step(`^each picture is drawn on the back of its own scheme$`, w.eachPictureOnItsOwnBack)
	sc.Step(`^the README links to the site's page for "([^"]*)"$`, w.theReadmeLinksToThePageFor)
	sc.Step(`^every link from the README to the site lands on a page of the site$`, w.everyReadmeLinkLandsOnAPage)
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

func (w *siteWorld) theReferenceListsEveryKey() error {
	page, err := w.read(keysPage)
	if err != nil {
		return err
	}
	for _, k := range files.Keys() {
		for _, said := range []string{k.Name, k.Says, takes(k)} {
			if !strings.Contains(page, html.EscapeString(said)) {
				return fmt.Errorf("the keys page does not say %q for %s", said, k.Name)
			}
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

var toTheSite = regexp.MustCompile(regexp.QuoteMeta(address) + `([^)#"\s]*)`)

func readme() (string, error) {
	said, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	return string(said), err
}

func (w *siteWorld) theReadmeLinksToThePageFor(source string) error {
	said, err := readme()
	if err != nil {
		return err
	}
	name, ok := pageFor(source)
	if !ok {
		return fmt.Errorf("%s is not published", source)
	}
	if !strings.Contains(said, "("+address+name) {
		return fmt.Errorf("the README has no link to %s%s", address, name)
	}
	return nil
}

func (w *siteWorld) everyReadmeLinkLandsOnAPage() error {
	said, err := readme()
	if err != nil {
		return err
	}
	links := toTheSite.FindAllStringSubmatch(said, -1)
	if len(links) == 0 {
		return fmt.Errorf("the README has no link to the site")
	}
	for _, link := range links {
		page := link[1]
		if page == "" {
			page = "index.html"
		}
		if _, err := os.Stat(filepath.Join(w.out, page)); err != nil {
			return fmt.Errorf("the README links to %s%s, which is not on the site", address, link[1])
		}
	}
	return nil
}
