package engine

import (
	"slices"

	"cairn/internal/config"
)

// The three pages a board draws, by the names the configuration turns them off by.
const (
	PageTasks      = "tasks"
	PageObjectives = "objectives"
	PageTime       = "time"
)

var pagesOf = map[string][]string{
	"task":      {PageTasks},
	"series":    {PageTasks},
	"step":      {PageTasks},
	"objective": {PageObjectives},
	"label":     {PageTasks, PageObjectives},
	"timesheet": {PageTime},
}

// Pages is which of the three pages are shown, every one unless the configuration turned it off, and how many days a week of the time page shows.
type Pages struct {
	off          []string
	noWeekend    bool
	noRecurrence bool
}

// PagesOff is the pages shown when the ones named are turned off.
func PagesOff(off ...string) Pages { return Pages{off: off} }

// PagesFrom reads which pages the configuration turns off; a configuration that cannot be read turns none off, and says so where it is read for its places.
func PagesFrom(h Home) Pages {
	if h.Dir == "" {
		return Pages{}
	}
	read, _, err := h.settled()
	if err != nil {
		return Pages{}
	}
	return Pages{off: read.PagesOff(), noWeekend: read.TimeWeekends() == config.WeekendsHidden, noRecurrence: !read.Recurrence()}
}

// WeekDays is how many days a week of the time page shows, counted from Monday: seven, or five once the configuration hides the weekends.
func (p Pages) WeekDays() int {
	if p.noWeekend {
		return 5
	}
	return 7
}

// WithoutWeekends is the pages shown with a week cut to its five working days.
func (p Pages) WithoutWeekends() Pages {
	p.noWeekend = true
	return p
}

// Shows says whether a page is drawn.
func (p Pages) Shows(page string) bool { return !slices.Contains(p.off, page) }

// None says every page is turned off.
func (p Pages) None() bool {
	return !p.Shows(PageTasks) && !p.Shows(PageObjectives) && !p.Shows(PageTime)
}

// Offers says whether an act is offered: one that belongs to pages is offered only while one of them is shown — a step with the tasks, a label while there is a task or an objective to label — and the rest always. Every surface asks it: the command line, the MCP server, the boards' tabs, menus and command lines.
func (p Pages) Offers(op Operation) bool {
	if op.Repeats && p.noRecurrence {
		return false
	}
	pages, belongs := pagesOf[op.Thing]
	return !belongs || slices.ContainsFunc(pages, p.Shows)
}

// Recurs says whether a task may come back as a series: the series' acts and a rule given to a task are offered only then.
func (p Pages) Recurs() bool { return !p.noRecurrence }

// Shaped is an act as it is offered: without what belongs to the series while recurrence is off.
func (p Pages) Shaped(op Operation) Operation {
	if !p.noRecurrence {
		return op
	}
	op.Takes = slices.DeleteFunc(slices.Clone(op.Takes), func(a Argument) bool { return a.Repeats })
	return op
}

// Called is the act a thing and a verb name, when there is one and it is offered.
func (p Pages) Called(thing, verb string) (Operation, bool) {
	op, found := Called(thing, verb)
	if !found || !p.Offers(op) {
		return Operation{}, false
	}
	return op, true
}
