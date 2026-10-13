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

// View is what bubbletea asks for.
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

func (b Board) tooSmall() string {
	return b.skin.Title.Render(" cairn needs a bigger window.") + "\n" +
		b.skin.Faint.Render(fmt.Sprintf(" %d by %d at least. This one is %d by %d.",
			Narrowest, Shortest, b.width, b.height))
}

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

func (b Board) tasksBoard(l Layout) []string {
	all, band := engine.Deal(b.cards(), b.archived)
	top := l.Top
	var out []string
	if b.inside != 0 {
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

func (b Board) bandRows(band []engine.Card, l Layout) []string {
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

func (b Board) header(name string, held, width int, under bool) string {
	if under {
		return Spread(b.skin.HeadOn.Render("▌"+name), b.skin.Count.Render(strconv.Itoa(held)), width)
	}
	return Spread(b.skin.Head.Render(name), b.skin.Count.Render(strconv.Itoa(held)), width)
}

func (b Board) rule(width int, under bool) string {
	if under {
		return b.skin.RuleOn.Render(strings.Repeat("━", width))
	}
	return b.skin.Rule.Render(strings.Repeat("─", width))
}

func (b Board) across(cards []engine.Card, l Layout) string {
	return through(b.laidOut(cards, l), l.Band, l.Rows, b.bandAt)
}

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

func (b Board) stack(cards []engine.Card, l Layout, rows, at int) string {
	drawn := make([]string, 0, len(cards))
	for _, c := range cards {
		drawn = append(drawn, b.card(c, l))
	}
	return through(strings.Join(drawn, "\n"), l.Column, rows, at)
}

func (b Board) card(c engine.Card, l Layout) string {
	return b.skin.Card(l.Column, c.Priority, c.InFocus, b.chosen(c)).
		Render(strings.Join(Rows(c, b.glyphs, b.skin, b.now(), l.Interior), "\n"))
}

func through(content string, width, height, at int) string {
	if height < 1 {
		return ""
	}
	pane := viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
	pane.SetContent(content)
	pane.SetYOffset(at)
	return pane.View()
}

func (b Board) chosen(c engine.Card) bool { return b.mode == Insert && c.Number == b.task }

func (b Board) now() time.Time { return b.clock() }

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

func (b Board) objectivesList(l Layout) []string {
	goals := b.goals()
	width := b.width - margin - margin - frame
	drawn := make([]string, 0, len(goals))
	for _, g := range goals {
		drawn = append(drawn, b.goalCard(g, width))
	}
	rows := l.Top + l.Rows + 2
	return []string{
		" " + Spread(b.skin.Head.Render("OBJECTIVES"), b.skin.Count.Render(strconv.Itoa(len(goals))), b.width-2),
		" " + b.rule(b.width-2, false),
		lipgloss.NewStyle().PaddingLeft(margin).Render(
			through(strings.Join(drawn, "\n"), width+frame, rows, b.goalAt)),
	}
}

func (b Board) goalCard(g engine.Goal, width int) string {
	return b.skin.Card(width+frame, g.Priority, false, g.Number == b.goal).
		Render(strings.Join(b.goalRows(g, width), "\n"))
}

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

func (b Board) bar(offers ...string) string {
	return " " + Join("   ", b.width-margin, offers...)
}

func (b Board) offer(key, does string) string {
	return b.skin.Key.Render(key) + " " + b.skin.Says.Render(does)
}

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

func (b Board) promptLine() string {
	mark := "/"
	if b.prompt == Commanding {
		mark = ":"
	}
	return " " + b.skin.Prompt.Render(mark+b.typed) + b.skin.Key.Render("▏")
}

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

func (b Board) kindPopup() string {
	out := make([]string, 0, len(engine.Stalled))
	for i, kind := range engine.Stalled {
		if i == b.choice {
			out = append(out, b.skin.Picked.Render(" "+kind+" "))
			continue
		}
		out = append(out, b.skin.Faint.Render(" "+kind+" "))
	}
	return " " + b.skin.Title.Render("which kind of stopped?") + "  " + strings.Join(out, " ")
}
