#!/usr/bin/env bash
# fixtures.sh — a store full of believable work, to look at a board with.
#
#   scripts/fixtures.sh [folder]
#
# Makes a store under folder (/tmp/cairn-fixtures when none is named) holding
# objectives in every standing and tasks in every column, parts, blockers, a
# deadline past and one to come, two tasks in focus and two off the board —
# then opens the window on it. Run again, it starts from nothing.
#
# It only ever deletes a folder it made itself: one it made carries a mark,
# and a folder without it is refused rather than emptied — a typo in the
# argument must not be able to take a home directory with it.

set -euo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
store=${1:-/tmp/cairn-fixtures}
mark=.made-by-cairn-fixtures

if [ -e "$store" ]; then
  if [ ! -f "$store/$mark" ]; then
    echo "fixtures.sh: $store is there and was not made by this script; it is left alone" >&2
    exit 1
  fi
  rm -rf "$store"
fi
mkdir -p "$store"
touch "$store/$mark"

export CAIRN_STORE="$store/cairn.db"
export CAIRN_NOTES="$store/notes"

cairn() { go run "$here/cmd/cairn" "$@" >/dev/null; }

cairn objective add "Ship cairn 1.0" --state active --priority high --deadline 2026-12-15 --label product
cairn objective add "Make the store dependable" --state active --priority normal --label technical
cairn objective add "Redesign the board" --state on-hold --priority low
cairn objective add "Move to Postgres" --state reached --priority normal
cairn objective add "Port to mobile" --state dropped --priority low

cairn task add "Write the packaging spec" --state current --priority high --deadline 2026-10-04 --estimate 8h --spent 3h --label backend --label sprint-172 --objective 1 --desc "Three ADRs to read before starting."
cairn task add "Read ADR 0014" --parent 1 --state todo --estimate 2h --spent 1h
cairn task add "Settle the multi-chart question" --parent 1 --state done --estimate 1h --spent 4h
cairn task add "Write the Flows section" --parent 1 --state todo

cairn task add "Rework the migrations" --state current --priority critical --deadline 2026-09-28 --estimate 2h --spent 9h --label store --objective 2
cairn task add "Back up before rolling back" --parent 5 --state done
cairn task add "Test a rollback on an empty store" --parent 5 --state todo

cairn task add "The client's answer on quotas" --state blocked --because "waiting on the client" --priority high --label product --objective 1
cairn task add "Review PR 31745" --state watching --blocker 1 --priority normal --label backend
cairn task add "Decide on naming" --state paused --because "team call" --priority low

cairn task add "Clean the lockfiles" --state todo --priority low --label technical --objective 2
cairn task add "Update the README" --state todo --deadline 2026-10-06 --label doc
cairn task add "Encrypt the backups" --state backlog --priority normal --objective 2
cairn task add "Draw the columns" --state backlog --priority low --objective 3
cairn task add "Benchmark the store" --state backlog --label technical

cairn task add "Switch to Postgres 16" --state done --priority high --label store --objective 4
cairn task add "Drop the old schema" --state done --label store --objective 4
cairn task add "iOS prototype" --state done --priority low --objective 5

cairn task archive 17
cairn task archive 18

cairn task focus 1
cairn task focus 5

go run "$here/cmd/cairn" gui
