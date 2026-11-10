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
	ruleWide     = 400
	ruleControl  = 26
	ruleRow      = ruleControl + between
	ruleCaption  = 48
	ruleValue    = 40
	ruleTime     = 56
	ruleDate     = 88
	ruleFromSaid = 40
	ruleEndsWide = 180
)

var ruleCaptions = []string{"every", "", "on", "at", "from", "ends", "", "the next one comes"}

// RulePickerHigh is how high the picker stands: four lines with their caption beside them, the last under its own, then its offers.
func RulePickerHigh() int {
	return padding + 4*ruleRow + lineHigh + inset + ruleControl + between + padding + barHigh + padding
}

// RulePickerFor lays the picker out at a corner: the number and the unit side by side, the days, the time and the first day, the end and how far, and what it counts from under its caption.
func RulePickerFor(at Point, lines []RuleLine) RulePicker {
	box := Box{X: at.X, Y: at.Y, W: ruleWide, H: RulePickerHigh()}
	left, full := box.X+padding, ruleWide-padding-padding
	row := func(n int) (caption, control Box) {
		y := box.Y + padding + n*ruleRow
		return Box{X: left, Y: y + (ruleControl-lineHigh)/2, W: ruleCaption, H: lineHigh}, Box{X: left + ruleCaption, Y: y, W: full - ruleCaption, H: ruleControl}
	}
	for i := range lines {
		lines[i].Said = ruleCaptions[i]
	}
	caption, control := row(0)
	stepper(&lines[ruleEvery], caption, control, ruleValue)
	after := lines[ruleEvery].Track.X + lines[ruleEvery].Track.W + padding
	spread(&lines[ruleUnit], Box{X: after, Y: control.Y, W: control.X + control.W - after, H: ruleControl})
	caption, control = row(1)
	lines[ruleOn].Caption = caption
	spread(&lines[ruleOn], control)
	caption, control = row(2)
	stepper(&lines[ruleAt], caption, control, ruleTime)
	from := lines[ruleAt].Track.X + lines[ruleAt].Track.W + padding
	stepper(&lines[ruleFrom], Box{X: from, Y: caption.Y, W: ruleFromSaid, H: lineHigh}, Box{X: from + ruleFromSaid, Y: control.Y, W: control.X + control.W - from - ruleFromSaid, H: ruleControl}, ruleDate)
	caption, control = row(3)
	lines[ruleEnds].Caption = caption
	spread(&lines[ruleEnds], Box{X: control.X, Y: control.Y, W: ruleEndsWide, H: ruleControl})
	end := control.X + ruleEndsWide + padding
	stepper(&lines[ruleEnd], Box{}, Box{X: end, Y: control.Y, W: control.X + control.W - end, H: ruleControl}, ruleDate)
	top := box.Y + padding + 4*ruleRow
	lines[ruleComes].Caption = Box{X: left, Y: top, W: full, H: lineHigh}
	spread(&lines[ruleComes], Box{X: left, Y: top + lineHigh + inset, W: full, H: ruleControl})
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
