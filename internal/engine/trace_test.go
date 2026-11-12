package engine

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"cairn/internal/domain"
)

func addedAt(t *testing.T, path, title string, at time.Time) {
	t.Helper()
	now := at.Format(time.RFC3339)
	h := Home{Dir: t.TempDir(), Lookup: func(k string) (string, bool) {
		switch k {
		case "CAIRN_STORE":
			return path, true
		case NowVariable:
			return now, true
		}
		return "", false
	}}
	if _, err := DoFrom(TaskNew, h, given("title", title), Through("cli", h)); err != nil {
		t.Fatal(err)
	}
}

func tasksActedOn(answer Answer) []int {
	var numbers []int
	for _, act := range answer.Since {
		if n := act.Carried().Task; !slices.Contains(numbers, n) {
			numbers = append(numbers, n)
		}
	}
	return numbers
}

func TestTraceList_untilADayKeepsTheWholeOfItAndNothingAfter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	addedAt(t, path, "Monday's", time.Date(2026, 9, 28, 23, 59, 0, 0, time.Local))
	addedAt(t, path, "Tuesday's first", time.Date(2026, 9, 29, 0, 0, 0, 0, time.Local))
	addedAt(t, path, "Tuesday's last", time.Date(2026, 9, 29, 23, 59, 0, 0, time.Local))
	addedAt(t, path, "Wednesday's", time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local))
	listed, err := Do(TraceList, path, Input{Given: map[string]string{"since": "2026-09-29", "until": "2026-09-29"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := tasksActedOn(listed); !slices.Equal(got, []int{2, 3}) {
		t.Errorf("tasks acted on = %v, want Tuesday's two, from its first minute to its last, in the day's own time", got)
	}
}

func TestTraceList_untilSomethingThatIsNotADayIsRefused(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	addedAt(t, path, "Monday's", time.Date(2026, 9, 28, 9, 0, 0, 0, time.Local))
	_, err := Do(TraceList, path, given("until", "pasunjour"))
	if !errors.Is(err, domain.ErrNotADay) {
		t.Errorf("err = %v, want it refused as not a day", err)
	}
}
