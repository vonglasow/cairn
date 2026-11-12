package store

import (
	"fmt"
)

// View is a view as the store keeps it: its name, its place, and the filters it holds — an objective by its number, a label by its name.
type View struct {
	Name         string
	Place        int
	State        string
	Priority     string
	DueBefore    string
	Focus        string
	Search       string
	Objectives   []int
	Labels       []string
	Itself       bool
	Archived     bool
	WithArchived bool
}

const (
	everyView     = `SELECT id, name, place, state, priority, due_before, focus, search, itself, archived, with_archived FROM views ORDER BY place, id`
	keepView      = `INSERT INTO views (name, place, state, priority, due_before, focus, search, itself, archived, with_archived) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	viewObjective = `INSERT OR IGNORE INTO view_objectives (view_id, objective_number) VALUES (?, ?)`
	viewLabel     = `INSERT OR IGNORE INTO view_labels (view_id, label) VALUES (?, ?)`
)

// Views is every view, in its place.
func (s *Store) Views() ([]View, error) {
	ids, views, err := s.viewRows()
	if err != nil {
		return nil, fmt.Errorf("reading the views: %w", err)
	}
	for i, id := range ids {
		if views[i].Objectives, err = rowsOf(s, one[int], `SELECT objective_number FROM view_objectives WHERE view_id = ? ORDER BY objective_number`, id); err != nil {
			return nil, fmt.Errorf("reading what a view holds: %w", err)
		}
		if views[i].Labels, err = rowsOf(s, one[string], `SELECT label FROM view_labels WHERE view_id = ? ORDER BY label`, id); err != nil {
			return nil, fmt.Errorf("reading what a view holds: %w", err)
		}
	}
	return views, nil
}

func (s *Store) viewRows() ([]int64, []View, error) {
	rows, err := s.on.Query(everyView)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	var ids []int64
	views := []View{}
	for rows.Next() {
		var id int64
		var v View
		if err := rows.Scan(&id, &v.Name, &v.Place, &v.State, &v.Priority, &v.DueBefore, &v.Focus, &v.Search, &v.Itself, &v.Archived, &v.WithArchived); err != nil {
			return nil, nil, err
		}
		ids, views = append(ids, id), append(views, v)
	}
	return ids, views, rows.Err()
}

// KeepViews keeps these views in place of every one there was, each in the place its order gives it.
func (s *Store) KeepViews(views []View) error {
	for _, clear := range []string{`DELETE FROM view_labels`, `DELETE FROM view_objectives`, `DELETE FROM views`} {
		if _, err := s.on.Exec(clear); err != nil {
			return fmt.Errorf("keeping the views: %w", err)
		}
	}
	for i, v := range views {
		res, err := s.on.Exec(keepView, v.Name, i+1, v.State, v.Priority, v.DueBefore, v.Focus, v.Search, v.Itself, v.Archived, v.WithArchived)
		if err != nil {
			return fmt.Errorf("keeping the view %q: %w", v.Name, err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("keeping the view %q: %w", v.Name, err)
		}
		for _, n := range v.Objectives {
			if _, err := s.on.Exec(viewObjective, id, n); err != nil {
				return fmt.Errorf("keeping the view %q: %w", v.Name, err)
			}
		}
		for _, label := range v.Labels {
			if _, err := s.on.Exec(viewLabel, id, label); err != nil {
				return fmt.Errorf("keeping the view %q: %w", v.Name, err)
			}
		}
	}
	return nil
}
