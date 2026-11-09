package objective

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// ErrWaitsOnItself is returned when an objective would wait on itself, directly or through others.
var ErrWaitsOnItself = errors.New("an objective does not wait on itself")

// MayWaitOn says whether o may wait on what it names, given every objective by its number: never on itself, nor on one that already waits on it, however far through others.
func MayWaitOn(o Objective, by map[int]Objective) error {
	if slices.Contains(o.waitsOn, o.number) {
		return fmt.Errorf("%s: %w", refOf(o), ErrWaitsOnItself)
	}
	for _, first := range o.waitsOn {
		if path := pathTo(first, o.number, by); path != nil {
			said := fmt.Sprintf("%s already waits on %s", refOf(by[first]), refOf(o))
			if between := path[1 : len(path)-1]; len(between) > 0 {
				names := make([]string, 0, len(between))
				for _, n := range between {
					names = append(names, refOf(by[n]))
				}
				said += ", through " + strings.Join(names, ", ")
			}
			return fmt.Errorf("%s; %w, even through others", said, ErrWaitsOnItself)
		}
	}
	return nil
}

func pathTo(from, to int, by map[int]Objective) []int {
	came := map[int]int{from: 0}
	queue := []int{from}
	for len(queue) > 0 {
		at := queue[0]
		queue = queue[1:]
		if at == to {
			path := []int{at}
			for at != from {
				at = came[at]
				path = append([]int{at}, path...)
			}
			return path
		}
		for _, next := range by[at].waitsOn {
			if _, seen := came[next]; !seen {
				came[next] = at
				queue = append(queue, next)
			}
		}
	}
	return nil
}

// Unmet names, a sentence each, what o waits on that is not reached yet.
func Unmet(o Objective, by map[int]Objective) []string {
	var said []string
	for _, n := range o.waitsOn {
		if on := by[n]; on.state != Reached {
			said = append(said, fmt.Sprintf("%s waits on %s, not reached", refOf(o), refOf(on)))
		}
	}
	return said
}

// Awaited names, a line each, the objectives that wait on the one shown; nothing when none does.
func Awaited(waiting []Objective) string {
	if len(waiting) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString("\n")
	for _, o := range waiting {
		fmt.Fprintf(&out, "%s waits on it\n", refOf(o))
	}
	return out.String()
}
