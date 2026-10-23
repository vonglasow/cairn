package engine

import (
	"fmt"
	"slices"
	"time"

	"cairn/internal/config"
	"cairn/internal/domain"
	"cairn/internal/trace"
)

// NowVariable is the variable that names the moment cairn takes for now when it writes an act and measures a focus session, for a script or a scenario to stand later.
const NowVariable = "CAIRN_NOW"

func (h Home) now() time.Time {
	if named, ok := h.lookup()(NowVariable); ok {
		if moment, err := time.Parse(time.RFC3339, named); err == nil {
			return moment
		}
	}
	return time.Now()
}

// FocusSession is how long a task has been in focus since the focus was taken or last renewed, and, once past the length the configuration gives a session, the question it asks.
type FocusSession struct {
	Task  int       `json:"task"`
	Since time.Time `json:"since" jsonschema:"when the focus was taken, or last renewed"`
	Up    string    `json:"up" jsonschema:"how long the session has been up, as 1h30m; empty under a minute"`
	Ask   string    `json:"ask,omitempty" jsonschema:"once the session is past its length: whether to renew it or end it"`
}

func sessionOf(number int, acts []trace.Act, now time.Time, length time.Duration) FocusSession {
	var since time.Time
	for _, a := range acts {
		c := a.Carried()
		if c.Task == number && (c.Did == trace.FocusTaken || c.Did == trace.FocusRenewed) && c.At.After(since) {
			since = c.At
		}
	}
	up := now.Sub(since).Truncate(time.Minute)
	session := FocusSession{Task: number, Since: since, Up: domain.FormatDuration(up)}
	if length > 0 && up >= length {
		session.Ask = fmt.Sprintf("this focus session has been up %s: renew it by focusing task %d again, or end it by unfocusing task %d",
			session.Up, number, number)
	}
	return session
}

func (e *Engine) sessionLength() (time.Duration, error) {
	if e.home.Dir == "" {
		return 0, nil
	}
	read, err := config.Read(e.home.Dir, e.home.lookup())
	if err != nil {
		return 0, nil
	}
	said := read.FocusSession()
	if said == "" {
		return 0, nil
	}
	length, err := time.ParseDuration(said)
	if err != nil || length <= 0 {
		return 0, fmt.Errorf("focus.session %q: %w", said, config.ErrFocus)
	}
	return length, nil
}

func (e *Engine) sessionsOf(held []int, length time.Duration) ([]FocusSession, error) {
	now := e.home.now()
	out := make([]FocusSession, 0, len(held))
	for _, number := range held {
		acts, err := e.history.ActsOn(number)
		if err != nil {
			return nil, err
		}
		out = append(out, sessionOf(number, acts, now, length))
	}
	return out, nil
}

func (e *Engine) sessions() ([]FocusSession, error) {
	length, err := e.sessionLength()
	if err != nil {
		return nil, err
	}
	held, err := e.places.Focus()
	if err != nil {
		return nil, err
	}
	return e.sessionsOf(held, length)
}

func (e *Engine) asked() ([]string, error) {
	length, err := e.sessionLength()
	if err != nil || length == 0 {
		return nil, err
	}
	held, err := e.places.Focus()
	if err != nil {
		return nil, err
	}
	sessions, err := e.sessionsOf(held, length)
	if err != nil {
		return nil, err
	}
	var asks []string
	for _, one := range sessions {
		if one.Ask != "" {
			asks = append(asks, one.Ask)
		}
	}
	return asks, nil
}

func (e *Engine) renewed(number int) error {
	length, err := e.sessionLength()
	if err != nil || length == 0 {
		return err
	}
	sessions, err := e.sessionsOf([]int{number}, length)
	if err != nil || slices.IndexFunc(sessions, func(s FocusSession) bool { return s.Ask != "" }) < 0 {
		return err
	}
	return e.record(number, trace.FocusRenewed, "", "", "")
}
