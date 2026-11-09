package store

import (
	"slices"
	"testing"
)

func TestKeptLabels_standOnTheirOwnAndAreRenamedAndLetGo(t *testing.T) {
	s := openStore(t)
	for _, label := range []string{"release", "doc", "release"} {
		if err := s.KeepLabel(label); err != nil {
			t.Fatalf("KeepLabel(%q): %v", label, err)
		}
	}
	if kept, err := s.KeptLabels(); err != nil || !slices.Equal(kept, []string{"doc", "release"}) {
		t.Fatalf("kept %v, %v", kept, err)
	}
	if err := s.RenameKeptLabel("release", "train"); err != nil {
		t.Fatal(err)
	}
	if err := s.ForgetLabel("doc"); err != nil {
		t.Fatal(err)
	}
	if kept, err := s.KeptLabels(); err != nil || !slices.Equal(kept, []string{"train"}) {
		t.Errorf("after a rename and a delete, kept %v, %v", kept, err)
	}
}

func TestLabelColours_areKeptRenamedAndLetGo(t *testing.T) {
	s := openStore(t)
	if err := s.ColourLabel("release", "peach"); err != nil {
		t.Fatal(err)
	}
	if err := s.ColourLabel("release", "teal"); err != nil {
		t.Fatal(err)
	}
	if err := s.ColourLabel("doc", "blue"); err != nil {
		t.Fatal(err)
	}
	if got, err := s.LabelColours(); err != nil || got["release"] != "teal" || got["doc"] != "blue" {
		t.Fatalf("the colours are %v, %v", got, err)
	}
	if err := s.RenameKeptLabel("release", "train"); err != nil {
		t.Fatal(err)
	}
	if err := s.ForgetLabel("doc"); err != nil {
		t.Fatal(err)
	}
	if got, err := s.LabelColours(); err != nil || len(got) != 1 || got["train"] != "teal" {
		t.Errorf("after a rename and a delete the colours are %v, %v", got, err)
	}
}

func TestLabelScopes_areKeptRenamedAndLetGo(t *testing.T) {
	s := openStore(t)
	if err := s.ScopeLabel("area", "tasks"); err != nil {
		t.Fatal(err)
	}
	if err := s.ScopeLabel("area", "objectives"); err != nil {
		t.Fatal(err)
	}
	if err := s.ScopeLabel("chore", "tasks"); err != nil {
		t.Fatal(err)
	}
	if got, err := s.LabelScopes(); err != nil || got["area"] != "objectives" || got["chore"] != "tasks" {
		t.Fatalf("the scopes are %v, %v", got, err)
	}
	if err := s.RenameKeptLabel("area", "domain"); err != nil {
		t.Fatal(err)
	}
	if err := s.ForgetLabel("chore"); err != nil {
		t.Fatal(err)
	}
	if got, err := s.LabelScopes(); err != nil || len(got) != 1 || got["domain"] != "objectives" {
		t.Errorf("after a rename and a delete the scopes are %v, %v", got, err)
	}
	if err := s.ScopeLabel("domain", "both"); err != nil {
		t.Fatal(err)
	}
	if got, err := s.LabelScopes(); err != nil || len(got) != 0 {
		t.Errorf("a label for both again keeps %v, %v", got, err)
	}
}
