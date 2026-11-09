package gui

import (
	"strconv"

	"gioui.org/font"
	"gioui.org/text"

	"cairn/internal/engine"
)

func drawGoals(k Kit, l Layout, w Window) {
	list, detail := SplitOf(l)
	Ground{}.Draw(k, list)
	on := k.inside("ground")
	tree := w.Tree()
	Head{Said: "Objectives", Count: strconv.Itoa(len(tree))}.Draw(on, GoalsHead(l))
	in, clipped := on.within(GoalsRun(l))
	for at, b := range tree {
		g := b.Goal
		item := ListItem{Number: "#" + strconv.Itoa(g.Number), Title: g.Title, Standing: g.State, Priority: g.Priority,
			Counts: Counted(g), Done: g.Done, Total: g.Total, Depth: depthOf(b), Last: b.Last,
			Chosen: w.chosenGoal && g.Number == w.Chosen() || w.isTaken(g.Number), Hovered: GoalAt(l, at, w.goals) == w.pointed}
		if len(g.Labels) > 0 {
			item.Chip, item.ChipColour = g.Labels[0], w.LabelColour(g.Labels[0])
		}
		item.Draw(in, GoalAt(l, at, w.goals))
	}
	clipped.Pop()
	Ground{}.Draw(k, detail)
	drawDetail(k.inside("ground"), w, w.Detail())
}

func drawDetail(k Kit, w Window, d Detail) {
	if len(d.Buttons) == 0 {
		return
	}
	g := d.Goal
	Heading{Said: g.Title, Number: "#" + strconv.Itoa(g.Number)}.Draw(k, d.Heading)
	for i, offer := range d.Offers {
		Button{Said: offer, Key: w.OfferKey(offer), Keyed: true, Look: w.buttonLook(offer, d.Buttons[i])}.Draw(k, d.Buttons[i])
	}
	m := w.measured()
	x := d.Line.X
	pill := func(said string) Box {
		return Box{X: x, Y: d.Line.Y + (d.Line.H-pillHigh)/2, W: ends(m, said, small) + 2*pillPad, H: pillHigh}
	}
	standing := Pill{Said: g.State, Tone: ToneStanding, Standing: g.State}.Draw(k, pill(g.State))
	x = standing.X + standing.W + pillGap
	for _, label := range g.Labels {
		chip := LabelChip{Label: label, Colour: w.LabelColour(label)}.Draw(k, pill(label))
		x = chip.X + chip.W + pillGap
	}
	facts := Box{X: x + pillGap, Y: d.Line.Y, W: max(0, d.Line.X+d.Line.W-x-pillGap), H: d.Line.H}
	say(k.gtx, k.th, k.skin.Colour.Ink.Faint, facts, fits(k.gtx, k.th, d.Facts, facts.W, small), small)
	if d.Progress.W > 0 {
		Progress{Done: g.Done, Total: g.Total, Said: d.Counted}.Draw(k, d.Progress)
		Figures{Said: []Figure{{Figure: d.Spent, Word: "spent"}, {Figure: d.Left, Word: d.LeftSays}}}.Draw(k, d.Figures)
	}
	if d.Note.W > 0 {
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, d.Note, d.NoteSays, small)
	}
	if d.Proposes {
		Banner{Said: d.ProposedSays, Warn: true}.Draw(k, d.Proposed)
	}
	if d.WaitsAt.H > 0 {
		d.Waits.Draw(k, d.WaitsAt)
	}
	if d.SubsHead.H > 0 {
		Head{Said: "Sub-objectives", Count: strconv.Itoa(len(g.Children))}.Draw(k, d.SubsHead)
	}
	for _, t := range d.Tiles {
		tile := t.Tile
		Card{Goal: &tile, Chosen: false, Hovered: t.At == w.pointed}.Draw(k, t.At)
	}
	if d.OwnHead.H > 0 {
		Head{Said: d.OwnSaid, Count: d.OwnCount}.Draw(k, d.OwnHead)
	}
	drawTaskLines(k, w, d.Own)
	if d.InSubsHead.H > 0 {
		Head{Said: "In its sub-objectives", Count: d.InSubsCount}.Draw(k, d.InSubsHead)
	}
	drawTaskLines(k, w, d.InSubs)
	if d.More != "" {
		say(k.gtx, k.th, k.skin.Colour.Accent.Base, d.MoreAt, d.More, small)
	}
}

func drawTaskLines(k Kit, w Window, lines []TaskLine) {
	for _, line := range lines {
		faded := k.faded(1)
		if engine.Finished(line.Card) {
			faded.Pop()
			faded = k.faded(0.5)
		}
		tile := Tile{}.Draw(k, SubPanel(line.At))
		in := k.inside("tile")
		row := SubWriting(line.At)
		row.X, row.W = tile.X+padding, tile.W-2*padding
		numbered := monoWide("#"+strconv.Itoa(line.Card.Number), small)
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: row.X, Y: row.Y, W: numbered, H: row.H}, "#"+strconv.Itoa(line.Card.Number), small, k.skin.Type.Mono, text.Start)
		right := line.Right
		if line.Serves != "" {
			right = line.Serves + " · " + right
		}
		said := ends(w.measured(), right, small) + 2
		end := row.X + row.W - said
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: end, Y: row.Y, W: said, H: row.H}, right, small, font.Font{}, text.Start)
		if line.Serves != "" {
			mark := ServesMark{Ink: k.skin.Colour.Ink.Faint}.Draw(in, Box{X: end - servesSide - servesGap, Y: row.Y + (row.H-servesSide)/2, W: servesSide, H: servesSide})
			end = mark.X
		}
		title := Box{X: row.X + numbered + 2*pillGap + 4, Y: row.Y, W: max(0, end-padding-row.X-numbered-2*pillGap-4), H: row.H}
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, title, fits(k.gtx, k.th, line.Card.Title, title.W, reads), reads)
		faded.Pop()
	}
}
