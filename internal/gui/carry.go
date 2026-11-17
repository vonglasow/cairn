package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"slices"
	"strconv"
)

type carried struct {
	on, lane  bool
	at, place int
}

func (w Window) carrying(across, down int) Window {
	if w.Tab != board.Tasks || w.Chosen == 0 {
		return w
	}
	if !w.carry.on {
		column, inBand, at, found := board.Where(w.Showing().Cards, w.Archived, w.Chosen)
		if !found {
			return w
		}
		w.held = Held{Number: w.Chosen, Moved: true}
		w.carry = carried{on: true, at: column}
		all := w.InColumn(column)
		if inBand {
			w.carry = carried{on: true, lane: true, at: slices.Index(board.Stalled, w.ChosenCard().State)}
			all, at = w.InLane(w.carry.at), slices.IndexFunc(w.InLane(w.carry.at), func(c board.Card) bool { return c.Number == w.Chosen })
		}
		w.carry.place = w.leftBefore(all, at)
	}
	w.carry = w.carry.stepped(w, across, down)
	return w.carriedTo()
}

func (c carried) stepped(w Window, across, down int) carried {
	if across != 0 {
		last := board.Columns - 1
		if c.lane {
			last = len(board.Stalled) - 1
		}
		if c.lane && c.at+across < 0 {
			c.lane = false
		}
		c.at = max(0, min(last, c.at+across))
		c.place = min(c.place, len(w.besideHeld(c.cards(w))))
		return c
	}
	c.place += down
	room := len(w.besideHeld(c.cards(w)))
	switch {
	case c.place > room && !c.lane && c.at >= 1:
		return carried{on: true, lane: true, at: c.at - 1}
	case c.place < 0 && c.lane:
		up := carried{on: true, at: c.at + 1}
		up.place = len(w.besideHeld(up.cards(w)))
		return up
	}
	c.place = max(0, min(room, c.place))
	return c
}

func (c carried) cards(w Window) []board.Card {
	if c.lane {
		return w.InLane(c.at)
	}
	return w.InColumn(c.at)
}

func (w Window) carriedTo() Window {
	all := w.carry.cards(w)
	kept := w.besideHeld(all)
	full := len(all)
	if w.carry.place < len(kept) {
		full = slices.IndexFunc(all, func(c board.Card) bool { return c.Number == kept[w.carry.place].Number })
	}
	l := w.Layout()
	var slot geom.Box
	if w.carry.lane {
		w = w.followingLane(w.carry.at, full)
		slot = l.LaneCardAt(w.carry.at, full, w.lanes[w.carry.at])
	} else {
		w = w.following(w.carry.at, full)
		slot = l.CardAt(w.carry.at, full, w.scroll[w.carry.at])
	}
	w.held.At = geom.Point{X: slot.X, Y: slot.Y}
	if own, ok := w.chosenSlot(); ok {
		w.held.From = geom.Point{X: own.X - lifted, Y: own.Y - lifted}
	}
	return w
}

const lifted = 16

// ShiftLetGo is ⇧ let go: the card the keyboard carries drops where it is carried to.
func (w Window) ShiftLetGo() Window {
	if !w.carry.on {
		return w
	}
	_, elsewhere := w.Landing()
	w.carry = carried{}
	if !elsewhere {
		w.held = Held{}
		return w
	}
	return w.dropped(w.held.Number, w.held.At)
}

func (w Window) whileCarrying(name string) Window {
	switch name {
	case "shift+left":
		return w.carrying(-1, 0)
	case "shift+right":
		return w.carrying(1, 0)
	case "shift+up":
		return w.carrying(0, -1)
	case "shift+down":
		return w.carrying(0, 1)
	case "esc":
		w.carry, w.held = carried{}, Held{}
	}
	return w
}

// CommandSays is what the command bar's line says at rest; while the keyboard carries a card, how to drop it or put it back.
func (w Window) CommandSays() string {
	if w.carry.on {
		return "Carrying #" + strconv.Itoa(w.held.Number) + " — let ⇧ go to drop it, esc to put it back"
	}
	return w.Board.CommandSays()
}
