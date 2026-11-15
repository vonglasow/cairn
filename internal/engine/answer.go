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
	"time"

	"cairn/internal/engine/act"
	"cairn/internal/engine/migration"
	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
)

// Task is how a task crosses to a machine: every fact task.Shape carries, plus how far along it is.
type Task struct {
	task.Shape
	Progress   *item.ProgressShape `json:"progress"`
	Context    string              `json:"context,omitempty"`
	StoodSince *time.Time          `json:"in_column_since,omitempty"`
	Blocking   []int               `json:"blocking,omitempty" jsonschema:"the numbers of the open tasks naming it as a blocker; left out when none does"`
	Added      *sets               `json:"added,omitempty" jsonschema:"what the change put on it; left out when it put nothing"`
	TakenOff   *sets               `json:"taken_off,omitempty" jsonschema:"what the change took off it; left out when it took nothing"`
	Warnings   []string            `json:"warnings,omitempty" jsonschema:"what the act did and should be known: each objective it counts in waiting on one not reached, as work starts on it; left out when there is nothing to say"`
	Series     *series.Occurs      `json:"series,omitempty" jsonschema:"the series it is an occurrence of, while series are on; left out otherwise"`
	Made       []series.MadeNext   `json:"made,omitempty" jsonschema:"each occurrence a series made because this act left it with none open, after the days it missed; left out when none was"`
}

// ofTask turns a task and how far along it is into what crosses to a machine.
func ofTask(t task.Task, p item.Progress) Task {
	return Task{Shape: task.OfTask(t), Progress: item.OfProgress(p)}
}

// Row is a task as a listing carries it: its facts, and whether it says anything — not what it says; only the fields asked for, when some were.
type Row struct {
	task.Facts
	Described     bool `json:"described"`
	Series        int  `json:"series" jsonschema:"the number of the series it is an occurrence of; 0 when it is none, or series are off"`
	Blocks        int  `json:"blocks" jsonschema:"how many open tasks name it as a blocker"`
	Overdue       *int `json:"overdue" jsonschema:"how many days past its deadline it is; null when it has none, is done, or is not past it"`
	DueIn         *int `json:"due_in" jsonschema:"how many days are left before its deadline, 0 on the day; null when it has none, is done, or is past it"`
	Age           int  `json:"age" jsonschema:"how many days it has sat in its column"`
	ObjectiveLate bool `json:"objective_late" jsonschema:"true when an objective it serves is past its deadline, neither reached nor dropped"`
	kept          []string
}

func rowOf(t task.Task, kept []string) Row {
	shape := task.OfTask(t)
	return Row{Facts: shape.Facts, Described: shape.Description != "", kept: kept}
}

// Tasks is how a listing of tasks crosses: an object naming what it carries, never a bare list.
type Tasks struct {
	Tasks    []Row             `json:"tasks"`
	InFocus  []int             `json:"in_focus"`
	Crossed  *task.Crossed     `json:"crossed,omitempty" jsonschema:"the column a move took a task out of and into; left out when it stayed where it was"`
	Matched  map[string][]int  `json:"matched,omitempty" jsonschema:"each thing a search for any of it named, and the numbers of the tasks it matched — none when it matched nothing; given a page, only the tasks on it; left out unless any was asked"`
	Unknown  []int             `json:"unknown,omitempty" jsonschema:"each number a listing was asked for that no task carries, in the order asked; left out when every one is a task"`
	Total    *int              `json:"total,omitempty" jsonschema:"how many tasks the listing or the search kept before a page was cut from them, or before they were counted; left out unless a limit, an offset or a count was asked"`
	Counts   *map[string]int   `json:"counts,omitempty" jsonschema:"each column, or each label, as count asked, and how many of the tasks kept sit in it or carry it; one no task has is left out; left out unless count was asked"`
	Warnings []string          `json:"warnings,omitempty" jsonschema:"what the move did and should be known: each objective a task moved into current counts in, waiting on one not reached; left out when there is nothing to say"`
	View     string            `json:"view,omitempty" jsonschema:"the view whose filters kept the tasks; left out unless one was named"`
	Gone     []string          `json:"gone,omitempty" jsonschema:"each condition of the view whose objective or label is gone, which narrows nothing; when every one is, the view keeps nothing; left out when none is"`
	Outside  []view.Outside    `json:"outside,omitempty" jsonschema:"the tasks outside the view that block, wait on or are linked to one listed, each once with every way it does, never counted; left out unless a view was named"`
	Made     []series.MadeNext `json:"made,omitempty" jsonschema:"each occurrence a series made because this act left it with none open, after the days it missed; left out when none was"`
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

func ofSheet(s act.Sheet, titles map[int]string, gone map[int]bool) Timesheet {
	days := make([]TimesheetDay, 0, len(s.Days))
	for _, d := range s.Days {
		day := TimesheetDay{Day: calendar.FormatDay(d.Day), Minutes: int(d.Total.Minutes()), Tasks: []TimesheetTask{}}
		for _, w := range d.Tasks {
			day.Tasks = append(day.Tasks, TimesheetTask{Number: w.Task, Title: titles[w.Task],
				Minutes: int(w.Took.Minutes()), LeftOpenAtMidnight: w.LeftOpen, Deleted: gone[w.Task]})
		}
		days = append(days, day)
	}
	return Timesheet{Days: days}
}

// objectivesListed is how a listing of objectives crosses.
type objectivesListed struct {
	Objectives []objectiveRow `json:"objectives"`
}

// kept is how a listing of backups crosses — named for what is kept beside the store, since Backups already names the way to reach them.
type kept struct {
	Backups []BackupShape `json:"backups"`
}

// ledger is how a listing of changes to the shape of the store crosses, with what the store holds — which the text has always said and the JSON used to lose, having nowhere to put it.
type ledger struct {
	Changes []migration.Shape `json:"changes"`
	Holds   *backup.Held      `json:"holds"`
}

// acts is how what was done to a task, or over a period, crosses.
type acts struct {
	Acts []act.Shape `json:"acts"`
}

// Focus is how the task being worked on crosses, or the fact that none is.
type Focus struct {
	Tasks    []Task         `json:"tasks"`
	InFocus  []int          `json:"in_focus"`
	Sessions []focusSession `json:"sessions"`
	Warnings []string       `json:"warnings,omitempty" jsonschema:"what the focus did and should be known: each objective the task brought into current counts in, waiting on one not reached; left out when there is nothing to say"`
}

// objectiveRow is how an objective crosses to a machine.
type objectiveRow struct {
	objective.Shape
	Tasks      []int               `json:"tasks"`
	Progress   *item.ProgressShape `json:"progress"`
	TimeSpent  string              `json:"time_spent" jsonschema:"how long has been spent on the tasks serving it, as 1h30m; empty when nothing has"`
	Left       []int               `json:"left" jsonschema:"the numbers of the tasks serving it that are neither done nor archived"`
	Parent     *int                `json:"parent" jsonschema:"the number of the objective it sits under; null when it stands on its own"`
	Children   []int               `json:"children" jsonschema:"the numbers of its sub-objectives; empty when it holds none"`
	WaitsOn    []int               `json:"waits_on" jsonschema:"the numbers of the objectives it waits on, which warn and never block; empty when it waits on none"`
	WaitedOnBy []int               `json:"waited_on_by" jsonschema:"the numbers of the objectives that wait on it; empty when none does"`
	Proposed   *int                `json:"proposed_reached,omitempty" jsonschema:"the number of an objective every sub-objective of which is reached, so it may be put reached too — itself on show, its parent on the act that reached the last one; left out otherwise"`
	Added      *sets               `json:"added,omitempty" jsonschema:"what the change put on it; left out when it put nothing"`
	TakenOff   *sets               `json:"taken_off,omitempty" jsonschema:"what the change took off it; left out when it took nothing"`
}

// ofObjective turns an objective into what crosses.
func ofObjective(o objective.Objective, under []task.Task, p item.Progress) objectiveRow {
	numbers := make([]int, 0, len(under))
	for _, t := range under {
		numbers = append(numbers, t.Carried().Number)
	}
	return objectiveRow{
		Shape:      objective.OfObjective(o),
		Tasks:      numbers,
		Progress:   item.OfProgress(p),
		TimeSpent:  calendar.FormatDuration(reviewOf(under).Spent),
		Left:       leftOf(under),
		Children:   []int{},
		WaitsOn:    []int{},
		WaitedOnBy: []int{},
	}
}

func seated(said objectiveRow, o objective.Objective, a Answer) objectiveRow {
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

// labelled is how labelling answers: how many it put one on.
type labelled struct {
	Labelled int `json:"labelled"`
}

// swept is how archiving answers: how many were taken off the board and, when one was named by its number, that task as it ends up.
type swept struct {
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

// restored confirms a backup was restored, replacing the store with what it held.
type restored struct {
	Restored int `json:"restored"`
}

// applied names what a change command made — every name it applied, or none when nothing was waiting.
type applied struct {
	Applied []string `json:"applied"`
}

// Of turns what an operation answered into the shape a machine reads — the shape the operation declared it answers, and no other.
func Of(op Operation, a Answer) any {
	switch op.Answers.(type) {
	case workspace.ConfigChanged:
		return *a.ConfigChanged
	case workspace.Config:
		return *a.Config
	case workspace.Workspaces:
		return *a.Workspaces
	case workspace.Added:
		return *a.WorkspaceAdded
	case workspace.Used:
		return *a.WorkspaceUsed
	case workspace.Removed:
		return *a.WorkspaceRemoved
	case workspace.Shown:
		return *a.WorkspaceShown
	case Timesheet:
		return ofSheet(*a.Sheet, a.Named, a.Gone)
	case review:
		r := ofReview(*a.Objective, a.Under, a.Progress, *a.Period)
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
		said := ofTask(*a.Task, a.Progress)
		said.Context = a.Context
		if !a.StoodSince.IsZero() {
			said.StoodSince = &a.StoodSince
		}
		said.Blocking = a.Choosing[said.Number].Blocks
		said.Added, said.TakenOff = a.Moved.added(), a.Moved.takenOff()
		said.Warnings = a.Warnings
		if o, is := a.Occurring[said.Number]; is {
			said.Series = &o
		}
		said.Made = a.Followed
		return said
	case swept:
		swept := swept{Archived: *a.Archived}
		if a.Task != nil {
			one := ofTask(*a.Task, a.Progress)
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
		said.View, said.Gone, said.Outside = a.InView, a.Lost, a.Outside
		for i, row := range said.Tasks {
			said.Tasks[i].Series = a.Occurring[row.Number].Series
			c := a.Choosing[row.Number]
			said.Tasks[i].Blocks, said.Tasks[i].Overdue, said.Tasks[i].DueIn = len(c.Blocks), c.Overdue, c.DueIn
			said.Tasks[i].Age, said.Tasks[i].ObjectiveLate = c.Age, c.ObjectiveLate
		}
		said.Made = a.Followed
		return said
	case objectiveRow:
		said := seated(ofObjective(*a.Objective, a.Under, a.Progress), *a.Objective, a)
		said.Added, said.TakenOff = a.Moved.added(), a.Moved.takenOff()
		said.Proposed = proposedIn(a)
		return said
	case objectivesListed:
		if a.MatchedObjectives != nil {
			return objectivesListed{Objectives: ofObjectives(a.MatchedObjectives, a)}
		}
		return objectivesListed{Objectives: ofObjectives(a.Objectives, a)}
	case BackupShape:
		return ofBackup(*a.Backup, a.Holds)
	case kept:
		return kept{Backups: ofBackups(a.Backups)}
	case migration.Shape:
		return migration.OfMigration(*a.Change)
	case ledger:
		return ledger{Changes: ofChanges(a.Changes), Holds: a.Holds}
	case migration.RolledBack:
		return migration.OfRolledBack(*a.RolledBack)
	case labelled:
		return labelled{Labelled: *a.Labelled}
	case label.List:
		return label.List{Labels: *a.Labels, Warnings: a.Warnings}
	case View:
		return *a.Viewed
	case Views:
		return *a.ViewsKept
	case view.Deleted:
		return view.Deleted{Deleted: *a.ViewGone}
	case Series:
		return *a.SeriesShown
	case seriesListed:
		return *a.SeriesKept
	case series.Deleted:
		return *a.SeriesGone
	case workspace.ConfigWritten:
		return workspace.ConfigWritten{Written: *a.Written}
	case debugReported:
		return debugReported{Written: *a.Written, Files: *a.Packed}
	case task.Steps:
		return task.OfSteps(*a.Task)
	case Deleted:
		said := Deleted{Deleted: *a.Deleted, Also: a.Also, WasBlocking: a.Unblocked}
		if a.Went != nil {
			said.Kept, said.SubObjectives, said.Detached, said.TasksDeleted = a.Went.Kept, a.Went.Gone, a.Went.Detached, a.Went.Deleted
		}
		return said
	case restored:
		return restored{Restored: *a.Restored}
	case app.PluginWritten:
		return app.PluginWritten{Plugin: *a.Plugin}
	case app.PluginShown:
		return *a.Install
	case app.Upgraded:
		return *a.Upgraded
	case app.Installed:
		return app.Installed{Installed: *a.Installed, Runs: *a.Runs, OpenedFrom: *a.OpenedFrom}
	case exchange.Exported:
		return exchange.Exported{Exported: *a.Exported, Holds: a.Holds}
	case exchange.ImportedFile:
		return exchange.ImportedFile{Imported: *a.Imported, Added: *a.Holds}
	case applied:
		return applied{Applied: a.Applied}
	case acts:
		if a.Since != nil {
			return acts{Acts: ofActs(a.Since)}
		}
		return acts{Acts: ofActs(a.Acts)}
	}
	return nil
}

func ofActs(acts []act.Act) []act.Shape {
	out := make([]act.Shape, 0, len(acts))
	for _, one := range acts {
		out = append(out, act.OfAct(one))
	}
	return out
}

func ofFocus(held []task.Task, p item.Progress, numbers []int, sessions []focusSession) Focus {
	out := make([]Task, 0, len(held))
	for _, t := range held {
		out = append(out, ofTask(t, p))
		p = item.Progress{}
	}
	if sessions == nil {
		sessions = []focusSession{}
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

func listed[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}

func ofObjectives(all []objective.Objective, a Answer) []objectiveRow {
	out := make([]objectiveRow, 0, len(all))
	for _, o := range all {
		said := seated(ofObjective(o, nil, a.Along[o.Carried().Number]), o, a)
		said.TimeSpent = calendar.FormatDuration(a.Spent[o.Carried().Number])
		out = append(out, said)
	}
	return out
}

func ofBackups(all []backup.Backup) []BackupShape {
	out := make([]BackupShape, 0, len(all))
	for _, b := range all {
		out = append(out, ofBackup(b, nil))
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
	Holds  *backup.Held `json:"holds,omitempty"`
}

// ofBackup turns a backup into what crosses, holding what it held when that was counted, and nothing when it was not: making a backup does not open it, and a zero nobody measured would read as an empty store.
func ofBackup(b backup.Backup, held *backup.Held) BackupShape {
	out := BackupShape{Number: b.Number, Made: calendar.FormatMoment(b.Made)}
	if held != nil {
		out.Holds = &backup.Held{Tasks: held.Tasks, Objectives: held.Objectives}
	}
	return out
}
