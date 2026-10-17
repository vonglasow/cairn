package trace

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"cairn/internal/domain"
)

const went = "→"

const nest = "    "

// Text renders a history, oldest first.
func Text(acts []Act) string {
	var out strings.Builder
	tw := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)
	for _, a := range acts {
		_, _ = fmt.Fprintln(tw, columns(a.Carried()))
	}
	_ = tw.Flush()
	return out.String()
}

// List renders the acts of a period across tasks, each naming its task, and says so when there were none.
func List(acts []Act) string {
	shown := make([]Act, 0, len(acts))
	for _, a := range acts {
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
		l := a.Carried()
		_, _ = fmt.Fprintf(tw, "%s\t#%d\t%s\n",
			domain.FormatMoment(l.At), l.Task, strings.TrimPrefix(columns(l), domain.FormatMoment(l.At)+"\t"))
	}
	_ = tw.Flush()
	return out.String()
}

func born(a Act) bool {
	l := a.Carried()
	return l.Did == Created && l.Field != ""
}

func columns(l Carried) string {
	var out strings.Builder
	if l.Did == Created && l.Field != "" {
		fmt.Fprintf(&out, "%s\t%s\t%s", nest, l.Field, l.Became)
		return out.String()
	}
	fmt.Fprintf(&out, "%s\t%s", domain.FormatMoment(l.At), what(l))
	if l.Did == TimeAdded && l.Was == "" {
		fmt.Fprintf(&out, "\t%s", l.Became)
	} else if l.Did == TimeAdded {
		fmt.Fprintf(&out, "\t%s for %s", l.Became, l.Was)
	} else if l.Field == "" && l.Became != "" {
		fmt.Fprintf(&out, "\t%s", l.Became)
	} else if l.Field == "" && l.Was != "" {
		fmt.Fprintf(&out, "\t%s", l.Was)
	} else if l.Field != "" && (l.Was != "" || l.Became != "") {
		fmt.Fprintf(&out, "\t%s %s %s",
			domain.OrNothing(l.Was), went, domain.OrNothing(l.Became))
	}
	out.WriteString(Credited(l.By, l.Through))
	return out.String()
}

// Credited is what follows an act on its line: who did it and through what, as much of it as was written down, and nothing when neither was.
func Credited(by, through string) string {
	var said []string
	if by != "" {
		said = append(said, "by "+by)
	}
	if through != "" {
		said = append(said, "through "+through)
	}
	if len(said) == 0 {
		return ""
	}
	return "  (" + strings.Join(said, ", ") + ")"
}

func what(l Carried) string {
	if l.Field != "" {
		return l.Field
	}
	return l.Did
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
