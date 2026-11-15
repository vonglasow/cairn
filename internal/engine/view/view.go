// Package view holds a view: a set of filters kept under a name, its place, what it keeps.
package view

import (
	"cairn/internal/engine/item"
	"cairn/internal/engine/task"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// View is a view: its name, its place, and the filters it holds — an objective by its number, a label by its name.
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
	Unblocked    bool
}

// Objective is an objective a view holds, with the store's figures for it: the same in a view and out of one.
type Objective struct {
	Number    int                 `json:"number"`
	Title     string              `json:"title"`
	Progress  *item.ProgressShape `json:"progress" jsonschema:"how far along it is over its whole subtree, each task once; null when nothing serves it"`
	TimeSpent string              `json:"time_spent" jsonschema:"how long has been spent on the tasks serving it, as 1h30m; empty when nothing has"`
}

// Deleted is how deleting a view answers: the name of the view deleted.
type Deleted struct {
	Deleted string `json:"deleted"`
}

// Outside is a task outside a view that touches one it keeps, and every way it does.
type Outside struct {
	Number int     `json:"number"`
	Title  string  `json:"title"`
	State  string  `json:"state"`
	Why    []Touch `json:"why"`
}

// Touch is one way a task outside a view touches one inside it: it blocks it, waits on it, or is linked to it.
type Touch struct {
	Is string `json:"is" jsonschema:"blocks, waits on or linked to"`
	Of int    `json:"of" jsonschema:"the number of the task inside the view"`
}

// NarrowsAnything says whether a view holds any filter.
func NarrowsAnything(v View) bool {
	return v.Focus != "" || v.Unblocked || v.State != "" || v.Priority != "" || v.DueBefore != "" || v.Archived || v.Search != "" || len(v.Objectives) > 0 || len(v.Labels) > 0
}

// Seen is a view as a listing reads it: its name, the filters it lost, whether it keeps nothing, its search words.
type Seen struct {
	Name    string
	Lost    []string
	Nothing bool
	Words   []string
}

// Kept is the tasks the view keeps.
func (s Seen) Kept(tasks []task.Task) []task.Task {
	if s.Nothing {
		return []task.Task{}
	}
	return slices.DeleteFunc(tasks, func(t task.Task) bool { return len(s.Words) > 0 && !task.Matching(t, s.Words) })
}

// Touching is each way a task outside a view touches one inside it.
var Touching = []string{"blocks", "waits on", "linked to"}

// Touches says whether a task outside a view touches one inside it in the way named.
func Touches(is string, outside, inside task.Carried) bool {
	switch is {
	case "blocks":
		return slices.Contains(inside.Blockers, outside.Number)
	case "waits on":
		return slices.Contains(outside.Blockers, inside.Number)
	}
	return slices.Contains(outside.Linked, inside.Number) || slices.Contains(inside.Linked, outside.Number)
}

// Naming is the views that name one of the objectives deleted.
func Naming(views []View, goners []int) []string {
	var naming []string
	for _, v := range views {
		if slices.ContainsFunc(v.Objectives, func(n int) bool { return slices.Contains(goners, n) }) {
			naming = append(naming, v.Name)
		}
	}
	return naming
}

// CmdKey is the key that opens the view at a place, ⌘1 to ⌘9, or its place past them.
func CmdKey(place int) string {
	if place <= 9 {
		return "⌘" + strconv.Itoa(place)
	}
	return "place " + strconv.Itoa(place)
}

// StoreFigures is, for a person, how far the store says an objective of a view stands.
func StoreFigures(o Objective) string {
	if o.Progress == nil {
		return "nothing serves it yet"
	}
	var said []string
	if t := o.Progress.Tasks; t != nil {
		said = append(said, fmt.Sprintf("%d/%d", t.Done, t.Total))
	}
	if s := o.Progress.Steps; s != nil {
		said = append(said, fmt.Sprintf("%d/%d steps", s.Done, s.Total))
	}
	said = append(said, fmt.Sprintf("%d%%", o.Progress.Percent))
	if o.TimeSpent != "" {
		said = append(said, o.TimeSpent)
	}
	return strings.Join(said, " · ") + " (the store's)"
}

// TypedName is a view name as it is typed on a command line.
func TypedName(name string) string {
	if strings.ContainsAny(name, " \t'\"") {
		return strconv.Quote(name)
	}
	return name
}

// OutsideText is, for a person, the tasks outside a view that touch one inside it.
func OutsideText(outside []Outside) string {
	if len(outside) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString("\noutside this view, never counted:\n")
	for _, o := range outside {
		said := []string{fmt.Sprintf("#%d  %s ", o.Number, o.Title), o.State}
		for _, w := range o.Why {
			said = append(said, fmt.Sprintf("%s #%d", w.Is, w.Of))
		}
		out.WriteString(strings.Join(said, " · ") + "\n")
	}
	return out.String()
}
