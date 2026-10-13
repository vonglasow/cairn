package gui

// Several cards taken at once: gathered with ⌘ and a click, stretched with ⇧
// and a click, and deleted or archived in one act once asked.

import (
	"slices"
	"strconv"

	"gioui.org/io/key"

	"cairn/internal/engine"
)

// KeysDown is told which keys are held while the pointer acts: ⌘ gathers
// what is clicked, ⇧ stretches to it. The loop says so before every pointer
// event, since a click arrives with its keys and the window keeps nothing of
// them otherwise.
func (w Window) KeysDown(held key.Modifiers) Window {
	w.gathering, w.stretching = held.Contain(key.ModShortcut), held.Contain(key.ModShift)
	return w
}

// Taken is what is taken together, in the order it was taken: nothing, or
// two and more.
func (w Window) Taken() []int {
	if len(w.taken) < 2 {
		return nil
	}
	return w.taken
}

// isTaken says whether a card is among what is taken together.
func (w Window) isTaken(number int) bool { return slices.Contains(w.Taken(), number) }

// choosingAt is a click on a card of the board or an objective: chosen alone,
// or — with a key held — gathered with what was chosen, or every card between
// the two.
func (w Window) choosingAt(number int) Window {
	if number != 0 && w.gathering {
		return w.gather(number)
	}
	if number != 0 && w.stretching {
		return w.stretch(number)
	}
	w.taken = nil
	w.chosen, w.chosenGoal = number, w.tab == Objectives
	return w
}

// gather adds a card to what is taken, or takes it back out.
func (w Window) gather(number int) Window {
	if len(w.taken) == 0 && w.chosen != 0 {
		w.taken = []int{w.chosen}
	}
	if at := slices.Index(w.taken, number); at >= 0 {
		w.taken = slices.Delete(slices.Clone(w.taken), at, at+1)
		return w
	}
	w.taken = append(slices.Clone(w.taken), number)
	w.chosen, w.chosenGoal = number, w.tab == Objectives
	return w
}

// stretch takes every card from the one chosen to the one clicked, in the
// order they are drawn — in one column, one lane, or the objectives. Across
// two of them there is no between, and the click gathers instead.
func (w Window) stretch(number int) Window {
	for _, run := range w.runs() {
		from, to := slices.Index(run, w.chosen), slices.Index(run, number)
		if from < 0 || to < 0 {
			continue
		}
		w.taken = slices.Clone(run[min(from, to) : max(from, to)+1])
		w.chosen, w.chosenGoal = number, w.tab == Objectives
		return w
	}
	return w.gather(number)
}

// runs is every list of cards a stretch can follow: the objectives on their
// page, the columns and the lanes on the board.
func (w Window) runs() [][]int {
	numbers := func(cards []engine.Card) []int {
		var out []int
		for _, c := range cards {
			out = append(out, c.Number)
		}
		return out
	}
	if w.tab == Objectives {
		var goals []int
		for _, g := range engine.Sorted(w.Showing().Goals) {
			goals = append(goals, g.Number)
		}
		return [][]int{goals}
	}
	var runs [][]int
	for i := range engine.Columns {
		runs = append(runs, numbers(w.inColumn(i)))
	}
	for i := range engine.Stalled {
		runs = append(runs, numbers(w.inLane(i)))
	}
	return runs
}

// asksAboutSeveral says whether the question is about what is taken together.
func (w Window) asksAboutSeveral() bool {
	return w.question == aboutDeletingSeveral || w.question == aboutArchivingSeveral
}

// askingSeveral asks about deleting, or archiving, what is taken.
func (w Window) askingSeveral(about Question) Window {
	w.screen, w.question = OnAsking, about
	return w
}

// flags is what the question about what is taken names first.
func (w Window) flags() []engine.Flag {
	if w.question == aboutArchivingSeveral {
		return engine.ArchivingTasks(w.snap, w.Taken())
	}
	if w.chosenGoal {
		return engine.DeletingGoals(w.snap, w.Taken())
	}
	return engine.DeletingTasks(w.snap, w.Taken())
}

// severalSays is what the question about what is taken says: a line for each
// thing it names first, or that nothing is named.
func (w Window) severalSays() []Said {
	flags := w.flags()
	if len(flags) == 0 {
		return []Said{{Means: "Nothing here takes anything else with it."}}
	}
	var said []Said
	for _, f := range flags {
		said = append(said, Said{Key: w.kind() + " " + strconv.Itoa(f.Number), Means: f.Says})
	}
	return said
}

// severalOffers is what the question about what is taken offers: all that
// can go, all but what it named, or nothing. The second only when it named
// something that could go all the same — otherwise it is the first again.
func (w Window) severalOffers() []string {
	if slices.ContainsFunc(w.flags(), func(f engine.Flag) bool { return !f.Refused }) {
		return []string{"All possible", "Skip flagged", "Cancel"}
	}
	return []string{"All possible", "Cancel"}
}

// actingOnSeveral does what was asked to what is taken, leaving out what was
// named — every one of it, or only what cannot go — in one act.
func (w Window) actingOnSeveral(onlyRefused bool) Window {
	kept := engine.LeavingOut(w.Taken(), w.flags(), onlyRefused)
	op := engine.TaskDelete
	if w.chosenGoal {
		op = engine.ObjectiveDelete
	}
	if w.question == aboutArchivingSeveral {
		op = engine.TaskArchive
	}
	w.screen, w.question = OnTheBoard, aboutDeleting
	if len(kept) == 0 {
		return w
	}
	w.taken = nil
	return w.asking(op, engine.Agreed(engine.Several(kept)))
}

// severalHead is how the question about what is taken is headed.
func (w Window) severalHead() string {
	verb := "Delete "
	if w.question == aboutArchivingSeveral {
		verb = "Archive "
	}
	return verb + strconv.Itoa(len(w.Taken())) + " " + w.kind() + "s?"
}
