#!/usr/bin/env bash
# red-first.sh — refuse to write production Go until a test has been seen red.
#
# Claude Code runs this before every Write, Edit and Bash. It reads the hook's
# JSON on stdin and either says nothing (the write goes ahead) or answers with
# a refusal naming the package that owes a failing test.
#
# Why it watches Bash and not only Write|Edit: measured, on 2026-08-30, a whole
# package — internal/tui, 4000 lines — was written through `cat > file <<'GO'`
# and through python scripts. A gate on Write|Edit alone would have watched it
# happen and said nothing.
#
# Why internal/ only: cmd/ holds `main`, which no test can call without ending
# the process that runs it, and test/ is the harness itself. Gating either would
# block work no test can precede.

set -uo pipefail

root=${CLAUDE_PROJECT_DIR:-$PWD}
receipts="$root/.claude/red"
# How long a red stays good. Long enough to write the code it asked for, short
# enough that yesterday's failure cannot authorise today's.
keeps=1800

refuse() {
  jq -nc --arg why "$1" '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: $why
    }
  }'
  exit 0
}

# Without jq the payload cannot be read, and a gate that cannot read what it
# guards has to close rather than open: before this, a missing jq left `tool`
# empty and the catch-all below let every write through. The refusal is
# written by hand because the usual one needs jq to be built.
if ! command -v jq >/dev/null; then
  printf '%s\n' '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"red-first.sh needs jq to read what is about to be written, and jq is not installed."}}'
  exit 0
fi

payload=$(cat)
tool=$(printf '%s' "$payload" | jq -r '.tool_name // ""')

case $tool in
  Write|Edit)
    targets=$(printf '%s' "$payload" | jq -r '[.tool_input.file_path // empty] | .[]')
    ;;
  Bash)
    command=$(printf '%s' "$payload" | jq -r '.tool_input.command // ""')
    # What is looked for is the *target* of a write, not the presence of a
    # chevron. Measured: matching any `>` refused a plain grep, because `2>&1`
    # carries one and names no file at all — and it refused this very script's
    # own repair, because the repair mentioned a path in a comment.
    #
    # A command that only reads is none of our business either: reading a
    # production file goes through, redirecting into one does not.
    targets=$(
      {
        # `> path` and `>> path`. `2>&1` sends one stream to another and stops
        # at the ampersand, so it yields nothing here.
        printf '%s' "$command" |
          grep -oE '>>?[[:space:]]*[^&[:space:];|)]+' |
          sed -E 's/^>>?[[:space:]]*//'
        # The tools that write in place name their target among the arguments,
        # so for those the whole command is read for paths. This is the clause
        # that stays a guess: a path a script builds at runtime is invisible.
        if printf '%s' "$command" |
          grep -Eq '\btee\b|\bsed +-i\b|\bperl +-[a-z]*i\b|\btruncate\b|\b(python3?|ruby|node|osascript)\b|\bgit +(checkout|restore)\b'; then
          printf '%s' "$command" | grep -oE '[A-Za-z0-9_./-]+[.]go'
        fi
        # The tools that copy or move write to their last argument and only
        # read the others, so only that one is a target: copying a production
        # file aside to keep it is a read. Measured as a hole before this: a
        # file could be written in /tmp and copied into place unseen.
        printf '%s' "$command" | tr ';&|' '\n\n\n' |
          awk '$1 ~ /^(cp|mv|install|rsync|ditto)$/ { print $NF }'
        # A patch names what it changes inside itself, on its +++ lines.
        for patch in $(printf '%s' "$command" | grep -oE '[A-Za-z0-9_./-]+[.](patch|diff)'); do
          [ -f "$patch" ] && sed -nE 's|^\+\+\+ (b/)?([^[:space:]]+).*|\2|p' "$patch"
        done
      } | grep -E '[.]go$' || true
    )
    ;;
  *)
    exit 0
    ;;
esac

owed=""
for target in $targets; do
  # Relative to the repository, whatever shape the caller wrote it in.
  where=${target#"$root"/}
  where=${where#./}
  case $where in
    internal/*) ;;
    *) continue ;;
  esac
  case $where in
    *_test.go) continue ;;
    *.go) ;;
    *) continue ;;
  esac

  package=$(dirname "$where")
  receipt="$receipts/${package//\//_}.json"

  if [ ! -f "$receipt" ]; then
    owed="$owed $package"
    continue
  fi
  age=$(( $(date +%s) - $(stat -f %m "$receipt" 2>/dev/null || stat -c %Y "$receipt") ))
  if [ "$age" -gt "$keeps" ]; then
    owed="$owed $package"
  fi
done

[ -n "$owed" ] || exit 0

first=$(printf '%s' "$owed" | tr ' ' '\n' | grep -v '^$' | head -1)
refuse "$(printf 'Writing production Go in %s is refused: no test in that package has been seen red.

CLAUDE.md: "Scenario, then test, then code. Never the reverse."

Write the test that states what is missing, watch it fail, and record the failure:

    make red PKG=./%s

A test that passes records nothing — the point is that the failure existed.' "$first" "$first")"
