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

var pageOf = map[string]string{"task": PageTasks, "objective": PageObjectives, "timesheet": PageTime}

// Pages is which of the three pages are shown: every one, unless the configuration turned it off.
type Pages struct {
	off []string
}

// PagesOff is the pages shown when the ones named are turned off.
func PagesOff(off ...string) Pages { return Pages{off: off} }

// PagesFrom reads which pages the configuration turns off; a configuration that cannot be read turns none off, and says so where it is read for its places.
func PagesFrom(h Home) Pages {
	if h.Dir == "" {
		return Pages{}
	}
	read, err := config.Read(h.Dir, h.lookup())
	if err != nil {
		return Pages{}
	}
	return Pages{off: read.PagesOff()}
}

// Shows says whether a page is drawn.
func (p Pages) Shows(page string) bool { return !slices.Contains(p.off, page) }

// None says every page is turned off.
func (p Pages) None() bool {
	return !p.Shows(PageTasks) && !p.Shows(PageObjectives) && !p.Shows(PageTime)
}

// Offers says whether an act is offered: one that belongs to a page is offered only while that page is shown, and the rest always.
func (p Pages) Offers(op Operation) bool {
	page, belongs := pageOf[op.Thing]
	return !belongs || p.Shows(page)
}
