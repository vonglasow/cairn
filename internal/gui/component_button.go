package gui

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/text"
)

// Button is a word to click in one of its looks; keyed, the word then the key that does it on a chip of its own.
type Button struct {
	Said, Key string
	Look      buttonLook
	Keyed     bool
	Small     bool
}

// Draw draws the button in at.
func (b Button) Draw(k Kit, at Box) Box {
	k.put("button", at)
	back, ink, chip := k.skin.button(b.Look)
	if b.Look == buttonGhost {
		ringed(k, at, k.skin.Radius.Control, k.skin.Colour.Line.Edge)
	} else {
		rounded(k.gtx, at, k.skin.Radius.Control, back)
	}
	if b.Small {
		sayStyled(k.gtx, k.th, ink, Box{X: at.X, Y: at.Y + (at.H-lineHigh)/2, W: at.W, H: lineHigh}, b.Said, small, font.Medium, 1, text.Middle)
		return at
	}
	if !b.Keyed {
		sayStyled(k.gtx, k.th, ink, ButtonSaid(at), b.Said, reads, font.Medium, 1, text.Middle)
		return at
	}
	said, key := ButtonParts(at, k.measured(), b.Said, b.Key)
	sayStyled(k.gtx, k.th, ink, said, b.Said, reads, font.Medium, 1, text.Start)
	if b.Key != "" {
		KeyChip{Key: b.Key}.grounded(k.inside("button"), key, chip, ink)
	}
	return at
}

type buttonLook int

func litLook(look buttonLook) buttonLook {
	switch look {
	case buttonOn:
		return buttonOnLit
	case buttonDanger:
		return buttonDangerLit
	}
	return buttonLit
}

func lookOf(on bool) buttonLook {
	if on {
		return buttonOn
	}
	return buttonPlain
}

const (
	buttonPlain buttonLook = iota
	buttonOn
	buttonLit
	buttonFaint
	buttonDanger
	buttonOnLit
	buttonDangerLit
	buttonGhost
)

func (s Skin) button(look buttonLook) (back, ink, chip color.NRGBA) {
	c := s.Colour
	switch look {
	case buttonOn:
		return c.Accent.Base, c.Ink.OnAccent, c.Accent.Chip
	case buttonOnLit:
		return c.Accent.Lit, c.Ink.OnAccent, c.Accent.LitChip
	case buttonLit:
		return c.Surface.ControlLit, c.Ink.Text, c.Surface.ChipOnControlLit
	case buttonFaint:
		return c.Surface.Control, c.Ink.Faint, c.Surface.ChipOnControl
	case buttonDanger:
		return c.State.Danger, c.State.Wrong, c.State.DangerChip
	case buttonDangerLit:
		return c.State.DangerLit, c.State.Wrong, c.State.DangerLitChip
	}
	return c.Surface.Control, c.Ink.Text, c.Surface.ChipOnControl
}
