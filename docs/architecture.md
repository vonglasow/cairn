# Architecture

Read this before adding a package or a command.

## A notion keeps its rules, and hands over its facts

A task's state is its own — no exported field — and every rule about it lives
on it: whether it may hang under another, archive, finish, serve an objective.
What it carries it hands over as one plain record, `task.Carried`, and `Made`
turns a record back into a task. The card a person reads, the shape a machine
reads, the row the store keeps, what a search looks in and the tally of what is
done all read that one record.

It used to tell its facts one method at a time to whoever implemented a
membrane, absence included. Eighteen membranes copied the same fields into
eighteen structures, and every fact added touched all of them. The rules are
still the task's; only the copying went. An objective, an act of the trace and
a migration work the same way.

An act names its task **by number** and never carries the task, which is why
`internal/trace` is forbidden `internal/task` outright: a task may be gone by
the time its acts are read. That is in `.golangci.yml`, where it is enforced.

**Errors do not go east.** Telling is for reading; a command that fails returns
its error the way Go returns errors. That line is deliberate: pushed further,
every call would grow a `Failed` method and Go would fight back on every one.

## The layers

The outer axis is the **notion** — the thing a person can name and ask for.
What is left over is either carried by every notion, or is the plumbing under
all of them.

| package | holds | knows about |
|---|---|---|
| `internal/domain` | what is **carried**: columns, priorities, labels, days, durations, progress, the title rule, what an absence looks like, the set grammar, a narrowing | nobody |
| `internal/build` | which build this is | nobody |
| `plugin/claude` | the Claude Code plugin's files, carried in the binary, and how they are written out as a marketplace — outside `internal/` because Go embeds only what lies under the package, and the files are read from there by a clone too | nobody |
| `internal/config` | the file that names cairn's two places, read the way a person wrote it | nobody of ours |
| `internal/task` | the task: its cell, its rules, and its two renderings | `domain` |
| `internal/objective` | the objective: its cell, its rules, and its two renderings | `domain` |
| `internal/migration` | a change to the shape of the store, and the list of them | `domain` |
| `internal/trace` | one act performed on a task, and what a period of them comes to | `domain` |
| `internal/store` | keeping the notions, and handing them back | `domain`, the notions |
| `internal/engine` | every operation and every decision: the calculations, where the places are found, the only door to the store, and what every board reads — cards and goals, which fields a form asks about, how the cards deal into columns and where a pushed card goes | `domain`, `config`, the notions, `store` |
| `internal/cli` | parsing and choosing a rendering, and which exit code an answer is | everything below |
| `internal/mcp` | the same, for an agent | everything below but `cli` |
| `internal/tui` | the board on a terminal | everything below but `cli`, `mcp` and `gui` |
| `internal/gui` | the same board in a window | everything below but `cli`, `mcp` and `tui` |
| `cmd/cairn` | hands the process to `cli.Main` and exits with what it returns | `cli` |
| `test` | the acceptance world the surfaces share, and the givens both boards write | `engine`, `config`, the notions |

**A surface is an entry point and nothing more.** The command line, the agent's
server, the terminal board and the window are four ways in to the same
actions. Each turns what it was given — a flag, a JSON argument, a key, a
click — into a call to the engine, and shows what the engine answered in its
own medium. None of them decides anything: not what a search reaches, not what
a card says, not how long a week was. Every one of those is the engine's, said
once, so one question asked on four surfaces cannot come back with four
answers.

What a board reads therefore lives in the engine too: the two listings turned
into cards and goals, that a deadline four days out reads `4d`, that the three
stopped states go under the last three columns, that a form's `labels` box
fills the engine's `label` argument. It used to sit in a package of its own
between the engine and the boards, and that package grew decisions the engine
did not know about — the boards searched with a rule the command line did not
have. `depguard` denies the engine every drawing library outright: it says what
a card carries, never how it is drawn.

**Where a card goes is shared; how the cursor walks is not.** `engine.Step`,
`Beside` and `Where` say where a pushed card lands and which column is beside
it, once, for both boards. The cursor is each board's own, because the two
draw the stalled band differently — the terminal as a grid of three across,
the window as three lanes under the columns they belong to — and the same key
is a different journey across a different shape. A cursor shared between them
would make one of the two drawings walk like the other.

**A label is not a notion.** Neither is a priority, nor a column, nor a
deadline. None of them has an identity: you cannot list one, show one or delete
one — they exist only on a task or an objective. So they are carried, they live
in `domain`, and they needed no home of their own. That is measured, not
asserted: `attributes.go`, `labels.go` and `spent.go` mention neither task nor
objective.

`domain` imports nothing of ours, and neither does a notion beyond it. An
entity that knew how it was written to disk would drag a database into every
test that touches it.

**A notion never imports another notion.** A task serves an objective by
carrying its number, not by knowing what one is; the store and the engine are
where the two are tied together. That is the one rule the notion-first shape
lives or dies by, and depguard holds it.

`cmd/cairn` stays one line. Everything put there is out of reach of the tests
forever — the suite calls `cli.Main`, never the binary — so each line added
there is a line the coverage floor pays for and nothing can check.

## Every change of state goes through the engine

A surface parses input, calls one thing, and prints the result. It decides
nothing.

Each of the things cairn can be asked to do is **described once**, in
`internal/engine`: its name, its arguments and what each accepts, the access
level it sits at, what it needs from the store, and how to carry it out. Both
surfaces are built from that description — the command line builds a cobra tree
from it, the MCP server builds its tool list — so a flag added tomorrow reaches
an agent without a line being written twice, and "one tool per command" is true
by construction rather than by a test.

The description sits with the engine and not with the notion it names, and that
is deliberate: `task show`'s work is three readings — the task, what hangs
under it, how far along it is — orchestrated. It does not talk about a task, it
ties three answers together, and tying things together is what the engine is
for.

An engine held open exports three methods and no more: `Ask`, `Close` and
`Version`. Every change goes through an operation, so it is checked the way a
command line's would be and kept as one transaction; the engine's own methods
for adding, editing or archiving are its operations' to call, and a test that
needs a store in a certain state writes it through the operations too — see
`test.Givens`.

A surface that stays on screen goes through `Engine.Ask` rather than
`engine.Do`. Same description, same refusals, one difference: `Do` opens the
store, runs the operation and closes it, and `Ask` runs it against an engine
somebody is already holding. A board redrawing every two seconds would
otherwise open a connection per keystroke to no purpose, and the engine keeps
the path it was opened on so an operation reaching for what sits beside the
store is still handed somewhere to look.

**The store is opened where it is needed, not at startup**, and that now lives
in one place: the description says what it needs — a store, a store whatever
shape it is in, or none at all — and one function opens it and closes it on the
way out. A command that has no use for it never opens it, which is why `cairn
--version` creates no file and why a number that is not a number is refused
before anything is written. One invocation opens one connection, so nothing
survives between two runs — which is what makes "a task added by one command is
found by the next" a claim about the disk rather than about a cache.

**The engine is one package, split into files by notion.** It is the one door
— `Ask`, and nothing else a surface calls to change something — so it holds
every operation and every rule a surface must not hold. Grown past ten
thousand lines, it is split by what each file is about — `task.go` for what is
done to a task, `moving.go` for its place and the focus, `trace.go` for its
history, `answer.go` for the shapes that cross to a machine and `render.go`
for the text a person reads, `refusal.go` for what kind of answer a failure
is — and not into sub-packages: a sub-package would need a door of its own, and
two doors is how a rule comes to be checked at one and not the other.

## More than one cairn at a time

One person, one machine, and still two cairns at once: a second terminal, a
shell loop, a server answering while the terminal is in use. They share a file,
so the store is opened in WAL with a wait of five seconds.

Both are measured, not precautionary. Without the wait, four processes writing
together lost fourteen writes out of twenty. With the wait but without WAL, six
mixed reads and writes spent the whole five seconds and still lost five of
them: in the default journal a reader blocks a writer, so writers starve. In
WAL the same six take fifty milliseconds.

**An operation that writes is one transaction.** A task is several rows — the
task, its labels, what blocks it, what it serves, its place among its sisters —
and what was done to it is one more. The engine runs every operation that
writes inside one transaction, so they are kept together or not at all, and the
transaction begins as a write: two cairns adding to one column at once take
their turn rather than both reading the same last place. The operations that
change the shape of the store are the exception, and keep their own: they copy
the file first.

WAL keeps recent writes beside the store, in a log, until the store is closed —
and a board holds it open for as long as it is on screen. So a backup is not a
copy of the file: it is the store as it stands, log included, read in one go
with `VACUUM INTO`. Only a file SQLite cannot read at all is copied as it lies,
because a damaged store is exactly when a backup is wanted. Putting one back
copies it beside the store, clears the log the old store left, and moves it into
place in one step.

**Opening reads the shape, it never lays it down.** `migration apply` is what
changes a store; opening one only looks. The ledger is the single exception,
created when it is missing because nothing can be recorded without it.

## The store is spoken to through an interface the engine owns

`internal/store` is where SQLite lives, and it is the only place that knows
that. It speaks in notions: it takes a `task.Task` and gives one back. No
`*sql.Rows` crosses its boundary, or every caller would need to know the
schema.

It never reaches inside one either. A task hands the store its record on the
way in, and the store makes a task out of a row's record on the way out.

The engine talks to a **storage interface it declares itself** for what it
reads and writes of the notions, and `store` satisfies it. What that buys is
said exactly: the rules are tested against what the engine names, and the
engine names only what it uses. It does not yet make a second backing — a
file, a remote — arrive unnoticed: the one transaction an act runs in, the
migrations and the backups still reach the store itself, and its refusals are
named after it. That door is opened the day a second backing is wanted, not
before. That is `docs/design-
principles.md` applied as written: accept interfaces, and let the consumer
define them. Go needs no import for that, so `store` still knows nothing of
`engine`.

Two things follow, and both are enforced:

- **The engine is the only caller.** Nothing but `internal/engine` talks to
  storage, whatever the implementation.
- **A surface never opens the store.** A command that read the database itself
  would compute a rule the engine already owns, and the next surface would
  compute it slightly differently. So `cli` cannot import `store`, nor
  `database/sql` — it asks the engine, which is the whole point of the engine.

## The shape of the store changes, and never on its own

What a task can carry grows, so the shape of the store has to grow with it. Two
rules hold that in place, and both are about who decides.

**Making a store is one act.** The whole shape goes down in a single
transaction, and whoever finds the file already made steps aside rather than
laying a second shape on it. Measured, and it is why: laying it down change by
change let another command open in the middle, read a ledger with one of the
three recorded, and report a store as behind when it was only being made. A
store half made is a store nobody may see.

**A store somebody else is making is waited out**, the way one somebody else is
holding already was. That wait is cairn's own and not SQLite's, because
SQLite's does not cover it: switching a fresh file to the journal cairn needs
takes a lock of its own. Both shapes of being held — a plain lock and a
transaction that began before somebody else committed — are read as the same
moment through the result's primary code.

**Nothing changes the shape unless it was asked to.** A store missing a change
this cairn knows is reported and left alone — `store.Open` refuses it, and only
the commands whose job is to change the shape use `store.OpenBehind`. Migrating
on the way past would mean `cairn task list` rewriting somebody's database as a
side effect of listing their tasks.

**A change is made whole or not at all.** Each one — the SQL that makes it and
the line in the ledger that records it — runs as one transaction, and so does
undoing one. A change of several statements failing half way used to leave the
first statements made and the ledger silent about them: a store the next
`migration apply` could only refuse, with nothing saying where it had stopped.
The backup set aside first was a way back, by hand; the transaction is the way
not to need it.

**Every change carries its own undoing.** A migration holds the SQL that makes
it and the SQL that unmakes it, and both are printable — `migration show`
exists so a person can read what a tool is about to do before letting it. The
first shape is the floor: it has no undoing, because unmaking it would mean
throwing the store away, which is what deleting the file is for.

## The backups are reachable without the store

`store.Vault` opens no database. That is not an oversight: a backup is most
needed exactly when the store cannot be opened — because it is corrupt, or
because it comes from a cairn that knows more. Needing a working store to reach
the backups would put them out of reach on the one day they matter.

**A backup did not become a notion**, and that was measured rather than
assumed. `store/backup.go` opens databases: saying what a backup *holds* means
reading the saved file as one. So it is not the plain file copy it looks like,
and it stays in the layer that knows what a database is. `domain.Backup` names
one for a surface, the way it always did.

A change to the shape did become one — `internal/migration` holds what a change
is, the list of them with the SQL that makes and unmakes each, and its two
renderings. The **ledger** stayed behind, in `internal/store`: what has already
run is a table in the database, and it belongs with the database.

## Half of this is enforced

`depguard`, inside `golangci-lint`, holds the **direction** of dependencies: a
notion cannot import another notion, nothing below a surface can import one,
`engine` cannot import a surface, and `cli` cannot reach the store or a
database. There is one rule per package, and the reasons are written beside
them in `.golangci.yml`. A violation fails the commit:

```
import 'database/sql' is not allowed from list 'domain' (depguard)
```

`domain`, `build` and `config` are held differently, by an **allow-list**:
each may import what its list names and nothing else. A deny-list only forbids
what somebody thought to name, and the import that would hurt most in something
that knows nobody is the one nobody thought of. Widening any of those lists is
therefore a decision, made once, in the open.

The notions are held by deny-lists instead, because what they must never reach
is a short named list — each other, the store, the engine, the surfaces, a
database — rather than everything unnamed. Each notion names every other
notion in its list, so "a notion never imports another notion" is held for all
four, not only for the pair somebody thought of first.

One thing no linter can hold: a surface computing a rule inline instead of
asking the engine — that one sits in `CLAUDE.md` rather than only here.

**Adding a package means adding its rule in the same commit.** A layer nothing
enforces is a layer that lasts until the first hurry.

## The two things a second surface forced out of `internal/cli`

They sat there while one surface existed, on the rule the storage interface got
in reverse: extract from two real cases, never from one real and one imagined.
The MCP server is the second case, so both moved — and what each move cost is
worth writing down, because the cost is the reason they waited.

**The build stamps are `internal/build`.** `version`, `commit` and `built` are
injected there now (see `PKG` in the Makefile), and both surfaces read
`build.Read`. What made this awkward is what the note here always said: a
scenario has to be able to choose what the binary claims to be, and no test is
linked. So the surface reads its build through a variable — `stamps` in
`cli.go` — and a scenario replaces that, not the stamps. `build.Stamped` takes
the three rather than reading them, so what a build would say can be worked out
without being that build.

**The acceptance world is `test`.** It is handed a runner — run this command
line and tell me what came back — and never learns who answers. That is not
tidiness: had it imported a surface, that surface's own scenarios would import
a package which imports them back, and Go refuses the cycle. Each surface keeps
its own steps, and only what more than one of them claims lives there.

Two consequences, both measured rather than expected:

- It imports `testing` outside a `_test.go` file, because Go compiles a
  package's test files only for that package. `no-go-testing` is told to leave
  it alone, and `make harness-check` replaces it with the narrower thing that
  worry was always about: `go list -deps ./cmd/cairn` must not reach
  `cairn/test`.
- It is measured like any other package. A package with no test of its own
  still appears in the coverage profile at 0%, so a harness nothing tested took
  the whole repository below the floor. Its steps are therefore named methods
  rather than closures inside the registration, which is what makes them
  reachable without godog — and a step that quietly stopped asserting would
  otherwise have shown up as somebody else's scenario passing for the wrong
  reason.

## The four exit codes, and what holds them true

`cmd/cairn` exits with what `cli.Main` returns, and a shell is told which of
four kinds of answer it got:

| | |
|---|---|
| `0` | it worked |
| `1` | the request was understood, and could not be satisfied — or something nobody foresaw went wrong |
| `2` | the request was not understood at all |
| `3` | the store could not be used |

The difference between 1 and 2 is the one worth getting right. A script that
only knew "it failed" would have to read the message to decide whether to fix
its arguments or to wait and retry. A refusal answered with 2 says "I did not
understand you" about a request that was understood perfectly well.

**What kind of answer a failure is, is the engine's to say.**
`engine.RefusalOf` reads an error and answers its kind — `refused`, `usage`
(not understood), `store`, or `failed` (nobody foresaw it) — and a reason, a
word like `no-such-task` that stays the same when the sentence is reworded. The
command line turns the kind into an exit code; the MCP server puts both under
`_meta`; `--json` prints them beside the message. The list used to live in
`internal/cli/exit.go`, and a script and an agent were answered from two
places.

Two things changed with the move. **What is not understood is marked, not
guessed**: the engine's readers and `Check` mark their refusals with
`engine.Misread`, and so does the command line for cobra's own — so a failure
nobody foresaw is no longer answered as a request somebody mistyped, and gets
1, not 2. And **a surface's own refusal is made with `engine.Refusing`**, a
reason and a sentence — a deletion not confirmed, a set of glyphs that is not
one — so it is answered like any refusal the engine listed.

The list is held true by a test rather than by care:
`TestEveryFailureCairnForesawHasAReason` reads the source of every package
below the surfaces, and every `Err…` it declares has to be in `reasons`. An
error added anywhere fails that test until somebody has said which kind of
answer it is. Nineteen were missing from the old list the day it was first
checked.

## A part of the work is not work in its own right

A sub-task is a line of a list its task owns: a title, and whether it is done.
It keeps a number and a row of its own in the store — so the trace of it is the
trace of anything else, and `task show 2` still answers — and it is reachable on
neither board except in the form of the task it hangs under.

Three rules follow, and they are one rule seen from three sides:

- it stands in `todo` or `done` and nowhere else. A part being "blocked" is its
  task being blocked; a part being "current" is its task being worked on;
- it is archived with the work it belongs to, and it goes when that work
  is deleted. A deletion that refused while a part was there could not be
  carried out at all once the boards stopped drawing parts;
- what it carries beyond a title and a tick is not asked for. Measured on a
  store of 219 tasks: its 49 parts were done 47 times and todo twice, and
  carried no deadline, no label, no priority, no objective and no blocker.

The command line and the MCP server keep `--parent`, so an agent can still cut a
task into pieces. What they do not have is a way to reach a part that is not a
way to reach any task, which is the point: the constraint is on where a part is
*edited*, not on what the engine will hold.
