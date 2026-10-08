package gui

import (
	"errors"
	"slices"
	"strings"

	"cairn/internal/domain"
	"cairn/internal/reading"
)

// The list labels are chosen from: the one objectives are chosen from, with
// names where it had numbers, and one row more — the one that creates a label
// nobody carries yet, under the name the engine will keep it by
// (docs/labels.md). The name shown is the rule's own, read from the domain,
// so the row never promises a spelling the save would not keep.

// LabelsNoneSays is what an empty list of labels reads: nobody carries one
// yet, and what is typed creates the first.
const LabelsNoneSays = "no label yet: type one"

// openingLabels offers every label some task or objective carries, and those
// the field already holds — typed or created before the list was opened —
// with the field's own ticked.
func (w Window) openingLabels(held string) Window {
	known := map[string]bool{}
	for _, c := range w.snap.Cards {
		for _, label := range c.Labels {
			known[label] = true
		}
	}
	for _, g := range w.snap.Goals {
		for _, label := range g.Labels {
			known[label] = true
		}
	}
	holding := reading.Words(held)
	for _, label := range holding {
		known[label] = true
	}
	w.labelsOffered = w.labelsOffered[:0:0]
	for label := range known {
		w.labelsOffered = append(w.labelsOffered, label)
	}
	slices.Sort(w.labelsOffered)
	for i, label := range w.labelsOffered {
		if slices.Contains(holding, label) {
			w.ticked[i+1] = true
		}
	}
	return w
}

// labelChoices is the list as it reads: the labels the typing reaches, in
// alphabetical order, after the row that creates what was typed when that is
// no label yet — or says why it cannot be one.
func (w Window) labelChoices() []Choice {
	looking := strings.TrimSpace(w.filtering.Reads())
	written, refused := domain.ParseLabels([]string{looking})
	reach := strings.ToLower(looking)
	if refused == nil && len(written) == 1 {
		reach = written[0]
	}
	var found []Choice
	for i, label := range w.labelsOffered {
		if strings.Contains(label, reach) {
			found = append(found, Choice{Number: i + 1, Title: label})
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
	return append([]Choice{{Title: "create " + written[0], Creates: written[0]}}, found...)
}

// refusal is why what was typed cannot be a label, without the typing it
// quotes: the line above the list already shows that.
func refusal(err error) string {
	for _, why := range []error{domain.ErrLabelHasNoFamily, domain.ErrLabelHolds, domain.ErrLabelNotOneLine} {
		if errors.Is(err, why) {
			return why.Error()
		}
	}
	return err.Error()
}

// takingLabel ticks the label the keyboard is on, or unticks it; on the row
// that creates one, creates it ticked and clears the line, ready for the next.
// A row saying why something cannot be a label takes nothing.
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
		return w
	}
	if row.Number > 0 {
		w.ticked[row.Number] = !w.ticked[row.Number]
	}
	return w
}

// keepingLabels closes the list and puts the labels ticked in the field, in
// alphabetical order, the way the field writes them.
func (w Window) keepingLabels() Window {
	var kept []string
	for i, label := range w.labelsOffered {
		if w.ticked[i+1] {
			kept = append(kept, label)
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
