package gui

import (
	"image/color"
	"strconv"

	"gioui.org/font"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"

	"cairn/internal/engine"
)

// Card is a task's card, or an objective's, in one of its states: at rest, under the pointer, chosen, dimmed, dragged, or a reference to one the board leaves out.
type Card struct {
	Face                                        engine.Card
	Chosen, Hovered, Dimmed, Dragged, Reference bool
	Standing                                    string
	Pages                                       engine.Pages
	Colour                                      func(label string) string
}

// Draw draws the card in the slot at, which holds the gap under it.
func (c Card) Draw(k Kit, at Box) Box {
	if c.Dragged {
		return c.dragged(k, at)
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
	for i, line := range TitleLines(c.Face.Title, first, rest, wideAt(reads)) {
		sayStyled(k.gtx, k.th, c.writing(k.skin), Box{X: parts.Title.X, Y: parts.Title.Y + i*lineHigh, W: parts.Title.W, H: lineHigh}, line, reads, font.Medium, 1, text.Start)
	}
	if mark, ok := FocusMark(c.Face, parts.Title); ok {
		FocusDot{}.Draw(in, mark)
	}

	mono := k.skin.Type.Mono
	laid := CardRowOf(c.Face, parts.Number, m, k.gtx.Now, c.Pages)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, parts.Number, laid.Number, small, mono, text.Start)
	if laid.Facts != "" {
		ink := k.skin.Colour.Ink.Low
		if laid.Late {
			ink = k.skin.Colour.State.Late
		}
		sayIn(k.gtx, k.th, ink, laid.FactsAt, laid.Facts, small, mono, text.End)
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
