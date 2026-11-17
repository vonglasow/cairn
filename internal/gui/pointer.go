package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"slices"
	"strconv"
	"strings"
	"time"

	"cairn/internal/engine"
)

// Click chooses whatever is under the pointer, and nothing when there is nothing there.
func (w Window) Click(p geom.Point) Window {
	if w.Stranded() {
		if BlankOffer(w.width, w.height).In(p) {
			return w.openingConfiguration()
		}
		return w
	}
	if copied, ok := w.copyingAt(p); ok {
		return copied
	}
	if tabbed, ok := w.clickingBesideTab(p); ok {
		return tabbed
	}
	if w.Screen == board.OnForm {
		if opened, ok := w.clickingWork(p); ok {
			return opened
		}
	}
	if w.Screen == board.OnTheBoard && !w.native && MenuButton(w.Layout()).In(p) {
		return w.openingMenuBar()
	}
	if w.Screen == board.OnMenu {
		return w.clickingMenu(p)
	}
	if w.Screen == board.OnWorkspaces {
		return w.clickingWorkspaces(p)
	}
	if w.Screen == board.OnTheBoard && w.Layout().Place.In(p) {
		return w.openingSwitcher()
	}
	if w.Screen == board.OnPalette {
		return w.clickingPalette(p)
	}
	if w.Screen == board.OnTheBoard && CommandLine(w.Layout(), w.BarOffers()).In(p) {
		return w.openingPalette()
	}
	if clicked, on := w.clickingPills(p); on {
		return clicked
	}
	if clicked, on := w.clickingDetail(p); on {
		return clicked
	}
	if clicked, on := w.clickingSeries(p); on {
		return clicked
	}
	if cross, ok := w.PanelClose(); ok && cross.In(p) {
		return w.Choose(w.leavingOffer())
	}
	if clicked, on := w.clickingBanners(p); on {
		return clicked
	}
	if w.Screen == board.OnConfiguration {
		if offer, on := w.offerAt(p); on {
			return w.Choose(offer)
		}
		if !w.IntrosOpen && !w.ChooserOpen && !w.SettingsBox().In(p) && len(w.ColourSwatches()) == 0 {
			return w.leavingSettings()
		}
		return w.clickingSettings(p)
	}
	if w.Screen == board.OnBackups {
		if offer, on := w.offerAt(p); on {
			return w.Choose(offer)
		}
		return w.clickingBackups(p)
	}
	if w.Screen == board.OnLabels {
		if offer, on := w.offerAt(p); on {
			return w.Choose(offer)
		}
		return w.clickingLabels(p)
	}
	if w.Screen == board.OnFiltering {
		if offer, on := w.offerAt(p); on {
			return w.Choose(offer)
		}
		return w.clickingFilters(p)
	}
	if w.Screen == board.OnKeeping || w.Screen == board.OnViews {
		if offer, on := w.offerAt(p); on {
			return w.Choose(offer)
		}
		if w.Screen == board.OnKeeping {
			return w.clickingKeeping(p)
		}
		return w.clickingViewsPanel(p)
	}
	if offer, on := w.offerAt(p); on {
		return w.Choose(offer)
	}
	if w.Screen == board.OnFates {
		return w.clickingFates(p)
	}
	if w.Screen == board.OnSeriesFates {
		return w.clickingSeriesFates(p)
	}
	if w.Screen == board.OnChoosing {
		return w.clickingList(p)
	}
	if w.Screen == board.OnRule {
		return w.clickingRule(p)
	}
	if w.Screen == board.OnCalendar && !w.CalendarBox().In(p) {
		return w.clickedOff(p)
	}
	if w.Screen == board.OnCalendar {
		if walked, on := w.monthClicked(p); on {
			return walked
		}
		w, _ = w.dayClicked(p)
		return w
	}
	if edit, ok := w.OccurrenceEdit(); ok && w.Screen == board.OnForm && edit.In(p) {
		return w.Choose(EditTheSeries)
	}
	if at, on := w.fieldAt(p); on {
		reached := w.reaching(at, p)
		w.Form = w.Form.Onto(at)
		if chosenFromAList(w.Form.Fields[at]) {
			return w.choosingList()
		}
		if chosenOnACalendar(w.Form.Fields[at]) {
			return w.openingCalendar()
		}
		if chosenFromRules(w.Form.Fields[at]) {
			return w.openingRule()
		}
		return w.CaretInto(reached)
	}
	if w.ClosesOnAClickOff() && !w.panel().In(p) {
		return w.clickedOff(p)
	}
	if w.Screen == board.OnSearch {
		if Searching(w.Layout()).In(p) {
			return w
		}
	} else if w.Screen != board.OnTheBoard {
		return w
	}
	if tab, on := w.tabAt(p); on {
		return w.onPage(tab)
	}
	if w.Tab == board.Time {
		return w.clickingTime(p)
	}
	if n, on := w.seriesChipAt(p); on {
		return w.onSeries(n)
	}
	at := w.at(p)
	if at == 0 && w.Tab == board.Objectives {
		return w
	}
	return w.aiming(p).choosingAt(at)
}

func (w Window) aiming(p geom.Point) Window {
	if w.Tab != board.Tasks {
		return w
	}
	if state, _, found := w.where(w.Layout(), p); found {
		w.aimed = state
	}
	return w
}

// Hovering answers the window the pointer resting at a point leaves behind.
func (w Window) Hovering(p geom.Point) Window {
	w.over, w.hovered = w.offerAt(p)
	w.explaining = w.nameAt(p)
	w.resting, w.restingAt, w.restingWide = w.cutTitleAt(p)
	if w.resting == nil {
		w.resting, w.restingAt, w.restingWide = w.servedAt(p)
	}
	if w.resting == nil {
		w.resting, w.restingAt, w.restingWide = w.cutValueAt(p)
	}
	if w.resting == nil {
		w.resting, w.restingAt, w.restingWide = w.untailedAt(p)
	}
	if w.resting == nil {
		w.resting, w.restingAt, w.restingWide = w.cutFilterAt(p)
	}
	if w.resting == nil {
		w.resting, w.restingAt, w.restingWide = w.switcherPathAt(p)
	}
	if w.resting == nil {
		w.resting, w.restingAt, w.restingWide = w.hiddenAt(p)
	}
	w.pointed = w.slotAt(p)
	w.bubbled, w.bubbledFrom = w.countAt(p)
	if w.Screen == board.OnConfiguration {
		w = w.hoveringSettings(p)
	}
	return w
}

// Bubble is what the bubble the pointer opened says, a line to each thing it names: the whole title of the cut card the pointer rests on, or the objectives its mark stands for; nothing when it rests on neither.
func (w Window) Bubble() string { return strings.Join(w.resting, "\n") }

// BubbleBox is where the bubble is drawn and the lines it says, and whether there is one.
func (w Window) BubbleBox() (geom.Box, []string, bool) {
	if w.resting == nil {
		return geom.Box{}, nil, false
	}
	m := w.measured()
	room := min(2*w.restingWide, w.width-2*geom.Margin) - 2*padding
	var lines []string
	for _, said := range w.resting {
		lines = append(lines, m.Wrapped(said, evenly(m, said, room, plain), plain).Said...)
	}
	widest := 0
	for _, line := range lines {
		widest = max(widest, ends(m, line, plain))
	}
	wide := widest + 2*padding
	high := len(lines)*lineHigh + 2*inset
	near := w.restingAt
	x := max(geom.Margin, min(near.X-padding, w.width-geom.Margin-wide))
	y := near.Y - high - inset
	if y < geom.Margin {
		y = near.Y + near.H + inset
	}
	return geom.Box{X: x, Y: max(geom.Margin, min(y, w.height-geom.Margin-high)), W: wide, H: high}, lines, true
}

func (w Window) cutTitleAt(p geom.Point) ([]string, geom.Box, int) {
	if w.Screen != board.OnTheBoard || w.Tab != board.Tasks {
		return nil, geom.Box{}, 0
	}
	for _, placed := range w.Placed() {
		title := PartsOf(placed.Slot).Title
		if placed.Within.In(p) && title.In(p) && w.titleCut(placed.Card, placed.Slot) {
			return []string{placed.Card.Title}, title, placed.Slot.W
		}
	}
	return nil, geom.Box{}, 0
}

func (w Window) servedAt(p geom.Point) ([]string, geom.Box, int) {
	if w.Screen != board.OnTheBoard || w.Tab != board.Tasks {
		return nil, geom.Box{}, 0
	}
	for _, placed := range w.Placed() {
		if !placed.Within.In(p) || !placed.Slot.In(p) {
			continue
		}
		mark := CardRowOf(placed.Card, PartsOf(placed.Slot).Number, w.measured(), w.Now, w.Pages).ServesAt
		if !mark.In(p) {
			continue
		}
		var said []string
		for _, n := range placed.Card.Objectives {
			line := "#" + strconv.Itoa(n)
			if g, ok := w.Snap.Goal(n); ok {
				line += "  " + g.Title
			}
			said = append(said, line)
		}
		return said, mark, placed.Slot.W
	}
	return nil, geom.Box{}, 0
}

func (w Window) titleCut(c board.Card, slot geom.Box) bool {
	m := w.measured()
	first, rest := TitleRoom(c, PartsOf(slot).Title)
	lines := TitleLines(c.Title, first, rest, func(said string) int { return ends(m, said, reads) })
	return strings.HasSuffix(lines[len(lines)-1], "…") && !strings.HasSuffix(c.Title, "…")
}

func (w Window) countAt(p geom.Point) ([]string, geom.Box) {
	if w.Screen != board.OnTheBoard || w.Tab != board.Tasks {
		return nil, geom.Box{}
	}
	for _, placed := range w.Placed() {
		if !placed.Within.In(p) || !placed.Slot.In(p) {
			continue
		}
		if more := CardRowOf(placed.Card, PartsOf(placed.Slot).Number, w.measured(), w.Now, w.Pages).More; more.In(p) {
			return placed.Card.Labels, more
		}
	}
	return nil, geom.Box{}
}

// LabelsBubbled is every label of the card whose count the pointer rests on, and nothing when it rests on none.
func (w Window) LabelsBubbled() []string { return w.bubbled }

// LabelsBubbleBox is where the bubble of those labels is drawn, under their count, and each label's pill in it, wrapped to its width.
func (w Window) LabelsBubbleBox() (geom.Box, []geom.Box, bool) {
	if len(w.bubbled) == 0 {
		return geom.Box{}, nil, false
	}
	m := w.measured()
	inner := labelsBubbleWide - 2*inset*2
	var pills []geom.Box
	x, y := 0, 0
	for _, label := range w.bubbled {
		wide := min(inner, ends(m, label, small)+2*pillPad)
		if x > 0 && x+wide > inner {
			x, y = 0, y+pillHigh+pillGap
		}
		pills = append(pills, geom.Box{X: x, Y: y, W: wide, H: pillHigh})
		x += wide + pillGap
	}
	high := y + pillHigh + 2*inset*2
	at := geom.Point{X: min(w.bubbledFrom.X, w.width-geom.Margin-labelsBubbleWide), Y: w.bubbledFrom.Y + w.bubbledFrom.H + inset}
	if at.Y+high > w.height-geom.Margin {
		at.Y = w.bubbledFrom.Y - inset - high
	}
	box := geom.Box{X: at.X, Y: at.Y, W: labelsBubbleWide, H: high}
	for i := range pills {
		pills[i].X += box.X + 2*inset
		pills[i].Y += box.Y + 2*inset
	}
	return box, pills, true
}

const labelsBubbleWide = 180

func (w Window) slotAt(p geom.Point) geom.Box {
	if w.Screen == board.OnTheBoard && w.Tab == board.Objectives {
		return w.goalSlotAt(p)
	}
	if w.Screen == board.OnTheBoard && w.Tab == board.Series {
		return w.seriesSlotAt(p)
	}
	if w.Screen != board.OnTheBoard || w.Tab != board.Tasks {
		return geom.Box{}
	}
	for _, placed := range w.Placed() {
		if placed.Within.In(p) && placed.Slot.In(p) {
			return placed.Slot
		}
	}
	return geom.Box{}
}

// Away is the window the pointer having left it leaves behind.
func (w Window) Away() Window {
	w.over, w.hovered, w.explaining, w.resting, w.restingAt, w.pointed = "", false, 0, nil, geom.Box{}, geom.Box{}
	w.bubbled, w.bubbledFrom = nil, geom.Box{}
	return w
}

func (w Window) nameAt(p geom.Point) int {
	if w.Screen != board.OnForm {
		return 0
	}
	fields := w.Form.Fields
	for i := range fields {
		if name, _, _ := FormPlace(w.formBox(), fields, i); name.In(p) {
			return i + 1
		}
	}
	return 0
}

// Explains is what the field whose name the pointer rests on is, as the engine says it; nothing when it rests on no field's name.
func (w Window) Explains() string {
	fields := w.Form.Fields
	if w.Screen != board.OnForm || w.explaining == 0 || w.explaining > len(fields) {
		return ""
	}
	return fields[w.explaining-1].Means
}

// ExplainsBy is the example of what to type in the field whose name the pointer rests on, and nothing when there is none to give.
func (w Window) ExplainsBy() string {
	fields := w.Form.Fields
	if w.Explains() == "" {
		return ""
	}
	return fields[w.explaining-1].Example
}

// ExampleSays is how an explanation introduces its example.
const ExampleSays = "e.g. "

// Explanation is the box a field's explanation is drawn in and the lines it says — what the field is, then an example of what to type — as wide and as high as those lines and no more, and whether anything is explained at all.
func (w Window) Explanation() (geom.Box, []string, bool) {
	said := w.Explains()
	if said == "" {
		return geom.Box{}, nil, false
	}
	m := w.measured()
	lines := m.Wrapped(said, evenly(m, said, tipWide-2*padding, small), small).Said
	if by := w.ExplainsBy(); by != "" {
		lines = append(lines, m.Wrapped(ExampleSays+by, tipWide-2*padding, small).Said...)
	}
	widest := 0
	for _, line := range lines {
		widest = max(widest, ends(m, line, small))
	}
	name, _, _ := FormPlace(w.formBox(), w.Form.Fields, w.explaining-1)
	return ExplainedAt(name, w.width, widest+2*padding, len(lines)), lines, true
}

func evenly(m Measure, said string, room, size int) int {
	lines := len(m.Wrapped(said, room, size).Said)
	least, most := room/2, room
	for least < most {
		mid := (least + most) / 2
		if len(m.Wrapped(said, mid, size).Said) > lines {
			least = mid + 1
		} else {
			most = mid
		}
	}
	return most
}

func (w Window) cutValueAt(p geom.Point) ([]string, geom.Box, int) {
	if w.Screen != board.OnForm {
		return nil, geom.Box{}, 0
	}
	fields := w.Form.Fields
	for i, field := range fields {
		_, value, rows := FormPlace(w.formBox(), fields, i)
		if !value.In(p) || field.Long || field.Arg == "label" || field.Name == "steps" || i == w.Form.On() {
			continue
		}
		said := w.FieldSays(field)
		if ends(w.measured(), said, FormSize(field)) > Inside(Framed(value, rows)).W {
			return []string{said}, value, value.W
		}
	}
	return nil, geom.Box{}, 0
}

// ExplainedAt is where a field's explanation is drawn: just under its name, as wide and as high as asked, kept inside the window.
func ExplainedAt(name geom.Box, width, wide, lines int) geom.Box {
	x := min(name.X, width-geom.Margin-wide)
	return geom.Box{X: max(geom.Margin, x), Y: name.Y + name.H + inset, W: wide, H: lines*lineHigh + 2*inset}
}

// ExplainedLine is where the line at of an explanation is written inside its box.
func ExplainedLine(tip geom.Box, at int) geom.Box {
	return geom.Box{X: tip.X + padding, Y: tip.Y + inset + at*lineHigh, W: tip.W - 2*padding, H: lineHigh}
}

const tipWide = 360

// Lighting says whether an offer is the one the pointer is resting on, which is the one thing in the window that says a word can be clicked.
func (w Window) Lighting(offer string) bool { return w.hovered && w.over == offer }

func (w Window) goalAt(p geom.Point) int {
	l := w.Layout()
	if !GoalsRun(l).In(p) {
		return 0
	}
	for i, b := range w.Tree() {
		if ItemOf(GoalAt(l, i, w.goals), depthOf(b)).In(p) {
			return b.Goal.Number
		}
	}
	return 0
}

func (w Window) tabAt(p geom.Point) (board.Tab, bool) {
	l := w.Layout()
	for i, tab := range w.Shown() {
		if TabAt(l, i).In(p) {
			return tab, true
		}
	}
	return w.Tab, false
}

// DoubleClick opens the detail of the card under the pointer, which is where the description and the history are.
func (w Window) DoubleClick(p geom.Point) Window {
	if w.Stranded() {
		return w
	}
	if widening, on := gutterAt(w.Layout(), w.sizes, p); on && w.Tab == board.Tasks && w.Screen == board.OnTheBoard {
		return w.resized(widening.Evened())
	}
	if w.Screen == board.OnChoosing {
		box := w.ChoosingBox()
		if at, on := ChoosingAtPoint(box, len(w.Choosing()), w.ChoosingFrom, p); on {
			w.ChoosingOn = at
			w = w.taking()
			if w.Screen == board.OnChoosing {
				return w.keeping()
			}
		}
		return w
	}
	if at, on := w.partUnder(p); on && w.Screen == board.OnForm {
		field, _ := w.Form.PartsField()
		w.Form = w.Form.Onto(field)
		w.Form.PartAt = at
		w.Form = w.Form.Naming()
		return w
	}
	if held, ok := w.holdingSetting(p, Words); ok {
		return held
	}
	if w.Screen == board.OnForm {
		if at, on := w.fieldAt(p); on {
			return w.holdingBy(at, p, Words)
		}
		return w.word(p)
	}
	if w.Screen != board.OnTheBoard && w.Screen != board.OnSearch {
		return w
	}
	number := w.at(p)
	if _, _, inColumn := w.where(w.Layout(), p); number == 0 && inColumn && w.Tab == board.Tasks && w.Screen == board.OnTheBoard {
		w = w.aiming(p)
		behind := w.above(p)
		w = w.startingNew()
		w.Form.Behind = behind
		return w
	}
	if number == 0 {
		return w
	}
	w.Chosen, w.ChosenGoal = number, w.Tab == board.Objectives
	return w.opening()
}

// TripleClick holds the paragraph under the pointer, on an open task and in a field of the form: the third press of the gesture every text answers.
func (w Window) TripleClick(p geom.Point) Window {
	if at, on := w.fieldAt(p); on {
		return w.holdingBy(at, p, Paragraphs)
	}
	if w.Screen == board.OnForm {
		return w.paragraph(p)
	}
	return w
}

func (w Window) holdingBy(at int, p geom.Point, by Unit) Window {
	reached := w.reaching(at, p)
	w.Form = w.Form.Onto(at)
	from, to := runOf(w.Form.Writing().Reads(), by, reached)
	held := w.Form.Writing()
	held.At, held.Held = to, from-to
	w.Form = w.Form.Writes(held)
	w.holdingText, w.growing, w.anchor = true, by, Picked{From: from, To: to}
	return w
}

// PressAt takes hold of whatever is under the pointer, without deciding yet whether this is a click or a drag.
func (w Window) PressAt(p geom.Point) Window {
	w.pressed, w.pressedOn = true, w.Screen
	if w.Stranded() {
		return w
	}
	if w.Screen == board.OnChoosing {
		w.draggingChoosing = ChoosingBar(w.ChoosingBox()).In(p)
		return w
	}
	if taken, ok := w.takingPanel(p); ok {
		return taken
	}
	if slid, ok := w.pressingSlider(p); ok {
		return slid
	}
	if held, ok := w.holdingSetting(p, Characters); ok {
		return held
	}
	if w.Screen == board.OnForm {
		if _, bar, _, _, long := w.sayingOn(); long && bar.In(p) {
			w.draggingSaying = true
			return w
		}
		if at, on := w.partUnder(p); on {
			field, _ := w.Form.PartsField()
			w = w.keepingPartInSight()
			w.Form = w.Form.Onto(field)
			if w.gathering || w.stretching || !slices.Contains(w.Form.TakenParts(), at) {
				w.Form = w.Form.TakingPart(at, w.gathering, w.stretching)
			} else {
				w.Form.PartAt = at
			}
			w.Form.Carried = at + 1
			w.held = Held{From: p, At: p}
			return w
		}
		if FormGrip(w.formPanel()).In(p) {
			w.resizingForm, w.resizedFrom, w.held = true, w.formPanel(), Held{From: p, At: p}
			return w
		}
		if at, on := w.fieldAt(p); on {
			w.picked = Picked{}
			return w.holdingBy(at, p, Characters).inSightOfTheCaret()
		}
		if w.Tracing() && BesideBar(w.formPanel()).In(p) {
			w.draggingBar = true
			return w
		}
		if w.Form.ForObjective && w.BesideShown() && w.workBar().In(p) {
			w.draggingServing = true
			return w
		}
		if w.formPanel().In(p) && !w.textAt(p) {
			w.draggingForm, w.dragged, w.formGrip, w.held = true, false, p, Held{From: p, At: p}
			return w
		}
		return w.picking(p)
	}
	if w.Screen != board.OnTheBoard {
		return w
	}
	if w.Tab == board.Objectives {
		if n := w.goalAt(p); n != 0 {
			w.grabbedGoal, w.held = n, Held{From: p, At: p}
		}
		return w
	}
	if widening, on := gutterAt(w.Layout(), w.sizes, p); on && w.Tab == board.Tasks {
		w.widening = widening
		return w
	}
	w.held = Held{Number: w.at(p), From: p, At: p}
	return w
}

// MoveTo follows the pointer, and past Wobble says the card is being dragged.
func (w Window) MoveTo(p geom.Point) Window {
	if w.widening.On {
		w.sizes = w.widening.To(p)
		return w
	}
	if w.resizingForm {
		return w.resizingTheForm(p)
	}
	if w.movingPanel || w.resizingPanel {
		return w.carryingPanel(p)
	}
	if w.sliding != "" {
		return w.slidingTo(p)
	}
	if w.draggingForm {
		return w.movingForm(p)
	}
	if w.draggingBar {
		track := BesideBar(w.formPanel())
		w.history = w.settledHistory(Reached(track, w.Told(), w.Rows(), p))
		return w
	}
	if w.holdingText && w.Screen == board.OnConfiguration {
		return w.stretchingSetting(p)
	}
	if w.holdingText && w.Screen == board.OnForm {
		held := w.Form.Writing()
		fixed, moving := Stretched(held.Reads(), w.growing, w.anchor, w.reaching(w.Form.On(), p))
		held.At, held.Held = moving, fixed-moving
		w.Form = w.Form.Writes(held)
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
		w.ChoosingFrom = w.settledChoosing(
			Reached(ChoosingBar(box), len(w.Choosing()), ChoosingRows(box), p))
		return w
	}
	if w.draggingServing {
		room := max(1, w.workFrame().H/geom.SubHeight)
		w.serving = w.settledServing(Reached(w.workBar(), len(w.Under())+1, room, p) * geom.SubHeight)
		return w
	}
	if w.held.Number == 0 && w.Form.Carried == 0 && w.grabbedGoal == 0 {
		return w
	}
	w.held.At = p
	if abs(p.X-w.held.From.X) > Wobble || abs(p.Y-w.held.From.Y) > Wobble {
		w.held.Moved = true
	}
	if w.GoalDragging() {
		return w.hoveringGoal(p)
	}
	return w
}

// LetGo drops the card, or lets a click be a click.
func (w Window) LetGo(p geom.Point) Window {
	if w.widening.On {
		w.widening, w.pressed = Widening{}, false
		return w.resized(w.sizes)
	}
	if w.resizingForm {
		w.resizingForm, w.held, w.pressed = false, Held{}, false
		return w.resized(w.sizes)
	}
	if w.movingPanel || w.resizingPanel {
		return w.lettingGoOfPanel(p)
	}
	if w.sliding != "" {
		return w.slidingTo(p).lettingGoOfSlider()
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
		if w.Form.Writing().Holding() || w.Screen == board.OnConfiguration && w.KeyWritten.Holding() {
			return w
		}
	}
	waiting, was := w.pressed, w.pressedOn
	w.pressed = false
	if !waiting || was != w.Screen {
		w.held, w.grabbedGoal = Held{}, 0
		return w
	}
	if w.grabbedGoal != 0 {
		return w.lettingGoOfAGoal(p)
	}
	if w.Form.Carried != 0 {
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

func (w Window) wheeling(by int, p geom.Point, now time.Time) Window {
	goingOn := !w.wheelAt.IsZero() && now.Sub(w.wheelAt) < Still()
	w.wheelAt = now
	if goingOn && w.wheelOn != w.Screen {
		return w
	}
	w.wheelOn = w.Screen
	return w.Scroll(p, by)
}

// Scroll moves the column under the pointer, and no other.
func (w Window) Scroll(p geom.Point, by int) Window {
	if w.Screen == board.OnConfiguration {
		return w.scrollingSettings(by)
	}
	if w.Screen == board.OnForm {
		if frame, ok := w.partsFrame(); ok && frame.In(p) {
			return w.scrollingParts(by)
		}
		if w.Form.ForObjective && w.BesideShown() && w.workFrame().In(p) {
			w.serving = w.settledServing(w.serving + by)
			return w
		}
		if w.Tracing() && BesideBox(w.formPanel()).In(p) {
			lines, left := turning(w.spareHistory, by)
			w.history, w.spareHistory = w.settledHistory(w.history+lines), left
			return w
		}
		return w.scrollingSaying(p, by)
	}
	if w.Screen == board.OnHelp || w.Screen == board.OnAsking || w.Screen == board.OnMenu {
		return w
	}
	if w.Screen == board.OnCalendar {
		return w.scrollingMonths(p, by)
	}
	if w.Screen == board.OnBackups {
		return w.scrollingBackups(p, by)
	}
	if w.Screen == board.OnLabels {
		return w.scrollingLabels(p, by)
	}
	if w.Screen == board.OnChoosing {
		box := w.ChoosingBox()
		if ChoosingRun(box).In(p) || ChoosingBar(box).In(p) {
			lines, left := turning(w.spareChoosing, by)
			w.ChoosingFrom, w.spareChoosing = w.settledChoosing(w.ChoosingFrom+lines), left
		}
		return w
	}
	l := w.Layout()
	if w.Tab == board.Time {
		return w.scrollingDay(p, by)
	}
	if w.Tab == board.Objectives {
		if GoalsRun(l).In(p) {
			w.goals = w.settledGoals(w.goals + by)
		}
		return w
	}
	for i := range board.Columns {
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

func (w Window) at(p geom.Point) int {
	if w.Tab == board.Objectives {
		return w.goalAt(p)
	}
	if w.Tab == board.Time {
		return w.timeCardAt(p)
	}
	for _, placed := range w.Placed() {
		if placed.Within.In(p) && PartsOf(placed.Slot).Panel.In(p) {
			return placed.Card.Number
		}
	}
	return 0
}

func (w Window) above(p geom.Point) int {
	l := w.Layout()
	for _, placed := range w.Placed() {
		if placed.Within.In(p) && slices.Contains(l.Column[:], placed.Within) && placed.Slot.In(p) && !PartsOf(placed.Slot).Panel.In(p) {
			return placed.Card.Number
		}
	}
	return 0
}

func (w Window) dropped(number int, p geom.Point) Window {
	state, place, found := w.where(w.Layout(), p)
	w.held = Held{}
	if !found {
		return w
	}
	given := map[string]string{"number": strconv.Itoa(number)}
	if state != w.StateOf(number) {
		given["state"] = state
	}
	for name, value := range place {
		given[name] = value
	}
	var along []string
	if w.IsTaken(number) {
		for _, other := range w.Taken() {
			if other != number {
				along = append(along, strconv.Itoa(other))
			}
		}
		w.Gathered = nil
	}
	in := engine.Input{Given: given, Listed: map[string][]string{"also": along}}
	if len(given) == 1 && len(along) == 0 {
		return w.Look()
	}
	return w.asking(engine.TaskMove, in)
}

func (w Window) where(l geom.Layout, p geom.Point) (state string, place map[string]string, found bool) {
	if lane, in := l.LaneUnder(p); in {
		all := w.InLane(lane)
		return board.Stalled[lane], placeAfter(w.besideHeld(all), w.leftBefore(all, landed(l.Lane[lane], w.lanes[lane], p, len(all)))), true
	}
	for i := range board.Columns {
		if !l.Column[i].In(p) {
			continue
		}
		all := w.InColumn(i)
		return board.StateAt(i), placeAfter(w.besideHeld(all), w.leftBefore(all, landed(l.Column[i], w.scroll[i], p, len(all)))), true
	}
	return "", nil, false
}

func (w Window) leftBefore(all []board.Card, at int) int {
	kept := len(w.besideHeld(all[:min(at, len(all))]))
	return kept
}

func (w Window) ownPlace(all []board.Card, at int) bool {
	for i, c := range all {
		if c.Number == w.held.Number {
			return at == i || at == i+1
		}
	}
	return false
}

func placeAfter(here []board.Card, at int) map[string]string {
	if at == 0 {
		return map[string]string{"first": "true"}
	}
	return map[string]string{"after": strconv.Itoa(here[at-1].Number)}
}

func landed(body geom.Box, scrolled int, p geom.Point, room int) int {
	return max(0, min(room, (p.Y-body.Y+scrolled+geom.CardHeight/2)/geom.CardHeight))
}

func (w Window) besideHeld(cards []board.Card) []board.Card {
	if !w.Dragging() {
		return cards
	}
	kept := make([]board.Card, 0, len(cards))
	for _, c := range cards {
		if c.Number != w.held.Number && (!w.IsTaken(w.held.Number) || !w.IsTaken(c.Number)) {
			kept = append(kept, c)
		}
	}
	return kept
}

func (w Window) openedAt(l geom.Layout) (column, at int, open bool) {
	if !w.Dragging() {
		return 0, 0, false
	}
	if _, inLane := l.LaneUnder(w.held.At); inLane {
		return 0, 0, false
	}
	for i := range board.Columns {
		if l.Column[i].In(w.held.At) {
			return i, landed(l.Column[i], w.scroll[i], w.held.At, len(w.InColumn(i))), true
		}
	}
	return 0, 0, false
}

func (w Window) openedInLane(l geom.Layout) (lane, at int, open bool) {
	if !w.Dragging() {
		return 0, 0, false
	}
	lane, in := l.LaneUnder(w.held.At)
	if !in {
		return 0, 0, false
	}
	return lane, landed(l.Lane[lane], w.lanes[lane], w.held.At, len(w.InLane(lane))), true
}

func (w Window) clickedOff(p geom.Point) Window {
	if w.PickingLabels && w.Screen == board.OnChoosing {
		return w.leavingThePicker()
	}
	if w.Screen == board.OnCalendar && w.calendarKey != "" {
		return w.leavingCalendar()
	}
	overForm := w.Screen == board.OnChoosing || w.Screen == board.OnCalendar || w.Screen == board.OnRule
	if w.Screen == board.OnCalendar || w.Screen == board.OnRule {
		w.Screen = board.OnForm
	} else if w.Screen == board.OnChoosing {
		w = w.Choose("Cancel")
	} else if w.Screen != board.OnForm {
		w = w.Choose("Close")
	}
	if overForm && w.formPanel().In(p) {
		return w
	}
	if w.Screen != board.OnForm {
		return w
	}
	if w.Form.Changed() {
		w.Screen, w.Question = board.OnAsking, board.AboutLeaving
		return w
	}
	if w.Form.Adding && !w.Form.ForObjective {
		return w.leavingUnsaved()
	}
	return w.Choose("Close")
}
