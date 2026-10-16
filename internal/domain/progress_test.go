package domain

import "testing"

func TestProgress_countsWhatIsThere(t *testing.T) {
	p := Progress{Tasks: Tally{Total: 3, Done: 1}, Steps: Tally{Total: 4, Done: 2}}
	if !p.Any() {
		t.Error("Any() = false, want true: seven things are under it")
	}
	if got := p.Percent(); got != 43 {
		t.Errorf("Percent() = %d, want 43", got)
	}
}

func TestProgress_ofATaskCountsItsStepsOnly(t *testing.T) {
	p := Progress{Steps: Tally{Total: 2, Done: 1}}
	if got := p.Percent(); got != 50 {
		t.Errorf("Percent() = %d, want 50", got)
	}
}

func TestProgress_ofNothingIsNothing(t *testing.T) {
	var p Progress
	if p.Any() {
		t.Error("Any() = true, want false: nothing is under it")
	}
	if got := p.Percent(); got != 0 {
		t.Errorf("Percent() = %d, want 0: a share of nothing divides by nothing", got)
	}
}

func TestProgress_isWholeOnlyWhenEverythingIsDone(t *testing.T) {
	whole := Progress{Tasks: Tally{Total: 3, Done: 3}}
	if got := whole.Percent(); got != 100 {
		t.Errorf("Percent() = %d, want 100", got)
	}
}

func TestProgress_neverRoundsUpToWhole(t *testing.T) {
	almost := Progress{Tasks: Tally{Total: 300, Done: 299}}
	if got := almost.Percent(); got != 99 {
		t.Errorf("Percent() = %d, want 99: one task is still not done", got)
	}
}

func TestProgress_neverRoundsDownToNothing(t *testing.T) {
	barely := Progress{Tasks: Tally{Total: 300, Done: 1}}
	if got := barely.Percent(); got != 1 {
		t.Errorf("Percent() = %d, want 1: one task is done", got)
	}
}

func TestProgress_isNothingWhenNothingIsDone(t *testing.T) {
	started := Progress{Tasks: Tally{Total: 3}}
	if got := started.Percent(); got != 0 {
		t.Errorf("Percent() = %d, want 0: nothing is done", got)
	}
}

func TestProgress_roundsToTheNearestWhole(t *testing.T) {
	for _, c := range []struct {
		done, total, want int
	}{
		{1, 3, 33},
		{2, 3, 67},
		{1, 8, 13},
		{1, 2, 50},
		{5, 6, 83},
	} {
		p := Progress{Tasks: Tally{Total: c.total, Done: c.done}}
		if got := p.Percent(); got != c.want {
			t.Errorf("%d of %d: Percent() = %d, want %d", c.done, c.total, got, c.want)
		}
	}
}

func TestTally_addsUpAcrossBoth(t *testing.T) {
	p := Progress{Tasks: Tally{Total: 3, Done: 1}, Steps: Tally{Total: 4, Done: 2}}
	if got := p.Total(); got != 7 {
		t.Errorf("Total() = %d, want 7", got)
	}
	if got := p.Done(); got != 3 {
		t.Errorf("Done() = %d, want 3", got)
	}
}

func TestOfProgress_isNilWhenNothingIsUnder(t *testing.T) {
	if got := OfProgress(Progress{}); got != nil {
		t.Errorf("OfProgress(Progress{}) = %+v, want nil", got)
	}
}

func TestOfProgress_omitsATallyWithNothingInIt(t *testing.T) {
	got := OfProgress(Progress{Tasks: Tally{Total: 2, Done: 2}})
	if got == nil {
		t.Fatalf("OfProgress = nil, want a shape")
	}
	if got.Steps != nil {
		t.Errorf("Steps = %+v, want nil: no step is under it", got.Steps)
	}
	if got.Tasks == nil || *got.Tasks != (Tally{Total: 2, Done: 2}) {
		t.Errorf("Tasks = %+v, want {2 2}", got.Tasks)
	}
}

func TestOfProgress_carriesTheTallyAndTheShare(t *testing.T) {
	got := OfProgress(Progress{Steps: Tally{Total: 4, Done: 1}})
	if got == nil {
		t.Fatalf("OfProgress = nil, want a shape")
	}
	if got.Tasks != nil {
		t.Errorf("Tasks = %+v, want nil: no task is under it", got.Tasks)
	}
	if got.Steps == nil || *got.Steps != (Tally{Total: 4, Done: 1}) {
		t.Errorf("Steps = %+v, want {4 1}", got.Steps)
	}
	if got.Percent != 25 {
		t.Errorf("Percent = %d, want 25", got.Percent)
	}
}
