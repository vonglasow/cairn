package gui

import (
	"strconv"
	"time"

	"cairn/internal/engine"
)

// Click chooses whatever is under the pointer, and nothing when there is nothing there.
func (w Window) Click(p Point) Window {
	if w.screen == OnTheBoard && !w.native && MenuButton(w.Layout()).In(p) {
		return w.openingMenuBar()
	}
	if w.screen == OnMenu {
		return w.clickingMenu(p)
	}
	if w.screen == OnBackups {
		if offer, on := w.offerAt(p); on {
			return w.Choose(offer)
		}
		return w.clickingBackups(p)
	}
	if offer, on := w.offerAt(p); on {
		return w.Choose(offer)
	}
	if w.screen == OnChoosing {
		return w.clickingList(p)
	}
	if w.screen == OnCalendar && !w.CalendarBox().In(p) {
		return w.clickedOff(p)
	}
	if w.screen == OnCalendar {
		if walked, on := w.monthClicked(p); on {
			return walked
		}
		w, _ = w.dayClicked(p)
		return w
	}
	if at, on := w.fieldAt(p); on {
		reached := w.reaching(at, p)
		w.form = w.form.onto(at)
		if chosenFromAList(w.form.Fields()[at]) {
			return w.choosingList()
		}
		if chosenOnACalendar(w.form.Fields()[at]) {
			return w.openingCalendar()
		}
		return w.CaretInto(reached)
	}
	if w.closesOnAClickOff() && !w.panel().In(p) {
		return w.clickedOff(p)
	}
	if w.screen == OnSearch {
		if Searching(w.Layout()).In(p) {
			return w
		}
	} else if w.screen != OnTheBoard {
		return w
	}
	if tab, on := w.tabAt(p); on {
		return w.onPage(tab)
	}
	if w.tab == Time {
		return w.clickingTime(p)
	}
	return w.choosingAt(w.at(p))
}

// Hovering answers the window the pointer resting at a point leaves behind.
func (w Window) Hovering(p Point) Window {
	w.over, w.hovered = w.offerAt(p)
	return w
}

// Away is the window the pointer having left it leaves behind.
func (w Window) Away() Window {
	w.over, w.hovered = "", false
	return w
}

// Lighting says whether an offer is the one the pointer is resting on, which is the one thing in the window that says a word can be clicked.
func (w Window) Lighting(offer string) bool { return w.hovered && w.over == offer }

func (w Window) goalAt(p Point) int {
	l := w.Layout()
	here := engine.Sorted(w.Showing().Goals)
	for i := range here {
		if GoalAt(l, i, w.goals).In(p) {
			return here[i].Number
		}
	}
	return 0
}

func (w Window) tabAt(p Point) (Tab, bool) {
	l := w.Layout()
	for tab := range Pages {
		if TabAt(l, tab).In(p) {
			return Tab(tab), true
		}
	}
	return w.tab, false
}

// DoubleClick opens the detail of the card under the pointer, which is where the description and the history are.
func (w Window) DoubleClick(p Point) Window {
	if w.screen == OnChoosing {
		box := w.ChoosingBox()
		if at, on := ChoosingAtPoint(box, len(w.Choosing()), w.choosingFrom, p); on {
			w.choosing = at
			w = w.taking()
			if w.screen == OnChoosing {
				return w.keeping()
			}
		}
		return w
	}
	if at, on := w.partUnder(p); on && w.screen == OnForm {
		field, _ := w.form.partsField()
		w.form = w.form.onto(field)
		w.form.partAt = at
		w.form = w.form.naming()
		return w
	}
	if w.screen == OnForm {
		if at, on := w.fieldAt(p); on {
			return w.holdingBy(at, p, Words)
		}
		return w.word(p)
	}
	if w.screen != OnTheBoard && w.screen != OnSearch {
		return w
	}
	number := w.at(p)
	if number == 0 {
		return w
	}
	w.chosen, w.chosenGoal = number, w.tab == Objectives
	return w.opening()
}

// TripleClick holds the paragraph under the pointer, on an open task and in a field of the form: the third press of the gesture every text answers.
func (w Window) TripleClick(p Point) Window {
	if at, on := w.fieldAt(p); on {
		return w.holdingBy(at, p, Paragraphs)
	}
	if w.screen == OnForm {
		return w.paragraph(p)
	}
	return w
}

func (w Window) holdingBy(at int, p Point, by Unit) Window {
	reached := w.reaching(at, p)
	w.form = w.form.onto(at)
	from, to := runOf(w.form.Writing().Reads(), by, reached)
	held := w.form.Writing()
	held.At, held.Held = to, from-to
	w.form = w.form.Writes(held)
	w.holdingText, w.growing, w.anchor = true, by, Picked{From: from, To: to}
	return w
}

// PressAt takes hold of whatever is under the pointer, without deciding yet whether this is a click or a drag.
func (w Window) PressAt(p Point) Window {
	w.pressed, w.pressedOn = true, w.screen
	if w.screen == OnChoosing {
		w.draggingChoosing = ChoosingBar(w.ChoosingBox()).In(p)
		return w
	}
	if w.screen == OnForm {
		if _, bar, _, _, long := w.sayingOn(); long && bar.In(p) {
			w.draggingSaying = true
			return w
		}
		if at, on := w.partUnder(p); on {
			field, _ := w.form.partsField()
			w = w.keepingPartInSight()
			w.form = w.form.onto(field)
			w.form.partAt, w.form.carried = at, at+1
			w.held = Held{From: p, At: p}
			return w
		}
		if at, on := w.fieldAt(p); on {
			w.picked = Picked{}
			return w.holdingBy(at, p, Characters).inSightOfTheCaret()
		}
		if !w.form.Adding && BesideBar(w.formPanel()).In(p) {
			w.draggingBar, w.draggingServing = !w.form.ForObjective, w.form.ForObjective
			return w
		}
		return w.picking(p)
	}
	if w.screen != OnTheBoard {
		return w
	}
	if w.tab == Objectives {
		return w
	}
	w.held = Held{Number: w.at(p), From: p, At: p}
	return w
}

// MoveTo follows the pointer, and past Grip says the card is being dragged.
func (w Window) MoveTo(p Point) Window {
	if w.draggingBar {
		track := BesideBar(w.formPanel())
		w.history = w.settledHistory(Reached(track, w.Told(), w.Rows(), p))
		return w
	}
	if w.holdingText && w.screen == OnForm {
		held := w.form.Writing()
		fixed, moving := Stretched(held.Reads(), w.growing, w.anchor, w.reaching(w.form.On(), p))
		held.At, held.Held = moving, fixed-moving
		w.form = w.form.Writes(held)
		return w.inSightOfTheCaret()
	}
	if w.draggingSaying {
		_, track, lines, rows, _ := w.sayingOn()
		w.saying = w.settledSaying(Reached(track, len(lines.Said), rows, p))
		return w
	}
	if w.pickingText {
		return w.draggingText(p)
	}
	if w.draggingChoosing {
		box := w.ChoosingBox()
		w.choosingFrom = w.settledChoosing(
			Reached(ChoosingBar(box), len(w.Choosing()), ChoosingRows(box), p))
		return w
	}
	if w.draggingServing {
		track := BesideBar(w.formPanel())
		room := max(1, BesideBox(w.formPanel()).H/SubHeight)
		w.serving = w.settledServing(Reached(track, len(w.Under()), room, p) * SubHeight)
		return w
	}
	if w.held.Number == 0 && w.form.carried == 0 {
		return w
	}
	w.held.At = p
	if abs(p.X-w.held.From.X) > Grip || abs(p.Y-w.held.From.Y) > Grip {
		w.held.Moved = true
	}
	return w
}

// LetGo drops the card, or lets a click be a click.
func (w Window) LetGo(p Point) Window {
	if w.pickingText {
		w.pickingText = false
		return w
	}
	if w.draggingBar || w.draggingSaying || w.draggingServing || w.draggingChoosing {
		w.draggingBar, w.draggingSaying = false, false
		w.draggingServing, w.draggingChoosing = false, false
		return w
	}
	if w.holdingText {
		w.holdingText = false
		if w.form.Writing().Holding() {
			return w
		}
	}
	waiting, was := w.pressed, w.pressedOn
	w.pressed = false
	if !waiting || was != w.screen {
		w.held = Held{}
		return w
	}
	if w.form.carried != 0 {
		return w.lettingGoOfAPart(p)
	}
	held := w.held
	w.held = Held{}
	if held.Number != 0 && held.Moved {
		return w.dropped(held.Number, p)
	}
	return w.Click(p)
}

func turning(spare, by int) (lines, left int) {
	total := spare + by
	return total / lineHigh, total % lineHigh
}

func (w Window) wheeling(by int, p Point, now time.Time) Window {
	goingOn := !w.wheelAt.IsZero() && now.Sub(w.wheelAt) < Still()
	w.wheelAt = now
	if goingOn && w.wheelOn != w.screen {
		return w
	}
	w.wheelOn = w.screen
	return w.Scroll(p, by)
}

// Scroll moves the column under the pointer, and no other.
func (w Window) Scroll(p Point, by int) Window {
	if w.screen == OnForm {
		if frame, ok := w.partsFrame(); ok && frame.In(p) {
			return w.scrollingParts(by)
		}
		if !w.form.Adding && BesideBox(w.formPanel()).In(p) {
			if w.form.ForObjective {
				w.serving = w.settledServing(w.serving + by)
				return w
			}
			lines, left := turning(w.spareHistory, by)
			w.history, w.spareHistory = w.settledHistory(w.history+lines), left
			return w
		}
		return w.scrollingSaying(p, by)
	}
	if w.screen == OnHelp {
		if w.HelpBox().In(p) {
			lines, left := turning(w.spareHelp, by)
			w.helpFrom, w.spareHelp = w.settledHelp(w.HelpFrom()+lines), left
		}
		return w
	}
	if w.screen == OnAsking || w.screen == OnMenu {
		return w
	}
	if w.screen == OnCalendar {
		return w.scrollingMonths(p, by)
	}
	if w.screen == OnBackups {
		return w.scrollingBackups(p, by)
	}
	if w.screen == OnChoosing {
		box := w.ChoosingBox()
		if ChoosingRun(box).In(p) || ChoosingBar(box).In(p) {
			lines, left := turning(w.spareChoosing, by)
			w.choosingFrom, w.spareChoosing = w.settledChoosing(w.choosingFrom+lines), left
		}
		return w
	}
	l := w.Layout()
	if w.tab == Time {
		return w.scrollingDay(p, by)
	}
	if w.tab == Objectives {
		if GoalsRun(l).In(p) {
			w.goals = w.settledGoals(w.goals + by)
		}
		return w
	}
	for i := range engine.Columns {
		if !l.Column[i].In(p) {
			continue
		}
		w.scroll[i] = w.settled(i, w.scroll[i]+by)
		return w
	}
	if lane, found := l.LaneUnder(p); found {
		w.lanes[lane] = w.settledLane(lane, w.lanes[lane]+by)
	}
	return w
}

func (w Window) at(p Point) int {
	if w.tab == Objectives {
		return w.goalAt(p)
	}
	if w.tab == Time {
		return w.timeCardAt(p)
	}
	l := w.Layout()
	if column, _, found := l.Under(p, 0); found {
		here := w.inColumn(column)
		index := (p.Y - l.Column[column].Y + w.scroll[column]) / CardHeight
		if index < len(here) {
			return here[index].Number
		}
		return 0
	}
	if lane, found := l.LaneUnder(p); found {
		here := w.inLane(lane)
		index := (p.Y - l.Lane[lane].Y + w.lanes[lane]) / CardHeight
		if index < len(here) {
			return here[index].Number
		}
	}
	return 0
}

func (w Window) dropped(number int, p Point) Window {
	state, place, found := w.where(w.Layout(), p)
	if !found {
		return w
	}
	given := map[string]string{"number": strconv.Itoa(number)}
	if state != w.stateOf(number) {
		given["state"] = state
	}
	for name, value := range place {
		given[name] = value
	}
	if len(given) == 1 {
		return w.Look()
	}
	return w.asking(engine.TaskMove, engine.Input{Given: given})
}

func (w Window) where(l Layout, p Point) (state string, place map[string]string, found bool) {
	if lane, in := l.LaneUnder(p); in {
		return engine.Stalled[lane], nil, true
	}
	for i := range engine.Columns {
		if !l.Column[i].In(p) {
			continue
		}
		here := w.inColumn(i)
		at := (p.Y - l.Column[i].Y + w.scroll[i]) / CardHeight
		if at <= 0 {
			return engine.StateAt(i), map[string]string{"first": "true"}, true
		}
		if at-1 < len(here) {
			return engine.StateAt(i), map[string]string{"after": strconv.Itoa(here[at-1].Number)}, true
		}
		return engine.StateAt(i), map[string]string{"last": "true"}, true
	}
	return "", nil, false
}

func (w Window) clickedOff(p Point) Window {
	overForm := w.screen == OnChoosing || w.screen == OnCalendar
	if w.screen == OnCalendar {
		w.screen = OnForm
	} else if w.screen != OnForm {
		w = w.Choose("Close")
	}
	if overForm && w.formPanel().In(p) {
		return w
	}
	if w.screen != OnForm {
		return w
	}
	if w.form.Changed() {
		return w.Save()
	}
	return w.Choose("Close")
}
