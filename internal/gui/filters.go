package gui

import (
	"strings"

	"cairn/internal/engine"
)

// Filters are the board's filters, as the panel walks them.
func (w Window) Filters() engine.Filters { return w.filters }

// FilterOn is the line of the panel the keyboard is on.
func (w Window) FilterOn() int { return w.filterOn }

// FilteredBy is what the board is filtered by, written the way the bar says it: nothing when no filter narrows it.
func (w Window) FilteredBy() string { return w.filters.Said() }

// FilterSays is one line of the panel: what the filter is called, and what it is on.
func FilterSays(f engine.Filter) string {
	return f.Name + strings.Repeat(" ", max(1, filterNamed-len(f.Name))) + f.Says()
}

const filterNamed = 12

func (w Window) filteringBoard() Window {
	w.filters = engine.BoardFilters(w.snap, w.today()).Keeping(w.filters)
	w.screen = OnFiltering
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

func (w Window) pickingFilter(value int) Window {
	w.filters = w.filters.Walked(w.filterOn, value-w.filters[w.filterOn].At)
	w.filterOpen = false
	return w
}

// FilterList is the values of the filter whose list is open, where each is drawn, and whether one is open.
func (w Window) FilterList() ([]Box, bool) {
	if !w.filterOpen || w.screen != OnFiltering {
		return nil, false
	}
	row := ListRowAt(w.FiltersBox(), w.filterOn)
	f := w.filters[w.filterOn]
	wide := listLeast / 2
	for _, said := range f.Shown {
		wide = max(wide, ends(w.measured(), said, reads)+4*padding)
	}
	x := row.X + row.W - wide
	y := row.Y + row.H
	if y+len(f.Shown)*menuRow+2*inset > w.height-margin {
		y = max(margin, row.Y-len(f.Shown)*menuRow-2*inset)
	}
	boxes := make([]Box, len(f.Shown))
	for i := range f.Shown {
		boxes[i] = Box{X: x + inset, Y: y + inset + i*menuRow, W: wide - 2*inset, H: menuRow}
	}
	return boxes, true
}

func (w Window) clickingFilters(p Point) Window {
	if values, open := w.FilterList(); open {
		for i, value := range values {
			if value.In(p) {
				return w.pickingFilter(i)
			}
		}
		w.filterOpen = false
		return w
	}
	box := w.FiltersBox()
	for i := range w.filters {
		if ListRowAt(box, i).In(p) {
			return w.openingFilter(i)
		}
	}
	if !box.In(p) {
		w.screen = OnTheBoard
	}
	return w
}

// FiltersBox is the panel of the board's filters: across the middle of the window, a line a filter.
func (w Window) FiltersBox() Box {
	said := make([]string, 0, len(w.filters))
	for _, f := range w.filters {
		said = append(said, FilterSays(f))
	}
	return w.listBox(said)
}

// FilterPill is a filter in force as the pill it stands as at the right of the tabs: where, what it says, and which filter it is.
type FilterPill struct {
	Box  Box
	Said string
	At   int
}

// FilterPills is every filter in force as a pill, right to left from the search's pill, or from the end of the tabs when nothing is searched.
func (w Window) FilterPills() []FilterPill {
	if w.screen != OnTheBoard || w.tab == Objectives {
		return nil
	}
	l := w.Layout()
	tab := TabAt(l, 0)
	end := l.Tabs.X + l.Tabs.W
	if !w.native {
		end = MenuButton(l).X - between
	}
	if search, ok := w.SearchPill(); ok {
		end = search.X - pillGap
	}
	if count, ok := w.SearchCountPill(); ok {
		end = count.X - pillGap
	}
	var pills []FilterPill
	for i := len(w.filters) - 1; i >= 0; i-- {
		f := w.filters[i]
		if f.At == 0 {
			continue
		}
		said := f.Name + ": " + f.Says()
		wide := ends(w.measured(), said, small) + 2*pillPad + pillCross
		end -= wide
		pills = append([]FilterPill{{Box: Box{X: end, Y: tab.Y + (tab.H-searchPillHigh)/2, W: wide, H: searchPillHigh}, Said: said, At: i}}, pills...)
		end -= pillGap
	}
	return pills
}

func (w Window) clickingFilterPill(p Point) (Window, bool) {
	for _, pill := range w.FilterPills() {
		if !pill.Box.In(p) {
			continue
		}
		if p.X >= pill.Box.X+pill.Box.W-pillCross {
			kept := append(engine.Filters{}, w.filters...)
			kept[pill.At].At = 0
			w.filters = kept
		}
		return w, true
	}
	return w, false
}
