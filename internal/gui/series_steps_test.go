package gui

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/config"
	"cairn/internal/engine"
)

func (w *world) registerSeries(sc *godog.ScenarioContext) {
	sc.Given(`^recurrence is (on|off) in the window$`, func(said string) error {
		w.recurs = said == "on"
		return os.WriteFile(config.SettingsBeside(w.path), []byte("[recurrence]\non = "+fmt.Sprint(w.recurs)+"\n"), 0o600)
	})
	sc.Then(`^the card "([^"]*)" says its rule "([^"]*)"$`, func(title, rule string) error {
		slot, err := w.cardSlotOf(title)
		if err != nil {
			return err
		}
		for _, one := range textDrawn(w.w) {
			if one.said == rule && slot.In(Point{X: one.box.X, Y: one.box.Y}) {
				return nil
			}
		}
		return fmt.Errorf("the card %q does not write its rule %q", title, rule)
	})
	sc.Then(`^the card "([^"]*)" leads to series (\d+)$`, func(title string, n int) error {
		row, err := w.rowOf(title)
		if err != nil {
			return err
		}
		if row.Series != n || row.SeriesAt.W == 0 {
			return fmt.Errorf("the card %q leads to series %d at %+v", title, row.Series, row.SeriesAt)
		}
		return nil
	})
	sc.Then(`^the card "([^"]*)" says nothing of a series$`, func(title string) error {
		row, err := w.rowOf(title)
		if err != nil {
			return err
		}
		if row.Series != 0 {
			return fmt.Errorf("the card %q leads to series %d", title, row.Series)
		}
		c, err := w.cardOf(title)
		if err != nil {
			return err
		}
		if said := RuleOf(c); said != "" {
			return fmt.Errorf("the card %q says its rule %q", title, said)
		}
		return nil
	})
	sc.Step(`^I click the series of the card "([^"]*)"$`, func(title string) error {
		row, err := w.rowOf(title)
		if err != nil {
			return err
		}
		w.w = w.w.clicked(Point{X: row.SeriesAt.X + row.SeriesAt.W/2, Y: row.SeriesAt.Y + row.SeriesAt.H/2})
		return w.w.Wrong()
	})
	sc.Then(`^the series shown is series (\d+)$`, func(n int) error {
		if s, ok := w.w.SeriesShown(); !ok || s.Number != n {
			return fmt.Errorf("the series shown is %d", s.Number)
		}
		return nil
	})
	sc.Then(`^the list of series reads (".*")$`, func(list string) error {
		var got []string
		for _, row := range w.w.SeriesRows() {
			if row.Head != "" {
				got = append(got, row.Head)
				continue
			}
			got = append(got, row.Series.Title)
		}
		if want := quoted(list); !slices.Equal(got, want) {
			return fmt.Errorf("the list of series reads %q, want %q", got, want)
		}
		return nil
	})
	sc.Then(`^the series shown says "([^"]*)"$`, func(said string) error {
		box := w.w.SeriesDetail().Box
		var all []string
		for _, one := range textDrawn(w.w) {
			if box.In(Point{X: one.box.X, Y: one.box.Y}) {
				all = append(all, one.said)
			}
		}
		if !strings.Contains(strings.Join(all, " "), said) {
			return fmt.Errorf("the series shown says %q, and nothing of %q", all, said)
		}
		return nil
	})
	sc.Then(`^the series shown offers "([^"]*)"$`, func(offer string) error {
		if offers := w.w.SeriesDetail().Offers; !slices.Contains(offers, offer) {
			return fmt.Errorf("the series shown offers %q, not %q", offers, offer)
		}
		return nil
	})
	sc.Step(`^I click the series "([^"]*)"$`, func(title string) error {
		for _, row := range w.w.SeriesRows() {
			if row.Head == "" && row.Series.Title == title {
				w.w = w.w.clicked(Point{X: row.At.X + row.At.W/2, Y: row.At.Y + row.At.H/2})
				return w.w.Wrong()
			}
		}
		return fmt.Errorf("no series titled %q is listed", title)
	})
	sc.Step(`^I click the occurrence of task (\d+)$`, func(n int) error {
		for _, line := range w.w.SeriesDetail().Occurrences {
			if line.Occurrence.Task == n {
				w.w = w.w.clicked(Point{X: line.At.X + line.At.W/2, Y: line.At.Y + line.At.H/2})
				return w.w.Wrong()
			}
		}
		return fmt.Errorf("the series shown lists no occurrence of task %d", n)
	})
	sc.Then(`^the window asks what becomes of the tasks of series (\d+)$`, func(n int) error {
		if w.w.Screen() != OnSeriesFates || w.w.SeriesFatesAsked().Number != n {
			return fmt.Errorf("the window is on %v, asking about series %d", w.w.Screen(), w.w.SeriesFatesAsked().Number)
		}
		return nil
	})
	sc.Then(`^the choice taken is "([^"]*)"$`, func(said string) error {
		f := w.w.SeriesFatesAsked()
		if got := f.Rows[f.On].Said; got != said {
			return fmt.Errorf("the choice taken is %q, not %q", got, said)
		}
		return nil
	})
	sc.Step(`^I click the choice "([^"]*)"$`, func(said string) error {
		for _, row := range w.w.SeriesFatesAsked().Rows {
			if row.Said == said {
				w.w = w.w.clicked(Point{X: row.At.X + row.At.W/2, Y: row.At.Y + row.At.H/2})
				return nil
			}
		}
		return fmt.Errorf("no choice says %q", said)
	})
	sc.Step(`^I choose the rule "([^"]*)"$`, func(rule string) error {
		for i, one := range w.w.Choosing() {
			if one.Title == rule {
				w.w.choosing = i
				w.w = w.w.Choose("Choose")
				return w.w.Wrong()
			}
		}
		return fmt.Errorf("the list offers no rule %q", rule)
	})
	sc.Then(`^the form has no field "([^"]*)"$`, func(name string) error {
		if err := w.detailShowsField(name); err == nil {
			return fmt.Errorf("the form has a field %q", name)
		}
		return nil
	})
	sc.Then(`^the form says of its series "([^"]*)"$`, func(said string) error {
		box, ok := w.w.OccurrenceBox()
		if !ok {
			return fmt.Errorf("the form says nothing of a series")
		}
		var all []string
		for _, one := range textDrawn(w.w) {
			if box.In(Point{X: one.box.X, Y: one.box.Y}) {
				all = append(all, one.said)
			}
		}
		if !strings.Contains(strings.Join(all, " "), said) {
			return fmt.Errorf("the form says of its series %q, and nothing of %q", all, said)
		}
		return nil
	})
	sc.Then(`^the form offers "([^"]*)"$`, func(offer string) error {
		if w.w.Screen() != OnForm || !slices.Contains(w.w.Offers(), offer) {
			return fmt.Errorf("the window is on %v and offers %q", w.w.Screen(), w.w.Offers())
		}
		return nil
	})
	sc.Then(`^the command bar says "([^"]*)"$`, func(said string) error {
		if got, _ := w.w.PaletteBanner(); !strings.Contains(got, said) {
			return fmt.Errorf("the command bar says %q, not %q", got, said)
		}
		return nil
	})
}

func (w *world) shownPages() engine.Pages {
	if w.recurs {
		return w.pages
	}
	return w.pages.WithoutRecurrence()
}

func (w *world) cardSlotOf(title string) (Box, error) {
	for _, placed := range w.w.Placed() {
		if placed.Card.Title == title {
			return placed.Slot, nil
		}
	}
	return Box{}, fmt.Errorf("no card titled %q is on the board", title)
}

func (w *world) rowOf(title string) (CardRow, error) {
	slot, err := w.cardSlotOf(title)
	if err != nil {
		return CardRow{}, err
	}
	c, err := w.cardOf(title)
	if err != nil {
		return CardRow{}, err
	}
	return CardRowOf(c, PartsOf(slot).Number, w.w.measured(), moment, w.w.pages), nil
}
