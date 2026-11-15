package tui

import (
	"cairn/internal/theme"
	"fmt"
	"image/color"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

var graphic = regexp.MustCompile("\x1b\\[([0-9;:]*)m")

var painted = regexp.MustCompile(`(?:^|;)(?:38|48|58);2;(\d+);(\d+);(\d+)|(?:^|;)(3[1-6]|4[0-7]|9[0-7]|10[0-7]|38;5|48;5)(?:;|$)`)

func (w *world) registerColours(sc *godog.ScenarioContext) {
	sc.Then(`^no screen wears a colour the scheme does not have$`, w.noScreenWearsAColourOfItsOwn)
}

func schemeColours(c theme.Colours) map[string]bool {
	seen := map[string]bool{}
	add := func(c color.NRGBA) { seen[fmt.Sprintf("%d;%d;%d", c.R, c.G, c.B)] = true }
	for _, c := range []color.NRGBA{c.Ink.Text, c.Ink.Faint, c.Ink.Low, c.Line.Rule, c.Line.Edge, c.Line.Raised, c.Line.Selected,
		c.Accent.Base, c.State.Late, c.State.Done, c.Surface.Back} {
		add(c)
	}
	for _, c := range c.Priority {
		add(c)
	}
	for _, c := range c.Standing {
		add(c.Ink)
	}
	for _, c := range c.Labels {
		add(c)
	}
	return seen
}

func outsideTheScheme(drawn string, c theme.Colours) []string {
	known := schemeColours(c)
	var found []string
	for _, match := range graphic.FindAllStringSubmatch(drawn, -1) {
		for _, p := range painted.FindAllStringSubmatch(match[1], -1) {
			if p[4] != "" || !known[p[1]+";"+p[2]+";"+p[3]] {
				found = append(found, match[0])
			}
		}
	}
	return found
}

func (w *world) noScreenWearsAColourOfItsOwn() error {
	var problems []string
	look := func(name string) {
		if found := outsideTheScheme(w.b.Draw(), theme.Default()); len(found) > 0 {
			problems = append(problems, fmt.Sprintf("%s wears %q", name, found))
		}
	}
	w.everyScreen(look)
	if len(problems) > 0 {
		return fmt.Errorf("%d screen(s) wear a colour the scheme does not have:\n%s", len(problems), strings.Join(problems, "\n"))
	}
	return nil
}
