package gui

import (
	"image/color"
	"strconv"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
)

// Connector is the line joining a sub-objective to its parent in a tree: down its left side, and an elbow into it; the last under its parent stops at its elbow, rounded.
type Connector struct {
	Last  bool
	Elbow int
}

// Draw draws the connector down the left of at, its elbow Elbow down from at's top, or in its middle when Elbow is nought.
func (c Connector) Draw(k Kit, at Box) Box {
	k.put("connector", at)
	x, top := float32(at.X)+0.5, float32(at.Y)
	elbow := float32(at.Y+c.Elbow) + 0.5
	if c.Elbow == 0 {
		elbow = float32(at.Y+at.H/2) + 0.5
	}
	right, foot := float32(at.X+at.W), float32(at.Y+at.H)
	var lines clip.Path
	lines.Begin(k.gtx.Ops)
	lines.MoveTo(f32.Pt(x, top))
	if c.Last {
		r := float32(k.skin.Radius.Chip)
		lines.LineTo(f32.Pt(x, elbow-r))
		lines.QuadTo(f32.Pt(x, elbow), f32.Pt(x+r, elbow))
	} else {
		lines.LineTo(f32.Pt(x, foot))
		lines.MoveTo(f32.Pt(x, elbow))
	}
	lines.LineTo(f32.Pt(right, elbow))
	paint.FillShape(k.gtx.Ops, k.skin.Colour.Line.Edge, clip.Stroke{Path: lines.End(), Width: float32(k.skin.Size.Rule)}.Op())
	return at
}

// ListItem is one objective or one series in the list beside a detail: two lines, its standing's mark, its number — a series' behind its loop — its title and its type or its state on the first, its progress and its counts, or what it says and a word at the right, on the second; moved right and joined to its parent below level one; no ground at rest, the hover's under the pointer, a card's and a ring when chosen.
type ListItem struct {
	Number, Title, Standing, Priority string
	Chip, ChipColour, Counts          string
	Says, Aside, Badge, BadgeStanding string
	Done, Total, Depth                int
	Last, Chosen, Hovered, Loop       bool
}

// Draw draws the item in its slot at, the gap under it left empty.
func (i ListItem) Draw(k Kit, at Box) Box {
	if i.Depth > 0 {
		_, first, _ := ItemLine(ItemOf(at, 0))
		Connector{Last: i.Last, Elbow: first.Y + first.H/2 - at.Y}.Draw(k, JoinOf(at, i.Depth))
	}
	item := k.put("item", ItemOf(at, i.Depth))
	in := k.inside("item")
	switch {
	case i.Chosen:
		rounded(k.gtx, item, k.skin.Radius.Card, k.skin.Colour.Surface.Card)
		Ring{Radius: k.skin.Radius.Card}.Draw(in, item)
	case i.Hovered:
		rounded(k.gtx, item, k.skin.Radius.Card, k.skin.Colour.Surface.Hover)
	}
	Stripe{Priority: i.Priority}.Draw(in, Box{X: item.X, Y: item.Y + edgeInset, W: edgeWide, H: item.H - 2*edgeInset})
	mark, first, second := ItemLine(item)
	if i.Standing != "" {
		StandingMark{State: i.Standing}.Draw(in, mark)
	}
	m := k.measured()
	end := first.X + first.W
	if i.Chip != "" {
		chip := Box{X: end - ends(m, i.Chip, small) - 2*pillPad, Y: first.Y + (first.H-pillHigh)/2, W: ends(m, i.Chip, small) + 2*pillPad, H: pillHigh}
		LabelChip{Label: i.Chip, Colour: i.ChipColour}.Draw(in, chip)
		end = chip.X - pillGap
	}
	if i.Badge != "" {
		pill := Box{X: end - ends(m, i.Badge, small) - 2*pillPad, Y: first.Y + (first.H-pillHigh)/2, W: ends(m, i.Badge, small) + 2*pillPad, H: pillHigh}
		Pill{Said: i.Badge, Tone: ToneStanding, Standing: i.BadgeStanding}.Draw(in, pill)
		end = pill.X - pillGap
	}
	numbered := monoWide(i.Number, small)
	if n, err := strconv.Atoi(i.Number); i.Loop && err == nil {
		numbered = SeriesNumberWide(n, small)
		SeriesNumber{Number: n, Ink: k.skin.Colour.Accent.Base}.Draw(in, Box{X: first.X, Y: first.Y, W: numbered, H: first.H})
	} else {
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: first.X, Y: first.Y, W: numbered, H: first.H}, i.Number, small, k.skin.Type.Mono, text.Start)
	}
	title := Box{X: first.X + numbered + markSpace, Y: first.Y, W: max(0, end-first.X-numbered-markSpace), H: first.H}
	sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Text, title, fits(k.gtx, k.th, i.Title, title.W, reads), reads, font.Medium, 1, text.Start)
	if i.Total > 0 {
		Progress{Done: i.Done, Total: i.Total, Said: i.Counts}.Draw(in, second)
	}
	if i.Says != "" || i.Aside != "" {
		aside := min(second.W, ends(m, i.Aside, small)+2)
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: second.X + second.W - aside, Y: second.Y, W: aside, H: second.H}, i.Aside, small, font.Font{}, text.End)
		says := Box{X: second.X, Y: second.Y, W: max(0, second.W-aside-padding), H: second.H}
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, says, fits(k.gtx, k.th, i.Says, says.W, small), small)
	}
	return item
}

// StandingMark is where an objective stands, drawn rather than written: a triangle when active, two bars on hold, a tick reached, a dash dropped, each in its standing's ink.
type StandingMark struct{ State string }

// Draw draws the mark in at.
func (s StandingMark) Draw(k Kit, at Box) Box {
	k.put("standing", at)
	ink := k.skin.Colour.Ink.Faint
	if inked, ok := k.skin.Colour.Standing[s.State]; ok {
		ink = inked.Ink
	}
	x, y, w, h := float32(at.X), float32(at.Y), float32(at.W), float32(at.H)
	var shape clip.Path
	shape.Begin(k.gtx.Ops)
	switch s.State {
	case "active":
		shape.MoveTo(f32.Pt(x+w*0.25, y+h*0.15))
		shape.LineTo(f32.Pt(x+w*0.85, y+h*0.5))
		shape.LineTo(f32.Pt(x+w*0.25, y+h*0.85))
		shape.Close()
		paint.FillShape(k.gtx.Ops, ink, clip.Outline{Path: shape.End()}.Op())
		return at
	case "on-hold":
		for _, bar := range []float32{0.25, 0.6} {
			shape.MoveTo(f32.Pt(x+w*bar, y+h*0.2))
			shape.LineTo(f32.Pt(x+w*bar+2, y+h*0.2))
			shape.LineTo(f32.Pt(x+w*bar+2, y+h*0.8))
			shape.LineTo(f32.Pt(x+w*bar, y+h*0.8))
			shape.Close()
		}
		paint.FillShape(k.gtx.Ops, ink, clip.Outline{Path: shape.End()}.Op())
		return at
	case "reached":
		shape.MoveTo(f32.Pt(x+w*0.15, y+h*0.55))
		shape.LineTo(f32.Pt(x+w*0.4, y+h*0.8))
		shape.LineTo(f32.Pt(x+w*0.85, y+h*0.25))
	default:
		shape.MoveTo(f32.Pt(x+w*0.2, y+h*0.5))
		shape.LineTo(f32.Pt(x+w*0.8, y+h*0.5))
	}
	paint.FillShape(k.gtx.Ops, ink, clip.Stroke{Path: shape.End(), Width: k.skin.Size.Stroke}.Op())
	return at
}

// ServesMark is the diamond in front of what a task serves, drawn rather than written so no face has to carry it.
type ServesMark struct{ Ink color.NRGBA }

// Draw draws the diamond in at.
func (s ServesMark) Draw(k Kit, at Box) Box {
	k.put("serves", at)
	x, y := float32(at.X)+float32(at.W)/2, float32(at.Y)+float32(at.H)/2
	reach := float32(min(at.W, at.H))/2 - 1
	var shape clip.Path
	shape.Begin(k.gtx.Ops)
	shape.MoveTo(f32.Pt(x, y-reach))
	shape.LineTo(f32.Pt(x+reach, y))
	shape.LineTo(f32.Pt(x, y+reach))
	shape.LineTo(f32.Pt(x-reach, y))
	shape.Close()
	paint.FillShape(k.gtx.Ops, s.Ink, clip.Stroke{Path: shape.End(), Width: k.skin.Size.Stroke}.Op())
	return at
}

// Fact is one name and its value, with a faint line under the value when there is more to say.
type Fact struct {
	Name, Value, Aside string
	Late               bool
}

// Facts is a list of names and their values: the name faint in a column of its own, the value in the ink beside it.
type Facts struct{ Lines []Fact }

// High is how much room the facts take.
func (f Facts) High() int {
	high := 0
	for _, one := range f.Lines {
		high += lineHigh
		if one.Aside != "" {
			high += lineHigh
		}
	}
	return high
}

// Draw writes the facts down at.
func (f Facts) Draw(k Kit, at Box) Box {
	k.put("facts", at)
	y := at.Y
	for _, one := range f.Lines {
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: at.X, Y: y, W: factName, H: lineHigh}, one.Name, small)
		value := Box{X: at.X + factName, Y: y, W: max(0, at.W-factName), H: lineHigh}
		ink := k.skin.Colour.Ink.Text
		if one.Late {
			ink = k.skin.Colour.State.Late
		}
		say(k.gtx, k.th, ink, value, fits(k.gtx, k.th, one.Value, value.W, reads), reads)
		y += lineHigh
		if one.Aside != "" {
			say(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: value.X, Y: y, W: value.W, H: lineHigh}, fits(k.gtx, k.th, one.Aside, value.W, small), small)
			y += lineHigh
		}
	}
	return at
}

// Figure is one number and the word that says what it counts, late when it warns.
type Figure struct {
	Figure, Word string
	Late         bool
}

// Figures is a row of numbers at the figures' size, each followed by its word.
type Figures struct{ Said []Figure }

// Draw writes the figures along at.
func (f Figures) Draw(k Kit, at Box) Box {
	k.put("figures", at)
	m := k.measured()
	x := at.X
	for _, one := range f.Said {
		if x+ends(m, one.Figure, k.skin.Type.Figure) > at.X+at.W {
			break
		}
		ink := k.skin.Colour.Ink.Text
		if one.Late {
			ink = k.skin.Colour.State.Late
		}
		wide := ends(m, one.Figure, k.skin.Type.Figure) + 2
		sayStyled(k.gtx, k.th, ink, Box{X: x, Y: at.Y, W: max(0, min(wide, at.X+at.W-x)), H: at.H}, one.Figure, k.skin.Type.Figure, font.SemiBold, 1, text.Start)
		x += wide + inset
		said := min(ends(m, one.Word, small)+2, at.X+at.W-x)
		if said <= 0 {
			break
		}
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: x, Y: at.Y + at.H - lineHigh - 2, W: said, H: lineHigh}, one.Word, small, font.Font{}, text.Start)
		x += said + 2*padding
	}
	return at
}

// Banner is a short boxed message inside a panel or a detail: what was refused, in the wrong's tones, or what to watch, in on hold's; then what to do, on a line of its own.
type Banner struct {
	Said, Then string
	Warn       bool
}

// BannerHigh is how tall a banner saying that is at a width: its words wrapped, and its second line under them.
func BannerHigh(m Measure, said, then string, wide int) int {
	high := 2*bannerPad + len(m.Wrapped(said, wide-2*bannerPad, reads).Said)*lineHigh
	if then != "" {
		high += len(m.Wrapped(then, wide-2*bannerPad, small).Said) * lineHigh
	}
	return high
}

const bannerPad = 8

// Draw draws the banner over at and writes it inside.
func (b Banner) Draw(k Kit, at Box) Box {
	k.put("banner", at)
	ground, ink := k.skin.Colour.State.Danger, k.skin.Colour.State.Wrong
	if b.Warn {
		ground, ink = k.skin.Colour.Standing["on-hold"].Back, k.skin.Colour.Standing["on-hold"].Ink
	}
	rounded(k.gtx, at, k.skin.Radius.Card, ground)
	m := k.measured()
	inside := Box{X: at.X + bannerPad, Y: at.Y + bannerPad, W: at.W - 2*bannerPad}
	said := m.Wrapped(b.Said, inside.W, reads).Said
	for i, line := range said {
		sayStyled(k.gtx, k.th, ink, Box{X: inside.X, Y: inside.Y + i*lineHigh, W: inside.W, H: lineHigh}, line, reads, font.SemiBold, 1, text.Start)
	}
	y := inside.Y + len(said)*lineHigh
	for i, line := range m.Wrapped(b.Then, inside.W, small).Said {
		say(k.gtx, k.th, ink, Box{X: inside.X, Y: y + i*lineHigh, W: inside.W, H: lineHigh}, line, small)
	}
	return at
}
