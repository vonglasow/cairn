# Tasks that come back

A stand-up every weekday, a review every other Thursday, the plants three days
after they were last watered: work that reaches `done` and then comes back. In
cairn that is a **series**, and each time it comes back is a task of its own.

## Turning it on

Recurrence is off until a store's `settings.toml`, or the configuration, says:

```toml
[series]
on = true
```

Off, there is no `--every`, no `series` command and no series tool for an
agent, and nothing is made; series already there sleep as they were, and wake
up as they were when it is turned back on. See
[configuration](config.md#tasks-that-come-back).

## A series

A series has a number of its own, written `↻3`, apart from the tasks' `#312`.
It holds:

- **a rule**: how often, and whether it counts on the calendar or from the day
  the last one was done;
- **a template**, what the next task is made from: a title, a priority, labels,
  steps, an estimate and the objectives it serves — no description and no
  deadline, since those belong to one task;
- **a state**: active or stopped;
- **its occurrences**: the tasks it made, each with its own number, notes and
  time.

```sh
cairn task new "Team stand-up" --every "every weekday 10:30" --label work
cairn task new Retrospective --every "every other thursday 14:00" --from 2026-11-19
cairn task new "Water the plants" --every "every 3 days" --after-done
cairn task edit 12 --every "every monday"    # an ordinary task becomes the first
```

A series reads its rule back as a sentence, with its next days:

```
repeats  ↻1 every weekday at 10:30, on the calendar, from Mon 9 Nov 2026, never ending
day      Mon 9 Nov 10:30
then     Tue 10 Nov · Wed 11 Nov · Thu 12 Nov
```

## The rule, in plain words

| written | read |
|---|---|
| `every day`, `every 3 days`, `every other day` | so many days apart |
| `every week`, `every 2 weeks` | so many weeks apart, on the weekday it starts on |
| `every month`, `every 3 months` | so many months apart, on the day of the month it starts on; a month without that day takes its last |
| `every monday`, `every mon, thu`, `every other thursday`, `every 4 weeks on monday` | on those days of the week, every so many weeks |
| `every weekday`, `every weekend` | Monday to Friday, Saturday and Sunday |
| `10:30`, `at 9:00` | the time of day it is for, shown and never counted |
| `after done`, `after the last is done`, `--after-done` | counted from the day the last one is done |
| `10 times`, `until 2026-12-31` | when it ends: a number of times or a day, never both |
| `--from 2026-11-20` | the day it counts from; today when not said |

The leading `every` may be left out. A word cairn does not know is refused,
naming it, and so is a rule that contradicts itself.

## One open at a time

At most one occurrence is open. The next is made:

- **at once**, when the open one is done, skipped, archived or deleted. It
  takes the first day of the rule after the last one made. Days that passed
  meanwhile are written **missed**, and never made or caught up. Counted after
  done, nothing is ever missed.
- **on its day, never ahead**, when nothing is open — for instance a series
  whose last task was done while recurrence was off.

The answer says it: `↻1 made #2 for Tue 10 Nov`, after
`↻1 missed Tue 10 Nov, Wed 11 Nov` when days were missed.

An occurrence lands in `todo` unless the series says another column
(`backlog`, `todo` or `current`), due on its own day, carrying the template,
and serving the objectives the template serves.

## Working with a series

| command | what it does |
|---|---|
| `task skip N` | archives the open occurrence, writes it skipped, and makes the next |
| `task edit N … --series` | changes this occurrence and the ones to come: title, priority, labels, estimate, steps, objectives |
| `series list [--state active\|stopped] [--search …]` | every series, its next day, what it did so far |
| `series show N` | its rule and next days, template, every occurrence, and its trace |
| `series edit N` | the template, the rule, the day it counts from, the column; tasks already made never change |
| `series stop N` | nothing more is made; the open occurrence stays an ordinary task |
| `series resume N` | from today; nothing is caught up |
| `series delete N --keep-tasks \| --delete-open \| --delete-tasks --yes` | refused until you say what becomes of its tasks |

On a board, saving an occurrence asks "this one or the whole series"; on the
command line, `--series` is the answer. Both boards have a series tab, show
`↻` and the rule on a card, and offer the same three choices on deleting.

## A series never takes back what it made

Stopping a series, or turning recurrence off, never removes a task. Only
deleting a series can, and only as you choose: `--keep-tasks` keeps them all as
ordinary tasks, `--delete-open` deletes the open one, `--delete-tasks` deletes
them all — each one's trace kept.

## Carried in a file

`task export` writes every series beside the tasks: its rule, its days, its
template and its trace. `task import` makes each one a new series, its tasks
tied to it again at their new numbers, so the open one goes on. A series comes
in whether recurrence is on or not.
