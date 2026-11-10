package tui

import (
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

func (b Board) allSeries() []engine.Series { return engine.ShownSeries(b.snap.Series, b.query) }

func (b Board) chosenSeries() (engine.Series, bool) {
	all := b.allSeries()
	for _, s := range all {
		if s.Number == b.seriesOn {
			return s, true
		}
	}
	if len(all) > 0 {
		return all[0], true
	}
	return engine.Series{}, false
}

func (b Board) onSeries(n int) Board {
	b.tab, b.mode, b.task, b.inside = Series, Normal, 0, 0
	b.inOccurrences, b.occurrenceAt = false, 0
	b.seriesOn = n
	if s, found := b.chosenSeries(); found {
		b.seriesOn = s.Number
	}
	return b
}

func (b Board) walkSeries(way string) Board {
	s, found := b.chosenSeries()
	if b.inOccurrences && found {
		switch way {
		case goDown:
			b.occurrenceAt = min(len(s.Occurrences)-1, b.occurrenceAt+1)
		case goUp:
			b.occurrenceAt = max(0, b.occurrenceAt-1)
		}
		return b
	}
	all := b.allSeries()
	at := slices.IndexFunc(all, func(one engine.Series) bool { return one.Number == s.Number })
	switch way {
	case goDown:
		at++
	case goUp:
		at--
	default:
		return b
	}
	if at >= 0 && at < len(all) {
		b.seriesOn = all[at].Number
	}
	return b
}

func (b Board) inSeries(key string) (tea.Model, tea.Cmd, bool) {
	if way, ok := direction(key); ok {
		return b.walkSeries(way), nil, true
	}
	s, found := b.chosenSeries()
	switch key {
	case "esc":
		if b.inOccurrences {
			b.inOccurrences = false
			return b, nil, true
		}
		return b, nil, false
	case "enter":
		return b.intoOccurrences(s, found), nil, true
	case "n":
		model, cmd := b.newSeriesForm()
		return model, cmd, true
	}
	if !found {
		return b, nil, slices.Contains([]string{"s", "x", "r", "e", "D", "shift+d", "backspace", "delete", "a", "i"}, key)
	}
	switch key {
	case "s":
		if len(s.Open) == 0 {
			return b, nil, true
		}
		return b, b.carry(engine.TaskSkip, engine.Numbered(s.Open[0])), true
	case "x":
		return b, b.carry(engine.SeriesStop, engine.Numbered(s.Number)), true
	case "r":
		return b, b.carry(engine.SeriesResume, engine.Numbered(s.Number)), true
	case "e":
		model, cmd := b.seriesForm(s)
		return model, cmd, true
	case "D", "shift+d", "backspace", "delete":
		b.seriesGoing = &seriesGoing{number: s.Number}
		return b, nil, true
	case "a", "i":
		return b, nil, true
	}
	return b, nil, false
}

func (b Board) intoOccurrences(s engine.Series, found bool) Board {
	if !found || len(s.Occurrences) == 0 {
		return b
	}
	if !b.inOccurrences {
		b.inOccurrences, b.occurrenceAt = true, 0
		return b
	}
	at := s.Occurrences[min(b.occurrenceAt, len(s.Occurrences)-1)]
	if _, there := b.snap.Card(at.Task); at.Task == 0 || !there {
		return b
	}
	b.task, b.wasScreen, b.screen, b.stepOn = at.Task, OnTheBoard, OnDetail, 0
	return b
}

func (b Board) seriesAct(key string) (tea.Model, tea.Cmd) {
	c, ok := b.held()
	if !ok || c.Series == nil {
		return b, nil
	}
	switch key {
	case "s":
		return b, b.carry(engine.TaskSkip, engine.Numbered(c.Number))
	case "x":
		return b, b.carry(engine.SeriesStop, engine.Numbered(c.Series.Series))
	case "r":
		return b, b.carry(engine.SeriesResume, engine.Numbered(c.Series.Series))
	case "e":
		b.screen = OnTheBoard
		return b.onSeries(c.Series.Series), nil
	}
	return b, nil
}

func (b Board) heldRecurs() bool {
	c, ok := b.held()
	return ok && c.Series != nil && b.pages.Recurs()
}

func (b Board) seriesOffers() []string {
	if !b.heldRecurs() {
		return nil
	}
	c, _ := b.held()
	offers := []string{b.offer("r s", "skip"), b.offer("r x", "stop series")}
	if c.Series.Stopped {
		offers = []string{b.offer("r r", "resume series")}
	}
	return append(offers, b.offer("r e", "its series"))
}

func (b Board) occursSaid(c engine.Card) string {
	o := c.Series
	said := "↻" + strconv.Itoa(o.Series) + " " + o.Often
	if !strings.Contains(o.Says, "after the last is done") {
		said += ", on the calendar"
	}
	if s, found := b.snap.SeriesOf(o.Series); found {
		said += " · " + strconv.Itoa(s.Done) + " done · " + strconv.Itoa(s.Skipped) + " skipped · " + strconv.Itoa(s.Missed) + " missed"
	}
	switch {
	case o.Stopped:
		said += " · stopped"
	case len(o.Then) > 0:
		said += " · next " + engine.ShortDays(o.Then[:1])
	}
	return said
}

func (b Board) seriesTab() []string {
	rows := max(1, b.height-4)
	width := b.width - margin - margin
	left := width * 2 / 5
	right := width - left - 2
	list := b.seen(left, rows)(b.seriesList(left))
	chosen := b.seen(right, rows)(b.seriesChosen(right))
	return []string{lipgloss.NewStyle().PaddingLeft(margin).Render(
		lipgloss.JoinHorizontal(lipgloss.Top, list, b.skin.Rule.Render(strings.TrimSuffix(strings.Repeat(" │\n", rows), "\n")), chosen))}
}

func (b Board) seen(width, rows int) func([]string, int) string {
	return func(lines []string, chosen int) string {
		for i := range lines {
			lines[i] = Clip(lines[i], width)
		}
		at := max(0, chosen+2-rows)
		return lipgloss.NewStyle().Width(width).Render(through(strings.Join(lines, "\n"), width, rows, at))
	}
}

func (b Board) seriesList(width int) ([]string, int) {
	all := b.allSeries()
	if len(all) == 0 {
		return []string{b.skin.Faint.Render("no series yet — n makes one")}, 0
	}
	var out []string
	chosen := 0
	on, _ := b.chosenSeries()
	heads := map[string]string{"active": "Active", "stopped": "Stopped"}
	state := ""
	for _, s := range all {
		if s.State != state {
			state = s.State
			count := 0
			for _, other := range all {
				if other.State == state {
					count++
				}
			}
			out = append(out, Spread(b.skin.Head.Render(heads[state]), b.skin.Count.Render(strconv.Itoa(count)), width))
		}
		edge := "  "
		if s.Number == on.Number {
			edge, chosen = b.skin.Title.Render("┃ "), len(out)
		}
		when := b.whenSaid(s.Next[:min(1, len(s.Next))]...)
		if s.State == "stopped" {
			when = "stopped since " + b.whenSaid(s.Stopped)
		}
		head := b.skin.Faint.Render("↻"+strconv.Itoa(s.Number)) + "  " + b.skin.Bright.Render(s.Title)
		out = append(out, edge+Spread(head, b.skin.Says.Render(when), width-2))
		so := s.Rule + " · " + strconv.Itoa(s.Done) + " done · " + strconv.Itoa(s.Skipped) + " skipped · " + strconv.Itoa(s.Missed) + " missed"
		out = append(out, edge+"    "+b.skin.Faint.Render(so))
	}
	return out, chosen
}

func (b Board) seriesChosen(width int) ([]string, int) {
	s, found := b.chosenSeries()
	if !found {
		return []string{b.skin.Faint.Render("n makes a series: a task's form with its rule to fill in")}, 0
	}
	standing := b.skin.Bright.Render(engine.StandingSaid(s))
	out := []string{
		Spread(b.skin.Faint.Render("↻"+strconv.Itoa(s.Number))+"  "+b.skin.Title.Render(s.Title), standing, width),
		b.skin.Says.Render(s.Says),
		b.skin.Head.Render("NEXT") + "  " + b.skin.Bright.Render(engine.NextSaid(s)),
		"",
		b.skin.Head.Render("TEMPLATE") + b.skin.Faint.Render(" — what the next one is made from"),
		"  " + b.skin.Bright.Render(engine.TemplateSaid(s.Template)),
		b.skin.Head.Render("SO FAR"),
		"  " + b.skin.Bright.Render(strconv.Itoa(s.Done)+" done · "+strconv.Itoa(s.Skipped)+" skipped · "+strconv.Itoa(s.Missed)+" missed") +
			b.skin.Faint.Render(spentSaid(s.TimeSpent)),
		b.skin.Head.Render("ON THE BOARD NOW") + "  " + b.skin.Bright.Render(domain.OrNothing(numbersSaid(s.Open))),
		"",
		b.skin.Head.Render("OCCURRENCES · "+strconv.Itoa(len(s.Occurrences))) + b.skin.Faint.Render("  archived ones included · ⏎ opens one"),
	}
	chosen := 0
	for i, o := range s.Occurrences {
		mark := "  "
		if b.inOccurrences && i == b.occurrenceAt {
			mark, chosen = b.skin.Key.Render("›")+" ", len(out)
		}
		what := "  —  "
		if o.Task != 0 {
			what = pad("#"+strconv.Itoa(o.Task), 5)
		}
		is := o.Is
		if o.Archived {
			is += " · archived"
		}
		out = append(out, mark+b.skin.Says.Render(pad(engine.ShortDays([]string{o.Day}), 11))+b.skin.Faint.Render(what)+"  "+
			b.skin.Bright.Render(pad(is, 18))+b.skin.Faint.Render(o.TimeSpent))
	}
	return out, chosen
}

func (b Board) whenSaid(days ...string) string {
	if len(days) == 0 {
		return ""
	}
	return engine.DaySaid(days[0], b.now())
}

func spentSaid(spent string) string {
	if spent == "" {
		return ""
	}
	return " · " + spent
}

func numbersSaid(all []int) string {
	said := make([]string, 0, len(all))
	for _, n := range all {
		said = append(said, "#"+strconv.Itoa(n))
	}
	return strings.Join(said, " ")
}

func (b Board) seriesTabOffers() []string {
	if b.inOccurrences {
		return []string{b.offer("jk ↓↑", "occurrences"), b.offer("⏎", "open the task"), b.offer("esc", "the series"), b.offer("?", "help")}
	}
	offers := []string{b.offer("jk ↓↑", "series"), b.offer("⏎", "occurrences")}
	if s, found := b.chosenSeries(); found {
		offers = append(offers, b.offer("e", "edit"))
		if s.State == "stopped" {
			offers = append(offers, b.offer("r", "resume"))
		} else {
			offers = append(offers, b.offer("s", "skip"), b.offer("x", "stop"))
		}
		offers = append(offers, b.offer("⌫", "delete"))
	}
	return append(offers, b.offer("n", "new series"), b.offer("/", "search"), b.offer("⇥", "tab"),
		b.offer("?", "help"), b.offer(":", "command"), b.roaming())
}

type seriesGoing struct {
	number int
	choice int
	again  bool
}

func (b Board) choosingSeriesFate(key string) (tea.Model, tea.Cmd) {
	going := *b.seriesGoing
	s, found := b.snap.SeriesOf(going.number)
	if !found {
		b.seriesGoing = nil
		return b, nil
	}
	switch key {
	case "esc":
		if going.again {
			going.again = false
			b.seriesGoing = &going
			return b, nil
		}
		b.seriesGoing = nil
		return b, nil
	case "enter", "return":
		if engine.SeriesFates[going.choice] == "delete-tasks" && !going.again {
			going.again = true
			b.seriesGoing = &going
			return b, nil
		}
		b.seriesGoing = nil
		in := engine.Agreed(engine.Numbered(s.Number))
		in.Given[engine.SeriesFates[going.choice]] = "true"
		return b, b.carry(engine.SeriesDelete, in)
	}
	if way, ok := direction(key); ok && !going.again {
		switch way {
		case goDown:
			going.choice = min(len(engine.SeriesFates)-1, going.choice+1)
		case goUp:
			going.choice = max(0, going.choice-1)
		}
		b.seriesGoing = &going
	}
	return b, nil
}

func (b Board) drawSeriesGoing() string {
	return lipgloss.NewCompositor(
		lipgloss.NewLayer(b.drawBoard()),
		lipgloss.NewLayer(b.seriesGoingBox()).X(margin+1).Y(2).Z(1),
	).Render()
}

func (b Board) seriesGoingBox() string {
	going := *b.seriesGoing
	s, _ := b.snap.SeriesOf(going.number)
	made := engine.MadeStill(s)
	wide := min(switcherWide, b.width-margin-2)
	inner := wide - 2
	edge := b.skin.Rule
	title := Clip("─ DELETE THE SERIES ↻"+strconv.Itoa(s.Number)+" "+strings.ToUpper(s.Title)+" ", inner)
	lines := []string{edge.Render("┌") + b.skin.Title.Render(title) +
		edge.Render(strings.Repeat("─", max(0, inner-lipgloss.Width(title)))+"┐")}
	row := func(said string) {
		said = Clip(said, inner)
		lines = append(lines, edge.Render("│")+said+strings.Repeat(" ", inner-lipgloss.Width(said))+edge.Render("│"))
	}
	row("")
	row(" " + b.skin.Says.Render("It made "+strconv.Itoa(len(made))+" tasks still here; its rule and template go, its trace stays."))
	row("")
	open := "none is open"
	if len(s.Open) > 0 {
		open = numbersSaid(s.Open)
	}
	for i, said := range []string{
		"keep its tasks, as ordinary tasks — they lose ↻ and keep the rest",
		"keep the tasks done, delete the open one (" + open + ")",
		"delete all " + strconv.Itoa(len(made)) + " tasks too — their notes and time go with them",
	} {
		mark, dressed := "  ", b.skin.Says.Render(said)
		if i == going.choice {
			mark, dressed = b.skin.Key.Render("›")+" ", b.skin.Picked.Render(said)
		}
		row(" " + mark + dressed)
	}
	row("")
	lines = append(lines, edge.Render("└"+strings.Repeat("─", inner)+"┘"))
	return strings.Join(lines, "\n")
}

func (b Board) seriesGoingFoot() string {
	going := *b.seriesGoing
	s, _ := b.snap.SeriesOf(going.number)
	if going.again {
		return " " + b.skin.Title.Render("delete all "+strconv.Itoa(len(engine.MadeStill(s)))+" tasks too?") + "  " +
			b.bar(b.offer("⏎", "delete them all"), b.offer("esc", "choose again"))
	}
	return " " + b.skin.Title.Render("delete the series ↻"+strconv.Itoa(s.Number)+" "+s.Title+"?") + "  " +
		b.bar(b.offer("jk", "choose"), b.offer("⏎", "delete the series"), b.offer("esc", "cancel"))
}

func (b Board) taskFields(c engine.Card) []engine.Field {
	fields := engine.EditFields(c)
	if c.Series == nil && b.pages.Recurs() {
		fields = append(fields, engine.RepeatsField())
	}
	return fields
}

func (b Board) newSeriesForm() (Board, tea.Cmd) {
	fields := append(engine.AddingFields(engine.StateAt(1)), engine.RepeatsField())
	b.form = asking(fields, b.formWidth(), b.formHeight(), b.skin, b.about(false, 0))
	b.form.Adding = true
	b.screen = OnForm
	b, cmd := b.started()
	return b.focusingOn(engine.RepeatsField().Name, len(fields)), cmd
}

func (b Board) seriesForm(s engine.Series) (Board, tea.Cmd) {
	b.form = asking(engine.SeriesFields(s), b.formWidth(), b.formHeight(), b.skin, nil)
	b.form.Number, b.form.ForSeries = s.Number, true
	b.screen = OnForm
	return b.started()
}

func (b Board) asksWhole(form Form) bool {
	if form.Adding || form.ForObjective || form.ForSeries || !b.pages.Recurs() {
		return false
	}
	c, found := b.snap.Card(form.Number)
	return found && engine.ChangesTheSeries(c, form.typed())
}

func (b Board) answeringWhole(key string) (tea.Model, tea.Cmd) {
	form := *b.wholeSave
	switch key {
	case "enter", "return":
		b.asking, b.wholeSave = false, nil
		return b, b.saveTask(form, false)
	case "s":
		b.asking, b.wholeSave = false, nil
		return b, b.saveTask(form, true)
	case "esc":
		b.asking, b.wholeSave = false, nil
		b.form, b.screen = form, OnForm
		return b, nil
	}
	return b, nil
}
