package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine"
	"cairn/internal/gui/geom"
)

const (
	sectionWide     = 72
	formNamed       = 76
	formValueFewest = 110
)

const (
	saidFewest   = 4
	saidMost     = 10
	listedFewest = 3
)

// FormPlace is where one field of a form is drawn: the name it is written under, the line its value starts on, and how many rows it is given.
func FormPlace(over geom.Box, fields []board.Field, at int) (name, value geom.Box, rows int) {
	places, _, _, _ := formLayout(over, fields)
	p := places[at]
	return p.name, p.value, p.rows
}

// FormHeads are the heads of the sections a form groups its short fields in, each where it is written: a section none of the form's fields belong to has no head.
func FormHeads(over geom.Box, fields []board.Field) []Readable {
	_, heads, _, _ := formLayout(over, fields)
	return heads
}

// RepeatsReadAt is where a form reads back the rule typed in its repeats field: two lines under the line the field stands on, and nothing for a form with no such field.
func RepeatsReadAt(over geom.Box, fields []board.Field) (geom.Box, bool) {
	_, _, _, read := formLayout(over, fields)
	return read, read.H > 0
}

// FormOrder is the fields of a form in the order they are drawn: the lone fields, the long ones and the sections, read down each column, the column beside last.
func FormOrder(over geom.Box, fields []board.Field) []int {
	alone, long, sections := formGroups(fields)
	order := append(alone, long[:min(1, len(long))]...)
	if over.Beside > 0 {
		order = append(order, long[min(1, len(long)):]...)
	}
	for _, row := range sections {
		order = append(order, row...)
	}
	if over.Beside == 0 {
		order = append(order, long[min(1, len(long)):]...)
	}
	return order
}

func formGroups(fields []board.Field) (alone, long []int, rows [][]int) {
	rows = make([][]int, len(board.FormSections))
	for i, field := range fields {
		section, in := board.SectionOf(field)
		switch {
		case field.Long:
			long = append(long, i)
		case in:
			rows[section] = append(rows[section], i)
		default:
			alone = append(alone, i)
		}
	}
	return alone, long, rows
}

const readBackHigh = 2 * lineHigh

func readsBack(row []int, fields []board.Field) bool {
	for _, i := range row {
		if fields[i].Arg == board.RepeatsField().Arg {
			return true
		}
	}
	return false
}

// FormSize is the size a field of a form is written at: the title as large as a panel's heading, since it is what the task is called and the first thing read; everything else at the size the board reads at.
func FormSize(field board.Field) int {
	if field.Arg == "title" {
		return heading
	}
	return reads
}

// TickedPlaces is where the steps of a task being edited are drawn as tiles in their frame: one under the other from its top, from the first in sight, as many as fit above the line that adds another, and that line under the last of them.
func TickedPlaces(frame geom.Box, parts, from int) (tiles []geom.Box, adding geom.Box) {
	inside := Inside(frame)
	shown := min(parts-from, ticksFit(frame))
	for at := range shown {
		tiles = append(tiles, geom.Box{X: inside.X, Y: inside.Y + at*geom.SubHeight, W: inside.W, H: geom.SubHeight})
	}
	return tiles, geom.Box{X: inside.X + padding, Y: inside.Y + shown*geom.SubHeight, W: inside.W - padding, H: lineHigh}
}

func ticksFit(frame geom.Box) int { return max(0, (Inside(frame).H-lineHigh)/geom.SubHeight) }

// AddingAPart is what the steps of a task being edited say under their tiles: another is added by typing a line, which is where the keyboard goes.
const AddingAPart = "+ new step — a line of its own"

// FormRule is the line across the foot of a form, above what it offers: what says the offers are a line of their own and not one more field.
func FormRule(over geom.Box) geom.Box {
	offers := PanelOfferAt(over, 0)
	return geom.Box{X: over.X + padding, Y: offers.Y - between, W: over.W - padding - padding, H: 1}
}

type placed struct {
	name, value geom.Box
	rows        int
}

const longHigh = 44

// Shifted is a panel moved by so much, and held wholly inside a window of that size.
func Shifted(b geom.Box, by geom.Point, width, height int) geom.Box {
	b.X = max(0, min(b.X+by.X, width-b.W))
	b.Y = max(0, min(b.Y+by.Y, height-b.H))
	return b
}

func formLayout(over geom.Box, fields []board.Field) ([]placed, []Readable, int, geom.Box) {
	places := make([]placed, len(fields))
	left, full := over.X+padding, over.W-padding-padding
	top := FirstRow(over)

	alone, long, rows := formGroups(fields)
	sectionLeft, sectionFull, sectionTop := left, full, -1
	if over.Beside > 0 {
		sectionLeft, sectionFull, sectionTop = over.X+over.W+padding, over.Beside-padding-padding, top+apart
	}
	across := fieldsAcross(sectionFull)
	lines := 0
	readHigh := 0
	for _, row := range rows {
		lines += (len(row) + across - 1) / across
		if readsBack(row, fields) {
			readHigh = readBackHigh
		}
	}
	readable := readHigh > 0
	if sectionTop >= 0 {
		readable = readable && sectionTop+lines*apart+readHigh <= FormRule(over).Y-between
		lines, readHigh = 0, 0
	}
	wanted := make(map[int]int, len(long))
	for _, i := range long {
		wanted[i] = rowsWanted(fields[i], full)
	}
	fixed := len(alone)*apart + lines*apart + readHigh + between + len(long)*longHigh
	room := (FormRule(over).Y - top - fixed) / lineHigh
	given := shared(long, wanted, room)
	if sectionTop >= 0 && len(long) == 2 {
		given = map[int]int{long[0]: room * 3 / 5, long[1]: room - room*3/5}
	}
	if sectionTop >= 0 && len(long) == 1 {
		given = map[int]int{long[0]: room * 3 / 5}
	}

	y := top
	for _, i := range alone {
		places[i] = placed{name: geom.Box{X: left, Y: y, H: lineHigh}, value: geom.Box{X: left, Y: y, W: full, H: lineHigh}, rows: 1}
		y += apart
	}
	placeLong := func(i int) {
		places[i] = placed{
			name:  geom.Box{X: left, Y: y, W: full, H: lineHigh},
			value: geom.Box{X: left, Y: y + lineHigh + between + inset, W: full, H: lineHigh},
			rows:  given[i],
		}
		y += longHigh + given[i]*lineHigh
	}
	if len(long) > 0 {
		placeLong(long[0])
	}
	var heads []Readable
	var read geom.Box
	sy := y
	if sectionTop >= 0 {
		sy = sectionTop
	}
	headWide := sectionWide
	if sectionFull-sectionWide-formNamed < formValueFewest {
		headWide = 0
	}
	for s, row := range rows {
		if len(row) == 0 {
			continue
		}
		if headWide > 0 {
			heads = append(heads, one("head:"+board.FormSections[s].Head, geom.Box{X: sectionLeft, Y: sy, W: headWide, H: lineHigh}, board.FormSections[s].Head, small))
		}
		wide := min(across, len(row))
		cell := (sectionFull - headWide - (wide-1)*padding) / wide
		for at, i := range row {
			x := sectionLeft + headWide + at%wide*(cell+padding)
			places[i] = placed{
				name:  geom.Box{X: x, Y: sy, W: formNamed, H: lineHigh},
				value: geom.Box{X: x + formNamed, Y: sy, W: cell - formNamed, H: lineHigh},
				rows:  1,
			}
			if at%wide == wide-1 || at == len(row)-1 {
				sy += apart
				if line := row[at-at%wide : at+1]; readable && readsBack(line, fields) {
					read = geom.Box{X: sectionLeft, Y: sy - between, W: sectionFull, H: readBackHigh}
					sy += readBackHigh
				}
			}
		}
	}
	if sectionTop < 0 {
		y = sy
	}
	y += between
	for _, i := range long[min(1, len(long)):] {
		placeLong(i)
	}
	return places, heads, y, read
}

func fieldsAcross(full int) int {
	return max(1, (full-sectionWide+padding)/(formNamed+formValueFewest+padding))
}

func rowsWanted(field board.Field, wide int) int {
	if field.Name == "steps" {
		tiles := len(engine.PartsIn(field.Value))*geom.SubHeight + lineHigh
		return max(listedFewest, (tiles+lineHigh-1)/lineHigh)
	}
	inside := Inside(Framed(geom.Box{W: wide, H: lineHigh}, 1))
	return min(saidMost, max(saidFewest, len(Wrapped(field.Value, room(inside)).Said)))
}

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

// FieldMark is where a field chosen from a list, or on a calendar, says so: a mark at the right end of its frame.
func FieldMark(frame geom.Box) geom.Box {
	return geom.Box{X: frame.X + frame.W - padding - markWide, Y: frame.Y + (frame.H-markWide)/2, W: markWide, H: markWide}
}

const markWide = 10

// WalkMark is where a chevron walking back or on is drawn, in the middle of the room it is clicked in.
func WalkMark(walk geom.Box) geom.Box {
	return geom.Box{X: walk.X + (walk.W-markWide)/2, Y: walk.Y + (walk.H-markWide)/2, W: markWide, H: markWide}
}
