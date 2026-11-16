package engine

import (
	"cairn/internal/engine/app"
	"cairn/internal/engine/backup"
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/item"
	"cairn/internal/engine/label"
	"cairn/internal/engine/series"
	"cairn/internal/engine/view"
	"cairn/internal/engine/workspace"
	"errors"
	"fmt"
	"os/user"
	"strconv"
	"strings"
	"time"

	"cairn/internal/engine/act"
	"cairn/internal/engine/migration"
	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
)

// Level is how far an agent may reach.
type Level int

const (
	_ Level = iota
	// Read is level one: listing and showing, nothing that changes anything.
	Read
	// write is level two: also adding and changing, and what sets a backup aside first.
	write
	// Full is level three: also what loses work already done.
	Full
)

// errNotALevel is the reason a level is refused: what was given spelled none of "read", "write" or "full".
var errNotALevel = errors.New("not one of the access levels")

// ParseLevel reads a level by name, the way a person types it when starting a server.
func ParseLevel(name string) (Level, error) {
	switch name {
	case "read":
		return Read, nil
	case "write":
		return write, nil
	case "full":
		return Full, nil
	}
	return 0, fmt.Errorf("%q: expected \"read\", \"write\" or \"full\": %w", name, errNotALevel)
}

// again is what calling an operation a second time, with the same arguments, does to the store.
type again int

const (
	_ again = iota
	// ChangesNothingMore is a second call leaving the store as the first did — or being refused for having nothing left to do.
	ChangesNothingMore
	// doesItAgain is a second call adding again — another task, another backup, more time spent, one more change undone.
	doesItAgain
)

// needs says which way the store has to be opened.
type needs int

const (
	_ needs = iota
	// aStore refuses a store behind this cairn: a request about what it holds does not change its shape on the way past.
	aStore
	// aStoreBehind opens the store whatever shape it is in, for the operations whose job is to change that shape.
	aStoreBehind
	// noStore opens nothing at all, for what is reachable without the store — a backup is most needed exactly when the store cannot be opened.
	noStore
	// aConfiguration opens no store and has no use for its path either: it reads the configuration itself, as config show and the workspace commands do, so a store the configuration cannot find yet stops none of them.
	aConfiguration
)

// Kind is what sort of value an argument carries.
type Kind int

const (
	_ Kind = iota
	// Text is a plain string.
	Text
	// Number is a whole number.
	Number
	// day is a calendar day.
	day
	// duration is a length of time.
	duration
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
	Repeats    bool
	Bare       string
}

// answerShape is the one shape an operation's answer takes — an empty value of it, carried so a surface can read the shape off the type.
type answerShape any

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

// yes reports whether a Kind: Bool argument was named as true.
func (in Input) yes(name string) bool { return in.Given[name] == "true" }

// change answers what to hand Changes for a name: nothing when it was not named, and what it was named as otherwise.
func (in Input) change(name string) *string {
	given, named := in.Given[name]
	if !named {
		return nil
	}
	return &given
}

// changed answers what to hand Changes for a list of strings, on the same three states as Change.
func (in Input) changed(name string) *[]string {
	listed, named := in.Listed[name]
	if !named {
		return nil
	}
	return &listed
}

// whole answers the number that was named for a Kind: Number argument, and whether it was named at all.
func (in Input) whole(name string) (int, bool) {
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
	Needs      needs
	Again      again
	Overwrites bool
	Takes      []Argument
	Run        func(workedOn, Input) (Answer, error)
	Answers    answerShape
	Sweeps     func(Input) bool
	Finishes   bool
	Repeats    bool
}

// Agreed is what was given, said yes to: what a surface hands over once the person it asked has agreed.
func Agreed(in Input) Input {
	in.Confirmed = true
	return in
}

// errNotConfirmed is an act that asks first, asked for without saying yes.
var errNotConfirmed = Refusing("not-confirmed", "needs --yes to confirm")

// ErrStepStillOpen is returned when a task is put in done while one of its steps is not, and the act was not confirmed.
var ErrStepStillOpen = task.ErrStepStillOpen

// asksFirst says whether an act waits to be confirmed: one that loses work, or one that sweeps up whatever a narrowing reaches.
func (o Operation) asksFirst(in Input) bool {
	return o.Level == Full || o.Sweeps != nil && o.Sweeps(in)
}

func (o Operation) admitted(in Input) error {
	if err := o.check(in); err != nil {
		return err
	}
	return o.waiting(in)
}

func (o Operation) waiting(in Input) error {
	if !o.asksFirst(in) || in.Confirmed {
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
	return fmt.Errorf("%s %w", target, errNotConfirmed)
}

// Name is how an operation is named as a command or a tool: "task_new".
func (o Operation) Name() string { return strings.TrimPrefix(o.Thing+"_"+o.Verb, "_") }

// Line is how an operation reads on the command line: "task new".
func (o Operation) Line() string { return strings.TrimSpace(o.Thing + " " + o.Verb) }

// All is everything cairn can be asked to do, in a stable order.
func All() []Operation {
	return []Operation{
		TaskNew, TaskList, TaskShow, TaskEdit, TaskDelete, TaskSearch,
		TaskArchive, TaskUnarchive, TaskLabel,
		TaskMove, TaskFocus, TaskUnfocus, TaskExport, TaskImport, TaskSkip,
		StepNew, stepList, StepEdit, StepDelete,
		ObjectiveNew, ObjectiveList, ObjectiveShow, objectiveReport, ObjectiveEdit, ObjectiveMove, ObjectiveDelete, ObjectiveSearch,
		BackupNew, BackupList, backupShow, backupDelete, BackupRestore,
		MigrationList, migrationShow, MigrationApply, MigrationRollback,
		LabelList, LabelNew, LabelColour, labelScope, LabelDelete, LabelRename,
		ViewList,
		SeriesList, SeriesShow, SeriesEdit, SeriesStop, SeriesResume, SeriesDelete,
		TraceShow, traceList,
		TimesheetShow, TimesheetAdd,
		ConfigShow, ConfigInit, debugReport,
		WorkspaceList, WorkspaceShow,
		pluginWrite, pluginShow, appInstall,
	}
}

// ForAPersonOnly is what only a person at the command line may do, offered to no agent: adding a workspace, switching to another and removing one change where every later command reads and writes; a view is how the person looks at their work, so an agent reads the views and keeps, changes or deletes none.
func ForAPersonOnly() []Operation {
	return []Operation{WorkspaceAdd, WorkspaceUse, workspaceRemove, NotesMove, ViewNew, viewShow, ViewEdit, ViewDelete, ConfigSet, ConfigUnset, ConfigFill, Upgrade}
}

// check refuses what Run could not read anyway, before the store is opened so a refusal never touches the disk:.
func (o Operation) check(in Input) error {
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
		case day:
			if _, err := calendar.ParseDeadline(given); err != nil {
				return err
			}
		}
	}
	return nil
}

// workedOn is what an operation was given to work on.
type workedOn struct {
	Engine  *Engine
	Backups backup.Vault
	Path    string
	Notes   string
	Plugin  string
	Home    Home
}

// Answer is what an operation answered.
type Answer struct {
	Task              *task.Task
	Sessions          []focusSession
	Asked             []string
	Context           string
	StoodSince        time.Time
	Tasks             []task.Task
	Matched           []task.Task
	Terms             map[string][]int
	Unknown           []int
	Counts            []count
	Total             *int
	Offset            int
	Fields            []string
	Under             []task.Task
	Family            *objective.Family
	Tree              map[int][]int
	Waiters           map[int][]int
	Warnings          []string
	Spent             map[int]time.Duration
	Proposes          *objective.Objective
	Went              *objective.WentWith
	Objective         *objective.Objective
	Objectives        []objective.Objective
	MatchedObjectives []objective.Objective
	Backup            *backup.Backup
	Backups           []backup.Backup
	Change            *migration.Migration
	Changes           []migration.Migration
	RolledBack        *migration.Migration
	Progress          item.Progress
	Holds             *backup.Held
	Deleted           *int
	Also              []int
	Unblocked         []int
	Restored          *int
	Exported          *string
	Imported          *string
	Plugin            *string
	Install           *app.PluginShown
	Written           *string
	Archived          *int
	Labelled          *int
	Labels            *[]label.Carried
	Applied           []string
	Since             []act.Act
	Acts              []act.Act
	InFocus           *[]int
	Held              []task.Task
	NothingBlocks     []int
	Along             map[int]item.Progress
	Contexts          map[int]string
	Config            *workspace.Config
	ConfigChanged     *workspace.ConfigChanged
	NotesMoved        *workspace.NotesMove
	Sheet             *act.Sheet
	Period            *objective.Period
	Named             map[int]string
	Gone              map[int]bool
	Packed            *int
	Moved             *moves
	Crossed           *task.Crossed
	Installed         *string
	Upgraded          *app.Upgraded
	Runs              *string
	OpenedFrom        *string
	Workspaces        *workspace.Workspaces
	WorkspaceAdded    *workspace.Added
	WorkspaceUsed     *workspace.Used
	WorkspaceRemoved  *workspace.Removed
	WorkspaceShown    *workspace.Shown
	Viewed            *View
	ViewsKept         *Views
	ViewGone          *string
	InView            string
	Lost              []string
	KeepsNothing      bool
	Outside           []view.Outside
	Occurring         map[int]series.Occurs
	Choosing          map[int]task.Choosing
	Followed          []series.MadeNext
	SeriesShown       *Series
	SeriesKept        *SeriesListed
	SeriesGone        *series.Deleted
	Untimed           bool
	Kept              *string
}

// Do runs an operation: it opens the store the way op.Needs says to, hands Run an engine open on it — or none, when nothing was asked for — and closes what it opened on the way out, whatever Run does.
func Do(op Operation, path string, in Input) (Answer, error) {
	return doThere(op, Places{Store: path}, in, Acting{})
}

// Acting is where an act comes from, written down with it: the surface it came through, and who did it.
type Acting struct {
	Through string
	By      string
}

// AuthorVariable is the variable that names who acts, for a script or a session to say it is the one acting.
const AuthorVariable = "CAIRN_AUTHOR"

// author is who the acts of a surface are written down as: the one the environment names, or else the one logged in.
func author(h Home) string {
	if named, ok := h.lookup()(AuthorVariable); ok && named != "" {
		return named
	}
	if me, err := user.Current(); err == nil {
		return me.Username
	}
	return ""
}

// Through is a surface acting for the one its home names.
func Through(surface string, h Home) Acting { return Acting{Through: surface, By: author(h)} }

// doThere runs an operation at the places a surface found, and says which surface asked, so what it changes is written down as that surface's doing.
func doThere(op Operation, at Places, in Input, acting Acting) (Answer, error) {
	return admittedAndNoted(op, in, func() (Answer, error) { return doAt(op, at, in, acting) })
}

func doAt(op Operation, at Places, in Input, acting Acting) (Answer, error) {
	switch op.Needs {
	case aStoreBehind:
		return doWith(OpenBehind, op, at, in, acting)
	case aStore:
		return doWith(Open, op, at, in, acting)
	default:
		return op.Run(workedOn{Backups: backupsAt(at.Store), Path: at.Store, Notes: at.Notes, Plugin: at.Plugin, Home: at.Home}, in)
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
	if op.Thing == TaskList.Thing && op.Verb == TaskList.Verb {
		if err := eng.archivingSwept(); err != nil {
			return Answer{}, storeProblem(err)
		}
	}
	if op.Needs == aStore {
		if err := eng.timerKept(); err != nil {
			return Answer{}, storeProblem(err)
		}
		if err := eng.seriesDue(); err != nil {
			return Answer{}, storeProblem(err)
		}
	}
	answer, err = eng.run(op, in)
	if err != nil || op.Needs != aStore {
		return answer, err
	}
	answer.Asked, err = eng.asked()
	return answer, err
}
