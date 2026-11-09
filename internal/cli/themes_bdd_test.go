package cli

import (
	"context"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"cairn/internal/debug"
	"cairn/internal/engine"
	"cairn/internal/gui"
)

func initThemes(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := &workspaces{world: newWorld(t)}
		w.register(sc)
		w.registerConfig(sc)
		w.registerWorkspace(sc)
		w.registerSettings(sc)
		w.registerThemes(sc)
	}
}

func TestAcceptance_Themes(t *testing.T) {
	runAcceptance(t, "themes.feature", initThemes(t), "~@later")
}

const ownScheme = `system: "base16"
name: %q
author: "me"
variant: "dark"
palette:
  base00: %q
  base01: "#1f2024"
  base02: "#303238"
  base03: "#5f6067"
  base04: "#9a9ba2"
  base05: %q
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

func hexOf(c color.NRGBA) string { return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B) }

func (w *workspaces) registerThemes(sc *godog.ScenarioContext) {
	var worn *gui.Skin
	was := openGui
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		worn = nil
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		openGui = was
		return ctx, err
	})
	opening := func(line string) error {
		openGui = func(_ engine.Work, _ engine.Pages, look gui.Skin, _, _ string, _ *debug.Log, _, _ int, _ gui.Here) error {
			worn = &look
			return nil
		}
		w.Run(line)
		if w.Out.Code != exitOK || worn == nil {
			return fmt.Errorf("the window did not open: exit %d (%s)", w.Out.Code, w.Out.Stderr)
		}
		return nil
	}
	ownWith := func(slug, back, ink string) error {
		folder := filepath.Join(w.Defaults, ".config", "cairn", "themes")
		if err := os.MkdirAll(folder, 0o700); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(folder, slug+".yaml"), fmt.Appendf(nil, ownScheme, slug, back, ink), 0o600)
	}
	sc.Step(`^I open the window$`, func() error { return opening("cairn gui") })
	sc.Step(`^I open the window on the workspace "([^"]*)"$`, func(name string) error {
		return opening("cairn gui --workspace " + name)
	})
	sc.Step(`^the window wears the colours it always wore$`, func() error {
		if !reflect.DeepEqual(worn.Colour, gui.Dark().Colour) {
			return fmt.Errorf("the window wears other colours: its back is %s", hexOf(worn.Colour.Surface.Back))
		}
		return nil
	})
	sc.Step(`^the window's back is "([^"]*)"$`, func(want string) error {
		if got := hexOf(worn.Colour.Surface.Back); got != want {
			return fmt.Errorf("the window's back is %s, want %s", got, want)
		}
		return nil
	})
	sc.Step(`^the window's ink is "([^"]*)"$`, func(want string) error {
		if got := hexOf(worn.Colour.Ink.Text); got != want {
			return fmt.Errorf("the window's ink is %s, want %s", got, want)
		}
		return nil
	})
	sc.Step(`^a scheme of my own "([^"]*)" whose back is "([^"]*)"$`, func(slug, back string) error {
		return ownWith(slug, back, "#e7e7ea")
	})
	sc.Step(`^a scheme of my own "([^"]*)" whose back is "([^"]*)" and ink "([^"]*)"$`, ownWith)
	sc.Step(`^it says on standard error that "([^"]*)" is not a scheme it sees$`, func(slug string) error {
		return wantAll(w.Out.Stderr, "standard error", fmt.Sprintf("%q is not a scheme cairn sees", slug), "the window wears default")
	})
	sc.Step(`^it says on standard error that "([^"]*)" in "([^"]*)" is not a colour$`, func(said, slug string) error {
		return wantAll(w.Out.Stderr, "standard error", slug, fmt.Sprintf("%q is not a colour", said), "the window wears default")
	})
	sc.Step(`^it lists the schemes (.*) as shipped$`, func(listed string) error {
		for _, slug := range regexp.MustCompile(`"([^"]*)"`).FindAllStringSubmatch(listed, -1) {
			if !regexp.MustCompile(`(?m)^#\s+` + regexp.QuoteMeta(slug[1]) + `\s+shipped\b`).MatchString(w.Out.Stdout) {
				return fmt.Errorf("the answer does not list %q as shipped:\n%s", slug[1], w.Out.Stdout)
			}
		}
		return nil
	})
	sc.Step(`^it lists the scheme "([^"]*)" from my themes folder$`, func(slug string) error {
		at := filepath.Join(w.Defaults, ".config", "cairn", "themes", slug+".yaml")
		if !regexp.MustCompile(`(?m)^#\s+` + regexp.QuoteMeta(slug) + `\s+` + regexp.QuoteMeta(at)).MatchString(w.Out.Stdout) {
			return fmt.Errorf("the answer does not list %q from %s:\n%s", slug, at, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^it says "([^"]*)" is the one in force$`, func(slug string) error {
		if !regexp.MustCompile(`(?m)^#\s+` + regexp.QuoteMeta(slug) + `\s+.*, in force$`).MatchString(w.Out.Stdout) {
			return fmt.Errorf("the answer does not say %q is in force:\n%s", slug, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^it says "([^"]*)" is not a scheme it sees, and the window wears "([^"]*)"$`, func(slug, worn string) error {
		return wantAll(w.Out.Stdout, "the answer", fmt.Sprintf("# theme: %q is not a scheme cairn sees", slug), "the window wears "+worn)
	})
	sc.Step(`^it warns that "([^"]*)" falls short, naming "([^"]*)" on "([^"]*)" and the ratio it reads at$`, func(slug, said, on string) error {
		line := regexp.MustCompile(`(?m)^# ` + regexp.QuoteMeta(slug) + ` falls short: ` + regexp.QuoteMeta(said) + ` on ` + regexp.QuoteMeta(on) + ` reads at \d+\.\d\d, and needs \d`)
		if !line.MatchString(w.Out.Stdout) {
			return fmt.Errorf("the answer does not warn that %s falls short on %s on %s:\n%s", slug, said, on, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^it warns of no scheme falling short$`, func() error {
		if w.Out.Code != exitOK || strings.Contains(w.Out.Stdout, "falls short") {
			return fmt.Errorf("exit %d, and the answer warns:\n%s%s", w.Out.Code, w.Out.Stdout, w.Out.Stderr)
		}
		return nil
	})
}

func wantAll(where, label string, said ...string) error {
	for _, one := range said {
		if !strings.Contains(where, one) {
			return fmt.Errorf("%s does not say %q:\n%s", label, one, where)
		}
	}
	return nil
}
