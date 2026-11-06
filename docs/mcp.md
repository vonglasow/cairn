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
hands an agent in one answer. And `task_list` and `task_search` take `fields`,
the fields each row carries — `["number","state","title"]` — for a board too
long even then. `task_search` takes `any`, to keep what carries any one of
what is named rather than all of it, and then says which tasks each one
matched. `task_list` takes `number`, the tasks to answer and no other —
`[25,18,330]` — and says under `unknown` which number is no task, so a
handful of known tasks is one call rather than one `task_show` each.
`task_list` also takes `limit` and `offset`, a page of what it keeps in the
board's order, and then says under `total` how many there are in all; and
`count` — `"state"` or `"label"` — to answer no row, only how many sit in each
column or carry each label. See [listing.md](listing.md).

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

## A cairn installed while it runs takes over, with no reconnect

A client starts the server once a session and keeps it. Before, an upgrade —
`make install`, `brew upgrade`, `cairn upgrade`, a release — left the session talking to the old
binary until somebody reconnected it by hand (`/mcp` in Claude Code): the tools
added since were missing, and a configuration the new cairn wrote could be
refused by the old one. Now nothing is left to do
(`@mcp-a-newer-cairn-takes-over`).

Between two requests, the server compares the `cairn` the client started — its
path, found on the `PATH` when it was a bare name, links followed — with the
one it is: the file it resolves to, its size, its time. One `stat` a request.
When they differ, the request just read is not answered there. The server
waits for the calls it is still answering, then replaces itself with the cairn
on disk, on the same standard input and output, and hands it what it holds:

- the greeting — what the client said at `initialize` and that it was
  initialized, and the protocol they agreed on — so the client greets nobody
  again;
- the `subscriptions/listen` it holds open, the one call never answered, so
  the newer protocol's notifications still reach the client;
- what it had read and not answered, the request first;
- the names of the tools it offered.

The newer cairn answers the request first. If its tools are not the ones
handed over, it then tells the client the list changed
(`notifications/tools/list_changed`) and the client reads it again; with the
same tools it says nothing (`@mcp-the-same-tools-are-not-announced`).

On macOS and Linux the server `exec`s the newer cairn: the same process, so the
client sees no process end. Windows has no `exec`; the server starts the newer
cairn on the same handles, waits for it, and leaves with its code. What is
handed over travels in one environment variable, `CAIRN_MCP_CARRIED`, which
the newer cairn reads and removes as it starts.

Three choices, and why:

- **Reading one message a line, ourselves.** The SDK's `IOTransport` reads
  ahead of what it answers, so bytes would be lost at the hand-over. cairn's
  transport reads a line at a time and hands over whatever it buffered. It
  takes no JSON-RPC batch, which protocol 2025-06-18 removed.
- **Something that cannot be run leaves the running cairn answering**
  (`@mcp-what-cannot-run-leaves-the-running-cairn`). A file half copied, or
  not a program: the `exec` fails, the request is answered by the cairn still
  running, and the file is taken as seen — the next request does not try it
  again; the next install does. A cairn missing from its path, between the
  removal and the copy of an install, is no change at all.
- **The identity is the file, not the version it says.** Asking a binary its
  version means running it at every request; a `stat` costs nothing. A cairn
  installed again, unchanged, is handed over to all the same, which costs one
  `exec`.

What it cannot do: hand over to a cairn older than this one, which knows
nothing of `CAIRN_MCP_CARRIED`. Going back below it needs a reconnect, once.

Set aside, on 2026-11-03: running the binary for every call (that is the
command line over MCP, which the server is there to avoid); exiting when stale
(Claude Code does not start again a server that exits); a server on a local
port, which still needs restarting after an update and opens a port — a task
of its own, decided with [sharing.md](sharing.md).

## A page turned on or off while it runs changes its tools at the next call

The tools a server offers are those of the pages the configuration turns on
(see [config.md](config.md#the-pages-turned-off-one-by-one)). They used to be
read once, when the server started, so a page turned back on mid-session
offered nothing until somebody reconnected — while every other part of the
configuration was already read at each call. Now, between two requests, the
server reads the pages again, and when the tools they give differ from those it
last told the client of, it takes away the ones gone, offers the ones new, and
the SDK tells the client the list changed
(`@mcp-a-page-turned-back-on-is-offered-at-the-next-call`,
`@mcp-a-page-turned-off-is-taken-away-at-the-next-call`). The request just
read is answered with the tools as they are now: a tool of a page just turned
off is unknown to it, as it would be to a server started now.

- **It is the same rule as at the start, asked again.** `engine.PagesFrom`, then
  the access level: a page turned back on under `--access read` brings back its
  listing and showing, not its writing
  (`@mcp-a-page-turned-back-on-stays-within-the-access`). A configuration that
  cannot be read turns no page off, at the start or mid-session, so a file
  broken while the server runs brings every page's tools back, and every call
  that needs the file refuses and says why
  (`@mcp-a-configuration-that-cannot-be-read-mid-session-turns-no-page-off`).
  Keeping the last list instead would have needed a second reading of the
  file, with its own rule for what an error means; a server started on that
  same file offers those tools anyway.
- **Nothing changed, nothing said.** A configuration written again with the
  same pages, or a call with nothing written, sends no notification
  (`@mcp-pages-unchanged-are-not-announced`): the names are compared, not the
  file.
- **It goes through the hook the hand-over built.** The same transport, the
  same moment — once the server has answered once, then before each request it
  reads — and the same way of telling: adding and removing tools through the
  SDK, never a notification written by hand. Not before the first answer, for
  the reason a newer cairn waits for it: a `subscriptions/listen` handed over is
  only parked again once that answer is out. What a server hands over to a
  newer cairn is the list it last told the client of, so the newer one tells
  the client only what is new to it.

The price is one read of the configuration a request, which every call that
opens the store already pays.

## It tells an agent where its work goes, as it greets it

The server's greeting carries instructions, and they say one thing: a task's
description says what the task is, and what was done or decided on it goes in
its notes — the directory `task_show` names as its context — never in its
description (`@mcp-greeting-says-where-the-work-goes`). `task_edit`'s `desc`
says the same where it is filled in. An agent driving the command line reads
it in `--desc`'s help, which is the same sentence: it is the engine's.

It is said because it was not: an agent rewrote a description with a
session's log for want of being told where else to put it — see
[context.md](context.md). Nothing refuses a description that reads like a log:
a description may well hold dates, and a refusal would forbid that too.

## A piece of a description changes in one call

`task_edit` and `objective_edit` take `replace` and `by`, as the command line
takes `--replace` and `--by`: the piece named is put in place of, and the rest
is left as it was. Fixing one commit hash on twenty tasks used to be, for each,
`task_show`, an edit outside cairn and the whole description sent back — three
calls and a full rewrite where a slip anywhere lost text. It is one call now.

The piece has to be there exactly once, as the artifact store's `str_replace`
asks: not there, or there twice (the refusal says how many times), and nothing
is written — more of it is given, and the place is beyond doubt. `by` empty
takes the piece out. `desc` beside `replace` is refused: the description is
changed whole or in part, not both. No regular expression, no several pieces in
one call, nothing but the description: what was asked for, and nothing beside.

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

The engine refuses an act that asks first unless it comes confirmed. A tool
call comes confirmed: the access the server was started with is how the person
said yes to what an agent may do, once, ahead of time.

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

`cairn upgrade` is not a tool either, and not an operation of the engine's
list: it reaches GitHub, and every tool says it reaches nothing outside the
store (`@mcp-marks-what-it-never-reaches`). Replacing the binary an agent's
own server runs from is for the person at the terminal to decide; once they
have, the server takes the new one over on its own, above.

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

## A focus session past its length asks, in every answer

Once a focus session has been up for the length `[focus] session` gives, every
answer carries one more text after its own: the question whether to renew the
session or end it (`@mcp-any-answer-asks-to-renew-a-focus-session`, see
[config](config.md#a-focus-session-that-asks-to-be-renewed)). It goes in the
text, not the structured answer: the text is what an agent reads whatever it
called, while the structured shape is the tool's own and says only what the
tool answers. The engine says the question (`Answer.Asked`); the server only
puts it after the answer.

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

depguard's `testing` rule, which refuses the import anywhere but a test, leaves
`test/` alone, since it does import `testing` on purpose.

What guards `test/` instead is narrower and truer to the worry behind it — not
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
an empty string for a deadline or an estimate, an empty list or `null` for
labels, blockers, objectives and steps — and each field's description in the
schema says which. What a whole answer may leave out — the context, since when
a task has stood in its column — is left out entirely.

The one exception is a row of `task_list` or `task_search` called with
`fields`: it carries the fields named and no other. So a row's schema requires
none of its fields, on every tool whose answer holds rows. A client measures
every answer against the schema the tool declared — the scenarios do too
(`@mcp-a-listing-carries-the-fields-asked-for`) — and a row carrying three
fields where fourteen were required would have refused the very answer that
was asked for. Called without `fields`, a row still carries every one
(`@mcp-a-listing-asked-for-no-field-carries-every-one`). The schema is read off
`engine.Row` as before, and only its `required` is dropped, in `internal/mcp`,
where the schema is made.

A list argument that takes a closed set says it on its items: `fields` is an
array whose items are an `enum` of the fields a row carries
(`@mcp-names-the-fields-it-can-answer`), where a single value says it on
itself.

**A list of numbers takes numbers.** A list that names tasks or objectives by their numbers — `number` on
`task_list`, `blocker`, `link`, `objective`, `also` — takes each one written
as a number or as a string: its items are `{"type": ["integer", "string"]}`
(`@mcp-a-list-of-numbers-says-it-takes-numbers`). An agent writes
`"number": [25, 18]` because that is what a number is, and it used to be
refused, the refusal printing the integer in quotes as if a string had been
sent (cairn#241).

- **Strings stay.** `blocker` takes `"+3"`, `"-3"` and `"=3"`, which are no
  numbers, and every list was strings until now: taking them away would be a
  break.
- **A number is read as the digits it is written with**, then handed on
  exactly as the same string would be, so the rules stay where they were and
  are written once: `-3` among blockers takes 3 away as `"-3"` does, and
  `-3` among the numbers `task_list` answers is refused as `"-3"` is.
- **Only a whole number written whole is one.** `25.0`, `25.5` and `1e3` are
  refused as their strings are — the same rule a single number argument such
  as `task_show`'s already followed — rather than rounded into a task nobody
  named (`@mcp-what-is-no-number-is-refused-as-it-was-sent`).
- **The rule is the engine's.** `Argument.TakesNumbers` — a list whose `Of`
  names what it numbers — is read by `FromJSON` to take the numbers and by the
  server to write the schema, so the two cannot disagree. The command line is
  untouched: a word typed is a string either way.
- **A refusal says what was sent, as JSON.** A number among labels is refused
  with `got 3`, a string with `got "3"`, a yes with `got true`
  (`@mcp-a-number-among-labels-is-refused-as-a-number`): the value is written
  back the way JSON writes it, so the quotes say which it was.

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
