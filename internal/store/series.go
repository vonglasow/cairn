package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrNoSuchSeries is returned when a number refers to no series.
var ErrNoSuchSeries = errors.New("there is no such series")

// Series is a series as the store keeps it: its rule, the day it starts, the day it was stopped, and what its next occurrence is made from.
type Series struct {
	Number     int
	Rule       string
	Start      string
	Stopped    string
	Title      string
	Priority   string
	Estimate   time.Duration
	Steps      []string
	Labels     []string
	Objectives []int
}

// SeriesAct is one thing that happened to a series: an occurrence made, missed or skipped, a change, a stop, a resumption.
type SeriesAct struct {
	Series             int
	Did                string
	Task               int
	Day                string
	Field, Was, Became string
	Through, By        string
	At                 time.Time
}

const (
	seriesColumns = `number, rule, start, stopped, title, priority, estimate, steps`
	seriesActs    = `series_number, did, task_number, day, field, was, became, through, author, at`
)

// EverySeries is every series, by its number.
func (s *Store) EverySeries() ([]Series, error) {
	every, err := rowsOf(s, scanSeries, `SELECT `+seriesColumns+` FROM series ORDER BY number`)
	if err != nil {
		return nil, fmt.Errorf("reading the series: %w", err)
	}
	for i := range every {
		if err := s.templateOf(&every[i]); err != nil {
			return nil, err
		}
	}
	return every, nil
}

// FindSeries hands back the series carrying number, and says so when there is none.
func (s *Store) FindSeries(number int) (Series, error) {
	found, err := scanSeries(s.on.QueryRow(`SELECT `+seriesColumns+` FROM series WHERE number = ?`, number))
	if errors.Is(err, sql.ErrNoRows) {
		return Series{}, fmt.Errorf("series %d: %w", number, ErrNoSuchSeries)
	}
	if err != nil {
		return Series{}, fmt.Errorf("reading series %d: %w", number, err)
	}
	return found, s.templateOf(&found)
}

func scanSeries(row scanner) (Series, error) {
	var one Series
	var estimate int64
	var steps string
	if err := row.Scan(&one.Number, &one.Rule, &one.Start, &one.Stopped, &one.Title, &one.Priority, &estimate, &steps); err != nil {
		return Series{}, err
	}
	one.Estimate = time.Duration(estimate)
	one.Steps = []string{}
	if steps != "" {
		one.Steps = strings.Split(steps, "\n")
	}
	return one, nil
}

func (s *Store) templateOf(of *Series) error {
	var err error
	if of.Labels, err = rowsOf(s, one[string], `SELECT label FROM series_labels WHERE series_number = ? ORDER BY label`, of.Number); err != nil {
		return fmt.Errorf("reading what series %d carries: %w", of.Number, err)
	}
	if of.Objectives, err = rowsOf(s, one[int], `SELECT objective_number FROM series_objectives WHERE series_number = ? ORDER BY objective_number`, of.Number); err != nil {
		return fmt.Errorf("reading what series %d serves: %w", of.Number, err)
	}
	return nil
}

// KeepSeries records a new series and hands it back carrying the number it was given.
func (s *Store) KeepSeries(one Series) (Series, error) {
	res, err := s.on.Exec(`INSERT INTO series (rule, start, stopped, title, priority, estimate, steps) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		one.Rule, one.Start, one.Stopped, one.Title, one.Priority, int64(one.Estimate), strings.Join(one.Steps, "\n"))
	if err != nil {
		return Series{}, fmt.Errorf("keeping the series: %w", err)
	}
	number, err := res.LastInsertId()
	if err != nil {
		return Series{}, fmt.Errorf("numbering the series: %w", err)
	}
	one.Number = int(number)
	return one, s.keepTemplate(one)
}

// UpdateSeries replaces what the store holds for a series.
func (s *Store) UpdateSeries(one Series) error {
	if _, err := s.on.Exec(`UPDATE series SET rule = ?, start = ?, stopped = ?, title = ?, priority = ?, estimate = ?, steps = ? WHERE number = ?`,
		one.Rule, one.Start, one.Stopped, one.Title, one.Priority, int64(one.Estimate), strings.Join(one.Steps, "\n"), one.Number); err != nil {
		return fmt.Errorf("changing series %d: %w", one.Number, err)
	}
	return s.keepTemplate(one)
}

func (s *Store) keepTemplate(one Series) error {
	if err := s.clearTemplate(one.Number); err != nil {
		return err
	}
	for _, label := range one.Labels {
		if _, err := s.on.Exec(`INSERT OR IGNORE INTO series_labels (series_number, label) VALUES (?, ?)`, one.Number, label); err != nil {
			return fmt.Errorf("keeping what series %d carries: %w", one.Number, err)
		}
	}
	for _, n := range one.Objectives {
		if _, err := s.on.Exec(`INSERT OR IGNORE INTO series_objectives (series_number, objective_number) VALUES (?, ?)`, one.Number, n); err != nil {
			return fmt.Errorf("keeping what series %d serves: %w", one.Number, err)
		}
	}
	return nil
}

func (s *Store) clearTemplate(number int) error {
	for _, clear := range []string{`DELETE FROM series_labels WHERE series_number = ?`, `DELETE FROM series_objectives WHERE series_number = ?`} {
		if _, err := s.on.Exec(clear, number); err != nil {
			return fmt.Errorf("clearing what series %d carries: %w", number, err)
		}
	}
	return nil
}

// DeleteSeries removes a series and what its next occurrence was made from; its history stays.
func (s *Store) DeleteSeries(number int) error {
	if _, err := s.on.Exec(`DELETE FROM series WHERE number = ?`, number); err != nil {
		return fmt.Errorf("deleting series %d: %w", number, err)
	}
	return s.clearTemplate(number)
}

func (s *Store) letSeriesGoOf(objective int) error {
	if _, err := s.on.Exec(`DELETE FROM series_objectives WHERE objective_number = ?`, objective); err != nil {
		return fmt.Errorf("letting the series go of objective %d: %w", objective, err)
	}
	return nil
}

// RecordSeries writes down one thing that happened to a series.
func (s *Store) RecordSeries(a SeriesAct) error {
	if _, err := s.on.Exec(`INSERT INTO series_acts (`+seriesActs+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.Series, a.Did, a.Task, a.Day, a.Field, a.Was, a.Became, a.Through, a.By, a.At.UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("writing down what happened to series %d: %w", a.Series, err)
	}
	return nil
}

// SeriesActs hands back everything that happened to a series, oldest first, whether or not it is still there.
func (s *Store) SeriesActs(number int) ([]SeriesAct, error) {
	acts, err := rowsOf(s, scanSeriesAct, `SELECT `+seriesActs+` FROM series_acts WHERE series_number = ? ORDER BY id`, number)
	if err != nil {
		return nil, fmt.Errorf("reading what happened to series %d: %w", number, err)
	}
	return acts, nil
}

func scanSeriesAct(row scanner) (SeriesAct, error) {
	var a SeriesAct
	var at string
	if err := row.Scan(&a.Series, &a.Did, &a.Task, &a.Day, &a.Field, &a.Was, &a.Became, &a.Through, &a.By, &at); err != nil {
		return SeriesAct{}, err
	}
	when, err := time.Parse(time.RFC3339, at)
	if err != nil {
		return SeriesAct{}, fmt.Errorf("the moment %q cannot be read: %w", at, err)
	}
	a.At = when
	return a, nil
}

// Occurrences says, for every task a series still there made, which series made it.
func (s *Store) Occurrences() (map[int]int, error) {
	rows, err := s.on.Query(`SELECT task_number, series_number FROM series_acts
		WHERE did = 'made' AND series_number IN (SELECT number FROM series)`)
	if err != nil {
		return nil, fmt.Errorf("reading the occurrences: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := map[int]int{}
	for rows.Next() {
		var task, series int
		if err := rows.Scan(&task, &series); err != nil {
			return nil, err
		}
		out[task] = series
	}
	return out, rows.Err()
}
