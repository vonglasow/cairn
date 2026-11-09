package tui

import (
	"strings"
	"testing"

	"cairn/internal/engine"
)

func treeFake() *fake {
	f := newFake()
	f.goals = []engine.Goal{
		{Number: 1, Title: "Home automation", State: "active", Children: []int{2, 3}, Total: 3},
		{Number: 2, Title: "Enclosure", State: "active", Parent: 1, Total: 1},
		{Number: 3, Title: "Firmware", State: "active", Parent: 1, Total: 1},
		{Number: 4, Title: "Features", State: "on-hold", WaitsOn: []int{3}},
	}
	f.cards = []engine.Card{
		{Number: 1, Title: "Vent slots", State: "backlog", Priority: "normal", Objectives: []int{2}, Serves: []string{"Enclosure"}},
		{Number: 2, Title: "OTA updates", State: "backlog", Priority: "normal", Objectives: []int{3}, Serves: []string{"Firmware"}, Blockers: []int{4}},
		{Number: 3, Title: "Buy the relay", State: "backlog", Priority: "normal", Objectives: []int{1}, Serves: []string{"Home automation"}},
		{Number: 4, Title: "Order the board", State: "todo", Priority: "normal"},
	}
	return f
}

func TestAnObjectiveIsTwoRows(t *testing.T) {
	b := standing(treeFake()).with(plainSkin)
	g, _ := b.snap.Goal(4)
	rows := b.goalRows(g, 80)
	if len(rows) != 2 {
		t.Fatalf("an objective's row is %d lines, want 2", len(rows))
	}
	if !strings.Contains(rows[0], "Features") || !strings.Contains(rows[1], "waits on Firmware") {
		t.Errorf("rows = %q", rows)
	}
}

func TestTheObjectivesAreDrawnAsATree(t *testing.T) {
	b := hit(standing(treeFake()).with(plainSkin), "tab")
	drawn := strip(b.Draw())
	for _, want := range []string{"├─", "└─", "Home automation", "Firmware"} {
		if !strings.Contains(drawn, want) {
			t.Errorf("the objectives never say %q:\n%s", want, drawn)
		}
	}
	if strings.Index(drawn, "Firmware") > strings.Index(drawn, "Features") {
		t.Errorf("a sub-objective is drawn after an objective of level one:\n%s", drawn)
	}
	b = hit(b, "j")
	if b.goal != 2 {
		t.Errorf("j walked to %d, want the sub-objective 2", b.goal)
	}
}

func TestBracketsWalkAnObjectivesSubObjectives(t *testing.T) {
	b := hit(standing(treeFake()).with(plainSkin), "tab", "enter")
	if got := b.Where().Showing; got != 3 {
		t.Fatalf("the objective's board carries %d tasks, want 3", got)
	}
	if !strings.Contains(strip(b.Draw()), "all 3") {
		t.Errorf("the band does not count all of it:\n%s", strip(b.Draw()))
	}
	b = hit(b, "]")
	if got := b.Where().Showing; got != 1 {
		t.Errorf("] narrowed to %d tasks, want 1", got)
	}
	b = hit(b, "[", "[")
	if got := b.cards(); len(got) != 1 || got[0].Number != 3 {
		t.Errorf("[ from all reached %v, want the objective itself", got)
	}
	b = hit(b, "esc", "enter")
	if got := b.Where().Showing; got != 3 {
		t.Errorf("opening it again kept the narrowing: %d tasks", got)
	}
}

func TestACardNamesWhatItServesButTheObjectiveOpen(t *testing.T) {
	b := standing(treeFake()).with(plainSkin)
	if !strings.Contains(strip(b.Draw()), "◇ Enclosure") {
		t.Errorf("the card does not name what it serves:\n%s", strip(b.Draw()))
	}
	b = hit(b, "tab", "enter")
	if strings.Contains(strip(b.Draw()), "◇ Home automation") {
		t.Errorf("inside Home automation, a card names it again:\n%s", strip(b.Draw()))
	}
}

func TestABlockerOutsideIsDrawnAndNotCounted(t *testing.T) {
	b := hit(standing(treeFake()).with(plainSkin), "tab", "enter")
	drawn := strip(b.Draw())
	if !strings.Contains(drawn, "Order the board") {
		t.Errorf("the blocker outside is not drawn:\n%s", drawn)
	}
	if !strings.Contains(drawn, "TODO") || strings.Contains(drawn, "TODO                   1") {
		t.Errorf("the reference is counted in its column:\n%s", drawn)
	}
	if !strings.Contains(drawn, "┆") {
		t.Errorf("the reference is not drawn dashed:\n%s", drawn)
	}
}

func TestTShowsTheTasksOfTheTree(t *testing.T) {
	b := hit(standing(treeFake()).with(plainSkin), "tab", "t")
	if b.tab != Tasks || b.inside != 0 {
		t.Fatalf("t left the board on tab %d, inside %d", b.tab, b.inside)
	}
	if got := b.Where().Showing; got != 3 {
		t.Errorf("t shows %d tasks, want 3", got)
	}
}

func TestAWarningIsSaidAboveTheBar(t *testing.T) {
	b := standing(treeFake()).with(plainSkin)
	model, _ := b.Update(Looked{Snapshot: b.snap, Warned: []string{"#4 Features waits on #3 Firmware, not reached"}})
	b = model.(Board)
	if !strings.Contains(strip(b.Draw()), "waits on #3 Firmware") {
		t.Fatalf("the warning is not said:\n%s", strip(b.Draw()))
	}
	b = hit(b, "j")
	if strings.Contains(strip(b.Draw()), "waits on #3 Firmware") {
		t.Errorf("the warning outlived the next key")
	}
}

func TestDeletingAParentAsksEachChoice(t *testing.T) {
	f := treeFake()
	b := hit(standing(f).with(plainSkin), "tab", "shift+d")
	if b.deleting == nil {
		t.Fatal("deleting a parent asked nothing")
	}
	b = hit(b, "enter")
	if len(f.deletedGoal) != 0 || !strings.Contains(strip(b.Draw()), "choose for each first") {
		t.Fatalf("enter deleted before every choice was made:\n%s", strip(b.Draw()))
	}
	b = hit(b, "l", "j", "l", "l", "j", "l", "enter")
	if len(f.deletedGoal) != 1 {
		t.Fatalf("every choice made, and nothing deleted: %+v", b.deleting)
	}
	var in engine.Input
	for _, a := range f.asked {
		if a.op.Verb == "delete" {
			in = a.in
		}
	}
	if strings.Join(in.Listed["keep"], " ") != "2" || strings.Join(in.Listed["delete"], " ") != "3" || in.Given["tasks"] != "detach" {
		t.Errorf("asked %+v", in)
	}
}

func TestEscapeKeepsAnObjectiveAskedAbout(t *testing.T) {
	f := treeFake()
	b := hit(standing(f).with(plainSkin), "tab", "shift+d", "l", "esc")
	if b.deleting != nil || len(f.deletedGoal) != 0 {
		t.Errorf("esc did not keep it: %+v, %v", b.deleting, f.deletedGoal)
	}
}

func TestTheParentFieldOffersTheObjectivesOfLevelOne(t *testing.T) {
	b := hit(standing(treeFake()).with(plainSkin), "tab", "shift+g", "e")
	drawn := strip(b.Draw())
	if !strings.Contains(drawn, "1 Home automation") || strings.Contains(drawn, "2 Enclosure") {
		t.Errorf("the parent field offers the wrong objectives:\n%s", drawn)
	}
}

func TestTheServesFieldIsATree(t *testing.T) {
	f := treeFake()
	f.cards[0].Objectives = []int{2, 3}
	b := hit(standing(f).with(plainSkin), "i", "e")
	if !strings.Contains(strip(b.Draw()), "counted through 2 below") {
		t.Errorf("the parent is not said to be counted through what is below:\n%s", strip(b.Draw()))
	}
}
