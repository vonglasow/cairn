package tui

import "cairn/internal/engine"

func (b Board) openingFilters() Board {
	b.filters = engine.BoardFilters(b.snap, engine.DayOf(b.clock())).Keeping(b.filters)
	b.filtering = true
	return b
}

func (b Board) filteringBy(key string) Board {
	switch key {
	case "esc", "enter":
		b.filtering = false
	case "backspace":
		b.filters = b.filters.Cleared()
	}
	way, ok := direction(key)
	if !ok {
		return b
	}
	switch way {
	case goLeft:
		b.filterOn = max(0, b.filterOn-1)
	case goRight:
		b.filterOn = min(len(b.filters)-1, b.filterOn+1)
	case goDown:
		b.filters = b.filters.Walked(b.filterOn, 1)
	case goUp:
		b.filters = b.filters.Walked(b.filterOn, -1)
	}
	return b
}

func (b Board) filtersLine() string {
	said := make([]string, 0, len(b.filters))
	for i, f := range b.filters {
		one := f.Name + " " + f.Says()
		if i == b.filterOn {
			said = append(said, b.skin.Key.Render(one))
			continue
		}
		said = append(said, b.skin.Says.Render(one))
	}
	return b.bar(append(said, b.offer("←→ ↑↓", "change"), b.offer("⌫", "clear all"), b.offer("esc", "close"))...)
}
