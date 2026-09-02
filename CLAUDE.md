# cairn

A personal objective and task manager. CLI first, then MCP; the terminal and
desktop boards come later, once the engine stands on its own.

## Always

**English.** Code, comments, commit messages, scenarios, documentation.

**`prek` runs on its own at every commit**, and whatever it reports gets fixed —
every time. Not worked around, not deferred.

**Nothing exists twice.** Before writing anything, look for it first.

**Scenario, then test, then code.** Never the reverse. A write to a
production `.go` file under `internal/` is refused until `make red
PKG=./<package>` has recorded a failing test in that package — the rule is
no longer only written down, see
[.claude/hooks/red-first.sh](.claude/hooks/red-first.sh).

**Nothing is committed until the definition of done is met.** Walk
[docs/definition-of-done.md](docs/definition-of-done.md) before every commit.
A box you cannot tick is not a box you skip: say which one, say why, and ask
before committing.

**Every change of state goes through `internal/engine`.** A surface parses,
calls it, and prints — it decides nothing.

## Read before acting

| doing this | read |
|---|---|
| starting any piece of work, writing a scenario, committing | [docs/workflow.md](docs/workflow.md) |
| about to commit, or wondering whether something is finished | [docs/definition-of-done.md](docs/definition-of-done.md) |
| adding a package, a command, anything that changes state | [docs/architecture.md](docs/architecture.md) |
| writing code | [docs/design-principles.md](docs/design-principles.md) |
| adding a tool, changing a threshold, unblocking a hook | [docs/checks.md](docs/checks.md) |
| touching the MCP server, or adding a command an agent will see | [docs/mcp.md](docs/mcp.md) |
| touching the order of tasks, a task's place, or the focus | [docs/ordering.md](docs/ordering.md) |
| narrowing a listing, or changing what a search reaches | [docs/listing.md](docs/listing.md) |
| touching what stops a task, or what starts it again | [docs/blocking.md](docs/blocking.md) |
| touching what a task serves, or how an objective is counted | [docs/objectives.md](docs/objectives.md) |
| adding anything that changes a task, or touching what is recorded about it | [docs/trace.md](docs/trace.md) |
| touching the board — its shape, its keys, its modes — or anything a person sees on screen | [docs/tui.md](docs/tui.md) |

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
