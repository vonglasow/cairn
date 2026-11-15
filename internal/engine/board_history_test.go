package engine

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/item"
	"cairn/internal/engine/series"
	"strings"
	"testing"
	"time"

	"cairn/internal/engine/act"
)

func anAct(did, field, was, became, through string, at time.Time) act.Act {
	return act.Made(act.Carried{Task: 1, Did: did, Field: field, Was: was, Became: became, Through: through, At: at})
}

func TestSaid_isOneLineOfAHistoryTheWayAPersonReadsIt(t *testing.T) {
	at := time.Date(2026, 9, 29, 9, 30, 0, 0, time.Local)
	changed := SaidOf(anAct("changed", "priority", "normal", "high", "mcp", at))
	if want := calendar.FormatMoment(at) + "  changed  priority  normal → high  (through mcp)"; changed.Line() != want {
		t.Errorf("Line = %q, want %q", changed.Line(), want)
	}
	if changed.Was != "normal" || changed.Became != "high" || changed.Through != "mcp" {
		t.Errorf("said %+v", changed)
	}
	emptied := SaidOf(anAct("changed", "deadline", "2026-10-01", "", "", at))
	if line := emptied.Line(); !strings.HasSuffix(line, "deadline  2026-10-01 → "+item.Absent) {
		t.Errorf("an emptied field reads %q", line)
	}
	created := SaidOf(anAct("created", "", "", "", "", at))
	if want := calendar.FormatMoment(at) + "  created"; created.Line() != want {
		t.Errorf("Line = %q, want %q", created.Line(), want)
	}
	made := SaidOf(anAct(series.MadeFromSeries, "", "", "↻3", "", at))
	if want := calendar.FormatMoment(at) + "  made from series  ↻3"; made.Line() != want {
		t.Errorf("Line = %q, want %q", made.Line(), want)
	}
}
