package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"slices"
	"strconv"
	"strings"
)

func (w Window) filteringBoard() Window {
	w.Filters = board.FiltersOf(w.Snap, w.Today()).Keeping(w.Filters)
	w.Screen, w.FilterOpen, w.FilterFromPill, w.ViewWalked = board.OnFiltering, false, false, 0
	return w
}

func (w Window) walkingTheMenu(by int) Window {
	views := len(w.Snap.Views)
	if w.ViewsAbove() == 0 {
		views = 0
	}
	switch {
	case w.ViewWalked > 0 && by > 0 && w.ViewWalked == views:
		w.ViewWalked, w.FilterOn = 0, 0
	case w.ViewWalked > 0:
		w.ViewWalked = max(1, w.ViewWalked+by)
	case by < 0 && w.FilterOn == 0 && views > 0:
		w.ViewWalked = views
	default:
		w.FilterOn = min(len(w.Filters)-1, max(0, w.FilterOn+by))
	}
	return w
}

func (w Window) onFilters(name string) Window {
	if w.FilterOpen {
		switch name {
		case "down", "j":
			return w.walkingValues(1)
		case "up", "k":
			return w.walkingValues(-1)
		case "right":
			return w.openingDeeper()
		case "left":
			return w.goingBack()
		case "space":
			return w.takingValue(w.FilterPick, false)
		}
		return w
	}
	if opened, ok := w.openingByFigure(name); ok {
		return opened
	}
	switch name {
	case "v":
		return w.startingKeep()
	case "down", "j":
		return w.walkingTheMenu(1)
	case "up", "k":
		return w.walkingTheMenu(-1)
	case "enter", "space", "right":
		if w.ViewWalked > 0 {
			return w.openingView(w.Snap.Views[w.ViewWalked-1].Place)
		}
		return w.openingFilter(w.FilterOn)
	}
	for i, f := range w.Filters {
		if FilterKey(f) == name {
			return w.openingFilter(i)
		}
	}
	return w
}

// FilterKey is the key that opens a filter from the menu of filters: the first letter of its name.
func FilterKey(f board.Filter) string { return f.Name[:1] }

func (w Window) openingFilter(at int) Window {
	w.FilterOn, w.FilterOpen, w.FilterPick, w.ViewWalked, w.FilterTrail = at, true, max(1, w.Filters[at].At), 0, nil
	if first := w.ValueColumns()[0]; len(first) > 0 && !slices.Contains(first, w.FilterPick) {
		w.FilterPick = first[0]
	}
	return w
}

func (w Window) walkingValues(by int) Window {
	columns := w.ValueColumns()
	last := columns[len(columns)-1]
	if len(last) == 0 {
		return w
	}
	at := max(0, slices.Index(last, w.FilterPick))
	w.FilterPick = last[min(len(last)-1, max(0, at+by))]
	return w
}

func (w Window) openingDeeper() Window {
	columns := w.ValueColumns()
	if !w.OpensAColumn(w.FilterPick, len(columns)-1) {
		return w
	}
	g, _ := w.GoalOfValue(w.Filters[w.FilterOn].Values[w.FilterPick])
	w.FilterTrail = append(slices.Clone(w.FilterTrail), g.Number)
	return w
}

func (w Window) goingBack() Window {
	if len(w.FilterTrail) == 0 {
		return w.closingFilter()
	}
	n := w.FilterTrail[len(w.FilterTrail)-1]
	w.FilterTrail = w.FilterTrail[:len(w.FilterTrail)-1]
	w.FilterPick = max(1, slices.Index(w.Filters[w.FilterOn].Values, strconv.Itoa(n)))
	return w
}

func (w Window) choosingValue() Window {
	if w.OpensAColumn(w.FilterPick, len(w.ValueColumns())-1) {
		return w.openingDeeper()
	}
	return w.takingValue(w.FilterPick, true)
}

func (w Window) takingValue(value int, closing bool) Window {
	f := w.Filters[w.FilterOn]
	if f.Many {
		w.Filters = w.Filters.Toggled(w.FilterOn, f.Values[value])
	} else {
		if value == f.At {
			value = 0
		}
		w.Filters = w.Filters.Walked(w.FilterOn, value-f.At)
	}
	if closing {
		return w.closingMenu()
	}
	return w
}

func (w Window) closingMenu() Window {
	w.Screen, w.FilterOpen, w.FilterFromPill, w.FilterTrail = board.OnTheBoard, false, false, nil
	return w
}

func (w Window) openingFilterFromPill(at int) Window {
	w = w.filteringBoard().openingFilter(at)
	w.FilterFromPill = true
	return w
}

// closingFilter shuts the open filter's values: back to the menu of filters when they were opened from it, back to the board when from a pill.
func (w Window) closingFilter() Window {
	w.FilterOpen, w.FilterTrail = false, nil
	if w.FilterFromPill {
		w.Screen, w.FilterFromPill = board.OnTheBoard, false
	}
	return w
}

const (
	filterColumnWide = 232
	keepingWide      = 420
	filterCountWide  = 32
	dropHead         = 36
	dropFoot         = 36
)

// FilterValueParts is where a value of a filter writes its name and how many tasks it keeps, the name stopping short of the count.
func FilterValueParts(row geom.Box) (geom.Box, geom.Box) {
	line := Inside(row)
	named := PastCheck(line)
	count := geom.Box{X: line.X + line.W - filterCountWide, Y: line.Y, W: filterCountWide, H: line.H}
	named.W = count.X - between - named.X
	return named, count
}

func (w Window) cutFilterAt(p geom.Point) ([]string, geom.Box, int) {
	if !w.FilterOpen || w.Screen != board.OnFiltering {
		return nil, geom.Box{}, 0
	}
	f := w.Filters[w.FilterOn]
	columns, _ := w.FilterColumns()
	for _, column := range columns[1:] {
		values, at := w.ValueRows(column)
		for i, value := range values {
			named, _ := FilterValueParts(at[i])
			if named.In(p) && ends(w.measured(), f.Shown[value], reads) > named.W {
				return []string{f.Shown[value]}, at[i], at[i].W
			}
		}
	}
	return nil, geom.Box{}, 0
}

// FilterMenuSays is what heads the menu of filters.
const FilterMenuSays = "Filter by"

// FilterFoot is the keys the foot of the open menu says, each with what it does.
func (w Window) FilterFoot() [][2]string {
	if !w.FilterOpen {
		foot := [][2]string{{"↑↓", "walk"}, {"→", "open"}, {"⌫", "take off"}, {"⌥⌫", "all"}}
		if len(w.Snap.Views) > 0 {
			foot = append(foot, [2]string{w.keyboard.Spelled("⌘") + "1–9", "view"})
		}
		if w.Keepable() {
			foot = append(foot, [2]string{"v", "keep"})
		}
		return append(foot, [2]string{"esc", "close"})
	}
	if w.Filters[w.FilterOn].Many {
		return [][2]string{{"↑↓", "walk"}, {"←", "back"}, {"space", "tick"}, {"⏎", "take"}, {"esc", "close"}}
	}
	return [][2]string{{"↑↓", "walk"}, {"←→", "back, open"}, {"⏎", "take"}, {"esc", "close"}}
}

func (w Window) footWide() int {
	wide := 2 * padding
	for _, key := range w.FilterFoot() {
		wide += KeyChipWide(key[0]) + hintSpace + ends(w.measured(), key[1], small) + padding
	}
	return wide
}

// FilterColumn is one column of the menu of filters: which it is, from 0 for the filters themselves, where it is drawn, and the places among the open filter's values of those it lists.
type FilterColumn struct {
	At     int
	Box    geom.Box
	Values []int
}

const filterStubWide = 28

// FilterColumns is the columns of the menu of filters in sight, left to right, the filters first and each value opened at the right of the one it opened from; and the stub standing at their left for the columns slid out of sight when the window has no room for them all.
func (w Window) FilterColumns() ([]FilterColumn, geom.Box) {
	anchor := w.filterAnchor()
	var values [][]int
	if w.FilterOpen {
		values = w.ValueColumns()
	}
	rows := w.ViewsAbove() + len(w.Filters)
	for _, column := range values {
		rows = max(rows, len(column))
	}
	fixed := dropHead + dropFoot + inset
	most := max(1, (w.height-geom.Margin-anchor.Y-anchor.H-between-fixed)/menuRow)
	high := fixed + min(rows, most)*menuRow
	all, shown, stub := 1+len(values), 1+len(values), 0
	if room := min(w.width-2*geom.Margin, anchor.X+anchor.W-geom.Margin); all*filterColumnWide > room {
		stub = filterStubWide
		shown = max(1, (room-stub)/filterColumnWide)
	}
	wide := stub + shown*filterColumnWide
	x := max(geom.Margin, min(anchor.X+anchor.W-wide, w.width-geom.Margin-wide))
	y := anchor.Y + anchor.H + between
	var out []FilterColumn
	for c := all - shown; c < all; c++ {
		column := FilterColumn{At: c, Box: geom.Box{X: x + stub + (c-all+shown)*filterColumnWide, Y: y, W: filterColumnWide, H: high}}
		if c > 0 {
			column.Values = values[c-1]
		}
		out = append(out, column)
	}
	if stub == 0 {
		return out, geom.Box{}
	}
	return out, geom.Box{X: x, Y: y, W: stub, H: high}
}

// FiltersBox is the menu dropped under the pill it opened from, every column of it in sight: the filters, a line each with the key that opens it, then the values of the one open, each with how many tasks it would keep; its keys at its foot.
func (w Window) FiltersBox() geom.Box {
	columns, stub := w.FilterColumns()
	last := columns[len(columns)-1].Box
	left := columns[0].Box.X
	if stub.W > 0 {
		left = stub.X
	}
	return geom.Box{X: left, Y: last.Y, W: last.X + last.W - left, H: last.H}
}

func (w Window) filtersColumn() geom.Box {
	columns, _ := w.FilterColumns()
	if columns[0].At != 0 {
		return geom.Box{}
	}
	return columns[0].Box
}

// ValueRows is the values a column of the menu lists in sight, and where each is drawn: scrolled to the one the keyboard is on when it is in that column.
func (w Window) ValueRows(column FilterColumn) ([]int, []geom.Box) {
	room := dropRoom(column.Box)
	from := 0
	if at := slices.Index(column.Values, w.FilterPick); at >= 0 {
		from = max(0, min(at-room+1, len(column.Values)-room))
	}
	var values []int
	var boxes []geom.Box
	for i := from; i < len(column.Values) && i < from+room; i++ {
		values = append(values, column.Values[i])
		boxes = append(boxes, DropRowAt(column.Box, i-from))
	}
	return values, boxes
}

func (w Window) droppedUnder(anchor geom.Box, wide, rows int) geom.Box {
	fixed := dropHead + dropFoot + inset
	most := max(1, (w.height-geom.Margin-anchor.Y-anchor.H-between-fixed)/menuRow)
	high := fixed + min(rows, most)*menuRow
	wide = min(w.width-2*geom.Margin, wide)
	x := max(geom.Margin, min(anchor.X+anchor.W-wide, w.width-geom.Margin-wide))
	return geom.Box{X: x, Y: anchor.Y + anchor.H + between, W: wide, H: high}
}

// DropRowAt is where the i-th line of a dropped menu is, under its head.
func DropRowAt(box geom.Box, i int) geom.Box {
	return geom.Box{X: box.X + inset, Y: box.Y + dropHead + i*menuRow, W: box.W - 2*inset, H: menuRow}
}

func dropRoom(box geom.Box) int { return max(1, (box.H-dropHead-dropFoot-inset)/menuRow) }

func (w Window) filterAnchor() geom.Box {
	for _, pill := range w.Pills() {
		if w.Screen == board.OnKeeping && w.OfView && pill.Kind == PillView {
			return pill.Box
		}
	}
	for _, pill := range w.Pills() {
		if w.FilterFromPill && pill.Kind == PillFilter && pill.At == w.FilterOn {
			return pill.Box
		}
	}
	for _, pill := range w.Pills() {
		if pill.Kind == PillAdd || pill.Kind == PillFolded {
			return pill.Box
		}
	}
	l := w.Layout()
	return geom.Box{X: l.Tabs.X, Y: l.Tabs.Y, W: 0, H: l.Tabs.H}
}

// FilterRowAt is where the i-th filter's line is in the menu of filters, nowhere when the column of the filters has slid out of sight.
func (w Window) FilterRowAt(i int) geom.Box {
	column := w.filtersColumn()
	if column.W == 0 {
		return geom.Box{}
	}
	return DropRowAt(column, w.ViewsAbove()+i)
}

// FilterList is the values of the filter whose list is open, where each is drawn in the furthest column listing it, and whether one is open.
func (w Window) FilterList() ([]geom.Box, bool) {
	if !w.FilterOpen || w.Screen != board.OnFiltering {
		return nil, false
	}
	boxes := make([]geom.Box, len(w.Filters[w.FilterOn].Shown))
	columns, _ := w.FilterColumns()
	for _, column := range columns[1:] {
		values, at := w.ValueRows(column)
		for i, value := range values {
			boxes[value] = at[i]
		}
	}
	return boxes, true
}

func (w Window) clickingFilters(p geom.Point) Window {
	for _, pill := range w.Pills() {
		if pill.Box.In(p) {
			return w.clickingPill(pill, p)
		}
	}
	columns, stub := w.FilterColumns()
	if stub.In(p) {
		return w.goingBack()
	}
	for _, column := range columns {
		if !column.Box.In(p) {
			continue
		}
		if column.At > 0 {
			return w.clickingValue(column, p)
		}
		if viewed, on := w.clickingViews(p); on {
			return viewed
		}
		for i := range w.Filters {
			if w.FilterRowAt(i).In(p) {
				return w.openingFilter(i)
			}
		}
		return w
	}
	if w.FiltersBox().In(p) {
		return w
	}
	return w.closingMenu()
}

func (w Window) clickingValue(column FilterColumn, p geom.Point) Window {
	values, at := w.ValueRows(column)
	for i, value := range values {
		if !at[i].In(p) {
			continue
		}
		w.FilterTrail, w.FilterPick = slices.Clone(w.FilterTrail[:column.At-1]), value
		if w.OpensAColumn(value, column.At-1) {
			return w.openingDeeper()
		}
		return w.takingValue(value, !w.Filters[w.FilterOn].Many)
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
	// PillView is the view in force, standing for every filter it holds, its cross leaving it.
	PillView
	// PillState is the state the series' tab is narrowed to, walked by a click, its cross listing them all again.
	PillState
)

// FilterPill is one pill of the row at the right of the tabs: where it is, what it says and the name it says it under, what it stands for, for a filter's which one, whether it is faint for meaning nothing on the page in sight, for a view whether a filter or a search stands on top of it, and what it says faint after its value.
type FilterPill struct {
	Box     geom.Box
	Said    string
	Name    string
	Kind    PillKind
	At      int
	Faint   bool
	Serves  bool
	Changed bool
	Tail    string
}

// Cross is where a pill that clears what it stands for draws the cross that does it.
func (p FilterPill) Cross() (geom.Box, bool) {
	if p.Kind != PillSearch && p.Kind != PillFilter && p.Kind != PillView && p.Kind != PillState {
		return geom.Box{}, false
	}
	return geom.Box{X: p.Box.X + p.Box.W - pillCross, Y: p.Box.Y, W: pillCross, H: p.Box.H}, true
}

// Pills is the row at the right of the tabs: the search kept, then every filter in force, the pill clearing them, and against the menu the pill adding one, which stands where it stood whatever is added; folded into one pill, in that pill's place, when they no longer fit between the tabs and the menu.
func (w Window) Pills() []FilterPill {
	if w.Screen != board.OnTheBoard && w.Screen != board.OnFiltering && w.Screen != board.OnSearch && w.Screen != board.OnKeeping {
		return nil
	}
	tab := w.TabsAt()[0]
	y, high := tab.Y+(tab.H-searchPillHigh)/2, searchPillHigh
	start, end := w.pillsRun()
	if bar := w.Narrowing(); bar.Box.W > 0 {
		end = bar.Box.X - between
	}
	m := w.measured()
	var row []FilterPill
	crossed := func(name, value string, kind PillKind, at int) {
		wide := ends(m, name, small) + nameGap + ends(m, value, small) + 2*pillPad + pillCross
		row = append(row, FilterPill{Box: geom.Box{W: wide}, Said: name + " " + value, Name: name, Kind: kind, At: at})
	}
	plain := func(said string, kind PillKind) {
		row = append(row, FilterPill{Box: geom.Box{W: ends(m, said, small) + 2*pillPad}, Said: said, Kind: kind})
	}
	taken, viewed := 0, 0
	if w.Tab == board.Series {
		plain(AddFilterSays, PillAdd)
		crossed(StateSays, w.SeriesState(), PillState, -1)
		return placedPills(row, start, end, y, high)
	}
	if v, in := w.InView(); in {
		crossed("view", v.Name, PillView, -1)
		row[0].Changed = w.ViewChanged()
		if row[0].Changed {
			row[0].Box.W += dotSide + between
		}
		viewed = 1
	}
	if !w.Looking.Empty() && w.Screen != board.OnSearch {
		crossed("/", w.Looking.Reads(), PillSearch, -1)
		taken++
	}
	bar := w.Narrowing()
	narrowed := bar.Box.W > 0
	for i, f := range w.Filters {
		if len(f.Taken()) > 0 && narrowed && f.Arg == "objective" {
			root, _ := w.Snap.Goal(bar.Root)
			crossed("", "#"+strconv.Itoa(root.Number)+" "+root.Title, PillFilter, i)
			pill := &row[len(row)-1]
			pill.Said, pill.Serves, pill.Box.W = strings.TrimSpace(pill.Said), true, pill.Box.W+servesSide+servesGap
			if tailed := pill.Box.W + nameGap + ends(m, WithItsSubsSays, small); widthOf(row)-pill.Box.W+tailed <= end-start {
				pill.Said, pill.Tail, pill.Box.W = pill.Said+" "+WithItsSubsSays, WithItsSubsSays, tailed
			}
			taken++
			continue
		}
		if len(f.Taken()) > 0 {
			crossed(f.Name+":", f.TakenSays(), PillFilter, i)
			row[len(row)-1].Faint = w.Tab == board.Objectives && !f.ReachesGoals()
			taken++
		}
	}
	if taken > 0 && !narrowed {
		plain(ClearSays, PillClear)
	}
	if !narrowed {
		plain(AddFilterSays, PillAdd)
	}
	if widthOf(row) > end-start && taken+viewed > 0 {
		row = nil
		if !narrowed {
			plain(ClearSays, PillClear)
		}
		plain(FoldedSays(taken+viewed), PillFolded)
	}
	return placedPills(row, start, end, y, high)
}

// StateSays is what the pill narrowing the series' tab to a state is named.
const StateSays = "state:"

func placedPills(row []FilterPill, start, end, y, high int) []FilterPill {
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

func (w Window) untailedAt(p geom.Point) ([]string, geom.Box, int) {
	for _, pill := range w.Pills() {
		if pill.Serves && pill.Tail == "" && pill.Box.In(p) {
			return []string{pill.Said + ", " + WithItsSubsSays}, pill.Box, pill.Box.W
		}
	}
	return nil, geom.Box{}, 0
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

func (w Window) takingOffLast() Window {
	for i := len(w.Filters) - 1; i >= 0; i-- {
		if len(w.Filters[i].Taken()) > 0 {
			return w.takingOff(i)
		}
	}
	w.Looking = board.Written{}
	return w
}

func (w Window) takingOff(at int) Window {
	cleared := slices.Clone(w.Filters)
	cleared[at].At, cleared[at].Ticked = 0, nil
	w.Filters = cleared
	return w
}

func (w Window) clickingPill(pill FilterPill, p geom.Point) Window {
	if cross, ok := pill.Cross(); ok && cross.In(p) {
		switch pill.Kind {
		case PillSearch:
			w.Looking = board.Written{}
			return w
		case PillView:
			return w.leavingView()
		case PillState:
			w.SeriesIn = 0
			return w
		}
		return w.takingOff(pill.At)
	}
	switch {
	case pill.Kind == PillState, pill.Kind == PillAdd && w.Tab == board.Series:
		return w.walkingSeriesState()
	}
	switch pill.Kind {
	case PillView:
		return w.openingViewPopover()
	case PillFilter:
		return w.openingFilterFromPill(pill.At)
	case PillAdd, PillFolded:
		return w.filteringBoard()
	case PillClear:
		w.Filters, w.Looking = w.Filters.Cleared(), board.Written{}
		w.Screen = board.OnTheBoard
		return w
	case PillSearch:
		w.Screen, w.SearchFrom = board.OnSearch, board.OnTheBoard
		return w
	}
	return w
}

func (w Window) clickingPills(p geom.Point) (Window, bool) {
	if w.Screen != board.OnTheBoard {
		return w, false
	}
	if narrowed, ok := w.clickingNarrowing(p); ok {
		return narrowed, true
	}
	for _, pill := range w.Pills() {
		if pill.Box.In(p) {
			return w.clickingPill(pill, p), true
		}
	}
	return w, false
}
