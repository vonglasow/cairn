package board

import (
	"cairn/internal/engine"
	"cairn/internal/engine/label"
	"slices"
	"testing"
	"time"
)

func homeTree(t *testing.T) *engine.Engine {
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

func lookAt(t *testing.T, e *engine.Engine) Snapshot {
	t.Helper()
	found, err := Look(e)
	if err != nil {
		t.Fatalf("Look: %v", err)
	}
	return found
}

func TestLook_keepsWhatEachLabelIsFor(t *testing.T) {
	e := openEngine(t)
	for name, scope := range map[string]string{"area": "objectives", "backend": "tasks", "spec": ""} {
		if _, err := e.Ask(engine.LabelNew, engine.Input{Given: map[string]string{"label": name, "scope": scope}}); err != nil {
			t.Fatalf("label new %s: %v", name, err)
		}
	}
	s := lookAt(t, e)
	if got := s.LabelsFor(label.OnObjectives); !slices.Equal(got, []string{"area", "spec"}) {
		t.Errorf("labels for objectives = %v, want area and spec", got)
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
	var values []string
	for _, seg := range got {
		values = append(values, seg.Value)
	}
	if !slices.Equal(values, []string{"1", "2", "3", "1 itself"}) {
		t.Errorf("the segments stand for the objective filter's values %q", values)
	}
	if none := s.Segments(4, s.Cards); none != nil {
		t.Errorf("an objective with no sub-objective has segments %v", none)
	}
}

func TestBoardFilters_widenAParentToItsSubObjectives(t *testing.T) {
	s := lookAt(t, homeTree(t))
	fs := FiltersOf(s, time.Now()).On("objective", "1")
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

func TestCountedOnce_refusesAParentTickedBesideItsSubObjectives(t *testing.T) {
	s := lookAt(t, homeTree(t))
	err := s.CountedOnce(44, []int{2, 1, 3})
	if err == nil || err.Error() != "#44 is already counted in #1 Home automation through #2 Enclosure and #3 Firmware; nothing was changed" {
		t.Errorf("CountedOnce = %v", err)
	}
	if err := s.CountedOnce(0, []int{2, 3, 4}); err != nil {
		t.Errorf("two sub-objectives of one parent and another tree were refused: %v", err)
	}
	if err := s.CountedOnce(0, []int{1, 99}); err != nil {
		t.Errorf("an objective that is not there was judged: %v", err)
	}
}

func TestLeft_countsWhatIsOpenOverTheSubtreeOnce(t *testing.T) {
	e := homeTree(t)
	if _, err := e.Ask(engine.TaskEdit, engine.Input{Given: map[string]string{"number": "1", "deadline": "2026-12-01"}}); err != nil {
		t.Fatalf("deadline: %v", err)
	}
	s := lookAt(t, e)
	if got := s.Left(1); got.Tasks != 2 || got.On != 1 {
		t.Errorf("Home automation has %+v left, want 2 tasks, the next due #1", got)
	}
	if got := s.Left(4); got.Tasks != 0 {
		t.Errorf("Tidy the workshop has %+v left, want nothing", got)
	}
}

func TestSnapshot_treeOfIsTheLevelOneObjectiveAboveIt(t *testing.T) {
	s := lookAt(t, homeTree(t))
	for n, want := range map[int]int{1: 1, 2: 1, 3: 1, 4: 4, 9: 9} {
		if got := s.TreeOf(n); got != want {
			t.Errorf("TreeOf(%d) = %d, want %d", n, got, want)
		}
	}
}

func TestDeleting_countsTheTasksOfASubDeletedAfterOneKept(t *testing.T) {
	s := Snapshot{
		Goals: []Goal{{Number: 1, Children: []int{2, 3}}, {Number: 2, Parent: 1}, {Number: 3, Parent: 1}},
		Cards: []Card{{Number: 7, Objectives: []int{2}}, {Number: 8, Objectives: []int{3}}},
	}
	d := s.Deleting(1).Choose(2, KeepIt).Choose(3, DeleteIt)
	if touched := d.Touched(); len(touched) != 1 || touched[0] != 8 {
		t.Errorf("touched %v", touched)
	}
	if d.FateOf(2) != KeepIt {
		t.Errorf("the first sub's fate is %q", d.FateOf(2))
	}
}

func TestLook_namesWhatACardServesUnderItsParent(t *testing.T) {
	s := lookAt(t, homeTree(t))
	var got [][]string
	for _, c := range s.Cards {
		got = append(got, c.Serves)
	}
	want := [][]string{{"Home automation › Enclosure", "Home automation › Firmware"}, {"Home automation"}}
	if !slices.EqualFunc(got, want, slices.Equal) {
		t.Errorf("serves = %q, want %q", got, want)
	}
}
