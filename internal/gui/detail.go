package gui

import (
	"cairn/internal/domain"
	"cairn/internal/engine"
	"fmt"
	"strconv"
	"time"
)

// Ghost is the card in mid-air under the pointer, and whether there is one.
func (w Window) Ghost() (Box, bool) {
	if !w.Dragging() {
		return Box{}, false
	}
	l := w.Layout()
	return Box{
		X: w.held.At.X - l.Column[0].W/2,
		Y: w.held.At.Y - CardHeight/2,
		W: l.Column[0].W,
		H: CardHeight,
	}, true
}

// Landing is the room the card would drop into, a card high, the cards after it moved down to make it; and whether it is over one at all.
func (w Window) Landing() (Box, bool) {
	if !w.Dragging() {
		return Box{}, false
	}
	l := w.Layout()
	if lane, at, open := w.openedInLane(l); open {
		return l.LaneCardAt(lane, at, w.lanes[lane]), true
	}
	if column, at, open := w.openedAt(l); open {
		return l.CardAt(column, at, w.scroll[column]), true
	}
	return Box{}, false
}

const landingHigh = 3

// Lines is the history as the lines it is drawn on, wrapped to the room the panel gives it.
func (w Window) Lines() []string {
	inside := Inside(BesideBox(w.formPanel()))
	wrap := func(said string) Lines { return w.wrapping(said, inside, small) }
	if w.chosenGoal {
		return wrap("an objective has no trace of its own").Said
	}
	lines := Told(w.acts, wrap)
	if before := w.seriesBefore(); len(before) > 0 && len(w.acts) > 0 {
		lines = append(lines, wrap(BeforeThisOneSays).Said...)
		for _, said := range before {
			lines = append(lines, wrap(said).Said...)
		}
	}
	return lines
}

// BeforeThisOneSays is what heads, in the history of an occurrence, what its series did before it.
const BeforeThisOneSays = "in the series, before this one"

func (w Window) seriesBefore() []string {
	c, found := w.snap.Card(w.chosen)
	if !found || c.Series == nil || w.chosenGoal {
		return nil
	}
	s, found := w.snap.SeriesOf(c.Series.Series)
	if !found {
		return nil
	}
	return engine.BeforeIn(s, c.Number)
}

// ServingSays is what that box reads when there is no card to draw in it.
const ServingSays = "nothing serves it yet"

// Told is how many of those lines there are, and Rows how many are shown at once.
func (w Window) Told() int { return len(w.Lines()) }

// Rows is how many lines of history the panel shows at once.
func (w Window) Rows() int { return BesideRows(w.formPanel()) }

// History is the first line of the history that is shown.
func (w Window) History() int { return w.history }

// ChoosingFromRow is the first row of the open list in sight, and ChoosingRoom how many of them are in sight at once.
func (w Window) ChoosingFromRow() int { return w.choosingFrom }

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

func (w Window) sayingOn() (frame, bar Box, lines Lines, rows int, ok bool) {
	switch w.screen {
	case OnForm:
		at, long := w.sayingAt()
		if !long {
			return Box{}, Box{}, Lines{}, 0, false
		}
		fields := w.form.Fields()
		var value Box
		_, value, rows = FormPlace(w.formBox(), fields, at)
		frame, bar = Saying(value, rows)
		said := fields[at].Value
		if at == w.form.On() {
			said = w.form.Writing().Reads()
		}
		return frame, bar, w.wrapping(said, Inside(frame), reads), rows, true
	}
	return Box{}, Box{}, Lines{}, 0, false
}

func (w Window) sayingAt() (int, bool) {
	fields := w.form.Fields()
	if len(fields) == 0 {
		return 0, false
	}
	if fields[w.form.On()].Long {
		return w.form.On(), true
	}
	for i, field := range fields {
		if field.Long {
			return i, true
		}
	}
	return 0, false
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
	return min(max(0, (len(w.Under())+1)*SubHeight-w.workFrame().H), max(0, want))
}

// ServingCardAt is where one of those tiles is drawn, counting the scroll — the same reckoning a column makes of its cards.
func (w Window) ServingCardAt(at int) Box {
	box := w.workFrame()
	return Box{X: box.X, Y: box.Y + at*SubHeight - w.serving, W: box.W, H: SubHeight}
}

// TabsBeside says whether the column beside the form is split in two tabs: a task's, whose history is behind one; an objective's holds its details alone.
func (w Window) TabsBeside() bool {
	return w.besideShown() && !w.form.ForObjective && !w.form.ForSeries
}

func (w Window) descriptionFrame() Box {
	fields := w.form.Fields()
	for i, f := range fields {
		if f.Name == "description" {
			_, value, rows := FormPlace(w.formBox(), fields, i)
			return Framed(value, rows)
		}
	}
	return Box{}
}

// WorkName is where an objective's form names the work serving it: under its description.
func (w Window) WorkName() Box {
	said := w.descriptionFrame()
	return Box{X: said.X, Y: said.Y + said.H + between, W: said.W, H: lineHigh}
}

func (w Window) workFrame() Box {
	name := w.WorkName()
	top := name.Y + name.H + between
	return Box{X: name.X, Y: top, W: name.W, H: max(lineHigh, FormRule(w.formBox()).Y-between-top)}
}

func (w Window) workBar() Box {
	frame := w.workFrame()
	return Box{X: frame.X + frame.W + between, Y: frame.Y, W: barWide, H: frame.H}
}

// WorkTile is one line of the work an objective's form lists: a task serving it, or the line that adds one.
type WorkTile struct {
	Said   string
	Number int
	At     Box
}

func workTileAt(tile WorkTile) Box {
	panel := SubPanel(tile.At)
	return Box{X: panel.X + inset, Y: panel.Y + 1, W: panel.W - 2*inset, H: panel.H - 2}
}

// WorkTiles is the lines of the work an objective's form lists, where each is drawn: the tasks serving it, then the line that adds one.
func (w Window) WorkTiles() []WorkTile {
	if !w.besideShown() || !w.form.ForObjective {
		return nil
	}
	var tiles []WorkTile
	for at, c := range w.Under() {
		tiles = append(tiles, WorkTile{Said: c.Title, Number: c.Number, At: w.ServingCardAt(at)})
	}
	return append(tiles, WorkTile{Said: AddingServingSays, At: w.ServingCardAt(len(tiles))})
}

// AddingServingSays is the last line of an objective's work, which adds a task serving it.
const AddingServingSays = "+ new task serving it"

func (w Window) clickingWork(p Point) (Window, bool) {
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
	if w.form.Changed() {
		return w.telling("save or cancel the objective first")
	}
	goal := w.form.Number
	if tile.Number == 0 {
		w = w.startingNew()
		w.form, w.formMoved = asOpened(Form{Adding: true, fields: engine.ServingTaskFields(goal)}), Point{}
		return w
	}
	w.chosen, w.chosenGoal, w.screen = tile.Number, false, OnTheBoard
	return w.Edit()
}

// WorkOn is the line of an objective's work the keyboard is on, and whether it is on one.
func (w Window) WorkOn() (WorkTile, bool) {
	tiles := w.WorkTiles()
	if !w.form.onWork || w.form.workAt >= len(tiles) {
		return WorkTile{}, false
	}
	return tiles[w.form.workAt], true
}

func (w Window) onWork(name string) (Window, bool) {
	tiles := w.WorkTiles()
	switch name {
	case "down", "j":
		w.form.workAt = min(len(tiles)-1, w.form.workAt+1)
	case "up", "k":
		w.form.workAt = max(0, w.form.workAt-1)
	case "enter":
		if tile, on := w.WorkOn(); on {
			return w.openingWork(tile), true
		}
	case "tab":
		w.form = w.form.onto(w.form.On() + 1)
		return w, true
	case "shift+tab":
		w.form.onWork = false
		return w, true
	default:
		return w, true
	}
	frame := w.workFrame()
	top := w.form.workAt * SubHeight
	w.serving = w.settledServing(min(top, max(w.serving, top+SubHeight-frame.H)))
	return w, true
}

func (w Window) reachingWork() bool {
	fields := w.form.Fields()
	return w.form.ForObjective && w.besideShown() && w.form.On() < len(fields) && fields[w.form.On()].Name == "description"
}

func (w Window) panel() Box {
	switch w.screen {
	case OnForm:
		return w.formPanel()
	case OnChoosing:
		return w.ChoosingBox()
	case OnAsking:
		return w.QuestionBox()
	case OnFates:
		return w.FatesAsked().Box
	case OnSeriesFates:
		return w.SeriesFatesAsked().Box
	case OnHelp:
		return w.HelpBox()
	case OnCalendar:
		return w.CalendarBox()
	case OnRule:
		return w.RulePicked().Box
	case OnMenu:
		return w.MenuBox()
	case OnBackups:
		return w.BackupsBox()
	case OnLabels:
		return w.LabelsBox()
	case OnFiltering:
		return w.FiltersBox()
	case OnConfiguration:
		return w.SettingsBox()
	case OnKeeping:
		return w.KeepingBox()
	case OnViews:
		return w.ViewsBox()
	}
	return Over(w.width, w.height)
}

func (w Window) formPanel() Box {
	return Shifted(w.formOpensIn(), w.formMoved, w.width, w.height)
}

func (w Window) formOpensIn() Box {
	over := Over(w.width, w.height)
	if given := w.sizes.Form; given != ([2]int{}) {
		over.W = max(FormLeast.X, min(given[0], w.width-over.X-margin))
		over.H = max(FormLeast.Y, min(given[1], w.height-over.Y-margin))
	}
	return over
}

// FormLeast is the smallest a form is resized to: room for its title, a few lines of description and of steps, and its short fields beside them.
var FormLeast = Point{X: 480, Y: 400}

// FormGrip is the corner of a form that resizes it when dragged.
func FormGrip(panel Box) Box {
	return Box{X: panel.X + panel.W - gripSide - inset, Y: panel.Y + panel.H - gripSide - inset, W: gripSide, H: gripSide}
}

const gripSide = 14

func (w Window) resizingTheForm(p Point) Window {
	from := w.resizedFrom
	w.sizes.Form = [2]int{
		max(FormLeast.X, from.W+p.X-w.held.From.X),
		max(FormLeast.Y, from.H+p.Y-w.held.From.Y),
	}
	return w
}

func (w Window) movingForm(p Point) Window {
	opens := w.formOpensIn()
	moved := Shifted(opens, Point{X: w.formMoved.X + p.X - w.formGrip.X, Y: w.formMoved.Y + p.Y - w.formGrip.Y}, w.width, w.height)
	w.formMoved, w.formGrip = Point{X: moved.X - opens.X, Y: moved.Y - opens.Y}, p
	if abs(p.X-w.held.From.X) > Wobble || abs(p.Y-w.held.From.Y) > Wobble {
		w.dragged = true
	}
	return w
}

func (w Window) formBox() Box {
	if !w.besideShown() {
		return w.formPanel()
	}
	left := FormFields(w.formPanel())
	left.Beside = FormBeside(w.formPanel()).W
	return left
}

func (w Window) besideShown() bool { return !w.form.Adding || !w.form.ForObjective }

func (w Window) closesOnAClickOff() bool {
	switch w.screen {
	case OnHelp, OnAsking, OnChoosing:
		return true
	case OnForm:
		return true
	}
	return false
}

// ServingAt is how far that box has been scrolled.
func (w Window) ServingAt() int { return w.serving }

// Reading is the work as it was last read, which is what everything draws from.
func (w Window) Reading() engine.Snapshot { return w.snap }

// Says is what a window too small to draw in says instead.
func (w Window) Says() string {
	if w.screen != TooSmallToDraw {
		return ""
	}
	return fmt.Sprintf(
		"This window is %d by %d. The board needs %d by %d.",
		w.width, w.height, Narrowest, Shortest)
}

func (w Window) scrollingSaying(p Point, by int) Window {
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
	writing := w.form.Writing()
	line, column := lines.At(writing.At)
	to := min(len(lines.Said)-1, max(0, line+by))
	at := lines.From[to] + min(column, len([]rune(lines.Said[to])))
	w.form = w.form.Writes(writing.CaretAt(at))
	return w.inSightOfTheCaret()
}

// Tracing says whether the column beside an open form shows what was done to it — its history, or the work serving an objective — rather than what it carries: chosen with its tab, and given up for the details when the keyboard goes to one of them.
func (w Window) Tracing() bool { return w.form.tracing && w.TabsBeside() }

func (w Window) onASectionField() bool {
	fields := w.form.Fields()
	if w.form.On() >= len(fields) {
		return false
	}
	_, sectional := sectionOf(fields[w.form.On()])
	return sectional
}

func (w Window) togglingTrace() Window {
	w.form.tracing = !w.form.tracing
	if w.form.tracing && w.onASectionField() {
		for i, field := range w.form.Fields() {
			if field.Name == "description" {
				w.form = w.form.onto(i)
			}
		}
	}
	return w
}

// BesideTabs is the two tabs at the head of the column beside a form: what it carries, and what was done to it.
func BesideTabs(panel Box) [2]Box {
	head := BesideHead(panel)
	half := head.W / 2
	return [2]Box{{X: head.X, Y: head.Y - inset, W: half, H: head.H + 2*inset}, {X: head.X + half, Y: head.Y - inset, W: head.W - half, H: head.H + 2*inset}}
}

// BesideTabsSay is what the two tabs are called.
func (w Window) BesideTabsSay() [2]string {
	if w.form.ForObjective {
		return [2]string{"Details", "Work"}
	}
	return [2]string{"Details", "History"}
}

func (w Window) clickingBesideTab(p Point) (Window, bool) {
	if w.screen != OnForm || !w.TabsBeside() {
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

func (w Window) hidden(field engine.Field) bool {
	_, sectional := sectionOf(field)
	return sectional && w.Tracing()
}

// The heads an objective's form writes over its details, its work and its progress.
const (
	DetailsHead  = "Details"
	ServingHead  = "Serving it"
	ProgressHead = "Progress"
)

// ProgressBox is where an objective's form says how far it has got: in the column beside, under its details.
func (w Window) ProgressBox() Box {
	fields := w.form.Fields()
	bottom := FirstRow(w.formPanel())
	for i, f := range fields {
		if _, sectional := sectionOf(f); !sectional {
			continue
		}
		_, value, rows := FormPlace(w.formBox(), fields, i)
		frame := Framed(value, rows)
		bottom = max(bottom, frame.Y+frame.H)
	}
	right := FormBeside(w.formPanel())
	g, _ := w.snap.Goal(w.form.Number)
	high := progressHead + w.GoalProgress(g).High()
	box := Box{X: right.X + padding, Y: bottom + padding, W: right.W - 2*padding, H: high}
	if along := w.offersAlong(w.formPanel()); len(along) > 0 {
		box.H = max(0, min(box.H, along[0].Y-between-padding-box.Y))
	}
	return box
}

const progressHead = lineHigh + between + 2*padding + lineHigh

// GoalProgress is what an objective's form says of how far it has got: what is done, the time spent against the estimate, what is left, where else it is counted, and the days to its deadline.
func (w Window) GoalProgress(g engine.Goal) Facts {
	spent := "—"
	if g.Spent > 0 {
		spent = domain.FormatDuration(g.Spent)
	}
	time := Fact{Name: "time spent", Value: spent}
	if g.Estimate != "" {
		time.Aside = "of " + g.Estimate + " estimated"
	}
	f := Facts{Lines: []Fact{
		{Name: "done", Value: strconv.Itoa(g.Done) + " of " + strconv.Itoa(g.Total) + " tasks"},
		time,
		{Name: "left", Value: w.snap.Left(g.Number).Line()},
	}}
	if parent, ok := w.snap.Goal(g.Parent); ok && g.Parent != 0 {
		f.Lines = append(f.Lines, Fact{Name: "counted in", Value: "#" + strconv.Itoa(parent.Number) + " " + parent.Title,
			Aside: strconv.Itoa(parent.Done) + " of " + strconv.Itoa(parent.Total) + " there, each task once"})
	}
	if !g.Deadline.IsZero() {
		f.Lines = append(f.Lines, Fact{Name: "deadline", Value: domain.FormatDay(g.Deadline) + ", " + daysTo(engine.DayOf(g.Deadline), w.today())})
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
