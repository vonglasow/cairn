package migration

import (
	"testing"
	"time"
)

// facts reads back what a test needs from a change, without reaching into
// fields that stopped being exported.
type facts struct {
	number                    int
	name, does, undoes, loses string
	applied                   bool
}

func (f *facts) Numbered(n int)      { f.number = n }
func (f *facts) Named(s string)      { f.name = s }
func (f *facts) Does(s string)       { f.does = s }
func (f *facts) Undoes(s string)     { f.undoes = s }
func (f *facts) NotUndoable()        { f.undoes = "" }
func (f *facts) Loses(s string)      { f.loses = s }
func (f *facts) NothingLost()        { f.loses = "" }
func (f *facts) Applied(t time.Time) { f.applied = true }
func (f *facts) NotApplied()         { f.applied = false }

func factsOf(m Migration) facts {
	var f facts
	m.Tell(&f)
	return f
}

func TestKnown_namesEveryChangeCairnKnows(t *testing.T) {
	if len(Known) != 5 {
		t.Fatalf("len(Known) = %d, want 5", len(Known))
	}
	for i, m := range Known {
		f := factsOf(m)
		if f.name == "" || f.does == "" {
			t.Errorf("change %d is incomplete: %+v", i, f)
		}
	}
}

// The first change is the floor: it has no undoing, because unmaking it would
// mean throwing the store away, which is what deleting the file is for.
func TestKnown_theFirstChangeCannotBeUndone(t *testing.T) {
	f := factsOf(Known[0])
	if f.undoes != "" {
		t.Errorf("Known[0] undoes %q, want the floor to carry no undoing", f.undoes)
	}
}
