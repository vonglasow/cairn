package engine

import (
	"errors"

	"cairn/internal/config"
	"cairn/internal/domain"
	"cairn/internal/linuxapp"
	"cairn/internal/macapp"
	"cairn/internal/objective"
	"cairn/internal/store"
	"cairn/internal/task"
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
	{ErrStoreBehind, KindStore, "store-behind"},
	{ErrNewerStore, KindStore, "store-newer"},
	{errNotAStore, KindStore, "not-a-store"},
	{errStore, KindStore, "store-unreadable"},
	{errMisread, KindUsage, "not-understood"},
	{errNotAnExport, KindRefused, "not-an-export"},
	{errExportPointsOutside, KindRefused, "export-points-outside"},
	{domain.ErrTitleRequired, KindRefused, "title-required"},
	{domain.ErrTitleNotOneLine, KindRefused, "title-not-one-line"},
	{domain.ErrNotText, KindRefused, "not-text"},
	{domain.ErrLabelRequired, KindRefused, "label-required"},
	{domain.ErrNoSuchColour, KindRefused, "no-such-colour"},
	{domain.ErrNoSuchScope, KindRefused, "no-such-scope"},
	{domain.ErrOutOfScope, KindRefused, "label-out-of-scope"},
	{objective.ErrWaitsOnItself, KindRefused, "waits-on-itself"},
	{errWaitsOnRequired, KindRefused, "waits-on-required"},
	{errWaitsOnBothWays, KindRefused, "waits-on-both-ways"},
	{errWaitsOnSignAtBirth, KindRefused, "waits-on-sign-at-birth"},
	{domain.ErrLabelNotOneLine, KindRefused, "label-not-one-line"},
	{domain.ErrLabelHasNoFamily, KindRefused, "label-has-no-family"},
	{domain.ErrLabelHolds, KindRefused, "label-holds"},
	{domain.ErrLabelsBothWays, KindRefused, "labels-both-ways"},
	{domain.ErrReplacedNowhere, KindRefused, "replaced-nowhere"},
	{domain.ErrNotADayFromToday, KindRefused, "not-a-day-from-today"},
	{errNotMadeByFixtures, KindRefused, "not-made-by-fixtures"},
	{errNoApplicationHere, KindRefused, "no-application-here"},
	{macapp.ErrNotOurs, KindRefused, "app-not-ours"},
	{linuxapp.ErrNotOurs, KindRefused, "entry-not-ours"},
	{errPassing, KindRefused, "passing-build"},
	{errBuiltFromSource, KindRefused, "built-from-source"},
	{errKeptByHomebrew, KindRefused, "kept-by-homebrew"},
	{upgrade.ErrOutOfReach, KindRefused, "releases-out-of-reach"},
	{upgrade.ErrNoBinaryHere, KindRefused, "no-binary-here"},
	{upgrade.ErrNotItsSum, KindRefused, "not-its-sum"},
	{upgrade.ErrNotWritable, KindRefused, "not-writable"},
	{domain.ErrReplacedMoreThanOnce, KindRefused, "replaced-more-than-once"},
	{errReplaceAlone, KindRefused, "replace-alone"},
	{errDescriptionBothWays, KindRefused, "description-both-ways"},
	{domain.ErrNotAState, KindRefused, "not-a-state"},
	{objective.ErrNotAnObjectiveState, KindRefused, "not-an-objective-state"},
	{domain.ErrNotAPriority, KindRefused, "not-a-priority"},
	{domain.ErrNotADay, KindRefused, "not-a-day"},
	{domain.ErrNotARule, KindRefused, "not-a-rule"},
	{domain.ErrNotADuration, KindRefused, "not-a-duration"},
	{domain.ErrNegativeAmount, KindRefused, "negative-amount"},
	{domain.ErrFinerThanASecond, KindRefused, "finer-than-a-second"},
	{errNoSuchTask, KindRefused, "no-such-task"},
	{store.ErrNoSuchSeries, KindRefused, "no-such-series"},
	{errNothingToChange, KindRefused, "nothing-to-change"},
	{errLabelSignAtCreation, KindRefused, "label-sign-at-creation"},
	{errNoSuchObjective, KindRefused, "no-such-objective"},
	{errNoSuchLabel, KindRefused, "no-such-label"},
	{errBlockerSignAtBirth, KindRefused, "blocker-sign-at-creation"},
	{errBlockerRequired, KindRefused, "blocker-required"},
	{errBlockedByItself, KindRefused, "blocked-by-itself"},
	{errLinkedToItself, KindRefused, "linked-to-itself"},
	{errBlockingCycle, KindRefused, "blocking-cycle"},
	{errBlockersBothWays, KindRefused, "blockers-both-ways"},
	{errLinkRequired, KindRefused, "link-required"},
	{errLinksBothWays, KindRefused, "links-both-ways"},
	{errLinkSignAtBirth, KindRefused, "link-sign-at-creation"},
	{errStillBlocked, KindRefused, "still-blocked"},
	{errObjectiveRequired, KindRefused, "objective-required"},
	{errObjectivesBothWays, KindRefused, "objectives-both-ways"},
	{errObjectiveSignAtBirth, KindRefused, "objective-sign-at-creation"},
	{errNowhereToMove, KindRefused, "nowhere-to-move"},
	{errObjectiveNowhereToGo, KindRefused, "objective-nowhere-to-go"},
	{errNotAnArrival, KindRefused, "not-an-arrival"},
	{errObjectiveParentAndAfter, KindRefused, "objective-parent-and-after"},
	{errMoreThanOnePlace, KindRefused, "more-than-one-place"},
	{errNothingToLabel, KindRefused, "nothing-to-label"},
	{errNothingToArchive, KindRefused, "nothing-to-archive"},
	{errAlsoWithoutNumber, KindRefused, "also-without-number"},
	{errDayToCome, KindRefused, "day-to-come"},
	{errNotATime, KindRefused, "not-a-time"},
	{task.ErrMovedBehindItself, KindRefused, "moved-behind-itself"},
	{errDoneHoldsNoFocus, KindRefused, "done-holds-no-focus"},
	{task.ErrOtherColumn, KindRefused, "other-column"},
	{task.ErrAlreadyArchived, KindRefused, "already-archived"},
	{task.ErrNotArchived, KindRefused, "not-archived"},
	{task.ErrStepStillOpen, KindRefused, "step-still-open"},
	{task.ErrNoSuchStep, KindRefused, "no-such-step"},
	{errConfigAlreadyThere, KindRefused, "config-already-there"},
	{errDoneAndNotDone, KindRefused, "done-and-not-done"},
	{errObjectiveCarriesTasks, KindRefused, "objective-carries-tasks"},
	{errObjectiveHoldsSubObjectives, KindRefused, "objective-holds-sub-objectives"},
	{errObjectiveNamedByAView, KindRefused, "objective-named-by-a-view"},
	{errNotASubObjective, KindRefused, "not-a-sub-objective"},
	{objective.ErrTooDeep, KindRefused, "objective-too-deep"},
	{objective.ErrUnderItself, KindRefused, "objective-under-itself"},
	{objective.ErrCountedTwice, KindRefused, "counted-twice"},
	{errNoSuchBackup, KindRefused, "no-such-backup"},
	{errNoBackup, KindRefused, "no-backup"},
	{errDamagedBackup, KindRefused, "damaged-backup"},
	{errNoSuchChange, KindRefused, "no-such-change"},
	{errNothingWaiting, KindRefused, "nothing-waiting"},
	{errAlreadyApplied, KindRefused, "already-applied"},
	{errNothingToRollBack, KindRefused, "nothing-to-roll-back"},
	{errUndoneMeanwhile, KindRefused, "undone-meanwhile"},
	{errEarlierWaiting, KindRefused, "earlier-change-waiting"},
	{errNotOneOf, KindRefused, "not-one-of"},
	{errNotAField, KindRefused, "not-a-field"},
	{errNotALevel, KindRefused, "not-a-level"},
	{config.ErrNotTOML, KindRefused, "config-not-toml"},
	{config.ErrUnknownKey, KindRefused, "config-unknown-key"},
	{config.ErrNotAbsolute, KindRefused, "config-not-absolute"},
	{config.ErrArchive, KindRefused, "config-archive"},
	{config.ErrFocus, KindRefused, "config-focus"},
	{config.ErrTime, KindRefused, "config-time"},
	{config.ErrWeekends, KindRefused, "config-weekends"},
	{config.ErrTimeMode, KindRefused, "config-time-mode"},
	{config.ErrKeepSizes, KindRefused, "config-window"},
	{config.ErrIntro, KindRefused, "config-intro"},
	{config.ErrGhost, KindRefused, "config-ghost"},
	{config.ErrRecurrence, KindRefused, "config-recurrence"},
	{config.ErrNotWritten, KindRefused, "config-not-written"},
	{config.ErrMachineKey, KindRefused, "config-machine-key"},
	{config.ErrEmpty, KindRefused, "config-empty"},
	{config.ErrNotAKey, KindRefused, "config-not-a-key"},
	{config.ErrValue, KindRefused, "config-value"},
	{config.ErrNotEdited, KindRefused, "config-not-edited"},
	{errStoreKey, KindRefused, "config-store-key"},
	{errAnsweredKey, KindRefused, "config-answered-key"},
	{errEmptyGlobal, KindRefused, "config-empty-global"},
	{config.ErrNotSet, KindRefused, "config-not-set"},
	{errNoSuchWorkspace, KindRefused, "no-such-workspace"},
	{ErrWorkspaceNotThere, KindRefused, "workspace-not-there"},
	{ErrWorkspaceKnown, KindRefused, "workspace-known"},
	{errNotAWorkspaceName, KindRefused, "not-a-workspace-name"},
	{errRemovingTheDefault, KindRefused, "removing-the-default"},
	{config.ErrNotAWorkspace, KindRefused, "config-not-a-workspace"},
	{config.ErrNotAWorkspaceStore, KindRefused, "config-not-a-workspace-store"},
	{errNotesNotMoved, KindRefused, "notes-not-moved"},
	{errNotThere, KindRefused, "not-there"},
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
	if errors.Is(err, ErrStoreBehind) {
		return Agreeing{Reason: RefusalOf(err).Reason, Op: MigrationApply, In: Input{}}, true
	}
	if errors.Is(err, ErrNewerStore) {
		return Agreeing{Reason: RefusalOf(err).Reason}, true
	}
	if !op.Finishes || in.Confirmed || !errors.Is(err, ErrStepStillOpen) {
		return Agreeing{}, false
	}
	return Agreeing{Reason: RefusalOf(err).Reason, Op: op, In: Agreed(in)}, true
}
