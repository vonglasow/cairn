package tui

import (
	"cairn/internal/board"
	"cairn/internal/engine/label"
	"slices"
	"strconv"
	"strings"
)

type describer func(field board.Field, value *string) func() string

func (b Board) about(forObjective bool, number int) describer {
	return func(field board.Field, value *string) func() string {
		switch {
		case forObjective && field.Arg == "parent":
			said := b.parentsOffered(number)
			return func() string { return said }
		case forObjective && field.Arg == "label":
			said := b.skin.Faint.Render("for objectives: " + orNone(strings.Join(b.snap.LabelsFor(label.OnObjectives), " · ")))
			return func() string { return said }
		case !forObjective && field.Arg == "objective":
			return func() string { return b.servesTree(*value) }
		case !forObjective && (field.Arg == "blocker" || field.Arg == "link"):
			return func() string { return b.tasksOffered(*value, number) }
		}
		return nil
	}
}

func treeOf(p board.Pick) int {
	if p.Under {
		return p.Goal.Parent
	}
	return p.Goal.Number
}

func orNone(said string) string {
	if said == "" {
		return "none yet"
	}
	return said
}

func (b Board) parentsOffered(number int) string {
	offered := []string{}
	for _, g := range b.snap.Parents(number) {
		offered = append(offered, strconv.Itoa(g.Number)+" "+g.Title)
	}
	return b.skin.Faint.Render("offered: " + strings.Join(append(offered, "none"), " · "))
}

func (b Board) servesTree(typed string) string {
	var lines []string
	ticked := board.NumbersIn(typed)
	for _, p := range b.snap.Picks(ticked) {
		if !slices.ContainsFunc(b.snap.Reach(treeOf(p)), func(n int) bool { return slices.Contains(ticked, n) }) {
			continue
		}
		lead := ""
		if p.Under {
			lead = "├ "
			if p.Last {
				lead = "└ "
			}
		}
		box := "[ ]"
		if p.Ticked {
			box = "[x]"
		}
		line := lead + box + " " + strconv.Itoa(p.Goal.Number) + " " + p.Goal.Title
		if !p.Ticked && p.Through > 0 {
			lines = append(lines, b.skin.Faint.Render(line), b.skin.Faint.Render("    counted through "+strconv.Itoa(p.Through)+" below"))
			continue
		}
		if p.Ticked {
			lines = append(lines, b.skin.Bright.Render(line))
			continue
		}
		lines = append(lines, b.skin.Says.Render(line))
	}
	return strings.Join(lines, "\n")
}

const tasksShown = 5

func (b Board) tasksOffered(typed string, number int) string {
	ticked := board.NumbersIn(typed)
	var looked []string
	for _, word := range strings.Fields(typed) {
		if _, err := strconv.Atoi(word); err != nil {
			looked = append(looked, word)
		}
	}
	found := board.FoundCards(b.snap.Cards, strings.Join(looked, " "), number)
	if len(looked) == 0 {
		found = slices.DeleteFunc(found, func(c board.Card) bool { return !slices.Contains(ticked, c.Number) })
	}
	if len(found) == 0 && len(looked) > 0 {
		return b.skin.Faint.Render(board.TasksNoneSays)
	}
	var lines []string
	for _, c := range found[:min(len(found), tasksShown)] {
		box, style := "[ ] ", b.skin.Says
		if slices.Contains(ticked, c.Number) {
			box, style = "[x] ", b.skin.Bright
		}
		lines = append(lines, style.Render(box+strconv.Itoa(c.Number)+" "+c.Title))
	}
	if more := len(found) - tasksShown; more > 0 {
		lines = append(lines, b.skin.Faint.Render("… "+strconv.Itoa(more)+" more"))
	}
	return strings.Join(lines, "\n")
}
