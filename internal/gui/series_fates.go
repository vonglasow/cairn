package gui

import (
	"strconv"

	"cairn/internal/engine"
)

// SeriesFateRow is one choice deleting a series offers: what it says, what it adds under, and where it stands.
type SeriesFateRow struct {
	Said, Then string
	At         Box
}

// SeriesFates is the question deleting a series asks: what becomes of the tasks it made, its tasks kept unless told.
type SeriesFates struct {
	Number             int
	Box, Heading, Said Box
	Asks               string
	Rows               []SeriesFateRow
	On                 int
}

// DeleteTheSeriesSays is the offer that deletes the series once its tasks' fate is chosen.
const DeleteTheSeriesSays = "Delete the series"

const (
	seriesFatesWide = 560
	seriesFateHigh  = 2*lineHigh + 2*inset + between
)

// SeriesFatesAsked lays the question out over the window.
func (w Window) SeriesFatesAsked() SeriesFates {
	s, _ := w.snap.SeriesOf(w.fateOf)
	made := engine.MadeStill(s)
	f := SeriesFates{Number: s.Number, On: w.seriesFate}
	f.Asks = SeriesMadeSays(s) + " Its rule and template are deleted; its trace says so."
	open := "none is open"
	if len(s.Open) > 0 {
		open = numbersSaid(s.Open)
	}
	f.Rows = []SeriesFateRow{
		{Said: "Keep its tasks, as ordinary tasks", Then: "They lose the link to the series; notes, time, steps and trace stay."},
		{Said: "Keep the tasks done, delete the open one (" + open + ")", Then: "The history stays; nothing more lands on the board."},
		{Said: "Delete all " + strconv.Itoa(len(made)) + " tasks too", Then: "Their notes and time go with them. Asks once more, naming the count."},
	}
	wide := min(seriesFatesWide, w.width-2*margin)
	high := FirstRow(Box{}) + 2*lineHigh + padding + len(f.Rows)*seriesFateHigh + padding + barHigh + padding
	f.Box = Box{X: (w.width - wide) / 2, Y: max(margin, (w.height-high)/3), W: wide, H: high}
	f.Heading = FormHeading(f.Box)
	y := FirstRow(f.Box)
	f.Said = RowIn(f.Box, y, 2*lineHigh)
	y += 2*lineHigh + padding
	for i := range f.Rows {
		f.Rows[i].At = RowIn(f.Box, y, seriesFateHigh-between)
		y += seriesFateHigh
	}
	return f
}

// SeriesMadeSays is what deleting a series says it made: how many tasks still here, how many open, done and skipped.
func SeriesMadeSays(s engine.Series) string {
	made := engine.MadeStill(s)
	said := "It made " + strconv.Itoa(len(made)) + " task" + plural(len(made)) + " still here: "
	said += strconv.Itoa(len(s.Open)) + " open, " + strconv.Itoa(s.Done) + " done, " + strconv.Itoa(s.Skipped) + " skipped."
	return said
}

func (w Window) onSeriesFates(name string) Window {
	switch name {
	case "down", "j":
		w.seriesFate = min(len(engine.SeriesFates)-1, w.seriesFate+1)
	case "up", "k":
		w.seriesFate = max(0, w.seriesFate-1)
	}
	return w
}

func (w Window) clickingSeriesFates(p Point) Window {
	for i, row := range w.SeriesFatesAsked().Rows {
		if row.At.In(p) {
			w.seriesFate = i
		}
	}
	return w
}

func (w Window) confirmingSeriesFate() Window {
	if engine.SeriesFates[w.seriesFate] == "delete-tasks" {
		w.screen, w.question = OnAsking, aboutDeletingSeriesTasks
		return w
	}
	return w.deletingSeries()
}

func (w Window) deletingSeries() Window {
	w.screen, w.question = OnTheBoard, aboutDeleting
	in := engine.Agreed(engine.Numbered(w.fateOf))
	in.Given[engine.SeriesFates[w.seriesFate]] = "true"
	w.seriesOn = 0
	return w.asking(engine.SeriesDelete, in)
}

func (w Window) deletingAllSays() string {
	s, _ := w.snap.SeriesOf(w.fateOf)
	return "Delete all " + strconv.Itoa(len(engine.MadeStill(s))) + " tasks too?"
}

func drawSeriesFates(k Kit, w Window) {
	f := w.SeriesFatesAsked()
	Veil{}.Draw(k, Box{W: w.width, H: w.height})
	k = k.inside("veil")
	Panel{}.Draw(k, f.Box)
	in := k.inside("panel")
	s, _ := w.snap.SeriesOf(f.Number)
	Heading{Said: "Delete the series " + s.Title + "?", Number: strconv.Itoa(s.Number), Loop: true}.Draw(in, f.Heading)
	sayOver(k.gtx, k.th, k.skin.Colour.Ink.Faint, f.Said, f.Asks, reads, 2)
	for i, row := range f.Rows {
		r := Row{Picked: i == f.On}
		r.Draw(in, row.At)
		r.Check(in, Check{Taken: i == f.On}, CheckAt(Box{X: row.At.X + inset, Y: row.At.Y + inset, W: row.At.W, H: lineHigh}))
		text := Box{X: row.At.X + inset + checkSide + padding, Y: row.At.Y + inset, W: max(0, row.At.W-2*inset-checkSide-padding), H: lineHigh}
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, text, fits(k.gtx, k.th, row.Said, text.W, reads), reads)
		then := Box{X: text.X, Y: text.Y + lineHigh, W: text.W, H: lineHigh}
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, then, fits(k.gtx, k.th, row.Then, then.W, small), small)
	}
	drawPanelOffers(in, f.Box, w)
}
