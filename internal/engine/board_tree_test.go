package engine

import (
	"slices"
	"testing"
	"time"

	"cairn/internal/domain"
)

func homeTree(t *testing.T) *Engine {
	t.Helper()
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	under(t, e, "Enclosure", "1")
	under(t, e, "Firmware", "1")
	addObjective(t, e, "Tidy the workshop")
	servingAll(t, e, "Pin map", 2, 3)
	servingAll(t, e, "Buy the relay", 1)
	return e
}

func lookAt(t *testing.T, e *Engine) Snapshot {
	t.Helper()
	found, err := Look(e)
	if err != nil {
		t.Fatalf("Look: %v", err)
	}
	return found
}

func TestLook_seatsEachObjectiveInItsTree(t *testing.T) {
	e := homeTree(t)
	if _, err := e.Ask(ObjectiveEdit, Input{Given: map[string]string{"number": "4"}, Listed: map[string][]string{"waits-on": {"3"}}}); err != nil {
		t.Fatalf("waits-on: %v", err)
	}
	s := lookAt(t, e)
	home, _ := s.Goal(1)
	if !slices.Equal(home.Children, []int{2, 3}) || home.Parent != 0 {
		t.Errorf("Home automation = parent %d, children %v", home.Parent, home.Children)
	}
	firmware, _ := s.Goal(3)
	if firmware.Parent != 1 {
		t.Errorf("Firmware sits under %d, want 1", firmware.Parent)
	}
	tidy, _ := s.Goal(4)
	if !slices.Equal(tidy.WaitsOn, []int{3}) {
		t.Errorf("Tidy waits on %v, want [3]", tidy.WaitsOn)
	}
	if home.Total != 2 {
		t.Errorf("Home automation counts %d tasks, want 2, each once", home.Total)
	}
}

func TestLook_spendsTheTimeOfASubObjectiveOnItsParentOnce(t *testing.T) {
	e := homeTree(t)
	if _, err := e.Ask(TimesheetAdd, Input{Given: map[string]string{"number": "1", "time": "1h", "on": "2026-09-01"}}); err != nil {
		t.Fatalf("timesheet: %v", err)
	}
	s := lookAt(t, e)
	home, _ := s.Goal(1)
	if home.Spent != time.Hour {
		t.Errorf("Home automation spent %v, want 1h, counted once", home.Spent)
	}
}

func TestLook_namesWhatACardServes(t *testing.T) {
	s := lookAt(t, homeTree(t))
	pin, _ := s.Card(1)
	if !slices.Equal(pin.Serves, []string{"Enclosure", "Firmware"}) {
		t.Errorf("Pin map serves %v, want Enclosure and Firmware by name", pin.Serves)
	}
}

func TestLook_keepsWhatEachLabelIsFor(t *testing.T) {
	e := openEngine(t)
	for label, scope := range map[string]string{"area": "objectives", "backend": "tasks", "spec": ""} {
		if _, err := e.Ask(LabelNew, Input{Given: map[string]string{"label": label, "scope": scope}}); err != nil {
			t.Fatalf("label new %s: %v", label, err)
		}
	}
	s := lookAt(t, e)
	if got := s.LabelsFor(domain.OnObjectives); !slices.Equal(got, []string{"area", "spec"}) {
		t.Errorf("labels for objectives = %v, want area and spec", got)
	}
}

func TestBranches_drawEachSubObjectiveUnderItsParent(t *testing.T) {
	s := lookAt(t, homeTree(t))
	var got []int
	var under, last []bool
	for _, b := range Branches(s.Goals) {
		got = append(got, b.Goal.Number)
		under, last = append(under, b.Under), append(last, b.Last)
	}
	if !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Errorf("the tree reads %v, want 1 2 3 4", got)
	}
	if !slices.Equal(under, []bool{false, true, true, false}) || !slices.Equal(last, []bool{false, false, true, false}) {
		t.Errorf("under %v, last %v", under, last)
	}
}

func TestBranches_putASubObjectiveWhoseParentIsLeftOutAtLevelOne(t *testing.T) {
	s := lookAt(t, homeTree(t))
	firmware, _ := s.Goal(3)
	got := Branches([]Goal{firmware})
	if len(got) != 1 || got[0].Under {
		t.Errorf("a sub-objective alone = %+v, want it at level one", got)
	}
}

func TestSnapshot_reachesAnObjectiveAndItsSubObjectives(t *testing.T) {
	s := lookAt(t, homeTree(t))
	if got := s.Reach(1); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("Reach(1) = %v, want 1 2 3", got)
	}
	if got := s.Reach(3); !slices.Equal(got, []int{3}) {
		t.Errorf("Reach(3) = %v, want 3", got)
	}
}

func TestSnapshot_segmentsAnObjectiveIntoItsSubObjectives(t *testing.T) {
	s := lookAt(t, homeTree(t))
	got := s.Segments(1, s.Cards)
	var said []string
	for _, seg := range got {
		said = append(said, seg.Name+" "+Numbers([]int{seg.Count}))
	}
	if !slices.Equal(said, []string{"all 2", "Enclosure 1", "Firmware 1", "itself 1"}) {
		t.Errorf("segments = %v", said)
	}
	if !slices.Equal(got[3].Reach, []int{1}) || !slices.Equal(got[0].Reach, []int{1, 2, 3}) {
		t.Errorf("itself reaches %v, all reaches %v", got[3].Reach, got[0].Reach)
	}
	if none := s.Segments(4, s.Cards); none != nil {
		t.Errorf("an objective with no sub-objective has segments %v", none)
	}
}

func TestServing_keepsWhatServesAnyOfThem(t *testing.T) {
	cards := []Card{{Number: 1, Objectives: []int{2}}, {Number: 2, Objectives: []int{3}}, {Number: 3}}
	if got := filteredNumbers(Serving(cards, 2, 3)); got != "1 2" {
		t.Errorf("Serving(2, 3) = %q, want 1 2", got)
	}
}

func TestBlocking_findsTheBlockersLeftOut(t *testing.T) {
	all := []Card{{Number: 1, Title: "Order the relay"}, {Number: 2, Blockers: []int{1, 9}}, {Number: 3, Blockers: []int{2}}}
	shown := []Card{all[1], all[2]}
	if got := filteredNumbers(Blocking(shown, all)); got != "1" {
		t.Errorf("Blocking = %q, want 1, the one left out", got)
	}
}

func TestBoardFilters_widenAParentToItsSubObjectives(t *testing.T) {
	s := lookAt(t, homeTree(t))
	fs := BoardFilters(s, time.Now()).On("objective", "1")
	if !fs.NarrowsBy("objective") || fs.NarrowsBy("label") {
		t.Errorf("NarrowsBy reads %v for objective, %v for label", fs.NarrowsBy("objective"), fs.NarrowsBy("label"))
	}
	kept, err := fs.Kept(s.Cards)
	if err != nil {
		t.Fatalf("Kept: %v", err)
	}
	if got := filteredNumbers(kept); got != "1 2" {
		t.Errorf("the parent keeps %q, want 1 2", got)
	}
	if said := fs.Said(); said != "objective #1 Home automation, with its sub-objectives" {
		t.Errorf("Said = %q", said)
	}
	itself := fs.On("objective", "1 itself")
	kept, _ = itself.Kept(s.Cards)
	if got := filteredNumbers(kept); got != "2" {
		t.Errorf("the parent itself keeps %q, want 2", got)
	}
	if said := itself.Said(); said != "objective #1 Home automation, itself" {
		t.Errorf("Said = %q", said)
	}
	if goals := itself.KeptGoals(s.Goals); len(goals) != 1 {
		t.Errorf("the parent itself keeps %d objectives, want 1", len(goals))
	}
	if goals := fs.KeptGoals(s.Goals); len(goals) != 3 {
		t.Errorf("the parent keeps %d objectives, want 3", len(goals))
	}
}

func TestPicks_countAParentThroughWhatIsTickedBelowIt(t *testing.T) {
	s := lookAt(t, homeTree(t))
	picks := s.Picks([]int{2, 3})
	if picks[0].Goal.Number != 1 || picks[0].Ticked || picks[0].Through != 2 {
		t.Errorf("the parent = %+v, want counted through 2 below", picks[0])
	}
	if !picks[1].Ticked || !picks[1].Under {
		t.Errorf("Enclosure = %+v, want ticked, under", picks[1])
	}
}

func TestParents_areTheObjectivesOfLevelOneButItself(t *testing.T) {
	s := lookAt(t, homeTree(t))
	var got []int
	for _, g := range s.Parents(4) {
		got = append(got, g.Number)
	}
	if !slices.Equal(got, []int{1}) {
		t.Errorf("Parents(4) = %v, want 1", got)
	}
}

func TestGoalFields_sayWhereItSitsAndWhatItWaitsOn(t *testing.T) {
	fields := GoalFields(Goal{Parent: 4, WaitsOn: []int{2, 3}})
	parent, _ := valueOf(fields, "parent")
	waits, _ := valueOf(fields, "waits on")
	if parent != "4" || waits != "2 3" {
		t.Errorf("parent %q, waits on %q", parent, waits)
	}
	if none, _ := valueOf(GoalFields(Goal{}), "parent"); none != "" {
		t.Errorf("a parent never set reads %q, want nothing", none)
	}
}

func TestDeleting_asksForEachSubObjectiveAndTheTasksTouched(t *testing.T) {
	s := lookAt(t, homeTree(t))
	d := s.Deleting(1)
	if !slices.Equal(d.Subs, []int{2, 3}) {
		t.Fatalf("Subs = %v", d.Subs)
	}
	if !d.Asks() || d.Ready() {
		t.Fatalf("a parent with sub-objectives was ready before anything was chosen")
	}
	if !d.AsksTasks() {
		t.Errorf("Home automation carries a task itself, so its tasks are asked about")
	}
	d = d.Choose(2, KeepIt).Choose(3, DeleteIt).Tasks(DetachThem)
	if !d.Ready() {
		t.Fatalf("every choice made, and still not ready: %+v", d)
	}
	in := d.Input()
	if in.Given["number"] != "1" || in.Given["tasks"] != "detach" ||
		!slices.Equal(in.Listed["keep"], []string{"2"}) || !slices.Equal(in.Listed["delete"], []string{"3"}) {
		t.Errorf("Input = %+v", in)
	}
}

func TestDeleting_asksNothingOfAnObjectiveAlone(t *testing.T) {
	s := lookAt(t, homeTree(t))
	if s.Deleting(4).Asks() {
		t.Error("an objective with nothing under it asked what goes with it")
	}
}

func TestDeleting_asksAboutTasksOnlyOnceTheyAreTouched(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	under(t, e, "Enclosure", "1")
	servingAll(t, e, "Vent slots", 2)
	d := lookAt(t, e).Deleting(1)
	if d.AsksTasks() {
		t.Error("the tasks were asked about before any sub-objective holding them was deleted")
	}
	if got := d.Choose(2, DeleteIt).Touched(); !slices.Equal(got, []int{1}) {
		t.Errorf("Touched = %v, want the task of the sub-objective deleted", got)
	}
	if !d.Choose(2, DeleteIt).AsksTasks() {
		t.Error("deleting the sub-objective holding a task did not ask about it")
	}
	if d.Choose(2, KeepIt).Ready() != true {
		t.Error("keeping the only sub-objective left nothing to ask, and it was not ready")
	}
}
