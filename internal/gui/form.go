package gui

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

// Form is the modal over the board while it is open.
type Form struct {
	Adding       bool
	ForObjective bool
	Number       int

	fields      []engine.Field
	at          int
	caret       int
	held        int
	partAt      int
	partsTaken  []int
	partsTop    int
	spareParts  int
	writingPart bool
	named       Written
	carried     int
	opened      []string
	behind      int
}

// Changed says whether anything in the form differs from what it opened with.
func (f Form) Changed() bool {
	if len(f.opened) != len(f.fields) {
		return false
	}
	for i := range f.fields {
		if f.fields[i].Value != f.opened[i] {
			return true
		}
	}
	return false
}

func asOpened(f Form) Form {
	f.opened = make([]string, len(f.fields))
	for i := range f.fields {
		f.opened[i] = f.fields[i].Value
	}
	return f
}

// Writing is the field the keyboard is on, as a text with its caret.
func (f Form) Writing() Written {
	if len(f.fields) == 0 {
		return Written{}
	}
	return Written{Value: f.fields[f.at].Value, At: f.caret, Held: f.held}
}

// Writes puts a written text back into the field the keyboard is on.
func (f Form) Writes(written Written) Form {
	if len(f.fields) == 0 {
		return f
	}
	f = f.Set(f.fields[f.at].Name, written.Reads())
	f.caret, f.held = written.At, written.Held
	return f
}

func (f Form) onto(at int) Form {
	if f.writingPart {
		f = f.committing()
		f.writingPart = false
	}
	arriving := at != f.at
	if arriving {
		f.partAt = 0
	}
	f.at = min(len(f.fields)-1, max(0, at))
	if arriving && f.onParts() {
		f = f.takingTheOneOn()
	}
	f.caret, f.held = len([]rune(f.fields[f.at].Value)), 0
	return f
}

func adding(tab Tab, state string) Form {
	if tab == Objectives {
		return asOpened(Form{Adding: true, ForObjective: true, fields: engine.NewGoalFields()})
	}
	return asOpened(Form{Adding: true, fields: engine.AddingFields(state)})
}

func editing(c engine.Card) Form {
	return asOpened(Form{Number: c.Number, fields: engine.EditFields(c)})
}

func (w Window) freshened() Window {
	if w.form.Adding || w.form.Number == 0 || len(w.form.opened) != len(w.form.fields) {
		return w
	}
	var fresh []engine.Field
	if w.form.ForObjective {
		g, found := w.snap.Goal(w.form.Number)
		if !found {
			return w
		}
		fresh = engine.GoalFields(g)
	} else {
		c, found := w.snap.Card(w.form.Number)
		if !found {
			return w
		}
		fresh = engine.EditFields(c)
	}
	if len(fresh) != len(w.form.fields) {
		return w
	}
	f := w.form
	f.fields, f.opened = slices.Clone(f.fields), slices.Clone(f.opened)
	for i := range fresh {
		if f.fields[i].Value == f.opened[i] && f.fields[i].Value != fresh[i].Value {
			f.fields[i].Value = fresh[i].Value
			if i == f.at {
				f.caret, f.held = min(f.caret, len([]rune(fresh[i].Value))), 0
			}
		}
		f.opened[i] = fresh[i].Value
	}
	w.form = f
	return w
}

func editingGoal(g engine.Goal) Form {
	return asOpened(Form{Number: g.Number, ForObjective: true, fields: engine.GoalFields(g)})
}

// Fields is what the form asks about, in the order it lays them out.
func (f Form) Fields() []engine.Field { return f.fields }

// Set writes a field.
func (f Form) Set(name, value string) Form {
	fields := make([]engine.Field, len(f.fields))
	copy(fields, f.fields)
	for i := range fields {
		if fields[i].Name == name {
			fields[i].Value = value
			f.at = i
		}
	}
	f.fields = fields
	return f
}

// Types puts what was typed into the field the keyboard is on.
func (w Window) Types(text string) Window {
	if text == "" {
		return w
	}
	return w.writing(func(said Written) Written { return said.Types(text) })
}

// Edits puts what the system sent in place of the run it named, in the field the keyboard is on.
func (w Window) Edits(from, to int, text string) Window {
	return w.writing(func(said Written) Written { return said.Replaces(from, to, text) })
}

// Typed is the text the keys are writing in, which the system is told so that what it composes lands where it thinks: nothing when no text takes them.
func (w Window) Typed() (Written, bool) {
	var said Written
	took := false
	w.writing(func(on Written) Written {
		said, took = on, true
		return on
	})
	return said, took
}

func (w Window) writing(change func(Written) Written) Window {
	if w.screen == OnSearch {
		w.looking = change(w.looking)
		return w
	}
	if w.screen == OnLabels && w.renaming {
		w.renamed = change(w.renamed)
		return w
	}
	if w.screen == OnChoosing {
		if !w.ChoosingSearched() {
			return w
		}
		w.filtering, w.choosing, w.choosingFrom = change(w.filtering), 0, 0
		return w
	}
	if w.screen != OnForm {
		return w
	}
	if w.form.onParts() {
		if w.form.writingPart {
			w.form.named = change(w.form.named)
		}
		return w
	}
	if w.onAClosedSet() {
		return w.choosingList().writing(change)
	}
	w.form = w.form.Writes(change(w.form.Writing()))
	return w.inSightOfTheCaret()
}

func (w Window) onAClosedSet() bool { return len(w.form.Fields()[w.form.On()].Choices) > 0 }

// On is the field the keyboard is on.
func (f Form) On() int { return f.at }

func (w Window) back() Window {
	if w.screen != OnForm || w.onAClosedSet() {
		return w
	}
	w.form = w.form.Writes(w.form.Writing().Back())
	return w
}

func (w Window) walks(by int) Window {
	if w.screen == OnSearch {
		w.looking = w.looking.CaretAt(w.looking.At + by)
		return w
	}
	if w.screen != OnForm {
		return w
	}
	if w.onAClosedSet() {
		return w.cycle(by)
	}
	w.form = w.form.Writes(w.form.Writing().CaretAt(w.form.caret + by))
	return w
}

// CaretInto puts the caret a given number of characters into the field the keyboard is on, which is what a click in the text means.
func (w Window) CaretInto(characters int) Window {
	if w.screen != OnForm {
		return w
	}
	w.form = w.form.Writes(w.form.Writing().CaretAt(characters))
	return w
}

func (w Window) cycle(by int) Window {
	if w.screen != OnForm {
		return w
	}
	on := w.form.Fields()[w.form.On()]
	if len(on.Choices) == 0 {
		return w
	}
	at := 0
	for i, choice := range on.Choices {
		if choice == on.Value {
			at = i
		}
	}
	at = min(len(on.Choices)-1, max(0, at+by))
	w.form = w.form.Writes(Ending(on.Choices[at]))
	return w
}

func (w Window) fieldAt(p Point) (int, bool) {
	if w.screen != OnForm {
		return 0, false
	}
	box := w.formBox()
	fields := w.form.Fields()
	for i := range fields {
		name, value, rows := FormPlace(box, fields, i)
		value.H = lineHigh * rows
		if WordsOf(w.measured(), name, fields[i].Name, small).In(p) || value.In(p) {
			return i, true
		}
	}
	return 0, false
}

func into(value Box, p Point, said string, m Measure) int {
	return m.Nearest(said, reads, p.X-value.X)
}

func (w Window) reaching(at int, p Point) int {
	fields := w.form.Fields()
	_, value, rows := FormPlace(w.formBox(), fields, at)
	inside := Inside(Framed(value, rows))
	on := at == w.form.On()
	said := Written{Value: fields[at].Value}
	if on {
		said = w.form.Writing()
	}
	lines := w.wrapping(said.Reads(), inside, FormSize(fields[at]))
	line, _ := lines.At(said.At)
	return lines.Reaches(inside, p, StartsAt(lines, rows, line, w.scrolledOf(at)), w.measured(), FormSize(fields[at]))
}

func (w Window) scrolledOf(at int) int {
	if !w.form.Fields()[at].Long {
		return Follow
	}
	if moved, _ := w.sayingAt(); moved == at {
		return w.Saying()
	}
	return 0
}

// Edit opens the form on whatever is chosen.
func (w Window) Edit() Window {
	if w.chosen == 0 {
		return w
	}
	if w.chosenGoal {
		if g, found := w.snap.Goal(w.chosen); found {
			w.form, w.screen, w.formMoved = editingGoal(g), OnForm, Point{}
		}
		return w.beside()
	}
	if c, found := w.snap.Card(w.chosen); found {
		w.form, w.screen, w.formMoved = editing(c), OnForm, Point{}
	}
	return w.beside()
}

func (w Window) beside() Window {
	w.history, w.serving, w.saying = 0, 0, 0
	w.spareHistory, w.spareSaying = 0, 0
	w.picked, w.pickingText = Picked{}, false
	return w.readingActs()
}

func (w Window) readingActs() Window {
	w.acts = nil
	if w.chosenGoal || w.screen != OnForm || w.form.Adding {
		return w
	}
	shown, err := w.work.Ask(engine.TraceShow, engine.Numbered(w.chosen))
	if err != nil {
		return w.failing(err)
	}
	for _, act := range shown.Acts {
		w.acts = append(w.acts, line(act))
	}
	return w
}

// Form is what the modal holds, for whoever draws it.
func (w Window) Form() Form { return w.form }

func formShown(s Screen) bool { return s == OnForm || s == OnChoosing || s == OnCalendar }

func (w Window) formUnder() bool {
	return formShown(w.screen) || w.screen == OnAsking && (w.question == aboutDeletingPart || w.question == aboutDeletingParts || w.question == aboutLeaving)
}

// PartsTicked is what the steps field of an open form is drawn as: a tile for each step it says now, done or not, unsaved lines included.
func (w Window) PartsTicked() []engine.Card {
	if !w.formUnder() {
		return nil
	}
	for _, field := range w.form.Fields() {
		if field.Name != "steps" {
			continue
		}
		var cards []engine.Card
		for _, part := range engine.PartsIn(field.Value) {
			state := domain.Todo
			if part.Done {
				state = domain.Done
			}
			cards = append(cards, engine.Card{Number: part.Number, Title: part.Title, State: string(state)})
		}
		return cards
	}
	return nil
}

// Save writes what the form holds and closes it.
func (w Window) Save() Window {
	if w.screen != OnForm {
		return w
	}
	if w.form.writingPart {
		w.form = w.form.committing()
		w.form.writingPart = false
	}
	w, _, err := w.stored()
	if err != nil {
		return w
	}
	w.screen, w.form = OnTheBoard, Form{}
	return w.Look()
}

func (w Window) stored() (Window, int, error) {
	op, asked := w.saving()
	answer, err := w.work.Ask(op, asked)
	if err != nil {
		return w.failing(err), 0, err
	}
	number := w.form.Number
	if w.form.Adding {
		number = engine.Added(answer)
		w.form.Adding, w.form.Number = false, number
		if !w.form.ForObjective {
			if err := w.drafts.Drop(); err != nil {
				return w.failing(err), 0, err
			}
		}
		if err := w.puttingBehind(number); err != nil {
			return w.failing(err), 0, err
		}
	}
	return w, number, nil
}

func (w Window) puttingBehind(number int) error {
	if w.form.behind == 0 {
		return nil
	}
	_, err := w.work.Ask(engine.TaskMove, engine.Input{Given: map[string]string{
		"number": strconv.Itoa(number), "after": strconv.Itoa(w.form.behind)}})
	if errors.Is(err, engine.ErrOtherColumn) {
		return nil
	}
	return err
}

// Cancel closes whatever is open and writes nothing.
func (w Window) Cancel() Window {
	if w.screen == OnTheBoard {
		w.looking = Written{}
		return w
	}
	if w.screen == OnSearch {
		w.screen, w.looking = OnTheBoard, Written{}
		return w
	}
	if w.screen == OnForm && w.form.writingPart {
		w.form.writingPart, w.form.named = false, Written{}
		return w
	}
	if w.screen == OnForm && w.form.Adding && !w.form.ForObjective {
		return w.throwingAway()
	}
	if w.screen == OnForm && w.form.Changed() {
		w.screen, w.question = OnAsking, aboutLeaving
		return w
	}
	if w.screen == OnAsking && w.question == aboutLeaving {
		return w.throwingAway()
	}
	w.screen, w.form, w.question = OnTheBoard, Form{}, aboutDeleting
	return w
}

func (w Window) throwingAway() Window {
	if w.form.Adding && !w.form.ForObjective {
		w.screen, w.form, w.question = OnTheBoard, Form{}, aboutDeleting
		return w.droppingDraft()
	}
	w.draft = w.form
	w.screen, w.form, w.question = OnTheBoard, Form{}, aboutDeleting
	return w
}

func (w Window) recoverable() bool {
	d := w.draft
	return d.fields != nil && !w.form.Changed() &&
		d.Adding == w.form.Adding && d.ForObjective == w.form.ForObjective && d.Number == w.form.Number
}

// Choose is a button pressed by name, which is how a scenario says it and how a pointer says it.
func (w Window) Choose(name string) Window {
	switch strings.ToLower(name) {
	case "quit":
		w.quitting = true
		return w
	case "edit":
		return w.Edit()
	case "new linked task":
		return w.startingLinked()
	case "save":
		return w.Save()
	case "cmd+enter":
		return w.Save()
	case "choose":
		if chosenOnACalendar(w.form.Fields()[w.form.On()]) {
			return w.openingCalendar()
		}
		return w.choosingList()
	case "no day":
		return w.dated(time.Time{})
	case "cancel":
		if w.screen == OnCalendar {
			w.screen = OnForm
			return w
		}
		if w.screen == OnChoosing {
			return w.leaving()
		}
		if w.question == aboutSweeping || w.question == aboutPuttingBack || w.screen == OnBackups {
			w.screen, w.question = OnTheBoard, aboutDeleting
			return w
		}
		if w.question == aboutDeletingLabel {
			w.screen, w.question = OnLabels, aboutDeleting
			return w
		}
		if w.question == aboutDeletingPart || w.question == aboutDeletingParts {
			w.screen, w.question = OnForm, aboutDeleting
			return w
		}
		if w.question == aboutHiding {
			return w.answeringHiding(false)
		}
		if w.screen == OnLabels && w.renaming {
			w.renaming = false
			return w
		}
		if w.screen == OnLabels {
			w.screen = OnTheBoard
			return w
		}
		return w.Cancel()
	case "new":
		return w.startingNew()
	case "pick it up":
		return w.pickingUp()
	case "start afresh":
		return w.startingAfresh()
	case "take":
		if w.screen == OnCalendar {
			return w.dated(w.day)
		}
		if w.screen == OnLabels {
			return w.renamingLabel()
		}
		return w.taking()
	case "keep":
		return w.keeping()
	case "close":
		if w.screen == OnChoosing {
			return w.keeping()
		}
		return w.Cancel()
	case "search":
		w.screen = OnSearch
		return w
	case "filter":
		return w.filteringBoard()
	case "reset sizes":
		return w.resized(Sizes{})
	case "no filter":
		w.filters = w.filters.Cleared()
		return w
	case "help":
		w.screen, w.helpFrom, w.spareHelp = OnHelp, 0, 0
		return w
	case "about cairn":
		w.screen, w.question = OnAsking, aboutCairn
		return w
	case "clear":
		if len(w.taken) > 0 {
			w.taken = nil
			return w
		}
		w.looking = Written{}
		return w
	case "archived":
		w.archived = !w.archived
		return w
	case "tasks":
		return w.onPage(Tasks)
	case "objectives":
		return w.onPage(Objectives)
	case "time":
		return w.onPage(Time)
	case "export":
		return w.exportingTasks()
	case "import":
		return w.importingTasks()
	case "new backup":
		return w.backingUp()
	case "backups":
		return w.listingBackups()
	case "labels":
		return w.listingLabels()
	case "rename":
		return w.startingRename()
	case "all possible":
		return w.actingOnSeveral(true)
	case "skip flagged":
		return w.actingOnSeveral(false)
	case "delete", "agree", "apply":
		if w.screen == OnTheBoard && len(w.Taken()) > 0 {
			return w.askingSeveral(aboutDeletingSeveral)
		}
		if w.question == aboutSweeping {
			return w.sweeping()
		}
		if w.question == aboutPuttingBack {
			return w.puttingBack()
		}
		if w.question == aboutDeletingLabel {
			return w.deletingLabel()
		}
		if w.question == aboutFinishingParts || w.question == aboutApplying {
			return w.agreed()
		}
		if w.question == aboutDeletingPart {
			return w.deletingPart()
		}
		if w.question == aboutDeletingParts {
			return w.deletingParts()
		}
		if w.question == aboutHiding {
			return w.answeringHiding(true)
		}
		if w.screen == OnLabels {
			w.screen, w.question = OnAsking, aboutDeletingLabel
			return w
		}
		if w.question == aboutLeaving {
			return w.throwingAway()
		}
		return w.Delete()
	case "throw away":
		return w.throwingAway()
	case "back":
		w.screen, w.question = OnForm, aboutDeleting
		return w
	case "recover":
		w.form, w.draft = w.draft, Form{}
		return w
	case "archive", "unarchive":
		if w.screen == OnTheBoard && len(w.Taken()) > 0 {
			return w.askingSeveral(aboutArchivingSeveral)
		}
		return w.puttingAway()
	}
	return w
}

func (w Window) saving() (engine.Operation, engine.Input) {
	adding, editing := engine.TaskNew, engine.TaskEdit
	if w.form.ForObjective {
		adding, editing = engine.ObjectiveNew, engine.ObjectiveEdit
	}
	return engine.Saved(w.form.fields, w.form.Adding, w.form.Number, adding, editing)
}

func (w Window) addingOn() string {
	column, inBand, _, found := engine.Where(w.Showing().Cards, w.archived, w.chosen)
	if !found && w.aimed != "" {
		return w.aimed
	}
	if !found {
		return engine.StateAt(0)
	}
	if inBand {
		return engine.Stalled[0]
	}
	return engine.StateAt(column)
}

// Delete asks first, and deletes the second time.
func (w Window) Delete() Window {
	if w.screen == OnAsking && w.question == aboutDeleting {
		return w.deleting()
	}
	if w.chosen == 0 {
		return w
	}
	w.screen = OnAsking
	return w
}

func (w Window) deleting() Window {
	op := engine.TaskDelete
	if w.chosenGoal {
		op = engine.ObjectiveDelete
	}
	w.screen = OnTheBoard
	if _, err := w.work.Ask(op, engine.Agreed(engine.Numbered(w.chosen))); err != nil {
		return w.failing(err)
	}
	w.chosen, w.chosenGoal = 0, false
	return w.Look()
}

// Choosing is what the open list offers: every objective there is, in the order the objectives tab shows them, so that a list and the tab behind it cannot disagree about what comes first.
func (w Window) Choosing() []Choice {
	if w.screen != OnChoosing {
		return nil
	}
	return w.choices()
}

// ChoosingSays is what the list reads when there is nothing to serve.
const ChoosingSays = "nothing to serve yet"

// ChoosingAt is where in that list the keyboard sits.
func (w Window) ChoosingAt() int { return w.choosing }

// Ticked says whether an objective is in, as the open list has it.
func (w Window) Ticked(number int) bool { return w.ticked[number] }

// FieldSays is what a field reads on screen, which is not always what it carries: the engine names the numbers a field holds.
func (w Window) FieldSays(field engine.Field) string { return w.snap.Named(field) }

func (w Window) choosingList() Window {
	on := w.form.Fields()[w.form.On()]
	if !chosenFromAList(on) {
		return w
	}
	w.ticked, w.choosing, w.choosingFrom = map[int]bool{}, 0, 0
	w.choosingFor, w.filtering = on.Arg, Written{}
	if on.Arg == "label" {
		w = w.openingLabels(on.Value)
	} else {
		for _, number := range engine.NumbersIn(on.Value) {
			w.ticked[number] = true
		}
	}
	if len(on.Choices) > 0 {
		w.ticked = map[int]bool{}
		for i, choice := range on.Choices {
			if choice == on.Value {
				w.ticked[i+1] = true
			}
		}
	}
	w.screen = OnChoosing
	for at, one := range w.Choosing() {
		if w.ticked[one.Number] {
			w.choosing = at
			break
		}
	}
	w.choosingFrom = w.bringingIntoSight()
	return w
}

// Choice is one line of the list: what it is numbered and what it is called.
type Choice struct {
	Number  int
	Title   string
	Creates string
}

// RowSays is what a row of the open list reads: whether it is taken, then what it is known by.
func (w Window) RowSays(one Choice) string {
	said := "   "
	if w.Ticked(one.Number) {
		said = TickedMark + "  "
	}
	if len(w.closedSet()) == 0 && w.choosingFor != "label" {
		said += Numbered(engine.Card{Number: one.Number}) + "  "
	}
	return said + one.Title
}

func (w Window) choices() []Choice {
	if w.choosingFor == "label" {
		return w.labelChoices()
	}
	looking := w.filtering.Reads()
	var out []Choice
	if set := w.closedSet(); len(set) > 0 {
		for i, choice := range set {
			if strings.Contains(choice, strings.ToLower(looking)) {
				out = append(out, Choice{Number: i + 1, Title: choice})
			}
		}
		return out
	}
	for _, g := range engine.FoundGoals(w.snap.Goals, looking) {
		out = append(out, Choice{Number: g.Number, Title: g.Title})
	}
	return out
}

// Filtering is what has been typed into the open list.
func (w Window) Filtering() Written { return w.filtering }

// OneOnly says whether the open list takes one answer rather than several.
func (w Window) OneOnly() bool { return len(w.closedSet()) > 0 }

func (w Window) closedSet() []string {
	if w.screen != OnChoosing {
		return nil
	}
	return w.form.Fields()[w.form.On()].Choices
}

func (w Window) inChoosing(name string) Window {
	offered := w.Choosing()
	switch name {
	case "down", "j":
		w.choosing = min(max(0, len(offered)-1), w.choosing+1)
		w.choosingFrom = w.bringingIntoSight()
	case "up", "k":
		w.choosing = max(0, w.choosing-1)
		w.choosingFrom = w.bringingIntoSight()
	case "backspace":
		w.filtering, w.choosing, w.choosingFrom = w.filtering.Back(), 0, 0
	case "esc":
		return w.leaving()
	case "enter":
		return w.taking()
	case "cmd+enter":
		return w.keeping()
	}
	return w
}

func (w Window) taking() Window {
	if w.choosingFor == "label" {
		return w.takingLabel()
	}
	offered := w.Choosing()
	if w.choosing >= len(offered) {
		return w
	}
	number := offered[w.choosing].Number
	if w.OneOnly() {
		if w.ticked[number] && len(w.closedSet()) == 0 {
			w.ticked = map[int]bool{}
		} else {
			w.ticked = map[int]bool{number: true}
		}
		return w.keeping()
	}
	w.ticked[number] = !w.ticked[number]
	return w
}

func (w Window) bringingIntoSight() int {
	room := w.ChoosingRoom()
	return min(w.choosing, max(w.choosingFrom, w.choosing-room+1))
}

func (w Window) leaving() Window {
	w.screen = OnForm
	return w
}

func (w Window) keeping() Window {
	if w.choosingFor == "label" {
		return w.keepingLabels()
	}
	if set := w.closedSet(); len(set) > 0 {
		kept := w.form.Fields()[w.form.On()].Value
		for i, choice := range set {
			if w.ticked[i+1] {
				kept = choice
			}
		}
		w.form = w.form.Writes(Ending(kept))
		w.screen = OnForm
		return w
	}
	var numbers []int
	for _, g := range w.Choosing() {
		if w.ticked[g.Number] {
			numbers = append(numbers, g.Number)
		}
	}
	w.form = w.form.Writes(Ending(engine.Numbers(numbers)))
	w.screen = OnForm
	return w
}

// ChoosingSearched says whether the open list has a line to type into.
func (w Window) ChoosingSearched() bool { return len(w.closedSet()) == 0 }
