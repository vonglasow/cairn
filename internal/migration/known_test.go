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
	if len(Known) != 22 {
		t.Fatalf("len(Known) = %d, want 22", len(Known))
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
	f := factsOf(Known[17])
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

func TestKnown_givesALabelAScope(t *testing.T) {
	f := factsOf(Known[18])
	if f.Name != "a label has a scope" {
		t.Fatalf("the change is %+v", f)
	}
	if !strings.Contains(f.Does, "CREATE TABLE IF NOT EXISTS label_scopes") || !strings.Contains(f.Undoes, "DROP TABLE IF EXISTS label_scopes") {
		t.Errorf("the change does %q and undoes %q", f.Does, f.Undoes)
	}
	if !strings.Contains(f.Loses, "the scope of each label") {
		t.Errorf("rolling it back loses %q", f.Loses)
	}
}

func TestKnown_letsAnObjectiveWaitOnAnother(t *testing.T) {
	f := factsOf(Known[19])
	if f.Name != "an objective can wait on another" {
		t.Fatalf("the change is %+v", f)
	}
	if !strings.Contains(f.Does, "CREATE TABLE IF NOT EXISTS objective_waits") || !strings.Contains(f.Undoes, "DROP TABLE IF EXISTS objective_waits") {
		t.Errorf("the change does %q and undoes %q", f.Does, f.Undoes)
	}
	if !strings.Contains(f.Loses, "which objective waits on which") {
		t.Errorf("rolling it back loses %q", f.Loses)
	}
}

func TestKnown_keepsViews(t *testing.T) {
	f := factsOf(Known[20])
	if f.Name != "a view keeps the filters under a name" {
		t.Fatalf("the change is %+v", f)
	}
	for _, table := range []string{"views", "view_objectives", "view_labels"} {
		if !strings.Contains(f.Does, "CREATE TABLE IF NOT EXISTS "+table+" ") || !strings.Contains(f.Undoes, "DROP TABLE IF EXISTS "+table+";") {
			t.Errorf("the change does %q and undoes %q, want the table %s made and dropped", f.Does, f.Undoes, table)
		}
	}
	if !strings.Contains(f.Loses, "every view") {
		t.Errorf("rolling it back loses %q", f.Loses)
	}
}

func TestKnown_keepsSeries(t *testing.T) {
	f := factsOf(Known[21])
	if f.Name != "a task can come back, as a series" {
		t.Fatalf("the change is %+v", f)
	}
	for _, table := range []string{"series", "series_labels", "series_objectives", "series_acts"} {
		if !strings.Contains(f.Does, "CREATE TABLE IF NOT EXISTS "+table+" ") || !strings.Contains(f.Undoes, "DROP TABLE IF EXISTS "+table+";") {
			t.Errorf("the change does %q and undoes %q, want the table %s made and dropped", f.Does, f.Undoes, table)
		}
	}
	if !strings.Contains(f.Loses, "every series") || !strings.Contains(f.Loses, "every task") {
		t.Errorf("rolling it back loses %q", f.Loses)
	}
}
