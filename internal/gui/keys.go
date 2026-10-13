package gui

import (
	"strconv"

	"cairn/internal/engine"
)

// What the keyboard does: the keys the bar offers and the few it does not,
// walking the board and moving what is chosen, and holding, copying and
// pasting text.

// spellings are the keys that arrive under more than one name, each read as
// the one the bar's shortcuts are written in — said once, before anything
// looks at the key, so no screen has to know there were two.
var spellings = map[string]string{
	"ctrl+n":       "cmd+n",
	"cmd+f":        "/",
	"ctrl+f":       "/",
	"return":       "enter",
	"shift+return": "shift+enter",
	" ":            "space",
}

// Key is the few flat shortcuts this board keeps. There are no modes: what a
// person points at is what happens, and these are the things a pointer is bad
// at saying.
func (w Window) Key(name string) Window {
	if said, ok := spellings[name]; ok {
		name = said
	}
	// A key does what the bar offers, where the bar offers it, and nothing
	// where the bar says nothing. One rule and not a list of guards: the bar
	// used to offer Edit on an open task while e did nothing there, and a key
	// that adds used to work over an open form and throw away what was in it.
	//
	// Read off the offers rather than written out beside them, so an offer
	// added to one screen cannot arrive without its key.
	for _, offer := range w.Offers() {
		if Shortcut(offer) == name {
			return w.Choose(offer)
		}
	}
	// Escape is not here: every screen that has a way out offers it as Close
	// or Cancel, and Shortcut maps both to esc — so it is served by the loop
	// above before it ever reaches a screen.
	if on, ok := screens[w.screen]; ok {
		return on(w, name)
	}
	return w
}

// screens is what a key the bar does not offer does, by the screen it lands
// on. One table rather than a guard per key on which screen it is: each
// screen's keys are read together, in one place.
var screens = map[Screen]func(Window, string) Window{
	OnTheBoard: Window.onBoard,
	OnForm:     func(w Window, name string) Window { return w.inForm(name).inSightOfTheCaret() },
	OnSearch:   Window.onSearch,
	OnChoosing: Window.inChoosing,
	OnHelp:     Window.inHelp,
	OnCalendar: Window.onCalendar,
	OnMenu:     Window.onMenu,
	OnBackups:  Window.onBackups,
}

// onBoard is what a key does on the board itself.
func (w Window) onBoard(name string) Window {
	if w.tab == Time {
		mine, shared := weekKeys(name)
		if mine {
			return w.onWeek(name)
		}
		if !shared {
			return w
		}
	}
	switch name {
	case "f":
		// Taking a task up is not offered on the bar: the bar has room for
		// what changes a task, and this changes what somebody is working on.
		// It is in the help, under the same key the terminal board uses.
		if w.chosen != 0 && !w.chosenGoal {
			return w.working()
		}
	case "cmd+b":
		// Not on the bar, which has room for what changes a task: this keeps
		// the whole store, and is in the help, under the key the system's own
		// save-a-copy chords are near.
		return w.backingUp()
	case "shift+b":
		return w.listingBackups()
	case "cmd+e":
		return w.exportingTasks()
	case "cmd+o":
		return w.importingTasks()
	case "shift+x":
		// The whole column off the board at once. Not on the bar either: the
		// bar has room for what one card does, and this does what a column
		// does. The engine has swept by a narrowing all along and only the
		// command line ever asked it to.
		if w.tab == Tasks && w.sweeps() != "" {
			w.screen, w.question = OnAsking, aboutSweeping
		}
	case "tab":
		return w.other(1)
	case "shift+tab":
		return w.other(-1)
	case "left":
		return w.walking(-1, 0)
	case "right":
		return w.walking(1, 0)
	case "up":
		return w.walking(0, -1)
	case "down":
		return w.walking(0, 1)
	case "enter":
		return w.opening()
	case "shift+left":
		return w.shifting(-1, 0)
	case "shift+right":
		return w.shifting(1, 0)
	case "shift+up":
		return w.shifting(0, -1)
	case "shift+down":
		return w.shifting(0, 1)
	}
	return w
}

// onSearch is what a key does while the search box is open.
func (w Window) onSearch(name string) Window {
	if edit, ok := textKey(name); ok {
		w.looking = edit(w.looking)
		return w
	}
	switch name {
	case "backspace":
		w.looking = w.looking.Back()
	case "left":
		return w.walks(-1)
	case "right":
		return w.walks(1)
	case "enter":
		// The box closes and what it narrowed stays narrowed. A search that
		// vanished the moment it was asked for would be a search nobody
		// could act on.
		w.screen = OnTheBoard
	}
	return w
}

// inForm is what a key does while the form has the window. Everything here is
// about one field, because a form is a column of fields and the keyboard is on
// one of them.
func (w Window) inForm(name string) Window {
	if w.form.onParts() {
		if done, took := w.onParts(name); took {
			return done.keepingPartInSight()
		}
	}
	long := w.form.Fields()[w.form.On()].Long
	if edit, ok := textKey(name); ok && !w.onAClosedSet() {
		w.form = w.form.Writes(edit(w.form.Writing()))
		return w.inSightOfTheCaret()
	}
	if long && (name == "shift+up" || name == "shift+down") {
		by := 1
		if name == "shift+up" {
			by = -1
		}
		return w.holdingLines(by)
	}
	switch name {
	case "tab":
		w.form = w.form.onto(w.form.On() + 1)
	case "shift+tab":
		w.form = w.form.onto(w.form.On() - 1)
	case "down", "up":
		// Inside the description the arrows walk its lines, so every one of
		// them can be reached. Down used to leave for the next field, and the
		// only way through a long text was a character at a time — which is
		// why it read as cut off. Tab is what leaves it.
		by := 1
		if name == "up" {
			by = -1
		}
		if long {
			return w.walksLines(by)
		}
		w.form = w.form.onto(w.form.On() + by)
	case "backspace":
		return w.back()
	case "cmd+a":
		// What was last pressed in: a field, or what the column beside the
		// form writes, which holds no caret and is held all the same.
		if w.picked.Key != "" {
			return w.all()
		}
		w.form = w.form.Writes(w.form.Writing().All())
	case "cmd+c":
		if w.picked.Key != "" {
			return w.copyingPicked()
		}
		w.copying = w.form.Writing().Holds()
	case "cmd+x":
		w.copying = w.form.Writing().Holds()
		if !w.onAClosedSet() {
			w.form = w.form.Writes(w.form.Writing().Back())
		}
	case "cmd+v":
		// Asked for here and answered later: a clipboard is read by the
		// system, which takes as long as it takes, and the answer comes back
		// as an event like any other.
		w.pasting = true
	case "cmd+s":
		// The terminal's save, which the window takes too: a terminal cannot
		// be sent ⌘, so the board saves on ^S, and the same fingers should
		// work in both.
		return w.Save()
	case "shift+left":
		return w.holding(-1)
	case "shift+right":
		return w.holding(1)
	case "left":
		return w.walks(-1)
	case "right":
		return w.walks(1)
	case "enter", "shift+enter":
		// A line where a line can be written, and on to the next field
		// everywhere else — the same on both boards. It never saves: saving
		// is a chord of its own, so it cannot happen by reaching for the key
		// that breaks a line.
		if long {
			w.form = w.form.Writes(w.form.Writing().Types("\n"))
			return w.inSightOfTheCaret()
		}
		w.form = w.form.onto(w.form.On() + 1)
	}
	return w
}

// inSightOfTheCaret scrolls the long field the keyboard is on just far enough
// for the caret's line to be in the box. The wheel moves the box on its own; the
// caret moving brings it back.
func (w Window) inSightOfTheCaret() Window {
	_, _, lines, rows, ok := w.sayingOn()
	if !ok {
		return w
	}
	line, _ := lines.At(w.form.Writing().At)
	w.saying = w.settledSaying(inSight(line, 1, rows, w.saying))
	return w
}

// holding walks the caret and leaves the anchor, which is how a run is held
// from the keyboard.
func (w Window) holding(by int) Window {
	writing := w.form.Writing()
	w.form = w.form.Writes(writing.Selects(writing.At + by))
	return w
}

// Copying is what the window wants put on the clipboard, and whether it wants
// anything at all.
func (w Window) Copying() (string, bool) { return w.copying, w.copying != "" }

// Copied says the loop has done it.
func (w Window) Copied() Window {
	w.copying = ""
	return w
}

// Pasting says the window has asked for what is on the clipboard.
func (w Window) Pasting() bool { return w.pasting }

// Pasted hands over what was on it. Empty is still an answer: the clipboard
// held nothing, and the window has stopped waiting either way.
func (w Window) Pasted(text string) Window {
	w.pasting = false
	if text == "" {
		return w
	}
	if w.screen == OnChoosing {
		w.filtering, w.choosing, w.choosingFrom = w.filtering.Types(text), 0, 0
		return w
	}
	if w.screen != OnForm {
		return w
	}
	if w.form.onParts() || w.onAClosedSet() {
		return w.Types(text)
	}
	w.form = w.form.Writes(w.form.Writing().Types(text))
	return w
}

// walking moves what is chosen by the arrows: across the columns, or down the
// cards of one. A hand on the keyboard should not have to reach for the mouse
// to look at the next task.
//
// It stops at the edges rather than wrapping round, and walks past a column
// with nothing in it rather than stopping in one — an empty column is not a
// place to be.
func (w Window) walking(across, down int) Window {
	if w.tab == Objectives {
		return w.walkingGoals(across, down)
	}
	// The band is walked like the columns. A task that is blocked, watching or
	// paused used to be beyond the arrows altogether — and those are the ones
	// that want a decision.
	if lane, at, inBand := w.whereInBand(); inBand {
		return w.walkingBand(lane, at, across, down)
	}
	column, at := w.whereChosen()
	if down > 0 {
		// Off the bottom of a column is into the lane beneath it. The backlog
		// stands alone, full height, with no lane under it.
		if here := w.inColumn(column); column > 0 && at == len(here)-1 {
			if lane := w.inLane(column - 1); len(lane) > 0 {
				w.chosen, w.chosenGoal = lane[0].Number, false
				return w
			}
		}
	}
	if across != 0 {
		for next := column + across; next >= 0 && next < Columns; next += across {
			here := w.inColumn(next)
			if len(here) == 0 {
				continue
			}
			at = min(at, len(here)-1)
			w.chosen, w.chosenGoal = here[at].Number, false
			return w.following(next, at)
		}
		return w
	}
	here := w.inColumn(column)
	if len(here) == 0 {
		return w
	}
	at = min(len(here)-1, max(0, at+down))
	w.chosen, w.chosenGoal = here[at].Number, false
	return w.following(column, at)
}

// shifting moves the chosen card itself rather than the choice: sideways it
// changes column, up and down it changes place among the ones it sits with.
// The terminal has done this from the beginning; the window had only the drag,
// so the central gesture of the product wanted a mouse.
func (w Window) shifting(across, down int) Window {
	if w.tab == Objectives || w.chosen == 0 {
		return w
	}
	showing := w.Showing().Cards
	column, inBand, _, found := engine.Where(showing, w.archived, w.chosen)
	if !found {
		return w
	}
	if across != 0 {
		state, ok := engine.Beside(column, inBand, across)
		if !ok {
			return w
		}
		return w.asking(engine.TaskEdit, engine.Input{Given: map[string]string{
			"number": strconv.Itoa(w.chosen), "state": state}})
	}
	if inBand {
		return w
	}
	given, ok := engine.Step(w.inColumn(column), w.chosen, down)
	if !ok {
		return w
	}
	given["number"] = strconv.Itoa(w.chosen)
	return w.asking(engine.TaskMove, engine.Input{Given: given})
}

// whereInBand says which lane the chosen task sits in, and where in it.
func (w Window) whereInBand() (lane, at int, ok bool) {
	for i := range Lanes {
		for j, c := range w.inLane(i) {
			if c.Number == w.chosen {
				return i, j, true
			}
		}
	}
	return 0, 0, false
}

// walkingBand moves inside the band, and back up into the column a lane sits
// under. The lanes are one row, so down walks along a lane and up leaves it.
func (w Window) walkingBand(lane, at, across, down int) Window {
	if down < 0 && at == 0 {
		if here := w.inColumn(lane + 1); len(here) > 0 {
			last := len(here) - 1
			w.chosen, w.chosenGoal = here[last].Number, false
			return w.following(lane+1, last)
		}
		return w
	}
	if across != 0 {
		for next := lane + across; next >= 0 && next < len(Lanes); next += across {
			if here := w.inLane(next); len(here) > 0 {
				landed := min(at, len(here)-1)
				w.chosen, w.chosenGoal = here[landed].Number, false
				return w.followingLane(next, landed)
			}
		}
		return w
	}
	here := w.inLane(lane)
	landed := min(len(here)-1, max(0, at+down))
	w.chosen, w.chosenGoal = here[landed].Number, false
	return w.followingLane(lane, landed)
}

// followingLane is following for a lane of the band.
func (w Window) followingLane(lane, at int) Window {
	run := w.Layout().Lane[lane]
	w.lanes[lane] = w.settledLane(lane, inSight(at, CardHeight, run.H, w.lanes[lane]))
	return w
}

// following moves a column so that the card just chosen is in sight. Walking
// used to move the choice and leave the column where it was, which chose a
// card nobody could see — and a card chosen out of sight is a card whose
// title, whose state and whose place are all unknown to whoever chose it.
func (w Window) following(column, at int) Window {
	run := w.Layout().Column[column]
	w.scroll[column] = w.settled(column, inSight(at, CardHeight, run.H, w.scroll[column]))
	return w
}

// followingGoal is the same for the objectives, which are one list rather than
// four columns and so have one number of their own.
func (w Window) followingGoal(at int) Window {
	run := GoalsRun(w.Layout())
	w.goals = w.settledGoals(inSight(at/GoalsAcross(), CardHeight, run.H, w.goals))
	return w
}

// settledGoals keeps the objectives between their two ends.
func (w Window) settledGoals(want int) int {
	run := GoalsRun(w.Layout())
	return min(max(0, GoalRows(len(w.Showing().Goals))*CardHeight-run.H), max(0, want))
}

// walkingGoals is the same down a list, which is what the objectives are.
func (w Window) walkingGoals(across, down int) Window {
	here := engine.Sorted(w.Showing().Goals)
	if len(here) == 0 {
		return w
	}
	at := 0
	for i, g := range here {
		if g.Number == w.chosen {
			at = i
		}
	}
	// Across moves to the next objective, down to the one a row below: the
	// objectives are a grid, and an arrow means in a grid what it looks like
	// it means.
	at = min(len(here)-1, max(0, at+across+down*GoalsAcross()))
	w.chosen, w.chosenGoal = here[at].Number, true
	return w.followingGoal(at)
}

// WhereChosen is where what is chosen is drawn and the run of the window it is
// drawn in, so that a test can ask whether one is inside the other.
func (w Window) WhereChosen() (box, run Box, ok bool) {
	l := w.Layout()
	if w.tab == Objectives {
		for i, g := range engine.Sorted(w.Showing().Goals) {
			if g.Number == w.chosen {
				return GoalAt(l, i, w.goals), GoalsRun(l), true
			}
		}
		return Box{}, Box{}, false
	}
	dealt, _ := engine.Deal(w.Showing().Cards, w.archived)
	for i, held := range dealt {
		for j, c := range held.Cards {
			if c.Number == w.chosen {
				return l.CardAt(i, j, w.scroll[i]), l.Column[i], true
			}
		}
	}
	for lane := range Lanes {
		for j, c := range w.inLane(lane) {
			if c.Number == w.chosen {
				return l.LaneCardAt(lane, j, w.lanes[lane]), l.Lane[lane], true
			}
		}
	}
	return Box{}, Box{}, false
}

// asking runs one operation and reads the board back, the way every gesture
// that writes ends; a refusal is kept to be said.
func (w Window) asking(op engine.Operation, in engine.Input) Window {
	if _, err := w.work.Ask(op, in); err != nil {
		return w.failing(err)
	}
	return w.Look()
}

// whereChosen is which column what is chosen stands in and how far down it, and
// the first card of the first column that has one when nothing is chosen — an
// arrow on an untouched board has to begin somewhere.
func (w Window) whereChosen() (column, at int) {
	// What is walked is what is showing: under a search, the arrows reach
	// what the search left and nothing behind it.
	dealt, _ := engine.Deal(w.Showing().Cards, w.archived)
	for i, held := range dealt {
		for j, c := range held.Cards {
			if c.Number == w.chosen {
				return i, j
			}
		}
	}
	for i, held := range dealt {
		if len(held.Cards) > 0 {
			return i, 0
		}
	}
	return 0, 0
}

// opening is what enter does: it opens whatever the arrows walked to, the way a
// double-click opens whatever the pointer landed on — to be edited: reading a
// task and changing it are one screen, because a task is opened to be worked
// on and a second gesture to start writing in it was a gesture spent for
// nothing.
func (w Window) opening() Window { return w.Edit() }
