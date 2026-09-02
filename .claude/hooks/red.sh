#!/usr/bin/env bash
# red.sh — run a package's tests, and record the failure if there is one.
#
# This is the only thing that mints a receipt, and it mints one only when `go
# test` refused. A passing suite leaves nothing behind: what the receipt attests
# is not that a test exists, it is that the failure was seen.
#
# A build failure counts as red, and on purpose. In Go the first honest red for
# a function that does not exist yet is `undefined: Scrolled` — refusing it
# would make the gate unusable for exactly the case it is meant to serve.

set -uo pipefail

package=${1:-}
only=${2:-}

if [ -z "$package" ]; then
  echo "usage: make red PKG=./internal/board [RUN=TestSomething]" >&2
  exit 2
fi

root=${CLAUDE_PROJECT_DIR:-$PWD}
receipts="$root/.claude/red"
mkdir -p "$receipts"

where=${package#./}
where=${where#cairn/}
receipt="$receipts/${where//\//_}.json"

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
[ -n "$failing" ] || failing="(build failed)"

jq -n \
  --arg package "$where" \
  --arg failing "$failing" \
  --arg at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --arg output "$(printf '%s' "$output" | head -40)" \
  '{package: $package, failing: $failing, at: $at, output: $output}' > "$receipt"

printf '\xe2\x9c\x93 red recorded for %s: %s\n' "$where" "$failing"
echo "  You may now write the code that turns it green."
