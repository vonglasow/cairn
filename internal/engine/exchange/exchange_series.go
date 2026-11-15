package exchange

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/series"
	"cairn/internal/store/sqlite"
	"fmt"
	"slices"
	"time"
)

// SeriesFiled is a series as an export files it.
type SeriesFiled struct {
	Number   int             `json:"number"`
	Rule     string          `json:"rule"`
	From     string          `json:"from"`
	Stopped  string          `json:"stopped,omitempty"`
	Template series.Template `json:"template"`
	Acts     []series.Act    `json:"acts,omitempty"`
}

var seriesDid = []string{"made", "missed", "skipped", "changed", "stopped", "resumed", "deleted"}

// SeriesFromToday reads the days of the series an export files, relative ones from today.
func SeriesFromToday(read []SeriesFiled, today time.Time) ([]SeriesFiled, error) {
	for i, s := range read {
		for _, day := range []*string{&read[i].From, &read[i].Stopped} {
			counted, err := calendar.FromToday(*day, today)
			if err != nil {
				return nil, fmt.Errorf("series %d: %w", s.Number, err)
			}
			*day = counted
		}
		for j, a := range s.Acts {
			counted, err := calendar.FromToday(a.Day, today)
			if err != nil {
				return nil, fmt.Errorf("series %d: %w", s.Number, err)
			}
			read[i].Acts[j].Day = counted
		}
	}
	return read, nil
}

// ActFiled is an act of a series as an import keeps it, refused when it is nothing a series does.
func ActFiled(a series.Act) (sqlite.SeriesAct, error) {
	if !slices.Contains(seriesDid, a.Did) {
		return sqlite.SeriesAct{}, fmt.Errorf("%w: %q is nothing a series does — %v", ErrNotAnExport, a.Did, seriesDid)
	}
	if _, err := calendar.ParseDeadline(a.Day); err != nil {
		return sqlite.SeriesAct{}, err
	}
	at := time.Now()
	if a.At != "" {
		read, err := time.Parse(time.RFC3339, a.At)
		if err != nil {
			return sqlite.SeriesAct{}, fmt.Errorf("%w: the moment %q is written %s", ErrNotAnExport, a.At, time.RFC3339)
		}
		at = read
	}
	return sqlite.SeriesAct{Did: a.Did, Task: a.Task, Day: a.Day, Field: a.Field, Was: a.Was, Became: a.Became,
		Through: a.Through, By: a.By, At: at}, nil
}
