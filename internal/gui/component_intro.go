package gui

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

func seenAs(ink color.NRGBA, seen float64) color.NRGBA {
	ink.A = uint8(float64(ink.A) * clamp01(seen))
	return ink
}

// Draw draws the ring the ripple sends out; at is the window, which it may run past.
func (v Wave) Draw(k Kit, at Box) Box {
	k.put("wave", at)
	box := image.Rect(int(v.X-v.Radius), int(v.Y-v.Radius), int(v.X+v.Radius), int(v.Y+v.Radius))
	glow, line := k.skin.Size.Stroke*float32(k.skin.Space.Halo)*2, k.skin.Size.Stroke
	paint.FillShape(k.gtx.Ops, seenAs(k.skin.Colour.Accent.Glow, v.Seen), clip.Stroke{Path: clip.Ellipse(box).Path(k.gtx.Ops), Width: glow}.Op())
	paint.FillShape(k.gtx.Ops, seenAs(k.skin.Colour.Accent.Base, v.Seen), clip.Stroke{Path: clip.Ellipse(box).Path(k.gtx.Ops), Width: line}.Op())
	return at
}

// Draw draws the band of light across at, the window, leaning as it says, brightest in its middle.
func (s Sweep) Draw(k Kit, at Box) Box {
	k.put("sweep", at)
	x, half := float32(s.X), float32(s.Wide)/2
	turned := op.Affine(f32.Affine2D{}.Rotate(f32.Pt(x, float32(at.Y+at.H/2)), float32(s.Tilt*math.Pi/180))).Push(k.gtx.Ops)
	defer turned.Pop()
	peak := seenAs(k.skin.Colour.Accent.Lit, s.Seen*0.14)
	clear := seenAs(k.skin.Colour.Accent.Base, 0)
	top, bottom := at.Y-at.H/4, at.Y+at.H+at.H/4
	for _, side := range [][2]float32{{x - half, x}, {x + half, x}} {
		left, right := min(side[0], side[1]), max(side[0], side[1])
		area := clip.Rect{Min: image.Pt(int(left), top), Max: image.Pt(int(right), bottom)}.Push(k.gtx.Ops)
		paint.LinearGradientOp{Stop1: f32.Pt(side[0], 0), Color1: clear, Stop2: f32.Pt(side[1], 0), Color2: peak}.Add(k.gtx.Ops)
		paint.PaintOp{}.Add(k.gtx.Ops)
		area.Pop()
	}
	return at
}

// Grains is the dust a board condenses out of, each speck drawn in the colour of the part of its card it lands on.
type Grains []Grain

// Draw draws every speck; at is the window.
func (g Grains) Draw(k Kit, at Box) Box {
	k.put("grains", at)
	for _, one := range g {
		ink := k.skin.Colour.Line.Edge
		switch one.Part {
		case GrainStripe:
			if by, ok := k.skin.Colour.Priority[one.Priority]; ok {
				ink = by
			}
		case GrainTitle:
			ink = k.skin.Colour.Ink.Text
			if one.Focus {
				ink = k.skin.Colour.Ink.OnTint
			}
		}
		speck := clip.Rect{Min: image.Pt(int(one.X), int(one.Y)), Max: image.Pt(int(math.Ceil(one.X+one.Side)), int(math.Ceil(one.Y+one.Side)))}
		paint.FillShape(k.gtx.Ops, seenAs(ink, one.Seen), speck.Op())
	}
	return at
}

// Draw draws the focus mark speaking at at: its dot grown and seen as far as it has come, and the ring it sends out.
func (b Pulse) Draw(k Kit, at Box) Box {
	glow := GlowOf(at)
	if b.Ring > 0 {
		r := float64(glow.Dot) * b.Ring
		x, y := float64(glow.X), float64(glow.Y)
		box := image.Rect(int(x-r), int(y-r), int(math.Ceil(x+r)), int(math.Ceil(y+r)))
		paint.FillShape(k.gtx.Ops, seenAs(k.skin.Colour.Accent.Base, b.RingSeen), clip.Stroke{Path: clip.Ellipse(box).Path(k.gtx.Ops), Width: k.skin.Size.Stroke}.Op())
	}
	if b.Dot <= 0 || b.DotSeen <= 0 {
		return at
	}
	seen := paint.PushOpacity(k.gtx.Ops, float32(clamp01(b.DotSeen)))
	defer seen.Pop()
	grown := op.Affine(f32.Affine2D{}.Scale(f32.Pt(glow.X, glow.Y), f32.Pt(float32(b.Dot), float32(b.Dot)))).Push(k.gtx.Ops)
	defer grown.Pop()
	return FocusDot{}.Draw(k, at)
}

// Wash is light, or shadow, laid over a card or a ground as it arrives, by so much.
type Wash struct {
	Dark, Ground bool
	By           float64
}

// Draw lays the wash over at.
func (w Wash) Draw(k Kit, at Box) Box {
	ink, radius := k.skin.Colour.Ink.Text, k.skin.Radius.Card
	if w.Dark {
		ink = k.skin.Colour.Surface.Back
	}
	if w.Ground {
		radius = k.skin.Radius.Ground
	}
	rounded(k.gtx, at, radius, seenAs(ink, w.By))
	return at
}

// Drafted is the outline a card or a ground is drawn as before it is filled, seen by so much.
type Drafted struct {
	Ground bool
	By     float64
}

// Draw draws the outline around at.
func (d Drafted) Draw(k Kit, at Box) Box {
	radius := k.skin.Radius.Card
	if d.Ground {
		radius = k.skin.Radius.Ground
	}
	ringed(k, at, radius, seenAs(k.skin.Colour.Line.Edge, d.By))
	return at
}

// Lifted is the deeper shadow a card casts while it flies, by how far it is lifted.
type Lifted struct{ By float64 }

// Draw casts the shadow under at.
func (l Lifted) Draw(k Kit, at Box) Box {
	for _, s := range k.skin.Shadow {
		down, spread := int(float64(s.Down)*(1+l.By)), int(float64(s.Spread)*(1+l.By))
		under := Box{X: at.X - spread, Y: at.Y + down - spread, W: at.W + 2*spread, H: at.H + 2*spread}
		rounded(k.gtx, under, k.skin.Radius.Card+spread, seenAs(s.Ink, l.By*2))
	}
	return at
}
