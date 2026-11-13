# Decisions about how cairn is built

Read this before proposing a change to how cairn is built, checked or
released. The decisions a person using cairn would ask about are in
[docs/decisions.md](../../docs/decisions.md); these are the ones only whoever
develops cairn meets: one line per decision, and where its reason lives.

## What a task is, and what it carries

| decided | why, in one line | where |
|---|---|---|
| the rule grammar is written by hand, rrule-go is not added | the words typed are not RRULE, the subset is small, the library has not moved since 2023 | [dependencies.md](dependencies.md#weighed-and-not-added) |
| a study and what it decides to build land in the same version | #222 and #226 landed a version apart, and a version shipped a decision with nothing of it built | [workflow.md](workflow.md#a-study-and-its-build) |

## What is remembered, and what is not

| decided | why, in one line | where |
|---|---|---|
| the plugin is installed from GitHub at the binary's tag; cairn says the commands, never runs them | the plugin then is the one the version was made with, and Claude Code versions it by its manifest, which the release writes; Claude Code's own commands are the person's to type | [claude-code.md](claude-code.md#where-it-lives-and-how-it-is-installed), [release.md](release.md) |

## The store

| decided | why, in one line | where |
|---|---|---|
| no index on the tasks table | a store holds hundreds of tasks; an index is added the day a listing is measured slow | this page |
| SQLite, in Go, one file | one static binary, no cgo for the store | [dependencies.md](dependencies.md) |
| every write is one transaction, every migration one per change | an act, or a change of shape, half kept is worse than none | [architecture.md](architecture.md) |
| integrity between tables is kept by the code, not by foreign keys | SQLite leaves them off per connection; the declared `ON DELETE` clauses of the first shape are inert, and what refers to a task is removed by hand on a deletion, inside its transaction | `internal/store/store.go`, `Delete` |
| closed sets — a column, a priority — are refused in `domain`, not by a `CHECK` | one place says what a column is, and a `CHECK` would be a second list to keep true | `internal/domain/attributes.go` |
| a change that alters data and not the shape is forgotten, not refused, when rolled back | it has nothing to undo, and refusing it would wall off every older shape behind it; what it leaves as it was is said by its `loses`, and applying it again changes nothing | `internal/store/migrate.go`, `RollBack` |
| a backup is refused as a restore when SQLite cannot read it whole | a damaged backup put in place trades a store that opens for one that does not | `internal/store/backup.go` |
| an engine held open follows a store put back under it, rather than a restore being refused while a board is open | measured: a board's connection stayed on the file the restore replaced, did not see what came back, and wrote where nobody reads, without an error; the engine now sees its file was replaced before every act and every beat, and reopens. A lock held against boards would be a lock a forgotten window holds for ever | `internal/engine/engine.go`, `following` |
| looking at a backup, or at the store before backing it up, opens the file to be read and nothing else | opened the shared way, a store that was not there was made readable by anybody, and a backup looked at was rewritten and its date moved | `internal/store/backup.go`, `heldIn` |

## The surfaces

| decided | why, in one line | where |
|---|---|---|
| which acts ask first is the engine's, and every surface asks the same ones | the boards asked before sweeping a column and the command line did not: one act, two rules. The engine refuses an act that loses work, or sweeps up what a narrowing reaches, unless it comes confirmed; `--yes`, a board's question and an agent's access are only how each surface says it asked | `Operation.AsksFirst`, `Input.Confirmed` |
| the Claude sessions holding a task are the hook's, not the engine's | the engine would write and read the session entries in a task's notes, which cairn never looks at; it keeps only what every surface needs, letting go of a focus and asking for a session to be renewed | [claude-code.md](claude-code.md#where-it-lives-and-how-it-is-installed) |
| every change goes through the engine; a surface decides nothing | the rules live once | [architecture.md](architecture.md) |
| what kind of answer a failure is — refused, not understood, store, unforeseen — is the engine's | a script's exit code and an agent's refusal are read off the same list | `internal/engine/refusal.go` |
| changes to the shape of a store are made in order, and a rollback reads the ledger again inside its transaction | each change was written on the shape the ones before it leave; two rollbacks read before either ran both said they undid the same change | `internal/store/migrate.go` |
| a database holding tables, and none of a store's, is refused before anything is written into it | a path named by mistake can lead to somebody else's database, and opening it the usual way already writes into it | `internal/store/open.go`, `refuseAnotherDatabase` |
| several tasks or objectives are deleted or archived in one act, under `--also`, all of them or none | three deletions asked one after another left, when the third was refused, two gone and somebody to work out which; one act is one transaction, and a new optional argument breaks no caller | `internal/engine/together.go` |
| a version's number is read off the commits by svu, and its tag is made by `make release` once every check has passed | a number typed by hand is a number somebody misread the history for, and a tag made before the checks says a failed build was a version | [release.md](release.md) |
| a release runs every check but the mutation sweep, which is run on its own and leaves a report | the sweep took the better part of an hour, and a release that waited on it was a release nobody made; every commit still mutates the lines it changed, and the report says what the last sweep found | [checks.md](checks.md#the-report-a-sweep-leaves) |
| a version is handed over as binaries, each built on the system it runs on, and no source archive | the window needs cgo and the system's graphics headers, which a cross-compile cannot reach, so each binary is built where it runs; whoever builds cairn clones it | [release.md](release.md) |
| `--json` is a flag of every command, including those that never answer in JSON | one flag at the root reads the same everywhere; on `completion` or `gui` it changes nothing and costs nothing | `internal/cli/cli.go` |
| no comments in the code but one line per exported symbol, and no prose in the scenarios | a quarter of the code was comments, much of it the history of a line rather than its reason; the reasons live in `docs/` and in the commits | [design-principles.md](design-principles.md#comments) |
| what a terminal would obey, or what formats text rather than shows it — every Unicode Cf character, emoji joiners included — is refused when written and replaced when shown | the category is the rule, since a list of the characters somebody thought of let the rest through, tag characters an agent reads included; the refusal keeps it out from now on; showing it replaced covers what was kept before the refusal existed, or comes back with an older backup | `internal/domain/title.go`, `Prose`, `Shown` |
| a tool whose answer shape cannot be derived is offered without one, rather than stopping the server | a test proves every tool derives one, so this can only happen to code nobody ran | `internal/mcp/tools.go` |
| the two boards share where a card goes, not how the cursor walks | they draw the stalled band differently — a grid of three in the terminal, lanes under the columns in the window — so the same key is a different journey | [architecture.md](architecture.md) |
| what a board reads lives in the engine, not in a package between it and the boards | a surface is an entry point; the package in between grew decisions of its own, and its search drifted from the engine's | [architecture.md](architecture.md) |
| the engine is one package, split into files by notion | one door (`Ask`) and one place for the rules; sub-packages wait until a file cannot be read on its own | [architecture.md](architecture.md) |

## The checks

| decided | why, in one line | where |
|---|---|---|
| what the tools find stays current; what the tools are is pinned | a new rule found something already true; a tool changing its own behaviour found nothing | [checks.md](checks.md#getting-them) |
| no threshold is lowered | a gate that bends is not a gate | [checks.md](checks.md#thresholds) |
| `&&`/`||` and `continue`/`break` mutated in the core only | that is where the rules are; the surfaces gave mostly mutants no test can tell apart | [checks.md](checks.md#thresholds) |
| the surfaces are never mutated; the core is mutated in CI, and no sweep is run by hand before a push or a version | the scenarios judge what is on screen, and a sweep of every package was estimated at forty to fifty hours, the window's for most of it | [checks.md](checks.md#thresholds) |
| packages are mutated one after another | gremlins already uses every core inside a package; two at once only made timeouts | [checks.md](checks.md#running-them-by-hand) |
| Linux is checked by hand, in a container, before a version | it needs Docker, which no hook may assume | [release.md](release.md) |
| the container's image is named by tag and Debian release, not digest; commitlint's own dependencies are not locked | the first is a check run by hand, not a gate, and a digest would be moved at every Go release for nothing anyone notices; the second runs only on a message, locally, and locking npm's tree would mean a lockfile in a Go repository for one hook | `scripts/linux.sh`, `.pre-commit-config.yaml` |
| a form's commands are waited for in the terminal board's tests, all but the cursor's blink, passed over at once | the blink is a timer of 530 ms nobody can wait out, and its function names it before it runs; racing every command against five milliseconds, as before, dropped one a machine busy with the rest of the suite was slow to schedule (cairn#298), and raising that deadline made the suite three times slower | `internal/tui/fake_test.go`, `quickly` |

## Known and not yet answered

What a review found and nobody has decided yet. It is written here so it is
not found a second time as news, and it leaves this list when it is fixed or
decided.

Nothing, at the moment.

## One word for each notion, on every surface

The window, the terminal, the command line, the MCP server and these pages
used to name one notion several ways: a task that comes back was a series, a
recurrence and something that repeats; a day due was a deadline and a due;
the settings were a configuration. Each notion now has one word, the one the
person reads (#292, decided 2026-11-11):

- **series** — the setting reads Series, its key `series.on` (the old
  `[recurrence] on` is still read), the filter `series` on `all`,
  `task list --series` alone keeps every occurrence and `--series 3` those of
  ↻3; the refusal says *series are off*;
- **rule** — the field a series comes back by; `--every` stays the flag;
- **occurrence** — what a series made, opened as one;
- **settings** — every screen; `cairn config` and the file names stay;
- **tab** — what was a page, in the settings' section *Tabs*;
- **trace** — what was the history of a task;
- **column** — the filter pill reads `column: todo`; every filter pill reads
  `name: value`, the view's and the search's as before;
- **state** — an objective's, in its form;
- **deadline** — `--deadline-before`, the filter `deadline`, `deadline` where
  a day was `due`;
- **notes** — what a form called its context;
- **left** — `3 left · 5 done`, never `to do` beside a column TODO;
- **blockers**, **parent** — the form's sections;
- **save** — to keep something written, *Save as a view…*, *Save search*;
  *keep* only when a deletion is refused;
- **new, list, show, edit, delete** — the words the commands' groups say;
- **report** — `objective report`, never a review;
- **details**, **help** — the headings.

- **forget** — `workspace forget`, which was `workspace remove`;
- **description** — `--description` is `--desc` spelled whole;
- **theme** — never a scheme, in what an agent reads too;
- **Series** joins the View menu; outside the Series tab the series' acts
  read *Open the series*, *Skip this one*, *Stop the series*, *Resume the
  series*; in the tab, where the series is said already, *Edit the series*,
  *Skip the next*, *Stop*, *Delete…*;
- **`…`** ends a line only when a choice follows it, and a line opening with
  `+` is written in sentence case: *+ New task serving it*.
