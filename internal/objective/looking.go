package objective

import (
	"time"

	"cairn/internal/domain"
)

// looking is an objective's membrane for a search: it is told every fact an
// objective carries and keeps, for each thing named, whether it has turned up
// in any of them — read the way a person reads an objective, not the way it
// is stored, and case-insensitive. A number is heard and ignored: what an
// objective is numbered never contributes, because looking for "1" is not
// what tells an objective apart from the one numbered 1.
type looking struct {
	*domain.Looking
}

// Where it stands, the priority and a label are heard and dropped: a listing
// narrows by those, and looking for them too would be a second way to ask one
// question. The task side says the same, at greater length.

func (l *looking) Numbered(int)              {}
func (l *looking) Titled(s string)           { l.Consider(s) }
func (l *looking) Standing(State)            {}
func (l *looking) Carries(domain.Priority)   {}
func (l *looking) Described(s string)        { l.Consider(s) }
func (l *looking) NotDescribed()             {}
func (l *looking) Labelled(string)           {}
func (l *looking) NotLabelled()              {}
func (l *looking) LabelsTold()               {}
func (l *looking) Due(t time.Time)           { l.Consider(domain.FormatDay(t)) }
func (l *looking) NotDue()                   {}
func (l *looking) Estimated(d time.Duration) { l.Consider(domain.FormatDuration(d)) }
func (l *looking) NotEstimated()             {}

// Matching says whether o carries every one of named, read the way a person
// reads an objective rather than the way it is stored, case-insensitive,
// wherever among its facts each one happens to sit.
func Matching(o Objective, named []string) bool {
	l := &looking{domain.LookingFor(named)}
	o.Tell(l)
	return l.All()
}
