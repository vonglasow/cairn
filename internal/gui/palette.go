package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"slices"
	"strconv"
	"strings"
	"time"

	"cairn/internal/engine"
)

type paletteState struct {
	opened  time.Time
	typed   board.Written
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
func PaletteKeyAt(line geom.Box, key string) geom.Box {
	wide := KeyChipWide(key)
	return geom.Box{X: line.X + line.W - padding/2 - wide, Y: line.Y + (line.H-chipHigh)/2, W: wide, H: chipHigh}
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
	w.Screen = board.OnPalette
	w.palette = paletteState{recent: w.palette.recent, opened: w.Now}
	return w
}

// leavingSearch gives a search up and goes back to where it was started: the command bar when a slash was typed into it, the board otherwise.
func (w Window) leavingSearch() Window {
	w.Looking = board.Written{}
	if w.SearchFrom == board.OnPalette {
		return w.openingPalette()
	}
	w.Screen = board.OnTheBoard
	return w
}

func (w Window) shuttingPalette() Window {
	w.Screen = board.OnTheBoard
	w.palette.typed, w.palette.row, w.palette.waiting = board.Written{}, 0, ""
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
		if c, ok := w.Snap.Card(n); ok {
			rows, found[n] = append(rows, openingRow(c)), true
		}
	}
	rows = append(rows, w.viewCommandRows(typed)...)
	for _, action := range append(w.actionRows(), w.viewRows()...) {
		if reaching(action.Says, typed) {
			rows = append(rows, action)
		}
	}
	rows = append(rows, w.commandRows(typed)...)
	if typed != "" {
		matched := 0
		for _, c := range w.Showing().Cards {
			if matched < paletteTasksAtMost && !found[c.Number] && board.ReachesCard(c, typed) {
				rows, matched = append(rows, w.taskRow(c)), matched+1
			}
		}
		return rows
	}
	return append(rows, w.palette.recent...)
}

func (w Window) commandRows(typed string) []PaletteRow {
	number, ok := w.CommandedTask(typed)
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
		for _, value := range w.ValuesOf(command.field) {
			if says := command.says + value; engine.Reaches([]string{says}, typed) {
				rows = append(rows, w.settingRow(command.field, strconv.Itoa(number), value, says))
			}
		}
	}
	return rows
}

func (w Window) taskRow(c board.Card) PaletteRow {
	number := c.Number
	return PaletteRow{Group: paletteTasksGroup, Says: "#" + strconv.Itoa(number) + "  " + c.Title, does: func(w Window) Window {
		w.Chosen, w.ChosenGoal = number, false
		return w.opening()
	}}
}

func reaching(said, typed string) bool {
	return strings.TrimSpace(typed) == "" || engine.Reaches([]string{said}, typed)
}

func openingRow(c board.Card) PaletteRow {
	number := c.Number
	return PaletteRow{Group: paletteTasksGroup, Says: "Open #" + strconv.Itoa(number) + " · " + c.Title, does: func(w Window) Window {
		w.Gathered, w.Chosen, w.ChosenGoal = nil, number, false
		return w.chosenInSight().opening()
	}}
}

func (w Window) chosenInSight() Window {
	spot, ok := board.SpotOf(w.Showing().Cards, w.Archived, w.Chosen)
	switch {
	case !ok:
		return w
	case spot.InBand:
		return w.followingLane(spot.Column, spot.At)
	}
	return w.following(spot.Column, spot.At)
}

// typedIntoPalette is the open command bar with what is typed into it changed: a slash typed first turns it into the board's search, with what follows the slash.
func (w Window) typedIntoPalette(typed board.Written) Window {
	if said := typed.Reads(); w.palette.waiting == "" && strings.HasPrefix(said, "/") {
		w = w.shuttingPalette()
		w.Screen, w.Looking, w.SearchFrom = board.OnSearch, board.Written{}.Types(strings.TrimPrefix(said, "/")), board.OnPalette
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
	if c, ok := w.Snap.Card(w.Chosen); ok && w.Chosen != 0 && !w.ChosenGoal {
		n := "#" + strconv.Itoa(c.Number)
		for _, line := range w.menuFor() {
			action(n+" "+line, MenuKey(line), paletteDoing(line))
		}
		if c.Series != nil && w.Pages.Recurs() {
			series := c.Series.Series
			action("Open the series of "+n, "", func(w Window) Window { return w.onSeries(series) })
		}
	}
	if w.Pages.Recurs() {
		action("Series", "", func(w Window) Window { return w.onPage(board.Series) })
		action(NewSeriesSays, "", func(w Window) Window { return w.startingSeries() })
	}
	for _, general := range []struct{ says, key string }{
		{"New task", "cmd+n"}, {"Search", "/"}, {"Filters", "shift+f"}, {"Labels", "shift+l"},
		{"Archived", "shift+a"}, {"Backups", "shift+b"}, {"New backup", "cmd+b"}, {"Export", "cmd+e"},
		{"Import", "cmd+o"}, {"Open a workspace…", "cmd+shift+o"}, {"Settings…", "cmd+,"}, {"Help", "?"},
	} {
		action(general.says, general.key, pressing(general.key))
	}
	action("Reset sizes", "", func(w Window) Window { return w.Choose("Reset sizes") })
	return rows
}

func paletteDoing(line string) func(Window) Window {
	switch line {
	case "Move to…":
		return waitingFor("state")
	case "Priority…":
		return waitingFor("priority")
	case "Labels…":
		return waitingFor("label")
	}
	return func(w Window) Window { return w.doingLine(line) }
}

func waitingFor(field string) func(Window) Window {
	return func(w Window) Window {
		w.Screen = board.OnPalette
		w.palette.waiting, w.palette.typed, w.palette.row = field, board.Written{}, 0
		return w
	}
}

func (w Window) valueRows(typed string) []PaletteRow {
	field, number := w.palette.waiting, strconv.Itoa(w.Chosen)
	values := w.ValuesOf(field)
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

func (w Window) settingRow(field, number, value, says string) PaletteRow {
	given := value
	if field == "label" {
		given = "+" + value
	}
	return PaletteRow{Group: paletteActionsGroup, Says: says, does: func(w Window) Window {
		return w.asking(engine.TaskEdit, engine.Input{Given: map[string]string{"number": number, field: given}})
	}}
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
		kept := []PaletteRow{{Group: paletteRecentGroup, Says: row.Says, does: row.does, done: w.Now}}
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
	Box   geom.Box
	Group string
	Row   int
}

// PaletteBox is the open command bar's panel: over the top of the window, under the tabs and in its middle, half the window wide, as tall as its lines, up to the bar at the foot, which stays in sight.
func (w Window) PaletteBox() geom.Box {
	l := w.Layout()
	wide := min(w.width-2*geom.Margin, max(paletteLeast, w.width/2))
	high := paletteHead + padding
	if banner, ok := w.paletteBannerIn(geom.Box{W: wide}); ok {
		high = max(high, banner.Y+banner.H+padding)
	}
	for _, line := range w.paletteLines(geom.Box{W: wide}) {
		high = max(high, line.Box.Y+line.Box.H+padding)
	}
	top := l.Tabs.Y + l.Tabs.H
	high = min(high, l.Bar.Y-geom.Margin-top)
	return geom.Box{X: (w.width - wide) / 2, Y: top, W: wide, H: high}
}

// PaletteLines is every line of the open command bar where it is drawn.
func (w Window) PaletteLines() []PaletteLine { return w.paletteLines(w.PaletteBox()) }

func (w Window) paletteLines(box geom.Box) []PaletteLine {
	var lines []PaletteLine
	y, group := box.Y+paletteHead, ""
	if banner, ok := w.paletteBannerIn(box); ok {
		y = banner.Y + banner.H + padding
	}
	for i, row := range w.PaletteRows() {
		if row.Group != group {
			group = row.Group
			lines = append(lines, PaletteLine{Box: geom.Box{X: box.X + padding, Y: y, W: box.W - 2*padding, H: paletteGroupLow}, Group: group, Row: -1})
			y += paletteGroupLow
		}
		lines = append(lines, PaletteLine{Box: geom.Box{X: box.X + padding/2, Y: y, W: box.W - padding, H: paletteRowHigh}, Row: i})
		y += paletteRowHigh
	}
	return lines
}

// PaletteTyped is where what is typed into the open command bar is written: across its head, short of where it says esc shuts it.
func PaletteTyped(box geom.Box) geom.Box {
	return geom.Box{X: box.X + padding + inset, Y: box.Y + (paletteHead-lineHigh)/2, W: box.W - 2*padding - paletteShut, H: lineHigh}
}

const paletteShut = 90

// CommandLine is the part of the command bar left of its hints, which a click opens.
func CommandLine(l geom.Layout, hints []geom.Box) geom.Box {
	end := l.Bar.X + l.Bar.W
	if len(hints) > 0 {
		end = hints[0].X
	}
	return geom.Box{X: l.Bar.X, Y: l.Bar.Y, W: end - l.Bar.X, H: l.Bar.H}
}

func (w Window) clickingPalette(p geom.Point) Window {
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
func (w Window) SearchPill() (geom.Box, bool) {
	for _, pill := range w.Pills() {
		if pill.Kind == PillSearch {
			return pill.Box, true
		}
	}
	return geom.Box{}, false
}

const searchPillHigh = 22

// paletteRises is how long the open command bar takes to drop into place.
const paletteRises = 120 * time.Millisecond

// paletteRiseBy is how far above its place the command bar starts dropping from.
const paletteRiseBy = 8

// PaletteRise is how far the open command bar has dropped into place, from nought to one, eased to land softly, and whether it is still moving.
func (w Window) PaletteRise() (float64, bool) {
	if w.Screen != board.OnPalette {
		return 1, false
	}
	share := float64(w.Now.Sub(w.palette.opened)) / float64(paletteRises)
	if share >= 1 || share < 0 {
		return 1, false
	}
	return 1 - (1-share)*(1-share), true
}

// PaletteBanner is what the open command bar says over its rows when what is typed reaches for what is turned off — a series while recurrence is off — and whether it says anything.
func (w Window) PaletteBanner() (string, bool) {
	typed := strings.TrimSpace(w.palette.typed.Reads())
	if w.Pages.Recurs() || typed == "" || !engine.Reaches([]string{paletteOffWords}, typed) {
		return "", false
	}
	return engine.ErrRecurrenceOff.Error(), true
}

const paletteOffWords = "series recurrence repeats every skip"

func paletteBannerSaid(said string) (string, string) {
	first, then, _ := strings.Cut(said, ": ")
	return first, then
}

// PaletteBannerAt is where the open command bar draws what it says over its rows.
func (w Window) PaletteBannerAt() (geom.Box, bool) { return w.paletteBannerIn(w.PaletteBox()) }

func (w Window) paletteBannerIn(box geom.Box) (geom.Box, bool) {
	said, ok := w.PaletteBanner()
	if !ok {
		return geom.Box{}, false
	}
	first, then := paletteBannerSaid(said)
	wide := box.W - 2*padding
	return geom.Box{X: box.X + padding, Y: box.Y + paletteHead + padding, W: wide, H: BannerHigh(w.measured(), first, then, wide)}, true
}
