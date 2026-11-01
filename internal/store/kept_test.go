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
