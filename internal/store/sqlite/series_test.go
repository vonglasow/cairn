package sqlite

import (
	"testing"
)

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
