package tui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

var terminalColours = map[string]string{
	"red": "31", "green": "32", "yellow": "33", "blue": "34", "magenta": "35", "cyan": "36",
	"bright red": "91", "bright green": "92", "bright yellow": "93",
	"bright blue": "94", "bright magenta": "95", "bright cyan": "96",
}

var labelsInTheTerminal = map[string]string{
	"rosewater": "bright magenta", "flamingo": "bright red", "pink": "magenta", "mauve": "bright magenta",
	"red": "red", "maroon": "bright red", "peach": "bright yellow", "yellow": "yellow",
	"green": "green", "teal": "cyan", "sky": "bright cyan", "sapphire": "bright blue",
	"blue": "blue", "lavender": "bright blue",
}

var graphic = regexp.MustCompile("\x1b\\[([0-9;:]*)m")

func (w *world) registerColours(sc *godog.ScenarioContext) {
	sc.Then(`^no screen wears a colour the terminal's theme does not set$`, w.noScreenWearsAColourOfItsOwn)
	sc.Then(`^the rule between two columns is faint, in no colour of its own$`, w.ruleIsFaint)
	sc.Then(`^the left edge of the card for task (\d+) wears the terminal's (.+)$`, w.leftEdgeWears)
	sc.Then(`^the tab "([^"]*)" is drawn reversed in the terminal's (.+)$`, w.tabReversed)
	sc.Then(`^the dot of the header wears the terminal's (.+)$`, w.dotWearsTheTerminals)
}

func outsideTheTheme(drawn string) []string {
	var found []string
	for _, match := range graphic.FindAllStringSubmatch(drawn, -1) {
		for _, param := range strings.FieldsFunc(match[1], func(r rune) bool { return r == ';' || r == ':' }) {
			n, err := strconv.Atoi(param)
			if err == nil && paintsOfItsOwn(n) {
				found = append(found, match[0])
				break
			}
		}
	}
	return found
}

func paintsOfItsOwn(n int) bool {
	return n == 38 || n == 48 || n == 58 || n == 90 || n == 100
}

func (w *world) noScreenWearsAColourOfItsOwn() error {
	var problems []string
	look := func(name string) {
		if found := outsideTheTheme(w.b.Draw()); len(found) > 0 {
			problems = append(problems, fmt.Sprintf("%s wears %q", name, found))
		}
	}
	w.everyScreen(look)
	if len(problems) > 0 {
		return fmt.Errorf("%d screen(s) wear a colour of their own:\n%s", len(problems), strings.Join(problems, "\n"))
	}
	return nil
}

func (w *world) ruleIsFaint() error {
	for _, line := range strings.Split(w.b.Draw(), "\n") {
		if strings.Contains(strip(line), "BACKLOG") {
			continue
		}
		if strings.Contains(line, "\x1b[2m│") {
			return nil
		}
	}
	return fmt.Errorf("no rule is drawn faint and bare:\n%q", w.b.Draw())
}

func (w *world) leftEdgeWears(number int, colour string) error {
	title := ""
	for _, c := range w.cards() {
		if c.Number == number {
			title = c.Title
		}
	}
	if title == "" {
		return fmt.Errorf("there is no task %d", number)
	}
	for _, line := range strings.Split(w.b.Draw(), "\n") {
		if !strings.Contains(strip(line), title) {
			continue
		}
		edge := strings.Index(line, "┃")
		if edge < 0 {
			return fmt.Errorf("the card for task %d has no left edge: %q", number, line)
		}
		before := line[:edge]
		if colour == "own foreground" {
			if last := graphic.FindAllStringSubmatchIndex(before, -1); len(last) > 0 && last[len(last)-1][1] == len(before) && last[len(last)-1][3] > last[len(last)-1][2] {
				return fmt.Errorf("the left edge of task %d wears a colour: %q", number, line)
			}
			return nil
		}
		code, ok := terminalColours[colour]
		if !ok {
			return fmt.Errorf("%q is none of the terminal's colours", colour)
		}
		if !strings.HasSuffix(before, "\x1b["+code+"m") {
			return fmt.Errorf("the left edge of task %d does not wear the terminal's %s (%s): %q", number, colour, code, line)
		}
		return nil
	}
	return fmt.Errorf("task %d is not on the board:\n%s", number, w.screen())
}

func (w *world) tabReversed(tab, colour string) error {
	code, ok := terminalColours[colour]
	if !ok {
		return fmt.Errorf("%q is none of the terminal's colours", colour)
	}
	header := strings.SplitN(w.b.Draw(), "\n", 2)[0]
	if want := "\x1b[1;7;" + code + "m" + tab; !strings.Contains(header, want) {
		return fmt.Errorf("the tab %q is not drawn reversed in the terminal's %s (%q):\n%q", tab, colour, want, header)
	}
	return nil
}

func (w *world) dotWearsTheTerminals(colour string) error {
	code, ok := terminalColours[colour]
	if !ok {
		return fmt.Errorf("%q is none of the terminal's colours", colour)
	}
	worn := "\x1b[" + code + "m● "
	if header := strings.SplitN(w.b.Draw(), "\n", 2)[0]; !strings.Contains(header, worn) {
		return fmt.Errorf("the dot of the header does not wear the terminal's %s (%q):\n%q", colour, worn, header)
	}
	return nil
}
