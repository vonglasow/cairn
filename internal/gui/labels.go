package gui

import (
	"cairn/internal/domain"
	"slices"
	"strconv"
	"strings"

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
	seeking    bool
	sought     Written
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
	w.seeking, w.sought = false, Written{}
	if listed.Labels != nil {
		w.carried = *listed.Labels
	}
	w.labelOn = max(0, min(w.labelOn, len(w.carried)-1))
	w.labelsFrom = min(w.labelsFrom, w.labelOn)
	return w
}

func (w Window) labelChosen() string { return w.LabelsShown()[w.labelOn].Label }

// LabelsShown is the labels the list holds: every one, or those whose name holds what was typed after a slash.
func (w Window) LabelsShown() []engine.LabelCarried {
	sought := strings.ToLower(strings.TrimSpace(w.sought.Reads()))
	if sought == "" {
		return w.carried
	}
	var shown []engine.LabelCarried
	for _, one := range w.carried {
		if strings.Contains(one.Label, sought) {
			shown = append(shown, one)
		}
	}
	return shown
}

// Seeking says what is typed narrows the list of labels, and Sought is what has been typed of it.
func (w Window) Seeking() bool { return w.seeking }

// Sought is what was typed after a slash to narrow the list of labels.
func (w Window) Sought() Written { return w.sought }

// LabelCreating is the label the list offers to make, what was sought when no label holds it, and whether it offers one.
func (w Window) LabelCreating() (string, bool) {
	sought := strings.TrimSpace(w.sought.Reads())
	return sought, sought != "" && len(w.LabelsShown()) == 0
}

func (w Window) seekingLabels(change func(Written) Written) Window {
	next := change(w.sought)
	if !w.seeking {
		w.seeking = next.Reads() == "/"
		return w
	}
	if w.sought.Empty() && next.Reads() == "/" {
		return w
	}
	w.sought, w.labelOn, w.labelsFrom = next, 0, 0
	return w
}

func (w Window) leavingTheSearch() Window {
	w.seeking, w.sought, w.labelOn, w.labelsFrom = false, Written{}, 0, 0
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
	if w.seeking {
		switch name {
		case "backspace":
			w.sought, w.labelOn, w.labelsFrom = w.sought.Back(), 0, 0
			return w
		case "enter":
			w.seeking = false
			return w
		case "down", "up":
		default:
			return w
		}
	}
	switch name {
	case "/":
		w.seeking = true
	case "r":
		return w.startingRename()
	case "c":
		return w.startingColour()
	case "cmd+backspace":
		if len(w.LabelsShown()) > 0 {
			return w.Choose("delete")
		}
	case "down", "j":
		w.labelOn = min(len(w.LabelsShown())-1, w.labelOn+1)
		w.labelsFrom = max(w.labelsFrom, w.labelOn-w.LabelsRoom()+1)
	case "up", "k":
		w.labelOn = max(0, w.labelOn-1)
		w.labelsFrom = min(w.labelsFrom, w.labelOn)
	}
	return w
}

func (w Window) startingRename() Window {
	if len(w.LabelsShown()) == 0 {
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
	w.labelsFrom = max(0, min(w.labelsFrom+lines, len(w.LabelsShown())-w.LabelsRoom()))
	return w
}

// LabelsRoom is how many labels the list shows at once: the lines between its search and its offers.
func (w Window) LabelsRoom() int { return listRoom(w.LabelsBox()) - 1 }

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
	if w.LabelSearchAt().In(p) && !w.renaming {
		w.seeking = true
		return w
	}
	if _, ok := w.LabelCreating(); ok && ListRowAt(box, 1).In(p) {
		return w.creatingSought()
	}
	for _, row := range w.LabelRows() {
		if !row.At.In(p) || (row.Typing && w.making) {
			continue
		}
		for _, action := range row.Actions {
			if action.At.In(p) {
				return w.Choose(strings.ToLower(action.Offer))
			}
		}
		w.labelOn = slices.IndexFunc(w.LabelsShown(), func(one engine.LabelCarried) bool { return one.Label == row.Carried.Label })
		w.renaming, w.making = false, false
		if row.Dot.In(p) {
			return w.startingColour()
		}
		return w
	}
	if !box.In(p) {
		w.screen, w.renaming, w.making = OnTheBoard, false, false
	}
	return w
}

func (w Window) labelsOpens() Box {
	m := w.measured()
	said := []string{LabelSearchSays}
	if w.making {
		said = append(said, w.renamed.Reads())
	}
	least := 0
	for _, one := range w.carried {
		said = append(said, LabelSays(one))
		least = max(least, w.labelLineWide(m, one))
	}
	if creating, ok := w.LabelCreating(); ok {
		said = append(said, CreatingSays(creating))
	} else if len(w.carried) == 0 {
		said = append(said, NoLabelSays)
	}
	box := w.listBox(said)
	if wide := min(Over(w.width, w.height).W, least+2*padding); wide > box.W {
		box.X, box.W = (w.width-wide)/2, wide
	}
	return box
}

func (w Window) labelLineWide(m Measure, one engine.LabelCarried) int {
	actions := 0
	for _, wide := range w.actionWidths(m) {
		actions += wide + inset
	}
	return 2*inset + dotSide + inset*2 + ends(m, one.Label, small) + 2*pillPad + padding +
		ends(m, carriedSays(one), small) + padding + actions
}

func carriedSays(one engine.LabelCarried) string {
	return countOf(one.Tasks, "task") + " · " + countOf(one.Objectives, "objective")
}

// NoLabelSays is what the list of labels says when there is none.
const NoLabelSays = "No label yet: n makes one"

// LabelSearchSays is what the search of the list of labels says while nothing is typed in it.
const LabelSearchSays = "Find or name a new label"

// CreatingSays is the line that offers to make the label sought when none holds it.
func CreatingSays(sought string) string { return "Create “" + sought + "”" }

// LabelSearchAt is where the search of the list of labels is: its first line.
func (w Window) LabelSearchAt() Box { return ListRowAt(w.LabelsBox(), 0) }

// LabelActions is what the row of the label chosen offers, in its order.
var LabelActions = []string{"Rename", "Colour", "Delete"}

// RowAction is one thing the row of the label chosen offers, and where its button is.
type RowAction struct {
	Offer string
	At    Box
}

// LabelRow is one line of the list of labels where it is drawn: a label and what carries it, or the name being typed for one, whether the keyboard is on it, where its dot and its name are, and the actions it offers when it is the one chosen.
type LabelRow struct {
	At      Box
	Carried engine.LabelCarried
	Typing  bool
	On      bool
	Dot     Box
	Name    Box
	Actions []RowAction
}

// LabelRows is the lines of the list of labels in sight, under its search, the one being named first while a new label is.
func (w Window) LabelRows() []LabelRow {
	box := w.LabelsBox()
	m := w.measured()
	var rows []LabelRow
	if w.making {
		rows = append(rows, LabelRow{At: ListRowAt(box, 1), Typing: true, On: true})
	}
	shown := w.LabelsShown()
	from := w.labelsFrom
	last := min(len(shown), from+w.LabelsRoom()-len(rows))
	for i := from; i < last; i++ {
		on := i == w.labelOn && !w.making
		row := LabelRow{At: ListRowAt(box, 1+len(rows)), Carried: shown[i], On: on, Typing: on && w.renaming}
		line := Inside(row.At)
		row.Dot = Box{X: line.X, Y: row.At.Y + (row.At.H-dotSide)/2, W: dotSide, H: dotSide}
		row.Name = Box{X: row.Dot.X + dotSide + 2*inset, Y: line.Y + (line.H-pillHigh)/2, W: ends(m, shown[i].Label, small) + 2*pillPad, H: pillHigh}
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

func (w Window) actionsIn(m Measure, row Box) []RowAction {
	widths := w.actionWidths(m)
	x := row.X + row.W - inset
	actions := make([]RowAction, len(LabelActions))
	for i := len(LabelActions) - 1; i >= 0; i-- {
		x -= widths[i]
		actions[i] = RowAction{Offer: LabelActions[i], At: Box{X: x, Y: row.Y + 2, W: widths[i], H: row.H - 4}}
		x -= inset
	}
	return actions
}

// LabelColour is the colour a label wears on this board: the one the engine read for it, or else the one read off its name.
func (w Window) LabelColour(label string) string {
	if colour, ok := w.snap.Colours[label]; ok {
		return colour
	}
	return domain.ColourOf(label)
}

func (w Window) startingColour() Window {
	shown := w.LabelsShown()
	if len(shown) == 0 {
		return w
	}
	w.colouring, w.seeking = true, false
	w.colourPick = max(0, slices.Index(domain.LabelColours, shown[w.labelOn].Colour))
	return w
}

func (w Window) colouredSetting(at int) Window {
	key := w.colourKey
	w.colourKey = ""
	return w.setting(key, domain.LabelColours[at])
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
	var row Box
	switch {
	case w.colouring && w.screen == OnLabels:
		for _, one := range w.LabelRows() {
			if one.On {
				row = one.At
			}
		}
	case w.colourKey != "" && w.screen == OnConfiguration:
		for _, one := range w.SettingRows() {
			if one.Key.Name == w.colourKey {
				row = one.Control
			}
		}
	default:
		return nil
	}
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

// LabelsBox is the panel of the list of labels: across the middle of the window, its search first, then every label up to what the window has room for, wide enough for the actions of the one chosen; a search narrows the list, not the panel.
func (w Window) LabelsBox() Box { return w.placedPanel(OnLabels, w.labelsOpens()) }
