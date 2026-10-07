package gui

import (
	"strconv"
	"time"

	"cairn/internal/engine"
	"cairn/internal/reading"
)

// What the pointer does: a click, a double and a triple, a press carried into
// a drag and let go, a wheel turned, and what it rests on.

// Click chooses whatever is under the pointer, and nothing when there is
// nothing there.
func (w Window) Click(p Point) Window {
	// What is offered answers first, wherever it is drawn: it is what closes
	// a panel, so it is the one thing a panel does not swallow.
	if offer, on := w.offerAt(p); on {
		return w.Choose(offer)
	}
	// A list answers before the form does. It is drawn over one, and the boxes
	// underneath are still where they were — a click meant for a row used to
	// fall through and move the keyboard to whatever field was behind it.
	if w.screen == OnChoosing {
		return w.clickingList(p)
	}
	// A form's own fields answer, because that is how the keyboard gets to one
	// with a mouse.
	if at, on := w.fieldAt(p); on {
		reached := w.reaching(at, p)
		w.form = w.form.onto(at)
		// A field a list answers opens it — a closed set among them. There is
		// nothing to type there, so a click that only moved the keyboard onto
		// it left a pointer with nowhere to go and nothing saying a list was
		// behind it.
		if chosenFromAList(w.form.Fields()[at]) {
			return w.choosingList()
		}
		return w.CaretInto(reached)
	}
	// Off a panel is out of it, where getting out costs nothing. A hand on the
	// mouse should not have to reach for the keyboard to put something away.
	//
	// A form is the exception and the only one: it holds what somebody typed
	// and has not saved, and a click landing anywhere outside would throw that
	// away without asking. esc is still there, and the bar offers it.
	if w.closesOnAClickOff() && !w.panel().In(p) {
		return w.Choose("Close")
	}
	// A search is not a panel over the board: the board is still there, still
	// showing what the search left, and a card under the pointer is a card.
	// Making it wait for enter meant closing the box to reach what the box had
	// just found. Over the box itself a click belongs to what is being typed.
	if w.screen == OnSearch {
		if Searching(w.Layout()).In(p) {
			return w
		}
	} else if w.screen != OnTheBoard {
		// And a panel over the board does have the window. Letting a click
		// through to the board behind it is how an open detail came to show
		// something nobody asked it for.
		return w
	}
	if tab, on := w.tabAt(p); on {
		w.tab, w.chosen, w.chosenGoal, w.taken = tab, 0, false, nil
		return w
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

// Hovered is the offer the pointer is resting on.
func (w Window) Hovered() (string, bool) { return w.over, w.hovered }

// Lighting says whether an offer is the one the pointer is resting on, which
// is the one thing in the window that says a word can be clicked. Asked here
// rather than worked out in the drawing, so a scenario can put the question.
func (w Window) Lighting(offer string) bool { return w.hovered && w.over == offer }

// goalAt says which objective a point lands on in the list.
func (w Window) goalAt(p Point) int {
	l := w.Layout()
	here := reading.Sorted(w.Showing().Goals)
	for i := range here {
		if GoalAt(l, i, w.goals).In(p) {
			return here[i].Number
		}
	}
	return 0
}

// tabAt says which tab a point falls on, and whether it falls on one at all.
// The strip is wider than the two tabs together, so clicking beside them is
// clicking nothing.
func (w Window) tabAt(p Point) (Tab, bool) {
	l := w.Layout()
	for _, tab := range []Tab{Tasks, Objectives} {
		if TabAt(l, int(tab)).In(p) {
			return tab, true
		}
	}
	return w.tab, false
}

// DoubleClick opens the detail of the card under the pointer, which is where
// the description and the history are.
func (w Window) DoubleClick(p Point) Window {
	// In a list it takes the row and is done with it, which is the gesture a
	// hand already has: one click says which row, two say that one.
	if w.screen == OnChoosing {
		box := ChoosingOver(w.width, w.height)
		if at, on := ChoosingAtPoint(box, len(w.Choosing()), w.choosingFrom, p); on {
			w.choosing = at
			w = w.taking()
			if w.screen == OnChoosing {
				return w.keeping()
			}
		}
		return w
	}
	// In a field of the form it holds the word it lands on, and beside the
	// form the same: every piece of text written can be held and copied.
	if w.screen == OnForm {
		if at, on := w.fieldAt(p); on {
			return w.holdingBy(at, p, Words)
		}
		return w.word(p)
	}
	// The same two screens a single click answers on, for the same reason: a
	// search leaves the board showing, so a card under the pointer is a card.
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

// TripleClick holds the paragraph under the pointer, on an open task and in a
// field of the form: the third press of the gesture every text answers.
func (w Window) TripleClick(p Point) Window {
	if at, on := w.fieldAt(p); on {
		return w.holdingBy(at, p, Paragraphs)
	}
	if w.screen == OnForm {
		return w.paragraph(p)
	}
	return w
}

// holdingBy holds the unit of a field a press lands on, puts the keyboard on
// that field, and keeps the run as the anchor a drag grows from. The character
// is read before the keyboard moves, off the field as it was on screen.
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

// PressAt takes hold of whatever is under the pointer, without deciding yet
// whether this is a click or a drag.
func (w Window) PressAt(p Point) Window {
	w.pressed, w.pressedOn = true, w.screen
	// A press in a field is where a run of text may be held from, so the caret
	// goes where the press landed and not to the field's end: the end is where
	// somebody arriving at a field types, and this one is not arriving.
	if w.screen == OnChoosing {
		w.draggingChoosing = ChoosingBar(ChoosingOver(w.width, w.height)).In(p)
		return w
	}
	if w.screen == OnForm {
		// The bar beside the long field the keyboard is on is taken hold of,
		// and a press on it is not a press in the text.
		if _, bar, _, _, long := w.sayingOn(); long && bar.In(p) {
			w.draggingSaying = true
			return w
		}
		if at, on := w.fieldAt(p); on {
			w.picked = Picked{}
			return w.holdingBy(at, p, Characters).inSightOfTheCaret()
		}
		// Beside a form opened on something already there: the bar of that
		// column is taken hold of, and anywhere else is where a run of what it
		// writes may be held from.
		if !w.form.Adding && BesideBar(w.formPanel()).In(p) {
			w.draggingBar, w.draggingServing = !w.form.ForObjective, w.form.ForObjective
			return w
		}
		return w.picking(p)
	}
	if w.screen != OnTheBoard {
		return w
	}
	// Nothing is taken hold of on the objectives page. An objective and a task
	// can carry the same number, and what is dragged is let go into a column —
	// which only a task has. Dragging one used to move the other.
	if w.tab == Objectives {
		return w
	}
	w.held = Held{Number: w.at(p), From: p, At: p}
	return w
}

// MoveTo follows the pointer, and past Grip says the card is being dragged.
func (w Window) MoveTo(p Point) Window {
	if w.draggingBar {
		track := BesideBar(Over(w.width, w.height))
		w.history = w.settledHistory(Reached(track, w.Told(), w.Rows(), p))
		return w
	}
	// A pointer held down in a field holds a run of its text, and goes on
	// holding it once the pointer leaves the field: a hand overshoots, and a
	// run that stopped where the frame did stopped short of where it was going.
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
		box := ChoosingOver(w.width, w.height)
		w.choosingFrom = w.settledChoosing(
			Reached(ChoosingBar(box), len(w.Choosing()), ChoosingRows(box), p))
		return w
	}
	if w.draggingServing {
		// Counted in tiles and settled in pixels, since that box is scrolled
		// by pixels and not by lines.
		track := BesideBar(Over(w.width, w.height))
		room := max(1, BesideBox(Over(w.width, w.height)).H/SubHeight)
		w.serving = w.settledServing(Reached(track, len(w.Under()), room, p) * SubHeight)
		return w
	}
	if w.held.Number == 0 {
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
		// Let go where it was let go. A run held with the mouse is finished by
		// lifting it, and the click that would follow would collapse the run
		// it just made.
		w.holdingText = false
		if w.form.Writing().Holding() {
			return w
		}
	}
	// A release with no press behind it on this screen is not a click here.
	// A double click opens a task on the press, and the release that follows
	// landed on the board it had just left — closing what it had just opened,
	// for anyone whose pointer was on the part of a card the panel does not
	// cover.
	waiting, was := w.pressed, w.pressedOn
	w.pressed = false
	if !waiting || was != w.screen {
		w.held = Held{}
		return w
	}
	held := w.held
	w.held = Held{}
	if held.Number != 0 && held.Moved {
		return w.dropped(held.Number, p)
	}
	// Everything else is a click, including a press that took hold of nothing.
	// A tab is not a card and neither is an offer in the bar: making this the
	// card's business is what left both unreachable in a real window, where
	// every click arrives as a press and then a release.
	return w.Click(p)
}

// turning adds what a wheel sent to what the turns before it left over, and
// answers whole lines and the new leftover. A trackpad sends a few pixels at a
// time: taken one by one they each round to nothing and are thrown away, and
// added up they are a line like any other.
func turning(spare, by int) (lines, left int) {
	total := spare + by
	return total / lineHigh, total % lineHigh
}

// wheeling is one part of a turn of the wheel. A part close enough to the last
// one goes on the same gesture, and is dropped if what that gesture began on
// has gone: a trackpad goes on sending after the fingers leave it, and a task
// put away with esc while it was still sending handed the rest of the turn to
// the board — the column under the pointer scrolled though nobody had turned
// anything over it. A part after a pause is a turn of its own.
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
	// A panel stops the wheel. Over a form, the help or the question it used to
	// reach the board behind, and scrolled something nobody could see. What
	// it reaches on a form is the long field the keyboard is on, and the
	// column beside a form opened on something already there: a task's
	// history, or the work serving an objective.
	if w.screen == OnForm {
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
	if w.screen == OnHelp || w.screen == OnAsking {
		return w
	}
	if w.screen == OnChoosing {
		box := ChoosingOver(w.width, w.height)
		if ChoosingRun(box).In(p) || ChoosingBar(box).In(p) {
			lines, left := turning(w.spareChoosing, by)
			w.choosingFrom, w.spareChoosing = w.settledChoosing(w.choosingFrom+lines), left
		}
		return w
	}
	l := w.Layout()
	if w.tab == Objectives {
		// One grid across the four columns' worth of width, so anywhere over
		// the board is over the objectives.
		if GoalsRun(l).In(p) {
			w.goals = w.settledGoals(w.goals + by)
		}
		return w
	}
	for i := range Columns {
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

// at says which task is under a point, in a column or in one of the band's
// lanes, and zero when there is none.
func (w Window) at(p Point) int {
	// The objectives tab is a list across the whole board, and its rows lie
	// over the columns behind them. Asking the columns there answers whichever
	// task happens to sit at the same height, which is not what was pointed at.
	if w.tab == Objectives {
		return w.goalAt(p)
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

// dropped is what a card let go somewhere means. A drop says two things at
// once — which column, and which place among the cards there — and the engine
// is told them in that order, because it refuses to do both in one act.
//
// There is no transaction between the two. If the second is refused, the task
// has changed column and not place, and the board says so when it looks again:
// visible rather than silent.
func (w Window) dropped(number int, p Point) Window {
	l := w.Layout()
	state, place, found := w.where(l, p)
	if !found {
		return w
	}
	if state != w.stateOf(number) {
		asked := engine.Input{Given: map[string]string{
			"number": strconv.Itoa(number),
			"state":  state,
		}}
		if _, err := w.work.Ask(engine.TaskEdit, asked); err != nil {
			w.err = err
			return w
		}
	}
	if len(place) > 0 {
		given := map[string]string{"number": strconv.Itoa(number)}
		for name, value := range place {
			given[name] = value
		}
		return w.asking(engine.TaskMove, engine.Input{Given: given})
	}
	return w.Look()
}

func (w Window) where(l Layout, p Point) (state string, place map[string]string, found bool) {
	if lane, in := l.LaneUnder(p); in {
		return Lanes[lane], nil, true
	}
	for i := range Columns {
		if !l.Column[i].In(p) {
			continue
		}
		here := w.inColumn(i)
		at := (p.Y - l.Column[i].Y + w.scroll[i]) / CardHeight
		if at <= 0 {
			return reading.StateAt(i), map[string]string{"first": "true"}, true
		}
		if at-1 < len(here) {
			return reading.StateAt(i), map[string]string{"after": strconv.Itoa(here[at-1].Number)}, true
		}
		return reading.StateAt(i), map[string]string{"last": "true"}, true
	}
	return "", nil, false
}
