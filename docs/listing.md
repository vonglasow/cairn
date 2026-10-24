# Narrowing a listing, and what looking still reaches

Two commands answer with several tasks. `task list` narrows by the facts that
come from a closed set; `task search` looks in everything a task says, free
text and closed sets alike. It is the one search, the engine's, and every
surface asks it.

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
tolerated: looking gave up the fields a listing narrows by.

## And reversed again

On 2026-10-13 looking took the column, the priority and the labels back. The
boards have one search box and no narrowing, so a label could only be found
there by searching for it — and the boards had kept a rule of their own that
did, while the command line's did not. Two rules for one word is the thing
this repository refuses, so there is one now, the engine's, and it reaches
everything. `search high` answers what carries priority high as well as what
says the word; `--priority high` is still there for the exact question.

## What each one reaches

| fact | narrowed by | looked in |
|---|---|---|
| title | | ✓ |
| description | | ✓ |
| deadline | | ✓ |
| estimate | | ✓ |
| time spent | | ✓ |
| column / where it stands | ✓ | ✓ |
| priority | ✓ | ✓ |
| label | ✓ | ✓ |
| objectives served | ✓ | |
| number | | ✓ |
| archived | ✓ | |
| deadline, as a range | ✓ | |

A closed set can be matched exactly, so it is narrowed by; everything a task
says, closed or not, is looked in as a substring.

A task's own number is looked in too, and it used not to be. The old reason —
that looking for "1" would answer 1, 10, 21, and that asking by number is what
`task show` is for — held only where there is a `task show` to type. On a board
the search box is the only way in, so somebody holding a number had no way to
use it. Amended in use, which is where that kind of thing is found.

It is looked for as a substring like everything else, so 228 answers 228 and
1228 both. Weighed and kept: a second kind of matching in a search that has
one costs more than a glance at one extra answer, and being handed nothing
costs the way back.

What a task *carries* about other things stays out of reach — what it serves,
what stops it. Reaching those would answer every task waiting on 228 to
somebody looking for 228, which is the half of the old reason that was always
true. Its steps are not about other things: they are part of what the task
says, and a word in one reaches the task (below).

## What was archived is off every listing until it is asked for

Archiving keeps a task without deleting it, out of its column. It is a flag beside
the column, not one of them: a task keeps the column it finished in, its
history and everything it carries, and stops being answered.

Every listing therefore narrows on it whether or not anybody named it — the
default is `archived = 0`, written in `whereTasks` before the clauses a caller
asked for. Two arguments change that: `--archived` keeps only what was taken
off, `--with-archived` keeps both. Objectives have no such flag and go through
`where`, which does not add the clause.

Looking is the exception a person expects: `task show 228` answers 228 whether
it is on the board or not, because somebody typing a number already knows which
task they mean.

Archiving takes a number or a narrowing, and `task archive --state done` sweeps
a column in one go. `task label` puts one label on everything a narrowing
reaches, which is how a sprint gets marked before it is swept. The label it
puts is its argument `put`, and `label` keeps the name it has in every other
narrowing — the labels a task already carries to be reached. On the command line
the first is positional and the second a flag, so the two never met; an agent
names both in one object, where a name is held once, so they could not share it.
Archiving answers one shape however many it took off: how many, and — when one
was named — the task as it ends up.

**A sweep acts on what the narrowing reaches, and both sweeps reach the
same.** Archiving and labelling go through `sweeping`, so the two answer one
board. They once did not: while sub-tasks existed, one sweep counted them and
the other did not, and archiving a finished column took 120 tasks where
labelling the same narrowing took 167. Steps are no tasks, so there is nothing
left for a sweep to count apart: a task's steps are archived with it, brought
back with it, and never reached on their own.

Both directions refuse a task already in the state asked for. Archiving twice
was refused first, because a sweep counts what it actually took and skips the
rest on that error; bringing back what never went away was not, for a while,
since nothing was counting. That asymmetry was an accident of what the sweep
needed, not a decision, and a pair of gestures that answer the same mistake two
different ways is a pair nobody can reason about. A board never meets either
refusal — `x` reads the card and picks the verb — so they exist for the command
line and the MCP, where a number is typed blind.

## A listing names the tasks, not their steps

A step is left out of `task list` (`@task-list-leaves-steps-out`). It is a
piece of its task, read where the task is read: `task show` names its steps and
says which are done (`@task-show-names-its-steps`), which is the question that
section answers. When the pieces were sub-tasks and listed under their parents,
a listing of 219 tasks showed 219 lines when 170 were the work and the rest its
pieces.

A machine is handed them where the task is shown: `task show --json` carries
`steps`, each a title and whether it is done
(`@task-show-json-carries-its-steps`), so reading a task answers what its steps
are called and which are done without asking again. `step list` answers the
same for one task alone; see [steps.md](steps.md).

## A listing hands over the facts, and leaves what a task says to be shown

A row of `task list --json` and `task search --json` — and of `task_list` and
`task_search`, which answer the same shape — carries every fact of a task and
`described`, whether it says anything; never the description itself. That is
read with `task show`, the way a person opens a card the board marks as
described.

Measured on a real board: 105 tasks handed over with their descriptions came to
106 000 characters, more than a client hands an agent in one answer — the
listing an agent most needs was the one it could not read. The descriptions
were seven tenths of it, and they grow: a task worked on for weeks carries its
account. So the shape keeps them apart (`task.Facts`, then `task.Shape` as the
facts and the description) rather than cutting a listing down after the fact.
A row carries no `progress` either, since a listing counts none and a field
that is always null is paid for on every row.

A step has a shape of its own — a title and whether it is done — rather than
being a task again: a task holding tasks is a shape no schema can describe, and
eight MCP tools once lost the shape of their answer the moment a task held a
task. The type says the rule.

## A word in a step answers the task it belongs to

A step's title is part of what its task says, so `task.Said` carries it, and a
word found only in a step reaches the task: `task search pipeline` answers the
task one of whose steps says it (`@task-search-reaches-a-step`), and the boards
raise that task, once however many of its steps say it
(`@tui-searching-raises-the-task-a-step-belongs-to`,
`@tui-searching-raises-each-task-once`,
`@gui-searching-raises-the-task-a-step-belongs-to`).

It has to, because a step has no number and no card of its own to appear on,
so a search reaching one and answering nothing would come back empty for a
word plainly in the store — measured on a real store, 44 of the 49 sub-tasks
that became steps carry a word their task does not, so that is most of them.
The command line and the boards used to differ here — the line named the
sub-task, the boards its parent — and with steps there is one answer, the task,
since it is the thing that can be acted on.

## One rule, the engine's

What a search looks in is `task.Said` — `objective.Said` for an objective — and
whether what was named sits in it is `domain.Reaches`. `task search` asks them
of the store; a board asks `engine.Reaches` of the card it already holds, whose
`Said` the engine filled when the board was read. There is no second list of
fields anywhere to drift.

There were three once: the engine looked in six fields, the terminal board in
three and the window in three, and then the boards in more than the engine.
Nothing here could have caught it. Coverage and mutation measure whether the
tests judge the code that exists, and three tested copies of one rule score
three perfect marks — mutation *rewards* duplication. The only cure was to
keep one.

An objective is narrowed the same way, minus what it serves — nothing puts an
objective under another, and an objective carries no steps.

## Everything named has to hold

Two facts named are two conditions, not a choice between them. `--state todo
--priority high` keeps what is both, and `--label team --label urgent` keeps
what carries both. This is the rule looking already followed, kept rather than
invented, so the two commands do not disagree about what naming twice means.

## A listing is flat

No task hangs under another, so every listing is flat, narrowed or not, and
comes back in the store's one order: the column first, then the position, then
the number. The position is counted within a column and nothing else (see
[ordering.md](ordering.md)), so a narrowed listing reads in the order the board
holds — no listing has an order of its own, and the same tasks come back in the
same order whether a filter was named or not.

While sub-tasks existed, an unnarrowed listing was a tree and a narrowed one
flattened it, so a sub-task and its aunt, both at position 1 under their
parents, interleaved and read as shuffled. That was accepted as it was rather
than sorted by number, and the fix it pointed to — a position counted across a
whole column — is what steps made of the order.

This is the same rendering `task search` already used, and it is the same field
of `Answer` that carries it. The field was called `Found` and is now called
`Matched`: what a search found and what a narrowing kept are one notion — these
matched, and where each sits is not what was asked — so they are one field.

## A row carries what a decision is made on

`#12  Write the spec` was the whole of a row: no column, no priority, no
deadline, no labels. Deciding anything meant opening every task one at a time,
and that listing is what an agent sees through the MCP server as well.

A row is now the number, the title, and then what it carries, each piece behind
a `·`:

```
#12  Write the spec  · current  · high  · 2026-10-09  · backend sprint-172
```

A task carrying none of them is the number and the title and nothing else, so
the shape does not grow for work that has not been filled in yet. An
objective's row says where it stands the same way — `reached` and `dropped`
were invisible in a listing of objectives.

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

**A refusal lists the vocabulary.** `"nowhere": not one of the columns —
backlog, todo, current, done`, and the same for a priority, for a day (`a day
written 2006-01-02`) and for a duration (`an amount and a unit, as in 2h or
1h30m`). All four used to say only what was wrong, which leaves the next
attempt to guesswork; all four are closed sets or one fixed shape, and saying
so costs one clause. The words come from `domain.States` and
`domain.Priorities` rather than being written out again beside the refusal.

Held against what the argument itself declares it accepts — `Argument.OneOf`,
which is already read off `domain.States` — so the closed set stays in the one
place it is written, and a listing of objectives is held to the objective
states rather than to a task's columns.

**One mistake, one reason.** The refusal wraps the sentinel the attribute is
refused with when it is given — `Argument.Outside`: a column that does not
exist is `not-a-state` to a machine whether a listing narrows by it or an
edit gives it, where it used to be `not-one-of` from one and `not-a-state`
from the other.

An objective nobody added is a different case: `--objective 9` succeeds and
answers nothing. There is no closed set to fall outside of, and whether an
objective exists is a question about objectives that reading tasks does not
ask. What it is still held to is being a number at all: `--objective 0` is
refused, because zero numbers nothing and answering "there is no task" would
read as an objective nobody is working towards rather than a filter nobody
could satisfy.

A task serves a set, so naming two keeps what serves **both**, the way two
labels do.

## A deadline narrows as a range, and is looked in as text

`--due-before 2026-10-01` keeps what is due on or before that day, the day
itself included: a deadline is the last day something may be done, so a listing
of what is due by the first that left out the first would answer a different
question.

What carries no deadline is left out. It is not due by that day — it is not due
at all — and a listing of what is coming up that answered every dateless task
would answer most of the board.

It is the one narrowing that is also looked in: a deadline is free text to
`task search`, which answers a month or a day as a substring. The two do not
collide because they answer different questions — "which tasks say October" and
"which tasks are due by the first".

A day that is not one is now refused by `Check`, beside the number that is not
one. It had to be: a narrowing compares a column of days against what was
typed, and left unread it answered whatever that sorted beside. Every `Day`
argument gains the refusal, not only this one.

## A search narrows the same way a listing does

`task search spec --state todo` is the one place the two vocabularies meet:
what is looked for is free text, what narrows it is a closed set. It went in
when the board grew long enough that a word worth searching for was a word
answering the whole of it.

It reads the narrowed listing and looks in that, rather than reading everything
and narrowing afterwards — so `--archived` reaches what was archived
and nothing else does, which is the same rule every other listing follows.

**The boards search the same way.** They narrow what they already hold rather
than asking the store on every keystroke, but what reaches a card is what
reaches a row: every fact a task carries, labels, column and priority included,
and every word typed has to be carried rather than one of them. They used to
look in six fields with one word, so one word, "search", meant two things
depending on where it was typed — and searching by label was impossible on
either board.

**The boards filter by the same narrowing.** Beside the search, both boards
offer the closed set a listing narrows by — column, priority, label, objective,
a day due by — as filters walked through their values (cairn#59). They narrow
what the board already holds, in memory, but by the narrowing `task list`
builds from the same flags (`engine.Filters.Kept`, through `keeping`), told to
each card instead of to SQL: one rule, read twice.

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

- **Any of, rather than all of** — `--state todo,current`. The one shape asked
  for was all of it; a set would need a second spelling and both commands would
  have to agree on it.
- **Narrowing what a trace answers** — `trace list` reads a period, which is
  already a narrowing of a different kind.
