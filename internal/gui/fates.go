package gui

import (
	"strconv"
	"strings"

	"cairn/internal/engine"
)

// FateRow is one line of the question an objective's deletion asks: what it names, the choices it offers, the one chosen, and where it stands.
type FateRow struct {
	Said    string
	Sub     int
	Choices []string
	Fates   []engine.Fate
	On      int
	At      Box
	Choice  Box
	Picks   []Box
}

// Fates is the question deleting an objective asks, row by row: each sub-objective deleted or kept as level one, the tasks it touches deleted or detached; nothing chosen for the person.
type Fates struct {
	Box, Heading, Said, SubsHead, TasksHead, Note Box
	Rows                                          []FateRow
	Asks, NoteSays                                string
}

var (
	subChoices  = []string{"Delete", "Keep as level one"}
	subFates    = []engine.Fate{engine.DeleteIt, engine.KeepIt}
	taskChoices = []string{"Delete", "Detach"}
	taskFates   = []engine.Fate{engine.DeleteIt, engine.DetachThem}
)

const (
	fatesWide = 560
	fateHigh  = 30
)

func (w Window) deletingGoal(n int) Window {
	d := w.snap.Deleting(n)
	if !d.Asks() {
		w.screen = OnAsking
		return w
	}
	w.screen, w.fates, w.fateRow = OnFates, d, 0
	return w
}

// FatesAsked lays the question out over the window.
func (w Window) FatesAsked() Fates {
	d := w.fates
	g, _ := w.snap.Goal(d.Number)
	var f Fates
	for _, sub := range d.Subs {
		s, _ := w.snap.Goal(sub)
		f.Rows = append(f.Rows, FateRow{Said: "#" + strconv.Itoa(sub) + " " + s.Title, Sub: sub, Choices: subChoices, Fates: subFates, On: fateAt(subFates, d.FateOf(sub))})
	}
	touched := d.Touched()
	if d.AsksTasks() {
		said := numbersSaid(touched)
		if len(touched) == len(engine.Serving(w.snap.Cards, d.Number)) {
			said += " on " + g.Title + " itself"
		}
		f.Rows = append(f.Rows, FateRow{Said: said, Choices: taskChoices, Fates: taskFates, On: fateAt(taskFates, d.TasksFate())})
	}
	f.Asks = "It holds " + strconv.Itoa(len(d.Subs)) + " sub-objective" + plural(len(d.Subs)) + " and touches " + strconv.Itoa(len(touched)) + " task" + plural(len(touched)) + ". Say what becomes of each."
	f.NoteSays = "Detached tasks stay on the board and serve nothing. The tasks of a sub-objective kept follow it."
	wide := min(fatesWide, w.width-2*margin)
	high := FirstRow(Box{}) + lineHigh + padding + len(f.Rows)*fateHigh + 2*(lineHigh+inset) + 2*lineHigh + padding + barHigh + padding
	f.Box = Box{X: (w.width - wide) / 2, Y: max(margin, (w.height-high)/3), W: wide, H: high}
	f.Heading = FormHeading(f.Box)
	y := FirstRow(f.Box)
	f.Said = RowIn(f.Box, y, lineHigh)
	y += lineHigh + padding
	m := w.measured()
	for i := range f.Rows {
		if i == 0 && len(d.Subs) > 0 {
			f.SubsHead = RowIn(f.Box, y, lineHigh)
			y += lineHigh + inset
		}
		if i == len(d.Subs) {
			f.TasksHead = RowIn(f.Box, y, lineHigh)
			y += lineHigh + inset
		}
		row := &f.Rows[i]
		row.At = RowIn(f.Box, y, fateHigh-2)
		x := row.At.X + row.At.W - inset
		for j := len(row.Choices) - 1; j >= 0; j-- {
			seg := ends(m, row.Choices[j], small) + 2*padding
			x -= seg
			row.Picks = append([]Box{{X: x, Y: row.At.Y + 3, W: seg, H: row.At.H - 6}}, row.Picks...)
		}
		row.Choice = Box{X: x, Y: row.At.Y + 3, W: row.At.X + row.At.W - inset - x, H: row.At.H - 6}
		y += fateHigh
	}
	f.Note = RowIn(f.Box, y+inset, 2*lineHigh)
	return f
}

func fateAt(fates []engine.Fate, f engine.Fate) int {
	for i, one := range fates {
		if one == f {
			return i
		}
	}
	return -1
}

func numbersSaid(numbers []int) string {
	said := make([]string, 0, len(numbers))
	for _, n := range numbers {
		said = append(said, "#"+strconv.Itoa(n))
	}
	return strings.Join(said, ", ")
}

// DeletesSays is what the button that deletes says: the objectives it deletes, and the tasks when they go too.
func (w Window) DeletesSays() string {
	d := w.fates
	gone := []int{d.Number}
	for _, sub := range d.Subs {
		if d.FateOf(sub) == engine.DeleteIt {
			gone = append(gone, sub)
		}
	}
	said := numbersSaid(gone)
	if at := strings.LastIndex(said, ", "); at >= 0 {
		said = said[:at] + " and " + said[at+2:]
	}
	if d.TasksFate() == engine.DeleteIt {
		n := len(d.Touched())
		said += " with " + strconv.Itoa(n) + " task" + plural(n)
	}
	return "Delete " + said
}

func (w Window) fated(row, at int) Window {
	f := w.FatesAsked()
	if row < 0 || row >= len(f.Rows) || at < 0 || at >= len(f.Rows[row].Fates) {
		return w
	}
	fate := f.Rows[row].Fates[at]
	if f.Rows[row].Sub != 0 {
		w.fates = w.fates.Choose(f.Rows[row].Sub, fate)
	} else {
		w.fates = w.fates.Tasks(fate)
	}
	w.fateRow = min(row, len(w.FatesAsked().Rows)-1)
	return w
}

func (w Window) onFates(name string) Window {
	f := w.FatesAsked()
	switch name {
	case "down", "j":
		w.fateRow = min(len(f.Rows)-1, w.fateRow+1)
	case "up", "k":
		w.fateRow = max(0, w.fateRow-1)
	case "left", "h", "right", "l":
		row := f.Rows[w.fateRow]
		at := 0
		if name == "right" || name == "l" {
			at = len(row.Choices) - 1
		}
		return w.fated(w.fateRow, at)
	}
	return w
}

func (w Window) confirmingFates() Window {
	if !w.fates.Ready() {
		return w
	}
	w.screen = OnTheBoard
	if _, err := w.work.Ask(engine.ObjectiveDelete, w.fates.Input()); err != nil {
		return w.failing(err)
	}
	w.chosen, w.chosenGoal = 0, false
	return w.Look()
}

func (w Window) clickingFates(p Point) Window {
	if !w.FatesAsked().Box.In(p) {
		return w
	}
	for i, row := range w.FatesAsked().Rows {
		for j, pick := range row.Picks {
			if pick.In(p) {
				return w.fated(i, j)
			}
		}
	}
	return w
}

func drawFates(k Kit, w Window) {
	f := w.FatesAsked()
	Veil{}.Draw(k, Box{W: w.width, H: w.height})
	k = k.inside("veil")
	Panel{}.Draw(k, f.Box)
	in := k.inside("panel")
	g, _ := w.snap.Goal(w.fates.Number)
	Heading{Said: "Delete " + g.Title, Number: "#" + strconv.Itoa(g.Number)}.Draw(in, f.Heading)
	say(k.gtx, k.th, k.skin.Colour.Ink.Faint, f.Said, fits(k.gtx, k.th, f.Asks, f.Said.W, reads), reads)
	if f.SubsHead.H > 0 {
		Head{Said: "Sub-objectives", Count: strconv.Itoa(len(w.fates.Subs))}.Draw(in, f.SubsHead)
	}
	if f.TasksHead.H > 0 {
		Head{Said: "Tasks", Count: strconv.Itoa(len(w.fates.Touched()))}.Draw(in, f.TasksHead)
	}
	for i, row := range f.Rows {
		r := Row{Picked: i == w.fateRow}
		r.Draw(in, row.At)
		say(k.gtx, k.th, k.skin.Colour.Ink.Text, Box{X: row.At.X + padding, Y: row.At.Y + (row.At.H-lineHigh)/2, W: max(0, row.Choice.X-row.At.X-2*padding), H: lineHigh}, row.Said, reads)
		Segmented{Said: row.Choices, At: row.Picks, On: row.On, Small: true}.Draw(in.inside("row"), row.Choice)
	}
	sayOver(k.gtx, k.th, k.skin.Colour.Ink.Faint, f.Note, f.NoteSays, small, 2)
	drawPanelOffers(in, f.Box, w)
}
