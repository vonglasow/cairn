package tui

import (
	"errors"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/engine"
)

const viewsByFigure = 9

func (b Board) inView() (engine.Viewing, bool) {
	if b.viewing == "" {
		return engine.Viewing{}, false
	}
	return b.snap.View(b.viewing)
}

func (b Board) viewChanged() bool {
	_, in := b.inView()
	return in && (b.filters.Narrows() || b.query != "")
}

func (b Board) choosingAView(key string) Board {
	b.choosingView = false
	if key == "esc" {
		return b
	}
	place, err := strconv.Atoi(key)
	if err != nil || len(key) != 1 {
		b.choosingView = true
		return b
	}
	if place == 0 {
		return b.wholeBoard()
	}
	if place > len(b.snap.Views) {
		b.choosingView = true
		return b
	}
	return b.openingView(b.snap.Views[place-1].Name)
}

func (b Board) openingView(name string) Board {
	b.viewing, b.query = name, ""
	b.filters = b.filters.Cleared()
	b.column, b.inBand, b.inside, b.offsets = 0, false, 0, [engine.Columns]int{}
	return b
}

func (b Board) wholeBoard() Board {
	b.viewing, b.query = "", ""
	b.filters = b.filters.Cleared()
	return b
}

func (b Board) viewsLine() string {
	var offers []string
	for _, v := range b.snap.Views {
		if v.Place <= viewsByFigure {
			offers = append(offers, b.offer(strconv.Itoa(v.Place), v.Name+" "+strconv.Itoa(len(v.Kept))))
		}
	}
	return b.bar(append(offers, b.offer("0", "whole board"), b.offer("esc", "close"))...)
}

func (b Board) viewSays() string {
	v, in := b.inView()
	if !in {
		return ""
	}
	said := "view " + v.Name
	if b.viewChanged() {
		said += " *"
	}
	said += " · " + strconv.Itoa(len(engine.OnTheBoard(b.cards(), b.archived))) + " of " +
		strconv.Itoa(len(engine.OnTheBoard(b.snap.Cards, b.archived))) + " tasks"
	if gone := engine.GoneSays(v.View); gone != "" {
		said += " · " + gone
	}
	if b.filters.Narrows() {
		said += " · filtered by " + b.filters.Said()
	}
	return said
}

var errNoViewHeld = errors.New("no view is in force: v and a figure opens one")

func (b Board) viewCommand(words []string) (tea.Model, tea.Cmd, bool) {
	if len(words) < 2 || words[0] != "view" {
		return b, nil, false
	}
	switch {
	case words[1] == "save":
		v, in := b.inView()
		if !in {
			b.err = errNoViewHeld
			return b, nil, true
		}
		saved := b.filters.Saving(v.View, b.query)
		b.filters, b.query = b.filters.Cleared(), ""
		return b, b.telling(engine.ViewEdit, saved), true
	case words[1] == "new" && !strings.HasPrefix(strings.Join(words[2:], " "), "-") && !strings.Contains(strings.Join(words[2:], " "), " --"):
		v, _ := b.inView()
		name := strings.Join(words[2:], " ")
		if name == "" {
			name = b.filters.Proposed(b.snap, b.query)
		}
		in := b.filters.Holding(v.View, b.query)
		in.Given["name"] = name
		b = b.openingView(name)
		return b, b.telling(engine.ViewNew, in), true
	}
	if _, called := b.pages.Called(words[0], words[1]); called {
		return b, nil, false
	}
	name := strings.Join(words[1:], " ")
	v, found := b.snap.View(name)
	if !found {
		b.err = errors.New("no view is named " + name + ": view list names them")
		return b, nil, true
	}
	return b.openingView(v.Name), nil, true
}
