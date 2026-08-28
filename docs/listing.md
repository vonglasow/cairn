# Narrowing a listing, and what looking still reaches

Two commands answer with several tasks. `task list` narrows by the facts that
come from a closed set; `task search` looks in what somebody typed themselves.
Neither reaches what the other does, and that boundary is the whole design.

## Why a listing narrows at all

A list of everything stops being readable once there is a lot of it. It also
stops being usable to an agent, which pays for every line it is handed — a
store with two hundred tasks in it hands over two hundred tasks to answer a
question about four.

Both halves matter, and only one of them is new. A person scrolls; an agent
runs out of room. The narrowing was asked for because of the second and is
worth having for the first.

## The reversal, written down

`docs/ordering.md` used to say a column filter had been weighed and left out,
on this reasoning: `task search` already answered close to it, since looking
reached every field including the column, and a filter would have been a second
way of asking one question.

The measurement in that paragraph was right and the conclusion was wrong.
`search todo` does return a task sitting in the backlog whose title happens to
contain the word — that was recorded as acceptable in a search. It is not
acceptable as the only way to ask for a column, which is what it had become.

So the decision reversed, and the second way of asking was removed rather than
tolerated: looking gave up the fields a listing narrows by. There is still
exactly one way to ask each question. It is just a different one.

## What each one reaches

| fact | narrowed by | looked in |
|---|---|---|
| title | | ✓ |
| description | | ✓ |
| deadline | | ✓ |
| estimate | | ✓ |
| time spent | | ✓ |
| column / where it stands | ✓ | |
| priority | ✓ | |
| label | ✓ | |
| objectives served | ✓ | |
| number | | |

The line falls between **a closed set or something numbered** and **free text**.
A closed set can be matched exactly, so it is narrowed by; free text cannot, so
it is looked in as a substring. A number is reached by neither, for the reason
`docs/ordering.md` gives: looking for "1" would answer 1, 10, 21 and everything
hanging under them, and asking by number is what `task show` is for.

What this cost: `search high` no longer answers a task carrying the high
priority. What it bought: `search high` no longer answers a task whose
description merely contains the word "highlight". The second was measured; the
first is what `--priority high` is for.

An objective is narrowed the same way, minus what it serves — nothing puts an
objective under another.

## Everything named has to hold

Two facts named are two conditions, not a choice between them. `--state todo
--priority high` keeps what is both, and `--label team --label urgent` keeps
what carries both. This is the rule looking already followed, kept rather than
invented, so the two commands do not disagree about what naming twice means.

## A narrowed listing is flat

An unnarrowed listing is a tree: sub-tasks nested under the task they hang
from. A narrowed one is flat, and a sub-task that matched hangs under nothing —
what matched is what is shown, and where each task sits is not what was asked.

Bringing the parent along was weighed and left out. A parent that does not
carry what was asked for is not part of the answer, and showing it to hold its
child would put a task in a listing of tasks that do not match it.

This is the same rendering `task search` already used, and it is the same field
of `Answer` that carries it. The field was called `Found` and is now called
`Matched`: what a search found and what a narrowing kept are one notion — these
matched, and where each sits is not what was asked — so they are one field.

## The focus is a flag on a task

Which task is being worked on is a flag the task carries, not a fact about the
listing. So a listing the task was narrowed out of says nothing about it:
`in_focus` reads null rather than naming a task the answer does not carry.

An unnarrowed listing carries every task, so nothing changes there. The case
only arises when narrowing excluded it, and there the alternative was worse — a
caller reading a number it cannot find among the tasks it was handed.

## A value outside a closed set is refused

`task list --state nowhere` fails and names what it accepts instead. It does
not answer "there is no task", which would be a lie a caller cannot tell from
the truth: asked for a column that does not exist, an empty answer reads as an
empty column.

Held against what the argument itself declares it accepts — `Argument.OneOf`,
which is already read off `domain.States` — so the closed set stays in the one
place it is written, and a listing of objectives is held to the objective
states rather than to a task's columns.

An objective nobody added is a different case: `--objective 9` succeeds and
answers nothing. There is no closed set to fall outside of, and whether an
objective exists is a question about objectives that reading tasks does not
ask. What it is still held to is being a number at all: `--objective 0` is
refused, because zero numbers nothing and answering "there is no task" would
read as an objective nobody is working towards rather than a filter nobody
could satisfy.

A task serves a set, so naming two keeps what serves **both**, the way two
labels do.

## Where the narrowing happens

In SQL, in `internal/store/narrowing.go`, not by reading everything and
dropping most of it above. What a narrowed listing is for is not handing over
what nobody asked for, and a filter that reads the whole store first only hides
the cost.

Every value crosses behind its own placeholder; the query string is built from
clauses written out in that file and from nothing a caller typed. A label named
`' OR 1=1` narrows to the tasks carrying that label and to nothing else.

`All` and `AllObjectives` are now the narrowed read with nothing named, rather
than a second loop over the same rows.

## Deliberately not there yet

- **Narrowing by a deadline** — `--due-before`, `--overdue`. Looking still
  reaches a deadline as a substring, which answers a month or a day but not a
  range. A range is a different kind of argument and has not been asked for.
- **Narrowing a search** — `task search spec --state todo`. It would work, and
  it would be the only place the two vocabularies meet. Left out until
  something needs it.
- **Any of, rather than all of** — `--state todo,current`. The one shape asked
  for was all of it; a set would need a second spelling and both commands would
  have to agree on it.
- **Narrowing what a trace answers** — `trace list` reads a period, which is
  already a narrowing of a different kind.
