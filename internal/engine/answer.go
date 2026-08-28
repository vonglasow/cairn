package engine

// The two renderings of an Answer, kept in one place so neither surface
// carries its own copy and the two can never drift apart: the shape a machine
// reads — the JSON `--json` prints on the command line and the shape an MCP
// tool answers with — and the card a person reads. Both live here because
// every one of these is about an Answer, which is the engine's.

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
}

// OfTask turns a task and how far along it is into what crosses to a machine.
func OfTask(t task.Task, p domain.Progress) Task {
	return Task{Shape: task.OfTask(t), Progress: domain.OfProgress(p)}
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
	Tasks []Task `json:"tasks"`
	// InFocus is which task is being worked on, said once. Null when none is.
	InFocus *int `json:"in_focus"`
}

// Objectives is how a listing of objectives crosses.
type Objectives struct {
	Objectives []Objective `json:"objectives"`
}

// Kept is how a listing of backups crosses — named for what is kept beside the
// store, since Backups already names the way to reach them.
type Kept struct {
	Backups []domain.BackupShape `json:"backups"`
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
	Task    *Task `json:"task"`
	InFocus *int  `json:"in_focus"`
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
		numbers = append(numbers, task.OfTask(t).Number)
	}
	return Objective{
		Shape:    objective.OfObjective(o),
		Tasks:    numbers,
		Progress: domain.OfProgress(p),
	}
}

// Deleted confirms an act on something numbered — a task deleted, an
// objective deleted, a backup deleted. Named once and reused, since deleting
// any of the three answers the same shape.
type Deleted struct {
	Deleted int `json:"deleted"`
}

// PutBack confirms a backup was restored, replacing the store with what it
// held.
type PutBack struct {
	PutBack int `json:"put back"`
}

// Applied names what a change command made — every name it applied, or none
// when nothing was waiting. Nothing waiting answers with an empty list, never
// null: that is a success, not a refusal, and a script reading the field
// should not have to tell the two apart.
type Applied struct {
	Applied []string `json:"applied"`
}

// Refusal is how a failure crosses to a machine, on the one stream it reads,
// even when the answer is a refusal it cannot otherwise parse.
type Refusal struct {
	Error string `json:"error"`
}

// Of turns what an operation answered into the shape a machine reads. Answer
// carries every field an operation might fill but only the two or three that
// are its own, so the fields are checked in turn, the most specific first: a
// single entity outranks a list of them, and a list outranks the bookkeeping
// shapes that carry no entity at all.
//
// Written with ifs rather than a switch on purpose: a tagless switch puts its
// conditions outside any covered block, so a mutation of one of them reports
// as unreachable and the gate cannot judge it.
func Of(a Answer) any {
	// A task alongside the focus is the focus being told, and nothing else
	// tells it that way: a card asked for by number says nothing about the
	// focus, and a listing carries its rows too. Checked first, because the
	// plain-task branch below would otherwise swallow it.
	if a.Task != nil && a.InFocus != nil {
		return ofFocus(*a.Task, a.Progress, *a.InFocus)
	}
	if a.Task != nil {
		return OfTask(*a.Task, a.Progress)
	}
	if a.Tasks != nil {
		return ofTasks(a.Tasks, a.InFocus)
	}
	if a.Matched != nil {
		return ofTasks(a.Matched, a.InFocus)
	}
	if a.Objective != nil {
		return OfObjective(*a.Objective, a.Under, a.Progress)
	}
	if a.Objectives != nil {
		return Objectives{Objectives: ofObjectives(a.Objectives)}
	}
	if a.MatchedObjectives != nil {
		return Objectives{Objectives: ofObjectives(a.MatchedObjectives)}
	}
	if a.Backup != nil {
		return domain.OfBackup(*a.Backup, a.Holds)
	}
	if a.Backups != nil {
		return Kept{Backups: ofBackups(a.Backups)}
	}
	if a.Change != nil {
		return migration.OfMigration(*a.Change)
	}
	if a.Changes != nil {
		return Ledger{Changes: ofChanges(a.Changes), Holds: a.Holds}
	}
	if a.RolledBack != nil {
		return migration.OfRolledBack(*a.RolledBack)
	}
	if a.Deleted != nil {
		return Deleted{Deleted: *a.Deleted}
	}
	if a.PutBack != nil {
		return PutBack{PutBack: *a.PutBack}
	}
	if a.Applied != nil {
		return Applied{Applied: a.Applied}
	}
	if a.Holds != nil {
		return *a.Holds
	}
	if a.Acts != nil {
		return Acts{Acts: ofActs(a.Acts)}
	}
	if a.Since != nil {
		return Acts{Acts: ofActs(a.Since)}
	}
	if a.InFocus != nil {
		return nothingInFocus(*a.InFocus)
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

// ofFocus turns the task being worked on into what crosses. Both fields are
// filled: the number is what a machine tests, the task is what it would
// otherwise have to ask for straight after.
func ofFocus(t task.Task, p domain.Progress, number int) Focus {
	held := OfTask(t, p)
	return Focus{Task: &held, InFocus: heldOrNothing(&number)}
}

// nothingInFocus turns "no task is being worked on" into what crosses — the
// same shape as ofFocus with both fields empty, so a machine reads one field
// of one shape whether or not anything is held. Null rather than zero: zero is
// a number no task carries, and a machine testing for it would have to know
// that.
func nothingInFocus(number int) Focus {
	return Focus{InFocus: heldOrNothing(&number)}
}

// ofTasks turns a listing of tasks into what crosses, with no progress counted
// per item: a listing that counted what hangs under each task would ask for it
// once per task to answer one line each.
func ofTasks(tasks []task.Task, inFocus *int) Tasks {
	out := make([]Task, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, OfTask(t, domain.Progress{}))
	}
	return Tasks{Tasks: out, InFocus: heldOrNothing(inFocus)}
}

// heldOr reads which task an answer says is in focus as a number, answering
// none when the answer said nothing about it. What a person reads wants a
// number, since it marks a row by comparing it.
func heldOr(inFocus *int) int {
	if inFocus == nil {
		return task.NothingInFocus
	}
	return *inFocus
}

// heldOrNothing reads the same thing for a machine, which wants the absence
// told rather than spelled zero: null when none is held or when the answer
// said nothing about it — a listing that was not asked does not get to claim
// none.
func heldOrNothing(inFocus *int) *int {
	if heldOr(inFocus) == task.NothingInFocus {
		return nil
	}
	return inFocus
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
func ofBackups(all []domain.Backup) []domain.BackupShape {
	out := make([]domain.BackupShape, 0, len(all))
	for _, b := range all {
		out = append(out, domain.OfBackup(b, nil))
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

// RenderText turns an answer into what a person reads. It works from which
// field of the answer was set, the same way Of works out the JSON — with one
// exception that field cannot resolve on its own: Deleted names three
// different things (a task, an objective, a backup), and only the operation
// says which. A change shown in full and a change just rolled back need no
// such help — Answer keeps them apart itself, in Change and RolledBack.
//
// Written with ifs rather than a switch on purpose: a tagless switch puts its
// conditions outside any covered block, so a mutation of one of them reports
// as unreachable and the gate cannot judge it.
func RenderText(op Operation, a Answer) string {
	if a.Task != nil {
		return task.Text(*a.Task, a.Progress) + task.Under(a.Under)
	}
	if a.Tasks != nil {
		return tasksText(a.Tasks, a.InFocus)
	}
	if a.Matched != nil {
		return foundText(a.Matched)
	}
	if a.Objective != nil {
		return objective.Text(*a.Objective, a.Progress) + objective.Serving(linesOf(a.Under))
	}
	if a.Objectives != nil {
		return objectivesText(a.Objectives)
	}
	if a.MatchedObjectives != nil {
		return objectivesText(a.MatchedObjectives)
	}
	if a.Backup != nil {
		return backupLine(*a.Backup, *a.Holds)
	}
	if a.Backups != nil {
		return backupsText(a.Backups)
	}
	if a.RolledBack != nil {
		return migration.RolledBackText(*a.RolledBack)
	}
	if a.Change != nil {
		return migration.Text(*a.Change)
	}
	if a.Changes != nil {
		return migration.List(a.Changes, *a.Holds)
	}
	if a.Deleted != nil {
		return fmt.Sprintf("%s %d was deleted\n", op.Thing, *a.Deleted)
	}
	if a.PutBack != nil {
		return fmt.Sprintf("backup %d was put back\n", *a.PutBack)
	}
	if a.Applied != nil {
		return appliedText(a.Applied)
	}
	if a.Acts != nil {
		return trace.Text(a.Acts)
	}
	if a.Since != nil {
		return trace.List(a.Since) + trace.Summary(trace.AddedUp(a.Since, time.Now()))
	}
	if a.InFocus != nil {
		return "nothing is in focus\n"
	}
	return ""
}

// tasksText renders every task, saying so when there is none.
func tasksText(tasks []task.Task, inFocus *int) string {
	if len(tasks) == 0 {
		return "there is no task\n"
	}
	return task.Tree(tasks, heldOr(inFocus))
}

// foundText renders every task a search found, flat — a found sub-task hangs
// under nothing, since what matched is what is shown rather than where it
// sits — saying so when nothing was found.
func foundText(tasks []task.Task) string {
	if len(tasks) == 0 {
		return "there is no task\n"
	}
	var out strings.Builder
	for _, t := range tasks {
		out.WriteString(task.Line(t))
		out.WriteString("\n")
	}
	return out.String()
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
		_, _ = fmt.Fprintf(&out, "#%d  made %s\n", one.Number, domain.FormatMoment(one.Made))
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
		_, _ = fmt.Fprintf(&out, "applied %q\n", name)
	}
	return out.String()
}
