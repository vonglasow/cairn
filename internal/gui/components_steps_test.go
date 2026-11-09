package gui

import (
	"fmt"
	"image"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/cucumber/godog"

	"cairn/internal/engine"
)

func partsDrawn(w Window) []Drawing {
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Constraints: layout.Exact(image.Pt(w.width, w.height)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
	var drawn Drawn
	Draw(NewKit(gtx, Theme(), Dark(), &drawn), w)
	return drawn.All()
}

func (w *world) componentsSteps(sc *godog.ScenarioContext) {
	sc.Then(`^the card "([^"]*)" counts the labels it leaves out in a pill with no ring$`, w.countWithNoRing)
	sc.Then(`^every key the command bar names sits on a chip$`, w.paletteKeysOnChips)
	sc.Then(`^the slot of the column "([^"]*)" is rounded as a card is$`, w.emptyRounded)
	sc.Then(`^the dot beside the label "([^"]*)" is as wide as the workspace's dot$`, w.labelDotAsWide)
	sc.Then(`^the switcher is drawn as a popover$`, func() error {
		return w.drawnAsAPopover("the switcher", w.w.SwitcherBox())
	})
	sc.Then(`^the labels are shown in a popover$`, func() error {
		bubble, _, ok := w.w.LabelsBubbleBox()
		if !ok {
			return fmt.Errorf("no bubble shows labels")
		}
		return w.drawnAsAPopover("the bubble", bubble)
	})
}

func (w *world) countWithNoRing(title string) error {
	for _, p := range w.w.Placed() {
		if p.Card.Title != title {
			continue
		}
		laid := CardRowOf(p.Card, PartsOf(p.Slot).Number, Measured(Theme()), time.Time{}, w.w.pages)
		if laid.Count == "" {
			return fmt.Errorf("the card %q leaves no label out", title)
		}
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		if got := pixel(shot, laid.More.X+laid.More.W/2, laid.More.Y); got != Dark().Colour.Surface.Control {
			return fmt.Errorf("the count %q is edged in %v rather than drawn as a plain pill", laid.Count, got)
		}
		return nil
	}
	return fmt.Errorf("no card titled %q on the board", title)
}

func (w *world) paletteKeysOnChips() error {
	shown := w.w
	shown.palette.opened = shown.now.Add(-time.Hour)
	shot, err := drawnInAStep(shown)
	if err != nil {
		return err
	}
	c := Dark().Colour.Surface
	rows, keyed := shown.PaletteRows(), 0
	for _, line := range shown.PaletteLines() {
		if line.Row < 0 || rows[line.Row].Key == "" || line.Box.Y+line.Box.H > shown.height {
			continue
		}
		ground := c.Control
		if line.Row == shown.palette.row {
			ground = c.ChipOnControl
		}
		right := Box{X: line.Box.X + line.Box.W/2, Y: line.Box.Y, W: line.Box.W / 2, H: line.Box.H}
		if painted(shot, right, ground) == 0 {
			return fmt.Errorf("the key of %q stands on no chip", rows[line.Row].Says)
		}
		keyed++
	}
	if keyed < 2 {
		return fmt.Errorf("the command bar names %d key", keyed)
	}
	return nil
}

func (w *world) emptyRounded(column string) error {
	slot := EmptyAt(w.w.Layout().CardAt(engine.IndexOf(column), 0, 0))
	shot, err := drawnInAStep(w.w)
	if err != nil {
		return err
	}
	s := Dark()
	if got := pixel(shot, slot.X, slot.Y); got == s.Colour.Line.Edge {
		return fmt.Errorf("the slot of %q is square at its corner", column)
	}
	if painted(shot, Box{X: slot.X + s.Radius.Card, Y: slot.Y, W: slot.W - 2*s.Radius.Card, H: 1}, s.Colour.Line.Edge) == 0 {
		return fmt.Errorf("the slot of %q is drawn with no line", column)
	}
	return nil
}

func (w *world) labelDotAsWide(label string) error {
	shot, err := drawnInAStep(w.w)
	if err != nil {
		return err
	}
	labels := Dark().Colour.Labels
	across := func(dot Box, ink string) int {
		return painted(shot, Box{X: dot.X - dot.W, Y: dot.Y + dot.H/2, W: 3 * dot.W, H: 1}, labels[ink])
	}
	place, _, _ := PlaceParts(w.w.Layout().Place)
	for _, row := range w.w.LabelRows() {
		if row.Carried.Label != label {
			continue
		}
		mine, theirs := across(row.Dot, w.w.LabelColour(label)), across(place, w.w.Place().Colour)
		if mine == 0 || mine != theirs {
			return fmt.Errorf("the dot beside %q is %d wide, and the workspace's %d", label, mine, theirs)
		}
		return nil
	}
	return fmt.Errorf("no label %q is listed", label)
}

func (w *world) drawnAsAPopover(what string, box Box) error {
	for _, d := range partsDrawn(w.w) {
		if d.At == box {
			if d.Part != "popover" || d.Layer != LayerPopover {
				return fmt.Errorf("%s is drawn as a %s on layer %v", what, d.Part, d.Layer)
			}
			return nil
		}
	}
	return fmt.Errorf("%s is not drawn at %+v", what, box)
}
