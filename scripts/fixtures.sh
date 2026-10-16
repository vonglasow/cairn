#!/usr/bin/env bash
# fixtures.sh — a store full of believable work, to look at a board with.
#
#   scripts/fixtures.sh [folder]
#
# Makes a store under folder (/tmp/cairn-fixtures when none is named) holding
# objectives in every standing and tasks in every column, steps, blockers, a
# deadline past and one to come, two tasks in focus and two archived — then
# opens the window on it. Run again, it starts from nothing.
#
#   FIXTURES_TRACE=1 scripts/fixtures.sh     (make trace)
#
# opens it traced, for a bug seen only on one system, such as a drag drawn in
# the wrong place on Windows. cairn is built from the tree as it stands — a
# binary from before the trace learned something cannot write it down — and
# the record, under traces/, starts with what the machine is: its system,
# display session, input method, keyboard, language.
# Do the one thing that goes wrong, close the window, send the file.
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

# Days written relative to today, so "a deadline past and one to come" stays
# true whichever day the fixtures are made: fixed dates went stale the day
# after they were written. BSD date spells it -v, GNU date -d.
day() { date -v"$1"d +%F 2>/dev/null || date -d "$1 days" +%F; }

cairn objective new "Ship cairn 1.0" --state active --priority high --deadline "$(day +71)" --label product
cairn objective new "Make the store dependable" --state active --priority normal --label technical
cairn objective new "Redesign the board" --state on-hold --priority low
cairn objective new "Move to Postgres" --state reached --priority normal
cairn objective new "Port to mobile" --state dropped --priority low

cairn task new "Write the packaging spec" --state current --priority high --deadline "$(day -1)" --estimate 8h --spent 3h --label backend --label sprint-172 --objective 1 --desc "Three ADRs to read before starting."
cairn step new 1 "Read ADR 0014"
cairn step new 1 "Settle the multi-chart question" --done
cairn step new 1 "Write the Flows section"

cairn task new "Rework the migrations" --state current --priority critical --deadline "$(day -7)" --estimate 2h --spent 9h --label store --objective 2
cairn step new 2 "Back up before rolling back" --done
cairn step new 2 "Test a rollback on an empty store"

cairn task new "The client's answer on quotas" --state blocked --because "waiting on the client" --priority high --label product --objective 1
cairn task new "Review PR 31745" --state watching --blocker 1 --priority normal --label backend
cairn task new "Decide on naming" --state paused --because "team call" --priority low

cairn task new "Clean the lockfiles" --state todo --priority low --label technical --objective 2
cairn task new "Update the README" --state todo --deadline "$(day +1)" --label doc
cairn task new "Encrypt the backups" --state backlog --priority normal --objective 2
cairn task new "Draw the columns" --state backlog --priority low --objective 3
cairn task new "Benchmark the store" --state backlog --label technical

cairn task new "Switch to Postgres 16" --state "done" --priority high --label store --objective 4
cairn task new "Drop the old schema" --state "done" --label store --objective 4
cairn task new "iOS prototype" --state "done" --priority low --objective 5

cairn task archive 12
cairn task archive 13

cairn task focus 1
cairn task focus 2

if [ -z "${FIXTURES_TRACE:-}" ]; then
  exec go run "$here/cmd/cairn" gui
fi

mkdir -p "$here/traces"
log="$here/traces/trace-$(date +%Y%m%d-%H%M%S).log"
bin="$store/bin/cairn"
case "$(uname -s)" in MINGW* | MSYS* | CYGWIN*) bin="$bin.exe" ;; esac
(cd "$here" && go build -o "$bin" ./cmd/cairn)

# Each fact on a line of its own, and a missing tool says so rather than
# stopping the record: the machine that misbehaves is rarely the one every
# tool is installed on.
fact() { printf '# %-16s %s\n' "$1" "$2"; }
ran() { "$@" 2>/dev/null | tr '\n' ' ' | sed 's/  */ /g' || true; }
{
  fact "recorded" "$(date '+%F %T %z')"
  fact "commit" "$(cd "$here" && git describe --tags --always --dirty 2>/dev/null || echo unknown)"
  fact "cairn" "$(ran "$bin" --version)"
  fact "go" "$(go version)"
  fact "system" "$(uname -srm)"
  case "$(uname -s)" in
    Darwin) fact "macos" "$(ran sw_vers -productVersion)" ;;
    Linux)
      fact "distribution" "$(sed -n 's/^PRETTY_NAME=//p' /etc/os-release 2>/dev/null | tr -d '"')"
      fact "session" "${XDG_SESSION_TYPE:-unset}"
      fact "desktop" "${XDG_CURRENT_DESKTOP:-unset}"
      fact "wayland" "${WAYLAND_DISPLAY:-unset}"
      fact "x11" "${DISPLAY:-unset}"
      fact "gtk im" "${GTK_IM_MODULE:-unset}"
      fact "qt im" "${QT_IM_MODULE:-unset}"
      fact "xmodifiers" "${XMODIFIERS:-unset}"
      fact "ibus" "$(pgrep -x ibus-daemon >/dev/null && echo running || echo 'not running')"
      fact "fcitx" "$(pgrep -f fcitx >/dev/null && echo running || echo 'not running')"
      fact "keyboard" "$(ran setxkbmap -query)$(ran localectl status)"
      ;;
    MINGW* | MSYS* | CYGWIN*) fact "windows" "$(ran cmd //c ver)" ;;
  esac
  fact "language" "${LC_ALL:-${LANG:-unset}}"
} >"$log"

cat <<SAID

The window opens on a board of believable work, and writes down every pointer
and keyboard event. Do the one thing that goes wrong, then close the window:

  a drag drawn in the wrong place   press a card, move it a few centimetres
                                    with the button held, let go

SAID

CAIRN_GUI_TRACE=1 "$bin" gui 2>>"$log" || true

echo
echo "the record is $log — send that file"
