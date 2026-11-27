package gui

import (
	"fmt"
	"testing"

	"cairn/internal/board"
	"cairn/internal/engine/agent"
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
)

func columnOf(state string) int {
	for i := range board.Columns {
		if board.StateAt(i) == state {
			return i
		}
	}
	return -1
}

func manyIn(state string, from int) []board.Card {
	var cards []board.Card
	for i := range 40 {
		cards = append(cards, card(from+i, fmt.Sprintf("Task %d", from+i), state))
	}
	return cards
}

func TestEveryListThatOverflowsHasABarToPull(t *testing.T) {
	todo := columnOf("todo")
	l := New(newFake(), 1200, 800).Layout()
	tasks := holding(newFake(), append(manyIn("todo", 1), manyIn("blocked", 100)...)...)
	var goals []board.Goal
	for i := 1; i <= 40; i++ {
		goals = append(goals, board.Goal{Number: i, Title: fmt.Sprintf("Objective %d", i)})
	}
	objectives := holding(newFake())
	objectives.Snap.Goals, objectives.Tab = goals, board.Objectives
	var many []agent.Agent
	for i := range 30 {
		one := loose
		one.ID = fmt.Sprintf("s%02d", i)
		many = append(many, one)
	}
	sessions := agentsWindow(newFake(), pagesWith(true, true, false), many...)
	for _, c := range []struct {
		name string
		w    Window
		list geom.Box
		read func(Window) int
	}{
		{"a column", tasks, l.Column[todo], func(w Window) int { return w.scroll[todo] }},
		{"a lane", tasks, l.Lane[0], func(w Window) int { return w.lanes[0] }},
		{"the objectives", objectives, GoalsRun(l), func(w Window) int { return w.goals }},
		{"the agents", sessions, AgentsRun(l), func(w Window) int { return w.sessions.scroll }},
	} {
		t.Run(c.name, func(t *testing.T) {
			track, found := geom.Box{}, false
			for _, bar := range c.w.ListBars() {
				if bar.Track.X >= c.list.X && bar.Track.X+bar.Track.W <= c.list.X+c.list.W && bar.Track.Y >= c.list.Y && bar.Track.Y < c.list.Y+c.list.H {
					track, found = bar.Track, true
				}
			}
			if !found {
				t.Fatalf("no bar beside %s that runs past its room: %v", c.name, c.w.ListBars())
			}
			if painted(drawn(t, c.w), track, design.Dark().Colour.Line.Edge) == 0 {
				t.Errorf("the bar beside %s is not drawn", c.name)
			}
			top := geom.Point{X: track.X + track.W/2, Y: track.Y + 2}
			far := geom.Point{X: top.X, Y: track.Y + track.H + 999}
			pulled := c.w.PressAt(top).MoveTo(far).LetGo(far)
			if c.read(pulled) == c.read(c.w) {
				t.Errorf("pulling the bar beside %s left it at %d", c.name, c.read(c.w))
			}
			if pulled.Screen != c.w.Screen {
				t.Errorf("pulling the bar beside %s opened %v", c.name, pulled.Screen)
			}
		})
	}
}

func TestABarTakesItsRoomFromTheCardsOnlyWhenTheyOverflow(t *testing.T) {
	todo, current := columnOf("todo"), columnOf("current")
	w := holding(newFake(), append(manyIn("todo", 1), card(99, "Alone", "current"))...)
	l := w.Layout()
	for _, p := range w.Placed() {
		switch {
		case p.Card.State == "todo" && p.Slot.W != l.Column[todo].W-(barWide+between)/2:
			t.Fatalf("a card of a column that overflows is %d wide, want %d", p.Slot.W, l.Column[todo].W-(barWide+between)/2)
		case p.Card.State == "current" && p.Slot.W != l.Column[current].W:
			t.Fatalf("a card of a column that fits is %d wide, want %d", p.Slot.W, l.Column[current].W)
		}
	}
	for _, bar := range w.ListBars() {
		if bar.Track.X >= l.Column[current].X && bar.Track.X < l.Column[current].X+l.Column[current].W {
			t.Errorf("a column that fits has a bar: %v", bar.Track)
		}
	}
}
