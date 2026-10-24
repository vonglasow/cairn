package gui

import (
	"fmt"
	"strings"

	"cairn/internal/engine"

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
