package gui

import (
	"image/color"

	"gioui.org/text"
)

// Ground is what a column, a lane or a day of the week stands on: a column's ground, or a raised one for a day.
type Ground struct{ Raised bool }

// Draw lays the ground over at.
func (g Ground) Draw(k Kit, at Box) Box {
	k.put("ground", at)
	ink := k.skin.Colour.Surface.Ground
	if g.Raised {
		ink = k.skin.Colour.Surface.Raised
	}
	rounded(k.gtx, at, k.skin.Radius.Ground, ink)
	return at
}

// Panel is a raised box over the board, a form's or a list's, headed when it has a head.
type Panel struct{ Head string }

// Draw raises the panel over at, its shadow under it, and writes its head.
func (p Panel) Draw(k Kit, at Box) Box {
	k.put("panel", at)
	raised(k, at, k.skin.Radius.Panel, k.skin.Colour.Surface.Raised)
	if p.Head != "" {
		Heading{Said: p.Head}.Draw(k.inside("panel"), FormHeading(at))
	}
	return at
}

// Heading is the title of what a panel or a detail shows, and the number of what it shows beside it when it has one.
type Heading struct{ Said, Number string }

// Draw writes the heading in at.
func (h Heading) Draw(k Kit, at Box) Box {
	k.put("heading", at)
	if h.Number == "" {
		sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Text, at, h.Said, HeadingSize, HeadingWeight, 1, text.Start)
		return at
	}
	numbered := monoWide(h.Number, reads)
	room := max(0, at.W-numbered-markSpace)
	said := fits(k.gtx, k.th, h.Said, room, HeadingSize)
	sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Text, Box{X: at.X, Y: at.Y, W: room, H: at.H}, said, HeadingSize, HeadingWeight, 1, text.Start)
	x := at.X + min(room, ends(measuredIn(k.th, HeadingWeight), said, HeadingSize)) + markSpace
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: x, Y: at.Y + (at.H-lineHigh)/2 + 1, W: numbered, H: lineHigh}, h.Number, reads, k.skin.Type.Mono, text.Start)
	return at
}

// Popover is a small raised box dropped over the rest: a menu, a bubble, the swatches, a tip, the switcher.
type Popover struct{}

// Draw raises the popover over at, its shadow under it.
func (Popover) Draw(k Kit, at Box) Box {
	k.put("popover", at)
	raised(k, at, k.skin.Radius.Ground, k.skin.Colour.Surface.Raised)
	return at
}

// Veil dims what is under a question, or, light, under the command bar.
type Veil struct{ Light bool }

// Draw lays the veil over at.
func (v Veil) Draw(k Kit, at Box) Box {
	k.put("veil", at)
	fill(k.gtx, at, v.ink(k.skin))
	return at
}

func (v Veil) ink(s Skin) color.NRGBA {
	if v.Light {
		return s.Colour.Surface.VeilLight
	}
	return s.Colour.Surface.Veil
}

// Rule is a line across what it is drawn on, between what that shows and what it offers.
type Rule struct{}

// Draw draws the rule along the top of at, in the line of the layer it is on.
func (Rule) Draw(k Kit, at Box) Box {
	line := k.skin.Colour.Line.Rule
	if k.layer == LayerPanel || k.layer == LayerPopover {
		line = k.skin.Colour.Line.Raised
	}
	rule := k.put("rule", Box{X: at.X, Y: at.Y, W: at.W, H: k.skin.Size.Rule})
	fill(k.gtx, rule, line)
	return rule
}
