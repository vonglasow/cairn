#!/usr/bin/env bash
# Claude Code hook: ask the cairn window for a permission, or the answers to
# Claude's questions, while one listens.
# cairn prints the answer for Claude Code, or nothing and the terminal asks.
# It never fails Claude Code: it always exits 0. See DECISIONS.md.

input=$(cat)
cd "$(jq -r '.cwd // empty' <<<"$input" 2>/dev/null)" 2>/dev/null || true
printf '%s' "$input" | cairn agent ask --request - 2>/dev/null
exit 0
