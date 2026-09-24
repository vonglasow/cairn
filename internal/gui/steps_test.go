package gui

// What the scenarios look at once a gesture has been made. Everything here
// reads the window's own state or the store through the second connection — a
// precondition or an assertion that went through the code under test would
// cancel out with it and prove nothing.

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cucumber/godog"

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
	dealt, _ := reading.Deal(w.w.Reading().Cards)
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
	dealt, _ := reading.Deal(w.w.Showing().Cards)
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
	_, stalled := reading.Deal(w.w.Reading().Cards)
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
	if w.w.Screen() != OnDetail {
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

// servingCardsAreCards is what tells a card from a line of text: it is drawn at
// the height every other card on this board is drawn at.
func (w *world) servingCardsAreCards() error {
	for at := range w.w.Under() {
		if got := w.w.ServingCardAt(at).H; got != CardHeight {
			return fmt.Errorf("the card at %d is %d high, and a card is %d", at, got, CardHeight)
		}
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
	if got := w.w.PanelHead(); got != head {
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

func (w *world) detailClosed() error {
	if w.w.Screen() != OnTheBoard {
		return fmt.Errorf("the window is on %v", w.w.Screen())
	}
	return nil
}

func (w *world) isScrolled(column string) error {
	at := reading.IndexOf(strings.ToLower(column))
	if w.w.scroll[at] == w.scrolled[at] {
		return fmt.Errorf("%q did not move", column)
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
	found, err := w.other.Show(number)
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

func (w *world) expects(sc *godog.ScenarioContext) {
	sc.Then(`^the columns read (".*")$`, w.columnsRead)
	sc.Then(`^"([^"]*)" carries "([^"]*)"$`, w.columnCarries)
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
	sc.Then(`^the detail is open on "([^"]*)"$`, w.detailOpenOn)
	sc.Then(`^the detail shows "([^"]*)"$`, w.detailShows)
	sc.Then(`^the detail draws a card for "([^"]*)" under its fields$`, w.detailLists)
	sc.Then(`^the detail draws no card for "([^"]*)"$`, w.detailDoesNotList)
	sc.Then(`^the box under the fields is empty$`, w.detailSaysNothingServes)
	sc.Then(`^the cards under the fields are the height of a card$`, w.servingCardsAreCards)
	sc.Then(`^the right column is still headed "([^"]*)"$`, w.rightColumnHead)
	sc.Then(`^the detail says nothing serves it yet$`, w.detailSaysNothingServes)
	sc.Then(`^the detail is closed$`, w.detailClosed)
	sc.Then(`^"([^"]*)" is scrolled$`, w.isScrolled)
	sc.Then(`^"([^"]*)" is not scrolled$`, w.isNotScrolled)
	sc.Then(`^nothing was asked of the engine$`, w.askedNothing)
	sc.Then(`^the engine was asked "([^"]*)" for task (\d+)$`, w.wasAsked)
	sc.Then(`^the engine was not asked "([^"]*)"$`, w.wasNotAsked)
	sc.Then(`^the engine was asked "([^"]*)" once$`, w.wasAskedOnce)
	sc.Then(`^task (\d+) is in the column "([^"]*)"$`, w.taskInColumn)
	sc.Then(`^task (\d+) is called "([^"]*)"$`, w.taskCalled)
	sc.Then(`^the form asks about (".*")$`, w.formAsksAbout)
	sc.Then(`^the field "([^"]*)" reads "([^"]*)"$`, w.fieldReads)
	sc.Then(`^the store was read once$`, w.readOnce)
	sc.Then(`^it says what went wrong$`, w.saysWhatWentWrong)
	sc.Then(`^the detail shows the field "([^"]*)"$`, w.detailShowsField)
	sc.Then(`^the detail shows the field "([^"]*)" reading "([^"]*)"$`, w.detailFieldReads)
	sc.Then(`^the card being dragged follows the pointer$`, w.ghostFollows)
	sc.Then(`^it shows where the card would land in "([^"]*)"$`, w.landsIn)
	sc.Then(`^it shows nowhere for the card to land$`, w.landsNowhere)
	sc.Then(`^"([^"]*)" does not carry "([^"]*)"$`, w.columnDoesNotCarry)
	sc.Then(`^the help is open$`, w.helpOpen)
	sc.Then(`^the help is closed$`, w.helpClosed)
	sc.Then(`^the help says what "([^"]*)" does$`, w.helpSaysAbout)
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
	sc.Then(`^what is chosen is in sight$`, w.chosenInSight)
	sc.Then(`^the objectives have moved$`, w.objectivesMoved)
	sc.Then(`^the caret is on the second field$`, w.caretOnTheSecondField)
	sc.Then(`^the tab "([^"]*)" says it holds (\d+)$`, w.tabHolds)
	sc.Then(`^the form is open on task (\d+)$`, w.formOpenOnTask)
	sc.Then(`^the list of objectives is open$`, w.listOpen)
	sc.Then(`^the form is still drawn behind it$`, w.formBehindTheList)
	sc.Then(`^the field "([^"]*)" is drawn by name and not by number$`, w.fieldDrawnByName)
	sc.Then(`^the list offers "([^"]*)"$`, w.listOffers)
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
	sc.Then(`^the history is at its beginning$`, w.historyAtTheBeginning)
	sc.Then(`^the history is at its end$`, w.historyAtTheEnd)
	sc.Then(`^the field "([^"]*)" is written on more than one line$`, w.fieldOnMoreThanOneLine)
	sc.Then(`^no line of the field "([^"]*)" runs out of it$`, w.noLineRunsOut)
	sc.Then(`^the bar offers "([^"]*)" under "([^"]*)"$`, w.barOffersUnder)
	sc.Then(`^the window is asking what to look for$`, w.askingWhatToLookFor)
	// Asked of every card the board deals, columns and band alike: "not in
	// BACKLOG" would pass for a card that merely sits in another column.
	sc.Then(`^the board says nothing of "([^"]*)"$`, func(title string) error {
		dealt, stalled := reading.Deal(w.w.Showing().Cards)
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
	sc.Then(`^the search box is open$`, w.searchBoxOpen)
	sc.Then(`^the search box is closed$`, w.searchBoxClosed)
	sc.Then(`^the search is still in force$`, w.searchStillInForce)
	sc.Then(`^the caret is at (\d+)$`, w.caretIsAt)
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
	found, err := w.other.ShowObjective(number)
	if err != nil {
		return fmt.Errorf("objective %d: %w", number, err)
	}
	if got := objective.OfObjective(found).Title; got != title {
		return fmt.Errorf("objective %d is called %q, want %q", number, got, title)
	}
	return nil
}

func (w *world) detailShowsField(name string) error {
	for _, field := range w.w.Shown() {
		if field.Name == name {
			return nil
		}
	}
	return fmt.Errorf("the detail does not show %q", name)
}

func (w *world) detailFieldReads(name, value string) error {
	for _, field := range w.w.Shown() {
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
		if strings.Contains(line.Name, about) || strings.Contains(line.Value, about) {
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
	for _, field := range w.w.Form().Fields() {
		if field.Name == name {
			if RowsOf(field) <= 1 {
				return fmt.Errorf("%q has room for %d line", name, RowsOf(field))
			}
			return nil
		}
	}
	return fmt.Errorf("the form has no field called %q", name)
}

func (w *world) fieldHasRoomForOne(name string) error {
	for _, field := range w.w.Form().Fields() {
		if field.Name == name {
			if RowsOf(field) != 1 {
				return fmt.Errorf("%q has room for %d lines", name, RowsOf(field))
			}
			return nil
		}
	}
	return fmt.Errorf("the form has no field called %q", name)
}

func (w *world) fieldShowsTheEnd(name string) error {
	box := Over(w.w.Layout().Bar.W+2*margin, w.w.Layout().Bar.Y+barHigh+margin)
	fields := w.w.Form().Fields()
	for i, field := range fields {
		if field.Name != name {
			continue
		}
		_, value := FieldRow(box, FieldTop(box, fields, i))
		lines := Wrapped(field.Value, room(Inside(Framed(value, RowsOf(field)))))
		if len(lines.Said) < 2 {
			return fmt.Errorf("%q is not longer than its room, so this proves nothing", name)
		}
		line, _ := lines.At(w.w.Form().Writing().At)
		shown := lines.Said[lines.Window(RowsOf(field), line)]
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
		_, value := FieldRow(Over(w.w.width, w.w.height), FieldTop(Over(w.w.width, w.w.height), fields, i))
		fits := room(Inside(Framed(value, RowsOf(field))))
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
	fields, history := DetailFields(box), HistoryBox(box)
	if history.X < fields.X+fields.W {
		return fmt.Errorf("the history starts at %d and the fields end at %d: it is under them, not beside them",
			history.X, fields.X+fields.W)
	}
	return nil
}

func (w *world) historyInsideThePanel() error {
	box := Over(w.w.width, w.w.height)
	for what, part := range map[string]Box{"the history": HistoryBox(box), "its bar": HistoryBar(box)} {
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
	box := ServingBox(Over(w.w.width, w.w.height), w.w.Shown())
	last := len(w.w.Under())*CardHeight - box.H
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
	_, value := FieldRow(over, FieldTop(over, fields, 0))
	frame := Framed(value, RowsOf(fields[0]))
	if frame.Y+frame.H > list.Y {
		return fmt.Errorf("the first field reaches %d and the list starts at %d: they overlap, so this proves nothing",
			frame.Y+frame.H, list.Y)
	}
	// Its outline, and not its fill: the fill is the same colour as the ground
	// the whole window is painted on, so counting that would pass even with no
	// form drawn at all. Measured — it did.
	if painted(drawn(w.t, w.w), frame, Dark().Edge) == 0 {
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
		box := Over(w.w.width, w.w.height)
		_, value := FieldRow(box, FieldTop(box, fields, at))
		frame := Framed(value, RowsOf(f))
		// Past where the numbers themselves could have ended, and inside the
		// frame rather than on it: the frame's own outline is not the field's
		// writing, and a fraction of the width would only say the text is
		// long. What proves the names are drawn is ink beyond the reach of
		// what the field carries.
		inside := Inside(frame)
		from := inside.X + (len(f.Value)+2)*perGlyph
		far := Box{X: from, Y: inside.Y, W: inside.X + inside.W - from, H: inside.H}
		if inked(drawn(w.t, w.w), far, Dark().Back) == 0 {
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
func goalTitles(goals []reading.Goal) []string {
	var said []string
	for _, g := range goals {
		said = append(said, g.Title)
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
	for _, screen := range []Screen{OnTheBoard, OnDetail, OnForm, OnHelp, OnAsking} {
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
	case OnDetail:
		return w.Key("enter"), nil
	case OnForm:
		return w.Choose("Edit"), nil
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
