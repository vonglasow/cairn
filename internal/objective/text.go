package objective

import (
	"fmt"
	"strings"

	"cairn/internal/domain"
)

// Text renders an objective in full, empty attributes included, the way a person reads it.
func Text(o Objective, p domain.Progress) string {
	c := o.Carried()
	return domain.Card(c.Number, c.Title, []domain.Fact{
		{Key: "state", Value: string(c.State)},
		{Key: "priority", Value: string(c.Priority)},
		{Key: "description", Value: c.Description},
		{Key: "labels", Value: strings.Join(c.Labels, ", ")},
		{Key: "deadline", Value: domain.FormatDay(c.Deadline)},
		{Key: "estimate", Value: domain.FormatDuration(c.Estimate)},
		{Key: "progress", Value: p.Line("the objective")},
	})
}

// Serving renders the tasks put under an objective as the lines already built for each, and says so when there are none — an objective with an empty heading would read as one whose tasks failed to load.
func Serving(lines []string) string {
	if len(lines) == 0 {
		return "\nnothing is under it\n"
	}
	var out strings.Builder
	out.WriteString("\ntasks\n")
	for _, l := range lines {
		fmt.Fprintf(&out, "  %s\n", l)
	}
	return out.String()
}

// List renders every objective as one row each.
func List(all []Objective) string {
	var out strings.Builder
	for _, o := range all {
		r := o.Carried()
		line := fmt.Sprintf("#%d  %s", r.Number, r.Title)
		for _, said := range []string{
			string(r.State), string(r.Priority), domain.FormatDay(r.Deadline),
		} {
			if said != "" {
				line += "  · " + said
			}
		}
		fmt.Fprintf(&out, "%s\n", line)
	}
	return out.String()
}
