#!/usr/bin/env bash
# Claude Code hook: tie this session to its cairn task through a session entry
# in the task's notes, and hold or release that task's focus. It never blocks
# Claude Code: it prints nothing and always exits 0. See docs/claude-code.md.

idle=${CAIRN_IDLE_SECONDS:-600}
input=$(cat)
field() { jq -r "$1 // empty" <<<"$input"; }
session=$(field .session_id)
notes=$(cairn config show --json 2>/dev/null | jq -r .notes.at)
waits="${TMPDIR:-/tmp}/cairn-idle-$session"

open() { grep -L '^released' "$@" 2>/dev/null; }
entry=$(open "$notes"/*/*-session-"$session".md | head -1)
task=""
[ -z "$entry" ] || task=$(basename "$(dirname "$entry")")

close_link() { echo "released $(date -u +%FT%TZ)" >>"$entry"; }
let_go() {
  # Another session still working on the task keeps it in focus.
  open "$notes/$task"/*-session-*.md | grep -qv -- "-session-$session.md" ||
    cairn task unfocus "$task" >/dev/null 2>&1
}

case $(field .hook_event_name) in
  PostToolUse)
    n=$(field .tool_input.number)
    case $(field .tool_name) in
      *task_focus)
        [ "$n" != "$task" ] || exit 0
        shown=$(cairn task show "$n" --json 2>/dev/null) || exit 0
        [ -z "$entry" ] || { let_go; close_link; }
        dir=$(jq -r .context <<<"$shown")
        mkdir -p "$dir"
        printf 'cd %s\nclaude --resume %s\n' "$(field .cwd)" "$session" \
          >"$dir/$(date +%F)-session-$session.md" ;;
      *task_unfocus)
        [ -n "$n" ] && [ "$n" = "$task" ] && close_link ;;
    esac ;;
  UserPromptSubmit)
    # A message calls off every wait the session's answers started.
    rm -f "$waits".*
    [ -z "$task" ] || cairn task focus "$task" >/dev/null 2>&1 ;;
  Stop)
    [ -z "$task" ] || {
      wait=$(mktemp "$waits.XXXXXX")
      (sleep "$idle"; [ ! -e "$wait" ] || { let_go; rm -f "$wait"; }) </dev/null &>/dev/null &
    } ;;
  SessionEnd)
    [ -z "$task" ] || { let_go; close_link; } ;;
esac
exit 0
