package gui

import (
	"slices"
	"strconv"
	"time"

	"cairn/internal/engine"
)

// Click chooses whatever is under the pointer, and nothing when there is nothing there.
func (w Window) Click(p Point) Window {
	if w.Blank() {
		return w
	}
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
	if w.screen == OnLabels {
		if offer, on := w.offerAt(p); on {
			return w.Choose(offer)
		}
		return w.clickingLabels(p)
	}
	if w.screen == OnFiltering {
		if offer, on := w.offerAt(p); on {
			return w.Choose(offer)
		}
		return w.clickingFilters(p)
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
	return w.aiming(p).choosingAt(w.at(p))
}

func (w Window) aiming(p Point) Window {
	if w.tab != Tasks {
		return w
	}
	if state, _, found := w.where(w.Layout(), p); found {
		w.aimed = state
	}
	return w
}

// Hovering answers the window the pointer resting at a point leaves behind.
func (w Window) Hovering(p Point) Window {
	w.over, w.hovered = w.offerAt(p)
	w.explaining = w.nameAt(p)
	w.resting, w.restingSlot = w.cutTitleAt(p)
	return w
}

// Bubble is the whole title of the cut card the pointer rests on, and nothing when it rests on none.
func (w Window) Bubble() string { return w.resting }

// BubbleBox is where the bubble is drawn and the lines it says, and whether there is one.
func (w Window) BubbleBox() (Box, []string, bool) {
	if w.resting == "" {
		return Box{}, nil, false
	}
	m := w.measured()
	room := min(2*w.restingSlot.W, w.width-2*margin) - 2*padding
	lines := m.Wrapped(w.resting, room, plain).Said
	widest := 0
	for _, line := range lines {
		widest = max(widest, ends(m, line, plain))
	}
	wide := widest + 2*padding
	high := len(lines)*lineHigh + 2*inset
	title := PartsOf(w.restingSlot).Title
	x := max(margin, min(title.X-padding, w.width-margin-wide))
	y := title.Y - high - inset
	if y < margin {
		y = title.Y + title.H + inset
	}
	return Box{X: x, Y: max(margin, min(y, w.height-margin-high)), W: wide, H: high}, lines, true
}

func (w Window) cutTitleAt(p Point) (string, Box) {
	if w.screen != OnTheBoard || w.tab != Tasks {
		return "", Box{}
	}
	for _, placed := range w.Placed() {
		if placed.Within.In(p) && placed.Slot.In(p) && TitleOf(placed.Card.Title, placed.Slot.W) != placed.Card.Title {
			return placed.Card.Title, placed.Slot
		}
	}
	return "", Box{}
}

// Away is the window the pointer having left it leaves behind.
func (w Window) Away() Window {
	w.over, w.hovered, w.explaining, w.resting, w.restingSlot = "", false, 0, "", Box{}
	return w
}

func (w Window) nameAt(p Point) int {
	if w.screen != OnForm {
		return 0
	}
	fields := w.form.Fields()
	for i := range fields {
		if name, _, _ := FormPlace(w.formBox(), fields, i); name.In(p) {
			return i + 1
		}
	}
	return 0
}

// Explains is what the field whose name the pointer rests on is, as the engine says it; nothing when it rests on no field's name.
func (w Window) Explains() string {
	fields := w.form.Fields()
	if w.screen != OnForm || w.explaining == 0 || w.explaining > len(fields) {
		return ""
	}
	return fields[w.explaining-1].Means
}

// ExplainsBy is the example of what to type in the field whose name the pointer rests on, and nothing when there is none to give.
func (w Window) ExplainsBy() string {
	fields := w.form.Fields()
	if w.Explains() == "" {
		return ""
	}
	return fields[w.explaining-1].Example
}

// ExampleSays is how an explanation introduces its example.
const ExampleSays = "e.g. "

// Explanation is the box a field's explanation is drawn in and the lines it says — what the field is, then an example of what to type — as wide and as high as those lines and no more, and whether anything is explained at all.
func (w Window) Explanation() (Box, []string, bool) {
	said := w.Explains()
	if said == "" {
		return Box{}, nil, false
	}
	m := w.measured()
	lines := m.Wrapped(said, tipWide-2*padding, small).Said
	if by := w.ExplainsBy(); by != "" {
		lines = append(lines, ExampleSays+by)
	}
	widest := 0
	for _, line := range lines {
		widest = max(widest, ends(m, line, small))
	}
	name, _, _ := FormPlace(w.formBox(), w.form.Fields(), w.explaining-1)
	return ExplainedAt(name, w.width, widest+2*padding, len(lines)), lines, true
}

// ExplainedAt is where a field's explanation is drawn: just under its name, as wide and as high as asked, kept inside the window.
func ExplainedAt(name Box, width, wide, lines int) Box {
	x := min(name.X, width-margin-wide)
	return Box{X: max(margin, x), Y: name.Y + name.H + inset, W: wide, H: lines*lineHigh + 2*inset}
}

// ExplainedLine is where the line at of an explanation is written inside its box.
func ExplainedLine(tip Box, at int) Box {
	return Box{X: tip.X + padding, Y: tip.Y + inset + at*lineHigh, W: tip.W - 2*padding, H: lineHigh}
}

const tipWide = 360

// OfferLit is where an offer under the pointer is lit: around its words, as wide as they are, with as much room on every side of them.
func OfferLit(words Box, wide int) Box {
	return Box{X: words.X - inset, Y: words.Y - inset, W: wide + 2*inset, H: lineHigh + 2*inset}
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
	for i, tab := range w.Shown() {
		if TabAt(l, i).In(p) {
			return tab, true
		}
	}
	return w.tab, false
}

// DoubleClick opens the detail of the card under the pointer, which is where the description and the history are.
func (w Window) DoubleClick(p Point) Window {
	if w.Blank() {
		return w
	}
	if widening, on := gutterAt(w.Layout(), w.sizes, p); on && w.tab == Tasks && w.screen == OnTheBoard {
		return w.resized(widening.Evened())
	}
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
	if _, _, inColumn := w.where(w.Layout(), p); number == 0 && inColumn && w.tab == Tasks && w.screen == OnTheBoard {
		w = w.aiming(p)
		behind := w.above(p)
		w = w.startingNew()
		w.form.behind = behind
		return w
	}
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
	if w.Blank() {
		return w
	}
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
			if w.gathering || w.stretching || !slices.Contains(w.form.takenParts(), at) {
				w.form = w.form.takingPart(at, w.gathering, w.stretching)
			} else {
				w.form.partAt = at
			}
			w.form.carried = at + 1
			w.held = Held{From: p, At: p}
			return w
		}
		if at, on := w.fieldAt(p); on {
			w.picked = Picked{}
			return w.holdingBy(at, p, Characters).inSightOfTheCaret()
		}
		if w.besideShown() && BesideBar(w.formPanel()).In(p) {
			w.draggingBar, w.draggingServing = !w.form.ForObjective, w.form.ForObjective
			return w
		}
		if w.formPanel().In(p) && !w.textAt(p) {
			w.draggingForm, w.dragged, w.formGrip, w.held = true, false, p, Held{From: p, At: p}
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
	if widening, on := gutterAt(w.Layout(), w.sizes, p); on && w.tab == Tasks {
		w.widening = widening
		return w
	}
	w.held = Held{Number: w.at(p), From: p, At: p}
	return w
}

// MoveTo follows the pointer, and past Grip says the card is being dragged.
func (w Window) MoveTo(p Point) Window {
	if w.widening.On {
		w.sizes = w.widening.To(p)
		return w
	}
	if w.draggingForm {
		return w.movingForm(p)
	}
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
	if w.widening.On {
		w.widening, w.pressed = Widening{}, false
		return w.resized(w.sizes)
	}
	if w.draggingForm {
		w.draggingForm, w.held, w.pressed = false, Held{}, false
		if w.dragged {
			return w
		}
		return w.Click(p)
	}
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
	if w.Dragging() {
		return w.dropped(w.held.Number, p)
	}
	w.held = Held{}
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
		if w.besideShown() && BesideBox(w.formPanel()).In(p) {
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
	if w.screen == OnLabels {
		return w.scrollingLabels(p, by)
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
		return cardOn(w.inColumn(column), p.Y-l.Column[column].Y+w.scroll[column])
	}
	if lane, found := l.LaneUnder(p); found {
		return cardOn(w.inLane(lane), p.Y-l.Lane[lane].Y+w.lanes[lane])
	}
	return 0
}

func cardOn(here []engine.Card, down int) int {
	index := down / CardHeight
	if index >= len(here) || down%CardHeight >= CardHeight-gap {
		return 0
	}
	return here[index].Number
}

func (w Window) above(p Point) int {
	l := w.Layout()
	column, _, found := l.Under(p, 0)
	if !found {
		return 0
	}
	here := w.inColumn(column)
	down := p.Y - l.Column[column].Y + w.scroll[column]
	if index := down / CardHeight; index < len(here) && down%CardHeight >= CardHeight-gap {
		return here[index].Number
	}
	return 0
}

func (w Window) dropped(number int, p Point) Window {
	state, place, found := w.where(w.Layout(), p)
	w.held = Held{}
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
	var along []string
	if w.isTaken(number) {
		for _, other := range w.Taken() {
			if other != number {
				along = append(along, strconv.Itoa(other))
			}
		}
		w.taken = nil
	}
	in := engine.Input{Given: given, Listed: map[string][]string{"also": along}}
	if len(given) == 1 && len(along) == 0 {
		return w.Look()
	}
	return w.asking(engine.TaskMove, in)
}

func (w Window) where(l Layout, p Point) (state string, place map[string]string, found bool) {
	if lane, in := l.LaneUnder(p); in {
		here := w.besideHeld(w.inLane(lane))
		return engine.Stalled[lane], placeAfter(here, landed(l.Lane[lane], w.lanes[lane], p, len(here))), true
	}
	for i := range engine.Columns {
		if !l.Column[i].In(p) {
			continue
		}
		here := w.besideHeld(w.inColumn(i))
		return engine.StateAt(i), placeAfter(here, landed(l.Column[i], w.scroll[i], p, len(here))), true
	}
	return "", nil, false
}

func placeAfter(here []engine.Card, at int) map[string]string {
	if at == 0 {
		return map[string]string{"first": "true"}
	}
	return map[string]string{"after": strconv.Itoa(here[at-1].Number)}
}

func landed(body Box, scrolled int, p Point, room int) int {
	return max(0, min(room, (p.Y-body.Y+scrolled+CardHeight/2)/CardHeight))
}

func (w Window) besideHeld(cards []engine.Card) []engine.Card {
	if !w.Dragging() {
		return cards
	}
	kept := make([]engine.Card, 0, len(cards))
	for _, c := range cards {
		if c.Number != w.held.Number && (!w.isTaken(w.held.Number) || !w.isTaken(c.Number)) {
			kept = append(kept, c)
		}
	}
	return kept
}

func (w Window) openedAt(l Layout) (column, at int, open bool) {
	if !w.Dragging() {
		return 0, 0, false
	}
	if _, inLane := l.LaneUnder(w.held.At); inLane {
		return 0, 0, false
	}
	for i := range engine.Columns {
		if l.Column[i].In(w.held.At) {
			return i, landed(l.Column[i], w.scroll[i], w.held.At, len(w.besideHeld(w.inColumn(i)))), true
		}
	}
	return 0, 0, false
}

func (w Window) openedInLane(l Layout) (lane, at int, open bool) {
	if !w.Dragging() {
		return 0, 0, false
	}
	lane, in := l.LaneUnder(w.held.At)
	if !in {
		return 0, 0, false
	}
	return lane, landed(l.Lane[lane], w.lanes[lane], w.held.At, len(w.besideHeld(w.inLane(lane)))), true
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
	if w.form.Adding && !w.form.ForObjective {
		return w.leavingUnsaved()
	}
	if w.form.Changed() {
		return w.Save()
	}
	return w.Choose("Close")
}
