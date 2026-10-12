package gui

// A form saves itself. ⌘S was the only way a change reached the store, and
// every other place a person writes in saves on its own: so a form that was
// changed, and has not changed since the window last looked at the store, is
// saved where it stands and stays open over what it saved. Waiting for a beat
// with no typing keeps a save from landing in the middle of a word. A task
// being added is made once it has a title — before that there is nothing to
// make — and the form goes on over it. See docs/gui.md.

import (
	"slices"
	"strconv"
	"strings"
)

// autosaved saves the open form when it was changed and the typing has paused
// for a beat: its values are the ones the beat before saw.
func (w Window) autosaved() Window {
	if w.screen != OnForm || w.form.writingPart {
		w.paused = nil
		return w
	}
	now := w.form.values()
	paused := slices.Equal(now, w.paused)
	w.paused = now
	// What the store refused is not asked again until it changes: asked a
	// beat, the refusal never went, and the store was written to for ever.
	if !paused || !w.form.Changed() || !w.form.titled() || slices.Equal(now, w.refused) {
		return w
	}
	saved := w.savedInPlace()
	if saved.err != nil {
		saved.refused = now
	}
	return saved
}

// values is what the form holds, struck parts included, as one list a beat
// can be compared with the next by.
func (f Form) values() []string {
	var held []string
	for _, field := range f.fields {
		held = append(held, field.Value)
	}
	for _, n := range f.struck {
		held = append(held, "struck "+strconv.Itoa(n))
	}
	return held
}

// titled says the form names what it writes: a task or an objective with no
// title is nothing yet, and making it would make an empty card.
func (f Form) titled() bool {
	for _, field := range f.fields {
		if field.Name == "title" {
			return strings.TrimSpace(field.Value) != ""
		}
	}
	return true
}

// savedInPlace writes the form and opens it again over what it wrote, read
// back from the store: the keyboard where it was, the caret where it was. Read
// back rather than kept, so a part just added carries its number, and a label
// written another way by the store reads the way it was kept.
func (w Window) savedInPlace() Window {
	kept := w.form
	w, number, err := w.stored()
	if err != nil {
		return w
	}
	w = w.Look()
	if kept.ForObjective {
		g, found := w.snap.Goal(number)
		if !found {
			return w
		}
		w.form = editingGoal(g)
	} else {
		c, found := w.snap.Card(number)
		if !found {
			return w
		}
		w.form = editing(c, w.snap.Cards)
	}
	w.chosen, w.chosenGoal = number, kept.ForObjective
	w.form = w.form.onto(kept.at)
	// The field being typed in keeps what was typed: the store trims a title,
	// and a space typed before the next word went with it. It counts as saved
	// as it stands, so the next beat does not save it again.
	typed := kept.fields[kept.at].Value
	w.form = w.form.Writes(Written{Value: typed, At: kept.caret})
	w.form.opened[kept.at] = typed
	w.form.partAt, w.form.partsTop = kept.partAt, kept.partsTop
	w.paused = w.form.values()
	return w.readingActs()
}
