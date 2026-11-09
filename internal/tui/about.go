package tui

import (
	"slices"
	"strconv"
	"strings"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

type describer func(field engine.Field, value *string) func() string

func (b Board) about(forObjective bool, number int) describer {
	return func(field engine.Field, value *string) func() string {
		switch {
		case forObjective && field.Arg == "parent":
			said := b.parentsOffered(number)
			return func() string { return said }
		case forObjective && field.Arg == "label":
			said := b.skin.Faint.Render("for objectives: " + orNone(strings.Join(b.snap.LabelsFor(domain.OnObjectives), " · ")))
			return func() string { return said }
		case !forObjective && field.Arg == "objective":
			return func() string { return b.servesTree(*value) }
		}
		return nil
	}
}

func treeOf(p engine.Pick) int {
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
	ticked := engine.NumbersIn(typed)
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
