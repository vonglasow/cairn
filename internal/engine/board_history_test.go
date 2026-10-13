package engine

import (
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/trace"
)

// act builds one act of a history.
func act(did, field, was, became, through string, at time.Time) trace.Act {
	var a trace.AsTold
	a.On(1)
	a.Did(did)
	domain.TellOrNot(field, a.Changing, a.NothingNamed)
	domain.TellOrNot(was, a.Was, a.WasNothing)
	domain.TellOrNot(became, a.Became, a.BecameNothing)
	domain.TellOrNot(through, a.Through, a.ByHand)
	a.At(at)
	return a.Act()
}

// A line of a history says when, what was done, and — when it was about
// something in particular — what it went from and to, and which surface did it
// when it was not the person reading. Both boards say it the same way: the
// window used to write the moment the way a machine reads one, and leave out
// what a value was.
func TestSaid_isOneLineOfAHistoryTheWayAPersonReadsIt(t *testing.T) {
	at := time.Date(2026, 9, 29, 9, 30, 0, 0, time.Local)
	changed := SaidOf(act("changed", "priority", "normal", "high", "mcp", at))
	if want := domain.FormatMoment(at) + "  changed  priority  normal → high  (mcp)"; changed.Line() != want {
		t.Errorf("Line = %q, want %q", changed.Line(), want)
	}
	if changed.Was != "normal" || changed.Became != "high" || changed.Through != "mcp" {
		t.Errorf("said %+v", changed)
	}
	emptied := SaidOf(act("changed", "deadline", "2026-10-01", "", "", at))
	if line := emptied.Line(); !strings.HasSuffix(line, "deadline  2026-10-01 → "+domain.Absent) {
		t.Errorf("an emptied field reads %q", line)
	}
	created := SaidOf(act("created", "", "", "", "", at))
	if want := domain.FormatMoment(at) + "  created"; created.Line() != want {
		t.Errorf("Line = %q, want %q", created.Line(), want)
	}
}
