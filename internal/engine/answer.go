package engine

// The two renderings of an Answer, kept in one place so neither surface
// carries its own copy and the two can never drift apart: the shape a machine
// reads — the JSON `--json` prints on the command line and the shape an MCP
// tool answers with — and the card a person reads. Both live here because
// every one of these is about an Answer, which is the engine's.

import (
	"time"

	"cairn/internal/domain"
	"cairn/internal/migration"
	"cairn/internal/objective"
	"cairn/internal/task"
	"cairn/internal/trace"
)

// Task is how a task crosses to a machine: every fact task.Shape carries, plus
// how far along it is. Progress sits here rather than in task.Shape itself:
// task and objective share the notion and it belongs to neither.
//
// Which task is being worked on is not here. It is a fact of the store, not of
// a task, and a listing says it once at the top rather than once per row — said
// on every row, the rows can disagree and a reader has to pick one to believe.
// It also makes the two renderings agree: the card a person reads has never
// shown it either, only the listing marks it.
type Task struct {
	task.Shape
	Progress *domain.ProgressShape `json:"progress"`
	// Context is where what was written about this task elsewhere lives, said
	// when it is the task asked about and the caller said where the notes are.
	// A harness writes there and reads there; cairn only says where.
	Context string `json:"context,omitempty"`
	// StoodSince is the moment it last changed column, or was made: how long
	// it has sat where it is. Said when it is the task asked about.
	StoodSince *time.Time `json:"in_column_since,omitempty"`
	// SubTasks is what hangs under this one, and only when it is what was
	// asked about. A listing carries rows, not this — see Row.
	SubTasks []SubTask `json:"sub_tasks,omitempty"`
}

// SubTask is a part of the work a task names, as it crosses to a machine.
//
// Its own shape rather than a Task again, and that is not only convenience: a
// sub-task cannot itself carry sub-tasks, and a Task holding Tasks would say
// it could. The type says the rule — and a self-referential shape is one no
// schema can describe, which is how this was found: eight MCP tools lost the
// shape of their answer the moment Task held a Task.
//
// Three facts and no more, because three are what a part is asked about: which
// one it is, what it is called, and whether it is done. Everything else about
// it is reached by its number.
type SubTask struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
}

// OfTask turns a task and how far along it is into what crosses to a machine.
func OfTask(t task.Task, p domain.Progress) Task {
	return Task{Shape: task.OfTask(t), Progress: domain.OfProgress(p)}
}

// ofTaskUnder is the same, plus what hangs under it.
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

// Row is a task as a listing carries it: its facts, and whether it says
// anything — not what it says. Measured on a real board, 105 tasks handed over
// with their descriptions came to 106 000 characters, past what a client hands
// an agent in one answer; the descriptions were seven tenths of it. What a
// task says is read by showing it, the way a person opens a card the board
// marks as described.
//
// No progress and nothing hanging under it either: a listing counts neither,
// and a field that is always null is a field every row pays for to say
// nothing. The type says it — a row cannot carry what it has no field for.
type Row struct {
	task.Facts
	Described bool `json:"described"`
}

// rowOf turns one task into a row of a listing.
func rowOf(t task.Task) Row {
	shape := task.OfTask(t)
	return Row{Facts: shape.Facts, Described: shape.Description != ""}
}

// Tasks is how a listing of tasks crosses: an object naming what it carries,
// never a bare list. So is every other listing below, for three reasons, and
// only the first is somebody else's rule. A client refuses a schema that does
// not describe an object, and one refused tool costs the whole list. A bare
// list can never gain a sibling field, so what a listing knows besides its rows
// — which task is in focus, what the store holds — had nowhere to go and was
// either repeated on every row or dropped. And a reader of the JSON can tell
// one listing from another without inspecting a row.
type Tasks struct {
	Tasks []Row `json:"tasks"`
	// InFocus is which tasks are being worked on, said once for the whole
	// listing. Always a list, empty when none are: a machine reads the one
	// field either way rather than testing for null first.
	InFocus []int `json:"in_focus"`
}

// Timesheet is how long the work took, as it crosses to a machine: every day
// something was done on, in order, with each task's minutes — rounded to the
// quarter hour — and the day's, the sum of those. Minutes and not a written
// length, so whoever fills a sheet in adds them up without reading "1h30".
type Timesheet struct {
	Days []TimesheetDay `json:"days"`
}

// TimesheetDay is one day of it.
type TimesheetDay struct {
	Day     string          `json:"day"`
	Minutes int             `json:"minutes"`
	Tasks   []TimesheetTask `json:"tasks"`
}

// TimesheetTask is one task on one day, and whether it was still in focus when
// the day ended — the time then stopped at midnight, not where the work did.
type TimesheetTask struct {
	Number             int    `json:"number"`
	Title              string `json:"title"`
	Minutes            int    `json:"minutes"`
	LeftOpenAtMidnight bool   `json:"left_open_at_midnight"`
}

// ofSheet turns a sheet into what crosses.
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

// Kept is how a listing of backups crosses — named for what is kept beside the
// store, since Backups already names the way to reach them.
type Kept struct {
	Backups []BackupShape `json:"backups"`
}

// Ledger is how a listing of changes to the shape of the store crosses, with
// what the store holds — which the text has always said and the JSON used to
// lose, having nowhere to put it.
//
// Named for the ledger a store keeps beside its shape, and not Changes: that
// name is taken by the attributes an edit was given, and one word for two
// notions is one place for a reader to be wrong.
type Ledger struct {
	Changes []migration.Shape `json:"changes"`
	Holds   *domain.Held      `json:"holds"`
}

// Acts is how what was done to a task, or over a period, crosses.
type Acts struct {
	Acts []trace.Shape `json:"acts"`
}

// Focus is how the task being worked on crosses, or the fact that none is. One
// shape either way rather than two: a tool that answers either of two things
// can be described only as "either", and a client that will not read an either
// reads nothing at all.
type Focus struct {
	Tasks   []Task `json:"tasks"`
	InFocus []int  `json:"in_focus"`
}

// Objective is how an objective crosses to a machine. Tasks holds what was
// put directly under it, which is why it can hold three numbers while
// progress counts seven things: the sub-tasks under those tasks are counted,
// not listed. Neither Tasks nor Progress sits in objective.Shape: both need
// a task's number to say anything, and an objective may not know of one.
type Objective struct {
	objective.Shape
	Tasks    []int                 `json:"tasks"`
	Progress *domain.ProgressShape `json:"progress"`
}

// OfObjective turns an objective into what crosses.
func OfObjective(o objective.Objective, under []task.Task, p domain.Progress) Objective {
	numbers := make([]int, 0, len(under))
	for _, t := range under {
		numbers = append(numbers, factsOf(t).number)
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

// Swept is how archiving answers: how many were taken off the board and, when
// one was named by its number, that task as it ends up. One shape whether one
// was named or a narrowing reached several, because a description that says
// "either of these" is one a client has to resolve at runtime.
type Swept struct {
	Archived int   `json:"archived"`
	Task     *Task `json:"task,omitempty"`
}

// Deleted confirms an act on something numbered — a task deleted, an
// objective deleted, a backup deleted. Named once and reused, since deleting
// any of the three answers the same shape.
type Deleted struct {
	Deleted int `json:"deleted"`
	// Also names what was deleted in the same act, when more than one was.
	Also []int `json:"also,omitempty"`
	// WasBlocking names the tasks a deleted task was in the way of, when it
	// was; their link to it went with it.
	WasBlocking []int `json:"was_blocking,omitempty"`
}

// Restored confirms a backup was restored, replacing the store with what it
// held.
type Restored struct {
	Restored int `json:"restored"`
}

// Applied names what a change command made — every name it applied, or none
// when nothing was waiting. Nothing waiting answers with an empty list, never
// null: that is a success, not a refusal, and a script reading the field
// should not have to tell the two apart.
type Applied struct {
	Applied []string `json:"applied"`
}

// Of turns what an operation answered into the shape a machine reads — the
// shape the operation declared it answers, and no other.
//
// Read off the declaration rather than worked out from which field of the
// Answer happened to be filled. Worked out, the order of the checks decided the
// answer: a task beside the focus had to be checked before a plain task, a sweep
// answered one shape when it took one task off and another when it took two,
// and the schema an agent was handed described only one of them.
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
	case Deleted:
		return Deleted{Deleted: *a.Deleted, Also: a.Also, WasBlocking: a.Unblocked}
	case Restored:
		return Restored{Restored: *a.Restored}
	case PluginWritten:
		return PluginWritten{Plugin: *a.Plugin}
	case Exported:
		return Exported{Exported: *a.Exported, Holds: a.Holds}
	case Imported:
		return Imported{Imported: *a.Imported, Backup: *a.Restored}
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

// ofActs turns everything done to a task into what crosses.
func ofActs(acts []trace.Act) []trace.Shape {
	out := make([]trace.Shape, 0, len(acts))
	for _, one := range acts {
		out = append(out, trace.OfAct(one))
	}
	return out
}

// ofFocus turns the tasks being worked on into what crosses. Both fields are
// filled: the numbers are what a machine tests, the tasks are what it would
// otherwise have to ask for straight after.
//
// Progress is counted for the first and not for the rest, for the reason
// ofTasks gives: counting what hangs under each would ask for it once per
// task. The operation that names one task counts it; the one that names a list
// counts none.
func ofFocus(held []task.Task, p domain.Progress, numbers []int) Focus {
	out := make([]Task, 0, len(held))
	for _, t := range held {
		out = append(out, OfTask(t, p))
		p = domain.Progress{}
	}
	return Focus{Tasks: out, InFocus: listed(numbers)}
}

// nothingInFocus turns "nothing is being worked on" into what crosses — the
// same shape as ofFocus with both lists empty, so a machine reads one field of
// one shape whether or not anything is held.

// ofTasks turns a listing of tasks into what crosses, a row each.
func ofTasks(tasks []task.Task, inFocus *[]int) Tasks {
	out := make([]Row, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, rowOf(t))
	}
	// Null when the answer was not asked about the focus at all, and a list —
	// empty or not — when it was. A listing nobody asked does not get to claim
	// that nothing is being worked on.
	if inFocus == nil {
		return Tasks{Tasks: out}
	}
	return Tasks{Tasks: out, InFocus: listed(*inFocus)}
}

// heldOr reads which tasks an answer says are in focus, answering none when
// the answer said nothing about it. A listing that was not asked does not get
// to claim that nothing is being worked on.
func heldOr(inFocus *[]int) []int {
	if inFocus == nil {
		return nil
	}
	return *inFocus
}

// listed is a run of numbers as a machine reads it: a list either way, empty
// rather than null when nothing is held, so a client reads one field of one
// shape whether or not anything is being worked on.
func listed(numbers []int) []int {
	if numbers == nil {
		return []int{}
	}
	return numbers
}

// ofObjectives turns every objective into what crosses, with nothing counted
// under any of them, for the same reason ofTasks counts nothing per task.
func ofObjectives(all []objective.Objective) []Objective {
	out := make([]Objective, 0, len(all))
	for _, o := range all {
		out = append(out, OfObjective(o, nil, domain.Progress{}))
	}
	return out
}

// ofBackups turns every backup into what crosses, with nothing said about
// what each held: listing them does not open any of them to count.
func ofBackups(all []domain.Backup) []BackupShape {
	out := make([]BackupShape, 0, len(all))
	for _, b := range all {
		out = append(out, OfBackup(b, nil))
	}
	return out
}

// ofChanges turns every change into what crosses.
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

// OfBackup turns a backup into what crosses, holding what it held when that
// was counted, and nothing when it was not: making a backup does not open it,
// and a zero nobody measured would read as an empty store.
func OfBackup(b domain.Backup, held *domain.Held) BackupShape {
	out := BackupShape{Number: b.Number, Made: domain.FormatMoment(b.Made)}
	if held != nil {
		out.Holds = &domain.Held{Tasks: held.Tasks, Objectives: held.Objectives}
	}
	return out
}
