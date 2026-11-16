package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/label"
	"cairn/internal/gui/geom"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/engine"
)

// LabelSays is one label's line: its name, then how many tasks and objectives carry it.
func LabelSays(one label.Carried) string {
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
	w.Carried, w.Screen, w.Renaming, w.Making, w.Colouring = []label.Carried{}, board.OnLabels, false, false, false
	w.Seeking, w.Sought = false, board.Written{}
	if listed.Labels != nil {
		w.Carried = *listed.Labels
	}
	w.LabelOn = max(0, min(w.LabelOn, len(w.Carried)-1))
	w.LabelsFrom = min(w.LabelsFrom, w.LabelOn)
	return w
}

func (w Window) seekingLabels(change func(board.Written) board.Written) Window {
	next := change(w.Sought)
	if !w.Seeking {
		w.Seeking = next.Reads() == "/"
		return w
	}
	if w.Sought.Empty() && next.Reads() == "/" {
		return w
	}
	w.Sought, w.LabelOn, w.LabelsFrom = next, 0, 0
	return w
}

func (w Window) leavingTheSearch() Window {
	w.Seeking, w.Sought, w.LabelOn, w.LabelsFrom = false, board.Written{}, 0, 0
	return w
}

func (w Window) creatingSought() Window {
	sought, ok := w.LabelCreating()
	if !ok {
		return w
	}
	if _, err := w.work.Ask(engine.LabelNew, engine.Input{Given: map[string]string{"label": sought}}); err != nil {
		return w.failing(err)
	}
	return w.Look().listingLabels()
}

func (w Window) onLabels(name string) Window {
	if w.Colouring {
		switch name {
		case "right", "l":
			w.ColourPick = min(len(label.LabelColours)-1, w.ColourPick+1)
		case "left", "h":
			w.ColourPick = max(0, w.ColourPick-1)
		}
		return w
	}
	if w.Renaming {
		if name == "backspace" {
			w.Renamed = w.Renamed.Back()
		}
		return w
	}
	if w.Seeking {
		switch name {
		case "backspace":
			w.Sought, w.LabelOn, w.LabelsFrom = w.Sought.Back(), 0, 0
			return w
		case "enter":
			w.Seeking = false
			return w
		case "down", "up":
		default:
			return w
		}
	}
	switch name {
	case "/":
		w.Seeking = true
	case "r":
		return w.startingRename()
	case "c":
		return w.startingColour()
	case "cmd+backspace":
		if len(w.LabelsShown()) > 0 {
			return w.Choose("delete")
		}
	case "down", "j":
		w.LabelOn = min(len(w.LabelsShown())-1, w.LabelOn+1)
		w.LabelsFrom = max(w.LabelsFrom, w.LabelOn-w.LabelsRoom()+1)
	case "up", "k":
		w.LabelOn = max(0, w.LabelOn-1)
		w.LabelsFrom = min(w.LabelsFrom, w.LabelOn)
	}
	return w
}

func (w Window) startingRename() Window {
	if len(w.LabelsShown()) == 0 {
		return w
	}
	w.Renaming, w.Making, w.Renamed = true, false, board.Written{}.Types(w.LabelChosen())
	return w
}

func (w Window) makingLabel() Window {
	w.Renaming, w.Making, w.Renamed = true, true, board.Written{}
	return w
}

func (w Window) renamingLabel() Window {
	if w.Making {
		if _, err := w.work.Ask(engine.LabelNew, engine.Input{Given: map[string]string{"label": w.Renamed.Reads()}}); err != nil {
			return w.failing(err)
		}
		return w.Look().listingLabels()
	}
	in := engine.Input{Given: map[string]string{"label": w.LabelChosen(), "to": w.Renamed.Reads()}}
	if _, err := w.work.Ask(engine.LabelRename, in); err != nil {
		return w.failing(err)
	}
	return w.Look().listingLabels()
}

func (w Window) deletingLabel() Window {
	name := w.LabelChosen()
	w.Screen, w.Question = board.OnLabels, board.AboutDeleting
	if _, err := w.work.Ask(engine.LabelDelete, engine.Agreed(engine.Input{Given: map[string]string{"label": name}})); err != nil {
		return w.failing(err)
	}
	w = w.Look().listingLabels()
	if w.Screen != board.OnLabels {
		w.Screen = board.OnTheBoard
	}
	return w.telling("the label " + name + " was taken off everything")
}

func (w Window) scrollingLabels(p geom.Point, by int) Window {
	if !w.LabelsBox().In(p) {
		return w
	}
	lines, left := turning(w.spareChoosing, by)
	w.spareChoosing = left
	w.LabelsFrom = max(0, min(w.LabelsFrom+lines, len(w.LabelsShown())-w.LabelsRoom()))
	return w
}

// LabelsRoom is how many labels the list shows at once: the lines between its search and its offers.
func (w Window) LabelsRoom() int { return listRoom(w.LabelsBox()) - 1 }

func (w Window) clickingLabels(p geom.Point) Window {
	if w.Colouring {
		for i, swatch := range w.ColourSwatches() {
			if swatch.At.In(p) {
				return w.colouringLabel(i)
			}
		}
		w.Colouring = false
		return w
	}
	box := w.LabelsBox()
	if w.LabelSearchAt().In(p) && !w.Renaming {
		w.Seeking = true
		return w
	}
	if _, ok := w.LabelCreating(); ok && ListRowAt(box, 1).In(p) {
		return w.creatingSought()
	}
	for _, row := range w.LabelRows() {
		if !row.At.In(p) || (row.Typing && w.Making) {
			continue
		}
		for _, action := range row.Actions {
			if action.At.In(p) {
				return w.Choose(strings.ToLower(action.Offer))
			}
		}
		w.LabelOn = slices.IndexFunc(w.LabelsShown(), func(one label.Carried) bool { return one.Label == row.Carried.Label })
		w.Renaming, w.Making = false, false
		if row.Dot.In(p) {
			return w.startingColour()
		}
		return w
	}
	if !box.In(p) {
		w.Screen, w.Renaming, w.Making = board.OnTheBoard, false, false
	}
	return w
}

func (w Window) labelsOpens() geom.Box {
	m := w.measured()
	said := []string{LabelSearchSays}
	if w.Making {
		said = append(said, w.Renamed.Reads())
	}
	least := 0
	for _, one := range w.Carried {
		said = append(said, LabelSays(one))
		least = max(least, w.labelLineWide(m, one))
	}
	if creating, ok := w.LabelCreating(); ok {
		said = append(said, CreatingSays(creating))
	} else if len(w.Carried) == 0 {
		said = append(said, NoLabelSays)
	}
	box := w.listBox(said)
	if wide := min(Over(w.width, w.height).W, least+2*padding); wide > box.W {
		box.X, box.W = (w.width-wide)/2, wide
	}
	return box
}

func (w Window) labelLineWide(m Measure, one label.Carried) int {
	actions := 0
	for _, wide := range w.actionWidths(m) {
		actions += wide + inset
	}
	return 2*inset + dotSide + inset*2 + ends(m, one.Label, small) + 2*pillPad + padding +
		ends(m, carriedSays(one), small) + padding + actions
}

func carriedSays(one label.Carried) string {
	return countOf(one.Tasks, "task") + " · " + countOf(one.Objectives, "objective")
}

// NoLabelSays is what the list of labels says when there is none.
const NoLabelSays = "No label yet: `n` makes one"

// LabelSearchSays is what the search of the list of labels says while nothing is typed in it.
const LabelSearchSays = "Find or name a new label"

// CreatingSays is the line that offers to make the label sought when none holds it.
func CreatingSays(sought string) string { return "Create “" + sought + "”" }

// LabelSearchAt is where the search of the list of labels is: its first line.
func (w Window) LabelSearchAt() geom.Box { return ListRowAt(w.LabelsBox(), 0) }

// LabelActions is what the row of the label chosen offers, in its order.
var LabelActions = []string{"Rename", "Colour", "Delete"}

// RowAction is one thing the row of the label chosen offers, and where its button is.
type RowAction struct {
	Offer string
	At    geom.Box
}

// LabelRow is one line of the list of labels where it is drawn: a label and what carries it, or the name being typed for one, whether the keyboard is on it, where its dot and its name are, and the actions it offers when it is the one chosen.
type LabelRow struct {
	At      geom.Box
	Carried label.Carried
	Typing  bool
	On      bool
	Dot     geom.Box
	Name    geom.Box
	Actions []RowAction
}

// LabelRows is the lines of the list of labels in sight, under its search, the one being named first while a new label is.
func (w Window) LabelRows() []LabelRow {
	box := w.LabelsBox()
	m := w.measured()
	var rows []LabelRow
	if w.Making {
		rows = append(rows, LabelRow{At: ListRowAt(box, 1), Typing: true, On: true})
	}
	shown := w.LabelsShown()
	from := w.LabelsFrom
	last := min(len(shown), from+w.LabelsRoom()-len(rows))
	for i := from; i < last; i++ {
		on := i == w.LabelOn && !w.Making
		row := LabelRow{At: ListRowAt(box, 1+len(rows)), Carried: shown[i], On: on, Typing: on && w.Renaming}
		line := Inside(row.At)
		row.Dot = geom.Box{X: line.X, Y: row.At.Y + (row.At.H-dotSide)/2, W: dotSide, H: dotSide}
		row.Name = geom.Box{X: row.Dot.X + dotSide + 2*inset, Y: line.Y + (line.H-pillHigh)/2, W: ends(m, shown[i].Label, small) + 2*pillPad, H: pillHigh}
		if on && !row.Typing {
			row.Actions = w.actionsIn(m, row.At)
		}
		rows = append(rows, row)
	}
	return rows
}

func (w Window) actionWidths(m Measure) []int {
	widths := make([]int, 0, len(LabelActions))
	for _, offer := range LabelActions {
		widths = append(widths, buttonWordWide(m, offer)+KeyChipWide(w.OfferKey(offer))+hintSpace+padding)
	}
	return widths
}

func (w Window) actionsIn(m Measure, row geom.Box) []RowAction {
	widths := w.actionWidths(m)
	x := row.X + row.W - inset
	actions := make([]RowAction, len(LabelActions))
	for i := len(LabelActions) - 1; i >= 0; i-- {
		x -= widths[i]
		actions[i] = RowAction{Offer: LabelActions[i], At: geom.Box{X: x, Y: row.Y + 2, W: widths[i], H: row.H - 4}}
		x -= inset
	}
	return actions
}

func (w Window) startingColour() Window {
	shown := w.LabelsShown()
	if len(shown) == 0 {
		return w
	}
	w.Colouring, w.Seeking = true, false
	w.ColourPick = max(0, slices.Index(label.LabelColours, shown[w.LabelOn].Colour))
	return w
}

func (w Window) colouredSetting(at int) Window {
	key := w.colourKey
	w.colourKey = ""
	return w.setting(key, label.LabelColours[at])
}

func (w Window) colouringLabel(at int) Window {
	in := engine.Input{Given: map[string]string{"label": w.LabelChosen(), "colour": label.LabelColours[at]}}
	if _, err := w.work.Ask(engine.LabelColour, in); err != nil {
		return w.failing(err)
	}
	return w.Look().listingLabels()
}

// ColourSwatch is one colour a label can wear, where the list of labels draws it to be picked.
type ColourSwatch struct {
	Colour string
	At     geom.Box
}

// ColourSwatches is the fourteen colours laid out under the label being coloured, a swatch each, and nothing while no label is.
func (w Window) ColourSwatches() []ColourSwatch {
	var row geom.Box
	switch {
	case w.Colouring && w.Screen == board.OnLabels:
		for _, one := range w.LabelRows() {
			if one.On {
				row = one.At
			}
		}
	default:
		return nil
	}
	wide := len(label.LabelColours)*(swatchSide+swatchGap) - swatchGap
	x := max(geom.Margin+inset, min(row.X+inset, w.width-geom.Margin-inset-wide))
	y := row.Y + row.H + inset
	swatches := make([]ColourSwatch, 0, len(label.LabelColours))
	for i, colour := range label.LabelColours {
		swatches = append(swatches, ColourSwatch{Colour: colour, At: geom.Box{X: x + i*(swatchSide+swatchGap), Y: y + inset, W: swatchSide, H: swatchSide}})
	}
	return swatches
}

const (
	swatchSide = 18
	swatchGap  = 6
)

// LabelsBox is the panel of the list of labels: across the middle of the window, its search first, then every label up to what the window has room for, wide enough for the actions of the one chosen; a search narrows the list, not the panel.
func (w Window) LabelsBox() geom.Box { return w.placedPanel(board.OnLabels, w.labelsOpens()) }
