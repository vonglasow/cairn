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

// What an operation answered, as text for a person to read: the same answer
// a machine reads as JSON, written the way the command line prints it.

// contextText is the line of a card that says where a task's outside context
// lives, apart from the facts above it: the place is not something the task
// carries but somewhere the task's material is kept.
func contextText(at string) string {
	if at == "" {
		return ""
	}
	return "\ncontext  " + at + "\n"
}

// RenderText turns an answer into what a person reads: the rendering of the
// shape the operation declared it answers, the way Of builds the JSON. The
// operation also names what Deleted counts — a task, an objective, a backup.
//
// It is shown, never obeyed: what was written before cairn refused escapes,
// or comes back with an older backup, reaches a terminal with every rune it
// would obey replaced (domain.Shown).
func RenderText(op Operation, a Answer) string {
	return domain.Shown(rendered(op, a))
}

// rendered is the text of an answer, as it was kept.
func rendered(op Operation, a Answer) string {
	switch op.Answers.(type) {
	case Config:
		return configText(*a.Config)
	case Timesheet:
		return trace.SheetText(*a.Sheet, a.Named)
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
		return objective.Text(*a.Objective, a.Progress) + objective.Serving(linesOf(a.Under))
	case Objectives:
		if a.MatchedObjectives != nil {
			return objectivesText(a.MatchedObjectives)
		}
		return objectivesText(a.Objectives)
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
	case Deleted:
		said := fmt.Sprintf("%s %d was deleted\n", op.Thing, *a.Deleted)
		for _, n := range a.Unblocked {
			said += fmt.Sprintf("task %d was waiting on it, and nothing names what stops it now\n", n)
		}
		return said
	case PutBack:
		return fmt.Sprintf("backup %d was put back\n", *a.PutBack)
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

// cardText is one task in full, what hangs under it, and where its context
// lives.
func cardText(a Answer) string {
	return task.Text(*a.Task, a.Progress) + task.Under(a.Under) + stoodText(*a.Task, a.StoodSince) + contextText(a.Context)
}

// stoodText says since when a task has stood in its column, and for how many
// days that is, when it is the task asked about.
func stoodText(t task.Task, since time.Time) string {
	if since.IsZero() {
		return ""
	}
	return fmt.Sprintf("\nin %s since %s · %s\n", factsOf(t).state, domain.FormatDay(since), daysSince(since, time.Now()))
}

// daysSince is how many days lie between a moment and now, counted as a
// person counts them: by the days on the calendar, not by the hours.
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

// heldText renders the tasks being worked on, saying so plainly when there are
// none. Rendered as a listing rather than as one card each: several can be
// held, and a dozen cards one after another is a page nobody reads.
func heldText(held []task.Task, a Answer) string {
	if len(held) == 0 {
		return "nothing is in focus\n"
	}
	return task.Tree(held, marksOf(a))
}

// tasksText renders every task, saying so when there is none.
func tasksText(tasks []task.Task, a Answer) string {
	if len(tasks) == 0 {
		return "there is no task\n"
	}
	return task.Tree(tasks, marksOf(a))
}

// foundText renders every task that matched, flat — a sub-task that matched
// hangs under nothing, since what matched is what is shown rather than where it
// sits — saying so when nothing matched.
func foundText(tasks []task.Task, a Answer) string {
	if len(tasks) == 0 {
		return "there is no task\n"
	}
	return task.Lines(tasks, marksOf(a))
}

// marksOf reads what a listing has to say about its own rows. Both marks come
// from the store rather than from any task, so both are carried on the answer
// and neither is asked of a task.
func marksOf(a Answer) task.Marks {
	return task.Marks{InFocus: heldOr(a.InFocus), NothingBlocks: a.NothingBlocks}
}

// linesOf renders every task as one listing line each, for objective's
// serving to lay out: an objective knows nothing of a task, so this builds
// the lines rather than handing tasks across that boundary.
func linesOf(tasks []task.Task) []string {
	out := make([]string, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, task.Line(t))
	}
	return out
}

// objectivesText renders every objective, saying so when there is none.
func objectivesText(all []objective.Objective) string {
	if len(all) == 0 {
		return "there is no objective\n"
	}
	return objective.List(all)
}

// backupLine renders one backup: its number, when it was made, and what it
// holds.
func backupLine(b domain.Backup, holds domain.Held) string {
	return fmt.Sprintf("#%d  made %s  %s\n", b.Number, domain.FormatMoment(b.Made), domain.HeldText(holds))
}

// backupsText renders every backup, saying so when there is none.
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

// appliedText names what a change command made — every name it applied, or
// says that nothing was waiting when none was.
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
