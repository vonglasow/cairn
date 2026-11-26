# Architecture

Read this before adding a package or a command. Why each choice was made is
one line in [DECISIONS.md](DECISIONS.md); the rules of work are in
[CLAUDE.md](CLAUDE.md); building is in [INSTALL.md](INSTALL.md), and
`make help` lists the targets. `make fixtures` opens a window on a sample
board made afresh, in debug mode.

## Four layers

```
surfaces   cli   mcp   gui   tui          parse, show, read keys
board      internal/board                 what is shown, computed once
engine     internal/engine, engine/...    what cairn knows and does
storage    store/sqlite   store/files     where it is kept
```

A layer imports only the layers below it. `depguard`, in
[.golangci.yml](.golangci.yml), holds it: one rule per package, the reason
beside it. `make rule-check` refuses a package under `internal/` that no rule
names.

| package | holds |
|---|---|
| `engine` | the catalogue of every operation, `Ask` the one door, and what reads several notions at once |
| `engine/item` | what a task and an objective both have: column, priority, title, card, progress |
| `engine/change` | the changes an input says: `+x -x =x`, replace … by …, more or less time |
| `engine/calendar` | days, deadlines, estimates, durations, periods |
| `engine/task`, `objective` | the two entities: their rules, their record, the store they need as an interface |
| `engine/act` | the trace: one act done to a task or an objective |
| `engine/filter`, `view` | what a read keeps; a set of filters under a name |
| `engine/label`, `series`, `backup`, `workspace`, `migration`, `exchange`, `app` | one notion each |
| `engine/agent` | a Claude session running, read from Claude Code's own files, and a permission it asks the window |
| `store/sqlite` | the store; the only package that knows a database exists |
| `store/files` | the files beside it: configuration, drafts, window sizes, recent workspaces, and what Claude Code keeps of its sessions |
| `board` | what every board reads: cards, goals, forms, menus, the walk, `Board` the state on screen |
| `cli`, `mcp` | the command tree and the tools, both made from the catalogue |
| `cli/render` | an answer as a person reads it, for every surface |
| `tui`, `tui/geom`, `tui/design` | the board in a terminal: the program, cells, styles |
| `gui`, `gui/geom`, `gui/design`, `gui/motion` | the board in a window: the program, pixels, skin, motion |
| `theme` | a base16 scheme turned into every colour, and its contrasts |
| `debug` | what debug mode writes down |
| `build`, `upgrade`, `logo`, `macapp`, `linuxapp` | the binary, its upgrade, and cairn as an application |
| `plugin/claude` | the Claude Code plugin, embedded in the binary |
| `site`, `cmd/site` | the documentation site, never part of the binary |
| `test`, `test/claude` | the acceptance world the surfaces share |

## A notion keeps its rules and hands over its facts

A task's state has no exported field. Every rule about it is a method on it:
whether it may archive, finish with a step open, serve an objective. What it
carries it hands over as one record, `task.Carried`; `Made` turns a record
back into a task. The card, the JSON, the row in the store and the search all
read that one record. An objective, an act and a migration work the same way.

**A notion never imports another notion.** A task serves an objective by its
number. The engine and the store tie them together.

An act names its task by number and never carries it: the task may be gone
when its acts are read.

## Every change of state goes through the engine

Each operation is described once in `internal/engine`: its name, its
arguments and what each takes, its access level, what it needs of the store,
whether it asks first, and how it runs. The command line builds a cobra tree
from the catalogue, the MCP server its tools, the boards their menus and
forms. One flag added reaches every surface.

An engine held open exports `Ask`, `Close` and `Version`. `engine.Do` opens
the store, asks and closes; a board holds one engine and calls `Ask`. A test
that needs a store in some state writes it through the operations
(`test.Givens`).

**A surface decides nothing.** It turns a flag, a JSON argument, a key or a
click into an input, calls `Ask`, and shows the answer in its medium. `cli`
cannot import the store or `database/sql`. A surface computing a rule inline
is the one thing no linter holds: `CLAUDE.md` says it.

The store is opened only by the operation that needs it, so `cairn --version`
writes no file, and a bad argument is refused before anything is written.

## What is shown, computed once

`internal/board` holds what both boards show: the facts of a card (each a
word, a mark and a tone, mapped by each board to its glyph and colour), the
columns and the band of three lanes, the walk between cards, the menus made
from the catalogue, the forms made from an operation's fields, the settings
from the configuration keys. `Board` is the state of a board on screen; the
window and the terminal embed it. It changes nothing but through `Ask`.

A board places what `board` gives in its own units (`geom`: pixels or cells)
and draws it with its design (`design`: the theme's colours, its faces or
glyphs). The window answers the pointer from the same boxes it draws with
(`Placed`, `CardRowOf`, `PillRowOf`, `FormPlace`).

## Failures

`engine.RefusalOf` reads an error and says its kind and a stable reason
(`no-such-task`). The command line turns the kind into an exit code, MCP puts
both under `_meta`, `--json` prints them.

| exit | kind |
|---|---|
| 0 | it worked |
| 1 | `refused`: understood and not satisfied; or `failed`: unforeseen |
| 2 | `usage`: not understood (`engine.Misread`) |
| 3 | `store`: the store could not be used |

A surface's own refusal is made with `engine.Refusing`.
`TestEveryFailureCairnForesawHasAReason` fails until every `Err…` below the
surfaces has a reason.

## The store

SQLite in Go, one file, WAL, a busy wait of five seconds: two cairns at once
are the normal case (a board and an agent). Each operation that writes is one
transaction, begun as a write.

The store speaks in notions: it takes a `task.Task` and gives one back,
through an interface the engine declares. No `*sql.Rows` crosses it.

**Opening reads the shape and never changes it.** Only `migration apply`
changes it, one transaction per change, each carrying its undoing, both
printable with `migration show`. A one-shot command is refused a store
behind; a board holds it (`Hold`, `OpenBehind`) and its operations that need
a current store are refused `ErrStoreBehind` until the person agrees. A store
from a newer cairn is refused to everything.

A store is made in one transaction; a store being made by another cairn is
waited for.

A backup is the store read whole with `VACUUM INTO`, log included.
`store.Vault` opens no database: the backups are reachable when the store is
not.

## A step is part of its task

A step is a title and a tick, kept by its task's number and its place. It has
no number, no focus and no trace of its own. It moves, archives and is
deleted with its task; a change to it is a change to its task. A task is
refused `done` while a step is open, unless it comes confirmed.

## Tests

- **Behaviour** is tested once, in table tests in `engine` and `board`.
- **A surface** keeps one contract test, and the tests that alone reach its
  own lines.
- **A scenario** in `bdd/` says what a person sees, through a surface. A
  scenario or a unit test is kept when it reaches a line nothing else does,
  and one scenario per rule.
- `test` is the acceptance world: handed a runner, it never knows which
  surface answers, so no surface imports another. `make harness-check` keeps
  it out of the binary.
- No test waits on time: clocks and processes are injected; a program a test
  writes and runs goes through `test.WriteProgram`.
- Coverage floors: 90 % in all, 85 % per package (`make cover-check`).
  `cmd/cairn` stays one line, since no test reaches it.

## Checks

`prek` runs at every commit: gofmt, goimports, golangci-lint with depguard,
shellcheck, semgrep, govulncheck, the tested, rule, harness, windows, quiet
and cover checks, and commitlint ([.pre-commit-config.yaml](.pre-commit-config.yaml)).
`.github/workflows/check.yml` runs the same on every pull request, on Linux.
Mutation (`gremlins`) is a report run by hand on the core, never a gate.
Tools are pinned; what they find is fixed, and no threshold is lowered.

## Release

`make release` reads the version off the commits (svu), runs every check on
a clean `main`, and tags. Pushing the tag runs
`.github/workflows/release.yml`: the checks again, then one binary built on
each system, since the window needs cgo. The plugin's manifest carries the
version; the plugin is installed from GitHub at the binary's tag.
