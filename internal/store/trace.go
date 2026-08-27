package store

// What was done to a task, kept. The table is appended to and never rewritten,
// and an act names its task by number rather than referring to it: a task may
// be gone by the time its acts are read, and the numbers are never handed out
// twice, so a number goes on naming the one task it always named.

import (
	"fmt"
	"time"

	"cairn/internal/trace"
)

// traceColumns is the reading order every query below hands to scanAct, written
// once so recording and reading cannot drift apart.
const traceColumns = `task_number, did, field, was, became, through, at`

// boundAct holds an act's facts in the shape SQL will bind. An act tells them
// here rather than handing over its fields, which is what lets this keep
// working the day those fields change.
type boundAct struct {
	task                             int
	did, field, was, became, through string
	at                               time.Time
}

func (b *boundAct) On(n int)               { b.task = n }
func (b *boundAct) Did(what string)        { b.did = what }
func (b *boundAct) Changing(field string)  { b.field = field }
func (b *boundAct) NothingNamed()          { b.field = "" }
func (b *boundAct) Was(value string)       { b.was = value }
func (b *boundAct) WasNothing()            { b.was = "" }
func (b *boundAct) Became(value string)    { b.became = value }
func (b *boundAct) BecameNothing()         { b.became = "" }
func (b *boundAct) Through(surface string) { b.through = surface }
func (b *boundAct) ByHand()                { b.through = "" }
func (b *boundAct) At(when time.Time)      { b.at = when }

// Record writes down one act. Nothing is ever updated here: a history that
// could be rewritten would not be one.
func (s *Store) Record(a trace.Act) error {
	var b boundAct
	a.Tell(&b)
	if _, err := s.db.Exec(
		`INSERT INTO trace (`+traceColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		b.task, b.did, b.field, b.was, b.became, b.through,
		b.at.UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("writing down what happened to task %d: %w", b.task, err)
	}
	return nil
}

// ActsOn hands back everything done to a task, oldest first. A task that never
// existed and one nothing was ever done to answer the same way, since a task
// always carries at least its creation: the caller asks about the task itself
// when it needs to tell them apart.
func (s *Store) ActsOn(number int) ([]trace.Act, error) {
	rows, err := s.db.Query(
		`SELECT `+traceColumns+` FROM trace WHERE task_number = ? ORDER BY id`, number)
	if err != nil {
		return nil, fmt.Errorf("reading what happened to task %d: %w", number, err)
	}
	defer func() { _ = rows.Close() }()

	acts := []trace.Act{}
	for rows.Next() {
		one, err := scanAct(rows)
		if err != nil {
			return nil, fmt.Errorf("reading an act: %w", err)
		}
		acts = append(acts, one)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading what happened to task %d: %w", number, err)
	}
	return acts, nil
}

// ActsSince hands back everything done from a moment onward, across every task,
// oldest first. A task deleted since keeps its acts, so a period can name a task
// that is no longer there — which is the whole reason a history outlives what it
// describes.
func (s *Store) ActsSince(from time.Time) ([]trace.Act, error) {
	rows, err := s.db.Query(
		`SELECT `+traceColumns+` FROM trace WHERE at >= ? ORDER BY id`,
		from.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("reading what happened since %s: %w", from, err)
	}
	defer func() { _ = rows.Close() }()

	acts := []trace.Act{}
	for rows.Next() {
		one, err := scanAct(rows)
		if err != nil {
			return nil, fmt.Errorf("reading an act: %w", err)
		}
		acts = append(acts, one)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading what happened since %s: %w", from, err)
	}
	return acts, nil
}

// scanAct reads one act by scanning a row into plain variables and telling them
// to a builder, rather than into an act's fields directly — the membrane Tell
// uses, the other way round.
func scanAct(row scanner) (trace.Act, error) {
	var (
		task                             int
		did, field, was, became, through string
		at                               string
	)
	if err := row.Scan(&task, &did, &field, &was, &became, &through, &at); err != nil {
		return trace.Act{}, err
	}
	when, err := time.Parse(time.RFC3339, at)
	if err != nil {
		return trace.Act{}, fmt.Errorf("the moment %q cannot be read: %w", at, err)
	}

	var a trace.AsTold
	a.On(task)
	a.Did(did)
	trace.TellOrNot(field, a.Changing, a.NothingNamed)
	trace.TellOrNot(was, a.Was, a.WasNothing)
	trace.TellOrNot(became, a.Became, a.BecameNothing)
	trace.TellOrNot(through, a.Through, a.ByHand)
	a.At(when)
	return a.Act(), nil
}
