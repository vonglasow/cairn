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
| no recurrence, no work that never finishes | cairn models work that reaches `done`; turning a real problem into a mechanism was not asked for, and a meeting that comes back belongs to the calendar | [ordering.md](ordering.md#weighed-and-left-out) |
| no precedence between two tasks | measured: it constrains without ordering | [ordering.md](ordering.md#weighed-and-left-out) |
| how long a task has sat in its column is read off the trace, never stored | the one date nothing else implies, and the trace already holds it | [ordering.md](ordering.md#what-the-store-gains) |
| no `unblock` verb | `--blocker -3` does it, and returning to `todo` belongs to the edit | [blocking.md](blocking.md#deliberately-not-there-yet) |
| no "what makes a task doable now" | two properties were found and the list plainly does not close | [ordering.md](ordering.md#deliberately-not-here-yet) |
| no task hangs under another; a task is split into steps, a title and a tick each | a sub-task was a whole task for what was most often one line to tick, and its parent and sisters ran through every rule and every surface | [steps.md](steps.md#why-steps-replaced-sub-tasks) |
| a link between two tasks has no kind and no direction; a task done is never reopened, a linked task is opened | a kind would say again what a label already says, and reopening a task rewrites the period, the time and the story it was finished in | [links.md](links.md) |
| no `task next` | what it would answer is already asked plainly — `--due-before`, `--state todo`, `task focus` — and choosing among them is the person's order, which cairn does not override | this page |
| no follow-up day and no person awaited on a task watched | its `because` says who, its deadline says when, and both boards already mark a deadline passed | [blocking.md](blocking.md#three-words-for-stopped-not-one) |
| no ticket field; a ticket is named in the title or the description | the search reaches both, and a link is between two tasks, never to a place outside cairn | [links.md](links.md) |

## What is remembered, and what is not

| decided | why, in one line | where |
|---|---|---|
| a deletion keeps the task's history | otherwise the past changes every time something is deleted | [trace.md](trace.md#the-trace-outlives-the-task) |
| no undo, and no putting a deleted task back | reading and undoing are two different asks; the history makes it possible to work out, nothing does it | [trace.md](trace.md#deliberately-not-here-yet) |
| a deletion keeps what the task held, whole, in its last act, and nothing purges it | the history of a task is the reason the trace exists, and a deletion that kept nothing made even putting it back by hand impossible; a purge — `trace purge`, a task named, `--yes` — is built the day something written by mistake has to go, not before. Until then, what must not be kept is not written | [trace.md](trace.md#a-deletion-writes-down-what-went-with-it) |
| the timesheet gives the hours raw, by day and task | what a good day looks like is a judgement; grouping by objective or label was not asked for | [trace.md](trace.md#deliberately-not-here-yet) |
| no stand-up or weekly report in cairn | the trace (`trace list --since`) and the timesheet (`timesheet show --since --until`) already hand an agent every fact one would hold; writing it up is the /daily and /weekly skills', and the same report in two places is one too many | [trace.md](trace.md#deliberately-not-here-yet) |
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
| a refusal quotes what a person typed exactly as typed, with what a terminal would act on written out | what somebody typed is theirs and is never changed — not even its case, which fang's style used to change — and it reaches a screen shown, never obeyed | the command line's `saidToAPerson` |
| changes to the shape of a store are made in order, and a rollback reads the ledger again inside its transaction | each change was written on the shape the ones before it leave; two rollbacks read before either ran both said they undid the same change | `internal/store/migrate.go` |
| a database holding tables, and none of a store's, is refused before anything is written into it | a path named by mistake can lead to somebody else's database, and opening it the usual way already writes into it | `internal/store/open.go`, `refuseAnotherDatabase` |
| several tasks or objectives are deleted or archived in one act, under `--also`, all of them or none | three deletions asked one after another left, when the third was refused, two gone and somebody to work out which; one act is one transaction, and a new optional argument breaks no caller | `internal/engine/together.go` |
| a version's number is read off the commits by svu, and its tag is made by `make release` once every check has passed | a number typed by hand is a number somebody misread the history for, and a tag made before the checks says a failed build was a version | [release.md](release.md) |
| a release runs every check but the mutation sweep, which is run on its own and leaves a report | the sweep took the better part of an hour, and a release that waited on it was a release nobody made; every commit still mutates the lines it changed, and the report says what the last sweep found | [checks.md](checks.md#the-report-a-sweep-leaves) |
| a version is handed over as binaries, each built on the system it runs on, and no source archive | the window needs cgo and the system's graphics headers, which a cross-compile cannot reach, so each binary is built where it runs; whoever builds cairn clones it | [release.md](release.md) |
| `--json` is a flag of every command, including those that never answer in JSON | one flag at the root reads the same everywhere; on `completion` or `gui` it changes nothing and costs nothing | `internal/cli/cli.go` |
| no `--dry-run` on what loses work | without `--yes` it is refused before anything is touched, and the refusal names what would have gone — that is the dry run | `Operation.AsksFirst` in the engine |
| no comments in the code but one line per exported symbol, and no prose in the scenarios | a quarter of the code was comments, much of it the history of a line rather than its reason; the reasons live in `docs/` and in the commits | [design-principles.md](design-principles.md#comments) |
| which acts ask first is the engine's, and every surface asks the same ones | the boards asked before sweeping a column and the command line did not: one act, two rules. The engine refuses an act that loses work, or sweeps up what a narrowing reaches, unless it comes confirmed; `--yes`, a board's question and an agent's access are only how each surface says it asked | `Operation.AsksFirst`, `Input.Confirmed` |
| what a terminal would obey, or what formats text rather than shows it — every Unicode Cf character, emoji joiners included — is refused when written and replaced when shown | the category is the rule, since a list of the characters somebody thought of let the rest through, tag characters an agent reads included; the refusal keeps it out from now on; showing it replaced covers what was kept before the refusal existed, or comes back with an older backup | `internal/domain/title.go`, `Prose`, `Shown` |
| no limit on a listing | a row carries facts and never what a task says, which is what made a listing too long for a client; a limit comes the day a board that large is measured | [listing.md](listing.md#a-listing-hands-over-the-facts-and-leaves-what-a-task-says-to-be-shown) |
| the contract with an agent is versioned by its commits until the first tag | a break says `BREAKING CHANGE` in the commit that makes it, and the schemas are pinned so it cannot be made without one; a version number with nothing tagged would number nothing | [mcp.md](mcp.md#what-changes-in-a-tools-answer-and-what-does-not) |
| the MCP server looks only, unless told otherwise | an agent can be talked into a write by what it reads | [mcp.md](mcp.md#what-an-agent-may-do-is-settled-when-the-server-starts) |
| a tool above the access level is absent, not refused | a refusal invites another try; an absence ends it | [mcp.md](mcp.md#what-an-agent-may-do-is-settled-when-the-server-starts) |
| a refusal's kind travels under `_meta`, not `structuredContent` | the client measures structured content against the answer's shape even on a refusal | [mcp.md](mcp.md) |
| a tool whose answer shape cannot be derived is offered without one, rather than stopping the server | a test proves every tool derives one, so this can only happen to code nobody ran | `internal/mcp/tools.go` |
| the Claude sessions holding a task are the hook's, not the engine's | the engine would write and read the session entries in a task's notes, which cairn never looks at; it keeps only what every surface needs, letting go of a focus and asking for a session to be renewed | [claude-code.md](claude-code.md#where-it-lives-and-how-it-is-installed) |
| the two boards share where a card goes, not how the cursor walks | they draw the stalled band differently — a grid of three in the terminal, lanes under the columns in the window — so the same key is a different journey | [architecture.md](architecture.md) |
| what a board reads lives in the engine, not in a package between it and the boards | a surface is an entry point; the package in between grew decisions of its own, and its search drifted from the engine's | [architecture.md](architecture.md) |
| the engine is one package, split into files by notion | one door (`Ask`) and one place for the rules; sub-packages wait until a file cannot be read on its own | [architecture.md](architecture.md) |

## Sharing

| decided | why, in one line | where |
|---|---|---|
| no decentralised store (IPFS), no CRDT library, no synced store file | public by default and undeletable; every CRDT engine fitting SQLite needs C; a file synced mid-write is a corrupt store | [sharing.md](sharing.md#study-1--a-decentralised-store-16) |
| sharing, when it is needed, is the trace merged over a relay of the same binary | the trace is already a log, and a relay is where a deletion reaches everybody | [sharing.md](sharing.md#study-1--a-decentralised-store-16) |
| no blockchain | colleagues of one employer trust each other, and a chain cannot forget what the GDPR says must be forgettable | [sharing.md](sharing.md#study-2--a-blockchain-17) |

## The checks

| decided | why, in one line | where |
|---|---|---|
| what the tools find stays current; what the tools are is pinned | a new rule found something already true; a tool changing its own behaviour found nothing | [checks.md](checks.md#getting-them) |
| no threshold is lowered | a gate that bends is not a gate | [checks.md](checks.md#thresholds) |
| `&&`/`||` and `continue`/`break` mutated in the core only | that is where the rules are; the surfaces gave mostly mutants no test can tell apart | [checks.md](checks.md#thresholds) |
| packages are mutated one after another | gremlins already uses every core inside a package; two at once only made timeouts | [checks.md](checks.md#running-them-by-hand) |
| Linux is checked by hand, in a container, before a version | it needs Docker, which no hook may assume | [release.md](release.md) |
| the container's image is named by tag and Debian release, not digest; commitlint's own dependencies are not locked | the first is a check run by hand, not a gate, and a digest would be moved at every Go release for nothing anyone notices; the second runs only on a message, locally, and locking npm's tree would mean a lockfile in a Go repository for one hook | `scripts/linux.sh`, `.pre-commit-config.yaml` |
| a form's commands are raced against five milliseconds in the terminal board's tests | a cursor blink is among them and is a timer nobody can wait out, and a command cannot say what it is before it runs; raising the deadline made the suite twelve times slower, as the test itself says | `internal/tui/fake_test.go`, `quickly` |

## Known and not yet answered

What a review found and nobody has decided yet. It is written here so it is
not found a second time as news, and it leaves this list when it is fixed or
decided.

Nothing, at the moment.
