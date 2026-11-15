package engine

import (
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
	"math"
	"slices"
	"strings"
	"time"

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

// RenderText turns an answer into what a person reads: the rendering of the shape the operation declared it answers, the way Of builds the JSON.
func RenderText(op Operation, a Answer) string {
	if _, isSeries := op.Answers.(Series); isSeries {
		return item.Shown(rendered(op, a))
	}
	return item.Shown(rendered(op, a) + series.FollowedText(a.Followed))
}

func rendered(op Operation, a Answer) string {
	switch op.Answers.(type) {
	case workspace.ConfigChanged:
		return workspace.ConfigChangedText(*a.ConfigChanged)
	case workspace.Config:
		return workspace.ConfigText(*a.Config)
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
	case Timesheet:
		return act.SheetText(*a.Sheet, namedAsShown(a.Named, a.Gone))
	case review:
		return objective.Report(*a.Objective, a.Progress, a.review(), *a.Period)
	case Focus:
		if a.Task != nil {
			return cardText(a) + warningsText(a.Warnings)
		}
		return heldText(a.Held, a) + sessionAsksText(a.Sessions)
	case Task:
		if op.Verb == TaskSkip.Verb {
			return fmt.Sprintf("#%d skipped\n", (*a.Task).Carried().Number)
		}
		return cardText(a) + movedText(a.Moved) + warningsText(a.Warnings)
	case swept:
		if a.Task != nil {
			return cardText(a)
		}
		return fmt.Sprintf("%d %s(s) archived\n", *a.Archived, op.Thing)
	case Tasks:
		if a.Counts != nil {
			return countsText(a.Counts, *a.Total)
		}
		if a.Matched != nil {
			return lostText(a) + foundText(a.Matched, a) + pageText(a, len(a.Matched)) + view.OutsideText(a.Outside)
		}
		return lostText(a) + crossedText(a.Crossed) + warningsText(a.Warnings) + tasksText(a.Tasks, a) + pageText(a, len(a.Tasks)) + view.OutsideText(a.Outside)
	case objectiveRow:
		return objectiveText(a) + movedText(a.Moved)
	case objectivesListed:
		if a.MatchedObjectives != nil {
			return objectivesText(a.MatchedObjectives, a)
		}
		return objectivesText(a.Objectives, a)
	case BackupShape:
		return backupLine(*a.Backup, *a.Holds)
	case kept:
		return backupsText(a.Backups)
	case migration.RolledBack:
		return migration.RolledBackText(*a.RolledBack)
	case migration.Shape:
		return migration.Text(*a.Change)
	case ledger:
		return migration.List(a.Changes, *a.Holds)
	case labelled:
		return fmt.Sprintf("%d %s(s) labelled\n", *a.Labelled, op.Thing)
	case label.List:
		return labelsText(*a.Labels) + warningsText(a.Warnings)
	case View:
		if a.Viewed.Outside != nil {
			return viewShownText(*a.Viewed)
		}
		return viewKeptText(*a.Viewed, map[bool]string{true: "kept view ", false: "changed view "}[op.Verb == "new"]+a.Viewed.Name)
	case Views:
		return viewsText(a.ViewsKept.Views)
	case Series:
		return seriesActedText(op.Verb, a)
	case seriesListed:
		return seriesListText(a.SeriesKept.Series)
	case series.Deleted:
		return seriesDeletedText(*a.SeriesGone)
	case view.Deleted:
		return "deleted view " + *a.ViewGone + " · no task changed\n"
	case workspace.ConfigWritten:
		return workspace.ConfigWrittenText(*a.Written)
	case debugReported:
		return fmt.Sprintf("debug report written to %s, %d debug file(s) in it\n", *a.Written, *a.Packed)
	case task.Steps:
		return stepsText(*a.Task)
	case Deleted:
		said := fmt.Sprintf("%s %d was deleted\n", op.Thing, *a.Deleted)
		for _, n := range a.Also {
			said += fmt.Sprintf("%s %d was deleted\n", op.Thing, n)
		}
		for _, n := range a.Unblocked {
			said += fmt.Sprintf("task %d was waiting on it, and nothing names what stops it now\n", n)
		}
		return said + objective.WentText(a.Went)
	case restored:
		return fmt.Sprintf("backup %d was restored\n", *a.Restored)
	case app.PluginWritten:
		return app.PluginText(*a.Plugin)
	case app.PluginShown:
		return app.PluginShownText(*a.Install)
	case app.Upgraded:
		return a.Upgraded.String()
	case app.Installed:
		return app.Text(Of(op, a).(app.Installed))
	case exchange.Exported:
		return *a.Exported + " " + backup.HeldText(*a.Holds) + "\n"
	case exchange.ImportedFile:
		return fmt.Sprintf("%d task(s) and %d objective(s) were added from %s\n", a.Holds.Tasks, a.Holds.Objectives, *a.Imported)
	case applied:
		return appliedText(a.Applied)
	case acts:
		if a.Since != nil {
			return act.List(a.Since)
		}
		return act.Text(a.Acts)
	}
	return ""
}

func cardText(a Answer) string {
	said := task.Text(*a.Task, a.Progress, !a.Untimed, a.Choosing[(*a.Task).Carried().Number].Blocks...) + task.StepsText((*a.Task).Carried().Steps) + stoodText(*a.Task, a.StoodSince) + contextText(a.Context)
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

func daysFrom(since, now time.Time) int {
	from, to := since.Local(), now.Local()
	first := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.Local)
	last := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.Local)
	return int(math.Round(last.Sub(first).Hours() / 24))
}

func daysSince(since, now time.Time) string {
	days := daysFrom(since, now)
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

func heldText(held []task.Task, a Answer) string {
	if len(held) == 0 {
		return "nothing is in focus\n"
	}
	return task.Lines(held, marksOf(a))
}

func sessionAsksText(sessions []focusSession) string {
	var b strings.Builder
	for _, s := range sessions {
		if s.Ask != "" {
			b.WriteString(s.Ask + "\n")
		}
	}
	return b.String()
}

func tasksText(tasks []task.Task, a Answer) string {
	if len(tasks) == 0 {
		return "there is no task\n"
	}
	return task.Lines(tasks, marksOf(a))
}

func foundText(tasks []task.Task, a Answer) string {
	if len(tasks) == 0 {
		return "there is no task\n"
	}
	return task.Lines(tasks, marksOf(a))
}

func marksOf(a Answer) task.Marks {
	return task.Marks{Blocks: task.BlocksOf(a.Choosing), InFocus: heldOr(a.InFocus), NothingBlocks: a.NothingBlocks, Repeats: series.RepeatsOf(a.Occurring)}
}

func linesOf(tasks []task.Task) []string {
	out := make([]string, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, task.Line(t))
	}
	return out
}

func objectivesText(all []objective.Objective, a Answer) string {
	if len(all) == 0 {
		return "there is no objective\n"
	}
	return objective.List(all, a.Along, a.Spent)
}

func objectiveText(a Answer) string {
	var f objective.Family
	if a.Family != nil {
		f = *a.Family
	}
	said := objective.Text(*a.Objective, a.Progress, a.review(), f)
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
			shown[number] = title + deletedMark
		}
	}
	return shown
}
