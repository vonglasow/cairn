# Architecture

Read this before adding a package or a command.

## A notion is a cell

A task is a cell. Its state is its own — no exported field, no accessor — and
it never hands data back. It **tells**: one method per fact it carries, absence
included, to whoever asked to be told. Whoever wants to know something about a
task implements that membrane and is told it.

```go
type Told interface {
	Numbered(int)
	Titled(string)
	Described(string)
	NotDescribed()
	Due(time.Time)
	NotDue()
	…
}

func (t Task) Tell(to Told)
```

Four membranes exist for a task today and the cell knows none of them: one
builds the card a person reads, one the shape a machine reads, one binds the
SQL that keeps it, one counts what is done. Adding a fifth costs the cell
nothing, which is the whole point.

The act in `internal/trace` is a cell of the same kind, and it shows what a
membrane buys where the shapes differ: one renders a line of a task's history,
one a line of a period across tasks, one binds the SQL, one adds up how long
the work ran. Four readings of the same facts, and the act knows about none of
them.

It also shows what a cell may not reach. An act names its task **by number**
and never carries the task, which is why `internal/trace` is forbidden
`internal/task` outright: a task may be gone by the time its acts are read, and
an act holding one could not outlive it. That is a rule about what a notion is
allowed to know, and it is in `.golangci.yml` where it can be enforced rather
than remembered.

An **absence is a fact that gets told**, never a silence a membrane has to
notice. `NotDue` means there is no deadline, and a membrane that ignored it
would render whatever it happened to start with. A set is closed the same way:
labels are told one at a time and then `LabelsTold` says there are no more,
because otherwise a membrane cannot place its row until it is too late. Both of
those were found by writing them wrong first.

The membrane runs the other way too. `AsTold` is told a task's facts and hands
back the task they make — that is how the store builds one out of a row, and
how every test builds a fixture. There is one way to make a task, and it is the
same one whichever direction the facts are travelling.

**Errors do not go east.** Telling is for reading; a command that fails returns
its error the way Go returns errors. That line is deliberate: pushed further,
every call would grow a `Failed` method and Go would fight back on every one.

## The layers

The outer axis is the **notion** — the thing a person can name and ask for.
What is left over is either carried by every notion, or is the plumbing under
all of them.

| package | holds | knows about |
|---|---|---|
| `internal/domain` | what is **carried**: columns, priorities, labels, days, durations, progress, the title rule, what an absence looks like | nobody |
| `internal/build` | which build this is | nobody |
| `internal/task` | the task: its cell, its rules, and its two renderings | `domain` |
| `internal/objective` | the objective, in the same five shapes | `domain` |
| `internal/migration` | a change to the shape of the store, and the list of them | `domain` |
| `internal/trace` | one act performed on a task, and what a period of them comes to | `domain` |
| `internal/store` | keeping the notions, and handing them back | `domain`, the notions |
| `internal/engine` | the operations, the calculations, and the only door to the store | `domain`, the notions, `store` |
| `internal/reading` | what every board reads and no board draws: the two listings turned into cards and goals, which fields a form asks about, what a saved form asks the engine for, how the cards deal into columns | `domain`, the notions, `engine` |
| `internal/cli` | parsing and choosing a rendering | everything below |
| `internal/mcp` | the same, for an agent | everything below but `cli` |
| `internal/tui` | the board on a terminal | everything below but `cli` and `mcp` |
| `cmd/cairn` | hands the process to `cli.Main` and exits with what it returns | `cli` |
| `test` | the acceptance world the surfaces share | `engine` |

**`reading` is not a surface, and that is why it exists.** Two boards draw the
same work — one in a terminal, one in a window — and they differ in how a box
is drawn, not in what goes in it. That a task with nothing written about it
carries no description mark, that a deadline four days out reads `4d`, that the
three stopped states go under the last three columns, that a form's `labels`
box fills the engine's `label` argument: none of it is about cells or pixels,
and all of it would otherwise be written twice.

So `reading` is held to something stricter than a surface. `depguard` denies it
lipgloss, bubbletea, huh and cobra outright — not because they would break it,
but because reaching for one is the moment it stopped being shareable. It was
lifted out of `internal/tui` the day the second board was specified and not
before: a package made for one consumer is a boundary guessed rather than
measured.

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

## More than one cairn at a time

One person, one machine, and still two cairns at once: a second terminal, a
shell loop, a server answering while the terminal is in use. They share a file,
so the store is opened in WAL with a wait of five seconds.

Both are measured, not precautionary. Without the wait, four processes writing
together lost fourteen writes out of twenty. With the wait but without WAL, six
mixed reads and writes spent the whole five seconds and still lost five of
them: in the default journal a reader blocks a writer, so writers starve. In
WAL the same six take fifty milliseconds.

WAL keeps recent writes beside the store, which matters for `backup add`
because it copies the file. Nothing is lost today: a command opens the store in
its RunE and closes it on the way out, and closing checkpoints — so between two
commands there is only ever one file. A surface that holds the store open, as a
server would, has to checkpoint before a backup is taken.

**Opening reads the shape, it never lays it down.** `migration apply` is what
changes a store; opening one only looks. The ledger is the single exception,
created when it is missing because nothing can be recorded without it.

## The store is spoken to through an interface the engine owns

`internal/store` is where SQLite lives, and it is the only place that knows
that. It speaks in notions: it takes a `task.Task` and gives one back. No
`*sql.Rows` crosses its boundary, or every caller would need to know the
schema.

It never reaches inside one either. A task tells the store its facts on the way
in, and the store tells a builder a row's facts on the way out — the same
membrane, run both directions. That is what lets a cell keep its insides to
itself while still being written to disk.

A second backing is a real prospect, not an imagined one — a file, a remote,
something that is not SQLite — so the engine talks to a **storage interface it
declares itself**, and `store` satisfies it. That is `docs/design-
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

`domain` and `build` are held differently, by an **allow-list**: each may
import what its list names and nothing else. A deny-list only forbids what
somebody thought to name, and the import that would hurt most in something that
knows nobody is the one nobody thought of. Widening either list is therefore a
decision, made once, in the open.

The notions are held by deny-lists instead, because what they must never reach
is a short named list — each other, the store, the engine, the surfaces, a
database — rather than everything unnamed.

Two things no linter can hold. A surface computing a rule inline instead of
asking the engine — that one sits in `CLAUDE.md` rather than only here. And a
cell growing an accessor: nothing stops `func (t Task) Title() string` from
being written, and the day it is, every membrane becomes optional and the shape
unravels one caller at a time. What guards it is that no such method exists to
copy from.

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
