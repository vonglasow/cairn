package task

import (
	"strconv"
	"time"

	"cairn/internal/domain"
)

// looking is a task's membrane for a search: it is told every fact a task
// carries and keeps, for each thing named, whether it has turned up in any of
// them — read the way a person reads a task, not the way it is stored, and
// case-insensitive. Its own number contributes; the numbers it carries about
// other things do not. What a task hangs under or serves says nothing about
// that task — reaching those would answer every sub-task of 228 to somebody
// looking for 228, which is answering about the wrong task entirely.
//
// Its own number is looked for as a substring like everything else, so 228
// answers 228 and 1228 both. Weighed and kept: a second kind of matching in a
// membrane that has one costs more than a glance at one extra answer, and the
// reason the number was left out altogether — that `task show` is how you ask
// by number — holds only where there is a `task show` to type. On a board the
// search box is the only way in.
//
// The column, the priority and the labels are looked in too. A listing
// narrows by them, but a board has one search box and no listing: left out
// here, a label could not be searched for there, and the boards kept a rule of
// their own that did. One search, on every surface.
type looking struct {
	said []string
}

// Consider keeps one thing the task says, to be looked in.
func (l *looking) Consider(s string) { l.said = append(l.said, s) }

func (l *looking) Numbered(n int)            { l.Consider(strconv.Itoa(n)) }
func (l *looking) Titled(s string)           { l.Consider(s) }
func (l *looking) InColumn(s domain.State)   { l.Consider(string(s)) }
func (l *looking) Carries(p domain.Priority) { l.Consider(string(p)) }
func (l *looking) Described(s string)        { l.Consider(s) }
func (l *looking) NotDescribed()             {}
func (l *looking) Labelled(s string)         { l.Consider(s) }
func (l *looking) NotLabelled()              {}
func (l *looking) LabelsTold()               {}
func (l *looking) BlockedBy(int)             {}
func (l *looking) NotBlocked()               {}
func (l *looking) BlockersTold()             {}
func (l *looking) Because(s string)          { l.Consider(s) }
func (l *looking) NotBecause()               {}
func (l *looking) Due(t time.Time)           { l.Consider(domain.FormatDay(t)) }
func (l *looking) NotDue()                   {}
func (l *looking) Estimated(d time.Duration) { l.Consider(domain.FormatDuration(d)) }
func (l *looking) NotEstimated()             {}
func (l *looking) Spent(d time.Duration)     { l.Consider(domain.FormatDuration(d)) }
func (l *looking) NothingSpent()             {}
func (l *looking) Archived()                 {}
func (l *looking) NotArchived()              {}
func (l *looking) Under(int)                 {}
func (l *looking) UnderNothing()             {}
func (l *looking) Serves(int)                {}
func (l *looking) ServesNothing()            {}
func (l *looking) ObjectivesTold()           {}

// Said is everything a search looks in, read the way a person reads a task
// rather than the way it is stored.
func Said(t Task) []string {
	l := &looking{}
	t.Tell(l)
	return l.said
}

// Matching says whether t carries every one of named, case-insensitive,
// wherever among its facts each one happens to sit.
func Matching(t Task, named []string) bool {
	return domain.Reaches(Said(t), named)
}
