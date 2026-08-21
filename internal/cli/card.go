package cli

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"cairn/internal/domain"
)

// absent is what an attribute a task does not carry looks like. A blank would
// read as a value nobody typed rather than as an absence.
const absent = "—"

// card renders a task in full, empty attributes included. Rendering lives here
// and only here: the day a second surface shows a task, it asks for this
// rather than inventing a layout of its own.
func card(t domain.Task, p domain.Progress) string {
	var out strings.Builder
	_, _ = fmt.Fprintf(&out, "#%d  %s\n\n", t.Number, t.Title)

	tw := tabwriter.NewWriter(&out, 0, 2, 2, ' ', 0)
	for _, field := range []struct{ key, value string }{
		{"state", string(t.State)},
		{"priority", string(t.Priority)},
		{"description", t.Description},
		{"labels", strings.Join(t.Labels, ", ")},
		{"deadline", domain.FormatDay(t.Deadline)},
		{"estimate", domain.FormatDuration(t.Estimate)},
		{"time spent", domain.FormatDuration(t.TimeSpent)},
		{"parent", parentOf(t)},
		{"objective", servesOf(t)},
		{"progress", progressLine(p, "task")},
	} {
		_, _ = fmt.Fprintf(tw, "%s\t%s\n", field.key, orNothing(field.value))
	}
	_ = tw.Flush()
	return out.String()
}

// parentOf writes the number a task hangs under, or nothing when it hangs
// under nothing.
func parentOf(t domain.Task) string {
	if t.Parent == domain.NoParent {
		return ""
	}
	return fmt.Sprintf("%d", t.Parent)
}

// servesOf writes the number of the objective a task serves, or nothing when
// it serves none.
func servesOf(t domain.Task) string {
	if t.Objective == domain.NoObjective {
		return ""
	}
	return fmt.Sprintf("%d", t.Objective)
}

// progressLine says how far along something is, and says nothing at all when
// nothing is under it. Zero of zero is not "none of it done": there is no work
// to be done, and a share of nothing would read as a standstill.
//
// Tasks and sub-tasks are named apart because they are counted apart, while
// the share is over both — three tasks and four sub-tasks is never seven
// tasks, but all seven have to happen.
func progressLine(p domain.Progress, what string) string {
	if !p.Any() {
		return ""
	}
	var counts []string
	if p.Tasks.Total > 0 {
		counts = append(counts, fmt.Sprintf("%d of %d tasks", p.Tasks.Done, p.Tasks.Total))
	}
	if p.SubTasks.Total > 0 {
		counts = append(counts, fmt.Sprintf("%d of %d sub-tasks", p.SubTasks.Done, p.SubTasks.Total))
	}
	return fmt.Sprintf("%s are done — the %s is %d%% done",
		strings.Join(counts, " and "), what, p.Percent())
}

// objectiveCard renders an objective in full. It is not a task's card with
// fields removed: an objective has no column, no parent and no time spent, and
// a shared renderer would have to be told which half to leave out every time.
func objectiveCard(o domain.Objective, p domain.Progress) string {
	var out strings.Builder
	_, _ = fmt.Fprintf(&out, "#%d  %s\n\n", o.Number, o.Title)

	tw := tabwriter.NewWriter(&out, 0, 2, 2, ' ', 0)
	for _, field := range []struct{ key, value string }{
		{"state", string(o.State)},
		{"priority", string(o.Priority)},
		{"description", o.Description},
		{"labels", strings.Join(o.Labels, ", ")},
		{"deadline", domain.FormatDay(o.Deadline)},
		{"estimate", domain.FormatDuration(o.Estimate)},
		{"progress", progressLine(p, "objective")},
	} {
		_, _ = fmt.Fprintf(tw, "%s\t%s\n", field.key, orNothing(field.value))
	}
	_ = tw.Flush()
	return out.String()
}

// serving renders the tasks put under an objective, and says so when there are
// none — an objective with an empty heading would read as one whose tasks
// failed to load.
func serving(tasks []domain.Task) string {
	if len(tasks) == 0 {
		return "\nnothing is under it\n"
	}
	var out strings.Builder
	out.WriteString("\ntasks\n")
	for _, t := range tasks {
		_, _ = fmt.Fprintf(&out, "  %s\n", line(t))
	}
	return out.String()
}

// objectiveList renders every objective as one row each. No tree: what is
// under an objective is shown by `objective show`, and a listing that unfolded
// every task would stop being a listing of objectives.
func objectiveList(all []domain.Objective) string {
	var out strings.Builder
	for _, o := range all {
		_, _ = fmt.Fprintf(&out, "#%d  %s\n", o.Number, o.Title)
	}
	return out.String()
}

// under renders what hangs beneath a task, and nothing at all when nothing
// does — an empty heading would read as a section somebody forgot to fill.
func under(tasks []domain.Task) string {
	if len(tasks) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString("\nsub-tasks\n")
	for _, t := range tasks {
		_, _ = fmt.Fprintf(&out, "  %s\n", line(t))
	}
	return out.String()
}

// nest is what sets a sub-task apart in a listing. Indentation rather than a
// drawn tree: one level does not need corners, and an indented line still
// reads when someone greps for it.
const nest = "    "

// tree renders a listing with each sub-task under the task it hangs from.
// Everything keeps its own number, so a line found here can be acted on
// directly — the indentation says where it belongs, not what it is.
func tree(tasks []domain.Task) string {
	hanging := map[int][]domain.Task{}
	for _, t := range tasks {
		if t.Parent != domain.NoParent {
			hanging[t.Parent] = append(hanging[t.Parent], t)
		}
	}

	var out strings.Builder
	for _, t := range tasks {
		if t.Parent != domain.NoParent {
			continue
		}
		out.WriteString(line(t))
		out.WriteString("\n")
		for _, under := range hanging[t.Number] {
			out.WriteString(nest)
			out.WriteString(line(under))
			out.WriteString("\n")
		}
	}
	return out.String()
}

// line renders a task as one row of a listing: its number, then its title.
func line(t domain.Task) string {
	return fmt.Sprintf("#%d  %s", t.Number, t.Title)
}

// orNothing shows an absence as an absence.
func orNothing(value string) string {
	if value == "" {
		return absent
	}
	return value
}
