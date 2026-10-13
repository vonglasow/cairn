package gui

import (
	"strconv"

	"cairn/internal/engine"
)

var spellings = map[string]string{
	"ctrl+n":       "cmd+n",
	"cmd+f":        "/",
	"ctrl+f":       "/",
	"return":       "enter",
	"shift+return": "shift+enter",
	" ":            "space",
}

// Key is the few flat shortcuts this board keeps.
func (w Window) Key(name string) Window {
	if said, ok := spellings[name]; ok {
		name = said
	}
	for _, offer := range w.Offers() {
		if Shortcut(offer) == name {
			return w.Choose(offer)
		}
	}
	if on, ok := screens[w.screen]; ok {
		return on(w, name)
	}
	return w
}

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
		if w.chosen != 0 && !w.chosenGoal {
			return w.working()
		}
	case "cmd+b":
		return w.backingUp()
	case "shift+b":
		return w.listingBackups()
	case "cmd+e":
		return w.exportingTasks()
	case "cmd+o":
		return w.importingTasks()
	case "shift+x":
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
		w.screen = OnTheBoard
	}
	return w
}

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
		w.pasting = true
	case "cmd+s":
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
		if long {
			w.form = w.form.Writes(w.form.Writing().Types("\n"))
			return w.inSightOfTheCaret()
		}
		w.form = w.form.onto(w.form.On() + 1)
	}
	return w
}

func (w Window) inSightOfTheCaret() Window {
	_, _, lines, rows, ok := w.sayingOn()
	if !ok {
		return w
	}
	line, _ := lines.At(w.form.Writing().At)
	w.saying = w.settledSaying(inSight(line, 1, rows, w.saying))
	return w
}

func (w Window) holding(by int) Window {
	writing := w.form.Writing()
	w.form = w.form.Writes(writing.Selects(writing.At + by))
	return w
}

// Copying is what the window wants put on the clipboard, and whether it wants anything at all.
func (w Window) Copying() (string, bool) { return w.copying, w.copying != "" }

// Copied says the loop has done it.
func (w Window) Copied() Window {
	w.copying = ""
	return w
}

// Pasting says the window has asked for what is on the clipboard.
func (w Window) Pasting() bool { return w.pasting }

// Pasted hands over what was on it.
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

func (w Window) walking(across, down int) Window {
	if w.tab == Objectives {
		return w.walkingGoals(across, down)
	}
	if lane, at, inBand := w.whereInBand(); inBand {
		return w.walkingBand(lane, at, across, down)
	}
	column, at := w.whereChosen()
	if down > 0 {
		if here := w.inColumn(column); column > 0 && at == len(here)-1 {
			if lane := w.inLane(column - 1); len(lane) > 0 {
				w.chosen, w.chosenGoal = lane[0].Number, false
				return w
			}
		}
	}
	if across != 0 {
		for next := column + across; next >= 0 && next < engine.Columns; next += across {
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

func (w Window) whereInBand() (lane, at int, ok bool) {
	for i := range engine.Stalled {
		for j, c := range w.inLane(i) {
			if c.Number == w.chosen {
				return i, j, true
			}
		}
	}
	return 0, 0, false
}

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
		for next := lane + across; next >= 0 && next < len(engine.Stalled); next += across {
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

func (w Window) followingLane(lane, at int) Window {
	run := w.Layout().Lane[lane]
	w.lanes[lane] = w.settledLane(lane, inSight(at, CardHeight, run.H, w.lanes[lane]))
	return w
}

func (w Window) following(column, at int) Window {
	run := w.Layout().Column[column]
	w.scroll[column] = w.settled(column, inSight(at, CardHeight, run.H, w.scroll[column]))
	return w
}

func (w Window) followingGoal(at int) Window {
	run := GoalsRun(w.Layout())
	w.goals = w.settledGoals(inSight(at/GoalsAcross(), CardHeight, run.H, w.goals))
	return w
}

func (w Window) settledGoals(want int) int {
	run := GoalsRun(w.Layout())
	return min(max(0, GoalRows(len(w.Showing().Goals))*CardHeight-run.H), max(0, want))
}

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
	at = min(len(here)-1, max(0, at+across+down*GoalsAcross()))
	w.chosen, w.chosenGoal = here[at].Number, true
	return w.followingGoal(at)
}

// WhereChosen is where what is chosen is drawn and the run of the window it is drawn in, so that a test can ask whether one is inside the other.
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
	for lane := range engine.Stalled {
		for j, c := range w.inLane(lane) {
			if c.Number == w.chosen {
				return l.LaneCardAt(lane, j, w.lanes[lane]), l.Lane[lane], true
			}
		}
	}
	return Box{}, Box{}, false
}

func (w Window) asking(op engine.Operation, in engine.Input) Window {
	if _, err := w.work.Ask(op, in); err != nil {
		return w.failing(err)
	}
	return w.Look()
}

func (w Window) whereChosen() (column, at int) {
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

func (w Window) opening() Window { return w.Edit() }
