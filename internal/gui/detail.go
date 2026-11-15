package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine"
	"cairn/internal/engine/calendar"
	"cairn/internal/gui/geom"
	"fmt"
	"strconv"
	"time"
)

// Ghost is the card in mid-air: the card itself, its own size, held where it was taken; and whether there is one.
func (w Window) Ghost() (geom.Box, bool) {
	slot, ok := w.ghostSlot()
	if !ok {
		return geom.Box{}, false
	}
	panel := PartsOf(slot).Panel
	return geom.Box{X: slot.X, Y: slot.Y, W: panel.W, H: panel.H}, true
}

func (w Window) ghostSlot() (geom.Box, bool) {
	if !w.Dragging() {
		return geom.Box{}, false
	}
	for _, p := range w.Placed() {
		if p.Card.Number == w.held.Number {
			by := geom.Point{X: w.held.At.X - w.held.From.X, Y: w.held.At.Y - w.held.From.Y}
			return geom.Box{X: p.Slot.X + by.X, Y: p.Slot.Y + by.Y, W: p.Slot.W, H: p.Slot.H}, true
		}
	}
	return geom.Box{}, false
}

// LineAbove is the line a drop lands on just above a slot: in the gap between it and the card before.
func LineAbove(slot geom.Box) geom.Box {
	panel := PartsOf(slot).Panel
	gap := slot.H - panel.H
	return geom.Box{X: panel.X, Y: slot.Y - gap/2 - landingHigh/2 - landingHigh%2, W: panel.W, H: landingHigh}
}

// Landing is the line where the card would drop, between two cards, the board not moving; and whether it is over a place it would move to.
func (w Window) Landing() (geom.Box, bool) {
	if !w.Dragging() {
		return geom.Box{}, false
	}
	l := w.Layout()
	if lane, at, open := w.openedInLane(l); open {
		if w.ownPlace(w.InLane(lane), at) {
			return geom.Box{}, false
		}
		return LineAbove(l.LaneCardAt(lane, at, w.lanes[lane])), true
	}
	if column, at, open := w.openedAt(l); open {
		if w.ownPlace(w.InColumn(column), at) {
			return geom.Box{}, false
		}
		return LineAbove(l.CardAt(column, at, w.scroll[column])), true
	}
	return geom.Box{}, false
}

// OriginShows is how much of a carried thing's own place still shows: faint, so it reads as where it came from.
const OriginShows = 0.25

const landingHigh = 3

// Lines is the history as the lines it is drawn on, wrapped to the room the panel gives it.
func (w Window) Lines() []string {
	inside := Inside(BesideBox(w.formPanel()))
	wrap := func(said string) Lines { return w.wrapping(said, inside, small) }
	if w.ChosenGoal {
		return wrap("an objective has no trace of its own").Said
	}
	lines := Told(w.Acts, wrap)
	if before := w.SeriesBefore(); len(before) > 0 && len(w.Acts) > 0 {
		lines = append(lines, wrap(BeforeThisOneSays).Said...)
		for _, said := range before {
			lines = append(lines, wrap(said).Said...)
		}
	}
	return lines
}

// BeforeThisOneSays is what heads, in the history of an occurrence, what its series did before it.
const BeforeThisOneSays = "in the series, before this one"

// ServingSays is what that box reads when there is no card to draw in it.
const ServingSays = "nothing serves it yet"

// Told is how many of those lines there are, and Rows how many are shown at once.
func (w Window) Told() int { return len(w.Lines()) }

// Rows is how many lines of history the panel shows at once.
func (w Window) Rows() int { return BesideRows(w.formPanel()) }

// History is the first line of the history that is shown.
func (w Window) History() int { return w.history }

// ChoosingRoom is how many rows the open list shows at once.
func (w Window) ChoosingRoom() int { return ChoosingRows(w.ChoosingBox()) }

func (w Window) settledChoosing(want int) int {
	return min(max(0, len(w.Choosing())-w.ChoosingRoom()), max(0, want))
}

// Saying is how far the description on an open task has been scrolled.
func (w Window) Saying() int { return w.settledSaying(w.saying) }

// SaidLines is the description as it is laid out in its box, and SaidRows how many of them are in sight.
func (w Window) SaidLines() []string {
	_, _, lines, _, _ := w.sayingOn()
	return lines.Said
}

func (w Window) sayingOn() (frame, bar geom.Box, lines Lines, rows int, ok bool) {
	switch w.Screen {
	case board.OnForm:
		at, long := w.SayingAt()
		if !long {
			return geom.Box{}, geom.Box{}, Lines{}, 0, false
		}
		fields := w.Form.Fields
		var value geom.Box
		_, value, rows = FormPlace(w.formBox(), fields, at)
		frame, bar = Saying(value, rows)
		said := fields[at].Value
		if at == w.Form.On() {
			said = w.Form.Writing().Reads()
		}
		return frame, bar, w.wrapping(said, Inside(frame), reads), rows, true
	}
	return geom.Box{}, geom.Box{}, Lines{}, 0, false
}

// SaidRows is how many lines of the description its box holds.
func (w Window) SaidRows() int {
	_, _, _, rows, _ := w.sayingOn()
	return rows
}

func (w Window) settledSaying(want int) int {
	return min(max(0, len(w.SaidLines())-w.SaidRows()), max(0, want))
}

func (w Window) settledHistory(want int) int {
	return min(max(0, w.Told()-w.Rows()), max(0, want))
}

func (w Window) settledServing(want int) int {
	return min(max(0, (len(w.Under())+1)*geom.SubHeight-w.workFrame().H), max(0, want))
}

// ServingCardAt is where one of those tiles is drawn, counting the scroll — the same reckoning a column makes of its cards.
func (w Window) ServingCardAt(at int) geom.Box {
	box := w.workFrame()
	return geom.Box{X: box.X, Y: box.Y + at*geom.SubHeight - w.serving, W: box.W, H: geom.SubHeight}
}

func (w Window) descriptionFrame() geom.Box {
	fields := w.Form.Fields
	for i, f := range fields {
		if f.Name == "description" {
			_, value, rows := FormPlace(w.formBox(), fields, i)
			return Framed(value, rows)
		}
	}
	return geom.Box{}
}

// WorkName is where an objective's form names the work serving it: under its description.
func (w Window) WorkName() geom.Box {
	said := w.descriptionFrame()
	return geom.Box{X: said.X, Y: said.Y + said.H + between, W: said.W, H: lineHigh}
}

func (w Window) workFrame() geom.Box {
	name := w.WorkName()
	top := name.Y + name.H + between
	return geom.Box{X: name.X, Y: top, W: name.W, H: max(lineHigh, FormRule(w.formBox()).Y-between-top)}
}

func (w Window) workBar() geom.Box {
	frame := w.workFrame()
	return geom.Box{X: frame.X + frame.W + between, Y: frame.Y, W: barWide, H: frame.H}
}

// WorkTile is one line of the work an objective's form lists: a task serving it, or the line that adds one.
type WorkTile struct {
	Said   string
	Number int
	At     geom.Box
}

func workTileAt(tile WorkTile) geom.Box {
	panel := SubPanel(tile.At)
	return geom.Box{X: panel.X + inset, Y: panel.Y + 1, W: panel.W - 2*inset, H: panel.H - 2}
}

// WorkTiles is the lines of the work an objective's form lists, where each is drawn: the tasks serving it, then the line that adds one.
func (w Window) WorkTiles() []WorkTile {
	if !w.BesideShown() || !w.Form.ForObjective {
		return nil
	}
	var tiles []WorkTile
	for at, c := range w.Under() {
		tiles = append(tiles, WorkTile{Said: c.Title, Number: c.Number, At: w.ServingCardAt(at)})
	}
	return append(tiles, WorkTile{Said: AddingServingSays, At: w.ServingCardAt(len(tiles))})
}

// AddingServingSays is the last line of an objective's work, which adds a task serving it.
const AddingServingSays = "+ New task serving it"

func (w Window) clickingWork(p geom.Point) (Window, bool) {
	if !w.workFrame().In(p) {
		return w, false
	}
	for _, tile := range w.WorkTiles() {
		if tile.At.In(p) {
			return w.openingWork(tile), true
		}
	}
	return w, false
}

func (w Window) openingWork(tile WorkTile) Window {
	if w.Form.Changed() {
		return w.telling("save or cancel the objective first")
	}
	goal := w.Form.Number
	if tile.Number == 0 {
		w = w.startingNew()
		w.Form, w.formMoved = board.AsOpened(board.Form{Adding: true, Fields: board.ServingTaskFields(goal)}), geom.Point{}
		return w
	}
	w.Chosen, w.ChosenGoal, w.Screen = tile.Number, false, board.OnTheBoard
	return w.Edit()
}

// WorkOn is the line of an objective's work the keyboard is on, and whether it is on one.
func (w Window) WorkOn() (WorkTile, bool) {
	tiles := w.WorkTiles()
	if !w.Form.OnWork || w.Form.WorkAt >= len(tiles) {
		return WorkTile{}, false
	}
	return tiles[w.Form.WorkAt], true
}

func (w Window) onWork(name string) (Window, bool) {
	tiles := w.WorkTiles()
	switch name {
	case "down", "j":
		w.Form.WorkAt = min(len(tiles)-1, w.Form.WorkAt+1)
	case "up", "k":
		w.Form.WorkAt = max(0, w.Form.WorkAt-1)
	case "enter":
		if tile, on := w.WorkOn(); on {
			return w.openingWork(tile), true
		}
	case "tab":
		w.Form = w.Form.Onto(w.Form.On() + 1)
		return w, true
	case "shift+tab":
		w.Form.OnWork = false
		return w, true
	default:
		return w, true
	}
	frame := w.workFrame()
	top := w.Form.WorkAt * geom.SubHeight
	w.serving = w.settledServing(min(top, max(w.serving, top+geom.SubHeight-frame.H)))
	return w, true
}

func (w Window) panel() geom.Box {
	switch w.Screen {
	case board.OnForm:
		return w.formPanel()
	case board.OnChoosing:
		return w.ChoosingBox()
	case board.OnAsking:
		return w.QuestionBox()
	case board.OnFates:
		return w.FatesAsked().Box
	case board.OnSeriesFates:
		return w.SeriesFatesAsked().Box
	case board.OnHelp:
		return w.HelpBox()
	case board.OnCalendar:
		return w.CalendarBox()
	case board.OnRule:
		return w.RulePicked().Box
	case board.OnMenu:
		return w.MenuBox()
	case board.OnBackups:
		return w.BackupsBox()
	case board.OnLabels:
		return w.LabelsBox()
	case board.OnFiltering:
		return w.FiltersBox()
	case board.OnConfiguration:
		return w.SettingsBox()
	case board.OnKeeping:
		return w.KeepingBox()
	case board.OnViews:
		return w.ViewsBox()
	}
	return Over(w.width, w.height)
}

func (w Window) formPanel() geom.Box {
	return Shifted(w.formOpensIn(), w.formMoved, w.width, w.height)
}

func (w Window) formOpensIn() geom.Box {
	over := Over(w.width, w.height)
	if given := w.sizes.Form; given != ([2]int{}) {
		over.W = max(FormLeast.X, min(given[0], w.width-over.X-geom.Margin))
		over.H = max(FormLeast.Y, min(given[1], w.height-over.Y-geom.Margin))
	}
	return over
}

// FormLeast is the smallest a form is resized to: room for its title, a few lines of description and of steps, and its short fields beside them.
var FormLeast = geom.Point{X: 480, Y: 400}

// FormGrip is the corner of a form that resizes it when dragged.
func FormGrip(panel geom.Box) geom.Box {
	return geom.Box{X: panel.X + panel.W - gripSide - inset, Y: panel.Y + panel.H - gripSide - inset, W: gripSide, H: gripSide}
}

const gripSide = 14

func (w Window) resizingTheForm(p geom.Point) Window {
	from := w.resizedFrom
	w.sizes.Form = [2]int{
		max(FormLeast.X, from.W+p.X-w.held.From.X),
		max(FormLeast.Y, from.H+p.Y-w.held.From.Y),
	}
	return w
}

func (w Window) movingForm(p geom.Point) Window {
	opens := w.formOpensIn()
	moved := Shifted(opens, geom.Point{X: w.formMoved.X + p.X - w.formGrip.X, Y: w.formMoved.Y + p.Y - w.formGrip.Y}, w.width, w.height)
	w.formMoved, w.formGrip = geom.Point{X: moved.X - opens.X, Y: moved.Y - opens.Y}, p
	if abs(p.X-w.held.From.X) > Wobble || abs(p.Y-w.held.From.Y) > Wobble {
		w.dragged = true
	}
	return w
}

func (w Window) formBox() geom.Box {
	if !w.BesideShown() {
		return w.formPanel()
	}
	left := FormFields(w.formPanel())
	left.Beside = FormBeside(w.formPanel()).W
	return left
}

// ServingAt is how far that box has been scrolled.
func (w Window) ServingAt() int { return w.serving }

// Says is what a window too small to draw in says instead.
func (w Window) Says() string {
	if w.Screen != board.TooSmallToDraw {
		return ""
	}
	return fmt.Sprintf(
		"This window is %d by %d. The board needs %d by %d.",
		w.width, w.height, geom.Narrowest, geom.Shortest)
}

func (w Window) scrollingSaying(p geom.Point, by int) Window {
	frame, bar, _, _, ok := w.sayingOn()
	if !ok || !frame.In(p) && !bar.In(p) {
		return w
	}
	lines, left := turning(w.spareSaying, by)
	w.saying, w.spareSaying = w.settledSaying(w.saying+lines), left
	return w
}

func (w Window) walksLines(by int) Window {
	_, _, lines, _, ok := w.sayingOn()
	if !ok {
		return w
	}
	writing := w.Form.Writing()
	line, column := lines.At(writing.At)
	to := min(len(lines.Said)-1, max(0, line+by))
	at := lines.From[to] + min(column, len([]rune(lines.Said[to])))
	w.Form = w.Form.Writes(writing.CaretAt(at))
	return w.inSightOfTheCaret()
}

func (w Window) togglingTrace() Window {
	w.Form.Tracing = !w.Form.Tracing
	if w.Form.Tracing && w.OnASectionField() {
		for i, field := range w.Form.Fields {
			if field.Name == "description" {
				w.Form = w.Form.Onto(i)
			}
		}
	}
	return w
}

// BesideTabs is the two tabs at the head of the column beside a form: what it carries, and what was done to it.
func BesideTabs(panel geom.Box) [2]geom.Box {
	head := BesideHead(panel)
	half := head.W / 2
	return [2]geom.Box{{X: head.X, Y: head.Y - inset, W: half, H: head.H + 2*inset}, {X: head.X + half, Y: head.Y - inset, W: head.W - half, H: head.H + 2*inset}}
}

func (w Window) clickingBesideTab(p geom.Point) (Window, bool) {
	if w.Screen != board.OnForm || !w.TabsBeside() {
		return w, false
	}
	for i, tab := range BesideTabs(w.formPanel()) {
		if tab.In(p) {
			if (i == 1) != w.Tracing() {
				w = w.togglingTrace()
			}
			return w, true
		}
	}
	return w, false
}

// The heads an objective's form writes over its details, its work and its progress.
const (
	DetailsHead  = "Details"
	ServingHead  = "Serving it"
	ProgressHead = "Progress"
)

// ProgressBox is where an objective's form says how far it has got: in the column beside, under its details.
func (w Window) ProgressBox() geom.Box {
	fields := w.Form.Fields
	bottom := FirstRow(w.formPanel())
	for i, f := range fields {
		if _, sectional := board.SectionOf(f); !sectional {
			continue
		}
		_, value, rows := FormPlace(w.formBox(), fields, i)
		frame := Framed(value, rows)
		bottom = max(bottom, frame.Y+frame.H)
	}
	right := FormBeside(w.formPanel())
	g, _ := w.Snap.Goal(w.Form.Number)
	high := progressHead + w.GoalProgress(g).High()
	box := geom.Box{X: right.X + padding, Y: bottom + padding, W: right.W - 2*padding, H: high}
	if along := w.offersAlong(w.formPanel()); len(along) > 0 {
		box.H = max(0, min(box.H, along[0].Y-between-padding-box.Y))
	}
	return box
}

const progressHead = lineHigh + between + 2*padding + lineHigh

// GoalProgress is what an objective's form says of how far it has got: what is done, the time spent against the estimate, what is left, where else it is counted, and the days to its deadline.
func (w Window) GoalProgress(g board.Goal) Facts {
	f := Facts{Lines: []Fact{{Name: "done", Value: strconv.Itoa(g.Done) + " of " + strconv.Itoa(g.Total) + " tasks"}}}
	if w.Pages.Shows(engine.PageTime) {
		time := Fact{Name: "time spent", Value: w.SpentSaid(calendar.FormatDuration(g.Spent))}
		if g.Estimate != "" {
			time.Aside = "of " + g.Estimate + " estimated"
		}
		f.Lines = append(f.Lines, time)
	}
	f.Lines = append(f.Lines, Fact{Name: "left", Value: w.Snap.Left(g.Number).Line()})
	if parent, ok := w.Snap.Goal(g.Parent); ok && g.Parent != 0 {
		f.Lines = append(f.Lines, Fact{Name: "counted in", Value: "#" + strconv.Itoa(parent.Number) + " " + parent.Title,
			Aside: strconv.Itoa(parent.Done) + " of " + strconv.Itoa(parent.Total) + " there, each task once"})
	}
	if !g.Deadline.IsZero() {
		f.Lines = append(f.Lines, Fact{Name: "deadline", Value: calendar.FormatDay(g.Deadline) + ", " + daysTo(board.DayOf(g.Deadline), w.Today())})
	}
	return f
}

func daysTo(day, today time.Time) string {
	days := int(day.Sub(today).Hours() / 24)
	switch {
	case days == 0:
		return "today"
	case days < 0:
		return strconv.Itoa(-days) + " day" + plural(-days) + " past"
	}
	return "in " + strconv.Itoa(days) + " day" + plural(days)
}
