package gui

import (
	"fmt"
	"image"
	"slices"
	"strings"
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
	sc.Then(`^the card "([^"]*)" on the time page writes its number whole, its facts clear of it$`, w.numberClearOfFacts)
	sc.Then(`^the form's grip is clear of its buttons$`, func() error {
		panel := w.w.formPanel()
		for _, along := range w.w.offersAlong(panel) {
			if cross(ButtonIn(along), FormGrip(panel)) {
				return fmt.Errorf("the grip %+v lies over the button %+v", FormGrip(panel), ButtonIn(along))
			}
		}
		return nil
	})
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
	sc.Then(`^the date of "([^"]*)" is ringed in the accent and not filled$`, w.todayRinged)
	sc.Then(`^the search of the labels is lit with a halo$`, w.labelSearchHaloed)
	sc.Then(`^the value "([^"]*)" of the filter stands in a box filled with the accent$`, func(value string) error {
		return w.valueBoxed(value, true)
	})
	sc.Then(`^the value "([^"]*)" of the filter stands in an empty box$`, func(value string) error {
		return w.valueBoxed(value, false)
	})
	sc.Then(`^every line of the command bar is written in the reading ink$`, w.paletteLinesRead)
	sc.Then(`^the groups of the command bar are headed in sentence case$`, w.paletteGroupsInSentenceCase)
	sc.Then(`^the pill of the workspace is ringed in the accent$`, w.placeRinged)
	sc.Then(`^the steps stand in a field framed as every field is$`, w.stepsInAField)
}

func (w *world) todayRinged(date string) error {
	day, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return err
	}
	shot, err := drawnInAStep(w.w)
	if err != nil {
		return err
	}
	s := Dark()
	for i, one := range w.w.WeekDays() {
		if !one.Equal(day) {
			continue
		}
		pill := TodayPill(WeekHeadAt(WeekDayAt(w.w.Layout(), i)), w.w.measured())
		if got := pixel(shot, pill.X+pill.W/2, pill.Y); got != s.Colour.Accent.Base {
			return fmt.Errorf("the date of %s is edged in %v, not in the accent", date, got)
		}
		if got := pixel(shot, pill.X+2, pill.Y+pill.H/2); got == s.Colour.Accent.Base {
			return fmt.Errorf("the date of %s is filled with the accent", date)
		}
		return nil
	}
	return fmt.Errorf("the week shown holds no %s", date)
}

func (w *world) labelSearchHaloed() error {
	if !w.w.Seeking() {
		return fmt.Errorf("the labels are not being searched")
	}
	shot, err := drawnInAStep(w.w)
	if err != nil {
		return err
	}
	row := w.w.LabelSearchAt()
	got := pixel(shot, row.X+row.W/2, row.Y+1-Dark().Space.Halo+1)
	if got == Dark().Colour.Surface.Raised || got.B <= got.R {
		return fmt.Errorf("the search of the labels has no halo above it: %v", got)
	}
	return nil
}

func (w *world) valueBoxed(value string, taken bool) error {
	values, open := w.w.FilterList()
	if !open {
		return fmt.Errorf("no filter lists its values")
	}
	f := w.w.Filters()[w.w.FilterOn()]
	at := slices.Index(f.Shown, value)
	if at < 0 || at >= len(values) {
		return fmt.Errorf("the filter %q lists no %q: %v", f.Name, value, f.Shown)
	}
	shot, err := drawnInAStep(w.w)
	if err != nil {
		return err
	}
	s := Dark()
	box := CheckAt(Inside(values[at]))
	accent := painted(shot, box, s.Colour.Accent.Base)
	switch {
	case taken && accent < box.W*box.H/3:
		return fmt.Errorf("the value %q stands in no box filled with the accent", value)
	case !taken && (accent != 0 || painted(shot, box, s.Colour.Line.Edge) == 0):
		return fmt.Errorf("the value %q does not stand in an empty box", value)
	}
	return nil
}

func (w *world) paletteLinesRead() error {
	shown := w.w
	shown.palette.opened = shown.now.Add(-time.Hour)
	shot, err := drawnInAStep(shown)
	if err != nil {
		return err
	}
	read := 0
	for _, line := range shown.PaletteLines() {
		if line.Row < 0 || line.Box.Y+line.Box.H > shown.height {
			continue
		}
		words := Box{X: line.Box.X, Y: line.Box.Y, W: line.Box.W / 3, H: line.Box.H}
		if lettered(shot, words, Dark().Colour.Ink.Text) == 0 {
			return fmt.Errorf("the line %q is not written in the reading ink", shown.PaletteRows()[line.Row].Says)
		}
		read++
	}
	if read < 2 {
		return fmt.Errorf("the command bar shows %d line", read)
	}
	return nil
}

func (w *world) paletteGroupsInSentenceCase() error {
	shown := w.w
	shown.palette.opened = shown.now.Add(-time.Hour)
	var said []string
	for _, one := range textDrawn(shown) {
		said = append(said, one.said)
	}
	heads := 0
	for _, line := range shown.PaletteLines() {
		if line.Row >= 0 {
			continue
		}
		if !slices.Contains(said, line.Group) || (strings.ToUpper(line.Group) != line.Group && slices.Contains(said, strings.ToUpper(line.Group))) {
			return fmt.Errorf("the group %q is not headed as it is named: %v", line.Group, said)
		}
		heads++
	}
	if heads == 0 {
		return fmt.Errorf("the command bar heads no group")
	}
	return nil
}

func (w *world) placeRinged() error {
	if w.w.Screen() != OnWorkspaces {
		return fmt.Errorf("the switcher is not open")
	}
	shot, err := drawnInAStep(w.w)
	if err != nil {
		return err
	}
	place := w.w.Layout().Place
	if got := pixel(shot, place.X+place.W/2, place.Y); got != Dark().Colour.Accent.Base {
		return fmt.Errorf("the pill of the workspace is edged in %v", got)
	}
	return nil
}

func (w *world) stepsInAField() error {
	fields := w.w.Form().Fields()
	for i, field := range fields {
		if field.Name != "steps" {
			continue
		}
		_, value, rows := FormPlace(w.w.formBox(), fields, i)
		frame := Framed(value, rows)
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		s := Dark()
		if got := pixel(shot, frame.X, frame.Y); got == s.Colour.Line.Rule || got == s.Colour.Line.Edge {
			return fmt.Errorf("the steps stand in a square box")
		}
		if got := pixel(shot, frame.X+frame.W/2, frame.Y); got != s.Colour.Line.Rule {
			return fmt.Errorf("the steps' frame is edged in %v, not a field's rule", got)
		}
		if got := pixel(shot, frame.X+4, frame.Y+frame.H-5); got != s.Colour.Surface.Field {
			return fmt.Errorf("the steps stand on %v, not on a field's ground", got)
		}
		return nil
	}
	return fmt.Errorf("the form has no steps")
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

func (w *world) numberClearOfFacts(title string) error {
	l := w.w.Layout()
	for i := range 7 {
		for j, one := range w.w.DayShown(i) {
			if w.w.DayCard(one).Title == title {
				return numberApart(partsDrawn(w.w), PartsOf(DayCardAt(WeekDayAt(l, i), j)).Panel)
			}
		}
	}
	return fmt.Errorf("no card titled %q is on the time page", title)
}

func numberApart(drawn []Drawing, card Box) error {
	var number, facts []Drawing
	in := false
	for _, d := range drawn {
		switch {
		case strings.HasSuffix(d.Part, "card"):
			in = d.At == card
		case in && strings.HasSuffix(d.Part, "card/number"):
			number = append(number, d)
		case in && strings.HasSuffix(d.Part, "card/facts"):
			facts = append(facts, d)
		}
	}
	if len(number) != 1 || number[0].At.W < monoWide("#1", small) {
		return fmt.Errorf("the card at %+v writes its number in %+v", card, number)
	}
	for _, f := range facts {
		if cross(f.At, number[0].At) {
			return fmt.Errorf("the card's facts %+v cover its number %+v", f.At, number[0].At)
		}
	}
	return nil
}
