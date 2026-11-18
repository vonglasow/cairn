package board

import (
	"strings"
	"time"

	"cairn/internal/engine"
	"cairn/internal/engine/item"
)

// Tone is what a piece of the board means, for each surface to give it its colour.
type Tone string

// The tones a board speaks in.
const (
	Normal Tone = "normal"
	Faint  Tone = "faint"
	Late   Tone = "late"
	Passed Tone = "passed"
	Says   Tone = "says"
	Accent Tone = "accent"
)

// Mark is what the board says without words, for each surface to give it its glyph.
type Mark string

// The marks a card wears.
const (
	Blocked   Mark = "blocked"
	Watching  Mark = "watching"
	Paused    Mark = "paused"
	Archived  Mark = "archived"
	Repeats   Mark = "repeats"
	Described Mark = "described"
	Steps     Mark = "steps"
	Label     Mark = "label"
	Serves    Mark = "serves"
	InTheWay  Mark = "in-the-way"
	Freed     Mark = "freed"
)

// Fact is one thing a card says: its words, the mark before them, and what they mean.
type Fact struct {
	Said string
	Mark Mark
	Tone Tone
}

// FreedSaid is what a stopped task says once nothing it waits on is open.
const FreedSaid = "nothing blocks it"

// FactsOf is every fact a card says, in the order both surfaces show them: what is in its way, or that nothing is; how many tasks it is in the way of; when it comes back or is wanted; its cost while the time is shown; its steps done; its labels; what it serves.
func FactsOf(c Card, now time.Time, p engine.Pages) []Fact {
	if c.Freed {
		return []Fact{{Said: FreedSaid, Mark: Freed, Tone: Accent}}
	}
	var facts []Fact
	if way := c.InTheWay(); way != "" && Stopped(c.State) {
		facts = append(facts, Fact{Said: way, Mark: InTheWay, Tone: Faint})
	}
	if blocks := c.BlocksSaid(); blocks != "" {
		facts = append(facts, Fact{Said: blocks, Tone: Normal})
	}
	if text, show, late, passed := OccurrenceDay(c, now); show {
		facts = append(facts, Fact{Said: text, Tone: toneOf(late, passed)})
	} else if text, show, late := CardDeadline(c, now); show && c.Series == nil {
		facts = append(facts, Fact{Said: text, Tone: toneOf(late, false)})
	}
	if spent, over := Cost(p, c); spent != "" {
		facts = append(facts, Fact{Said: spent, Tone: toneOf(over, false)})
	}
	if done := c.PartsDone(); done != "" {
		facts = append(facts, Fact{Said: done, Mark: Steps, Tone: Faint})
	}
	for _, l := range c.Labels {
		facts = append(facts, Fact{Said: l, Mark: Label, Tone: Says})
	}
	for _, name := range c.Serves {
		facts = append(facts, Fact{Said: name, Mark: Serves, Tone: Says})
	}
	return facts
}

// CutFromStart is said cut from its start to what fits, the cut marked by an ellipsis, so what it ends with stays read; nothing when not even its last letter fits.
func CutFromStart(said string, fits func(string) bool) string {
	if fits(said) {
		return said
	}
	runes := []rune(said)
	for i := 1; i < len(runes); i++ {
		if cut := "…" + strings.TrimLeft(string(runes[i:]), " "); fits(cut) {
			return cut
		}
	}
	return ""
}

func toneOf(late, passed bool) Tone {
	switch {
	case late:
		return Late
	case passed:
		return Passed
	}
	return Normal
}

// MarksOf is what a card wears beside its number: archived, coming back, described, and the state it is stopped in.
func MarksOf(c Card) []Mark {
	var marks []Mark
	if c.Archived {
		marks = append(marks, Archived)
	}
	if c.Series != nil {
		marks = append(marks, Repeats)
	}
	if c.Described() {
		marks = append(marks, Described)
	}
	switch c.State {
	case string(item.Blocked):
		marks = append(marks, Blocked)
	case string(item.Watching):
		marks = append(marks, Watching)
	case string(item.Paused):
		marks = append(marks, Paused)
	}
	return marks
}
