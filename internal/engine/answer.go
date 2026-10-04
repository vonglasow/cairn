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
	// Context is where what was written about this task elsewhere lives, said
	// when it is the task asked about and the caller said where the notes are.
	// A harness writes there and reads there; cairn only says where.
	Context string `json:"context,omitempty"`
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
		numbers = append(numbers, task.OfTask(t).Number)
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

// Swept is how a sweep answers: how many were taken off the board.
type Swept struct {
	Archived int `json:"archived"`
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
	// Config is its own answer, sharing no field with any other operation's —
	// checked first for that reason alone, not because it could be mistaken
	// for one of the branches below.
	if a.Config != nil {
		return *a.Config
	}
	// A task alongside the focus is the focus being told, and nothing else
	// tells it that way: a card asked for by number says nothing about the
	// focus, and a listing carries its rows too. Checked first, because the
	// plain-task branch below would otherwise swallow it.
	if a.Task != nil && a.InFocus != nil {
		return ofFocus([]task.Task{*a.Task}, a.Progress, *a.InFocus)
	}
	if a.Held != nil && a.InFocus != nil {
		return ofFocus(a.Held, domain.Progress{}, *a.InFocus)
	}
	if a.Task != nil {
		said := ofTaskUnder(*a.Task, a.Progress, a.Under)
		said.Context = a.Context
		return said
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
	if a.Archived != nil {
		return Swept{Archived: *a.Archived}
	}
	if a.Labelled != nil {
		return Labelled{Labelled: *a.Labelled}
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
func nothingInFocus(numbers []int) Focus {
	return Focus{Tasks: []Task{}, InFocus: listed(numbers)}
}

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

// contextText is the line of a card that says where a task's outside context
// lives, apart from the facts above it: the place is not something the task
// carries but somewhere the task's material is kept.
func contextText(at string) string {
	if at == "" {
		return ""
	}
	return "\ncontext  " + at + "\n"
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
		return task.Text(*a.Task, a.Progress) + task.Under(a.Under) + contextText(a.Context)
	}
	if a.Tasks != nil {
		return tasksText(a.Tasks, a)
	}
	if a.Matched != nil {
		return foundText(a.Matched, a)
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
	if a.Archived != nil {
		return fmt.Sprintf("%d %s(s) archived\n", *a.Archived, op.Thing)
	}
	if a.Labelled != nil {
		return fmt.Sprintf("%d %s(s) labelled\n", *a.Labelled, op.Thing)
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
	if a.Held != nil {
		return heldText(a.Held, a)
	}
	if a.InFocus != nil {
		return "nothing is in focus\n"
	}
	return ""
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
