package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"fmt"
	"strings"

	"gioui.org/io/pointer"
	"github.com/cucumber/godog"
)

func (w *world) gutters(sc *godog.ScenarioContext) {
	var (
		opened geom.Layout
		taken  bool
	)
	drag := func(gutter geom.Box, right int) {
		if !taken {
			opened, taken = w.w.Layout(), true
		}
		from := geom.Point{X: gutter.X + gutter.W/2, Y: gutter.Y + gutter.H/2}
		to := geom.Point{X: from.X + right, Y: from.Y}
		w.w = w.w.PressAt(from).MoveTo(to).LetGo(to)
	}
	lift := func(by int) {
		if !taken {
			opened, taken = w.w.Layout(), true
		}
		edge := geom.BandEdge(w.w.Layout())
		from := geom.Point{X: edge.X + edge.W/2, Y: edge.Y + edge.H/2}
		to := geom.Point{X: from.X, Y: from.Y - by}
		w.w = w.w.PressAt(from).MoveTo(to).LetGo(to)
	}
	twice := func(on geom.Box) {
		at := geom.Point{X: on.X + on.W/2, Y: on.Y + on.H/2}
		var g Gesture
		for range 2 {
			w.w, g = Pointing(w.w, g, pointer.Press, at, 0, w.today)
			w.w, g = Pointing(w.w, g, pointer.Release, at, 0, w.today)
		}
	}
	column := func(name string) (int, error) {
		at := board.IndexOf(strings.ToLower(name))
		if at < 0 {
			return 0, fmt.Errorf("no column called %q", name)
		}
		return at, nil
	}

	sc.Step(`^I drag the gutter between "([^"]*)" and "([^"]*)" (\d+) to the right$`, func(left, right string, by int) error {
		at, err := column(left)
		if err != nil {
			return err
		}
		if next, err := column(right); err != nil || next != at+1 {
			return fmt.Errorf("%q does not stand right of %q", right, left)
		}
		drag(geom.Gutter(w.w.Layout(), at), by)
		return nil
	})
	sc.Step(`^the pointer over the gutter after "([^"]*)" is the one for widening sideways$`, func(name string) error {
		at, err := column(name)
		if err != nil {
			return err
		}
		return pointerIs(w.w, geom.Gutter(w.w.Layout(), at), ColResize)
	})
	sc.Step(`^I drag the gutter between the "([^"]*)" and "([^"]*)" lanes (\d+) to the right$`, func(left, right string, by int) error {
		at, err := laneNamed(left)
		if err != nil {
			return err
		}
		if next, err := laneNamed(right); err != nil || next != at+1 {
			return fmt.Errorf("%q does not stand right of %q", right, left)
		}
		drag(geom.LaneGutter(w.w.Layout(), at), by)
		return nil
	})
	wider := func(was, is, by int, name string) error {
		if is-was != by {
			return fmt.Errorf("%q is %d wide, opened at %d: %+d, not %+d", name, is, was, is-was, by)
		}
		return nil
	}
	sc.Step(`^"([^"]*)" is (\d+) wider than it opened$`, func(name string, by int) error {
		at, err := column(name)
		if err != nil {
			return err
		}
		return wider(opened.Column[at].W, w.w.Layout().Column[at].W, by, name)
	})
	sc.Step(`^no card is being dragged$`, func() error {
		if _, dragging := w.w.Ghost(); dragging || w.w.held.Number != 0 {
			return fmt.Errorf("a card is held: %+v", w.w.held)
		}
		return nil
	})
	sc.Given(`^the configuration keeps the window's sizes$`, func() error {
		w.keepSizes = true
		return nil
	})
	sc.Step(`^I drag the edge above the band (\d+) up$`, func(by int) error {
		lift(by)
		return nil
	})
	sc.Step(`^I double-click the edge above the band$`, func() error {
		twice(geom.BandEdge(w.w.Layout()))
		return nil
	})
	sc.Step(`^I double-click the gutter between "([^"]*)" and "([^"]*)"$`, func(left, _ string) error {
		at, err := column(left)
		if err != nil {
			return err
		}
		twice(geom.Gutter(w.w.Layout(), at))
		return nil
	})
	sc.Step(`^I double-click the gutter between the "([^"]*)" and "([^"]*)" lanes$`, func(left, _ string) error {
		at, err := laneNamed(left)
		if err != nil {
			return err
		}
		twice(geom.LaneGutter(w.w.Layout(), at))
		return nil
	})
	taller := func(was, is, by int, name string) error {
		if is-was != by {
			return fmt.Errorf("%s is %d tall, opened at %d: %+d, not %+d", name, is, was, is-was, by)
		}
		return nil
	}
	sc.Step(`^the band is (\d+) taller than it opened$`, func(by int) error {
		return taller(opened.Band.H, w.w.Layout().Band.H, by, "the band")
	})
	sc.Step(`^the band is as tall as it opened$`, func() error {
		return taller(geom.LayoutFor(w.w.width, w.w.height).Band.H, w.w.Layout().Band.H, 0, "the band")
	})
	sc.Step(`^"([^"]*)" is (\d+) shorter than it opened$`, func(name string, by int) error {
		at, err := column(name)
		if err != nil {
			return err
		}
		return taller(opened.Column[at].H, w.w.Layout().Column[at].H, -by, name)
	})
	sc.Step(`^"([^"]*)" is as tall as it opened$`, func(name string) error {
		at, err := column(name)
		if err != nil {
			return err
		}
		return taller(opened.Column[at].H, w.w.Layout().Column[at].H, 0, name)
	})
	sc.Step(`^every lane is as wide as the others$`, func() error {
		l := w.w.Layout()
		for at := range l.Lane {
			if l.Lane[at].W != l.Lane[0].W {
				return fmt.Errorf("lane %d is %d wide, the first %d", at, l.Lane[at].W, l.Lane[0].W)
			}
		}
		return nil
	})
	sc.Step(`^every column is as wide as the others$`, func() error {
		l := w.w.Layout()
		for at := range board.Columns {
			if l.Column[at].W != l.Column[0].W {
				return fmt.Errorf("column %d is %d wide, the first %d", at, l.Column[at].W, l.Column[0].W)
			}
		}
		return nil
	})
}

func pointerIs(w Window, over geom.Box, want Pointer) error {
	if got := w.PointerAt(geom.Point{X: over.X + over.W/2, Y: over.Y + over.H/2}); got != want {
		return fmt.Errorf("the pointer over %+v is %v, not %v", over, got, want)
	}
	return nil
}
