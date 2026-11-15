package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"slices"
	"testing"
	"time"

	"cairn/internal/engine"
)

func homeWindow(width, height int) Window {
	w := holding(newFake(),
		board.Card{Number: 1, Title: "Pin map", State: "backlog", Objectives: []int{2, 3}, Serves: []string{"Enclosure", "Firmware"}},
		board.Card{Number: 2, Title: "OTA updates", State: "done", Objectives: []int{3}, Serves: []string{"Firmware"}},
		board.Card{Number: 3, Title: "Buy the relay", State: "current", Objectives: []int{1}, Serves: []string{"Home automation"}},
	).Resized(width, height)
	w.Snap.Goals = []board.Goal{
		{Number: 1, Title: "Home automation", State: "active", Children: []int{2, 3}, Done: 1, Total: 3},
		{Number: 2, Title: "Enclosure", State: "active", Parent: 1, Done: 0, Total: 1},
		{Number: 3, Title: "Firmware", State: "active", Parent: 1, Done: 1, Total: 2, WaitedOnBy: []int{4}},
		{Number: 4, Title: "Tidy the workshop", State: "active", WaitsOn: []int{3}},
	}
	w.Tab = board.Objectives
	return w
}

func onField(w Window, arg string) Window {
	for i, f := range w.Form.Fields {
		if f.Arg == arg {
			w.Form = w.Form.Onto(i)
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
	w.Chosen, w.ChosenGoal = 4, true
	w = onField(w.Edit(), "parent")
	if got := choiceTitles(w); !slices.Equal(got, []string{"none", "Home automation"}) {
		t.Fatalf("the parent list reads %v", got)
	}
	if !w.OneOnly() {
		t.Error("the parent list takes more than one")
	}
	w.ChoosingOn = 1
	w = w.taking()
	if got := w.Form.Fields[w.Form.On()].Value; got != "1" {
		t.Errorf("choosing Home automation writes %q", got)
	}
	w = onField(w, "parent")
	w.ChoosingOn = 0
	if got := w.taking().Form.Fields[w.Form.On()].Value; got != "none" {
		t.Errorf("choosing none writes %q", got)
	}
}

func TestTheLabelsOfAFormAreThoseMeantForWhatItIsAbout(t *testing.T) {
	w := homeWindow(1200, 800)
	w.Snap.Goals[3].Labels = []string{"area"}
	w.Snap.Cards[0].Labels = []string{"backend"}
	w.Snap.Scopes = map[string]string{"area": "objectives", "backend": "tasks"}
	w.Chosen, w.ChosenGoal = 1, true
	if got := choiceTitles(onField(w.Edit(), "label")); !slices.Equal(got, []string{"area"}) {
		t.Errorf("an objective's labels offer %v", got)
	}
	w.Tab, w.Chosen, w.ChosenGoal = board.Tasks, 1, false
	if got := choiceTitles(onField(w.Edit(), "label")); !slices.Equal(got, []string{"backend"}) {
		t.Errorf("a task's labels offer %v", got)
	}
}

func TestDeletingAnObjectiveAsksRowByRowAndNamesWhatGoes(t *testing.T) {
	w := homeWindow(1200, 800)
	w.Chosen, w.ChosenGoal = 1, true
	w = w.Delete()
	if w.Screen != board.OnFates {
		t.Fatalf("deleting a parent left the window on %v", w.Screen)
	}
	f := w.FatesAsked()
	if len(f.Rows) != 3 || f.Rows[2].Sub != 0 || f.Rows[0].On != -1 {
		t.Fatalf("the question asks %+v", f.Rows)
	}
	for _, row := range f.Rows {
		if !f.Box.In(geom.Point{X: row.Choice.X, Y: row.Choice.Y}) || row.Choice.X < row.At.X {
			t.Errorf("the row %q chooses at %+v, outside %+v", row.Said, row.Choice, f.Box)
		}
	}
	if w.Key("enter").Screen != board.OnFates {
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
	gone.Chosen, gone.ChosenGoal = 1, true
	gone = gone.Delete().Key("l").Key("j").Key("l").Key("j").Key("l").Key("enter")
	in, asked := f2.gave("objective", "delete")
	if !asked || in.Given["tasks"] != "detach" || gone.Screen != board.OnTheBoard {
		t.Errorf("confirming asked %+v %v, on %v", in, asked, gone.Screen)
	}
	if cancelled := w.Choose("Cancel"); cancelled.Screen != board.OnTheBoard {
		t.Errorf("cancel left the window on %v", cancelled.Screen)
	}
	alone := homeWindow(1200, 800)
	alone.Chosen, alone.ChosenGoal = 4, true
	if alone.Delete().Screen != board.OnAsking {
		t.Error("an objective holding nothing is not asked about plainly")
	}
}

func TestANarrowWindowShowsThePartInForceBetweenTwoChevrons(t *testing.T) {
	w := homeWindow(geom.Narrowest, geom.Shortest).At(time.Now()).Key("t")
	n := w.Narrowing()
	if !n.Compact || n.Back.W == 0 || n.Next.W == 0 || n.At[n.On].W == 0 {
		t.Fatalf("at %d the bar is %+v", geom.Narrowest, n)
	}
	if n.Back.X+n.Back.W > n.At[n.On].X || n.Next.X < n.At[n.On].X+n.At[n.On].W || n.Box.X+n.Box.W > w.width-geom.Margin {
		t.Errorf("the chevrons %+v %+v do not stand either side of %+v inside %+v", n.Back, n.Next, n.At[n.On], n.Box)
	}
	next := w.clicked(geom.Point{X: n.Next.X + n.Next.W/2, Y: n.Next.Y + n.Next.H/2})
	if taken, _ := next.ObjectiveTaken(); taken != "2" {
		t.Errorf("the next chevron narrowed to %q", taken)
	}
	n = next.Narrowing()
	back := next.clicked(geom.Point{X: n.Back.X + n.Back.W/2, Y: n.Back.Y + n.Back.H/2})
	if taken, _ := back.ObjectiveTaken(); taken != "1" {
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
	w.Chosen, w.ChosenGoal = 1, true
	d := w.Detail()
	at := slices.Index(d.Offers, AddSubSays)
	if at < 0 || len(d.Buttons) != len(d.Offers) {
		t.Fatalf("the detail offers %v at %v", d.Offers, d.Buttons)
	}
	b := d.Buttons[at]
	opened := w.clicked(geom.Point{X: b.X + b.W/2, Y: b.Y + b.H/2})
	if opened.Screen != board.OnForm || !opened.Form.Adding || !opened.Form.ForObjective || valueOf(opened.Form, "parent") != "1" {
		t.Errorf("the offer opened %v adding %v with the parent %q", opened.Screen, opened.Form.Adding, valueOf(opened.Form, "parent"))
	}
	if keyed := w.Key("shift+n"); valueOf(keyed.Form, "parent") != "1" || keyed.Screen != board.OnForm {
		t.Errorf("⇧N opened %v with the parent %q", keyed.Screen, valueOf(keyed.Form, "parent"))
	}
	w.Chosen = 2
	if slices.Contains(w.Detail().Offers, AddSubSays) {
		t.Error("a sub-objective offers a sub-objective under it")
	}
	if keyed := w.Key("shift+n"); keyed.Screen == board.OnForm {
		t.Error("⇧N on a sub-objective opened a form")
	}
	for _, size := range pinnedSizes {
		w := homeWindow(size.X, size.Y)
		w.Chosen, w.ChosenGoal = 1, true
		d := w.Detail()
		if len(d.Buttons) == 0 || d.Heading.W == 0 || d.Buttons[0].X < d.Heading.X+d.Heading.W {
			t.Errorf("at %v the offers %+v leave the heading %+v", size, d.Buttons, d.Heading)
		}
	}
}

func TestTheTreeTheTilesAndTheOffersAreLitUnderThePointer(t *testing.T) {
	w := homeWindow(1200, 800)
	w.Chosen, w.ChosenGoal = 1, true
	l := w.Layout()
	item := GoalAt(l, 1, 0)
	if got := w.Hovering(geom.Point{X: item.X + 4, Y: item.Y + 4}).pointed; got != item {
		t.Errorf("over the second objective the pointer lights %+v", got)
	}
	d := w.Detail()
	tile := d.Tiles[0].At
	if got := w.Hovering(geom.Point{X: tile.X + 4, Y: tile.Y + 4}).pointed; got != tile {
		t.Errorf("over a tile the pointer lights %+v", got)
	}
	button := d.Buttons[0]
	if got := w.Hovering(geom.Point{X: button.X + 4, Y: button.Y + 4}).pointed; got != button {
		t.Errorf("over an offer the pointer lights %+v", got)
	}
	if got := w.Hovering(geom.Point{X: 1, Y: 1}).pointed; got != (geom.Box{}) {
		t.Errorf("over nothing the pointer lights %+v", got)
	}
}

func TestTheRefusalOfADoubleCountSaysWhatToUntick(t *testing.T) {
	w := homeWindow(1200, 800)
	w.Tab, w.Chosen = board.Tasks, 1
	w = onField(w.Edit(), "objective")
	w.ChoosingOn = 0
	if got := w.taking().ChoosingThen(); got != "Untick Enclosure and Firmware to put it on Home automation itself. Nothing was changed." {
		t.Errorf("ticking the parent says %q", got)
	}
	w.Chosen = 3
	w = onField(w.Edit(), "objective")
	w.ChoosingOn = 1
	if got := w.taking().ChoosingThen(); got != "Untick Home automation to put it on Enclosure. Nothing was changed." {
		t.Errorf("ticking a sub-objective says %q", got)
	}
}

func TestAnObjectiveSaysNoTimeSpentWhileTheTimeIsNotShown(t *testing.T) {
	w := homeWindow(1200, 800)
	w.Pages = engine.PagesOff(engine.PageTime)
	for _, f := range w.GoalProgress(w.Snap.Goals[2]).Lines {
		if f.Name == "time spent" {
			t.Errorf("the time not shown, the progress says %+v", f)
		}
	}
	if d := w.Detail(); d.Spent != "" {
		t.Errorf("the time not shown, the detail says %q spent", d.Spent)
	}
	if said := w.SpentSaid(""); said != "" {
		t.Errorf("the time not shown, a tile says %q spent", said)
	}
}

func TestTheProgressSaysTheDaysToTheDeadline(t *testing.T) {
	w := homeWindow(1200, 800).At(time.Date(2026, 11, 9, 10, 0, 0, 0, time.Local))
	g := w.Snap.Goals[0]
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
	w.Snap.Goals[0].Title = "Home automation of the house"
	w = w.Key("t")
	pill := w.Pills()[0]
	if pill.Tail != "" || !pill.Serves {
		t.Fatalf("the pill %+v kept its tail", pill)
	}
	if got := w.Hovering(geom.Point{X: pill.Box.X + 4, Y: pill.Box.Y + 4}).Bubble(); got != pill.Said+", "+WithItsSubsSays {
		t.Errorf("under the pointer the pill says %q", got)
	}
}
