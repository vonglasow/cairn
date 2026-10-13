package tui

import (
	"cairn/internal/engine"
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Drawing. Nothing here changes anything: a board is read, and a string comes
// out. That is what lets a scenario press a key and then read what is on the
// screen, with no terminal anywhere near it.

// View is what bubbletea asks for. It is one line over Draw, because what the
// board says has to be readable without a terminal: a scenario reads Draw.
//
// The board takes the whole terminal and gives it back on the way out, which is
// what the alternate screen is for: a board is a place you go to, and the
// scrollback you left behind should still be there when you come out.
func (b Board) View() tea.View {
	return tea.View{Content: b.Draw(), AltScreen: true}
}

// Draw is the whole screen, as text.
func (b Board) Draw() string {
	if TooSmall(b.width, b.height) {
		return b.tooSmall()
	}
	switch b.screen {
	case OnHelp:
		return b.drawHelp()
	case OnDetail:
		return b.drawDetail()
	case OnForm:
		return b.drawForm()
	case OnSaid:
		return b.drawSaid()
	}
	return b.drawBoard()
}

// tooSmall is the whole screen when the board will not fit. It says what is
// needed and what there is, because "too small" without a number leaves
// somebody dragging a window and guessing.
func (b Board) tooSmall() string {
	return b.skin.Title.Render(" cairn needs a bigger window.") + "\n" +
		b.skin.Faint.Render(fmt.Sprintf(" %d by %d at least. This one is %d by %d.",
			Narrowest, Shortest, b.width, b.height))
}

// drawBoard is the ordinary screen: the tabs, the columns, the band, the bar.
func (b Board) drawBoard() string {
	l := LayoutFor(b.width, b.height)
	var out []string
	out = append(out, b.tabs(), "")
	switch b.tab {
	case Objectives:
		out = append(out, b.objectivesList(l)...)
	case Time:
		out = append(out, b.timeWeek(l)...)
	default:
		out = append(out, b.tasksBoard(l)...)
	}
	return strings.Join(append(out, "", b.foot()), "\n")
}

// tabs is the top line: which tab is open, where inside it, and which mode.
func (b Board) tabs() string {
	tasks, goals, spent := b.skin.TabOff.Render("tasks"), b.skin.TabOff.Render("objectives"), b.skin.TabOff.Render("time")
	switch b.tab {
	case Tasks:
		tasks = b.skin.TabOn.Render("TASKS")
	case Objectives:
		goals = b.skin.TabOn.Render("OBJECTIVES")
	case Time:
		spent = b.skin.TabOn.Render("TIME")
	}
	left := " " + tasks + " " + goals + " " + spent
	if b.inside != 0 {
		left += "  " + b.skin.Crumb.Render("▸ "+strconv.Itoa(b.inside))
	}
	return Spread(left, b.skin.Mode.Render(b.modeName())+" ", b.width)
}

// modeName is what the mode is called on screen.
//
// Written as ifs rather than a tagless switch, and not for taste: a switch with
// no expression puts its conditions outside any covered block, so the mutation
// gate cannot judge them and a wrong one would go out unnoticed. The same trap
// is written up in test/harness.go.
func (b Board) modeName() string {
	if b.screen == OnHelp {
		return "HELP"
	}
	if b.screen == OnDetail {
		return "DETAIL"
	}
	if b.screen == OnForm {
		return "EDIT"
	}
	if b.prompt == Searching {
		return "SEARCH"
	}
	if b.prompt == Commanding {
		return "COMMAND"
	}
	if b.mode == Insert {
		return "INSERT"
	}
	return "NORMAL"
}

// tasksBoard is the four columns, the band under three of them, and the
// objective's own card above the lot when one is open.
func (b Board) tasksBoard(l Layout) []string {
	all, band := engine.Deal(b.cards(), b.archived)
	top := l.Top
	var out []string
	if b.inside != 0 {
		// The objective's card takes two rows out of the columns rather than
		// out of the band: what is stalled stays visible whatever else is on
		// screen.
		out = append(out, b.insideBand(), "")
		top = max(0, top-insideRows)
	}
	heads := make([]string, 0, len(all))
	rules := make([]string, 0, len(all))
	stacks := make([]string, 0, len(all))
	for i, column := range all {
		under := !b.inBand && b.mode == Normal && b.column == i
		heads = append(heads, b.header(column.Name, len(column.Cards), l.Column, under))
		rules = append(rules, b.rule(l.Column, under))
		stacks = append(stacks, b.stack(column.Cards, l, top, b.offsets[i]))
	}
	out = append(out,
		" "+strings.Join(heads, " "),
		" "+strings.Join(rules, " "),
		lipgloss.NewStyle().PaddingLeft(margin).Render(lipgloss.JoinHorizontal(lipgloss.Top, spaced(stacks)...)),
	)
	return append(out, b.bandRows(band, l)...)
}

// bandRows is the stalled band: its own header, and its cards three across.
func (b Board) bandRows(band []engine.Card, l Layout) []string {
	// The band starts where the second column starts: the left margin, the
	// whole of the backlog, and the gap after it.
	offset := margin + l.Column + 1
	under := strings.Repeat(" ", offset)
	held := b.inBand && b.mode == Normal
	head := Spread(b.skin.Band.Render(engine.Band), b.skin.Count.Render(strconv.Itoa(len(band))), l.Band)
	if held {
		head = Spread(b.skin.HeadOn.Render("▌"+engine.Band), b.skin.Count.Render(strconv.Itoa(len(band))), l.Band)
	}
	return []string{
		under + head,
		under + b.rule(l.Band, held),
		lipgloss.NewStyle().PaddingLeft(offset).Render(b.across(band, l)),
	}
}

// header is the line above a column: its name on the left, how many it holds on
// the right, and a bar in front of the name when the cursor stands on it.
func (b Board) header(name string, held, width int, under bool) string {
	if under {
		return Spread(b.skin.HeadOn.Render("▌"+name), b.skin.Count.Render(strconv.Itoa(held)), width)
	}
	return Spread(b.skin.Head.Render(name), b.skin.Count.Render(strconv.Itoa(held)), width)
}

// rule is the line under a header, drawn heavier where the cursor stands.
func (b Board) rule(width int, under bool) string {
	if under {
		return b.skin.RuleOn.Render(strings.Repeat("━", width))
	}
	return b.skin.Rule.Render(strings.Repeat("─", width))
}

// across lays the band's cards out three to a row, and scrolls when there are
// more of them than the band has room for.
func (b Board) across(cards []engine.Card, l Layout) string {
	return through(b.laidOut(cards, l), l.Band, l.Rows, b.bandAt)
}

// laidOut is the band's cards, three to a row, before anything is put behind a
// window. Separated so a test can count the rows it really made rather than the
// rows the window padded it out to.
func (b Board) laidOut(cards []engine.Card, l Layout) string {
	wide := engine.Columns - 1
	var rows []string
	for at := 0; at < len(cards); at += wide {
		end := min(at+wide, len(cards))
		drawn := make([]string, 0, wide)
		for _, c := range cards[at:end] {
			drawn = append(drawn, b.card(c, l))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, spaced(drawn)...))
	}
	return strings.Join(rows, "\n")
}

// stack draws one column's cards down the screen, scrolling to whatever the
// cursor is standing on.
func (b Board) stack(cards []engine.Card, l Layout, rows, at int) string {
	drawn := make([]string, 0, len(cards))
	for _, c := range cards {
		drawn = append(drawn, b.card(c, l))
	}
	return through(strings.Join(drawn, "\n"), l.Column, rows, at)
}

// card draws one, in the frame its priority and its focus decide.
func (b Board) card(c engine.Card, l Layout) string {
	return b.skin.Card(l.Column, c.Priority, c.InFocus, b.chosen(c)).
		Render(strings.Join(Rows(c, b.glyphs, b.skin, b.now(), l.Interior), "\n"))
}

// through puts a drawing behind a window of a given size, scrolled to a given
// row. It is what makes a column longer than the screen usable rather than
// merely truncated: what is past the bottom is reachable, it is just not on
// screen yet.
func through(content string, width, height, at int) string {
	if height < 1 {
		return ""
	}
	pane := viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
	pane.SetContent(content)
	pane.SetYOffset(at)
	return pane.View()
}

// chosen says whether a card is the one the cursor is holding.
func (b Board) chosen(c engine.Card) bool { return b.mode == Insert && c.Number == b.task }

// now is what the board calls the time. Taken from the outside so a scenario
// can say what day it is and a deadline reads the same every run.
func (b Board) now() time.Time { return b.clock() }

// spaced puts a single blank column between things drawn side by side.
func spaced(parts []string) []string {
	var out []string
	for i, part := range parts {
		if i > 0 {
			out = append(out, " ")
		}
		out = append(out, part)
	}
	return out
}

// insideBand is the objective's own card, flattened onto two rows above its
// board.
func (b Board) insideBand() string {
	for _, g := range b.snap.Goals {
		if g.Number != b.inside {
			continue
		}
		mark := lipgloss.NewStyle().Foreground(b.skin.palette.Standing(g.State)).Render(StandingGlyph(g.State, b.glyphs))
		head := Spread(" "+b.skin.Title.Render(strconv.Itoa(g.Number)+"   "+g.Title), mark+" ", b.width)
		return head + "\n " + b.progress(g.Done, g.Total, barWidth) + "  " + b.skin.Faint.Render(g.Counted())
	}
	return ""
}

// objectivesList is the second tab: one card per objective, down the screen,
// scrolling to whichever the cursor is on.
func (b Board) objectivesList(l Layout) []string {
	goals := b.goals()
	width := b.width - margin - margin - frame
	drawn := make([]string, 0, len(goals))
	for _, g := range goals {
		drawn = append(drawn, b.goalCard(g, width))
	}
	rows := l.Top + l.Rows + 2 // the band's header and rule are the columns' here
	return []string{
		" " + Spread(b.skin.Head.Render("OBJECTIVES"), b.skin.Count.Render(strconv.Itoa(len(goals))), b.width-2),
		" " + b.rule(b.width-2, false),
		lipgloss.NewStyle().PaddingLeft(margin).Render(
			through(strings.Join(drawn, "\n"), width+frame, rows, b.goalAt)),
	}
}

// goalCard draws one objective.
func (b Board) goalCard(g engine.Goal, width int) string {
	return b.skin.Card(width+frame, g.Priority, false, g.Number == b.goal).
		Render(strings.Join(b.goalRows(g, width), "\n"))
}

// foot is the bar at the bottom, and whatever is being typed above it. It only
// offers what works here: a key on the bar does something in this view, and a
// key that does nothing here is not on it.
func (b Board) foot() string {
	if b.err != nil {
		return " " + b.skin.Wrong.Render(b.err.Error())
	}
	if b.asking {
		return b.question()
	}
	if b.choosing {
		return b.kindPopup()
	}
	if b.prompt != NoPrompt {
		return b.promptLine()
	}
	return b.bar(b.offers()...)
}

// bar is the line at the foot of any screen: what can be pressed, cut to the
// width, dropping whole entries from the end rather than wrapping. A wrapped
// bar pushes everything above it up and out of the window.
func (b Board) bar(offers ...string) string {
	return " " + Join("   ", b.width-margin, offers...)
}

// offer writes one entry of the bar: the key, then what it does.
func (b Board) offer(key, does string) string {
	return b.skin.Key.Render(key) + " " + b.skin.Says.Render(does)
}

// offers is what the bar says, for this tab and this mode.
func (b Board) offers() []string {
	if b.tab == Objectives {
		return []string{
			b.offer("hjkl ←↓↑→", "objectives"), b.offer("⏎", "open"), b.offer("a", "new"),
			b.offer("e", "edit"), b.offer("d", "delete"), b.offer("/", "search"),
			b.offer("⇥", "tab"), b.offer("?", "help"), b.offer(":", "command"),
		}
	}
	if b.mode == Insert {
		return []string{
			b.offer("hjkl ←↓↑→", "tasks"), b.offer("⇧hjkl", "move"), b.offer("⏎", "open"),
			b.offer("e", "edit"), b.offer("f", "focus"), b.offer("d", "delete"),
			b.offer("x", "archive"), b.offer("?", "help"), b.offer("esc", "back"),
		}
	}
	out := []string{
		b.offer("hjkl ←↓↑→", "columns"), b.offer("i", "insert"),
	}
	if b.inside != 0 {
		out = append(out, b.offer("a", "new here"), b.offer("/", "search"),
			b.offer("?", "help"), b.offer("esc", "objectives"))
		return out
	}
	return append(out, b.offer("a", "new"), b.offer("/", "search"),
		b.offer("⇥", "tab"), b.offer("?", "help"), b.offer(":", "command"))
}

// promptLine is what a prompt looks like while it is taking what is typed.
func (b Board) promptLine() string {
	mark := "/"
	if b.prompt == Commanding {
		mark = ":"
	}
	return " " + b.skin.Prompt.Render(mark+b.typed) + b.skin.Key.Render("▏")
}

// question is the deletion waiting to be agreed to, said out loud: what would
// go, and the two keys that answer. Nothing undoes a deletion, so it is named
// by number rather than left to be guessed at.
func (b Board) question() string {
	if b.leaving {
		return " " + b.skin.Title.Render("throw away what you typed?") + "  " +
			b.bar(b.offer("⏎", "throw it away"), b.offer("esc", "back to the form"))
	}
	if b.sweeping {
		return " " + b.skin.Title.Render("archive every task in "+
			strings.ToUpper(engine.StateAt(b.column))+"?") + "  " +
			b.bar(b.offer("⏎", "archive them"), b.offer("esc", "keep them"))
	}
	what, number := "task", b.task
	if b.tab == Objectives {
		what, number = "objective", b.goal
	}
	asked := b.skin.Title.Render("delete " + what + " " + strconv.Itoa(number) + "?")
	return " " + asked + "  " + b.bar(b.offer("⏎", "delete"), b.offer("esc", "keep"))
}

// kindPopup asks which of the three kinds of stopped a task is going into.
func (b Board) kindPopup() string {
	out := make([]string, 0, len(kinds))
	for i, kind := range kinds {
		if i == b.choice {
			out = append(out, b.skin.Picked.Render(" "+kind+" "))
			continue
		}
		out = append(out, b.skin.Faint.Render(" "+kind+" "))
	}
	return " " + b.skin.Title.Render("which kind of stopped?") + "  " + strings.Join(out, " ")
}
