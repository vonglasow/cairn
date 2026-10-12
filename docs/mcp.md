# The MCP server

Read this before touching `internal/mcp`, or before adding a command that an
agent will see.

The scenarios are in [bdd/mcp.feature](../bdd/mcp.feature); this file carries
the reasons behind them, because a scenario says what happens and not why it
was chosen that way. They are being built one rule at a time, and what a rule
cost is written down here as it lands.

## It is a surface, and it decides nothing

`internal/mcp` sits beside `internal/cli`: it reads a call, hands it to the
engine, and answers. Every rule stays where it already lives. A rule that got
written here would be a rule an agent obeys and a person does not, which is two
cairns wearing one name.

The consequence worth stating: nothing in this file changes what cairn does. It
changes who may ask, and how the asking is shaped.

It also means what an agent reads is what a person reads. A row of a listing
said only `#12  Write the spec` for a long time, which left an agent opening
every task one at a time to find out which was urgent or late; the row now
carries the column, the priority, the deadline and the labels, and it does so
for both of them at once because there is only one place it is written.

It carries whether a task has a description, and not the description: a board
of 105 tasks listed with theirs came to 106 000 characters, past what a client
hands an agent in one answer. See [listing.md](listing.md).

## It is started like every other command

`cairn mcp serve`, because everything else reads `cairn <thing> <verb>` and a
server is no reason to break that. `mcp` is a group like `task` or `backup`,
and it has room for a second verb the day one is wanted.

It speaks over standard input and output. A client starts it as a child process
of its own, and no open port is the right default for something one person runs
on one machine.

```json
{
  "cairn": {
    "command": "cairn",
    "args": ["mcp", "serve", "--access", "full"],
    "env": { "CAIRN_STORE": "/Users/…/.cairn/cairn.db" }
  }
}
```

## What an agent may do is settled when the server starts

One flag, three levels, `read` by default:

| `--access` | what it reaches |
|---|---|
| `read` | listing and showing, nothing else |
| `write` | also adding, changing, `task_export`, `task_import`, `backup_new`, `migration_apply` |
| `full` | also `task_delete`, `objective_delete`, `backup_delete`, `backup_restore`, `migration_rollback` |

The line between `write` and `full` is not "destructive" in the technical sense
— it is **what can lose work already done**. Adding a backup loses nothing, so
it sits in `write`; rolling a migration back drops what that migration carried,
so it sits in `full`. `migration_apply` is in `write` for the same reason: it
changes the shape but sets a backup aside first.

**Nothing given is looking and nothing else.** The default was `write`, and an
agent reads what it is handed — a description, a label, a note — before it
acts; one talked into a write by what it read reached as far as nobody had
decided it should. A server started without a word now reaches no further than
listing and showing, and every level above that is one somebody typed. The
configuration above names `full` because that is a decision its owner made,
not one the server made for them.

**A tool above the level is not exposed at all**, rather than exposed and
refused. A refusal invites another attempt, and an agent that has been told no
will reword and try again; an absence ends the matter. It also keeps the
conversation shorter, which is its own kind of safety.

Editing is ordinary. `task_edit` can empty a description or replace a whole set
of labels, and that was weighed and left alone: editing a task is what a task
manager is for, and a confirmation asked twenty times is a confirmation nobody
reads by the twenty-first.

## Two guardrails, and only one of them is ours

The MCP protocol lets a tool carry annotations — `readOnlyHint`,
`destructiveHint` — and clients use them to decide when to ask the person
before calling. That is the natural shape of the command line's `--yes`: the
question goes to a human, asked by the client, at the moment of the call.

But an annotation is a hint. Nothing obliges a client to honour it. So the
access level is enforced server-side, where no client can talk it away. The two
are not redundant: the annotation asks, the level forbids.

What the annotations say is read off the operation, never written per tool:

| annotation | said when |
|---|---|
| `title` | always — the operation's own short sentence |
| `readOnlyHint` | the operation is at `read` |
| `destructiveHint` | it loses work (`full`), **or it can replace what was written** — `task_edit` empties a description as surely as a deletion does, which is the protocol's sense of the word and wider than the access level's |
| `idempotentHint` | a second call with the same arguments does nothing more than the first, or is refused for having nothing left to do: every read, an edit, a move, a focus — never an addition, a backup made, time added |
| `openWorldHint` | never: cairn reaches nothing outside the store |

The last two facts are the engine's — `Again` and `Overwrites` on an
operation, beside its level — since whether a second call costs anything is a
property of what the operation does, not of the surface that offers it.

Each tool's input schema says what it takes and nothing else
(`additionalProperties: false`), with the type of what a list holds. The engine
refused a name an operation does not take all along; said in the schema, a
client knows before it calls.

## The store is opened for each call and closed again

A command opens the store when it needs it and closes it on the way out, and
[architecture.md](architecture.md) says that is what makes "a task added by one
command is found by the next" a claim about the disk rather than about a cache.
A server is only a command asked more often, so it does the same.

Measured, on a store of 200 tasks:

| | |
|---|---|
| opening and closing, doing nothing | 0.66 ms |
| reading all 200 tasks, already open | 1.57 ms |
| the whole call | 2.89 ms |

Opening is a quarter of the work it serves, and invisible next to a round trip
through a client and a model.

Three things follow, and each would otherwise be a problem of its own:

- The invariant holds. The server is not the one place where "written by one,
  seen by the next" stops being true.
- Nothing holds a write-ahead log open between two calls. A backup reads the
  store as it stands either way, but a log nobody holds is a log nothing can
  lose.
- Nothing holds the file while `backup_restore` or `migration_rollback` swaps
  it underneath. Under a live handle that is corruption, not staleness.

The places are found at each call too, not once when the server starts: the
server is handed the home and the environment, and the engine works out where
the store and the notes are each time it is asked. A configuration changed
while the server runs is read by the next call, and a configuration every other
call refuses still lets `config_show` say why — a server that refused to start
over it would have left an agent nothing to ask.

## A client that hangs up is not a failure

Measured on the built binary, and it was not obvious: a client that greets the
server and closes its end in the same breath used to make `cairn mcp serve`
exit 2 with `Server is closing: EOF`, and the greeting went unanswered. An
empty standard input already exited 0, and so did a client that greeted,
waited, then closed. Only the request still in flight failed.

Half of that is ours and half is not, and the scenario says only the half that
is. A client owns the process it started and ends it by closing the pipe, so
ending is not failing — that part is now true. Making the greeting be
*answered* before the server stops is not: the SDK tears the session down on
EOF and drops what is in flight, and holding EOF back would mean guessing when
the server is done. It costs nothing real, since the asker has gone by then.

Recognising it costs one number. The SDK's own sentinel lives in its
`internal/jsonrpc2`, which nothing outside may name; the public handle is the
`jsonrpc.Error` alias and its code, `-32004`, which the SDK exports no name
for. So cairn writes the number, in a statement rather than a constant — Go
instruments statements, so a constant would put the one thing this depends on
outside anything a test can be seen exercising, and the mutation gate would
report it unjudged rather than proven. A test pins it: if the SDK changes it,
that turns red here instead of going quiet in somebody's terminal.

## One tool per command, named after it

`task_new`, `objective_show`, `migration_rollback` — the same grammar as the
command line, so anybody who knows one knows the other.

The count is never written down. A scenario saying how many tools there are
would freeze a number in text the process forbids rewording, and the twentieth
command would break the contract instead of the code. The scenario says **a
tool for every command**, and a test walks the cobra tree to prove it.

`cairn mcp serve` is a command and not a tool. It is how the server is started,
not something the server can be asked to do, and a scenario keeps it out.

## Each tool says what shape its answer takes, and that shape is an object

A tool describes its arguments and used to say nothing about its answer, so a
client had to guess. Declaring the shape closed that. But the first attempt at
declaring it made things strictly worse, and the reason is worth keeping.

**What was wrong, and how it was found.** The declaration was read off the Go
type, and several operations answered a bare list — `[]Task`, `[]trace.Shape`.
Their schemas therefore said `"type": ["null", "array"]`. `task_focus` answered
two shapes, so its schema said `anyOf`. All of that is legal: the SDK's own type
says `structuredContent` "may marshal to any valid JSON value (object, array,
or primitive) conforming to the tool's OutputSchema", per SEP-2106. Legal and
unusable — the client on the other end validates every `outputSchema` as an
object, refuses the ten that were not, and **a single refused tool costs the
whole list**. Before declaring, that client fetched the tools with a complaint.
After declaring, it fetched none. Measured on the real client, not reasoned
about.

Two mistakes, and only one of them is technical. Declaring a schema was right;
declaring it without asking the one client that was going to read it was not.
And no gate could catch it, because every gate here tests cairn against cairn.
A scenario now states the client's constraint outright — *every tool says it
answers one object* — so the requirement lives in the same place as everything
else that must hold.

**Every answer is one object naming what it carries.** A listing of tasks is
`{"tasks": [...], "in_focus": 2}`, not `[...]`. Three reasons, and only the
first is somebody else's rule:

- a client refuses a description that is not of an object;
- a bare list can never gain a sibling field, so what a listing knows besides
  its rows had nowhere to go — the focus was repeated on every row, and what
  the store holds was dropped from `migration_list` entirely though the text
  had always said it;
- a reader can tell one listing from another without inspecting a row.

**The focus is said once, at the top.** Every task in a listing used to carry
`in_focus`, which is a fact of the store restated once per row. Restated, the
rows can disagree, and a reader has to pick one to believe. It also puts the
two renderings back in step: the card a person reads has never shown the focus,
only the listing marks it.

**An operation answers one shape, not a set of them.** `task_focus` used to
answer two — the task being worked on, or nothing in focus — and its schema
said "either". A description that says "either" is one the reader has to
resolve at runtime, which is the thing that broke. So the operation was changed
rather than the description: it answers `{"task": …, "in_focus": …}` either
way, with the fields it cannot fill left empty. A machine reads one field of
one shape whether or not anything is held.

**The shape is read off the Go type, never written out.** An operation declares
what it answers — `Answers: Tasks{}` — and the schema is inferred from that
type. The same struct that crosses the wire is the one the description is taken
from, so the two cannot drift: change the struct and the schema changes with it.

**A refusal says what kind it is, beside the sentence.** Under `_meta`, at
`cairn/refusal`, a refusal carries `{"error", "kind", "reason"}`: the kind is
`refused`, `usage`, `store` or `failed`, and the reason is a word that does not
change when the sentence is reworded — `no-such-task`, `title-required`,
`not-understood`. An agent branching on "is this a typo or a missing task"
read prose until then. The kinds and reasons are the engine's
(`engine.RefusalOf`), the same ones the command line's exit code and its JSON
refusal are read off, so a script and an agent are told the same thing.

**A call the client gave up on is not run.** A cancellation that arrives
before a call starts stops it there. One that arrives after has nothing to
stop: a call is one transaction of a few milliseconds, and stopping it half way
is what the transaction is there to prevent.

**A refusal is not an answer of that shape.** It comes back as `IsError` with
the reason as text and carries no `structuredContent`, so nothing measures it
against a schema it was never meant to satisfy. Measured, then frozen in a
scenario: without it, every refusal would read as a violation. That is why its
kind travels under `_meta` and not there.

And the declaration bought something that was not the point of it. `Of` works
out which shape to build by looking at which field of an `Answer` was filled —
knowable only after the operation has run, and until now comparable to nothing.
A test now runs every operation `All()` names and checks that what `Of` builds is what
the operation said it answers. The description and the behaviour check each
other.

## Why the SDK, written down

`CLAUDE.md` asks for a written reason before a dependency, so here it is — and
its line in [dependencies.md](dependencies.md), where every one is listed.

`modelcontextprotocol/go-sdk` v1.7.0 carries the protocol and derives each
tool's JSON schema from its Go argument type. Writing it by hand was measured
at 250–350 lines against 35 with the SDK, and the schema would then be a second
thing to keep in step with the signature.

It brings seven transitive dependencies it actually reaches —
`github.com/google/jsonschema-go`, `golang.org/x/oauth2`,
`github.com/segmentio/asm`, `github.com/segmentio/encoding`,
`github.com/yosida95/uritemplate/v3`, `golang.org/x/sync`, `golang.org/x/time`
— most of them for the HTTP transport a stdio server never calls. That was the
argument against, and it does not hold: Go links only what is reachable, and
`govulncheck` runs over the lot at every commit. The argument that does hold is
the project's own: look for it first, and do not write twice what exists once.

## The test harness has to be shared, and how

The scenarios need a store, an environment, and a way to run a command. All of
that was `internal/cli/world_test.go` — 726 lines that `internal/mcp` could not
import, because Go compiles `_test.go` files only for their own package.

The shape that works:

```
test/          the harness and the world. Imports "testing", and no
               surface: it is handed a "run this command" function and
               never learns who answers. Tested like anything else.
internal/cli/  keeps its own scenarios, in package cli
internal/mcp/  keeps its own, in package mcp_test
```

`internal/mcp`'s scenarios are in a **package of their own** — `mcp_test`, not
`mcp` — and that is forced, not stylistic. The `mcp serve` command lives in
`internal/cli`, so `cli` imports `mcp`; a scenario file in `package mcp` that
imported `cli` to run `cairn task new` would close the cycle. An external test
package may import both, and still counts toward `internal/mcp`'s coverage.

It also settles where the parity proof goes. Walking the cobra tree needs `cli`
and the tool list needs `mcp`, and a package that may import both already
exists — so there is no separate place for it to live.

Two things forced that shape. Had `test/` imported `cli` to run commands, then
`internal/cli`'s own scenarios — which are in `package cli` — would import a
package that imports them back, and Go refuses the cycle. And moving the
scenarios out of `internal/cli` altogether would empty its coverage, which is
measured per package.

That last line was wrong, and measuring it is what caught it. A package with no
`_test.go` does **not** fall outside the coverage gate: it appears in the
profile at 0%, and a harness of that size took the repository from 91.8% to
86.0% — below the floor. It only falls outside the mutation loop, which skips
packages with no test file.

So `test/` is tested like anything else, and that shaped it. Each shared step
is a named method on `World`, not a closure inside `Register`; a closure can
only be reached through godog, and a step that quietly stopped asserting would
have surfaced as somebody else's scenario passing for the wrong reason.
`Register` is then pure wiring and holds no judgement. What stayed out is the
one thing that cannot be reached without a real godog run: `test.Suite`
describes the run, and each surface keeps the four lines that fail its own test
when it does not pass.

`no-go-testing` is told to leave `test/` alone, since it does import `testing`
on purpose.

What replaces that guard is narrower and truer to the worry behind it — not
"nothing imports testing", but "nothing the binary is made of imports the
harness":

```sh
# finding something is the failure, not the other way round
go list -deps ./cmd/cairn | grep -q 'cairn/test'
```

Measured: `go list -deps ./cmd/cairn` reports none of godog, where the same
command with `-test` reports eleven. It sees what the binary is made of, and
nothing else.

## Nothing is open any more, and both answers came from the same place

Two things were open here, and neither was solved where it was written down.

**The shape a machine reads** was going to be lifted out of
`internal/cli/output.go` into a package both surfaces could read. It was, and
then that package was dissolved: a notion carries its own two renderings — the
one a person reads and the one a machine reads — and neither surface owns
either. So `task.Shape` is what `--json` prints and what a tool answers with,
and there is no third place for the two to drift apart in. See
[architecture.md](architecture.md).

**The build stamps** are `internal/build`, injected there by the linker and
read by both surfaces through `build.Read`.

What settled both is the same thing: the question "which surface owns this?"
had no good answer because neither did. A build is not the command line's, and
the shape of a task is not the server's.

## What changes in a tool's answer, and what does not

An agent holds on to the shape of an answer the way a script holds on to a
column. So the answers change by one rule:

- **adding a field breaks nothing** — a reader that does not know it ignores
  it;
- **renaming or removing one, or changing what it holds, is a break** — it is
  said in the commit that makes it, under `BREAKING CHANGE`, and the scenarios
  that pin the old shape are changed in the same commit, on purpose;
- **an argument is the same**: a new optional one breaks nothing, a renamed or
  newly required one does.

**A field that has nothing to say is there, empty.** A task's row carries all
its fields every time; one that names nothing says so with its empty value —
`0` for a parent when the task is its own, an empty string for a deadline or an
estimate, an empty list or `null` for labels, blockers and objectives — and
each field's description in the schema says which. What a whole answer may
leave out — the context, the sub-tasks, since when a task has stood in its
column — is left out entirely.

**Arguments are named the way the command line names them, answers the way
JSON does.** An argument is `due-before`, `with-archived`: it is the same
description the command line reads its flags from, and a second spelling
would be a second list to keep true. A field of an answer is `time_spent`,
`blocked_by`, `in_focus`. The rule is written here so it is a rule and not an
accident; changing either would be a break.

The schemas themselves are pinned: `internal/mcp/testdata/schemas.json` holds
every tool's input and output schema, its annotations and the level it is
offered from, as the server hands them out, and a test
fails when what the server says differs from it. A difference is either the
break the commit meant — and the file is rewritten with it, by running the
test with `UPDATE_SCHEMAS=1` — or one nobody meant, found before an agent did.
