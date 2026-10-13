package objective

import (
	"strconv"
	"time"

	"cairn/internal/domain"
)

type looking struct {
	said []string
}

// Consider keeps one thing the objective says, to be looked in.
func (l *looking) Consider(s string) { l.said = append(l.said, s) }

func (l *looking) Numbered(n int)            { l.Consider(strconv.Itoa(n)) }
func (l *looking) Titled(s string)           { l.Consider(s) }
func (l *looking) Standing(s State)          { l.Consider(string(s)) }
func (l *looking) Carries(p domain.Priority) { l.Consider(string(p)) }
func (l *looking) Described(s string)        { l.Consider(s) }
func (l *looking) NotDescribed()             {}
func (l *looking) Labelled(s string)         { l.Consider(s) }
func (l *looking) NotLabelled()              {}
func (l *looking) LabelsTold()               {}
func (l *looking) Due(t time.Time)           { l.Consider(domain.FormatDay(t)) }
func (l *looking) NotDue()                   {}
func (l *looking) Estimated(d time.Duration) { l.Consider(domain.FormatDuration(d)) }
func (l *looking) NotEstimated()             {}

// Said is everything a search looks in, read the way a person reads an objective rather than the way it is stored.
func Said(o Objective) []string {
	l := &looking{}
	o.Tell(l)
	return l.said
}

// Matching says whether o carries every one of named, case-insensitive, wherever among its facts each one happens to sit.
func Matching(o Objective, named []string) bool {
	return domain.Reaches(Said(o), named)
}
