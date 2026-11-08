package gui

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"cairn/internal/engine"
)

type paletteState struct {
	opened  time.Time
	typed   Written
	row     int
	waiting string
	recent  []PaletteRow
}

// PaletteRow is one line of the command bar opened: the group it is listed under, what it says, the key that does the same on the board, and what choosing it does.
type PaletteRow struct {
	Group, Says, Key string
	does             func(Window) Window
	done             time.Time
}

// Ago is how long ago a recent line was done, as the command bar says it beside it: just now under a minute, then in minutes and hours.
func (r PaletteRow) Ago(now time.Time) string {
	since := now.Sub(r.done)
	switch {
	case r.done.IsZero():
		return ""
	case since < time.Minute:
		return "just now"
	case since < time.Hour:
		return strconv.Itoa(int(since/time.Minute)) + " min ago"
	}
	return strconv.Itoa(int(since/time.Hour)) + " h ago"
}

// PaletteKeyAt is where the chosen line of the command bar draws the key that does it: at the right of the line.
func PaletteKeyAt(line Box, key string) Box {
	wide := KeyChipWide(key)
	return Box{X: line.X + line.W - padding/2 - wide, Y: line.Y + (line.H-chipHigh)/2, W: wide, H: chipHigh}
}

// The groups the open command bar lists its lines under, in the order they come.
const (
	paletteActionsGroup = "Actions"
	paletteTasksGroup   = "Tasks"
	paletteRecentGroup  = "Recent"
)

const (
	paletteTasksAtMost  = 5
	paletteRecentAtMost = 5
)

func (w Window) openingPalette() Window {
	w.screen = OnPalette
	w.palette = paletteState{recent: w.palette.recent, opened: w.now}
	return w
}

// leavingSearch gives a search up and goes back to where it was started: the command bar when a slash was typed into it, the board otherwise.
func (w Window) leavingSearch() Window {
	w.looking = Written{}
	if w.searchFrom == OnPalette {
		return w.openingPalette()
	}
	w.screen = OnTheBoard
	return w
}

func (w Window) shuttingPalette() Window {
	w.screen = OnTheBoard
	w.palette.typed, w.palette.row, w.palette.waiting = Written{}, 0, ""
	return w
}

// PaletteRows is every line the open command bar offers for what was typed: the value an action waits for, or the actions of the moment, then the tasks that match, then what was done from it lately.
func (w Window) PaletteRows() []PaletteRow {
	typed := w.palette.typed.Reads()
	if w.palette.waiting != "" {
		return w.valueRows(typed)
	}
	var rows []PaletteRow
	found := map[int]bool{}
	if n, err := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(typed), "#")); err == nil {
		if c, ok := w.snap.Card(n); ok {
			rows, found[n] = append(rows, openingRow(c)), true
		}
	}
	for _, action := range w.actionRows() {
		if reaching(action.Says, typed) {
			rows = append(rows, action)
		}
	}
	rows = append(rows, w.commandRows(typed)...)
	if typed != "" {
		matched := 0
		for _, c := range w.Showing().Cards {
			if matched < paletteTasksAtMost && !found[c.Number] && engine.ReachesCard(c, typed) {
				rows, matched = append(rows, w.taskRow(c)), matched+1
			}
		}
		return rows
	}
	return append(rows, w.palette.recent...)
}

func (w Window) commandRows(typed string) []PaletteRow {
	number, ok := w.commandedTask(typed)
	if !ok {
		return nil
	}
	n := "#" + strconv.Itoa(number)
	var rows []PaletteRow
	for _, command := range []struct{ field, bare, says string }{
		{"state", "Move " + n + " to…", "Move " + n + " to "},
		{"priority", "Priority of " + n + "…", "Priority of " + n + ": "},
		{"label", "Label " + n + " with…", "Label " + n + " with "},
	} {
		if reaching(command.bare, typed) {
			continue
		}
		for _, value := range w.valuesOf(command.field) {
			if says := command.says + value; engine.Reaches([]string{says}, typed) {
				rows = append(rows, w.settingRow(command.field, strconv.Itoa(number), value, says))
			}
		}
	}
	return rows
}

func (w Window) commandedTask(typed string) (int, bool) {
	for _, word := range strings.Fields(typed) {
		if n, err := strconv.Atoi(strings.TrimPrefix(word, "#")); err == nil {
			_, ok := w.snap.Card(n)
			return n, ok
		}
	}
	return w.chosen, typed != "" && w.chosen != 0 && !w.chosenGoal
}

func (w Window) taskRow(c engine.Card) PaletteRow {
	number := c.Number
	return PaletteRow{Group: paletteTasksGroup, Says: "#" + strconv.Itoa(number) + "  " + c.Title, does: func(w Window) Window {
		w.chosen, w.chosenGoal = number, false
		return w.opening()
	}}
}

func reaching(said, typed string) bool {
	return strings.TrimSpace(typed) == "" || engine.Reaches([]string{said}, typed)
}

func openingRow(c engine.Card) PaletteRow {
	number := c.Number
	return PaletteRow{Group: paletteTasksGroup, Says: "Open #" + strconv.Itoa(number) + " · " + c.Title, does: func(w Window) Window {
		w.taken, w.chosen, w.chosenGoal = nil, number, false
		return w.chosenInSight().opening()
	}}
}

func (w Window) chosenInSight() Window {
	for column := range engine.Columns {
		for at, c := range w.inColumn(column) {
			if c.Number == w.chosen {
				return w.following(column, at)
			}
		}
	}
	if lane, at, ok := w.whereInBand(); ok {
		return w.followingLane(lane, at)
	}
	return w
}

// typedIntoPalette is the open command bar with what is typed into it changed: a slash typed first turns it into the board's search, with what follows the slash.
func (w Window) typedIntoPalette(typed Written) Window {
	if said := typed.Reads(); w.palette.waiting == "" && strings.HasPrefix(said, "/") {
		w = w.shuttingPalette()
		w.screen, w.looking, w.searchFrom = OnSearch, Written{}.Types(strings.TrimPrefix(said, "/")), OnPalette
		return w
	}
	w.palette.typed, w.palette.row = typed, 0
	return w
}

func pressing(key string) func(Window) Window { return func(w Window) Window { return w.Key(key) } }

func (w Window) actionRows() []PaletteRow {
	var rows []PaletteRow
	action := func(says, key string, does func(Window) Window) {
		rows = append(rows, PaletteRow{Group: paletteActionsGroup, Says: says, Key: key, does: does})
	}
	if c, ok := w.snap.Card(w.chosen); ok && w.chosen != 0 && !w.chosenGoal {
		n := "#" + strconv.Itoa(c.Number)
		action("Move "+n+" to…", "m", waitingFor("state"))
		action("Priority of "+n+"…", "p", waitingFor("priority"))
		action("Label "+n+" with…", "l", waitingFor("label"))
		action("Edit "+n, "e", pressing("e"))
		focus := "Focus " + n
		if c.InFocus {
			focus = "Unfocus " + n
		}
		action(focus, "f", pressing("f"))
		action("Archive "+n, "x", func(w Window) Window { return w.Choose("Archive") })
		action("Delete "+n, "cmd+backspace", func(w Window) Window { return w.Choose("Delete") })
	}
	for _, general := range []struct{ says, key string }{
		{"New task", "cmd+n"}, {"Search", "/"}, {"Filters", "shift+f"}, {"Labels", "shift+l"},
		{"Archived", "shift+a"}, {"Backups", "shift+b"}, {"New backup", "cmd+b"}, {"Export", "cmd+e"},
		{"Import", "cmd+o"}, {"Configuration", "cmd+,"}, {"Help", "?"},
	} {
		action(general.says, general.key, pressing(general.key))
	}
	return rows
}

func waitingFor(field string) func(Window) Window {
	return func(w Window) Window {
		w.screen = OnPalette
		w.palette.waiting, w.palette.typed, w.palette.row = field, Written{}, 0
		return w
	}
}

func (w Window) valueRows(typed string) []PaletteRow {
	field, number := w.palette.waiting, strconv.Itoa(w.chosen)
	values := w.valuesOf(field)
	var rows []PaletteRow
	for _, value := range values {
		if !reaching(value, typed) {
			continue
		}
		rows = append(rows, w.settingRow(field, number, value, Titled(value)))
	}
	if field == "label" && typed != "" && !slices.Contains(values, typed) {
		rows = append(rows, w.settingRow(field, number, typed, "Create "+typed))
	}
	return rows
}

func (w Window) valuesOf(field string) []string {
	switch field {
	case "state":
		return engine.States()
	case "priority":
		return engine.Priorities()
	case "label":
		return w.boardLabels()
	}
	return nil
}

func (w Window) settingRow(field, number, value, says string) PaletteRow {
	given := value
	if field == "label" {
		given = "+" + value
	}
	return PaletteRow{Group: paletteActionsGroup, Says: says, does: func(w Window) Window {
		return w.asking(engine.TaskEdit, engine.Input{Given: map[string]string{"number": number, field: given}})
	}}
}

func (w Window) boardLabels() []string {
	var labels []string
	for _, c := range w.snap.Cards {
		for _, label := range c.Labels {
			if !slices.Contains(labels, label) {
				labels = append(labels, label)
			}
		}
	}
	slices.Sort(labels)
	return labels
}

func (w Window) onPalette(name string) Window {
	if edit, ok := textKey(name); ok {
		return w.typedIntoPalette(edit(w.palette.typed))
	}
	rows := w.PaletteRows()
	switch name {
	case "esc":
		return w.shuttingPalette()
	case "backspace":
		if w.palette.typed.Reads() == "" {
			if w.palette.waiting != "" {
				w.palette.waiting = ""
				return w
			}
			return w.shuttingPalette()
		}
		w.palette.typed, w.palette.row = w.palette.typed.Back(), 0
	case "down":
		w.palette.row = min(max(0, len(rows)-1), w.palette.row+1)
	case "up":
		w.palette.row = max(0, w.palette.row-1)
	case "enter":
		if w.palette.row < len(rows) {
			return w.runningPalette(rows[w.palette.row])
		}
	}
	return w
}

func (w Window) runningPalette(row PaletteRow) Window {
	waits := w.palette.waiting == ""
	recent := w.palette.recent
	w = w.shuttingPalette()
	done := row.does(w)
	if waits && row.Group != paletteTasksGroup {
		kept := []PaletteRow{{Group: paletteRecentGroup, Says: row.Says, does: row.does, done: w.now}}
		for _, was := range recent {
			if was.Says != row.Says && len(kept) < paletteRecentAtMost {
				kept = append(kept, was)
			}
		}
		done.palette.recent = kept
	}
	return done
}

const (
	paletteHead     = 44
	paletteGroupLow = 24
	paletteRowHigh  = 28
	paletteLeast    = 480
)

// PaletteLine is one line of the open command bar as it is drawn: a group's name, or the i-th row.
type PaletteLine struct {
	Box   Box
	Group string
	Row   int
}

// PaletteBox is the open command bar's panel: over the top of the window, under the tabs and in its middle, half the window wide, as tall as its lines, up to the bar at the foot, which stays in sight.
func (w Window) PaletteBox() Box {
	l := w.Layout()
	wide := min(w.width-2*margin, max(paletteLeast, w.width/2))
	high := paletteHead + padding
	for _, line := range w.paletteLines(Box{}) {
		high = max(high, line.Box.Y+line.Box.H+padding)
	}
	top := l.Tabs.Y + l.Tabs.H
	high = min(high, l.Bar.Y-margin-top)
	return Box{X: (w.width - wide) / 2, Y: top, W: wide, H: high}
}

// PaletteLines is every line of the open command bar where it is drawn.
func (w Window) PaletteLines() []PaletteLine { return w.paletteLines(w.PaletteBox()) }

func (w Window) paletteLines(box Box) []PaletteLine {
	var lines []PaletteLine
	y, group := box.Y+paletteHead, ""
	for i, row := range w.PaletteRows() {
		if row.Group != group {
			group = row.Group
			lines = append(lines, PaletteLine{Box: Box{X: box.X + padding, Y: y, W: box.W - 2*padding, H: paletteGroupLow}, Group: group, Row: -1})
			y += paletteGroupLow
		}
		lines = append(lines, PaletteLine{Box: Box{X: box.X + padding/2, Y: y, W: box.W - padding, H: paletteRowHigh}, Row: i})
		y += paletteRowHigh
	}
	return lines
}

// PaletteTyped is where what is typed into the open command bar is written: across its head, short of where it says esc shuts it.
func PaletteTyped(box Box) Box {
	return Box{X: box.X + padding + inset, Y: box.Y + (paletteHead-lineHigh)/2, W: box.W - 2*padding - paletteShut, H: lineHigh}
}

const paletteShut = 90

// CommandLine is the part of the command bar left of its hints, which a click opens.
func CommandLine(l Layout, hints []Box) Box {
	end := l.Bar.X + l.Bar.W
	if len(hints) > 0 {
		end = hints[0].X
	}
	return Box{X: l.Bar.X, Y: l.Bar.Y, W: end - l.Bar.X, H: l.Bar.H}
}

func (w Window) clickingPalette(p Point) Window {
	if !w.PaletteBox().In(p) {
		return w.shuttingPalette()
	}
	rows := w.PaletteRows()
	for _, line := range w.PaletteLines() {
		if line.Row >= 0 && line.Box.In(p) && line.Box.Y+line.Box.H <= w.PaletteBox().Y+w.PaletteBox().H {
			return w.runningPalette(rows[line.Row])
		}
	}
	return w
}

const pillCross = 18

// SearchPill is the pill a search kept stands as, first in the row after the tabs, its cross clearing it, and whether a search is kept at all.
func (w Window) SearchPill() (Box, bool) {
	for _, pill := range w.Pills() {
		if pill.Kind == PillSearch {
			return pill.Box, true
		}
	}
	return Box{}, false
}

const searchPillHigh = 22

// paletteRises is how long the open command bar takes to drop into place.
const paletteRises = 120 * time.Millisecond

// paletteRiseBy is how far above its place the command bar starts dropping from.
const paletteRiseBy = 8

// PaletteRise is how far the open command bar has dropped into place, from nought to one, eased to land softly, and whether it is still moving.
func (w Window) PaletteRise() (float64, bool) {
	if w.screen != OnPalette {
		return 1, false
	}
	share := float64(w.now.Sub(w.palette.opened)) / float64(paletteRises)
	if share >= 1 || share < 0 {
		return 1, false
	}
	return 1 - (1-share)*(1-share), true
}
