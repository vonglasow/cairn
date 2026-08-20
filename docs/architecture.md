# Architecture

Read this before adding a package or a command.

## Every change goes through the engine

A command parses input, calls one thing, and prints the result. It decides
nothing. All the rules live in `internal/engine`, so the next surface asking for
the same thing gets the same behaviour without a line being written twice.

Concretely: a command declares the shape it accepts, hands the arguments to the
engine, and prints what comes back. Any `if` in a command that is not about
input or output belongs in the engine.

## The layers

| package | holds | knows about |
|---|---|---|
| `internal/domain` | the entities and what makes them valid | nobody |
| `internal/store` | keeping them, and handing them back | `domain` |
| `internal/engine` | the operations | `domain`, `store` |
| `internal/cli` | parsing and printing | `domain`, `engine` |
| `cmd/cairn` | hands the process to `cli.Main` and exits with what it returns | `cli` |

`domain` imports nothing of ours. An entity that knew how it was displayed
would drag a terminal into every test that touches it, and an entity that knew
how it was written to disk would drag a database in.

`cmd/cairn` stays one line. Everything put there is out of reach of the tests
forever — the suite calls `cli.Main`, never the binary — so each line added
there is a line the coverage floor pays for and nothing can check.

**The store is opened where it is needed, not at startup.** A command that has
no use for it never opens it, which is why `cairn --version` creates no file
and why the default place stays untouched until something is actually written.
One invocation opens one connection and closes it on the way out, so nothing
survives between two runs of the command — which is what makes "a task added by
one command is found by the next" a claim about the disk rather than about a
cache.

## The store is spoken to through an interface the engine owns

`internal/store` is where SQLite lives, and it is the only place that knows
that. It speaks in entities: it takes a `domain.Task` and gives one back. No
`*sql.Rows` crosses its boundary, or every caller would need to know the schema.

A second backing is a real prospect, not an imagined one — a file, a remote,
something that is not SQLite — so the engine talks to a **storage interface it
declares itself**, and `store` satisfies it. That is `docs/design-principles.md`
applied as written: accept interfaces, and let the consumer define them. Go
needs no import for that, so `store` still knows nothing of `engine`.

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

**Nothing changes the shape unless it was asked to.** A store missing a change
this cairn knows is reported and left alone — `store.Open` refuses it, and only
the commands whose job is to change the shape use `store.OpenBehind`. Migrating
on the way past would mean `cairn task list` rewriting somebody's database as a
side effect of listing their tasks.

**Every change carries its own undoing.** A migration holds the SQL that makes
it and the SQL that unmakes it, and both are printable — `migration show` exists
so a person can read what a tool is about to do before letting it. The first
shape is the floor: it has no undoing, because unmaking it would mean throwing
the store away, which is what deleting the file is for.

## The backups are reachable without the store

`store.Vault` opens no database. That is not an oversight: a backup is most needed
exactly when the store cannot be opened — because it is corrupt, or because it
comes from a cairn that knows more. Needing a working store to reach the backups
would put them out of reach on the one day they matter.

The same reason keeps `domain.Backup` and `domain.Migration` in `domain`: a
surface has to name a backup and a change without being allowed to know how either
is made.

## Half of this is enforced

`depguard`, inside `golangci-lint`, holds the **direction** of dependencies:
`engine` cannot import a surface, `cli` cannot reach the store or a database.
The rules and their reasons are in `.golangci.yml`, and a violation fails the
commit:

```
import 'database/sql' is not allowed from list 'domain' (depguard)
```

`domain` is held differently, by an **allow-list**: it may import what that
list names and nothing else. A deny-list only forbids what somebody thought to
name, and the import that would hurt most in an entity is the one nobody
thought of. Widening that list is therefore a decision, made once, in the open.

What no linter can hold is a command computing a rule inline instead of calling
the engine. That one is on us, and it is why it sits in `CLAUDE.md` rather than
only here.

**Adding a package means adding its rule in the same commit.** A layer nothing
enforces is a layer that lasts until the first hurry.

## What is deliberately not split yet

Two things sit in `internal/cli` that a second surface will eventually want,
and they stay there until that surface exists:

- **The build stamps.** `version`, `commit` and `built` are injected by the
  linker into `internal/cli` (see `PKG` in the Makefile). An MCP surface will
  also need to say which build it is, and that is the moment to lift them into
  a package of their own — not before, since moving them now would mean the
  scenarios could no longer set them.
- **The acceptance suite.** `runAcceptance` is an unexported test helper, so a
  second package cannot call it. Making it shareable means a package importing
  `testing` outside a test file, which `no-go-testing` refuses for good reason.

Both are the same judgement the storage interface got, in reverse: extract from
two real cases, never from one real and one imagined.
