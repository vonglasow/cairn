package trace

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"cairn/internal/domain"
)

const went = "→"

const nest = "    "

// Text renders a history, oldest first.
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

// List renders the acts of a period across tasks, each naming its task, and says so when there were none.
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

func born(a Act) bool {
	var l line
	a.Tell(&l)
	return l.did == Created && l.field != ""
}

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

func (l *line) columns() string {
	var out strings.Builder
	if l.did == Created && l.field != "" {
		fmt.Fprintf(&out, "%s\t%s\t%s", nest, l.field, l.became)
		return out.String()
	}
	fmt.Fprintf(&out, "%s\t%s", domain.FormatMoment(l.at), l.what())
	if l.did == TimeAdded {
		fmt.Fprintf(&out, "\t%s for %s", l.became, l.was)
	} else if l.field == "" && l.became != "" {
		fmt.Fprintf(&out, "\t%s", l.became)
	} else if l.field == "" && l.was != "" {
		fmt.Fprintf(&out, "\t%s", l.was)
	} else if l.field != "" && (l.was != "" || l.became != "") {
		fmt.Fprintf(&out, "\t%s %s %s",
			domain.OrNothing(l.was), went, domain.OrNothing(l.became))
	}
	if l.through != "" {
		fmt.Fprintf(&out, "  (%s)", l.through)
	}
	return out.String()
}

func (l *line) moment() string { return domain.FormatMoment(l.at) }

func (l *line) what() string {
	if l.field != "" {
		return l.field
	}
	return l.did
}

// SheetText writes how long the work took, day by day: each task's time before its title, so a title holding "1h" is never read as one, then what the day came to.
func SheetText(s Sheet, titles map[int]string) string {
	if len(s.Days) == 0 {
		return "nothing was in focus\n"
	}
	var out strings.Builder
	for i, day := range s.Days {
		if i > 0 {
			out.WriteString("\n")
		}
		fmt.Fprintf(&out, "%s\n", domain.FormatDay(day.Day))
		tw := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)
		for _, w := range day.Tasks {
			left := ""
			if w.LeftOpen {
				left = " — still in focus at midnight"
			}
			_, _ = fmt.Fprintf(tw, "  #%d\t%s\t%s%s\n", w.Task, Hours(w.Took), titles[w.Task], left)
		}
		_ = tw.Flush()
		fmt.Fprintf(&out, "  total  %s\n", Hours(day.Total))
	}
	return out.String()
}
