package gui

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"gioui.org/font"
	"gioui.org/io/key"
	"github.com/cucumber/godog"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/linuxapp"
	"cairn/internal/migration"
	"cairn/internal/objective"
	"cairn/internal/task"
)

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
	dealt, _ := engine.Deal(w.w.Reading().Cards, w.w.archived)
	for _, column := range dealt {
		got = append(got, column.Name)
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		return fmt.Errorf("the columns read %v, want %v", got, want)
	}
	return nil
}

func (w *world) cardsOf(column string) ([]engine.Card, error) {
	at := engine.IndexOf(strings.ToLower(column))
	if at < 0 {
		return nil, fmt.Errorf("no column called %q", column)
	}
	dealt, _ := engine.Deal(w.w.Showing().Cards, w.w.archived)
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
	if engine.Band != name {
		return fmt.Errorf("the band is named %q", engine.Band)
	}
	return nil
}

func (w *world) lanesRead(list string) error {
	want := quoted(list)
	if strings.Join(engine.Stalled, "|") != strings.Join(want, "|") {
		return fmt.Errorf("the lanes read %v, want %v", engine.Stalled, want)
	}
	return nil
}

func (w *world) laneReads(kind, list string) error {
	lane, err := laneNamed(kind)
	if err != nil {
		return err
	}
	_, stalled := engine.Deal(w.w.Showing().Cards, w.w.archived)
	var got []string
	for _, c := range engine.InLanes(stalled)[lane] {
		got = append(got, c.Title)
	}
	want := quoted(list)
	if strings.Join(got, "|") != strings.Join(want, "|") {
		return fmt.Errorf("the lane %q reads %v, want %v", kind, got, want)
	}
	return nil
}

func (w *world) laneCarries(kind, title string) error {
	_, stalled := engine.Deal(w.w.Reading().Cards, w.w.archived)
	for _, c := range stalled {
		if c.State == kind && c.Title == title {
			return nil
		}
	}
	return fmt.Errorf("the lane %q does not carry %q", kind, title)
}

func (w *world) tabsRead(list string) error {
	want := quoted(list)
	var got []string
	for _, tab := range w.w.Shown() {
		got = append(got, Pages[tab])
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		return fmt.Errorf("the tabs read %v, want %v", got, want)
	}
	return nil
}

func (w *world) tabIs(name string) error {
	got := Pages[w.w.Tab()]
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
	at := engine.IndexOf(strings.ToLower(column))
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
	at := engine.IndexOf(strings.ToLower(column))
	title := PartsOf(w.w.Layout().CardAt(at, 0, 0)).Title
	m := w.w.measured()
	for _, c := range here {
		first, rest := TitleRoom(c, title)
		lines := TitleLines(c.Title, first, rest, func(said string) int { return ends(m, said, reads) })
		if strings.HasSuffix(lines[len(lines)-1], ending) {
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

func titles(cards []engine.Card) []string {
	var said []string
	for _, c := range cards {
		said = append(said, c.Title)
	}
	return said
}

func (w *world) rightColumnTabs(first, second string) error {
	if got := w.w.BesideTabsSay(); got != [2]string{first, second} {
		return fmt.Errorf("the column beside the form offers the tabs %q", got)
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
	at := engine.IndexOf(strings.ToLower(column))
	if w.w.scroll[at] == w.scrolled[at] {
		return fmt.Errorf("%q did not move", column)
	}
	return nil
}

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

func (w *world) acrossTheForm(name string, above []string) error {
	over := w.w.formBox()
	fields := w.w.Form().Fields()
	frames := make([]Box, len(fields))
	for i := range fields {
		_, value, rows := FormPlace(over, fields, i)
		frames[i] = Framed(value, rows)
	}
	left, right, at := frames[0].X, 0, -1
	for i, frame := range frames {
		if frame.X >= over.X+over.W {
			continue
		}
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
		if i == at || slices.Contains(above, fields[i].Name) || frame.X >= over.X+over.W {
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
	at := engine.IndexOf(strings.ToLower(column))
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

func (w *world) formHeadedBold(want string) error {
	if w.w.Screen() != OnForm {
		return fmt.Errorf("no form is open")
	}
	if got := HeadsOn(w.w, OnForm); !strings.HasPrefix(got, want) {
		return fmt.Errorf("the form is headed %q, want it to begin %q", got, want)
	}
	if HeadingWeight < font.Bold {
		return fmt.Errorf("the heading is drawn at weight %d, want bold", HeadingWeight)
	}
	for _, f := range w.w.Form().Fields() {
		if HeadingSize <= FormSize(f) {
			return fmt.Errorf("the heading is %d, no larger than the field %q at %d", HeadingSize, f.Name, FormSize(f))
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

func (w *world) cardNumbered(n int) (engine.Card, error) {
	for _, c := range w.w.Showing().Cards {
		if c.Number == n {
			return c, nil
		}
	}
	return engine.Card{}, fmt.Errorf("the board carries no card numbered %d", n)
}

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

func (w *world) listInsideThePanel() error {
	box := w.w.ChoosingBox()
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
		if ink := Dark().Writing(c); ink != Dark().Colour.Ink.Faint {
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
		if ink := Dark().Writing(c); ink != Dark().Colour.Ink.Text {
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
	sc.Then(`^the lane "([^"]*)" reads (".*")$`, w.laneReads)
	sc.Then(`^the band is named "([^"]*)"$`, w.bandNamed)
	sc.Then(`^the band's lanes read (".*")$`, w.lanesRead)
	sc.Then(`^the lane "([^"]*)" carries "([^"]*)"$`, w.laneCarries)
	sc.Then(`^the tabs read (".*")$`, w.tabsRead)
	sc.Then(`^the tab is "([^"]*)"$`, w.tabIs)
	sc.Then(`^every card in "([^"]*)" is the same height$`, w.sameHeight)
	sc.Then(`^the card "([^"]*)" is striped in its priority's colour (\d+) points wide$`, w.stripedWide)
	sc.Then(`^"([^"]*)" carries a card whose title ends in "([^"]*)"$`, w.titleEndsIn)
	sc.Then(`^it says the window is too small$`, w.tooSmall)
	sc.Then(`^it says what size it needs$`, w.saysWhatItNeeds)
	sc.Then(`^the card "([^"]*)" is chosen$`, w.isChosen)
	sc.Then(`^the card "([^"]*)" is not chosen$`, w.isNotChosen)
	sc.Then(`^"([^"]*)" is open to be edited$`, w.detailOpenOn)
	sc.Then(`^the form shows "([^"]*)"$`, w.detailShows)
	sc.Then(`^the work beside the form carries "([^"]*)"$`, w.detailLists)
	sc.Then(`^the work beside the form does not carry "([^"]*)"$`, w.detailDoesNotList)
	sc.Then(`^what serves it is drawn one line to a tile$`, w.servingCardsAreCards)
	sc.Then(`^the column beside the form offers the tabs "([^"]*)" and "([^"]*)"$`, w.rightColumnTabs)
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
	sc.Then(`^the steps are drawn as a list to tick, "([^"]*)" first$`, func(title string) error {
		if parts := w.w.PartsTicked(); len(parts) == 0 || parts[0].Title != title {
			return fmt.Errorf("the parts are drawn as %v, and %q is not first", parts, title)
		}
		return nil
	})
	sc.Then(`^the steps are drawn as a list to tick, "([^"]*)" among them$`, func(title string) error {
		for _, c := range w.w.PartsTicked() {
			if c.Title == title {
				return nil
			}
		}
		return fmt.Errorf("the parts are drawn as %v, and %q is not among them", w.w.PartsTicked(), title)
	})
	sc.Then(`^the trace begins with the last thing done to task (\d+)$`, func(n int) error {
		shown, err := w.held.Ask(engine.TraceShow, engine.Numbered(n))
		if err != nil {
			return err
		}
		last, lines := line(shown.Acts[len(shown.Acts)-1]), w.w.Lines()
		if len(lines) == 0 || !strings.HasPrefix(last, lines[0]) {
			return fmt.Errorf("the history begins %q, and the last thing done was %q", lines, last)
		}
		return nil
	})
	sc.Then(`^no line of the trace is empty$`, func() error {
		for i, said := range w.w.Lines() {
			if strings.TrimSpace(said) == "" {
				return fmt.Errorf("line %d of the history is empty: %q", i, w.w.Lines())
			}
		}
		return nil
	})
	sc.Then(`^every field of the form stands left of the trace$`, func() error {
		history := BesideBox(w.w.formPanel())
		_, frames := w.formFrames()
		for i, frame := range frames {
			if frame.X+frame.W > history.X {
				return fmt.Errorf("%q runs to %d and the history starts at %d", w.w.Form().Fields()[i].Name, frame.X+frame.W, history.X)
			}
		}
		return nil
	})
	sc.Then(`^the work beside the form reads (.+)$`, func(list string) error {
		var got []string
		for _, c := range w.w.Under() {
			got = append(got, c.Title)
		}
		if want := strings.Join(quoted(list), "|"); strings.Join(got, "|") != want {
			return fmt.Errorf("the work beside the form reads %q, want %q", got, want)
		}
		return nil
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
		for i, offer := range w.w.offersAlong(box) {
			if offer.X < box.X || offer.X+offer.W > box.X+box.W || offer.Y+offer.H > box.Y+box.H {
				return fmt.Errorf("offer %d at %v runs out of the question %v", i, offer, box)
			}
		}
		return nil
	})
	sc.Then(`^the steps are at their beginning$`, func() error {
		if from := w.w.partsFrom(); from != 0 {
			return fmt.Errorf("the parts are shown from %d", from)
		}
		return nil
	})
	sc.Then(`^the steps have moved$`, func() error {
		if from := w.w.partsFrom(); from == w.was {
			return fmt.Errorf("the parts are still shown from %d", from)
		}
		return nil
	})
	sc.Then(`^the last step is in sight$`, func() error {
		parts, tiles := w.w.PartsTicked(), w.w.partTiles()
		if last := w.w.partsFrom() + len(tiles) - 2; last != len(parts)-1 {
			return fmt.Errorf("the last part in sight is %d of %d", last+1, len(parts))
		}
		return nil
	})
	sc.Then(`^a step is in hand, under the pointer$`, func() error {
		ghost, ok := w.w.PartGhost()
		if !ok || !ghost.In(w.pointer) {
			return fmt.Errorf("the part in hand is drawn at %v (%v), and the pointer is at %v", ghost, ok, w.pointer)
		}
		return nil
	})
	sc.Then(`^it would land below "([^"]*)"$`, func(title string) error {
		line, ok := w.w.PartLanding()
		at := slices.IndexFunc(w.w.PartsTicked(), func(c engine.Card) bool { return c.Title == title })
		tile := w.w.partTiles()[at-w.w.partsFrom()]
		if !ok || line.Y != tile.Y+tile.H-landingHigh {
			return fmt.Errorf("it would land at %v (%v), and %q ends at %d", line, ok, title, tile.Y+tile.H)
		}
		return nil
	})
	sc.Then(`^the step the keyboard is on is "([^"]*)"$`, func(want string) error {
		if got, on := w.w.PartOn(); !on || got != want {
			return fmt.Errorf("the keyboard is on the part %q (%v), want %q", got, on, want)
		}
		return nil
	})
	sc.Then(`^the step "([^"]*)" is ticked$`, func(title string) error {
		for _, c := range w.w.PartsTicked() {
			if c.Title == title && engine.Finished(c) {
				return nil
			}
		}
		return fmt.Errorf("the parts are %+v, and %q is not ticked", w.w.PartsTicked(), title)
	})
	sc.Step(`^I drag across the trace from the start of its first line to the end of its second$`, func() error {
		r, ok := w.w.readable("trace")
		if !ok || len(r.Lines.Said) < 2 {
			return fmt.Errorf("the trace is not one text of two lines or more: %+v", r)
		}
		from := Point{X: r.Box.X, Y: r.Box.Y + (0-r.From)*lineHigh + lineHigh/2}
		to := Point{X: r.Box.X + r.Box.W - 1, Y: r.Box.Y + (1-r.From)*lineHigh + lineHigh/2}
		w.w = w.w.PressAt(from).MoveTo(to).LetGo(to)
		return nil
	})
	sc.Then(`^the clipboard reads the first (\d+) lines of the trace$`, func(n int) error {
		lines := w.w.Lines()
		if len(lines) < n {
			return fmt.Errorf("the trace has %d lines", len(lines))
		}
		if want := strings.Join(lines[:n], "\n"); w.clipboard != want {
			return fmt.Errorf("the clipboard reads %q, want %q", w.clipboard, want)
		}
		return nil
	})
	sc.Step(`^I click the step "([^"]*)"$`, func(title string) error {
		at, err := w.stepTile(title)
		if err != nil {
			return err
		}
		w.w = w.w.clicked(at)
		return nil
	})
	sc.Step(`^I ⌘-click the step "([^"]*)"$`, func(title string) error {
		return w.clickWith(key.ModShortcut, w.stepTile, title)
	})
	sc.Step(`^I ⇧-click the step "([^"]*)"$`, func(title string) error {
		return w.clickWith(key.ModShift, w.stepTile, title)
	})
	sc.Step(`^I double-click the step "([^"]*)"$`, func(title string) error {
		at, err := w.stepTile(title)
		if err != nil {
			return err
		}
		w.w = w.w.PressAt(at).LetGo(at)
		w.w = w.w.doubleClicked(at).LetGo(at)
		return nil
	})
	sc.Then(`^the form's steps read (.+)$`, func(list string) error {
		var want, got []string
		for _, m := range regexp.MustCompile(`"([^"]*)"`).FindAllStringSubmatch(list, -1) {
			want = append(want, m[1])
		}
		for _, part := range w.w.form.parts() {
			got = append(got, part.Title)
		}
		if !slices.Equal(got, want) {
			return fmt.Errorf("the form's steps read %q, want %q", got, want)
		}
		return nil
	})
	sc.Then(`^the last act on task (\d+) changed its steps from "([^"]*)" to "([^"]*)"$`, func(n int, was, became string) error {
		shown, err := w.other.Ask(engine.TraceShow, engine.Input{Given: map[string]string{"number": strconv.Itoa(n)}})
		if err != nil {
			return err
		}
		if len(shown.Acts) == 0 {
			return fmt.Errorf("task %d has no act", n)
		}
		last := shown.Acts[len(shown.Acts)-1].Carried()
		if last.Field != "steps" || last.Was != was || last.Became != became {
			return fmt.Errorf("the last act on task %d is %+v", n, last)
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
	sc.Then(`^the question says "([^"]*)"$`, func(want string) error {
		for _, line := range w.w.Asks() {
			if line.Means == want {
				return nil
			}
		}
		return fmt.Errorf("the question says %+v, not %q", w.w.Asks(), want)
	})
	sc.Then(`^the question says nothing of how to answer it$`, func() error {
		for _, line := range w.w.Asks() {
			if strings.Contains(line.Means, "agree") || strings.Contains(line.Means, "Cancel") {
				return fmt.Errorf("the question says %q, which its buttons already say", line.Means)
			}
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
	sc.Then(`^the window asks to put task (\d+) in done, ticking its steps$`, func(n int) error {
		if w.w.Screen() != OnAsking || w.w.question != aboutFinishingParts {
			return fmt.Errorf("the window is on %v asking %v", w.w.Screen(), w.w.question)
		}
		if head := Heads(w.w); !strings.Contains(head, strconv.Itoa(n)) {
			return fmt.Errorf("the question reads %q, naming no task %d", head, n)
		}
		return nil
	})
	sc.Then(`^the window says the store comes from a newer cairn$`, func() error {
		if w.w.Screen() != OnAsking || w.w.question != aboutNewer {
			return fmt.Errorf("the window is on %v asking %v", w.w.Screen(), w.w.question)
		}
		return nil
	})
	sc.Then(`^the window is leaving$`, func() error {
		if !w.w.Leaving() {
			return fmt.Errorf("the window is not leaving")
		}
		return nil
	})
	sc.Then(`^the window asks to apply the change that is waiting, naming it$`, func() error {
		if w.w.Screen() != OnAsking || w.w.question != aboutApplying {
			return fmt.Errorf("the window is on %v asking %v", w.w.Screen(), w.w.question)
		}
		listed, err := w.work.Ask(engine.MigrationList, engine.Input{})
		if err != nil {
			return err
		}
		said := Heads(w.w) + " " + sentences(w.w.Asks())
		for _, m := range listed.Changes {
			if shape := migration.OfMigration(m); !shape.Applied && !strings.Contains(said, shape.Name) {
				return fmt.Errorf("the question reads %q, naming no %q", said, shape.Name)
			}
		}
		return nil
	})
	sc.Then(`^task (\d+) is called "([^"]*)"$`, w.taskCalled)
	sc.Then(`^there (?:is|are) (\d+) tasks?$`, func(n int) error {
		for i := 1; i <= n; i++ {
			if _, err := w.tell(i); err != nil {
				return fmt.Errorf("task %d is not there: %w", i, err)
			}
		}
		if _, err := w.tell(n + 1); err == nil {
			return fmt.Errorf("there is a task %d, want %d tasks", n+1, n)
		}
		return nil
	})
	sc.Then(`^there are no tasks$`, func() error {
		if _, err := w.tell(1); err == nil {
			return fmt.Errorf("there is a task 1, want none")
		}
		return nil
	})
	sc.Then(`^the form asks about (".*")$`, w.formAsksAbout)
	sc.Then(`^the form is headed "([^"]*)" in bold, larger than its fields$`, w.formHeadedBold)
	sc.Then(`^the field "([^"]*)" reads "([^"]*)"$`, w.fieldReads)
	sc.Then(`^the store was read once$`, w.readOnce)
	sc.Then(`^it says what went wrong$`, w.saysWhatWentWrong)
	sc.Then(`^the week shown starts on "([^"]*)"$`, func(day string) error {
		if got := domain.FormatDay(w.w.week); got != day {
			return fmt.Errorf("the week shown starts on %s, want %s", got, day)
		}
		return nil
	})
	dayOfWeek := func(day string) (engine.TimesheetDay, bool) {
		for _, d := range w.w.sheet.Days {
			if d.Day == day {
				return d, true
			}
		}
		return engine.TimesheetDay{}, false
	}
	sc.Then(`^the day "([^"]*)" says "([^"]*)"$`, func(day, total string) error {
		d, _ := dayOfWeek(day)
		if got := d.Says(); got != total {
			return fmt.Errorf("%s says %q, want %q", day, got, total)
		}
		return nil
	})
	sc.Then(`^the day "([^"]*)" carries "([^"]*)" for "([^"]*)"$`, func(day, title, length string) error {
		d, _ := dayOfWeek(day)
		for _, one := range d.Tasks {
			if one.Title == title {
				if got := one.Says(); !strings.Contains(got, length) {
					return fmt.Errorf("%s carries %q, want it to say %s", day, got, length)
				}
				return nil
			}
		}
		return fmt.Errorf("%s carries nothing titled %q", day, title)
	})
	weekday := func(day string) (int, error) {
		for i := range 7 {
			if domain.FormatDay(w.w.week.AddDate(0, 0, i)) == day {
				return i, nil
			}
		}
		return 0, fmt.Errorf("the week shown has no %s", day)
	}
	sc.Given(`^I spent "([^"]*)" on each of tasks (\d+) to (\d+) on "([^"]*)"$`, func(length string, from, to int, day string) error {
		for n := from; n <= to; n++ {
			in := engine.Input{Given: map[string]string{"number": strconv.Itoa(n), "time": length, "on": day}}
			if _, err := w.other.Ask(engine.TimesheetAdd, in); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Then(`^the day "([^"]*)" has more to show than it has room for$`, func(day string) error {
		i, err := weekday(day)
		if err != nil {
			return err
		}
		if all, room := len(w.w.sheet.On(w.w.week.AddDate(0, 0, i)).Tasks), DayRoom(WeekDayAt(w.w.Layout(), i)); all <= room {
			return fmt.Errorf("%s holds %d tasks in room for %d", day, all, room)
		}
		return nil
	})
	sc.Step(`^I scroll down over the day "([^"]*)" (\d+) times$`, func(day string, n int) error {
		i, err := weekday(day)
		if err != nil {
			return err
		}
		box := WeekDayAt(w.w.Layout(), i)
		for range n {
			w.w = w.w.Scroll(Point{X: box.X + 4, Y: box.Y + box.H/2}, weekRow)
		}
		return nil
	})
	shows := func(day, title string) (bool, error) {
		i, err := weekday(day)
		if err != nil {
			return false, err
		}
		for _, one := range w.w.DayShown(i) {
			if one.Title == title {
				return true, nil
			}
		}
		return false, nil
	}
	sc.Then(`^the day "([^"]*)" shows "([^"]*)"$`, func(day, title string) error {
		shown, err := shows(day, title)
		if err != nil {
			return err
		}
		if !shown {
			return fmt.Errorf("%s does not show %q", day, title)
		}
		return nil
	})
	sc.Then(`^the day "([^"]*)" does not show "([^"]*)"$`, func(day, title string) error {
		shown, err := shows(day, title)
		if err != nil {
			return err
		}
		if shown {
			return fmt.Errorf("%s shows %q", day, title)
		}
		return nil
	})
	sc.Step(`^I click "([^"]*)" on the time page$`, func(chip string) error {
		box := TimeChip(w.w.Layout(), chip)
		w.w = w.w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		return w.w.Wrong()
	})
	sc.Then(`^the time page offers the twelve months$`, func() error {
		if got := w.w.TimeCells(); len(got) != 12 || got[0] != "Jan" {
			return fmt.Errorf("the time page offers %v", got)
		}
		return nil
	})
	sc.Step(`^I click the cell "([^"]*)" on the time page$`, func(name string) error {
		for i, cell := range w.w.TimeCells() {
			if cell == name {
				box := TimeCellAt(w.w.Layout(), i)
				w.w = w.w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
				return w.w.Wrong()
			}
		}
		return fmt.Errorf("the time page offers %v, not %q", w.w.TimeCells(), name)
	})
	sc.Then(`^every cell of the time page is drawn as a button$`, func() error {
		var cells []Box
		for i := range w.w.TimeCells() {
			cells = append(cells, TimeCellAt(w.w.Layout(), i))
		}
		return buttonsDrawn(w.w, cells, Dark().Colour.Surface.Control, Dark().Colour.Accent.Base)
	})
	sc.Step(`^I rest the pointer on the (count of labels|title) of "([^"]*)"$`, func(what, title string) error {
		for _, placed := range w.w.Placed() {
			if placed.Card.Title != title {
				continue
			}
			parts := PartsOf(placed.Slot)
			at := Point{X: parts.Title.X + 4, Y: parts.Title.Y + 4}
			if what == "count of labels" {
				more := CardRowOf(placed.Card, parts.Number, w.w.measured(), moment, engine.Pages{}).More
				if more == (Box{}) {
					return fmt.Errorf("the card for %q counts no labels", title)
				}
				at = Point{X: more.X + more.W/2, Y: more.Y + more.H/2}
			}
			w.w = w.w.Hovering(at)
			return nil
		}
		return fmt.Errorf("the board carries no card called %q", title)
	})
	sc.Then(`^a bubble shows the labels (".*")$`, func(list string) error {
		if got := w.w.LabelsBubbled(); strings.Join(got, "|") != strings.Join(quoted(list), "|") {
			return fmt.Errorf("the bubble shows the labels %v", got)
		}
		return nil
	})
	sc.Then(`^no bubble shows labels$`, func() error {
		if got := w.w.LabelsBubbled(); len(got) != 0 {
			return fmt.Errorf("a bubble shows the labels %v", got)
		}
		return nil
	})
	sc.Then(`^the day shown is "([^"]*)"$`, func(day string) error {
		if !w.w.Daily() || domain.FormatDay(w.w.week) != day {
			return fmt.Errorf("the time page shows %q, daily %v", TimeSays(w.w), w.w.Daily())
		}
		return nil
	})
	aloneShows := func(title string) bool {
		for _, one := range w.w.AloneShown() {
			if one.Title == title {
				return true
			}
		}
		return false
	}
	sc.Then(`^the day shown shows "([^"]*)"$`, func(title string) error {
		if !aloneShows(title) {
			return fmt.Errorf("the day shown does not show %q", title)
		}
		return nil
	})
	sc.Then(`^the day shown does not show "([^"]*)"$`, func(title string) error {
		if aloneShows(title) {
			return fmt.Errorf("the day shown shows %q", title)
		}
		return nil
	})
	sc.Step(`^I click the head of the day "([^"]*)"$`, func(day string) error {
		i, err := weekday(day)
		if err != nil {
			return err
		}
		head := WeekHeadAt(WeekDayAt(w.w.Layout(), i))
		w.w = w.w.clicked(Point{X: head.X + 4, Y: head.Y + head.H/2})
		return w.w.Wrong()
	})
	sc.Then(`^the day "([^"]*)" shows a total of "([^"]*)"$`, func(day, total string) error {
		i, err := weekday(day)
		if err != nil {
			return err
		}
		if got := w.w.DayHeld(i).Says(); got != total {
			return fmt.Errorf("%s shows a total of %q, want %q", day, got, total)
		}
		return nil
	})
	sc.Then(`^the month shown is "([^"]*)"$`, func(month string) error {
		if !w.w.Monthly() || w.w.week.Format("January 2006") != month {
			return fmt.Errorf("the time page shows %q, monthly %v", TimeSays(w.w), w.w.Monthly())
		}
		return nil
	})
	dayOfMonth := func(day string) (int, error) {
		for i, shown := range engine.MonthGrid(w.w.week) {
			if domain.FormatDay(shown) == day {
				return i, nil
			}
		}
		return 0, fmt.Errorf("the month shown has no %s", day)
	}
	sc.Then(`^the month's day "([^"]*)" shows "([^"]*)"$`, func(day, title string) error {
		i, err := dayOfMonth(day)
		if err != nil {
			return err
		}
		lines := CellLines(w.w.MonthDay(i).Tasks, CellRoom(MonthCellAt(w.w.Layout(), i)))
		for _, line := range lines {
			if strings.Contains(line, title) {
				return nil
			}
		}
		return fmt.Errorf("%s reads %q", day, lines)
	})
	sc.Step(`^I click the month's day "([^"]*)"$`, func(day string) error {
		i, err := dayOfMonth(day)
		if err != nil {
			return err
		}
		box := MonthCellAt(w.w.Layout(), i)
		w.w = w.w.clicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
		return nil
	})
	cardOfWeek := func(title string) (Point, error) {
		l := w.w.Layout()
		for i := range 7 {
			for j, one := range w.w.DayShown(i) {
				if w.w.DayCard(one).Title == title {
					box := DayCardAt(WeekDayAt(l, i), j)
					return Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, nil
				}
			}
		}
		return Point{}, fmt.Errorf("no card titled %q is on the time page", title)
	}
	sc.Then(`^the day "([^"]*)" shows the card "([^"]*)" for "([^"]*)"$`, func(day, title, length string) error {
		i, err := weekday(day)
		if err != nil {
			return err
		}
		for _, one := range w.w.DayShown(i) {
			c := w.w.DayCard(one)
			if c.Title != title {
				continue
			}
			if c.Estimate == "" {
				return fmt.Errorf("%s shows %q without what its card carries: %+v", day, title, c)
			}
			if !strings.Contains(one.Says(), length) {
				return fmt.Errorf("%s says %q, want %s", day, one.Says(), length)
			}
			return nil
		}
		return fmt.Errorf("%s shows no card titled %q", day, title)
	})
	sc.Step(`^I click the card "([^"]*)" on the time page$`, func(title string) error {
		at, err := cardOfWeek(title)
		if err != nil {
			return err
		}
		w.w = w.w.clicked(at)
		return nil
	})
	sc.Step(`^I double-click the card "([^"]*)" on the time page$`, func(title string) error {
		at, err := cardOfWeek(title)
		if err != nil {
			return err
		}
		w.w = w.w.doubleClicked(at)
		return nil
	})
	sc.Step(`^I double-click "([^"]*)" in the month's day "([^"]*)"$`, func(title, day string) error {
		i, err := dayOfMonth(day)
		if err != nil {
			return err
		}
		for j, one := range w.w.MonthDay(i).Tasks {
			if one.Title == title {
				box := CellLineAt(MonthCellAt(w.w.Layout(), i), j)
				w.w = w.w.doubleClicked(Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
				return nil
			}
		}
		return fmt.Errorf("%s holds no task titled %q", day, title)
	})
	sc.Step(`^I rest the pointer on the name of the field "([^"]*)"$`, func(name string) error {
		for i, field := range w.w.form.Fields() {
			if field.Name == name {
				box, _, _ := FormPlace(w.w.formBox(), w.w.form.Fields(), i)
				w.w = w.w.Hovering(Point{X: box.X + 2, Y: box.Y + box.H/2})
				return nil
			}
		}
		return fmt.Errorf("the form has no field %q", name)
	})
	sc.Step(`^I rest the pointer on the (title|foot|objective mark) of the card "([^"]*)"$`, func(part, title string) error {
		for _, placed := range w.w.Placed() {
			if placed.Card.Title != title {
				continue
			}
			parts := PartsOf(placed.Slot)
			on := map[string]Box{
				"title":          parts.Title,
				"foot":           parts.Steps,
				"objective mark": CardRowOf(placed.Card, parts.Number, w.w.measured(), w.w.now, engine.Pages{}).ServesAt,
			}[part]
			w.w = w.w.Hovering(Point{X: on.X + on.W/2, Y: on.Y + on.H/2})
			return nil
		}
		return fmt.Errorf("the board holds no card titled %q", title)
	})
	sc.Then(`^a bubble shows "([^"]*)"$`, func(said string) error {
		if got := w.w.Bubble(); got != said {
			return fmt.Errorf("the bubble shows %q, not %q", got, said)
		}
		return nil
	})
	titleOfCard := func(title string) (Box, Box, error) {
		bubble, _, shown := w.w.BubbleBox()
		if !shown {
			return Box{}, Box{}, fmt.Errorf("no bubble shows")
		}
		for _, placed := range w.w.Placed() {
			if placed.Card.Title == title {
				return bubble, PartsOf(placed.Slot).Title, nil
			}
		}
		return Box{}, Box{}, fmt.Errorf("no card titled %q is on the board", title)
	}
	sc.Then(`^the bubble stands just above the title of the card "([^"]*)"$`, func(title string) error {
		bubble, under, err := titleOfCard(title)
		if err != nil {
			return err
		}
		if gap := under.Y - (bubble.Y + bubble.H); gap < 0 || gap > 6 {
			return fmt.Errorf("the bubble ends %dpx above the title, want 0 to 6", gap)
		}
		return nil
	})
	sc.Then(`^the bubble's text begins where the title of the card "([^"]*)" begins$`, func(title string) error {
		bubble, under, err := titleOfCard(title)
		if err != nil {
			return err
		}
		if begins := ExplainedLine(bubble, 0).X; begins != under.X {
			return fmt.Errorf("the bubble's text begins at %d, the title at %d", begins, under.X)
		}
		return nil
	})
	sc.Then(`^no bubble shows$`, func() error {
		if got := w.w.Bubble(); got != "" {
			return fmt.Errorf("a bubble shows %q", got)
		}
		return nil
	})
	sc.Step(`^the pointer leaves the window$`, func() error {
		w.w = w.w.Away()
		return nil
	})
	sc.Then(`^the window explains "([^"]*)"$`, func(said string) error {
		if got := w.w.Explains(); !strings.Contains(got, said) {
			return fmt.Errorf("the window explains %q, want %q", got, said)
		}
		return nil
	})
	sc.Then(`^the explanation gives the example "([^"]*)"$`, func(want string) error {
		if got := w.w.ExplainsBy(); got != want {
			return fmt.Errorf("the explanation gives the example %q, want %q", got, want)
		}
		return nil
	})
	sc.Then(`^the explanation is no bigger than what it says$`, func() error {
		tip, said, ok := w.w.Explanation()
		if !ok {
			return fmt.Errorf("nothing is explained")
		}
		inside := Inside(tip)
		if inside.H > len(said)*lineHigh+inset {
			return fmt.Errorf("an explanation of %d lines is %d high, room for %d", len(said), tip.H, inside.H/lineHigh)
		}
		widest := 0
		for _, line := range said {
			widest = max(widest, ends(w.w.measured(), line, small))
		}
		if inside.W > widest+2*padding+glyphOf(small) {
			return fmt.Errorf("an explanation whose widest line is %d is %d wide", widest, tip.W)
		}
		return nil
	})
	sc.Then(`^the explanation stays inside the window$`, func() error {
		tip, _, ok := w.w.Explanation()
		if !ok {
			return fmt.Errorf("nothing is explained")
		}
		if tip.X < 0 || tip.Y < 0 || tip.X+tip.W > w.w.width || tip.Y+tip.H > w.w.height {
			return fmt.Errorf("the explanation runs to %+v in a window of %dx%d", tip, w.w.width, w.w.height)
		}
		return nil
	})
	sc.Then(`^the window explains nothing$`, func() error {
		if got := w.w.Explains(); got != "" {
			return fmt.Errorf("the window explains %q", got)
		}
		return nil
	})
	offerBox := func(offer string) (Box, error) {
		bar := w.w.BarOffers()
		for i, one := range w.w.Offers() {
			if one == offer {
				return bar[i], nil
			}
		}
		return Box{}, fmt.Errorf("the bar does not offer %q", offer)
	}
	sc.Step(`^I rest the pointer on the offer "([^"]*)"$`, func(offer string) error {
		box, err := offerBox(offer)
		if err != nil {
			return err
		}
		w.w = w.w.Hovering(Point{X: box.X + 2, Y: box.Y + 2})
		return nil
	})
	sc.Then(`^the offer "([^"]*)" is a lit hint, its key on a chip$`, func(offer string) error {
		box, err := offerBox(offer)
		if err != nil {
			return err
		}
		if !w.w.Lighting(offer) {
			return fmt.Errorf("the offer %q is not lit", offer)
		}
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		if painted(shot, Box{X: box.X + 1, Y: box.Y + box.H/2, W: 1, H: 1}, Dark().Colour.Surface.Hover) != 1 {
			return fmt.Errorf("the hint %q at %+v is not lit", offer, box)
		}
		chip := Box{X: box.X + hintSpace/2 + chipInset - 2, Y: box.Y + box.H/2 - chipHigh/2 + 2, W: 1, H: 1}
		if painted(shot, chip, Dark().Colour.Surface.Control) != 1 {
			return fmt.Errorf("the key of %q is not on a chip at %+v", offer, chip)
		}
		return nil
	})
	sc.Then(`^every offer of the panel is drawn as a button$`, func() error {
		slots := w.w.offersAlong(w.w.panel())
		grounds := map[buttonLook]color.NRGBA{buttonPlain: Dark().Colour.Surface.Control, buttonOn: Dark().Colour.Accent.Base, buttonDanger: Dark().Colour.State.Danger}
		for i, slot := range slots {
			look := w.w.OfferLook(i, w.w.Offers()[i])
			if err := buttonsDrawn(w.w, []Box{ButtonIn(slot)}, grounds[look]); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Then(`^the day held is drawn as a filled button$`, func() error {
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		at := DayAt(w.w.CalendarBox(), dayIndex(w.w, w.w.day))
		middle := Box{X: at.X + Dark().Radius.Control, Y: at.Y + 1, W: at.W - 2*Dark().Radius.Control, H: 2}
		if got := painted(shot, middle, Dark().Colour.Accent.Base); got != middle.W*middle.H {
			return fmt.Errorf("the day held, at %+v, is filled on %d of %d pixels", at, got, middle.W*middle.H)
		}
		return nil
	})
	sc.Then(`^today is drawn as an edged button$`, func() error {
		at := DayAt(w.w.CalendarBox(), dayIndex(w.w, w.w.today()))
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		if inside := (Box{X: at.X + Dark().Radius.Control, Y: at.Y + 3, W: at.W - 2*Dark().Radius.Control, H: 2}); painted(shot, inside, Dark().Colour.Surface.Raised) != inside.W*inside.H {
			return fmt.Errorf("today, at %+v, is filled", at)
		}
		if edge := (Box{X: at.X + Dark().Radius.Control, Y: at.Y, W: at.W - 2*Dark().Radius.Control, H: 1}); painted(shot, edge, Dark().Colour.Accent.Base) != edge.W {
			return fmt.Errorf("today, at %+v, is not ringed", at)
		}
		return nil
	})
	sc.Then(`^the other days are drawn without a box$`, func() error {
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		for i, day := range engine.MonthGrid(w.w.day) {
			if day.Equal(w.w.day) || day.Equal(w.w.today()) {
				continue
			}
			if edge := Edges(DayAt(w.w.CalendarBox(), i))[0]; painted(shot, edge, Dark().Colour.Surface.Raised) != edge.W {
				return fmt.Errorf("%s is drawn in a box", day.Format(time.DateOnly))
			}
		}
		return nil
	})
	sc.Then(`^every month is drawn as a button$`, func() error {
		var cells []Box
		for i := range w.w.CalendarCells() {
			cells = append(cells, CalendarCell(w.w.CalendarBox(), i))
		}
		return buttonsDrawn(w.w, cells, Dark().Colour.Surface.Control, Dark().Colour.Accent.Base)
	})
	sc.Then(`^the calendar's month, year and arrows are drawn as buttons$`, func() error {
		box := w.w.CalendarBox()
		return buttonsDrawn(w.w, []Box{CalendarMonth(box), CalendarYear(box), MonthBefore(box), MonthAfter(box)}, Dark().Colour.Surface.Control, Dark().Colour.Accent.Base)
	})
	sc.Then(`^the labels list "([^"]*)" carried by (\d+) tasks? and (\d+) objectives?$`, func(label string, tasks, objectives int) error {
		if w.w.Screen() != OnLabels {
			return fmt.Errorf("the labels are not listed: the window is on %v", w.w.Screen())
		}
		for _, one := range w.w.LabelsCarried() {
			if one.Label == label {
				if one.Tasks != tasks || one.Objectives != objectives {
					return fmt.Errorf("the label %q is carried by %+v", label, one)
				}
				return nil
			}
		}
		return fmt.Errorf("the labels list no %q: %+v", label, w.w.LabelsCarried())
	})
	sc.Then(`^the labels list no "([^"]*)"$`, func(label string) error {
		for _, one := range w.w.LabelsCarried() {
			if one.Label == label {
				return fmt.Errorf("the labels still list %q", label)
			}
		}
		return nil
	})
	sc.Then(`^the labels are listed$`, func() error {
		if w.w.Screen() != OnLabels {
			return fmt.Errorf("the window is on %v", w.w.Screen())
		}
		return nil
	})
	labelRow := func(label string) (LabelRow, error) {
		for _, row := range w.w.LabelRows() {
			if !row.Typing && row.Carried.Label == label {
				return row, nil
			}
		}
		return LabelRow{}, fmt.Errorf("no label %q is in sight", label)
	}
	clicking := func(b Box) { w.w = w.w.clicked(Point{X: b.X + b.W/2, Y: b.Y + b.H/2}) }
	sc.Step(`^I click the label "([^"]*)"$`, func(label string) error {
		row, err := labelRow(label)
		if err != nil {
			return err
		}
		clicking(row.Name)
		return nil
	})
	sc.Step(`^I click the dot of the label "([^"]*)"$`, func(label string) error {
		row, err := labelRow(label)
		if err != nil {
			return err
		}
		clicking(row.Dot)
		return nil
	})
	sc.Step(`^I click "([^"]*)" in the row of the label "([^"]*)"$`, func(offer, label string) error {
		row, err := labelRow(label)
		if err != nil {
			return err
		}
		for _, one := range row.Actions {
			if one.Offer == offer {
				clicking(one.At)
				return w.w.Wrong()
			}
		}
		return fmt.Errorf("the row of %q offers no %q: %+v", label, offer, row.Actions)
	})
	rowOffers := func(label string) ([]string, error) {
		row, err := labelRow(label)
		if err != nil {
			return nil, err
		}
		var offers []string
		for _, one := range row.Actions {
			offers = append(offers, one.Offer)
		}
		return offers, nil
	}
	sc.Then(`^the row of the label "([^"]*)" offers "([^"]*)", "([^"]*)" and "([^"]*)"$`, func(label, a, b, c string) error {
		offers, err := rowOffers(label)
		if err != nil {
			return err
		}
		if !slices.Equal(offers, []string{a, b, c}) {
			return fmt.Errorf("the row of %q offers %v", label, offers)
		}
		return nil
	})
	sc.Then(`^the row of the label "([^"]*)" offers nothing$`, func(label string) error {
		offers, err := rowOffers(label)
		if err != nil {
			return err
		}
		if len(offers) > 0 {
			return fmt.Errorf("the row of %q offers %v", label, offers)
		}
		return nil
	})
	sc.Then(`^the panel offers "([^"]*)" and "([^"]*)"$`, func(a, b string) error {
		if got := w.w.Offers(); !slices.Equal(got, []string{a, b}) {
			return fmt.Errorf("the panel offers %v", got)
		}
		return nil
	})
	sc.Then(`^the labels in sight are "([^"]*)"$`, func(want string) error {
		var names []string
		for _, row := range w.w.LabelRows() {
			if !row.Typing {
				names = append(names, row.Carried.Label)
			}
		}
		if got := strings.Join(names, ", "); got != want {
			return fmt.Errorf("the labels in sight are %q", got)
		}
		return nil
	})
	sc.Then(`^the labels are picked under the card "([^"]*)"$`, func(title string) error {
		if w.w.Screen() != OnChoosing || !w.w.Picking() {
			return fmt.Errorf("the window is on %v, picking %v", w.w.Screen(), w.w.Picking())
		}
		for _, placed := range w.w.Placed() {
			if placed.Card.Title != title {
				continue
			}
			box, slot := w.w.ChoosingBox(), placed.Slot
			if box.Y < slot.Y+slot.H || box.X != slot.X {
				return fmt.Errorf("the labels are at %+v, not under the card at %+v", box, slot)
			}
			return nil
		}
		return fmt.Errorf("no card %q is placed", title)
	})
	sc.Then(`^the list says "([^"]*)" is carried by (\d+) tasks?$`, func(label string, n int) error {
		for _, one := range w.w.Choosing() {
			if one.Title == label {
				if one.Carried != n {
					return fmt.Errorf("the list says %q is carried by %d", label, one.Carried)
				}
				return nil
			}
		}
		return fmt.Errorf("the list offers no %q", label)
	})
	sc.Then(`^the labels taken read "([^"]*)"$`, func(want string) error {
		var got []string
		for _, one := range w.w.LabelsTaken() {
			got = append(got, one.Label)
		}
		if strings.Join(got, ", ") != want {
			return fmt.Errorf("the labels taken read %q", strings.Join(got, ", "))
		}
		return nil
	})
	sc.When(`^I click the cross of the label taken "([^"]*)"$`, func(label string) error {
		for _, one := range w.w.LabelsTaken() {
			if one.Label == label {
				clicking(one.Cross)
				return w.w.Wrong()
			}
		}
		return fmt.Errorf("no label taken is %q", label)
	})
	sc.Then(`^the labels offer to create "([^"]*)"$`, func(want string) error {
		said, ok := w.w.LabelCreating()
		if !ok || said != want {
			return fmt.Errorf("the labels offer to create %q (%v)", said, ok)
		}
		return nil
	})
	sc.Then(`^the window asks "([^"]*)"$`, func(question string) error {
		if w.w.Screen() != OnAsking || Heads(w.w) != question {
			return fmt.Errorf("the window is on %v and asks %q", w.w.Screen(), Heads(w.w))
		}
		return nil
	})
	sc.Then(`^task (\d+) carries exactly the labels "([^"]*)"$`, func(n int, want string) error {
		shown, err := w.other.Ask(engine.TaskShow, engine.Input{Given: map[string]string{"number": strconv.Itoa(n)}})
		if err != nil {
			return err
		}
		if got := strings.Join(task.OfTask(*shown.Task).Labels, ", "); got != want {
			return fmt.Errorf("task %d carries %q, want %q", n, got, want)
		}
		return nil
	})
	sc.Then(`^the window is known to the system as "([^"]*)"$`, func(name string) error {
		if Known != name {
			return fmt.Errorf("the window is known as %q", Known)
		}
		return nil
	})
	sc.Then(`^the icon it hands the system is cairn's logo at (\d+) pixels, its corners clear$`, func(size int) error {
		drawn, err := png.Decode(bytes.NewReader(Logo()))
		if err != nil {
			return err
		}
		if b := drawn.Bounds(); b.Dx() != size || b.Dy() != size {
			return fmt.Errorf("the logo is %v", b)
		}
		if _, _, _, a := drawn.At(1, 1).RGBA(); a != 0 {
			return fmt.Errorf("the logo's corner is not clear")
		}
		if _, _, _, a := drawn.At(size/2, size/2).RGBA(); a == 0 {
			return fmt.Errorf("the logo's middle is clear")
		}
		return nil
	})
	sc.Then(`^the icon it hands X11 holds cairn's logo at (\d+), (\d+), (\d+), (\d+) and (\d+) pixels, its corners clear$`, func(a, b, c, d, e int) error {
		held := X11Icon()
		for _, size := range []int{a, b, c, d, e} {
			if len(held) < 2 || held[0] != uint32(size) || held[1] != uint32(size) || len(held) < 2+size*size {
				return fmt.Errorf("the icon does not hold the logo at %d pixels next", size)
			}
			pixels := held[2 : 2+size*size]
			if pixels[0]>>24 != 0 {
				return fmt.Errorf("the logo at %d pixels has a corner that is not clear", size)
			}
			if pixels[size*size/2+size/2]>>24 == 0 {
				return fmt.Errorf("the logo at %d pixels has a clear middle", size)
			}
			held = held[2+size*size:]
		}
		if len(held) != 0 {
			return fmt.Errorf("the icon holds %d values more than the logos named", len(held))
		}
		return nil
	})
	var data string
	dataHome := func() string {
		if data == "" {
			data = w.t.TempDir()
		}
		return data
	}
	entry := func(name string) string { return filepath.Join(dataHome(), "applications", name+".desktop") }
	elsewhere := ""
	place := func() string {
		if elsewhere == "" {
			elsewhere = w.t.TempDir()
		}
		return elsewhere
	}
	running := func() string { return filepath.Join(place(), "running", Known) }
	pathTo := ""
	onPath := func(name string) (string, error) {
		if pathTo == "" {
			return "", fmt.Errorf("%s is not on the PATH", name)
		}
		return pathTo, nil
	}
	sc.Given(`^the cairn running is the one the PATH leads to as "([^"]*)"$`, func(as string) error {
		if err := os.MkdirAll(filepath.Dir(running()), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(running(), nil, 0o755); err != nil {
			return err
		}
		pathTo = filepath.Join(place(), as)
		if err := os.MkdirAll(filepath.Dir(pathTo), 0o755); err != nil {
			return err
		}
		return os.Symlink(running(), pathTo)
	})
	sc.Given(`^the cairn running is not the one the PATH leads to$`, func() error {
		pathTo = ""
		return nil
	})
	sc.Step(`^the window settles in on a Linux desktop$`, func() error {
		return linuxapp.SettleIn(dataHome(), linuxapp.Starter(running(), onPath))
	})
	sc.Then(`^the desktop entry "([^"]*)" starts the full path of "([^"]*)" with "([^"]*)"$`, func(name, as, with string) error {
		return entrySays(entry(name), "Exec="+filepath.Join(place(), as)+" "+with)
	})
	sc.Then(`^the desktop entry "([^"]*)" starts the cairn running with "([^"]*)"$`, func(name, with string) error {
		return entrySays(entry(name), "Exec="+running()+" "+with)
	})
	sc.Given(`^a desktop entry "([^"]*)" written by somebody else$`, func(name string) error {
		if err := os.MkdirAll(filepath.Dir(entry(name)), 0o755); err != nil {
			return err
		}
		return os.WriteFile(entry(name), []byte("[Desktop Entry]\nName=mine\n"), 0o644)
	})
	sc.Given(`^a desktop entry "([^"]*)" written by somebody else, showing the icon "([^"]*)"$`, func(name, icon string) error {
		if err := os.MkdirAll(filepath.Dir(entry(name)), 0o755); err != nil {
			return err
		}
		return os.WriteFile(entry(name), []byte("[Desktop Entry]\nName=Cairn\nIcon="+icon+"\nExec=cairn gui\n"), 0o644)
	})
	sc.Step(`^the window opens on a Linux desktop$`, func() error {
		if err := w.open(); err != nil {
			return err
		}
		w.w = w.w.Settling(dataHome(), linuxapp.Starter(running(), onPath))
		return nil
	})
	sc.Step(`^somebody changes the desktop entry "([^"]*)"$`, func(name string) error {
		was, err := os.ReadFile(entry(name))
		if err != nil {
			return err
		}
		return os.WriteFile(entry(name), append(was, []byte("Comment=Mes tâches\n")...), 0o644)
	})
	sc.Then(`^the desktop entry "([^"]*)" still shows the icon "([^"]*)"$`, func(name, icon string) error {
		return entrySays(entry(name), "Icon="+icon)
	})
	sc.Then(`^the desktop entry "([^"]*)" is the one somebody else wrote$`, func(name string) error {
		got, err := os.ReadFile(entry(name))
		if err != nil {
			return err
		}
		if string(got) != "[Desktop Entry]\nName=mine\n" {
			return fmt.Errorf("the entry was rewritten:\n%s", got)
		}
		return nil
	})
	sc.Then(`^the desktop entry "([^"]*)" names the icon "([^"]*)"$`, func(name, icon string) error {
		return entrySays(entry(name), "Icon="+icon, "StartupWMClass="+Known)
	})
	sc.Then(`^the icon "([^"]*)" is there at ([\d, ]+) and (\d+) pixels$`, func(icon, some, last string) error {
		for _, named := range append(strings.Split(some, ", "), last) {
			size, err := strconv.Atoi(named)
			if err != nil {
				return err
			}
			held, err := os.ReadFile(filepath.Join(dataHome(), "icons", "hicolor", named+"x"+named, "apps", icon+".png"))
			if err != nil {
				return err
			}
			drawn, err := png.Decode(bytes.NewReader(held))
			if err != nil {
				return err
			}
			if b := drawn.Bounds(); b.Dx() != size || b.Dy() != size {
				return fmt.Errorf("the icon at %d pixels is %v", size, b)
			}
		}
		return nil
	})
	sc.Then(`^the day "([^"]*)" carries nothing$`, func(day string) error {
		if d, _ := dayOfWeek(day); len(d.Tasks) > 0 {
			return fmt.Errorf("%s carries %d tasks", day, len(d.Tasks))
		}
		return nil
	})
	sc.Then(`^the window tells "([^"]*)"$`, func(said string) error {
		if !strings.Contains(w.w.Note(), said) {
			return fmt.Errorf("the window tells %q, want it to say %q", w.w.Note(), said)
		}
		return nil
	})
	sc.Then(`^there (?:is|are) (\d+) backups?$`, func(n int) error {
		kept, err := w.work.Ask(engine.BackupList, engine.Input{})
		if err != nil {
			return err
		}
		if len(kept.Backups) != n {
			return fmt.Errorf("there are %d backups, want %d", len(kept.Backups), n)
		}
		return nil
	})
	sc.Then(`^the backups listed are (\d+)$`, func(n int) error {
		if w.w.Screen() != OnBackups {
			return fmt.Errorf("the window is on %v, not on the backups", w.w.Screen())
		}
		if got := len(w.w.Backups()); got != n {
			return fmt.Errorf("%d backups are listed, want %d", got, n)
		}
		return nil
	})
	sc.Then(`^the window is asking$`, func() error {
		if w.w.Screen() != OnAsking {
			return fmt.Errorf("the window is on %v, not asking", w.w.Screen())
		}
		return nil
	})
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
	sc.Then(`^what is in hand says what it carries$`, func() error {
		ghost, dragging := w.w.Ghost()
		if !dragging {
			return fmt.Errorf("nothing is in hand")
		}
		inside := Box{X: ghost.X + 2, Y: ghost.Y + 2, W: ghost.W - 4, H: ghost.H - 4}
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		if inked(shot, inside, Dark().Colour.Surface.Ghost) == 0 {
			return fmt.Errorf("nothing is written on what is in hand, at %+v", ghost)
		}
		return nil
	})

	sc.Then(`^it shows where the tile would land$`, func() error {
		landing, somewhere := w.w.Landing()
		if !somewhere {
			return fmt.Errorf("nothing says where the tile would land")
		}
		box := BesideBox(w.w.formPanel())
		if landing.X != box.X || landing.W != box.W {
			return fmt.Errorf("it lands at %+v, and the box is %+v", landing, box)
		}
		return nil
	})
	sc.Then(`^the board behind shows no landing$`, func() error {
		landing, somewhere := w.w.Landing()
		if !somewhere {
			return nil
		}
		for i := range engine.Columns {
			if w.w.Layout().Column[i].X == landing.X {
				return fmt.Errorf("the landing at %+v is column %d's, not the box's", landing, i)
			}
		}
		return nil
	})
	sc.Then(`^it shows where the card would land in "([^"]*)"$`, w.landsIn)
	sc.Then(`^it shows nowhere for the card to land$`, w.landsNowhere)
	sc.Then(`^"([^"]*)" does not carry "([^"]*)"$`, w.columnDoesNotCarry)
	sc.When(`^I click the filter "([^"]*)"$`, func(name string) error {
		for i, f := range w.w.Filters() {
			if f.Name == name {
				row := w.w.FilterRowAt(i)
				w.w = w.w.clicked(Point{X: row.X + row.W/2, Y: row.Y + row.H/2})
				return nil
			}
		}
		return fmt.Errorf("no filter %q", name)
	})
	sc.Then(`^the filter "([^"]*)" lists (its values|nothing)$`, func(name, what string) error {
		values, open := w.w.FilterList()
		if what == "nothing" {
			if open {
				return fmt.Errorf("a filter's list is still open: %v", values)
			}
			return nil
		}
		if !open || w.w.Filters()[w.w.FilterOn()].Name != name || len(values) != len(w.w.Filters()[w.w.FilterOn()].Shown) {
			return fmt.Errorf("the filter %q lists %v, open %v", name, values, open)
		}
		return nil
	})
	sc.Then(`^the labels offer (\d+) colours$`, func(n int) error {
		if got := len(w.w.ColourSwatches()); got != n {
			return fmt.Errorf("the labels offer %d colours", got)
		}
		return nil
	})
	sc.When(`^I click the colour "([^"]*)"$`, func(colour string) error {
		for _, swatch := range w.w.ColourSwatches() {
			if swatch.Colour == colour {
				w.w = w.w.clicked(Point{X: swatch.At.X + swatch.At.W/2, Y: swatch.At.Y + swatch.At.H/2})
				return w.w.Wrong()
			}
		}
		return fmt.Errorf("no swatch is %q", colour)
	})
	sc.Then(`^the label "([^"]*)" wears "([^"]*)"$`, func(label, colour string) error {
		for _, one := range w.w.LabelsCarried() {
			if one.Label == label {
				if one.Colour != colour {
					return fmt.Errorf("%q wears %q, not %q", label, one.Colour, colour)
				}
				return nil
			}
		}
		return fmt.Errorf("no label %q is listed", label)
	})
	sc.Then(`^the field "([^"]*)" stands beside the description$`, func(name string) error {
		fields := w.w.Form().Fields()
		at, description := -1, -1
		for i, f := range fields {
			switch f.Name {
			case name:
				at = i
			case "description":
				description = i
			}
		}
		if at < 0 || description < 0 {
			return fmt.Errorf("the form holds no %q or no description", name)
		}
		_, value, _ := FormPlace(w.w.formBox(), fields, at)
		_, said, _ := FormPlace(w.w.formBox(), fields, description)
		if value.X < said.X+said.W {
			return fmt.Errorf("%q at %+v is not beside the description at %+v", name, value, said)
		}
		return nil
	})
	sc.Then(`^the description takes (\d+) parts of the height and the steps (\d+)$`, func(description, steps int) error {
		fields := w.w.Form().Fields()
		rows := map[string]int{}
		for i, f := range fields {
			_, _, given := FormPlace(w.w.formBox(), fields, i)
			rows[f.Name] = given
		}
		total := rows["description"] + rows["steps"]
		if want := total * description / (description + steps); rows["description"] != want {
			return fmt.Errorf("the description is given %d rows of %d, want %d", rows["description"], total, want)
		}
		return nil
	})
	sc.Then(`^the column beside the form shows the (details|trace)$`, func(what string) error {
		if got := w.w.Tracing(); got != (what == "trace") {
			return fmt.Errorf("the column beside the form shows the trace: %v", got)
		}
		return nil
	})
	sc.Then(`^the work stands under the description$`, func() error {
		fields := w.w.Form().Fields()
		for i, f := range fields {
			if f.Name == "description" {
				_, value, rows := FormPlace(w.w.formBox(), fields, i)
				said, work := Framed(value, rows), w.w.workFrame()
				if work.Y < said.Y+said.H || work.X != said.X {
					return fmt.Errorf("the work at %+v is not under the description at %+v", work, said)
				}
				return nil
			}
		}
		return fmt.Errorf("the form has no description")
	})
	sc.Then(`^the column beside the form offers no tabs$`, func() error {
		if w.w.TabsBeside() {
			return fmt.Errorf("the column beside the form offers tabs")
		}
		return nil
	})
	sc.When(`^I click the work "([^"]*)"$`, func(title string) error {
		for _, tile := range w.w.WorkTiles() {
			if tile.Said == title {
				w.w = w.w.clicked(Point{X: tile.At.X + tile.At.W/2, Y: tile.At.Y + tile.At.H/2})
				return w.w.Wrong()
			}
		}
		return fmt.Errorf("no work reads %q: %+v", title, w.w.WorkTiles())
	})
	sc.Then(`^the keyboard is on the work "([^"]*)"$`, func(title string) error {
		tile, on := w.w.WorkOn()
		if !on || tile.Said != title {
			return fmt.Errorf("the keyboard is on the work %+v (%v), not %q", tile, on, title)
		}
		return nil
	})
	sc.Then(`^the help is open$`, w.helpOpen)
	sc.Then(`^the help is closed$`, w.helpClosed)
	sc.Then(`^the help says what "([^"]*)" does$`, w.helpSaysAbout)
	sc.Then(`^the help says nothing of "([^"]*)"$`, func(about string) error {
		if w.helpSaysAbout(about) == nil {
			return fmt.Errorf("the help still says %q", about)
		}
		return nil
	})
	sc.Then(`^the help shows every line$`, func() error {
		box := w.w.HelpBox()
		foot := PanelOfferAt(box, 0).Y
		for _, line := range w.w.HelpLines() {
			if line.At.Y < box.Y || line.At.Y+line.At.H > foot {
				return fmt.Errorf("the line %q at %+v is out of the help %+v, whose foot is at %d", line.Key+line.Head, line.At, box, foot)
			}
		}
		return nil
	})
	sc.Then(`^the help lays its keys out in (\d+) columns$`, func(n int) error {
		columns := map[int]bool{}
		for _, line := range w.w.HelpLines() {
			columns[line.At.X] = true
		}
		if len(columns) != n {
			return fmt.Errorf("the help lays its keys out in %d columns", len(columns))
		}
		return nil
	})
	sc.Then(`^the help groups its keys under (.+)$`, func(list string) error {
		var heads []string
		for _, line := range w.w.HelpLines() {
			if line.Head != "" {
				heads = append(heads, line.Head)
			}
		}
		if want := quoted(strings.ReplaceAll(list, " and ", ", ")); strings.Join(heads, "|") != strings.Join(want, "|") {
			return fmt.Errorf("the help groups its keys under %v, want %v", heads, want)
		}
		return nil
	})
	sc.Then(`^the help is no taller than what it says$`, func() error {
		box, over := w.w.HelpBox(), Over(w.width, w.high)
		if box.H >= over.H {
			return fmt.Errorf("the help is %d high, as tall as the room it was given", box.H)
		}
		return nil
	})
	sc.Then(`^every line of the help fits inside its panel$`, func() error {
		box := w.w.HelpBox()
		for _, line := range w.w.HelpLines() {
			said := ends(w.w.measured(), line.Means+line.Head, reads)
			if end := line.MeansAt.X + said; line.Head == "" && end > line.At.X+line.At.W {
				return fmt.Errorf("%q ends at %d, past its column %+v", line.Means, end, line.At)
			}
			if line.At.X+line.At.W > box.X+box.W-padding {
				return fmt.Errorf("the column %+v runs past the help %+v", line.At, box)
			}
		}
		return nil
	})
	sc.Then(`^the calendar is open$`, func() error {
		if w.w.Screen() != OnCalendar {
			return fmt.Errorf("the window is on %v", w.w.Screen())
		}
		return nil
	})
	sc.Then(`^the time page shows a (week|month)$`, func(what string) error {
		if w.w.Tab() != Time || w.w.Monthly() != (what == "month") || w.w.Daily() {
			return fmt.Errorf("the time page on %v shows monthly %v daily %v", w.w.Tab(), w.w.Monthly(), w.w.Daily())
		}
		return nil
	})
	sc.Then(`^the window is searching "([^"]*)"$`, func(said string) error {
		if w.w.Screen() != OnSearch || w.w.Searched().Reads() != said {
			return fmt.Errorf("the window is on %v searching %q", w.w.Screen(), w.w.Searched().Reads())
		}
		return nil
	})
	sc.Then(`^the search "([^"]*)" is kept as a pill$`, func(said string) error {
		if _, ok := w.w.SearchPill(); !ok || w.w.Screen() != OnTheBoard || w.w.Searched().Reads() != said {
			return fmt.Errorf("the window on %v keeps the search %q, pill %v", w.w.Screen(), w.w.Searched().Reads(), ok)
		}
		return nil
	})
	sc.Step(`^I click the cross of the search's pill$`, func() error {
		pill, ok := w.w.SearchPill()
		if !ok {
			return fmt.Errorf("no search is kept as a pill")
		}
		w.w = w.w.clicked(Point{X: pill.X + pill.W - pillCross/2, Y: pill.Y + pill.H/2})
		return nil
	})
	sc.Then(`^no search is kept$`, func() error {
		if _, ok := w.w.SearchPill(); ok || !w.w.Searched().Empty() {
			return fmt.Errorf("the search %q is kept", w.w.Searched().Reads())
		}
		return nil
	})
	sc.Then(`^the command bar is (open|shut)$`, func(how string) error {
		if open := w.w.Screen() == OnPalette; open != (how == "open") {
			return fmt.Errorf("the command bar open is %v, the window on %v", open, w.w.Screen())
		}
		return nil
	})
	sc.Then(`^the command bar offers "([^"]*)" first$`, func(said string) error {
		rows := w.w.PaletteRows()
		if len(rows) == 0 || rows[0].Says != said {
			return fmt.Errorf("the command bar offers %v", rows)
		}
		return nil
	})
	sc.Then(`^the command bar stands at the top of the window, in its middle$`, func() error {
		box, l := w.w.PaletteBox(), w.w.Layout()
		if box.Y > l.Tabs.Y+l.Tabs.H+padding || box.Y+box.H > l.Bar.Y || box.X+box.W/2 != w.w.width/2 {
			return fmt.Errorf("the command bar is at %+v, the tabs at %+v, the bar at %+v", box, l.Tabs, l.Bar)
		}
		return nil
	})
	sc.Then(`^the command bar offers the task "([^"]*)"$`, func(title string) error {
		for _, row := range w.w.PaletteRows() {
			if row.Group == paletteTasksGroup && strings.HasSuffix(row.Says, title) {
				return nil
			}
		}
		return fmt.Errorf("the command bar offers %v", w.w.PaletteRows())
	})
	sc.Then(`^the command bar reads "([^"]*)"$`, func(said string) error {
		if got := w.w.CommandSays(); got != said {
			return fmt.Errorf("the command bar reads %q", got)
		}
		return nil
	})
	sc.Then(`^the bar's line has room to be read$`, func() error {
		if line := CommandLine(w.w.Layout(), w.w.BarOffers()); line.W < barLineLeast {
			return fmt.Errorf("the bar's line is %d wide, under the %d it keeps", line.W, barLineLeast)
		}
		return nil
	})
	sc.Then(`^every hint of the bar is its key alone, inside the window$`, func() error {
		if w.w.barWorded() {
			return fmt.Errorf("the bar's hints keep their words at %d wide", w.w.width)
		}
		for i, box := range w.w.BarOffers() {
			if box.X < 0 || box.X+box.W > w.w.width {
				return fmt.Errorf("hint %d runs from %d to %d, outside a window %d wide", i, box.X, box.X+box.W, w.w.width)
			}
		}
		return nil
	})
	sc.Then(`^every hint of the bar has its words$`, func() error {
		if !w.w.barWorded() {
			return fmt.Errorf("the bar's hints are their keys alone at %d wide", w.w.width)
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
	sc.Then(`^the card "([^"]*)" carries a glowing dot at the right of its first line$`, w.cardCarriesAGlow)
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

	sc.Then(`^the second tile says what it holds$`, func() error {
		tile := w.w.ServingCardAt(1)
		line := RowIn(Box{X: tile.X, Y: tile.Y, W: tile.W, H: tile.H - between}, tile.Y+inset, lineHigh)
		band := Box{X: line.X, Y: line.Y, W: line.W, H: line.H - inset}
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		var reached int
		for x := band.X; x < band.X+band.W; x++ {
			if inked(shot, Box{X: x, Y: band.Y, W: 1, H: band.H}, Dark().Colour.Surface.Raised) > 0 {
				reached = x - band.X
			}
		}
		if reached < line.W/2 {
			return fmt.Errorf("the tile writes across %d of its %d: a long title fills its line", reached, line.W)
		}
		return nil
	})

	sc.Then(`^the second tile leaves room before its frame$`, func() error {
		tile := w.w.ServingCardAt(1)
		edge := Box{X: tile.X + tile.W - padding, Y: tile.Y, W: padding, H: tile.H}
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		for what, ink := range map[string]color.NRGBA{"faint": Dark().Colour.Ink.Faint, "ink": Dark().Colour.Ink.Text} {
			if got := painted(shot, edge, ink); got != 0 {
				return fmt.Errorf("%d pixels of %s writing reach the tile's last %d", got, what, padding)
			}
		}
		return nil
	})

	sc.Then(`^the search is not in force$`, func() error {
		if !w.w.Searched().Empty() {
			return fmt.Errorf("the board is still narrowed to %q", w.w.looking.Reads())
		}
		return nil
	})

	sc.Then(`^the list of objectives is open$`, w.listOpen)
	sc.Then(`^the form is still drawn behind it$`, w.formBehindTheList)
	sc.Then(`^the form is drawn behind the question$`, func() error {
		asked, over := w.w.QuestionBox(), Over(w.w.width, w.w.height)
		fields := w.w.Form().Fields()
		frame := Framed(formValue(over, fields, 0), formRows(over, fields, 0))
		if frame.Y+frame.H > asked.Y {
			return fmt.Errorf("the first field reaches %d and the question starts at %d: they overlap, so this proves nothing", frame.Y+frame.H, asked.Y)
		}
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		if paintedNear(shot, frame, dimmed(Dark().Colour.Line.Edge), 6) == 0 {
			return fmt.Errorf("the form's first field is gone behind the question")
		}
		return nil
	})
	sc.Then(`^the question stands raised, its corners rounded$`, func() error {
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		asked := w.w.QuestionBox()
		if painted(shot, Box{X: asked.X, Y: asked.Y, W: 1, H: 1}, Dark().Colour.Surface.Raised) != 0 {
			return fmt.Errorf("the question's corner at %+v is not rounded", asked)
		}
		if painted(shot, Box{X: asked.X + asked.W/2, Y: asked.Y + 2, W: 1, H: 1}, Dark().Colour.Surface.Raised) != 1 {
			return fmt.Errorf("the question at %+v does not stand on a panel's ground", asked)
		}
		if painted(shot, Edges(asked)[0], Dark().Colour.Accent.Base) != 0 {
			return fmt.Errorf("the question is still edged in the accent")
		}
		return nil
	})
	sc.Then(`^what is behind the question is dimmed$`, func() error {
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		over := Over(w.w.width, w.w.height)
		fields := w.w.Form().Fields()
		frame := Framed(formValue(over, fields, 0), formRows(over, fields, 0))
		if painted(shot, frame, Dark().Colour.Line.Edge) != 0 || paintedNear(shot, frame, dimmed(Dark().Colour.Line.Edge), 6) == 0 {
			return fmt.Errorf("the form behind the question is drawn as bright as without it")
		}
		return nil
	})
	sc.Then(`^the field "([^"]*)" is drawn by name and not by number$`, w.fieldDrawnByName)
	sc.Then(`^the list offers "([^"]*)"$`, w.listOffers)
	sc.Then(`^the (list|calendar) has no heading$`, func(string) error {
		if said := Heads(w.w); said != "" {
			return fmt.Errorf("the panel is headed %q", said)
		}
		return nil
	})
	sc.Then(`^the list begins at the top of its panel$`, func() error {
		box := w.w.ChoosingBox()
		if run := ChoosingRun(box); run.Y != box.Y+padding {
			return fmt.Errorf("the list's rows begin at %d, its panel at %d", run.Y, box.Y)
		}
		return nil
	})
	sc.Then(`^the calendar's month, year and arrows stand whole in a row at its top$`, func() error {
		box := w.w.CalendarBox()
		row := []Box{CalendarMonth(box), CalendarYear(box), MonthBefore(box), MonthAfter(box)}
		for i, one := range row {
			if one.Y != box.Y+padding || Inside(one).H < heading*5/4 {
				return fmt.Errorf("%+v is not a whole line at the top of %+v", one, box)
			}
			if one.Y+one.H > WeekdayAt(box, 0).Y || one.X < box.X || one.X+one.W > box.X+box.W {
				return fmt.Errorf("%+v leaves the head of %+v", one, box)
			}
			if i > 0 && row[i-1].X+row[i-1].W > one.X {
				return fmt.Errorf("%+v runs over %+v", row[i-1], one)
			}
		}
		return nil
	})
	sc.Then(`^the list offers (a|no) search$`, func(which string) error {
		if got := w.w.ChoosingSearched(); got != (which == "a") {
			return fmt.Errorf("the list searched is %v, want %v", got, which == "a")
		}
		return nil
	})
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
		box := w.w.ChoosingBox()
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
	sc.Then(`^the list hangs under the field "([^"]*)"$`, func(name string) error {
		box := w.w.ChoosingBox()
		fields := w.w.Form().Fields()
		for i, field := range fields {
			if field.Name != name {
				continue
			}
			_, value, rows := FormPlace(w.w.formBox(), fields, i)
			frame := Framed(value, rows)
			if box.X > frame.X || box.X+box.W < frame.X+frame.W || box.Y != frame.Y+frame.H+between {
				return fmt.Errorf("the list is at %d,%d and the field %q ends at %d,%d", box.X, box.Y, name, frame.X, frame.Y+frame.H)
			}
			return nil
		}
		return fmt.Errorf("the form has no field %q", name)
	})
	sc.Then(`^the list has room for (\d+) rows and no more$`, func(rows int) error {
		if got := ChoosingRows(w.w.ChoosingBox()); got != rows {
			return fmt.Errorf("the list has room for %d rows", got)
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
		if said := w.w.looking.Reads(); said != "" {
			return fmt.Errorf("the search box already reads %q", said)
		}
		return nil
	})
	sc.Then(`^every key the bar offers does what it offers$`, w.everyKeyTheBarOffers)
	sc.Then(`^the trace is beside the fields$`, w.historyBesideTheFields)
	sc.Then(`^the form writes the trace, empty$`, func() error {
		if !w.w.Tracing() {
			return fmt.Errorf("the form writes no trace beside it")
		}
		if told := w.w.Told(); told != 0 {
			return fmt.Errorf("the trace of a task not yet saved holds %d lines", told)
		}
		return nil
	})
	sc.Then(`^the trace stays inside the panel$`, w.historyInsideThePanel)
	sc.Then(`^the trace has more to show than it has room for$`, w.historyHasMore)
	sc.Then(`^the trace has moved$`, w.historyMoved)
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
	sc.Then(`^the trace is at its beginning$`, w.historyAtTheBeginning)
	sc.Then(`^the trace is at its end$`, w.historyAtTheEnd)
	sc.Then(`^the field "([^"]*)" is written on more than one line$`, w.fieldOnMoreThanOneLine)
	sc.Then(`^no line of the field "([^"]*)" runs out of it$`, w.noLineRunsOut)
	sc.Then(`^the bar offers "([^"]*)" under "([^"]*)"$`, w.barOffersUnder)
	sc.Then(`^the window is asking what to look for$`, w.askingWhatToLookFor)
	sc.Then(`^the board says nothing of "([^"]*)"$`, func(title string) error {
		dealt, stalled := engine.Deal(w.w.Showing().Cards, w.w.archived)
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
	sc.Then(`^the rule over what the panel offers stands out from the panel$`, w.panelRuleSeen)
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
	for _, g := range engine.Sorted(w.w.Reading().Goals) {
		got = append(got, g.Title)
	}
	want := quoted(list)
	if strings.Join(got, "|") != strings.Join(want, "|") {
		return fmt.Errorf("the objectives read %v, want %v", got, want)
	}
	return nil
}

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

func (w *world) landsNowhere() error {
	if landing, somewhere := w.w.Landing(); somewhere {
		return fmt.Errorf("the board says the card would land at %v, and it cannot land there", landing)
	}
	return nil
}

func (w *world) landsIn(column string) error {
	at := engine.IndexOf(strings.ToLower(column))
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
	if got := w.w.looking.Reads(); got != value {
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
	if w.w.looking.Reads() == "" {
		return fmt.Errorf("nothing is being searched for any more")
	}
	return nil
}

func (w *world) barOffersUnder(offer, key string) error {
	for _, said := range w.w.Offers() {
		if said != offer {
			continue
		}
		if got := Pretty(w.w.shortcut(offer)); got != key {
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

func (w *world) panelRuleSeen() error {
	along := w.w.PanelOffers()
	if len(along) == 0 {
		return fmt.Errorf("the panel offers nothing, so it draws no rule")
	}
	shot, err := drawnInAStep(w.w)
	if err != nil {
		return err
	}
	box := w.w.panel()
	at := shot.RGBAAt(box.X+box.W/2, along[0].Y-between)
	rule := color.NRGBA{R: at.R, G: at.G, B: at.B, A: 255}
	if got := Contrast(rule, Dark().Colour.Surface.Raised); got < 1.2 {
		return fmt.Errorf("the rule %+v stands at %.2f from the panel, and a rule needs 1.2", rule, got)
	}
	return nil
}

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
	box := w.w.workFrame()
	last := (len(w.w.Under())+1)*SubHeight - box.H
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

func (w *world) formBehindTheList() error {
	over, list := Over(w.w.width, w.w.height), w.w.ChoosingBox()
	if list.W >= over.W || list.H >= over.H {
		return fmt.Errorf("the list is %dx%d over a form of %dx%d: it covers it rather than sitting on it",
			list.W, list.H, over.W, over.H)
	}
	fields := w.w.Form().Fields()
	value := formValue(over, fields, 0)
	frame := Framed(value, formRows(over, fields, 0))
	if frame.Y+frame.H > list.Y {
		return fmt.Errorf("the first field reaches %d and the list starts at %d: they overlap, so this proves nothing",
			frame.Y+frame.H, list.Y)
	}
	shot, err := drawnInAStep(w.w)
	if err != nil {
		return err
	}
	if painted(shot, frame, Dark().Colour.Surface.Field) == 0 {
		return fmt.Errorf("the form's first field is gone: the list covered it rather than sitting on it")
	}
	return nil
}

func (w *world) fieldDrawnByName(field string) error {
	for _, f := range w.w.Form().Fields() {
		if f.Name != field {
			continue
		}
		said := w.w.FieldSays(f)
		if said == f.Value {
			return fmt.Errorf("the field %q is drawn as %q, which is what it carries", field, said)
		}
		fields := w.w.Form().Fields()
		at := slices.IndexFunc(fields, func(x engine.Field) bool { return x.Name == field })
		box := w.w.formBox()
		value := formValue(box, fields, at)
		frame := Framed(value, formRows(box, fields, at))
		inside := Inside(frame)
		from := inside.X + (len(f.Value)+2)*perGlyph
		far := Box{X: from, Y: inside.Y, W: inside.X + inside.W - from, H: inside.H}
		shot, err := drawnInAStep(w.w)
		if err != nil {
			return err
		}
		if inked(shot, far, Dark().Colour.Surface.Back) == 0 {
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

func (w *world) everyKeyTheBarOffers() error {
	from := w.w
	for _, screen := range []Screen{OnTheBoard, OnForm, OnHelp, OnAsking} {
		at, err := onScreen(from, screen)
		if err != nil {
			return err
		}
		for _, offer := range at.Offers() {
			key := at.shortcut(offer)
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

func (w *world) whereGoal() (int, error) {
	for i, g := range engine.Sorted(w.w.Showing().Goals) {
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
	last := l.Column[engine.Columns-1]
	for i := range engine.Sorted(w.w.Showing().Goals) {
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
	last := len(engine.Sorted(w.w.Showing().Goals)) - 1
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

func (w *world) cardOf(title string) (engine.Card, error) {
	for _, c := range w.w.Showing().Cards {
		if c.Title == title {
			return c, nil
		}
	}
	return engine.Card{}, fmt.Errorf("no card titled %q", title)
}

func (w *world) cardSays(title, said string) error {
	c, err := w.cardOf(title)
	if err != nil {
		return err
	}
	facts, _ := Facts(c, moment, w.w.pages)
	m := w.w.measured()
	shown, more := Pills(c.Labels, PartsOf(w.w.Layout().CardAt(0, 0, 0)).Labels.W, func(said string) int { return ends(m, said, small) })
	line := strings.Join(shown, "  ") + "  +" + strconv.Itoa(more) + "  " + facts
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
	if _, mark := FocusMark(c, Box{W: 200, H: 2 * lineHigh}); !mark {
		return fmt.Errorf("the card for %q draws no glowing dot, and says nothing about being worked on", title)
	}
	return nil
}

func (w *world) cardCarriesAGlow(title string) error {
	c, err := w.cardOf(title)
	if err != nil {
		return err
	}
	line := Box{X: 0, Y: 0, W: 200, H: 2 * lineHigh}
	mark, ok := FocusMark(c, line)
	if !ok {
		return fmt.Errorf("the card for %q carries no mark at all", title)
	}
	glow := GlowOf(mark)
	if glow.Dot <= 0 || glow.Halo <= glow.Dot {
		return fmt.Errorf("the card for %q carries %+v, not a dot in a halo", title, glow)
	}
	if glow.X+glow.Halo > float32(line.W) || glow.X < float32(line.W-markSide) || glow.Y+glow.Halo > float32(lineHigh) {
		return fmt.Errorf("the glow of %q is at %+v, not at the right of its first line", title, glow)
	}
	return nil
}

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

func entrySays(path string, lines ...string) error {
	got, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for _, line := range lines {
		if !strings.Contains(string(got), line+"\n") {
			return fmt.Errorf("the entry does not say %q:\n%s", line, got)
		}
	}
	return nil
}

func (w *world) stepTile(title string) (Point, error) {
	parts := w.w.form.parts()
	tiles := w.w.partTiles()
	for i, part := range parts {
		shown := i - w.w.partsFrom()
		if part.Title != title || shown < 0 || shown >= len(tiles)-1 {
			continue
		}
		writ := SubWriting(tiles[shown])
		return Point{X: writ.X + writ.W/2, Y: writ.Y + writ.H/2}, nil
	}
	return Point{}, fmt.Errorf("no step titled %q is in sight", title)
}

func (w *world) stripedWide(title string, wide int) error {
	card, err := w.cardOf(title)
	if err != nil {
		return err
	}
	shot, err := drawnInAStep(w.w)
	if err != nil {
		return err
	}
	at := PartsOf(w.w.Layout().CardAt(engine.IndexOf(card.State), 0, 0)).Edge
	want := Dark().Colour.Priority[card.Priority]
	stripe := Box{X: at.X, Y: at.Y + Dark().Radius.Stripe, W: wide, H: at.H - 2*Dark().Radius.Stripe}
	if got := painted(shot, stripe, want); got != stripe.W*stripe.H {
		return fmt.Errorf("the stripe fills %d of %d pixels of its %d points", got, stripe.W*stripe.H, wide)
	}
	past := Box{X: at.X + wide, Y: at.Y + 1, W: 1, H: at.H - 2}
	if got := painted(shot, past, want); got != 0 {
		return fmt.Errorf("the stripe goes on past %d points, for %d pixels", wide, got)
	}
	return nil
}

func buttonsDrawn(w Window, buttons []Box, grounds ...color.NRGBA) error {
	shot, err := drawnInAStep(w)
	if err != nil {
		return err
	}
	for _, button := range buttons {
		row := Box{X: button.X + Dark().Radius.Control, Y: button.Y + 2, W: button.W - 2*Dark().Radius.Control, H: 1}
		got := 0
		for _, ground := range grounds {
			got += painted(shot, row, ground)
		}
		if got != row.W {
			return fmt.Errorf("the button at %+v stands on %d of %d pixels of its ground along %+v", button, got, row.W, row)
		}
	}
	return nil
}

func dayIndex(w Window, day time.Time) int {
	for i, one := range engine.MonthGrid(w.day) {
		if one.Equal(day) {
			return i
		}
	}
	return -1
}
