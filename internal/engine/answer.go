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
	Added      *Sets                 `json:"added,omitempty" jsonschema:"what the change put on it; left out when it put nothing"`
	TakenOff   *Sets                 `json:"taken_off,omitempty" jsonschema:"what the change took off it; left out when it took nothing"`
	Warnings   []string              `json:"warnings,omitempty" jsonschema:"what the act did and should be known: each objective it counts in waiting on one not reached, as work starts on it; left out when there is nothing to say"`
}

// OfTask turns a task and how far along it is into what crosses to a machine.
func OfTask(t task.Task, p domain.Progress) Task {
	return Task{Shape: task.OfTask(t), Progress: domain.OfProgress(p)}
}

// Row is a task as a listing carries it: its facts, and whether it says anything — not what it says; only the fields asked for, when some were.
type Row struct {
	task.Facts
	Described bool `json:"described"`
	kept      []string
}

func rowOf(t task.Task, kept []string) Row {
	shape := task.OfTask(t)
	return Row{Facts: shape.Facts, Described: shape.Description != "", kept: kept}
}

// Tasks is how a listing of tasks crosses: an object naming what it carries, never a bare list.
type Tasks struct {
	Tasks    []Row            `json:"tasks"`
	InFocus  []int            `json:"in_focus"`
	Crossed  *Crossed         `json:"crossed,omitempty" jsonschema:"the column a move took a task out of and into; left out when it stayed where it was"`
	Matched  map[string][]int `json:"matched,omitempty" jsonschema:"each thing a search for any of it named, and the numbers of the tasks it matched — none when it matched nothing; given a page, only the tasks on it; left out unless any was asked"`
	Unknown  []int            `json:"unknown,omitempty" jsonschema:"each number a listing was asked for that no task carries, in the order asked; left out when every one is a task"`
	Total    *int             `json:"total,omitempty" jsonschema:"how many tasks the listing or the search kept before a page was cut from them, or before they were counted; left out unless a limit, an offset or a count was asked"`
	Counts   *map[string]int  `json:"counts,omitempty" jsonschema:"each column, or each label, as count asked, and how many of the tasks kept sit in it or carry it; one no task has is left out; left out unless count was asked"`
	Warnings []string         `json:"warnings,omitempty" jsonschema:"what the move did and should be known: each objective a task moved into current counts in, waiting on one not reached; left out when there is nothing to say"`
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
	Deleted            bool   `json:"deleted,omitempty" jsonschema:"the task was deleted, keeping its time; its title is the one it went with"`
}

func ofSheet(s trace.Sheet, titles map[int]string, gone map[int]bool) Timesheet {
	days := make([]TimesheetDay, 0, len(s.Days))
	for _, d := range s.Days {
		day := TimesheetDay{Day: domain.FormatDay(d.Day), Minutes: int(d.Total.Minutes()), Tasks: []TimesheetTask{}}
		for _, w := range d.Tasks {
			day.Tasks = append(day.Tasks, TimesheetTask{Number: w.Task, Title: titles[w.Task],
				Minutes: int(w.Took.Minutes()), LeftOpenAtMidnight: w.LeftOpen, Deleted: gone[w.Task]})
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
	Tasks    []Task         `json:"tasks"`
	InFocus  []int          `json:"in_focus"`
	Sessions []FocusSession `json:"sessions"`
	Warnings []string       `json:"warnings,omitempty" jsonschema:"what the focus did and should be known: each objective the task brought into current counts in, waiting on one not reached; left out when there is nothing to say"`
}

// Objective is how an objective crosses to a machine.
type Objective struct {
	objective.Shape
	Tasks      []int                 `json:"tasks"`
	Progress   *domain.ProgressShape `json:"progress"`
	TimeSpent  string                `json:"time_spent" jsonschema:"how long has been spent on the tasks serving it, as 1h30m; empty when nothing has"`
	Left       []int                 `json:"left" jsonschema:"the numbers of the tasks serving it that are neither done nor archived"`
	Parent     *int                  `json:"parent" jsonschema:"the number of the objective it sits under; null when it stands on its own"`
	Children   []int                 `json:"children" jsonschema:"the numbers of its sub-objectives; empty when it holds none"`
	WaitsOn    []int                 `json:"waits_on" jsonschema:"the numbers of the objectives it waits on, which warn and never block; empty when it waits on none"`
	WaitedOnBy []int                 `json:"waited_on_by" jsonschema:"the numbers of the objectives that wait on it; empty when none does"`
	Proposed   *int                  `json:"proposed_reached,omitempty" jsonschema:"the number of an objective every sub-objective of which is reached, so it may be put reached too — itself on show, its parent on the act that reached the last one; left out otherwise"`
	Added      *Sets                 `json:"added,omitempty" jsonschema:"what the change put on it; left out when it put nothing"`
	TakenOff   *Sets                 `json:"taken_off,omitempty" jsonschema:"what the change took off it; left out when it took nothing"`
}

// OfObjective turns an objective into what crosses.
func OfObjective(o objective.Objective, under []task.Task, p domain.Progress) Objective {
	numbers := make([]int, 0, len(under))
	for _, t := range under {
		numbers = append(numbers, factsOf(t).Number)
	}
	return Objective{
		Shape:      objective.OfObjective(o),
		Tasks:      numbers,
		Progress:   domain.OfProgress(p),
		TimeSpent:  domain.FormatDuration(reviewOf(under).Spent),
		Left:       leftOf(under),
		Children:   []int{},
		WaitsOn:    []int{},
		WaitedOnBy: []int{},
	}
}

func seated(said Objective, o objective.Objective, a Answer) Objective {
	c := o.Carried()
	if c.Parent != 0 {
		said.Parent = &c.Parent
	}
	said.Children = listed(a.Tree[c.Number])
	said.WaitsOn = listed(c.WaitsOn)
	said.WaitedOnBy = listed(a.Waiters[c.Number])
	return said
}

func proposedIn(a Answer) *int {
	if a.Proposes != nil {
		n := a.Proposes.Carried().Number
		return &n
	}
	if a.Family != nil && a.Family.Proposed {
		n := a.Objective.Carried().Number
		return &n
	}
	return nil
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
	Deleted       int   `json:"deleted"`
	Also          []int `json:"also,omitempty"`
	WasBlocking   []int `json:"was_blocking,omitempty"`
	Kept          []int `json:"kept,omitempty" jsonschema:"the sub-objectives kept, each standing on its own now; left out when none was"`
	SubObjectives []int `json:"sub_objectives_deleted,omitempty" jsonschema:"the sub-objectives deleted with it; left out when none was"`
	Detached      []int `json:"detached,omitempty" jsonschema:"the tasks detached from what was deleted; left out when none was"`
	TasksDeleted  []int `json:"tasks_deleted,omitempty" jsonschema:"the tasks deleted with it; left out when none was"`
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
	case Workspaces:
		return *a.Workspaces
	case WorkspaceAdded:
		return *a.WorkspaceAdded
	case WorkspaceUsed:
		return *a.WorkspaceUsed
	case WorkspaceRemoved:
		return *a.WorkspaceRemoved
	case WorkspaceShown:
		return *a.WorkspaceShown
	case Timesheet:
		return ofSheet(*a.Sheet, a.Named, a.Gone)
	case Review:
		r := OfReview(*a.Objective, a.Under, a.Progress, *a.Period)
		r.Objective = seated(r.Objective, *a.Objective, a)
		return r
	case Focus:
		held := a.Held
		if a.Task != nil {
			held = []task.Task{*a.Task}
		}
		said := ofFocus(held, a.Progress, heldOr(a.InFocus), a.Sessions)
		said.Warnings = a.Warnings
		return said
	case Task:
		said := OfTask(*a.Task, a.Progress)
		said.Context = a.Context
		if !a.StoodSince.IsZero() {
			said.StoodSince = &a.StoodSince
		}
		said.Added, said.TakenOff = a.Moved.added(), a.Moved.takenOff()
		said.Warnings = a.Warnings
		return said
	case Swept:
		swept := Swept{Archived: *a.Archived}
		if a.Task != nil {
			one := OfTask(*a.Task, a.Progress)
			swept.Task = &one
		}
		return swept
	case Tasks:
		var said Tasks
		if a.Matched != nil {
			said = ofTasks(a.Matched, a.InFocus, a.Fields)
			said.Matched, said.Unknown = a.Terms, a.Unknown
		} else {
			said = ofTasks(a.Tasks, a.InFocus, a.Fields)
			said.Crossed = a.Crossed
		}
		said.Total, said.Counts, said.Warnings = a.Total, ofCounts(a.Counts), a.Warnings
		return said
	case Objective:
		said := seated(OfObjective(*a.Objective, a.Under, a.Progress), *a.Objective, a)
		said.Added, said.TakenOff = a.Moved.added(), a.Moved.takenOff()
		said.Proposed = proposedIn(a)
		return said
	case Objectives:
		if a.MatchedObjectives != nil {
			return Objectives{Objectives: ofObjectives(a.MatchedObjectives, a)}
		}
		return Objectives{Objectives: ofObjectives(a.Objectives, a)}
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
	case ConfigWritten:
		return ConfigWritten{Written: *a.Written}
	case DebugReported:
		return DebugReported{Written: *a.Written, Files: *a.Packed}
	case Steps:
		return ofSteps(*a.Task)
	case Deleted:
		said := Deleted{Deleted: *a.Deleted, Also: a.Also, WasBlocking: a.Unblocked}
		if a.Went != nil {
			said.Kept, said.SubObjectives, said.Detached, said.TasksDeleted = a.Went.Kept, a.Went.Gone, a.Went.Detached, a.Went.Deleted
		}
		return said
	case Restored:
		return Restored{Restored: *a.Restored}
	case PluginWritten:
		return PluginWritten{Plugin: *a.Plugin}
	case AppInstalled:
		return AppInstalled{Installed: *a.Installed, Runs: *a.Runs, OpenedFrom: *a.OpenedFrom}
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

func ofFocus(held []task.Task, p domain.Progress, numbers []int, sessions []FocusSession) Focus {
	out := make([]Task, 0, len(held))
	for _, t := range held {
		out = append(out, OfTask(t, p))
		p = domain.Progress{}
	}
	if sessions == nil {
		sessions = []FocusSession{}
	}
	return Focus{Tasks: out, InFocus: listed(numbers), Sessions: sessions}
}

func ofTasks(tasks []task.Task, inFocus *[]int, kept []string) Tasks {
	out := make([]Row, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, rowOf(t, kept))
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

func ofObjectives(all []objective.Objective, a Answer) []Objective {
	out := make([]Objective, 0, len(all))
	for _, o := range all {
		said := seated(OfObjective(o, nil, a.Along[o.Carried().Number]), o, a)
		said.TimeSpent = domain.FormatDuration(a.Spent[o.Carried().Number])
		out = append(out, said)
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
