package engine

import (
	"errors"
	"fmt"
	"os/user"
	"strconv"
	"strings"
	"time"

	"cairn/internal/domain"
	"cairn/internal/migration"
	"cairn/internal/objective"
	"cairn/internal/task"
	"cairn/internal/trace"
)

// Level is how far an agent may reach.
type Level int

const (
	_ Level = iota
	// Read is level one: listing and showing, nothing that changes anything.
	Read
	// Write is level two: also adding and changing, and what sets a backup aside first.
	Write
	// Full is level three: also what loses work already done.
	Full
)

// ErrNotALevel is the reason a level is refused: what was given spelled none of "read", "write" or "full".
var ErrNotALevel = errors.New("not one of the access levels")

// ParseLevel reads a level by name, the way a person types it when starting a server.
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

// Again is what calling an operation a second time, with the same arguments, does to the store.
type Again int

const (
	_ Again = iota
	// ChangesNothingMore is a second call leaving the store as the first did — or being refused for having nothing left to do.
	ChangesNothingMore
	// DoesItAgain is a second call adding again — another task, another backup, more time spent, one more change undone.
	DoesItAgain
)

// Needs says which way the store has to be opened.
type Needs int

const (
	_ Needs = iota
	// AStore refuses a store behind this cairn: a request about what it holds does not change its shape on the way past.
	AStore
	// AStoreBehind opens the store whatever shape it is in, for the operations whose job is to change that shape.
	AStoreBehind
	// NoStore opens nothing at all, for what is reachable without the store — a backup is most needed exactly when the store cannot be opened.
	NoStore
	// AConfiguration opens no store and has no use for its path either: it asks for a tolerant reading of the configuration instead, config show alone.
	AConfiguration
)

// Kind is what sort of value an argument carries.
type Kind int

const (
	_ Kind = iota
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
	// Bool is named or not named, and carries nothing else.
	Bool
)

// Argument describes one thing an operation takes, once, so a surface can build a flag or a schema field from it without knowing the other's shape.
type Argument struct {
	Name       string
	Kind       Kind
	Positional bool
	Required   bool
	OneOf      []string
	Outside    error
	Means      string
	Of         string
}

// Answered is the one shape an operation's answer takes — an empty value of it, carried so a surface can read the shape off the type.
type Answered any

// Input is what a caller handed over, one entry per Argument.Name that was actually named.
type Input struct {
	Given     map[string]string
	Listed    map[string][]string
	Confirmed bool
}

// Named reports whether the caller named this argument at all.
func (in Input) Named(name string) bool {
	if _, named := in.Given[name]; named {
		return true
	}
	_, named := in.Listed[name]
	return named
}

// Yes reports whether a Kind: Bool argument was named as true.
func (in Input) Yes(name string) bool { return in.Given[name] == "true" }

// Change answers what to hand Changes for a name: nothing when it was not named, and what it was named as otherwise.
func (in Input) Change(name string) *string {
	given, named := in.Given[name]
	if !named {
		return nil
	}
	return &given
}

// Changed answers what to hand Changes for a list of strings, on the same three states as Change.
func (in Input) Changed(name string) *[]string {
	listed, named := in.Listed[name]
	if !named {
		return nil
	}
	return &listed
}

// Whole answers the number that was named for a Kind: Number argument, and whether it was named at all.
func (in Input) Whole(name string) (int, bool) {
	given, named := in.Given[name]
	if !named {
		return 0, false
	}
	n, _ := strconv.Atoi(given)
	return n, true
}

// Operation is one of the things cairn can be asked to do, described once so both surfaces are built from it.
type Operation struct {
	Thing      string
	Verb       string
	Short      string
	Does       string
	Example    string
	Level      Level
	Needs      Needs
	Again      Again
	Overwrites bool
	Takes      []Argument
	Run        func(On, Input) (Answer, error)
	Answers    Answered
	Sweeps     func(Input) bool
	Finishes   bool
}

// Agreed is what was given, said yes to: what a surface hands over once the person it asked has agreed.
func Agreed(in Input) Input {
	in.Confirmed = true
	return in
}

// ErrNotConfirmed is an act that asks first, asked for without saying yes.
var ErrNotConfirmed = Refusing("not-confirmed", "needs --yes to confirm")

// ErrStepStillOpen is returned when a task is put in done while one of its steps is not, and the act was not confirmed.
var ErrStepStillOpen = task.ErrStepStillOpen

// ErrNoSuchStep is re-exported so a surface can recognise it.
var ErrNoSuchStep = task.ErrNoSuchStep

// AsksFirst says whether an act waits to be confirmed: one that loses work, or one that sweeps up whatever a narrowing reaches.
func (o Operation) AsksFirst(in Input) bool {
	return o.Level == Full || o.Sweeps != nil && o.Sweeps(in)
}

func (o Operation) admitted(in Input) error {
	if err := o.Check(in); err != nil {
		return err
	}
	return o.waiting(in)
}

func (o Operation) waiting(in Input) error {
	if !o.AsksFirst(in) || in.Confirmed {
		return nil
	}
	var named []string
	for _, argument := range o.Takes {
		if argument.Positional && strings.TrimSpace(in.Given[argument.Name]) != "" {
			named = append(named, strings.TrimSpace(in.Given[argument.Name]))
		}
	}
	target := strings.TrimSpace(o.Thing + " " + strings.Join(named, " "))
	if len(named) == 0 && o.Sweeps != nil && o.Sweeps(in) {
		target = "every " + o.Thing + " it reaches"
	}
	for _, more := range in.Listed["also"] {
		target += ", " + o.Thing + " " + strings.TrimSpace(more)
	}
	return fmt.Errorf("%s %w", target, ErrNotConfirmed)
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
		TaskMove, TaskFocus, TaskUnfocus, TaskExport, TaskImport,
		StepNew, StepList, StepEdit, StepDelete,
		ObjectiveNew, ObjectiveList, ObjectiveShow, ObjectiveEdit, ObjectiveDelete, ObjectiveSearch,
		BackupNew, BackupList, BackupShow, BackupDelete, BackupRestore,
		MigrationList, MigrationShow, MigrationApply, MigrationRollback,
		LabelList, LabelDelete, LabelRename,
		TraceShow, TraceList,
		TimesheetShow, TimesheetAdd,
		ConfigShow, ConfigInit,
		PluginWrite,
	}
}

// Check refuses what Run could not read anyway, before the store is opened so a refusal never touches the disk:.
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
			if _, err := domain.ParseDeadline(given); err != nil {
				return err
			}
		}
	}
	return nil
}

// On is what an operation was given to work on.
type On struct {
	Engine  *Engine
	Backups *Backups
	Path    string
	Notes   string
	Plugin  string
	Home    Home
}

// Answer is what an operation answered.
type Answer struct {
	Task              *task.Task
	Context           string
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
	Also              []int
	Unblocked         []int
	Restored          *int
	Exported          *string
	Imported          *string
	Plugin            *string
	Written           *string
	Archived          *int
	Labelled          *int
	Labels            *[]LabelCarried
	Applied           []string
	Since             []trace.Act
	Acts              []trace.Act
	InFocus           *[]int
	Held              []task.Task
	NothingBlocks     []int
	Along             map[int]domain.Progress
	Contexts          map[int]string
	Config            *Config
	Sheet             *trace.Sheet
	Named             map[int]string
}

// Do runs an operation: it opens the store the way op.Needs says to, hands Run an engine open on it — or none, when nothing was asked for — and closes what it opened on the way out, whatever Run does.
func Do(op Operation, path string, in Input) (Answer, error) {
	return DoAt(op, Places{Store: path}, in, Acting{})
}

// Acting is where an act comes from, written down with it: the surface it came through, and who did it.
type Acting struct {
	Through string
	By      string
}

// AuthorVariable is the variable that names who acts, for a script or a session to say it is the one acting.
const AuthorVariable = "CAIRN_AUTHOR"

// Author is who the acts of a surface are written down as: the one the environment names, or else the one logged in.
func Author(h Home) string {
	if named, ok := h.lookup()(AuthorVariable); ok && named != "" {
		return named
	}
	if me, err := user.Current(); err == nil {
		return me.Username
	}
	return ""
}

// Through is a surface acting for the one its home names.
func Through(surface string, h Home) Acting { return Acting{Through: surface, By: Author(h)} }

// DoAt runs an operation at the places a surface found, and says which surface asked, so what it changes is written down as that surface's doing.
func DoAt(op Operation, at Places, in Input, acting Acting) (Answer, error) {
	if err := op.admitted(in); err != nil {
		return Answer{}, err
	}
	return doAt(op, at, in, acting)
}

func doAt(op Operation, at Places, in Input, acting Acting) (Answer, error) {
	switch op.Needs {
	case AStoreBehind:
		return doWith(OpenBehind, op, at, in, acting)
	case AStore:
		return doWith(Open, op, at, in, acting)
	default:
		return op.Run(On{Backups: BackupsAt(at.Store), Path: at.Store, Notes: at.Notes, Plugin: at.Plugin, Home: at.Home}, in)
	}
}

func doWith(
	open func(string) (*Engine, error),
	op Operation, at Places, in Input, acting Acting,
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
	eng.acting, eng.notes, eng.home = acting, at.Notes, at.Home
	return eng.run(op, in)
}
