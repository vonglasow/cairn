package engine

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"cairn/internal/theme"
)

const darkInk = `system: "base16"
name: "dim"
palette:
  base00: "#222222"
  base01: "#1f2024"
  base02: "#303238"
  base03: "#5f6067"
  base04: "#9a9ba2"
  base05: "#444444"
  base06: "#e7e7ea"
  base07: "#ffffff"
  base08: "#e5605e"
  base09: "#f0ba60"
  base0A: "#e5a24a"
  base0B: "#7fb69f"
  base0C: "#4fae8c"
  base0D: "#7b93ff"
  base0E: "#cba6f7"
  base0F: "#f2cdcd"
`

func homeWith(t *testing.T, content string, own map[string]string) Home {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(filepath.Join(dir, "themes"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	for slug, data := range own {
		if err := os.WriteFile(filepath.Join(dir, "themes", slug+".yaml"), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return Home{Dir: home, Lookup: func(string) (string, bool) { return "", false }}
}

func TestWearing_isTheDefaultWhenNothingNamesAScheme(t *testing.T) {
	worn, unworn := Wearing(homeWith(t, "", nil))
	if !reflect.DeepEqual(worn, theme.Default()) || unworn != "" {
		t.Errorf("with nothing said, the window wears other colours, or says %q", unworn)
	}
}

func TestWearing_isTheSchemeNamed(t *testing.T) {
	worn, unworn := Wearing(homeWith(t, "[window]\ntheme = \"nord\"\n", nil))
	nord, err := theme.Find("", "nord")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(worn, nord.Colours()) || unworn != "" {
		t.Errorf("named nord, the window wears other colours, or says %q", unworn)
	}
}

func TestWearing_aSchemeOfOnesOwnIsReadFromTheThemesFolder(t *testing.T) {
	worn, _ := Wearing(homeWith(t, "[window]\ntheme = \"nord\"\n", map[string]string{"nord": darkInk}))
	if worn.Surface.Back.R != 0x22 {
		t.Errorf("the window's back is %v, not the scheme of my own", worn.Surface.Back)
	}
}

func TestWearing_aSchemeThatCannotBeWornIsSaidAndTheDefaultWorn(t *testing.T) {
	for named, says := range map[string]string{"neon": `"neon" is not a scheme cairn sees`, "bad": `"#12" is not a colour`} {
		worn, unworn := Wearing(homeWith(t, "[window]\ntheme = \""+named+"\"\n", map[string]string{"bad": strings.Replace(darkInk, `"#222222"`, `"#12"`, 1)}))
		if !reflect.DeepEqual(worn, theme.Default()) {
			t.Errorf("%s: the window does not wear the default", named)
		}
		if !strings.Contains(unworn, says) || !strings.HasSuffix(unworn, "; the window wears default") {
			t.Errorf("%s: it says %q", named, unworn)
		}
	}
}

func TestWearing_aConfigurationThatCannotBeReadLeavesTheDefault(t *testing.T) {
	worn, unworn := Wearing(homeWith(t, "[window\n", nil))
	if !reflect.DeepEqual(worn, theme.Default()) || unworn != "" {
		t.Errorf("an unreadable configuration: %q", unworn)
	}
	if worn, _ := Wearing(Home{}); !reflect.DeepEqual(worn, theme.Default()) {
		t.Error("with no home, the window does not wear the default")
	}
}

func TestConfigShow_saysTheSchemeAndEverySchemeSeen(t *testing.T) {
	c := shownWithHome(t, homeWith(t, "[window]\ntheme = \"dim\"\n", map[string]string{"dim": darkInk}))
	if c.Window == nil || c.Window.Theme != "dim" || c.Window.Unworn != "" {
		t.Fatalf("Window = %+v", c.Window)
	}
	var dim, def ThemeSeen
	for _, seen := range c.Themes {
		switch seen.Slug {
		case "dim":
			dim = seen
		case "default":
			def = seen
		}
	}
	if !dim.InForce || !strings.HasSuffix(dim.From, filepath.Join("themes", "dim.yaml")) || len(dim.Short) == 0 {
		t.Errorf("dim is %+v", dim)
	}
	if def.InForce || def.From != theme.Carried || len(def.Short) != 0 {
		t.Errorf("default is %+v", def)
	}
	said := configText(c)
	for _, line := range []string{
		"\n[window]\nkeep_sizes = false\ntheme = \"dim\"\n",
		"\n# the schemes cairn sees, any of them a theme:\n",
		"\n#   default            shipped\n",
		"\n#   dim                " + dim.From + ", in force\n",
		"\n# dim falls short: ink.text on surface.back reads at ",
	} {
		if !strings.Contains(said, line) {
			t.Errorf("config show does not say %q:\n%s", line, said)
		}
	}
}

func TestConfigShow_saysWhyTheSchemeNamedIsNotWorn(t *testing.T) {
	c := shownWithHome(t, homeWith(t, "[window]\ntheme = \"bad\"\n", map[string]string{"bad": strings.Replace(darkInk, `"#222222"`, `"#12"`, 1)}))
	if !strings.Contains(c.Window.Unworn, `"#12" is not a colour`) {
		t.Errorf("Unworn = %q", c.Window.Unworn)
	}
	said := configText(c)
	for _, line := range []string{"\n# theme: bad (", "; the window wears default\n", "\n#   default            shipped, in force\n", "bad.yaml, cannot be worn: "} {
		if !strings.Contains(said, line) {
			t.Errorf("config show does not say %q:\n%s", line, said)
		}
	}
	if again := shownWith(t, configText(c)); again.Problem != "" || again.Window.Theme != "bad" {
		t.Errorf("read back: %+v (%s)", again.Window, again.Problem)
	}
}

func shownWithHome(t *testing.T, h Home) Config {
	t.Helper()
	answer, err := ConfigShow.Run(On{Home: h}, Input{})
	if err != nil || answer.Config == nil {
		t.Fatalf("config show: %v", err)
	}
	return *answer.Config
}

func TestSchemeOf_findsACarriedSchemeAndOneOfOnesOwnWhereEachIs(t *testing.T) {
	nord, err := os.ReadFile(filepath.Join("..", "theme", "schemes", "nord.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	own := t.TempDir()
	if err := os.WriteFile(filepath.Join(own, "mine.yaml"), nord, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := SchemeOf(ThemeSeen{Slug: "mine", From: filepath.Join(own, "mine.yaml")}); err != nil {
		t.Errorf("a scheme of one's own: %v", err)
	}
	here := t.TempDir()
	if err := os.WriteFile(filepath.Join(here, "nord.yaml"), []byte("not a scheme"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(here)
	if _, err := SchemeOf(ThemeSeen{Slug: "nord", From: theme.Carried}); err != nil {
		t.Errorf("a carried scheme was looked for beside the working folder: %v", err)
	}
}
