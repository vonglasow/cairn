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
func card(t domain.Task) string {
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
	} {
		_, _ = fmt.Fprintf(tw, "%s\t%s\n", field.key, orNothing(field.value))
	}
	_ = tw.Flush()
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
