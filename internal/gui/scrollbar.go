package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
)

// ListBar is the bar beside a list that runs past its room: its track, and how far the list runs, shows and is scrolled, in pixels.
type ListBar struct {
	Track           geom.Box
	Lines, Rows, At int
	to              func(Window, int) Window
}

func (w Window) barred(l geom.Layout) geom.Layout {
	for i := range board.Columns {
		if len(w.InColumn(i))*geom.CardHeight > l.Column[i].H {
			l.Barred[i] = barRoom
		}
	}
	for lane := range board.Stalled {
		if len(w.InLane(lane))*geom.CardHeight > l.Lane[lane].H {
			l.LaneBarred[lane] = barRoom
		}
	}
	switch w.Tab {
	case board.Objectives:
		if len(w.Tree())*ItemHigh > GoalsRun(l).H {
			l.Listed = barRoom
		}
	case board.Agents:
		if laid := LaidAgents(l, w.AgentsGrouped(), 0, w.agentHigh); laid.High > laid.Shown.H {
			l.Listed = barRoom
		}
	}
	return l
}

const (
	listBarWide = barWide / 2
	barRoom     = (barWide + between) / 2
)

func beside(b geom.Box) geom.Box {
	return geom.Box{X: b.X + b.W - listBarWide, Y: b.Y, W: listBarWide, H: b.H}
}

// ListBars is the bars of the lists the page shows that run past their room.
func (w Window) ListBars() []ListBar {
	l := w.Layout()
	var bars []ListBar
	switch w.Tab {
	case board.Time, board.Series:
	case board.Objectives:
		if l.Listed > 0 {
			run := GoalsRun(l)
			bars = append(bars, ListBar{Track: beside(run), Lines: len(w.Tree()) * ItemHigh, Rows: run.H, At: w.goals,
				to: func(w Window, at int) Window { w.goals = w.settledGoals(at); return w }})
		}
	case board.Agents:
		if l.Listed > 0 {
			laid := w.AgentsLaid()
			track := beside(geom.Box{X: laid.Run.X, Y: laid.Shown.Y, W: laid.Run.W, H: laid.Shown.H})
			bars = append(bars, ListBar{Track: track, Lines: laid.High, Rows: laid.Shown.H, At: w.sessions.scroll,
				to: func(w Window, at int) Window { w.sessions.scroll = w.settledAgents(at); return w }})
		}
	default:
		for i := range board.Columns {
			if l.Barred[i] > 0 {
				bars = append(bars, ListBar{Track: beside(l.Column[i]), Lines: len(w.InColumn(i)) * geom.CardHeight, Rows: l.Column[i].H, At: w.scroll[i],
					to: func(w Window, at int) Window { w.scroll[i] = w.settled(i, at); return w }})
			}
		}
		for lane := range board.Stalled {
			if l.LaneBarred[lane] > 0 {
				bars = append(bars, ListBar{Track: beside(l.Lane[lane]), Lines: len(w.InLane(lane)) * geom.CardHeight, Rows: l.Lane[lane].H, At: w.lanes[lane],
					to: func(w Window, at int) Window { w.lanes[lane] = w.settledLane(lane, at); return w }})
			}
		}
	}
	return bars
}

func drawListBar(k Kit, w Window, track geom.Box) {
	for _, b := range w.ListBars() {
		if b.Track == track {
			Scrollbar{Lines: b.Lines, Rows: b.Rows, At: b.At}.Draw(k, track)
		}
	}
}

func (w Window) pullingTo(p geom.Point) Window {
	bars := w.ListBars()
	if w.pulling > len(bars) {
		return w
	}
	bar := bars[w.pulling-1]
	return bar.to(w, Reached(bar.Track, bar.Lines, bar.Rows, p))
}

func (w Window) pressingBar(p geom.Point) (Window, bool) {
	if w.Screen != board.OnTheBoard {
		return w, false
	}
	for i, bar := range w.ListBars() {
		if bar.Track.In(p) {
			w.pulling = i + 1
			return w.pullingTo(p), true
		}
	}
	return w, false
}
