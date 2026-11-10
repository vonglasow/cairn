package gui

import "cairn/internal/engine"

func drawRule(k Kit, w Window) {
	picker := w.RulePicked()
	Panel{}.Draw(k, picker.Box)
	k = k.inside("panel")
	for i, line := range picker.Lines {
		if line.Said != "" {
			ink := k.skin.Colour.Ink.Faint
			if captionOf(picker.Lines, w.ruling.line) == i {
				ink = k.skin.Colour.Accent.Lit
			}
			say(k.gtx, k.th, ink, line.Caption, line.Said, small)
		}
		switch {
		case line.Ticked != nil:
			for j, pick := range line.Picks {
				day := ButtonIn(pick)
				Button{Said: line.Choices[j], Look: lookOf(line.Ticked[j])}.Draw(k, day)
				if i == w.ruling.line && j == w.ruling.day {
					Ring{Radius: k.skin.Radius.Control}.Draw(k.inside("button"), day)
				}
			}
		case line.Compact:
			Segmented{Said: []string{line.Choices[line.On]}, At: line.Picks, Small: true, Back: line.Back, Next: line.Next}.Draw(k, line.Track)
		default:
			Segmented{Said: line.Choices, At: line.Picks, On: line.On, Small: true}.Draw(k, line.Track)
		}
		if i == w.ruling.line && line.Ticked == nil {
			Ring{Radius: k.skin.Radius.Control}.Draw(k.inside("segmented"), line.Track)
		}
	}
	drawPanelOffers(k, picker.Box, w)
}

func drawReadBack(k Kit, w Window) {
	read, ok := RepeatsReadAt(w.formBox(), w.form.Fields())
	if !ok {
		return
	}
	for _, field := range w.form.Fields() {
		if field.Arg != engine.RepeatsField().Arg {
			continue
		}
		said := engine.RuleReadBack(field.Value, w.now)
		top := Box{X: read.X, Y: read.Y, W: read.W, H: lineHigh}
		under := Box{X: read.X, Y: read.Y + lineHigh, W: read.W, H: lineHigh}
		if said.Wrong != "" {
			say(k.gtx, k.th, k.skin.Colour.State.Wrong, top, fits(k.gtx, k.th, said.Wrong, top.W, small), small)
			continue
		}
		if said.Says == "" {
			continue
		}
		LoopMark{Ink: k.skin.Colour.Accent.Base}.Draw(k, Box{X: top.X, Y: top.Y + (top.H-loopSide)/2, W: loopSide, H: loopSide})
		top.X, top.W = top.X+loopSide+inset, top.W-loopSide-inset
		say(k.gtx, k.th, k.skin.Colour.Ink.OnTint, top, fits(k.gtx, k.th, said.Says, top.W, small), small)
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, under, fits(k.gtx, k.th, said.Next, under.W, small), small)
	}
}

func captionOf(lines []RuleLine, line int) int {
	for line > 0 && lines[line].Said == "" {
		line--
	}
	return line
}
