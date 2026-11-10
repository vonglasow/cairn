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
	if b.pages.None() {
		return ""
	}
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
	case OnWorkspaces:
		return b.drawSwitcher()
	}
	if b.deleting != nil {
		return b.drawDeleting()
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
	return strings.Join(append(out, b.titleLine(), b.foot()), "\n")
}

func (b Board) tabs() string {
	names := map[Tab]string{Tasks: "tasks", Objectives: "objectives", Time: "time"}
	left := " " + b.placeMark()
	for _, tab := range b.shown() {
		name := b.skin.TabOff.Render(names[tab])
		if tab == b.tab {
			name = b.skin.TabOn.Render(strings.ToUpper(names[tab]))
		}
		left += name + " "
	}
	left = strings.TrimSuffix(left, " ")
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
	if b.screen == OnWorkspaces {
		return "WORKSPACES"
	}
	if b.filtering {
		return "FILTER"
	}
	if b.choosingView {
		return "VIEWS"
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
	var outside []engine.Card
	for _, r := range b.references() {
		outside = append(outside, r.Card)
	}
	far, farBand := engine.Deal(outside, b.archived)
	top := l.Top
	var out []string
	if b.inside != 0 {
		out = append(out, b.insideBand(), "")
		top = max(0, top-insideRows)
	}
	columns := make([]string, 0, len(all))
	for i, column := range all {
		under := !b.inBand && b.mode == Normal && b.column == i
		columns = append(columns, strings.Join([]string{
			b.header(column.Name, len(column.Cards), l.Column, under),
			b.rule(l.Column, under),
			b.stack(column.Cards, far[i].Cards, l, top, b.offsets[i]),
		}, "\n"))
	}
	out = append(out, lipgloss.NewStyle().PaddingLeft(margin).Render(b.apart(columns)))
	return append(out, b.bandRows(band, farBand, l)...)
}

func (b Board) bandRows(band, far []engine.Card, l Layout) []string {
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
		lipgloss.NewStyle().PaddingLeft(offset).Render(b.across(band, far, l)),
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

func (b Board) across(cards, far []engine.Card, l Layout) string {
	return through(b.laidOut(cards, far, l), l.Band, l.Rows, b.bandAt)
}

func (b Board) laidOut(cards, far []engine.Card, l Layout) string {
	wide := engine.Columns - 1
	drawn := make([]string, 0, len(cards)+len(far))
	for _, c := range cards {
		drawn = append(drawn, b.card(c, l))
	}
	for _, c := range far {
		drawn = append(drawn, b.reference(c, l))
	}
	var rows []string
	for at := 0; at < len(drawn); at += wide {
		end := min(at+wide, len(drawn))
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, spaced(drawn[at:end])...))
	}
	return strings.Join(rows, "\n")
}

func (b Board) stack(cards, far []engine.Card, l Layout, rows, at int) string {
	drawn := make([]string, 0, len(cards)+len(far))
	for _, c := range cards {
		drawn = append(drawn, b.card(c, l))
	}
	for _, c := range far {
		drawn = append(drawn, b.reference(c, l))
	}
	return through(strings.Join(drawn, "\n"), l.Column, rows, at)
}

func (b Board) card(c engine.Card, l Layout) string {
	return b.skin.Card(l.Column, c.Priority, c.InFocus, b.chosen(c)).
		Render(strings.Join(Rows(b.servingNamed(c), b.glyphs, b.skin, b.now(), l.Interior), "\n"))
}

func (b Board) reference(c engine.Card, l Layout) string {
	rows := Rows(b.servingNamed(c), b.glyphs, SkinOf(Faded), b.now(), l.Interior)
	for _, r := range b.references() {
		if r.Card.Number == c.Number {
			rows[len(rows)-1] = b.skin.Late.Render(Clip(r.Touches(""), l.Interior))
		}
	}
	return b.skin.Reference(l.Column).Render(strings.Join(rows, "\n"))
}

func (b Board) servingNamed(c engine.Card) engine.Card {
	opened, found := b.snap.Goal(b.inside)
	if b.inside == 0 || !found {
		return c
	}
	var kept []string
	for _, name := range c.Serves {
		if name != opened.Title {
			kept = append(kept, name)
		}
	}
	c.Serves = kept
	return c
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
	g, found := b.snap.Goal(b.inside)
	if !found {
		return ""
	}
	mark := b.skin.palette.Standing(g.State).Style().Render(StandingGlyph(g.State, b.glyphs) + " " + g.State)
	right := mark + " "
	if waits := b.waitsOn(g, true); waits != "" {
		right = b.skin.Band.Render(waits) + "   " + right
	}
	head := Spread(" "+b.skin.Title.Render(strconv.Itoa(g.Number)+"   "+g.Title), right, b.width)
	line := " " + b.progress(g.Done, g.Total, barWidth) + "  " + b.skin.Faint.Render(g.Counted())
	if chips := b.chips(); chips != "" {
		line += "   " + chips
	}
	return head + "\n" + Clip(line, b.width)
}

func (b Board) chips() string {
	segments := b.segments()
	said := make([]string, 0, len(segments))
	for i, seg := range segments {
		chip := seg.Name + " " + strconv.Itoa(seg.Count)
		if i == b.segmentOn() {
			said = append(said, b.skin.Picked.Render("‹ "+chip+" ›"))
			continue
		}
		said = append(said, b.skin.Says.Render(chip))
	}
	return strings.Join(said, "  ")
}

func (b Board) objectivesList(l Layout) []string {
	branches := engine.Branches(b.goals())
	width := b.width - margin - margin
	drawn := make([]string, 0, len(branches))
	for _, br := range branches {
		drawn = append(drawn, b.goalCard(br, width))
	}
	rows := l.Top + l.Rows + 2
	return []string{
		" " + Spread(b.skin.Head.Render("OBJECTIVES"), b.skin.Count.Render(strconv.Itoa(len(branches))), b.width-2),
		" " + b.rule(b.width-2, false),
		lipgloss.NewStyle().PaddingLeft(margin).Render(
			through(strings.Join(drawn, "\n"), width, rows, b.goalAt)),
	}
}

const gutter = 2

func (b Board) goalCard(br engine.Branch, width int) string {
	edge := b.skin.Faint.Render("  ")
	if br.Goal.Number == b.goal {
		edge = b.skin.Title.Render("┃ ")
	}
	first, second := "", ""
	if br.Under {
		first, second = "├─ ", "│  "
		if br.Last {
			first, second = "└─ ", "   "
		}
	}
	indent := lipgloss.Width(first)
	rows := b.goalRows(br.Goal, width-gutter-indent)
	return edge + b.skin.Rule.Render(first) + rows[0] + "\n" + edge + b.skin.Rule.Render(second) + rows[1]
}

func (b Board) foot() string {
	if b.err != nil {
		return " " + b.skin.Wrong.Render(b.err.Error())
	}
	if b.asking {
		return b.question()
	}
	if b.deleting != nil {
		return b.deletingFoot()
	}
	if b.screen == OnWorkspaces {
		return b.switcherBar()
	}
	if b.choosing {
		return b.kindPopup()
	}
	if b.filtering {
		return b.filtersLine()
	}
	if b.choosingView {
		return b.viewsLine()
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
			b.offer("jk ↓↑", "objectives"), b.offer("⏎", "open"), b.offer("t", "its tasks"), b.offer("a", "new"),
			b.offer("e", "edit"), b.offer("D", "delete"), b.offer("?", "help"),
			b.offer("N", "under it"), b.offer("z", "fold"), b.offer("/", "search"), b.offer("⇥", "tab"), b.offer(":", "command"), b.roaming(),
		}
	}
	if b.mode == Insert {
		return []string{
			b.offer("hjkl ←↓↑→", "tasks"), b.offer("⇧hjkl", "move"), b.offer("⏎", "open"),
			b.offer("e", "edit"), b.offer("f", "focus"), b.offer("D", "delete"),
			b.offer("x", "archive"), b.offer("?", "help"), b.offer("esc", "back"), b.roaming(), b.offer("o", "linked"),
		}
	}
	out := []string{
		b.offer("hjkl ←↓↑→", "columns"), b.offer("i", "insert"),
	}
	if b.inside != 0 {
		if len(b.segments()) > 0 {
			out = append(out, b.offer("[ ]", "sub-objective"))
		}
		out = append(out, b.offer("a", "new here"), b.offer("/", "search"),
			b.offer("?", "help"), b.offer("esc", "objectives"))
		return out
	}
	if len(b.segments()) > 0 {
		out = append(out, b.offer("[ ]", "sub-objective"))
	}
	return append(out, b.offer("a", "new"), b.offer("/", "search"),
		b.offer("⇥", "tab"), b.offer("?", "help"), b.offer(":", "command"), b.roaming())
}

func (b Board) roaming() string {
	if !b.roams() {
		return ""
	}
	return b.offer("^O", "workspace")
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
	if b.going != nil {
		return " " + b.skin.Title.Render("delete step "+b.going.In.Given["step"]+" of task "+b.going.In.Given["number"]+", “"+b.going.Title+"”?") + "  " +
			b.bar(b.offer("⏎", "delete it"), b.offer("esc", "keep it"))
	}
	if b.aboutReason(engine.ErrNewerStore) {
		return " " + b.skin.Title.Render("this store comes from a newer cairn") + "  " +
			b.bar(b.offer("esc", "close the board, then open cairn again"))
	}
	if b.aboutReason(engine.ErrStoreBehind) {
		return " " + b.skin.Title.Render("apply what this store is waiting for?") + "  " +
			b.bar(b.offer("⏎", "apply, a backup is set aside first"), b.offer("esc", "quit"))
	}
	if b.agreeing != nil {
		return " " + b.skin.Title.Render("put task "+b.agreeing.In.Given["number"]+" in done, ticking its steps?") + "  " +
			b.bar(b.offer("⏎", "finish them all"), b.offer("esc", "keep them as they are"))
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

func (b Board) titleLine() string {
	if len(b.warned) > 0 {
		return " " + b.skin.Band.Render(Clip(strings.Join(b.warned, " · "), b.width-margin-margin))
	}
	if b.opened != "" {
		return " " + b.skin.Faint.Render(Clip(b.opened, b.width-margin-margin))
	}
	title := b.wholeTitle()
	if said := b.viewSays(); title == "" && said != "" {
		return " " + b.skin.Faint.Render(Clip(said, b.width-margin-margin))
	}
	if title == "" && b.filters.Narrows() && !b.filtering {
		said := b.skin.Faint.Render("filtered by " + b.filters.Said())
		if chips := b.chips(); chips != "" {
			said += "   " + chips
		}
		return " " + Clip(said, b.width-margin-margin)
	}
	if title == "" {
		return ""
	}
	return " " + b.skin.Faint.Render(Clip(title, b.width-margin-margin))
}

// wholeTitle is the title of the card under the cursor in insert mode when the card cuts it, and nothing otherwise.
func (b Board) wholeTitle() string {
	c, ok := b.held()
	if !ok || b.mode != Insert {
		return ""
	}
	lines := Wrap(c.Title, LayoutFor(b.width, b.height).Interior, titleLines)
	if !strings.HasSuffix(lines[titleLines-1], ellipsis) {
		return ""
	}
	return c.Title
}

// apart sets columns side by side with a dim rule in the one-cell gap between each two, on every row.
func (b Board) apart(columns []string) string {
	height := 0
	for _, column := range columns {
		height = max(height, lipgloss.Height(column))
	}
	gap := b.skin.Rule.Render(strings.TrimSuffix(strings.Repeat("│\n", height), "\n"))
	var out []string
	for i, column := range columns {
		if i > 0 {
			out = append(out, gap)
		}
		out = append(out, column)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, out...)
}
