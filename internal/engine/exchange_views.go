package engine

import (
	"fmt"
	"slices"
	"strings"

	"cairn/internal/domain"
	"cairn/internal/store"
)

type viewFiled struct {
	Name string `json:"name"`
	Holds
}

func (e *Engine) viewsWritten(t tree) ([]viewFiled, error) {
	views, err := e.store.Views()
	if err != nil {
		return nil, err
	}
	var out []viewFiled
	for _, v := range views {
		v.Objectives = slices.DeleteFunc(slices.Clone(v.Objectives), func(n int) bool { _, there := t.by[n]; return !there })
		if !narrowsAnything(v) {
			continue
		}
		out = append(out, viewFiled{Name: v.Name, Holds: Holds{State: v.State, Priority: v.Priority, Label: v.Labels, Objective: v.Objectives,
			Itself: v.Itself, DueBefore: v.DueBefore, Focus: v.Focus, Archived: v.Archived, WithArchived: v.WithArchived, Search: v.Search}})
	}
	return out, nil
}

func checkedViews(read []viewFiled, objectives map[int]bool) error {
	seen := map[string]bool{}
	for _, v := range read {
		if seen[domain.Folded(v.Name)] {
			return fmt.Errorf("%w: view %q is in the file twice", ErrNotAnExport, v.Name)
		}
		seen[domain.Folded(v.Name)] = true
		for _, o := range v.Objective {
			if !objectives[o] {
				return fmt.Errorf("%w: view %s names objective %d — objective %d is not in the file", ErrExportPointsOutside, v.Name, o, o)
			}
		}
		if _, err := viewOfFile(v); err != nil {
			return fmt.Errorf("view %s: %w", v.Name, err)
		}
	}
	return nil
}

func viewOfFile(v viewFiled) (store.View, error) {
	name, err := domain.Title("a view", v.Name)
	if err != nil {
		return store.View{}, err
	}
	in := Input{Given: map[string]string{"name": name}, Listed: map[string][]string{}}
	for argument, said := range map[string]string{"state": v.State, "priority": v.Priority, "deadline-before": v.DueBefore} {
		if said != "" {
			in.Given[argument] = said
		}
	}
	if err := ViewNew.Check(in); err != nil {
		return store.View{}, err
	}
	if v.Focus != "" && v.Focus != "in" && v.Focus != "out" {
		return store.View{}, fmt.Errorf("%w: view %s keeps the focus %q — in or out", ErrNotAnExport, v.Name, v.Focus)
	}
	for _, argument := range viewFilters {
		if err := accepted(in, argument); err != nil {
			return store.View{}, err
		}
	}
	labels, err := domain.ParseLabels(v.Label)
	if err != nil {
		return store.View{}, err
	}
	kept := store.View{Name: name, State: v.State, Priority: v.Priority, DueBefore: v.DueBefore, Focus: v.Focus, Search: strings.Join(strings.Fields(v.Search), " "), Labels: labels,
		Objectives: slices.Clone(v.Objective), Itself: v.Itself, Archived: v.Archived, WithArchived: v.WithArchived}
	if !narrowsAnything(kept) {
		return store.View{}, Refusing("view-holds-nothing", "a view holds at least one filter or a search")
	}
	return kept, nil
}

func (e *Engine) viewsFiled(read []viewFiled, objectives map[int]int) error {
	if len(read) == 0 {
		return nil
	}
	views, err := e.store.Views()
	if err != nil {
		return err
	}
	for _, v := range read {
		if slices.ContainsFunc(views, func(here store.View) bool { return domain.Folded(here.Name) == domain.Folded(v.Name) }) {
			continue
		}
		kept, _ := viewOfFile(v)
		for i, n := range kept.Objectives {
			kept.Objectives[i] = objectives[n]
		}
		views = append(views, kept)
	}
	return e.store.KeepViews(views)
}
