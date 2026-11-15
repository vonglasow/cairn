package engine

import (
	"cairn/internal/engine/item"
	"fmt"
	"slices"
	"strings"

	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
)

// sets are the members of what a task or an objective carries that one change moved the same way.
type sets struct {
	Labels     []string `json:"labels,omitempty" jsonschema:"the labels, lower case"`
	Blockers   []int    `json:"blockers,omitempty" jsonschema:"the numbers of the tasks in its way"`
	Links      []int    `json:"links,omitempty" jsonschema:"the numbers of the tasks it is linked to"`
	Objectives []int    `json:"objectives,omitempty" jsonschema:"the numbers of the objectives it serves"`
}

// moves is what a change put on a task or an objective, and what it took off.
type moves struct {
	Added    sets
	TakenOff sets
}

func between[T comparable](was, now []T) (added, gone []T) {
	for _, one := range now {
		if !slices.Contains(was, one) {
			added = append(added, one)
		}
	}
	for _, one := range was {
		if !slices.Contains(now, one) {
			gone = append(gone, one)
		}
	}
	return added, gone
}

func movedOf(was, now task.Task) *moves {
	before, after := was.Carried(), now.Carried()
	var m moves
	m.Added.Labels, m.TakenOff.Labels = between(before.Labels, after.Labels)
	m.Added.Blockers, m.TakenOff.Blockers = between(before.Blockers, after.Blockers)
	m.Added.Links, m.TakenOff.Links = between(before.Linked, after.Linked)
	m.Added.Objectives, m.TakenOff.Objectives = between(before.Objectives, after.Objectives)
	return m.said()
}

func movedOfObjective(was, now objective.Objective) *moves {
	var m moves
	m.Added.Labels, m.TakenOff.Labels = between(was.Carried().Labels, now.Carried().Labels)
	return m.said()
}

func (m moves) said() *moves {
	if m.Added.empty() && m.TakenOff.empty() {
		return nil
	}
	return &m
}

func (s sets) empty() bool {
	return len(s.Labels) == 0 && len(s.Blockers) == 0 && len(s.Links) == 0 && len(s.Objectives) == 0
}

func (m *moves) added() *sets {
	if m == nil || m.Added.empty() {
		return nil
	}
	return &m.Added
}

func (m *moves) takenOff() *sets {
	if m == nil || m.TakenOff.empty() {
		return nil
	}
	return &m.TakenOff
}

func movedText(m *moves) string {
	if m == nil {
		return ""
	}
	var lines []string
	for _, of := range []struct {
		name        string
		added, gone []string
	}{
		{"labels", m.Added.Labels, m.TakenOff.Labels},
		{"blockers", refs(m.Added.Blockers), refs(m.TakenOff.Blockers)},
		{"links", refs(m.Added.Links), refs(m.TakenOff.Links)},
		{"objectives", refs(m.Added.Objectives), refs(m.TakenOff.Objectives)},
	} {
		lines = append(lines, movedLine(of.name+" added", of.added)...)
		lines = append(lines, movedLine(of.name+" taken off", of.gone)...)
	}
	return "\n" + strings.Join(lines, "")
}

func movedLine(said string, members []string) []string {
	if len(members) == 0 {
		return nil
	}
	return []string{fmt.Sprintf("%-16s  %s\n", said, strings.Join(members, ", "))}
}

func refs(numbers []int) []string {
	out := make([]string, 0, len(numbers))
	for _, n := range numbers {
		out = append(out, item.Ref(n))
	}
	return out
}
