package tui

import (
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"cairn/internal/engine"
)

var subFates = []engine.Fate{engine.Undecided, engine.KeepIt, engine.DeleteIt}

var taskFates = []engine.Fate{engine.Undecided, engine.DetachThem, engine.DeleteIt}

func (b Board) deletingGoal() Board {
	d := b.snap.Deleting(b.goal)
	if !d.Asks() {
		b.asking = true
		return b
	}
	b.deleting, b.deleteRow, b.refusing = &d, 0, ""
	return b
}

func (b Board) fateRows() int {
	rows := len(b.deleting.Subs)
	if b.deleting.AsksTasks() {
		rows++
	}
	return rows
}

func (b Board) choosingFates(key string) (tea.Model, tea.Cmd) {
	d := *b.deleting
	b.refusing = ""
	switch key {
	case "esc":
		b.deleting = nil
		return b, nil
	case "enter":
		if !d.Ready() {
			b.refusing = "choose for each first"
			return b, nil
		}
		b.deleting = nil
		return b, b.carry(engine.ObjectiveDelete, d.Input())
	}
	way, ok := direction(key)
	if !ok {
		return b, nil
	}
	switch way {
	case goDown:
		b.deleteRow = min(b.fateRows()-1, b.deleteRow+1)
	case goUp:
		b.deleteRow = max(0, b.deleteRow-1)
	case goRight, goLeft:
		by := 1
		if way == goLeft {
			by = -1
		}
		d = fated(d, b.deleteRow, by)
		b.deleting = &d
		b.deleteRow = min(b.deleteRow, b.fateRows()-1)
	}
	return b, nil
}

func fated(d engine.Deleting, row, by int) engine.Deleting {
	if row < len(d.Subs) {
		sub := d.Subs[row]
		return d.Choose(sub, stepped(subFates, d.FateOf(sub), by))
	}
	return d.Tasks(stepped(taskFates, d.TasksFate(), by))
}

func stepped(fates []engine.Fate, was engine.Fate, by int) engine.Fate {
	at := slices.Index(fates, was) + by
	return fates[min(len(fates)-1, max(1, at))]
}

func fateSaid(f engine.Fate, many bool) string {
	switch f {
	case engine.KeepIt:
		return "keep it, on its own"
	case engine.DetachThem:
		return "detach them"
	case engine.DeleteIt:
		if many {
			return "delete them"
		}
		return "delete it"
	}
	return "choose"
}

func (b Board) drawDeleting() string {
	return lipgloss.NewCompositor(
		lipgloss.NewLayer(b.drawBoard()),
		lipgloss.NewLayer(b.deletingBox()).X(margin+1).Y(2).Z(1),
	).Render()
}

func (b Board) deletingBox() string {
	d := *b.deleting
	g, _ := b.snap.Goal(d.Number)
	wide := min(switcherWide, b.width-margin-2)
	inner := wide - 2
	edge := b.skin.Rule
	title := "─ DELETE #" + strconv.Itoa(d.Number) + " " + strings.ToUpper(g.Title) + " "
	title = Clip(title, inner)
	lines := []string{edge.Render("┌") + b.skin.Title.Render(title) +
		edge.Render(strings.Repeat("─", max(0, inner-lipgloss.Width(title)))+"┐")}
	row := func(s string) {
		s = Clip(s, inner)
		lines = append(lines, edge.Render("│")+s+strings.Repeat(" ", inner-lipgloss.Width(s))+edge.Render("│"))
	}
	row("")
	name := inner / 2
	choice := func(at int, label string, f engine.Fate, many bool) {
		mark := "  "
		said := b.skin.Says.Render(fateSaid(f, many))
		if at == b.deleteRow {
			mark = b.skin.Key.Render("›") + " "
			said = b.skin.Picked.Render("‹ " + fateSaid(f, many) + " ›")
		}
		row(" " + mark + pad(Clip(label, name-4), name-3) + said)
	}
	for i, sub := range d.Subs {
		s, _ := b.snap.Goal(sub)
		choice(i, "sub-objective #"+strconv.Itoa(sub)+" "+s.Title, d.FateOf(sub), false)
	}
	if d.AsksTasks() {
		touched := make([]string, 0, len(d.Touched()))
		for _, n := range d.Touched() {
			touched = append(touched, "#"+strconv.Itoa(n))
		}
		choice(len(d.Subs), "tasks "+strings.Join(touched, " "), d.TasksFate(), true)
	}
	row("")
	lines = append(lines, edge.Render("└"+strings.Repeat("─", inner)+"┘"))
	return strings.Join(lines, "\n")
}

func (b Board) deletingFoot() string {
	asked := " " + b.skin.Title.Render("what goes with objective "+strconv.Itoa(b.deleting.Number)+"?") + "  "
	if b.refusing != "" {
		asked += b.skin.Wrong.Render(b.refusing) + "  "
	}
	return asked + b.bar(b.offer("jk", "choose"), b.offer("hl", "change"),
		b.offer("⏎", "delete"), b.offer("esc", "keep it"))
}
