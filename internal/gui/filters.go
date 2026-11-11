package gui

import (
	"slices"
	"strconv"
	"strings"

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
	w.screen, w.filterOpen, w.filterFromPill, w.viewWalked = OnFiltering, false, false, 0
	return w
}

// ViewWalked is the view of the filter menu the keyboard is on, from 1; 0 while it is on the filters.
func (w Window) ViewWalked() int { return w.viewWalked }

func (w Window) walkingTheMenu(by int) Window {
	views := len(w.snap.Views)
	if w.viewsAbove() == 0 {
		views = 0
	}
	switch {
	case w.viewWalked > 0 && by > 0 && w.viewWalked == views:
		w.viewWalked, w.filterOn = 0, 0
	case w.viewWalked > 0:
		w.viewWalked = max(1, w.viewWalked+by)
	case by < 0 && w.filterOn == 0 && views > 0:
		w.viewWalked = views
	default:
		w.filterOn = min(len(w.filters)-1, max(0, w.filterOn+by))
	}
	return w
}

func (w Window) onFilters(name string) Window {
	if w.filterOpen {
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
			return w.takingValue(w.filterPick, false)
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
		if w.viewWalked > 0 {
			return w.openingView(w.snap.Views[w.viewWalked-1].Place)
		}
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
	w.filterOn, w.filterOpen, w.filterPick, w.viewWalked, w.filterTrail = at, true, max(1, w.filters[at].At), 0, nil
	if first := w.valueColumns()[0]; len(first) > 0 && !slices.Contains(first, w.filterPick) {
		w.filterPick = first[0]
	}
	return w
}

func (w Window) valueColumns() [][]int {
	f := w.filters[w.filterOn]
	if f.Arg != "objective" {
		all := make([]int, 0, len(f.Values))
		for i := 1; i < len(f.Values); i++ {
			all = append(all, i)
		}
		return [][]int{all}
	}
	under := func(parent int) []int {
		var out []int
		for i, value := range f.Values {
			if g, found := w.goalOfValue(value); found && (g.Parent == parent || parent == 0 && !slices.Contains(f.Values, strconv.Itoa(g.Parent))) {
				out = append(out, i)
			}
		}
		return out
	}
	columns := [][]int{under(0)}
	for _, n := range w.filterTrail {
		own := []int{slices.Index(f.Values, strconv.Itoa(n))}
		if itself := slices.Index(f.Values, strconv.Itoa(n)+" itself"); itself >= 0 {
			own = append(own, itself)
		}
		columns = append(columns, append(own, under(n)...))
	}
	return columns
}

func (w Window) goalOfValue(value string) (engine.Goal, bool) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return engine.Goal{}, false
	}
	return w.snap.Goal(n)
}

func (w Window) holdsValue(value, want int) bool {
	f := w.filters[w.filterOn]
	g, found := w.goalOfValue(f.Values[value])
	if !found || len(g.Children) == 0 {
		return false
	}
	reached := w.snap.Reach(g.Number)
	n, _ := strings.CutSuffix(f.Values[want], " itself")
	number, err := strconv.Atoi(n)
	return err == nil && number != g.Number && slices.Contains(reached, number)
}

func (w Window) opensAColumn(value, column int) bool {
	f := w.filters[w.filterOn]
	g, found := w.goalOfValue(f.Values[value])
	if !found || len(g.Children) == 0 {
		return false
	}
	return column == 0 || w.filterTrail[column-1] != g.Number
}

func (w Window) goalOpened(value, column int) bool {
	g, found := w.goalOfValue(w.filters[w.filterOn].Values[value])
	return found && w.filterTrail[column] == g.Number
}

func (w Window) walkingValues(by int) Window {
	columns := w.valueColumns()
	last := columns[len(columns)-1]
	if len(last) == 0 {
		return w
	}
	at := max(0, slices.Index(last, w.filterPick))
	w.filterPick = last[min(len(last)-1, max(0, at+by))]
	return w
}

func (w Window) openingDeeper() Window {
	columns := w.valueColumns()
	if !w.opensAColumn(w.filterPick, len(columns)-1) {
		return w
	}
	g, _ := w.goalOfValue(w.filters[w.filterOn].Values[w.filterPick])
	w.filterTrail = append(slices.Clone(w.filterTrail), g.Number)
	return w
}

func (w Window) goingBack() Window {
	if len(w.filterTrail) == 0 {
		return w.closingFilter()
	}
	n := w.filterTrail[len(w.filterTrail)-1]
	w.filterTrail = w.filterTrail[:len(w.filterTrail)-1]
	w.filterPick = max(1, slices.Index(w.filters[w.filterOn].Values, strconv.Itoa(n)))
	return w
}

func (w Window) choosingValue() Window {
	if w.opensAColumn(w.filterPick, len(w.valueColumns())-1) {
		return w.openingDeeper()
	}
	return w.takingValue(w.filterPick, true)
}

func (w Window) takingValue(value int, closing bool) Window {
	f := w.filters[w.filterOn]
	if f.Many {
		w.filters = w.filters.Toggled(w.filterOn, f.Values[value])
	} else {
		if value == f.At {
			value = 0
		}
		w.filters = w.filters.Walked(w.filterOn, value-f.At)
	}
	if closing {
		return w.closingMenu()
	}
	return w
}

func (w Window) closingMenu() Window {
	w.screen, w.filterOpen, w.filterFromPill, w.filterTrail = OnTheBoard, false, false, nil
	return w
}

func (w Window) openingFilterFromPill(at int) Window {
	w = w.filteringBoard().openingFilter(at)
	w.filterFromPill = true
	return w
}

// closingFilter shuts the open filter's values: back to the menu of filters when they were opened from it, back to the board when from a pill.
func (w Window) closingFilter() Window {
	w.filterOpen, w.filterTrail = false, nil
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
	filterColumnWide = 232
	keepingWide      = 420
	filterCountWide  = 32
	dropHead         = 36
	dropFoot         = 36
)

// FilterValueParts is where a value of a filter writes its name and how many tasks it keeps, the name stopping short of the count.
func FilterValueParts(row Box) (Box, Box) {
	line := Inside(row)
	named := PastCheck(line)
	count := Box{X: line.X + line.W - filterCountWide, Y: line.Y, W: filterCountWide, H: line.H}
	named.W = count.X - between - named.X
	return named, count
}

func (w Window) cutFilterAt(p Point) ([]string, Box, int) {
	if !w.filterOpen || w.screen != OnFiltering {
		return nil, Box{}, 0
	}
	f := w.filters[w.filterOn]
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
	return nil, Box{}, 0
}

// FilterMenuSays is what heads the menu of filters.
const FilterMenuSays = "Filter by"

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
		foot := [][2]string{{"↑↓", "walk"}, {"→", "open"}, {"⌫", "take off"}, {"⌥⌫", "all"}}
		if len(w.snap.Views) > 0 {
			foot = append(foot, [2]string{w.keyboard.Spelled("⌘") + "1–9", "view"})
		}
		if w.keepable() {
			foot = append(foot, [2]string{"v", "keep"})
		}
		return append(foot, [2]string{"esc", "close"})
	}
	if w.filters[w.filterOn].Many {
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
	Box    Box
	Values []int
}

const filterStubWide = 28

// FilterColumns is the columns of the menu of filters in sight, left to right, the filters first and each value opened at the right of the one it opened from; and the stub standing at their left for the columns slid out of sight when the window has no room for them all.
func (w Window) FilterColumns() ([]FilterColumn, Box) {
	anchor := w.filterAnchor()
	var values [][]int
	if w.filterOpen {
		values = w.valueColumns()
	}
	rows := w.viewsAbove() + len(w.filters)
	for _, column := range values {
		rows = max(rows, len(column))
	}
	fixed := dropHead + dropFoot + inset
	most := max(1, (w.height-margin-anchor.Y-anchor.H-between-fixed)/menuRow)
	high := fixed + min(rows, most)*menuRow
	all, shown, stub := 1+len(values), 1+len(values), 0
	if room := min(w.width-2*margin, anchor.X+anchor.W-margin); all*filterColumnWide > room {
		stub = filterStubWide
		shown = max(1, (room-stub)/filterColumnWide)
	}
	wide := stub + shown*filterColumnWide
	x := max(margin, min(anchor.X+anchor.W-wide, w.width-margin-wide))
	y := anchor.Y + anchor.H + between
	var out []FilterColumn
	for c := all - shown; c < all; c++ {
		column := FilterColumn{At: c, Box: Box{X: x + stub + (c-all+shown)*filterColumnWide, Y: y, W: filterColumnWide, H: high}}
		if c > 0 {
			column.Values = values[c-1]
		}
		out = append(out, column)
	}
	if stub == 0 {
		return out, Box{}
	}
	return out, Box{X: x, Y: y, W: stub, H: high}
}

// FiltersBox is the menu dropped under the pill it opened from, every column of it in sight: the filters, a line each with the key that opens it, then the values of the one open, each with how many tasks it would keep; its keys at its foot.
func (w Window) FiltersBox() Box {
	columns, stub := w.FilterColumns()
	last := columns[len(columns)-1].Box
	left := columns[0].Box.X
	if stub.W > 0 {
		left = stub.X
	}
	return Box{X: left, Y: last.Y, W: last.X + last.W - left, H: last.H}
}

func (w Window) filtersColumn() Box {
	columns, _ := w.FilterColumns()
	if columns[0].At != 0 {
		return Box{}
	}
	return columns[0].Box
}

// ValueRows is the values a column of the menu lists in sight, and where each is drawn: scrolled to the one the keyboard is on when it is in that column.
func (w Window) ValueRows(column FilterColumn) ([]int, []Box) {
	room := dropRoom(column.Box)
	from := 0
	if at := slices.Index(column.Values, w.filterPick); at >= 0 {
		from = max(0, min(at-room+1, len(column.Values)-room))
	}
	var values []int
	var boxes []Box
	for i := from; i < len(column.Values) && i < from+room; i++ {
		values = append(values, column.Values[i])
		boxes = append(boxes, DropRowAt(column.Box, i-from))
	}
	return values, boxes
}

func (w Window) droppedUnder(anchor Box, wide, rows int) Box {
	fixed := dropHead + dropFoot + inset
	most := max(1, (w.height-margin-anchor.Y-anchor.H-between-fixed)/menuRow)
	high := fixed + min(rows, most)*menuRow
	wide = min(w.width-2*margin, wide)
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
		if w.screen == OnKeeping && w.ofView && pill.Kind == PillView {
			return pill.Box
		}
	}
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

// FilterRowAt is where the i-th filter's line is in the menu of filters, nowhere when the column of the filters has slid out of sight.
func (w Window) FilterRowAt(i int) Box {
	column := w.filtersColumn()
	if column.W == 0 {
		return Box{}
	}
	return DropRowAt(column, w.viewsAbove()+i)
}

// FilterList is the values of the filter whose list is open, where each is drawn in the furthest column listing it, and whether one is open.
func (w Window) FilterList() ([]Box, bool) {
	if !w.filterOpen || w.screen != OnFiltering {
		return nil, false
	}
	boxes := make([]Box, len(w.filters[w.filterOn].Shown))
	columns, _ := w.FilterColumns()
	for _, column := range columns[1:] {
		values, at := w.ValueRows(column)
		for i, value := range values {
			boxes[value] = at[i]
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
		for i := range w.filters {
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

func (w Window) clickingValue(column FilterColumn, p Point) Window {
	values, at := w.ValueRows(column)
	for i, value := range values {
		if !at[i].In(p) {
			continue
		}
		w.filterTrail, w.filterPick = slices.Clone(w.filterTrail[:column.At-1]), value
		if w.opensAColumn(value, column.At-1) {
			return w.openingDeeper()
		}
		return w.takingValue(value, !w.filters[w.filterOn].Many)
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
	Box     Box
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
func (p FilterPill) Cross() (Box, bool) {
	if p.Kind != PillSearch && p.Kind != PillFilter && p.Kind != PillView && p.Kind != PillState {
		return Box{}, false
	}
	return Box{X: p.Box.X + p.Box.W - pillCross, Y: p.Box.Y, W: pillCross, H: p.Box.H}, true
}

// Pills is the row at the right of the tabs: the search kept, then every filter in force, the pill clearing them, and against the menu the pill adding one, which stands where it stood whatever is added; folded into one pill, in that pill's place, when they no longer fit between the tabs and the menu.
func (w Window) Pills() []FilterPill {
	if w.screen != OnTheBoard && w.screen != OnFiltering && w.screen != OnSearch && w.screen != OnKeeping {
		return nil
	}
	l := w.Layout()
	tab := TabAt(l, 0)
	y, high := tab.Y+(tab.H-searchPillHigh)/2, searchPillHigh
	start, end := w.pillsRun()
	if bar := w.Narrowing(); bar.Box.W > 0 {
		end = bar.Box.X - between
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
	taken, viewed := 0, 0
	if w.tab == Series {
		plain(AddFilterSays, PillAdd)
		crossed(StateSays, w.SeriesState(), PillState, -1)
		return placedPills(row, start, end, y, high)
	}
	if v, in := w.inView(); in {
		crossed("view", v.Name, PillView, -1)
		row[0].Changed = w.ViewChanged()
		if row[0].Changed {
			row[0].Box.W += dotSide + between
		}
		viewed = 1
	}
	if !w.looking.Empty() && w.screen != OnSearch {
		crossed("/", w.looking.Reads(), PillSearch, -1)
		taken++
	}
	bar := w.Narrowing()
	narrowed := bar.Box.W > 0
	for i, f := range w.filters {
		if len(f.Taken()) > 0 && narrowed && f.Arg == "objective" {
			root, _ := w.snap.Goal(bar.Root)
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
			row[len(row)-1].Faint = w.tab == Objectives && !f.ReachesGoals()
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

func (w Window) untailedAt(p Point) ([]string, Box, int) {
	for _, pill := range w.Pills() {
		if pill.Serves && pill.Tail == "" && pill.Box.In(p) {
			return []string{pill.Said + ", " + WithItsSubsSays}, pill.Box, pill.Box.W
		}
	}
	return nil, Box{}, 0
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
	if page == Series && (w.seriesIn != 0 || !w.looking.Empty()) {
		return shown + "/" + strconv.Itoa(len(w.snap.Series))
	}
	if page == Series ||
		page == Time || (w.looking.Empty() && !w.filters.Narrows() && w.viewing == "") {
		return shown
	}
	if page == Objectives {
		if all := len(w.Reading().Goals); w.Holds(page) != all {
			return shown + "/" + strconv.Itoa(all)
		}
		return shown
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
		switch pill.Kind {
		case PillSearch:
			w.looking = Written{}
			return w
		case PillView:
			return w.leavingView()
		case PillState:
			w.seriesIn = 0
			return w
		}
		return w.takingOff(pill.At)
	}
	switch {
	case pill.Kind == PillState, pill.Kind == PillAdd && w.tab == Series:
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
