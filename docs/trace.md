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

## Nineteen acts, and reading is not one of them

Thirteen of them are the attributes a task carries, one act each: its title, its
description, its column, its priority, its labels, what blocks it, the words
beside them, the tasks it is linked to, its deadline, its estimate, its steps, the
objectives it serves, and whether it was archived — archiving and
bringing back are changes like any other, written `yes` and nothing. Six are not
about a field: **created**, **moved**, **focus taken**, **focus released**,
**focus renewed**, **deleted**, **time added**. A focus renewed restarts the
focus's session and leaves its time running (see
[config](config.md#a-focus-session-that-asks-to-be-renewed)).

**Steps are one attribute.** A step added, ticked, renamed or taken away is a
*changed* act on its task with the field `steps`, the values before and after
written as the whole list, `[x] a; [ ] b` (`@step-traced`): a step has no
number, so it has no trace of its own, and the list is what the task said
before and says after. See [steps.md](steps.md).

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
description accepts tabs and newlines both. Any one field holding thirteen values
needs a format invented for it and a parser to match, and the two disagree
eventually.

So a creation writes the act itself, then one act per attribute the task
carries, each value alone in its own field. Exact, with nothing invented. What
it costs is four lines at a birth instead of one, and that is paid back in the
rendering: those acts are laid out **indented under the creation** and share
its moment. A history then reads as one line for the birth and one line per
thing that has happened since.

## A deletion writes down what went with it

The act that says a task was deleted carries what the task held, as one line of
`field: value` pairs. It used to carry nothing at all: cairn recorded that a
task had gone and not what had gone with it — no title, no description — so
even putting it back by hand was impossible.

That is a different thing from not having undo. Reading and undoing are two
different asks and cairn answers only the first, deliberately; a trace that
kept no content made the second unanswerable by anybody, which was not a
decision, only an omission.

It is written on the one act rather than one act per attribute. A deletion is
one thing that happened, and a history reading as thirteen would say a task was
deleted thirteen times.

## The trace outlives the task

Deleting a task does not take its history. Otherwise **the past changes every
time something is deleted**: the hours spent on Tuesday on a task thrown away
on Friday would vanish from Tuesday, and "how was my week" would stop being a
question with a stable answer. That is precisely what a trace exists to
prevent.

It works because numbers are never handed out twice — measured: delete `#2` and
the next task added takes `#4`. So an entry naming task 7 goes on naming that
one task after it is gone, and never quietly means another.

The price is real: the trace cannot be a child that leaves with its task, and
reading the history of a deleted task needs a way in that is not `task show`.

**A deleted task keeps its time on the sheet, said to be deleted**
(`@timesheet-a-deleted-task-keeps-its-time-and-says-it-is-gone`): its row
carries the title its deletion took with it, `deleted` in the JSON and
"(deleted)" after the title, and the window says "task 7 was deleted" rather
than choose a card that is gone (`@gui-a-deleted-tasks-time-says-so`). For
whoever would rather five minutes on a task made by mistake left with it,
`[time] deleted = "taken"` makes every deletion take its time
(`@timesheet-a-task-deleted-with-its-time-leaves-the-sheet`): the deletion
writes `time: taken` last among what went with the task, and the sheet leaves
every act of that task out. Its history is still there to read; only the
sheet stops counting it, and only because the configuration asked. A task
deleted before the key was set keeps its time. Taking the time back with one
`time added` of less than nothing per day was weighed and left: as many acts
as days, and a row at nothing left behind (cairn#122).

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
places of the tasks it sits with is not written down.

**A link is written on both tasks it joins.** It is the one act cairn writes on
a task nobody named: `task edit 2 --link +1` writes `links — → #1` on task 2
and `links — → #2` on task 1 (`@trace-a-link-is-written-on-both-tasks`). It
is not a consequence like a shifted place: a link has no direction, so it is
an attribute of both tasks, and task 1's own `links` changed. The same holds
when a task is deleted: each task it was linked to is written `#2 → —`
(`@task-link-gone-with-the-task`), where the blockers a deleted task held are
let go of without a word. See [links.md](links.md).

A move to the place a task already holds is no act at all, and nothing is
written: a card dropped back where it stood, or the only card of a column moved
to the head of it, changed nothing — like an edit giving a field the value it
holds (`@trace-a-move-to-where-it-stands-is-not-an-act`).

What is given up with it, said plainly: the order of the whole list as it stood
on a past day cannot be read back. It could be recomputed by replaying the
acts, which is not part of this and may never be — and a replay would depend on
the ordering rules never having changed. A single task, on the other hand,
comes back exactly, since every change to what it carries is an act of its own.

## How long the work took: the timesheet

A listing of acts over a period used to end with what could be worked out from
it: how many stretches of work, how many switches, the time per task. It went,
on a measurement — a week of a real board said 117h45 on each of two tasks,
because both had been left in focus for days and a stretch ran to whenever it
was closed. A sum that counts the nights says nothing anybody can put on a
sheet. What replaced it answers the one question that sum was for: how long the
work took, day by day. `cairn timesheet show`, and `timesheet_show` for an
agent. The scenarios are in `bdd/timesheet.feature`.

**Focus is what is measured**, and nothing else is: focused when the work starts,
unfocused when it stops, at the moment it is done and no other. Placing it after
the fact was built and taken back out the same day: a focus that can be put at
any moment past is a timesheet anybody can write, and the time it measures
says what was typed rather than what was done. What was done without focus — a
meeting, an afternoon nobody focused — is added by hand (`timesheet add`), and a
focus left on too long is taken back the same way, with less than nothing. Both
are acts in the history like any other, read back from there like the rest.

**The time spent on a task is read from here, and from nowhere else.** It is
the time the task was in focus and the time added by hand, summed over every
day; `task edit --spent` is time added by hand today, the timesheet's `--on`
another day. It used to be a number written on the task beside this history, and
the two disagreed: the card said what was typed, the timesheet what was
focused. The number written before is in the history now, as time added with no
day — no day is invented for it, so it counts on the task and on no day of the
sheet. A form sends only what was changed in it, since a focus still running
makes the time it opened with already behind by the time it is saved.

**A run of focus stops at the first midnight after it began.** One left open
would otherwise count the night, and every day after it; the day it crossed
says so, task by task. The cost, known and taken: work truly carried on past
midnight loses what came after it, and is added by hand.

**Two tasks in focus at once share the minutes they overlap**, so a day comes
to what the clock says and never to more. Several are often in focus at once,
one per agent session; counted in full on each, a day of three sessions read as
three days.

**Each task is said to the minute, a minute begun counted whole**, and a day is
the sum of those — the sum of what is written, so the lines add up to their
total, and a task's time spent is the sum of its days. It was rounded to the
nearest quarter of an hour, and that hid every task worked on for less than
seven and a half minutes: a review answered, a ticket read, gone from the day.
Bringing a day to quarters is the sheet's business, done when it is filled in. A
task with no time is not written; a day nothing was done on is, coming to `0m`,
so a week reads as a week.

**There is no ceiling and no kind of work.** A sheet may accept eight hours a
day and sort them into development, fixes, meetings; cairn says the real time,
every hour of it, and holds no category. Bringing a day down and sorting it out
is done when the sheet is filled in, by whoever fills it in — the categories
change, and a task does not know which one it was until it is over.

## It says who acted, and through what

Every act carries who did it and the surface it came through, and says both
on its line: `(by sam, through gui)` (`@trace-says-who-acted`). It used to
carry the surface alone, and only when it was the MCP server: the terminal,
the board in the terminal and the window all wrote nothing, so they could not
be told apart, and nothing at all said who — a person, a script, an agent, a
Claude session (cairn#83).

The surface is `cli`, `tui`, `gui`, `mcp`, `migration` for what a change
of the store's shape wrote, or `archive` for a task archived on its own
(see [config.md](config.md#archiving-on-its-own)). Who is the one `CAIRN_AUTHOR` names when it is
set, and the one logged in otherwise (`@trace-author-is-the-login-by-default`);
over MCP, the client that greeted the server names itself — `claude-code` —
unless `CAIRN_AUTHOR` is set for the server
(`@mcp-what-an-agent-does-is-written-down-as-its-own`); the plugin's hook sets it to
`claude session <id>` for what it does to the focus, so a session's acts are
told from the person's (`@claude-session-acts-are-the-sessions`). The engine
takes both as one value, `Acting`, from the surface that asked — `DoFrom` for
a command, `Hold` for a board — and writes them on every act it records.

What one author did is read apart: `cairn trace list --by sam`
(`@trace-list-by-author`), and the JSON carries `by` and `through` on every act
(`@trace-json-says-who`). Acts written before keep what they had: no author,
and a surface only when it was the server. The column is `author`, not `by`,
which SQL keeps for itself.

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
- **A reading of a day or a week.** The timesheet gives the hours raw on
  purpose. What a good day looks like, what counts as an interruption — those
  are judgements, and cairn does not make them. Hours gathered by objective or
  by label are not a judgement, and they are not here either: nothing has asked
  for them, and a sheet is filled in from the raw hours by whoever fills it.
- **The order of the whole list at a past moment.** See above: the act is
  traced, not its consequences.
- **Putting a task back to how it was.** The history makes it possible to work
  out; nothing here does it. Reading and undoing are two different asks.

## How long a task has sat where it is

The one date nothing a task carries implies is read from here, not kept beside
the task: `task show` says `in todo since 2026-09-26 · 9 days`, the moment of
the last act that named the column, or of the creation when none has. The rule
is `trace.LastNamed`, arithmetic over acts like the timesheet's, so it lives
with them; the engine only asks it of a task's history.

## One thing a restore takes with it

`backup restore` replaces the store with an older one, and the trace goes back
with it: what happened after that backup was made is gone from the history as
well as from the tasks. That is inherent — the trace lives in the store it
describes — and it is written here so it is not discovered.
