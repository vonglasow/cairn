# A Claude Code session tied to a cairn task

Read this before touching the Claude Code plugin, its hook or its skill, or
before changing what a session reads from a task or writes into it.

The first piece of a larger Claude Code plugin for cairn; the other pieces — putting a column back in order, bringing external
tickets in, the daily report — get their own designs.

## What it is for

A working session starts with the context of the task it serves, writes what
it learns where the next session will find it, and counts its time against
that task — without the person repeating any of it by hand, and without
cairn knowing anything about Claude.

## Words

"task" alone, in a conversation with Claude, is Claude's own to-do list. A
cairn task is said **"task cairn"** and written **`cairn#123`**, beside
`AB#123` for Azure Boards, `PROJ-123` for Jira and `owner/repo#123` for
GitHub, which keep their own notation. Nothing is renamed inside cairn.

## What it does

**Attaching.** The person says which task they work on — `cairn#123` — or asks
what to work on and picks one of the list (below). Claude puts the task in
focus through the cairn MCP server; the hook writes the session entry. Claude
then reads, from the task's notes directory (`task show` gives it):

- `summary.md` — where the task stands: decisions, state, next step;
- the three most recent dated entries, and the names of the older ones;
- for a sub-task, the parent's `summary.md`, and nothing else of the parent;
- `ticket.md`, when there is one.

A session that is a short aside attaches to nothing, and nothing is written.

**The external ticket is optional, and on demand.** Only when the person asks
does Claude read the ticket named by the task's labels — `ado`, and its id as a
label of digits alone (docs/labels.md) — through
whatever Azure DevOps access is there, and rewrite `ticket.md` from it. When
that fails, the copy already there is read and the failure is said. No ticket
tool is required for anything else to work. Azure DevOps only, for now.

**Choosing.** Asked what to work on, Claude lists the first five tasks of the
to-do and in-progress columns that are not blocked, in the board's order, each
with its parent. Claude ranks nothing: the board's order is the priority, and
changing it is the person's act. A passed deadline may be pointed out; it
moves nothing.

**Detaching.** When the subject changes or the person says they stop, Claude
updates `summary.md`, then releases the focus through the MCP server; the hook
closes the link. A session closed abruptly releases the focus through the
hook, and its summary is caught up at the next attachment.

**Idleness.** Ten minutes without a message release the focus; the link stays
open, and the next message takes the focus again.

**Several sessions, one task.** A session that ends or falls idle lets go of
its task only when no other session has a link open on it: the focus is held
per task, and the other session is still working.

**The focus released from a board is taken back.** A person who releases the
focus from the window or the terminal, while a session is linked to that task,
sees the session's next message take it again. The session says it works on
the task; to stop it, detach the session.

## Who does what

| what | done by |
|---|---|
| listing tasks, taking and releasing the focus on request, showing a task | cairn MCP server |
| reading and writing notes, the summary, the ticket on demand | Claude, through the skill |
| the session entry, which session holds which task, idleness, session end | the hook script — deterministic |

The trace already records every focus taken and released, with its time; it is
the source of the time spent and of the daily report. It does not record which
session acted, and recording it would tie cairn to Claude, so the link lives
in the notes instead.

## The link is the session entry

When the focus is taken through the MCP server, the hook writes
`<notes>/<task>/<YYYY-MM-DD>-session-<session-id>.md`:

```
cd <the session's working directory>
claude --resume <session-id>
```

The open entry of a session — one not ending in a `released <time>` line — is
the session's link; the directory it sits in is the task's number. One session,
one task: taking another lets go of the previous task and closes its link.
Asking what is in focus, a focus cairn refused, and letting go of some other
task leave the link alone. A bare unfocus — every task at once — is not the
session's to answer, and the hook does nothing about it. Nothing is stored
outside the notes but the idle waits, in the temporary directory.

## The hook

`plugin/claude/hooks/cairn-session.sh`, called by Claude Code on `PostToolUse`
(matching any tool ending in `__task_focus` or `__task_unfocus`),
`UserPromptSubmit`, `Stop` and `SessionEnd`:

```bash
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
```

The hook never stands in Claude Code's way: it prints nothing — a hook's
output on a message can reach the model's context — and exits 0 whatever cairn
answered, because a non-zero code on a message can block it, and the plugin is
on in every project. Only the idle delay changes when tested:
`CAIRN_IDLE_SECONDS`, six hundred when unset.

What it rests on, each tried with `claude -p` on 2026-10-08 rather than taken
from the documentation, which is silent on most of it:

- **The tool is matched by its suffix.** The cairn server is named
  `mcp__cairn__…` when the person declared it, and `mcp__plugin_cairn_cairn__…`
  when the plugin did; a matcher on one name misses the other.
- **The idle wait outlives the hook.** Claude Code does not wait for a child
  sent to the background with its output closed, and the child survives the
  session's end.
- **The session's end fits.** Claude Code gives every `SessionEnd` hook 1.5
  seconds between them; this one takes under a tenth of one.
- **Idleness is a wait a message calls off, not a count of minutes.** Each
  answer leaves a file in the temporary directory and a wait on it; a message
  removes the session's files, and a wait that finds its file gone does
  nothing. `find -mmin` was the first measure and held at ten minutes, but at
  zero macOS and GNU disagree, so no test could reach it; comparing times was
  the second, and macOS's bash 3.2 compares them to the second, so a message
  in the same second as the answer went unseen.
- **A resumed session keeps its id**, and `SessionStart` says `resume`: the
  link is found again.

## The skill

One skill, `cairn`, whose description carries the words above so that
"task cairn" and `cairn#123` call it. It holds the three gestures — attaching,
choosing, detaching — and the rule of words. Nothing mechanical is in it.

## Where it lives, and how it is installed

```
.claude-plugin/marketplace.json
plugin/claude/
  plugin.go                     the plugin, carried inside the binary
  .claude-plugin/plugin.json
  .mcp.json                     cairn mcp serve --access full
  skills/cairn/SKILL.md
  hooks/hooks.json
  hooks/cairn-session.sh
```

The plugin is independent of cairn's code: it reaches cairn only through its
command line and its MCP server. Needed: `cairn` and `jq` on the `PATH`.
Nothing else.

**The binary carries the plugin.** A person handed only the binary — by brew,
or as a file — has no repository to add as a marketplace. So `plugin.go`, in
the plugin's own folder, embeds its files, and `cairn plugin write [--dir D]`
writes them out as a marketplace: `D/.claude-plugin/marketplace.json`, read
off the plugin's manifest rather than written a second time, and the plugin
under `D/cairn`, its hook executable. `D` is `claude-plugin` beside the store
unless named. Then `/plugin marketplace add D` and `/plugin install
cairn@cairn`, which the command prints. The plugin installed is always the one
the cairn it calls was built with, and says so: the manifest written, and the
marketplace's entry for it, carry the version of the cairn that wrote them —
Claude Code reads it to tell an update from what it has, and warned of a
plugin with none.

Writing again replaces the plugin written before, so nothing an older cairn
carried is left behind; and so a folder holding anything else — a home named by
mistake — is refused rather than cleared (`ErrNotItsFolder`). The repository
stays a marketplace of its own, for a clone: `/plugin marketplace add <the
cairn repository>`. `bdd/claude-plugin.feature` holds both, against the real
binary, file by file against this folder; `claude plugin validate` accepts
what it writes.

The plugin declares the cairn MCP server, so that installing it is enough. A
person who already declared one keeps theirs: Claude Code drops a plugin's
server when one with the very same command is already there, and loads both
when the commands differ — tools twice over, harmless to the hook, noisy for
the model. `INSTALL.md` says to declare it with the plugin's command, or to
remove it.

## How it is held

- `bdd/claude-session.feature`: each scenario hands the hook the JSON Claude
  Code hands it, against the real cairn binary on a store and notes of its own,
  with the idle delay at one second. `test/claude` runs it, two levels under
  the root like every package holding scenarios, and in none of cairn's.
- `test/claude/plugin_test.go` holds the wiring no scenario reaches: the four
  events call a script that is there and runs, the matcher hears the focus
  from either server and nothing else, cairn answers to the server the plugin
  declares, and the marketplace points at the plugin.
- `shellcheck` runs at every commit on every shell script (docs/checks.md).
- Tried for real with `claude -p --plugin-dir plugin/claude`: the skill was
  called by "on travaille sur cairn#1", the focus taken, the entry written with
  the session's directory and id, the summary read, and the session's end
  released the focus and closed the link.

## Not in this piece

Putting a column back in order on request, bringing external tickets in, the
daily report, ticket systems other than Azure DevOps, a setting that fetches
the ticket on every attachment.
