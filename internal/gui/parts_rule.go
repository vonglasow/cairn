package gui

// RuleLine is one line of the picker a rule is built in: what it names, what it offers, the one chosen or the days ticked, and where each is drawn and clicked.
type RuleLine struct {
	Said       string
	Choices    []string
	On         int
	Ticked     []bool
	Compact    bool
	Caption    Box
	Track      Box
	Picks      []Box
	Back, Next Box
}

// RulePicker is the picker a rule is built in by its parts, hung under what the repeats field reads back: a line a part, then its offers.
type RulePicker struct {
	Box   Box
	Lines []RuleLine
}

const (
	ruleWide    = 380
	ruleControl = 28
	ruleBlock   = lineHigh + inset + ruleControl + between
	ruleBlocks  = 5
	ruleValue   = 64
	ruleDate    = 92
	ruleEndsAt  = 200
)

var ruleCaptions = []string{"every", "", "on", "at", "ends", "", "the next one comes"}

// RulePickerHigh is how high the picker stands: its five blocks of a caption over its controls, then its offers.
func RulePickerHigh() int { return padding + ruleBlocks*ruleBlock + padding + barHigh + padding }

// RulePickerFor lays the picker out at a corner: the number and the unit side by side, the days, the time, the end and how far, and what it counts from.
func RulePickerFor(at Point, lines []RuleLine) RulePicker {
	box := Box{X: at.X, Y: at.Y, W: ruleWide, H: RulePickerHigh()}
	left, full := box.X+padding, ruleWide-padding-padding
	block := func(n int) (caption, control Box) {
		y := box.Y + padding + n*ruleBlock
		return Box{X: left, Y: y, W: full, H: lineHigh}, Box{X: left, Y: y + lineHigh + inset, W: full, H: ruleControl}
	}
	for i := range lines {
		lines[i].Said = ruleCaptions[i]
	}
	caption, control := block(0)
	stepper(&lines[ruleEvery], caption, control, ruleValue)
	after := lines[ruleEvery].Track.X + lines[ruleEvery].Track.W + padding
	spread(&lines[ruleUnit], Box{X: after, Y: control.Y, W: control.X + control.W - after, H: ruleControl})
	caption, control = block(1)
	lines[ruleOn].Caption = caption
	spread(&lines[ruleOn], control)
	caption, control = block(2)
	stepper(&lines[ruleAt], caption, control, ruleValue)
	caption, control = block(3)
	lines[ruleEnds].Caption = caption
	spread(&lines[ruleEnds], Box{X: control.X, Y: control.Y, W: ruleEndsAt, H: ruleControl})
	from := control.X + ruleEndsAt + padding
	stepper(&lines[ruleEnd], Box{}, Box{X: from, Y: control.Y, W: control.X + control.W - from, H: ruleControl}, ruleDate)
	caption, control = block(4)
	lines[ruleComes].Caption = caption
	spread(&lines[ruleComes], control)
	return RulePicker{Box: box, Lines: lines}
}

func stepper(line *RuleLine, caption, control Box, value int) {
	line.Caption, line.Compact = caption, true
	line.Back = Box{X: control.X, Y: control.Y, W: chevronRoom, H: control.H}
	shown := Box{X: control.X + chevronRoom, Y: control.Y, W: value, H: control.H}
	line.Next = Box{X: shown.X + shown.W, Y: control.Y, W: chevronRoom, H: control.H}
	line.Picks = []Box{shown}
	line.Track = Box{X: control.X, Y: control.Y, W: 2*chevronRoom + value, H: control.H}
}

func spread(line *RuleLine, track Box) {
	line.Track = track
	n := len(line.Choices)
	if n == 0 {
		return
	}
	each := track.W / n
	line.Picks = make([]Box, n)
	for i := range n {
		line.Picks[i] = Box{X: track.X + i*each, Y: track.Y, W: each, H: track.H}
	}
	line.Picks[n-1].W = track.X + track.W - line.Picks[n-1].X
}
