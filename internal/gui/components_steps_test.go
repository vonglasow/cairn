package gui

import (
	"cairn/internal/gui/design"
	"fmt"
	"image"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/cucumber/godog"
)

func partsDrawn(w Window) []Drawing { return partsDrawnWith(drawingTheme(), w) }

func partsDrawnWith(th *material.Theme, w Window) []Drawing {
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Constraints: layout.Exact(image.Pt(w.width, w.height)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
	var drawn Drawn
	Draw(NewKit(gtx, th, design.Dark(), &drawn), w)
	return drawn.All()
}

func (w *world) componentsSteps(sc *godog.ScenarioContext) {
	sc.Then(`^the card "([^"]*)" counts the labels it leaves out in a pill with no ring$`, w.countWithNoRing)
}

func (w *world) countWithNoRing(title string) error {
	for _, p := range w.w.Placed() {
		if p.Card.Title != title {
			continue
		}
		laid := CardRowOf(p.Card, PartsOf(p.Slot).Number, Measured(w.theme), time.Time{}, w.w.Pages)
		if laid.Count == "" {
			return fmt.Errorf("the card %q leaves no label out", title)
		}
		shot, err := drawnInAStep(w.theme, w.w)
		if err != nil {
			return err
		}
		if got := pixel(shot, laid.More.X+laid.More.W/2, laid.More.Y); got != design.Dark().Colour.Surface.Control {
			return fmt.Errorf("the count %q is edged in %v rather than drawn as a plain pill", laid.Count, got)
		}
		return nil
	}
	return fmt.Errorf("no card titled %q on the board", title)
}
