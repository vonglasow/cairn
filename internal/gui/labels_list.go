package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/label"
	"cairn/internal/gui/geom"
	"errors"
	"maps"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/engine"
)

// LabelsNoneSays is what an empty list of labels reads: nobody carries one yet, and what is typed creates the first.
const LabelsNoneSays = "no label yet: type one"

func (w Window) openingLabels(held string) Window {
	known := map[string]bool{}
	for _, c := range w.Snap.Cards {
		for _, name := range c.Labels {
			known[name] = true
		}
	}
	for _, g := range w.Snap.Goals {
		for _, name := range g.Labels {
			known[name] = true
		}
	}
	w.LabelsCounted = map[string]int{}
	scopes := maps.Clone(w.Snap.Scopes)
	if listed, err := w.work.Ask(engine.LabelList, engine.Input{}); err == nil && listed.Labels != nil {
		for _, one := range *listed.Labels {
			known[one.Label] = true
			w.LabelsCounted[one.Label] = one.Tasks
			if scopes == nil {
				scopes = map[string]string{}
			}
			scopes[one.Label] = one.Scope
		}
	}
	bearer := label.OnTasks
	if w.Form.ForObjective {
		bearer = label.OnObjectives
	}
	for name := range known {
		if scope, err := label.ParseScope(scopes[name]); err == nil && !scope.Admits(bearer) {
			delete(known, name)
		}
	}
	holding := board.Words(held)
	for _, name := range holding {
		known[name] = true
	}
	w.LabelsOffered = w.LabelsOffered[:0:0]
	for name := range known {
		w.LabelsOffered = append(w.LabelsOffered, name)
	}
	slices.Sort(w.LabelsOffered)
	for i, name := range w.LabelsOffered {
		if slices.Contains(holding, name) {
			w.TickedSet[i+1] = true
		}
	}
	return w
}

func (w Window) labelChoices() []Choice {
	looking := strings.TrimSpace(w.Filtering.Reads())
	written, refused := label.ParseLabels([]string{looking})
	reach := strings.ToLower(looking)
	if refused == nil && len(written) == 1 {
		reach = written[0]
	}
	var found []Choice
	for i, name := range w.LabelsOffered {
		if strings.Contains(name, reach) {
			found = append(found, Choice{Number: i + 1, Title: name, Carried: w.LabelsCounted[name]})
		}
	}
	slices.SortFunc(found, func(a, b Choice) int { return strings.Compare(a.Title, b.Title) })
	if looking == "" {
		return found
	}
	if refused != nil {
		return append([]Choice{{Title: refusal(refused)}}, found...)
	}
	if slices.Contains(w.LabelsOffered, written[0]) {
		return found
	}
	return append(found, Choice{Title: "create " + written[0], Creates: written[0]})
}

func refusal(err error) string {
	for _, why := range []error{label.ErrLabelHasNoFamily, label.ErrLabelHolds, label.ErrLabelNotOneLine} {
		if errors.Is(err, why) {
			return why.Error()
		}
	}
	return err.Error()
}

func (w Window) takingLabel() Window {
	offered := w.Choosing()
	if w.ChoosingOn >= len(offered) {
		return w
	}
	row := offered[w.ChoosingOn]
	if row.Creates != "" {
		w.LabelsOffered = append(slices.Clone(w.LabelsOffered), row.Creates)
		w.TickedSet[len(w.LabelsOffered)] = true
		w.Filtering, w.ChoosingOn, w.ChoosingFrom = board.Written{}, 0, 0
		return w.writingPicked("+" + row.Creates)
	}
	if row.Number > 0 {
		w.TickedSet[row.Number] = !w.TickedSet[row.Number]
		if w.TickedSet[row.Number] {
			return w.writingPicked("+" + row.Title)
		}
		return w.writingPicked("-" + row.Title)
	}
	return w
}

// LabelTaken is one label the open list has taken, as a pill at the head of the list, and the cross that takes it off.
type LabelTaken struct {
	Label string
	Pill  geom.Box
	Cross geom.Box
}

// LabelsTaken is the labels the open list has taken, in their order, laid as pills at the start of the line the list is searched from.
func (w Window) LabelsTaken() []LabelTaken {
	if w.Screen != board.OnChoosing || w.ChoosingFor != "label" {
		return nil
	}
	m := w.measured()
	line := Inside(ChoosingLooking(w.ChoosingBox()))
	x := line.X
	var taken []LabelTaken
	for i, name := range w.LabelsOffered {
		if !w.TickedSet[i+1] {
			continue
		}
		pill := geom.Box{X: x, Y: line.Y + (line.H-pillHigh)/2, W: ends(m, name, small) + 2*pillPad + crossWide, H: pillHigh}
		taken = append(taken, LabelTaken{Label: name, Pill: pill, Cross: geom.Box{X: pill.X + pill.W - crossWide - pillPad/2, Y: pill.Y, W: crossWide + pillPad/2, H: pill.H}})
		x += pill.W + pillGap
	}
	return taken
}

const crossWide = 10

func (w Window) untaking(name string) Window {
	at := slices.Index(w.LabelsOffered, name)
	if at < 0 || !w.TickedSet[at+1] {
		return w
	}
	w.TickedSet[at+1] = false
	return w.writingPicked("-" + name)
}

func (w Window) untakingLast() Window {
	taken := w.LabelsTaken()
	if len(taken) == 0 {
		return w
	}
	return w.untaking(taken[len(taken)-1].Label)
}

func (w Window) writingPicked(change string) Window {
	if !w.PickingLabels {
		return w
	}
	in := engine.Input{Given: map[string]string{"number": strconv.Itoa(w.Chosen)}, Listed: map[string][]string{"label": {change}}}
	if _, err := w.work.Ask(engine.TaskEdit, in); err != nil {
		return w.failing(err)
	}
	screen := w.Screen
	w = w.Look()
	w.Screen = screen
	return w
}

func (w Window) leavingThePicker() Window {
	w.Screen, w.PickingLabels = board.OnTheBoard, false
	return w
}

func (w Window) chosenSlot() (geom.Box, bool) {
	for _, placed := range w.Placed() {
		if placed.Card.Number == w.Chosen {
			return placed.Slot, true
		}
	}
	return geom.Box{}, false
}

func (w Window) keepingLabels() Window {
	var kept []string
	for i, name := range w.LabelsOffered {
		if w.TickedSet[i+1] {
			kept = append(kept, name)
		}
	}
	slices.Sort(kept)
	w.Form = w.Form.Writes(board.Ending(strings.Join(kept, " ")))
	w.Screen = board.OnForm
	return w
}

// EmptyListSays is what the open list reads when it offers nothing.
func (w Window) EmptyListSays() string {
	if w.ChoosingFor == "label" {
		return LabelsNoneSays
	}
	if w.ChoosingSearched() {
		return board.TasksNoneSays
	}
	return ChoosingSays
}
