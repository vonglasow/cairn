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

## Two places, and their names

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

## Written for you

`cairn config init` writes the file where it is read, whole: every key,
commented, set to what cairn uses when nothing is said
(`@config-init-writes-every-key`). Writing one by hand meant finding its keys
here and in the README first (cairn#99). It is the README's example to the
byte — one constant, `config.Example`, that a test holds the README to, and
that another holds to every key the reader accepts — so neither can drift.
It never writes over a file already there, which may hold choices nobody
would want reset (`@config-init-leaves-a-file-alone`); and it writes what is
used when nothing is said, not what the environment says right now, because
a passing choice does not settle into the file.

## The plugin's folder

```toml
plugin = "~/.cairn/claude-plugin"   # where cairn writes its Claude Code plugin
```

Where `cairn plugin write` writes the plugin, and where `cairn mcp serve` keeps
it at its own version ([claude-code.md](claude-code.md)). Not said, it is
`claude-plugin` beside the store, wherever the store is. It is a key of the
file and nothing else — no flag, no variable — because two commands have to
agree on it: `plugin write --dir` named a folder the server never looked at,
and was taken out for it (cairn#98). Absolute, or under `~`, like the others
(`@config-names-the-plugin-folder`,
`@config-a-plugin-folder-that-is-not-absolute-is-refused`); `config show` says
where it is and whether the file named it.

## The pages, turned off one by one

```toml
[pages]
time = false        # tasks, objectives and time; each is on unless it says false
```

A page turned off is gone from every surface, and the engine is what says so
(`engine.Pages`): the boards draw no tab for it and `tab` walks past it, the
command line offers none of its commands — `task` and `step` with the tasks,
`objective`, `timesheet`, and `label` once neither tasks nor objectives are
left to label — and the MCP server lists none of its tools (`bdd/pages.feature`,
`@pages-off-takes-the-steps-with-the-tasks`,
`@pages-labels-go-with-the-last-page-they-label`, `@gui-a-page-off-has-no-tab`,
`@tui-a-page-off-has-no-tab`, `@mcp-a-page-off-offers-none-of-its-tools`).

**One rule, asked by everything that offers an act.** It is `engine.Pages`:
`Offers` for an operation, `Called` for a thing and a verb typed. The window's
menu bar asked nothing at first — one fixed list, `MenuBar`, read by the native
bars of macOS and Windows and by the button standing in for them — so a page
turned off kept its lines there, and the terminal's `:` line still ran its
commands (cairn#100). Now each line of the menu bar names the acts it does,
and is offered while one of them is (`@gui-a-page-off-takes-its-menus`,
`@gui-the-tasks-off-take-their-menus`); a menu left with no line goes, and the
`:` line runs only what is offered (`@tui-a-page-off-takes-its-commands`).

With every page off a board opens with nothing in it, its menu bar with
nothing either (`@gui-every-page-off-leaves-what-belongs-to-none`), and only
closing it does anything. What belongs to no page — the configuration,
backups, migrations, the trace, the plugin — stays on the command line and
for an agent. A
page cairn does not have is refused like any other unknown name, and `config
show` says which pages are off. A configuration that cannot be read turns no
page off: every command that needs it refuses anyway, and says why.



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

## Archiving on its own

```toml
[archive]
start = 2026-10-21   # the day the count starts
days = 14            # how long each period lasts; 14 when it is not said
```

A done column fills up: 55 tasks sat there next to 129 archived before anybody
swept it by hand (cairn#56). With `[archive]` said, days are counted from
`start` in periods of `days` calendar days, and a task that went done before
the period under way began is archived. The periods are arbitrary: nothing in
cairn gives them a meaning beyond the boundary they draw. The moment a task
went done is its last move into `done` in the trace. This runs when a board
opens (`engine.Hold`) and when the tasks are listed, through the command line
or the MCP server (`@archive-what-went-done-before-the-period`,
`@archive-keeps-what-went-done-in-the-period`), through the same act as
`task archive`, written down as done through `archive`
(`@archive-on-its-own-is-said-in-the-trace`). Nothing said, nothing is
archived on its own (`@archive-nothing-on-its-own-unless-told`), nor before
the day the count starts (`@archive-nothing-before-the-count-starts`). A start
that is not a day, or periods of no day, are refused like any key the file
gets wrong (`@config-archiving-in-periods-of-no-day-is-refused`,
`@config-archiving-counts-from-a-day`), and `config show` says it
(`@config-show-says-what-is-archived-on-its-own`).

Calendar days, not working days: counting working days would ask which days
are holidays, which cairn does not know. The day the period under way began is
not kept anywhere: it is computed from `start` each time, so changing `start`
or `days` takes effect the next time it runs.

Today is the day the clock says, unless `CAIRN_TODAY` names another, written
2026-11-04: a script, or a scenario, stands on that day to see what would be
archived then. A scenario cannot move the trace back — it asks the engine,
which only writes the present, and never opens the store — so it moves today
forward instead. A day it cannot read is taken as no day named.

## A focus session that asks to be renewed

```toml
[focus]
session = "1h"   # how long a focus session lasts before it asks
```

A focus left open counts time nobody spent, but no clock can tell that the work
stopped: cutting it at a time of day, or after a while unused, was built and
taken back the same day (cairn#51). The question goes to whoever holds the
focus instead. Each task in focus has a session, up since the focus was taken
or last renewed, read from the trace; the focus answer says it for every task
in focus (`sessions`, with `since` and `up`,
`@focus-session-says-how-long-it-has-been-up`). Once a session has been up for
`session`, it asks to be renewed or ended (`ask`,
`@focus-session-asks-to-be-renewed-past-its-length`), and every answer the MCP
server gives carries that question after its own text, whatever tool was
called (`@mcp-any-answer-asks-to-renew-a-focus-session`): an agent rarely calls
`task_focus`. Focusing the task again renews the session, written `focus
renewed` in the trace (`@focus-session-renewed`); within its length it writes
nothing. The time runs on through a renewal: only a release stops it. Nothing
said, nothing is asked. A length that is not a duration is refused, naming
`focus.session` (`@config-a-focus-session-that-cannot-be-read-is-refused`), and
`config show` says it (`@config-show-says-how-long-a-focus-session-lasts`).

The Claude hook takes the focus at each prompt, so a prompt past the length
renews the session: a person writing is a person at work. The question is left
for an agent working alone for longer than a session.

Now is the clock, unless `CAIRN_NOW` names another moment, written in RFC 3339:
a scenario stands later to see a session grow, and the acts written meanwhile
are written at that moment.

## What a deleted task does with its time

```toml
[time]
deleted = "taken"   # or "kept", what is done when it is not said
```

Kept, a deleted task's time stays on the sheet, said to be deleted; taken, it
leaves with the task. See [trace](trace.md#the-trace-outlives-the-task). Any
other value is refused naming `time.deleted`
(`@config-what-a-deleted-task-does-with-its-time-is-one-of-two`), and
`config show` says when it is taken
(`@config-show-says-a-deleted-task-takes-its-time`).

## The environment names the same two places

`CAIRN_STORE` and `CAIRN_NOTES`, one per key, and nothing else. They replace the
answer rather than taking a turn before it — see the order above — and what they
name goes through the same expansion and the same absoluteness rule as what the
file names. A place is read the same way wherever it was named.

## The store wants a disk of its own

The store keeps its latest writes in a log beside it (`-wal`, `-shm`) until it
is closed — the journal that lets a board read while a command writes. That
log is shared through the machine's own memory, which a network share does not
offer and a synchronised folder does not respect: put `store` on NFS, SMB, or
in a folder iCloud or Dropbox keeps in step, and two machines can each see
half a write. The store goes on a local disk; what travels between machines is
a backup, which is one file written whole.

## TOML, because a person opens this file

This file is for someone to read and change, so it has to be able to carry a
line above each place saying what that place is. That requirement decided the form on
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

## Not posed yet

Cairn does not write this file. Nothing has to exist for it to work — no file is
the ordinary case — but nothing leaves one behind either: a person who wants one
writes it.

Posing it at the first run is specified in `bdd/config.feature` — once, never
completed, never holding what the environment said, never by a server told to
only read — and every one of those scenarios carries `@later`. They are not run,
and this page does not describe them as what cairn does, until they are built.

## Reading is not the same as not being there

A file that is there and is not TOML stops the command, naming the file. No file
at all is the ordinary case and stops nothing. A file that is there and cannot be
read stops the command too, but today as a store that cannot be used (exit 3);
answering it as a refusal that names the file is specified, and carries `@later`.

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
- **Anything but places — and the pages.** Nothing about how cairn behaves
  belongs here. The pages were that decision, taken on its own on 2026-10-16
  (cairn#67): which of them somebody uses is a fact about the person, read the
  same way by every surface, so it lives where every surface already looks.
  Any other preference is still a decision of its own.
- **A file per project.** One file, at the place configuration lives. A store
  somewhere else is what the environment is for.
