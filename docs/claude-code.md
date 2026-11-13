# With Claude Code

cairn ships a Claude Code plugin. With it, a session starts with the context
of the task it serves, writes what it learns where the next session will find
it, and counts its time against that task — without you repeating any of it.

## Installing it

```sh
cairn plugin show
```

prints the commands that install the plugin of the version of cairn you run,
to paste into Claude Code:

```
/plugin marketplace remove cairn
/plugin marketplace add vonglasow/cairn@vX.Y.Z
/plugin install cairn@cairn
```

A cairn built from source has no tag; `cairn plugin write` writes the plugin
into a folder beside the store — or the one the configuration's `plugin` key
names — and prints the two commands that install it from there. Written that
way, it is kept current: the next MCP server started by an upgraded cairn
writes it again. `cairn upgrade` prints the commands for the version it put in
place.

The plugin needs `cairn` and `jq` on the `PATH`. It declares the cairn MCP
server itself, at the `full` access level; if you declared one by hand, remove
it or declare it with the same command, or Claude Code loads both and lists
every tool twice.

## Words

In a conversation with Claude, "task" alone is Claude's own to-do list. A
cairn task is said **"task cairn"** and written **`cairn#123`** — beside
`AB#123` for Azure Boards, `PROJ-123` for Jira and `owner/repo#123` for GitHub.

## What a session does

**Attaching.** Say which task you work on — "on travaille sur cairn#123",
"work on cairn#42". Claude puts it in focus, which moves waiting work into
`current`, then reads from the task's notes `summary.md`, the latest dated
entries, its steps, and `ticket.md` when there is one.

**Choosing.** Ask what to work on. Claude lists the first five tasks of `todo`
and `current` that nothing blocks, in the board's order, and you pick one.
Claude ranks nothing.

**Working.** What was done or decided goes in the task's notes: a dated entry
per piece of work, and `summary.md` rewritten whole — never in the
description, which says what the task is. A task is split into steps rather
than into more tasks. A task done whose work does not hold is not reopened: a
new task linked to it is opened.

**Closing.** When the work is delivered, Claude writes what was delivered and
what was left out, rewrites the summary, puts the task in `done` and lets go
of the focus.

**Detaching.** When the subject changes or you stop, Claude updates the
summary and lets go of the focus. A session closed abruptly lets go of it
through the hook.

**The ticket, on demand.** A task tied to Azure DevOps carries the label `ado`
and the work item's id as a label of digits alone. Only when you ask ("charge
le ticket") does Claude read it and rewrite `ticket.md` from it.

**An agreed approach.** When the configuration names an
[`[approach] label`](config.md#an-approach-agreed-first), a task carrying it
is not coded before you agree on how: Claude writes `approach.md` in its notes
and waits.

## The session and the focus

The plugin's hook ties each session to its task through a session entry in the
task's notes, `<date>-session-<id>.md`, which says how to resume the session:

```
cd <the session's working directory>
claude --resume <session-id>
```

- Ten minutes without a message let go of the focus; the next message takes
  it again.
- A task worked on by several sessions keeps its focus until the last one
  stops.
- A focus let go from a board while a session is attached is taken again at the
  session's next message.
- A focus that has run long asks to be renewed; see
  [configuration](config.md#focus).
- What the hook does is written in the task's trace as
  `claude session <id>`, and what the agent does as `claude-code`.

## Which board a session uses

A session uses the workspace of the project it runs in — a project holding a
workspace's folder, its `.cairn` say — the way the command line does from
there. At its start, the hook says which workspace answers, and why. A project
that holds no workspace can name one in its `.claude/settings.json`:

```json
{
  "env": { "CAIRN_WORKSPACE": "work" }
}
```

See [workspaces](config.md#workspaces).

## Looking back

Two more skills read without changing the work:

- **`recap`** reads back the last day worked, or a week: what was done, the
  time per task, and for a week the objectives that moved. It also reads the
  Claude Code sessions of the period that no task names, and proposes a task
  for each; on your word it attaches them and adds their time.
- **`status`** says where things stand: the tasks in focus and in `current`
  with their open steps, what was done since the last day worked, what is left
  in the version, and what waits on you.
