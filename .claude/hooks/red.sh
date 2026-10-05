#!/usr/bin/env bash
# red.sh — run a package's tests, and record the failure if there is one.
#
# This is the only thing that mints a receipt, and it mints one only when `go
# test` refused. A passing suite leaves nothing behind: what the receipt attests
# is not that a test exists, it is that the failure was seen.
#
# A build failure counts as red, and on purpose. In Go the first honest red for
# a function that does not exist yet is `undefined: Scrolled` — refusing it
# would make the gate unusable for exactly the case it is meant to serve. But
# only a build failure a test caused: the error has to sit in a _test.go file.
# Measured, before that line was drawn: a package that did not exist, a -run
# pattern that was not a valid expression, and production code that did not
# compile all minted a receipt, and none of them was a test seen red.

set -uo pipefail

package=${1:-}
only=${2:-}
moving=${3:-}

if [ -z "$package" ]; then
  echo "usage: make red PKG=./internal/tui [RUN=TestSomething]" >&2
  echo "       make red PKG=./internal/tui MOVING=\"what moves, and where to\"" >&2
  exit 2
fi

command -v jq >/dev/null || { echo "red.sh needs jq, and jq is not installed" >&2; exit 2; }

root=${CLAUDE_PROJECT_DIR:-$PWD}
receipts="$root/.claude/red"
mkdir -p "$receipts"

where=${package#./}
where=${where#cairn/}
receipt="$receipts/${where//\//_}.json"

# A package that is not there has nothing to fail, so go test failing on it
# says nothing about a test.
if [ ! -d "$root/$where" ]; then
  printf '\xe2\x9c\x97 %s is not a package directory. There is nothing red to record.\n' "$package" >&2
  exit 2
fi

# Code that moves out of a package has no absent behaviour to write a test for,
# and a suite that still passes mints nothing — so without this, the gate
# refuses the one kind of change it has nothing to say about.
#
# It is a door and not a hole: it demands a sentence saying what moves and
# where to, that sentence is kept in the receipt beside the timestamp, and the
# receipt says `moving` rather than naming a failing test. Anybody reading the
# receipts can see which writes were let through this way and judge them.
if [ -n "$moving" ]; then
  jq -n \
    --arg package "$where" \
    --arg moving "$moving" \
    --arg at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    '{package: $package, moving: $moving, at: $at}' > "$receipt"
  printf '\xe2\x86\xb3 moving, recorded for %s: %s\n' "$where" "$moving"
  echo "  No test was run and none was owed: nothing new is being asked for."
  echo "  What is owed is that the suite still passes when the move is done."
  exit 0
fi

echo "→ go test ${only:+-run $only }$package"
if [ -n "$only" ]; then
  output=$(go test -count=1 -run "$only" "$package" 2>&1)
else
  output=$(go test -count=1 "$package" 2>&1)
fi
status=$?

if [ $status -eq 0 ]; then
  printf '\xe2\x9c\x97 %s passes. There is nothing red to record.\n' "$package"
  echo "  Write the test that states what is missing, then run this again."
  exit 1
fi

failing=$(printf '%s' "$output" | grep -E '^\s*--- FAIL: ' | sed -E 's/^\s*--- FAIL: ([^ ]+).*/\1/' | paste -sd, -)
if [ -z "$failing" ]; then
  # No test ran and failed, so the only red that counts is a test that does
  # not compile yet — its error located in a _test.go file of this package.
  if ! printf '%s' "$output" | grep -qE '_test\.go:[0-9]+(:[0-9]+)?: '; then
    printf '\xe2\x9c\x97 %s failed, but no test failed and no test failed to build:\n' "$package"
    printf '%s\n' "$output" | head -20
    echo "  Nothing was recorded. Fix what is broken, then write the test that is red."
    exit 1
  fi
  failing="(a test does not build yet)"
fi

jq -n \
  --arg package "$where" \
  --arg failing "$failing" \
  --arg at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --arg output "$(printf '%s' "$output" | head -40)" \
  '{package: $package, failing: $failing, at: $at, output: $output}' > "$receipt"

printf '\xe2\x9c\x93 red recorded for %s: %s\n' "$where" "$failing"
echo "  You may now write the code that turns it green."
