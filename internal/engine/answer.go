package engine

import (
	"time"

	"cairn/internal/domain"
	"cairn/internal/migration"
	"cairn/internal/objective"
	"cairn/internal/task"
	"cairn/internal/trace"
)

// Task is how a task crosses to a machine: every fact task.Shape carries, plus how far along it is.
type Task struct {
	task.Shape
	Progress   *domain.ProgressShape `json:"progress"`
	Context    string                `json:"context,omitempty"`
	StoodSince *time.Time            `json:"in_column_since,omitempty"`
	SubTasks   []SubTask             `json:"sub_tasks,omitempty"`
}

// SubTask is a part of the work a task names, as it crosses to a machine.
type SubTask struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
}

// OfTask turns a task and how far along it is into what crosses to a machine.
func OfTask(t task.Task, p domain.Progress) Task {
	return Task{Shape: task.OfTask(t), Progress: domain.OfProgress(p)}
}

func ofTaskUnder(t task.Task, p domain.Progress, under []task.Task) Task {
	said := OfTask(t, p)
	for _, u := range under {
		shape := task.OfTask(u)
		said.SubTasks = append(said.SubTasks, SubTask{
			Number: shape.Number,
			Title:  shape.Title,
			State:  shape.State,
		})
	}
	return said
}

// Row is a task as a listing carries it: its facts, and whether it says anything — not what it says.
type Row struct {
	task.Facts
	Described bool `json:"described"`
}

func rowOf(t task.Task) Row {
	shape := task.OfTask(t)
	return Row{Facts: shape.Facts, Described: shape.Description != ""}
}

// Tasks is how a listing of tasks crosses: an object naming what it carries, never a bare list.
type Tasks struct {
	Tasks   []Row `json:"tasks"`
	InFocus []int `json:"in_focus"`
}

// Timesheet is how long the work took, as it crosses to a machine: every day something was done on, in order, with each task's minutes — a minute begun counted whole — and the day's, the sum of those.
type Timesheet struct {
	Days []TimesheetDay `json:"days"`
}

// TimesheetDay is one day of it.
type TimesheetDay struct {
	Day     string          `json:"day"`
	Minutes int             `json:"minutes"`
	Tasks   []TimesheetTask `json:"tasks"`
}

// TimesheetTask is one task on one day, and whether it was still in focus when the day ended — the time then stopped at midnight, not where the work did.
type TimesheetTask struct {
	Number             int    `json:"number"`
	Title              string `json:"title"`
	Minutes            int    `json:"minutes"`
	LeftOpenAtMidnight bool   `json:"left_open_at_midnight"`
}

func ofSheet(s trace.Sheet, titles map[int]string) Timesheet {
	days := make([]TimesheetDay, 0, len(s.Days))
	for _, d := range s.Days {
		day := TimesheetDay{Day: domain.FormatDay(d.Day), Minutes: int(d.Total.Minutes()), Tasks: []TimesheetTask{}}
		for _, w := range d.Tasks {
			day.Tasks = append(day.Tasks, TimesheetTask{Number: w.Task, Title: titles[w.Task],
				Minutes: int(w.Took.Minutes()), LeftOpenAtMidnight: w.LeftOpen})
		}
		days = append(days, day)
	}
	return Timesheet{Days: days}
}

// Objectives is how a listing of objectives crosses.
type Objectives struct {
	Objectives []Objective `json:"objectives"`
}

// Kept is how a listing of backups crosses — named for what is kept beside the store, since Backups already names the way to reach them.
type Kept struct {
	Backups []BackupShape `json:"backups"`
}

// Ledger is how a listing of changes to the shape of the store crosses, with what the store holds — which the text has always said and the JSON used to lose, having nowhere to put it.
type Ledger struct {
	Changes []migration.Shape `json:"changes"`
	Holds   *domain.Held      `json:"holds"`
}

// Acts is how what was done to a task, or over a period, crosses.
type Acts struct {
	Acts []trace.Shape `json:"acts"`
}

// Focus is how the task being worked on crosses, or the fact that none is.
type Focus struct {
	Tasks   []Task `json:"tasks"`
	InFocus []int  `json:"in_focus"`
}

// Objective is how an objective crosses to a machine.
type Objective struct {
	objective.Shape
	Tasks    []int                 `json:"tasks"`
	Progress *domain.ProgressShape `json:"progress"`
}

// OfObjective turns an objective into what crosses.
func OfObjective(o objective.Objective, under []task.Task, p domain.Progress) Objective {
	numbers := make([]int, 0, len(under))
	for _, t := range under {
		numbers = append(numbers, factsOf(t).Number)
	}
	return Objective{
		Shape:    objective.OfObjective(o),
		Tasks:    numbers,
		Progress: domain.OfProgress(p),
	}
}

// Labelled is how labelling answers: how many it put one on.
type Labelled struct {
	Labelled int `json:"labelled"`
}

// Swept is how archiving answers: how many were taken off the board and, when one was named by its number, that task as it ends up.
type Swept struct {
	Archived int   `json:"archived"`
	Task     *Task `json:"task,omitempty"`
}

// Deleted confirms an act on something numbered — a task deleted, an objective deleted, a backup deleted.
type Deleted struct {
	Deleted     int   `json:"deleted"`
	Also        []int `json:"also,omitempty"`
	WasBlocking []int `json:"was_blocking,omitempty"`
}

// Restored confirms a backup was restored, replacing the store with what it held.
type Restored struct {
	Restored int `json:"restored"`
}

// Applied names what a change command made — every name it applied, or none when nothing was waiting.
type Applied struct {
	Applied []string `json:"applied"`
}

// Of turns what an operation answered into the shape a machine reads — the shape the operation declared it answers, and no other.
func Of(op Operation, a Answer) any {
	switch op.Answers.(type) {
	case Config:
		return *a.Config
	case Timesheet:
		return ofSheet(*a.Sheet, a.Named)
	case Focus:
		held := a.Held
		if a.Task != nil {
			held = []task.Task{*a.Task}
		}
		return ofFocus(held, a.Progress, heldOr(a.InFocus))
	case Task:
		said := ofTaskUnder(*a.Task, a.Progress, a.Under)
		said.Context = a.Context
		if !a.StoodSince.IsZero() {
			said.StoodSince = &a.StoodSince
		}
		return said
	case Swept:
		swept := Swept{Archived: *a.Archived}
		if a.Task != nil {
			one := OfTask(*a.Task, a.Progress)
			swept.Task = &one
		}
		return swept
	case Tasks:
		if a.Matched != nil {
			return ofTasks(a.Matched, a.InFocus)
		}
		return ofTasks(a.Tasks, a.InFocus)
	case Objective:
		return OfObjective(*a.Objective, a.Under, a.Progress)
	case Objectives:
		if a.MatchedObjectives != nil {
			return Objectives{Objectives: ofObjectives(a.MatchedObjectives)}
		}
		return Objectives{Objectives: ofObjectives(a.Objectives)}
	case BackupShape:
		return OfBackup(*a.Backup, a.Holds)
	case Kept:
		return Kept{Backups: ofBackups(a.Backups)}
	case migration.Shape:
		return migration.OfMigration(*a.Change)
	case Ledger:
		return Ledger{Changes: ofChanges(a.Changes), Holds: a.Holds}
	case migration.RolledBack:
		return migration.OfRolledBack(*a.RolledBack)
	case Labelled:
		return Labelled{Labelled: *a.Labelled}
	case Labels:
		return Labels{Labels: *a.Labels}
	case Deleted:
		return Deleted{Deleted: *a.Deleted, Also: a.Also, WasBlocking: a.Unblocked}
	case Restored:
		return Restored{Restored: *a.Restored}
	case PluginWritten:
		return PluginWritten{Plugin: *a.Plugin}
	case Exported:
		return Exported{Exported: *a.Exported, Holds: a.Holds}
	case Imported:
		return Imported{Imported: *a.Imported, Added: *a.Holds}
	case Applied:
		return Applied{Applied: a.Applied}
	case Acts:
		if a.Since != nil {
			return Acts{Acts: ofActs(a.Since)}
		}
		return Acts{Acts: ofActs(a.Acts)}
	}
	return nil
}

func ofActs(acts []trace.Act) []trace.Shape {
	out := make([]trace.Shape, 0, len(acts))
	for _, one := range acts {
		out = append(out, trace.OfAct(one))
	}
	return out
}

func ofFocus(held []task.Task, p domain.Progress, numbers []int) Focus {
	out := make([]Task, 0, len(held))
	for _, t := range held {
		out = append(out, OfTask(t, p))
		p = domain.Progress{}
	}
	return Focus{Tasks: out, InFocus: listed(numbers)}
}

func ofTasks(tasks []task.Task, inFocus *[]int) Tasks {
	out := make([]Row, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, rowOf(t))
	}
	if inFocus == nil {
		return Tasks{Tasks: out}
	}
	return Tasks{Tasks: out, InFocus: listed(*inFocus)}
}

func heldOr(inFocus *[]int) []int {
	if inFocus == nil {
		return nil
	}
	return *inFocus
}

func listed(numbers []int) []int {
	if numbers == nil {
		return []int{}
	}
	return numbers
}

func ofObjectives(all []objective.Objective) []Objective {
	out := make([]Objective, 0, len(all))
	for _, o := range all {
		out = append(out, OfObjective(o, nil, domain.Progress{}))
	}
	return out
}

func ofBackups(all []domain.Backup) []BackupShape {
	out := make([]BackupShape, 0, len(all))
	for _, b := range all {
		out = append(out, OfBackup(b, nil))
	}
	return out
}

func ofChanges(all []migration.Migration) []migration.Shape {
	out := make([]migration.Shape, 0, len(all))
	for _, c := range all {
		out = append(out, migration.OfMigration(c))
	}
	return out
}

// BackupShape is how a backup crosses to a machine.
type BackupShape struct {
	Number int          `json:"number"`
	Made   string       `json:"made"`
	Holds  *domain.Held `json:"holds,omitempty"`
}

// OfBackup turns a backup into what crosses, holding what it held when that was counted, and nothing when it was not: making a backup does not open it, and a zero nobody measured would read as an empty store.
func OfBackup(b domain.Backup, held *domain.Held) BackupShape {
	out := BackupShape{Number: b.Number, Made: domain.FormatMoment(b.Made)}
	if held != nil {
		out.Holds = &domain.Held{Tasks: held.Tasks, Objectives: held.Objectives}
	}
	return out
}
