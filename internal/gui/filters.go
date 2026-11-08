package gui

import (
	"slices"
	"strconv"

	"cairn/internal/engine"
)

// Filters are the board's filters, as the menus walk them.
func (w Window) Filters() engine.Filters { return w.filters }

// FilterOn is the line of the menu of filters the keyboard is on.
func (w Window) FilterOn() int { return w.filterOn }

// FilterPick is the value of the open filter the keyboard is on.
func (w Window) FilterPick() int { return w.filterPick }

// FilteredBy is what the board is filtered by, written the way the bar says it: nothing when no filter narrows it.
func (w Window) FilteredBy() string { return w.filters.Said() }

func (w Window) filteringBoard() Window {
	w.filters = engine.BoardFilters(w.snap, w.today()).Keeping(w.filters)
	w.screen, w.filterOpen, w.filterFromPill = OnFiltering, false, false
	return w
}

func (w Window) onFilters(name string) Window {
	if w.filterOpen {
		switch name {
		case "down", "j":
			w.filterPick = min(len(w.filters[w.filterOn].Shown)-1, w.filterPick+1)
		case "up", "k":
			w.filterPick = max(0, w.filterPick-1)
		}
		return w
	}
	switch name {
	case "down", "j":
		w.filterOn = min(len(w.filters)-1, w.filterOn+1)
	case "up", "k":
		w.filterOn = max(0, w.filterOn-1)
	case "right", "l":
		w.filters = w.filters.Walked(w.filterOn, 1)
	case "left", "h":
		w.filters = w.filters.Walked(w.filterOn, -1)
	case "enter", "space":
		return w.openingFilter(w.filterOn)
	}
	return w
}

func (w Window) openingFilter(at int) Window {
	w.filterOn, w.filterOpen, w.filterPick = at, true, w.filters[at].At
	return w
}

func (w Window) openingFilterFromPill(at int) Window {
	w = w.filteringBoard().openingFilter(at)
	w.filterFromPill = true
	return w
}

func (w Window) pickingFilter(value int) Window {
	f := w.filters[w.filterOn]
	if f.Many && value > 0 {
		w.filters = w.filters.Toggled(w.filterOn, f.Values[value])
		return w
	}
	w.filters = w.filters.Walked(w.filterOn, value-f.At)
	return w.closingFilter()
}

// closingFilter shuts the open filter's values: back to the menu of filters when they were opened from it, back to the board when from a pill.
func (w Window) closingFilter() Window {
	w.filterOpen = false
	if w.filterFromPill {
		w.screen, w.filterFromPill = OnTheBoard, false
	}
	return w
}

// FilterCounts is how many tasks each value of the open filter would keep under the other filters, in the order its values are listed.
func (w Window) FilterCounts() []int {
	if !w.filterOpen || w.screen != OnFiltering {
		return nil
	}
	counts, err := w.filters.Counts(w.Reading().Cards, w.filterOn)
	if err != nil {
		return nil
	}
	return counts
}

// FilterTaken says whether a value of the open filter is one it is on.
func (w Window) FilterTaken(value int) bool {
	f := w.filters[w.filterOn]
	return slices.Contains(f.Taken(), f.Values[value]) || (value == 0 && len(f.Taken()) == 0)
}

const (
	filterMenuWide  = 260
	filterValueWide = 280
)

// FiltersBox is the menu of filters opened under the pill that adds one: a line a filter, what it is on at its right, its offers at its foot.
func (w Window) FiltersBox() Box {
	anchor := w.filterAnchor()
	if w.filterOpen {
		return w.droppedUnder(anchor, filterValueWide, len(w.filters[w.filterOn].Shown))
	}
	return w.droppedUnder(anchor, filterMenuWide, len(w.filters))
}

func (w Window) droppedUnder(anchor Box, wide, rows int) Box {
	fixed := FirstRow(Box{}) + between + padding + barHigh + padding
	most := max(1, (w.height-margin-anchor.Y-anchor.H-between-fixed)/menuRow)
	high := fixed + min(rows, most)*menuRow
	wide = min(w.width-2*margin, max(wide, w.offersWide(), w.headWide()+padding+pillCross))
	x := max(margin, min(anchor.X, w.width-margin-wide))
	return Box{X: x, Y: anchor.Y + anchor.H + between, W: wide, H: high}
}

func (w Window) filterAnchor() Box {
	for _, pill := range w.Pills() {
		if w.filterFromPill && pill.Kind == PillFilter && pill.At == w.filterOn {
			return pill.Box
		}
	}
	for _, pill := range w.Pills() {
		if pill.Kind == PillAdd || pill.Kind == PillFolded {
			return pill.Box
		}
	}
	l := w.Layout()
	return Box{X: l.Tabs.X, Y: l.Tabs.Y, W: 0, H: l.Tabs.H}
}

// FilterRowAt is where the i-th filter's line is in the menu of filters.
func (w Window) FilterRowAt(i int) Box { return ListRowAt(w.FiltersBox(), i) }

// FilterList is the values of the filter whose list is open, where each is drawn, and whether one is open.
func (w Window) FilterList() ([]Box, bool) {
	if !w.filterOpen || w.screen != OnFiltering {
		return nil, false
	}
	box := w.FiltersBox()
	room := listRoom(box)
	shown := w.filters[w.filterOn].Shown
	from := max(0, min(w.filterPick-room+1, len(shown)-room))
	boxes := make([]Box, len(shown))
	for i := range shown {
		boxes[i] = ListRowAt(box, i-from)
		if i < from || i >= from+room {
			boxes[i] = Box{}
		}
	}
	return boxes, true
}

func (w Window) clickingFilters(p Point) Window {
	for _, pill := range w.Pills() {
		if pill.Box.In(p) {
			return w.clickingPill(pill, p)
		}
	}
	box := w.FiltersBox()
	if values, open := w.FilterList(); open {
		for i, value := range values {
			if value.In(p) {
				w.filterPick = i
				return w.pickingFilter(i)
			}
		}
		if !box.In(p) {
			w.screen, w.filterOpen, w.filterFromPill = OnTheBoard, false, false
		}
		return w
	}
	for i := range w.filters {
		if w.FilterRowAt(i).In(p) {
			return w.openingFilter(i)
		}
	}
	if !box.In(p) {
		w.screen = OnTheBoard
	}
	return w
}

// PillKind says what a pill after the tabs stands for.
type PillKind int

const (
	// PillSearch is the search kept, its cross clearing it.
	PillSearch PillKind = iota
	// PillFilter is a filter in force, its cross clearing it.
	PillFilter
	// PillAdd opens the menu of filters.
	PillAdd
	// PillClear takes every filter and the search away.
	PillClear
	// PillFolded stands for every filter and the search when they no longer fit.
	PillFolded
	// PillCount says how many tasks are kept of how many, at the end of the row.
	PillCount
)

// Pill is one pill of the row after the tabs: where it is, what it says, what it stands for, for a filter's which one, and whether it is faint for meaning nothing on the page in sight.
type Pill struct {
	Box   Box
	Said  string
	Kind  PillKind
	At    int
	Faint bool
}

// Cross is where a pill that clears what it stands for draws the cross that does it.
func (p Pill) Cross() (Box, bool) {
	if p.Kind != PillSearch && p.Kind != PillFilter {
		return Box{}, false
	}
	return Box{X: p.Box.X + p.Box.W - pillCross, Y: p.Box.Y, W: pillCross, H: p.Box.H}, true
}

// Pills is the row after the tabs: the search kept, then every filter in force, a pill to add one and one to clear them, folded into one pill when they no longer fit before the menu; and at its end how many tasks are kept of how many.
func (w Window) Pills() []Pill {
	if w.screen != OnTheBoard && w.screen != OnFiltering && w.screen != OnSearch {
		return nil
	}
	l := w.Layout()
	tab := TabAt(l, 0)
	y, high := tab.Y+(tab.H-searchPillHigh)/2, searchPillHigh
	start := TabAt(l, len(w.Shown())).X + between
	end := l.Tabs.X + l.Tabs.W
	if !w.native {
		end = MenuButton(l).X - between
	}
	m := w.measured()
	var row []Pill
	crossed := func(said string, kind PillKind, at int) {
		row = append(row, Pill{Box: Box{W: ends(m, said, small) + 2*pillPad + pillCross}, Said: said, Kind: kind, At: at})
	}
	plain := func(said string, kind PillKind) {
		row = append(row, Pill{Box: Box{W: ends(m, said, small) + 2*pillPad}, Said: said, Kind: kind})
	}
	taken := 0
	if !w.looking.Empty() && w.screen != OnSearch {
		crossed("/ "+w.looking.Reads(), PillSearch, -1)
		taken++
	}
	for i, f := range w.filters {
		if len(f.Taken()) > 0 {
			crossed(f.Name+" "+f.TakenSays(), PillFilter, i)
			row[len(row)-1].Faint = w.tab == Objectives && !f.ReachesGoals()
			taken++
		}
	}
	plain(AddFilterSays, PillAdd)
	if taken > 0 {
		plain(ClearSays, PillClear)
	}
	var count []Pill
	if said, ok := w.KeptCount(); ok {
		count = []Pill{{Box: Box{W: monoWide(said, small) + 2*pillPad}, Said: said, Kind: PillCount}}
	}
	room := func() int {
		if len(count) > 0 {
			return end - count[0].Box.W - pillGap - start
		}
		return end - start
	}
	if widthOf(row) > room() && taken > 0 {
		row = nil
		plain(FoldedSays(taken), PillFolded)
		plain(ClearSays, PillClear)
	}
	if widthOf(row) > room() {
		count = nil
	}
	var pills []Pill
	for _, pill := range count {
		pill.Box.X, pill.Box.Y, pill.Box.H = end-pill.Box.W, y, high
		pills = append(pills, pill)
	}
	x := start
	for _, pill := range row {
		if x+pill.Box.W > end {
			break
		}
		pill.Box.X, pill.Box.Y, pill.Box.H = x, y, high
		pills = append(pills, pill)
		x += pill.Box.W + pillGap
	}
	return pills
}

func widthOf(row []Pill) int {
	wide := 0
	for _, pill := range row {
		wide += pill.Box.W + pillGap
	}
	return wide - pillGap
}

// AddFilterSays is what the pill that opens the menu of filters says.
const AddFilterSays = "+ Filter"

// ClearSays is what the pill that takes every filter and the search away says.
const ClearSays = "Clear"

// FoldedSays is what the pill standing for every filter and the search says once they no longer fit: how many there are.
func FoldedSays(n int) string { return "Filters · " + strconv.Itoa(n) }

// KeptCount is how many tasks the search and the filters keep of how many the board holds — objectives on the objectives' page — and whether anything narrows it at all.
func (w Window) KeptCount() (string, bool) {
	if w.looking.Empty() && !w.filters.Narrows() {
		return "", false
	}
	if w.tab == Objectives {
		return strconv.Itoa(len(w.Showing().Goals)) + " of " + strconv.Itoa(len(w.Reading().Goals)) + " objectives", true
	}
	return strconv.Itoa(len(w.Showing().Cards)) + " of " + strconv.Itoa(len(w.Reading().Cards)) + " tasks", true
}

func (w Window) clickingPill(pill Pill, p Point) Window {
	if cross, ok := pill.Cross(); ok && cross.In(p) {
		if pill.Kind == PillSearch {
			w.looking = Written{}
			return w
		}
		cleared := slices.Clone(w.filters)
		cleared[pill.At].At, cleared[pill.At].Ticked = 0, nil
		w.filters = cleared
		return w
	}
	switch pill.Kind {
	case PillFilter:
		return w.openingFilterFromPill(pill.At)
	case PillAdd, PillFolded:
		return w.filteringBoard()
	case PillClear:
		w.filters, w.looking = w.filters.Cleared(), Written{}
		w.screen = OnTheBoard
		return w
	case PillSearch:
		w.screen, w.searchFrom = OnSearch, OnTheBoard
		return w
	}
	return w
}

func (w Window) clickingPills(p Point) (Window, bool) {
	if w.screen != OnTheBoard {
		return w, false
	}
	for _, pill := range w.Pills() {
		if pill.Box.In(p) && pill.Kind != PillCount {
			return w.clickingPill(pill, p), true
		}
	}
	return w, false
}
