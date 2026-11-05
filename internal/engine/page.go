package engine

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"

	"cairn/internal/domain"
	"cairn/internal/task"
)

var pageArguments = []Argument{
	{Name: "limit", Kind: Number, Of: "a", Means: "answer at most this many tasks, in the board's order, the answer then saying under total how many there are in all"},
	{Name: "offset", Kind: Number, Of: "a", Means: "leave out this many tasks first, in the board's order, the answer then saying under total how many there are in all; past the last one, no task"},
}

var countArgument = Argument{
	Name:  "count",
	Kind:  Text,
	OneOf: []string{"state", "label"},
	Means: "answer no row, only how many of the tasks kept sit in each column (state) or carry each label (label), and under total how many there are in all",
}

// Count is how many of the tasks a listing kept sit in one column, or carry one label.
type Count struct {
	Of    string
	Tasks int
}

type page struct {
	asked         bool
	offset, limit int
}

func askedPage(in Input) (page, error) {
	var p page
	if limit, named := in.Whole("limit"); named {
		if limit <= 0 {
			return page{}, Misread(fmt.Errorf("%q: %q is at least 1", in.Given["limit"], "limit"))
		}
		p.asked, p.limit = true, limit
	}
	if offset, named := in.Whole("offset"); named {
		if offset < 0 {
			return page{}, Misread(fmt.Errorf("%q: %q is 0 or more", in.Given["offset"], "offset"))
		}
		p.asked, p.offset = true, offset
	}
	return p, nil
}

func (p page) of(tasks []task.Task) []task.Task {
	from := min(p.offset, len(tasks))
	to := len(tasks)
	if p.limit > 0 {
		to = min(from+p.limit, to)
	}
	return tasks[from:to]
}

func askedCount(in Input, fields []string, p page) (string, error) {
	by, named := in.Given[countArgument.Name]
	if !named {
		return "", nil
	}
	if err := accepted(in, countArgument); err != nil {
		return "", err
	}
	if fields != nil || p.asked {
		return "", Misread(fmt.Errorf("%q: a count answers no row, so it takes no fields, limit or offset", by))
	}
	return by, nil
}

func countedBy(tasks []task.Task, by string) []Count {
	tally := map[string]int{}
	for _, t := range tasks {
		f := factsOf(t)
		groups := f.Labels
		if by == "state" {
			groups = []string{string(f.State)}
		}
		for _, group := range groups {
			tally[group]++
		}
	}
	order := slices.Sorted(maps.Keys(tally))
	if by == "state" {
		order = slices.DeleteFunc(oneOf(domain.States), func(state string) bool { return tally[state] == 0 })
	}
	counts := make([]Count, 0, len(order))
	for _, group := range order {
		counts = append(counts, Count{Of: group, Tasks: tally[group]})
	}
	return counts
}

func ofCounts(counts []Count) *map[string]int {
	if counts == nil {
		return nil
	}
	said := make(map[string]int, len(counts))
	for _, c := range counts {
		said[c.Of] = c.Tasks
	}
	return &said
}

func countsText(counts []Count, total int) string {
	width := 0
	for _, c := range counts {
		width = max(width, utf8.RuneCountInString(c.Of))
	}
	var said strings.Builder
	for _, c := range counts {
		fmt.Fprintf(&said, "%-*s  %d\n", width, c.Of, c.Tasks)
	}
	fmt.Fprintf(&said, "%d task(s) in all\n", total)
	return said.String()
}

func pageText(a Answer, shown int) string {
	if a.Total == nil {
		return ""
	}
	if shown == 0 {
		return fmt.Sprintf("no task past %d, of %d\n", a.Offset, *a.Total)
	}
	return fmt.Sprintf("tasks %d to %d of %d\n", a.Offset+1, a.Offset+shown, *a.Total)
}
