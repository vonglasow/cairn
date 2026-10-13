package engine

import (
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/trace"
)

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
