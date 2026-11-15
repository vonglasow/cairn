package board

import (
	"cairn/internal/engine"
	"cairn/internal/engine/label"
	"cairn/internal/engine/series"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Kind is what is chosen: a task or an objective.
func (b Board) Kind() string {
	if b.ChosenGoal {
		return "objective"
	}
	return "task"
}

// FormKind is what the form adds or edits: a task, an objective or a series.
func (b Board) FormKind() string {
	if b.Form.ForObjective {
		return "objective"
	}
	if b.Form.ForSeries {
		return "series"
	}
	return "task"
}

// CommandSays is what the command bar's line says at rest: the task chosen, by its number and title, or what can be typed into it.
func (b Board) CommandSays() string {
	if b.Chosen != 0 && !b.ChosenGoal {
		if c, ok := b.Snap.Card(b.Chosen); ok {
			return "#" + strconv.Itoa(c.Number) + " · " + c.Title
		}
	}
	return "Command, number or title…"
}

// SavedTitle is the title held by what the form edits, as it was saved.
func (b Board) SavedTitle() string {
	if b.Form.ForObjective {
		g, _ := b.Snap.Goal(b.Form.Number)
		return g.Title
	}
	if b.Form.ForSeries {
		s, _ := b.Snap.SeriesOf(b.Form.Number)
		return s.Title
	}
	c, _ := b.Snap.Card(b.Form.Number)
	return c.Title
}

// Sweeps is the column a sweep from the chosen task acts on, and nothing when it may not sweep.
func (b Board) Sweeps() string {
	if b.Chosen == 0 || b.ChosenGoal {
		return ""
	}
	state := b.StateOf(b.Chosen)
	if Stopped(state) {
		return ""
	}
	return state
}

// Searched is the search as a written text, caret and all.
func (b Board) Searched() Written { return b.Looking }

// Blank says every page is turned off: the window is drawn, and nothing in it.
func (b Board) Blank() bool { return b.Pages.None() }

// Stranded says whether the board has nothing to show and no settings open.
func (b Board) Stranded() bool { return b.Blank() && b.Screen != OnConfiguration }

// Today is the day the board stands on.
func (b Board) Today() time.Time { return DayOf(b.Now) }

// SeriesBefore is what the series of the chosen occurrence did before it, line by line.
func (b Board) SeriesBefore() []string {
	c, found := b.Snap.Card(b.Chosen)
	if !found || c.Series == nil || b.ChosenGoal {
		return nil
	}
	s, found := b.Snap.SeriesOf(c.Series.Series)
	if !found {
		return nil
	}
	return BeforeIn(s, c.Number)
}

// SayingAt is the long field the form writes in, and whether it has one.
func (b Board) SayingAt() (int, bool) {
	fields := b.Form.Fields
	if len(fields) == 0 {
		return 0, false
	}
	if fields[b.Form.On()].Long {
		return b.Form.On(), true
	}
	for i, field := range fields {
		if field.Long {
			return i, true
		}
	}
	return 0, false
}

// TabsBeside says whether the column beside the form is split in two tabs: a task's, whose history is behind one; an objective's holds its details alone.
func (b Board) TabsBeside() bool {
	return b.BesideShown() && !b.Form.ForObjective && !b.Form.ForSeries
}

// ReachingWork says whether the keyboard is on an objective's description, where the work it is made of is reached.
func (b Board) ReachingWork() bool {
	fields := b.Form.Fields
	return b.Form.ForObjective && b.BesideShown() && b.Form.On() < len(fields) && fields[b.Form.On()].Name == "description"
}

// BesideShown says whether the form shows what stands beside its fields: every form but a new objective's.
func (b Board) BesideShown() bool { return !b.Form.Adding || !b.Form.ForObjective }

// ClosesOnAClickOff says whether a click outside what is open closes it.
func (b Board) ClosesOnAClickOff() bool {
	switch b.Screen {
	case OnHelp, OnAsking, OnChoosing:
		return true
	case OnForm:
		return true
	}
	return false
}

// Reading is the work as it was last read, which is what everything draws from.
func (b Board) Reading() Snapshot { return b.Snap }

// Tracing says whether the column beside an open form shows what was done to it — its history, or the work serving an objective — rather than what it carries: chosen with its tab, and given up for the details when the keyboard goes to one of them.
func (b Board) Tracing() bool { return b.Form.Tracing && b.TabsBeside() }

// OnASectionField says whether the field the form is on is laid out under a section.
func (b Board) OnASectionField() bool {
	fields := b.Form.Fields
	if b.Form.On() >= len(fields) {
		return false
	}
	_, sectional := SectionOf(fields[b.Form.On()])
	return sectional
}

// BesideTabsSay is what the two tabs are called.
func (b Board) BesideTabsSay() [2]string {
	if b.Form.ForObjective {
		return [2]string{"Details", "Work"}
	}
	return [2]string{"Details", "Trace"}
}

// Hidden says whether a field is out of sight while the trace is shown.
func (b Board) Hidden(field Field) bool {
	_, sectional := SectionOf(field)
	return sectional && b.Tracing()
}

// FilteredBy is what the board is filtered by, written the way the bar says it: nothing when no filter narrows it.
func (b Board) FilteredBy() string { return b.Filters.Said() }

// GoalOfValue is the objective a filter value names, and whether there is one.
func (b Board) GoalOfValue(value string) (Goal, bool) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return Goal{}, false
	}
	return b.Snap.Goal(n)
}

// UnreachedTitles is the titles of the objectives one waits on that are not reached.
func (b Board) UnreachedTitles(g Goal) []string {
	var titles []string
	for _, n := range g.WaitsOn {
		if on, ok := b.Snap.Goal(n); ok && on.State != "reached" {
			titles = append(titles, on.Title)
		}
	}
	return titles
}

// HiddenSays is what the bubble over a folded parent's signal says: what it hides, each whole, and how to see it.
func (b Board) HiddenSays(n int) []string {
	g, _ := b.Snap.Goal(n)
	lines := []string{"Hidden under " + g.Title + ":"}
	for _, h := range b.Snap.Hidden(n, b.Today()) {
		named := "#" + strconv.Itoa(h.Goal.Number) + " " + h.Goal.Title
		if h.Late {
			lines = append(lines, named+", its deadline "+series.ShortDay(h.Goal.Deadline)+" past, not reached.")
			continue
		}
		lines = append(lines, named+" waits on #"+strconv.Itoa(h.On.Number)+" "+h.On.Title+", not reached.")
	}
	return append(lines, "→ or a click on its chevron shows it.")
}

// OnAClosedSet says whether the field the form is on takes one of a closed set of values.
func (b Board) OnAClosedSet() bool { return len(b.Form.Fields[b.Form.On()].Choices) > 0 }

// Saving is the operation saving the form asks, and what it is asked with.
func (b Board) Saving() (engine.Operation, engine.Input) {
	adding, editing := engine.TaskNew, engine.TaskEdit
	if b.Form.ForObjective {
		adding, editing = engine.ObjectiveNew, engine.ObjectiveEdit
	}
	if b.Form.ForSeries {
		editing = engine.SeriesEdit
	}
	op, asked := Saved(b.Form.Fields, b.Form.Adding, b.Form.Number, adding, editing)
	if b.Form.Whole {
		asked.Given["series"] = "true"
	}
	return op, asked
}

// FieldSays is what a field reads on screen, which is not always what it carries: the engine names the numbers a field holds.
func (b Board) FieldSays(field Field) string { return b.Snap.Named(field) }

// ClosedSet is the values the field being chosen for may take, and nothing when no list is open.
func (b Board) ClosedSet() []string {
	if b.Screen != OnChoosing {
		return nil
	}
	return b.SetOf(b.Form.Fields[b.Form.On()])
}

// SetOf is the values a field may take.
func (b Board) SetOf(field Field) []string { return field.Choices }

// LabelColour is the colour a label wears on this board: the one the engine read for it, or else the one read off its name.
func (b Board) LabelColour(name string) string {
	if colour, ok := b.Snap.Colours[name]; ok {
		return colour
	}
	return label.ColourOf(name)
}

// ChosenCard is the card chosen, and an empty one when none is.
func (b Board) ChosenCard() Card {
	card, _ := b.Snap.Card(b.Chosen)
	return card
}

// ObjectiveTaken is the one objective the filters narrow to, and whether they narrow to one.
func (b Board) ObjectiveTaken() (string, bool) {
	for _, f := range b.Filters {
		if taken := f.Taken(); f.Arg == "objective" && len(taken) == 1 {
			return taken[0], true
		}
	}
	return "", false
}

// SpentSaid is the time spent as a board writes it: a dash for none, nothing with the time page off.
func (b Board) SpentSaid(spent string) string {
	switch {
	case !b.Pages.Shows(engine.PageTime):
		return ""
	case spent == "":
		return "—"
	}
	return spent
}

// Named is objectives written by number and title, joined.
func (b Board) Named(numbers []int) string {
	var said []string
	for _, n := range numbers {
		g, _ := b.Snap.Goal(n)
		said = append(said, "#"+strconv.Itoa(n)+" "+g.Title)
	}
	return strings.Join(said, ", ")
}

// ServedAmong is the titles of the objectives a task serves among those named, joined.
func (b Board) ServedAmong(c Card, among []int) string {
	var said []string
	for _, g := range b.Snap.Served(c) {
		if slices.Contains(among, g.Number) {
			said = append(said, g.Title)
		}
	}
	return strings.Join(said, ", ")
}

// CommandedTask is the task a typed command acts on: the number it names, or else the one chosen.
func (b Board) CommandedTask(typed string) (int, bool) {
	for _, word := range strings.Fields(typed) {
		if n, err := strconv.Atoi(strings.TrimPrefix(word, "#")); err == nil {
			_, ok := b.Snap.Card(n)
			return n, ok
		}
	}
	return b.Chosen, typed != "" && b.Chosen != 0 && !b.ChosenGoal
}

// ValuesOf is the values a field offers: the columns, the priorities, or the labels carried.
func (b Board) ValuesOf(field string) []string {
	switch field {
	case "state":
		return engine.States()
	case "priority":
		return Priorities()
	case "label":
		return b.BoardLabels()
	}
	return nil
}

// BoardLabels is every label a card carries, each once.
func (b Board) BoardLabels() []string {
	var labels []string
	for _, c := range b.Snap.Cards {
		for _, label := range c.Labels {
			if !slices.Contains(labels, label) {
				labels = append(labels, label)
			}
		}
	}
	slices.Sort(labels)
	return labels
}

// PartTitle is the title of the step the form is on, and nothing on the line adding one.
func (b Board) PartTitle() string {
	parts := b.Form.Parts()
	if b.Form.PartAt >= len(parts) {
		return ""
	}
	return parts[b.Form.PartAt].Title
}

// SeriesWhen is what a series' line says at its right: its next day, or the day it was stopped.
func (b Board) SeriesWhen(s engine.Series) string {
	if s.State == "stopped" {
		return "since " + DaySaid(s.Stopped, b.Now)
	}
	return NextDaySaid(s, b.Now)
}

// NowSaid is what a series has open now, line by line.
func (b Board) NowSaid(s engine.Series) []string {
	if len(s.Open) == 0 {
		return []string{"Nothing open.", "The next is made on its day."}
	}
	c, found := b.Snap.Card(s.Open[0])
	if !found {
		return []string{"#" + strconv.Itoa(s.Open[0])}
	}
	said := c.State
	if text, show, _, _ := OccurrenceDay(c, b.Now); show {
		said += " · " + text
	}
	if c.SubTotal > 0 {
		said += " · " + strconv.Itoa(c.SubDone) + " of " + strconv.Itoa(c.SubTotal) + " steps"
	}
	return []string{"#" + strconv.Itoa(c.Number) + " " + c.Title, said, "One open at a time; the next is made when it is done or skipped."}
}

// AsksWhole says whether saving the form asks first whether the change goes to the whole series: an occurrence edited in what its series carries.
func (b Board) AsksWhole() bool {
	if b.Form.Adding || b.Form.ForObjective || b.Form.ForSeries || b.Form.Answered || !b.Pages.Recurs() {
		return false
	}
	c, found := b.Snap.Card(b.Form.Number)
	return found && ChangesTheSeries(c, b.Form.Fields)
}

// ChangedOfTheSeries is what the form changes that its series carries, named.
func (b Board) ChangedOfTheSeries() string {
	var names []string
	for _, f := range b.Form.Fields {
		if f.Value != f.Was && ChangesTheSeries(Card{Series: &series.Occurs{}}, []Field{f}) {
			names = append(names, f.Name)
		}
	}
	if len(names) < 2 {
		return strings.Join(names, "")
	}
	return strings.Join(names[:len(names)-1], ", ") + " and its " + names[len(names)-1]
}

// OccurrenceChosen is the occurrence of a series chosen, and whether one is, recurrence on.
func (b Board) OccurrenceChosen() (Card, bool) {
	c, found := b.Snap.Card(b.Chosen)
	return c, found && b.Chosen != 0 && !b.ChosenGoal && c.Series != nil && b.Pages.Recurs()
}

// Taken is what is taken together, in the order it was taken: nothing, or two and more.
func (b Board) Taken() []int {
	if len(b.Gathered) < 2 {
		return nil
	}
	return b.Gathered
}

// IsTaken says whether a task or an objective is among those gathered.
func (b Board) IsTaken(number int) bool { return slices.Contains(b.Taken(), number) }

// TitleOf is the title of a task, or of an objective when one is chosen.
func (b Board) TitleOf(n int) string {
	if b.ChosenGoal {
		g, _ := b.Snap.Goal(n)
		return g.Title
	}
	c, _ := b.Snap.Card(n)
	return c.Title
}

// Keepable says whether there is something to keep as a view: a filter on, or a search.
func (b Board) Keepable() bool {
	return b.Filters.Narrows() || !b.Looking.Empty()
}

// ViewsAbove is how many lines the views take at the head of the filter menu: none without a view.
func (b Board) ViewsAbove() int {
	if len(b.Snap.Views) == 0 {
		return 0
	}
	return len(b.Snap.Views) + 2
}

// DayCard is a task of the time page as the board holds it, so it is drawn the way the board draws it; a task the board no longer holds is drawn from what the sheet says of it.
func (b Board) DayCard(one engine.TimesheetTask) Card {
	if c, found := b.Snap.Card(one.Number); found {
		return c
	}
	return Card{Number: one.Number, Title: one.Shown()}
}

// Context is where what was written about the task a form is open on lives elsewhere, or nothing: no form, a new task, an objective, or a window not told where the notes are.
func (b Board) Context() string {
	if b.Screen != OnForm || b.Form.Adding || b.Form.ForObjective || b.Form.ForSeries {
		return ""
	}
	c, _ := b.Snap.Card(b.Form.Number)
	return c.Context
}

// StateOf is the column a task sits in.
func (b Board) StateOf(number int) string {
	c, _ := b.Snap.Card(number)
	return c.State
}

// Under is what the box beside the fields of an open objective holds: the work that serves it.
func (b Board) Under() []Card {
	serving := slices.Clone(Serving(b.Snap.Cards, b.Chosen))
	rank := func(c Card) int {
		switch {
		case c.State == "current":
			return 0
		case Finished(c):
			return 2
		}
		return 1
	}
	slices.SortStableFunc(serving, func(a, b Card) int { return rank(a) - rank(b) })
	return serving
}
