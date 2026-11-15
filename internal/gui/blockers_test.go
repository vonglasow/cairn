package gui

import (
	"cairn/internal/board"
	"slices"
	"testing"
)

func TestTheBlockersListOffersTheOtherTasksAndWritesTheTicked(t *testing.T) {
	w := holding(newFake(),
		board.Card{Number: 1, Title: "Fix the login", State: "todo", Blockers: []int{3}},
		board.Card{Number: 2, Title: "Write the docs", State: "backlog", Labels: []string{"login"}},
		board.Card{Number: 3, Title: "Plan it", State: "done"},
	)
	w = onField(w.clicked(middleOf(w, 1, 0)).Choose("Edit"), "blocker")
	if got := choiceTitles(w); !slices.Equal(got, []string{"Plan it", "Write the docs"}) {
		t.Fatalf("the blockers list offers %v, want the other tasks, the newest first", got)
	}
	w = w.Types("nothing")
	if got := w.EmptyListSays(); len(w.Choosing()) != 0 || got != board.TasksNoneSays {
		t.Fatalf("a search reaching nothing offers %v and reads %q", choiceTitles(w), got)
	}
	w = w.Key("backspace").Key("backspace").Key("backspace").Key("backspace").Key("backspace").Key("backspace").Key("backspace").Types("login")
	if got := choiceTitles(w); !slices.Equal(got, []string{"Write the docs"}) {
		t.Fatalf("typing login offers %v, want the task it reaches", got)
	}
	w = w.Key("enter").Key("cmd+enter")
	if got := valueOf(w.Form, "blockers"); got != "2 3" {
		t.Errorf("the blockers field reads %q, want 2 and the 3 the search hid", got)
	}
}
