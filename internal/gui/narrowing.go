package gui

import (
	"strconv"
	"strings"

	"cairn/internal/engine"
)

// Narrowing is the bar beside the objective's pill when the board shows an objective holding sub-objectives: all of it, each sub-objective, or the objective itself, each counted, the one in force lit.
type Narrowing struct {
	Root     int
	Segments []engine.Segment
	Values   []string
	On       int
	Box      Box
	At       []Box
}

const narrowingLeast = 160

// Narrowing is the bar the board draws beside its pills, and nothing when the objective filter is on no objective holding sub-objectives, or when the bar has no room.
func (w Window) Narrowing() Narrowing {
	if w.tab != Tasks || (w.screen != OnTheBoard && w.screen != OnFiltering && w.screen != OnSearch && w.screen != OnKeeping) {
		return Narrowing{}
	}
	value, ok := w.objectiveTaken()
	if !ok {
		return Narrowing{}
	}
	root, _ := strconv.Atoi(strings.TrimSuffix(value, " itself"))
	if g, found := w.snap.Goal(root); found && len(g.Children) == 0 && g.Parent != 0 {
		root = g.Parent
	}
	n := Narrowing{Root: root, Segments: w.snap.Segments(root, engine.OnTheBoard(w.snap.Cards, w.archived))}
	if len(n.Segments) == 0 {
		return Narrowing{}
	}
	g, _ := w.snap.Goal(root)
	n.Values = append(n.Values, strconv.Itoa(root))
	for _, c := range g.Children {
		n.Values = append(n.Values, strconv.Itoa(c))
	}
	n.Values = append(n.Values, strconv.Itoa(root)+" itself")
	for i, v := range n.Values {
		if v == value {
			n.On = i
		}
	}
	start, end := w.pillsRun()
	m := w.measured()
	wide := 0
	for _, s := range n.Segments {
		wide += narrowingWide(m, s)
	}
	if end-wide-narrowingLeast < start {
		return Narrowing{}
	}
	tab := TabAt(w.Layout(), 0)
	n.Box = Box{X: end - wide, Y: tab.Y + (tab.H-searchPillHigh)/2, W: wide, H: searchPillHigh}
	x := n.Box.X
	for _, s := range n.Segments {
		seg := narrowingWide(m, s)
		n.At = append(n.At, Box{X: x, Y: n.Box.Y, W: seg, H: n.Box.H})
		x += seg
	}
	return n
}

func narrowingWide(m Measure, s engine.Segment) int {
	return ends(m, s.Name, small) + countGap + monoWide(strconv.Itoa(s.Count), small) + 2*segmentPad
}

const segmentPad = 9

func (w Window) objectiveTaken() (string, bool) {
	for _, f := range w.filters {
		if taken := f.Taken(); f.Arg == "objective" && len(taken) == 1 {
			return taken[0], true
		}
	}
	return "", false
}

func (w Window) pillsRun() (start, end int) {
	l := w.Layout()
	start = TabAt(l, len(w.Shown())).X + between
	end = l.Tabs.X + l.Tabs.W
	if !w.native {
		end = MenuButton(l).X - between
	}
	return start, end
}

func (w Window) narrowedTo(at int) Window {
	n := w.Narrowing()
	if at < 0 || at >= len(n.Values) {
		return w
	}
	w.filters = w.filters.On("objective", n.Values[at])
	return w
}

func (w Window) clickingNarrowing(p Point) (Window, bool) {
	if w.screen != OnTheBoard {
		return w, false
	}
	for i, at := range w.Narrowing().At {
		if at.In(p) {
			return w.narrowedTo(i), true
		}
	}
	return w, false
}

func (w Window) walkingNarrowing(by int) Window {
	n := w.Narrowing()
	if len(n.Values) == 0 {
		return w
	}
	return w.narrowedTo((n.On + by + len(n.Values)) % len(n.Values))
}

func drawNarrowing(k Kit, w Window) {
	n := w.Narrowing()
	if len(n.At) == 0 {
		return
	}
	s := Segmented{At: n.At, On: n.On, Small: true}
	for _, seg := range n.Segments {
		s.Said = append(s.Said, seg.Name)
		s.Counts = append(s.Counts, strconv.Itoa(seg.Count))
	}
	s.Draw(k, n.Box)
}

// References are the tasks the board leaves out that block, wait on or are linked to one it shows, each where its column would have it after the cards it holds: shown while a view or the objective filter narrows the board, and never counted.
func (w Window) References() []Placed {
	if (w.viewing == "" && !w.filters.NarrowsBy("objective")) || w.tab != Tasks {
		return nil
	}
	shown := w.Showing().Cards
	refs := engine.Touching(shown, engine.OnTheBoard(w.snap.Cards, w.archived))
	if len(refs) == 0 {
		return nil
	}
	l := w.Layout()
	dealt, _ := engine.Deal(shown, w.archived)
	var cards []engine.Card
	says := map[int]string{}
	for _, r := range refs {
		cards, says[r.Card.Number] = append(cards, r.Card), r.Says()
	}
	theirs, _ := engine.Deal(cards, w.archived)
	var out []Placed
	for i, column := range theirs {
		for at, c := range column.Cards {
			out = append(out, Placed{Card: c, Slot: l.CardAt(i, len(dealt[i].Cards)+at, w.scroll[i]), Within: l.Column[i], Reference: true, Says: says[c.Number]})
		}
	}
	return out
}

// Why is what a reference says of itself: that the board leaves it out, and every way it touches what is shown.
func (p Placed) Why() string { return p.Says }
