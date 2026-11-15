package tui

import (
	"cairn/internal/board"
	"strings"
	"testing"

	"cairn/internal/engine"
)

func treeFake() *fake {
	f := newFake()
	f.goals = []board.Goal{
		{Number: 1, Title: "Home automation", State: "active", Children: []int{2, 3}, Total: 3},
		{Number: 2, Title: "Enclosure", State: "active", Parent: 1, Total: 1},
		{Number: 3, Title: "Firmware", State: "active", Parent: 1, Total: 1},
		{Number: 4, Title: "Features", State: "on-hold", WaitsOn: []int{3}},
	}
	f.cards = []board.Card{
		{Number: 1, Title: "Vent slots", State: "backlog", Priority: "normal", Objectives: []int{2}, Serves: []string{"Enclosure"}},
		{Number: 2, Title: "OTA updates", State: "backlog", Priority: "normal", Objectives: []int{3}, Serves: []string{"Firmware"}, Blockers: []int{4}},
		{Number: 3, Title: "Buy the relay", State: "backlog", Priority: "normal", Objectives: []int{1}, Serves: []string{"Home automation"}},
		{Number: 4, Title: "Order the board", State: "todo", Priority: "normal"},
	}
	return f
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

func TestZFoldsAParentAndTheCursorClimbsToIt(t *testing.T) {
	b := hit(standing(treeFake()).with(plainSkin), "tab", "j", "z")
	if b.goal != 1 {
		t.Errorf("folding from a sub-objective left the cursor on %d, want its parent 1", b.goal)
	}
	drawn := strip(b.Draw())
	if strings.Contains(drawn, "Enclosure") || !strings.Contains(drawn, "+2") {
		t.Errorf("a folded parent still shows its sub-objectives, or not how many:\n%s", drawn)
	}
	if b = hit(b, "j"); b.goal != 4 {
		t.Errorf("j from a folded parent reached %d, want 4", b.goal)
	}
	b = hit(b, "k", "z")
	if !strings.Contains(strip(b.Draw()), "Enclosure") {
		t.Errorf("z again did not unfold:\n%s", strip(b.Draw()))
	}
}

func TestBracketsWalkTheTasksAnObjectiveShows(t *testing.T) {
	b := hit(standing(treeFake()).with(plainSkin), "tab", "t")
	drawn := strip(b.Draw())
	if !strings.Contains(drawn, "‹ all 3 ›") || !strings.Contains(drawn, "[ ]") {
		t.Errorf("the tasks shown carry no sub-objective to walk:\n%s", drawn)
	}
	b = hit(b, "]")
	if got := b.cards(); len(got) != 1 || got[0].Number != 1 {
		t.Errorf("] reached %v, want the task of Enclosure", got)
	}
	b = hit(b, "[", "[")
	if got := b.cards(); len(got) != 1 || got[0].Number != 3 {
		t.Errorf("[ from all reached %v, want the objective itself", got)
	}
	if said := b.filters.Said(); said != "objective #1 Home automation, itself" {
		t.Errorf("the filter says %q", said)
	}
}
