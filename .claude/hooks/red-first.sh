#!/usr/bin/env bash
# red-first.sh — refuse to write production Go until a test has been seen red.
#
# Claude Code runs this before every Write, Edit and Bash. It reads the hook's
# JSON on stdin and either says nothing (the write goes ahead) or answers with
# a refusal naming the package that owes a failing test.
#
# Why it watches Bash and not only Write|Edit: measured, on 2026-08-30, a whole
# package — internal/board, 4000 lines — was written through `cat > file <<'GO'`
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

payload=$(cat)
tool=$(printf '%s' "$payload" | jq -r '.tool_name // ""')

case $tool in
  Write|Edit)
    targets=$(printf '%s' "$payload" | jq -r '[.tool_input.file_path // empty] | .[]')
    ;;
  Bash)
    command=$(printf '%s' "$payload" | jq -r '.tool_input.command // ""')
    # A command that only reads is none of our business. `cat internal/x.go`
    # goes through; `cat > internal/x.go` does not.
    if ! printf '%s' "$command" | grep -Eq '>|\btee\b|\bsed +-i\b|\bperl +-[a-z]*i\b|\btruncate\b|\bdd\b.*\bof=|\b(python3?|ruby|node|osascript)\b'; then
      exit 0
    fi
    targets=$(printf '%s' "$command" | grep -oE '[A-Za-z0-9_./-]+\.go' || true)
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
