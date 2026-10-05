package reading

import "testing"

// One key takes a task up or lets it go, takes it off the board or brings it
// back: which, is what the card already says. Each board answered that its own
// way — the terminal from the focus list, the window from the card.
func TestToggles_areReadOffTheCard(t *testing.T) {
	if FocusOp(Card{}).Name() != "task_focus" || FocusOp(Card{InFocus: true}).Name() != "task_unfocus" {
		t.Error("taking up and letting go are not read off the card")
	}
	if AwayOp(Card{}).Name() != "task_archive" || AwayOp(Card{Archived: true}).Name() != "task_unarchive" {
		t.Error("taking off and bringing back are not read off the card")
	}
}

// A card or a goal found by its number in what was read. Looked for a dozen
// times over, by hand, in both boards.
func TestSnapshot_findsByNumber(t *testing.T) {
	s := Snapshot{Cards: []Card{{Number: 2, Title: "Two"}}, Goals: []Goal{{Number: 5, Title: "Five"}}}
	if c, ok := s.Card(2); !ok || c.Title != "Two" {
		t.Errorf("Card(2) = %+v, %v", c, ok)
	}
	if _, ok := s.Card(3); ok {
		t.Error("Card(3) was found")
	}
	if g, ok := s.Goal(5); !ok || g.Title != "Five" {
		t.Errorf("Goal(5) = %+v, %v", g, ok)
	}
	if _, ok := s.Goal(6); ok {
		t.Error("Goal(6) was found")
	}
}
