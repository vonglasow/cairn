# Where cairn finds its own places

Read this before touching how the store or the notes are addressed, before
adding a key to the configuration, or before changing what happens when nothing
says where something is.

Cairn has two places to find before it can do anything: where the tasks are
kept, and where the notes taken elsewhere are looked for. Both are found the
same way, so there is one rule to know rather than two. The scenarios are in
`bdd/config.feature`; what the notes are for is in [context.md](context.md).

## The order, and why the environment is not part of it

```
the file            ~/.config/cairn/config.toml
   ↓ says nothing
the default         ~/.cairn/cairn.db   ~/.cairn/notes
   ↑
the environment     replaces the answer, whatever it was
```

The file is where a choice is made and kept. The default is what a choice
looks like when nobody made one — the same on every machine, so "where are my
tasks" never depends on which one you are sitting at.

The environment is neither. It is a passing choice: another store for one
command, another vault for one session. It does not take a turn after the file
and before the default, it answers instead of all of it — and it changes nothing
on disk, so the next command without it finds things exactly as they were.

Saying it as a chain of three was tried and read wrong: it makes the environment
look like a fallback, which is the one thing it is not.

**This changes something that already worked.** Before, the store was the
environment or the default, with nothing in between. The file is a new rung, and
a machine with a file now answers differently than it did.

## Two keys, and their names

```toml
# where cairn keeps your tasks
store = "~/.cairn/cairn.db"

# where cairn looks for notes taken elsewhere
notes = "~/.cairn/notes"
```

`store` and `notes`. Both name the place itself and not a folder to compose
something into: `store` is the file, `notes` is the directory task directories
sit under.

They fall under one folder by default, `~/.cairn`, so what cairn keeps stays
together and one directory is enough to move or back up by hand.

## Absolute, and nothing else

A place that depends on where the command was run from answers differently
depending on who asks, and a configuration is read from everywhere. So `notes`,
`./notes` and `../vault` are refused, naming what was wrong.

`~/vault` is absolute and is accepted — it is also what anyone writes by hand,
and it is the one form that means the same thing on macOS and on Linux, which
`/Users/me/vault` does not. A leading `~` is expanded first and the result is
what is judged: `filepath.IsAbs("~/vault")` is false, so the two steps cannot be
folded into one.

**A key written as the empty string is refused, and a variable set to nothing is
not.** The asymmetry is deliberate and reads as a bug if it is not said out
loud. `store = ""` is a choice somebody wrote down and that does nothing — the
same silence this file condemns two sections below, and it falls out of the
absoluteness rule for free, since an empty string is not an absolute path.
`CAIRN_STORE=` is a passing choice, and a passing choice that says nothing is
simply nothing: it leaves the answer as it was worked out. The file is written;
the environment is not.

## The environment names the same two places

`CAIRN_STORE` and `CAIRN_NOTES`, one per key, and nothing else. They replace the
answer rather than taking a turn before it — see the order above — and what they
name goes through the same expansion and the same absoluteness rule as what the
file names. A place is read the same way wherever it was named.

## TOML, because a person opens this file

Cairn poses this file for someone to read and change, so it has to carry a line
above each place saying what that place is. That requirement decided the form on
its own: JSON has no comments and turns paths into escapes, and a hand-rolled
`key = value` reader means inventing rules for quoting, whitespace and repeats
that a specification already answers.

The dependency is real and this is its written reason.

## A name cairn does not know is refused

Not passed over. A setting somebody believes is in force and that does nothing
is worse than no setting at all — this is the same silence that was already
found and fixed once, on arguments an agent sent that nobody read.

It is paid for: an older cairn refuses a file written by a newer one, and unlike
the store, a configuration has no ledger to carry it across versions. Taken
knowingly.

## Posed once, and never put back

The first command leaves a file behind, holding the places cairn would have used
anyway. Opening it is then enough to see where everything is, without knowing
anything beforehand.

| situation | what happens |
|---|---|
| no file | posed, with the defaults and their comments |
| a file that says one place and not the other | left half-said — never completed |
| the environment says something at the first run | the file still holds the default; the command still obeys the environment |
| the place cannot be written to | said, and the command carries on |
| a server told to only read | nothing posed |

The file is posed once. What is in it afterwards is the person's.

Three of those rows are each a way of not being clever. Completing a half-said
file would rewrite something somebody edited. Writing down what the environment
happened to say would freeze a passing choice into a permanent one. And a
command that failed because a convenience file could not be written would fail
for a reason that has nothing to do with what was asked — so it says so and goes
on, which is the one place this departs from an unreadable `--state`, and it
departs because posing is not reading.

A server told to only read poses nothing, and that has a consequence worth
knowing rather than discovering: on a new machine the first thing to run cairn is
often the server, so the file appears later, the first time something able to
write does.

## Reading is not the same as not being there

A file that is there and cannot be read stops the command, naming the file. A
file that is there and is not TOML stops it too, and says so differently. No file
at all is the ordinary case and stops nothing.

Three answers rather than one, because a permission and a typo are not an
absence, and a caller that cannot tell them apart looks for the wrong problem.

## `config show` reports where every other command refuses

`cairn config show` says which places cairn will use and what named each one:
the file, the environment, or nothing. It is the only command that does not
refuse a configuration it cannot make sense of — and that exception is the
point, not an oversight. A person reaches for it *because* something is not
where they expected it, so a diagnostic that answers with the same refusal the
last command gave is a diagnostic that is useless exactly when it is called.

It does not lie in the other direction either. On a file it cannot make sense
of it prints nothing about places — showing `store → somewhere` would say the
opposite of what is true, since that same file turns every other command away —
and it says so: the file, the problem, and the warning that the rest will refuse
until it is fixed. A scenario measures that last sentence by running another
command on the same file, rather than taking cairn's word for it.

Two consequences worth knowing rather than discovering. It carries
`Needs: AConfiguration`, which is what keeps the shared path from resolving a
store before it runs — without that it would be unreachable precisely when it
matters. And over MCP it answers normally rather than as a refusal, because a
refusal carries nothing for a machine, and the file's path is the one thing an
agent needs when everything else has started saying no.

## Deliberately not there yet

- **A command to write the file.** It is posed, and after that it is edited the
  way any file is. A `config set` would be a second way of doing one thing.
- **Saying what the file named when the environment replaced it.** `config show`
  says what won and who named it, never what lost. It was sketched and left out:
  no scenario observes it, and a capability nothing observes is either specified
  or dropped.
- **Anything but places.** Nothing about how cairn behaves belongs here. The day
  a preference wants a home, that is a decision to take on its own.
- **A file per project.** One file, at the place configuration lives. A store
  somewhere else is what the environment is for.
