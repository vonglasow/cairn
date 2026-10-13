package tui

import (
	"strconv"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/objective"
)

func (b Board) panel(title, body string, width int) string {
	return b.skin.Panel.Width(width).Render(title + "\n" + body)
}

func (b Board) named(title, body string, width int) string {
	return b.panel(b.skin.Title.Render(title), body, width)
}

func (b Board) drawHelp() string {
	half := (b.width - margin - margin - 3) / 2
	keys := b.named("KEYS", strings.Join(b.helpKeys(), "\n"), half)
	legend := b.named("LEGEND", strings.Join(b.legend(), "\n"), half)
	both := lipgloss.JoinHorizontal(lipgloss.Top, keys, " ", legend)
	rows := max(1, b.height-4)
	return strings.Join([]string{
		b.tabs(), "",
		lipgloss.NewStyle().PaddingLeft(margin).Render(through(both, half*2+1, rows, b.helpAt)),
		"", b.bar(b.offer("↓↑", "read on"), b.offer("esc", "close")),
	}, "\n")
}

func (b Board) group(name string, lines ...string) []string {
	return append([]string{"", b.skin.Head.Render(name)}, lines...)
}

func (b Board) says(key, does string) string {
	return "  " + b.skin.Key.Render(pad(key, 12)) + b.skin.Says.Render(does)
}

func (b Board) helpKeys() []string {
	if b.tab == Objectives {
		return b.group("NORMAL — the cursor on an objective",
			b.says("hjkl ←↓↑→", "change objective"),
			b.says("gg / G", "first / last"),
			b.says("⏎", "open it"),
			b.says("a", "a new objective"),
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
		b.says("a", "a new task"),
		b.says("⇧A", "show what was archived"),
		b.says("⇧X", "archive every task in this column — it asks first"),
		b.says("/", "search"),
		b.says("⇥", "next tab — tasks, objectives, the time spent"),
		b.says(":", "command — what it answers is said until esc"),
		b.says("b", "a new backup of the store"),
		b.says("⇧B", "the backups"),
		b.says("←→ t on time", "walk the weeks, back to today's"),
	)
	out = append(out, b.group("INSERT — the cursor on a task",
		b.says("hjkl ←↓↑→", "tasks, and the band"),
		b.says("gg / G", "first / last in the column"),
		b.says("⇧hjkl", "move it: sideways changes column, up and down change place"),
		b.says("⏎", "the detail"),
		b.says("e", "the form"),
		b.says("f", "focus or unfocus"),
		b.says("d", "delete"),
		b.says("x", "archive or unarchive"),
		b.says("esc", "back to normal"),
	)...)
	out = append(out, b.group("DETAIL — one task, open",
		b.says("^D ^U", "walk its description, half a box at a time"),
	)...)
	out = append(out, b.group("FORM — the same keys as the window",
		b.says("⏎", "a new line in the description, the next field anywhere else"),
		b.says("↑↓", "the description's lines — tab leaves it"),
		b.says("tab ⇧tab", "next / previous field"),
		b.says("^S", "save, from anywhere in it — ⏎ never does"),
	)...)
	return append(out, b.group("EVERYWHERE", b.says("?", "help"))...)
}

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
		mark(g.Active, p.Standing(string(objective.Active)), "active", "being worked towards"),
		mark(g.OnHold, p.Standing(string(objective.OnHold)), "on-hold", "not for now"),
		mark(g.Reached, p.Standing(string(objective.Reached)), "reached", "done with"),
		mark(g.Dropped, p.Standing(string(objective.Dropped)), "dropped", "given up"),
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

func (b Board) drawDetail() string {
	held, ok := b.held()
	if !ok {
		return strings.Join([]string{b.tabs(), "", " " + b.skin.Wrong.Render("there is no such task"),
			"", b.bar(b.offer("esc", "back"))}, "\n")
	}
	width := b.width - margin - margin
	left := b.detailWidth()
	right := width - left - 1
	body := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, b.detailLeft(held, left)...),
		" ",
		b.named("TRACE", strings.Join(b.history(held.Number, right-4, b.height-10), "\n"), right),
	)
	return strings.Join([]string{
		b.tabs(), "", b.detailBand(held), "",
		lipgloss.NewStyle().PaddingLeft(margin).Render(body),
		"", b.bar(b.detailOffers(held)...),
	}, "\n")
}

func (b Board) detailOffers(c engine.Card) []string {
	offers := []string{b.offer("hjkl ←↓↑→", "navigate"), b.offer("e", "edit"), b.offer("f", "focus")}
	if c.Description != "" {
		offers = append(offers, b.offer("^D ^U", "description"))
	}
	return append(offers, b.offer("?", "help"), b.offer("esc", "back"))
}

func (b Board) detailLeft(c engine.Card, width int) []string {
	under := b.detailLeftUnder(c, width)
	var out []string
	if c.Description != "" {
		lines, rows := b.saidOn(c, width, under)
		shown := InSight(lines, b.scrolledSaying(c, lines, rows), rows, width-describing)
		out = append(out, b.named("DESCRIPTION", b.skin.Bright.Render(strings.Join(shown, "\n")), width))
	}
	out = append(out, under...)
	if len(out) == 0 {
		return []string{b.named("NOTHING WRITTEN DOWN",
			b.skin.Faint.Render("no description, no sub-tasks, nothing served"), width)}
	}
	return out
}

const describing = 6

const aroundTheDetail = 7

func (b Board) saidOn(c engine.Card, width int, under []string) ([]string, int) {
	taken := 0
	for _, panel := range under {
		taken += lipgloss.Height(panel)
	}
	rows := max(3, b.height-aroundTheDetail-taken-3)
	return Paragraphs(c.Description, width-describing), rows
}

func (b Board) scrolledSaying(c engine.Card, lines []string, rows int) int {
	if b.sayingOf != c.Number {
		return 0
	}
	return min(max(0, len(lines)-rows), max(0, b.saying))
}

func (b Board) scrollingSaying(down bool) Board {
	c, ok := b.held()
	if !ok || c.Description == "" {
		return b
	}
	width := b.detailWidth()
	lines, rows := b.saidOn(c, width, b.detailLeftUnder(c, width))
	by := max(1, rows/2)
	if !down {
		by = -by
	}
	at := b.scrolledSaying(c, lines, rows)
	b.saying, b.sayingOf = min(max(0, len(lines)-rows), max(0, at+by)), c.Number
	return b
}

func (b Board) detailWidth() int { return (b.width-margin-margin)*3/5 - 1 }

func (b Board) detailLeftUnder(c engine.Card, width int) []string {
	var under []string
	if rows := b.subTasks(c, width-4); len(rows) > 0 {
		under = append(under, b.named(b.partsHead(c), strings.Join(rows, "\n"), width))
	}
	if len(c.Objectives) > 0 {
		under = append(under, b.named("SERVES", b.serves(c, width-4), width))
	}
	if c.Context != "" {
		under = append(under, b.named("CONTEXT", b.skin.Faint.Render(c.Context), width))
	}
	return under
}

func (b Board) detailBand(c engine.Card) string {
	head := " " + b.skin.Faint.Render(strconv.Itoa(c.Number)) + "   " + b.skin.Title.Render(c.Title)
	state := c.State
	if engine.Stopped(c.State) {
		state = lipgloss.NewStyle().Foreground(b.skin.palette.Stopped(c.State)).
			Render(StoppedGlyph(c.State, b.glyphs) + " " + c.State)
	}
	right := state + "  " + lipgloss.NewStyle().Foreground(b.skin.palette.Priority(c.Priority)).Render(c.Priority)
	var told []string
	for _, label := range c.Labels {
		told = append(told, b.skin.Says.Render(b.glyphs.Label+" "+label))
	}
	if day := domain.FormatDay(c.Deadline); day != "" {
		text, show, late := engine.CardDeadline(c, b.now())
		style := b.skin.Bright
		if late {
			style = b.skin.Late
		}
		if show {
			day += " · " + text
		}
		told = append(told, b.skin.Says.Render("due ")+style.Render(day))
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

func (b Board) subTasks(c engine.Card, width int) []string {
	var out []string
	for _, part := range engine.Hanging(b.snap.Cards, c.Number) {
		box, name := "[ ]", b.skin.Bright.Render(part.Title)
		if engine.Finished(part) {
			box, name = "[x]", b.skin.Faint.Strikethrough(true).Render(part.Title)
		}
		out = append(out, Clip(b.skin.Says.Render(box)+" "+b.skin.Faint.Render(strconv.Itoa(part.Number))+"  "+name, width))
	}
	return out
}

func (b Board) partsHead(c engine.Card) string {
	return "SUB-TASKS   " + engine.Tallied(engine.Hanging(b.snap.Cards, c.Number))
}

func (b Board) serves(c engine.Card, width int) string {
	var out []string
	for _, g := range b.snap.Served(c) {
		out = append(out, Spread(
			lipgloss.NewStyle().Foreground(b.skin.palette.Standing(g.State)).Render(StandingGlyph(g.State, b.glyphs))+
				" "+b.skin.Faint.Render(strconv.Itoa(g.Number))+"  "+b.skin.Bright.Render(g.Title),
			b.skin.Says.Render(g.Counted()), width))
	}
	return strings.Join(out, "\n")
}

func (b Board) history(number, width, rows int) []string {
	if b.heard.Number != number {
		return nil
	}
	if b.heard.Err != nil {
		return []string{b.skin.Wrong.Render(b.heard.Err.Error())}
	}
	drawn := make([]string, 0, len(b.heard.Acts))
	for _, act := range b.heard.Acts {
		drawn = append(drawn, Clip(b.line(act), width))
	}
	if rows < 1 {
		return drawn
	}
	pane := viewport.New(viewport.WithWidth(width), viewport.WithHeight(min(rows, len(drawn))))
	pane.SetContent(strings.Join(drawn, "\n"))
	pane.GotoBottom()
	return strings.Split(pane.View(), "\n")
}

func (b Board) hearing(number int) tea.Cmd {
	return func() tea.Msg {
		shown, err := b.work.Ask(engine.TraceShow, engine.Numbered(number))
		return Heard{Number: number, Acts: shown.Acts, Err: err}
	}
}

func pad(s string, width int) string {
	return s + strings.Repeat(" ", max(0, width-lipgloss.Width(s)))
}
