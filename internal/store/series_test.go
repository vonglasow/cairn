package store

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestSeries_areKeptChangedAndDeleted(t *testing.T) {
	s := openStore(t)
	if every, err := s.EverySeries(); err != nil || len(every) != 0 {
		t.Fatalf("a new store holds the series %v, %v", every, err)
	}
	made, err := s.KeepSeries(Series{
		Rule: "every weekday 10:30", Start: "2026-11-09", Title: "Stand-up", Priority: "normal",
		Estimate: 15 * time.Minute, Steps: []string{"Yesterday", "Today"}, Labels: []string{"work", "meeting"}, Objectives: []int{4, 2},
		Arrives: "current",
	})
	if err != nil {
		t.Fatal(err)
	}
	if made.Number != 1 {
		t.Fatalf("the first series is numbered %d", made.Number)
	}
	found, err := s.FindSeries(1)
	if err != nil {
		t.Fatal(err)
	}
	made.Labels, made.Objectives = []string{"meeting", "work"}, []int{2, 4}
	if !reflect.DeepEqual(found, made) {
		t.Errorf("found\n%+v\nwant\n%+v", found, made)
	}
	found.Stopped, found.Title, found.Labels, found.Objectives, found.Steps, found.Arrives = "2026-11-10", "Daily", []string{"team"}, nil, nil, "backlog"
	if err := s.UpdateSeries(found); err != nil {
		t.Fatal(err)
	}
	again, err := s.FindSeries(1)
	if err != nil {
		t.Fatal(err)
	}
	found.Objectives, found.Steps = []int{}, []string{}
	if !reflect.DeepEqual(again, found) {
		t.Errorf("changed, found\n%+v\nwant\n%+v", again, found)
	}
	if err := s.DeleteSeries(1); err != nil {
		t.Fatal(err)
	}
	if _, err := s.FindSeries(1); !errors.Is(err, ErrNoSuchSeries) {
		t.Errorf("a deleted series is found: %v", err)
	}
	if labels, err := rowsOf(s, one[string], `SELECT label FROM series_labels`); err != nil || len(labels) != 0 {
		t.Errorf("what a deleted series carried is still kept: %v, %v", labels, err)
	}
	if next, err := s.KeepSeries(Series{Rule: "every day", Start: "2026-11-09", Title: "Again", Priority: "low"}); err != nil || next.Number != 2 {
		t.Errorf("a series made after one deleted is numbered %d, %v", next.Number, err)
	}
}

func TestSeries_actsAreKeptAndTellWhichTaskBelongsWhere(t *testing.T) {
	s := openStore(t)
	for _, title := range []string{"one", "two"} {
		if _, err := s.KeepSeries(Series{Rule: "every day", Start: "2026-11-09", Title: title, Priority: "normal"}); err != nil {
			t.Fatal(err)
		}
	}
	at := time.Date(2026, 11, 9, 10, 0, 0, 0, time.UTC)
	for _, a := range []SeriesAct{
		{Series: 1, Did: "made", Task: 7, Day: "2026-11-09", Through: "cli", By: "me", At: at},
		{Series: 2, Did: "made", Task: 8, Day: "2026-11-09", At: at},
		{Series: 1, Did: "missed", Day: "2026-11-10", At: at},
		{Series: 1, Did: "changed", Field: "title", Was: "one", Became: "One", At: at},
	} {
		if err := s.RecordSeries(a); err != nil {
			t.Fatal(err)
		}
	}
	acts, err := s.SeriesActs(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(acts) != 3 || acts[0].Task != 7 || acts[0].By != "me" || !acts[0].At.Equal(at) || acts[2].Became != "One" {
		t.Errorf("the acts of series 1 read %+v", acts)
	}
	of, err := s.Occurrences()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(of, map[int]int{7: 1, 8: 2}) {
		t.Errorf("the occurrences read %v", of)
	}
	if err := s.DeleteSeries(2); err != nil {
		t.Fatal(err)
	}
	if of, _ := s.Occurrences(); !reflect.DeepEqual(of, map[int]int{7: 1}) {
		t.Errorf("a deleted series still has occurrences: %v", of)
	}
	if acts, _ := s.SeriesActs(2); len(acts) != 1 {
		t.Errorf("a deleted series lost its history: %v", acts)
	}
}

func TestSeries_followALabelAndLetGoOfAnObjective(t *testing.T) {
	s := openStore(t)
	if _, err := s.KeepSeries(Series{Rule: "every day", Start: "2026-11-09", Title: "one", Priority: "normal", Labels: []string{"a", "b"}, Objectives: []int{1}}); err != nil {
		t.Fatal(err)
	}
	if err := s.RenameKeptLabel("a", "c"); err != nil {
		t.Fatal(err)
	}
	if err := s.ForgetLabel("b"); err != nil {
		t.Fatal(err)
	}
	if err := s.letSeriesGoOf(1); err != nil {
		t.Fatal(err)
	}
	found, err := s.FindSeries(1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(found.Labels, []string{"c"}) || len(found.Objectives) != 0 {
		t.Errorf("the template carries %v and serves %v", found.Labels, found.Objectives)
	}
}

func TestSeries_sayWhenTheStoreCannotKeepThem(t *testing.T) {
	broken := func(t *testing.T, drop string) *Store {
		t.Helper()
		s := openStore(t)
		if _, err := s.KeepSeries(Series{Rule: "every day", Start: "2026-11-09", Title: "one", Priority: "normal", Labels: []string{"a"}, Objectives: []int{1}}); err != nil {
			t.Fatal(err)
		}
		if _, err := s.on.Exec(drop); err != nil {
			t.Fatal(err)
		}
		return s
	}
	one := Series{Number: 1, Rule: "every day", Start: "2026-11-09", Title: "one", Priority: "normal", Labels: []string{"a"}, Objectives: []int{1}}
	for _, c := range []struct {
		drop string
		does func(*Store) error
	}{
		{`DROP TABLE series_labels`, func(s *Store) error { _, err := s.EverySeries(); return err }},
		{`DROP TABLE series_labels`, func(s *Store) error { _, err := s.FindSeries(1); return err }},
		{`DROP TABLE series_objectives`, func(s *Store) error { _, err := s.FindSeries(1); return err }},
		{`DROP TABLE series_labels`, func(s *Store) error { return s.UpdateSeries(one) }},
		{`DROP TABLE series_objectives`, func(s *Store) error { return s.UpdateSeries(one) }},
		{`DROP TABLE series_labels`, func(s *Store) error { return s.DeleteSeries(1) }},
		{`DROP TABLE series_objectives`, func(s *Store) error { return s.letSeriesGoOf(1) }},
		{`DROP TABLE series`, func(s *Store) error { _, err := s.EverySeries(); return err }},
		{`DROP TABLE series`, func(s *Store) error { _, err := s.FindSeries(1); return err }},
		{`DROP TABLE series`, func(s *Store) error { _, err := s.KeepSeries(one); return err }},
		{`DROP TABLE series`, func(s *Store) error { return s.UpdateSeries(one) }},
		{`DROP TABLE series`, func(s *Store) error { return s.DeleteSeries(1) }},
		{`DROP TABLE series_labels`, func(s *Store) error { _, err := s.KeepSeries(one); return err }},
		{`DROP TABLE series_acts`, func(s *Store) error { return s.RecordSeries(SeriesAct{Series: 1, Did: "made"}) }},
		{`DROP TABLE series_acts`, func(s *Store) error { _, err := s.SeriesActs(1); return err }},
		{`DROP TABLE series_acts`, func(s *Store) error { _, err := s.Occurrences(); return err }},
		{`INSERT INTO series_acts (series_number, did, at) VALUES (1, 'made', 'not a moment')`, func(s *Store) error { _, err := s.SeriesActs(1); return err }},
		{`INSERT INTO series_acts (series_number, did, task_number, at) VALUES (1, 'made', 'seven', '2026-11-09T10:00:00Z')`, func(s *Store) error { _, err := s.Occurrences(); return err }},
	} {
		if err := c.does(broken(t, c.drop)); err == nil {
			t.Errorf("after %s, it went ahead", c.drop)
		}
	}
}
