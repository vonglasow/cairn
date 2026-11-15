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
	sc.Then(`^the rule between two columns wears the scheme's line$`, w.ruleWearsTheLine)
	sc.Then(`^the left edge of the card for task (\d+) wears the scheme's (\w+) priority$`, w.leftEdgeWears)
	sc.Then(`^the tab "([^"]*)" is drawn reversed in the accent of "([^"]*)"$`, w.tabReversed)
	sc.Then(`^the dot of the header wears the scheme's (\w+)$`, w.dotWearsTheSchemes)
}

func ink(c color.NRGBA) string { return fmt.Sprintf("38;2;%d;%d;%d", c.R, c.G, c.B) }

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

func (w *world) ruleWearsTheLine() error {
	want := ink(theme.Default().Line.Edge) + "m│"
	for _, line := range strings.Split(w.b.Draw(), "\n") {
		if strings.Contains(line, want) {
			return nil
		}
	}
	return fmt.Errorf("no rule wears the scheme's line (%q):\n%q", want, w.b.Draw())
}

func (w *world) leftEdgeWears(number int, priority string) error {
	title := ""
	for _, c := range w.cards() {
		if c.Number == number {
			title = c.Title
		}
	}
	if title == "" {
		return fmt.Errorf("there is no task %d", number)
	}
	want := ink(theme.Default().Priority[priority]) + "m"
	for _, line := range strings.Split(w.b.Draw(), "\n") {
		if !strings.Contains(strip(line), title) {
			continue
		}
		edge := strings.Index(line, "┃")
		if edge < 0 {
			return fmt.Errorf("the card for task %d has no left edge: %q", number, line)
		}
		if !strings.HasSuffix(line[:edge], want) {
			return fmt.Errorf("the left edge of task %d does not wear the scheme's %s priority (%s): %q", number, priority, want, line)
		}
		return nil
	}
	return fmt.Errorf("task %d is not on the board:\n%s", number, w.screen())
}

func (w *world) tabReversed(tab, scheme string) error {
	s, err := theme.Find("", scheme)
	if err != nil {
		return err
	}
	header := strings.SplitN(w.b.Draw(), "\n", 2)[0]
	if want := "\x1b[1;7;" + ink(s.Colours().Accent.Base) + "m" + tab; !strings.Contains(header, want) {
		return fmt.Errorf("the tab %q is not drawn reversed in the accent of %s (%q):\n%q", tab, scheme, want, header)
	}
	return nil
}

func (w *world) dotWearsTheSchemes(colour string) error {
	worn := "\x1b[" + ink(theme.Default().Labels[colour]) + "m● "
	if header := strings.SplitN(w.b.Draw(), "\n", 2)[0]; !strings.Contains(header, worn) {
		return fmt.Errorf("the dot of the header does not wear the scheme's %s (%q):\n%q", colour, worn, header)
	}
	return nil
}
