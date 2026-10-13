package engine

import (
	"slices"
	"testing"
)

func TestDeletingTasks_namesWhatGoesWithThem(t *testing.T) {
	snap := Snapshot{
		Cards: []Card{
			{Number: 1, Title: "Ship it"},
			{Number: 2, Title: "Part", Parent: 1},
			{Number: 3, Title: "Blocked", Blockers: []int{1}},
			{Number: 4, Title: "Plain"},
			{Number: 5, Title: "Waits on four", Blockers: []int{4}},
			{Number: 6, Title: "Both", Parent: 7},
			{Number: 7, Title: "Two parts"},
			{Number: 8, Title: "Second part", Parent: 7},
		},
		Focus: []int{4},
	}
	got := DeletingTasks(snap, []int{1, 4, 5, 7})
	want := []Flag{
		{Number: 1, Says: "its 1 part goes with it"},
		{Number: 1, Says: "task 3 waits on it"},
		{Number: 4, Says: "it is in focus"},
		{Number: 7, Says: "its 2 parts go with it"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("DeletingTasks = %+v, want %+v", got, want)
	}
}

func TestArchivingTasks_namesWhatCannotGo(t *testing.T) {
	snap := Snapshot{
		Cards: []Card{
			{Number: 1, Title: "Gone", Archived: true},
			{Number: 2, Title: "Ship it"},
			{Number: 3, Title: "Part", Parent: 2},
		},
		Focus: []int{2},
	}
	got := ArchivingTasks(snap, []int{1, 2})
	want := []Flag{
		{Number: 1, Says: "it is already archived", Refused: true},
		{Number: 2, Says: "its 1 part goes with it"},
		{Number: 2, Says: "it is in focus"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("ArchivingTasks = %+v, want %+v", got, want)
	}
}

func TestDeletingGoals_namesWhatWorkStillServes(t *testing.T) {
	snap := Snapshot{Cards: []Card{
		{Number: 1, Objectives: []int{2}},
		{Number: 2, Objectives: []int{2}},
		{Number: 3, Objectives: []int{3}},
	}}
	got := DeletingGoals(snap, []int{1, 2, 3})
	want := []Flag{
		{Number: 2, Says: "2 tasks serve it: it cannot go", Refused: true},
		{Number: 3, Says: "1 task serves it: it cannot go", Refused: true},
	}
	if !slices.Equal(got, want) {
		t.Errorf("DeletingGoals = %+v, want %+v", got, want)
	}
}

func TestLeavingOut_isEitherEveryFlagOrOnlyTheRefused(t *testing.T) {
	flags := []Flag{{Number: 1, Says: "x"}, {Number: 2, Says: "y", Refused: true}}
	if got := LeavingOut([]int{1, 2, 3}, flags, false); !slices.Equal(got, []int{3}) {
		t.Errorf("leaving out every flag = %v, want [3]", got)
	}
	if got := LeavingOut([]int{1, 2, 3}, flags, true); !slices.Equal(got, []int{1, 3}) {
		t.Errorf("leaving out only the refused = %v, want [1 3]", got)
	}
}

func TestSeveral_isTheFirstAndTheRestBesideIt(t *testing.T) {
	in := Several([]int{4, 7, 9})
	if in.Given["number"] != "4" || !slices.Equal(in.Listed["also"], []string{"7", "9"}) {
		t.Errorf("Several = %+v", in)
	}
}
