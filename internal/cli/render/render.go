// Package render writes an answer the way a person reads it, for every surface.
package render

import (
	"cairn/internal/engine"
	"cairn/internal/engine/agent"
	"cairn/internal/engine/app"
	"cairn/internal/engine/backup"
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/exchange"
	"cairn/internal/engine/item"
	"cairn/internal/engine/label"
	"cairn/internal/engine/series"
	"cairn/internal/engine/view"
	"cairn/internal/engine/workspace"
	"fmt"
	"slices"
	"strings"
	"text/tabwriter"
	"time"
	"unicode/utf8"

	"cairn/internal/engine/act"
	"cairn/internal/engine/migration"
	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
)

func contextText(at string) string {
	if at == "" {
		return ""
	}
	return "\ncontext  " + at + "\n"
}

// Text turns an answer into what a person reads: the rendering of the shape the operation declared it answers, the way engine.Of builds the JSON.
func Text(op engine.Operation, a engine.Answer) string {
	if _, isSeries := op.Answers.(engine.Series); isSeries {
		return item.Shown(rendered(op, a))
	}
	return item.Shown(rendered(op, a) + series.FollowedText(a.Followed))
}

func rendered(op engine.Operation, a engine.Answer) string {
	switch op.Answers.(type) {
	case workspace.ConfigChanged:
		return workspace.ConfigChangedText(*a.ConfigChanged)
	case workspace.Config:
		return workspace.ConfigText(*a.Config)
	case agent.Agents:
		return agent.AgentsText(*a.Agents)
	case agent.Agent:
		return agent.AgentsText(agent.Agents{Agents: []agent.Agent{*a.Agent}})
	case agent.Started:
		return agent.StartedText(*a.Started)
	case agent.Terminals:
		return agent.TerminalsText(*a.Terminals)
	case agent.Decision:
		return agent.DecisionText(*a.Decision)
	case workspace.Workspaces:
		return workspace.WorkspacesText(a.Workspaces.Workspaces) + workspace.RecentText(a.Workspaces.Recent)
	case workspace.Removed:
		return workspace.RemovedText(*a.WorkspaceRemoved)
	case workspace.Added:
		return workspace.AddedText(*a.WorkspaceAdded)
	case workspace.Used:
		return workspace.UsedText(*a.WorkspaceUsed)
	case workspace.Shown:
		return workspace.ShownText(*a.WorkspaceShown)
	case engine.Timesheet:
		return act.SheetText(*a.Sheet, namedAsShown(a.Named, a.Gone))
	case engine.Reported:
		return objective.Report(*a.Objective, a.Progress, a.ReviewOf(), *a.Period)
	case engine.Focus:
		if a.Task != nil {
			return cardText(a) + warningsText(a.Warnings)
		}
		return heldText(a.Held, a) + sessionAsksText(a)
	case engine.Task:
		if op.Verb == engine.TaskSkip.Verb {
			return fmt.Sprintf("#%d skipped\n", (*a.Task).Carried().Number)
		}
		return cardText(a) + movedText(a) + warningsText(a.Warnings)
	case engine.Swept:
		if a.Task != nil {
			return cardText(a)
		}
		return fmt.Sprintf("%d %s(s) archived\n", *a.Archived, op.Thing)
	case engine.Tasks:
		if a.Counts != nil {
			return countsText(a)
		}
		if a.Matched != nil {
			return lostText(a) + foundText(a.Matched, a) + pageText(a, len(a.Matched)) + view.OutsideText(a.Outside)
		}
		return lostText(a) + crossedText(a.Crossed) + warningsText(a.Warnings) + tasksText(a.Tasks, a) + pageText(a, len(a.Tasks)) + view.OutsideText(a.Outside)
	case engine.ObjectiveRow:
		return objectiveText(a) + movedText(a)
	case engine.ObjectivesListed:
		if a.MatchedObjectives != nil {
			return objectivesText(a.MatchedObjectives, a)
		}
		return objectivesText(a.Objectives, a)
	case engine.BackupShape:
		return backupLine(*a.Backup, *a.Holds)
	case engine.BackupsKept:
		return backupsText(a.Backups)
	case migration.RolledBack:
		return migration.RolledBackText(*a.RolledBack)
	case migration.Shape:
		return migration.Text(*a.Change)
	case engine.Ledger:
		return migration.List(a.Changes, *a.Holds)
	case engine.Labelled:
		return fmt.Sprintf("%d %s(s) labelled\n", *a.Labelled, op.Thing)
	case label.List:
		return labelsText(*a.Labels) + warningsText(a.Warnings)
	case engine.View:
		if a.Viewed.Outside != nil {
			return viewShownText(*a.Viewed)
		}
		return viewKeptText(*a.Viewed, map[bool]string{true: "kept view ", false: "changed view "}[op.Verb == "new"]+a.Viewed.Name)
	case engine.Views:
		return viewsText(a.ViewsKept.Views)
	case engine.Series:
		return seriesActedText(op.Verb, a)
	case engine.SeriesListed:
		return seriesListText(a.SeriesKept.Series)
	case series.Deleted:
		return seriesDeletedText(*a.SeriesGone)
	case view.Deleted:
		return "deleted view " + *a.ViewGone + " · no task changed\n"
	case workspace.ConfigWritten:
		return workspace.ConfigWrittenText(*a.Written)
	case engine.DebugReported:
		return fmt.Sprintf("debug report written to %s, %d debug file(s) in it\n", *a.Written, *a.Packed)
	case task.Steps:
		return stepsText(*a.Task)
	case engine.Deleted:
		said := fmt.Sprintf("%s %d was deleted\n", op.Thing, *a.Deleted)
		for _, n := range a.Also {
			said += fmt.Sprintf("%s %d was deleted\n", op.Thing, n)
		}
		for _, n := range a.Unblocked {
			said += fmt.Sprintf("task %d was waiting on it, and nothing names what stops it now\n", n)
		}
		return said + objective.WentText(a.Went)
	case engine.Restored:
		return fmt.Sprintf("backup %d was restored\n", *a.Restored)
	case app.PluginWritten:
		return app.PluginText(*a.Plugin)
	case app.PluginShown:
		return app.PluginShownText(*a.Install)
	case app.Upgraded:
		return a.Upgraded.String()
	case app.Installed:
		return app.Text(engine.Of(op, a).(app.Installed))
	case exchange.Exported:
		return *a.Exported + " " + backup.HeldText(*a.Holds) + "\n"
	case exchange.ImportedFile:
		return fmt.Sprintf("%d task(s) and %d objective(s) were added from %s\n", a.Holds.Tasks, a.Holds.Objectives, *a.Imported)
	case engine.Applied:
		return appliedText(a.Applied)
	case engine.Acts:
		if a.Since != nil {
			return act.List(a.Since)
		}
		return act.Text(a.Acts)
	}
	return ""
}

func cardText(a engine.Answer) string {
	said := task.Text(*a.Task, a.Progress, !a.Untimed, a.Choosing[(*a.Task).Carried().Number]) + task.StepsText((*a.Task).Carried().Steps) + stoodText(*a.Task, a.StoodSince) + contextText(a.Context)
	if o, is := a.Occurring[(*a.Task).Carried().Number]; is {
		said += series.OccursText(o)
	}
	return said
}

func stoodText(t task.Task, since time.Time) string {
	if since.IsZero() {
		return ""
	}
	return fmt.Sprintf("\nin %s since %s · %s\n", t.Carried().State, calendar.FormatDay(since.Local()), daysSince(since, time.Now()))
}

func daysSince(since, now time.Time) string {
	days := calendar.DaysFrom(since, now)
	if days == 0 {
		return "today"
	}
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}

func crossedText(c *task.Crossed) string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf("task %d moved from %s to %s\n", c.Task, c.From, c.To)
}

func heldText(held []task.Task, a engine.Answer) string {
	if len(held) == 0 {
		return "nothing is in focus\n"
	}
	return task.Lines(held, marksOf(a))
}

func sessionAsksText(a engine.Answer) string {
	var b strings.Builder
	for _, s := range a.Sessions {
		if s.Ask != "" {
			b.WriteString(s.Ask + "\n")
		}
	}
	return b.String()
}

func tasksText(tasks []task.Task, a engine.Answer) string {
	if len(tasks) == 0 {
		return "there is no task\n"
	}
	return task.Lines(tasks, marksOf(a))
}

func foundText(tasks []task.Task, a engine.Answer) string {
	if len(tasks) == 0 {
		return "there is no task\n"
	}
	return task.Lines(tasks, marksOf(a))
}

func marksOf(a engine.Answer) task.Marks {
	return task.Marks{Blocks: task.BlocksOf(a.Choosing), InFocus: engine.HeldOr(a.InFocus), NothingBlocks: a.NothingBlocks, Repeats: series.RepeatsOf(a.Occurring), Serves: task.ServesOf(a.Choosing)}
}

func linesOf(tasks []task.Task) []string {
	out := make([]string, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, task.Line(t))
	}
	return out
}

func objectivesText(all []objective.Objective, a engine.Answer) string {
	if len(all) == 0 {
		return "there is no objective\n"
	}
	return objective.List(all, a.Along, a.Spent)
}

func objectiveText(a engine.Answer) string {
	var f objective.Family
	if a.Family != nil {
		f = *a.Family
	}
	said := objective.Text(*a.Objective, a.Progress, a.ReviewOf(), f)
	if len(f.Children) == 0 {
		said += objective.Serving(linesOf(a.Under))
	} else {
		n := a.Objective.Carried().Number
		var own, through []task.Task
		for _, t := range a.Under {
			if slices.Contains(t.Carried().Objectives, n) {
				own = append(own, t)
			} else {
				through = append(through, t)
			}
		}
		said += objective.Children(f) + objective.Apart(linesOf(own), linesOf(through))
	}
	if f.Proposed {
		said += objective.Proposal(*a.Objective, true)
	}
	if a.Proposes != nil {
		said += objective.Proposal(*a.Proposes, false)
	}
	return said + objective.Awaited(f.Awaited)
}

func warningsText(said []string) string {
	var out strings.Builder
	for _, one := range said {
		out.WriteString("warning: " + one + "\n")
	}
	return out.String()
}

func backupLine(b backup.Backup, holds backup.Held) string {
	return fmt.Sprintf("#%d  made %s  %s\n", b.Number, calendar.FormatMoment(b.Made), backup.HeldText(holds))
}

func backupsText(all []backup.Backup) string {
	if len(all) == 0 {
		return "there is no backup\n"
	}
	var out strings.Builder
	for _, one := range all {
		fmt.Fprintf(&out, "#%d  made %s\n", one.Number, calendar.FormatMoment(one.Made))
	}
	return out.String()
}

func appliedText(made []string) string {
	if len(made) == 0 {
		return "there is nothing waiting to be applied\n"
	}
	var out strings.Builder
	for _, name := range made {
		fmt.Fprintf(&out, "applied %q\n", name)
	}
	return out.String()
}

func namedAsShown(titles map[int]string, gone map[int]bool) map[int]string {
	shown := make(map[int]string, len(titles))
	for number, title := range titles {
		shown[number] = title
		if gone[number] {
			shown[number] = title + engine.DeletedMark
		}
	}
	return shown
}

func lostText(a engine.Answer) string {
	if a.InView == "" || len(a.Lost) == 0 {
		return ""
	}
	name := view.TypedName(a.InView)
	hint := fmt.Sprintf("  cairn view edit %s, or cairn view delete %s\n", name, name)
	if a.KeepsNothing {
		return fmt.Sprintf("every condition view %s holds is gone, so it keeps nothing: %s\n", a.InView, strings.Join(a.Lost, ", ")) + hint
	}
	verb := "is gone and narrows nothing"
	if len(a.Lost) > 1 {
		verb = "are gone and narrow nothing"
	}
	return fmt.Sprintf("view %s: %s %s\n", a.InView, strings.Join(a.Lost, " and "), verb) + hint
}

func viewShownText(v engine.View) string {
	lines := []string{
		fmt.Sprintf("view     %s  (%s)", v.Name, view.CmdKey(v.Place)),
		"holds    " + engine.HoldsSays(v),
		fmt.Sprintf("keeps    %d of %s · %d done", v.Keeps, item.ManyOf(v.Of, "task", "tasks"), v.Done),
	}
	if len(v.Gone) > 0 {
		lines = append(lines, "gone     "+strings.Join(v.Gone, " · "))
	}
	if len(v.Outside) > 0 {
		var outside []string
		for _, o := range v.Outside {
			var why []string
			for _, w := range o.Why {
				why = append(why, fmt.Sprintf("%s #%d", w.Is, w.Of))
			}
			outside = append(outside, fmt.Sprintf("#%d %s", o.Number, strings.Join(why, ", ")))
		}
		lines = append(lines, "outside  "+strings.Join(outside, " · "))
	}
	for _, o := range v.Objectives {
		lines = append(lines, fmt.Sprintf("objective #%d %s  %s", o.Number, o.Title, view.StoreFigures(o)))
	}
	return strings.Join(lines, "\n") + "\n"
}

func viewKeptText(v engine.View, said string) string {
	if v.Was != "" {
		said += " (was " + v.Was + ")"
	}
	return fmt.Sprintf("%s · %s · keeps %d of %s\n", said, view.CmdKey(v.Place), v.Keeps, item.ManyOf(v.Of, "task", "tasks"))
}

func viewsText(views []engine.View) string {
	if len(views) == 0 {
		return "there is no view\n"
	}
	var out strings.Builder
	tw := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "#\tname\tholds\tkeeps")
	for _, v := range views {
		_, _ = fmt.Fprintf(tw, "%d\t%s\t%s\t%d\n", v.Place, v.Name, engine.HoldsSays(v), v.Keeps)
	}
	_ = tw.Flush()
	return out.String()
}

func pageText(a engine.Answer, shown int) string {
	if a.Total == nil {
		return ""
	}
	if shown == 0 {
		return fmt.Sprintf("no task past %d, of %d\n", a.Offset, *a.Total)
	}
	return fmt.Sprintf("tasks %d to %d of %d\n", a.Offset+1, a.Offset+shown, *a.Total)
}

func countsText(a engine.Answer) string {
	width := 0
	for _, c := range a.Counts {
		width = max(width, utf8.RuneCountInString(c.Of))
	}
	var said strings.Builder
	for _, c := range a.Counts {
		fmt.Fprintf(&said, "%-*s  %d\n", width, c.Of, c.Tasks)
	}
	fmt.Fprintf(&said, "%d task(s) in all\n", *a.Total)
	return said.String()
}

func movedText(a engine.Answer) string {
	m := a.Moved
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

func labelsText(listed []label.Carried) string {
	if len(listed) == 0 {
		return "no label is carried\n"
	}
	var out strings.Builder
	tw := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)
	for _, one := range listed {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\tfor %s\n", one.Label, item.ManyOf(one.Tasks, "task", "tasks"), item.ManyOf(one.Objectives, "objective", "objectives"), one.Scope)
	}
	_ = tw.Flush()
	return out.String()
}

func seriesActedText(verb string, a engine.Answer) string {
	s := *a.SeriesShown
	switch verb {
	case "stop":
		if len(s.Open) > 0 {
			return fmt.Sprintf("↻%d stopped; %s stays open, and nothing more is made\n", s.Number, act.NumbersOrNothing(s.Open))
		}
		return fmt.Sprintf("↻%d stopped; nothing more is made\n", s.Number)
	case "resume":
		if len(a.Followed) > 0 {
			return fmt.Sprintf("↻%d resumed · next %s, made as #%d\n", s.Number, series.ShortDays([]string{a.Followed[0].Day}), a.Followed[0].Task)
		}
		if len(s.Open) > 0 {
			return fmt.Sprintf("↻%d resumed; %s is still open, and the next comes when it is done or skipped\n", s.Number, act.NumbersOrNothing(s.Open))
		}
		return fmt.Sprintf("↻%d resumed; it has ended, and nothing more is made\n", s.Number)
	case "edit":
		return fmt.Sprintf("changed ↻%d %s\nrule      %s\ntemplate  %s\n", s.Number, s.Title, s.Says, engine.TemplateSaid(s.Template))
	}
	return seriesShownText(s)
}

func seriesListText(all []engine.Series) string {
	if len(all) == 0 {
		return "there is no series\n"
	}
	var out strings.Builder
	for _, s := range all {
		fmt.Fprintf(&out, "↻%d  %s  · %s", s.Number, s.Title, engine.StandingSaid(s))
		if s.State == "active" {
			next := s
			next.Next = s.Next[:min(1, len(s.Next))]
			fmt.Fprintf(&out, "  · next %s", engine.NextSaid(next))
		}
		fmt.Fprintf(&out, "\n    %s · %d done · %d skipped · %d missed\n", strings.SplitN(s.Says, ", ", 2)[0], s.Done, s.Skipped, s.Missed)
	}
	return out.String()
}

func seriesDeletedText(d series.Deleted) string {
	said := fmt.Sprintf("↻%d was deleted", d.Deleted)
	if len(d.TasksDeleted) > 0 {
		said += " · " + act.NumbersOrNothing(d.TasksDeleted) + " deleted"
	}
	if len(d.Kept) > 0 {
		said += fmt.Sprintf(" · %s kept as ordinary tasks", item.ManyOf(len(d.Kept), "task", "tasks"))
	}
	return said + "\n"
}

func seriesShownText(s engine.Series) string {
	var out strings.Builder
	fmt.Fprintf(&out, "↻%d  %s  %s\n\n", s.Number, s.Title, engine.StandingSaid(s))
	fmt.Fprintf(&out, "rule      %s\nnext      %s\n", s.Says, engine.NextSaid(s))
	if strings.Contains(s.Says, "after the last is done") && !s.Ended {
		rule, _ := series.ParseRule(s.Rule)
		fmt.Fprintf(&out, "then      %s\n", rule.Then())
	}
	fmt.Fprintf(&out, "template  %s\n", engine.TemplateSaid(s.Template))
	soFar := fmt.Sprintf("%d done · %d skipped · %d missed", s.Done, s.Skipped, s.Missed)
	if s.TimeSpent != "" {
		soFar += " · " + s.TimeSpent
	}
	fmt.Fprintf(&out, "so far    %s\nopen      %s\n", soFar, item.OrNothing(act.NumbersOrNothing(s.Open)))
	if len(s.Occurrences) > 0 {
		out.WriteString("\noccurrences\n")
		for _, o := range s.Occurrences {
			what := "—"
			if o.Task != 0 {
				what = item.Ref(o.Task)
			}
			is := o.Is
			if o.Archived {
				is += " · archived"
			}
			fmt.Fprintf(&out, "  %-10s  %-5s  %s", series.ShortDays([]string{o.Day}), what, is)
			if o.TimeSpent != "" {
				out.WriteString("  " + o.TimeSpent)
			}
			out.WriteString("\n")
		}
	}
	if len(s.Acts) > 0 {
		out.WriteString("\ntrace\n")
		for _, a := range s.Acts {
			at, _ := time.Parse(time.RFC3339, a.At)
			line := fmt.Sprintf("  %s  %s", calendar.FormatMoment(at), series.ActSaid(a))
			for _, more := range []string{a.Through, a.By} {
				if more != "" {
					line += " · " + more
				}
			}
			out.WriteString(line + "\n")
		}
	}
	return out.String()
}

func stepsText(t task.Task) string {
	steps := t.Carried().Steps
	if len(steps) == 0 {
		return fmt.Sprintf("task %d has no steps\n", t.Carried().Number)
	}
	return task.StepsText(steps)[1:]
}
