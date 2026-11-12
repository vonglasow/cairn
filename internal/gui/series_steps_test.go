package gui

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/config"
	"cairn/internal/domain"
	"cairn/internal/engine"
)

func (w *world) registerSeries(sc *godog.ScenarioContext) {
	sc.Given(`^recurrence is (on|off) in the window$`, func(said string) error {
		w.recurs = said == "on"
		return os.WriteFile(config.SettingsBeside(w.path), []byte("[series]\non = "+fmt.Sprint(w.recurs)+"\n"), 0o600)
	})
	sc.Then(`^the card "([^"]*)" says its rule "([^"]*)"$`, func(title, rule string) error {
		slot, err := w.cardSlotOf(title)
		if err != nil {
			return err
		}
		for _, one := range textDrawnWith(w.theme, w.w) {
			if one.said == rule && slot.In(Point{X: one.box.X, Y: one.box.Y}) {
				return nil
			}
		}
		return fmt.Errorf("the card %q does not write its rule %q", title, rule)
	})
	sc.Then(`^the trace says "([^"]*)"$`, func(want string) error {
		if !strings.Contains(strings.Join(w.w.Lines(), "\n"), want) {
			return fmt.Errorf("the trace reads %q, with no %q", w.w.Lines(), want)
		}
		return nil
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
		for _, one := range textDrawnWith(w.theme, w.w) {
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
		wanted, err := domain.ParseRule(rule)
		if err != nil {
			return err
		}
		parts := wanted.Parts()
		if err := w.walkRuleTo("every", strconv.Itoa(parts.Every)); err != nil {
			return err
		}
		if err := w.clickInRulePicker(parts.Unit); err != nil {
			return err
		}
		for _, day := range parts.Days {
			if err := w.clickInRulePicker(Weekdays[(int(day)+6)%7]); err != nil {
				return err
			}
		}
		if w.w.Form().Fields()[w.w.Form().On()].Value != rule {
			return fmt.Errorf("the picker wrote %q, not %q", w.w.Form().Fields()[w.w.Form().On()].Value, rule)
		}
		w.w = w.w.Choose("Choose")
		return w.w.Wrong()
	})
	sc.Then(`^the rule's picker offers "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)" and "([^"]*)"$`, func(a, b, c, d, e string) error {
		var said []string
		for _, line := range w.w.RulePicked().Lines {
			said = append(said, line.Said)
		}
		for _, want := range []string{a, b, c, d, e} {
			if !slices.Contains(said, want) {
				return fmt.Errorf("the rule's picker offers %q, with no %q", said, want)
			}
		}
		return nil
	})
	sc.Step(`^I click "([^"]*)" in the rule's picker$`, w.clickInRulePicker)
	sc.Step(`^I click the next of "([^"]*)" in the rule's picker$`, func(said string) error {
		for _, line := range w.w.RulePicked().Lines {
			if line.Said == said && line.Next.W > 0 {
				w.w = w.w.clicked(Point{X: line.Next.X + line.Next.W/2, Y: line.Next.Y + line.Next.H/2})
				return nil
			}
		}
		return fmt.Errorf("the rule's picker walks no %q on", said)
	})
	sc.Then(`^the rule's picker is closed$`, func() error {
		if w.w.Screen() == OnRule {
			return fmt.Errorf("the rule's picker is still open")
		}
		return nil
	})
	sc.Then(`^under the field "([^"]*)" the window reads "([^"]*)"$`, func(_, want string) error {
		if said := w.readBack(); !strings.Contains(said, want) {
			return fmt.Errorf("under the field the window reads %q, not %q", said, want)
		}
		return nil
	})
	sc.Then(`^under the field "([^"]*)" the window reads the next three days$`, func(string) error {
		if said := w.readBack(); strings.Count(said, " · ") != 2 {
			return fmt.Errorf("under the field the window reads %q, not three days", said)
		}
		return nil
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
		for _, one := range textDrawnWith(w.theme, w.w) {
			if box.In(Point{X: one.box.X, Y: one.box.Y}) {
				all = append(all, one.said)
			}
		}
		if !strings.Contains(strings.Join(all, " "), said) {
			return fmt.Errorf("the form says of its series %q, and nothing of %q", all, said)
		}
		return nil
	})
	sc.Then(`^the form's heading carries the loop$`, func() error {
		if _, ok := w.w.OccurrenceHeading(); !ok {
			return fmt.Errorf("the form's heading is %q, with no loop", HeadsOn(w.w, OnForm))
		}
		return nil
	})
	sc.Then(`^the form's heading says "([^"]*)"$`, func(want string) error {
		head, _ := w.w.OccurrenceHeading()
		if !strings.Contains(head.Title+" "+head.Number+" "+head.Under, want) {
			return fmt.Errorf("the form's heading says %+v, not %q", head, want)
		}
		return nil
	})
	sc.Step(`^I set the series' state to "([^"]*)"$`, func(state string) error {
		for range 4 {
			if w.w.SeriesState() == state {
				return nil
			}
			for _, pill := range w.w.Pills() {
				if pill.Kind == PillState {
					w.w = w.w.clicked(Point{X: pill.Box.X + inset, Y: pill.Box.Y + pill.Box.H/2})
				}
			}
		}
		return fmt.Errorf("the series' state never reached %q: %q", state, w.w.SeriesState())
	})
	sc.Then(`^the series listed are (.+)$`, func(list string) error {
		var got []string
		for _, s := range w.w.SeriesListed() {
			got = append(got, strconv.Quote(s.Title))
		}
		if strings.Join(got, ", ") != list {
			return fmt.Errorf("the series listed are %s, not %s", strings.Join(got, ", "), list)
		}
		return nil
	})
	sc.Then(`^the board shows the pill "([^"]*)"$`, func(said string) error {
		var all []string
		for _, pill := range w.w.Pills() {
			all = append(all, pill.Said)
			if pill.Said == said {
				return nil
			}
		}
		return fmt.Errorf("the pills read %q, with no %q", all, said)
	})
	sc.Then(`^the series tab says "([^"]*)"$`, func(want string) error {
		var all []string
		for _, one := range textDrawnWith(w.theme, w.w) {
			all = append(all, one.said)
		}
		if !strings.Contains(strings.Join(all, " "), want) {
			return fmt.Errorf("the series tab says %q, and nothing of %q", all, want)
		}
		return nil
	})
	sc.Then(`^the bar offers "([^"]*)"$`, func(offer string) error {
		if !slices.Contains(w.w.Offers(), offer) {
			return fmt.Errorf("the bar offers %q, not %q", w.w.Offers(), offer)
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

func (w *world) clickInRulePicker(said string) error {
	for _, line := range w.w.RulePicked().Lines {
		for i, choice := range line.Choices {
			if choice == said && i < len(line.Picks) {
				pick := line.Picks[i]
				w.w = w.w.clicked(Point{X: pick.X + pick.W/2, Y: pick.Y + pick.H/2})
				return nil
			}
		}
	}
	return fmt.Errorf("the rule's picker offers no %q", said)
}

func (w *world) walkRuleTo(said, choice string) error {
	for range 100 {
		for _, line := range w.w.RulePicked().Lines {
			if line.Said != said {
				continue
			}
			if line.On >= 0 && line.Choices[line.On] == choice {
				return nil
			}
			w.w = w.w.clicked(Point{X: line.Next.X + line.Next.W/2, Y: line.Next.Y + line.Next.H/2})
		}
	}
	return fmt.Errorf("the rule's picker never walks %q to %q", said, choice)
}

func (w *world) readBack() string {
	box, ok := RepeatsReadAt(w.w.formBox(), w.w.Form().Fields())
	if !ok {
		return ""
	}
	var said []string
	for _, one := range textDrawnWith(w.theme, w.w) {
		if box.In(Point{X: one.box.X, Y: one.box.Y}) {
			said = append(said, one.said)
		}
	}
	return strings.Join(said, " ")
}
