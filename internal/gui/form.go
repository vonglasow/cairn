package gui

// The modal that adds and edits, and the two buttons that end it.
//
// Nothing here lists the fields. They come from reading, which says the word a
// person reads, the closed set when there is one, the field that wants more
// than one line, and which engine argument each fills — so an attribute added
// to the engine and to the terminal's form arrives here with no line written.

import (
	"strconv"
	"strings"
	"time"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/reading"
)

// Form is the modal over the board while it is open.
type Form struct {
	// Adding tells a form over nothing from a form over something. Adding and
	// editing are one form, because they ask the same questions.
	Adding bool
	// ForObjective says which of the two things is being written.
	ForObjective bool
	// Number is the task or objective being edited, and nothing when adding.
	Number int

	fields []reading.Field
	// at is the field the keyboard is on, and caret where it stands inside
	// that field's value. Moving to another field puts the caret at its end,
	// because somebody arriving at a field is adding to it.
	at    int
	caret int
	// held is how far a run's anchor lies from the caret, nought for none.
	held int
	// partAt is which line of the list of parts the keyboard is on: a part,
	// one struck out, or the line that adds. writingPart says a line is being
	// written on — a part renamed, or a new one — and named is what is
	// written there. partsTop is the first line of the list in sight, moved
	// only as far as keeps the keyboard's line in sight. struck is the parts
	// struck out, by number, which stay in
	// their place until the form is saved; carried is the part being dragged,
	// counted from one.
	partAt      int
	partsTop    int
	spareParts  int
	writingPart bool
	named       Written
	struck      []int
	carried     int
	// opened is what the fields held when the form was opened, so leaving it
	// can tell a form somebody typed into from one they only looked at. Ten
	// minutes of typing used to go on an esc without a question, while a
	// deletion of two words was guarded.
	opened []string
}

// Changed says whether anything in the form differs from what it opened with.
func (f Form) Changed() bool {
	if len(f.opened) != len(f.fields) {
		return false
	}
	if len(f.struck) > 0 {
		return true
	}
	for i := range f.fields {
		if f.fields[i].Value != f.opened[i] {
			return true
		}
	}
	return false
}

// asOpened records what the fields held, which is what Changed compares with.
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

// onto moves the keyboard to a field and puts the caret at its end.
func (f Form) onto(at int) Form {
	// What was being written on the list of parts is kept on the way out.
	if f.writingPart {
		f = f.committing()
		f.writingPart = false
	}
	if at != f.at {
		f.partAt = 0
	}
	f.at = min(len(f.fields)-1, max(0, at))
	f.caret, f.held = len([]rune(f.fields[f.at].Value)), 0
	return f
}

// adding opens an empty form for whichever tab asked, a task's on the column
// it was asked on.
func adding(tab Tab, state string) Form {
	if tab == Objectives {
		return asOpened(Form{Adding: true, ForObjective: true, fields: reading.NewGoalFields()})
	}
	return asOpened(Form{Adding: true, fields: reading.AddingFields(state)})
}

// editing opens a form on what a task already carries, its parts included.
func editing(c reading.Card, cards []reading.Card) Form {
	return asOpened(Form{Number: c.Number, fields: reading.EditFields(c, cards)})
}

// editingGoal is the same for an objective.
func editingGoal(g reading.Goal) Form {
	return asOpened(Form{Number: g.Number, ForObjective: true, fields: reading.GoalFields(g)})
}

// Fields is what the form asks about, in the order it lays them out.
func (f Form) Fields() []reading.Field { return f.fields }

// On is the field the keyboard is on.
func (f Form) On() int { return f.at }

// ValueOf answers what a field holds, and nothing for a field that is not
// there.
func (f Form) ValueOf(name string) string {
	for _, field := range f.fields {
		if field.Name == name {
			return field.Value
		}
	}
	return ""
}

// Set writes a field. A form is a value like everything else here, so this
// answers the form it becomes.
func (f Form) Set(name, value string) Form {
	fields := make([]reading.Field, len(f.fields))
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

// Types puts what was typed into the field the keyboard is on. A form nobody
// opened has no field to type into, and typing on the board goes nowhere.
func (w Window) Types(text string) Window {
	if text == "" {
		return w
	}
	if w.screen == OnSearch {
		w.looking = w.looking.Types(text)
		return w
	}
	// A list narrows as it is typed into. It is the same gesture as a search
	// over the board, and the same matching rule answers it.
	if w.screen == OnChoosing {
		w.filtering, w.choosing, w.choosingFrom = w.filtering.Types(text), 0, 0
		return w
	}
	if w.screen != OnForm {
		return w
	}
	// On the list of parts, what is typed is a part's name — and nothing at
	// all while no line is being written: space ticks there.
	if w.form.onParts() {
		if w.form.writingPart {
			w.form.named = w.form.named.Types(text)
		}
		return w
	}
	w.form = w.form.Writes(w.form.Writing().Types(text))
	return w.inSightOfTheCaret()
}

// On puts the keyboard on a field by name, which is what a click does and what
// a test says.
func (w Window) On(name string) Window {
	for i, field := range w.form.Fields() {
		if field.Name == name {
			w.form = w.form.onto(i)
		}
	}
	return w
}

// back takes the last character off the field the keyboard is on. One
// character and not one byte: half of a letter is not a letter.
func (w Window) back() Window {
	if w.screen != OnForm {
		return w
	}
	w.form = w.form.Writes(w.form.Writing().Back())
	return w
}

// walks the caret through the field the keyboard is on. A field with a closed
// set has nothing to walk through: the same keys change its value instead.
func (w Window) walks(by int) Window {
	if w.screen == OnSearch {
		w.looking = w.looking.CaretAt(w.looking.At + by)
		return w
	}
	if w.screen != OnForm {
		return w
	}
	if len(w.form.Fields()[w.form.On()].Choices) > 0 {
		return w.cycle(by)
	}
	w.form = w.form.Writes(w.form.Writing().CaretAt(w.form.caret + by))
	return w
}

// CaretInto puts the caret a given number of characters into the field the
// keyboard is on, which is what a click in the text means.
func (w Window) CaretInto(characters int) Window {
	if w.screen != OnForm {
		return w
	}
	w.form = w.form.Writes(w.form.Writing().CaretAt(characters))
	return w
}

// cycle moves a field with a closed set to the next or previous of them, and
// stops at both ends. Seven columns and four priorities are a value to cycle,
// not a page to read — and a field with nothing to cycle is left alone, because
// the same key has to be free to do something else there one day.
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

// fieldAt says which of a form's fields a point lands on.
func (w Window) fieldAt(p Point) (int, bool) {
	if w.screen != OnForm {
		return 0, false
	}
	box := w.formBox()
	fields := w.form.Fields()
	for i := range fields {
		name, value, rows := FormPlace(box, fields, i)
		// A field that wants several rows is answered anywhere in them.
		value.H = lineHigh * rows
		if name.In(p) || value.In(p) {
			return i, true
		}
	}
	return 0, false
}

// into says how many characters into a line of text a point falls, so that
// clicking in the middle of a word puts the caret there rather than at the end.
func into(value Box, p Point, said string, m Measure) int {
	return m.Nearest(said, reads, p.X-value.X)
}

// reaching says which character of a field a point lands on, line and column
// both, in the field as it is drawn: from the line in sight at its top, and
// with the face its words are written in. Asked before the keyboard moves onto
// the field, because what was pressed is what was on screen.
func (w Window) reaching(at int, p Point) int {
	fields := w.form.Fields()
	_, value, rows := FormPlace(w.formBox(), fields, at)
	inside := Inside(Framed(value, rows))
	// What drawForm shows of it: the field the keyboard is on from what is
	// being typed and where the wheel left it, the others from their beginning.
	on := at == w.form.On()
	said := Written{Value: fields[at].Value}
	if on {
		said = w.form.Writing()
	}
	lines := w.wrapping(said.Reads(), inside, FormSize(fields[at]))
	line, _ := lines.At(said.At)
	return lines.Reaches(inside, p, StartsAt(lines, rows, line, w.scrolledOf(at)), w.measured(), FormSize(fields[at]))
}

// scrolledOf is where a field of the form shows its text from: a one-line
// field wherever its caret is, the long field the wheel moves wherever the
// wheel left it, and any other long field from its beginning, which is where
// a text is read from.
func (w Window) scrolledOf(at int) int {
	if !w.form.Fields()[at].Long {
		return Follow
	}
	if moved, _ := w.sayingAt(); moved == at {
		return w.Saying()
	}
	return 0
}

// Walk moves the keyboard from one field to the next, and stops at both ends
// rather than wrapping: a form that wraps is a form somebody falls off.
func (f Form) Walk(by int) Form {
	f.at = min(len(f.fields)-1, max(0, f.at+by))
	return f
}

// Edit opens the form on whatever is chosen. Nothing chosen opens nothing:
// there is no task to be a form about.
func (w Window) Edit() Window {
	if w.chosen == 0 {
		return w
	}
	if w.chosenGoal {
		if g, found := w.snap.Goal(w.chosen); found {
			w.form, w.screen = editingGoal(g), OnForm
		}
		return w.beside()
	}
	if c, found := w.snap.Card(w.chosen); found {
		w.form, w.screen = editing(c, w.snap.Cards), OnForm
	}
	return w.beside()
}

// beside reads what the column beside an opened form shows, and puts it at
// its top: the history of a task, or the work serving an objective, which is
// already on the board. An objective has no history of its own — the trace is
// kept of tasks — and the number it carries belongs to a task as well: asking
// for the trace of that number answered with a task nobody had chosen.
func (w Window) beside() Window {
	w.history, w.serving, w.saying = 0, 0, 0
	w.spareHistory, w.spareSaying = 0, 0
	w.picked, w.pickingText = Picked{}, false
	w.acts = nil
	if w.chosenGoal || w.screen != OnForm {
		return w
	}
	shown, err := w.work.Ask(engine.TraceShow, reading.Numbered(w.chosen))
	if err != nil {
		w.err, w.wrongFor = err, 0
		return w
	}
	for _, act := range shown.Acts {
		w.acts = append(w.acts, line(act))
	}
	return w
}

// Form is what the modal holds, for whoever draws it.
func (w Window) Form() Form { return w.form }

// PartsTicked is what the parts field of an open form is drawn as: a tile for
// each part it says now, done or not, unsaved lines included. Nothing on a
// form with no parts field.
//
// The form is open under a list being chosen from as much as on its own, and
// is drawn there too. Answered on the form alone, a task with parts opened a
// list and the window fell over: its tiles counted the parts, and this said
// there were none.
func (w Window) PartsTicked() []reading.Card {
	if w.screen != OnForm && w.screen != OnChoosing {
		return nil
	}
	for _, field := range w.form.Fields() {
		if field.Name != "parts" {
			continue
		}
		var cards []reading.Card
		for _, part := range reading.PartsIn(field.Value) {
			state := domain.Todo
			if part.Done {
				state = domain.Done
			}
			cards = append(cards, reading.Card{Number: part.Number, Title: part.Title, State: string(state)})
		}
		return cards
	}
	return nil
}

// Save writes what the form holds and closes it. One act in the history per
// save, which is the whole reason this is a modal and not twelve fields edited
// in place on a card.
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

// stored writes what the form holds, and says the number of what it wrote:
// the task or objective it was over, or the one it has just added. What went
// wrong is said on the window and handed back.
func (w Window) stored() (Window, int, error) {
	w.form = w.form.withoutStruck()
	op, asked := w.saving()
	answer, err := w.work.Ask(op, asked)
	if err != nil {
		w.err, w.wrongFor = err, 0
		return w, 0, err
	}
	// The parts go under the task the form was over — or, when it added one,
	// under the task it has just added.
	number := w.form.Number
	if w.form.Adding {
		number = reading.Added(answer)
	}
	if err := w.savingParts(number); err != nil {
		w.err, w.wrongFor = err, 0
		return w, 0, err
	}
	return w, number, nil
}

// savingParts settles what hangs under a task with what the form's parts
// field says — see reading.SettleParts.
func (w Window) savingParts(parent int) error {
	for _, field := range w.form.fields {
		if field.Name == "parts" {
			return reading.SettleParts(w.work, parent, field.Value, w.snap.Cards)
		}
	}
	return nil
}

// Cancel closes whatever is open and writes nothing. Whatever, and not only a
// form: the question asked before deleting is answered with the same word, and
// a Cancel that only knew about forms left it open.
func (w Window) Cancel() Window {
	// On the board, esc is what clears a search — there is nothing else here
	// to put away.
	if w.screen == OnTheBoard {
		w.looking = Written{}
		return w
	}
	// Backing out of the search box itself clears what it was narrowing to.
	// This used to be a branch on esc alone; the bar now offers the same act,
	// and an offer that did less than the key behind it would be a bar lying
	// about what it does.
	if w.screen == OnSearch {
		w.screen, w.looking = OnTheBoard, Written{}
		return w
	}
	// A line of the list of parts being written is put down, and nothing
	// else: it is the nearer thing to back out of.
	if w.screen == OnForm && w.form.writingPart {
		w.form.writingPart, w.form.named = false, Written{}
		return w
	}
	// A form somebody typed into is asked about before it goes. The window
	// guards a deletion of two words; ten minutes of typing went on an esc
	// with no question at all.
	if w.screen == OnForm && w.form.Changed() {
		w.screen, w.question = OnAsking, aboutLeaving
		return w
	}
	// And the question answered the way it was reached — esc, or a click off
	// it — throws the form away, setting what was typed aside.
	if w.screen == OnAsking && w.question == aboutLeaving {
		return w.throwingAway()
	}
	// Anywhere else it closes what is open and nothing more. A search
	// narrowing the board is not what was opened: a task found by searching,
	// opened, changed and put away used to give back the whole board, leaving
	// somebody to type their search again to carry on where they were.
	w.screen, w.form, w.question = OnTheBoard, Form{}, aboutDeleting
	return w
}

// throwingAway closes a form somebody typed into, setting what they typed
// aside rather than losing it.
func (w Window) throwingAway() Window {
	w.draft = w.form
	w.screen, w.form, w.question = OnTheBoard, Form{}, aboutDeleting
	return w
}

// recoverable says whether the form open is the one a form was thrown away
// from — the same task or objective, or a new one of the same kind — and
// nothing has been typed into it since: what was set aside can be brought
// back without writing over anything.
func (w Window) recoverable() bool {
	d := w.draft
	return d.fields != nil && !w.form.Changed() &&
		d.Adding == w.form.Adding && d.ForObjective == w.form.ForObjective && d.Number == w.form.Number
}

// Choose is a button pressed by name, which is how a scenario says it and how
// a pointer says it.
func (w Window) Choose(name string) Window {
	switch strings.ToLower(name) {
	case "edit":
		return w.Edit()
	case "save":
		return w.Save()
	case "cmd+enter":
		return w.Save()
	case "choose":
		// The same word twice, and two different acts: on a form it opens the
		// list, and in the list it takes what the keyboard is on. A bar that
		// said "Open a list" would be naming the mechanism rather than what
		// somebody came to do.
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
		// Backing out of a list is backing out of the list, not of the form
		// under it: the form holds what somebody typed, and throwing that away
		// is what Cancel on a form means.
		if w.screen == OnChoosing {
			return w.leaving()
		}
		if w.question == aboutSweeping {
			w.screen, w.question = OnTheBoard, aboutDeleting
			return w
		}
		return w.Cancel()
	case "new":
		w.form, w.screen = adding(w.tab, w.addingOn()), OnForm
		return w
	case "take":
		if w.screen == OnCalendar {
			return w.dated(w.day)
		}
		return w.taking()
	case "keep":
		return w.keeping()
	case "close":
		// Closing the list is not cancelling the form under it. What was
		// ticked goes back into the field and the form is still there, with
		// everything else somebody had already typed into it still in it.
		if w.screen == OnChoosing {
			return w.keeping()
		}
		return w.Cancel()
	case "search":
		w.screen = OnSearch
		return w
	case "help":
		w.screen, w.helpFrom, w.spareHelp = OnHelp, 0, 0
		return w
	case "clear":
		// What is taken together is let go of first: it is the nearer of
		// the two, and the search is still there for the esc after.
		if len(w.taken) > 0 {
			w.taken = nil
			return w
		}
		w.looking = Written{}
		return w
	case "archived":
		w.archived = !w.archived
		return w
	case "all possible":
		return w.actingOnSeveral(true)
	case "skip flagged":
		return w.actingOnSeveral(false)
	case "delete", "agree":
		if w.screen == OnTheBoard && len(w.Taken()) > 0 {
			return w.askingSeveral(aboutDeletingSeveral)
		}
		if w.question == aboutSweeping {
			return w.sweeping()
		}
		if w.question == aboutLeaving {
			return w.throwingAway()
		}
		return w.Delete()
	case "throw away":
		return w.throwingAway()
	case "back":
		// Going back to the form is what the question was there for.
		w.screen, w.question = OnForm, aboutDeleting
		return w
	case "recover":
		w.form, w.draft = w.draft, Form{}
		return w
	case "archive", "restore":
		if w.screen == OnTheBoard && len(w.Taken()) > 0 {
			return w.askingSeveral(aboutArchivingSeveral)
		}
		return w.puttingAway()
	}
	return w
}

// saving says which operation a saved form means and what it carries. Adding
// and editing are one form, so they are one function here: the difference is
// the operation and the number, and nothing else about a form knows which of
// the two it is.
//
// The title is trimmed on its way out and nowhere else: a title is what a task
// is called, and what somebody called it does not include the space they left
// after it.
func (w Window) saving() (engine.Operation, engine.Input) {
	adding, editing := engine.TaskAdd, engine.TaskEdit
	if w.form.ForObjective {
		adding, editing = engine.ObjectiveAdd, engine.ObjectiveEdit
	}
	return reading.Saved(w.form.fields, w.form.Adding, w.form.Number, adding, editing)
}

// addingOn is the column a task added now belongs to: the one what is chosen
// stands in, the backlog when nothing is — and the first of the stopped states
// when what is chosen is in the band, the way the terminal adds there.
func (w Window) addingOn() string {
	column, inBand, _, found := reading.Where(w.Showing().Cards, w.archived, w.chosen)
	if !found {
		return reading.StateAt(0)
	}
	if inBand {
		return reading.Stalled[0]
	}
	return reading.StateAt(column)
}

// Delete asks first, and deletes the second time. It is the one thing this
// window does that loses work, and work lost without being asked is work
// somebody has to remember they lost.
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

// deleting is the second time: it does what was agreed to.
func (w Window) deleting() Window {
	op := engine.TaskDelete
	if w.chosenGoal {
		op = engine.ObjectiveDelete
	}
	w.screen = OnTheBoard
	if _, err := w.work.Ask(op, reading.Numbered(w.chosen)); err != nil {
		w.err, w.wrongFor = err, 0
		return w
	}
	w.chosen, w.chosenGoal = 0, false
	return w.Look()
}

// Choosing is what the open list offers: every objective there is, in the order
// the objectives tab shows them, so that a list and the tab behind it cannot
// disagree about what comes first.
func (w Window) Choosing() []Choice {
	if w.screen != OnChoosing {
		return nil
	}
	return w.choices()
}

// ChoosingSays is what the list reads when there is nothing to serve. A store
// with no objectives in it says so rather than opening an empty panel, which
// reads as a list that failed to load.
const ChoosingSays = "nothing to serve yet"

// ChoosingAt is where in that list the keyboard sits.
func (w Window) ChoosingAt() int { return w.choosing }

// Ticked says whether an objective is in, as the open list has it.
func (w Window) Ticked(number int) bool { return w.ticked[number] }

// FieldSays is what a field reads on screen, which is not always what it
// carries. The objectives field carries the numbers the engine takes and shows
// the names a person knows: nobody knows an objective by its number, and the
// one place that says what a number means is the other tab.
func (w Window) FieldSays(field reading.Field) string {
	var said []string
	switch field.Arg {
	case "objective":
		for _, number := range reading.NumbersIn(field.Value) {
			said = append(said, "#"+strconv.Itoa(number)+" "+w.goalTitle(number))
		}
	case "parent":
		// The number and what it is called. A number alone is what somebody
		// has to go and look up, which is the whole of why the list exists.
		for _, number := range reading.NumbersIn(field.Value) {
			said = append(said, "#"+strconv.Itoa(number)+" "+w.cardTitle(number))
		}
	default:
		return field.Value
	}
	return strings.Join(said, ", ")
}

// cardTitle is what a task is called, or its number again when the snapshot
// has never heard of it — the same rule an objective's title goes by.
func (w Window) cardTitle(number int) string {
	if c, found := w.snap.Card(number); found {
		return c.Title
	}
	return strconv.Itoa(number)
}

// goalTitle is what an objective is called, or its number again when the
// snapshot has never heard of it — a number that names nothing is still truer
// than an empty space where a name should be.
func (w Window) goalTitle(number int) string {
	g, _ := w.snap.Goal(number)
	return g.Title
}

// choosing opens the list over the form, with what the field already carries
// ticked. Opened from the objectives field and from nowhere else: it is the
// one field a number was typed into, and the only one a list would answer.
func (w Window) choosingList() Window {
	on := w.form.Fields()[w.form.On()]
	if !chosenFromAList(on) {
		return w
	}
	w.ticked, w.choosing, w.choosingFrom = map[int]bool{}, 0, 0
	w.choosingFor, w.filtering = on.Arg, Written{}
	if on.Arg == "label" {
		// A label is named, not numbered: one made of digits is a name too.
		w = w.openingLabels(on.Value)
	} else {
		for _, number := range reading.NumbersIn(on.Value) {
			w.ticked[number] = true
		}
	}
	// A closed set is listed in its own order, and what the field holds is the
	// one of it already taken.
	if len(on.Choices) > 0 {
		w.ticked = map[int]bool{}
		for i, choice := range on.Choices {
			if choice == on.Value {
				w.ticked[i+1] = true
			}
		}
	}
	w.screen = OnChoosing
	// Opened on what it already carries, so taking it off is one key rather
	// than a walk down the list to find it.
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
// A task and an objective are the same thing to a list, which is why the list
// is one piece of code and not two.
type Choice struct {
	Number int
	Title  string
	// State is the column the task sits in, said on the row. Empty for an
	// objective, which sits in no column a list has any use for.
	State string
	// Creates is the label taking this row creates: the row that offers a
	// label nobody carries yet, under the name it will have.
	Creates string
}

// RowSays is what a row of the open list reads: whether it is taken, then what
// it is known by. A task and an objective are known by their number too; a
// column is known by its name alone, and the number it is listed under is the
// list's, not anything somebody has seen before.
func (w Window) RowSays(one Choice) string {
	said := "   "
	if w.Ticked(one.Number) {
		said = TickedMark + "  "
	}
	if len(w.closedSet()) == 0 && w.choosingFor != "label" {
		said += Numbered(reading.Card{Number: one.Number}) + "  "
	}
	return said + one.Title
}

// choices is everything the open list offers, narrowed by what is typed into
// it. The narrowing is the board's own — a list that matched its own way would
// answer a different set from the search box two keystrokes away.
func (w Window) choices() []Choice {
	if w.choosingFor == "label" {
		return w.labelChoices()
	}
	looking := w.filtering.Reads()
	var out []Choice
	// A closed set is offered in the order the domain gives it, numbered by
	// where it stands in that order, since what is kept is a name and not a
	// number anybody has seen.
	if set := w.closedSet(); len(set) > 0 {
		for i, choice := range set {
			if strings.Contains(choice, strings.ToLower(looking)) {
				out = append(out, Choice{Number: i + 1, Title: choice})
			}
		}
		return out
	}
	if w.choosingFor == "parent" {
		for _, c := range reading.Parentable(w.snap.Cards, w.form.Number, looking) {
			out = append(out, Choice{Number: c.Number, Title: c.Title, State: c.State})
		}
		return out
	}
	// The same search the search box makes, so the list and the box beside it
	// never disagree about what a word reaches.
	for _, g := range reading.FoundGoals(w.snap.Goals, looking) {
		out = append(out, Choice{Number: g.Number, Title: g.Title})
	}
	return out
}

// Filtering is what has been typed into the open list.
func (w Window) Filtering() Written { return w.filtering }

// OneOnly says whether the open list takes one answer rather than several. A
// task hangs under one parent; it serves as many objectives as it serves.
// A column, a priority or a standing is one of a set, and only one.
func (w Window) OneOnly() bool { return w.choosingFor == "parent" || len(w.closedSet()) > 0 }

// closedSet is what the field the list is open for may hold, when that is a
// closed set; empty for a list of tasks or objectives.
func (w Window) closedSet() []string {
	if w.screen != OnChoosing {
		return nil
	}
	return w.form.Fields()[w.form.On()].Choices
}

// inChoosing is what a key does while the list has the window.
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
		// A list is corrected the way every other text here is. Without it a
		// mistyped letter meant closing the list and opening it again.
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

// taking takes what the keyboard is on, or gives it back. One answer closes
// the list with it, since there is nothing left to say; several stay open,
// because taking them one at a time is what the list is for.
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
		// Taken again is taken off. One answer is still an answer that can be
		// withdrawn, and a field a list fills is a field only a list empties —
		// except a closed set, which has no empty answer: a task stands in a
		// column whatever is done to the list.
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

// bringingIntoSight moves the run of rows in sight so the one the keyboard is
// on is inside it, which is what a column does when a card is chosen off the
// bottom of it.
// Counted in rows and not in pixels. The same walk in pixels had to divide by
// a row's height at the end, and a run rarely divides into whole rows — the
// remainder was an arithmetic nothing could judge, which is what this file's
// own header says to write out as numbers instead.
func (w Window) bringingIntoSight() int {
	room := w.ChoosingRoom()
	return min(w.choosing, max(w.choosingFrom, w.choosing-room+1))
}

// leaving closes the list and puts nothing back, which is what backing out of
// one means.
func (w Window) leaving() Window {
	w.screen = OnForm
	return w
}

// keeping closes the list and puts what was ticked back in the field, as the
// numbers the engine takes. What a person reads there is FieldSays; what is
// sent is this, because a surface that made up its own spelling of an argument
// would be a surface deciding something.
func (w Window) keeping() Window {
	if w.choosingFor == "label" {
		return w.keepingLabels()
	}
	// A closed set keeps the name taken, and the one it held when nothing was:
	// a column is never left empty for a list closed on a narrowing.
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
	w.form = w.form.Writes(Ending(reading.Numbers(numbers)))
	w.screen = OnForm
	return w
}
