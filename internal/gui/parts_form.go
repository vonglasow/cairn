package gui

import (
	"slices"

	"cairn/internal/reading"
)

// Where the fields of a form go, read from the top: the title across the whole
// width, the description under it, the short fields grouped a section to a
// row, then the parts — and what the form offers on a line of its own, under a
// rule at its foot.
//
// It was one column, every field under the last, and then two: the short
// fields down the left, the long ones sharing the right. The first ran the
// parts over Save and Cancel once they became a second long field; the second
// left a hole under eleven fields and two frames as tall as the panel with
// three lines in them.

// formSections are the rows the short fields are grouped in, by what each
// field asks the engine for: a column and a standing are both a state.
var formSections = []struct {
	head string
	args []string
}{
	{"STATE", []string{"state", "priority", "label"}},
	{"TIME", []string{"deadline", "estimate", "spent"}},
	{"BLOCKED", []string{"blocker", "because"}},
	{"LINKS", []string{"parent", "objective"}},
}

// sectionWide is the room a section's head takes before its fields; formNamed
// is what a field's name takes before its box, inside a section.
const (
	sectionWide = 72
	formNamed   = 76
	// formValueFewest is the least a short field's box is given: a date, a
	// column's name, a few labels.
	formValueFewest = 110
)

// The rows a long field is given: the description a line for each it says,
// between four and ten; the parts a line for each tile and one to add another,
// three at the least.
const (
	saidFewest   = 4
	saidMost     = 10
	listedFewest = 3
)

// FormPlace is where one field of a form is drawn: the name it is written
// under, the line its value starts on, and how many rows it is given.
func FormPlace(over Box, fields []reading.Field, at int) (name, value Box, rows int) {
	places, _, _ := formLayout(over, fields)
	p := places[at]
	return p.name, p.value, p.rows
}

// FormHeads are the heads of the sections a form groups its short fields in,
// each where it is written: a section none of the form's fields belong to has
// no head.
func FormHeads(over Box, fields []reading.Field) []Readable {
	_, heads, _ := formLayout(over, fields)
	return heads
}

// FormSize is the size a field of a form is written at: the title as large as
// a panel's heading, since it is what the task is called and the first thing
// read; everything else at the size the board reads at.
func FormSize(field reading.Field) int {
	if field.Arg == "title" {
		return heading
	}
	return reads
}

// TickedPlaces is where the parts of a task being edited are drawn as tiles in
// their frame: one under the other from its top, from the first in sight, as
// many as fit above the line that adds another, and that line under the last
// of them.
func TickedPlaces(frame Box, parts, from int) (tiles []Box, adding Box) {
	inside := Inside(frame)
	shown := min(parts-from, ticksFit(frame))
	for at := range shown {
		tiles = append(tiles, Box{X: inside.X, Y: inside.Y + at*SubHeight, W: inside.W, H: SubHeight})
	}
	return tiles, Box{X: inside.X + padding, Y: inside.Y + shown*SubHeight, W: inside.W - padding, H: lineHigh}
}

// ticksFit is how many tiles a frame holds above the line that adds.
func ticksFit(frame Box) int { return max(0, (Inside(frame).H-lineHigh)/SubHeight) }

// AddingAPart is what the parts of a task being edited say under their tiles:
// another is added by typing a line, which is where the keyboard goes.
const AddingAPart = "+ new part — a line of its own"

// FormRule is the line across the foot of a form, above what it offers: what
// says the offers are a line of their own and not one more field.
func FormRule(over Box) Box {
	offers := PanelOfferAt(over, 0)
	return Box{X: over.X + padding, Y: offers.Y - between, W: over.W - padding - padding, H: 1}
}

type placed struct {
	name, value Box
	rows        int
}

// sectionOf is which section a field is grouped in, and false for one in none.
func sectionOf(field reading.Field) (int, bool) {
	for i, s := range formSections {
		if slices.Contains(s.args, field.Arg) {
			return i, true
		}
	}
	return 0, false
}

// longHigh is what a long field takes besides its rows: its name above it, the
// room its frame keeps, and the gap before what follows.
//
// Written as the number and not as that sum, for the reason apart is: a
// constant's arithmetic runs on no line, so no test reaches it. It is lineHigh
// + between + inset + inset + padding, which is 18 + 6 + 4 + 4 + 12, and the
// test beside FormPlace holds the two together.
const longHigh = 44

// FormOver is the panel a form is drawn in: the one every other screen uses,
// cut to what the form holds. Its top stays where it was, and so does every
// field; what rises is the rule and the offers under it, so a tall window
// does not put a floor of empty ground between the parts and Save.
func FormOver(width, height int, fields []reading.Field) Box {
	over := Over(width, height)
	return cutTo(over, over, fields)
}

// OpenOver is the panel a form open on something already there is drawn in:
// cut the same way, to the fields laid out in its left column, with what
// stands beside them as tall as they are.
func OpenOver(width, height int, fields []reading.Field) Box {
	over := Over(width, height)
	return cutTo(over, FormFields(over), fields)
}

// cutTo cuts a panel to the fields laid out in a part of it, with room under
// them for the rule and the offers.
func cutTo(over, laid Box, fields []reading.Field) Box {
	_, _, end := formLayout(laid, fields)
	over.H = min(over.H, end-over.Y+between+barHigh+padding)
	return over
}

// formLayout lays every field of a form out, and the heads of its sections,
// and says where what it laid out ends.
func formLayout(over Box, fields []reading.Field) ([]placed, []Readable, int) {
	places := make([]placed, len(fields))
	left, full := over.X+padding, over.W-padding-padding
	top := FirstRow(over)

	// What is laid out, in the order it is read: the short fields in no
	// section, the first long field, the sections, the other long fields.
	var alone, long []int
	rows := make([][]int, len(formSections))
	for i, field := range fields {
		section, in := sectionOf(field)
		switch {
		case field.Long:
			long = append(long, i)
		case in:
			rows[section] = append(rows[section], i)
		default:
			alone = append(alone, i)
		}
	}
	// A section with more fields than a line holds goes on over another
	// line, under the same head.
	across := fieldsAcross(full)
	lines := 0
	for _, row := range rows {
		lines += (len(row) + across - 1) / across
	}
	wanted := make(map[int]int, len(long))
	for _, i := range long {
		wanted[i] = rowsWanted(fields[i], full)
	}
	fixed := len(alone)*apart + lines*apart + between + len(long)*longHigh
	given := shared(long, wanted, (FormRule(over).Y-top-fixed)/lineHigh)

	y := top
	for _, i := range alone {
		places[i] = placed{name: Box{X: left, Y: y, H: lineHigh}, value: Box{X: left, Y: y, W: full, H: lineHigh}, rows: 1}
		y += apart
	}
	placeLong := func(i int) {
		places[i] = placed{
			name:  Box{X: left, Y: y, W: full, H: lineHigh},
			value: Box{X: left, Y: y + lineHigh + between + inset, W: full, H: lineHigh},
			rows:  given[i],
		}
		y += longHigh + given[i]*lineHigh
	}
	if len(long) > 0 {
		placeLong(long[0])
	}
	var heads []Readable
	for s, row := range rows {
		if len(row) == 0 {
			continue
		}
		heads = append(heads, one("head:"+formSections[s].head, Box{X: left, Y: y, W: sectionWide, H: lineHigh}, formSections[s].head, small))
		wide := min(across, len(row))
		cell := (full - sectionWide - (wide-1)*padding) / wide
		for at, i := range row {
			x := left + sectionWide + at%wide*(cell+padding)
			places[i] = placed{
				name:  Box{X: x, Y: y, W: formNamed, H: lineHigh},
				value: Box{X: x + formNamed, Y: y, W: cell - formNamed, H: lineHigh},
				rows:  1,
			}
			if at%wide == wide-1 || at == len(row)-1 {
				y += apart
			}
		}
	}
	y += between
	for _, i := range long[min(1, len(long)):] {
		placeLong(i)
	}
	return places, heads, y
}

// fieldsAcross is how many short fields a line of a form holds: as many as
// leave each its name and the least a box is given, and one at the least.
func fieldsAcross(full int) int {
	return max(1, (full-sectionWide+padding)/(formNamed+formValueFewest+padding))
}

// rowsWanted is how many rows a long field asks for, from what it holds now.
func rowsWanted(field reading.Field, wide int) int {
	if field.Name == "parts" {
		tiles := len(reading.PartsIn(field.Value))*SubHeight + lineHigh
		return max(listedFewest, (tiles+lineHigh-1)/lineHigh)
	}
	inside := Inside(Framed(Box{W: wide, H: lineHigh}, 1))
	return min(saidMost, max(saidFewest, len(Wrapped(field.Value, room(inside)).Said)))
}

// shared gives the long fields the rows the form has for them. The first — the
// description — keeps its fewest while the others can keep theirs, and gives
// way down to a line where they cannot: a list of parts nothing of which is in
// sight is a list nobody can work. The others are then given what they want
// of the rest, in turn, and the first takes what is left, up to what it wants.
// One row at the least for every one of them, the way every frame keeps one.
func shared(long []int, wanted map[int]int, room int) map[int]int {
	given := make(map[int]int, len(long))
	if len(long) == 0 {
		return given
	}
	kept := min(saidFewest, wanted[long[0]], max(1, room-(len(long)-1)*listedFewest))
	left := room - kept
	for _, i := range long[1:] {
		given[i] = max(1, min(wanted[i], left))
		left -= given[i]
	}
	given[long[0]] = max(1, min(wanted[long[0]], left+kept))
	return given
}
