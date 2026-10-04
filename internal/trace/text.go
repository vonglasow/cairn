package trace

// The two renderings of an act: the line a person reads in a history, and the
// shape a machine reads. Both live here, with the notion, so neither surface
// owns either and the two cannot drift apart.

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"cairn/internal/domain"
)

// went is what sits between what a value was and what it became. Written out
// rather than drawn with an arrow glyph: a history is grepped as often as it is
// read, and a word survives a terminal that a glyph may not.
const went = "→"

// nest is what sets apart what a task was born carrying. Indentation, the way a
// listing sets apart a sub-task: those acts belong to the creation above them
// and share its moment, so repeating it on each would say the same thing four
// times.
const nest = "    "

// Text renders a history, oldest first.
//
// Nothing to render is not a case this handles: a task always carries at least
// its creation, so an empty history means the task was never there, and that is
// a refusal the engine gives before anything gets here.
//
// Laid out through a tabwriter so the moments, the acts and the values line up
// in columns: a history is read down one of those columns far more often than
// across a line.
func Text(acts []Act) string {
	var out strings.Builder
	tw := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)
	for _, a := range acts {
		var l line
		a.Tell(&l)
		_, _ = fmt.Fprintln(tw, l.columns())
	}
	_ = tw.Flush()
	return out.String()
}

// List renders the acts of a period across tasks, each naming its task, and
// says so when there were none.
//
// What a task was born carrying is left out: those acts belong to that task's
// own history, and a day is unreadable with four lines of birth for every task
// created in it.
func List(acts []Act) string {
	shown := make([]Act, 0, len(acts))
	for _, a := range acts {
		var f fact
		a.Tell(&f)
		if born(a) {
			continue
		}
		shown = append(shown, a)
	}
	if len(shown) == 0 {
		return "nothing was done\n"
	}
	var out strings.Builder
	tw := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)
	for _, a := range shown {
		var l line
		a.Tell(&l)
		_, _ = fmt.Fprintf(tw, "%s\t#%d\t%s\n",
			domain.FormatMoment(l.at), l.task, strings.TrimPrefix(l.columns(), l.moment()+"\t"))
	}
	_ = tw.Flush()
	return out.String()
}

// born says whether an act is one of those recording what a task started as.
func born(a Act) bool {
	var l line
	a.Tell(&l)
	return l.did == Created && l.field != ""
}

// line is an act's membrane for one line of a history: it is told every fact
// and keeps the ones a line shows.
type line struct {
	task                             int
	did, field, was, became, through string
	at                               time.Time
}

func (l *line) On(n int)               { l.task = n }
func (l *line) Did(what string)        { l.did = what }
func (l *line) Changing(field string)  { l.field = field }
func (l *line) NothingNamed()          { l.field = "" }
func (l *line) Was(value string)       { l.was = value }
func (l *line) WasNothing()            { l.was = "" }
func (l *line) Became(value string)    { l.became = value }
func (l *line) BecameNothing()         { l.became = "" }
func (l *line) Through(surface string) { l.through = surface }
func (l *line) ByHand()                { l.through = "" }
func (l *line) At(when time.Time)      { l.at = when }

// columns lays the line out with a tab between each column, so a history can be
// aligned: the moment, what was done, and the change when there was one — with
// the surface beside it when it was not the terminal.
//
// An attribute a task was born carrying is laid out under the creation it
// belongs to, indented and without a moment of its own. Every other act carries
// its moment, then what was done, then the change when there was one, then the
// surface when it was not the terminal.
//
// A creation has no before, so it says its value plainly. An attribute that
// changed always has one — even when it was nothing, which is worth seeing.
func (l *line) columns() string {
	var out strings.Builder
	if l.did == Created && l.field != "" {
		_, _ = fmt.Fprintf(&out, "%s\t%s\t%s", nest, l.field, l.became)
		return out.String()
	}
	_, _ = fmt.Fprintf(&out, "%s\t%s", domain.FormatMoment(l.at), l.what())
	// Time added by hand counts for a day of its own, which is not the day it
	// was added on: the line says both, the moment it was written and the day.
	if l.did == TimeAdded {
		_, _ = fmt.Fprintf(&out, "\t%s for %s", l.became, l.was)
	} else if l.field == "" && l.became != "" {
		_, _ = fmt.Fprintf(&out, "\t%s", l.became)
	} else if l.field == "" && l.was != "" {
		// A deletion names no attribute and became nothing: what it says is
		// what the task held when it went. Printed, because a history that
		// keeps it and does not show it keeps it where nobody can read it.
		_, _ = fmt.Fprintf(&out, "\t%s", l.was)
	} else if l.field != "" && (l.was != "" || l.became != "") {
		_, _ = fmt.Fprintf(&out, "\t%s %s %s",
			domain.OrNothing(l.was), went, domain.OrNothing(l.became))
	}
	// Beside the value and not a column after it: a value written over several
	// lines ended its last line in that column's tab, became a cell, and pushed
	// the moments of every act after it to the right.
	if l.through != "" {
		_, _ = fmt.Fprintf(&out, "  (%s)", l.through)
	}
	return out.String()
}

// moment is how this line's moment reads, so a listing can take it back off
// the front and put the task number in its place.
func (l *line) moment() string { return domain.FormatMoment(l.at) }

// what an act is called in a history: the attribute when one changed, and what
// was done otherwise. A line saying "changed  priority" would say the same
// thing twice.
func (l *line) what() string {
	if l.field != "" {
		return l.field
	}
	return l.did
}

// SheetText writes how long the work took, day by day: each task's time before
// its title, so a title holding "1h" is never read as one, then what the day
// came to. A task still in focus when its day ended says so, since that time
// stopped at midnight rather than where the work did.
func SheetText(s Sheet, titles map[int]string) string {
	if len(s.Days) == 0 {
		return "nothing was in focus\n"
	}
	var out strings.Builder
	for i, day := range s.Days {
		if i > 0 {
			out.WriteString("\n")
		}
		_, _ = fmt.Fprintf(&out, "%s\n", domain.FormatDay(day.Day))
		tw := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)
		for _, w := range day.Tasks {
			left := ""
			if w.LeftOpen {
				left = " — still in focus at midnight"
			}
			_, _ = fmt.Fprintf(tw, "  #%d\t%s\t%s%s\n", w.Task, Hours(w.Took), titles[w.Task], left)
		}
		_ = tw.Flush()
		_, _ = fmt.Fprintf(&out, "  total  %s\n", Hours(day.Total))
	}
	return out.String()
}
