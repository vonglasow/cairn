package gui

import (
	"slices"
	"strconv"
	"strings"
)

func (w Window) autosaved() Window {
	if w.screen != OnForm || w.form.writingPart {
		w.paused = nil
		return w
	}
	now := w.form.values()
	paused := slices.Equal(now, w.paused)
	w.paused = now
	if !paused || !w.form.Changed() || !w.form.titled() || slices.Equal(now, w.refused) {
		return w
	}
	saved := w.savedInPlace()
	if saved.err != nil {
		saved.refused = now
	}
	return saved
}

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

func (f Form) titled() bool {
	for _, field := range f.fields {
		if field.Name == "title" {
			return strings.TrimSpace(field.Value) != ""
		}
	}
	return true
}

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
		w.form = editing(c)
	}
	w.chosen, w.chosenGoal = number, kept.ForObjective
	w.form = w.form.onto(kept.at)
	typed := kept.fields[kept.at].Value
	w.form = w.form.Writes(Written{Value: typed, At: kept.caret})
	w.form.opened[kept.at] = typed
	w.form.partAt, w.form.partsTop = kept.partAt, kept.partsTop
	w.paused = w.form.values()
	return w.readingActs()
}
