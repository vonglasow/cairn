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
| a backup is refused as a restore when SQLite cannot read it whole | a damaged backup put in place trades a store that opens for one that does not | `internal/store/backup.go` |

## The surfaces

| decided | why, in one line | where |
|---|---|---|
| every change goes through the engine; a surface decides nothing | the rules live once | [architecture.md](architecture.md) |
| what kind of answer a failure is — refused, not understood, store, unforeseen — is the engine's | a script's exit code and an agent's refusal are read off the same list | `internal/engine/refusal.go` |
| no `--dry-run` on what loses work | without `--yes` it is refused before anything is touched, and the refusal names what would have gone — that is the dry run | the command line's `newActionCmd` |
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
