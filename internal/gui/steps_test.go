package gui

// What the scenarios look at once a gesture has been made. Everything here
// reads the window's own state or the store through the second connection — a
// precondition or an assertion that went through the code under test would
// cancel out with it and prove nothing.

import (
	"fmt"
	"image/color"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
	"cairn/internal/objective"
	"cairn/internal/reading"
	"cairn/internal/task"
)

// quoted reads a list written the way a scenario writes one: "a", "b", "c".
func quoted(list string) []string {
	var out []string
	for _, part := range strings.Split(list, ",") {
		out = append(out, strings.Trim(strings.TrimSpace(part), `"`))
	}
	return out
}

func (w *world) columnsRead(list string) error {
	want := quoted(list)
	var got []string
	dealt, _ := reading.Deal(w.w.Reading().Cards, w.w.archived)
	for _, column := range dealt {
		got = append(got, column.Name)
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		return fmt.Errorf("the columns read %v, want %v", got, want)
	}
	return nil
}

func (w *world) cardsOf(column string) ([]reading.Card, error) {
	at := reading.IndexOf(strings.ToLower(column))
	if at < 0 {
		return nil, fmt.Errorf("no column called %q", column)
	}
	dealt, _ := reading.Deal(w.w.Showing().Cards, w.w.archived)
	return dealt[at].Cards, nil
}

func (w *world) columnCarries(column, title string) error {
	here, err := w.cardsOf(column)
	if err != nil {
		return err
	}
	for _, c := range here {
		if c.Title == title {
			return nil
		}
	}
	return fmt.Errorf("%q does not carry %q", column, title)
}

func (w *world) columnReads(column, list string) error {
	here, err := w.cardsOf(column)
	if err != nil {
		return err
	}
	var got []string
	for _, c := range here {
		got = append(got, c.Title)
	}
	want := quoted(list)
	if strings.Join(got, "|") != strings.Join(want, "|") {
		return fmt.Errorf("%q reads %v, want %v", column, got, want)
	}
	return nil
}

func (w *world) bandNamed(name string) error {
	if reading.Band != name {
		return fmt.Errorf("the band is named %q", reading.Band)
	}
	return nil
}

func (w *world) lanesRead(list string) error {
	want := quoted(list)
	if strings.Join(Lanes, "|") != strings.Join(want, "|") {
		return fmt.Errorf("the lanes read %v, want %v", Lanes, want)
	}
	return nil
}

func (w *world) laneCarries(kind, title string) error {
	_, stalled := reading.Deal(w.w.Reading().Cards, w.w.archived)
	for _, c := range stalled {
		if c.State == kind && c.Title == title {
			return nil
		}
	}
	return fmt.Errorf("the lane %q does not carry %q", kind, title)
}

func (w *world) tabsRead(list string) error {
	want := quoted(list)
	got := []string{"TASKS", "OBJECTIVES"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		return fmt.Errorf("the tabs read %v, want %v", got, want)
	}
	return nil
}

func (w *world) tabIs(name string) error {
	got := "TASKS"
	if w.w.Tab() == Objectives {
		got = "OBJECTIVES"
	}
	if got != name {
		return fmt.Errorf("the tab is %q, want %q", got, name)
	}
	return nil
}

func (w *world) sameHeight(column string) error {
	here, err := w.cardsOf(column)
	if err != nil {
		return err
	}
	at := reading.IndexOf(strings.ToLower(column))
	l := w.w.Layout()
	for i := range here {
		if got := l.CardAt(at, i, 0).H; got != CardHeight {
			return fmt.Errorf("card %d is %d tall, the rest are %d", i, got, CardHeight)
		}
	}
	return nil
}

func (w *world) titleEndsIn(column, ending string) error {
	here, err := w.cardsOf(column)
	if err != nil {
		return err
	}
	at := reading.IndexOf(strings.ToLower(column))
	width := w.w.Layout().Column[at].W
	for _, c := range here {
		if strings.HasSuffix(TitleOf(c.Title, width), ending) {
			return nil
		}
	}
	return fmt.Errorf("no card in %q has a title ending in %q", column, ending)
}

func (w *world) tooSmall() error {
	if w.w.Screen() != TooSmallToDraw {
		return fmt.Errorf("the window is on %v", w.w.Screen())
	}
	return nil
}

func (w *world) saysWhatItNeeds() error {
	said := w.w.Says()
	if !strings.Contains(said, fmt.Sprint(Narrowest)) || !strings.Contains(said, fmt.Sprint(Shortest)) {
		return fmt.Errorf("it says %q, which names neither size it needs", said)
	}
	return nil
}

func (w *world) numberOf(title string) (int, error) {
	for _, c := range w.w.Reading().Cards {
		if c.Title == title {
			return c.Number, nil
		}
	}
	return 0, fmt.Errorf("no task titled %q", title)
}

func (w *world) isChosen(title string) error {
	// An objective is chosen by the same field under a different kind, so the
	// title is looked for where the chosen thing actually lives.
	if w.w.ChosenGoal() {
		for _, g := range w.w.Reading().Goals {
			if g.Title == title {
				if w.w.Chosen() != g.Number {
					return fmt.Errorf("objective %d is chosen, not %q", w.w.Chosen(), title)
				}
				return nil
			}
		}
		return fmt.Errorf("no objective titled %q", title)
	}
	number, err := w.numberOf(title)
	if err != nil {
		return err
	}
	if w.w.Chosen() != number {
		return fmt.Errorf("task %d is chosen, not %q", w.w.Chosen(), title)
	}
	return nil
}

func (w *world) isNotChosen(title string) error {
	number, err := w.numberOf(title)
	if err != nil {
		return err
	}
	if w.w.Chosen() == number {
		return fmt.Errorf("%q is chosen and should not be", title)
	}
	return nil
}

func (w *world) detailOpenOn(title string) error {
	if w.w.Screen() != OnForm {
		return fmt.Errorf("the window is on %v", w.w.Screen())
	}
	return w.isChosen(title)
}

// detailLists is asked of the work an open objective says serves it, by title:
// a count says how far along it is and never what is left.
// Asked of the lines the panel draws and not of what the window knows: a
// window that has the answer and does not put it on screen has not answered.
// Asked of the cards the box under the fields draws, and not of what the
// window knows: a window that has the answer and does not put it on screen has
// not answered.
func (w *world) detailLists(title string) error {
	for _, c := range w.w.Under() {
		if c.Title == title {
			return nil
		}
	}
	return fmt.Errorf("the box under the fields holds %v, and no card for %q", titles(w.w.Serving()), title)
}

func (w *world) detailDoesNotList(title string) error {
	for _, c := range w.w.Under() {
		if c.Title == title {
			return fmt.Errorf("the box under the fields holds a card for %q, which serves something else", title)
		}
	}
	return nil
}

// servingCardsAreCards checks the tiles are a line high and no more: a card's
// 96 pixels were mostly room for facts a sub-task no longer carries, and the
// box holds five of these where it held one of those.
func (w *world) servingCardsAreCards() error {
	for at := range w.w.Under() {
		if got := w.w.ServingCardAt(at).H; got != SubHeight {
			return fmt.Errorf("the tile at %d is %d high, and a tile is %d", at, got, SubHeight)
		}
	}
	if SubHeight > lineHigh+padding {
		return fmt.Errorf("a tile is %d high, which is more than one line and its padding", SubHeight)
	}
	return nil
}

func (w *world) detailSaysNothingServes() error {
	if got := w.w.Under(); len(got) != 0 {
		return fmt.Errorf("the box under the fields holds %v, and nothing serves it", titles(got))
	}
	return nil
}

// titles is what a failure prints of a run of cards.
func titles(cards []reading.Card) []string {
	var said []string
	for _, c := range cards {
		said = append(said, c.Title)
	}
	return said
}

// rightColumnHead is the assertion that the history keeps its column: it is
// what a task opens onto, and an objective must not take it away.
func (w *world) rightColumnHead(head string) error {
	got := w.w.PanelHead()
	if w.w.ChosenGoal() {
		got = w.w.UnderHead()
	}
	if got != head {
		return fmt.Errorf("the right column is headed %q, not %q", got, head)
	}
	return nil
}

func (w *world) detailShows(text string) error {
	for _, c := range w.w.Reading().Cards {
		if c.Number == w.w.Chosen() && strings.Contains(c.Description, text) {
			return nil
		}
	}
	return fmt.Errorf("the detail does not show %q", text)
}

func (w *world) isScrolled(column string) error {
	at := reading.IndexOf(strings.ToLower(column))
	if w.w.scroll[at] == w.scrolled[at] {
		return fmt.Errorf("%q did not move", column)
	}
	return nil
}

// formFrames is where every field of the open form is drawn: its name and its
// frame, the bar beside a long one included.
func (w *world) formFrames() ([]Box, []Box) {
	over := w.w.formBox()
	fields := w.w.Form().Fields()
	var names, frames []Box
	for i := range fields {
		name, value, rows := FormPlace(over, fields, i)
		frame := Framed(value, rows)
		if fields[i].Long {
			var bar Box
			frame, bar = Saying(value, rows)
			frame.W = bar.X + bar.W - frame.X
		}
		names, frames = append(names, name), append(frames, frame)
	}
	return names, frames
}

func (w *world) formFitsAboveItsOffers() error {
	foot := FormRule(w.w.formPanel()).Y
	names, frames := w.formFrames()
	for i, frame := range frames {
		if bottom := max(frame.Y+frame.H, names[i].Y+names[i].H); bottom >= foot {
			return fmt.Errorf("field %q ends at %d and the offers start at %d", w.w.Form().Fields()[i].Name, bottom, foot)
		}
	}
	return nil
}

func overlap(a, b Box) bool {
	return a.X < b.X+b.W && b.X < a.X+a.W && a.Y < b.Y+b.H && b.Y < a.Y+a.H
}

func (w *world) formFieldsApart() error {
	names, frames := w.formFrames()
	fields := w.w.Form().Fields()
	all := append(append([]Box{}, names...), frames...)
	for i := range all {
		for j := i + 1; j < len(all); j++ {
			if overlap(all[i], all[j]) {
				return fmt.Errorf("%q and %q overlap: %v and %v",
					fields[i%len(fields)].Name, fields[j%len(fields)].Name, all[i], all[j])
			}
		}
	}
	return nil
}

// acrossTheForm says a field's frame runs from the left of the form to its
// right, as far as any frame does, and that only the fields named are above it.
func (w *world) acrossTheForm(name string, above []string) error {
	over := w.w.formBox()
	fields := w.w.Form().Fields()
	// The frames alone, without a long field's bar: it sits in a margin every
	// field keeps, which is what lines the frames up.
	frames := make([]Box, len(fields))
	for i := range fields {
		_, value, rows := FormPlace(over, fields, i)
		frames[i] = Framed(value, rows)
	}
	left, right, at := frames[0].X, 0, -1
	for i, frame := range frames {
		left, right = min(left, frame.X), max(right, frame.X+frame.W)
		if fields[i].Name == name {
			at = i
		}
	}
	if at < 0 {
		return fmt.Errorf("the form has no field called %q", name)
	}
	if frame := frames[at]; frame.X != left || frame.X+frame.W != right {
		return fmt.Errorf("%q runs from %d to %d, and the form from %d to %d", name, frame.X, frame.X+frame.W, left, right)
	}
	for i, frame := range frames {
		if i == at || slices.Contains(above, fields[i].Name) {
			continue
		}
		if frame.Y <= frames[at].Y {
			return fmt.Errorf("%q is at %d, not under %q at %d", fields[i].Name, frame.Y, name, frames[at].Y)
		}
	}
	return nil
}

func (w *world) laneIsScrolled(kind string) error {
	lane, err := laneNamed(kind)
	if err != nil {
		return err
	}
	if w.w.lanes[lane] == w.laneScrolled[lane] {
		return fmt.Errorf("the lane %q did not move", kind)
	}
	return nil
}

func (w *world) laneIsNotScrolled(kind string) error {
	lane, err := laneNamed(kind)
	if err != nil {
		return err
	}
	if w.w.lanes[lane] != w.laneScrolled[lane] {
		return fmt.Errorf("the lane %q moved and should not have", kind)
	}
	return nil
}

func (w *world) isNotScrolled(column string) error {
	at := reading.IndexOf(strings.ToLower(column))
	if w.w.scroll[at] != w.scrolled[at] {
		return fmt.Errorf("%q moved and should not have", column)
	}
	return nil
}

func (w *world) askedNothing() error {
	if n := len(w.work.wrote); n != 0 {
		return fmt.Errorf("the store was asked %d things", n)
	}
	return nil
}

func (w *world) wasAsked(what string, number int) error {
	thing, verb, _ := strings.Cut(what, " ")
	for _, a := range w.work.wrote {
		if a.thing == thing && a.verb == verb && a.number == fmt.Sprint(number) {
			return nil
		}
	}
	return fmt.Errorf("the store was never asked %q for task %d: it was asked %v", what, number, w.work.wrote)
}

func (w *world) wasNotAsked(what string) error {
	thing, verb, _ := strings.Cut(what, " ")
	for _, a := range w.work.wrote {
		if a.thing == thing && a.verb == verb {
			return fmt.Errorf("the store was asked %q and should not have been", what)
		}
	}
	return nil
}

func (w *world) wasAskedOnce(what string) error {
	thing, verb, _ := strings.Cut(what, " ")
	count := 0
	for _, a := range w.work.wrote {
		if a.thing == thing && a.verb == verb {
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("the store was asked %q %d times", what, count)
	}
	return nil
}

func (w *world) taskInColumn(number int, state string) error {
	told, err := w.tell(number)
	if err != nil {
		return err
	}
	if told.State != state {
		return fmt.Errorf("task %d is in %q, want %q", number, told.State, state)
	}
	return nil
}

func (w *world) taskCalled(number int, title string) error {
	told, err := w.tell(number)
	if err != nil {
		return err
	}
	if told.Title != title {
		return fmt.Errorf("task %d is called %q, want %q", number, told.Title, title)
	}
	return nil
}

// tell reads a task back through the second connection, so an assertion never
// goes through the window it is judging.
func (w *world) tell(number int) (task.Shape, error) {
	found, err := w.givens().Task(number)
	if err != nil {
		return task.Shape{}, fmt.Errorf("task %d: %w", number, err)
	}
	return task.OfTask(found), nil
}

func (w *world) formAsksAbout(list string) error {
	if w.w.Screen() != OnForm {
		return fmt.Errorf("no form is open")
	}
	var got []string
	for _, f := range w.w.Form().Fields() {
		got = append(got, f.Name)
	}
	for _, want := range quoted(list) {
		found := false
		for _, name := range got {
			found = found || name == want
		}
		if !found {
			return fmt.Errorf("the form does not ask about %q: it asks about %v", want, got)
		}
	}
	return nil
}

func (w *world) fieldReads(name, value string) error {
	if got := w.w.Form().ValueOf(name); got != value {
		return fmt.Errorf("the field %q reads %q, want %q", name, got, value)
	}
	return nil
}

func (w *world) readOnce() error {
	if w.work.reads != 1 {
		return fmt.Errorf("the store was read %d times", w.work.reads)
	}
	return nil
}

func (w *world) saysWhatWentWrong() error {
	if w.w.Wrong() == nil {
		return fmt.Errorf("the window says nothing went wrong")
	}
	return nil
}

// cardDimmed reads the ink the card would be written in, which is what the
// drawing asks the skin for.
// cardNumbered is the card a scenario names by its number.
func (w *world) cardNumbered(n int) (reading.Card, error) {
	for _, c := range w.w.Showing().Cards {
		if c.Number == n {
			return c, nil
		}
	}
	return reading.Card{}, fmt.Errorf("the board carries no card numbered %d", n)
}

// markedDescribed reads the mark a card writes beside its number.
func (w *world) markedDescribed(n int, want bool) error {
	c, err := w.cardNumbered(n)
	if err != nil {
		return err
	}
	if got := strings.Contains(Carried(c, Marks()), Marks().Described); got != want {
		return fmt.Errorf("the card for task %d is marked as described: %v, want %v", n, got, want)
	}
	return nil
}

// listInsideThePanel asks of every row drawn that it is inside the run the
// panel gives them. A list longer than its panel used to draw straight past it
// and over the window behind.
func (w *world) listInsideThePanel() error {
	box := ChoosingOver(w.w.width, w.w.height)
	run := ChoosingRun(box)
	offered, from, room := w.w.Choosing(), w.w.ChoosingFromRow(), w.w.ChoosingRoom()
	for i := from; i < min(len(offered), from+room); i++ {
		row := ChoosingRow(box, ChoosingFrom(box)+(i-from)*apart)
		if row.Y < run.Y || row.Y+row.H > run.Y+run.H {
			return fmt.Errorf("row %d is drawn at %d..%d, and the panel holds %d..%d",
				i, row.Y, row.Y+row.H, run.Y, run.Y+run.H)
		}
	}
	return nil
}

func (w *world) cardDimmed(title string) error {
	for _, c := range w.w.Showing().Cards {
		if c.Title != title {
			continue
		}
		if ink := Dark().Writing(c); ink != Dark().Faint {
			return fmt.Errorf("%q is written in %v, not the faint ink", title, ink)
		}
		return nil
	}
	return fmt.Errorf("the board carries no card called %q", title)
}

func (w *world) cardNotDimmed(title string) error {
	for _, c := range w.w.Showing().Cards {
		if c.Title != title {
			continue
		}
		if ink := Dark().Writing(c); ink != Dark().Ink {
			return fmt.Errorf("%q is written in %v, not the ordinary ink", title, ink)
		}
		return nil
	}
	return fmt.Errorf("the board carries no card called %q", title)
}

func (w *world) expects(sc *godog.ScenarioContext) {
	sc.Then(`^the columns read (".*")$`, w.columnsRead)
	sc.Then(`^"([^"]*)" carries "([^"]*)"$`, w.columnCarries)
	sc.Then(`^the card "([^"]*)" is dimmed$`, w.cardDimmed)
	sc.Then(`^the card for task (\d+) is marked as described$`, func(n int) error {
		return w.markedDescribed(n, true)
	})
	sc.Then(`^the card for task (\d+) is not marked as described$`, func(n int) error {
		return w.markedDescribed(n, false)
	})
	sc.Then(`^the card for task (\d+) says "([^"]*)"$`, func(n int, said string) error {
		c, err := w.cardNumbered(n)
		if err != nil {
			return err
		}
		if line := Carried(c, Marks()); !strings.Contains(line, said) {
			return fmt.Errorf("the card for task %d reads %q beside its number, and says nothing of %q",
				n, line, said)
		}
		return nil
	})
	sc.Then(`^the card "([^"]*)" is not dimmed$`, w.cardNotDimmed)
	sc.Then(`^"([^"]*)" reads (".*")$`, w.columnReads)
	sc.Then(`^the band is named "([^"]*)"$`, w.bandNamed)
	sc.Then(`^the band's lanes read (".*")$`, w.lanesRead)
	sc.Then(`^the lane "([^"]*)" carries "([^"]*)"$`, w.laneCarries)
	sc.Then(`^the tabs read (".*")$`, w.tabsRead)
	sc.Then(`^the tab is "([^"]*)"$`, w.tabIs)
	sc.Then(`^every card in "([^"]*)" is the same height$`, w.sameHeight)
	sc.Then(`^"([^"]*)" carries a card whose title ends in "([^"]*)"$`, w.titleEndsIn)
	sc.Then(`^it says the window is too small$`, w.tooSmall)
	sc.Then(`^it says what size it needs$`, w.saysWhatItNeeds)
	sc.Then(`^the card "([^"]*)" is chosen$`, w.isChosen)
	sc.Then(`^the card "([^"]*)" is not chosen$`, w.isNotChosen)
	sc.Then(`^"([^"]*)" is open to be edited$`, w.detailOpenOn)
	sc.Then(`^the form shows "([^"]*)"$`, w.detailShows)
	sc.Then(`^the work beside the form carries "([^"]*)"$`, w.detailLists)
	sc.Then(`^the work beside the form does not carry "([^"]*)"$`, w.detailDoesNotList)
	sc.Then(`^what hangs under it is drawn one line to a tile$`, w.servingCardsAreCards)
	sc.Then(`^the column beside the form is headed "([^"]*)"$`, w.rightColumnHead)
	sc.Then(`^the work beside the form says nothing serves it yet$`, w.detailSaysNothingServes)
	sc.Then(`^"([^"]*)" is scrolled$`, w.isScrolled)
	sc.Then(`^"([^"]*)" is not scrolled$`, w.isNotScrolled)
	sc.Then(`^no field of the form reaches what it offers$`, w.formFitsAboveItsOffers)
	sc.Then(`^no two fields of the form overlap$`, w.formFieldsApart)
	sc.Then(`^the title runs the width of the form, above everything else$`, func() error {
		return w.acrossTheForm("title", nil)
	})
	sc.Then(`^the description runs the width of the form, under the title$`, func() error {
		return w.acrossTheForm("description", []string{"title"})
	})
	sc.Then(`^the form groups its fields under (.+)$`, func(list string) error {
		var got []string
		for _, head := range FormHeads(w.w.formBox(), w.w.Form().Fields()) {
			got = append(got, head.Said)
		}
		if want := quoted(list); strings.Join(got, "|") != strings.Join(want, "|") {
			return fmt.Errorf("the form groups its fields under %v, want %v", got, want)
		}
		return nil
	})
	sc.Then(`^the parts are drawn as a list to tick, "([^"]*)" first$`, func(title string) error {
		if parts := w.w.PartsTicked(); len(parts) == 0 || parts[0].Title != title {
			return fmt.Errorf("the parts are drawn as %v, and %q is not first", parts, title)
		}
		return nil
	})
	sc.Then(`^the parts are drawn as a list to tick, "([^"]*)" among them$`, func(title string) error {
		for _, c := range w.w.PartsTicked() {
			if c.Title == title {
				return nil
			}
		}
		return fmt.Errorf("the parts are drawn as %v, and %q is not among them", w.w.PartsTicked(), title)
	})
	sc.Then(`^the history begins with the last thing done to task (\d+)$`, func(n int) error {
		shown, err := w.held.Ask(engine.TraceShow, reading.Numbered(n))
		if err != nil {
			return err
		}
		last, lines := line(shown.Acts[len(shown.Acts)-1]), w.w.Lines()
		if len(lines) == 0 || !strings.HasPrefix(last, lines[0]) {
			return fmt.Errorf("the history begins %q, and the last thing done was %q", lines, last)
		}
		return nil
	})
	sc.Then(`^no line of the history is empty$`, func() error {
		for i, said := range w.w.Lines() {
			if strings.TrimSpace(said) == "" {
				return fmt.Errorf("line %d of the history is empty: %q", i, w.w.Lines())
			}
		}
		return nil
	})
	sc.Then(`^every field of the form stands left of the history$`, func() error {
		history := BesideBox(Over(w.w.width, w.w.height))
		_, frames := w.formFrames()
		for i, frame := range frames {
			if frame.X+frame.W > history.X {
				return fmt.Errorf("%q runs to %d and the history starts at %d", w.w.Form().Fields()[i].Name, frame.X+frame.W, history.X)
			}
		}
		return nil
	})
	sc.Then(`^the work beside the form reads "([^"]*)"$`, func(want string) error {
		var got []string
		for _, c := range w.w.Under() {
			got = append(got, c.Title)
		}
		if strings.Join(got, "|") != want {
			return fmt.Errorf("the work beside the form reads %q, want %q", got, want)
		}
		return nil
	})
	sc.Then(`^task (\d+) hangs under task (\d+), titled "([^"]*)"$`, func(n, parent int, want string) error {
		found, err := w.givens().Tasks()
		if err != nil {
			return err
		}
		for _, t := range found {
			if shape := task.OfTask(t); shape.Number == n {
				if shape.Parent != parent || shape.Title != want {
					return fmt.Errorf("task %d hangs under %d, titled %q", n, shape.Parent, shape.Title)
				}
				return nil
			}
		}
		return fmt.Errorf("there is no task %d", n)
	})
	sc.Then(`^task (\d+) is titled "([^"]*)"$`, func(n int, want string) error {
		found, err := w.givens().Tasks()
		if err != nil {
			return err
		}
		for _, t := range found {
			if shape := task.OfTask(t); shape.Number == n {
				if shape.Title != want {
					return fmt.Errorf("task %d is titled %q, want %q", n, shape.Title, want)
				}
				return nil
			}
		}
		return fmt.Errorf("there is no task %d", n)
	})
	sc.Then(`^the question is no bigger than what it asks$`, func() error {
		box, over := w.w.QuestionBox(), Over(w.w.width, w.w.height)
		if box.W >= over.W || box.H >= over.H/2 {
			return fmt.Errorf("the question is %v in a window whose panels are %v", box, over)
		}
		for i := range w.w.Offers() {
			if offer := PanelOfferAt(box, i); offer.X+offer.W > box.X+box.W || offer.Y+offer.H > box.Y+box.H {
				return fmt.Errorf("offer %d at %v runs out of the question %v", i, offer, box)
			}
		}
		return nil
	})
	sc.Then(`^the parts are at their beginning$`, func() error {
		if from := w.w.partsFrom(); from != 0 {
			return fmt.Errorf("the parts are shown from %d", from)
		}
		return nil
	})
	sc.Then(`^the parts have moved$`, func() error {
		if from := w.w.partsFrom(); from == w.was {
			return fmt.Errorf("the parts are still shown from %d", from)
		}
		return nil
	})
	sc.Then(`^the last part is in sight$`, func() error {
		parts, tiles := w.w.PartsTicked(), w.w.partTiles()
		if last := w.w.partsFrom() + len(tiles) - 2; last != len(parts)-1 {
			return fmt.Errorf("the last part in sight is %d of %d", last+1, len(parts))
		}
		return nil
	})
	sc.Then(`^a part is in hand, under the pointer$`, func() error {
		ghost, ok := w.w.PartGhost()
		if !ok || !ghost.In(w.pointer) {
			return fmt.Errorf("the part in hand is drawn at %v (%v), and the pointer is at %v", ghost, ok, w.pointer)
		}
		return nil
	})
	sc.Then(`^it would land below "([^"]*)"$`, func(title string) error {
		line, ok := w.w.PartLanding()
		at := slices.IndexFunc(w.w.PartsTicked(), func(c reading.Card) bool { return c.Title == title })
		tile := w.w.partTiles()[at-w.w.partsFrom()]
		if !ok || line.Y != tile.Y+tile.H-landingHigh {
			return fmt.Errorf("it would land at %v (%v), and %q ends at %d", line, ok, title, tile.Y+tile.H)
		}
		return nil
	})
	sc.Then(`^the part the keyboard is on is "([^"]*)"$`, func(want string) error {
		if got, on := w.w.PartOn(); !on || got != want {
			return fmt.Errorf("the keyboard is on the part %q (%v), want %q", got, on, want)
		}
		return nil
	})
	sc.Then(`^the part "([^"]*)" is ticked$`, func(title string) error {
		for _, c := range w.w.PartsTicked() {
			if c.Title == title && reading.Finished(c) {
				return nil
			}
		}
		return fmt.Errorf("the parts are %+v, and %q is not ticked", w.w.PartsTicked(), title)
	})
	sc.Then(`^the part "([^"]*)" is struck out$`, func(title string) error {
		if !slices.ContainsFunc(w.w.PartsGone(), func(p reading.Part) bool { return p.Title == title }) {
			return fmt.Errorf("struck out are %+v, and not %q", w.w.PartsGone(), title)
		}
		return nil
	})
	sc.Then(`^the part "([^"]*)" is not struck out$`, func(title string) error {
		if slices.ContainsFunc(w.w.PartsGone(), func(p reading.Part) bool { return p.Title == title }) {
			return fmt.Errorf("%q is struck out", title)
		}
		return nil
	})
	sc.Then(`^the cards taken together are (.+)$`, func(list string) error {
		var got []string
		for _, n := range w.w.Taken() {
			if c, found := w.w.Reading().Card(n); found {
				got = append(got, c.Title)
			}
		}
		if want := quoted(list); strings.Join(got, "|") != strings.Join(want, "|") {
			return fmt.Errorf("the cards taken together are %v, want %v", got, want)
		}
		return nil
	})
	sc.Then(`^no card is taken with another$`, func() error {
		if taken := w.w.Taken(); len(taken) > 1 {
			return fmt.Errorf("%v are taken together", taken)
		}
		return nil
	})
	sc.Then(`^the question says of (task|objective) (\d+) "([^"]*)"$`, func(kind string, n int, want string) error {
		named := kind + " " + strconv.Itoa(n)
		for _, line := range w.w.Asks() {
			if line.Key == named && strings.Contains(line.Means, want) {
				return nil
			}
		}
		return fmt.Errorf("the question says %+v, and nothing of %s saying %q", w.w.Asks(), named, want)
	})
	sc.Then(`^the bar under the board reads (.+)$`, func(list string) error {
		if got, want := w.w.BoardOffers(), quoted(list); strings.Join(got, "|") != strings.Join(want, "|") {
			return fmt.Errorf("the bar under the board reads %v, want %v", got, want)
		}
		return nil
	})
	sc.Then(`^the bar under the board answers nothing$`, func() error {
		for i := range w.w.BoardOffers() {
			box := OfferAt(w.w.Layout(), i)
			if offer, on := w.w.offerAt(Point{X: box.X + box.W/2, Y: box.Y + box.H/2}); on {
				return fmt.Errorf("the bar under the board answers %q under a panel", offer)
			}
		}
		return nil
	})
	sc.Then(`^the lane "([^"]*)" is scrolled$`, w.laneIsScrolled)
	sc.Then(`^the lane "([^"]*)" is not scrolled$`, w.laneIsNotScrolled)
	sc.Then(`^nothing was asked of the engine$`, w.askedNothing)
	sc.Then(`^the engine was asked "([^"]*)" for task (\d+)$`, w.wasAsked)
	sc.Then(`^the engine was not asked "([^"]*)"$`, w.wasNotAsked)
	sc.Step(`^I triple-click the first word of "([^"]*)" on the open task$`, func(key string) error {
		at, err := w.firstWordOf(key)
		if err == nil {
			w.w = w.w.tripleClicked(at)
		}
		return err
	})
	sc.Step(`^I double-click the first word of "([^"]*)" on the open task$`, func(key string) error {
		at, err := w.firstWordOf(key)
		if err == nil {
			w.w = w.w.DoubleClick(at)
		}
		return err
	})
	sc.Then(`^the box under the task is headed "([^"]*)"$`, func(want string) error {
		if got := w.w.UnderHead(); got != want {
			return fmt.Errorf("the box under the task is headed %q, want %q", got, want)
		}
		return nil
	})
	sc.Then(`^what hangs under it reads (.+)$`, func(said string) error {
		var want []string
		for _, one := range strings.Split(said, ", ") {
			want = append(want, strings.Trim(one, `"`))
		}
		var got []string
		for _, c := range w.w.Under() {
			got = append(got, c.Title)
		}
		if !slices.Equal(got, want) {
			return fmt.Errorf("what hangs under it reads %v, want %v", got, want)
		}
		return nil
	})
	sc.Then(`^the engine was asked "([^"]*)" once$`, w.wasAskedOnce)
	sc.Then(`^task (\d+) is in the column "([^"]*)"$`, w.taskInColumn)
	sc.Then(`^task (\d+) is called "([^"]*)"$`, w.taskCalled)
	sc.Then(`^the form asks about (".*")$`, w.formAsksAbout)
	sc.Then(`^the field "([^"]*)" reads "([^"]*)"$`, w.fieldReads)
	sc.Then(`^the store was read once$`, w.readOnce)
	sc.Then(`^it says what went wrong$`, w.saysWhatWentWrong)
	sc.Then(`^it says nothing went wrong$`, func() error {
		if wrong := w.w.Wrong(); wrong != nil {
			return fmt.Errorf("the window still says: %w", wrong)
		}
		return nil
	})
	sc.Then(`^what went wrong is drawn clear of the bar$`, func() error {
		l := w.w.Layout()
		said, bar := WrongBox(l), l.Bar
		if said.Y < bar.Y+bar.H && bar.Y < said.Y+said.H {
			return fmt.Errorf("what went wrong, at %+v, lies over the bar at %+v", said, bar)
		}
		if len(w.w.Offers()) == 0 {
			return fmt.Errorf("the bar offers nothing while something went wrong")
		}
		return nil
	})
	sc.Step(`^the window has looked at the store (\d+) times?$`, func(n int) error {
		for range n {
			w.w = w.w.Wonder()
		}
		return nil
	})
	sc.Then(`^the form shows the field "([^"]*)"$`, w.detailShowsField)
	sc.Then(`^the open task says where the context of task (\d+) lives$`, func(n int) error {
		want := filepath.Join(notesRoot, strconv.Itoa(n))
		// Written where it can be held and copied, like every text it writes.
		for _, r := range w.w.Readables() {
			if r.Key == "field:context" {
				if r.Said != want {
					return fmt.Errorf("the open task says its context lives at %q, want %q", r.Said, want)
				}
				return nil
			}
		}
		return fmt.Errorf("the open task says nothing of where its context lives")
	})
	sc.Then(`^the form asks nothing about "([^"]*)"$`, func(name string) error {
		for _, field := range w.w.Form().Fields() {
			if field.Name == name {
				return fmt.Errorf("the form asks about %q", name)
			}
		}
		return nil
	})
	sc.Then(`^the form shows the field "([^"]*)" reading "([^"]*)"$`, w.detailFieldReads)
	sc.Then(`^the card being dragged follows the pointer$`, w.ghostFollows)
	sc.Then(`^the tile being dragged follows the pointer$`, w.ghostFollows)
	// Inside the ghost and not merely somewhere: a ghost drawn with a card's
	// layout put its title 28 pixels down, which is past the bottom of a tile
	// — the writing landed on whatever was behind it instead.
	sc.Then(`^what is in hand says what it carries$`, func() error {
		ghost, dragging := w.w.Ghost()
		if !dragging {
			return fmt.Errorf("nothing is in hand")
		}
		// Strictly inside its own outline, which would otherwise be what is
		// counted — measured, it was, and this passed with the writing landing
		// entirely outside the ghost.
		inside := Box{X: ghost.X + 2, Y: ghost.Y + 2, W: ghost.W - 4, H: ghost.H - 4}
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		if inked(shot, inside, Dark().Ghost) == 0 {
			return fmt.Errorf("nothing is written on what is in hand, at %+v", ghost)
		}
		return nil
	})

	sc.Then(`^it shows where the tile would land$`, func() error {
		landing, somewhere := w.w.Landing()
		if !somewhere {
			return fmt.Errorf("nothing says where the tile would land")
		}
		box := BesideBox(Over(w.w.width, w.w.height))
		if landing.X != box.X || landing.W != box.W {
			return fmt.Errorf("it lands at %+v, and the box is %+v", landing, box)
		}
		return nil
	})
	// The board is still drawn behind the panel, and it was still being asked
	// where a point over it would land: a blue line appeared under the panel
	// for a gesture that had nothing to do with the board.
	sc.Then(`^the board behind shows no landing$`, func() error {
		landing, somewhere := w.w.Landing()
		if !somewhere {
			return nil
		}
		for i := range Columns {
			if w.w.Layout().Column[i].X == landing.X {
				return fmt.Errorf("the landing at %+v is column %d's, not the box's", landing, i)
			}
		}
		return nil
	})
	sc.Then(`^it shows where the card would land in "([^"]*)"$`, w.landsIn)
	sc.Then(`^it shows nowhere for the card to land$`, w.landsNowhere)
	sc.Then(`^"([^"]*)" does not carry "([^"]*)"$`, w.columnDoesNotCarry)
	sc.Then(`^the help is open$`, w.helpOpen)
	sc.Then(`^the help is closed$`, w.helpClosed)
	sc.Then(`^the help says what "([^"]*)" does$`, w.helpSaysAbout)
	sc.Then(`^the help shows some of its lines and not all$`, func() error {
		if shown, all := w.w.HelpRoom(), len(w.w.HelpRows()); shown < 1 || shown >= all {
			return fmt.Errorf("the help shows %d of its %d lines", shown, all)
		}
		return nil
	})
	sc.Then(`^the help shows every line$`, func() error {
		if shown, all := w.w.HelpRoom(), len(w.w.HelpRows()); shown < all {
			return fmt.Errorf("the help shows %d of its %d lines", shown, all)
		}
		return nil
	})
	sc.Then(`^the help draws a bar to scroll it by$`, func() error {
		if _, ok := w.w.HelpBar(); !ok {
			return fmt.Errorf("the help draws no bar")
		}
		return nil
	})
	sc.Then(`^the help starts (\d+) lines down$`, func(n int) error {
		if got := w.w.HelpFrom(); got != n {
			return fmt.Errorf("the help starts %d lines down, want %d", got, n)
		}
		return nil
	})
	sc.Then(`^the help shows its last line$`, func() error {
		if last := w.w.HelpFrom() + w.w.HelpRoom(); last != len(w.w.HelpRows()) {
			return fmt.Errorf("the help ends at line %d of %d", last, len(w.w.HelpRows()))
		}
		return nil
	})
	sc.Then(`^the help is no taller than what it says$`, func() error {
		box, over := w.w.HelpBox(), Over(w.width, w.high)
		if box.H >= over.H {
			return fmt.Errorf("the help is %d high, as tall as the room it was given", box.H)
		}
		if _, ok := w.w.HelpBar(); ok {
			return fmt.Errorf("the help draws a bar with nothing to scroll")
		}
		return nil
	})
	sc.Then(`^every line of the help fits inside its panel$`, func() error {
		box := w.w.HelpBox()
		for i, row := range w.w.HelpRows() {
			_, value := HelpRowAt(box, i)
			if wide := Reckoned(row.Means, reads); wide[len(wide)-1] > value.W {
				return fmt.Errorf("line %d, %q, is %d wide in a column of %d", i, row.Means, wide[len(wide)-1], value.W)
			}
		}
		return nil
	})
	sc.Then(`^the bar does not offer "([^"]*)"$`, func(offer string) error {
		for _, on := range w.w.Offers() {
			if on == offer {
				return fmt.Errorf("the bar offers %q: %v", offer, w.w.Offers())
			}
		}
		return nil
	})
	sc.Then(`^the field "([^"]*)" shows a caret$`, w.fieldShowsACaret)
	sc.Then(`^no other field shows a caret$`, w.noOtherFieldShowsACaret)
	sc.Then(`^the field "([^"]*)" has room for more than one line$`, w.fieldHasRoomForMoreThanOneLine)
	sc.Then(`^the field "([^"]*)" has room for one$`, w.fieldHasRoomForOne)
	sc.Then(`^the field "([^"]*)" shows the end of what was typed$`, w.fieldShowsTheEnd)
	sc.Then(`^the objectives are laid out more than one to a row$`, w.objectivesAcross)
	sc.Then(`^every objective is drawn inside the board$`, w.objectivesInsideTheBoard)
	sc.Then(`^the objectives all fit without scrolling$`, w.objectivesAllFit)
	sc.Then(`^what is chosen has moved on one$`, w.chosenMovedOnOne)
	sc.Then(`^what is chosen has moved down one row$`, w.chosenMovedDownARow)
	sc.Then(`^the card "([^"]*)" says it is being worked on$`, w.cardSaysInFocus)
	sc.Then(`^the card "([^"]*)" says "([^"]*)"$`, w.cardSays)
	sc.Then(`^the card "([^"]*)" says nothing of "([^"]*)"$`, func(title, said string) error {
		if err := w.cardSays(title, said); err == nil {
			return fmt.Errorf("the card for %q says %q, and should not", title, said)
		}
		return nil
	})
	sc.Then(`^what is chosen is in sight$`, w.chosenInSight)
	sc.Then(`^the objectives have moved$`, w.objectivesMoved)
	sc.Then(`^the caret is on the second field$`, w.caretOnTheSecondField)
	sc.Then(`^the tab "([^"]*)" says it holds (\d+)$`, w.tabHolds)
	sc.Then(`^the form is open on task (\d+)$`, w.formOpenOnTask)

	// Asked before the one below it, and for the same reason it exists: that
	// one asks whether writing reaches the frame's edge, which is trivially
	// true of a tile with no writing on it at all. A measuring bug once cut
	// every title to a lone ellipsis and it passed.
	sc.Then(`^the second tile says what it holds$`, func() error {
		tile := w.w.ServingCardAt(1)
		line := Row(Box{X: tile.X, Y: tile.Y, W: tile.W, H: tile.H - between}, tile.Y+inset, lineHigh)
		// Strictly inside the tile's own border, which runs its whole width
		// and would ink every column asked about. Measured — it did, and this
		// read as a full line of writing where there was a lone ellipsis.
		band := Box{X: line.X, Y: line.Y, W: line.W, H: line.H - inset}
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		var reached int
		for x := band.X; x < band.X+band.W; x++ {
			if inked(shot, Box{X: x, Y: band.Y, W: 1, H: band.H}, Dark().Panel) > 0 {
				reached = x - band.X
			}
		}
		if reached < line.W/2 {
			return fmt.Errorf("the tile writes across %d of its %d: a long title fills its line", reached, line.W)
		}
		return nil
	})

	// Measured in pixels at the tile's right edge: a title cut by counting an
	// average glyph width ran all the way there and was clipped, which reads
	// as a word chopped against the scrollbar.
	sc.Then(`^the second tile leaves room before its frame$`, func() error {
		// The writing's own colour, and not everything that is not the tile:
		// the tile the keyboard is on is outlined, and its outline runs down
		// the very edge this asks about. Measured — it counted that.
		tile := w.w.ServingCardAt(1)
		edge := Box{X: tile.X + tile.W - padding, Y: tile.Y, W: padding, H: tile.H}
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		for what, ink := range map[string]color.NRGBA{"faint": Dark().Faint, "ink": Dark().Ink} {
			if got := painted(shot, edge, ink); got != 0 {
				return fmt.Errorf("%d pixels of %s writing reach the tile's last %d", got, what, padding)
			}
		}
		return nil
	})

	sc.Then(`^the search is not in force$`, func() error {
		if !w.w.Searched().Empty() {
			return fmt.Errorf("the board is still narrowed to %q", w.w.Looking())
		}
		return nil
	})

	sc.Then(`^the list of objectives is open$`, w.listOpen)
	sc.Then(`^the form is still drawn behind it$`, w.formBehindTheList)
	sc.Then(`^the field "([^"]*)" is drawn by name and not by number$`, w.fieldDrawnByName)
	sc.Then(`^the list offers "([^"]*)"$`, w.listOffers)
	sc.Then(`^the label "([^"]*)" is (not )?ticked$`, func(label, not string) error {
		for _, one := range w.w.Choosing() {
			if one.Title == label {
				if w.w.Ticked(one.Number) == (not == "") {
					return nil
				}
				return fmt.Errorf("the label %q ticked is %v", label, not != "")
			}
		}
		return fmt.Errorf("the list offers %v, and not %q", goalTitles(w.w.Choosing()), label)
	})
	sc.Then(`^the list offers to create "([^"]*)"$`, func(label string) error {
		for _, one := range w.w.Choosing() {
			if one.Creates == label {
				return nil
			}
		}
		return fmt.Errorf("the list offers %v, and does not offer to create %q", goalTitles(w.w.Choosing()), label)
	})
	sc.Then(`^the list offers to create nothing$`, func() error {
		for _, one := range w.w.Choosing() {
			if one.Creates != "" {
				return fmt.Errorf("the list offers to create %q", one.Creates)
			}
		}
		return nil
	})
	sc.Then(`^the list says a label has no family$`, func() error {
		for _, one := range w.w.Choosing() {
			if strings.Contains(one.Title, "a label has no family") {
				return nil
			}
		}
		return fmt.Errorf("the list reads %v, saying nothing of a family", goalTitles(w.w.Choosing()))
	})
	sc.Then(`^task (\d+) carries the labels "([^"]*)"$`, func(n int, want string) error {
		shape, err := w.tell(n)
		if err != nil {
			return err
		}
		if got := strings.Join(shape.Labels, ", "); got != want {
			return fmt.Errorf("task %d carries %q, want %q", n, got, want)
		}
		return nil
	})
	sc.Then(`^the list reads (".*")$`, func(said string) error {
		var want []string
		for _, one := range strings.Split(said, ",") {
			want = append(want, strings.Trim(strings.TrimSpace(one), `"`))
		}
		got := goalTitles(w.w.Choosing())
		if len(got) != len(want) {
			return fmt.Errorf("the list reads %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				return fmt.Errorf("the list reads %v, want %v", got, want)
			}
		}
		return nil
	})
	sc.Then(`^the row "([^"]*)" says "([^"]*)"$`, func(title, said string) error {
		for _, one := range w.w.Choosing() {
			if one.Title != title {
				continue
			}
			if one.State != said {
				return fmt.Errorf("the row for %q says %q, not %q", title, one.State, said)
			}
			return nil
		}
		return fmt.Errorf("the list offers no row called %q", title)
	})
	sc.Then(`^the row the keyboard is on is "([^"]*)"$`, func(title string) error {
		offered, at := w.w.Choosing(), w.w.ChoosingAt()
		if at >= len(offered) {
			return fmt.Errorf("the keyboard is on row %d of %d", at, len(offered))
		}
		if offered[at].Title != title {
			return fmt.Errorf("the keyboard is on %q, not %q", offered[at].Title, title)
		}
		return nil
	})
	sc.Then(`^the list offers nothing$`, func() error {
		if got := w.w.Choosing(); len(got) != 0 {
			return fmt.Errorf("the list offers %v", goalTitles(got))
		}
		return nil
	})
	sc.Then(`^the list has taken (\d+)$`, func(n int) error {
		taken := 0
		for _, one := range w.w.Choosing() {
			if w.w.Ticked(one.Number) {
				taken++
			}
		}
		if taken != n {
			return fmt.Errorf("the list has taken %d rows, not %d", taken, n)
		}
		return nil
	})
	sc.Then(`^the list does not offer "([^"]*)"$`, func(title string) error {
		for _, one := range w.w.Choosing() {
			if one.Title == title {
				return fmt.Errorf("the list still offers %q", title)
			}
		}
		return nil
	})
	sc.Then(`^the list has more to show than it has room for$`, func() error {
		if len(w.w.Choosing()) <= w.w.ChoosingRoom() {
			return fmt.Errorf("the list offers %d rows and shows %d",
				len(w.w.Choosing()), w.w.ChoosingRoom())
		}
		return nil
	})
	sc.Then(`^the list is at its beginning$`, func() error {
		if w.w.ChoosingFromRow() != 0 {
			return fmt.Errorf("the list starts at row %d", w.w.ChoosingFromRow())
		}
		return nil
	})
	sc.Then(`^the list is at its end$`, func() error {
		want := max(0, len(w.w.Choosing())-w.w.ChoosingRoom())
		if w.w.ChoosingFromRow() != want {
			return fmt.Errorf("the list starts at row %d and its end is %d", w.w.ChoosingFromRow(), want)
		}
		return nil
	})
	sc.Then(`^the list has moved$`, func() error {
		if w.w.ChoosingFromRow() == w.was {
			return fmt.Errorf("the list stayed at row %d", w.w.ChoosingFromRow())
		}
		return nil
	})
	sc.Then(`^every row of the list is inside the panel$`, w.listInsideThePanel)
	sc.Then(`^a bar sits beside the list$`, func() error {
		box := ChoosingOver(w.w.width, w.w.height)
		run, track := ChoosingRun(box), ChoosingBar(box)
		if track.W <= 0 || track.H <= 0 {
			return fmt.Errorf("the bar beside the list is %dx%d", track.W, track.H)
		}
		if track.X < run.X+run.W {
			return fmt.Errorf("the bar at %d sits over the rows, which end at %d", track.X, run.X+run.W)
		}
		return nil
	})
	sc.Then(`^the row the keyboard is on is in sight$`, func() error {
		at, from, room := w.w.ChoosingAt(), w.w.ChoosingFromRow(), w.w.ChoosingRoom()
		if at < from || at >= from+room {
			return fmt.Errorf("the keyboard is on row %d, and rows %d to %d are in sight",
				at, from, from+room-1)
		}
		return nil
	})
	sc.Then(`^the window is choosing$`, func() error {
		if w.w.Screen() != OnChoosing {
			return fmt.Errorf("the window is on %v, not the list", w.w.Screen())
		}
		return nil
	})
	sc.Then(`^the form is still on the field "([^"]*)"$`, func(name string) error {
		if got := w.w.Form().Fields()[w.w.Form().On()].Name; got != name {
			return fmt.Errorf("the keyboard is on the field %q, not %q", got, name)
		}
		return nil
	})
	sc.Then(`^the clipboard reads "([^"]*)"$`, func(said string) error {
		if w.clipboard != said {
			return fmt.Errorf("the clipboard reads %q, not %q", w.clipboard, said)
		}
		return nil
	})
	sc.Then(`^the field "([^"]*)" holds "([^"]*)"$`, func(name, said string) error {
		for _, f := range w.w.Form().Fields() {
			if f.Name != name {
				continue
			}
			if got := w.w.Form().Writing().Holds(); got != said {
				return fmt.Errorf("the field %q holds %q, not %q", name, got, said)
			}
			return nil
		}
		return fmt.Errorf("the form has no field %q", name)
	})
	sc.Then(`^the window is on the board$`, func() error {
		if w.w.Screen() != OnTheBoard {
			return fmt.Errorf("the window is on %v, not the board", w.w.Screen())
		}
		return nil
	})
	sc.Then(`^the window is on the form$`, func() error {
		if w.w.Screen() != OnForm {
			return fmt.Errorf("the window is on %v, not the form", w.w.Screen())
		}
		return nil
	})
	sc.Then(`^the list is on "([^"]*)"$`, w.listOn)
	sc.Then(`^objective (\d+) is ticked$`, func(n int) error { return w.ticked(n, true) })
	sc.Then(`^objective (\d+) is not ticked$`, func(n int) error { return w.ticked(n, false) })
	sc.Then(`^the field "([^"]*)" names "([^"]*)"$`, w.fieldNames)
	sc.Then(`^the list says there is nothing to serve yet$`, w.listSaysNothing)
	sc.Then(`^nothing has been looked for yet$`, func() error {
		if said := w.w.Looking(); said != "" {
			return fmt.Errorf("the search box already reads %q", said)
		}
		return nil
	})
	sc.Then(`^every key the bar offers does what it offers$`, w.everyKeyTheBarOffers)
	sc.Then(`^the history is beside the fields$`, w.historyBesideTheFields)
	sc.Then(`^the history stays inside the panel$`, w.historyInsideThePanel)
	sc.Then(`^the history has more to show than it has room for$`, w.historyHasMore)
	sc.Then(`^the history has moved$`, w.historyMoved)
	sc.Then(`^the work under it has more to show than it has room for$`, w.servingHasMore)
	sc.Then(`^the work under it has moved$`, w.servingMoved)
	sc.Then(`^the work under it is at its beginning$`, w.servingAtTheBeginning)
	sc.Then(`^the work under it is at its end$`, w.servingAtTheEnd)
	sc.Then(`^the description is at its beginning$`, func() error {
		if w.w.saying != 0 {
			return fmt.Errorf("the description is at line %d, not its beginning", w.w.saying)
		}
		return nil
	})
	sc.Then(`^the description is at its end$`, func() error {
		// Asked first, because the end of a description nobody is showing is
		// zero — and zero is where an untouched window already sits, so the
		// question would answer yes to a gesture that did nothing at all.
		if _, _, _, _, long := w.w.sayingOn(); !long {
			return fmt.Errorf("no description is open: the window is on %v", w.w.Screen())
		}
		want := max(0, len(w.w.SaidLines())-w.w.SaidRows())
		if w.w.saying != want {
			return fmt.Errorf("the description is at line %d and its end is %d", w.w.saying, want)
		}
		return nil
	})
	sc.Then(`^the description has more to show than it has room for$`, func() error {
		if len(w.w.SaidLines()) <= w.w.SaidRows() {
			return fmt.Errorf("the description is %d lines and its box holds %d",
				len(w.w.SaidLines()), w.w.SaidRows())
		}
		return nil
	})
	sc.Then(`^the description has moved$`, func() error {
		if w.w.saying == w.was {
			return fmt.Errorf("the description stayed at line %d", w.w.saying)
		}
		return nil
	})
	sc.Then(`^a bar sits beside the description$`, func() error {
		frame, track, _, _, long := w.w.sayingOn()
		if !long {
			return fmt.Errorf("the open task shows no description")
		}
		if track.W <= 0 || track.H <= 0 {
			return fmt.Errorf("the bar beside the description is %dx%d", track.W, track.H)
		}
		if track.X < frame.X+frame.W {
			return fmt.Errorf("the bar at %d sits over the description, which ends at %d",
				track.X, frame.X+frame.W)
		}
		return nil
	})
	sc.Then(`^the history is at its beginning$`, w.historyAtTheBeginning)
	sc.Then(`^the history is at its end$`, w.historyAtTheEnd)
	sc.Then(`^the field "([^"]*)" is written on more than one line$`, w.fieldOnMoreThanOneLine)
	sc.Then(`^no line of the field "([^"]*)" runs out of it$`, w.noLineRunsOut)
	sc.Then(`^the bar offers "([^"]*)" under "([^"]*)"$`, w.barOffersUnder)
	sc.Then(`^the window is asking what to look for$`, w.askingWhatToLookFor)
	// Asked of every card the board deals, columns and band alike: "not in
	// BACKLOG" would pass for a card that merely sits in another column.
	sc.Then(`^the board says nothing of "([^"]*)"$`, func(title string) error {
		dealt, stalled := reading.Deal(w.w.Showing().Cards, w.w.archived)
		for _, held := range dealt {
			for _, c := range held.Cards {
				if c.Title == title {
					return fmt.Errorf("the board deals %q into %q", title, held.Name)
				}
			}
		}
		for _, c := range stalled {
			if c.Title == title {
				return fmt.Errorf("the board deals %q into the band", title)
			}
		}
		return nil
	})

	sc.Then(`^no card is chosen$`, func() error {
		if w.w.Chosen() != 0 {
			return fmt.Errorf("task %d is chosen", w.w.Chosen())
		}
		return nil
	})
	sc.Then(`^the help is headed "([^"]*)"$`, w.helpHeaded)
	sc.Then(`^the panel is headed "([^"]*)"$`, w.helpHeaded)
	sc.Then(`^the search box is open$`, w.searchBoxOpen)
	sc.Then(`^the search box is closed$`, w.searchBoxClosed)
	sc.Then(`^the search is still in force$`, w.searchStillInForce)
	sc.Then(`^the caret is at (\d+)$`, w.caretIsAt)
	sc.Then(`^the caret is on line (\d+) of the field "([^"]*)"$`,
		func(want int, name string) error {
			if w.w.Screen() != OnForm {
				return fmt.Errorf("nothing is being typed into")
			}
			fields := w.w.Form().Fields()
			at := w.w.Form().On()
			if fields[at].Name != name {
				return fmt.Errorf("the keyboard is on %q, not on %q", fields[at].Name, name)
			}
			said := w.w.Form().Writing()
			_, _, lines, _, _ := w.w.sayingOn()
			line, _ := lines.At(said.At)
			if line+1 != want {
				return fmt.Errorf("the caret is on line %d of %q, want line %d", line+1, name, want)
			}
			return nil
		})
	sc.Then(`^the search reads "([^"]*)"$`, w.searchReads)
	sc.Then(`^the keyboard is on "([^"]*)"$`, w.keyboardIsOn)
	sc.Then(`^the objectives read (".*")$`, w.objectivesRead)
	sc.Then(`^the objective "([^"]*)" is chosen$`, w.objectiveChosen)
	sc.Then(`^the card for objective (\d+) says "([^"]*)"$`, w.goalCardSays)
	sc.Then(`^the card for objective (\d+) says nothing of how far along it is$`, w.goalCardSaysNoCount)
	sc.Then(`^objective (\d+) is called "([^"]*)"$`, w.objectiveCalled)
}

func (w *world) objectivesRead(list string) error {
	var got []string
	for _, g := range reading.Sorted(w.w.Reading().Goals) {
		got = append(got, g.Title)
	}
	want := quoted(list)
	if strings.Join(got, "|") != strings.Join(want, "|") {
		return fmt.Errorf("the objectives read %v, want %v", got, want)
	}
	return nil
}

// goalCardSays is asked of the line the objectives tab draws beside an
// objective's number: how much work it carries and how much of it is done.
func (w *world) goalCardSays(number int, text string) error {
	for _, g := range w.w.Reading().Goals {
		if g.Number == number {
			if said := Counted(g); !strings.Contains(said, text) {
				return fmt.Errorf("the card for objective %d says %q, and nothing of %q", number, said, text)
			}
			return nil
		}
	}
	return fmt.Errorf("there is no objective %d", number)
}

func (w *world) goalCardSaysNoCount(number int) error {
	for _, g := range w.w.Reading().Goals {
		if g.Number == number {
			if said := Counted(g); said != "" {
				return fmt.Errorf("the card for objective %d says %q, and nothing serves it", number, said)
			}
			return nil
		}
	}
	return fmt.Errorf("there is no objective %d", number)
}

func (w *world) objectiveChosen(title string) error {
	for _, g := range w.w.Reading().Goals {
		if g.Title == title {
			if w.w.Chosen() != g.Number {
				return fmt.Errorf("objective %d is chosen, not %q", w.w.Chosen(), title)
			}
			return nil
		}
	}
	return fmt.Errorf("no objective titled %q", title)
}

func (w *world) objectiveCalled(number int, title string) error {
	found, err := w.givens().Objective(number)
	if err != nil {
		return fmt.Errorf("objective %d: %w", number, err)
	}
	if got := objective.OfObjective(found).Title; got != title {
		return fmt.Errorf("objective %d is called %q, want %q", number, got, title)
	}
	return nil
}

func (w *world) detailShowsField(name string) error {
	for _, field := range w.w.Form().Fields() {
		if field.Name == name {
			return nil
		}
	}
	return fmt.Errorf("the detail does not show %q", name)
}

func (w *world) detailFieldReads(name, value string) error {
	for _, field := range w.w.Form().Fields() {
		if field.Name != name {
			continue
		}
		if field.Value != value {
			return fmt.Errorf("the detail shows %q as %q, want %q", name, field.Value, value)
		}
		return nil
	}
	return fmt.Errorf("the detail does not show %q", name)
}

func (w *world) ghostFollows() error {
	ghost, dragging := w.w.Ghost()
	if !dragging {
		return fmt.Errorf("no card is being dragged")
	}
	if !ghost.In(w.w.held.At) {
		return fmt.Errorf("the card at %+v does not hold the pointer at %+v", ghost, w.w.held.At)
	}
	return nil
}

// landsNowhere is the assertion that the window draws no landing line: the card
// is in hand and over a place it cannot take, so there is nothing true to draw.
func (w *world) landsNowhere() error {
	if landing, somewhere := w.w.Landing(); somewhere {
		return fmt.Errorf("the board says the card would land at %v, and it cannot land there", landing)
	}
	return nil
}

func (w *world) landsIn(column string) error {
	at := reading.IndexOf(strings.ToLower(column))
	if at < 0 {
		return fmt.Errorf("no column called %q", column)
	}
	landing, somewhere := w.w.Landing()
	if !somewhere {
		return fmt.Errorf("the board does not say where the card would land")
	}
	if landing.X != w.w.Layout().Column[at].X {
		return fmt.Errorf("it lands at %d and %q is at %d", landing.X, column, w.w.Layout().Column[at].X)
	}
	return nil
}

func (w *world) keyboardIsOn(name string) error {
	if w.w.Screen() != OnForm {
		return fmt.Errorf("no form is open")
	}
	got := w.w.Form().Fields()[w.w.Form().On()].Name
	if got != name {
		return fmt.Errorf("the keyboard is on %q, want %q", got, name)
	}
	return nil
}

func (w *world) columnDoesNotCarry(column, title string) error {
	if err := w.columnCarries(column, title); err == nil {
		return fmt.Errorf("%q carries %q and should not", column, title)
	}
	return nil
}

func (w *world) helpOpen() error {
	if w.w.Screen() != OnHelp {
		return fmt.Errorf("the window is on %v", w.w.Screen())
	}
	return nil
}

func (w *world) helpClosed() error {
	if w.w.Screen() == OnHelp {
		return fmt.Errorf("the help is still open")
	}
	return nil
}

func (w *world) helpSaysAbout(about string) error {
	for _, line := range w.w.Helping() {
		if strings.Contains(line.Key, about) || strings.Contains(line.Means, about) {
			return nil
		}
	}
	return fmt.Errorf("the help says nothing about %q", about)
}

func (w *world) fieldShowsACaret(name string) error {
	on := w.w.Form().Fields()[w.w.Form().On()].Name
	if on != name {
		return fmt.Errorf("the caret is on %q, not %q", on, name)
	}
	return nil
}

func (w *world) noOtherFieldShowsACaret() error {
	// One field carries the keyboard at a time, and the caret is drawn there
	// and nowhere else — so there is one caret by construction.
	if w.w.Screen() != OnForm {
		return fmt.Errorf("no form is open")
	}
	if on := w.w.Form().On(); on < 0 || on >= len(w.w.Form().Fields()) {
		return fmt.Errorf("the keyboard is on field %d, and there are %d", on, len(w.w.Form().Fields()))
	}
	return nil
}

func (w *world) fieldHasRoomForMoreThanOneLine(name string) error {
	rows, err := w.rowsOfField(name)
	if err != nil {
		return err
	}
	if rows <= 1 {
		return fmt.Errorf("%q has room for %d line", name, rows)
	}
	return nil
}

func (w *world) fieldHasRoomForOne(name string) error {
	rows, err := w.rowsOfField(name)
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("%q has room for %d lines", name, rows)
	}
	return nil
}

// rowsOfField is how many lines a field of the open form is drawn over.
func (w *world) rowsOfField(name string) (int, error) {
	fields := w.w.Form().Fields()
	for i := range fields {
		if fields[i].Name == name {
			return formRows(w.w.formBox(), fields, i), nil
		}
	}
	return 0, fmt.Errorf("the form has no field called %q", name)
}

func (w *world) fieldShowsTheEnd(name string) error {
	box := Over(w.w.Layout().Bar.W+2*margin, w.w.Layout().Bar.Y+barHigh+margin)
	fields := w.w.Form().Fields()
	for i, field := range fields {
		if field.Name != name {
			continue
		}
		value := formValue(box, fields, i)
		rows := formRows(box, fields, i)
		lines := Wrapped(field.Value, room(Inside(Framed(value, rows))))
		if len(lines.Said) < 2 {
			return fmt.Errorf("%q is not longer than its room, so this proves nothing", name)
		}
		line, _ := lines.At(w.w.Form().Writing().At)
		shown := lines.Said[lines.Window(rows, line)]
		if !strings.HasSuffix(field.Value, shown) {
			return fmt.Errorf("%q shows %q, which is not the end of %q", name, shown, field.Value)
		}
		return nil
	}
	return fmt.Errorf("the form has no field called %q", name)
}

func (w *world) caretIsAt(at int) error {
	if w.w.Screen() == OnSearch {
		if got := w.w.Searched().At; got != at {
			return fmt.Errorf("the caret is at %d in the search, want %d", got, at)
		}
		return nil
	}
	if w.w.Screen() != OnForm {
		return fmt.Errorf("nothing is being typed into")
	}
	if got := w.w.Form().Writing().At; got != at {
		return fmt.Errorf("the caret is at %d, want %d", got, at)
	}
	return nil
}

func (w *world) searchReads(value string) error {
	if got := w.w.Looking(); got != value {
		return fmt.Errorf("the search reads %q, want %q", got, value)
	}
	return nil
}

func (w *world) searchBoxOpen() error {
	if w.w.Screen() != OnSearch {
		return fmt.Errorf("the window is on %v", w.w.Screen())
	}
	return nil
}

func (w *world) searchBoxClosed() error {
	if w.w.Screen() == OnSearch {
		return fmt.Errorf("the search box is still open")
	}
	return nil
}

func (w *world) searchStillInForce() error {
	if w.w.Looking() == "" {
		return fmt.Errorf("nothing is being searched for any more")
	}
	return nil
}

func (w *world) barOffersUnder(offer, key string) error {
	for _, said := range w.w.Offers() {
		if said != offer {
			continue
		}
		if got := Pretty(Shortcut(offer)); got != key {
			return fmt.Errorf("%q is offered under %q, not %q", offer, got, key)
		}
		return nil
	}
	return fmt.Errorf("the bar does not offer %q", offer)
}

func (w *world) askingWhatToLookFor() error {
	if w.w.Screen() != OnSearch {
		return fmt.Errorf("the window is on %v", w.w.Screen())
	}
	return nil
}

func (w *world) helpHeaded(head string) error {
	if got := Heads(w.w); got != head {
		return fmt.Errorf("the help is headed %q, not %q", got, head)
	}
	return nil
}

// wrapping is a field's value broken into the lines it is drawn on, and the
// room one of those lines has.
func (w *world) wrapping(name string) (Lines, int, error) {
	fields := w.w.Form().Fields()
	for i, field := range fields {
		if field.Name != name {
			continue
		}
		box := w.w.formBox()
		value := formValue(box, fields, i)
		fits := room(Inside(Framed(value, formRows(box, fields, i))))
		return Wrapped(field.Value, fits), fits, nil
	}
	return Lines{}, 0, fmt.Errorf("the form has no field %q", name)
}

func (w *world) fieldOnMoreThanOneLine(name string) error {
	lines, _, err := w.wrapping(name)
	if err != nil {
		return err
	}
	if len(lines.Said) < 2 {
		return fmt.Errorf("%q is written on %d line, and what was typed is longer than one", name, len(lines.Said))
	}
	return nil
}

func (w *world) noLineRunsOut(name string) error {
	lines, fits, err := w.wrapping(name)
	if err != nil {
		return err
	}
	for i, line := range lines.Said {
		if len([]rune(line)) > fits {
			return fmt.Errorf("line %d of %q is %d characters in room for %d: %q", i, name, len([]rune(line)), fits, line)
		}
	}
	return nil
}

func (w *world) historyBesideTheFields() error {
	box := Over(w.w.width, w.w.height)
	fields, history := FormFields(box), BesideBox(box)
	if history.X < fields.X+fields.W {
		return fmt.Errorf("the history starts at %d and the fields end at %d: it is under them, not beside them",
			history.X, fields.X+fields.W)
	}
	return nil
}

func (w *world) historyInsideThePanel() error {
	box := Over(w.w.width, w.w.height)
	for what, part := range map[string]Box{"the history": BesideBox(box), "its bar": BesideBar(box)} {
		if part.X < box.X || part.X+part.W > box.X+box.W {
			return fmt.Errorf("%s runs from %d to %d in a panel from %d to %d",
				what, part.X, part.X+part.W, box.X, box.X+box.W)
		}
		if part.Y < box.Y || part.Y+part.H > PanelOfferAt(box, 0).Y {
			return fmt.Errorf("%s runs from %d to %d, and what the panel offers starts at %d",
				what, part.Y, part.Y+part.H, PanelOfferAt(box, 0).Y)
		}
	}
	// And nothing is shown past the end of what there is to show.
	if w.w.history+w.w.Rows() > w.w.Told()+w.w.Rows() {
		return fmt.Errorf("the history is scrolled to %d of %d lines", w.w.history, w.w.Told())
	}
	return nil
}

func (w *world) historyHasMore() error {
	if w.w.Told() <= w.w.Rows() {
		return fmt.Errorf("the history is %d lines in room for %d: it all fits, so this proves nothing",
			w.w.Told(), w.w.Rows())
	}
	return nil
}

func (w *world) servingHasMore() error {
	cards, rows := len(w.w.Under()), w.w.ServingRows()
	if cards <= rows {
		return fmt.Errorf("the box shows %d cards at once and there are %d: nothing is out of sight", rows, cards)
	}
	return nil
}

func (w *world) servingMoved() error {
	if w.w.serving == w.was {
		return fmt.Errorf("the work under it is still at %d", w.w.serving)
	}
	return nil
}

func (w *world) servingAtTheBeginning() error {
	if w.w.serving != 0 {
		return fmt.Errorf("the work under it is at %d, not at its beginning", w.w.serving)
	}
	return nil
}

func (w *world) servingAtTheEnd() error {
	box := BesideBox(Over(w.w.width, w.w.height))
	last := len(w.w.Under())*SubHeight - box.H
	if w.w.serving != last {
		return fmt.Errorf("the work under it is at %d and its end is %d", w.w.serving, last)
	}
	return nil
}

func (w *world) historyMoved() error {
	if w.w.history == w.was {
		return fmt.Errorf("the history is still at %d", w.w.history)
	}
	return nil
}

func (w *world) historyAtTheBeginning() error {
	if w.w.history != 0 {
		return fmt.Errorf("the history is at line %d, not its beginning", w.w.history)
	}
	return nil
}

func (w *world) historyAtTheEnd() error {
	if want := max(0, w.w.Told()-w.w.Rows()); w.w.history != want {
		return fmt.Errorf("the history is at line %d and its end is %d", w.w.history, want)
	}
	return nil
}

// formBehindTheList checks the list is an overlay and not a page of its own:
// the form's own panel is wider than the list's, and both are drawn.
func (w *world) formBehindTheList() error {
	over, list := Over(w.w.width, w.w.height), ChoosingOver(w.w.width, w.w.height)
	if list.W >= over.W || list.H >= over.H {
		return fmt.Errorf("the list is %dx%d over a form of %dx%d: it covers it rather than sitting on it",
			list.W, list.H, over.W, over.H)
	}
	// What makes it an overlay is that the form is still there underneath. The
	// first field's own frame lies outside the list's panel, so if the list
	// were a page of its own that frame would have been painted over.
	fields := w.w.Form().Fields()
	value := formValue(over, fields, 0)
	frame := Framed(value, formRows(over, fields, 0))
	if frame.Y+frame.H > list.Y {
		return fmt.Errorf("the first field reaches %d and the list starts at %d: they overlap, so this proves nothing",
			frame.Y+frame.H, list.Y)
	}
	// Its outline, and not its fill: the fill is the same colour as the ground
	// the whole window is painted on, so counting that would pass even with no
	// form drawn at all. Measured — it did.
	shot, err := drawnInAStep(w.w)
	if err != nil {
		return err
	}
	if painted(shot, frame, Dark().Edge) == 0 {
		return fmt.Errorf("the form's first field is gone: the list covered it rather than sitting on it")
	}
	return nil
}

// fieldDrawnByName checks the drawing and not the window's own answer: the
// field carries numbers and must be drawn as names, and a window that knows
// the names without putting them on screen has not answered.
func (w *world) fieldDrawnByName(field string) error {
	for _, f := range w.w.Form().Fields() {
		if f.Name != field {
			continue
		}
		said := w.w.FieldSays(f)
		if said == f.Value {
			return fmt.Errorf("the field %q is drawn as %q, which is what it carries", field, said)
		}
		// Text cannot be read back out of a screenshot, so what is measured is
		// reach: the names run far past where the numbers they stand for would
		// have stopped. Two numbers and a comma cannot fill a third of this
		// frame; two names do.
		fields := w.w.Form().Fields()
		at := slices.IndexFunc(fields, func(x reading.Field) bool { return x.Name == field })
		box := w.w.formBox()
		value := formValue(box, fields, at)
		frame := Framed(value, formRows(box, fields, at))
		// Past where the numbers themselves could have ended, and inside the
		// frame rather than on it: the frame's own outline is not the field's
		// writing, and a fraction of the width would only say the text is
		// long. What proves the names are drawn is ink beyond the reach of
		// what the field carries.
		inside := Inside(frame)
		from := inside.X + (len(f.Value)+2)*perGlyph
		far := Box{X: from, Y: inside.Y, W: inside.X + inside.W - from, H: inside.H}
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		if inked(shot, far, Dark().Back) == 0 {
			return fmt.Errorf("the field %q draws nothing past where %q would have ended", field, f.Value)
		}
		return nil
	}
	return fmt.Errorf("the form has no field %q", field)
}

func (w *world) listOpen() error {
	if w.w.Screen() != OnChoosing {
		return fmt.Errorf("the window is on %v", w.w.Screen())
	}
	return nil
}

func (w *world) listOffers(title string) error {
	for _, g := range w.w.Choosing() {
		if g.Title == title {
			return nil
		}
	}
	return fmt.Errorf("the list offers %v, and not %q", goalTitles(w.w.Choosing()), title)
}

func (w *world) listOn(title string) error {
	on := w.w.Choosing()
	at := w.w.ChoosingAt()
	if at < 0 || at >= len(on) {
		return fmt.Errorf("the list is on nothing, at %d of %d", at, len(on))
	}
	if on[at].Title != title {
		return fmt.Errorf("the list is on %q, not %q", on[at].Title, title)
	}
	return nil
}

func (w *world) ticked(number int, want bool) error {
	if got := w.w.Ticked(number); got != want {
		return fmt.Errorf("objective %d ticked = %v, want %v", number, got, want)
	}
	return nil
}

// fieldNames is asked of what the field says on screen, and not of what it
// carries: it carries the numbers the engine takes, and shows the names a
// person knows. Asking the value would be asking the wrong half.
func (w *world) fieldNames(field, title string) error {
	for _, f := range w.w.Form().Fields() {
		if f.Name != field {
			continue
		}
		if said := w.w.FieldSays(f); !strings.Contains(said, title) {
			return fmt.Errorf("the field %q says %q, and nothing of %q", field, said, title)
		}
		return nil
	}
	return fmt.Errorf("the form has no field %q", field)
}

func (w *world) listSaysNothing() error {
	if got := w.w.Choosing(); len(got) != 0 {
		return fmt.Errorf("the list offers %v, and there is nothing to serve", goalTitles(got))
	}
	return nil
}

// goalTitles is what a failure prints of a run of objectives.
func goalTitles(offered []Choice) []string {
	var said []string
	for _, one := range offered {
		said = append(said, one.Title)
	}
	return said
}

func (w *world) formOpenOnTask(number int) error {
	if w.w.Screen() != OnForm {
		return fmt.Errorf("the window is on %v, not a form", w.w.Screen())
	}
	if got := w.w.Form().Number; got != number {
		return fmt.Errorf("the form is open on %d, not %d", got, number)
	}
	return nil
}

// everyKeyTheBarOffers walks every screen and, on each, presses the key the
// bar writes beside every offer — then chooses the same offer with the
// pointer. The two have to land in the same place: the bar is what says what a
// key does, so a key that does something else makes the bar a lie.
func (w *world) everyKeyTheBarOffers() error {
	from := w.w
	for _, screen := range []Screen{OnTheBoard, OnForm, OnHelp, OnAsking} {
		at, err := onScreen(from, screen)
		if err != nil {
			return err
		}
		for _, offer := range at.Offers() {
			key := Shortcut(offer)
			if key == "" {
				return fmt.Errorf("on %v the bar offers %q under no key at all", screen, offer)
			}
			pressed, chosen := at.Key(key), at.Choose(offer)
			if pressed.Screen() != chosen.Screen() {
				return fmt.Errorf("on %v, %q left the window on %v and pressing %q left it on %v",
					screen, offer, chosen.Screen(), key, pressed.Screen())
			}
			if pressed.Screen() == at.Screen() && chosen.Screen() == at.Screen() && pressed.Chosen() != chosen.Chosen() {
				return fmt.Errorf("on %v, %q chose %d and pressing %q chose %d",
					screen, offer, chosen.Chosen(), key, pressed.Chosen())
			}
		}
	}
	return nil
}

// onScreen puts a window on a screen, so the step above can walk all of them.
func onScreen(w Window, screen Screen) (Window, error) {
	switch screen {
	case OnTheBoard:
		return w, nil
	case OnForm:
		return w.Key("enter"), nil
	case OnHelp:
		return w.Choose("Help"), nil
	case OnAsking:
		return w.Choose("Delete"), nil
	}
	return w, fmt.Errorf("no way to reach %v", screen)
}

func (w *world) caretOnTheSecondField() error {
	if got := w.w.Form().On(); got != 1 {
		return fmt.Errorf("the keyboard is on field %d, not the second", got)
	}
	return nil
}

func (w *world) tabHolds(name string, count int) error {
	for i, page := range Pages {
		if page != name {
			continue
		}
		if got := w.w.Holds(Tab(i)); got != count {
			return fmt.Errorf("the tab %q says it holds %d, not %d", name, got, count)
		}
		return nil
	}
	return fmt.Errorf("there is no tab called %q", name)
}

// chosenInSight asks where the chosen card is drawn and whether that is inside
// the run of the window it is drawn in.
func (w *world) chosenInSight() error {
	if w.w.Chosen() == 0 {
		return fmt.Errorf("nothing is chosen, so this proves nothing")
	}
	box, run, ok := w.w.WhereChosen()
	if !ok {
		return fmt.Errorf("task %d is chosen and drawn nowhere", w.w.Chosen())
	}
	if box.Y < run.Y || box.Y+box.H > run.Y+run.H {
		return fmt.Errorf("what is chosen is drawn from %d to %d, in a run from %d to %d",
			box.Y, box.Y+box.H, run.Y, run.Y+run.H)
	}
	return nil
}

func (w *world) objectivesMoved() error {
	if w.w.goals == w.was {
		return fmt.Errorf("the objectives are still at %d", w.w.goals)
	}
	return nil
}

// whereGoal is which of the objectives, in the order they are drawn, is chosen.
func (w *world) whereGoal() (int, error) {
	for i, g := range reading.Sorted(w.w.Showing().Goals) {
		if g.Number == w.w.Chosen() {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no objective is chosen")
}

func (w *world) objectivesAcross() error {
	l := w.w.Layout()
	first, second := GoalAt(l, 0, 0), GoalAt(l, 1, 0)
	if second.Y != first.Y {
		return fmt.Errorf("the second objective is a row below the first: they are one to a row")
	}
	if second.X <= first.X {
		return fmt.Errorf("the second objective is at %d and the first at %d", second.X, first.X)
	}
	return nil
}

func (w *world) objectivesInsideTheBoard() error {
	l := w.w.Layout()
	run := l.Column[0]
	last := l.Column[Columns-1]
	for i := range reading.Sorted(w.w.Showing().Goals) {
		box := GoalAt(l, i, 0)
		if box.X < run.X || box.X+box.W > last.X+last.W {
			return fmt.Errorf("objective %d runs from %d to %d, and the board from %d to %d",
				i, box.X, box.X+box.W, run.X, last.X+last.W)
		}
	}
	return nil
}

func (w *world) objectivesAllFit() error {
	if got := w.w.goals; got != 0 {
		return fmt.Errorf("the objectives start scrolled to %d", got)
	}
	l := w.w.Layout()
	last := len(reading.Sorted(w.w.Showing().Goals)) - 1
	box, run := GoalAt(l, last, 0), l.Column[0]
	if box.Y+box.H > run.Y+run.H {
		return fmt.Errorf("the last objective ends at %d and the board at %d", box.Y+box.H, run.Y+run.H)
	}
	return nil
}

func (w *world) chosenMovedOnOne() error {
	at, err := w.whereGoal()
	if err != nil {
		return err
	}
	if at != w.was+1 {
		return fmt.Errorf("what is chosen went from %d to %d", w.was, at)
	}
	w.was = at
	return nil
}

func (w *world) chosenMovedDownARow() error {
	at, err := w.whereGoal()
	if err != nil {
		return err
	}
	if want := w.was + GoalsAcross(); at != want {
		return fmt.Errorf("what is chosen went from %d to %d, and a row is %d", w.was, at, GoalsAcross())
	}
	w.was = at
	return nil
}

// cardOf finds a card by its title, so a step can ask what it says.
func (w *world) cardOf(title string) (reading.Card, error) {
	for _, c := range w.w.Showing().Cards {
		if c.Title == title {
			return c, nil
		}
	}
	return reading.Card{}, fmt.Errorf("no card titled %q", title)
}

// cardSays reads the line of labels a card draws, at the width a card in a
// column actually has.
func (w *world) cardSays(title, said string) error {
	c, err := w.cardOf(title)
	if err != nil {
		return err
	}
	// The whole bottom row: the labels on the left, and what is due, what it
	// has cost and what it serves on the right.
	facts, _ := Facts(c, moment)
	line := Tagged(c, PartsOf(w.w.Layout().CardAt(0, 0, 0)).Labels.W) + "  " + facts
	if !strings.Contains(line, said) {
		return fmt.Errorf("the card for %q reads %q, and says nothing of %q", title, line, said)
	}
	return nil
}

func (w *world) cardSaysInFocus(title string) error {
	c, err := w.cardOf(title)
	if err != nil {
		return err
	}
	if !c.InFocus {
		return fmt.Errorf("%q does not say it is being worked on", title)
	}
	if !strings.Contains(Numbered(c), InFocusMark) {
		return fmt.Errorf("the card for %q reads %q, and says nothing about being worked on",
			title, Numbered(c))
	}
	return nil
}

// firstWordOf is the point over the first character of a piece of text an open
// task writes: one beside its form, or one of the fields it is edited in —
// "field:description" is the description, wherever it is written.
func (w *world) firstWordOf(key string) (Point, error) {
	for _, r := range w.w.Readables() {
		if r.Key == key {
			return Point{X: r.Box.X + glyphOf(r.Size)/2, Y: r.Box.Y + lineHigh/2}, nil
		}
	}
	if name, ok := strings.CutPrefix(key, "field:"); ok {
		if at, err := w.inField(name); err == nil {
			return at(0), nil
		}
	}
	return Point{}, fmt.Errorf("the open task writes nothing called %q", key)
}
