package engine

import (
	"fmt"
	"strings"
	"time"

	"cairn/internal/domain"
	"cairn/internal/migration"
	"cairn/internal/objective"
	"cairn/internal/task"
	"cairn/internal/trace"
)

func contextText(at string) string {
	if at == "" {
		return ""
	}
	return "\ncontext  " + at + "\n"
}

// RenderText turns an answer into what a person reads: the rendering of the shape the operation declared it answers, the way Of builds the JSON.
func RenderText(op Operation, a Answer) string {
	return domain.Shown(rendered(op, a))
}

func rendered(op Operation, a Answer) string {
	switch op.Answers.(type) {
	case Config:
		return configText(*a.Config)
	case Timesheet:
		return trace.SheetText(*a.Sheet, namedAsShown(a.Named, a.Gone))
	case Review:
		return objective.Report(*a.Objective, a.Progress, reviewOf(a.Under), *a.Period)
	case Focus:
		if a.Task != nil {
			return cardText(a)
		}
		return heldText(a.Held, a)
	case Task:
		return cardText(a)
	case Swept:
		if a.Task != nil {
			return cardText(a)
		}
		return fmt.Sprintf("%d %s(s) archived\n", *a.Archived, op.Thing)
	case Tasks:
		if a.Matched != nil {
			return foundText(a.Matched, a)
		}
		return tasksText(a.Tasks, a)
	case Objective:
		return objective.Text(*a.Objective, a.Progress, reviewOf(a.Under)) + objective.Serving(linesOf(a.Under))
	case Objectives:
		if a.MatchedObjectives != nil {
			return objectivesText(a.MatchedObjectives, a.Along)
		}
		return objectivesText(a.Objectives, a.Along)
	case BackupShape:
		return backupLine(*a.Backup, *a.Holds)
	case Kept:
		return backupsText(a.Backups)
	case migration.RolledBack:
		return migration.RolledBackText(*a.RolledBack)
	case migration.Shape:
		return migration.Text(*a.Change)
	case Ledger:
		return migration.List(a.Changes, *a.Holds)
	case Labelled:
		return fmt.Sprintf("%d %s(s) labelled\n", *a.Labelled, op.Thing)
	case Labels:
		return labelsText(*a.Labels)
	case ConfigWritten:
		return configWrittenText(*a.Written)
	case Steps:
		return stepsText(*a.Task)
	case Deleted:
		said := fmt.Sprintf("%s %d was deleted\n", op.Thing, *a.Deleted)
		for _, n := range a.Also {
			said += fmt.Sprintf("%s %d was deleted\n", op.Thing, n)
		}
		for _, n := range a.Unblocked {
			said += fmt.Sprintf("task %d was waiting on it, and nothing names what stops it now\n", n)
		}
		return said
	case Restored:
		return fmt.Sprintf("backup %d was restored\n", *a.Restored)
	case PluginWritten:
		return pluginText(*a.Plugin)
	case Exported:
		return *a.Exported + " " + domain.HeldText(*a.Holds) + "\n"
	case Imported:
		return fmt.Sprintf("%d task(s) and %d objective(s) were added from %s\n", a.Holds.Tasks, a.Holds.Objectives, *a.Imported)
	case Applied:
		return appliedText(a.Applied)
	case Acts:
		if a.Since != nil {
			return trace.List(a.Since)
		}
		return trace.Text(a.Acts)
	}
	return ""
}

func cardText(a Answer) string {
	return task.Text(*a.Task, a.Progress) + task.StepsText(factsOf(*a.Task).Steps) + stoodText(*a.Task, a.StoodSince) + contextText(a.Context)
}

func stoodText(t task.Task, since time.Time) string {
	if since.IsZero() {
		return ""
	}
	return fmt.Sprintf("\nin %s since %s · %s\n", factsOf(t).State, domain.FormatDay(since.Local()), daysSince(since, time.Now()))
}

func daysSince(since, now time.Time) string {
	from, to := since.Local(), now.Local()
	first := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.Local)
	last := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.Local)
	days := int(last.Sub(first).Hours()+12) / 24
	if days == 0 {
		return "today"
	}
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}

func heldText(held []task.Task, a Answer) string {
	if len(held) == 0 {
		return "nothing is in focus\n"
	}
	return task.Lines(held, marksOf(a))
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
	return task.Marks{InFocus: heldOr(a.InFocus), NothingBlocks: a.NothingBlocks}
}

func linesOf(tasks []task.Task) []string {
	out := make([]string, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, task.Line(t))
	}
	return out
}

func objectivesText(all []objective.Objective, along map[int]domain.Progress) string {
	if len(all) == 0 {
		return "there is no objective\n"
	}
	return objective.List(all, along)
}

func backupLine(b domain.Backup, holds domain.Held) string {
	return fmt.Sprintf("#%d  made %s  %s\n", b.Number, domain.FormatMoment(b.Made), domain.HeldText(holds))
}

func backupsText(all []domain.Backup) string {
	if len(all) == 0 {
		return "there is no backup\n"
	}
	var out strings.Builder
	for _, one := range all {
		fmt.Fprintf(&out, "#%d  made %s\n", one.Number, domain.FormatMoment(one.Made))
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
