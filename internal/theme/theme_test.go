package theme

import (
	"errors"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
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

func writeOwn(t *testing.T, folder, slug, data string) string {
	t.Helper()
	at := filepath.Join(folder, slug+".yaml")
	if err := os.WriteFile(at, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	return at
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
