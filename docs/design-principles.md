# Design principles

Read this before writing code.

Written for Go: no inheritance, small interfaces, packages as the unit of
dependency. Principles that came from elsewhere are stated as they apply here,
not as they are usually recited.

## DRY — don't repeat yourself

Before writing anything — a function, a make target, a hook, a helper, a
constant — look for it first. If it already exists, use it. If something close
exists, extend it rather than adding a sibling beside it.

Two definitions of the same thing do not stay identical: one gets fixed, the
other does not, and the difference is found much later by someone who trusted
the wrong one.

This applies across files, not just within them. A command defined in the
Makefile and repeated in `.pre-commit-config.yaml` is a duplicate. A profile
built by two targets is a duplicate. When a check needs a command, it calls
the one that already exists.

## YAGNI — you aren't gonna need it

Build what the current scenario needs, nothing more. No option that nothing
asks for, no abstraction for a second case that does not exist yet, no
configuration knob nobody has requested.

Speculative code is never free: it has to be read, tested, mutated and carried
by everyone who comes after, and it is usually wrong anyway — the second case,
when it finally arrives, rarely looks like the one that was imagined.

When a second case does arrive, generalise then. That is the cheap moment,
because both cases are finally on the table.

## KISS — keep it simple

The simplest thing that satisfies the scenario wins. A clever solution that
needs a paragraph of explanation loses to a dull one that needs none.

Complexity is only worth its price when something concrete demands it — a
measured cost, a real case. Not a feeling that it might.

## SOLID

### SRP — single responsibility

A type or function has one reason to change. A command parses and prints; it
does not also decide business rules. When a name needs "and" to be accurate,
it is two things.

### OCP — open for extension, closed for modification

Adding a case should mean adding code, not editing a function everyone else
depends on. In practice here: a new command registers itself on the tree, it
does not require surgery inside an existing one.

Do not build for this in advance — see YAGNI. Reach for it when the second case
actually arrives.

### LSP — substitutability

Go has no inheritance, so this is about interfaces: any implementation must
honour what the interface promises. No panicking where the contract says an
error is returned, no silently doing nothing where it says the work happened.
A caller must never need to know which implementation it holds.

### ISP — interface segregation

Small interfaces, and **defined by the consumer, not the producer**. A function
needing to read one thing takes an interface with that one method, not the
whole service. Interfaces here are usually one or two methods.

### DIP — depend on abstractions

Accept interfaces, return structs. A package owning rules must not import the
package that talks to SQLite; the storage side satisfies an interface the rules
side declares. That is what keeps the rules testable without a database.

## Tell, don't ask

Send an object a message; do not interrogate it to decide on its behalf.

Pulling fields out to compute something the object could have computed puts its
rules in the caller — where they get duplicated, and drift the day the object
changes. A rule stays on the notion. Its facts, though, are handed over as one
plain record (`task.Carried`) rather than told one method at a time: rendering
a card or a row is not a rule, and eighteen copies of the same fields were the
duplication this principle is meant to prevent.

## SSOT — single source of truth

Every piece of information has one home. A value that is derived is derived,
never stored a second time; a rule lives in one place and everything else
calls it.

Two backups of a truth become two truths, and nothing says which one is right.

## Law of Demeter

Talk to your neighbours, not to strangers. `a.B().C().D()` reaches through
objects that are none of your business, and ties you to a shape you do not own —
it breaks the day any link in that chain changes.

Ask the neighbour for what you need, and let it deal with its own neighbours.

## Comments

The code carries no comments but a one-line doc comment on each exported
symbol, which the linter requires. Why a thing is the way it is lives in
`docs/`, in the scenarios' names and in the commit that made it so — not
beside the code, where it went stale and outweighed the code it described.
Scenarios carry their steps and no prose.
