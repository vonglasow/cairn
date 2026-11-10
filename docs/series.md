# A series: the tasks that come back

Read this before touching a series, a rule, or what is made again when a task
comes back.

A stand-up every weekday, a review every other Thursday, the plants three days
after they were last watered. Each of them is work that reaches `done`, and
then comes back. cairn left this out for a long time — see
[decisions.md](decisions.md#what-a-task-is-and-what-it-carries) — and took it
back in #247, for the reasons the study gave (`.cairn/notes/247/`): the work
is real, it is the same work every time, and copying a task by hand every
morning was the mechanism nobody wanted.

## A notion of its own

A **series** is a rule, what its next occurrence is made from, a state, and
the tasks it made. It has numbers of its own, written `↻3`, apart from the
tasks' `#312`, so a series is named even once every task it made is archived
or gone.

- **Its rule** says how often, in plain words (below), what it counts from —
  the calendar, or the day the last one was done — the day it counts from, and
  when it ends.
- **Its template** is what the next occurrence is made from: a title, a
  priority, labels, steps, an estimate and the objectives it serves. Not a
  description and not a deadline: a description is what one task is, and the
  deadline of an occurrence is its own day.
- **Its state** is active or stopped, and a stopped series keeps the day it was
  stopped.
- **Its occurrences** are tasks, each with a number, notes and time of its own,
  archived ones included. Nothing is summed per series in the timesheet and
  there is no notes folder per series: an occurrence is a task like any other.

There is no package for it, the way there is none for a view: its one rule with
weight — which days a rule falls on — is the rule's, and lives in
`internal/domain/rule.go` beside the days and durations; what ties a series to
its tasks is the engine's (`internal/engine/series.go`), and what keeps it is
the store's (`internal/store/series.go`).

## The rule, in plain words

`--every` takes the words a person says, and a series reads them back as a
sentence with its next three days:

```
$ cairn task new "Team stand-up" --every "every weekday 10:30"
…
repeats  ↻1 every weekday at 10:30, on the calendar, from Mon 9 Nov 2026, never ending
day      Mon 9 Nov 10:30
then     Tue 10 Nov · Wed 11 Nov · Thu 12 Nov
```

| written | read |
|---|---|
| `every day`, `every 3 days`, `every other day` | so many days apart |
| `every week`, `every 2 weeks` | so many weeks apart, on the weekday it starts on |
| `every month`, `every 3 months` | so many months apart, on the day of the month it starts on; a month without that day takes its last |
| `every monday`, `every mon, thu`, `every other thursday`, `every 4 weeks on monday` | on those days of the week, every so many weeks |
| `every weekday`, `every weekend` | Monday to Friday, Saturday and Sunday |
| `10:30`, `at 9:00` | the time of day it is for, shown and never counted |
| `after done`, `after the last is done`, `--after-done` | counted from the day the last one is done, not on the calendar |
| `10 times`, `until 2026-12-31` | when it ends: a number of times or a day, never both |
| `--from 2026-11-20` | the day it counts from; today when not said |

The leading `every` may be left out (`--every "weekday 10:30"`). A word it does
not know is refused naming it; an end said twice, a day of the week with days
or months, a day of the week counted after done, or anything past 999 is
refused saying why (`@series-grammar-refuses-two-ends`,
`@series-grammar-refuses-a-word-it-does-not-know`). What is kept is the rule
written one way, `every 2 weeks on thu 14:00`, which reads back as the same
rule.

"Every month on the 1st" was not asked for (answer 11 of the study) and is not
there: a month counts from the day the series starts.

### Hand-written, not rrule-go

RFC 5545's RRULE is the grammar behind every calendar, and
`github.com/teambition/rrule-go` reads and iterates it. It was weighed and
left out — see [dependencies.md](dependencies.md#weighed-and-not-added): the
subset asked for is days, weeks and months, days of the week, an interval and
an end, about two hundred lines with their reading; the words a person types
are not RRULE, so a reader of plain words was owed either way; and the library
has had no release since 2023.

## One open at a time, and when the next is made

At most one occurrence is open — neither done nor archived — and the next is
made in one of two ways.

- **An act that leaves a series with nothing open makes its next at once**:
  putting the open occurrence in done, skipping it, archiving or deleting it.
  The next is the first day of the rule after the last one made, and the days
  that passed meanwhile are **missed**: written in the series' trace, never
  made, never caught up (`@series-missed-are-not-made-up`). Counted after done,
  the next is so long after today, and nothing is ever missed.
- **With nothing open, cairn makes the next on its day, and never ahead**:
  before every act on the store, as it archives what it was told to and keeps
  the timer, it looks at each active series with nothing open — one whose last
  occurrence was done while recurrence was off — and makes its next on the day
  it falls, through `series`, missing nothing
  (`@series-made-on-its-day-and-never-ahead`).

The answer of the act that made one says so, `↻1 made #2 for Tue 10 Nov`, and
`↻1 missed Tue 10 Nov, Wed 11 Nov` before it; an agent reads the same under
`made`. An occurrence is made in `todo`, due on its own day, carrying the
template; a label the store no longer lets a task carry is left off it.

**An occurrence serves the objectives the template serves**, and its time and
its progress count there as any task's do
(`@series-an-occurrence-serves-the-objectives`).

## The acts

| act | what it does |
|---|---|
| `task new … --every "<rule>" [--after-done] [--from <day>]` | makes the task, the first occurrence of a new series, due on the rule's first day unless a deadline is given |
| `task edit N --every "<rule>"` | makes a task that is not one yet the first occurrence of a new series; refused on an occurrence, whose series' rule is changed with `series edit` |
| `task edit N … --series` | changes the occurrence, and the series too for the ones to come: whatever of title, priority, labels, estimate, steps and objectives is named |
| `task skip N` | archives the open occurrence as it stands, writes it skipped, and the next is made |
| `series list [--state active\|stopped] [--search …]` | every series, its next day, what it did so far |
| `series show ↻N` | its rule and next days, template, counts, every occurrence, archived ones included, and its trace |
| `series edit N` | the template, the rule, the day it counts from; the occurrences made never change |
| `series stop N` | nothing more is made; the open occurrence stays an ordinary open task |
| `series resume N` | from today: with nothing open, the next is made at once on the first day from today; nothing caught up, nothing written missed |
| `series delete N --keep-tasks \| --delete-open \| --delete-tasks --yes` | refused until what becomes of its tasks is said, and the refusal counts them |

On a board, "this one or the whole series" is a question asked on saving; on
the command line, `--series` is the answer.

## A series never takes back what it made

**Stopping a series or turning recurrence off never removes a task it made.**
Only an explicit choice when deleting a series can: `--keep-tasks` keeps every
one as an ordinary task, without its `↻`; `--delete-open` deletes the one open;
`--delete-tasks` deletes them all, each one as `task delete` would, its trace
kept. Deleting the series removes its rule and its template; its trace stays.

## Turned on per store

`[recurrence] on = false` is the default, and a store's `settings.toml` may say
`on = true` for its workspace alone — see
[config.md](config.md#tasks-that-come-back-turned-on-per-store). Off, there is
no `--every`, no `--repeats`, no `series` command and no series tool for an
agent; nothing is made; the series stay as they were, dormant; an occurrence
open is an ordinary task, with no `↻` on it (`@series-off-offers-nothing`,
`@series-off-leaves-the-tasks-ordinary`). Turned back on, every series is as it
was, and an active one with nothing open makes its next on its next day from
today (`@series-on-again-brings-it-back`).

A store opened with no home — the engine's own tests — reads no configuration,
and recurrence is on there, as every page is.

## Its trace: a table of its own

A series' acts — `made`, `missed`, `skipped`, `changed`, `stopped`, `resumed`,
`deleted` — are kept in `series_acts`, not in the tasks' `trace`. The trace is
read by task number everywhere — the timesheet, how long a task sat in its
column, the timer — and a missed day or a stop names no task: it would have
needed a number no task has, which every reader would then have to learn to
step over. What happens to an occurrence itself — made, archived when skipped,
done — is still in its own trace, as for any task. See
[trace.md](trace.md#a-series-keeps-a-trace-of-its-own).

Which series a task belongs to is read off that trace, never stored twice: an
occurrence is a task some `made` act of a series still there names. So is the
day an occurrence was made for, and how many times a series has come back,
against its end.

## In the store

Migration #22, "a task can come back, as a series": `series`, `series_labels`,
`series_objectives` and `series_acts`. Rolled back, every series, its rule, its
template and its trace go, and every task it made stays an ordinary task with
its notes, time and trace (`@series-migration-rollback-keeps-the-tasks`). A
label renamed or deleted follows onto the templates; an objective deleted is
let go of by them.

## Carried in a file

`task export` writes every series beside the tasks — its rule, the day it
counts from, the day it was stopped, its template and its trace — and the
objectives its template serves with the rest; `task import` makes each one a
new series, numbered after the ones already here, and its acts name the tasks
at their new numbers, so every occurrence is linked to its series again and the
one open makes the next when it is done
(`@exchange-a-series-comes-with-its-occurrences`,
`@exchange-an-imported-series-goes-on`). A task its series made and that was
deleted since is not in the file: its act comes back naming no task, and the
series' history still counts it. A series whose rule, days or template cannot be
read, or whose template serves an objective the file does not hold, refuses the
whole file before anything is added (`@exchange-a-series-outside-the-file`,
`@exchange-a-series-that-cannot-be-read`). Its days may be counted from today,
`"from": "-3d"`, as a task's are. The file went to version 6: an older cairn
would have read the tasks and dropped every series without a word.

A series comes in whether recurrence is on or not, and stays dormant while it
is off, as one already in the store does. `cairn fixtures` holds two, and turns
recurrence on in the settings of the store it makes.

## On a board

Both boards have them: the series tab, the `↻` and the rule on a card, the
`r` keys, the question asked on saving an occurrence, and deleting with its
three choices — see [tui.md](tui.md#series) and [gui.md](gui.md#series). What
a board reads of a series comes from the engine, written once for both:
`Look` reads every series whole, a card carries the series it is an occurrence
of, and `OccurrenceDay`, `SeriesFields` and `RepeatsField` say the day, the
form and the field; `Pages.Recurring` takes every series and every mark off
what a board shows while recurrence is off, `ShownSeries` lists them active
then stopped, `SeriesFates` and `MadeStill` are what deleting one offers and
counts, `ChangesTheSeries` says when saving an occurrence asks, and `DaySaid`
writes a series' day.
