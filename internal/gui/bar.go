package gui

import (
	"strconv"
	"strings"

	"cairn/internal/engine"
	"cairn/internal/reading"
)

// What the bar offers and what the help says: the pages and their heads, the
// key for each offer, the questions asked before losing work.

// Pages is what the two tabs are called, in the order they are drawn. Here
// and not in the drawing, because which page somebody is on is a thing the
// window knows and the drawing only shows.
var Pages = []string{"TASKS", "OBJECTIVES"}

// Heads is what the panel that is open says it is, written across its top.
// Nothing when no panel is open: the board is its own heading.
func Heads(w Window) string { return HeadsOn(w, w.screen) }

// HeadsOn is what a given panel says it is, whether or not it is the one on
// top. A panel drawn under another is still drawn, and asking the window what
// it was on made the form under an open list read the list's own heading.
func HeadsOn(w Window, screen Screen) string {
	switch screen {
	case OnChoosing:
		// What is being chosen, and what for. A list with no heading is a list
		// somebody has to remember the way into.
		what := "objectives for"
		if w.OneOnly() {
			what = "a " + w.form.Fields()[w.form.On()].Name + " for"
		}
		return what + " " + w.formKind() + " " + strconv.Itoa(w.form.Number)
	case OnHelp:
		return "What everything does"
	case OnAsking:
		if w.question == aboutLeaving {
			return "Throw away what you typed?"
		}
		if w.question == aboutSweeping {
			return "Take every task in " + strings.ToUpper(w.sweeps()) + " off the board?"
		}
		return "Delete " + w.kind() + " " + strconv.Itoa(w.chosen) + "?"
	case OnDetail:
		// The whole title, on the line beside the number. It is the one line
		// on the panel that was empty, it holds ninety-nine characters at the
		// smallest window this board draws in, and the title is what somebody
		// opened the task to read — the field below it cuts at thirty.
		return headed(w.kind()+" "+strconv.Itoa(w.chosen), w.titled())
	case OnForm:
		if w.form.Adding {
			return "New " + w.formKind()
		}
		return w.formKind() + " " + strconv.Itoa(w.form.Number)
	}
	return ""
}

// kind is the word for what is chosen, and formKind for what a form is about.
// Two of them because a form can be open over nothing chosen at all.
func (w Window) kind() string {
	if w.chosenGoal {
		return "objective"
	}
	return "task"
}

// titled is the title of whatever is open, read off the fields it shows so a
// task and an objective answer the same way.
func (w Window) titled() string {
	for _, field := range w.Shown() {
		if field.Arg == "title" {
			return field.Value
		}
	}
	return ""
}

// headed joins a head to what it is about, and is the head alone when there is
// nothing to join.
func headed(head, about string) string {
	if about == "" {
		return head
	}
	return head + " · " + about
}

func (w Window) formKind() string {
	if w.form.ForObjective {
		return "objective"
	}
	return "task"
}

// Shortcut is the key that does an offer without the pointer. Every offer has
// one: the bar is what somebody reads to find out what a key does, and an
// offer with no key is a word the bar cannot teach anything about.
//
// Delete is d and not the delete key, because that is what the terminal board
// has always used and one of the two had to give.
func Shortcut(offer string) string {
	switch strings.ToLower(offer) {
	case "new":
		return "cmd+n"
	case "search":
		return "/"
	case "help":
		return "?"
	case "archived":
		return "shift+a"
	case "archive", "restore":
		return "x"
	case "take":
		return "enter"
	case "keep":
		// Not space, which a list with a line to type into cannot have: a
		// title has spaces in it, so the filter needs them and the shortcut
		// cannot. Taking a row and being done with the list are two acts, and
		// the second gets the modifier.
		return "cmd+enter"
	case "edit":
		return "e"
	case "delete":
		return "d"
	case "agree":
		// Not d, which is what opened the question. A quick dd — which is how
		// a deletion is typed in the grammar this product borrows — used to
		// remove a task before a word of the question had been read.
		return "enter"
	case "choose":
		return "enter"
	case "save":
		// One key, and not enter. A form has two acts and enter is spent on
		// the other: it opens the list a field answers with, and it breaks a
		// line in a field with room for one. ⌘⏎ saves wherever the keyboard
		// is, which is what the bar says.
		return "cmd+enter"
	case "done":
		return "space"
	case "close", "cancel", "clear":
		return "esc"
	}
	return ""
}

// Pretty is how the bar writes a key, which is not how a key arrives. The two
// are told apart on purpose: a bar that wrote "cmd+n" would be reading out a
// name meant for a program, and a key looked up by "⌘N" would never be found.
func Pretty(key string) string {
	switch key {
	case "cmd+n":
		return "⌘N"
	case "cmd+enter":
		return "⌘⏎"
	case "shift+a":
		return "⇧A"
	case "enter":
		return "⏎"
	}
	return key
}

// Said is one line of a panel that only says something: a key, or nothing,
// and what it means. Its own type rather than a form's field, which carries an
// argument and a set of choices a panel of words has no use for.
type Said struct {
	Key, Means string
}

// Helping is what everything does, as a person reads it.
func (w Window) Helping() []Said {
	return []Said{
		{Key: "click", Means: "choose a card"},
		{Key: "double-click", Means: "open it, and see everything it carries"},
		{Key: "drag", Means: "move it — the column it lands in is its new state"},
		{Key: "drag into a lane", Means: "blocked, watching or paused"},
		{Key: "wheel", Means: "scroll what is under the pointer — a column, a lane, the objectives, a history"},
		{Key: "↑↓ under a task", Means: "walk what hangs under it"},
		{Key: "space under a task", Means: "mark the part the keyboard is on done, or not"},
		{Key: "wheel on a history", Means: "scroll it — the bar beside it can be pulled too"},
		{Key: "on an open task", Means: "every text can be held: drag across it, double-click a word, triple-click a paragraph — ⌘A holds it all, ⌘C copies"},
		{Key: "^D ^U on an open task", Means: "walk its description half a box at a time — the wheel and the bar do too"},
		{Key: "e", Means: "edit what is chosen, on the board or on an open task"},
		{Key: "f", Means: "take up what is chosen, or let it go — several at once is fine"},
		{Key: "d", Means: "delete what is chosen — it asks first"},
		{Key: "x", Means: "take what is chosen off the board, or put it back"},
		{Key: "⇧X", Means: "take every task in that card's column off the board — it asks first"},
		{Key: "←→↑↓", Means: "walk the board, the band included"},
		{Key: "⇧←→↑↓", Means: "move what is chosen — sideways changes its column"},
		{Key: "⏎ on the board", Means: "open what is chosen"},
		{Key: "⏎ on a question", Means: "agree to it"},
		{Key: "⏎ in a form", Means: "a new line in the description, the next field anywhere else — it never saves"},
		{Key: "↑↓ in the description", Means: "walk its lines — tab leaves it, and the wheel and the bar scroll it"},
		{Key: "⌘⏎ or ⌘S in a form", Means: "save, wherever the keyboard is"},
		{Key: "⌘A ⌘C ⌘X ⌘V", Means: "hold all of a field, copy, cut, paste"},
		{Key: "⇧←→ in a field", Means: "hold a run of text — dragging across it does the same"},
		{Key: "⏎ in a list", Means: "take the row · ⌘⏎ keeps the list · j and k walk it"},
		{Key: "⇧A", Means: "what was archived, on the board or off it — it reads faintly"},
		{Key: "⌘N", Means: "add"},
		{Key: "/ or ⌘F", Means: "search — it reaches the description and the reason"},
		{Key: "⏎ in a search", Means: "keep what it narrowed and go back to the board"},
		{Key: "tab", Means: "the other page — shift+tab the one before"},
		{Key: "tab, ↑↓ in a form", Means: "walk its fields — inside the description ↑↓ walk its lines"},
		{Key: "←→", Means: "change a field that has a closed set"},
		{Key: "esc", Means: "back out of whatever is open"},
		{Key: "?", Means: "this help"},
	}
}

// Offers is what the bar says can be done here. It changes with what is on
// screen and with what is chosen, because an offer that does nothing is worse
// than no offer.
func (w Window) Offers() []string {
	switch w.screen {
	case OnChoosing:
		// Take and Done, not Choose twice. Space takes what the keyboard is
		// on, which is the gesture a hand already has for a list; enter is
		// done with the whole list, which is a different act and says so.
		return []string{"Take", "Keep", "Cancel"}
	case OnForm:
		// The one field a list answers says so on the bar. enter is Save's key
		// everywhere else in a form, and a bar that went on offering Save here
		// while enter opened a list would be a bar that lies — which is the
		// one thing this bar exists not to do. Save comes back the moment the
		// keyboard leaves this field.
		// Save is always offered. It used to give way to Choose on a field a
		// list answers, and with the keyboard left on that field after the
		// list closed there was no way to save at all — the one act a form
		// exists for, unreachable for having chosen something.
		if chosenFromAList(w.form.Fields()[w.form.On()]) {
			return []string{"Choose", "Save", "Cancel"}
		}
		return []string{"Save", "Cancel"}
	case OnDetail:
		// Done only where there is a part to mark. The bar is where a key is
		// looked for, so it says what space does exactly where space does it,
		// and says nothing of it on a task nothing hangs under.
		if len(w.Under()) > 0 && !w.chosenGoal {
			return []string{"Done", "Edit", "Close"}
		}
		return []string{"Edit", "Close"}
	case OnAsking:
		return []string{"Agree", "Cancel"}
	case OnHelp:
		return []string{"Close"}
	case OnSearch:
		// One word, at the right of the bar, where the box no longer reaches.
		// Offering nothing left the pointer with no way out of a search and no
		// way to act on the card it had just found.
		return []string{"Close"}
	}
	offers := []string{"New", "Search", "Archived", "Help"}
	if !w.looking.Empty() {
		offers = append(offers, "Clear")
	}
	if w.chosen != 0 {
		// Archive and Restore are the same key on the same card: the bar names
		// which way it goes, so nobody has to work out which it will be.
		away := "Archive"
		if w.chosenIsArchived() {
			away = "Restore"
		}
		offers = append(offers, "Edit", away, "Delete")
	}
	return offers
}

// clickingList is what a click does inside an open list: a row is taken the
// way the keyboard takes it, and the line that narrows it takes the caret.
// Anywhere else inside the panel it does nothing, which is what a click on a
// panel's own ground has always done.
func (w Window) clickingList(p Point) Window {
	box := ChoosingOver(w.width, w.height)
	// A click puts the keyboard on the row; it does not take it. Taking is
	// space, and a double-click — a single click that took would make choosing
	// which row to read impossible without choosing it.
	if at, on := ChoosingAtPoint(box, len(w.Choosing()), w.choosingFrom, p); on {
		w.choosing = at
		return w
	}
	if looking := ChoosingLooking(box); looking.In(p) {
		w.filtering = w.filtering.CaretAt(into(Inside(looking), p, w.filtering.Reads(), w.measured()))
	}
	return w
}

// chosenFromAList says whether a field is answered by picking from a list
// rather than by typing. The objectives and the parent are numbers somebody
// would otherwise have to go and look up; a column or a priority is one of a
// closed set, and the list is where that set is read before one is picked.
func chosenFromAList(field reading.Field) bool {
	return field.Arg == "objective" || field.Arg == "parent" || len(field.Choices) > 0
}

// offerAt says which offer a point lands on, in the bar along the bottom or on
// the panel that has the window. Both say the same things and both answer,
// because a person clicks what they are looking at.
func (w Window) offerAt(p Point) (string, bool) {
	l := w.Layout()
	// The panel the offers are drawn on, which for a list is its own and not
	// the one everything else uses: drawn in one place and listened for in
	// another, three words looked like buttons and answered nowhere.
	panel := Over(w.width, w.height)
	if w.screen == OnChoosing {
		panel = ChoosingOver(w.width, w.height)
	}
	for i, offer := range w.Offers() {
		if OfferAt(l, i).In(p) {
			return offer, true
		}
		if w.screen != OnTheBoard && PanelOfferAt(panel, i).In(p) {
			return offer, true
		}
	}
	return "", false
}

// Question is what the one question screen is asking about.
type Question int

// The two things worth asking before doing: both lose work, and work lost
// without being asked is work somebody has to remember they lost.
const (
	aboutDeleting Question = iota
	aboutLeaving
	aboutSweeping
)

// sweeps is the column a sweep would take off the board: the one the chosen
// card stands in. Nothing when nothing is chosen, which is when there is no
// column to name and so nothing to ask about.
func (w Window) sweeps() string {
	if w.chosen == 0 || w.chosenGoal {
		return ""
	}
	state := w.stateOf(w.chosen)
	if reading.Stopped(state) {
		return ""
	}
	return state
}

// sweeping takes every task in that column off the board.
func (w Window) sweeping() Window {
	state := w.sweeps()
	w.screen, w.question = OnTheBoard, aboutDeleting
	if state == "" {
		return w
	}
	return w.asking(engine.TaskArchive, engine.Input{Given: map[string]string{"state": state}})
}
