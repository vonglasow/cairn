# Decisions

Read this before proposing something cairn does not do. Most of what it leaves
out was weighed and left out on purpose, and the reason is written beside the
rule it belongs to — which is the right place for it and the wrong place to
find it from. This is the index: one line per decision, and where the reason
lives.

A decision here is not a wall. It is what was true when it was made, with the
reason that made it; a new reason reopens it. What it saves is somebody
arriving at the same idea, building half of it, and finding the paragraph that
already said why not.

## What a task is, and what it carries

| decided | why, in one line | where |
|---|---|---|
| no urgency field | it says what the priority and the column already say | [ordering.md](ordering.md#weighed-and-left-out) |
| no recurrence, no work that never finishes | cairn models work that reaches `done`; turning a real problem into a mechanism was not asked for | [ordering.md](ordering.md#weighed-and-left-out) |
| no precedence between two tasks | measured: it constrains without ordering | [ordering.md](ordering.md#weighed-and-left-out) |
| how long a task has sat in its column is read off the trace, never stored | the one date nothing else implies, and the trace already holds it | [ordering.md](ordering.md#what-the-store-gains) |
| no `unblock` verb | `--blocker -3` does it, and returning to `todo` belongs to the edit | [blocking.md](blocking.md#deliberately-not-there-yet) |
| no "what makes a task doable now" | two properties were found and the list plainly does not close | [ordering.md](ordering.md#deliberately-not-here-yet) |
| no `task next` | what it would answer is already asked plainly — `--due-before`, `--state todo`, `task focus` — and choosing among them is the person's order, which cairn does not override | this page |

## What is remembered, and what is not

| decided | why, in one line | where |
|---|---|---|
| a deletion keeps the task's history | otherwise the past changes every time something is deleted | [trace.md](trace.md#the-trace-outlives-the-task) |
| no undo, and no putting a deleted task back | reading and undoing are two different asks; the history makes it possible to work out, nothing does it | [trace.md](trace.md#deliberately-not-here-yet) |
| a deletion keeps what the task held, whole, in its last act, and nothing purges it | the history of a task is the reason the trace exists, and a deletion that kept nothing made even putting it back by hand impossible; a purge — `trace purge`, a task named, `--yes` — is built the day something written by mistake has to go, not before. Until then, what must not be kept is not written | [trace.md](trace.md#a-deletion-writes-down-what-went-with-it) |
| the timesheet gives the hours raw, by day and task | what a good day looks like is a judgement; grouping by objective or label was not asked for | [trace.md](trace.md#deliberately-not-here-yet) |
| cairn says where a task's notes go and never looks there | a cairn that read them would compete with the harness it plugs into; counting the notes is looking | [context.md](context.md#it-never-looks-so-there-is-nothing-here-that-can-fail) |

## The store

| decided | why, in one line | where |
|---|---|---|
| SQLite, in Go, one file | one static binary, no cgo for the store | [dependencies.md](dependencies.md) |
| every write is one transaction, every migration one per change | an act, or a change of shape, half kept is worse than none | [architecture.md](architecture.md) |
| integrity between tables is kept by the code, not by foreign keys | SQLite leaves them off per connection; the declared `ON DELETE` clauses of the first shape are inert, and what refers to a task is removed by hand on a deletion, inside its transaction | `internal/store/store.go`, `Delete` |
| closed sets — a column, a priority — are refused in `domain`, not by a `CHECK` | one place says what a column is, and a `CHECK` would be a second list to keep true | `internal/domain/attributes.go` |
| no index on the tasks table | a store holds hundreds of tasks; an index is added the day a listing is measured slow | this page |
| a change that alters data and not the shape is forgotten, not refused, when rolled back | it has nothing to undo, and refusing it would wall off every older shape behind it; what it leaves as it was is said by its `loses`, and applying it again changes nothing | `internal/store/migrate.go`, `RollBack` |
| a backup is refused as a restore when SQLite cannot read it whole | a damaged backup put in place trades a store that opens for one that does not | `internal/store/backup.go` |
| an engine held open follows a store put back under it, rather than a restore being refused while a board is open | measured: a board's connection stayed on the file the restore replaced, did not see what came back, and wrote where nobody reads, without an error; the engine now sees its file was replaced before every act and every beat, and reopens. A lock held against boards would be a lock a forgotten window holds for ever | `internal/engine/engine.go`, `following` |
| looking at a backup, or at the store before backing it up, opens the file to be read and nothing else | opened the shared way, a store that was not there was made readable by anybody, and a backup looked at was rewritten and its date moved | `internal/store/backup.go`, `heldIn` |

## The surfaces

| decided | why, in one line | where |
|---|---|---|
| every change goes through the engine; a surface decides nothing | the rules live once | [architecture.md](architecture.md) |
| what kind of answer a failure is — refused, not understood, store, unforeseen — is the engine's | a script's exit code and an agent's refusal are read off the same list | `internal/engine/refusal.go` |
| `--json` is a flag of every command, including those that never answer in JSON | one flag at the root reads the same everywhere; on `completion` or `gui` it changes nothing and costs nothing | `internal/cli/cli.go` |
| no `--dry-run` on what loses work | without `--yes` it is refused before anything is touched, and the refusal names what would have gone — that is the dry run | the command line's `newActionCmd` |
| what a terminal would obey, or what formats text rather than shows it — every Unicode Cf character, emoji joiners included — is refused when written and replaced when shown | the category is the rule, since a list of the characters somebody thought of let the rest through, tag characters an agent reads included; the refusal keeps it out from now on; showing it replaced covers what was kept before the refusal existed, or comes back with an older backup | `internal/domain/title.go`, `Prose`, `Shown` |
| no limit on a listing | a row carries facts and never what a task says, which is what made a listing too long for a client; a limit comes the day a board that large is measured | [listing.md](listing.md#a-listing-hands-over-the-facts-and-leaves-what-a-task-says-to-be-shown) |
| the contract with an agent is versioned by its commits until the first tag | a break says `BREAKING CHANGE` in the commit that makes it, and the schemas are pinned so it cannot be made without one; a version number with nothing tagged would number nothing | [mcp.md](mcp.md#what-changes-in-a-tools-answer-and-what-does-not) |
| the MCP server looks only, unless told otherwise | an agent can be talked into a write by what it reads | [mcp.md](mcp.md#what-an-agent-may-do-is-settled-when-the-server-starts) |
| a tool above the access level is absent, not refused | a refusal invites another try; an absence ends it | [mcp.md](mcp.md#what-an-agent-may-do-is-settled-when-the-server-starts) |
| a refusal's kind travels under `_meta`, not `structuredContent` | the client measures structured content against the answer's shape even on a refusal | [mcp.md](mcp.md) |
| a tool whose answer shape cannot be derived is offered without one, rather than stopping the server | a test proves every tool derives one, so this can only happen to code nobody ran | `internal/mcp/tools.go` |
| the two boards share where a card goes, not how the cursor walks | they draw the stalled band differently — a grid of three in the terminal, lanes under the columns in the window — so the same key is a different journey | [architecture.md](architecture.md) |
| `internal/reading` both reads the board and builds what a board asks the engine | both boards need both halves, and neither half is used without the other | [architecture.md](architecture.md) |
| the engine is one package, split into files by notion | one door (`Ask`) and one place for the rules; sub-packages wait until a file cannot be read on its own | [architecture.md](architecture.md) |

## The checks

| decided | why, in one line | where |
|---|---|---|
| what the tools find stays current; what the tools are is pinned | a new rule found something already true; a tool changing its own behaviour found nothing | [checks.md](checks.md#getting-them) |
| no threshold is lowered | a gate that bends is not a gate | [checks.md](checks.md#thresholds) |
| `&&`/`||` and `continue`/`break` mutated in the core only | that is where the rules are; the surfaces gave mostly mutants no test can tell apart | [checks.md](checks.md#thresholds) |
| packages are mutated one after another | gremlins already uses every core inside a package; two at once only made timeouts | [checks.md](checks.md#running-them-by-hand) |
| Linux is checked by hand, in a container, before a version | it needs Docker, which no hook may assume | [release.md](release.md) |
| the container's image is named by tag and Debian release, not digest; commitlint's own dependencies are not locked | the first is a check run by hand, not a gate, and a digest would be moved at every Go release for nothing anyone notices; the second runs only on a message, locally, and locking npm's tree would mean a lockfile in a Go repository for one hook | `scripts/linux.sh`, `prek.toml` |
| a form's commands are raced against five milliseconds in the terminal board's tests | a cursor blink is among them and is a timer nobody can wait out, and a command cannot say what it is before it runs; raising the deadline made the suite twelve times slower, as the test itself says | `internal/tui/fake_test.go`, `quickly` |

## Known and not yet answered

What a review found and nobody has decided yet. It is written here so it is
not found a second time as news, and it leaves this list when it is fixed or
decided.

| what | why it matters | where |
|---|---|---|
| a person's error is printed with its first letter capitalised, quoted value included | `"abc"` reads `"Abc"`, a tab reads `\T`: what was typed is misquoted, and differs from the JSON | the command line's error handler |
| a closed set refused by a narrowing says `not-one-of`, refused by an edit says `not-a-state` | one mistake, two reasons to branch on | `internal/engine/refusal.go` |
| `migration rollback` does not read the ledger again inside its transaction, and `migration apply N` applies N with an earlier change missing | two at once both report undoing the same change; a change applied out of order gives a shape no sequence describes | `internal/store/migrate.go` |
| a SQLite file that is not cairn's, holding no table cairn knows, is given cairn's shape | it is reported as not a store only when it holds something cairn reads | `internal/store/migrate.go`, `isEmpty` |
