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
- its steps, which `task show` lists with it — what is ticked and what is left;
- `ticket.md`, when there is one.

A session that is a short aside attaches to nothing, and nothing is written.

**The external ticket is optional, and on demand.** Only when the person asks
does Claude read the ticket named by the task's labels — `ado`, and its id as a
label of digits alone (docs/labels.md) — through
whatever Azure DevOps access is there, and rewrite `ticket.md` from it. When
that fails, the copy already there is read and the failure is said. No ticket
tool is required for anything else to work. Azure DevOps only, for now.

**Choosing.** Asked what to work on, Claude lists the first five tasks of the
to-do and in-progress columns that are not blocked, in the board's order —
one listing per column, since a listing takes one, each asking only for the
number, the title, what blocks it and the deadline. Claude ranks nothing: the board's order is the priority, and
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
| reading back a day or a week, the sessions no task names, where things stand | Claude, through the `recap` and `status` skills |
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
Asking what is in focus, a focus cairn refused, and unfocusing some other
task leave the link alone. A bare unfocus — every task at once — is not the
session's to answer, and the hook does nothing about it. Nothing is stored
outside the notes but the idle waits, in the temporary directory.

## The hook

`plugin/claude/hooks/cairn-session.sh`, called by Claude Code on
`SessionStart`, `PostToolUse` (matching any tool ending in `__task_focus` or
`__task_unfocus`), `UserPromptSubmit`, `Stop` and `SessionEnd`:

```bash
#!/usr/bin/env bash
# Claude Code hook: tie this session to its cairn task through a session entry
# in the task's notes, and hold or release that task's focus. It never blocks
# Claude Code: it prints nothing but, at a session's start, which workspace
# cairn answers from, and always exits 0. See docs/claude-code.md.

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
  SessionStart)
    cairn workspace show --json 2>/dev/null | jq -r 'select(.why)
      | (if .workspace then "the workspace \(.workspace) (\(.folder))"
       else "the store \(.store), which no workspace names" end) as $where
      | "cairn: the cairn MCP server of this session answers from \($where), because \(.why.rung). "
      + "It keeps it while the session lasts; workspace_show says it again. "
      + "Switching is for the person: another workspace is another server."' ;;
  SessionEnd)
    [ -z "$task" ] || { let_go; close_link; } ;;
esac
exit 0
```

The hook never stands in Claude Code's way: it prints nothing but the
workspace at a session's start (below) — a hook's output on a message can
reach the model's context — and exits 0 whatever cairn
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

**A session's start names the workspace.** At `SessionStart` — and only
there, since what `SessionStart` prints is what Claude Code hands the model as
context — the hook prints one line: the workspace cairn answers from, its
folder and why, or the store no workspace names; that the server keeps it
while the session lasts; that `workspace_show` says it again; that switching
is the person's (`@claude-session-start-says-the-workspace`,
`@claude-session-start-says-a-store-no-workspace-names`). The agent knows
where it writes before its first write. The hook asks `cairn workspace show`
in the session's own environment, which is the plugin's server's: both are
started by Claude Code from the same environment, at the same moment, and the
server keeps what it found then ([mcp.md](mcp.md), "It keeps the store it
started on"). Nothing on `cairn` failing — a refusal answered as JSON too,
which the line keeps out by asking for its `why` — no line is printed, and the
session starts as before
(`@claude-session-start-says-nothing-when-the-environment-names-no-workspace`).

## The skills

The skill `cairn`, whose description carries the words above so that
"task cairn" and `cairn#123` call it. It holds the three gestures — attaching,
choosing, detaching — the rule of words, and one rule about a task done: it is
not reopened, a task linked to it is opened instead (see
[links.md](links.md)). Nothing mechanical is in it.

What it says beyond those came from real sessions, each a gap an agent filled
on its own and filled differently every time:

- **several tasks in one session** — each focused while its work runs, a
  subagent's task held by the session that sends it, each closed before the
  next; it used to assume one task per session;
- **closing** — a dated entry naming what was delivered and what was left out,
  summary.md rewritten as finished, then the focus let go;
- **a decision that changes the task** — the title becomes the work decided,
  the description says what the task is now, the question and the decision go
  in a dated entry, labels and blockers follow;
- **the approach**, only when `[approach] label` is set (see
  [config.md](config.md)): `approach.md` written and agreed before any code;
- **the command line** — what an agent driving `cairn` from the shell would
  not expect, for when the MCP server fails or a batch is long;
- **the notation** — `#123` will do while no other tracker is in the
  conversation, because both the person and the agent wrote it anyway;
- **abandoning** — a task given up is deleted, after a dated entry saying why
  and summary.md rewritten; `done` would say it was delivered and tick its
  steps (`@plugin-the-skill-says-what-abandoning-writes`).

Two more skills read without changing the work:

- **`recap`** reads back the last day worked — the last day before today with
  at least one act in the trace — or a week, Monday to Sunday: the acts
  (`trace_list` with `since` and `until`), the time per task
  (`timesheet_show`), and for a week the objectives that moved. It then reads
  the Claude Code sessions of the period, which cairn never does, leaves out
  those the trace names (`claude session <id>`) or a task's notes hold, and
  proposes a task for each one left, in a table. On the person's word, line by
  line, it attaches one: an entry in the task's notes, closed with `released`
  so the hook does not take it for a session still running, the summary when
  it changes, and its time through `timesheet_add`. A session not to propose
  again goes in `sessions-ignored`, beside the store, with its reason.
- **`status`** says where things stand at any moment: the tasks in focus and
  in `current` with their open steps, what was done since the last day worked,
  what is left in the version (its label, counted per column), and what waits
  on the person.

Neither ranks nor reports anything cairn does not filter: see
[decisions.md](decisions.md). Each call they write out whole is handed to the
MCP server (`@plugin-the-other-skills-ask-what-cairn-answers`).

## Where it lives, and how it is installed

```
.claude-plugin/marketplace.json
plugin/claude/
  plugin.go                     the plugin, carried inside the binary
  .claude-plugin/plugin.json
  .mcp.json                     cairn mcp serve --access full
  skills/cairn/SKILL.md
  skills/recap/SKILL.md
  skills/status/SKILL.md
  hooks/hooks.json
  hooks/cairn-session.sh
```

The plugin is independent of cairn's code: it reaches cairn only through its
command line and its MCP server. Needed: `cairn` and `jq` on the `PATH`.
Nothing else.

**The sessions are the hook's, not the engine's.** Which Claude session holds
a task, that another one still working on it keeps it in focus, and letting go
after a while with no prompt are the hook's rules. Moving them into the engine,
as session verbs, was weighed and left out on 2026-10-23: it would have the
engine write and read the session entries in a task's notes, which cairn never
looks at ([context.md](context.md)). What the engine took instead is what any
surface needs: it lets go of a focus whose task is done or blocked, and asks
for a session past its length to be renewed ([ordering.md](ordering.md)).

**From GitHub, at the binary's tag, first.** The repository is a marketplace,
and Claude Code pins one to a tag: `/plugin marketplace add
vonglasow/cairn@vX.Y.Z`, then `/plugin install cairn@cairn`. `cairn plugin
show` (`plugin_show`) prints those for the version of the cairn running
(`@plugin-installed-from-its-tag`), after `/plugin marketplace remove cairn`:
Claude Code's documentation says nothing of moving a marketplace already added
to another tag, and removing it — which uninstalls its plugin, hence the
install after — is the one way it does document. A cairn that is no version
released has no tag, and says to run `cairn plugin write`
(`@plugin-installed-from-a-build`). `cairn upgrade` prints the same lines for
the version it put in place (`@upgrade-says-how-to-install-its-plugin`). The
release writes the version into the plugin's manifest
([release.md](release.md)), so at a session's start the hook reads its own
manifest, compares it with `cairn --version`, each cut at its first `-` so a
build since a version counts as that version, and says when they differ and
both are versions (`@claude-session-start-says-the-plugin-is-of-another-version`).
cairn never runs Claude Code's commands itself.

**The binary carries the plugin.** A person handed only the binary — by brew,
or as a file — has no repository to add as a marketplace. So `plugin.go`, in
the plugin's own folder, embeds its files, and `cairn plugin write` writes
them out as a marketplace: `D/.claude-plugin/marketplace.json`, read off the
plugin's manifest rather than written a second time, and the plugin under
`D/cairn`, its hook executable. `D` is the folder the configuration's `plugin`
key names, `claude-plugin` beside the store when it names none
(`@config-names-the-plugin-folder`). There is no flag to name another: a
plugin written where the configuration does not say is a plugin nothing keeps
current (`@plugin-has-one-place`). Then `/plugin marketplace add D` and `/plugin install
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

**Written once, it follows the cairn that serves it.** `cairn mcp serve`,
which the plugin's `.mcp.json` has Claude Code start with every session, reads
the version the plugin beside the store says it is (`claudeplugin.WrittenAt`)
and, when another cairn wrote it, writes it again before answering
(`engine.KeepPluginCurrent`, `@plugin-kept-current`); a line on standard error
says so. Claude Code reads a marketplace on a local folder in place, so the
session after an upgrade loads the new plugin — the session that rewrote it
still runs the one it started with. Nothing is written where nobody wrote a
plugin (`@plugin-not-written-unasked`); the folder the configuration names
is followed as the one beside the store is
(`@plugin-kept-current-where-the-configuration-says`). The same for every way cairn is installed.

**The session's server follows an upgrade too.** The server Claude Code
started notices, at the next tool call, that the `cairn` on the `PATH` is
another one, and hands its session over to it in place: the call is answered
by the new cairn, its tools are the ones Claude Code lists from then on, and
nobody runs `/mcp` ([mcp.md](mcp.md), "A cairn installed while it runs takes
over"). The plugin itself — the skill and the hook — is still read from the
next session.

The plugin declares the cairn MCP server, so that installing it is enough. A
person who already declared one keeps theirs: Claude Code drops a plugin's
server when one with the very same command is already there, and loads both
when the commands differ — tools twice over, harmless to the hook, noisy for
the model. `INSTALL.md` says to declare it with the plugin's command, or to
remove it.

## A repository that names its workspace

A server keeps the workspace it started on, and no tool switches it
([mcp.md](mcp.md)). A repository points its sessions at a workspace in the
environment Claude Code starts them with — its `.claude/settings.json`:

```json
{
  "env": { "CAIRN_WORKSPACE": "cairn-dev" }
}
```

Claude Code starts the plugin's server and runs the hook in that environment,
so both are on `cairn-dev`: one server, `mcp__plugin_cairn_cairn__…`, whose
`workspace_show` says `when this server started, CAIRN_WORKSPACE — over the
default, …`; the line said at the session's start names the workspace and
`because CAIRN_WORKSPACE`
(`@claude-session-start-says-the-workspace-the-environment-names`); and the
session's entries are written in that workspace's notes
(`@claude-session-the-hook-works-in-the-workspace-the-environment-names`).
One name instead of two paths, and the name the person already uses with
`--workspace`. `CAIRN_STORE` and `CAIRN_NOTES`, set beside it, still win for
the place each names ([config.md](config.md)); a name no workspace has is
refused, and the server answers each call with that refusal until it is
fixed. That Claude Code hands `.claude/settings.json`'s `env` to a plugin's
server is what its documentation says; not tried with `claude -p` yet.

**Not a second server.** Declaring one in the repository's `.mcp.json` with
`--workspace cairn-dev` gives two servers whose commands differ, so Claude
Code loads both (above, "The plugin declares the cairn MCP server"): each tool
twice, on two workspaces, and the hook following the environment's, not that
server's — a focus taken through it was written in the notes of the other
workspace. That is what `CAIRN_WORKSPACE` replaces.

## How it is held

- `bdd/claude-session.feature`: each scenario hands the hook the JSON Claude
  Code hands it, against the real cairn binary on a store and notes of its own,
  with the idle delay at one second. `test/claude` runs it, two levels under
  the root like every package holding scenarios, and in none of cairn's.
- `test/claude/plugin_test.go` holds the wiring no scenario reaches: the five
  events call a script that is there and runs, the matcher hears the focus
  from either server and nothing else, cairn answers to the server the plugin
  declares, and the marketplace points at the plugin.
- `bdd/claude-plugin.feature` (`@plugin-the-skill-asks-what-cairn-answers`)
  hands every call the skill writes out whole — a code span
  `` `tool_name {"argument": …}` `` — to the MCP server of the built binary,
  and fails on the first one refused. The skill once told an agent to list
  `state: todo,current`, which a listing refuses, and nothing noticed
  (cairn#240). A call is written whole when its arguments do not depend on the
  task at hand; `task_focus` with the number stays prose, and unchecked.
- `shellcheck` runs at every commit on every shell script (docs/checks.md).
- Tried for real with `claude -p --plugin-dir plugin/claude`: the skill was
  called by "on travaille sur cairn#1", the focus taken, the entry written with
  the session's directory and id, the summary read, and the session's end
  released the focus and closed the link.

## Not in this piece

Putting a column back in order on request, bringing external tickets in,
ticket systems other than Azure DevOps, a setting that fetches
the ticket on every attachment.
