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
var Pages = []string{"TASKS", "OBJECTIVES", "TIME"}

// formNamed is what the open form is over, as a heading names it: its kind
// and number, or new when it is being added — a task being added has no
// number yet, and "task 0" is nothing anybody added.
func (w Window) formNamed() string {
	if w.form.Adding {
		return "the new " + w.formKind()
	}
	return w.formKind() + " " + strconv.Itoa(w.form.Number)
}

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
		if w.choosingFor == "label" {
			what = "labels for"
		}
		if w.OneOnly() {
			what = "a " + w.form.Fields()[w.form.On()].Name + " for"
		}
		return what + " " + w.formNamed()
	case OnCalendar:
		return "a deadline for " + w.formNamed() + " — " + w.day.Format("January 2006")
	case OnBackups:
		return "Backups — enter puts one back"
	case OnHelp:
		return "What everything does"
	case OnAsking:
		if w.question == aboutLeaving {
			return "Throw away what you typed?"
		}
		if w.asksAboutSeveral() {
			return w.severalHead()
		}
		if w.question == aboutPuttingBack {
			return "Put backup " + strconv.Itoa(w.backups[w.backupOn].Number) + " back?"
		}
		if w.question == aboutBringingIn {
			return "Bring in " + w.bringingName + "?"
		}
		if w.question == aboutSweeping {
			return "Take every task in " + strings.ToUpper(w.sweeps()) + " off the board?"
		}
		return "Delete " + w.kind() + " " + strconv.Itoa(w.chosen) + "?"
	case OnForm:
		if w.form.Adding {
			return "New " + w.formKind()
		}
		// The title as it was saved, beside the number: what was opened,
		// said where it is read first. Not as it is being typed — the field
		// below says that, and a heading following every keystroke would be
		// a second copy of it.
		head := w.formKind() + " " + strconv.Itoa(w.form.Number)
		if saved := w.savedTitle(); saved != "" {
			head += " · " + saved
		}
		return head
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
	case "export":
		return "cmd+e"
	case "import":
		return "cmd+o"
	case "back up":
		return "cmd+b"
	case "backups":
		return "shift+b"
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
	case "close", "cancel", "clear", "throw away":
		return "esc"
	case "back", "all possible":
		return "enter"
	case "skip flagged":
		return "s"
	case "recover":
		return "cmd+z"
	case "no day":
		return "backspace"
	}
	return ""
}

// Pretty is how the bar writes a key, which is not how a key arrives. The two
// are told apart on purpose: a bar that wrote "cmd+n" would be reading out a
// name meant for a program, and a key looked up by "⌘N" would never be found.
func Pretty(key string) string {
	switch key {
	case "cmd+enter":
		return "⌘⏎"
	case "enter":
		return "⏎"
	}
	// A letter under a modifier is written the way the key is printed: one
	// rule rather than a line for each chord, which every new one forgot.
	if letter, ok := strings.CutPrefix(key, "cmd+"); ok {
		return "⌘" + strings.ToUpper(letter)
	}
	if letter, ok := strings.CutPrefix(key, "shift+"); ok {
		return "⇧" + strings.ToUpper(letter)
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
	lines := helping()
	for i, line := range lines {
		lines[i] = Said{Key: w.keyboard.Spelled(line.Key), Means: w.keyboard.Spelled(line.Means)}
	}
	return lines
}

// helping is the help as a Mac writes it, spelled for the keyboard in front of
// the person by Helping.
func helping() []Said {
	return []Said{
		{Key: "click", Means: "choose a card"},
		{Key: "double-click", Means: "open it, and see everything it carries"},
		{Key: "drag", Means: "move it — the column it lands in is its new state"},
		{Key: "drag into a lane", Means: "blocked, watching or paused"},
		{Key: "right click", Means: "what can be done to what is under it — the board's offers where there is nothing"},
		{Key: "wheel", Means: "scroll what is under the pointer — a column, a lane, the objectives, a history"},
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
		{Key: "⌘⏎ or ⌘S in a form", Means: "save now and close — a form also saves itself once the typing pauses"},
		{Key: "esc esc in a form", Means: "throw it away — set aside: ⌘Z on the same form brings it back"},
		{Key: "⌘A ⌘C ⌘X ⌘V", Means: "hold all of a field, copy, cut, paste"},
		{Key: "⇧←→ in a field", Means: "hold a run of text — dragging across it does the same"},
		{Key: "⌘←→ or home end", Means: "the start or the end of the line — ⌥←→ walks a word at a time, and ⇧ holds as either goes"},
		{Key: "⌦ ⌘⌫ ⌥⌫", Means: "take out what is after the caret, back to the start of the line, back a word"},
		{Key: "⇧↑↓ in the description", Means: "hold whole lines of it"},
		{Key: "⏎ in a list", Means: "take the row · ⌘⏎ keeps the list · j and k walk it"},
		{Key: "⏎ on a deadline", Means: "a calendar — the arrows walk days and weeks, ⇧←→ months, t today, ⌫ no day"},
		{Key: "⇧A", Means: "what was archived, on the board or off it — it reads faintly"},
		{Key: "⌘N", Means: "new"},
		{Key: "⌘B", Means: "back the store up — it says where the copy lies"},
		{Key: "⇧B", Means: "the backups — enter puts one back, after asking, the store backed up first"},
		{Key: "⌘E", Means: "export the store to a file chosen in the system's dialog"},
		{Key: "⌘O", Means: "import a store from a file chosen there — after asking, the store backed up first"},
		{Key: "/ or ⌘F", Means: "search — it reaches the description and the reason"},
		{Key: "⏎ in a search", Means: "keep what it narrowed and go back to the board"},
		{Key: "tab", Means: "the next page — shift+tab the one before"},
		{Key: "←→ on TIME", Means: "the week before or after — t this week"},
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
		// Take and Keep, not Choose twice. Enter takes the row the keyboard is
		// on; ⌘⏎ keeps what is ticked and is done with the list, which is a
		// different act and says so. Not space: a list narrows as it is typed
		// into, and a title has spaces in it.
		return []string{"Take", "Keep", "Cancel"}
	case OnCalendar:
		return []string{"Take", "No day", "Cancel"}
	case OnBackups:
		return []string{"Cancel"}
	case OnMenu:
		// The menu is its own offers: the bar under it says what the board
		// will offer again, faintly, as it does under every panel.
		return nil
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
		offers := []string{"Save", "Cancel"}
		if on := w.form.Fields()[w.form.On()]; chosenFromAList(on) || chosenOnACalendar(on) {
			offers = []string{"Choose", "Save", "Cancel"}
		}
		if w.recoverable() {
			offers = append(offers, "Recover")
		}
		return offers
	case OnAsking:
		// Leaving a form is answered the way a hand leaving it answers: esc
		// again throws it away — set aside, not lost — and enter goes back.
		if w.question == aboutLeaving {
			return []string{"Throw away", "Back"}
		}
		if w.asksAboutSeveral() {
			return w.severalOffers()
		}
		return []string{"Agree", "Cancel"}
	case OnHelp:
		return []string{"Close"}
	case OnSearch:
		// One word, at the right of the bar, where the box no longer reaches.
		// Offering nothing left the pointer with no way out of a search and no
		// way to act on the card it had just found.
		return []string{"Close"}
	}
	return w.BoardOffers()
}

// BoardOffers is what the bar under the board says, whatever is open over it.
// A panel offers its own at its foot; the bar under the board used to change
// to say the same, which read as the board itself having changed underneath.
func (w Window) BoardOffers() []string {
	offers := []string{"New", "Search", "Archived", "Help"}
	// Clear lets go of what is taken together, and then of a search.
	if !w.looking.Empty() || len(w.Taken()) > 0 {
		offers = append(offers, "Clear")
	}
	// What is taken together is archived or deleted, and not edited: a form
	// is one task.
	if len(w.Taken()) > 0 {
		if w.chosenGoal {
			return append(offers, "Delete")
		}
		return append(offers, "Archive", "Delete")
	}
	// With nothing chosen, the store as a whole: carried out to a file, or
	// brought in from one. The window said nothing of either.
	if w.chosen == 0 {
		return append(offers, "Export", "Import")
	}
	{
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
	box := w.ChoosingBox()
	// A click puts the keyboard on the row; it does not take it. Taking is
	// enter, and a double-click — a single click that took would make choosing
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
func chosenOnACalendar(field reading.Field) bool { return field.Arg == "deadline" }

// chosenFromAList says whether a field is answered by a list.
func chosenFromAList(field reading.Field) bool {
	return field.Arg == "objective" || field.Arg == "parent" || field.Arg == "label" || len(field.Choices) > 0
}

// offerAt says which offer a point lands on: in the bar along the bottom while
// the board has the window, on the panel's own foot while a panel has it. A
// person clicks what they are looking at, and that is where it is drawn.
func (w Window) offerAt(p Point) (string, bool) {
	// The panel the offers are drawn on, which for a list is its own and not
	// the one everything else uses: drawn in one place and listened for in
	// another, three words looked like buttons and answered nowhere.
	panel := w.panel()
	// The bar under the board answers only while nothing is over it: under a
	// panel it is drawn faint, saying what the board will offer again, and
	// what can be done now is on the panel.
	alone := w.screen == OnTheBoard || w.screen == OnSearch
	along, bar := w.offersAlong(panel), w.BarOffers()
	for i, offer := range w.Offers() {
		if alone && bar[i].In(p) {
			return offer, true
		}
		if !alone && along[i].In(p) {
			return offer, true
		}
	}
	return "", false
}

// PanelOffers is where the offers of the panel that has the window are drawn,
// and so where they are clicked.
func (w Window) PanelOffers() []Box { return w.offersAlong(w.panel()) }

// offersAlong lays a panel's offers along its foot, each as wide as what it
// says and never narrower than an offer is: with one width for all,
// "backspace  No day" ran into the "esc  Cancel" after it.
func (w Window) offersAlong(panel Box) []Box {
	return laidAlong(w.widthsOf(w.Offers()), func(i int) Box { return PanelOfferAt(panel, i) }, panel.X+padding)
}

// BarOffers is where the bar under the board draws what it offers, and so
// where those are clicked: each as wide as what it says, as a panel's are.
func (w Window) BarOffers() []Box { return w.barAlong(w.Offers()) }

// barAlong lays offers along the bar under the board.
func (w Window) barAlong(offers []string) []Box {
	l := w.Layout()
	return laidAlong(w.widthsOf(offers), func(i int) Box { return OfferAt(l, i) }, l.Bar.X)
}

// laidAlong lays boxes of the widths given one after another from x, each
// on the row its slot is on.
func laidAlong(widths []int, slot func(int) Box, x int) []Box {
	var boxes []Box
	for i, wide := range widths {
		box := slot(i)
		box.X, box.W = x, wide
		boxes = append(boxes, box)
		x += wide
	}
	return boxes
}

// offerWidths is how wide each offer is drawn: what it says, and never
// narrower than an offer is.
func (w Window) offerWidths() []int { return w.widthsOf(w.Offers()) }

// widthsOf is how wide each of some offers is drawn.
func (w Window) widthsOf(offers []string) []int {
	var widths []int
	for _, offer := range offers {
		widths = append(widths, max(offerWide, ends(w.measured(), w.Said(offer), reads)+padding))
	}
	return widths
}

// savedTitle is the title of what a form is open on, as it was saved.
func (w Window) savedTitle() string {
	if w.form.ForObjective {
		g, _ := w.snap.Goal(w.form.Number)
		return g.Title
	}
	c, _ := w.snap.Card(w.form.Number)
	return c.Title
}

// Asks is what the question says, a sentence to a line.
func (w Window) Asks() []Said {
	if w.asksAboutSeveral() {
		return w.severalSays()
	}
	if w.question == aboutLeaving {
		return []Said{
			{Means: "It is set aside: this " + w.formKind() + "'s form brings it back."},
		}
	}
	if w.question == aboutBringingIn {
		return []Said{
			{Means: "It takes the place of the store, which is backed up first."},
			{Means: "Enter to agree, or Cancel."},
		}
	}
	if w.question == aboutPuttingBack {
		return []Said{
			{Means: "What the store holds now is backed up first."},
			{Means: "Enter to agree, or Cancel."},
		}
	}
	what := "task"
	if w.ChosenGoal() {
		what = "objective"
	}
	return []Said{
		{Means: "This " + what + " cannot be brought back."},
		{Means: "Delete again to agree, or Cancel."},
	}
}

// QuestionBox is the box a question is asked in: as wide as its sentences
// want and as tall as they are, not the panel a form is drawn in. Asked in
// the whole panel, two lines and two words sat in a box nine tenths empty.
func (w Window) QuestionBox() Box {
	return AskingOver(w.width, w.height, len(w.Asks()))
}

// Question is what the one question screen is asking about.
type Question int

// The two things worth asking before doing: both lose work, and work lost
// without being asked is work somebody has to remember they lost.
const (
	aboutDeleting Question = iota
	aboutLeaving
	aboutSweeping
	aboutDeletingSeveral
	aboutArchivingSeveral
	aboutPuttingBack
	aboutBringingIn
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
