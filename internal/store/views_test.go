package store

import (
	"reflect"
	"testing"
)

func TestViews_areKeptInTheirPlacesAndReplacedWhole(t *testing.T) {
	s := openStore(t)
	if views, err := s.Views(); err != nil || len(views) != 0 {
		t.Fatalf("a new store holds the views %v, %v", views, err)
	}
	kept := []View{
		{Name: "Home", Objectives: []int{4, 2}, Itself: true, Search: "spec", Labels: []string{"doc"}},
		{Name: "Next", State: "todo", Priority: "high", DueBefore: "2026-11-30", Focus: "in", Archived: true, WithArchived: true, Objectives: []int{}, Labels: []string{}},
	}
	if err := s.KeepViews(kept); err != nil {
		t.Fatal(err)
	}
	views, err := s.Views()
	if err != nil {
		t.Fatal(err)
	}
	kept[0].Place, kept[0].Objectives = 1, []int{2, 4}
	kept[1].Place = 2
	if !reflect.DeepEqual(views, kept) {
		t.Errorf("the views read\n%+v\nwant\n%+v", views, kept)
	}
	if err := s.KeepViews(kept[1:]); err != nil {
		t.Fatal(err)
	}
	if views, err := s.Views(); err != nil || len(views) != 1 || views[0].Name != "Next" || views[0].Place != 1 {
		t.Errorf("replaced by one, the views read %+v, %v", views, err)
	}
	if objectives, err := rowsOf(s, one[int], `SELECT objective_number FROM view_objectives`); err != nil || len(objectives) != 0 {
		t.Errorf("what a view let go of is still kept: %v, %v", objectives, err)
	}
}
