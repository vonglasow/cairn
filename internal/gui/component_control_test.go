package gui

import (
	"image/color"
	"slices"
	"testing"
)

func TestASegmentedControlIsOneTrackItsChosenSegmentInsetAllRound(t *testing.T) {
	s := Dark()
	segments := []Box{{X: 12, Y: 10, W: 60, H: 30}, {X: 72, Y: 10, W: 60, H: 30}, {X: 132, Y: 10, W: 60, H: 30}}
	track := Box{X: 10, Y: 10, W: 184, H: 30}
	seg := Segmented{Said: []string{"Day", "Week", "Month"}, At: segments, On: 1}
	shot, drawn := drawnAlone(t, 210, 50, func(k Kit) { seg.Draw(k, track) })
	if got := pixel(shot, track.X+2, track.Y+2); got == s.Colour.Surface.Control {
		t.Error("the track is rounded no more than a control's corners")
	}
	if got := pixel(shot, track.X+track.W/2-30, track.Y+track.H/2); got != s.Colour.Surface.Control {
		t.Errorf("the track stands on %v", got)
	}
	on := segments[1]
	hair := s.Space.Hair
	if pixel(shot, on.X+on.W/2, on.Y+hair-1) != s.Colour.Surface.Control || pixel(shot, on.X+hair-1, on.Y+on.H/2) != s.Colour.Surface.Control {
		t.Error("the chosen segment is not inset by a hair at its top and at its side")
	}
	if got := pixel(shot, on.X+hair+s.Radius.Control, on.Y+hair); got != s.Colour.Surface.ControlLit {
		t.Errorf("the chosen segment is lit %v, not %v", got, s.Colour.Surface.ControlLit)
	}
	if lettered(shot, on, s.Colour.Ink.Text) == 0 || lettered(shot, segments[0], s.Colour.Ink.Faint) == 0 || lettered(shot, segments[0], s.Colour.Ink.Text) != 0 {
		t.Error("the chosen segment is not written in the reading ink and the others faint")
	}
	if want := []string{"segmented", "segmented/segment", "segmented/segment", "segmented/segment"}; !slices.Equal(parts(drawn), want) || drawn[0].At != track || drawn[2].At != on {
		t.Errorf("the control recorded %+v", drawn)
	}
	counted, _ := drawnAlone(t, 210, 50, func(k Kit) {
		Segmented{Said: []string{"todo", "done"}, Counts: []string{"12", "3"}, At: segments[:2], Small: true}.Draw(k, track)
	})
	if lettered(counted, Box{X: segments[0].X + segments[0].W/2, Y: segments[0].Y, W: segments[0].W / 2, H: segments[0].H}, s.Colour.Ink.Low) == 0 {
		t.Error("a segment's count is not written after its word, low")
	}
}

func TestATabIsGroundedOnlyWhenItIsOpen(t *testing.T) {
	s := Dark()
	at := []Box{{X: 10, Y: 10, W: 140, H: 34}, {X: 150, Y: 10, W: 140, H: 34}}
	shot, drawn := drawnAlone(t, 300, 60, func(k Kit) {
		Tabs{Said: []string{"Tasks", "Time"}, Counts: []string{"38", "0"}, At: at, On: 0}.Draw(k, Box{X: 10, Y: 10, W: 280, H: 34})
	})
	if got := pixel(shot, at[0].X+at[0].W-4, at[0].Y+at[0].H/2); got != s.Colour.Surface.Control {
		t.Errorf("the open tab stands on %v", got)
	}
	if got := pixel(shot, at[1].X+at[1].W-4, at[1].Y+at[1].H/2); got != s.Colour.Surface.Back {
		t.Errorf("a closed tab stands on %v", got)
	}
	if lettered(shot, TabName(at[0]), s.Colour.Ink.Text) == 0 || lettered(shot, TabName(at[1]), s.Colour.Ink.Faint) == 0 || lettered(shot, TabName(at[1]), s.Colour.Ink.Low) == 0 {
		t.Error("the tabs do not write their names, open in the reading ink, and their counts low")
	}
	if want := []string{"tabs", "tabs/tab", "tabs/tab"}; !slices.Equal(parts(drawn), want) || drawn[1].At != at[0] {
		t.Errorf("the tabs recorded %+v", drawn)
	}
}

func TestARowIsGroundedByWhatItIsAndWritesInTheReadingInk(t *testing.T) {
	s := Dark()
	row := Box{X: 10, Y: 10, W: 200, H: 28}
	for _, c := range []struct {
		row         Row
		ground, ink color.NRGBA
	}{
		{Row{Said: "Backlog"}, s.Colour.Surface.Back, s.Colour.Ink.Text},
		{Row{Said: "Backlog", Picked: true}, s.Colour.Surface.Control, s.Colour.Ink.Text},
		{Row{Said: "Backlog", Taken: true}, s.Colour.Surface.Tint, s.Colour.Ink.Text},
		{Row{Said: "Backlog", Taken: true, Picked: true}, s.Colour.Surface.Control, s.Colour.Ink.Text},
		{Row{Said: "Delete", Danger: true}, s.Colour.Surface.Back, s.Colour.State.Wrong},
	} {
		shot, drawn := drawnAlone(t, 230, 50, func(k Kit) { c.row.Draw(k, row) })
		if got := pixel(shot, row.X+row.W-4, row.Y+row.H/2); got != c.ground {
			t.Errorf("%+v stands on %v, not %v", c.row, got, c.ground)
		}
		if lettered(shot, Inside(row), c.ink) == 0 {
			t.Errorf("%+v is not written in %v", c.row, c.ink)
		}
		if want := []Drawing{{Part: "row", At: row, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
			t.Errorf("%+v recorded %+v", c.row, drawn)
		}
	}
}

func TestARowGroundsTheChipAndThePillInsideIt(t *testing.T) {
	s := Dark()
	row := Box{X: 10, Y: 10, W: 200, H: 28}
	for _, picked := range []bool{false, true} {
		var chip, pill Box
		shot, drawn := drawnAlone(t, 230, 50, func(k Kit) {
			r := Row{Picked: picked}
			r.Draw(k, row)
			chip = r.Chip(k, KeyChip{Key: "esc"}, Box{X: row.X + 150, Y: row.Y, H: row.H})
			pill = r.Pill(k, Pill{Said: "deploy"}, Box{X: row.X + 60, Y: row.Y + 6, W: 60, H: pillHigh})
		})
		chipGround, pillGround := s.Colour.Surface.Control, s.Colour.Surface.Control
		if picked {
			chipGround, pillGround = s.Colour.Surface.ChipOnControl, s.Colour.Surface.Raised
		}
		if got := pixel(shot, chip.X+chip.W/2, chip.Y+1); got != chipGround {
			t.Errorf("a key on a row picked %v stands on %v, not %v", picked, got, chipGround)
		}
		if got := pixel(shot, pill.X+pill.W/2, pill.Y+1); got != pillGround {
			t.Errorf("a pill on a row picked %v stands on %v, not %v", picked, got, pillGround)
		}
		if want := []string{"row", "row/key", "row/pill"}; !slices.Equal(parts(drawn), want) {
			t.Errorf("the row recorded %v", parts(drawn))
		}
	}
}

func TestAFieldIsOneFrameLitWithItsHaloWhileWrittenInto(t *testing.T) {
	s := Dark()
	frame := Box{X: 20, Y: 20, W: 160, H: 30}
	rest, drawn := drawnAlone(t, 200, 70, func(k Kit) { Field{}.Draw(k, frame) })
	if got := pixel(rest, frame.X+frame.W/2, frame.Y); got != s.Colour.Line.Rule {
		t.Errorf("a field at rest is edged in %v", got)
	}
	if got := pixel(rest, frame.X+frame.W/2, frame.Y+frame.H/2); got != s.Colour.Surface.Field {
		t.Errorf("a field stands on %v", got)
	}
	if pixel(rest, frame.X, frame.Y) == s.Colour.Line.Rule {
		t.Error("a field is square")
	}
	if got := pixel(rest, frame.X+frame.W/2, frame.Y-2); got != s.Colour.Surface.Back {
		t.Errorf("a field at rest has a halo: %v", got)
	}
	if want := []Drawing{{Part: "field", At: frame, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the field recorded %+v", drawn)
	}
	lit, _ := drawnAlone(t, 200, 70, func(k Kit) { Field{Lit: true}.Draw(k, frame) })
	if got := pixel(lit, frame.X+frame.W/2, frame.Y); got != s.Colour.Accent.Base {
		t.Errorf("a lit field is edged in %v", got)
	}
	if got := pixel(lit, frame.X+frame.W/2, frame.Y-2); got == s.Colour.Surface.Back || got.B <= got.R {
		t.Errorf("a lit field has no halo: %v", got)
	}
}

func TestACheckIsABoxEmptyDoneOrTaken(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: checkSide, H: checkSide}
	empty, drawn := drawnAlone(t, 40, 40, func(k Kit) { Check{}.Draw(k, box) })
	if got := pixel(empty, box.X+box.W/2, box.Y); got != s.Colour.Line.Edge {
		t.Errorf("an empty box is edged in %v", got)
	}
	if got := pixel(empty, box.X+box.W/2, box.Y+box.H/2); got != s.Colour.Surface.Back {
		t.Errorf("an empty box is filled with %v", got)
	}
	if pixel(empty, box.X, box.Y) == s.Colour.Line.Edge {
		t.Error("a check box is square")
	}
	if want := []Drawing{{Part: "check", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the check recorded %+v", drawn)
	}
	for _, c := range []struct {
		check Check
		fill  color.NRGBA
	}{{Check{Done: true}, s.Colour.Ink.Faint}, {Check{Taken: true}, s.Colour.Accent.Base}} {
		shot, _ := drawnAlone(t, 40, 40, func(k Kit) { c.check.Draw(k, box) })
		if painted(shot, box, c.fill) < box.W*box.H/3 {
			t.Errorf("%+v is not filled with %v", c.check, c.fill)
		}
		if pixel(shot, box.X, box.Y) == c.fill {
			t.Errorf("%+v is square", c.check)
		}
	}
}

func TestProgressIsATrackOfThreeAndWhatIsDoneOnIt(t *testing.T) {
	s := Dark()
	at := Box{X: 10, Y: 10, W: 200, H: lineHigh}
	shot, drawn := drawnAlone(t, 230, 40, func(k Kit) { Progress{Done: 1, Total: 2, Said: "1/2"}.Draw(k, at) })
	mid := at.Y + at.H/2
	if pixel(shot, at.X+20, mid-1) != s.Colour.State.Done || pixel(shot, at.X+20, mid-3) == s.Colour.State.Done {
		t.Error("what is done is not three high")
	}
	if painted(shot, Box{X: at.X, Y: mid - 1, W: at.W, H: 1}, s.Colour.Surface.Control) == 0 {
		t.Error("what is left is not drawn on the track")
	}
	if lettered(shot, Box{X: at.X + at.W - 30, Y: at.Y, W: 30, H: at.H}, s.Colour.Ink.Low) == 0 {
		t.Error("the count is not written at the right")
	}
	if want := []Drawing{{Part: "progress", At: at, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the progress recorded %+v", drawn)
	}
	bare, _ := drawnAlone(t, 230, 40, func(k Kit) { Progress{Done: 1, Total: 1}.Draw(k, at) })
	if pixel(bare, at.X+at.W-4, mid-1) != s.Colour.State.Done {
		t.Error("a progress saying nothing does not run the whole width")
	}
	none, _ := drawnAlone(t, 230, 40, func(k Kit) { Progress{Total: 3}.Draw(k, at) })
	if painted(none, at, s.Colour.State.Done) != 0 {
		t.Error("nothing done is drawn done")
	}
}

func TestARingIsChosenOrTargetAtItsElementsRadius(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 80, H: 40}
	for _, c := range []struct {
		ring Ring
		ink  color.NRGBA
	}{{Ring{Radius: s.Radius.Card}, s.Colour.Line.Selected}, {Ring{Target: true, Radius: s.Radius.Row}, s.Colour.Accent.Base}} {
		shot, drawn := drawnAlone(t, 100, 60, func(k Kit) { c.ring.Draw(k, box) })
		if got := pixel(shot, box.X+box.W/2, box.Y); got != c.ink {
			t.Errorf("%+v is drawn in %v, not %v", c.ring, got, c.ink)
		}
		if got := pixel(shot, box.X+box.W/2, box.Y+1); got != s.Colour.Surface.Back {
			t.Errorf("%+v is thicker than a line, or fills what it rings: %v", c.ring, got)
		}
		if pixel(shot, box.X, box.Y) == c.ink {
			t.Errorf("%+v is square", c.ring)
		}
		if want := []Drawing{{Part: "ring", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
			t.Errorf("%+v recorded %+v", c.ring, drawn)
		}
	}
	today, drawn := drawnAlone(t, 100, 60, func(k Kit) { TodayMark{Radius: s.Radius.Today}.Draw(k, box) })
	if pixel(today, box.X+box.W/2, box.Y) != s.Colour.Accent.Base || pixel(today, box.X+box.W/2, box.Y+box.H/2) != s.Colour.Surface.Back {
		t.Error("today is not a ring in the accent")
	}
	if !slices.Equal(parts(drawn), []string{"today"}) {
		t.Errorf("today recorded %v", parts(drawn))
	}
}

func TestAHeadIsFaintAtFiveHundredItsCountLowAtTheRight(t *testing.T) {
	s := Dark()
	at := Box{X: 10, Y: 10, W: 200, H: lineHigh}
	shot, drawn := drawnAlone(t, 230, 40, func(k Kit) { Head{Said: "Recent", Count: "5"}.Draw(k, at) })
	if lettered(shot, Box{X: at.X, Y: at.Y, W: 60, H: at.H}, s.Colour.Ink.Faint) == 0 {
		t.Error("the head is not written faint")
	}
	if lettered(shot, Box{X: at.X + at.W - 20, Y: at.Y, W: 20, H: at.H}, s.Colour.Ink.Low) == 0 {
		t.Error("the count is not written low at the right")
	}
	if want := []Drawing{{Part: "head", At: at, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the head recorded %+v", drawn)
	}
}

func TestAScrollbarIsATrackAndAThumbWhereTheLinesAre(t *testing.T) {
	s := Dark()
	track := Box{X: 10, Y: 10, W: 10, H: 100}
	shot, drawn := drawnAlone(t, 40, 130, func(k Kit) {
		fill(k.gtx, Box{W: 40, H: 130}, s.Colour.Surface.Raised)
		Scrollbar{Lines: 20, Rows: 5, At: 15}.Draw(k, track)
	})
	thumb := Thumb(track, 20, 5, 15)
	if pixel(shot, track.X+5, track.Y+2) != s.Colour.Surface.Back || pixel(shot, thumb.X+5, thumb.Y+thumb.H/2) != s.Colour.Line.Edge {
		t.Error("the scrollbar is not a track of the back with a thumb of an edge")
	}
	if want := []Drawing{{Part: "scrollbar", At: track, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the scrollbar recorded %+v", drawn)
	}
}

func TestASwatchIsItsColourRingedWhenPicked(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 18, H: 18}
	shot, drawn := drawnAlone(t, 40, 40, func(k Kit) { Swatch{Colour: "mauve", Picked: true}.Draw(k, box) })
	if pixel(shot, box.X+9, box.Y+9) != s.Colour.Labels["mauve"] || pixel(shot, box.X+9, box.Y-1) != s.Colour.Ink.Text {
		t.Error("a picked swatch is not its colour ringed in the reading ink")
	}
	if !slices.Equal(parts(drawn), []string{"swatch"}) {
		t.Errorf("the swatch recorded %v", parts(drawn))
	}
	plain, _ := drawnAlone(t, 40, 40, func(k Kit) { Swatch{Colour: "mauve"}.Draw(k, box) })
	if pixel(plain, box.X+9, box.Y-1) != s.Colour.Surface.Back {
		t.Error("a swatch not picked is ringed")
	}
}

func TestADayCellIsPickedFilledTodayRingedOrPlain(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 40, H: 30}
	picked, drawn := drawnAlone(t, 60, 50, func(k Kit) { DayCell{Said: "9", Picked: true, Today: true}.Draw(k, box) })
	if pixel(picked, box.X+box.W/2, box.Y+1) != s.Colour.Accent.Base || lettered(picked, box, s.Colour.Ink.OnAccent) == 0 {
		t.Error("a picked day is not filled with the accent and written on it")
	}
	if !slices.Equal(parts(drawn), []string{"day"}) {
		t.Errorf("the day recorded %v", parts(drawn))
	}
	today, drawn := drawnAlone(t, 60, 50, func(k Kit) { DayCell{Said: "9", Today: true, Raised: true}.Draw(k, box) })
	if pixel(today, box.X+box.W/2, box.Y) != s.Colour.Accent.Base || pixel(today, box.X+3, box.Y+box.H/2) != s.Colour.Surface.Raised {
		t.Error("today is not a raised day ringed in the accent")
	}
	if !slices.Equal(parts(drawn), []string{"day", "day/today"}) {
		t.Errorf("today recorded %v", parts(drawn))
	}
	faded, _ := drawnAlone(t, 60, 50, func(k Kit) { DayCell{Said: "30", Faded: true}.Draw(k, box) })
	if lettered(faded, box, s.Colour.Ink.Faint) == 0 || pixel(faded, box.X+3, box.Y+box.H/2) != s.Colour.Surface.Back {
		t.Error("a day of another month is not written faint on nothing")
	}
}

func TestAWorkspacePillIsAControlRingedWhileItsSwitcherIsOpen(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 128, H: 34}
	shot, drawn := drawnAlone(t, 150, 60, func(k Kit) { WorkspacePill{Name: "work", Colour: "green", Open: true}.Draw(k, box) })
	if pixel(shot, box.X+box.W/2, box.Y) != s.Colour.Accent.Base || pixel(shot, box.X+box.W/2, box.Y+2) != s.Colour.Surface.Control {
		t.Error("an open pill is not a control ringed in the accent")
	}
	if want := []string{"workspace", "workspace/dot", "workspace/chevron", "workspace/ring"}; !slices.Equal(parts(drawn), want) {
		t.Errorf("the pill recorded %v", parts(drawn))
	}
	shut, _ := drawnAlone(t, 150, 60, func(k Kit) { WorkspacePill{Name: "…/scratch", Loose: true}.Draw(k, box) })
	if pixel(shut, box.X+box.W/2, box.Y) == s.Colour.Accent.Base {
		t.Error("a pill whose switcher is shut is ringed")
	}
}

func TestATileIsARoundedLineOfAListGhostedWhenPicked(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 200, H: 26}
	for _, c := range []struct {
		tile Tile
		want color.NRGBA
	}{{Tile{}, s.Colour.Surface.Raised}, {Tile{Picked: true}, s.Colour.Surface.Ghost}} {
		shot, drawn := drawnAlone(t, 230, 50, func(k Kit) { c.tile.Draw(k, box) })
		if pixel(shot, box.X+box.W/2, box.Y+box.H/2) != c.want || pixel(shot, box.X, box.Y) == c.want {
			t.Errorf("%+v is not a rounded tile of %v", c.tile, c.want)
		}
		if !slices.Equal(parts(drawn), []string{"tile"}) {
			t.Errorf("%+v recorded %v", c.tile, parts(drawn))
		}
	}
}

func TestTheCommandBarIsFramedAndItsHintsLitUnderThePointer(t *testing.T) {
	s := Dark()
	bar := Box{X: 10, Y: 10, W: 300, H: 32}
	shot, drawn := drawnAlone(t, 330, 60, func(k Kit) {
		Bar{}.Draw(k, bar)
		Hint{Key: "n", Word: "new", Lit: true}.Draw(k, Box{X: 100, Y: bar.Y, W: 80, H: bar.H})
		Hint{Key: "e", Word: "edit", Faint: true}.Draw(k, Box{X: 200, Y: bar.Y, W: 80, H: bar.H})
	})
	if pixel(shot, bar.X+bar.W/2, bar.Y) != s.Colour.Line.Rule || pixel(shot, bar.X+20, bar.Y+bar.H/2) != s.Colour.Surface.Raised {
		t.Error("the bar is not a raised ground edged with a rule")
	}
	if pixel(shot, 100+2, bar.Y+bar.H/2) != s.Colour.Surface.Hover {
		t.Error("a lit hint is not on the hover's ground")
	}
	if lettered(shot, Box{X: 200, Y: bar.Y, W: 80, H: bar.H}, s.Colour.Ink.Faint) == 0 {
		t.Error("a faint hint is not written faint")
	}
	if want := []string{"bar", "hint", "hint/key", "hint", "hint/key"}; !slices.Equal(parts(drawn), want) {
		t.Errorf("the bar recorded %v", parts(drawn))
	}
}

func TestTheSmallPartsAreDrawnInTheirBoxAndRecorded(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 14, H: 14}
	ink := s.Colour.Ink.Faint
	for _, c := range []struct {
		part string
		draw func(k Kit) Box
		ink  color.NRGBA
	}{
		{"cross", func(k Kit) Box { return Cross{Ink: ink}.Draw(k, box) }, ink},
		{"chevron", func(k Kit) Box { return Chevron{Ink: ink}.Draw(k, box) }, ink},
		{"calendar", func(k Kit) Box { return CalendarMark{Ink: ink}.Draw(k, box) }, ink},
		{"grip", func(k Kit) Box { return Grip{Ink: ink}.Draw(k, box) }, ink},
		{"focus", func(k Kit) Box { return FocusDot{}.Draw(k, box) }, s.Colour.Accent.Base},
		{"copy", func(k Kit) Box { return CopyMark{}.Draw(k, box) }, s.Colour.Ink.Low},
		{"menu", func(k Kit) Box { return MenuMark{}.Draw(k, box) }, s.Colour.Accent.Base},
		{"struck", func(k Kit) Box { return Struck{}.Draw(k, box) }, s.Colour.Ink.Faint},
		{"landing", func(k Kit) Box { return Landing{}.Draw(k, box) }, s.Colour.Accent.Base},
		{"spent", func(k Kit) Box { return Spent{}.Draw(k, box) }, s.Colour.Accent.Base},
		{"backdrop", func(k Kit) Box { return Backdrop{}.Draw(k, box) }, s.Colour.Surface.Back},
		{"info", func(k Kit) Box { return InfoBox{}.Draw(k, box) }, s.Colour.Surface.Card},
	} {
		var got Box
		shot, drawn := drawnAlone(t, 40, 40, func(k Kit) {
			fill(k.gtx, Box{W: 40, H: 40}, s.Colour.Surface.Raised)
			got = c.draw(k)
		})
		if got != box {
			t.Errorf("the %s gives back %+v", c.part, got)
		}
		if lettered(shot, box, c.ink) == 0 {
			t.Errorf("the %s is not drawn in %v", c.part, c.ink)
		}
		if painted(shot, Box{X: 0, Y: 0, W: 40, H: box.Y - 1}, c.ink) != 0 {
			t.Errorf("the %s is drawn outside its box", c.part)
		}
		if want := []Drawing{{Part: c.part, At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
			t.Errorf("the %s recorded %+v", c.part, drawn)
		}
	}
}

func TestAStoneIsAShadedShapeRecordedInItsBounds(t *testing.T) {
	s := Dark()
	shape := []Point{{X: 10, Y: 10}, {X: 30, Y: 12}, {X: 28, Y: 30}, {X: 12, Y: 28}}
	shot, drawn := drawnAlone(t, 40, 40, func(k Kit) { Stone{Shape: shape, Shade: s.stoneOf(0)}.Draw(k, Box{}) })
	if inked(shot, Box{X: 15, Y: 15, W: 10, H: 10}, s.Colour.Surface.Back) == 0 {
		t.Error("the stone is not drawn")
	}
	if want := []Drawing{{Part: "stone", At: Box{X: 10, Y: 10, W: 20, H: 20}, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the stone recorded %+v", drawn)
	}
}

func TestAFilterPillIsDrawnByItsKind(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: 120, H: 22}
	for _, c := range []struct {
		pill   FilterPill
		ground color.NRGBA
		parts  []string
	}{
		{FilterPill{Kind: PillFilter, Name: "label", Said: "label feature"}, s.Colour.Surface.Tint, []string{"filter", "filter/cross"}},
		{FilterPill{Kind: PillFilter, Said: "#8 Home automation", Serves: true}, s.Colour.Surface.Tint, []string{"filter", "filter/serves", "filter/cross"}},
		{FilterPill{Kind: PillFolded, Said: "3 filters"}, s.Colour.Surface.Tint, []string{"filter"}},
		{FilterPill{Kind: PillClear, Said: "clear"}, s.Colour.Surface.Back, []string{"filter"}},
		{FilterPill{Kind: PillAdd, Said: "+ filter"}, s.Colour.Surface.Back, []string{"filter", "filter/dashed"}},
	} {
		shot, drawn := drawnAlone(t, 140, 40, func(k Kit) { c.pill.Draw(k, box) })
		if got := pixel(shot, box.X+box.W/2, box.Y+box.H-3); got != c.ground && c.pill.Kind != PillAdd {
			t.Errorf("%+v stands on %v", c.pill.Kind, got)
		}
		if !slices.Equal(parts(drawn), c.parts) {
			t.Errorf("%+v recorded %v", c.pill.Kind, parts(drawn))
		}
	}
}

func TestAScrollbarWithNothingToScrollIsNotDrawn(t *testing.T) {
	s := Dark()
	track := Box{X: 10, Y: 10, W: 10, H: 100}
	shot, drawn := drawnAlone(t, 40, 130, func(k Kit) {
		fill(k.gtx, Box{W: 40, H: 130}, s.Colour.Surface.Raised)
		Scrollbar{Lines: 3, Rows: 5}.Draw(k, track)
	})
	if pixel(shot, track.X+5, track.Y+track.H/2) != s.Colour.Surface.Raised {
		t.Error("a scrollbar with nothing to scroll drew a track or a thumb")
	}
	if len(drawn) != 0 {
		t.Errorf("a scrollbar with nothing to scroll recorded %+v", drawn)
	}
}
