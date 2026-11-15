package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/item"
	"cairn/internal/engine/task"
	"cairn/internal/engine/workspace"
	"cairn/internal/gui/geom"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"cairn/internal/engine"
)

func (w Window) adding(tab board.Tab, state string) board.Form {
	if tab == board.Objectives {
		return board.AsOpened(board.Form{Adding: true, ForObjective: true, Fields: board.NewGoalFields()})
	}
	fields := board.AddingFields(state)
	if w.Pages.Recurs() {
		fields = withRepeats(fields)
	}
	return board.AsOpened(board.Form{Adding: true, Fields: fields})
}

func (w Window) editing(c board.Card) board.Form {
	return board.AsOpened(board.Form{Number: c.Number, Fields: w.taskFields(c)})
}

func (w Window) taskFields(c board.Card) []board.Field {
	fields := board.EditFields(c)
	if c.Series == nil && w.Pages.Recurs() {
		fields = withRepeats(fields)
	}
	return fields
}

func withRepeats(fields []board.Field) []board.Field {
	at := len(fields)
	if at > 0 && fields[at-1].Long {
		at--
	}
	return slices.Insert(slices.Clone(fields), at, board.RepeatsField(), board.FromField())
}

func (w Window) freshened() Window {
	if w.Form.Adding || w.Form.ForSeries || w.Form.Number == 0 || len(w.Form.Opened) != len(w.Form.Fields) {
		return w
	}
	var fresh []board.Field
	if w.Form.ForObjective {
		g, found := w.Snap.Goal(w.Form.Number)
		if !found {
			return w
		}
		fresh = board.GoalFields(g)
	} else {
		c, found := w.Snap.Card(w.Form.Number)
		if !found {
			return w
		}
		fresh = w.taskFields(c)
	}
	if len(fresh) != len(w.Form.Fields) {
		return w
	}
	f := w.Form
	f.Fields, f.Opened = slices.Clone(f.Fields), slices.Clone(f.Opened)
	for i := range fresh {
		if f.Fields[i].Value == f.Opened[i] && f.Fields[i].Value != fresh[i].Value {
			f.Fields[i].Value = fresh[i].Value
			if i == f.At {
				f.Caret, f.Held = min(f.Caret, len([]rune(fresh[i].Value))), 0
			}
		}
		f.Opened[i] = fresh[i].Value
	}
	w.Form = f
	return w
}

func editingGoal(g board.Goal) board.Form {
	return board.AsOpened(board.Form{Number: g.Number, ForObjective: true, Fields: board.GoalFields(g)})
}

// Types puts what was typed into the field the keyboard is on.
func (w Window) Types(text string) Window {
	if text == "" {
		return w
	}
	return w.writing(func(said board.Written) board.Written { return said.Types(text) })
}

// Edits puts what the system sent in place of the run it named, in the field the keyboard is on.
func (w Window) Edits(from, to int, text string) Window {
	return w.writing(func(said board.Written) board.Written { return said.Replaces(from, to, text) })
}

// Typed is the text the keys are writing in, which the system is told so that what it composes lands where it thinks: nothing when no text takes them.
func (w Window) Typed() (board.Written, bool) {
	var said board.Written
	took := false
	w.writing(func(on board.Written) board.Written {
		said, took = on, true
		return on
	})
	return said, took
}

func (w Window) writing(change func(board.Written) board.Written) Window {
	if w.Screen == board.OnPalette {
		return w.typedIntoPalette(change(w.palette.typed))
	}
	if w.Screen == board.OnSearch {
		w.Looking = change(w.Looking)
		return w
	}
	if w.Screen == board.OnWorkspaces {
		w.switcher.typed = change(w.switcher.typed)
		if w.switcher.asking == narrowing {
			w.switcher.on = 0
		}
		return w
	}
	if w.Screen == board.OnLabels && w.Renaming {
		w.Renamed = change(w.Renamed)
		return w
	}
	if w.Screen == board.OnConfiguration && w.KeyWriting != "" {
		w.KeyWritten = change(w.KeyWritten)
		return w
	}
	if w.Screen == board.OnKeeping && !w.OfView {
		w.KeepName = change(w.KeepName)
		return w
	}
	if w.Screen == board.OnViews && w.RenamingView {
		w.ViewRenamed = change(w.ViewRenamed)
		return w
	}
	if w.Screen == board.OnLabels {
		return w.seekingLabels(change)
	}
	if w.Screen == board.OnChoosing {
		if !w.ChoosingSearched() {
			return w
		}
		w.Filtering, w.ChoosingOn, w.ChoosingFrom = change(w.Filtering), 0, 0
		return w
	}
	if w.Screen != board.OnForm {
		return w
	}
	if w.Form.OnParts() {
		if w.Form.WritingPart {
			w.Form.Named = change(w.Form.Named)
		}
		return w
	}
	if w.OnAClosedSet() {
		return w.choosingList().writing(change)
	}
	w.Form = w.Form.Writes(change(w.Form.Writing()))
	return w.inSightOfTheCaret()
}

func (w Window) back() Window {
	if w.Screen != board.OnForm || w.OnAClosedSet() {
		return w
	}
	w.Form = w.Form.Writes(w.Form.Writing().Back())
	return w
}

func (w Window) walks(by int) Window {
	if w.Screen == board.OnSearch {
		w.Looking = w.Looking.CaretAt(w.Looking.At + by)
		return w
	}
	if w.Screen != board.OnForm {
		return w
	}
	if w.OnAClosedSet() {
		return w.cycle(by)
	}
	w.Form = w.Form.Writes(w.Form.Writing().CaretAt(w.Form.Caret + by))
	return w
}

// CaretInto puts the caret a given number of characters into the field the keyboard is on, which is what a click in the text means.
func (w Window) CaretInto(characters int) Window {
	if w.Screen != board.OnForm {
		return w
	}
	w.Form = w.Form.Writes(w.Form.Writing().CaretAt(characters))
	return w
}

func (w Window) cycle(by int) Window {
	if w.Screen != board.OnForm {
		return w
	}
	on := w.Form.Fields[w.Form.On()]
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
	w.Form = w.Form.Writes(board.Ending(on.Choices[at]))
	return w
}

func (w Window) fieldAt(p geom.Point) (int, bool) {
	if w.Screen != board.OnForm {
		return 0, false
	}
	box := w.formBox()
	fields := w.Form.Fields
	for i := range fields {
		if w.Hidden(fields[i]) {
			continue
		}
		name, value, rows := FormPlace(box, fields, i)
		value.H = lineHigh * rows
		if WordsOf(w.measured(), name, fields[i].Name, small).In(p) || value.In(p) {
			return i, true
		}
	}
	return 0, false
}

func into(value geom.Box, p geom.Point, said string, m Measure) int {
	return m.Nearest(said, reads, p.X-value.X)
}

func (w Window) reaching(at int, p geom.Point) int {
	fields := w.Form.Fields
	_, value, rows := FormPlace(w.formBox(), fields, at)
	inside := Inside(Framed(value, rows))
	on := at == w.Form.On()
	said := board.Written{Value: fields[at].Value}
	if on {
		said = w.Form.Writing()
	}
	lines := w.wrapping(said.Reads(), inside, FormSize(fields[at]))
	line, _ := lines.At(said.At)
	return lines.Reaches(inside, p, StartsAt(lines, rows, line, w.scrolledOf(at)), w.measured(), FormSize(fields[at]))
}

func (w Window) scrolledOf(at int) int {
	if !w.Form.Fields[at].Long {
		return Follow
	}
	if moved, _ := w.SayingAt(); moved == at {
		return w.Saying()
	}
	return 0
}

// Edit opens the form on whatever is chosen.
func (w Window) Edit() Window {
	if w.Chosen == 0 {
		return w
	}
	if w.ChosenGoal {
		if g, found := w.Snap.Goal(w.Chosen); found {
			w.Form, w.Screen, w.formMoved = editingGoal(g), board.OnForm, geom.Point{}
		}
		return w.beside()
	}
	if c, found := w.Snap.Card(w.Chosen); found {
		w.Form, w.Screen, w.formMoved = w.editing(c), board.OnForm, geom.Point{}
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
	w.Acts = nil
	if w.ChosenGoal || w.Screen != board.OnForm || w.Form.Adding || w.Form.ForSeries {
		return w
	}
	shown, err := w.work.Ask(engine.TraceShow, board.Numbered(w.Chosen))
	if err != nil {
		return w.failing(err)
	}
	for _, act := range shown.Acts {
		w.Acts = append(w.Acts, line(act))
	}
	return w
}

func formShown(s board.Screen) bool {
	return s == board.OnForm || s == board.OnChoosing || s == board.OnCalendar || s == board.OnRule
}

func (w Window) formUnder() bool {
	if w.PickingLabels && w.Screen == board.OnChoosing {
		return false
	}
	return formShown(w.Screen) || w.Screen == board.OnAsking && (w.Question == board.AboutDeletingPart || w.Question == board.AboutDeletingParts || w.Question == board.AboutLeaving)
}

// PartsTicked is what the steps field of an open form is drawn as: a tile for each step it says now, done or not, unsaved lines included.
func (w Window) PartsTicked() []board.Card {
	if !w.formUnder() {
		return nil
	}
	for _, field := range w.Form.Fields {
		if field.Name != "steps" {
			continue
		}
		var cards []board.Card
		for _, part := range engine.PartsIn(field.Value) {
			state := item.Todo
			if part.Done {
				state = item.Done
			}
			cards = append(cards, board.Card{Number: part.Number, Title: part.Title, State: string(state)})
		}
		return cards
	}
	return nil
}

// Save writes what the form holds and closes it.
func (w Window) Save() Window {
	if w.Screen != board.OnForm {
		return w
	}
	if w.Form.WritingPart {
		w.Form = w.Form.Committing()
		w.Form.WritingPart = false
	}
	if w.AsksWhole() {
		w.Screen, w.Question = board.OnAsking, board.AboutSeriesSave
		return w
	}
	w, _, err := w.stored()
	if err != nil {
		return w
	}
	w.Screen, w.Form = board.OnTheBoard, board.Form{}
	return w.Look()
}

func (w Window) stored() (Window, int, error) {
	op, asked := w.Saving()
	answer, err := w.work.Ask(op, asked)
	if err != nil {
		return w.failing(err), 0, err
	}
	w = w.warned(answer)
	number := w.Form.Number
	if w.Form.Adding {
		number = board.Added(answer)
		w.Form.Adding, w.Form.Number = false, number
		if !w.Form.ForObjective {
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
	if w.Form.Behind == 0 {
		return nil
	}
	_, err := w.work.Ask(engine.TaskMove, engine.Input{Given: map[string]string{
		"number": strconv.Itoa(number), "after": strconv.Itoa(w.Form.Behind), "state": w.Form.Argued("state")}})
	if errors.Is(err, task.ErrOtherColumn) {
		return nil
	}
	return err
}

// Cancel closes whatever is open and writes nothing.
func (w Window) Cancel() Window {
	if w.Screen == board.OnTheBoard {
		w.Looking = board.Written{}
		return w
	}
	if w.Screen == board.OnSearch {
		return w.leavingSearch()
	}
	if w.Screen == board.OnForm && w.Form.WritingPart {
		w.Form.WritingPart, w.Form.Named = false, board.Written{}
		return w
	}
	if w.Screen == board.OnForm && w.Form.Adding && !w.Form.ForObjective {
		return w.throwingAway()
	}
	if w.Screen == board.OnForm && w.Form.Changed() {
		w.Screen, w.Question = board.OnAsking, board.AboutLeaving
		return w
	}
	if w.Screen == board.OnAsking && w.Question == board.AboutLeaving {
		return w.throwingAway()
	}
	w.Screen, w.Form, w.Question = board.OnTheBoard, board.Form{}, board.AboutDeleting
	return w
}

func (w Window) throwingAway() Window {
	if w.Screen == board.OnAsking && w.Form.Adding && !w.Form.ForObjective {
		return w.leavingUnsaved()
	}
	if w.Form.Adding && !w.Form.ForObjective {
		w.Screen, w.Form, w.Question = board.OnTheBoard, board.Form{}, board.AboutDeleting
		return w.droppingDraft()
	}
	w.Draft = w.Form
	w.Screen, w.Form, w.Question = board.OnTheBoard, board.Form{}, board.AboutDeleting
	return w
}

// Choose is a button pressed by name, which is how a scenario says it and how a pointer says it.
func (w Window) Choose(name string) Window {
	if w.Screen == board.OnKeeping || w.Screen == board.OnViews {
		if done, took := w.choosingOnViews(name); took {
			return done
		}
	}
	if w.Screen == board.OnFates {
		if strings.HasPrefix(name, "Delete") {
			return w.confirmingFates()
		}
		w.Screen = board.OnTheBoard
		return w
	}
	if answered, took := w.choosingOnSettings(name); took {
		return answered
	}
	if answered, took := w.choosingOnSeries(name); took {
		return answered
	}
	switch strings.ToLower(name) {
	case "quit":
		w.quitting = true
		return w
	case "edit":
		return w.Edit()
	case "show its tasks":
		return w.detailing(ShowItsTasks)
	case strings.ToLower(AddSubSays):
		return w.detailing(AddSubSays)
	case "new linked task":
		return w.startingLinked()
	case "save":
		if w.Screen == board.OnLabels {
			return w.renamingLabel()
		}
		return w.Save()
	case "cmd+enter":
		return w.Save()
	case "choose":
		if w.Screen == board.OnRule {
			return w.closingRule(true)
		}
		if w.Screen == board.OnChoosing {
			return w.taking()
		}
		if w.Screen == board.OnFiltering {
			return w.choosingValue()
		}
		if w.Screen == board.OnLabels && w.Colouring {
			return w.colouringLabel(w.ColourPick)
		}
		if chosenOnACalendar(w.Form.Fields[w.Form.On()]) {
			return w.openingCalendar()
		}
		if chosenFromRules(w.Form.Fields[w.Form.On()]) {
			return w.openingRule()
		}
		return w.choosingList()
	case "no day":
		return w.dated(time.Time{})
	case "cancel":
		if w.Screen == board.OnFiltering && w.FilterOpen {
			return w.closingMenu()
		}
		if w.Screen == board.OnLabels && w.Colouring {
			w.Colouring = false
			return w
		}
		if w.Screen == board.OnCalendar && w.CalendarView != board.ViewDays {
			w.CalendarView = board.ViewDays
			return w
		}
		if w.Screen == board.OnCalendar {
			return w.leavingCalendar()
		}
		if w.Screen == board.OnRule {
			return w.closingRule(false)
		}
		if w.Screen == board.OnChoosing {
			return w.leaving()
		}
		if w.Question == board.AboutSweeping || w.Question == board.AboutPuttingBack || w.Screen == board.OnBackups {
			w.Screen, w.Question = board.OnTheBoard, board.AboutDeleting
			return w
		}
		if w.Question == board.AboutDeletingLabel {
			w.Screen, w.Question = board.OnLabels, board.AboutDeleting
			return w
		}
		if w.Question == board.AboutDeletingPart || w.Question == board.AboutDeletingParts {
			w.Screen, w.Question = board.OnForm, board.AboutDeleting
			return w
		}
		if w.Question == board.AboutHiding {
			return w.answeringHiding(false)
		}
		if w.Screen == board.OnLabels && w.Renaming {
			w.Renaming, w.Making = false, false
			return w
		}
		if w.Screen == board.OnLabels && w.Seeking {
			return w.leavingTheSearch()
		}
		if w.Screen == board.OnLabels {
			w.Screen = board.OnTheBoard
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
		if w.Screen == board.OnCalendar && w.CalendarView != board.ViewDays {
			return w.backFromGrid()
		}
		if w.Screen == board.OnCalendar {
			return w.dated(w.Day)
		}
		return w.taking()
	case "keep", "done":
		return w.keeping()
	case "close":
		if w.Screen == board.OnChoosing {
			return w.keeping()
		}
		return w.Cancel()
	case "search":
		w.Screen, w.SearchFrom = board.OnSearch, board.OnTheBoard
		return w
	case "filter":
		return w.filteringBoard()
	case strings.ToLower(KeepAsViewSays):
		return w.startingKeep()
	case strings.ToLower(ViewsSays) + "…":
		return w.listingViews()
	case "reset sizes":
		return w.resized(geom.Sizes{})
	case "no filter":
		w.Filters = w.Filters.Cleared()
		return w
	case "take off":
		return w.takingOff(w.FilterOn)
	case "help":
		w.Screen = board.OnHelp
		return w
	case "about cairn":
		w.Screen, w.Question = board.OnAsking, board.AboutCairn
		return w
	case "save search":
		w.Screen = board.OnTheBoard
		return w
	case "clear":
		if len(w.Gathered) > 0 {
			w.Gathered = nil
			return w
		}
		w.Looking = board.Written{}
		return w
	case "archived":
		w.Archived = !w.Archived
		return w
	case "tasks":
		return w.onPage(board.Tasks)
	case "objectives":
		return w.onPage(board.Objectives)
	case "time":
		return w.onPage(board.Time)
	case "series":
		return w.onPage(board.Series)
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
	case "settings…":
		return w.openingConfiguration()
	case "open a workspace…":
		return w.openingSwitcher()
	case "rename":
		return w.startingRename()
	case "new label":
		return w.makingLabel()
	case "create":
		return w.creatingSought()
	case "colour":
		return w.startingColour()
	case "skip flagged":
		return w.actingOnSeveral(false)
	case "delete", "agree", "apply", "restore", "finish", "replace":
		if w.Screen == board.OnBackups {
			w.Screen, w.Question = board.OnAsking, board.AboutPuttingBack
			return w
		}
		if w.Screen == board.OnAsking && w.AsksAboutSeveral() {
			return w.actingOnSeveral(true)
		}
		if w.Screen == board.OnTheBoard && len(w.Taken()) > 0 {
			return w.askingSeveral(board.AboutDeletingSeveral)
		}
		if w.Question == board.AboutSweeping {
			return w.sweeping()
		}
		if w.Question == board.AboutPuttingBack {
			return w.puttingBack()
		}
		if w.Question == board.AboutDeletingLabel {
			return w.deletingLabel()
		}
		if w.Question == board.AboutFinishingParts || w.Question == board.AboutApplying {
			return w.agreed()
		}
		if w.Question == board.AboutDeletingPart {
			return w.deletingPart()
		}
		if w.Question == board.AboutDeletingParts {
			return w.deletingParts()
		}
		if w.Question == board.AboutHiding {
			return w.answeringHiding(true)
		}
		if w.Screen == board.OnLabels {
			w.Screen, w.Question = board.OnAsking, board.AboutDeletingLabel
			return w
		}
		if w.Question == board.AboutLeaving {
			return w.throwingAway()
		}
		return w.Delete()
	case "throw away":
		if w.Question == board.AboutLeavingSettings {
			w.Screen, w.Question, w.KeyWriting = board.OnConfiguration, board.AboutDeleting, ""
			return w.Choose("Close")
		}
		return w.throwingAway()
	case "back", "keep editing":
		if w.Question == board.AboutLeavingSettings {
			w.Screen, w.Question = board.OnConfiguration, board.AboutDeleting
			return w
		}
		w.Screen, w.Question = board.OnForm, board.AboutDeleting
		return w
	case "recover":
		w.Form, w.Draft = w.Draft, board.Form{}
		return w
	case "archive", "unarchive":
		if w.Screen == board.OnAsking && w.Question == board.AboutSweeping {
			return w.sweeping()
		}
		if w.Screen == board.OnAsking && w.AsksAboutSeveral() {
			return w.actingOnSeveral(true)
		}
		if w.Screen == board.OnTheBoard && len(w.Taken()) > 0 {
			return w.askingSeveral(board.AboutArchivingSeveral)
		}
		return w.puttingAway()
	}
	return w
}

func (w Window) addingOn() string {
	column, inBand, _, found := board.Where(w.Showing().Cards, w.Archived, w.Chosen)
	if !found && w.aimed != "" {
		return w.aimed
	}
	if !found {
		return board.StateAt(0)
	}
	if inBand {
		return board.Stalled[0]
	}
	return board.StateAt(column)
}

// Delete asks first, and deletes the second time.
func (w Window) Delete() Window {
	if w.Screen == board.OnAsking && w.Question == board.AboutDeleting {
		return w.deleting()
	}
	if w.Chosen == 0 {
		return w
	}
	if w.ChosenGoal && len(w.Gathered) == 0 {
		return w.deletingGoal(w.Chosen)
	}
	w.Screen = board.OnAsking
	return w
}

func (w Window) deleting() Window {
	op := engine.TaskDelete
	if w.ChosenGoal {
		op = engine.ObjectiveDelete
	}
	w.Screen = board.OnTheBoard
	if _, err := w.work.Ask(op, engine.Agreed(board.Numbered(w.Chosen))); err != nil {
		return w.failing(err)
	}
	w.Chosen, w.ChosenGoal = 0, false
	return w.Look()
}

// Choosing is what the open list offers: every objective there is, in the order the objectives tab shows them, so that a list and the tab behind it cannot disagree about what comes first.
func (w Window) Choosing() []Choice {
	if w.Screen != board.OnChoosing {
		return nil
	}
	return w.choices()
}

// ChoosingSays is what the list reads when there is nothing to serve.
const ChoosingSays = "nothing to serve yet"

func (w Window) choosingList() Window {
	on := w.Form.Fields[w.Form.On()]
	if !chosenFromAList(on) {
		return w
	}
	w.TickedSet, w.ChoosingOn, w.ChoosingFrom, w.RefusedChoice = map[int]bool{}, 0, 0, ""
	w.ChoosingFor, w.Filtering, w.PickingLabels = on.Arg, board.Written{}, false
	if on.Arg == "label" {
		w = w.openingLabels(on.Value)
	} else {
		for _, number := range board.NumbersIn(on.Value) {
			w.TickedSet[number] = true
		}
		if on.Arg == "parent" && len(w.TickedSet) == 0 {
			w.TickedSet[0] = true
		}
	}
	if set := w.SetOf(on); len(set) > 0 {
		w.TickedSet = map[int]bool{}
		for i, choice := range set {
			if choice == on.Value {
				w.TickedSet[i+1] = true
			}
		}
	}
	w.Screen = board.OnChoosing
	for at, one := range w.Choosing() {
		if w.TickedSet[one.Number] {
			w.ChoosingOn = at
			break
		}
	}
	w.ChoosingFrom = w.bringingIntoSight()
	return w
}

// Choice is one line of the list: what it is numbered and what it is called, and for a label how many tasks carry it.
type Choice struct {
	Number  int
	Title   string
	Creates string
	Carried int
	Depth   int
	Last    bool
	Through int
}

// RowSays is what a row of the open list reads after its box: what it is known by.
func (w Window) RowSays(one Choice) string {
	said := ""
	if w.ChoosingFor == "parent" && one.Number == 0 {
		return one.Title
	}
	if len(w.ClosedSet()) == 0 && w.ChoosingFor != "label" {
		said += "#" + Numbered(board.Card{Number: one.Number}) + "  "
	}
	if w.ChoosingFor == "priority" {
		said += priorityRoom
	}
	return said + one.Title
}

func (w Window) choices() []Choice {
	if w.ChoosingFor == "label" {
		return w.labelChoices()
	}
	looking := w.Filtering.Reads()
	var out []Choice
	if set := w.ClosedSet(); len(set) > 0 {
		for i, choice := range set {
			if strings.Contains(choice, strings.ToLower(looking)) {
				out = append(out, Choice{Number: i + 1, Title: choice})
			}
		}
		return out
	}
	if w.ChoosingFor == "blocker" || w.ChoosingFor == "link" {
		for _, c := range board.FoundCards(w.Snap.Cards, looking, w.Form.Number) {
			out = append(out, Choice{Number: c.Number, Title: c.Title})
		}
		return out
	}
	if w.ChoosingFor == "parent" {
		out = append(out, Choice{Title: "none"})
		for _, g := range w.Snap.Parents(w.Form.Number) {
			out = append(out, Choice{Number: g.Number, Title: g.Title})
		}
		return out
	}
	if looking == "" {
		var ticked []int
		for n, on := range w.TickedSet {
			if on {
				ticked = append(ticked, n)
			}
		}
		for _, p := range w.Snap.Picks(ticked) {
			out = append(out, Choice{Number: p.Goal.Number, Title: p.Goal.Title, Depth: depthOf(p.Branch), Last: p.Last, Through: p.Through})
		}
		return out
	}
	for _, g := range board.FoundGoals(w.Snap.Goals, looking) {
		out = append(out, Choice{Number: g.Number, Title: g.Title})
	}
	return out
}

// RowAside is what a row of the open list says at its right: of an objective counted already through what is ticked below it, so.
func (w Window) RowAside(one Choice) string {
	if w.ChoosingFor == "objective" && one.Through > 0 && !w.TickedSet[one.Number] {
		return "counted through " + strconv.Itoa(one.Through) + " below"
	}
	return ""
}

// priorityRoom is the room a priority's row leaves before its name for the stripe of its colour.
const priorityRoom = "    "

// RowStripe is where a priority's row draws the stripe of its colour, in the room left before its name.
func (w Window) RowStripe(one Choice, line geom.Box) (geom.Box, bool) {
	if w.ChoosingFor != "priority" {
		return geom.Box{}, false
	}
	said := w.RowSays(one)
	before := strings.TrimSuffix(said, priorityRoom+one.Title)
	return geom.Box{X: line.X + tickRoom + ends(w.measured(), before, reads) + inset, Y: line.Y + (line.H-stripeHigh)/2, W: edgeWide, H: stripeHigh}, true
}

const stripeHigh = 12

func (w Window) inChoosing(name string) Window {
	offered := w.Choosing()
	switch name {
	case "down", "j":
		w.ChoosingOn = min(max(0, len(offered)-1), w.ChoosingOn+1)
		w.ChoosingFrom = w.bringingIntoSight()
	case "up", "k":
		w.ChoosingOn = max(0, w.ChoosingOn-1)
		w.ChoosingFrom = w.bringingIntoSight()
	case "backspace":
		if w.ChoosingFor == "label" && w.Filtering.Empty() {
			return w.untakingLast()
		}
		w.Filtering, w.ChoosingOn, w.ChoosingFrom = w.Filtering.Back(), 0, 0
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
	if w.ChoosingFor == "label" {
		return w.takingLabel()
	}
	offered := w.Choosing()
	if w.ChoosingOn >= len(offered) {
		return w
	}
	number := offered[w.ChoosingOn].Number
	if w.OneOnly() {
		if w.TickedSet[number] && len(w.ClosedSet()) == 0 {
			w.TickedSet = map[int]bool{}
		} else {
			w.TickedSet = map[int]bool{number: true}
		}
		return w.keeping()
	}
	if !w.TickedSet[number] && w.ChoosingFor == "objective" {
		numbers := []int{number}
		for n, on := range w.TickedSet {
			if on {
				numbers = append(numbers, n)
			}
		}
		if err := w.Snap.CountedOnce(w.Form.Number, numbers); err != nil {
			w.RefusedChoice, _, _ = strings.Cut(err.Error(), "; ")
			w.RefusedThen = w.untickSays(number)
			return w
		}
	}
	w.TickedSet[number], w.RefusedChoice = !w.TickedSet[number], ""
	return w
}

func (w Window) bringingIntoSight() int {
	room := w.ChoosingRoom()
	return min(w.ChoosingOn, max(w.ChoosingFrom, w.ChoosingOn-room+1))
}

func (w Window) leaving() Window {
	if w.PickingLabels {
		return w.leavingThePicker()
	}
	w.Screen = board.OnForm
	return w
}

func (w Window) keeping() Window {
	if w.PickingLabels {
		return w.leavingThePicker()
	}
	if w.ChoosingFor == "label" {
		return w.keepingLabels()
	}
	if set := w.ClosedSet(); len(set) > 0 {
		kept := w.Form.Fields[w.Form.On()].Value
		for i, choice := range set {
			if w.TickedSet[i+1] {
				kept = choice
			}
		}
		w.Form = w.Form.Writes(board.Ending(kept))
		w.Screen = board.OnForm
		return w
	}
	var numbers []int
	if w.ChoosingSearched() {
		for n, on := range w.TickedSet {
			if on {
				numbers = append(numbers, n)
			}
		}
		slices.Sort(numbers)
	}
	for _, g := range w.Choosing() {
		if w.TickedSet[g.Number] && g.Number != 0 && !w.ChoosingSearched() {
			numbers = append(numbers, g.Number)
		}
	}
	if w.ChoosingFor == "parent" && len(numbers) == 0 {
		w.Form = w.Form.Writes(board.Ending("none"))
		w.Screen = board.OnForm
		return w
	}
	w.Form = w.Form.Writes(board.Ending(board.Numbers(numbers)))
	w.Screen = board.OnForm
	return w
}

// NothingChanged is how a refusal at the foot of a list ends.
const NothingChanged = "Nothing was changed."

func (w Window) untickSays(number int) string {
	g, _ := w.Snap.Goal(number)
	if g.Parent != 0 {
		parent, _ := w.Snap.Goal(g.Parent)
		return "Untick " + parent.Title + " to put it on " + g.Title + ". " + NothingChanged
	}
	var below []string
	for _, c := range g.Children {
		if w.TickedSet[c] {
			sub, _ := w.Snap.Goal(c)
			below = append(below, sub.Title)
		}
	}
	return "Untick " + workspace.InWords(below) + " to put it on " + g.Title + " itself. " + NothingChanged
}

// RefusalBox is where the open list says what it refused: under it, or above it when there is no room below.
func (w Window) RefusalBox() (string, geom.Box, bool) {
	refused := w.ChoosingRefused()
	if refused == "" || w.Screen != board.OnChoosing {
		return "", geom.Box{}, false
	}
	box := w.ChoosingBox()
	high := BannerHigh(w.measured(), refused, w.ChoosingThen(), box.W)
	at := geom.Box{X: box.X, Y: box.Y + box.H + between, W: box.W, H: high}
	if at.Y+at.H > w.height-geom.Margin {
		at.Y = box.Y - between - high
	}
	return refused, at, true
}
