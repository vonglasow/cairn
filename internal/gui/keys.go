package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"strconv"
	"strings"
	"time"

	"cairn/internal/engine"
)

var goingTo = map[string]board.Tab{"t": board.Tasks, "o": board.Objectives, "i": board.Time}

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
	if w.Stranded() {
		if name == "cmd+," {
			return w.openingConfiguration()
		}
		return w
	}
	if w.Screen == board.OnPalette {
		return w.onPalette(name)
	}
	if w.SeriesChord {
		w.SeriesChord = false
		return w.chording(name)
	}
	if w.Screen == board.OnTheBoard && w.going {
		w.going = false
		if page, ok := goingTo[name]; ok && w.Now.Sub(w.goneAt) <= time.Second {
			return w.onPage(page)
		}
	}
	for _, offer := range w.Offers() {
		if w.shortcut(offer) == name {
			return w.Choose(offer)
		}
	}
	if on, ok := screens[w.Screen]; ok {
		return on(w, name)
	}
	return w
}

var screens = map[board.Screen]func(Window, string) Window{
	board.OnTheBoard:      Window.onBoard,
	board.OnForm:          func(w Window, name string) Window { return w.inForm(name).inSightOfTheCaret() },
	board.OnSearch:        Window.onSearch,
	board.OnChoosing:      Window.inChoosing,
	board.OnCalendar:      Window.onCalendar,
	board.OnMenu:          Window.onMenu,
	board.OnBackups:       Window.onBackups,
	board.OnLabels:        Window.onLabels,
	board.OnFiltering:     Window.onFilters,
	board.OnConfiguration: Window.onConfiguration,
	board.OnWorkspaces:    Window.onWorkspaces,
	board.OnFates:         Window.onFates,
	board.OnSeriesFates:   Window.onSeriesFates,
	board.OnRule:          Window.onRule,
	board.OnKeeping:       Window.onKeeping,
	board.OnViews:         Window.onViews,
}

func (w Window) onBoard(name string) Window {
	if w.Tab == board.Time {
		mine, shared := weekKeys(name)
		if mine {
			return w.onWeek(name)
		}
		if !shared {
			return w
		}
	}
	if w.Tab == board.Series {
		if done, took := w.seriesKey(name); took {
			return done
		}
	}
	if name == "esc" && w.Chosen != 0 {
		w.Chosen, w.ChosenGoal = 0, false
		return w
	}
	if w.Tab == board.Objectives && name == "t" {
		return w.detailing(ShowItsTasks)
	}
	if g, ok := w.Detailed(); w.Tab == board.Objectives && name == "shift+n" && ok && g.Parent == 0 {
		return w.detailing(AddSubSays)
	}
	if w.Tab == board.Objectives {
		if folded, took := w.foldingKey(name); took {
			return folded
		}
	}
	if w.Tab == board.Tasks && (name == "[" || name == "]") {
		by := 1
		if name == "[" {
			by = -1
		}
		return w.walkingNarrowing(by)
	}
	if card, ok := w.Snap.Card(w.Chosen); ok && w.Chosen != 0 && !w.ChosenGoal && w.Tab == board.Tasks {
		switch name {
		case "m":
			return w.settingUnderTheCard("state", engine.States(), card.State)
		case "p":
			return w.settingUnderTheCard("priority", board.Priorities(), card.Priority)
		case "l":
			return w.labellingChosen()
		case "t":
			return w.datingChosen()
		case "r":
			_, w.SeriesChord = w.OccurrenceChosen()
			return w
		}
	}
	if figure, ok := strings.CutPrefix(name, "cmd+shift+"); ok && aFigure(figure) {
		return w.choosingByPlace(name)
	}
	if opened, ok := w.openingByFigure(name); ok {
		return opened
	}
	switch name {
	case "g":
		w.going, w.goneAt = true, w.Now
		return w
	case "f":
		if w.Chosen != 0 && !w.ChosenGoal {
			return w.working()
		}
	case "o":
		if w.Chosen != 0 && !w.ChosenGoal {
			return w.startingLinked()
		}
	case "cmd+b":
		return w.backingUp()
	case "shift+b":
		return w.listingBackups()
	case "shift+l":
		return w.listingLabels()
	case "cmd+k":
		return w.openingPalette()
	case "cmd+,":
		return w.openingConfiguration()
	case "cmd+shift+o":
		return w.openingSwitcher()
	case "shift+f":
		return w.filteringBoard()
	case "backspace":
		return w.takingOffLast()
	case "alt+backspace":
		w.Filters, w.Looking = w.Filters.Cleared(), board.Written{}
		return w
	case "cmd+e":
		return w.exportingTasks()
	case "cmd+o":
		return w.importingTasks()
	case "shift+x":
		if w.Tab == board.Tasks && w.Sweeps() != "" {
			w.Screen, w.Question = board.OnAsking, board.AboutSweeping
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
		w.Looking = edit(w.Looking)
		return w
	}
	switch name {
	case "backspace":
		if w.Looking.Empty() {
			return w.leavingSearch()
		}
		w.Looking = w.Looking.Back()
	case "left":
		return w.walks(-1)
	case "right":
		return w.walks(1)
	case "enter":
		w.Screen = board.OnTheBoard
	}
	return w
}

func (w Window) inForm(name string) Window {
	if w.Form.OnWork {
		walked, _ := w.onWork(name)
		return walked
	}
	if name == "tab" && w.ReachingWork() {
		w.Form.OnWork, w.Form.WorkAt = true, 0
		return w
	}
	if w.Form.OnParts() {
		if done, took := w.onParts(name); took {
			return done.keepingPartInSight()
		}
	}
	long := w.Form.Fields[w.Form.On()].Long
	if edit, ok := textKey(name); ok && !w.OnAClosedSet() {
		w.Form = w.Form.Writes(edit(w.Form.Writing()))
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
		w.Form = w.Form.Onto(w.Form.On() + 1)
	case "shift+tab":
		w.Form = w.Form.Onto(w.Form.On() - 1)
	case "down", "up":
		by := 1
		if name == "up" {
			by = -1
		}
		if long {
			return w.walksLines(by)
		}
		w.Form = w.Form.Onto(w.Form.On() + by)
	case "backspace":
		return w.back()
	case "cmd+a":
		if w.picked.Key != "" {
			return w.all()
		}
		w.Form = w.Form.Writes(w.Form.Writing().All())
	case "cmd+c":
		if w.picked.Key != "" {
			return w.copyingPicked()
		}
		w.copying = w.Form.Writing().Holds()
	case "cmd+x":
		w.copying = w.Form.Writing().Holds()
		if !w.OnAClosedSet() {
			w.Form = w.Form.Writes(w.Form.Writing().Back())
		}
	case "cmd+v":
		w.pasting = true
	case "cmd+s":
		return w.Save()
	case "cmd+t":
		return w.togglingTrace()
	case "shift+left":
		return w.holding(-1)
	case "shift+right":
		return w.holding(1)
	case "left":
		return w.walks(-1)
	case "right":
		return w.walks(1)
	case "enter", "shift+enter":
		if on := w.Form.Fields[w.Form.On()]; chosenFromAList(on) || chosenOnACalendar(on) {
			return w.Choose("Choose")
		}
		if long {
			w.Form = w.Form.Writes(w.Form.Writing().Types("\n"))
			return w.inSightOfTheCaret()
		}
		w.Form = w.Form.Onto(w.Form.On() + 1)
	}
	return w
}

func (w Window) inSightOfTheCaret() Window {
	_, _, lines, rows, ok := w.sayingOn()
	if !ok {
		return w
	}
	line, _ := lines.At(w.Form.Writing().At)
	w.saying = w.settledSaying(board.InSight(line, 1, rows, w.saying))
	return w
}

func (w Window) holding(by int) Window {
	writing := w.Form.Writing()
	w.Form = w.Form.Writes(writing.Selects(writing.At + by))
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
	if w.Screen == board.OnChoosing {
		if !w.ChoosingSearched() {
			return w
		}
		w.Filtering, w.ChoosingOn, w.ChoosingFrom = w.Filtering.Types(text), 0, 0
		return w
	}
	if w.Screen == board.OnConfiguration {
		return w.Types(text)
	}
	if w.Screen != board.OnForm {
		return w
	}
	if w.Form.OnParts() || w.OnAClosedSet() {
		return w.Types(text)
	}
	w.Form = w.Form.Writes(w.Form.Writing().Types(text))
	return w
}

func (w Window) walking(across, down int) Window {
	if w.Tab == board.Objectives {
		return w.walkingGoals(across, down)
	}
	c, spot, moved := board.Walk(w.Showing().Cards, w.Archived, w.Chosen, across, down)
	if !moved {
		return w
	}
	w.Chosen, w.ChosenGoal = c.Number, false
	if spot.InBand {
		return w.followingLane(spot.Column, spot.At)
	}
	return w.following(spot.Column, spot.At)
}

func (w Window) shifting(across, down int) Window {
	if w.Tab == board.Objectives || w.Chosen == 0 {
		return w
	}
	showing := w.Showing().Cards
	column, inBand, _, found := board.Where(showing, w.Archived, w.Chosen)
	if !found {
		return w
	}
	if across != 0 {
		state, ok := board.Beside(column, inBand, across)
		if !ok {
			return w
		}
		return w.asking(engine.TaskEdit, engine.Input{Given: map[string]string{
			"number": strconv.Itoa(w.Chosen), "state": state}})
	}
	if inBand {
		return w
	}
	given, ok := board.Step(w.InColumn(column), w.Chosen, down)
	if !ok {
		return w
	}
	given["number"] = strconv.Itoa(w.Chosen)
	return w.asking(engine.TaskMove, engine.Input{Given: given})
}

func (w Window) followingLane(lane, at int) Window {
	run := w.Layout().Lane[lane]
	w.lanes[lane] = w.settledLane(lane, board.InSight(at, geom.CardHeight, run.H, w.lanes[lane]))
	return w
}

func (w Window) following(column, at int) Window {
	run := w.Layout().Column[column]
	w.scroll[column] = w.settled(column, board.InSight(at, geom.CardHeight, run.H, w.scroll[column]))
	return w
}

func (w Window) followingGoal(at int) Window {
	run := GoalsRun(w.Layout())
	w.goals = w.settledGoals(board.InSight(at, ItemHigh, run.H, w.goals))
	return w
}

func (w Window) settledGoals(want int) int {
	run := GoalsRun(w.Layout())
	return min(max(0, len(w.Tree())*ItemHigh-run.H), max(0, want))
}

func (w Window) walkingGoals(across, down int) Window {
	here := w.Tree()
	if len(here) == 0 {
		return w
	}
	at := -1
	for i, b := range here {
		if w.ChosenGoal && b.Goal.Number == w.Chosen {
			at = i
		}
	}
	if at < 0 {
		at = 0
	} else {
		at = min(len(here)-1, max(0, at+across+down))
	}
	return w.choosingGoal(here[at].Goal.Number)
}

func (w Window) asking(op engine.Operation, in engine.Input) Window {
	answer, err := w.work.Ask(op, in)
	if err != nil {
		if agreeing, ok := engine.AgreeingTo(op, in, err); ok {
			return w.agreeingTo(agreeing).Look()
		}
		return w.failing(err)
	}
	return w.warned(answer).Look()
}

func (w Window) warned(answer engine.Answer) Window {
	if len(answer.Warnings) == 0 {
		return w
	}
	return w.telling(strings.Join(answer.Warnings, "; "))
}

func (w Window) agreed() Window {
	agreeing := w.agreeing
	w.Screen, w.Question = board.OnTheBoard, board.AboutDeleting
	w.agreeing = engine.Agreeing{}
	w = w.asking(agreeing.Op, agreeing.In)
	if failed := w.Err; failed != nil {
		w = w.Look()
		w.Err = failed
	}
	return w
}

func (w Window) opening() Window { return w.Edit() }
