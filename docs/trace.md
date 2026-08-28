# The trace

Read this before touching what cairn records about what happened, or before
adding anything that changes a task.

The scenarios are in [bdd/trace.feature](../bdd/trace.feature); this file
carries the reasons behind them, because a scenario says what happens and not
why it was chosen that way.

It was written before any of this existed. Three things changed while it was
being built, and each is marked below with what forced the change.

## The question this answers

cairn holds no dates. A `critical` set in March reads exactly like one set this
morning, a task that has sat untouched for four months looks like one edited
yesterday, and nothing anywhere says what was done, when, or in what order. The
store knows the present and has never known anything else.

What was asked for is a trace: every act performed on a task, kept, so it can
be found again afterwards. Three uses came with the asking, and they are what
the design serves — knowing what was done on a task, how long it took, and
where the day was cut up.

## It is a history, not a log

A log of events says *the priority changed, on the 3rd of March*. You know
when, never what. A history says *the priority went from `normal` to
`critical`*, and that is a different claim: it lets a task be put back together
as it stood at any past moment, and it shows that a `critical` was raised
**from** something, which the date alone does not.

So every entry carries the field, the value before, the value after, and the
moment. The two extra columns are the whole difference between a trace you can
read and a trace you can rebuild from, and rebuilding is what was asked for.

## Seventeen acts, and reading is not one of them

Twelve of them are the attributes a task carries, one act each: its title, its
description, its column, its priority, its labels, what blocks it, the words
beside them, its deadline, its estimate, the time spent on it, its parent, the
objectives it serves. Five are not about a field: **created**, **moved**,
**focus taken**, **focus released**, **deleted**.

The number grows when a task gains an attribute, and nothing has to be written
here for it to: `theAttributes` in `internal/engine/recording.go` is one list,
walked both by the recording of a creation and by the recording of a change, so
an attribute added there is traced by both.

`list`, `show` and `search` are absent on purpose. A trace records what
changed, not what was looked at — traced, consulting a listing would fill the
journal with reading noise, and the switches that matter would be lost in it.

## A creation writes one act per attribute it started with

Not the bare word `created`. If creation said nothing and only the changes
after it were kept, then replaying a task would need the **defaults of the day
it was made** — `backlog`, `normal`. The day a default changes, every old
history replays wrong, and does so in silence. So what a task started as is
written down, and a replay then depends on nothing from that era.

**This began as one act carrying all of them, and it could not work.** The
values were laid out as `title=…, column=…` pairs, and measured: a title
carrying a comma splits into a pair and a word, so `Écrire la spéc, enfin`
cannot be read back apart. Nor is there a separator that would save it — a
description accepts tabs and newlines both. Any one field holding twelve values
needs a format invented for it and a parser to match, and the two disagree
eventually.

So a creation writes the act itself, then one act per attribute the task
carries, each value alone in its own field. Exact, with nothing invented. What
it costs is four lines at a birth instead of one, and that is paid back in the
rendering: those acts are laid out **indented under the creation** and share
its moment, the way a listing indents a sub-task. A history then reads as one
line for the birth and one line per thing that has happened since.

## The trace outlives the task

Deleting a task does not take its history. Otherwise **the past changes every
time something is deleted**: the hours spent on Tuesday on a task thrown away
on Friday would vanish from Tuesday, and "how was my week" would stop being a
question with a stable answer. That is precisely what a trace exists to
prevent.

It works because numbers are never handed out twice — measured: delete `#2` and
the next task added takes `#4`. So an entry naming task 7 goes on naming that
one task after it is gone, and never quietly means another.

The price is real: the trace cannot be a child that leaves with its parent, and
reading the history of a deleted task needs a way in that is not `task show`.

## What it replaces

| what stops needing its own place | where it comes from instead |
|---|---|
| when a task last changed column | the last entry that changed the column |
| when it was created, when anything last changed | the first entry, and the latest |
| when a priority was set | the entry that set it |
| where the focus has been | taking and releasing it are acts |
| which tasks arrived on top of the plan | arriving is an act like the rest |

The first of those was settled as an element while [ordering.md](ordering.md)
was written and deliberately not delivered, on the grounds that a column nobody
reads holds whatever the last few months left in it. It is not delivered now
either: it is derived.

## The act is traced, not its consequences

Moving one task shifts the place of every task behind it. Traced, that is
twenty entries for one act, nineteen of which nobody will ever read — measured
over a year at twenty moves a month: 240 entries against 4,800, of which 4,560
are noise, and the one line that says what a person did buried among them.

So an entry records the act: *moved to the first place*. What that did to the
places of its sisters is not written down.

What is given up with it, said plainly: the order of the whole list as it stood
on a past day cannot be read back. It could be recomputed by replaying the
acts, which is not part of this and may never be — and a replay would depend on
the ordering rules never having changed. A single task, on the other hand,
comes back exactly, since every change to what it carries is an act of its own.

## What it derives, in the raw

A listing of acts over a period ends with what can be worked out from them, the
same way `cairn migration list` ends with what the store holds. Numbers, not a
reading of them: how many stretches of work, on how many tasks, how many
switches, and the time per task.

A **stretch** is a `focus taken` closed by the `focus released` that follows.
There is no other definition available, and it is what every duration rests on.

**A release says what took over, and that was not in the first design.** Moving
the focus straight from one task to another writes two acts at the same moment
— a release, then a take — so nothing in the sequence alone tells that apart
from stopping and starting again. Measured: counting a switch as "a take while
a stretch is open" gave zero for a day of nothing but switching. So a release
names the task that took over, and a switch is a release that names one; a
release naming nothing is a stop, and picking up later is a resumption.

That is the distinction asked for — a cut is being pulled onto something else,
not finishing and going to a meeting — and it now rests on what was written
down rather than on reading between the lines. It is also what a person wants
to see: the line says what interrupted them.

Its limit, before anything is built on it: a stretch left open overnight runs
all night, and fourteen hours land on a task nobody touched. That is tolerable
here, and only because the two notions of time were kept apart from the start —
the total, for a timesheet, stays typed by hand, where `--spent` takes `+30m`
and `-15m`; the trace says the **shape** of a day, not what it is worth.

## It says who acted, when that is not you

An agent reaches cairn over MCP and can change a task. An entry carries which
surface did it, and shows it only when it was not the terminal: silent in the
ordinary case, visible in the one that matters.

## What the store gains

One table, appended to and never rewritten: which task, which act, which field,
what it was, what it became, when, and by which surface.

Two commands: `cairn trace show <number>` for one task's history, and `cairn
trace list` for the acts over a period with the derived numbers under them.
`--since` widens it; left out, it means today. `trace` is a group of its own
beside `backup` and `migration` — things about the tasks that are not tasks.

**`trace show` refuses a task nothing was ever written down about**, and the
reason is worth stating because the obvious check is the wrong one: it cannot
ask whether the task exists, since a deleted task keeps its history. The only
signal available is that there are no acts at all — and it is sound, because a
task always carries at least its creation. Before this, asking about a task
that never existed answered "nothing has been done to it" and exited 0, which
reads as an answer rather than as a mistake.

Nothing else changes. Every act already goes through `internal/engine`, which
is what makes one place enough to record them all.

## Weighed and left out

| | why |
|---|---|
| a log of events rather than a history | it says when, never what, and a task could not be put back together — which is the whole point |
| tracing what an act did to other tasks' places | 4,560 lines a year nobody reads, burying the 240 that say what somebody did |
| tracing what was read | consulting a listing would drown the switches that matter |
| **archiving old entries now** | measured: twenty years at 500 acts a month is 9.2 MB, and reading one task's history or one day's switches out of 120,000 entries is under the threshold this can measure. Nothing about size or speed asks for it |

The one argument for archiving that does hold is not about the store: `backup
add` copies the whole store, so every backup carries the whole trace. Fifty
backups is fifty times that. It is also answered by keeping fewer backups.

## Deliberately not here yet

- **Archiving.** Acted as an element, not built. The constraint to keep,
  because it is the real cost and not the storage: **a history cannot be cut
  without photographing where it was cut.** Take the earliest entries out and
  what remains begins in mid-air, so replaying a task — the thing this exists
  for — stops working. Cutting a period means writing the state of every task
  at the cut. And choosing a cut policy against a store with no history in it
  would be deciding blind; it is designed the day there are two years of trace
  to look at and a concrete reason to cut.
- **The longest stretch and the shortest.** They were in the first sketch of
  what a period says. No scenario asked for them, so they were not written.
- **A reading of a day or a week.** The numbers are given raw on purpose. What
  counts as a real interruption, what a good day looks like — those are
  judgements, and making them before a single day has been traced would be
  guessing. The raw numbers are what make the judgement possible later.
- **The order of the whole list at a past moment.** See above: the act is
  traced, not its consequences.
- **Putting a task back to how it was.** The history makes it possible to work
  out; nothing here does it. Reading and undoing are two different asks.

## One thing a restore takes with it

`backup restore` replaces the store with an older one, and the trace goes back
with it: what happened after that backup was made is gone from the history as
well as from the tasks. That is inherent — the trace lives in the store it
describes — and it is written here so it is not discovered.
