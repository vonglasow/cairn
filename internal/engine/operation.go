package engine

// This file describes what cairn can be asked to do, once, so both surfaces
// are built from the same description rather than each carrying its own copy
// of the things a caller can ask for.

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"cairn/internal/domain"
	"cairn/internal/migration"
	"cairn/internal/objective"
	"cairn/internal/task"
	"cairn/internal/trace"
)

// Level is how far an agent may reach. It also settles what a client is told a
// call would cost: Read is read-only, Full loses work, Write is neither.
type Level int

const (
	_ Level = iota // the zero value: never a level.
	// Read is level one: listing and showing, nothing that changes anything.
	Read
	// Write is level two: also adding and changing, and what sets a backup
	// aside first.
	Write
	// Full is level three: also what loses work already done.
	Full
)

// ErrNotALevel is the reason a level is refused: what was given spelled none
// of "read", "write" or "full".
var ErrNotALevel = errors.New("not one of the access levels")

// ParseLevel reads a level by name, the way a person types it when starting
// a server. Anything other than "read", "write" or "full" is refused, wrapping
// ErrNotALevel.
func ParseLevel(name string) (Level, error) {
	switch name {
	case "read":
		return Read, nil
	case "write":
		return Write, nil
	case "full":
		return Full, nil
	}
	return 0, fmt.Errorf("%q: expected \"read\", \"write\" or \"full\": %w", name, ErrNotALevel)
}

// Again is what calling an operation a second time, with the same arguments,
// does to the store. A client that retries a call it is unsure went through
// needs to know whether the retry costs anything.
type Again int

const (
	_ Again = iota // the zero value: never said.
	// ChangesNothingMore is a second call leaving the store as the first did —
	// or being refused for having nothing left to do.
	ChangesNothingMore
	// DoesItAgain is a second call adding again — another task, another backup,
	// more time spent, one more change undone.
	DoesItAgain
)

// Needs says which way the store has to be opened.
type Needs int

const (
	_ Needs = iota // the zero value: never a way of opening the store.
	// AStore refuses a store behind this cairn: a request about what it holds
	// does not change its shape on the way past.
	AStore
	// AStoreBehind opens the store whatever shape it is in, for the operations
	// whose job is to change that shape.
	AStoreBehind
	// NoStore opens nothing at all, for what is reachable without the store —
	// a backup is most needed exactly when the store cannot be opened. It
	// still asks for the path the store would live at, since that is where
	// a backup finds itself.
	NoStore
	// AConfiguration opens no store and has no use for its path either: it
	// asks for a tolerant reading of the configuration instead, config show
	// alone. Kept apart from NoStore because the two answer different
	// questions — NoStore says "no store, but still the path"; this says
	// "no store, and no path — a configuration in its place".
	AConfiguration
)

// Kind is what sort of value an argument carries. It exists so a surface can
// say so, and so whatever reads a caller's words reads them as what they are:
// a schema names a number, a yes is a flag that carries nothing.
type Kind int

const (
	_ Kind = iota // the zero value: never a kind.
	// Text is a plain string.
	Text
	// Number is a whole number.
	Number
	// Day is a calendar day.
	Day
	// Duration is a length of time.
	Duration
	// List is more than one string.
	List
	// Bool is named or not named, and carries nothing else. A flag on the
	// command line, a boolean in a schema.
	Bool
)

// Argument describes one thing an operation takes, once, so a surface can
// build a flag or a schema field from it without knowing the other's shape.
type Argument struct {
	// Name is how the argument is named, as a flag and as a JSON field.
	Name string
	// Kind is what sort of value it carries.
	Kind Kind
	// Positional says it is taken as an argument rather than a flag.
	Positional bool
	// Required says it must be named for the operation to make sense.
	Required bool
	// OneOf is the values it accepts, when there is a closed set.
	OneOf []string
	// Outside is what a value outside OneOf is refused with: the same
	// sentinel the attribute is refused with when it is given, so one mistake
	// has one reason whether a listing narrows by it or an edit gives it.
	Outside error
	// Means is one sentence: cobra's flag help and the schema's description.
	Means string
	// Of is what a Kind: Number argument names, written the way the refusal
	// has to read it: "a task", "an objective".
	Of string
}

// Answered is the one shape an operation's answer takes — an empty value of it,
// carried so a surface can read the shape off the type.
//
// Declared rather than inferred: a surface describing itself has to say
// beforehand what it will answer, and Of and RenderText build exactly that
// shape from it rather than working one out from which field was filled.
//
// One shape and not a set of them. It used to be a set, so that `task focus`
// could say "a task, or nothing in focus" — but a description that says "either
// of these" is a description a reader has to resolve at runtime, and at least
// one real client refuses it outright. So the operations that answered two
// things were made to answer one: `task focus` hands back the same shape held
// or not, with the fields it cannot fill left empty.
type Answered any

// Input is what a caller handed over, one entry per Argument.Name that was
// actually named.
//
// Not naming something and naming it as nothing are two different requests,
// and only one of them changes anything: a key absent was never named, a key
// present holding nothing was named as nothing, which is how a field is
// emptied. That is why these are maps and why the values are plain — a
// pointer would offer a second way to say "nothing", and every operation
// would each have to decide what the difference meant.
//
// Given carries every kind but List; Listed carries the arguments declared
// with Kind List.
type Input struct {
	Given  map[string]string
	Listed map[string][]string
}

// Named reports whether the caller named this argument at all. A Run body asks
// before reading, since what was never named is not a request to empty it.
func (in Input) Named(name string) bool {
	if _, named := in.Given[name]; named {
		return true
	}
	_, named := in.Listed[name]
	return named
}

// Yes reports whether a Kind: Bool argument was named as true. Named at all is
// not enough: a flag given as false was named, and reading it as a request
// would turn `--first=false` into moving a task first.
func (in Input) Yes(name string) bool { return in.Given[name] == "true" }

// Change answers what to hand Changes for a name: nothing when it was
// not named, and what it was named as otherwise. The three states the engine
// models are worked out here, once, rather than in each operation.
func (in Input) Change(name string) *string {
	given, named := in.Given[name]
	if !named {
		return nil
	}
	return &given
}

// Changed answers what to hand Changes for a list of strings, on the
// same three states as Change.
func (in Input) Changed(name string) *[]string {
	listed, named := in.Listed[name]
	if !named {
		return nil
	}
	return &listed
}

// Whole answers the number that was named for a Kind: Number argument, and
// whether it was named at all. It cannot fail because Do calls Check first,
// which refuses anything that would not parse before Run ever sees it.
func (in Input) Whole(name string) (int, bool) {
	given, named := in.Given[name]
	if !named {
		return 0, false
	}
	n, _ := strconv.Atoi(given)
	return n, true
}

// Operation is one of the things cairn can be asked to do, described
// once so both surfaces are built from it.
type Operation struct {
	// Thing is what the operation is about — "task", "backup".
	Thing string
	// Verb is what it does to it — "add", "list".
	Verb string
	// Short is cobra's Short.
	Short string
	// Does is one or two sentences: cobra's Long and the tool's description.
	Does string
	// Example is how it is typed, a line per example: cobra's Example. A
	// person reads one faster than a list of what the operation takes.
	Example string
	// Level is how far an agent may reach to call it.
	Level Level
	// Needs says which way the store has to be opened for it.
	Needs Needs
	// Again says what calling it a second time does. Every operation that
	// writes says it; one that only reads changes nothing however often.
	Again Again
	// Overwrites says it can replace what was already written — a
	// description emptied, a set of labels replaced — rather than only add
	// to it. Losing work outright is said by its Level.
	Overwrites bool
	// Takes is what it accepts.
	Takes []Argument
	// Run carries out the operation against whatever it was given to work on.
	Run func(On, Input) (Answer, error)
	// Answers is the shape this operation hands back. A surface that describes
	// itself to a machine says so before anything runs.
	Answers Answered
}

// Name is how an operation is named as a command or a tool: "task_new".
func (o Operation) Name() string { return o.Thing + "_" + o.Verb }

// Line is how an operation reads on the command line: "task new".
func (o Operation) Line() string { return o.Thing + " " + o.Verb }

// All is everything cairn can be asked to do, in a stable order.
func All() []Operation {
	return []Operation{
		TaskNew, TaskList, TaskShow, TaskEdit, TaskDelete, TaskSearch,
		TaskArchive, TaskUnarchive, TaskLabel,
		TaskMove, TaskFocus, TaskUnfocus,
		ObjectiveNew, ObjectiveList, ObjectiveShow, ObjectiveEdit, ObjectiveDelete, ObjectiveSearch,
		BackupNew, BackupList, BackupShow, BackupDelete, BackupRestore, BackupExport, BackupImport,
		MigrationList, MigrationShow, MigrationApply, MigrationRollback,
		TraceShow, TraceList,
		TimesheetShow, TimesheetAdd,
		ConfigShow,
		PluginWrite,
	}
}

// Check refuses what Run could not read anyway, before the store is opened
// so a refusal never touches the disk:
//
//   - a Required argument has to have been named, and a Required list has to
//     hold something. A caller that named nothing
//     it needs is told which argument that was, rather than answered as if it
//     had asked for nothing — an agent can correct a refusal and believes an
//     answer;
//   - a name that was given has to be one o.Takes lists. A typo in an
//     argument's name is not the same request as leaving it out, and letting
//     it through answers a question nobody asked;
//   - every Argument of Kind: Number that was named must read as a whole
//     number, empty excepted since that is how such an argument is cleared.
func (o Operation) Check(in Input) error {
	takes := map[string]bool{}
	for _, arg := range o.Takes {
		takes[arg.Name] = true
	}
	for name := range in.Given {
		if !takes[name] {
			return Misread(fmt.Errorf("%q: %s takes no argument by that name", name, o.Line()))
		}
	}
	for name := range in.Listed {
		if !takes[name] {
			return Misread(fmt.Errorf("%q: %s takes no argument by that name", name, o.Line()))
		}
	}
	for _, arg := range o.Takes {
		if arg.Required && !in.Named(arg.Name) {
			return Misread(fmt.Errorf("%s needs %q, and was not given one", o.Line(), arg.Name))
		}
		if listed, named := in.Listed[arg.Name]; arg.Required && named && len(listed) == 0 {
			return Misread(fmt.Errorf("%s needs at least one %q, and was given none", o.Line(), arg.Name))
		}
		given, named := in.Given[arg.Name]
		if arg.Required && arg.Kind == Number && named && given == "" {
			return Misread(fmt.Errorf("%s needs %q, and was given nothing", o.Line(), arg.Name))
		}
		if !named || given == "" {
			continue
		}
		switch arg.Kind {
		case Number:
			if _, err := strconv.Atoi(given); err != nil {
				return Misread(fmt.Errorf("%q: %s number is expected", given, arg.Of))
			}
		case Day:
			// A narrowing reads a day and compares it against a column of
			// days. Left unread, it compared against whatever was typed and
			// answered whatever that sorted beside — silence where a refusal
			// belonged.
			if _, err := domain.ParseDeadline(given); err != nil {
				return err
			}
		}
	}
	return nil
}

// On is what an operation was given to work on. All three come from the same
// place, so they travel together rather than as three parameters nobody can
// keep in the right order.
type On struct {
	// Engine is open on the store, or nil when the operation asked for none.
	Engine *Engine
	// Backups are what is kept beside the store, reachable whether or not the
	// store itself could be opened.
	Backups *Backups
	// Path is where the store is. The operations that change its shape set a
	// backup aside first, and that is what they need it for.
	Path string
	// Notes is the root every task's outside context sits under, as the surface
	// found it; empty when the surface did not say. See docs/context.md.
	Notes string
	// Home is what the places were found from, for `config show`, which
	// answers how they were found rather than where they are.
	Home Home
}

// Answer is what an operation answered. Every operation fills the fields that
// are its own and leaves the rest alone, and says beforehand — in Answers —
// which shape they make: Of and RenderText read that declaration, never which
// field happens to be filled.
//
// Deleted and Restored are numbers rather than entities because what they name
// is gone, or is a whole store just replaced: there is nothing left to hand
// back but which one it was.
//
// RolledBack is its own field rather than sharing Change: a change shown in
// full and a change just undone both carry a *migration.Migration, but they
// are not the same document — a rollback names what was undone and what it
// cost, where a change shown in full carries its SQL.
//
// Matched is its own field beside Tasks for the same reason: a listing renders
// as a tree, sub-tasks nested under the task they hang from, but what matched
// what was asked is listed flat — a sub-task that matched hangs under nothing,
// since what matched is what is shown, not where it sits. Its JSON is the same
// shape as Tasks, since a machine reading either does not care about
// indentation.
//
// Matched and not Found, because two questions land in it: what a search found,
// and what a narrowed listing kept. Both are "these matched, and where each
// sits is not what was asked", which is one notion and so one field.
//
// MatchedObjectives sits beside Matched for the reason Objective sits beside
// Task: an objective is not a task, and a field that carried either would
// make every reader ask which one it got before it could read it. Its JSON
// is the same shape as Objectives, for the same reason Matched's is the same
// shape as Tasks.
type Answer struct {
	Task *task.Task
	// Context is where the task shown keeps what was written about it
	// elsewhere: a directory named for its number under the notes. Composed,
	// never looked at — see docs/context.md. Empty when the caller did not say
	// where the notes are.
	Context string
	// StoodSince is the moment the task shown last changed column, or was
	// made: how long it has sat where it is. Read off its history, never
	// kept beside it — see docs/ordering.md.
	StoodSince        time.Time
	Tasks             []task.Task
	Matched           []task.Task
	Under             []task.Task
	Objective         *objective.Objective
	Objectives        []objective.Objective
	MatchedObjectives []objective.Objective
	Backup            *domain.Backup
	Backups           []domain.Backup
	Change            *migration.Migration
	Changes           []migration.Migration
	RolledBack        *migration.Migration
	Progress          domain.Progress
	Holds             *domain.Held
	Deleted           *int
	// Also names what was deleted in the same act as Deleted.
	Also []int
	// Unblocked names the tasks a deleted task was in the way of: the link
	// goes with it, and afterwards nothing would say they once waited on it.
	Unblocked []int
	Restored  *int
	// Exported is the file the store was written to, Imported the one it
	// was brought in from.
	Exported *string
	Imported *string
	// Plugin is where the Claude Code plugin was written.
	Plugin *string
	// Archived is how many a sweep took off the board, Labelled how many it
	// put a label on.
	Archived *int
	Labelled *int
	Applied  []string
	// Since is what was done over a period, across every task, oldest first.
	// Its own field beside Acts for the reason Matched sits beside Tasks: a
	// history is about one task and does not name it on every line, where a
	// period names the task each act was on.
	Since []trace.Act
	// Acts is what has been done to a task, oldest first.
	Acts []trace.Act
	// InFocus is which tasks are being worked on. Nil says the answer is not
	// about the focus at all; a list — empty or not — says it is. Every
	// listing carries it, because a listing marks it; and the operations that
	// answer about the focus carry it alone, since "nothing is in focus" is an
	// answer and not an absence.
	InFocus *[]int
	// Held are the tasks being worked on, for the operation that names them
	// rather than marking them.
	Held []task.Task
	// NothingBlocks are the tasks a listing marks as waiting on nothing any
	// more. A fact of the store rather than of a task — knowing it means
	// reading the state of every task that blocks a row, which a narrowed
	// listing does not carry and a task cannot see at all.
	NothingBlocks []int
	// Along is how far along each thing a listing names is, by its number, and
	// Contexts where each task's outside context lives — said once in the
	// listing, so a board drawing every card does not show each task in turn
	// to learn them.
	Along    map[int]domain.Progress
	Contexts map[int]string
	// Config is what `config show` answers: the file read and its two
	// emplacements, each with what named it.
	Config *Config
	// Sheet is how long the work took, day by day, and Named what each task
	// on it is called — a sheet holds numbers, and a person reads names.
	Sheet *trace.Sheet
	Named map[int]string
}

// Do runs an operation: it opens the store the way op.Needs says to, hands
// Run an engine open on it — or none, when nothing was asked for — and closes
// what it opened on the way out, whatever Run does. It names the store alone:
// an operation run this way is told nothing about the notes.
func Do(op Operation, path string, in Input) (Answer, error) {
	return DoAt(op, Places{Store: path}, in, ByHand)
}

// ByHand is the surface a history says nothing about: the person at the
// terminal. Every other surface names itself, and a line shows it only when it
// was not this one.
const ByHand = ""

// DoAt runs an operation at the places a surface found, and says which surface
// asked, so what it changes is written down as that surface's doing.
//
// Its own entry point rather than parameters on Do: a hundred callers name no
// surface and no notes because they are tests about something else, and
// threading two empty strings through all of them would say nothing a hundred
// times.
func DoAt(op Operation, at Places, in Input, through string) (Answer, error) {
	if err := op.Check(in); err != nil {
		return Answer{}, err
	}
	return doAt(op, at, in, through)
}

// doAt is DoAt once the request was checked: it opens what op.Needs says and
// runs it.
func doAt(op Operation, at Places, in Input, through string) (Answer, error) {
	switch op.Needs {
	case AStoreBehind:
		return doWith(OpenBehind, op, at, in, through)
	case AStore:
		return doWith(Open, op, at, in, through)
	default:
		return op.Run(On{Backups: BackupsAt(at.Store), Path: at.Store, Notes: at.Notes, Home: at.Home}, in)
	}
}

// doWith opens the store the way open does, hands Run an engine open on it,
// and closes it on the way out — whatever Run does or fails on.
//
// Closing is where the store writes back what it held aside, so a close that
// fails after Run worked is said rather than dropped.
func doWith(
	open func(string) (*Engine, error),
	op Operation, at Places, in Input, through string,
) (answer Answer, err error) {
	eng, err := open(at.Store)
	if err != nil {
		return Answer{}, Told(err)
	}
	defer func() {
		if closed := eng.Close(); err == nil && closed != nil {
			answer, err = Answer{}, storeProblem(closed)
		}
	}()
	eng.through, eng.notes, eng.home = through, at.Notes, at.Home
	// Check already ran above, so this takes the door that does not repeat it.
	return eng.run(op, in)
}
