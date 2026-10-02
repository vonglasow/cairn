package gui

// The modal that adds and edits, and the two buttons that end it.
//
// Nothing here lists the fields. They come from reading, which says the word a
// person reads, the closed set when there is one, the field that wants more
// than one line, and which engine argument each fills — so an attribute added
// to the engine and to the terminal's form arrives here with no line written.

import (
	"sort"
	"strconv"
	"strings"

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
	f.at = min(len(f.fields)-1, max(0, at))
	f.caret, f.held = len([]rune(f.fields[f.at].Value)), 0
	return f
}

// adding opens an empty form for whichever tab asked.
func adding(tab Tab) Form {
	if tab == Objectives {
		return asOpened(Form{Adding: true, ForObjective: true, fields: reading.GoalFields(reading.Goal{})})
	}
	return asOpened(Form{Adding: true, fields: reading.TaskFields(reading.Card{})})
}

// editing opens a form on what a task already carries, its parts included.
// They are reachable here and nowhere else: a part is off the board, so a task
// that carried one could neither be changed nor deleted once it did.
func editing(c reading.Card, parts []reading.Card) Form {
	fields := append(reading.TaskFields(c), reading.Field{
		Name: "parts", Value: reading.PartsSaid(parts), Long: true,
	})
	return asOpened(Form{Number: c.Number, fields: fields})
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
	// An opened field is the form's field still, with the panel to itself.
	if w.screen != OnForm && w.screen != OnWriting {
		return w
	}
	w.form = w.form.Writes(w.form.Writing().Types(text))
	return w
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
	box := Over(w.width, w.height)
	fields := w.form.Fields()
	for i := range fields {
		top := FieldTop(box, fields, i)
		name, value := FieldRow(box, top)
		// A field that wants several rows is answered anywhere in them.
		rows := RowsOf(box, fields, i)
		name.H, value.H = lineHigh*rows, lineHigh*rows
		if name.In(p) || value.In(p) {
			return i, true
		}
	}
	return 0, false
}

// into says how many characters into a field's value a point falls, so that
// clicking in the middle of a word puts the caret there rather than at the end.
func into(value Box, p Point) int {
	return max(0, (p.X-value.X)/perGlyph)
}

// reaching says which character of a field a point lands on, line and column
// both. It measures inside the frame the words are drawn in, where into()
// measured from the value box and so sat six points off, and it reads the line
// the point is down to, which into() never did.
func (w Window) reaching(at int, p Point) int {
	fields := w.form.Fields()
	box := Over(w.width, w.height)
	_, value := FieldRow(box, FieldTop(box, fields, at))
	inside := Inside(Framed(value, RowsOf(box, fields, at)))
	lines := Wrapped(fields[at].Value, room(inside))
	return lines.Reaches(inside, p, perGlyph)
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
		for _, g := range w.snap.Goals {
			if g.Number == w.chosen {
				w.form, w.screen = editingGoal(g), OnForm
				return w
			}
		}
		return w
	}
	for _, c := range w.snap.Cards {
		if c.Number == w.chosen {
			w.form, w.screen = editing(c, reading.Hanging(w.snap.Cards, c.Number)), OnForm
			return w
		}
	}
	return w
}

// Form is what the modal holds, for whoever draws it.
func (w Window) Form() Form { return w.form }

// Save writes what the form holds and closes it. One act in the history per
// save, which is the whole reason this is a modal and not twelve fields edited
// in place on a card.
func (w Window) Save() Window {
	// An opened field is the form still, so saving from inside it saves the
	// form: a key that did nothing there would make somebody close the field
	// to use it.
	if w.screen != OnForm && w.screen != OnWriting {
		return w
	}
	op, asked := w.saving()
	if _, err := w.work.Ask(op, asked); err != nil {
		w.err = err
		return w
	}
	if err := w.savingParts(); err != nil {
		w.err = err
		return w
	}
	w.screen, w.form = OnTheBoard, Form{}
	return w.Look()
}

// savingParts settles what hangs under the task with what the form's own field
// says: a line carrying a number is that part, named and ticked as it reads; a
// line carrying none is a part nobody has written down yet; a number the field
// no longer carries is a part that has gone.
//
// One operation each, through the engine like everything else. The engine has
// no argument for a set of parts and should not grow one: what it is asked for
// here is adding a task, changing a task and deleting a task, which is what
// these are.
func (w Window) savingParts() error {
	if w.form.Adding {
		return nil
	}
	said, holds := "", false
	for _, field := range w.form.fields {
		if field.Name == "parts" {
			said, holds = field.Value, true
		}
	}
	if !holds {
		return nil
	}
	parent := strconv.Itoa(w.form.Number)
	want := reading.PartsIn(said)
	kept := map[int]bool{}
	for _, part := range want {
		state := string(domain.Todo)
		if part.Done {
			state = string(domain.Done)
		}
		if part.Number == 0 {
			if _, err := w.work.Ask(engine.TaskAdd, engine.Input{Given: map[string]string{
				"title": part.Title, "parent": parent, "state": state,
			}}); err != nil {
				return err
			}
			continue
		}
		kept[part.Number] = true
		if _, err := w.work.Ask(engine.TaskEdit, engine.Input{Given: map[string]string{
			"number": strconv.Itoa(part.Number), "title": part.Title, "state": state,
		}}); err != nil {
			return err
		}
	}
	for _, was := range reading.Hanging(w.snap.Cards, w.form.Number) {
		if kept[was.Number] {
			continue
		}
		if _, err := w.work.Ask(engine.TaskDelete, reading.Numbered(was.Number)); err != nil {
			return err
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
	// A form somebody typed into is asked about before it goes. The window
	// guards a deletion of two words; ten minutes of typing went on an esc
	// with no question at all.
	if w.screen == OnForm && w.form.Changed() {
		w.screen, w.question = OnAsking, aboutLeaving
		return w
	}
	// Anywhere else it closes what is open and nothing more. A search
	// narrowing the board is not what was opened: a task found by searching,
	// opened, changed and put away used to give back the whole board, leaving
	// somebody to type their search again to carry on where they were.
	w.screen, w.form, w.question = OnTheBoard, Form{}, aboutDeleting
	return w
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
		return w.choosingList()
	case "done":
		return w.marking()
	case "cancel":
		// Backing out of a list is backing out of the list, not of the form
		// under it: the form holds what somebody typed, and throwing that away
		// is what Cancel on a form means.
		if w.screen == OnChoosing {
			return w.leaving()
		}
		// Backing out of the question about leaving a form goes back to the
		// form, which is the whole point of having asked.
		if w.question == aboutLeaving {
			w.screen, w.question = OnForm, aboutDeleting
			return w
		}
		if w.question == aboutSweeping {
			w.screen, w.question = OnTheBoard, aboutDeleting
			return w
		}
		return w.Cancel()
	case "new":
		w.form, w.screen = adding(w.tab), OnForm
		return w
	case "take":
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
		// Nor is closing an opened field: what was written is the field's
		// value already, and the form is still there underneath.
		if w.screen == OnWriting {
			w.screen = OnForm
			return w
		}
		return w.Cancel()
	case "search":
		w.screen = OnSearch
		return w
	case "help":
		w.screen = OnHelp
		return w
	case "clear":
		w.looking = Written{}
		return w
	case "archived":
		w.archived = !w.archived
		return w
	case "delete", "agree":
		if w.question == aboutSweeping {
			return w.sweeping()
		}
		if w.question == aboutLeaving {
			w.screen, w.form, w.question = OnTheBoard, Form{}, aboutDeleting
			return w
		}
		return w.Delete()
	case "archive", "restore":
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
	op := editing
	if w.form.Adding {
		op = adding
	}
	asked := reading.Asked(w.form.fields, op)
	asked.Given["title"] = strings.TrimSpace(w.form.ValueOf("title"))
	if !w.form.Adding {
		asked.Given["number"] = strconv.Itoa(w.form.Number)
	}
	return op, asked
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
	if _, err := w.work.Ask(op, reading.Numbered(w.chosen)); err != nil {
		w.err, w.screen = err, OnTheBoard
		return w
	}
	w.screen, w.chosen, w.chosenGoal = OnTheBoard, 0, false
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
	for _, c := range w.snap.Cards {
		if c.Number == number {
			return c.Title
		}
	}
	return strconv.Itoa(number)
}

// goalTitle is what an objective is called, or its number again when the
// snapshot has never heard of it — a number that names nothing is still truer
// than an empty space where a name should be.
func (w Window) goalTitle(number int) string {
	for _, g := range w.snap.Goals {
		if g.Number == number {
			return g.Title
		}
	}
	return ""
}

// choosing opens the list over the form, with what the field already carries
// ticked. Opened from the objectives field and from nowhere else: it is the
// one field a number was typed into, and the only one a list would answer.
func (w Window) choosingList() Window {
	on := w.form.Fields()[w.form.On()]
	if !chosenFromAList(on.Arg) {
		return w
	}
	w.ticked, w.choosing, w.choosingFrom = map[int]bool{}, 0, 0
	w.choosingFor, w.filtering = on.Arg, Written{}
	for _, number := range reading.NumbersIn(on.Value) {
		w.ticked[number] = true
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
}

// choices is everything the open list offers, narrowed by what is typed into
// it. The narrowing is the board's own — a list that matched its own way would
// answer a different set from the search box two keystrokes away.
func (w Window) choices() []Choice {
	looking := w.filtering.Reads()
	var out []Choice
	if w.choosingFor == "parent" {
		for _, c := range reading.OnTheBoard(w.snap.Cards, false) {
			// Nothing hangs under itself, and nothing hangs under a part.
			if c.Number == w.form.Number || c.Parent != 0 {
				continue
			}
			if looking == "" || reading.ReachesCard(c, looking) {
				out = append(out, Choice{Number: c.Number, Title: c.Title, State: c.State})
			}
		}
		// In the order somebody reaches for them, which is not the order the
		// board reads. A board goes from what has not been started to what is
		// finished; work is hung under what is being worked on, and the board's
		// order put two dozen backlog tasks in front of it.
		sort.SliceStable(out, func(i, j int) bool {
			return reachedFor(out[i].State) < reachedFor(out[j].State)
		})
		return out
	}
	for _, g := range reading.Sorted(w.snap.Goals) {
		if looking == "" || strings.Contains(strings.ToLower(g.Title), strings.ToLower(looking)) {
			out = append(out, Choice{Number: g.Number, Title: g.Title})
		}
	}
	return out
}

// reachedFor is how soon a column is reached for when something is being hung
// under it: what is being worked on, then what is next, then what is stopped,
// then what has not been started, and last what is over.
func reachedFor(state string) int {
	switch domain.State(state) {
	case domain.Current:
		return 0
	case domain.Todo:
		return 1
	case domain.Blocked, domain.Watching, domain.Paused:
		return 2
	case domain.Backlog:
		return 3
	}
	return 4
}

// Filtering is what has been typed into the open list.
func (w Window) Filtering() Written { return w.filtering }

// OneOnly says whether the open list takes one answer rather than several. A
// task hangs under one parent; it serves as many objectives as it serves.
func (w Window) OneOnly() bool { return w.choosingFor == "parent" }

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
	case "enter", "return":
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
	offered := w.Choosing()
	if w.choosing >= len(offered) {
		return w
	}
	number := offered[w.choosing].Number
	if w.OneOnly() {
		// Taken again is taken off. One answer is still an answer that can be
		// withdrawn, and a field a list fills is a field only a list empties.
		if w.ticked[number] {
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
