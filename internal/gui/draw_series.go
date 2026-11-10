package gui

import (
	"strconv"

	"gioui.org/font"
	"gioui.org/text"

	"cairn/internal/engine"
)

func drawSeries(k Kit, l Layout, w Window) {
	list, detail := SplitOf(l)
	Ground{}.Draw(k, list)
	on := k.inside("ground")
	shown, _ := w.SeriesShown()
	in, clipped := on.within(list)
	for _, row := range w.SeriesRows() {
		if row.Head != "" {
			Head{Said: row.Head, Count: row.Count}.Draw(in, row.At)
			continue
		}
		s := row.Series
		ListItem{Number: strconv.Itoa(s.Number), Loop: true, Title: s.Title, Priority: s.Template.Priority,
			Says: s.Rule, Aside: w.SeriesWhen(s), Badge: s.State, BadgeStanding: StandingOfSeries(s),
			Chosen: s.Number == shown.Number, Hovered: row.At == w.pointed}.Draw(in, row.At)
	}
	clipped.Pop()
	if len(w.SeriesListed()) == 0 {
		Empty{Said: NoSeriesSays}.Draw(on, EmptyAt(Box{X: list.X + gap, Y: list.Y + gap, W: list.W - 2*gap, H: CardHeight}))
	}
	Ground{}.Draw(k, detail)
	drawSeriesDetail(k.inside("ground"), w, w.SeriesDetail())
}

// NoSeriesSays is what the list of series says while there is none.
const NoSeriesSays = "No series yet. n makes one."

func drawSeriesDetail(k Kit, w Window, d SeriesDetail) {
	if len(d.Offers) == 0 {
		return
	}
	s := d.Series
	Heading{Said: s.Title, Number: strconv.Itoa(s.Number), Loop: true}.Draw(k, d.Heading)
	Pill{Said: engine.StandingSaid(s), Tone: ToneStanding, Standing: StandingOfSeries(s)}.Draw(k, d.Badge)
	say(k.gtx, k.th, k.skin.Colour.Ink.OnTint, d.Sentence, fits(k.gtx, k.th, s.Says, d.Sentence.W, plain), plain)
	say(k.gtx, k.th, k.skin.Colour.Ink.Faint, d.Next, fits(k.gtx, k.th, d.NextSays, d.Next.W, small), small)
	for i, line := range d.Lines {
		at := Box{X: d.LinesAt.X, Y: d.LinesAt.Y + i*lineHigh, W: d.LinesAt.W, H: lineHigh}
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, at, fits(k.gtx, k.th, line, at.W, small), small)
	}
	if d.Template.W > 0 {
		drawSeriesBox(k, d.Template, "Template — what the next one is made from", func(in Kit, room Box) {
			d.TemplateFacts.Draw(in, room)
		})
		drawSeriesBox(k, d.SoFar, "So far", func(in Kit, room Box) {
			Figures{Said: d.Figures}.Draw(in, Box{X: room.X, Y: room.Y, W: room.W, H: FiguresHigh})
			if s.TimeSpent != "" {
				said := s.TimeSpent + " over its occurrences"
				say(in.gtx, in.th, in.skin.Colour.Ink.Faint, Box{X: room.X, Y: room.Y + FiguresHigh + inset, W: room.W, H: lineHigh}, fits(in.gtx, in.th, said, room.W, small), small)
			}
		})
		drawSeriesBox(k, d.Now, "On the board now", func(in Kit, room Box) {
			for i, line := range d.NowSays {
				ink := in.skin.Colour.Ink.Text
				if i > 0 {
					ink = in.skin.Colour.Ink.Faint
				}
				at := Box{X: room.X, Y: room.Y + i*lineHigh, W: room.W, H: lineHigh}
				say(in.gtx, in.th, ink, at, fits(in.gtx, in.th, line, at.W, small), small)
			}
		})
	}
	Head{Said: d.OccurrencesSaid}.Draw(k, d.OccurrencesHead)
	aside := min(d.OccurrencesHead.W/2, ends(w.measured(), d.OccurrencesAside, small)+2)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: d.OccurrencesHead.X + d.OccurrencesHead.W - aside, Y: d.OccurrencesHead.Y, W: aside, H: lineHigh},
		fits(k.gtx, k.th, d.OccurrencesAside, aside, small), small, font.Font{}, text.End)
	for _, line := range d.Occurrences {
		drawOccurrence(k, w, line)
	}
	Rule{}.Draw(k, d.Foot)
	if d.FootSays != "" {
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, d.FootSaid, fits(k.gtx, k.th, d.FootSays, d.FootSaid.W, small), small, k.skin.Type.Mono, text.Start)
	}
	for i, offer := range d.Offers {
		look := buttonPlain
		switch offer {
		case DeleteSeriesSays:
			look = buttonDanger
		case ResumeSeriesSays:
			look = buttonOn
		}
		if w.pointed == d.Buttons[i] {
			look = litLook(look)
		}
		Button{Said: offer, Key: w.OfferKey(offer), Keyed: true, Look: look}.Draw(k, d.Buttons[i])
	}
}

func drawSeriesBox(k Kit, at Box, head string, inside func(Kit, Box)) {
	InfoBox{}.Draw(k, at)
	in := k.inside("info")
	Head{Said: head}.Draw(in, Box{X: at.X + padding, Y: at.Y + padding/2, W: at.W - 2*padding, H: lineHigh})
	inside(in, Box{X: at.X + padding, Y: at.Y + padding/2 + lineHigh + inset, W: at.W - 2*padding, H: at.H - padding - lineHigh - inset})
}

var occurrenceColumns = []int{56, 44, 0, 110, 48}

func drawOccurrence(k Kit, w Window, line OccurrenceLine) {
	o := line.Occurrence
	tile := Tile{Picked: line.At == w.pointed}.Draw(k, SubPanel(line.At))
	row := Box{X: tile.X + padding, Y: tile.Y + (tile.H-lineHigh)/2, W: tile.W - 2*padding, H: lineHigh}
	day, number, title, is, spent := w.OccurrenceSays(o)
	x := row.X
	fixed := 0
	for _, wide := range occurrenceColumns {
		fixed += wide
	}
	flexible := max(0, row.W-fixed-4*inset)
	isInk := k.skin.Colour.Ink.Faint
	switch {
	case o.Is == "missed":
		isInk = k.skin.Colour.State.Late
	case o.Is == "done":
		isInk = k.skin.Colour.State.Done
	case !o.Archived && o.Task != 0 && o.Is != "skipped" && o.Is != "deleted":
		isInk = k.skin.Colour.Accent.Base
	}
	for i, said := range []string{day, number, title, is, spent} {
		wide := occurrenceColumns[i]
		if wide == 0 {
			wide = flexible
		}
		at := Box{X: x, Y: row.Y, W: wide, H: row.H}
		switch i {
		case 1, 4:
			along := text.Start
			if i == 4 {
				along = text.End
			}
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, at, said, small, k.skin.Type.Mono, along)
		case 2:
			ink := k.skin.Colour.Ink.Text
			if o.Task == 0 {
				ink = k.skin.Colour.State.Late
			}
			say(k.gtx, k.th, ink, at, fits(k.gtx, k.th, said, at.W, reads), reads)
		case 3:
			say(k.gtx, k.th, isInk, at, fits(k.gtx, k.th, said, at.W, small), small)
		default:
			say(k.gtx, k.th, k.skin.Colour.Ink.Faint, at, fits(k.gtx, k.th, said, at.W, small), small)
		}
		x += wide + inset
	}
}
