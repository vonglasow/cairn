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

The plugin needs `cairn` and `jq` on the `PATH`. Its hook is a shell script:
on Windows, Claude Code runs it through Git Bash, and `jq` comes with
`winget install jqlang.jq`. Restart Claude Code once it is installed.

The plugin declares the cairn MCP server itself, at the `full` access level;
if you declared one by hand, remove it or declare it with the same command, or
Claude Code loads both and lists every tool twice:

```sh
claude mcp remove cairn
claude mcp add --scope user cairn -- cairn mcp serve --access full
```

## Words

In a conversation with Claude, "task" alone is Claude's own to-do list. A
cairn task is said **"task cairn"** and written **`cairn#123`** — beside
`AB#123` for Azure Boards, `PROJ-123` for Jira and `owner/repo#123` for GitHub.

## What a session does

**Attaching.** Say which task you work on — "let's work on cairn#123",
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
and the work item's id as a label of digits alone. Only when you ask ("load
the ticket") does Claude read it and rewrite `ticket.md` from it.

**An agreed approach.** When the configuration names an
[`[approach] label`](config.md#every-key), a task carrying it
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
  [configuration](config.md#every-key).
- What the hook does is written in the task's trace as
  `claude session <id>`, and what the agent does as `claude-code`.

## The sessions running

`cairn agent list` lists the Claude sessions running in the workspace, and the
window's Agents tab shows them. From there:

- `cairn agent go <session>` brings its terminal to the front: iTerm2,
  Terminal and Ghostty (1.4 and later) on macOS, kitty on Linux. Any other terminal, or an editor, says
  the command that resumes it instead. On Windows the session says what it
  runs in — Windows Terminal, VS Code, WezTerm or Claude Desktop — and
  `cairn agent go` says the command that resumes it, since none of them is
  brought to the front. That command is for PowerShell, Windows PowerShell
  5.1 included.
- `cairn agent start --folder <dir> --terminal <terminal> [--task <n>]`
  starts claude in a new tab, asked to work on the task when one is named;
  `cairn agent terminals` names the terminals it starts in.
- `cairn agent tie <session> --task <n>` ties a session to a task, and
  `--title "<text>"` to a task made for it, put in current; the task goes in
  focus.
- With `agents.permissions = true`, a session asking leave to use a tool asks
  the window while one is open: allow, deny, or allow for good in the
  project's or your settings — `cairn agent answer <session> <answer>`. With
  no window open, the terminal asks as it always does.
- The questions Claude asks you are answered the same way:
  `cairn agent answer <session> --choice Postgres --choice "Tests, Docs"`, one
  `--choice` for each question in its order, the labels joined by commas where
  it takes several.

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
