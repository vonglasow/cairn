package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"fmt"
	"slices"

	"github.com/cucumber/godog"
)

func (w *world) branchOf(title string) (board.Branch, geom.Box, error) {
	for i, b := range w.w.Tree() {
		if b.Goal.Title == title {
			return b, GoalAt(w.w.Layout(), i, w.w.goals), nil
		}
	}
	return board.Branch{}, geom.Box{}, fmt.Errorf("no objective titled %q is listed", title)
}

func (w *world) saidWithin(box geom.Box) []string {
	var said []string
	for _, one := range textDrawnWith(w.theme, w.w) {
		if box.In(geom.Point{X: one.box.X + one.box.W/2, Y: one.box.Y + one.box.H/2}) {
			said = append(said, one.said)
		}
	}
	return said
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
		w.w = w.w.Hovering(geom.Point{X: words.X + words.W/2, Y: words.Y + words.H/2})
		return nil
	})
	sc.Then(`^the bubble holds the line "([^"]*)"$`, func(line string) error {
		if lines := w.w.resting; !slices.Contains(lines, line) {
			return fmt.Errorf("the bubble says %q", lines)
		}
		return nil
	})
}
