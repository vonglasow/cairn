package gui

import (
	"cairn/internal/engine"
	"cairn/internal/trace"
	"fmt"
	"image/color"
	"math"
	"slices"
	"strings"
	"time"
)

// PanelOffers is where the offers of the panel that has the window are drawn, and so where they are clicked.
func (w Window) PanelOffers() []Box { return w.offersAlong(w.panel()) }

// ServingRows is how many tiles that box shows at once.
func (w Window) ServingRows() int {
	return max(1, BesideBox(w.formPanel()).H/SubHeight)
}

// Contrast is how far two colours stand apart, by the measure a person reading a screen is held to: 21 at most, 1 when they are the same.
func Contrast(a, b color.NRGBA) float64 {
	one, two := lit(a), lit(b)
	return (max(one, two) + 0.05) / (min(one, two) + 0.05)
}

func lit(c color.NRGBA) float64 {
	return 0.2126*straight(c.R) + 0.7152*straight(c.G) + 0.0722*straight(c.B)
}

func straight(v uint8) float64 {
	f := float64(v) / 255
	if v <= 10 {
		return f / 12.92
	}
	return math.Pow((f+0.055)/1.055, 2.4)
}

// ValueOf answers what a field holds, and nothing for a field that is not there.
func (f Form) ValueOf(name string) string {
	for _, field := range f.fields {
		if field.Name == name {
			return field.Value
		}
	}
	return ""
}

// Walk moves the keyboard from one field to the next, and stops at both ends rather than wrapping: a form that wraps is a form somebody falls off.
func (f Form) Walk(by int) Form {
	f.at = min(len(f.fields)-1, max(0, f.at+by))
	return f
}

// OnSystem is the window written for another system than the one it runs on: what a test asks for, so that what it reads does not depend on the machine.
func (w Window) OnSystem(system string) Window {
	w.keyboard = KeyboardOf(system)
	return w
}

// WhereChosen is where what is chosen is drawn and the run of the window it is drawn in, so that a test can ask whether one is inside the other.
func (w Window) WhereChosen() (box, run Box, ok bool) {
	l := w.Layout()
	if w.tab == Objectives {
		for i, g := range engine.Sorted(w.Showing().Goals) {
			if g.Number == w.chosen {
				return GoalAt(l, i, w.goals), GoalsRun(l), true
			}
		}
		return Box{}, Box{}, false
	}
	dealt, _ := engine.Deal(w.Showing().Cards, w.archived)
	for i, held := range dealt {
		for j, c := range held.Cards {
			if c.Number == w.chosen {
				return l.CardAt(i, j, w.scroll[i]), l.Column[i], true
			}
		}
	}
	for lane := range engine.Stalled {
		for j, c := range w.inLane(lane) {
			if c.Number == w.chosen {
				return l.LaneCardAt(lane, j, w.lanes[lane]), l.Lane[lane], true
			}
		}
	}
	return Box{}, Box{}, false
}

// Showing is how many whole cards a column has room for.
func (l Layout) Showing(column int) int { return l.Column[column].H / CardHeight }

// Written is the whole shape as a person can read it, so a test can pin it and a sum changed by accident shows up as a number that moved.
func (l Layout) Written() string {
	var out strings.Builder
	say := func(name string, b Box) {
		fmt.Fprintf(&out, "%-12s x=%-5d y=%-5d w=%-5d h=%d\n", name, b.X, b.Y, b.W, b.H)
	}
	say("tabs", l.Tabs)
	for i := range engine.Columns {
		say(fmt.Sprintf("head %d", i), l.Head[i])
		say(fmt.Sprintf("column %d", i), l.Column[i])
	}
	say("band", l.Band)
	for i, name := range engine.Stalled {
		say("lane "+name, l.Lane[i])
	}
	say("bar", l.Bar)
	return out.String()
}

// PartOn is the title of the part the keyboard is on, and whether it is on one.
func (w Window) PartOn() (string, bool) {
	parts := w.form.parts()
	if w.screen != OnForm || !w.form.onParts() || w.form.partAt >= len(parts) {
		return "", false
	}
	return parts[w.form.partAt].Title, true
}

// PartsGone is what is struck out, until the form is saved.
func (w Window) PartsGone() []engine.Part {
	return slices.DeleteFunc(w.form.parts(), func(part engine.Part) bool { return !w.form.isStruck(part) })
}

// Hovered is the offer the pointer is resting on.
func (w Window) Hovered() (string, bool) { return w.over, w.hovered }

// Serving is the work under the objective that is open, by title and by where each task stands.
func (w Window) Serving() []engine.Card {
	if !w.chosenGoal {
		return nil
	}
	return engine.Serving(w.snap.Cards, w.chosen)
}

// On puts the keyboard on a field by name, which is what a click does and what a test says.
func (w Window) On(name string) Window {
	for i, field := range w.form.Fields() {
		if field.Name == name {
			w.form = w.form.onto(i)
		}
	}
	return w
}

func anAct(did, field, became, through string) trace.Act {
	var told trace.AsTold
	told.On(1)
	told.Did(did)
	told.At(time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC))
	if field != "" {
		told.Changing(field)
		told.Became(became)
	} else {
		told.NothingNamed()
		told.BecameNothing()
	}
	if through != "" {
		told.Through(through)
	} else {
		told.ByHand()
	}
	return told.Act()
}

func narrow(line string, _ int) []int {
	edges := []int{0}
	for _, r := range line {
		wide := 9
		if r == 'i' {
			wide = 3
		}
		edges = append(edges, edges[len(edges)-1]+wide)
	}
	return edges
}
