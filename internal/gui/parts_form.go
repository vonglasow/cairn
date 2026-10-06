package gui

import "cairn/internal/reading"

// Where the fields of a form go: the short ones down the left, the long ones
// down the right, and what the form offers on a line of its own at its foot.
//
// It was one column, every field under the last. The form of a task grew a
// second long field — its parts — and in a window under eight hundred points
// the parts ran over what the form offers and out of the panel, the two
// buttons written across a field somebody was typing into.

// FormPlace is where one field of a form is drawn: the name it is written
// under, the line its value starts on, and how many rows it is given.
//
// A short field is a row of the left column, its name beside it, as on an open
// task. A long one is a block of the right column, its name above it; the long
// fields share that column between them, all the way down to the rule over the
// offers, so a description has the room the history has on an open task.
func FormPlace(over Box, fields []reading.Field, at int) (name, value Box, rows int) {
	left, right := DetailFields(over), DetailHistory(over)
	top := FirstRow(over)
	if !fields[at].Long {
		short := 0
		for i := range fields[:at] {
			if !fields[i].Long {
				short++
			}
		}
		name, value = FieldRow(left, top+short*apart)
		return name, value, 1
	}
	long, before := 0, 0
	for i := range fields {
		if fields[i].Long {
			long++
			if i < at {
				before++
			}
		}
	}
	// Each block ends with the gap before what follows it: the next one's
	// name, or the rule — and the last runs to the rule itself, so what the
	// share left over when it was divided is not lost under it.
	rule := FormRule(over).Y
	block := (rule - top) / long
	from := top + before*block
	till := from + block
	if before == long-1 {
		till = rule
	}
	name = Box{X: right.X + padding, Y: from, W: right.W - padding - padding, H: lineHigh}
	value = Box{X: name.X, Y: from + lineHigh + between + inset, W: name.W, H: lineHigh}
	// What the block has left once the name, the frame's own room and the gap
	// before the next are taken — and one row at least, the way every frame
	// keeps one.
	rows = max(1, (till-from-lineHigh-between-inset-inset-between)/lineHigh)
	return name, value, rows
}

// FormRule is the line across the foot of a form, above what it offers: what
// says the offers are a line of their own and not one more field.
func FormRule(over Box) Box {
	offers := PanelOfferAt(over, 0)
	return Box{X: over.X + padding, Y: offers.Y - between, W: over.W - padding - padding, H: 1}
}
