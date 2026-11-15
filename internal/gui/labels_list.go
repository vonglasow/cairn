package gui

import (
	"cairn/internal/engine/label"
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
	for _, c := range w.snap.Cards {
		for _, name := range c.Labels {
			known[name] = true
		}
	}
	for _, g := range w.snap.Goals {
		for _, name := range g.Labels {
			known[name] = true
		}
	}
	w.labelsCarried = map[string]int{}
	scopes := maps.Clone(w.snap.Scopes)
	if listed, err := w.work.Ask(engine.LabelList, engine.Input{}); err == nil && listed.Labels != nil {
		for _, one := range *listed.Labels {
			known[one.Label] = true
			w.labelsCarried[one.Label] = one.Tasks
			if scopes == nil {
				scopes = map[string]string{}
			}
			scopes[one.Label] = one.Scope
		}
	}
	bearer := label.OnTasks
	if w.form.ForObjective {
		bearer = label.OnObjectives
	}
	for name := range known {
		if scope, err := label.ParseScope(scopes[name]); err == nil && !scope.Admits(bearer) {
			delete(known, name)
		}
	}
	holding := engine.Words(held)
	for _, name := range holding {
		known[name] = true
	}
	w.labelsOffered = w.labelsOffered[:0:0]
	for name := range known {
		w.labelsOffered = append(w.labelsOffered, name)
	}
	slices.Sort(w.labelsOffered)
	for i, name := range w.labelsOffered {
		if slices.Contains(holding, name) {
			w.ticked[i+1] = true
		}
	}
	return w
}

func (w Window) labelChoices() []Choice {
	looking := strings.TrimSpace(w.filtering.Reads())
	written, refused := label.ParseLabels([]string{looking})
	reach := strings.ToLower(looking)
	if refused == nil && len(written) == 1 {
		reach = written[0]
	}
	var found []Choice
	for i, name := range w.labelsOffered {
		if strings.Contains(name, reach) {
			found = append(found, Choice{Number: i + 1, Title: name, Carried: w.labelsCarried[name]})
		}
	}
	slices.SortFunc(found, func(a, b Choice) int { return strings.Compare(a.Title, b.Title) })
	if looking == "" {
		return found
	}
	if refused != nil {
		return append([]Choice{{Title: refusal(refused)}}, found...)
	}
	if slices.Contains(w.labelsOffered, written[0]) {
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
	if w.choosing >= len(offered) {
		return w
	}
	row := offered[w.choosing]
	if row.Creates != "" {
		w.labelsOffered = append(slices.Clone(w.labelsOffered), row.Creates)
		w.ticked[len(w.labelsOffered)] = true
		w.filtering, w.choosing, w.choosingFrom = Written{}, 0, 0
		return w.writingPicked("+" + row.Creates)
	}
	if row.Number > 0 {
		w.ticked[row.Number] = !w.ticked[row.Number]
		if w.ticked[row.Number] {
			return w.writingPicked("+" + row.Title)
		}
		return w.writingPicked("-" + row.Title)
	}
	return w
}

// LabelTaken is one label the open list has taken, as a pill at the head of the list, and the cross that takes it off.
type LabelTaken struct {
	Label string
	Pill  Box
	Cross Box
}

// LabelsTaken is the labels the open list has taken, in their order, laid as pills at the start of the line the list is searched from.
func (w Window) LabelsTaken() []LabelTaken {
	if w.screen != OnChoosing || w.choosingFor != "label" {
		return nil
	}
	m := w.measured()
	line := Inside(ChoosingLooking(w.ChoosingBox()))
	x := line.X
	var taken []LabelTaken
	for i, name := range w.labelsOffered {
		if !w.ticked[i+1] {
			continue
		}
		pill := Box{X: x, Y: line.Y + (line.H-pillHigh)/2, W: ends(m, name, small) + 2*pillPad + crossWide, H: pillHigh}
		taken = append(taken, LabelTaken{Label: name, Pill: pill, Cross: Box{X: pill.X + pill.W - crossWide - pillPad/2, Y: pill.Y, W: crossWide + pillPad/2, H: pill.H}})
		x += pill.W + pillGap
	}
	return taken
}

const crossWide = 10

func (w Window) untaking(name string) Window {
	at := slices.Index(w.labelsOffered, name)
	if at < 0 || !w.ticked[at+1] {
		return w
	}
	w.ticked[at+1] = false
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
	if !w.pickingLabels {
		return w
	}
	in := engine.Input{Given: map[string]string{"number": strconv.Itoa(w.chosen)}, Listed: map[string][]string{"label": {change}}}
	if _, err := w.work.Ask(engine.TaskEdit, in); err != nil {
		return w.failing(err)
	}
	screen := w.screen
	w = w.Look()
	w.screen = screen
	return w
}

// Picking says the labels are picked from the board: the list hangs under the chosen card, no form behind it, and each label taken is written at once.
func (w Window) Picking() bool { return w.pickingLabels }

func (w Window) leavingThePicker() Window {
	w.screen, w.pickingLabels = OnTheBoard, false
	return w
}

func (w Window) chosenSlot() (Box, bool) {
	for _, placed := range w.Placed() {
		if placed.Card.Number == w.chosen {
			return placed.Slot, true
		}
	}
	return Box{}, false
}

func (w Window) keepingLabels() Window {
	var kept []string
	for i, name := range w.labelsOffered {
		if w.ticked[i+1] {
			kept = append(kept, name)
		}
	}
	slices.Sort(kept)
	w.form = w.form.Writes(Ending(strings.Join(kept, " ")))
	w.screen = OnForm
	return w
}

// EmptyListSays is what the open list reads when it offers nothing.
func (w Window) EmptyListSays() string {
	if w.choosingFor == "label" {
		return LabelsNoneSays
	}
	return ChoosingSays
}
