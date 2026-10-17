package migration

import (
	"testing"
)

func factsOf(m Migration) Carried {
	f := m.Carried()
	return f
}

func TestKnown_namesEveryChangeCairnKnows(t *testing.T) {
	if len(Known) != 14 {
		t.Fatalf("len(Known) = %d, want 14", len(Known))
	}
	for i, m := range Known {
		f := factsOf(m)
		if f.Name == "" || f.Does == "" {
			t.Errorf("change %d is incomplete: %+v", i, f)
		}
	}
}

func TestKnown_theFirstChangeCannotBeUndone(t *testing.T) {
	f := factsOf(Known[0])
	if f.Undoes != "" {
		t.Errorf("Known[0] undoes %q, want the floor to carry no undoing", f.Undoes)
	}
}
