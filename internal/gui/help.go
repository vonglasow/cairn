package gui

type helpGroup struct {
	name  string
	lines []Said
}

func helpGroups() []helpGroup {
	return []helpGroup{
		{"Board", []Said{
			{Key: "↑↓←→", Means: "walk the board"},
			{Key: "⇧←→↑↓", Means: "move the card"},
			{Key: "⏎", Means: "open"},
			{Key: "e", Means: "edit"},
			{Key: "f", Means: "focus"},
			{Key: "m", Means: "move to…"},
			{Key: "p", Means: "priority"},
			{Key: "l", Means: "labels"},
			{Key: "t", Means: "deadline"},
			{Key: "o", Means: "new linked task"},
			{Key: "x ⇧X", Means: "archive · column"},
			{Key: "⌘⌫", Means: "delete"},
			{Key: "esc", Means: "let the card go"},
		}},
		{"Forms", []Said{
			{Key: "⌘⏎", Means: "save · done"},
			{Key: "esc esc", Means: "throw away"},
			{Key: "⌘Z", Means: "bring it back"},
			{Key: "tab", Means: "next field"},
			{Key: "space", Means: "tick a step"},
		}},
		{"Find", []Said{
			{Key: "⌘K", Means: "command bar"},
			{Key: "/", Means: "search"},
			{Key: "⇧F ⌘1–9", Means: "filters, views"},
			{Key: "⇧A", Means: "archived"},
			{Key: "g t/o/i", Means: "go to a page"},
			{Key: "⌘N", Means: "new task"},
		}},
		{"Store", []Said{
			{Key: "⇧L", Means: "labels"},
			{Key: "⇧B", Means: "backups"},
			{Key: "⌘B", Means: "new backup"},
			{Key: "⌘,", Means: "configuration"},
			{Key: "⌘E ⌘O", Means: "export, import"},
			{Key: "?", Means: "this help"},
		}},
		{"Pointer", []Said{
			{Key: "double-click", Means: "open"},
			{Key: "drag", Means: "move the card"},
			{Key: "right click", Means: "its menu"},
			{Key: "drag a gutter", Means: "resize"},
		}},
	}
}

// Helping is what every key does, as a person reads it, spelled for this keyboard.
func (w Window) Helping() []Said {
	var lines []Said
	for _, group := range helpGroups() {
		for _, line := range group.lines {
			lines = append(lines, Said{Key: w.keyboard.Spelled(line.Key), Means: w.keyboard.Spelled(line.Means)})
		}
	}
	return lines
}

// HelpLine is one line of the help where it is laid out: the name of a group of keys, or a key on its chip and what it does beside it.
type HelpLine struct {
	Head, Key, Means   string
	At, KeyAt, MeansAt Box
}

// HelpLines is the help laid out in the panel: its groups one under another, the first half of them in the first column and the rest in the second when the panel is wide enough for two, what each key does starting where it does for every key of its column.
func (w Window) HelpLines() []HelpLine {
	box := w.HelpBox()
	split := w.helpSplit()
	wide := (box.W - 2*padding - (len(split)-1)*helpBetween) / len(split)
	var lines []HelpLine
	for c, groups := range split {
		x := box.X + padding + c*(wide+helpBetween)
		chips := 0
		for _, group := range groups {
			for _, line := range group.lines {
				chips = max(chips, KeyChipWide(w.keyboard.Spelled(line.Key)))
			}
		}
		row := 0
		place := func() Box {
			row++
			return Box{X: x, Y: FirstRow(box) + (row-1)*helpRow, W: wide, H: lineHigh}
		}
		for _, group := range groups {
			lines = append(lines, HelpLine{Head: group.name, At: place()})
			for _, line := range group.lines {
				slot := place()
				key := w.keyboard.Spelled(line.Key)
				chip := Box{X: slot.X, Y: slot.Y + (slot.H-chipHigh)/2, W: KeyChipWide(key), H: chipHigh}
				from := slot.X + chips + helpGap
				lines = append(lines, HelpLine{Key: key, Means: w.keyboard.Spelled(line.Means), At: slot, KeyAt: chip,
					MeansAt: Box{X: from, Y: slot.Y, W: slot.X + slot.W - from, H: slot.H}})
			}
		}
	}
	return lines
}

func (w Window) helpSplit() [][]helpGroup {
	groups := helpGroups()
	if w.helpColumns() == 1 {
		return [][]helpGroup{groups}
	}
	total := 0
	for _, group := range groups {
		total += 1 + len(group.lines)
	}
	first, count := 0, 0
	for first < len(groups) && 2*count < total {
		count += 1 + len(groups[first].lines)
		first++
	}
	return [][]helpGroup{groups[:first], groups[first:]}
}

func (w Window) helpRows() int {
	rows := 0
	for _, groups := range w.helpSplit() {
		count := 0
		for _, group := range groups {
			count += 1 + len(group.lines)
		}
		rows = max(rows, count)
	}
	return rows
}

func (w Window) helpColumns() int {
	if Over(w.width, w.height).W-2*padding >= 2*helpColumnLeast+helpBetween {
		return 2
	}
	return 1
}

// HelpBox is the panel the help is drawn in: across the middle of the window, as tall as its longest column.
func (w Window) HelpBox() Box {
	over := Over(w.width, w.height)
	rows := w.helpRows()
	fixed := FirstRow(over) - over.Y + padding + barHigh + padding
	high := min(over.H, fixed+rows*helpRow)
	return Box{X: over.X, Y: over.Y + (over.H-high)/2, W: over.W, H: high}
}

const (
	helpRow         = 20
	helpGap         = 10
	helpBetween     = 24
	helpColumnLeast = 230
)
