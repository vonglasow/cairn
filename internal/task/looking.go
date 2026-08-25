package task

import (
	"strings"
	"time"

	"cairn/internal/domain"
)

// looking is a task's membrane for a search: it is told every fact a task
// carries and keeps, for each thing named, whether it has turned up in any of
// them — read the way a person reads a task, not the way it is stored, and
// case-insensitive. A number is heard and ignored, on every one of its
// shapes: what a task is numbered, hangs under or serves never contributes,
// because looking for "1" is not what tells a task apart from the one
// hanging under it.
type looking struct {
	found map[string]bool
}

// consider marks every named thing still missing that sits inside text. An
// empty name can never be marked found, since an absent fact contributes
// nothing — nothing named as nothing is a thing looking for it should answer.
func (l *looking) consider(text string) {
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

func (l *looking) Numbered(int)              {}
func (l *looking) Titled(s string)           { l.consider(s) }
func (l *looking) InColumn(s domain.State)   { l.consider(string(s)) }
func (l *looking) Carries(p domain.Priority) { l.consider(string(p)) }
func (l *looking) Described(s string)        { l.consider(s) }
func (l *looking) NotDescribed()             {}
func (l *looking) Labelled(s string)         { l.consider(s) }
func (l *looking) NotLabelled()              {}
func (l *looking) LabelsTold()               {}
func (l *looking) Due(t time.Time)           { l.consider(domain.FormatDay(t)) }
func (l *looking) NotDue()                   {}
func (l *looking) Estimated(d time.Duration) { l.consider(domain.FormatDuration(d)) }
func (l *looking) NotEstimated()             {}
func (l *looking) Spent(d time.Duration)     { l.consider(domain.FormatDuration(d)) }
func (l *looking) NothingSpent()             {}
func (l *looking) Under(int)                 {}
func (l *looking) UnderNothing()             {}
func (l *looking) Serves(int)                {}
func (l *looking) ServesNothing()            {}

// all reports whether every named thing was found.
func (l *looking) all() bool {
	for _, found := range l.found {
		if !found {
			return false
		}
	}
	return true
}

// Matching says whether t carries every one of named, read the way a person
// reads a task rather than the way it is stored, case-insensitive, wherever
// among its facts each one happens to sit.
func Matching(t Task, named []string) bool {
	found := make(map[string]bool, len(named))
	for _, n := range named {
		found[strings.ToLower(n)] = false
	}
	l := &looking{found: found}
	t.Tell(l)
	return l.all()
}
