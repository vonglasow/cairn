package engine

import (
	"errors"

	"cairn/internal/config"
	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
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
	{ErrNotAStore, KindStore, "not-a-store"},
	{ErrStore, KindStore, "store-unreadable"},
	{errMisread, KindUsage, "not-understood"},
	{ErrNotAnExport, KindRefused, "not-an-export"},
	{ErrExportPointsOutside, KindRefused, "export-points-outside"},
	{domain.ErrTitleRequired, KindRefused, "title-required"},
	{domain.ErrTitleNotOneLine, KindRefused, "title-not-one-line"},
	{domain.ErrNotText, KindRefused, "not-text"},
	{domain.ErrLabelRequired, KindRefused, "label-required"},
	{domain.ErrLabelNotOneLine, KindRefused, "label-not-one-line"},
	{domain.ErrLabelHasNoFamily, KindRefused, "label-has-no-family"},
	{domain.ErrLabelHolds, KindRefused, "label-holds"},
	{domain.ErrLabelsBothWays, KindRefused, "labels-both-ways"},
	{domain.ErrNotAState, KindRefused, "not-a-state"},
	{objective.ErrNotAnObjectiveState, KindRefused, "not-an-objective-state"},
	{domain.ErrNotAPriority, KindRefused, "not-a-priority"},
	{domain.ErrNotADay, KindRefused, "not-a-day"},
	{domain.ErrNotADuration, KindRefused, "not-a-duration"},
	{domain.ErrNegativeAmount, KindRefused, "negative-amount"},
	{domain.ErrFinerThanASecond, KindRefused, "finer-than-a-second"},
	{ErrNoSuchTask, KindRefused, "no-such-task"},
	{ErrNothingToChange, KindRefused, "nothing-to-change"},
	{ErrLabelSignAtCreation, KindRefused, "label-sign-at-creation"},
	{ErrParentIsSelf, KindRefused, "parent-is-self"},
	{ErrParentIsSubTask, KindRefused, "parent-is-sub-task"},
	{ErrNoSuchObjective, KindRefused, "no-such-objective"},
	{ErrBlockerSignAtBirth, KindRefused, "blocker-sign-at-creation"},
	{ErrBlockerRequired, KindRefused, "blocker-required"},
	{ErrBlockedByItself, KindRefused, "blocked-by-itself"},
	{ErrBlockingCycle, KindRefused, "blocking-cycle"},
	{ErrBlockersBothWays, KindRefused, "blockers-both-ways"},
	{ErrStillBlocked, KindRefused, "still-blocked"},
	{ErrObjectiveRequired, KindRefused, "objective-required"},
	{ErrObjectivesBothWays, KindRefused, "objectives-both-ways"},
	{ErrObjectiveSignAtBirth, KindRefused, "objective-sign-at-creation"},
	{ErrNowhereToMove, KindRefused, "nowhere-to-move"},
	{ErrMoreThanOnePlace, KindRefused, "more-than-one-place"},
	{ErrNothingToLabel, KindRefused, "nothing-to-label"},
	{ErrNothingToArchive, KindRefused, "nothing-to-archive"},
	{ErrAlsoWithoutNumber, KindRefused, "also-without-number"},
	{ErrDayToCome, KindRefused, "day-to-come"},
	{ErrNotATime, KindRefused, "not-a-time"},
	{task.ErrMovedBehindItself, KindRefused, "moved-behind-itself"},
	{task.ErrOtherColumn, KindRefused, "other-column"},
	{task.ErrNotAmongItsSisters, KindRefused, "not-among-its-sisters"},
	{task.ErrAlreadyArchived, KindRefused, "already-archived"},
	{task.ErrNotArchived, KindRefused, "not-archived"},
	{task.ErrPartGoesWithItsWork, KindRefused, "part-goes-with-its-work"},
	{task.ErrSubTaskHasNoColumn, KindRefused, "sub-task-has-no-column"},
	{task.ErrSubTaskStillOpen, KindRefused, "sub-task-still-open"},
	{ErrSubTaskServesNoObjective, KindRefused, "sub-task-serves-no-objective"},
	{ErrObjectiveCarriesTasks, KindRefused, "objective-carries-tasks"},
	{ErrNoSuchBackup, KindRefused, "no-such-backup"},
	{ErrNoBackup, KindRefused, "no-backup"},
	{ErrDamagedBackup, KindRefused, "damaged-backup"},
	{ErrNoSuchChange, KindRefused, "no-such-change"},
	{ErrNothingWaiting, KindRefused, "nothing-waiting"},
	{ErrAlreadyApplied, KindRefused, "already-applied"},
	{ErrNothingToRollBack, KindRefused, "nothing-to-roll-back"},
	{ErrUndoneMeanwhile, KindRefused, "undone-meanwhile"},
	{ErrEarlierWaiting, KindRefused, "earlier-change-waiting"},
	{ErrNotOneOf, KindRefused, "not-one-of"},
	{ErrNotALevel, KindRefused, "not-a-level"},
	{config.ErrNotTOML, KindRefused, "config-not-toml"},
	{config.ErrUnknownKey, KindRefused, "config-unknown-key"},
	{config.ErrNotAbsolute, KindRefused, "config-not-absolute"},
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
	if StoreUnusable(err) {
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
