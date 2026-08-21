package domain

import "testing"

func TestProgress_countsWhatIsThere(t *testing.T) {
	p := Progress{Tasks: Tally{Total: 3, Done: 1}, SubTasks: Tally{Total: 4, Done: 2}}
	if !p.Any() {
		t.Error("Any() = false, want true: seven things are under it")
	}
	// 3 of 7, which is 42.857: the share is over everything, not over the
	// tasks alone, and it rounds to the nearest whole.
	if got := p.Percent(); got != 43 {
		t.Errorf("Percent() = %d, want 43", got)
	}
}

func TestProgress_ofATaskCountsItsSubTasksOnly(t *testing.T) {
	p := Progress{SubTasks: Tally{Total: 2, Done: 1}}
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

// 299 of 300 is 99.67, which rounds to 100 — and 100 reads as "finished" to
// everybody. The last thing left undone is exactly the one worth seeing, so
// the whole share is kept for the case where nothing is left.
func TestProgress_neverRoundsUpToWhole(t *testing.T) {
	almost := Progress{Tasks: Tally{Total: 300, Done: 299}}
	if got := almost.Percent(); got != 99 {
		t.Errorf("Percent() = %d, want 99: one task is still not done", got)
	}
}

// The other end, for the same reason: 1 of 300 is 0.33, and reporting 0 hides
// the fact that somebody has started.
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
		{1, 3, 33}, // 33.33 down
		{2, 3, 67}, // 66.67 up
		{1, 8, 13}, // 12.5 up, away from zero
		{1, 2, 50}, // exact
		{5, 6, 83}, // 83.33 down
	} {
		p := Progress{Tasks: Tally{Total: c.total, Done: c.done}}
		if got := p.Percent(); got != c.want {
			t.Errorf("%d of %d: Percent() = %d, want %d", c.done, c.total, got, c.want)
		}
	}
}

func TestTally_addsUpAcrossBoth(t *testing.T) {
	p := Progress{Tasks: Tally{Total: 3, Done: 1}, SubTasks: Tally{Total: 4, Done: 2}}
	if got := p.Total(); got != 7 {
		t.Errorf("Total() = %d, want 7", got)
	}
	if got := p.Done(); got != 3 {
		t.Errorf("Done() = %d, want 3", got)
	}
}
