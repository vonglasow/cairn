package tui

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/reading"
)

// The two screens that take the whole window without taking a form with them:
// the help, and a task's detail. Each hands the window back untouched, so what
// was on the board before is what is on it after.
//
// Both are built out of panels rather than out of headings and rules. A heading
// over a rule is what a terminal does when nobody has decided anything; a panel
// says where a thing begins and where it ends, and lets two of them sit side by
// side without either being read into the other.

// panel puts a titled box round something, of a given width. The title is
// dressed here when it is a plain word, and handed over already dressed when it
// is more than one thing — which is why it is not styled again.
func (b Board) panel(title, body string, width int) string {
	return b.skin.Panel.Width(width).Render(title + "\n" + body)
}

// named is a panel whose title is one word.
func (b Board) named(title, body string, width int) string {
	return b.panel(b.skin.Title.Render(title), body, width)
}

// drawHelp is the keys on the left and the legend on the right. It is
// contextual, like the bar it explains: the keys shown are the ones that work
// on the tab it was asked from, and a help listing keys that do nothing here is
// a help that has to be second-guessed.
func (b Board) drawHelp() string {
	half := (b.width - margin - margin - 3) / 2
	keys := b.named("KEYS", strings.Join(b.helpKeys(), "\n"), half)
	legend := b.named("LEGEND", strings.Join(b.legend(), "\n"), half)
	both := lipgloss.JoinHorizontal(lipgloss.Top, keys, " ", legend)
	// Behind a window like everything else: the legend is longer than a short
	// terminal, and help that pushed the way out of it off the bottom would be
	// help nobody could leave.
	//
	// The window is exactly as wide as the two panels it holds. Wider and it
	// would cut nothing, which means a width worked out wrong would never show.
	rows := max(1, b.height-4)
	return strings.Join([]string{
		b.tabs(), "",
		lipgloss.NewStyle().PaddingLeft(margin).Render(through(both, half*2+1, rows, b.helpAt)),
		"", b.bar(b.offer("↓↑", "read on"), b.offer("esc", "close")),
	}, "\n")
}

// group is a heading inside a panel, with its lines under it.
func (b Board) group(name string, lines ...string) []string {
	return append([]string{"", b.skin.Head.Render(name)}, lines...)
}

// says writes one key and what it does, lined up.
func (b Board) says(key, does string) string {
	return "  " + b.skin.Key.Render(pad(key, 12)) + b.skin.Says.Render(does)
}

// helpKeys is what can be pressed, grouped by mode.
func (b Board) helpKeys() []string {
	if b.tab == Objectives {
		return b.group("NORMAL — the cursor on an objective",
			b.says("hjkl ←↓↑→", "change objective"),
			b.says("gg / G", "first / last"),
			b.says("⏎", "open it"),
			b.says("a", "add an objective"),
			b.says("e", "the form"),
			b.says("d", "delete"),
			b.says("/", "search"),
			b.says("⇥", "next tab"),
			b.says(":", "command"),
			b.says("?", "help"),
		)
	}
	out := b.group("NORMAL — the cursor on a column",
		b.says("hjkl ←↓↑→", "columns, and the band"),
		b.says("i", "insert mode"),
		b.says("a", "add a task"),
		b.says("⇧A", "show what was archived"),
		b.says("/", "search"),
		b.says("⇥", "next tab"),
		b.says(":", "command"),
	)
	out = append(out, b.group("INSERT — the cursor on a task",
		b.says("hjkl ←↓↑→", "tasks, and the band"),
		b.says("gg / G", "first / last in the column"),
		b.says("⇧hjkl", "move it: sideways changes column, up and down change place"),
		b.says("⏎", "the detail"),
		b.says("e", "the form"),
		b.says("f", "take or let go the focus"),
		b.says("d", "delete"),
		b.says("x", "off the board, or back on"),
		b.says("esc", "back to normal"),
	)...)
	return append(out, b.group("EVERYWHERE", b.says("?", "help"))...)
}

// legend is what the marks on a card mean.
func (b Board) legend() []string {
	g, p := b.glyphs, b.skin.palette
	swatch := func(c any, name string) string {
		return "  " + lipgloss.NewStyle().Foreground(c.(interface {
			RGBA() (uint32, uint32, uint32, uint32)
		})).Render("▌") +
			" " + b.skin.Says.Render(name)
	}
	mark := func(glyph string, c any, name, means string) string {
		return "  " + lipgloss.NewStyle().Foreground(c.(interface {
			RGBA() (uint32, uint32, uint32, uint32)
		})).Render(glyph) +
			"  " + b.skin.Bright.Render(pad(name, 10)) + b.skin.Says.Render(means)
	}
	out := b.group("PRIORITY — the left edge of a card",
		swatch(p.Low, "low"), swatch(p.Normal, "normal"),
		swatch(p.High, "high"), swatch(p.Critical, "critical"),
	)
	out = append(out, b.group("STALLED — the top right of a card",
		mark(g.Blocked, p.Blocked, "blocked", "waiting on somebody else"),
		mark(g.Watching, p.Watching, "watching", "handed over, needs a look"),
		mark(g.Paused, p.Paused, "paused", "waiting on your decision"),
	)...)
	out = append(out, b.group("OBJECTIVES",
		mark(g.Active, p.Standing("active"), "active", "being worked towards"),
		mark(g.OnHold, p.Standing("on-hold"), "on-hold", "not for now"),
		mark(g.Reached, p.Standing("reached"), "reached", "done with"),
		mark(g.Dropped, p.Standing("dropped"), "dropped", "given up"),
	)...)
	return append(out, b.group("ON A CARD",
		mark(g.Described, p.Faint, "", "a description exists"),
		mark(g.SubTasks, p.Faint, "", "sub-tasks done / total"),
		mark(g.Label, p.Faint, "", "a label"),
		mark(g.InTheWay, p.Faint, "", "what is in the way"),
		"  "+b.skin.Bright.Render(pad("2d", 12))+b.skin.Says.Render("a deadline close by"),
		"  "+b.skin.Late.Render(pad("−3d", 12))+b.skin.Says.Render("one that has passed"),
		"  "+b.skin.Bright.Render(pad("╔═╗", 12))+b.skin.Says.Render("the task in focus"),
	)...)
}

// drawDetail is one task, whole: the short facts in a band, and the lists in
// panels underneath. A fact that is empty is not drawn — the opposite of a
// card, where an absence keeps its line to hold the height, because a full
// screen has no such constraint.
func (b Board) drawDetail() string {
	held, ok := b.held()
	if !ok {
		return strings.Join([]string{b.tabs(), "", " " + b.skin.Wrong.Render("there is no such task"),
			"", b.bar(b.offer("esc", "back"))}, "\n")
	}
	width := b.width - margin - margin
	left := (width * 3 / 5) - 1
	right := width - left - 1
	body := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, b.detailLeft(held, left)...),
		" ",
		b.named("HISTORY", strings.Join(b.history(held.Number, right-4, b.height-10), "\n"), right),
	)
	return strings.Join([]string{
		b.tabs(), "", b.detailBand(held), "",
		lipgloss.NewStyle().PaddingLeft(margin).Render(body),
		"", b.bar(b.offer("hjkl ←↓↑→", "navigate"), b.offer("e", "edit"),
			b.offer("f", "focus"), b.offer("?", "help"), b.offer("esc", "back")),
	}, "\n")
}

// detailLeft is what stacks down the left of the detail: what was written about
// the task, and what hangs under it.
func (b Board) detailLeft(c reading.Card, width int) []string {
	var out []string
	if c.Description != "" {
		out = append(out, b.named("DESCRIPTION", b.skin.Bright.Render(c.Description), width))
	}
	if rows := b.subTasks(c, width-4); len(rows) > 0 {
		out = append(out, b.named(fmt.Sprintf("SUB-TASKS   %d/%d", c.SubDone, c.SubTotal),
			strings.Join(rows, "\n"), width))
	}
	if len(c.Objectives) > 0 {
		out = append(out, b.named("SERVES", b.serves(c, width-4), width))
	}
	if len(out) == 0 {
		return []string{b.named("NOTHING WRITTEN DOWN",
			b.skin.Faint.Render("no description, no sub-tasks, nothing served"), width)}
	}
	return out
}

// detailBand is the short facts, on two rows above the panels.
func (b Board) detailBand(c reading.Card) string {
	head := " " + b.skin.Faint.Render(strconv.Itoa(c.Number)) + "   " + b.skin.Title.Render(c.Title)
	state := c.State
	if reading.Stopped(c.State) {
		state = lipgloss.NewStyle().Foreground(b.skin.palette.Stopped(c.State)).
			Render(StoppedGlyph(c.State, b.glyphs) + " " + c.State)
	}
	right := state + "  " + lipgloss.NewStyle().Foreground(b.skin.palette.Priority(c.Priority)).Render(c.Priority)
	var told []string
	for _, label := range c.Labels {
		told = append(told, b.skin.Says.Render(b.glyphs.Label+" "+label))
	}
	if day := domain.FormatDay(c.Deadline); day != "" {
		text, _, late := reading.Deadline(c.Deadline, b.now())
		style := b.skin.Bright
		if late {
			style = b.skin.Late
		}
		told = append(told, b.skin.Says.Render("due ")+style.Render(day+" · "+text))
	}
	if c.Estimate != "" {
		told = append(told, b.skin.Bright.Render(c.Estimate)+b.skin.Says.Render(" estimated"))
	}
	if c.Spent != "" {
		told = append(told, b.skin.Bright.Render(c.Spent)+b.skin.Says.Render(" spent"))
	}
	if way := inTheWay(c, b.glyphs); way != "" {
		told = append(told, b.skin.Wrong.Render(way))
	}
	first := Spread(head, right+" ", b.width)
	if len(told) == 0 {
		return first
	}
	return first + "\n " + strings.Join(told, "   ")
}

// subTasks is what hangs under a task, each with where it stands.
func (b Board) subTasks(c reading.Card, width int) []string {
	var out []string
	for _, other := range b.snap.Cards {
		if other.Parent != c.Number {
			continue
		}
		mark := b.skin.Faint.Render("·")
		if other.State == "done" {
			mark = b.skin.Filled.Render(b.glyphs.Reached)
		}
		out = append(out, Spread(mark+" "+b.skin.Faint.Render(strconv.Itoa(other.Number))+
			"  "+b.skin.Bright.Render(other.Title), b.skin.Says.Render(other.State), width))
	}
	return out
}

// serves is the objectives a task counts towards.
func (b Board) serves(c reading.Card, width int) string {
	var out []string
	for _, number := range c.Objectives {
		for _, g := range b.snap.Goals {
			if g.Number != number {
				continue
			}
			out = append(out, Spread(
				lipgloss.NewStyle().Foreground(b.skin.palette.Standing(g.State)).Render(StandingGlyph(g.State, b.glyphs))+
					" "+b.skin.Faint.Render(strconv.Itoa(g.Number))+"  "+b.skin.Bright.Render(g.Title),
				b.skin.Says.Render(fmt.Sprintf("%d/%d", g.Done, g.Total)), width))
		}
	}
	return strings.Join(out, "\n")
}

// history is every act performed on a task, newest last, behind a window of its
// own so a long life does not push everything else off the screen.
//
// This is the one place the trace is laid out beside the task it belongs to.
// `:trace show 14` prints the same history the way the command line prints it,
// and neither is the other's shortcut.
func (b Board) history(number, width, rows int) []string {
	shown, err := b.work.Ask(engine.TraceShow, reading.Numbered(number))
	if err != nil {
		return []string{b.skin.Wrong.Render(err.Error())}
	}
	drawn := make([]string, 0, len(shown.Acts))
	for _, act := range shown.Acts {
		drawn = append(drawn, Clip(b.line(act), width))
	}
	if rows < 1 {
		return drawn
	}
	pane := viewport.New(viewport.WithWidth(width), viewport.WithHeight(min(rows, len(drawn))))
	pane.SetContent(strings.Join(drawn, "\n"))
	// A history is read from its end: what happened last is what somebody came
	// to see. Asked for by name rather than worked out, so there is no sum to
	// get wrong.
	pane.GotoBottom()
	return strings.Split(pane.View(), "\n")
}

// pad puts a string in a fixed width, so a column of names lines up.
func pad(s string, width int) string {
	return s + strings.Repeat(" ", max(0, width-lipgloss.Width(s)))
}
