package gui

import (
	"cairn/internal/board"
	"slices"
	"strconv"

	"gioui.org/io/key"

	"cairn/internal/engine"
)

// KeysDown is told which keys are held while the pointer acts: ⌘ gathers what is clicked, ⇧ stretches to it.
func (w Window) KeysDown(held key.Modifiers) Window {
	w.gathering, w.stretching = held.Contain(key.ModShortcut), held.Contain(key.ModShift)
	return w
}

func (w Window) choosingAt(number int) Window {
	if number != 0 && w.gathering {
		return w.gather(number)
	}
	if number != 0 && w.stretching {
		return w.stretch(number)
	}
	w.Gathered = nil
	w.Chosen, w.ChosenGoal = number, w.Tab == board.Objectives
	if w.ChosenGoal {
		return w.openingParentOf(number)
	}
	return w
}

func (w Window) gather(number int) Window {
	if len(w.Gathered) == 0 && w.Chosen != 0 {
		w.Gathered = []int{w.Chosen}
	}
	if at := slices.Index(w.Gathered, number); at >= 0 {
		w.Gathered = slices.Delete(slices.Clone(w.Gathered), at, at+1)
		return w
	}
	w.Gathered = append(slices.Clone(w.Gathered), number)
	w.Chosen, w.ChosenGoal = number, w.Tab == board.Objectives
	return w
}

func (w Window) stretch(number int) Window {
	for _, run := range w.Runs() {
		from, to := slices.Index(run, w.Chosen), slices.Index(run, number)
		if from < 0 || to < 0 {
			continue
		}
		w.Gathered = slices.Clone(run[min(from, to) : max(from, to)+1])
		w.Chosen, w.ChosenGoal = number, w.Tab == board.Objectives
		return w
	}
	return w.gather(number)
}

func (w Window) askingSeveral(about board.Question) Window {
	w.Screen, w.Question = board.OnAsking, about
	return w
}

func (w Window) severalSays() []Said {
	flags := w.Flags()
	taken := w.Taken()
	var said []Said
	for i, n := range taken {
		if i == severalNamed && len(taken) > severalNamed+1 {
			said = append(said, Said{Means: "and " + strconv.Itoa(len(taken)-i) + " more"})
			break
		}
		line := "#" + strconv.Itoa(n) + "  " + w.TitleOf(n)
		for _, f := range flags {
			if f.Number == n {
				line += " — " + f.Says
			}
		}
		said = append(said, Said{Means: line})
	}
	return said
}

const severalNamed = 6

func (w Window) actingOnSeveral(onlyRefused bool) Window {
	kept := board.LeavingOut(w.Taken(), w.Flags(), onlyRefused)
	op := engine.TaskDelete
	if w.ChosenGoal {
		op = engine.ObjectiveDelete
	}
	if w.Question == board.AboutArchivingSeveral {
		op = engine.TaskArchive
	}
	w.Screen, w.Question = board.OnTheBoard, board.AboutDeleting
	if len(kept) == 0 {
		return w
	}
	w.Gathered = nil
	return w.asking(op, engine.Agreed(board.Several(kept)))
}
