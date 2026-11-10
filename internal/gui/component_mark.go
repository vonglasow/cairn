package gui

import (
	"cairn/internal/theme"
	"image"
	"image/color"
	"strings"
	"sync"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
)

// Cross is the stroked mark that closes or takes off what it stands on.
type Cross struct{ Ink color.NRGBA }

// Draw draws the cross in the middle of at.
func (c Cross) Draw(k Kit, at Box) Box {
	k.put("cross", at)
	reach := float32(at.W) / 4
	x, y := float32(at.X)+float32(at.W)/2, float32(at.Y)+float32(at.H)/2
	var strokes clip.Path
	strokes.Begin(k.gtx.Ops)
	strokes.MoveTo(f32.Pt(x-reach, y-reach))
	strokes.LineTo(f32.Pt(x+reach, y+reach))
	strokes.MoveTo(f32.Pt(x+reach, y-reach))
	strokes.LineTo(f32.Pt(x-reach, y+reach))
	paint.FillShape(k.gtx.Ops, c.Ink, clip.Stroke{Path: strokes.End(), Width: k.skin.Size.Stroke}.Op())
	return at
}

// Chevron is the stroked mark of what opens a list under it, or, turned to a side, of what walks back (-1) or on (1).
type Chevron struct {
	Ink  color.NRGBA
	Side int
}

// Draw draws the chevron in at.
func (c Chevron) Draw(k Kit, at Box) Box {
	k.put("chevron", at)
	var strokes clip.Path
	strokes.Begin(k.gtx.Ops)
	if c.Side != 0 {
		x, y := float32(at.X)+float32(at.W)*0.3, float32(at.Y)
		w, h := float32(at.W)*0.4, float32(at.H)
		if c.Side < 0 {
			x, w = x+w, -w
		}
		strokes.MoveTo(f32.Pt(x, y+1))
		strokes.LineTo(f32.Pt(x+w, y+h/2))
		strokes.LineTo(f32.Pt(x, y+h-1))
		paint.FillShape(k.gtx.Ops, c.Ink, clip.Stroke{Path: strokes.End(), Width: k.skin.Size.Stroke}.Op())
		return at
	}
	x, y := float32(at.X), float32(at.Y)+float32(at.H)*0.3
	w, h := float32(at.W), float32(at.H)*0.4
	strokes.MoveTo(f32.Pt(x+1, y))
	strokes.LineTo(f32.Pt(x+w/2, y+h))
	strokes.LineTo(f32.Pt(x+w-1, y))
	paint.FillShape(k.gtx.Ops, c.Ink, clip.Stroke{Path: strokes.End(), Width: k.skin.Size.Stroke}.Op())
	return at
}

// CalendarMark is the mark of a field chosen on a calendar.
type CalendarMark struct{ Ink color.NRGBA }

// Draw draws the mark in at.
func (c CalendarMark) Draw(k Kit, at Box) Box {
	k.put("calendar", at)
	outline(k.gtx, Box{X: at.X, Y: at.Y + 1, W: at.W, H: at.H - 1}, c.Ink)
	fill(k.gtx, Box{X: at.X, Y: at.Y + 1, W: at.W, H: 3}, c.Ink)
	return at
}

// Grip is the stroked corner a form is resized by.
type Grip struct{ Ink color.NRGBA }

// Draw draws the grip into the bottom right corner of at.
func (g Grip) Draw(k Kit, at Box) Box {
	k.put("grip", at)
	x, y := float32(at.X+at.W), float32(at.Y+at.H)
	var strokes clip.Path
	strokes.Begin(k.gtx.Ops)
	for _, reach := range []float32{float32(at.W), float32(at.W) / 2} {
		strokes.MoveTo(f32.Pt(x-reach, y))
		strokes.LineTo(f32.Pt(x, y-reach))
	}
	paint.FillShape(k.gtx.Ops, g.Ink, clip.Stroke{Path: strokes.End(), Width: k.skin.Size.Stroke}.Op())
	return at
}

// FocusDot is the glowing dot of the card in focus: a dot of the accent in its glow.
type FocusDot struct{}

// Draw draws the dot and its glow in at.
func (FocusDot) Draw(k Kit, at Box) Box {
	k.put("focus", at)
	glow := GlowOf(at)
	disc(k.gtx, glow.X, glow.Y, glow.Halo, k.skin.Colour.Accent.Glow)
	disc(k.gtx, glow.X, glow.Y, glow.Dot, k.skin.Colour.Accent.Base)
	return at
}

// Draw draws the mark at at: two sheets, one over the other.
func (CopyMark) Draw(k Kit, at Box) Box {
	k.put("copy", at)
	back := Box{X: at.X + 3, Y: at.Y, W: at.W - 3, H: at.H - 3}
	front := Box{X: at.X, Y: at.Y + 3, W: at.W - 3, H: at.H - 3}
	outline(k.gtx, back, k.skin.Colour.Ink.Low)
	fill(k.gtx, front, k.skin.Colour.Surface.Raised)
	outline(k.gtx, front, k.skin.Colour.Ink.Low)
	return at
}

// MenuMark is the three strokes of the button opening the menu where the system has no menu bar.
type MenuMark struct{}

// Draw draws the strokes in the button at.
func (MenuMark) Draw(k Kit, at Box) Box {
	k.put("menu", at)
	for _, stroke := range MenuStrokes(at) {
		fill(k.gtx, stroke, k.skin.Colour.Accent.Base)
	}
	return at
}

// Stone is one stone of the cairn the board is built from while it opens: a shape shaded from its top to its foot, grained.
type Stone struct {
	Shape []Point
	Shade theme.Shade
}

// Draw draws the stone where its shape stands; at is not read, the shape being where it is.
func (s Stone) Draw(k Kit, _ Box) Box {
	var path clip.Path
	path.Begin(k.gtx.Ops)
	top, bottom, left, right := s.Shape[0].Y, s.Shape[0].Y, s.Shape[0].X, s.Shape[0].X
	for i, p := range s.Shape {
		at := f32.Pt(float32(p.X), float32(p.Y))
		top, bottom, left, right = min(top, p.Y), max(bottom, p.Y), min(left, p.X), max(right, p.X)
		if i == 0 {
			path.MoveTo(at)
			continue
		}
		path.LineTo(at)
	}
	path.Close()
	bounds := k.put("stone", Box{X: left, Y: top, W: right - left, H: bottom - top})
	outlined := clip.Outline{Path: path.End()}.Op().Push(k.gtx.Ops)
	paint.LinearGradientOp{
		Stop1: f32.Pt(0, float32(top)), Color1: s.Shade.Light,
		Stop2: f32.Pt(0, float32(bottom)), Color2: s.Shade.Dark,
	}.Add(k.gtx.Ops)
	paint.PaintOp{}.Add(k.gtx.Ops)
	moved := op.Offset(image.Pt(left, top)).Push(k.gtx.Ops)
	grainOp().Add(k.gtx.Ops)
	paint.PaintOp{}.Add(k.gtx.Ops)
	moved.Pop()
	outlined.Pop()
	return bounds
}

var grainOp = sync.OnceValue(func() paint.ImageOp { return paint.NewImageOp(StoneGrain()) })

// Draw draws the pill in at by its kind: a search, a filter or a view in force with the cross taking it off — a view changed on top marked by a late dot — the filters folded into one, the act clearing them, the one adding one.
func (p FilterPill) Draw(k Kit, at Box) Box {
	k.put("filter", at)
	in := k.inside("filter")
	line := Box{X: at.X, Y: at.Y + (at.H-lineHigh)/2, W: at.W, H: lineHigh}
	switch p.Kind {
	case PillSearch, PillFilter, PillView:
		p.inForce(k, in, at)
	case PillAdd:
		Dashed{Radius: k.skin.Radius.Full(at)}.Draw(in, at)
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Text, line, p.Said, small, font.Font{}, text.Middle)
	case PillFolded:
		rounded(k.gtx, at, k.skin.Radius.Full(at), k.skin.Colour.Surface.Tint)
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Text, line, p.Said, small, font.Font{}, text.Middle)
	case PillClear:
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, line, p.Said, small, font.Font{}, text.Middle)
	}
	return at
}

func (p FilterPill) inForce(k, in Kit, at Box) {
	name, value := k.skin.Colour.Ink.Faint, k.skin.Colour.Ink.OnTint
	if p.Faint {
		name, value = k.skin.Colour.Ink.Low, k.skin.Colour.Ink.Faint
	}
	rounded(k.gtx, at, k.skin.Radius.Full(at), k.skin.Colour.Surface.Tint)
	line := Box{X: at.X + pillPad, Y: at.Y + (at.H-lineHigh)/2, W: at.W - pillPad - pillCross, H: lineHigh}
	if p.Serves {
		ServesMark{Ink: value}.Draw(in, Box{X: line.X, Y: at.Y + (at.H-servesSide)/2, W: servesSide, H: servesSide})
		line.X, line.W = line.X+servesSide+servesGap, line.W-servesSide-servesGap
	}
	sayIn(k.gtx, k.th, name, line, p.Name, small, font.Font{}, text.Start)
	said := ends(k.measured(), p.Name, small) + nameGap
	if p.Changed {
		line.W -= dotSide + between
		Dot{Late: true}.Draw(in, Box{X: line.X + line.W + between, Y: at.Y + (at.H-dotSide)/2, W: dotSide, H: dotSide})
	}
	valued := strings.TrimSuffix(strings.TrimPrefix(p.Said, p.Name+" "), " "+p.Tail)
	sayIn(k.gtx, k.th, value, Box{X: line.X + said, Y: line.Y, W: line.W - said, H: line.H}, valued, small, font.Font{Weight: font.Medium}, text.Start)
	if p.Tail != "" {
		after := said + ends(k.measured(), valued, small) + nameGap
		sayIn(k.gtx, k.th, name, Box{X: line.X + after, Y: line.Y, W: max(0, line.W-after), H: line.H}, p.Tail, small, font.Font{}, text.Start)
	}
	Cross{Ink: name}.Draw(in, Box{X: at.X + at.W - pillCross, Y: at.Y + (at.H-pillCross)/2, W: pillCross - 4, H: pillCross})
}
