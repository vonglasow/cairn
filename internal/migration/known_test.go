package migration

import (
	"strings"
	"testing"
)

func factsOf(m Migration) Carried {
	f := m.Carried()
	return f
}

func TestKnown_namesEveryChangeCairnKnows(t *testing.T) {
	if len(Known) != 18 {
		t.Fatalf("len(Known) = %d, want 18", len(Known))
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

func TestKnown_keepsALabelNothingCarries(t *testing.T) {
	f := factsOf(Known[15])
	if f.Name != "a label can stand on its own" || f.Loses == "" {
		t.Errorf("the last change is %+v", f)
	}
}

func TestKnown_givesALabelAColour(t *testing.T) {
	f := factsOf(Known[16])
	if f.Name != "a label wears a colour" || f.Loses == "" {
		t.Errorf("the last change is %+v", f)
	}
}

func TestKnown_putsAnObjectiveUnderAnother(t *testing.T) {
	f := factsOf(Known[len(Known)-1])
	if f.Name != "an objective can sit under another" {
		t.Fatalf("the last change is %+v", f)
	}
	if !strings.Contains(f.Does, "ADD COLUMN parent") || !strings.Contains(f.Undoes, "DROP COLUMN parent") {
		t.Errorf("the change does %q and undoes %q", f.Does, f.Undoes)
	}
	if !strings.Contains(f.Loses, "which objective each sub-objective sat under") {
		t.Errorf("rolling it back loses %q", f.Loses)
	}
}
