package store

// Where a task sits among the tasks it sits with, and how that place is
// changed. The place itself never leaves this package: a task does not carry
// it, the list does, and nothing above here has any use for the number.

import (
	"database/sql"
	"errors"
	"fmt"
)

// sisters is the group a task takes its place in: the tasks in the same column
// under the same parent. Written once because every statement below has to
// agree on it — a move that closed a gap in one group and opened one in
// another would leave both with a hole.
const sisters = `state = ? AND parent = ?`

// whereItSits reads back the group a task is in and the place it holds there.
func (s *Store) whereItSits(number int) (state string, parent, place int, err error) {
	row := s.on.QueryRow(
		`SELECT state, parent, position FROM tasks WHERE number = ?`, number)
	// Written with ifs rather than a tagless switch on purpose: a tagless
	// switch puts its conditions outside any covered block, so a mutation of
	// one of them reports as unreachable and the gate cannot judge it.
	err = row.Scan(&state, &parent, &place)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, 0, fmt.Errorf("task %d: %w", number, ErrNoSuchTask)
	}
	if err != nil {
		return "", 0, 0, fmt.Errorf("reading where task %d sits: %w", number, err)
	}
	return state, parent, place, nil
}

// MoveFirst puts a task at the head of the tasks it sits with.
func (s *Store) MoveFirst(number int) error {
	return s.place(number, func(_, _ int) int { return 1 })
}

// MoveLast puts a task behind all the tasks it sits with.
func (s *Store) MoveLast(number int) error {
	return s.place(number, func(_, among int) int { return among })
}

// MoveAfter puts a task directly behind another. Both are in the same group by
// the time this is called — that a task is moved among the ones it sits with is
// a rule of the notion, refused before the store is asked.
func (s *Store) MoveAfter(number, target int) error {
	_, _, behind, err := s.whereItSits(target)
	if err != nil {
		return err
	}
	return s.place(number, func(from, _ int) int {
		// Coming from in front, the target slides forward by one as the gap
		// closes, so landing on its place is landing behind it. Coming from
		// behind, the target does not move, so one past it is the place.
		//
		// The boundary is `<=` and not `<`: a task already where it is asked to
		// go has nowhere to travel, and reading that as "one past" would move
		// it off the place it holds.
		if from <= behind {
			return behind
		}
		return behind + 1
	})
}

// place moves a task to the place chosen among the tasks it sits with: the gap
// it leaves is closed and one is opened where it lands, so the places stay
// 1..N with no hole. The place is chosen by a function rather than passed,
// since each caller names it from something different — nothing, the size of
// the group, or where the task is coming from.
func (s *Store) place(number int, chosen func(from, among int) int) error {
	state, parent, from, err := s.whereItSits(number)
	if err != nil {
		return err
	}
	among, err := s.countSisters(state, parent)
	if err != nil {
		return err
	}
	// Nothing needs clamping. The places named are 1, the size of the group, or
	// one past the target — and one past the target only passes the group when
	// the target is last, which is a place a task cannot be travelling from.
	to := chosen(from, among)
	if to == from {
		return nil
	}
	if err := s.makeRoom(state, parent, from, to); err != nil {
		return err
	}
	if _, err := s.on.Exec(
		`UPDATE tasks SET position = ? WHERE number = ?`, to, number); err != nil {
		return fmt.Errorf("moving task %d: %w", number, err)
	}
	return nil
}

// countSisters says how many tasks sit in a group, which is the last place in
// it.
func (s *Store) countSisters(state string, parent int) (int, error) {
	var among int
	if err := s.on.QueryRow(
		`SELECT COUNT(*) FROM tasks WHERE `+sisters, state, parent).Scan(&among); err != nil {
		return 0, fmt.Errorf("counting what task sits with: %w", err)
	}
	return among, nil
}

// closeTheGapLeftIn pulls forward everything that sat behind the place a task
// has left, so the places there go on running 1..N.
func (s *Store) closeTheGapLeftIn(state string, parent, left int) error {
	if _, err := s.on.Exec(
		`UPDATE tasks SET position = position - 1
		 WHERE `+sisters+` AND position > ?`, state, parent, left); err != nil {
		return fmt.Errorf("closing the gap left behind: %w", err)
	}
	return nil
}

// makeRoom shifts by one everything lying between the place a task is leaving
// and the place it lands, so the gap it leaves is closed and one is opened
// where it goes.
//
// One statement either way, and the direction falls out of the two numbers
// inside it rather than being asked here. Asked here it could not be judged:
// place has already turned back when the two places are the same, so a
// comparison of them would have no boundary any test could stand on.
func (s *Store) makeRoom(state string, parent, from, to int) error {
	if _, err := s.on.Exec(
		`UPDATE tasks SET position = position + CASE WHEN ? < ? THEN -1 ELSE 1 END
		 WHERE `+sisters+` AND position BETWEEN ? AND ? AND position <> ?`,
		from, to, state, parent, min(from, to), max(from, to), from); err != nil {
		return fmt.Errorf("making room: %w", err)
	}
	return nil
}
