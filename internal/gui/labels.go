package gui

import (
	"cairn/internal/domain"
	"slices"
	"strconv"

	"cairn/internal/engine"
)

type labelsState struct {
	carried    []engine.LabelCarried
	labelOn    int
	labelsFrom int
	renaming   bool
	making     bool
	colouring  bool
	colourPick int
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
	w.carried, w.screen, w.renaming, w.making, w.colouring = []engine.LabelCarried{}, OnLabels, false, false, false
	if listed.Labels != nil {
		w.carried = *listed.Labels
	}
	w.labelOn = max(0, min(w.labelOn, len(w.carried)-1))
	w.labelsFrom = min(w.labelsFrom, w.labelOn)
	return w
}

func (w Window) labelChosen() string { return w.carried[w.labelOn].Label }

func (w Window) onLabels(name string) Window {
	if w.colouring {
		switch name {
		case "right", "l":
			w.colourPick = min(len(domain.LabelColours)-1, w.colourPick+1)
		case "left", "h":
			w.colourPick = max(0, w.colourPick-1)
		}
		return w
	}
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
	if len(w.carried) == 0 {
		return w
	}
	w.renaming, w.making, w.renamed = true, false, Written{}.Types(w.labelChosen())
	return w
}

func (w Window) makingLabel() Window {
	w.renaming, w.making, w.renamed = true, true, Written{}
	return w
}

func (w Window) renamingLabel() Window {
	if w.making {
		if _, err := w.work.Ask(engine.LabelNew, engine.Input{Given: map[string]string{"label": w.renamed.Reads()}}); err != nil {
			return w.failing(err)
		}
		return w.Look().listingLabels()
	}
	in := engine.Input{Given: map[string]string{"label": w.labelChosen(), "to": w.renamed.Reads()}}
	if _, err := w.work.Ask(engine.LabelRename, in); err != nil {
		return w.failing(err)
	}
	return w.Look().listingLabels()
}

// Making says the name being typed is a new label's rather than another name for one there is.
func (w Window) Making() bool { return w.making }

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
	if w.colouring {
		for i, swatch := range w.ColourSwatches() {
			if swatch.At.In(p) {
				return w.colouringLabel(i)
			}
		}
		w.colouring = false
		return w
	}
	box := w.LabelsBox()
	for _, row := range w.LabelRows() {
		if row.At.In(p) && (!row.Typing || !w.making) {
			w.labelOn = slices.IndexFunc(w.carried, func(one engine.LabelCarried) bool { return one.Label == row.Carried.Label })
			w.renaming, w.making = false, false
			return w
		}
	}
	if !box.In(p) {
		w.screen, w.renaming, w.making = OnTheBoard, false, false
	}
	return w
}

// LabelsBox is the panel of the list of labels: across the middle of the window, as long as the labels up to what the window has room for, a line more while a new one is named.
func (w Window) LabelsBox() Box {
	said := make([]string, 0, len(w.carried)+1)
	if w.making {
		said = append(said, w.renamed.Reads())
	}
	for _, one := range w.carried {
		said = append(said, LabelSays(one))
	}
	if len(said) == 0 {
		said = append(said, NoLabelSays)
	}
	return w.listBox(said)
}

// NoLabelSays is what the list of labels says when there is none.
const NoLabelSays = "No label yet: n makes one"

// LabelRow is one line of the list of labels where it is drawn: a label and what carries it, or the name being typed for one, and whether the keyboard is on it.
type LabelRow struct {
	At      Box
	Carried engine.LabelCarried
	Typing  bool
	On      bool
}

// LabelRows is the lines of the list of labels in sight, the one being named first while a new label is.
func (w Window) LabelRows() []LabelRow {
	box := w.LabelsBox()
	var rows []LabelRow
	if w.making {
		rows = append(rows, LabelRow{At: ListRowAt(box, 0), Typing: true, On: true})
	}
	from := w.labelsFrom
	last := min(len(w.carried), from+w.LabelsRoom()-len(rows))
	for i := from; i < last; i++ {
		on := i == w.labelOn && !w.making
		rows = append(rows, LabelRow{At: ListRowAt(box, len(rows)), Carried: w.carried[i], On: on, Typing: on && w.renaming})
	}
	return rows
}

// LabelColour is the colour a label wears on this board: the one the engine read for it, or else the one read off its name.
func (w Window) LabelColour(label string) string {
	if colour, ok := w.snap.Colours[label]; ok {
		return colour
	}
	return domain.ColourOf(label)
}

func (w Window) startingColour() Window {
	if len(w.carried) == 0 {
		return w
	}
	w.colouring = true
	w.colourPick = max(0, slices.Index(domain.LabelColours, w.carried[w.labelOn].Colour))
	return w
}

func (w Window) colouringLabel(at int) Window {
	in := engine.Input{Given: map[string]string{"label": w.labelChosen(), "colour": domain.LabelColours[at]}}
	if _, err := w.work.Ask(engine.LabelColour, in); err != nil {
		return w.failing(err)
	}
	return w.Look().listingLabels()
}

// ColourSwatch is one colour a label can wear, where the list of labels draws it to be picked.
type ColourSwatch struct {
	Colour string
	At     Box
}

// ColourSwatches is the fourteen colours laid out under the label being coloured, a swatch each, and nothing while no label is.
func (w Window) ColourSwatches() []ColourSwatch {
	if !w.colouring || w.screen != OnLabels {
		return nil
	}
	row := ListRowAt(w.LabelsBox(), w.labelOn-w.labelsFrom)
	wide := len(domain.LabelColours)*(swatchSide+swatchGap) - swatchGap
	x := max(margin+inset, min(row.X+inset, w.width-margin-inset-wide))
	y := row.Y + row.H + inset
	swatches := make([]ColourSwatch, 0, len(domain.LabelColours))
	for i, colour := range domain.LabelColours {
		swatches = append(swatches, ColourSwatch{Colour: colour, At: Box{X: x + i*(swatchSide+swatchGap), Y: y + inset, W: swatchSide, H: swatchSide}})
	}
	return swatches
}

// ColourPick is the swatch the keyboard is on.
func (w Window) ColourPick() int { return w.colourPick }

const (
	swatchSide = 18
	swatchGap  = 6
)
