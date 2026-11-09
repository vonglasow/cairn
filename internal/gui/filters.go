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
			w.filterPick = max(1, w.filterPick-1)
		}
		return w
	}
	switch name {
	case "down", "j":
		w.filterOn = min(len(w.filters)-1, w.filterOn+1)
	case "up", "k":
		w.filterOn = max(0, w.filterOn-1)
	case "enter", "space":
		return w.openingFilter(w.filterOn)
	}
	for i, f := range w.filters {
		if FilterKey(f) == name {
			return w.openingFilter(i)
		}
	}
	return w
}

// FilterKey is the key that opens a filter from the menu of filters: the first letter of its name.
func FilterKey(f engine.Filter) string { return f.Name[:1] }

func (w Window) openingFilter(at int) Window {
	w.filterOn, w.filterOpen, w.filterPick = at, true, max(1, w.filters[at].At)
	return w
}

func (w Window) openingFilterFromPill(at int) Window {
	w = w.filteringBoard().openingFilter(at)
	w.filterFromPill = true
	return w
}

func (w Window) pickingFilter(value int) Window {
	f := w.filters[w.filterOn]
	if f.Many {
		w.filters = w.filters.Toggled(w.filterOn, f.Values[value])
		return w
	}
	if value == f.At {
		value = 0
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
	return slices.Contains(f.Taken(), f.Values[value])
}

const (
	filterMenuWide  = 240
	filterValueWide = 300
	dropHead        = 36
	dropFoot        = 36
)

// FilterMenuSays is what heads the menu of filters.
const FilterMenuSays = "FILTER BY"

// FilterHint says how the open filter's values are taken: several ticked together, or one at a time.
func (w Window) FilterHint() string {
	if w.filters[w.filterOn].Many {
		return "ticked ones together"
	}
	return "one at a time"
}

// FilterFoot is the keys the foot of the open menu says, each with what it does.
func (w Window) FilterFoot() [][2]string {
	if !w.filterOpen {
		return [][2]string{{"↑↓", "walk"}, {"⏎", "open"}, {"⌫", "take off"}, {"⌥⌫", "all"}, {"esc", "close"}}
	}
	if w.filters[w.filterOn].Many {
		return [][2]string{{"↑↓", "walk"}, {"⏎", "tick"}, {"esc", "close"}}
	}
	return [][2]string{{"↑↓", "walk"}, {"⏎", "choose"}, {"esc", "close"}}
}

func (w Window) footWide() int {
	wide := 2 * padding
	for _, key := range w.FilterFoot() {
		wide += KeyChipWide(key[0]) + hintSpace + ends(w.measured(), key[1], small) + padding
	}
	return wide
}

// FiltersBox is the menu dropped under the pill it opened from: the filters, a line each with the key that opens it, or the open filter's values, each with how many tasks it would keep; its name at its head, its keys at its foot.
func (w Window) FiltersBox() Box {
	anchor := w.filterAnchor()
	if w.filterOpen {
		return w.droppedUnder(anchor, filterValueWide, len(w.filters[w.filterOn].Shown)-1)
	}
	return w.droppedUnder(anchor, filterMenuWide, len(w.filters))
}

func (w Window) droppedUnder(anchor Box, wide, rows int) Box {
	fixed := dropHead + dropFoot + inset
	most := max(1, (w.height-margin-anchor.Y-anchor.H-between-fixed)/menuRow)
	high := fixed + min(rows, most)*menuRow
	wide = min(w.width-2*margin, max(wide, w.footWide()))
	x := max(margin, min(anchor.X+anchor.W-wide, w.width-margin-wide))
	return Box{X: x, Y: anchor.Y + anchor.H + between, W: wide, H: high}
}

// DropRowAt is where the i-th line of a dropped menu is, under its head.
func DropRowAt(box Box, i int) Box {
	return Box{X: box.X + inset, Y: box.Y + dropHead + i*menuRow, W: box.W - 2*inset, H: menuRow}
}

func dropRoom(box Box) int { return max(1, (box.H-dropHead-dropFoot-inset)/menuRow) }

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
func (w Window) FilterRowAt(i int) Box { return DropRowAt(w.FiltersBox(), i) }

// FilterList is the values of the filter whose list is open, where each is drawn, and whether one is open.
func (w Window) FilterList() ([]Box, bool) {
	if !w.filterOpen || w.screen != OnFiltering {
		return nil, false
	}
	box := w.FiltersBox()
	room := dropRoom(box)
	shown := w.filters[w.filterOn].Shown
	from := 1 + max(0, min(w.filterPick-room, len(shown)-1-room))
	boxes := make([]Box, len(shown))
	for i := range shown {
		if i >= from && i < from+room {
			boxes[i] = DropRowAt(box, i-from)
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
)

// FilterPill is one pill of the row at the right of the tabs: where it is, what it says and the name it says it under, what it stands for, for a filter's which one, and whether it is faint for meaning nothing on the page in sight.
type FilterPill struct {
	Box   Box
	Said  string
	Name  string
	Kind  PillKind
	At    int
	Faint bool
}

// Cross is where a pill that clears what it stands for draws the cross that does it.
func (p FilterPill) Cross() (Box, bool) {
	if p.Kind != PillSearch && p.Kind != PillFilter {
		return Box{}, false
	}
	return Box{X: p.Box.X + p.Box.W - pillCross, Y: p.Box.Y, W: pillCross, H: p.Box.H}, true
}

// Pills is the row at the right of the tabs: the search kept, then every filter in force, the pill clearing them, and against the menu the pill adding one, which stands where it stood whatever is added; folded into one pill, in that pill's place, when they no longer fit between the tabs and the menu.
func (w Window) Pills() []FilterPill {
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
	var row []FilterPill
	crossed := func(name, value string, kind PillKind, at int) {
		wide := ends(m, name, small) + nameGap + ends(m, value, small) + 2*pillPad + pillCross
		row = append(row, FilterPill{Box: Box{W: wide}, Said: name + " " + value, Name: name, Kind: kind, At: at})
	}
	plain := func(said string, kind PillKind) {
		row = append(row, FilterPill{Box: Box{W: ends(m, said, small) + 2*pillPad}, Said: said, Kind: kind})
	}
	taken := 0
	if !w.looking.Empty() && w.screen != OnSearch {
		crossed("/", w.looking.Reads(), PillSearch, -1)
		taken++
	}
	for i, f := range w.filters {
		if len(f.Taken()) > 0 {
			crossed(f.Name, f.TakenSays(), PillFilter, i)
			row[len(row)-1].Faint = w.tab == Objectives && !f.ReachesGoals()
			taken++
		}
	}
	if taken > 0 {
		plain(ClearSays, PillClear)
	}
	plain(AddFilterSays, PillAdd)
	if widthOf(row) > end-start && taken > 0 {
		row = nil
		plain(ClearSays, PillClear)
		plain(FoldedSays(taken), PillFolded)
	}
	var pills []FilterPill
	x := max(start, end-widthOf(row))
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

func widthOf(row []FilterPill) int {
	wide := 0
	for _, pill := range row {
		wide += pill.Box.W + pillGap
	}
	return wide - pillGap
}

// nameGap is the room between what a filter's pill is named and the value it is on.
const nameGap = 5

// AddFilterSays is what the pill that opens the menu of filters says.
const AddFilterSays = "+ Filter"

// ClearSays is what the pill that takes every filter and the search away says.
const ClearSays = "Clear"

// FoldedSays is what the pill standing for every filter and the search says once they no longer fit: how many there are.
func FoldedSays(n int) string { return "Filters · " + strconv.Itoa(n) }

// TabCount is what a page's tab counts: what it shows, and of how many when the search or the filters narrow it.
func (w Window) TabCount(page Tab) string {
	shown := strconv.Itoa(w.Holds(page))
	if page == Time || (w.looking.Empty() && !w.filters.Narrows()) {
		return shown
	}
	if page == Objectives {
		return shown + "/" + strconv.Itoa(len(w.Reading().Goals))
	}
	return shown + "/" + strconv.Itoa(len(engine.OnTheBoard(w.Reading().Cards, w.archived)))
}

func (w Window) takingOffLast() Window {
	for i := len(w.filters) - 1; i >= 0; i-- {
		if len(w.filters[i].Taken()) > 0 {
			return w.takingOff(i)
		}
	}
	w.looking = Written{}
	return w
}

func (w Window) takingOff(at int) Window {
	cleared := slices.Clone(w.filters)
	cleared[at].At, cleared[at].Ticked = 0, nil
	w.filters = cleared
	return w
}

func (w Window) clickingPill(pill FilterPill, p Point) Window {
	if cross, ok := pill.Cross(); ok && cross.In(p) {
		if pill.Kind == PillSearch {
			w.looking = Written{}
			return w
		}
		return w.takingOff(pill.At)
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
		if pill.Box.In(p) {
			return w.clickingPill(pill, p), true
		}
	}
	return w, false
}
