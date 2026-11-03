package engine

import (
	"fmt"
	"slices"
	"strings"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
)

// Sets are the members of what a task or an objective carries that one change moved the same way.
type Sets struct {
	Labels     []string `json:"labels,omitempty" jsonschema:"the labels, lower case"`
	Blockers   []int    `json:"blockers,omitempty" jsonschema:"the numbers of the tasks in its way"`
	Links      []int    `json:"links,omitempty" jsonschema:"the numbers of the tasks it is linked to"`
	Objectives []int    `json:"objectives,omitempty" jsonschema:"the numbers of the objectives it serves"`
}

// Moved is what a change put on a task or an objective, and what it took off.
type Moved struct {
	Added    Sets
	TakenOff Sets
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

func movedOf(was, now task.Task) *Moved {
	before, after := factsOf(was), factsOf(now)
	var m Moved
	m.Added.Labels, m.TakenOff.Labels = between(before.Labels, after.Labels)
	m.Added.Blockers, m.TakenOff.Blockers = between(before.Blockers, after.Blockers)
	m.Added.Links, m.TakenOff.Links = between(before.Linked, after.Linked)
	m.Added.Objectives, m.TakenOff.Objectives = between(before.Objectives, after.Objectives)
	return m.said()
}

func movedOfObjective(was, now objective.Objective) *Moved {
	var m Moved
	m.Added.Labels, m.TakenOff.Labels = between(factsOfObjective(was).Labels, factsOfObjective(now).Labels)
	return m.said()
}

func (m Moved) said() *Moved {
	if m.Added.empty() && m.TakenOff.empty() {
		return nil
	}
	return &m
}

func (s Sets) empty() bool {
	return len(s.Labels) == 0 && len(s.Blockers) == 0 && len(s.Links) == 0 && len(s.Objectives) == 0
}

func (m *Moved) added() *Sets {
	if m == nil || m.Added.empty() {
		return nil
	}
	return &m.Added
}

func (m *Moved) takenOff() *Sets {
	if m == nil || m.TakenOff.empty() {
		return nil
	}
	return &m.TakenOff
}

func movedText(m *Moved) string {
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
		out = append(out, domain.Ref(n))
	}
	return out
}
