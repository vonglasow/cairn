package theme

import (
	"errors"
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"cairn/internal/domain"
)

const gruvbox = `system: "base16"
name: "Gruvbox dark"
author: "morhetz"
variant: "dark"
palette:
  base00: "#282828" # the window
  base01: "#3c3836"
  base02: "#504945"
  base03: "#665c54"
  base04: "#bdae93"
  base05: "#ebdbb2"
  base06: "#ebdbb2"
  base07: "#fbf1c7"
  base08: "#fb4934"
  base09: "#fe8019"
  base0A: "#fabd2f"
  base0B: "#b8bb26"
  base0C: "#8ec07c"
  base0D: "#83a598"
  base0E: "#d3869b"
  base0F: "#d65d0e"
`

func rgb(hex string) color.NRGBA {
	c, err := colourOf(hex)
	if err != nil {
		panic(err)
	}
	return c
}

func parsed(t *testing.T, data string) Scheme {
	t.Helper()
	s, err := Parse("mine", "test", []byte(data))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestABase16FileIsReadAsTintedThemingWritesIt(t *testing.T) {
	s := parsed(t, gruvbox)
	if s.Name != "Gruvbox dark" || s.Author != "morhetz" || s.Light() {
		t.Errorf("read %q by %q, light %t", s.Name, s.Author, s.Light())
	}
	c := s.Colours()
	if c.Surface.Back != rgb("#282828") || c.Ink.Text != rgb("#ebdbb2") || c.Accent.Base != rgb("#83a598") {
		t.Errorf("back %v, ink %v, accent %v", c.Surface.Back, c.Ink.Text, c.Accent.Base)
	}
}

func TestABase24FileIsReadAsBase16(t *testing.T) {
	s := parsed(t, strings.Replace(gruvbox, `system: "base16"`, `system: "base24"`, 1)+"  base10: \"#1d2021\"\n  base17: \"#fe8019\"\n")
	if s.Colours().Surface.Back != rgb("#282828") {
		t.Error("a base24 file is not read as the base16 it holds")
	}
}

func TestAColourThatIsNotOneIsSaidWithItsSlot(t *testing.T) {
	_, err := Parse("mine", "/somewhere/mine.yaml", []byte(strings.Replace(gruvbox, `"#282828"`, `"#12345"`, 1)))
	if !errors.Is(err, ErrNotAColour) {
		t.Fatalf("got %v", err)
	}
	for _, said := range []string{"mine", "/somewhere/mine.yaml", "palette.base00", `"#12345"`} {
		if !strings.Contains(err.Error(), said) {
			t.Errorf("%q does not say %s", err, said)
		}
	}
}

func TestASlotLeftOutIsSaid(t *testing.T) {
	_, err := Parse("mine", "test", []byte(strings.Replace(gruvbox, "  base0A: \"#fabd2f\"\n", "", 1)))
	if !errors.Is(err, ErrSlotMissing) || !strings.Contains(err.Error(), "base0A") {
		t.Fatalf("got %v", err)
	}
}

func TestAFileThatIsNotYAMLIsSaid(t *testing.T) {
	if _, err := Parse("mine", "test", []byte("palette: [unclosed")); !errors.Is(err, ErrNotAScheme) {
		t.Fatalf("got %v", err)
	}
}

func TestATokenIsOverriddenByItsName(t *testing.T) {
	s := parsed(t, gruvbox+"cairn:\n  tokens:\n    ink.faint: \"#aabbcc\"\n    surface.veil: \"#01020380\"\n    standing.on-hold.back: \"#010101\"\n    stones.2.dark: \"#020202\"\n    priority.low: \"#030303\"\n")
	c := s.Colours()
	if c.Ink.Faint != rgb("#aabbcc") || c.Surface.Veil != (color.NRGBA{R: 1, G: 2, B: 3, A: 0x80}) {
		t.Errorf("faint %v, veil %v", c.Ink.Faint, c.Surface.Veil)
	}
	if c.Standing["on-hold"].Back != rgb("#010101") || c.Stones[2].Dark != rgb("#020202") || c.Priority["low"] != rgb("#030303") {
		t.Errorf("standing %v, stone %v, priority %v", c.Standing["on-hold"], c.Stones[2], c.Priority["low"])
	}
}

func TestATokenCairnDoesNotKnowIsSaid(t *testing.T) {
	_, err := Parse("mine", "test", []byte(gruvbox+"cairn:\n  tokens:\n    ink.shiny: \"#aabbcc\"\n"))
	if !errors.Is(err, ErrNoSuchToken) || !strings.Contains(err.Error(), "ink.shiny") {
		t.Fatalf("got %v", err)
	}
}

func TestALabelIsOverriddenByItsName(t *testing.T) {
	s := parsed(t, gruvbox+"cairn:\n  labels:\n    mauve: \"#cba6f7\"\n")
	if got := s.Colours().Labels["mauve"]; got != rgb("#cba6f7") {
		t.Errorf("mauve is %v", got)
	}
	if _, err := Parse("mine", "test", []byte(gruvbox+"cairn:\n  labels:\n    beige: \"#cba6f7\"\n")); !errors.Is(err, ErrNoSuchToken) {
		t.Errorf("a label colour cairn does not know: %v", err)
	}
}

func TestEveryTokenGetsAValue(t *testing.T) {
	c := parsed(t, gruvbox).Colours()
	names := 0
	c.each(func(name string, at *color.NRGBA) {
		names++
		if at.A == 0 {
			t.Errorf("%s has no value", name)
		}
	})
	if want := 36 + len(domain.Priorities) + 4*2 + 3*2 + len(domain.LabelColours); names != want {
		t.Errorf("%d tokens, want %d", names, want)
	}
}

func TestTheEightAccentsAreLabelsStraight(t *testing.T) {
	c := parsed(t, gruvbox).Colours()
	for name, slot := range map[string]string{"red": "#fb4934", "peach": "#fe8019", "yellow": "#fabd2f", "green": "#b8bb26", "teal": "#8ec07c", "blue": "#83a598", "mauve": "#d3869b", "flamingo": "#d65d0e"} {
		if c.Labels[name] != rgb(slot) {
			t.Errorf("%s is %v, want %s", name, c.Labels[name], slot)
		}
	}
}

func TestALightSchemePullsItsLabelsTowardItsInk(t *testing.T) {
	light := strings.NewReplacer(`variant: "dark"`, `variant: "light"`, `"#282828"`, `"#fbf1c7"`, `base05: "#ebdbb2"`, `base05: "#3c3836"`).Replace(gruvbox)
	c := parsed(t, light).Colours()
	if want := mix(rgb("#fabd2f"), rgb("#3c3836"), 0.45); c.Labels["yellow"] != want {
		t.Errorf("yellow is %v, want %v", c.Labels["yellow"], want)
	}
}

func TestAVariantNotSaidIsReadOffTheGround(t *testing.T) {
	light := strings.NewReplacer("variant: \"dark\"\n", "", `"#282828"`, `"#fbf1c7"`, `base05: "#ebdbb2"`, `base05: "#3c3836"`).Replace(gruvbox)
	if !parsed(t, light).Light() {
		t.Error("a pale ground under a dark ink is not read as light")
	}
	if parsed(t, strings.Replace(gruvbox, "variant: \"dark\"\n", "", 1)).Light() {
		t.Error("a dark ground under a pale ink is read as light")
	}
}

func TestTheNineSchemesAreShipped(t *testing.T) {
	want := []string{"catppuccin-latte", "catppuccin-mocha", "default", "dracula", "gruvbox-dark", "gruvbox-light", "nord", "solarized-dark", "solarized-light"}
	if got := Shipped(); !slices.Equal(got, want) {
		t.Errorf("shipped %v, want %v", got, want)
	}
}

func TestEveryShippedSchemeIsReadAndCanBeRead(t *testing.T) {
	for _, slug := range Shipped() {
		s, err := Find("", slug)
		if err != nil {
			t.Fatalf("%s: %v", slug, err)
		}
		if s.From != Carried {
			t.Errorf("%s comes from %q", slug, s.From)
		}
		for _, short := range Shortfalls(s.Colours()) {
			t.Errorf("%s falls short: %s", slug, short)
		}
	}
}

func TestTheDefaultIsTheSchemeCalledDefault(t *testing.T) {
	s, err := Find("", "default")
	if err != nil {
		t.Fatal(err)
	}
	if !equal(Default(), s.Colours()) {
		t.Error("the default colours are not the default scheme's")
	}
}

func equal(a, b Colours) bool {
	same := true
	names := map[string]color.NRGBA{}
	a.each(func(name string, at *color.NRGBA) { names[name] = *at })
	b.each(func(name string, at *color.NRGBA) { same = same && names[name] == *at })
	return same
}

func writeOwn(t *testing.T, folder, slug, data string) string {
	t.Helper()
	at := filepath.Join(folder, slug+".yaml")
	if err := os.WriteFile(at, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	return at
}

func TestASchemeOfOnesOwnWinsOverAShippedOne(t *testing.T) {
	folder := t.TempDir()
	at := writeOwn(t, folder, "nord", gruvbox)
	s, err := Find(folder, "nord")
	if err != nil {
		t.Fatal(err)
	}
	if s.From != at || s.Colours().Surface.Back != rgb("#282828") {
		t.Errorf("from %q, back %v", s.From, s.Colours().Surface.Back)
	}
}

func TestASchemeNobodyHasIsSaid(t *testing.T) {
	for _, slug := range []string{"neon", "../default", "Nord", ""} {
		if _, err := Find(t.TempDir(), slug); !errors.Is(err, ErrUnknown) || !strings.Contains(err.Error(), fmt.Sprintf("%q", slug)) {
			t.Errorf("%q: %v", slug, err)
		}
	}
}

func TestAFileThatCannotBeReadIsSaid(t *testing.T) {
	folder := t.TempDir()
	if err := os.Mkdir(filepath.Join(folder, "odd.yaml"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := Find(folder, "odd"); err == nil || errors.Is(err, ErrUnknown) {
		t.Errorf("a file that cannot be read: %v", err)
	}
}

func TestSeenListsTheShippedAndOnesOwnOnce(t *testing.T) {
	folder := t.TempDir()
	writeOwn(t, folder, "nord", gruvbox)
	writeOwn(t, folder, "mine", strings.Replace(gruvbox, `"#282828"`, `"#nothex"`, 1))
	writeOwn(t, folder, "Not A Slug", gruvbox)
	if err := os.WriteFile(filepath.Join(folder, "notes.txt"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	seen := Seen(folder)
	var slugs []string
	for _, one := range seen {
		slugs = append(slugs, one.Slug)
		switch one.Slug {
		case "nord":
			if one.From != filepath.Join(folder, "nord.yaml") {
				t.Errorf("nord comes from %q", one.From)
			}
		case "mine":
			if !errors.Is(one.Err, ErrNotAColour) {
				t.Errorf("mine: %v", one.Err)
			}
		case "default":
			if one.From != Carried || one.Err != nil {
				t.Errorf("default: %q, %v", one.From, one.Err)
			}
		}
	}
	want := []string{"catppuccin-latte", "catppuccin-mocha", "default", "dracula", "gruvbox-dark", "gruvbox-light", "mine", "nord", "solarized-dark", "solarized-light"}
	if !slices.Equal(slugs, want) {
		t.Errorf("seen %v, want %v", slugs, want)
	}
	if got := Seen(filepath.Join(folder, "none")); len(got) != len(Shipped()) {
		t.Errorf("with no folder, %d seen", len(got))
	}
}

func TestAShortfallNamesTheTokensAndTheRatio(t *testing.T) {
	c := parsed(t, strings.Replace(gruvbox, `base05: "#ebdbb2"`, `base05: "#444444"`, 1)).Colours()
	shorts := Shortfalls(c)
	if len(shorts) == 0 {
		t.Fatal("dark ink on a dark ground falls short of nothing")
	}
	said := fmt.Sprint(shorts)
	want := fmt.Sprintf("ink.text on surface.back reads at %.2f, and needs 4.5", Contrast(rgb("#444444"), rgb("#282828")))
	if !strings.Contains(said, want) {
		t.Errorf("%s does not say %q", said, want)
	}
}

func TestAPillHardToReadIsAShortfall(t *testing.T) {
	c := parsed(t, gruvbox+"cairn:\n  labels:\n    yellow: \"#3a3a3a\"\n").Colours()
	if said := fmt.Sprint(Shortfalls(c)); !strings.Contains(said, "label.yellow on its pill") {
		t.Errorf("%s says nothing of the yellow pill", said)
	}
}

func TestContrastAnswersTheKnownRatios(t *testing.T) {
	black := color.NRGBA{A: 255}
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	for _, c := range []struct {
		a, b color.NRGBA
		want float64
	}{{white, black, 21}, {black, white, 21}, {black, black, 1}, {white, white, 1}} {
		if got := Contrast(c.a, c.b); math.Abs(got-c.want) > 0.01 {
			t.Errorf("%v on %v is %.2f, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestTheCurveIsStraightenedWhereItChangesShape(t *testing.T) {
	for _, c := range []struct {
		v    uint8
		want float64
	}{{0, 0}, {9, 0.0027317429}, {10, 0.0030352698}, {11, 0.0033465358}, {255, 1}} {
		if got := straight(c.v); math.Abs(got-c.want) > 0.000000001 {
			t.Errorf("straight(%d) = %.10f, want %.10f", c.v, got, c.want)
		}
	}
}

func TestMixMovesAFractionOfTheWay(t *testing.T) {
	if got := mix(rgb("#000000"), rgb("#ffffff"), 0.5); got != rgb("#808080") {
		t.Errorf("half way from black to white is %v", got)
	}
	if got := mix(rgb("#102030"), rgb("#ffffff"), 0); got != rgb("#102030") {
		t.Errorf("no way at all is %v", got)
	}
}

func TestAColourIsWrittenWithOrWithoutItsHashAndItsAlpha(t *testing.T) {
	for said, want := range map[string]color.NRGBA{"#0c0d0f": {R: 12, G: 13, B: 15, A: 255}, "0c0d0f": {R: 12, G: 13, B: 15, A: 255}, "#0607099e": {R: 6, G: 7, B: 9, A: 158}, "#ABCDEF": {R: 0xab, G: 0xcd, B: 0xef, A: 255}} {
		if got, err := colourOf(said); err != nil || got != want {
			t.Errorf("%q is %v (%v), want %v", said, got, err, want)
		}
	}
	for _, said := range []string{"#12345", "#gggggg", "", "#1234567", "red"} {
		if _, err := colourOf(said); !errors.Is(err, ErrNotAColour) {
			t.Errorf("%q: %v", said, err)
		}
	}
}
