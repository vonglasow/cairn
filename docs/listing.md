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
| number | | ✓ |

The line falls between **a closed set** and **free text**. A closed set can be
matched exactly, so it is narrowed by; free text cannot, so it is looked in as
a substring.

A task's own number is looked in too, and it used not to be. The old reason —
that looking for "1" would answer 1, 10, 21, and that asking by number is what
`task show` is for — held only where there is a `task show` to type. On a board
the search box is the only way in, so somebody holding a number had no way to
use it. Amended in use, which is where that kind of thing is found.

It is looked for as a substring like everything else, so 228 answers 228 and
1228 both. Weighed and kept: a second kind of matching in a membrane that has
one costs more than a glance at one extra answer, and being handed nothing
costs the way back.

What a task *carries* about other things stays out of reach — what it hangs
under, what it serves. Reaching those would answer every sub-task of 228 to
somebody looking for 228, which is the half of the old reason that was always
true.

## One rule, and it had three implementations

`reading.ReachesCard` is what a search reaches, and both boards ask it. They
did not: the engine looked in six fields, the terminal board in three and the
window in three, so `search 1h30m` answered on the command line and nothing at
all on either board. Nobody had decided that; it simply drifted.

Nothing here could have caught it. Coverage and mutation measure whether the
tests judge the code that exists, and three tested copies of one rule score
three perfect marks — mutation *rewards* duplication. depguard forbids a
dependency and cannot require one: it stops the window reaching the terminal,
which makes copying five lines easier than lifting them into `reading`. The
one rule that would have caught it — look for it before you write it — is the
one with no check behind it.

What holds the two together now is a test rather than a wish:
`TestReaches_answersWhatTheCommandLineAnswers` builds the same facts through
`task.AsTold` and as a `Card`, and asserts `ReachesCard` and `task.Matching`
answer the same of every query. If one moves without the other, it fails.

Its twin `reading.ReachesGoal` had been there all along, asked by both boards.
The place to share was never missing — only the looking.

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

### The order of a flat listing is the tree's, flattened

A narrowed listing comes back in the store's one order: the column first, then
the position, then the number. The position is counted **within a column and
under one parent**, so a sub-task and its aunt both hold position 1 — and
flattened, they interleave:

```
$ cairn task list --state todo
#1  Write the spec
#3  Write the tests      # a sub-task of #1, position 1 under it
#2  Ship it              # position 2 among the root tasks
```

Read as a tree that is exact; read flat it looks shuffled. Measured, put to the
user, and **accepted as it is** rather than sorted by number: the alternative
would give a narrowed listing an order of its own, so the same tasks would come
back in two different orders depending on whether a filter was named. One order
that reads oddly beats two that disagree.

If it ever stops being acceptable, the fix is not sorting the answer: it is
`docs/ordering.md`'s question of whether a position should be counted across a
whole column rather than within a parent — which is a change to the order
itself, not to how a listing is printed.

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
