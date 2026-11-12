package engine

import "testing"

func TestRender_writesTheContextAndMarksTheDeleted(t *testing.T) {
	if got := contextText("/notes/1"); got != "\ncontext  /notes/1\n" {
		t.Errorf("a context reads %q", got)
	}
	got := namedAsShown(map[int]string{1: "kept", 2: "gone"}, map[int]bool{2: true})
	if got[1] != "kept" || got[2] != "gone"+deletedMark {
		t.Errorf("the names read %v", got)
	}
}
