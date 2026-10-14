package gui

import (
	"strconv"

	"cairn/internal/engine"
)

type labelsState struct {
	carried    []engine.LabelCarried
	labelOn    int
	labelsFrom int
	renaming   bool
	renamed    Written
}

// LabelsCarried is every label, as the engine listed them, with what carries each.
func (w Window) LabelsCarried() []engine.LabelCarried { return w.carried }

// LabelOn is the line of the list the keyboard is on.
func (w Window) LabelOn() int { return w.labelOn }

// LabelsFrom is the first label of the list in sight.
func (w Window) LabelsFrom() int { return w.labelsFrom }

// Renaming says a label's new name is being typed, and Renamed is what has been typed of it.
func (w Window) Renaming() bool { return w.renaming }

// Renamed is the new name typed so far.
func (w Window) Renamed() Written { return w.renamed }

// LabelSays is one label's line: its name, then how many tasks and objectives carry it.
func LabelSays(one engine.LabelCarried) string {
	return one.Label + "   " + countOf(one.Tasks, "task") + " · " + countOf(one.Objectives, "objective")
}

func countOf(n int, one string) string {
	if n == 1 {
		return "1 " + one
	}
	return strconv.Itoa(n) + " " + one + "s"
}

func (w Window) listingLabels() Window {
	listed, err := w.work.Ask(engine.LabelList, engine.Input{})
	if err != nil {
		return w.failing(err)
	}
	if listed.Labels == nil || len(*listed.Labels) == 0 {
		return w.telling("no label yet: a task's form puts one on")
	}
	w.carried, w.screen, w.renaming = *listed.Labels, OnLabels, false
	w.labelOn = min(w.labelOn, len(w.carried)-1)
	w.labelsFrom = min(w.labelsFrom, w.labelOn)
	return w
}

func (w Window) labelChosen() string { return w.carried[w.labelOn].Label }

func (w Window) onLabels(name string) Window {
	if w.renaming {
		if name == "backspace" {
			w.renamed = w.renamed.Back()
		}
		return w
	}
	switch name {
	case "down", "j":
		w.labelOn = min(len(w.carried)-1, w.labelOn+1)
		w.labelsFrom = max(w.labelsFrom, w.labelOn-w.LabelsRoom()+1)
	case "up", "k":
		w.labelOn = max(0, w.labelOn-1)
		w.labelsFrom = min(w.labelsFrom, w.labelOn)
	}
	return w
}

func (w Window) startingRename() Window {
	w.renaming, w.renamed = true, Written{}.Types(w.labelChosen())
	return w
}

func (w Window) renamingLabel() Window {
	in := engine.Input{Given: map[string]string{"label": w.labelChosen(), "to": w.renamed.Reads()}}
	if _, err := w.work.Ask(engine.LabelRename, in); err != nil {
		return w.failing(err)
	}
	return w.Look().listingLabels()
}

func (w Window) deletingLabel() Window {
	label := w.labelChosen()
	w.screen, w.question = OnLabels, aboutDeleting
	if _, err := w.work.Ask(engine.LabelDelete, engine.Agreed(engine.Input{Given: map[string]string{"label": label}})); err != nil {
		return w.failing(err)
	}
	w = w.Look().listingLabels()
	if w.screen != OnLabels {
		w.screen = OnTheBoard
	}
	return w.telling("the label " + label + " was taken off everything")
}

func (w Window) scrollingLabels(p Point, by int) Window {
	if !w.LabelsBox().In(p) {
		return w
	}
	lines, left := turning(w.spareChoosing, by)
	w.spareChoosing = left
	w.labelsFrom = max(0, min(w.labelsFrom+lines, len(w.carried)-w.LabelsRoom()))
	return w
}

// LabelsRoom is how many labels the list shows at once: the lines between its heading and its offers.
func (w Window) LabelsRoom() int { return listRoom(w.LabelsBox()) }

func (w Window) clickingLabels(p Point) Window {
	box := w.LabelsBox()
	for i := w.labelsFrom; i < min(len(w.carried), w.labelsFrom+w.LabelsRoom()); i++ {
		if ListRowAt(box, i-w.labelsFrom).In(p) {
			w.labelOn, w.renaming = i, false
			return w
		}
	}
	if !box.In(p) {
		w.screen, w.renaming = OnTheBoard, false
	}
	return w
}

// LabelsBox is the panel of the list of labels: across the middle of the window, as long as the labels up to what the window has room for.
func (w Window) LabelsBox() Box {
	said := make([]string, 0, len(w.carried))
	for _, one := range w.carried {
		said = append(said, LabelSays(one))
	}
	return w.listBox(said)
}
