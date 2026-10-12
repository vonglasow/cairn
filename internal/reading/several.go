package reading

// What a board says before several things go in one act, and how it asks for
// them. Read off what the board already holds: what would go with a task, what
// waits on it, what serves an objective — the facts the engine goes by, said
// before rather than found out after.

import (
	"fmt"
	"slices"
	"strconv"

	"cairn/internal/engine"
)

// Flag is something one of several things about to go carries that somebody
// should know first: what it would take with it, or why it cannot go at all.
type Flag struct {
	Number  int
	Says    string
	Refused bool
}

// DeletingTasks is what a deletion of several tasks names first: the parts
// that go with one, a task waiting on one that is not going too, the task in
// focus. Nothing is refused — any task can be deleted.
func DeletingTasks(snap Snapshot, numbers []int) []Flag {
	var flags []Flag
	for _, n := range numbers {
		flags = append(flags, goingWith(snap, n)...)
		for _, c := range snap.Cards {
			if slices.Contains(c.Blockers, n) && !slices.Contains(numbers, c.Number) {
				flags = append(flags, Flag{Number: n, Says: "task " + strconv.Itoa(c.Number) + " waits on it"})
			}
		}
		flags = append(flags, inFocus(snap, n)...)
	}
	return flags
}

// ArchivingTasks is what an archiving of several tasks names first: a task
// already off the board cannot go again, and the parts and the focus go the
// way they go on a deletion.
func ArchivingTasks(snap Snapshot, numbers []int) []Flag {
	var flags []Flag
	for _, n := range numbers {
		if c, found := snap.Card(n); found && c.Archived {
			flags = append(flags, Flag{Number: n, Says: "it is already archived", Refused: true})
			continue
		}
		flags = append(flags, goingWith(snap, n)...)
		flags = append(flags, inFocus(snap, n)...)
	}
	return flags
}

// DeletingGoals is what a deletion of several objectives names first: one
// that work still serves cannot go, and says how much does.
func DeletingGoals(snap Snapshot, numbers []int) []Flag {
	var flags []Flag
	for _, n := range numbers {
		if serving := len(Serving(snap.Cards, n)); serving > 0 {
			flags = append(flags, Flag{Number: n, Says: counted(serving, "task serves", "tasks serve") + " it: it cannot go", Refused: true})
		}
	}
	return flags
}

// LeavingOut is what is still asked for once what was named is left out:
// everything flagged, or — when only what cannot go is left out — only that.
func LeavingOut(numbers []int, flags []Flag, onlyRefused bool) []int {
	var kept []int
	for _, n := range numbers {
		if !slices.ContainsFunc(flags, func(f Flag) bool { return f.Number == n && (f.Refused || !onlyRefused) }) {
			kept = append(kept, n)
		}
	}
	return kept
}

// Several is what several things taken together are asked of the engine as:
// the first as the number, every other beside it.
func Several(numbers []int) engine.Input {
	in := Numbered(numbers[0])
	also := []string{}
	for _, n := range numbers[1:] {
		also = append(also, strconv.Itoa(n))
	}
	in.Listed = map[string][]string{"also": also}
	return in
}

// goingWith names the parts a task takes with it.
func goingWith(snap Snapshot, n int) []Flag {
	if parts := len(Hanging(snap.Cards, n)); parts > 0 {
		return []Flag{{Number: n, Says: "its " + counted(parts, "part goes", "parts go") + " with it"}}
	}
	return nil
}

// inFocus names the task being worked on.
func inFocus(snap Snapshot, n int) []Flag {
	if slices.Contains(snap.Focus, n) {
		return []Flag{{Number: n, Says: "it is in focus"}}
	}
	return nil
}

// counted is a count and what it counts, said in the one or the many.
func counted(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", one)
	}
	return fmt.Sprintf("%d %s", n, many)
}
