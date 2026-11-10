package gui

import (
	"fmt"
	"slices"
	"strings"

	"cairn/internal/engine"

	"github.com/cucumber/godog"
)

func (w *world) branchOf(title string) (engine.Branch, Box, error) {
	for i, b := range w.w.Tree() {
		if b.Goal.Title == title {
			return b, GoalAt(w.w.Layout(), i, w.w.goals), nil
		}
	}
	return engine.Branch{}, Box{}, fmt.Errorf("no objective titled %q is listed", title)
}

func (w *world) saidWithin(box Box) []string {
	var said []string
	for _, one := range textDrawn(w.w) {
		if box.In(Point{X: one.box.X + one.box.W/2, Y: one.box.Y + one.box.H/2}) {
			said = append(said, one.said)
		}
	}
	return said
}

func (w *world) chevronDrawnAt(slot Box) bool {
	gutter := FoldAt(slot)
	for _, d := range partsDrawn(w.w) {
		if strings.HasSuffix(d.Part, "chevron") && d.At.X >= gutter.X && d.At.X+d.At.W <= gutter.X+gutter.W && d.At.Y >= gutter.Y && d.At.Y+d.At.H <= gutter.Y+gutter.H {
			return true
		}
	}
	return false
}

func (w *world) foldsSteps(sc *godog.ScenarioContext) {
	sc.When(`^I click the chevron of "([^"]*)"$`, func(title string) error {
		_, slot, err := w.branchOf(title)
		if err != nil {
			return err
		}
		w.clickAt(FoldAt(slot))
		return nil
	})
	sc.Then(`^the objective "([^"]*)" is folded, its chevron in the gutter$`, func(title string) error {
		b, slot, err := w.branchOf(title)
		if err != nil {
			return err
		}
		if !w.w.Folded(b.Goal.Number) {
			return fmt.Errorf("%q is unfolded", title)
		}
		if !w.chevronDrawnAt(slot) {
			return fmt.Errorf("no chevron is drawn in the gutter of %q, at %+v", title, FoldAt(slot))
		}
		return nil
	})
	sc.Then(`^the objective "([^"]*)" is unfolded$`, func(title string) error {
		b, _, err := w.branchOf(title)
		if err != nil {
			return err
		}
		if w.w.Folded(b.Goal.Number) {
			return fmt.Errorf("%q is folded", title)
		}
		return nil
	})
	sc.Then(`^the objective "([^"]*)" has no chevron$`, func(title string) error {
		_, slot, err := w.branchOf(title)
		if err != nil {
			return err
		}
		if w.chevronDrawnAt(slot) {
			return fmt.Errorf("a chevron is drawn beside %q", title)
		}
		return nil
	})
	sc.Then(`^the objective "([^"]*)" says "([^"]*)" after its title$`, func(title, kids string) error {
		_, slot, err := w.branchOf(title)
		if err != nil {
			return err
		}
		if said := w.saidWithin(slot); !slices.Contains(said, kids) {
			return fmt.Errorf("the objective %q says %q", title, said)
		}
		return nil
	})
	sc.Then(`^the objectives' head counts "([^"]*)"$`, func(want string) error {
		if got := w.w.GoalsCount(); got != want {
			return fmt.Errorf("the head counts %q", got)
		}
		return w.drawnSomewhere(want)
	})
	sc.Then(`^the objective "([^"]*)" is drawn faint, saying "([^"]*)"$`, func(title, says string) error {
		b, slot, err := w.branchOf(title)
		if err != nil {
			return err
		}
		if !b.Holds {
			return fmt.Errorf("%q is drawn as a match", title)
		}
		if said := w.saidWithin(slot); !slices.Contains(said, says) {
			return fmt.Errorf("the objective %q says %q", title, said)
		}
		return nil
	})
	sc.Then(`^the objective "([^"]*)" ends in the late ink with "([^"]*)"$`, func(title, want string) error {
		b, slot, err := w.branchOf(title)
		if err != nil {
			return err
		}
		if got, _ := w.w.SignalOf(b); got != want {
			return fmt.Errorf("the objective %q ends with %q", title, got)
		}
		if said := w.saidWithin(slot); !slices.Contains(said, want) {
			return fmt.Errorf("the objective %q is drawn saying %q", title, said)
		}
		return nil
	})
	sc.Then(`^the objective "([^"]*)" says nothing in the late ink$`, func(title string) error {
		b, _, err := w.branchOf(title)
		if err != nil {
			return err
		}
		if got, _ := w.w.SignalOf(b); got != "" {
			return fmt.Errorf("the objective %q ends with %q", title, got)
		}
		return nil
	})
	sc.When(`^I rest the pointer on what "([^"]*)" hides$`, func(title string) error {
		b, slot, err := w.branchOf(title)
		if err != nil {
			return err
		}
		said, dot := w.w.SignalOf(b)
		_, _, second := ItemLine(ItemOf(slot, 0))
		_, words := SignalAt(second, said, dot, w.w.measured())
		w.w = w.w.Hovering(Point{X: words.X + words.W/2, Y: words.Y + words.H/2})
		return nil
	})
	sc.Then(`^the bubble holds the line "([^"]*)"$`, func(line string) error {
		if lines := w.w.resting; !slices.Contains(lines, line) {
			return fmt.Errorf("the bubble says %q", lines)
		}
		return nil
	})
}
