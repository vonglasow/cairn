# cairn

A personal objective and task manager: a command line, an MCP server, a board
in the terminal and the same board in a window — four surfaces over one engine.

## Always

**English.** Code, commit messages, scenarios, documentation.

**No comments in the code** but a one-line doc comment on each exported symbol;
the reasons live in `.claude/docs/` and in the commits. See
[.claude/docs/design-principles.md](.claude/docs/design-principles.md#comments).

**`prek` runs on its own at every commit**, and whatever it reports gets fixed —
every time. Not worked around, not deferred.

**Nothing exists twice.** Before writing anything, look for it first.

**Scenario, then test, then code.** Never the reverse. A write to a
production `.go` file under `internal/` is refused until `make red
PKG=./<package>` has recorded a failing test in that package — the rule is
no longer only written down, see
[.claude/hooks/red-first.sh](.claude/hooks/red-first.sh).

**Nothing is committed until the definition of done is met.** Walk
[.claude/docs/definition-of-done.md](.claude/docs/definition-of-done.md) before every commit.
A box you cannot tick is not a box you skip: say which one, say why, and ask
before committing.

**A task cairn worked on is held, and its work goes in its notes.** The
board of this repository is `CAIRN_STORE=$PWD/.cairn/cairn.db` with
`CAIRN_NOTES=$PWD/.cairn/notes`. Before working on a task, put it `current`
and in focus. What was done or decided goes in
its notes, `.cairn/notes/<n>/`: a dated entry `YYYY-MM-DD-<subject>.md`, and
`summary.md` rewritten whole; **never in its description**, which says what the
task is. When it is committed, unfocus it and put it `done`. The cairn skill
(`plugin/claude/skills/cairn/SKILL.md`) has the rest.

**`docs/` is for the person using cairn; `.claude/docs/` is for whoever
develops it.** The site publishes `docs/` whole: how to use cairn, never why it
is built the way it is. The reasons, the decisions, the scenarios and the code
they name go in `.claude/docs/`, the page of the same name when there is one.
A behaviour that changes changes both.

**Every change of state goes through `internal/engine`.** A surface parses,
calls it, and prints — it decides nothing.

## Read before acting

| doing this | read |
|---|---|
| starting any piece of work, writing a scenario, committing | [.claude/docs/workflow.md](.claude/docs/workflow.md) |
| about to commit, or wondering whether something is finished | [.claude/docs/definition-of-done.md](.claude/docs/definition-of-done.md) |
| adding a package, a command, anything that changes state | [.claude/docs/architecture.md](.claude/docs/architecture.md) |
| writing code | [.claude/docs/design-principles.md](.claude/docs/design-principles.md) |
| adding a tool, changing a threshold, unblocking a hook | [.claude/docs/checks.md](.claude/docs/checks.md) |
| touching the MCP server, or adding a command an agent will see | [.claude/docs/mcp.md](.claude/docs/mcp.md) |
| touching the Claude Code plugin — its hook, its skill, how it is installed | [.claude/docs/claude-code.md](.claude/docs/claude-code.md) |
| touching the order of tasks, a task's place, or the focus | [.claude/docs/ordering.md](.claude/docs/ordering.md) |
| narrowing a listing, or changing what a search reaches | [.claude/docs/listing.md](.claude/docs/listing.md) |
| touching how a label is written, or how one is put on a task | [.claude/docs/labels.md](.claude/docs/labels.md) |
| touching what stops a task, or what starts it again | [.claude/docs/blocking.md](.claude/docs/blocking.md) |
| touching a task's steps, or what a task is split into | [.claude/docs/steps.md](.claude/docs/steps.md) |
| touching what links two tasks, or how a defect found in a task done is opened | [.claude/docs/links.md](.claude/docs/links.md) |
| touching what a task serves, or how an objective is counted | [.claude/docs/objectives.md](.claude/docs/objectives.md) |
| touching a series, a rule, or what is made when a task comes back | [.claude/docs/series.md](.claude/docs/series.md) |
| touching what showing a task says about its outside context, or where that context is looked for | [.claude/docs/context.md](.claude/docs/context.md) |
| adding a key to the configuration, or changing how a place is addressed | [.claude/docs/config.md](.claude/docs/config.md) |
| writing a key of the configuration from a surface, or reading it again while a board is open | [.claude/docs/config.md](.claude/docs/config.md#one-configuration-several-ways-in-to-it) |
| adding anything that changes a task, or touching what is recorded about it | [.claude/docs/trace.md](.claude/docs/trace.md) |
| touching the board — its shape, its keys, its modes — or anything a person sees on screen | [.claude/docs/tui.md](.claude/docs/tui.md) |
| touching the desktop window, or anything a person sees in one | [.claude/docs/gui.md](.claude/docs/gui.md) |
| touching how long the work took, or the timesheet | [.claude/docs/trace.md](.claude/docs/trace.md) |
| adding a dependency, or wondering why one is there | [.claude/docs/dependencies.md](.claude/docs/dependencies.md) |
| sharing a board with other people, syncing a store, or anything decentralised | [.claude/docs/sharing.md](.claude/docs/sharing.md) |
| proposing something cairn does not do, or wondering why it does not | [docs/decisions.md](docs/decisions.md), then [.claude/docs/decisions.md](.claude/docs/decisions.md) for how it is built |
| touching what debug mode writes down, or where | [.claude/docs/debug.md](.claude/docs/debug.md) |
| making a version, or handing a binary to anybody | [.claude/docs/release.md](.claude/docs/release.md) |

Those files carry the detail. This one carries what holds regardless of the
task, because a rule that has to be looked up is a rule that gets skipped.

## What we don't do

- **Never `git commit --no-verify`**, nor its short form `-n`, nor uninstalling
  the hooks. A commit that skips the checks is a commit nobody checked, and the
  problem it hid is found by someone else, later, with no idea where it came
  from. If a hook blocks, the hook is right until proven otherwise — fix what it
  reports, or say out loud that you cannot and stop there.
- No dependency added without a written reason.
- No threshold lowered to get a commit through.
- No scenario reworded to suit the code.
