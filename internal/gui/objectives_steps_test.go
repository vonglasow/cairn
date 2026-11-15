package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"fmt"
	"slices"
	"strings"

	"github.com/cucumber/godog"
)

func (w *world) objectivesSteps(sc *godog.ScenarioContext) {
	w.foldsSteps(sc)
	sc.Then(`^the detail is headed "([^"]*)"$`, w.detailHeaded)
	sc.Then(`^the detail says "([^"]*)"$`, w.detailSays)
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
				w.w = w.w.clicked(geom.Point{X: at.X + at.W/2, Y: at.Y + at.H/2})
				return nil
			}
		}
		return fmt.Errorf("the narrowing bar has no segment %q", name)
	})
	sc.When(`^I rest the pointer on the detail's offer "([^"]*)"$`, func(offer string) error {
		at, err := w.detailOffer(offer)
		w.w = w.w.Hovering(geom.Point{X: at.X + 2, Y: at.Y + 2})
		return err
	})
	sc.Then(`^the detail's offer "([^"]*)" is lit under the pointer$`, func(offer string) error {
		at, err := w.detailOffer(offer)
		if err != nil {
			return err
		}
		return w.litUnderThePointer(at, at)
	})
	sc.When(`^I rest the pointer on the objective "([^"]*)"$`, func(title string) error {
		at, err := w.findGoal(title)
		w.w = w.w.Hovering(at)
		return err
	})
	sc.Then(`^the objective "([^"]*)" is lit under the pointer$`, func(title string) error {
		for i, b := range w.w.Tree() {
			if b.Goal.Title == title {
				slot := GoalAt(w.w.Layout(), i, w.w.goals)
				return w.litUnderThePointer(slot, ItemOf(slot, depthOf(b)))
			}
		}
		return fmt.Errorf("no objective titled %q is on screen", title)
	})
	sc.When(`^I rest the pointer on the tile of "([^"]*)"$`, func(title string) error {
		at, err := w.tileOf(title)
		w.w = w.w.Hovering(geom.Point{X: at.X + 2, Y: at.Y + 2})
		return err
	})
	sc.Then(`^the tile of "([^"]*)" is lit under the pointer$`, func(title string) error {
		at, err := w.tileOf(title)
		if err != nil {
			return err
		}
		return w.litUnderThePointer(at, at)
	})
	sc.Then(`^the list says under it "([^"]*)"$`, func(said string) error {
		if got := w.w.ChoosingThen(); got != said {
			return fmt.Errorf("the list says under it %q", got)
		}
		return w.drawnSomewhere(strings.Fields(said)[0])
	})
	sc.Then(`^the list says "([^"]*)"$`, func(said string) error {
		if got := w.w.ChoosingRefused(); got != said {
			return fmt.Errorf("the list says %q", got)
		}
		for _, d := range partsDrawnWith(w.theme, w.w) {
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
		if w.w.Screen != board.OnFates || w.w.OfferLook(0, offers[0]) != buttonFaint {
			return fmt.Errorf("the deletion offers %v on %v, not faint", offers, w.w.Screen)
		}
		if w.w.Key("enter").Screen != board.OnFates {
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
					w.w = w.w.clicked(geom.Point{X: at.X + at.W/2, Y: at.Y + at.H/2})
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
		w.w = w.w.clicked(geom.Point{X: at.X + at.W/2, Y: at.Y + at.H/2})
		return nil
	})
}

func readAs(what string, got, want []string) error {
	if !slices.Equal(got, want) {
		return fmt.Errorf("%s read %q, want %q", what, got, want)
	}
	return nil
}

func (w *world) detailTexts() []string {
	d := w.w.Detail()
	var said []string
	for _, one := range textDrawnWith(w.theme, w.w) {
		if d.Box.In(geom.Point{X: one.box.X, Y: one.box.Y}) {
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
	for _, one := range textDrawnWith(w.theme, w.w) {
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

func (w *world) drawnSomewhere(said string) error {
	for _, one := range textDrawnWith(w.theme, w.w) {
		if strings.Contains(one.said, said) {
			return nil
		}
	}
	return fmt.Errorf("nothing on screen says %q", said)
}

func (w *world) detailOffer(offer string) (geom.Box, error) {
	d := w.w.Detail()
	for i, one := range d.Offers {
		if one == offer && i < len(d.Buttons) {
			return d.Buttons[i], nil
		}
	}
	return geom.Box{}, fmt.Errorf("the detail offers %v, not %q", d.Offers, offer)
}

func (w *world) tileOf(title string) (geom.Box, error) {
	for _, t := range w.w.Detail().Tiles {
		if t.Tile.Title == title {
			return t.At, nil
		}
	}
	return geom.Box{}, fmt.Errorf("the detail has no tile %q", title)
}

func (w *world) litUnderThePointer(slot, drawn geom.Box) error {
	if w.w.pointed != slot {
		return fmt.Errorf("the pointer lights %+v, not %+v", w.w.pointed, slot)
	}
	lit, err := drawnInAStep(w.theme, w.w)
	if err != nil {
		return err
	}
	away, err := drawnInAStep(w.theme, w.w.Away())
	if err != nil {
		return err
	}
	x, y := drawn.X+drawn.W/2, drawn.Y+4
	if pixel(lit, x, y) == pixel(away, x, y) {
		return fmt.Errorf("%+v is drawn the same under the pointer and away from it", drawn)
	}
	return nil
}
