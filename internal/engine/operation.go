package engine

// This file describes what cairn can be asked to do, once, so both surfaces
// are built from the same description rather than each carrying its own copy
// of the things a caller can ask for.

import (
	"errors"
	"fmt"
	"strconv"

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
	// a backup is most needed exactly when the store cannot be opened.
	NoStore
)

// Kind is what sort of value an argument carries. It exists so a surface can
// say so: a schema names a number, a flag takes a string either way.
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
	// Means is one sentence: cobra's flag help and the schema's description.
	Means string
	// Of is what a Kind: Number argument names, written the way the refusal
	// has to read it: "a task", "an objective".
	Of string
}

// Answered is the one shape an operation's answer takes — an empty value of it,
// carried so a surface can read the shape off the type.
//
// Declared rather than inferred, because nothing else knows it: Of works out
// which shape to build by looking at which field of an Answer was filled, and
// that is only knowable once the operation has run. A surface describing itself
// has to say beforehand.
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
	// Level is how far an agent may reach to call it.
	Level Level
	// Needs says which way the store has to be opened for it.
	Needs Needs
	// Takes is what it accepts.
	Takes []Argument
	// Run carries out the operation against whatever it was given to work on.
	Run func(On, Input) (Answer, error)
	// Answers is the shape this operation hands back. A surface that describes
	// itself to a machine says so before anything runs.
	Answers Answered
}

// Check refuses what Run could not read anyway, before the store is opened
// so a refusal never touches the disk:
//
//   - a Required argument has to have been named. A caller that named nothing
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
			return fmt.Errorf("%q: %s takes no argument by that name", name, o.Line())
		}
	}
	for name := range in.Listed {
		if !takes[name] {
			return fmt.Errorf("%q: %s takes no argument by that name", name, o.Line())
		}
	}
	for _, arg := range o.Takes {
		if arg.Required && !in.Named(arg.Name) {
			return fmt.Errorf("%s needs %q, and was not given one", o.Line(), arg.Name)
		}
		if arg.Kind != Number {
			continue
		}
		given, named := in.Given[arg.Name]
		if !named || given == "" {
			continue
		}
		if _, err := strconv.Atoi(given); err != nil {
			return fmt.Errorf("%q: %s number is expected", given, arg.Of)
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
}

// Answer is what an operation answered. Every operation fills the two or
// three fields that are its own and leaves the rest alone, so both surfaces
// read one shape rather than sorting fourteen.
//
// Deleted and PutBack are numbers rather than entities because what they name
// is gone, or is a whole store just replaced: there is nothing left to hand
// back but which one it was.
//
// RolledBack is its own field rather than sharing Change: a change shown in
// full and a change just undone both carry a *migration.Migration, but they
// are not the same document — a rollback names what was undone and what it
// cost, where a change shown in full carries its SQL. Keeping them apart here
// is what lets Of tell the two cases apart by which field is set, rather than
// a surface guessing from which operation was called.
//
// Found is its own field beside Tasks for the same reason: a listing renders
// as a tree, sub-tasks nested under the task they hang from, but a found task
// is listed flat — a found sub-task hangs under nothing in the listing, since
// what matched is what is shown, not where it sits. Its JSON is the same
// shape as Tasks, since a machine reading either does not care about
// indentation.
//
// FoundObjectives sits beside Found for the reason Objective sits beside
// Task: an objective is not a task, and a field that carried either would
// make every reader ask which one it got before it could read it. Its JSON
// is the same shape as Objectives, for the same reason Found's is the same
// shape as Tasks.
type Answer struct {
	Task            *task.Task
	Tasks           []task.Task
	Found           []task.Task
	Under           []task.Task
	Objective       *objective.Objective
	Objectives      []objective.Objective
	FoundObjectives []objective.Objective
	Backup          *domain.Backup
	Backups         []domain.Backup
	Change          *migration.Migration
	Changes         []migration.Migration
	RolledBack      *migration.Migration
	Progress        domain.Progress
	Holds           *domain.Held
	Deleted         *int
	PutBack         *int
	Applied         []string
	// Since is what was done over a period, across every task, oldest first.
	// Its own field beside Acts for the reason Found sits beside Tasks: a
	// history is about one task and does not name it on every line, where a
	// period names the task each act was on.
	Since []trace.Act
	// Acts is what has been done to a task, oldest first.
	Acts []trace.Act
	// InFocus is which task is being worked on — the number when one is, zero
	// when none is. Every listing carries it, because a listing marks it; and
	// the operations that answer about the focus carry it alone, since "nothing
	// is in focus" is an answer and not an absence.
	InFocus *int
}

// Do runs an operation: it opens the store the way op.Needs says to, hands
// Run an engine open on it — or none, when nothing was asked for — and closes
// what it opened on the way out, whatever Run does.
func Do(op Operation, path string, in Input) (Answer, error) {
	return DoThrough(op, path, in, ByHand)
}

// ByHand is the surface a history says nothing about: the person at the
// terminal. Every other surface names itself, and a line shows it only when it
// was not this one.
const ByHand = ""

// DoThrough runs an operation and says which surface asked, so what it changes
// is written down as that surface's doing.
//
// Its own entry point rather than a parameter on Do: seventy callers name no
// surface because they are the terminal, and threading an empty string through
// all of them would say nothing seventy times.
func DoThrough(op Operation, path string, in Input, through string) (Answer, error) {
	if err := op.Check(in); err != nil {
		return Answer{}, err
	}
	switch op.Needs {
	case AStoreBehind:
		return doWith(OpenBehind, op, path, in, through)
	case AStore:
		return doWith(Open, op, path, in, through)
	default:
		return op.Run(On{Backups: BackupsAt(path), Path: path}, in)
	}
}

// doWith opens the store the way open does, hands Run an engine open on it,
// and closes it on the way out — whatever Run does or fails on.
func doWith(
	open func(string) (*Engine, error),
	op Operation, path string, in Input, through string,
) (Answer, error) {
	eng, err := open(path)
	if err != nil {
		if errors.Is(err, ErrStoreBehind) {
			// The store knows it is behind; only a surface knows what
			// somebody has to type about it, which is why it is added here
			// rather than left for whoever calls Do to work out on their own.
			return Answer{}, fmt.Errorf("%w — run `cairn migration apply`", err)
		}
		return Answer{}, err
	}
	defer func() { _ = eng.Close() }()
	eng.through = through
	return op.Run(On{Engine: eng, Backups: BackupsAt(path), Path: path}, in)
}
