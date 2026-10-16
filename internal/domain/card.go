package domain

import (
	"fmt"
	"strings"
)

// Fact is one line of a card: what is carried, and its value as it is read — empty when nothing is carried.
type Fact struct {
	Key, Value string
}

// Card lays out something numbered the way a person reads it in full: the number and the title, then one fact a line, the values lined up in one column two spaces past the longest name, and an absence shown as one.
func Card(number int, title string, facts []Fact) string {
	wide := 0
	for _, f := range facts {
		wide = max(wide, len([]rune(f.Key)))
	}
	var out strings.Builder
	fmt.Fprintf(&out, "#%d  %s\n\n", number, title)
	for _, f := range facts {
		fmt.Fprintf(&out, "%s%s  %s\n", f.Key, strings.Repeat(" ", wide-len([]rune(f.Key))), OrNothing(f.Value))
	}
	return out.String()
}

// Ref writes a number the way a person reads it as naming something: #4.
func Ref(n int) string { return fmt.Sprintf("#%d", n) }

// Line says how far along something is, naming what it is about, and says nothing at all when nothing is under it.
func (p Progress) Line(about string) string {
	if !p.Any() {
		return ""
	}
	var counts []string
	if p.Tasks.Total > 0 {
		counts = append(counts, fmt.Sprintf("%d of %d tasks", p.Tasks.Done, p.Tasks.Total))
	}
	if p.Steps.Total > 0 {
		counts = append(counts, fmt.Sprintf("%d of %d steps", p.Steps.Done, p.Steps.Total))
	}
	return fmt.Sprintf("%s are done — %s is %d%% done", strings.Join(counts, " and "), about, p.Percent())
}

// Title is a title as it is kept: what surrounds it trimmed away, what sits inside it kept, refused when nothing is left or when it is more than one line.
func Title(of, raw string) (string, error) {
	title := strings.TrimSpace(raw)
	if title == "" {
		return "", fmt.Errorf("%s %w", of, ErrTitleRequired)
	}
	if err := OneLine(title); err != nil {
		return "", err
	}
	return title, nil
}

// Looking is a search through what something carries: each thing named, and whether it has turned up yet in anything it was shown.
type Looking struct {
	found map[string]bool
}

// LookingFor starts a search for every thing named.
func LookingFor(named []string) *Looking {
	found := make(map[string]bool, len(named))
	for _, n := range named {
		found[strings.ToLower(n)] = false
	}
	return &Looking{found: found}
}

// Consider marks every named thing still missing that sits inside text.
func (l *Looking) Consider(text string) {
	text = strings.ToLower(text)
	for name, already := range l.found {
		if already || name == "" {
			continue
		}
		if strings.Contains(text, name) {
			l.found[name] = true
		}
	}
}

// Reaches says whether every thing named sits somewhere in what was said.
func Reaches(said, named []string) bool {
	if len(named) == 0 {
		return false
	}
	l := LookingFor(named)
	for _, s := range said {
		l.Consider(s)
	}
	return l.All()
}

// All says whether every named thing was found.
func (l *Looking) All() bool {
	for _, found := range l.found {
		if !found {
			return false
		}
	}
	return true
}
