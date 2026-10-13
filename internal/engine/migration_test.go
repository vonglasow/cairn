package engine

import (
	"strconv"
	"testing"
	"time"

	"cairn/internal/migration"
)

// changeFacts reads back what a test needs from a change, without reaching
// into fields that stopped being exported.
type changeFacts struct {
	number  int
	name    string
	does    string
	applied bool
}

func (f *changeFacts) Numbered(n int)    { f.number = n }
func (f *changeFacts) Named(s string)    { f.name = s }
func (f *changeFacts) Does(s string)     { f.does = s }
func (f *changeFacts) Undoes(string)     {}
func (f *changeFacts) NotUndoable()      {}
func (f *changeFacts) Loses(string)      {}
func (f *changeFacts) NothingLost()      {}
func (f *changeFacts) Applied(time.Time) { f.applied = true }
func (f *changeFacts) NotApplied()       { f.applied = false }

func factsOfChange(m migration.Migration) changeFacts {
	var f changeFacts
	m.Tell(&f)
	return f
}

func TestMigrationShow_saysWhatAChangeDoes(t *testing.T) {
	path := storeAt(t)

	answer, err := Do(MigrationShow, path, Input{Given: map[string]string{"number": strconv.Itoa(1)}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Change == nil || factsOfChange(*answer.Change).number != 1 {
		t.Fatalf("answer.Change = %v, want change 1", answer.Change)
	}
}

func TestMigrationRollback_unmakesTheLastChangeApplied(t *testing.T) {
	path := storeAt(t)

	answer, err := Do(MigrationRollback, path, Input{Confirmed: true})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.RolledBack == nil {
		t.Fatalf("answer.RolledBack = nil, want the change that was rolled back")
	}

	list, err := Do(MigrationList, path, Input{})
	if err != nil {
		t.Fatalf("Do(MigrationList): %v", err)
	}
	rolledBack := factsOfChange(*answer.RolledBack)
	for _, one := range list.Changes {
		c := factsOfChange(one)
		if c.number == rolledBack.number && c.applied {
			t.Fatalf("change %d is still applied, want it rolled back", c.number)
		}
	}
}

func TestMigrationApply_appliesWhatWasNamedOrEverythingWaiting(t *testing.T) {
	upToDate := storeAt(t)

	// Nothing waiting is a success, not a refusal: Applied is empty, not nil.
	answer, err := Do(MigrationApply, upToDate, Input{})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Applied == nil || len(answer.Applied) != 0 {
		t.Fatalf("answer.Applied = %v, want an empty, non-nil set", answer.Applied)
	}

	// A number names exactly one change to make, not every change waiting.
	numbered := behindStore(t)
	waiting, err := Do(MigrationList, numbered, Input{})
	if err != nil {
		t.Fatalf("Do(MigrationList): %v", err)
	}
	var missing int
	for _, one := range waiting.Changes {
		c := factsOfChange(one)
		if !c.applied {
			missing = c.number
		}
	}
	if missing == 0 {
		t.Fatalf("waiting.Changes = %v, want one change not applied", waiting.Changes)
	}
	byNumber, err := Do(MigrationApply, numbered, Input{Given: map[string]string{"number": strconv.Itoa(missing)}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if len(byNumber.Applied) != 1 {
		t.Fatalf("byNumber.Applied = %v, want exactly the one change named", byNumber.Applied)
	}

	behind := behindStore(t)
	applied, err := Do(MigrationApply, behind, Input{})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if len(applied.Applied) != 1 {
		t.Fatalf("applied.Applied = %v, want exactly one change made", applied.Applied)
	}
}

func TestMigrationList_listsChangesAndWhatIsHeld(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskNew, path, Input{Given: map[string]string{"title": "one"}}); err != nil {
		t.Fatalf("Do(TaskNew): %v", err)
	}

	answer, err := Do(MigrationList, path, Input{})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if len(answer.Changes) == 0 {
		t.Fatalf("answer.Changes = %v, want at least the first change", answer.Changes)
	}
	if answer.Holds == nil || answer.Holds.Tasks != 1 {
		t.Fatalf("answer.Holds = %v, want a count of 1 task", answer.Holds)
	}
}
