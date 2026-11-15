package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"slices"
	"strconv"
	"strings"
)

// Narrowing is the bar beside the objective's pill when the board shows an objective holding sub-objectives: all of it, each sub-objective, or the objective itself, each counted, the one in force lit; compact, the one in force alone between two chevrons walking the others, when the whole bar has no room.
type Narrowing struct {
	Root       int
	Segments   []board.Segment
	On         int
	Box        geom.Box
	At         []geom.Box
	Compact    bool
	Back, Next geom.Box
	Said       string
}

const narrowingLeast = 160

// WithItsSubsSays is what the pill of an objective shown with its sub-objectives says after its name.
const WithItsSubsSays = "with its sub-objectives"

// Narrowing is the bar the board draws beside its pills, and nothing when neither the objective filter nor the view in force is on one objective holding sub-objectives, or when not even its compact form has room.
func (w Window) Narrowing() Narrowing {
	if w.Tab != board.Tasks || (w.Screen != board.OnTheBoard && w.Screen != board.OnFiltering && w.Screen != board.OnSearch && w.Screen != board.OnKeeping) {
		return Narrowing{}
	}
	value, ok := w.NarrowingTaken()
	if !ok {
		return Narrowing{}
	}
	root, _ := strconv.Atoi(strings.TrimSuffix(value, " itself"))
	root = w.Snap.TreeOf(root)
	n := Narrowing{Root: root, Segments: w.Snap.Segments(root, w.Snap.Viewed(w.Viewing, board.Laid(w.Snap.Cards, w.Archived)))}
	if len(n.Segments) == 0 {
		return Narrowing{}
	}
	for i, s := range n.Segments {
		if s.Value == value {
			n.On = i
		}
	}
	start, end := w.pillsRun()
	m := w.measured()
	wide := 0
	for _, s := range n.Segments {
		wide += narrowingWide(m, s)
	}
	tab := TabAt(w.Layout(), 0)
	y := tab.Y + (tab.H-searchPillHigh)/2
	if end-wide-narrowingLeast >= start {
		n.Box = geom.Box{X: end - wide, Y: y, W: wide, H: searchPillHigh}
		x := n.Box.X
		for _, s := range n.Segments {
			seg := narrowingWide(m, s)
			n.At = append(n.At, geom.Box{X: x, Y: y, W: seg, H: searchPillHigh})
			x += seg
		}
		return n
	}
	if end-start-2*chevronRoom < narrowingLeast/2 {
		return Narrowing{}
	}
	seg := min(narrowingWide(m, n.Segments[n.On]), end-start-2*chevronRoom)
	wide = seg + 2*chevronRoom
	on := n.Segments[n.On]
	n.Compact, n.Said = true, on.Name
	if room := seg - narrowingWide(m, board.Segment{Count: on.Count}); ends(m, on.Name, small) > room {
		n.Said = cutToFit(on.Name, room, func(said string) int { return ends(m, said, small) })
	}
	n.Box = geom.Box{X: end - wide, Y: y, W: wide, H: searchPillHigh}
	n.Back = geom.Box{X: n.Box.X, Y: y, W: chevronRoom, H: searchPillHigh}
	n.At = make([]geom.Box, len(n.Segments))
	n.At[n.On] = geom.Box{X: n.Back.X + chevronRoom, Y: y, W: seg, H: searchPillHigh}
	n.Next = geom.Box{X: n.At[n.On].X + seg, Y: y, W: chevronRoom, H: searchPillHigh}
	return n
}

const chevronRoom = 22

func narrowingWide(m Measure, s board.Segment) int {
	return ends(m, s.Name, small) + countGap + monoWide(strconv.Itoa(s.Count), small) + 2*segmentPad
}

const segmentPad = 7

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
	if at < 0 || at >= len(n.Segments) {
		return w
	}
	value := n.Segments[at].Value
	if v, in := w.InView(); in && len(v.Holds.Objective) == 1 && strconv.Itoa(v.Holds.Objective[0]) == value {
		if i := slices.IndexFunc(w.Filters, func(f board.Filter) bool { return f.Arg == "objective" }); i >= 0 {
			return w.takingOff(i)
		}
		return w
	}
	w.Filters = board.FiltersOf(w.Snap, w.Today()).Keeping(w.Filters).On("objective", value)
	return w
}

func (w Window) clickingNarrowing(p geom.Point) (Window, bool) {
	if w.Screen != board.OnTheBoard {
		return w, false
	}
	n := w.Narrowing()
	switch {
	case n.Compact && n.Back.In(p):
		return w.walkingNarrowing(-1), true
	case n.Compact && n.Next.In(p):
		return w.walkingNarrowing(1), true
	}
	for i, at := range n.At {
		if at.In(p) {
			return w.narrowedTo(i), true
		}
	}
	return w, false
}

func (w Window) walkingNarrowing(by int) Window {
	n := w.Narrowing()
	if len(n.Segments) == 0 {
		return w
	}
	return w.narrowedTo((n.On + by + len(n.Segments)) % len(n.Segments))
}

func drawNarrowing(k Kit, w Window) {
	n := w.Narrowing()
	if len(n.At) == 0 {
		return
	}
	s := Segmented{At: n.At, On: n.On, Small: true}
	if n.Compact {
		s.At, s.On, s.Back, s.Next = []geom.Box{n.At[n.On]}, 0, n.Back, n.Next
		s.Said, s.Counts = []string{n.Said}, []string{strconv.Itoa(n.Segments[n.On].Count)}
		s.Draw(k, n.Box)
		return
	}
	for _, seg := range n.Segments {
		s.Said = append(s.Said, seg.Name)
		s.Counts = append(s.Counts, strconv.Itoa(seg.Count))
	}
	s.Draw(k, n.Box)
}

// References are the tasks the board leaves out that block, wait on or are linked to one it shows, each where its column would have it after the cards it holds: shown while a view or the objective filter narrows the board, and never counted.
func (w Window) References() []Placed {
	if (w.Viewing == "" && !w.Filters.NarrowsBy("objective")) || w.Tab != board.Tasks {
		return nil
	}
	shown := w.Showing().Cards
	refs := board.Touching(shown, board.Laid(w.Snap.Cards, w.Archived))
	if len(refs) == 0 {
		return nil
	}
	l := w.Layout()
	dealt, _ := board.Deal(shown, w.Archived)
	var cards []board.Card
	says := map[int]string{}
	for _, r := range refs {
		cards, says[r.Card.Number] = append(cards, r.Card), r.Says()
	}
	theirs, _ := board.Deal(cards, w.Archived)
	var out []Placed
	for i, column := range theirs {
		after := len(dealt[i].Cards)
		if w.EmptySays(i) != "" {
			after++
		}
		for at, c := range column.Cards {
			out = append(out, Placed{Card: c, Slot: l.CardAt(i, after+at, w.scroll[i]), Within: l.Column[i], Reference: true, Says: says[c.Number]})
		}
	}
	return out
}

// Why is what a reference says of itself: that the board leaves it out, and every way it touches what is shown.
func (p Placed) Why() string { return p.Says }
