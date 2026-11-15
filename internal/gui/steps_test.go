package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/calendar"
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"gioui.org/io/key"
	"gioui.org/widget/material"
	"github.com/cucumber/godog"

	"cairn/internal/engine"
	"cairn/internal/engine/task"
	"cairn/internal/linuxapp"
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
	dealt, _ := board.Deal(w.w.Reading().Cards, w.w.Archived)
	for _, column := range dealt {
		got = append(got, column.Name)
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		return fmt.Errorf("the columns read %v, want %v", got, want)
	}
	return nil
}

func (w *world) cardsOf(column string) ([]board.Card, error) {
	at := board.IndexOf(strings.ToLower(column))
	if at < 0 {
		return nil, fmt.Errorf("no column called %q", column)
	}
	dealt, _ := board.Deal(w.w.Showing().Cards, w.w.Archived)
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
	got := Pages[w.w.Tab]
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
	at := board.IndexOf(strings.ToLower(column))
	l := w.w.Layout()
	for i := range here {
		if got := l.CardAt(at, i, 0).H; got != geom.CardHeight {
			return fmt.Errorf("card %d is %d tall, the rest are %d", i, got, geom.CardHeight)
		}
	}
	return nil
}

func (w *world) tooSmall() error {
	if w.w.Screen != board.TooSmallToDraw {
		return fmt.Errorf("the window is on %v", w.w.Screen)
	}
	return nil
}

func (w *world) saysWhatItNeeds() error {
	said := w.w.Says()
	if !strings.Contains(said, fmt.Sprint(geom.Narrowest)) || !strings.Contains(said, fmt.Sprint(geom.Shortest)) {
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
	if w.w.ChosenGoal {
		for _, g := range w.w.Reading().Goals {
			if g.Title == title {
				if w.w.Chosen != g.Number {
					return fmt.Errorf("objective %d is chosen, not %q", w.w.Chosen, title)
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
	if w.w.Chosen != number {
		return fmt.Errorf("task %d is chosen, not %q", w.w.Chosen, title)
	}
	return nil
}

func (w *world) detailOpenOn(title string) error {
	if w.w.Screen != board.OnForm {
		return fmt.Errorf("the window is on %v", w.w.Screen)
	}
	return w.isChosen(title)
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

func (w *world) wasAsked(what string, number int) error {
	thing, verb, _ := strings.Cut(what, " ")
	for _, a := range w.work.wrote {
		if a.thing == thing && a.verb == verb && a.number == fmt.Sprint(number) {
			return nil
		}
	}
	return fmt.Errorf("the store was never asked %q for task %d: it was asked %v", what, number, w.work.wrote)
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

func (w *world) fieldReads(name, value string) error {
	if got := valueOf(w.w.Form, name); got != value {
		return fmt.Errorf("the field %q reads %q, want %q", name, got, value)
	}
	return nil
}

func (w *world) saysWhatWentWrong() error {
	if w.w.Wrong() == nil {
		return fmt.Errorf("the window says nothing went wrong")
	}
	return nil
}

func (w *world) cardNumbered(n int) (board.Card, error) {
	for _, c := range w.w.Showing().Cards {
		if c.Number == n {
			return c, nil
		}
	}
	return board.Card{}, fmt.Errorf("the board carries no card numbered %d", n)
}

func (w *world) expects(sc *godog.ScenarioContext) {
	sc.Then(`^the columns read (".*")$`, w.columnsRead)
	sc.Then(`^"([^"]*)" carries "([^"]*)"$`, w.columnCarries)
	sc.Then(`^the card for task (\d+) says "([^"]*)"$`, func(n int, said string) error {
		c, err := w.cardNumbered(n)
		if err != nil {
			return err
		}
		line := Carried(c, Marks())
		facts, _ := CardFacts(c, time.Now(), engine.Pages{})
		if !strings.Contains(line, said) && !strings.Contains(facts, said) {
			return fmt.Errorf("the card for task %d reads %q beside its number and %q in its facts, and says nothing of %q",
				n, line, facts, said)
		}
		return nil
	})
	sc.Then(`^"([^"]*)" reads (".*")$`, w.columnReads)
	sc.Then(`^the tabs read (".*")$`, w.tabsRead)
	sc.Then(`^the tab is "([^"]*)"$`, w.tabIs)
	sc.Then(`^every card in "([^"]*)" is the same height$`, w.sameHeight)
	sc.Then(`^it says the window is too small$`, w.tooSmall)
	sc.Then(`^it says what size it needs$`, w.saysWhatItNeeds)
	sc.Then(`^the card "([^"]*)" is chosen$`, w.isChosen)
	sc.Then(`^"([^"]*)" is open to be edited$`, w.detailOpenOn)
	sc.Then(`^the steps are drawn as a list to tick, "([^"]*)" first$`, func(title string) error {
		if parts := w.w.PartsTicked(); len(parts) == 0 || parts[0].Title != title {
			return fmt.Errorf("the parts are drawn as %v, and %q is not first", parts, title)
		}
		return nil
	})
	sc.Then(`^the trace begins with the last thing done to task (\d+)$`, func(n int) error {
		shown, err := w.held.Ask(engine.TraceShow, board.Numbered(n))
		if err != nil {
			return err
		}
		last, lines := line(shown.Acts[len(shown.Acts)-1]), w.w.Lines()
		if len(lines) == 0 || !strings.HasPrefix(last, lines[0]) {
			return fmt.Errorf("the history begins %q, and the last thing done was %q", lines, last)
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
		for _, part := range w.w.Form.Parts() {
			got = append(got, part.Title)
		}
		if !slices.Equal(got, want) {
			return fmt.Errorf("the form's steps read %q, want %q", got, want)
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
	sc.Then(`^the lane "([^"]*)" is scrolled$`, w.laneIsScrolled)
	sc.Then(`^the lane "([^"]*)" is not scrolled$`, w.laneIsNotScrolled)
	sc.Then(`^the engine was asked "([^"]*)" for task (\d+)$`, w.wasAsked)
	sc.Step(`^I triple-click the first word of "([^"]*)" on the open task$`, func(key string) error {
		at, err := w.firstWordOf(key)
		if err == nil {
			w.w = w.w.tripleClicked(at)
		}
		return err
	})
	sc.Then(`^the engine was asked "([^"]*)" once$`, w.wasAskedOnce)
	sc.Then(`^task (\d+) is in the column "([^"]*)"$`, w.taskInColumn)
	sc.Then(`^the window asks to put task (\d+) in done, ticking its steps$`, func(n int) error {
		if w.w.Screen != board.OnAsking || w.w.Question != board.AboutFinishingParts {
			return fmt.Errorf("the window is on %v asking %v", w.w.Screen, w.w.Question)
		}
		if head := Heads(w.w); !strings.Contains(head, strconv.Itoa(n)) {
			return fmt.Errorf("the question reads %q, naming no task %d", head, n)
		}
		return nil
	})
	sc.Then(`^the window says the store comes from a newer cairn$`, func() error {
		if w.w.Screen != board.OnAsking || w.w.Question != board.AboutNewer {
			return fmt.Errorf("the window is on %v asking %v", w.w.Screen, w.w.Question)
		}
		return nil
	})
	sc.Then(`^task (\d+) is called "([^"]*)"$`, w.taskCalled)
	sc.Then(`^there are no tasks$`, func() error {
		if _, err := w.tell(1); err == nil {
			return fmt.Errorf("there is a task 1, want none")
		}
		return nil
	})
	sc.Then(`^the field "([^"]*)" reads "([^"]*)"$`, w.fieldReads)
	sc.Then(`^it says what went wrong$`, w.saysWhatWentWrong)
	sc.Then(`^the week shown starts on "([^"]*)"$`, func(day string) error {
		if got := calendar.FormatDay(w.w.Week); got != day {
			return fmt.Errorf("the week shown starts on %s, want %s", got, day)
		}
		return nil
	})
	weekday := func(day string) (int, error) {
		for i := range 7 {
			if calendar.FormatDay(w.w.Week.AddDate(0, 0, i)) == day {
				return i, nil
			}
		}
		return 0, fmt.Errorf("the week shown has no %s", day)
	}
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
		w.w = w.w.clicked(geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
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
				w.w = w.w.clicked(geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
				return w.w.Wrong()
			}
		}
		return fmt.Errorf("the time page offers %v, not %q", w.w.TimeCells(), name)
	})
	sc.Then(`^every cell of the time page is drawn as a button$`, func() error {
		var cells []geom.Box
		for i := range w.w.TimeCells() {
			cells = append(cells, TimeCellAt(w.w.Layout(), i))
		}
		return buttonsDrawn(w.theme, w.w, cells, design.Dark().Colour.Surface.Control, design.Dark().Colour.Accent.Base)
	})
	sc.Step(`^I rest the pointer on the (count of labels|title) of "([^"]*)"$`, func(what, title string) error {
		for _, placed := range w.w.Placed() {
			if placed.Card.Title != title {
				continue
			}
			parts := PartsOf(placed.Slot)
			at := geom.Point{X: parts.Title.X + 4, Y: parts.Title.Y + 4}
			if what == "count of labels" {
				more := CardRowOf(placed.Card, parts.Number, w.w.measured(), moment, engine.Pages{}).More
				if more == (geom.Box{}) {
					return fmt.Errorf("the card for %q counts no labels", title)
				}
				at = geom.Point{X: more.X + more.W/2, Y: more.Y + more.H/2}
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
		if !w.w.Daily || calendar.FormatDay(w.w.Week) != day {
			return fmt.Errorf("the time page shows %q, daily %v", TimeSays(w.w), w.w.Daily)
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
		w.w = w.w.clicked(geom.Point{X: head.X + 4, Y: head.Y + head.H/2})
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
		if !w.w.Monthly || w.w.Week.Format("January 2006") != month {
			return fmt.Errorf("the time page shows %q, monthly %v", TimeSays(w.w), w.w.Monthly)
		}
		return nil
	})
	dayOfMonth := func(day string) (int, error) {
		for i, shown := range board.MonthGrid(w.w.Week) {
			if calendar.FormatDay(shown) == day {
				return i, nil
			}
		}
		return 0, fmt.Errorf("the month shown has no %s", day)
	}
	cardOfWeek := func(title string) (geom.Point, error) {
		l := w.w.Layout()
		for i := range 7 {
			for j, one := range w.w.DayShown(i) {
				if w.w.DayCard(one).Title == title {
					box := DayCardAt(WeekDayAt(l, i), j, len(w.w.DayHeld(i).Tasks))
					return geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2}, nil
				}
			}
		}
		return geom.Point{}, fmt.Errorf("no card titled %q is on the time page", title)
	}
	sc.Step(`^I click the card "([^"]*)" on the time page$`, func(title string) error {
		at, err := cardOfWeek(title)
		if err != nil {
			return err
		}
		w.w = w.w.clicked(at)
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
				w.w = w.w.doubleClicked(geom.Point{X: box.X + box.W/2, Y: box.Y + box.H/2})
				return nil
			}
		}
		return fmt.Errorf("%s holds no task titled %q", day, title)
	})
	sc.Then(`^the labels list "([^"]*)" carried by (\d+) tasks? and (\d+) objectives?$`, func(label string, tasks, objectives int) error {
		if w.w.Screen != board.OnLabels {
			return fmt.Errorf("the labels are not listed: the window is on %v", w.w.Screen)
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
		if w.w.Screen != board.OnLabels {
			return fmt.Errorf("the window is on %v", w.w.Screen)
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
	clicking := func(b geom.Box) { w.w = w.w.clicked(geom.Point{X: b.X + b.W/2, Y: b.Y + b.H/2}) }
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
		if w.w.Screen != board.OnChoosing || !w.w.Picking() {
			return fmt.Errorf("the window is on %v, picking %v", w.w.Screen, w.w.Picking())
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
		if w.w.Screen != board.OnAsking || Heads(w.w) != question {
			return fmt.Errorf("the window is on %v and asks %q", w.w.Screen, Heads(w.w))
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
	sc.Then(`^the desktop entry "([^"]*)" names the icon "([^"]*)"$`, func(name, icon string) error {
		return entrySays(entry(name), "Icon="+icon, "StartupWMClass="+Known)
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
		if w.w.Screen != board.OnBackups {
			return fmt.Errorf("the window is on %v, not on the backups", w.w.Screen)
		}
		if got := len(w.w.Backups); got != n {
			return fmt.Errorf("%d backups are listed, want %d", got, n)
		}
		return nil
	})
	sc.Then(`^the window is asking$`, func() error {
		if w.w.Screen != board.OnAsking {
			return fmt.Errorf("the window is on %v, not asking", w.w.Screen)
		}
		return nil
	})
	sc.Then(`^it says nothing went wrong$`, func() error {
		if wrong := w.w.Wrong(); wrong != nil {
			return fmt.Errorf("the window still says: %w", wrong)
		}
		return nil
	})
	sc.Step(`^the window has looked at the store (\d+) times?$`, func(n int) error {
		for range n {
			w.w = w.w.Wonder()
		}
		return nil
	})

	sc.Then(`^"([^"]*)" does not carry "([^"]*)"$`, w.columnDoesNotCarry)
	sc.When(`^I click the filter "([^"]*)"$`, func(name string) error {
		for i, f := range w.w.Filters {
			if f.Name == name {
				row := w.w.FilterRowAt(i)
				w.w = w.w.clicked(geom.Point{X: row.X + row.W/2, Y: row.Y + row.H/2})
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
		if !open || w.w.Filters[w.w.FilterOn].Name != name || len(values) != len(w.w.Filters[w.w.FilterOn].Shown) {
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
				w.w = w.w.clicked(geom.Point{X: swatch.At.X + swatch.At.W/2, Y: swatch.At.Y + swatch.At.H/2})
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
	sc.When(`^I click the work "([^"]*)"$`, func(title string) error {
		for _, tile := range w.w.WorkTiles() {
			if tile.Said == title {
				w.w = w.w.clicked(geom.Point{X: tile.At.X + tile.At.W/2, Y: tile.At.Y + tile.At.H/2})
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
	sc.Then(`^the time page shows a (week|month)$`, func(what string) error {
		if w.w.Tab != board.Time || w.w.Monthly != (what == "month") || w.w.Daily {
			return fmt.Errorf("the time page on %v shows monthly %v daily %v", w.w.Tab, w.w.Monthly, w.w.Daily)
		}
		return nil
	})
	sc.Then(`^the search "([^"]*)" is kept as a pill$`, func(said string) error {
		if _, ok := w.w.SearchPill(); !ok || w.w.Screen != board.OnTheBoard || w.w.Searched().Reads() != said {
			return fmt.Errorf("the window on %v keeps the search %q, pill %v", w.w.Screen, w.w.Searched().Reads(), ok)
		}
		return nil
	})
	sc.Step(`^I click the cross of the search's pill$`, func() error {
		pill, ok := w.w.SearchPill()
		if !ok {
			return fmt.Errorf("no search is kept as a pill")
		}
		w.w = w.w.clicked(geom.Point{X: pill.X + pill.W - pillCross/2, Y: pill.Y + pill.H/2})
		return nil
	})
	sc.Then(`^no search is kept$`, func() error {
		if _, ok := w.w.SearchPill(); ok || !w.w.Searched().Empty() {
			return fmt.Errorf("the search %q is kept", w.w.Searched().Reads())
		}
		return nil
	})
	sc.Then(`^the command bar is (open|shut)$`, func(how string) error {
		if open := w.w.Screen == board.OnPalette; open != (how == "open") {
			return fmt.Errorf("the command bar open is %v, the window on %v", open, w.w.Screen)
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
	sc.Then(`^the bar does not offer "([^"]*)"$`, func(offer string) error {
		for _, on := range w.w.Offers() {
			if on == offer {
				return fmt.Errorf("the bar offers %q: %v", offer, w.w.Offers())
			}
		}
		return nil
	})
	sc.Then(`^the card "([^"]*)" says "([^"]*)"$`, w.cardSays)
	sc.Then(`^what is chosen is in sight$`, w.chosenInSight)
	sc.Then(`^the tab "([^"]*)" says it holds (\d+)$`, w.tabHolds)
	sc.Then(`^the form is open on task (\d+)$`, w.formOpenOnTask)

	sc.Then(`^the list offers "([^"]*)"$`, w.listOffers)
	sc.Then(`^the list offers to create "([^"]*)"$`, func(label string) error {
		for _, one := range w.w.Choosing() {
			if one.Creates == label {
				return nil
			}
		}
		return fmt.Errorf("the list offers %v, and does not offer to create %q", goalTitles(w.w.Choosing()), label)
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
	sc.Then(`^the list does not offer "([^"]*)"$`, func(title string) error {
		for _, one := range w.w.Choosing() {
			if one.Title == title {
				return fmt.Errorf("the list still offers %q", title)
			}
		}
		return nil
	})
	sc.Then(`^the window is choosing$`, func() error {
		if w.w.Screen != board.OnChoosing {
			return fmt.Errorf("the window is on %v, not the list", w.w.Screen)
		}
		return nil
	})
	sc.Then(`^the form is still on the field "([^"]*)"$`, func(name string) error {
		if got := w.w.Form.Fields[w.w.Form.On()].Name; got != name {
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
	sc.Then(`^the window is on the board$`, func() error {
		if w.w.Screen != board.OnTheBoard {
			return fmt.Errorf("the window is on %v, not the board", w.w.Screen)
		}
		return nil
	})
	sc.Then(`^the window is on the form$`, func() error {
		if w.w.Screen != board.OnForm {
			return fmt.Errorf("the window is on %v, not the form", w.w.Screen)
		}
		return nil
	})
	sc.Then(`^objective (\d+) is ticked$`, func(n int) error { return w.ticked(n, true) })
	sc.Then(`^objective (\d+) is not ticked$`, func(n int) error { return w.ticked(n, false) })
	sc.Then(`^the field "([^"]*)" names "([^"]*)"$`, w.fieldNames)
	sc.Then(`^every key the bar offers does what it offers$`, w.everyKeyTheBarOffers)
	sc.Then(`^the field "([^"]*)" is written on more than one line$`, w.fieldOnMoreThanOneLine)
	sc.Then(`^no line of the field "([^"]*)" runs out of it$`, w.noLineRunsOut)
	sc.Then(`^the bar offers "([^"]*)" under "([^"]*)"$`, w.barOffersUnder)
	sc.Then(`^the window is asking what to look for$`, w.askingWhatToLookFor)
	sc.Then(`^the board says nothing of "([^"]*)"$`, func(title string) error {
		dealt, stalled := board.Deal(w.w.Showing().Cards, w.w.Archived)
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
		if w.w.Chosen != 0 {
			return fmt.Errorf("task %d is chosen", w.w.Chosen)
		}
		return nil
	})
	sc.Then(`^the panel is headed "([^"]*)"$`, w.helpHeaded)
	sc.Then(`^the search box is open$`, w.searchBoxOpen)
	sc.Then(`^the caret is at (\d+)$`, w.caretIsAt)
	sc.Then(`^the search reads "([^"]*)"$`, w.searchReads)
	sc.Then(`^the objectives read (".*")$`, w.objectivesRead)
	sc.Then(`^the objective "([^"]*)" is chosen$`, w.objectiveChosen)
}

func (w *world) objectivesRead(list string) error {
	var got []string
	for _, b := range w.w.Tree() {
		got = append(got, b.Goal.Title)
	}
	want := quoted(list)
	if strings.Join(got, "|") != strings.Join(want, "|") {
		return fmt.Errorf("the objectives read %v, want %v", got, want)
	}
	return nil
}

func (w *world) objectiveChosen(title string) error {
	for _, g := range w.w.Reading().Goals {
		if g.Title == title {
			if w.w.Chosen != g.Number {
				return fmt.Errorf("objective %d is chosen, not %q", w.w.Chosen, title)
			}
			return nil
		}
	}
	return fmt.Errorf("no objective titled %q", title)
}

func (w *world) columnDoesNotCarry(column, title string) error {
	if err := w.columnCarries(column, title); err == nil {
		return fmt.Errorf("%q carries %q and should not", column, title)
	}
	return nil
}

func (w *world) helpOpen() error {
	if w.w.Screen != board.OnHelp {
		return fmt.Errorf("the window is on %v", w.w.Screen)
	}
	return nil
}

func (w *world) caretIsAt(at int) error {
	if w.w.Screen == board.OnSearch {
		if got := w.w.Searched().At; got != at {
			return fmt.Errorf("the caret is at %d in the search, want %d", got, at)
		}
		return nil
	}
	if w.w.Screen != board.OnForm {
		return fmt.Errorf("nothing is being typed into")
	}
	if got := w.w.Form.Writing().At; got != at {
		return fmt.Errorf("the caret is at %d, want %d", got, at)
	}
	return nil
}

func (w *world) searchReads(value string) error {
	if got := w.w.Looking.Reads(); got != value {
		return fmt.Errorf("the search reads %q, want %q", got, value)
	}
	return nil
}

func (w *world) searchBoxOpen() error {
	if w.w.Screen != board.OnSearch {
		return fmt.Errorf("the window is on %v", w.w.Screen)
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
	if w.w.Screen != board.OnSearch {
		return fmt.Errorf("the window is on %v", w.w.Screen)
	}
	return nil
}

func (w *world) helpHeaded(head string) error {
	if got := Heads(w.w); got != head {
		return fmt.Errorf("the help is headed %q, not %q", got, head)
	}
	return nil
}

func (w *world) wrapping(name string) (Lines, int, error) {
	fields := w.w.Form.Fields
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

func (w *world) listOffers(title string) error {
	for _, g := range w.w.Choosing() {
		if g.Title == title {
			return nil
		}
	}
	return fmt.Errorf("the list offers %v, and not %q", goalTitles(w.w.Choosing()), title)
}

func (w *world) ticked(number int, want bool) error {
	if got := w.w.Ticked(number); got != want {
		return fmt.Errorf("objective %d ticked = %v, want %v", number, got, want)
	}
	return nil
}

func (w *world) fieldNames(field, title string) error {
	for _, f := range w.w.Form.Fields {
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

func goalTitles(offered []Choice) []string {
	var said []string
	for _, one := range offered {
		said = append(said, one.Title)
	}
	return said
}

func (w *world) formOpenOnTask(number int) error {
	if w.w.Screen != board.OnForm {
		return fmt.Errorf("the window is on %v, not a form", w.w.Screen)
	}
	if got := w.w.Form.Number; got != number {
		return fmt.Errorf("the form is open on %d, not %d", got, number)
	}
	return nil
}

func (w *world) everyKeyTheBarOffers() error {
	from := w.w
	for _, screen := range []board.Screen{board.OnTheBoard, board.OnForm, board.OnHelp, board.OnAsking} {
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
			if pressed.Screen != chosen.Screen {
				return fmt.Errorf("on %v, %q left the window on %v and pressing %q left it on %v",
					screen, offer, chosen.Screen, key, pressed.Screen)
			}
			if pressed.Screen == at.Screen && chosen.Screen == at.Screen && pressed.Chosen != chosen.Chosen {
				return fmt.Errorf("on %v, %q chose %d and pressing %q chose %d",
					screen, offer, chosen.Chosen, key, pressed.Chosen)
			}
		}
	}
	return nil
}

func onScreen(w Window, screen board.Screen) (Window, error) {
	switch screen {
	case board.OnTheBoard:
		return w, nil
	case board.OnForm:
		return w.Key("enter"), nil
	case board.OnHelp:
		return w.Choose("Help"), nil
	case board.OnAsking:
		return w.Choose("Delete"), nil
	}
	return w, fmt.Errorf("no way to reach %v", screen)
}

func (w *world) tabHolds(name string, count int) error {
	for i, page := range Pages {
		if page != name {
			continue
		}
		if got := w.w.Holds(board.Tab(i)); got != count {
			return fmt.Errorf("the tab %q says it holds %d, not %d", name, got, count)
		}
		return nil
	}
	return fmt.Errorf("there is no tab called %q", name)
}

func (w *world) chosenInSight() error {
	if w.w.Chosen == 0 {
		return fmt.Errorf("nothing is chosen, so this proves nothing")
	}
	box, run, ok := w.w.WhereChosen()
	if !ok {
		return fmt.Errorf("task %d is chosen and drawn nowhere", w.w.Chosen)
	}
	if box.Y < run.Y || box.Y+box.H > run.Y+run.H {
		return fmt.Errorf("what is chosen is drawn from %d to %d, in a run from %d to %d",
			box.Y, box.Y+box.H, run.Y, run.Y+run.H)
	}
	return nil
}

func (w *world) cardOf(title string) (board.Card, error) {
	for _, c := range w.w.Showing().Cards {
		if c.Title == title {
			return c, nil
		}
	}
	return board.Card{}, fmt.Errorf("no card titled %q", title)
}

func (w *world) cardSays(title, said string) error {
	c, err := w.cardOf(title)
	if err != nil {
		return err
	}
	facts, _ := CardFacts(c, moment, w.w.Pages)
	m := w.w.measured()
	shown, more := Pills(c.Labels, PartsOf(w.w.Layout().CardAt(0, 0, 0)).Labels.W, func(said string) int { return ends(m, said, small) })
	line := strings.Join(shown, "  ") + "  +" + strconv.Itoa(more) + "  " + facts
	if !strings.Contains(line, said) {
		return fmt.Errorf("the card for %q reads %q, and says nothing of %q", title, line, said)
	}
	return nil
}

func (w *world) firstWordOf(key string) (geom.Point, error) {
	for _, r := range w.w.Readables() {
		if r.Key == key {
			return geom.Point{X: r.Box.X + glyphOf(r.Size)/2, Y: r.Box.Y + lineHigh/2}, nil
		}
	}
	if name, ok := strings.CutPrefix(key, "field:"); ok {
		if at, err := w.inField(name); err == nil {
			return at(0), nil
		}
	}
	return geom.Point{}, fmt.Errorf("the open task writes nothing called %q", key)
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

func (w *world) stepTile(title string) (geom.Point, error) {
	parts := w.w.Form.Parts()
	tiles := w.w.partTiles()
	for i, part := range parts {
		shown := i - w.w.partsFrom()
		if part.Title != title || shown < 0 || shown >= len(tiles)-1 {
			continue
		}
		writ := SubWriting(tiles[shown])
		return geom.Point{X: writ.X + writ.W/2, Y: writ.Y + writ.H/2}, nil
	}
	return geom.Point{}, fmt.Errorf("no step titled %q is in sight", title)
}

func buttonsDrawn(th *material.Theme, w Window, buttons []geom.Box, grounds ...color.NRGBA) error {
	shot, err := drawnInAStep(th, w)
	if err != nil {
		return err
	}
	for _, button := range buttons {
		row := geom.Box{X: button.X + design.Dark().Radius.Control, Y: button.Y + 2, W: button.W - 2*design.Dark().Radius.Control, H: 1}
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
