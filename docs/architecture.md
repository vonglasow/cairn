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
| `cmd/cairn` | opens the store, builds the engine, hands it to the commands | everything |

`domain` imports nothing of ours. An entity that knew how it was displayed
would drag a terminal into every test that touches it, and an entity that knew
how it was written to disk would drag a database in.

`cmd/cairn` is the only place allowed to know every layer, because someone has
to wire them together. It stays a handful of lines.

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
