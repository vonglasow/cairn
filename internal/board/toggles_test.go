package board

import "testing"

func TestToggles_areReadOffTheCard(t *testing.T) {
	if FocusOp(Card{}).Name() != "task_focus" || FocusOp(Card{InFocus: true}).Name() != "task_unfocus" {
		t.Error("taking up and letting go are not read off the card")
	}
	if AwayOp(Card{}).Name() != "task_archive" || AwayOp(Card{Archived: true}).Name() != "task_unarchive" {
		t.Error("taking off and bringing back are not read off the card")
	}
}
