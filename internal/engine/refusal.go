package engine

import (
	"cairn/internal/engine/app"
	"cairn/internal/engine/backup"
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/change"
	"cairn/internal/engine/exchange"
	"cairn/internal/engine/item"
	"cairn/internal/engine/label"
	"cairn/internal/engine/migration"
	"cairn/internal/engine/series"
	"cairn/internal/engine/workspace"
	"errors"

	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
	"cairn/internal/linuxapp"
	"cairn/internal/macapp"
	"cairn/internal/store/files"
	"cairn/internal/store/sqlite"
	"cairn/internal/upgrade"
)

// The kinds of failure.
const (
	// KindRefused: the request was understood and cannot be satisfied.
	KindRefused = "refused"
	// KindUsage: the request was not understood — a name nothing takes, a number that is not one.
	KindUsage = "usage"
	// KindStore: the store could not be used.
	KindStore = "store"
	// KindFailed: something nobody foresaw went wrong.
	KindFailed = "failed"
)

// Refusal is how a failure crosses to a machine, on the one stream it reads, even when the answer is a refusal it cannot otherwise parse.
type Refusal struct {
	Error  string `json:"error"`
	Kind   string `json:"kind"`
	Reason string `json:"reason,omitempty"`
}

type reason struct {
	err    error
	kind   string
	reason string
}

var reasons = []reason{
	{migration.ErrStoreBehind, KindStore, "store-behind"},
	{migration.ErrNewerStore, KindStore, "store-newer"},
	{backup.ErrNotAStore, KindStore, "not-a-store"},
	{errStore, KindStore, "store-unreadable"},
	{errMisread, KindUsage, "not-understood"},
	{exchange.ErrNotAnExport, KindRefused, "not-an-export"},
	{exchange.ErrExportPointsOutside, KindRefused, "export-points-outside"},
	{item.ErrTitleRequired, KindRefused, "title-required"},
	{item.ErrTitleNotOneLine, KindRefused, "title-not-one-line"},
	{item.ErrNotText, KindRefused, "not-text"},
	{label.ErrLabelRequired, KindRefused, "label-required"},
	{label.ErrNoSuchColour, KindRefused, "no-such-colour"},
	{label.ErrNoSuchScope, KindRefused, "no-such-scope"},
	{label.ErrOutOfScope, KindRefused, "label-out-of-scope"},
	{objective.ErrWaitsOnItself, KindRefused, "waits-on-itself"},
	{objective.ErrWaitsOnRequired, KindRefused, "waits-on-required"},
	{objective.ErrWaitsOnBothWays, KindRefused, "waits-on-both-ways"},
	{objective.ErrWaitsOnSignAtBirth, KindRefused, "waits-on-sign-at-birth"},
	{label.ErrLabelNotOneLine, KindRefused, "label-not-one-line"},
	{label.ErrLabelHasNoFamily, KindRefused, "label-has-no-family"},
	{label.ErrLabelHolds, KindRefused, "label-holds"},
	{label.ErrLabelsBothWays, KindRefused, "labels-both-ways"},
	{change.ErrReplacedNowhere, KindRefused, "replaced-nowhere"},
	{calendar.ErrNotADayFromToday, KindRefused, "not-a-day-from-today"},
	{errNotMadeByFixtures, KindRefused, "not-made-by-fixtures"},
	{app.ErrNoApplicationHere, KindRefused, "no-application-here"},
	{macapp.ErrNotOurs, KindRefused, "app-not-ours"},
	{linuxapp.ErrNotOurs, KindRefused, "entry-not-ours"},
	{app.ErrPassing, KindRefused, "passing-build"},
	{app.ErrBuiltFromSource, KindRefused, "built-from-source"},
	{app.ErrKeptByHomebrew, KindRefused, "kept-by-homebrew"},
	{upgrade.ErrOutOfReach, KindRefused, "releases-out-of-reach"},
	{upgrade.ErrNoBinaryHere, KindRefused, "no-binary-here"},
	{upgrade.ErrNotItsSum, KindRefused, "not-its-sum"},
	{upgrade.ErrNotWritable, KindRefused, "not-writable"},
	{change.ErrReplacedMoreThanOnce, KindRefused, "replaced-more-than-once"},
	{errReplaceAlone, KindRefused, "replace-alone"},
	{errDescriptionBothWays, KindRefused, "description-both-ways"},
	{item.ErrNotAState, KindRefused, "not-a-state"},
	{objective.ErrNotAnObjectiveState, KindRefused, "not-an-objective-state"},
	{item.ErrNotAPriority, KindRefused, "not-a-priority"},
	{calendar.ErrNotADay, KindRefused, "not-a-day"},
	{series.ErrNotARule, KindRefused, "not-a-rule"},
	{calendar.ErrNotADuration, KindRefused, "not-a-duration"},
	{calendar.ErrNegativeAmount, KindRefused, "negative-amount"},
	{calendar.ErrFinerThanASecond, KindRefused, "finer-than-a-second"},
	{task.ErrNoSuchTask, KindRefused, "no-such-task"},
	{sqlite.ErrNoSuchSeries, KindRefused, "no-such-series"},
	{errNothingToChange, KindRefused, "nothing-to-change"},
	{errLabelSignAtCreation, KindRefused, "label-sign-at-creation"},
	{objective.ErrNoSuchObjective, KindRefused, "no-such-objective"},
	{label.ErrNoSuchLabel, KindRefused, "no-such-label"},
	{task.ErrBlockerSignAtBirth, KindRefused, "blocker-sign-at-creation"},
	{task.ErrBlockerRequired, KindRefused, "blocker-required"},
	{task.ErrBlockedByItself, KindRefused, "blocked-by-itself"},
	{task.ErrLinkedToItself, KindRefused, "linked-to-itself"},
	{task.ErrBlockingCycle, KindRefused, "blocking-cycle"},
	{task.ErrBlockersBothWays, KindRefused, "blockers-both-ways"},
	{task.ErrLinkRequired, KindRefused, "link-required"},
	{task.ErrLinksBothWays, KindRefused, "links-both-ways"},
	{task.ErrLinkSignAtBirth, KindRefused, "link-sign-at-creation"},
	{task.ErrStillBlocked, KindRefused, "still-blocked"},
	{errObjectiveRequired, KindRefused, "objective-required"},
	{errObjectivesBothWays, KindRefused, "objectives-both-ways"},
	{errObjectiveSignAtBirth, KindRefused, "objective-sign-at-creation"},
	{task.ErrNowhereToMove, KindRefused, "nowhere-to-move"},
	{objective.ErrObjectiveNowhereToGo, KindRefused, "objective-nowhere-to-go"},
	{series.ErrNotAnArrival, KindRefused, "not-an-arrival"},
	{objective.ErrObjectiveParentAndAfter, KindRefused, "objective-parent-and-after"},
	{task.ErrMoreThanOnePlace, KindRefused, "more-than-one-place"},
	{task.ErrNothingToLabel, KindRefused, "nothing-to-label"},
	{task.ErrNothingToArchive, KindRefused, "nothing-to-archive"},
	{task.ErrAlsoWithoutNumber, KindRefused, "also-without-number"},
	{calendar.ErrDayToCome, KindRefused, "day-to-come"},
	{calendar.ErrNotATime, KindRefused, "not-a-time"},
	{task.ErrMovedBehindItself, KindRefused, "moved-behind-itself"},
	{task.ErrDoneHoldsNoFocus, KindRefused, "done-holds-no-focus"},
	{task.ErrOtherColumn, KindRefused, "other-column"},
	{task.ErrAlreadyArchived, KindRefused, "already-archived"},
	{task.ErrNotArchived, KindRefused, "not-archived"},
	{task.ErrStepStillOpen, KindRefused, "step-still-open"},
	{task.ErrNoSuchStep, KindRefused, "no-such-step"},
	{workspace.ErrConfigAlreadyThere, KindRefused, "config-already-there"},
	{task.ErrDoneAndNotDone, KindRefused, "done-and-not-done"},
	{objective.ErrObjectiveCarriesTasks, KindRefused, "objective-carries-tasks"},
	{objective.ErrObjectiveHoldsSubObjectives, KindRefused, "objective-holds-sub-objectives"},
	{objective.ErrObjectiveNamedByAView, KindRefused, "objective-named-by-a-view"},
	{objective.ErrNotASubObjective, KindRefused, "not-a-sub-objective"},
	{objective.ErrTooDeep, KindRefused, "objective-too-deep"},
	{objective.ErrUnderItself, KindRefused, "objective-under-itself"},
	{objective.ErrCountedTwice, KindRefused, "counted-twice"},
	{backup.ErrNoSuchBackup, KindRefused, "no-such-backup"},
	{backup.ErrNoBackup, KindRefused, "no-backup"},
	{backup.ErrDamagedBackup, KindRefused, "damaged-backup"},
	{migration.ErrNoSuchChange, KindRefused, "no-such-change"},
	{migration.ErrNothingWaiting, KindRefused, "nothing-waiting"},
	{migration.ErrAlreadyApplied, KindRefused, "already-applied"},
	{migration.ErrNothingToRollBack, KindRefused, "nothing-to-roll-back"},
	{migration.ErrUndoneMeanwhile, KindRefused, "undone-meanwhile"},
	{migration.ErrEarlierWaiting, KindRefused, "earlier-change-waiting"},
	{errNotOneOf, KindRefused, "not-one-of"},
	{errNotAField, KindRefused, "not-a-field"},
	{errNotALevel, KindRefused, "not-a-level"},
	{files.ErrNotTOML, KindRefused, "config-not-toml"},
	{files.ErrUnknownKey, KindRefused, "config-unknown-key"},
	{files.ErrNotAbsolute, KindRefused, "config-not-absolute"},
	{files.ErrArchive, KindRefused, "config-archive"},
	{files.ErrFocus, KindRefused, "config-focus"},
	{files.ErrTime, KindRefused, "config-time"},
	{files.ErrWeekends, KindRefused, "config-weekends"},
	{files.ErrTimeMode, KindRefused, "config-time-mode"},
	{files.ErrKeepSizes, KindRefused, "config-window"},
	{files.ErrIntro, KindRefused, "config-intro"},
	{files.ErrGhost, KindRefused, "config-ghost"},
	{files.ErrRecurrence, KindRefused, "config-recurrence"},
	{files.ErrNotWritten, KindRefused, "config-not-written"},
	{files.ErrMachineKey, KindRefused, "config-machine-key"},
	{files.ErrEmpty, KindRefused, "config-empty"},
	{files.ErrNotAKey, KindRefused, "config-not-a-key"},
	{files.ErrValue, KindRefused, "config-value"},
	{files.ErrNotEdited, KindRefused, "config-not-edited"},
	{workspace.ErrStoreKey, KindRefused, "config-store-key"},
	{workspace.ErrAnsweredKey, KindRefused, "config-answered-key"},
	{workspace.ErrEmptyGlobal, KindRefused, "config-empty-global"},
	{files.ErrNotSet, KindRefused, "config-not-set"},
	{workspace.ErrNoSuchWorkspace, KindRefused, "no-such-workspace"},
	{workspace.ErrWorkspaceNotThere, KindRefused, "workspace-not-there"},
	{workspace.ErrWorkspaceKnown, KindRefused, "workspace-known"},
	{workspace.ErrNotAWorkspaceName, KindRefused, "not-a-workspace-name"},
	{workspace.ErrRemovingTheDefault, KindRefused, "removing-the-default"},
	{files.ErrNotAWorkspace, KindRefused, "config-not-a-workspace"},
	{files.ErrNotAWorkspaceStore, KindRefused, "config-not-a-workspace-store"},
	{workspace.ErrNotesNotMoved, KindRefused, "notes-not-moved"},
	{workspace.ErrNotThere, KindRefused, "not-there"},
}

// RefusalOf says what kind of answer a failure is, and which one when cairn foresaw it.
func RefusalOf(err error) Refusal {
	said := Refusal{Error: err.Error(), Kind: KindFailed}
	for _, one := range reasons {
		if errors.Is(err, one.err) {
			said.Kind, said.Reason = one.kind, one.reason
			return said
		}
	}
	var named *refusing
	if errors.As(err, &named) {
		said.Kind, said.Reason = KindRefused, named.reason
		return said
	}
	if storeUnusable(err) {
		said.Kind, said.Reason = KindStore, "store-unusable"
	}
	return said
}

var errMisread = errors.New("not understood")

type misread struct{ err error }

func (m misread) Error() string { return m.err.Error() }

// Is lets errors.Is find the mark without the sentence carrying it.
func (m misread) Is(target error) bool { return target == errMisread }

func (m misread) Unwrap() error { return m.err }

// Misread marks an error as the request not having been understood, keeping its words.
func Misread(err error) error { return misread{err: err} }

type refusing struct{ reason, sentence string }

func (r *refusing) Error() string { return r.sentence }

// Refusing makes a refusal a surface foresaw: what it says, and the words a machine branches on.
func Refusing(reason, sentence string) error { return &refusing{reason: reason, sentence: sentence} }

// Agreeing is what a refusal asks the person, and the act that goes ahead once they agree.
type Agreeing struct {
	Reason string
	Op     Operation
	In     Input
}

// AgreeingTo says whether a refusal is a question for the person, and the act that agrees to it — none when the person can only leave.
func AgreeingTo(op Operation, in Input, err error) (Agreeing, bool) {
	if errors.Is(err, migration.ErrStoreBehind) {
		return Agreeing{Reason: RefusalOf(err).Reason, Op: MigrationApply, In: Input{}}, true
	}
	if errors.Is(err, migration.ErrNewerStore) {
		return Agreeing{Reason: RefusalOf(err).Reason}, true
	}
	if !op.Finishes || in.Confirmed || !errors.Is(err, ErrStepStillOpen) {
		return Agreeing{}, false
	}
	return Agreeing{Reason: RefusalOf(err).Reason, Op: op, In: Agreed(in)}, true
}
