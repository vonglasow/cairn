package gui

import (
	"image/color"
	"strconv"
	"strings"

	"gioui.org/font"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"

	"cairn/internal/engine"
)

// Card is a task's card, or an objective's, in one of its states: at rest, under the pointer, chosen, dimmed, dragged, or a reference to one the board leaves out, saying why it is there; laid out as a tile when it carries a sub-objective.
type Card struct {
	Face                                        engine.Card
	Chosen, Hovered, Dimmed, Dragged, Reference bool
	Standing, Why                               string
	Pages                                       engine.Pages
	Colour                                      func(label string) string
	Goal                                        *GoalTile
}

// GoalTile is what a card in tile layout carries: an objective's mark, number, title and type, its progress and counts, notes on its tasks, and a warning.
type GoalTile struct {
	Number, Title, State, Priority, Chip, ChipColour string
	Counts, Warn                                     string
	Done, Total                                      int
	Notes                                            []string
}

// Draw draws the card in the slot at, which holds the gap under it.
func (c Card) Draw(k Kit, at Box) Box {
	if c.Dragged {
		return c.dragged(k, at)
	}
	if c.Goal != nil {
		return c.tile(k, at)
	}
	if c.Dimmed {
		half := paint.PushOpacity(k.gtx.Ops, 0.5)
		defer half.Pop()
	}
	parts := PartsOf(at)
	k.put("card", parts.Panel)
	in := k.inside("card")
	ground := c.ground(k.skin)
	switch {
	case c.Reference:
		Dashed{Radius: k.skin.Radius.Card}.Draw(in, parts.Panel)
	case c.Chosen:
		rounded(k.gtx, parts.Panel, k.skin.Radius.Card, ground)
		Ring{Radius: k.skin.Radius.Card}.Draw(in, parts.Panel)
	default:
		rounded(k.gtx, parts.Panel, k.skin.Radius.Card, ground)
	}
	Stripe{Priority: priorityOf(c.Face)}.Draw(in, parts.Edge)

	m := Measured(k.th)
	wideAt := func(size int) func(string) int { return func(said string) int { return ends(m, said, size) } }
	first, rest := TitleRoom(c.Face, parts.Title)
	if rule := RuleOf(c.Face); rule != "" {
		LoopMark{Ink: k.skin.Colour.Ink.Faint}.Draw(in, Box{X: parts.Title.X, Y: parts.Title.Y + (lineHigh-loopSide)/2, W: loopSide, H: loopSide})
		title := Box{X: parts.Title.X + loopSide + markSpace, Y: parts.Title.Y, W: max(0, first-loopSide-markSpace), H: lineHigh}
		sayStyled(k.gtx, k.th, c.writing(k.skin), title, fits(k.gtx, k.th, c.Face.Title, title.W, reads), reads, font.Medium, 1, text.Start)
		said := Box{X: parts.Title.X, Y: parts.Title.Y + lineHigh, W: parts.Title.W, H: lineHigh}
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, said, fits(k.gtx, k.th, rule, said.W, small), small)
	} else {
		for i, line := range TitleLines(c.Face.Title, first, rest, wideAt(reads)) {
			sayStyled(k.gtx, k.th, c.writing(k.skin), Box{X: parts.Title.X, Y: parts.Title.Y + i*lineHigh, W: parts.Title.W, H: lineHigh}, line, reads, font.Medium, 1, text.Start)
		}
	}
	if mark, ok := FocusMark(c.Face, parts.Title); ok {
		FocusDot{}.Draw(in, mark)
	}

	mono := k.skin.Type.Mono
	laid := CardRowOf(c.Face, parts.Number, m, k.gtx.Now, c.Pages)
	in.put("number", laid.NumberAt)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, parts.Number, laid.Number, small, mono, text.Start)
	if laid.Series != 0 {
		SeriesNumber{Number: laid.Series, Ink: k.skin.Colour.Accent.Base}.Draw(in, laid.SeriesAt)
	}
	if c.Reference && c.Why != "" {
		laid.Facts, laid.Late, laid.FactsAt = c.Why, true, FactsAt(parts.Number, ends(m, c.Why, small))
	}
	if laid.Facts != "" {
		in.put("facts", laid.FactsAt)
		ink := k.skin.Colour.Ink.Low
		if laid.Late {
			ink = k.skin.Colour.State.Late
		}
		front := laid.FactsAt
		if laid.ServesAt.W > 0 {
			name := Serves(c.Face)
			ServesMark{Ink: k.skin.Colour.Ink.Faint}.Draw(in.inside("facts"), laid.MarkAt)
			at := laid.MarkAt.X + servesSide + servesGap
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: at, Y: laid.FactsAt.Y, W: laid.ServesAt.X + laid.ServesAt.W - at, H: laid.FactsAt.H}, name, small, mono, text.End)
			front.W = laid.ServesAt.X - front.X
			laid.Facts = strings.TrimSuffix(laid.Facts, name)
		}
		sayIn(k.gtx, k.th, ink, front, laid.Facts, small, mono, text.End)
	}
	for i, pill := range laid.Pills {
		if i == 0 && c.Standing != "" {
			Pill{Said: laid.Shown[i], Tone: ToneStanding, Standing: c.Standing}.Draw(in, pill)
			continue
		}
		LabelChip{Label: laid.Shown[i], Colour: c.Colour(laid.Shown[i])}.Draw(in, pill)
	}
	if laid.Count != "" {
		Pill{Said: laid.Count}.Draw(in, laid.More)
	}
	if c.Face.SubTotal > 0 {
		Progress{Done: c.Face.SubDone, Total: c.Face.SubTotal, Said: strconv.Itoa(c.Face.SubDone) + "/" + strconv.Itoa(c.Face.SubTotal)}.Draw(in, parts.Steps)
	}
	return parts.Panel
}

func (c Card) tile(k Kit, at Box) Box {
	g := c.Goal
	panel := k.put("card", PartsOf(at).Panel)
	in := k.inside("card")
	rounded(k.gtx, panel, k.skin.Radius.Card, c.ground(k.skin))
	if c.Chosen {
		Ring{Radius: k.skin.Radius.Card}.Draw(in, panel)
	}
	Stripe{Priority: g.Priority}.Draw(in, PartsOf(at).Edge)
	first, track, counts, notes := TileParts(panel)
	m := k.measured()
	mark := StandingMark{State: g.State}.Draw(in, Box{X: first.X, Y: first.Y + (lineHigh-markSide)/2, W: markSide, H: markSide})
	end := first.X + first.W
	if g.Chip != "" {
		wide := ends(m, g.Chip, small) + 2*pillPad
		chip := LabelChip{Label: g.Chip, Colour: g.ChipColour}.Draw(in, Box{X: end - wide, Y: first.Y + (lineHigh-pillHigh)/2, W: wide, H: pillHigh})
		end = chip.X - pillGap
	}
	x := mark.X + mark.W + markSpace
	numbered := monoWide(g.Number, small)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: x, Y: first.Y, W: numbered, H: lineHigh}, g.Number, small, k.skin.Type.Mono, text.Start)
	title := Box{X: x + numbered + markSpace, Y: first.Y, W: max(0, end-x-numbered-markSpace), H: lineHigh}
	sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Text, title, fits(k.gtx, k.th, g.Title, title.W, reads), reads, font.Medium, 1, text.Start)
	Progress{Done: g.Done, Total: g.Total}.Draw(in, track)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, counts, g.Counts, small, k.skin.Type.Mono, text.Start)
	lines := g.Notes
	if g.Warn != "" {
		lines = append(lines[:min(len(lines), 1):min(len(lines), 1)], g.Warn)
	}
	for i, line := range lines[:min(len(lines), 2)] {
		ink := k.skin.Colour.Ink.Faint
		if line == g.Warn && g.Warn != "" {
			ink = k.skin.Colour.State.Late
		}
		say(k.gtx, k.th, ink, Box{X: notes.X, Y: notes.Y + i*lineHigh, W: notes.W, H: lineHigh}, fits(k.gtx, k.th, line, notes.W, small), small)
	}
	return panel
}

func (c Card) dragged(k Kit, at Box) Box {
	k.put("card", at)
	raised(k, at, k.skin.Radius.Card, k.skin.Colour.Surface.Ghost)
	line := PartsOf(at).Title
	say(k.gtx, k.th, k.skin.Colour.Ink.Text, line, fits(k.gtx, k.th, c.Face.Title, line.W, plain), plain)
	return at
}

func (c Card) ground(s Skin) color.NRGBA {
	switch {
	case c.Face.InFocus:
		return s.Colour.Surface.Tint
	case c.Hovered:
		return s.Colour.Surface.Hover
	}
	return s.Colour.Surface.Card
}

func (c Card) writing(s Skin) color.NRGBA {
	if c.Reference {
		return s.Colour.Ink.Faint
	}
	return s.Writing(c.Face)
}

func priorityOf(c engine.Card) string {
	if c.Archived {
		return ""
	}
	return c.Priority
}

// Stripe is a priority's colour, 5 wide: square along the edge it is attached to, or rounded all round standing free.
type Stripe struct {
	Priority string
	Free     bool
}

// Draw draws the stripe in at.
func (s Stripe) Draw(k Kit, at Box) Box {
	k.put("stripe", at)
	ink, known := k.skin.Colour.Priority[s.Priority]
	if !known {
		ink = k.skin.Colour.Line.Edge
	}
	if s.Free {
		rounded(k.gtx, at, k.skin.Radius.Stripe, ink)
		return at
	}
	attached := clip.RRect{Rect: rect(at), NE: k.skin.Radius.Stripe, SE: k.skin.Radius.Stripe}
	paint.FillShape(k.gtx.Ops, ink, attached.Op(k.gtx.Ops))
	return at
}
