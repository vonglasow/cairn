package gui

import (
	"slices"
	"strconv"
	"testing"
	"time"

	"cairn/internal/engine"
)

func homeWindow(width, height int) Window {
	w := holding(newFake(),
		engine.Card{Number: 1, Title: "Pin map", State: "backlog", Objectives: []int{2, 3}, Serves: []string{"Enclosure", "Firmware"}},
		engine.Card{Number: 2, Title: "OTA updates", State: "done", Objectives: []int{3}, Serves: []string{"Firmware"}},
		engine.Card{Number: 3, Title: "Buy the relay", State: "current", Objectives: []int{1}, Serves: []string{"Home automation"}},
	).Resized(width, height)
	w.snap.Goals = []engine.Goal{
		{Number: 1, Title: "Home automation", State: "active", Children: []int{2, 3}, Done: 1, Total: 3},
		{Number: 2, Title: "Enclosure", State: "active", Parent: 1, Done: 0, Total: 1},
		{Number: 3, Title: "Firmware", State: "active", Parent: 1, Done: 1, Total: 2, WaitedOnBy: []int{4}},
		{Number: 4, Title: "Tidy the workshop", State: "active", WaitsOn: []int{3}},
	}
	w.tab = Objectives
	return w
}

func TestTheSplitPutsTheListBesideTheDetailAndBothInsideTheBoard(t *testing.T) {
	for _, size := range pinnedSizes {
		l := LayoutFor(size.X, size.Y)
		list, detail := SplitOf(l)
		if list.X != l.Column[0].X || list.W > splitTo || detail.X != list.X+list.W+gap {
			t.Errorf("at %v the list %+v and the detail %+v do not stand side by side", size, list, detail)
		}
		last := l.Column[len(l.Column)-1]
		if detail.X+detail.W != last.X+last.W || list.Y != l.Head[0].Y || list.H != detail.H {
			t.Errorf("at %v the split %+v %+v does not fill the board", size, list, detail)
		}
		if run := GoalsRun(l); run.Y <= list.Y || run.Y+run.H > list.Y+list.H || run.X < list.X || run.X+run.W > list.X+list.W {
			t.Errorf("at %v the run %+v is not inside the list %+v under its head", size, run, list)
		}
	}
}

func TestEachObjectiveHasItsSlotDownTheList(t *testing.T) {
	l := LayoutFor(1200, 800)
	first, second := GoalAt(l, 0, 0), GoalAt(l, 1, 0)
	if second.Y != first.Y+ItemHigh || second.X != first.X || first.H != ItemHigh {
		t.Errorf("the slots %+v and %+v are not one under the other", first, second)
	}
	if scrolled := GoalAt(l, 1, 10); scrolled.Y != second.Y-10 {
		t.Errorf("a list scrolled by 10 puts the second at %d", scrolled.Y)
	}
}

func TestTheTreeKeepsEachSubObjectiveUnderItsParent(t *testing.T) {
	var got []int
	for _, b := range homeWindow(1200, 800).unfolding(1).Tree() {
		got = append(got, b.Goal.Number)
	}
	if !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Errorf("the tree reads %v", got)
	}
}

func TestTheDetailShowsTheObjectiveChosenAndTheFirstWhileNoneIs(t *testing.T) {
	w := homeWindow(1200, 800)
	if g, ok := w.Detailed(); !ok || g.Number != 1 {
		t.Errorf("with none chosen the detail shows %d, want the first", g.Number)
	}
	w.chosen, w.chosenGoal = 4, true
	if g, _ := w.Detailed(); g.Number != 4 {
		t.Errorf("the detail shows %d, want the one chosen", g.Number)
	}
	empty := holding(newFake())
	if _, ok := empty.Detailed(); ok {
		t.Error("a store with no objective details one")
	}
}

func TestTheDetailOfAParentSplitsItsTasksAndTilesItsSubObjectives(t *testing.T) {
	for _, size := range pinnedSizes {
		d := homeWindow(size.X, size.Y).Detail()
		if d.Goal.Number != 1 || len(d.Tiles) != 2 {
			t.Fatalf("at %v the detail of %d tiles %d sub-objectives", size, d.Goal.Number, len(d.Tiles))
		}
		if len(d.Own) != 1 || d.Own[0].Card.Number != 3 {
			t.Errorf("at %v the tasks on it itself are %+v", size, d.Own)
		}
		var under []int
		for _, line := range d.InSubs {
			under = append(under, line.Card.Number)
		}
		if want := []int{1, 2}[:len(under)]; len(under) == 0 || !slices.Equal(under, want) {
			t.Errorf("at %v the tasks in its sub-objectives read %v, want open first", size, under)
		}
		if more := 2 - len(under); more > 0 && d.More != strconv.Itoa(more)+" more · t shows them all on the board" {
			t.Errorf("at %v %d tasks have no room and the detail says %q", size, more, d.More)
		}
		_, detail := SplitOf(homeWindow(size.X, size.Y).Layout())
		for _, b := range append([]Box{d.Heading, d.Line, d.Progress, d.Figures, d.SubsHead, d.OwnHead, d.InSubsHead}, d.Buttons...) {
			if b.X < detail.X || b.X+b.W > detail.X+detail.W || b.Y < detail.Y || b.Y+b.H > detail.Y+detail.H {
				t.Errorf("at %v a part %+v stands outside the detail %+v", size, b, detail)
			}
		}
		if d.Heading.X+d.Heading.W > d.Buttons[0].X {
			t.Errorf("at %v the heading %+v runs under the buttons %+v", size, d.Heading, d.Buttons[0])
		}
	}
}

func TestTheDetailNamesWhatTheObjectiveWaitsOnAndWhatWaitsOnIt(t *testing.T) {
	w := homeWindow(1200, 800)
	w.chosen, w.chosenGoal = 3, true
	if got := w.Detail().Waits.Lines; len(got) != 1 || got[0].Value != "#4 Tidy the workshop" {
		t.Errorf("Firmware's detail says %+v", got)
	}
	w.chosen = 4
	if got := w.Detail().Waits.Lines; len(got) != 1 || got[0].Value != "#3 Firmware" {
		t.Errorf("Tidy's detail says %+v", got)
	}
}

func TestATileSaysWhatElseItsTasksServeAndWarnsOfWhatItWaitsOn(t *testing.T) {
	w := homeWindow(1200, 800)
	w.snap.Goals[2].WaitsOn = []int{2}
	tiles := w.Detail().Tiles
	if got := tiles[0].Tile.Notes; !slices.Equal(got, []string{"#1 also in Firmware"}) {
		t.Errorf("Enclosure's tile notes %v", got)
	}
	if got := tiles[1].Tile.Warn; got != "waits on Enclosure" {
		t.Errorf("Firmware's tile warns %q", got)
	}
}

func TestTShowsTheTasksOfTheObjectiveDetailed(t *testing.T) {
	w := homeWindow(1200, 800).At(time.Now()).Key("t")
	if w.Tab() != Tasks {
		t.Fatalf("t left the window on %v", w.Tab())
	}
	if said := w.FilteredBy(); said != "objective #1 Home automation, with its sub-objectives" {
		t.Errorf("the board is filtered by %q", said)
	}
}

func TestWorkStartedEarlyIsToldAboveTheBar(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "OTA updates", State: "todo"})
	f.warn = []string{"#2 Features waits on #1 Firmware, not reached"}
	w.chosen = 1
	if got := w.working().Note(); got != "#2 Features waits on #1 Firmware, not reached" {
		t.Errorf("focusing says %q", got)
	}
	w.form = Form{Number: 1}
	w.screen = OnForm
	if got, _, _ := w.stored(); got.Note() != "#2 Features waits on #1 Firmware, not reached" {
		t.Errorf("saving says %q", got.Note())
	}
}

func onField(w Window, arg string) Window {
	for i, f := range w.form.Fields() {
		if f.Arg == arg {
			w.form = w.form.onto(i)
		}
	}
	return w.choosingList()
}

func choiceTitles(w Window) []string {
	var out []string
	for _, c := range w.Choosing() {
		out = append(out, c.Title)
	}
	return out
}

func TestTheParentOfAnObjectiveIsChosenAmongThoseOfLevelOneOrNone(t *testing.T) {
	w := homeWindow(1200, 800)
	w.chosen, w.chosenGoal = 4, true
	w = onField(w.Edit(), "parent")
	if got := choiceTitles(w); !slices.Equal(got, []string{"none", "Home automation"}) {
		t.Fatalf("the parent list reads %v", got)
	}
	if !w.OneOnly() {
		t.Error("the parent list takes more than one")
	}
	w.choosing = 1
	w = w.taking()
	if got := w.form.Fields()[w.form.On()].Value; got != "1" {
		t.Errorf("choosing Home automation writes %q", got)
	}
	w = onField(w, "parent")
	w.choosing = 0
	if got := w.taking().form.Fields()[w.form.On()].Value; got != "none" {
		t.Errorf("choosing none writes %q", got)
	}
}

func TestTheLabelsOfAFormAreThoseMeantForWhatItIsAbout(t *testing.T) {
	w := homeWindow(1200, 800)
	w.snap.Goals[3].Labels = []string{"area"}
	w.snap.Cards[0].Labels = []string{"backend"}
	w.snap.Scopes = map[string]string{"area": "objectives", "backend": "tasks"}
	w.chosen, w.chosenGoal = 1, true
	if got := choiceTitles(onField(w.Edit(), "label")); !slices.Equal(got, []string{"area"}) {
		t.Errorf("an objective's labels offer %v", got)
	}
	w.tab, w.chosen, w.chosenGoal = Tasks, 1, false
	if got := choiceTitles(onField(w.Edit(), "label")); !slices.Equal(got, []string{"backend"}) {
		t.Errorf("a task's labels offer %v", got)
	}
}

func TestTheObjectivesOfATaskAreATreeAndCountingOneTwiceIsRefused(t *testing.T) {
	w := homeWindow(1200, 800)
	w.tab, w.chosen = Tasks, 1
	w = onField(w.Edit(), "objective")
	var depths []int
	for _, c := range w.Choosing() {
		depths = append(depths, c.Depth)
	}
	if !slices.Equal(depths, []int{0, 1, 1, 0}) {
		t.Fatalf("the list's depths are %v", depths)
	}
	if got := w.RowAside(w.Choosing()[0]); got != "counted through 2 below" {
		t.Errorf("the parent's row says %q", got)
	}
	w.choosing = 0
	w = w.taking()
	if w.Ticked(1) {
		t.Error("the parent was ticked beside its sub-objectives")
	}
	if got := w.ChoosingRefused(); got != "#1 is already counted in #1 Home automation through #2 Enclosure and #3 Firmware" {
		t.Errorf("the list says %q", got)
	}
	w.choosing = 3
	if got := w.taking().ChoosingRefused(); got != "" {
		t.Errorf("ticking another tree still says %q", got)
	}
}

func TestTheProgressOfASubObjectiveSaysWhereItIsCounted(t *testing.T) {
	w := homeWindow(1200, 800)
	w.snap.Goals[2].Estimate = "12h"
	got := w.GoalProgress(w.snap.Goals[2]).Lines
	want := []Fact{
		{Name: "done", Value: "1 of 2 tasks"},
		{Name: "time spent", Value: "—", Aside: "of 12h estimated"},
		{Name: "left", Value: "1 task, none due"},
		{Name: "counted in", Value: "#1 Home automation", Aside: "1 of 3 there, each task once"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("the progress says %+v", got)
	}
}

func TestTheFixturesHoldEveryObjectiveScreen(t *testing.T) {
	parent, sub, refused := objectiveScreens(aFixturesBoard(t))
	if len(parent.Detail().Tiles) == 0 || sub.chosen == 0 {
		t.Errorf("the fixtures hold no objective with sub-objectives to draw")
	}
	if refused.ChoosingRefused() == "" {
		t.Errorf("the fixtures hold no task to refuse a double count on")
	}
	if len(parent.Key("t").Narrowing().At) == 0 {
		t.Errorf("showing the parent's tasks draws no narrowing bar")
	}
	if parent.Delete().Screen() != OnFates {
		t.Errorf("deleting the parent asks nothing")
	}
}

func TestDeletingAnObjectiveAsksRowByRowAndNamesWhatGoes(t *testing.T) {
	w := homeWindow(1200, 800)
	w.chosen, w.chosenGoal = 1, true
	w = w.Delete()
	if w.Screen() != OnFates {
		t.Fatalf("deleting a parent left the window on %v", w.Screen())
	}
	f := w.FatesAsked()
	if len(f.Rows) != 3 || f.Rows[2].Sub != 0 || f.Rows[0].On != -1 {
		t.Fatalf("the question asks %+v", f.Rows)
	}
	for _, row := range f.Rows {
		if !f.Box.In(Point{X: row.Choice.X, Y: row.Choice.Y}) || row.Choice.X < row.At.X {
			t.Errorf("the row %q chooses at %+v, outside %+v", row.Said, row.Choice, f.Box)
		}
	}
	if w.Key("enter").Screen() != OnFates {
		t.Error("enter deleted before everything was chosen")
	}
	w = w.Key("right").Key("down").Key("left").Key("down").Key("left")
	if got := w.DeletesSays(); got != "Delete #1 and #3 with 3 tasks" {
		t.Errorf("the button says %q", got)
	}
	w = w.Key("up").Key("up").Key("k").Key("h")
	if got := w.DeletesSays(); got != "Delete #1, #2 and #3 with 3 tasks" {
		t.Errorf("the button says %q", got)
	}
	f2 := newFake()
	gone := homeWindow(1200, 800)
	gone.work = f2
	gone.chosen, gone.chosenGoal = 1, true
	gone = gone.Delete().Key("l").Key("j").Key("l").Key("j").Key("l").Key("enter")
	in, asked := f2.gave("objective", "delete")
	if !asked || in.Given["tasks"] != "detach" || gone.Screen() != OnTheBoard {
		t.Errorf("confirming asked %+v %v, on %v", in, asked, gone.Screen())
	}
	if cancelled := w.Choose("Cancel"); cancelled.Screen() != OnTheBoard {
		t.Errorf("cancel left the window on %v", cancelled.Screen())
	}
	alone := homeWindow(1200, 800)
	alone.chosen, alone.chosenGoal = 4, true
	if alone.Delete().Screen() != OnAsking {
		t.Error("an objective holding nothing is not asked about plainly")
	}
}

func TestThePillOfAnObjectiveShownWithItsSubObjectivesSaysSo(t *testing.T) {
	w := homeWindow(1200, 800).At(time.Now()).Key("t")
	pills := w.Pills()
	if len(pills) != 1 || pills[0].Said != "#1 Home automation with its sub-objectives" || pills[0].Tail != WithItsSubsSays {
		t.Errorf("the pills are %+v", pills)
	}
}

func TestANarrowWindowShowsThePartInForceBetweenTwoChevrons(t *testing.T) {
	w := homeWindow(Narrowest, Shortest).At(time.Now()).Key("t")
	n := w.Narrowing()
	if !n.Compact || n.Back.W == 0 || n.Next.W == 0 || n.At[n.On].W == 0 {
		t.Fatalf("at %d the bar is %+v", Narrowest, n)
	}
	if n.Back.X+n.Back.W > n.At[n.On].X || n.Next.X < n.At[n.On].X+n.At[n.On].W || n.Box.X+n.Box.W > w.width-margin {
		t.Errorf("the chevrons %+v %+v do not stand either side of %+v inside %+v", n.Back, n.Next, n.At[n.On], n.Box)
	}
	next := w.clicked(Point{X: n.Next.X + n.Next.W/2, Y: n.Next.Y + n.Next.H/2})
	if taken, _ := next.objectiveTaken(); taken != "2" {
		t.Errorf("the next chevron narrowed to %q", taken)
	}
	n = next.Narrowing()
	back := next.clicked(Point{X: n.Back.X + n.Back.W/2, Y: n.Back.Y + n.Back.H/2})
	if taken, _ := back.objectiveTaken(); taken != "1" {
		t.Errorf("the chevron back narrowed to %q", taken)
	}
	for _, p := range next.Pills() {
		if p.Box.X+p.Box.W > n.Box.X {
			t.Errorf("the pill %+v runs into the bar %+v", p, n.Box)
		}
	}
	if wide := homeWindow(1200, 800).At(time.Now()).Key("t").Narrowing(); wide.Compact {
		t.Error("a wide window draws the bar compact")
	}
}

func TestTheDetailOfALevelOneOffersASubObjectiveUnderIt(t *testing.T) {
	w := homeWindow(1200, 800)
	w.chosen, w.chosenGoal = 1, true
	d := w.Detail()
	at := slices.Index(d.Offers, AddSubSays)
	if at < 0 || len(d.Buttons) != len(d.Offers) {
		t.Fatalf("the detail offers %v at %v", d.Offers, d.Buttons)
	}
	b := d.Buttons[at]
	opened := w.clicked(Point{X: b.X + b.W/2, Y: b.Y + b.H/2})
	if opened.Screen() != OnForm || !opened.form.Adding || !opened.form.ForObjective || opened.form.ValueOf("parent") != "1" {
		t.Errorf("the offer opened %v adding %v with the parent %q", opened.Screen(), opened.form.Adding, opened.form.ValueOf("parent"))
	}
	if keyed := w.Key("shift+n"); keyed.form.ValueOf("parent") != "1" || keyed.Screen() != OnForm {
		t.Errorf("⇧N opened %v with the parent %q", keyed.Screen(), keyed.form.ValueOf("parent"))
	}
	w.chosen = 2
	if slices.Contains(w.Detail().Offers, AddSubSays) {
		t.Error("a sub-objective offers a sub-objective under it")
	}
	if keyed := w.Key("shift+n"); keyed.Screen() == OnForm {
		t.Error("⇧N on a sub-objective opened a form")
	}
	for _, size := range pinnedSizes {
		w := homeWindow(size.X, size.Y)
		w.chosen, w.chosenGoal = 1, true
		d := w.Detail()
		if len(d.Buttons) == 0 || d.Heading.W == 0 || d.Buttons[0].X < d.Heading.X+d.Heading.W {
			t.Errorf("at %v the offers %+v leave the heading %+v", size, d.Buttons, d.Heading)
		}
	}
}

func TestTheTreeTheTilesAndTheOffersAreLitUnderThePointer(t *testing.T) {
	w := homeWindow(1200, 800)
	w.chosen, w.chosenGoal = 1, true
	l := w.Layout()
	item := GoalAt(l, 1, 0)
	if got := w.Hovering(Point{X: item.X + 4, Y: item.Y + 4}).pointed; got != item {
		t.Errorf("over the second objective the pointer lights %+v", got)
	}
	d := w.Detail()
	tile := d.Tiles[0].At
	if got := w.Hovering(Point{X: tile.X + 4, Y: tile.Y + 4}).pointed; got != tile {
		t.Errorf("over a tile the pointer lights %+v", got)
	}
	button := d.Buttons[0]
	if got := w.Hovering(Point{X: button.X + 4, Y: button.Y + 4}).pointed; got != button {
		t.Errorf("over an offer the pointer lights %+v", got)
	}
	if got := w.Hovering(Point{X: 1, Y: 1}).pointed; got != (Box{}) {
		t.Errorf("over nothing the pointer lights %+v", got)
	}
}

func TestTheRefusalOfADoubleCountSaysWhatToUntick(t *testing.T) {
	w := homeWindow(1200, 800)
	w.tab, w.chosen = Tasks, 1
	w = onField(w.Edit(), "objective")
	w.choosing = 0
	if got := w.taking().ChoosingThen(); got != "Untick Enclosure and Firmware to put it on Home automation itself. Nothing was changed." {
		t.Errorf("ticking the parent says %q", got)
	}
	w.chosen = 3
	w = onField(w.Edit(), "objective")
	w.choosing = 1
	if got := w.taking().ChoosingThen(); got != "Untick Home automation to put it on Enclosure. Nothing was changed." {
		t.Errorf("ticking a sub-objective says %q", got)
	}
}

func TestAnObjectiveSaysNoTimeSpentWhileTheTimeIsNotShown(t *testing.T) {
	w := homeWindow(1200, 800)
	w.pages = engine.PagesOff(engine.PageTime)
	for _, f := range w.GoalProgress(w.snap.Goals[2]).Lines {
		if f.Name == "time spent" {
			t.Errorf("the time not shown, the progress says %+v", f)
		}
	}
	if d := w.Detail(); d.Spent != "" {
		t.Errorf("the time not shown, the detail says %q spent", d.Spent)
	}
	if said := w.spentSaid(""); said != "" {
		t.Errorf("the time not shown, a tile says %q spent", said)
	}
}

func TestTheProgressSaysTheDaysToTheDeadline(t *testing.T) {
	w := homeWindow(1200, 800).At(time.Date(2026, 11, 9, 10, 0, 0, 0, time.Local))
	g := w.snap.Goals[0]
	for day, want := range map[int]string{19: "2026-11-19, in 10 days", 10: "2026-11-10, in 1 day", 9: "2026-11-09, today", 6: "2026-11-06, 3 days past"} {
		g.Deadline = time.Date(2026, 11, day, 0, 0, 0, 0, time.Local)
		lines := w.GoalProgress(g).Lines
		if got := lines[len(lines)-1]; got != (Fact{Name: "deadline", Value: want}) {
			t.Errorf("due on the %d the progress says %+v", day, got)
		}
	}
	g.Deadline = time.Time{}
	for _, f := range w.GoalProgress(g).Lines {
		if f.Name == "deadline" {
			t.Errorf("with no deadline the progress says %+v", f)
		}
	}
}

func TestAPillWithNoRoomForItsTailSaysItUnderThePointer(t *testing.T) {
	w := homeWindow(1200, 800).At(time.Now())
	w.snap.Goals[0].Title = "Home automation of the house"
	w = w.Key("t")
	pill := w.Pills()[0]
	if pill.Tail != "" || !pill.Serves {
		t.Fatalf("the pill %+v kept its tail", pill)
	}
	if got := w.Hovering(Point{X: pill.Box.X + 4, Y: pill.Box.Y + 4}).Bubble(); got != pill.Said+", "+WithItsSubsSays {
		t.Errorf("under the pointer the pill says %q", got)
	}
}
