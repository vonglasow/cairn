package gui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cucumber/godog"
)

func (w *world) objectivesSteps(sc *godog.ScenarioContext) {
	sc.Then(`^the objective "([^"]*)" is drawn indented under "([^"]*)" and joined to it$`, w.drawnUnder)
	sc.Then(`^the detail is headed "([^"]*)"$`, w.detailHeaded)
	sc.Then(`^the detail says "([^"]*)"$`, w.detailSays)
	sc.Then(`^the detail's sub-objectives read (".*")$`, func(list string) error {
		var got []string
		for _, t := range w.w.Detail().Tiles {
			got = append(got, t.Tile.Title)
		}
		return readAs("the detail's sub-objectives", got, quoted(list))
	})
	sc.Then(`^the detail's tasks on it itself read (".*")$`, func(list string) error {
		return readAs("the tasks on it itself", lineTitles(w.w.Detail().Own), quoted(list))
	})
	sc.Then(`^the detail's tasks in its sub-objectives read (".*")$`, func(list string) error {
		return readAs("the tasks in its sub-objectives", lineTitles(w.w.Detail().InSubs), quoted(list))
	})
	sc.Then(`^the tile of "([^"]*)" warns "([^"]*)"$`, w.tileWarns)
	sc.Then(`^the narrowing bar reads (".*")$`, func(list string) error {
		n := w.w.Narrowing()
		if len(n.At) == 0 {
			return fmt.Errorf("no narrowing bar is drawn")
		}
		var got []string
		for _, s := range n.Segments {
			got = append(got, fmt.Sprintf("%s %d", s.Name, s.Count))
		}
		return readAs("the narrowing bar", got, quoted(list))
	})
	sc.When(`^I click the segment "([^"]*)"$`, func(name string) error {
		n := w.w.Narrowing()
		for i, s := range n.Segments {
			if s.Name == name {
				at := n.At[i]
				w.w = w.w.clicked(Point{X: at.X + at.W/2, Y: at.Y + at.H/2})
				return nil
			}
		}
		return fmt.Errorf("the narrowing bar has no segment %q", name)
	})
	sc.Then(`^the card "([^"]*)" says it serves "([^"]*)"$`, w.cardServes)
	sc.Then(`^the progress says "([^"]*)" "([^"]*)"$`, func(name, value string) error {
		g, _ := w.w.snap.Goal(w.w.form.Number)
		for _, f := range w.w.GoalProgress(g).Lines {
			if f.Name == name && f.Value == value {
				return w.drawnSomewhere(value)
			}
		}
		return fmt.Errorf("the progress says %+v", w.w.GoalProgress(g).Lines)
	})
	sc.Then(`^the list draws "([^"]*)" indented under "([^"]*)" and joined to it$`, func(child, parent string) error {
		from := -1
		for i, c := range w.w.Choosing() {
			if c.Title == parent && c.Depth == 0 {
				from = i
			}
			if c.Title == child && from >= 0 && c.Depth == 1 {
				for _, d := range partsDrawn(w.w) {
					if strings.HasSuffix(d.Part, "row/connector") {
						return nil
					}
				}
				return fmt.Errorf("nothing joins %q to %q", child, parent)
			}
		}
		return fmt.Errorf("the list does not draw %q under %q: %+v", child, parent, w.w.Choosing())
	})
	sc.Then(`^the row "([^"]*)" of the list says "([^"]*)"$`, func(title, said string) error {
		for _, c := range w.w.Choosing() {
			if c.Title == title {
				if got := w.w.RowAside(c); got != said {
					return fmt.Errorf("the row %q says %q", title, got)
				}
				return w.drawnSomewhere(said)
			}
		}
		return fmt.Errorf("the list has no row %q", title)
	})
	sc.Then(`^the list says "([^"]*)"$`, func(said string) error {
		if got := w.w.ChoosingRefused(); got != said {
			return fmt.Errorf("the list says %q", got)
		}
		for _, d := range partsDrawn(w.w) {
			if strings.HasSuffix(d.Part, "banner") {
				return nil
			}
		}
		return fmt.Errorf("the refusal is not drawn")
	})
	sc.Then(`^the deletion asks about (".*")$`, func(list string) error {
		var got []string
		for _, row := range w.w.FatesAsked().Rows {
			if row.Sub != 0 {
				got = append(got, row.Said)
			}
		}
		return readAs("the deletion", got, quoted(strings.ReplaceAll(list, `" and "`, `", "`)))
	})
	sc.Then(`^the deletion asks about the tasks "([^"]*)"$`, func(said string) error {
		for _, row := range w.w.FatesAsked().Rows {
			if row.Sub == 0 && strings.HasPrefix(row.Said, said) {
				return nil
			}
		}
		return fmt.Errorf("the deletion asks nothing of the tasks %q", said)
	})
	sc.Then(`^the deletion cannot be confirmed yet$`, func() error {
		offers := w.w.Offers()
		if w.w.Screen() != OnFates || w.w.OfferLook(0, offers[0]) != buttonFaint {
			return fmt.Errorf("the deletion offers %v on %v, not faint", offers, w.w.Screen())
		}
		if w.w.Key("enter").Screen() != OnFates {
			return fmt.Errorf("enter deleted what was not chosen yet")
		}
		return nil
	})
	sc.When(`^I choose "([^"]*)" for (?:"([^"]*)"|the tasks)$`, func(choice, row string) error {
		for _, r := range w.w.FatesAsked().Rows {
			if row != "" && r.Said != row || row == "" && r.Sub != 0 {
				continue
			}
			for i, said := range r.Choices {
				if said == choice {
					at := r.Picks[i]
					w.w = w.w.clicked(Point{X: at.X + at.W/2, Y: at.Y + at.H/2})
					return nil
				}
			}
		}
		return fmt.Errorf("no row %q offers %q", row, choice)
	})
	sc.Then(`^the deletion is confirmed by "([^"]*)"$`, func(said string) error {
		offers := w.w.Offers()
		if offers[0] != said || w.w.OfferLook(0, said) != buttonDanger {
			return fmt.Errorf("the deletion offers %v", offers)
		}
		return nil
	})
	sc.When(`^I confirm the deletion$`, func() error {
		at := w.w.offersAlong(w.w.panel())[0]
		w.w = w.w.clicked(Point{X: at.X + at.W/2, Y: at.Y + at.H/2})
		return nil
	})
	sc.Then(`^"([^"]*)" is drawn as a reference saying "([^"]*)"$`, w.drawnAsAReference)
	sc.Then(`^the head of "([^"]*)" counts (\d+)$`, func(column string, n int) error {
		here, err := w.cardsOf(column)
		if err != nil {
			return err
		}
		if len(here) != n {
			return fmt.Errorf("the head of %q counts %d, want %d", column, len(here), n)
		}
		return nil
	})
}

func readAs(what string, got, want []string) error {
	if !slices.Equal(got, want) {
		return fmt.Errorf("%s read %q, want %q", what, got, want)
	}
	return nil
}

func lineTitles(lines []TaskLine) []string {
	var out []string
	for _, line := range lines {
		out = append(out, line.Card.Title)
	}
	return out
}

func (w *world) slotOf(title string) (Box, int, error) {
	for i, b := range w.w.Tree() {
		if b.Goal.Title == title {
			return GoalAt(w.w.Layout(), i, w.w.goals), i, nil
		}
	}
	return Box{}, 0, fmt.Errorf("no objective titled %q is listed", title)
}

func (w *world) drawnUnder(child, parent string) error {
	under, at, err := w.slotOf(child)
	if err != nil {
		return err
	}
	over, from, err := w.slotOf(parent)
	if err != nil {
		return err
	}
	if at <= from || ItemOf(under, 1).X != ItemOf(over, 0).X+indent {
		return fmt.Errorf("%q stands at %d, %+v, not indented under %q at %d", child, at, under, parent, from)
	}
	for _, d := range partsDrawn(w.w) {
		if strings.HasSuffix(d.Part, "connector") && d.At.X+d.At.W == ItemOf(under, 1).X && d.At.Y == under.Y {
			return nil
		}
	}
	return fmt.Errorf("nothing joins %q to %q", child, parent)
}

func (w *world) detailTexts() []string {
	d := w.w.Detail()
	var said []string
	for _, one := range textDrawn(w.w) {
		if d.Box.In(Point{X: one.box.X, Y: one.box.Y}) {
			said = append(said, one.said)
		}
	}
	return said
}

func (w *world) detailHeaded(title string) error {
	if got := w.w.Detail().Goal.Title; got != title {
		return fmt.Errorf("the detail shows %q", got)
	}
	heading := w.w.Detail().Heading
	for _, one := range textDrawn(w.w) {
		if one.said == title && one.box.Y == heading.Y && one.box.X == heading.X {
			return nil
		}
	}
	return fmt.Errorf("the detail's heading does not read %q", title)
}

func (w *world) detailSays(said string) error {
	texts := w.detailTexts()
	if all := strings.Join(texts, " "); !strings.Contains(all, said) {
		return fmt.Errorf("the detail says %q, and nothing of %q", texts, said)
	}
	return nil
}

func (w *world) tileWarns(title, warn string) error {
	for _, t := range w.w.Detail().Tiles {
		if t.Tile.Title != title {
			continue
		}
		if t.Tile.Warn != warn {
			return fmt.Errorf("the tile of %q warns %q", title, t.Tile.Warn)
		}
		for _, one := range textDrawn(w.w) {
			if one.said == warn && t.At.In(Point{X: one.box.X, Y: one.box.Y}) {
				return nil
			}
		}
		return fmt.Errorf("the tile of %q does not write %q", title, warn)
	}
	return fmt.Errorf("no tile for %q", title)
}

func (w *world) cardServes(title, served string) error {
	for _, p := range w.w.Placed() {
		if p.Card.Title != title {
			continue
		}
		row := CardRowOf(p.Card, PartsOf(p.Slot).Number, w.w.measured(), w.w.now, w.w.pages)
		if got := Serves(p.Card); got != served || row.ServesAt.W == 0 {
			return fmt.Errorf("the card %q says it serves %q at %+v", title, got, row.ServesAt)
		}
		for _, d := range partsDrawn(w.w) {
			if strings.HasSuffix(d.Part, "card/facts/serves") && d.At == row.MarkAt {
				return nil
			}
		}
		return fmt.Errorf("the card %q draws no mark before %q", title, served)
	}
	return fmt.Errorf("no card titled %q", title)
}

func (w *world) drawnAsAReference(title, why string) error {
	for _, p := range w.w.References() {
		if p.Card.Title != title {
			continue
		}
		if p.Why() != why {
			return fmt.Errorf("%q says %q", title, p.Why())
		}
		for _, d := range partsDrawn(w.w) {
			if strings.HasSuffix(d.Part, "card/dashed") && d.At == PartsOf(p.Slot).Panel {
				return nil
			}
		}
		return fmt.Errorf("%q is not drawn dashed", title)
	}
	return fmt.Errorf("%q is no reference", title)
}

func (w *world) drawnSomewhere(said string) error {
	for _, one := range textDrawn(w.w) {
		if strings.Contains(one.said, said) {
			return nil
		}
	}
	return fmt.Errorf("nothing on screen says %q", said)
}
