package site

import (
	"errors"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/theme"
)

func TestTerminalHTML_writesPlainTextEscaped(t *testing.T) {
	if got := terminalHTML("a <b> & c\n"); got != "a &lt;b&gt; &amp; c\n" {
		t.Errorf("terminalHTML = %q", got)
	}
}

func TestTerminalHTML_colouresWhatTheEscapesColour(t *testing.T) {
	for said, want := range map[string]string{
		"\x1b[34mx\x1b[m":          `<span class="t4">x</span>`,
		"\x1b[1;34mx\x1b[0m":       `<span class="t4 b">x</span>`,
		"\x1b[2mx\x1b[m":           `<span class="d">x</span>`,
		"\x1b[93mx\x1b[39m":        `<span class="t11">x</span>`,
		"\x1b[41mx\x1b[49m":        `<span class="u1">x</span>`,
		"\x1b[101mx\x1b[m":         `<span class="u9">x</span>`,
		"\x1b[38;5;12mx\x1b[m":     `<span class="t12">x</span>`,
		"\x1b[48;5;3mx\x1b[m":      `<span class="u3">x</span>`,
		"\x1b[7mx\x1b[m":           `<span class="tz uz">x</span>`,
		"\x1b[7;34mx\x1b[27mx":     `<span class="tz u4">x</span><span class="t4">x</span>`,
		"\x1b[1mx\x1b[22mx":        `<span class="b">x</span>x`,
		"\x1b[31ma\x1b[32mb\x1b[m": `<span class="t1">a</span><span class="t2">b</span>`,
		"\x1b[38;2;1;2;3mx\x1b[m":  `x`,
	} {
		if got := terminalHTML(said); got != want {
			t.Errorf("terminalHTML(%q) = %q, want %q", said, got, want)
		}
	}
}

func TestTerminal16_isTheSlotsATerminalTakesFromABase16Scheme(t *testing.T) {
	scheme, err := theme.Find("", "gruvbox-dark")
	if err != nil {
		t.Fatal(err)
	}
	slots, got := scheme.Slots(), terminal16(scheme)
	want := [16]color.NRGBA{
		slots[0x0], slots[0x8], slots[0xB], slots[0xA], slots[0xD], slots[0xE], slots[0xC], slots[0x5],
		slots[0x3], slots[0x8], slots[0xB], slots[0xA], slots[0xD], slots[0xE], slots[0xC], slots[0x7],
	}
	if got != want {
		t.Errorf("terminal16 = %v, want %v", got, want)
	}
}

func TestHeadingID_isTheAnchorGitHubGivesAHeading(t *testing.T) {
	for heading, want := range map[string]string{
		"One configuration, several ways in to it": "one-configuration-several-ways-in-to-it",
		"`[pages]` and `[archive]`":                "pages-and-archive",
		"What a task's notes are":                  "what-a-tasks-notes-are",
		"Hand-written, not rrule-go":               "hand-written-not-rrule-go",
		"Évolutions — v0.12":                       "évolutions--v012",
		"snake_case stays":                         "snake_case-stays",
	} {
		if got := headingID(heading); got != want {
			t.Errorf("headingID(%q) = %q, want %q", heading, got, want)
		}
	}
}

func TestLinked_takesALinkToWhereItLandsOnTheSite(t *testing.T) {
	for _, c := range []struct{ from, href, want string }{
		{"docs/tui.md", "gui.md", "gui.html"},
		{"docs/tui.md", "gui.md#the-design-system", "gui.html#the-design-system"},
		{"docs/tui.md", "#keys", "#keys"},
		{"docs/tui.md", "architecture.md#layers", "https://github.com/vonglasow/cairn/blob/v0.12.0/docs/architecture.md#layers"},
		{"docs/tui.md", "../INSTALL.md", "install.html"},
		{"docs/trace.md", "../bdd/trace.feature", "https://github.com/vonglasow/cairn/blob/v0.12.0/bdd/trace.feature"},
		{"INSTALL.md", "docs/claude-code.md", "claude-code.html"},
		{"docs/tui.md", "https://example.org/x.md", "https://example.org/x.md"},
		{"docs/tui.md", "mailto:a@b", "mailto:a@b"},
	} {
		if got := linked(c.from, c.href, "v0.12.0"); got != c.want {
			t.Errorf("linked(%q, %q) = %q, want %q", c.from, c.href, got, c.want)
		}
	}
}

func TestLinked_aBuildThatIsNoVersionLinksToMain(t *testing.T) {
	if got := linked("docs/tui.md", "architecture.md", "v0.12.0-3-gabc-dirty"); got != "https://github.com/vonglasow/cairn/blob/main/docs/architecture.md" {
		t.Errorf("linked = %q", got)
	}
}

func TestPageFor_namesThePageOfEverySourcePublished(t *testing.T) {
	for source, want := range map[string]string{"INSTALL.md": "install.html", "docs/usage.md": "usage.html", "docs/config.md": "config.html"} {
		if got, ok := pageFor(source); !ok || got != want {
			t.Errorf("pageFor(%q) = %q, %v; want %q", source, got, ok, want)
		}
	}
	if got, ok := pageFor(".claude/docs/checks.md"); ok {
		t.Errorf("a page for whoever develops cairn is published as %s", got)
	}
}

func TestSchemeRules_givesEveryTokenTheSiteDrawsWith(t *testing.T) {
	scheme, err := theme.Find("", "catppuccin-latte")
	if err != nil {
		t.Fatal(err)
	}
	rules := schemeRules(scheme)
	for _, want := range []string{`[data-scheme="catppuccin-latte"]{`, "--bg:#eff1f5", "color-scheme:light", "--t0:#eff1f5"} {
		if !strings.Contains(rules, want) {
			t.Errorf("schemeRules = %s\nwant it to hold %s", rules, want)
		}
	}
	if dark, _ := theme.Find("", "default"); !strings.Contains(schemeRules(dark), "color-scheme:dark") {
		t.Error("the default scheme is not said to be dark")
	}
}

func TestMain_refusesWhatItIsNotGiven(t *testing.T) {
	var said strings.Builder
	if code := Main([]string{"--nonsense"}, &said, &said); code == 0 {
		t.Errorf("Main ran with a flag it does not know: %s", said.String())
	}
	said.Reset()
	if code := Main([]string{}, &said, &said); code == 0 || !strings.Contains(said.String(), "--out") {
		t.Errorf("Main ran without being told where to write: %s", said.String())
	}
}

func TestBuild_drawsAPictureOfTheWindowInEveryScheme(t *testing.T) {
	out := t.TempDir()
	drawn := map[string]bool{}
	said := Said{Version: "v0.12.0", Repo: os.DirFS(filepath.Join("..", "..")), Window: func(slug string) (image.Image, error) {
		drawn[slug] = true
		return image.NewRGBA(image.Rect(0, 0, 2, 2)), nil
	}}
	if err := Build(out, said); err != nil {
		t.Fatal(err)
	}
	for _, slug := range theme.Shipped() {
		if !drawn[slug] {
			t.Errorf("the window was not drawn in %s", slug)
		}
		if _, err := os.Stat(filepath.Join(out, windowPicture(slug))); err != nil {
			t.Errorf("no picture of the window in %s: %v", slug, err)
		}
	}
	index, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(index), windowPicture(darkFirst)) {
		t.Error("the home page shows no picture of the window")
	}
}

func TestBuild_saysWhichSchemeTheWindowCouldNotBeDrawnIn(t *testing.T) {
	broken := errors.New("no GPU")
	said := Said{Version: "v0.12.0", Repo: os.DirFS(filepath.Join("..", "..")), Window: func(string) (image.Image, error) { return nil, broken }}
	if err := Build(t.TempDir(), said); !errors.Is(err, broken) || !strings.Contains(err.Error(), "the window in") {
		t.Errorf("Build = %v, want the scheme the window could not be drawn in", err)
	}
}

func TestMain_refusesAChangelogItCannotRead(t *testing.T) {
	var said strings.Builder
	missing := filepath.Join(t.TempDir(), "CHANGELOG.md")
	if code := Main([]string{"--out", t.TempDir(), "--changelog", missing}, &said, &said); code != 1 || !strings.Contains(said.String(), missing) {
		t.Errorf("Main = %d, %q; want 1, naming the changelog", code, said.String())
	}
}

func TestIDs_putCountsAnIDAlreadyGiven(t *testing.T) {
	seen := ids{}
	seen.Put([]byte("keys"))
	if got := string(seen.Generate([]byte("Keys"), 0)); got != "keys-1" {
		t.Errorf("Generate after Put = %q, want keys-1", got)
	}
}
