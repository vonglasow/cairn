package gui

import (
	"fmt"
	"strings"

	"cairn/internal/engine"

	"gioui.org/io/pointer"
	"github.com/cucumber/godog"
)

func (w *world) gutters(sc *godog.ScenarioContext) {
	var (
		opened Layout
		taken  bool
	)
	drag := func(gutter Box, right int) {
		if !taken {
			opened, taken = w.w.Layout(), true
		}
		from := Point{X: gutter.X + gutter.W/2, Y: gutter.Y + gutter.H/2}
		to := Point{X: from.X + right, Y: from.Y}
		w.w = w.w.PressAt(from).MoveTo(to).LetGo(to)
	}
	lift := func(by int) {
		if !taken {
			opened, taken = w.w.Layout(), true
		}
		edge := BandEdge(w.w.Layout())
		from := Point{X: edge.X + edge.W/2, Y: edge.Y + edge.H/2}
		to := Point{X: from.X, Y: from.Y - by}
		w.w = w.w.PressAt(from).MoveTo(to).LetGo(to)
	}
	twice := func(on Box) {
		at := Point{X: on.X + on.W/2, Y: on.Y + on.H/2}
		var g Gesture
		for range 2 {
			w.w, g = Pointing(w.w, g, pointer.Press, at, 0, w.today)
			w.w, g = Pointing(w.w, g, pointer.Release, at, 0, w.today)
		}
	}
	column := func(name string) (int, error) {
		at := engine.IndexOf(strings.ToLower(name))
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
		drag(Gutter(w.w.Layout(), at), by)
		return nil
	})
	sc.Step(`^I drag the gutter between "([^"]*)" and "([^"]*)" (\d+) to the left$`, func(left, right string, by int) error {
		at, err := column(left)
		if err != nil {
			return err
		}
		if next, err := column(right); err != nil || next != at+1 {
			return fmt.Errorf("%q does not stand right of %q", right, left)
		}
		drag(Gutter(w.w.Layout(), at), -by)
		return nil
	})
	sc.Step(`^I drag the gutter between "backlog" and "todo" from beside the band (\d+) to the right$`, func(by int) error {
		l := w.w.Layout()
		gutter := Gutter(l, 0)
		beside := Box{X: gutter.X, Y: l.Band.Y, W: gutter.W, H: l.Band.H}
		if !gutter.In(Point{X: beside.X + beside.W/2, Y: beside.Y + beside.H/2}) {
			return fmt.Errorf("the gutter %+v does not run beside the band %+v", gutter, l.Band)
		}
		drag(beside, by)
		return nil
	})
	sc.Step(`^the band starts where "todo" does$`, func() error {
		l := w.w.Layout()
		if l.Band.X != l.Column[1].X {
			return fmt.Errorf("the band starts at %d, todo at %d", l.Band.X, l.Column[1].X)
		}
		return nil
	})
	sc.Step(`^I drag the gutter between the "([^"]*)" and "([^"]*)" lanes (\d+) to the right$`, func(left, right string, by int) error {
		at, err := laneNamed(left)
		if err != nil {
			return err
		}
		if next, err := laneNamed(right); err != nil || next != at+1 {
			return fmt.Errorf("%q does not stand right of %q", right, left)
		}
		drag(LaneGutter(w.w.Layout(), at), by)
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
	sc.Step(`^"([^"]*)" is (\d+) narrower than it opened$`, func(name string, by int) error {
		at, err := column(name)
		if err != nil {
			return err
		}
		return wider(opened.Column[at].W, w.w.Layout().Column[at].W, -by, name)
	})
	sc.Step(`^"([^"]*)" is as wide as it opened$`, func(name string) error {
		at, err := column(name)
		if err != nil {
			return err
		}
		return wider(opened.Column[at].W, w.w.Layout().Column[at].W, 0, name)
	})
	sc.Step(`^the "([^"]*)" lane is (\d+) wider than it opened$`, func(name string, by int) error {
		at, err := laneNamed(name)
		if err != nil {
			return err
		}
		return wider(opened.Lane[at].W, w.w.Layout().Lane[at].W, by, name)
	})
	sc.Step(`^the "([^"]*)" lane is (\d+) narrower than it opened$`, func(name string, by int) error {
		at, err := laneNamed(name)
		if err != nil {
			return err
		}
		return wider(opened.Lane[at].W, w.w.Layout().Lane[at].W, -by, name)
	})
	sc.Step(`^"([^"]*)" is the narrowest a column goes$`, func(name string) error {
		at, err := column(name)
		if err != nil {
			return err
		}
		if got := w.w.Layout().Column[at].W; got != NarrowestColumn {
			return fmt.Errorf("%q is %d wide, not the %d a column goes down to", name, got, NarrowestColumn)
		}
		return nil
	})
	sc.Step(`^"([^"]*)" is wider than "([^"]*)"$`, func(one, other string) error {
		a, err := column(one)
		if err != nil {
			return err
		}
		b, err := column(other)
		if err != nil {
			return err
		}
		l := w.w.Layout()
		if l.Column[a].W <= l.Column[b].W {
			return fmt.Errorf("%q is %d wide and %q %d", one, l.Column[a].W, other, l.Column[b].W)
		}
		return nil
	})
	sc.Step(`^the columns fill the window's width$`, func() error {
		l := w.w.Layout()
		last := l.Column[engine.Columns-1]
		if end := last.X + last.W; end != l.Tabs.X+l.Tabs.W {
			return fmt.Errorf("the columns end at %d, the window's room at %d", end, l.Tabs.X+l.Tabs.W)
		}
		return nil
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
	sc.Step(`^I drag the edge above the band (\d+) down$`, func(by int) error {
		lift(-by)
		return nil
	})
	sc.Step(`^I double-click the edge above the band$`, func() error {
		twice(BandEdge(w.w.Layout()))
		return nil
	})
	sc.Step(`^I double-click the gutter between "([^"]*)" and "([^"]*)"$`, func(left, _ string) error {
		at, err := column(left)
		if err != nil {
			return err
		}
		twice(Gutter(w.w.Layout(), at))
		return nil
	})
	sc.Step(`^I double-click the gutter between the "([^"]*)" and "([^"]*)" lanes$`, func(left, _ string) error {
		at, err := laneNamed(left)
		if err != nil {
			return err
		}
		twice(LaneGutter(w.w.Layout(), at))
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
		return taller(LayoutFor(w.w.width, w.w.height).Band.H, w.w.Layout().Band.H, 0, "the band")
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
	sc.Step(`^the band's lanes hold one card's room$`, func() error {
		for at, lane := range w.w.Layout().Lane {
			if lane.H != CardHeight {
				return fmt.Errorf("lane %d is %d tall, not a card's %d", at, lane.H, CardHeight)
			}
		}
		return nil
	})
	sc.Step(`^"([^"]*)" holds one card's room$`, func(name string) error {
		at, err := column(name)
		if err != nil {
			return err
		}
		if got := w.w.Layout().Column[at].H; got != CardHeight {
			return fmt.Errorf("%q is %d tall, not a card's %d", name, got, CardHeight)
		}
		return nil
	})
	sc.Step(`^the band is taller than it opens at (\d+) by (\d+)$`, func(width, high int) error {
		if got, opens := w.w.Layout().Band.H, LayoutFor(width, high).Band.H; got <= opens {
			return fmt.Errorf("the band is %d tall, and opens %d tall at %d by %d", got, opens, width, high)
		}
		return nil
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
		for at := range engine.Columns {
			if l.Column[at].W != l.Column[0].W {
				return fmt.Errorf("column %d is %d wide, the first %d", at, l.Column[at].W, l.Column[0].W)
			}
		}
		return nil
	})
}
