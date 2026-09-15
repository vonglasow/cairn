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
}

// Writing is the field the keyboard is on, as a text with its caret.
func (f Form) Writing() Written {
	if len(f.fields) == 0 {
		return Written{}
	}
	return Written{Value: f.fields[f.at].Value, At: f.caret}
}

// Writes puts a written text back into the field the keyboard is on.
func (f Form) Writes(written Written) Form {
	if len(f.fields) == 0 {
		return f
	}
	f = f.Set(f.fields[f.at].Name, written.Reads())
	f.caret = written.At
	return f
}

// onto moves the keyboard to a field and puts the caret at its end.
func (f Form) onto(at int) Form {
	f.at = min(len(f.fields)-1, max(0, at))
	f.caret = len([]rune(f.fields[f.at].Value))
	return f
}

// adding opens an empty form for whichever tab asked.
func adding(tab Tab) Form {
	if tab == Objectives {
		return Form{Adding: true, ForObjective: true, fields: reading.GoalFields(reading.Goal{})}
	}
	return Form{Adding: true, fields: reading.TaskFields(reading.Card{})}
}

// editing opens a form on what a task already carries.
func editing(c reading.Card) Form {
	return Form{Number: c.Number, fields: reading.TaskFields(c)}
}

// editingGoal is the same for an objective.
func editingGoal(g reading.Goal) Form {
	return Form{Number: g.Number, ForObjective: true, fields: reading.GoalFields(g)}
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
	if w.screen != OnForm {
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
	w.form.caret = w.form.Writing().CaretAt(characters).At
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
	for i, field := range fields {
		top := FieldTop(box, fields, i)
		name, value := FieldRow(box, top)
		// A field that wants several rows is answered anywhere in them.
		name.H, value.H = lineHigh*RowsOf(field), lineHigh*RowsOf(field)
		if name.In(p) || value.In(p) {
			return i, true
		}
	}
	return 0, false
}

// into says how many characters into a field's value a point falls, so that
// clicking in the middle of a word puts the caret there rather than at the end.
func into(value Box, p Point) int {
	return max(0, (p.X-value.X-padding)/perGlyph)
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
			w.form, w.screen = editing(c), OnForm
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
	if w.screen != OnForm {
		return w
	}
	op, asked := w.saving()
	if _, err := w.work.Ask(op, asked); err != nil {
		w.err = err
		return w
	}
	w.screen, w.form = OnTheBoard, Form{}
	return w.Look()
}

// Cancel closes whatever is open and writes nothing. Whatever, and not only a
// form: the question asked before deleting is answered with the same word, and
// a Cancel that only knew about forms left it open.
func (w Window) Cancel() Window {
	if w.screen == OnTheBoard {
		return w
	}
	w.screen, w.form, w.looking = OnTheBoard, Form{}, Written{}
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
	case "choose":
		return w.choosingList()
	case "cancel":
		return w.Cancel()
	case "new":
		w.form, w.screen = adding(w.tab), OnForm
		return w
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
		w.screen = OnHelp
		return w
	case "clear":
		w.looking = Written{}
		return w
	case "delete":
		return w.Delete()
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
	if w.screen == OnAsking {
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
func (w Window) Choosing() []reading.Goal {
	if w.screen != OnChoosing {
		return nil
	}
	return reading.Sorted(w.snap.Goals)
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
	if field.Arg != "objective" {
		return field.Value
	}
	var said []string
	for _, number := range reading.NumbersIn(field.Value) {
		said = append(said, "#"+strconv.Itoa(number)+" "+w.goalTitle(number))
	}
	return strings.Join(said, ", ")
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
	if on.Arg != "objective" {
		return w
	}
	w.ticked, w.choosing = map[int]bool{}, 0
	for _, number := range reading.NumbersIn(on.Value) {
		w.ticked[number] = true
	}
	w.screen = OnChoosing
	return w
}

// inChoosing is what a key does while the list has the window.
func (w Window) inChoosing(name string) Window {
	offered := w.Choosing()
	switch name {
	case "down", "j":
		w.choosing = min(max(0, len(offered)-1), w.choosing+1)
	case "up", "k":
		w.choosing = max(0, w.choosing-1)
	case " ", "space":
		if w.choosing < len(offered) {
			number := offered[w.choosing].Number
			w.ticked[number] = !w.ticked[number]
		}
	case "esc", "enter", "return":
		return w.keeping()
	}
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
