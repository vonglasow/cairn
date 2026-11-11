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
| number | ✓ (`task list`) | ✓ |
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

## A listing carries only the fields asked for, when some are named

`task_list` and `task_search` take `fields` — `task list --json --fields number
--fields title` on the command line — and each row then carries those and no
other (`@mcp-a-listing-carries-the-fields-asked-for`,
`@mcp-a-search-carries-the-fields-asked-for`,
`@task-json-list-carries-the-fields-asked-for`). Nothing named, a row carries
every field, as it always did.

Leaving the descriptions out was not enough: on 2026-11-05 an agent listing a
board of 313 tasks, archived ones included, was handed some 90 000 characters,
past what its client hands over in one answer, when it needed the number, the
column and the title of each. Measured on this repository's board of 228 tasks:
62 899 characters whole, 20 462 with those three fields.

What was chosen, and why:

- **The vocabulary is the row's own.** A field is a key a row carries —
  `number`, `title`, `state`, `archived`, `priority`, `labels`, `blocked_by`,
  `linked`, `because`, `deadline`, `estimate`, `time_spent`, `objectives`,
  `described` — read off `engine.Row`, never written out beside it. The
  argument's description and the enum an agent reads are made from that one
  list, so a fact added to a row can be asked for the day it is there.
- **A name a row does not carry is refused**, `not-a-field`, naming every one it
  does (`@mcp-a-field-a-row-does-not-carry-is-refused`) — `description`
  included, since a listing never carries it: answering a row without the field
  asked for would read as a task that has none. A list naming nothing is
  refused too (`@mcp-asking-for-no-field-is-refused`): a row of no field is no
  answer, and answering every field would hand over the very answer that was
  too long.
- **A row keeps its own order**, not the order asked in, and a field named
  twice is carried once: the order of a row is then the same whoever asked.
- **The rest of the answer is untouched.** `in_focus` is said whatever was
  asked; it is not a field of a row.
- **The text is untouched too.** The line a person reads is already the short
  one — the number, the title, the column, the priority, the deadline, the
  labels — so `--fields` narrows `--json` and the answer a tool hands a machine,
  and `task list --fields title` without `--json` prints the line it always
  did. Shaping the line by the fields asked would be a second rendering of a
  row to keep true, for nobody who asked.
- **Several fields are several values**, the way two labels are: `--fields
  number --fields title`, `["number","title"]`. `--fields number,title` is one
  name that is not a field, and is refused saying so.
- **The command line has it because it is the same argument.** An argument is
  described once, in the engine, and both surfaces are built from it; an
  argument kept from one of them would be the first, and `--json` is there to
  print what a tool answers.

`fields` shapes what is handed back, never what is kept: it narrows nothing.

## A search can keep what carries any of what is named

`task_search` takes `any` — `task search T010 T013 --any` on the command line —
and then keeps every task carrying at least one of what is named rather than
all of it (`@task-search-any-of-what-is-named`,
`@mcp-a-search-finds-any-of-several-terms`). The answer then says, under
`matched`, which tasks each thing named matched:
`"matched":{"T010":[1],"T013":[2],"T999":[]}`
(`@task-json-search-any-says-what-each-term-matched`).

On 2026-11-05 an agent had to check whether ten ids from an old board each had
a task. Everything named having to hold, that was ten searches, or a whole
listing read by hand.

What was chosen, and why:

- **A switch on `named`, not a second list beside it.** `any` changes how what
  is named is read; it names nothing itself. A second list would have had to
  say what naming both means, and `named` being required would have had to
  give way to "one of the two", on both surfaces. Mixing all-of and any-of in
  one search is a question nobody asked: the case was a handful of ids, any of
  them. The narrowings still all hold, as they do on every search — `--state
  todo --any` keeps what is in todo and carries one of what is named.
- **Which term matched is said per term, beside the rows.** Each thing named
  maps to the numbers of the tasks it matched, in the listing's order, and a
  term that matched nothing maps to an empty list. The question asked was
  "which of these has a task", and a term with none is the half of the answer
  a row can never carry — it has no row to be said on. Inverting the map gives
  each task the terms it matched, so nothing is lost.
- **It is not a field, and `fields` does not touch it**
  (`@mcp-which-term-matched-whatever-the-fields`). It is about the search, not
  about a task, the way `in_focus` is; a row of three fields still comes with
  what matched. Put on the row it would have been a fact `task_list` could be
  asked for and never carry.
- **Only a search for any of it says it**
  (`@mcp-a-search-for-everything-named-says-no-term`). Asked for everything,
  every task kept matched every term, and saying so would be a copy of
  `named` per task.
- **A term is reported as it was typed.** The match ignores case, so `t010`
  and `T010` both match and are both said, each under its own spelling; the
  same term named twice is said once, and a task under it once.
- **The line a person reads is untouched.** The term is usually in what the
  line shows, and a second rendering of a search to keep true is not worth it
  for a person who can see the line; `--json` carries `matched`.

## A listing can answer only the tasks whose numbers are named

`task_list` takes `number` — `task list --number 25 --number 18` on the command
line — and answers those tasks and no other
(`@mcp-a-listing-answers-the-numbers-asked-for`, `@task-list-by-number`). A
number no task carries is said under `unknown`, in the order asked, and is no
refusal: `"unknown":[9]` (`@task-json-list-by-number-says-which-is-no-task`).

On 2026-11-05 an agent checking whether four tasks it already knew were done
had one `task_show` to make per task, or a whole board to list and read
through for four rows.

What was chosen, and why:

- **On `task_list`, not `task_show`.** `task_show` answers one task whole —
  its description, its steps, its context — and its answer is one task; made a
  list, it would answer two shapes, and a handful of descriptions is the very
  answer `fields` was made to keep short. `task_list` already answers rows,
  takes `fields`, and says the focus, so the numbers are one more narrowing of
  the shape it has.
- **Named `number`, given once per task**, the way a label is: `--number 25
  --number 18`, `[25,18]`. A list of task numbers is written as numbers or
  as strings, the way `blocker` and `link` are — see
  [mcp.md](mcp.md#what-changes-in-a-tools-answer-and-what-does-not).
- **Archived or not.** Somebody naming a number already knows which task they
  mean, which is why `task show` answers an archived task (above); a listing
  by number does the same, and a row says `archived`. `--archived` still keeps
  only what was taken off, so naming it narrows as it always did.
- **Everything else named still holds.** `--state done --number 1 --number 2`
  keeps what is both, like every narrowing. A number narrowed out is not
  `unknown` — the task exists and is not done, which is the answer asked for;
  `unknown` says only that no task carries it, deleted or never made
  (`@mcp-a-listing-by-number-is-narrowed-like-every-listing`).
- **`unknown` sits beside the rows, like `matched`.** A number with no task has
  no row to be said on, and it is about the asking, not about a task, so
  `fields` does not touch it
  (`@mcp-a-listing-by-number-carries-the-fields-asked-for`). It is left out
  when every number is a task (`@mcp-a-listing-by-number-reaches-an-archived-task`),
  the way `crossed` is when nothing moved.
- **Rows keep the board's order**, not the order asked in, and a number named
  twice is answered once — the same rule `fields` follows, so the same tasks
  read the same however they were asked for.
- **What is not a task number is refused**, `not-understood`: a word, zero, a
  negative, `1,2` as one value, or an empty list
  (`@mcp-asking-for-no-number-is-refused`,
  `@task-list-by-number-refuses-what-is-not-a-number`). An empty list answering
  the whole board would hand over what asking by number was there to avoid.
- **The line a person reads is untouched.** A person sees which tasks came
  back; `--json` carries `unknown`.

## A listing can answer a page of what it keeps, or only how many

`task_list` takes `limit` and `offset` — `task list --limit 50 --offset 100` —
and answers that page of what it keeps, the answer saying under `total` how
many it kept in all (`@mcp-a-listing-answers-a-page`,
`@task-json-list-a-page-says-how-many-in-all`). It takes `count`, `state` or
`label` — `task list --count state` — and then answers no row, only how many of
what it keeps sit in each column or carry each label, under `counts`, and how
many in all under `total` (`@mcp-a-listing-counts-by-state`,
`@mcp-a-listing-counts-by-label`, `@task-json-list-counts-by-label`).

`fields` keeps a row short, and a board keeps growing: past 300 tasks, archived
ones included, even short rows can exceed what an agent is handed in one tool
result. And "how many tasks in each column?" took the whole list, read and
added up by the agent.

What was chosen, and why:

- **`count` on `task_list` only.** A search keeps what carries what is named,
  and the case that grew too long was the listing of a whole board. A page
  came to `task_search` later, see below.
- **A page is cut from what everything else keeps, in the board's order** —
  the order a listing already answers in, so page after page reads the board
  through once (`@mcp-a-page-is-cut-from-what-is-kept`). The narrowings,
  `number` included, hold first; the page is cut last.
- **`total` sits beside the rows**, like `matched` and `unknown`: it is about
  the asking, not about a task, so `fields` leaves it alone. It is said
  whenever a limit, an offset or a count is asked, and left out otherwise
  (`@mcp-a-listing-asked-for-no-page-says-no-total`): without a page, the rows
  are the total.
- **`limit` alone, `offset` alone, or both.** A limit is at least 1 and an
  offset at least 0; anything else is refused `not-understood`
  (`@mcp-a-page-that-is-none-is-refused`, `@task-list-a-page-of-no-task-is-refused`)
  — a limit of 0 answering the whole board would hand over the very answer a
  limit was there to cut, and answering no row at all is what `count` is for.
  A limit past the last task answers what there is.
- **An offset past the last task is no refusal**: it answers no row, and
  `total` says how many there are (`@mcp-a-page-past-the-last-task-is-empty`).
  The board may have shrunk between two pages; that is an answer, not a
  mistake.
- **`count` answers no row**, and so refuses what shapes a row or a page —
  `fields`, `limit`, `offset` — `not-understood`, rather than ignoring them
  (`@mcp-counting-takes-no-row-argument`): an argument silently dropped reads
  as one that was obeyed. It counts what every narrowing keeps, `number`
  included, and `unknown` is still said. `tasks` stays, empty, since a listing
  always carries it.
- **Grouped by `state` or `label`**, the two named; anything else is refused
  `not-one-of`, naming both (`@mcp-counting-by-what-it-cannot-is-refused`). A
  column or a label no task kept has is left out, so a column missing counts
  none. A task carrying two labels counts under each, and one carrying none
  under no label, which is why `total` is said beside the counts.
- **A person reads it too.** Unlike `matched` or `unknown`, what was left out
  of a page cannot be seen from the lines, so the text ends with `tasks 101 to
  150 of 312`, or `no task past 400, of 312`; a count reads one column or label
  per line, columns in the board's order and labels by name, then `312 task(s)
  in all`. It is the text an agent is handed beside the JSON, too.

## A search pages its answer the same way

`task_search` takes `limit` and `offset` too — `task search spec --limit 50
--offset 100` — and answers that page of what the search kept, the answer
saying under `total` how many it kept in all
(`@mcp-a-search-answers-a-page`). A search keeping hundreds of tasks answered
them all at once, as a listing once did.

What was chosen, and why (cairn#242):

- **The page is cut after the search keeps its tasks**, in the board's order,
  the narrowings and what is named holding first — as a listing cuts its page
  last.
- **`total` counts every task the search kept**, not the rows on the page, so
  page after page reads the search through once. It is left out when no page
  was asked (`@mcp-a-search-asked-for-no-page-says-no-total`).
- **`matched` stays per row**: with `any`, it says which tasks each thing named
  matched among the rows on the page only, a term matching none of them
  answering none (`@mcp-a-search-page-says-what-matched-on-it`). `matched`
  explains the rows an agent was handed; saying of tasks it was not handed
  would name numbers it has no row for, and `total` already says how many
  there are beyond the page.
- **The same rules as a listing's page**: a limit below 1 or an offset below 0
  is refused `not-understood` (`@mcp-a-search-page-that-is-none-is-refused`),
  an offset past the last task answers no row and the total
  (`@mcp-a-search-page-past-the-last-task-is-empty`), and the text a person
  reads ends with which tasks of how many. A search has no `count`.

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

An objective is narrowed the same way, minus what it serves — an objective
serves nothing, though it may sit under another — and an objective carries no
steps.

## Everything named has to hold

Two facts named are two conditions, not a choice between them. `--state todo
--priority high` keeps what is both, and `--label team --label urgent` keeps
what carries both. This is the rule looking already followed, kept rather than
invented, so the two commands do not disagree about what naming twice means.
A search asked for `any` is the one exception, and only for what it looks for:
see above.

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
labels do. An objective named reaches its sub-objectives too: `--objective 4`
keeps what serves 4 or one of the objectives under it, as 4's numbers count
them, and `--itself` keeps to 4 alone. See
[objectives.md](objectives.md#two-levels).

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
a day due by — as filters walked through their values. They narrow
what the board already holds, in memory, but by the narrowing `task list`
builds from the same flags (`engine.Filters.Kept`, through `keeping`), told to
each card instead of to SQL: one rule, read twice. The label filter takes
several labels at once, a task kept when it carries every one — what
`--label` named twice already did (`Filter.Many`, `Toggled`). Each value says
how many tasks it would keep under the other filters as they stand
(`Filters.Counts`), counted by the same rule. The objectives are narrowed by
what an objective carries itself — its labels, its priority — or by which one
it is (`KeptGoals`), the rule `objective list` follows; a column or a day due
by means nothing to an objective and keeps every one (`ReachesGoals`).

## A view keeps the filters under a name

A view is the flags `task list` narrows by — a column, a priority, labels,
objectives with or without their sub-objectives, a day due by, the archived —
and a search, kept under a name and opened again in one act: `task list --view
<name>`, `view list`, `view show`. It holds nothing about layout: which tab,
which column the cursor is on, a sort. It is one narrowing, the engine's, and
no second language: a view held is told to the same `domain.Narrowing` the
flags are, and a flag given beside it narrows further, as two flags do
(`@view-new-keeps-the-filters`, `@view-flags-narrow-further`). Its search is
looked for the way `task search` looks, every word held
(`@view-new-holds-a-search`).

**It is kept in the store**, in three tables migration #21 adds — `views`,
`view_objectives`, `view_labels` — beside the work it narrows, carried by a
backup and a restore with it, never in `settings.toml`, which says how cairn
behaves and not what the work holds. Rolling the change back drops the three
and loses the views alone (`@view-migration-rollback-loses-the-views`).

**An export carries them.** `task export` writes every view in its place, its
name and its filters, and the objectives it names with the rest, even one no
task serves; `task import` adds each after the views already here, naming its
objectives at their new numbers (`@exchange-views-come-with-their-filters`). A
view whose name one here already has, folded the same way, is left out and the
one here stays as it was, as a label already here does
(`@exchange-a-view-named-here-stays-as-it-is`). A view that names an objective
the file does not hold, or that cannot be read — a name missing, a filter
outside what `view new` takes, nothing held — refuses the whole file before
anything is added (`@exchange-a-view-outside-the-file`,
`@exchange-a-view-that-cannot-be-read`). An objective gone is not written: it
narrows nothing already, and a view it leaves holding nothing is left out of
the file rather than come back as the whole board under a name.

**A name is one no view has**, whatever its case or its accents, folded the
way a label is (`@view-new-refuses-a-name-taken`), and a view holds at least
one filter or a search: one holding nothing would be the whole board under a
name (`@view-new-refuses-no-filter`). Its place, from 1, is the key that opens
it on a board — ⌘1…⌘9 in the window, `v` and a figure in the terminal; `view
edit --place` moves it and the others move along, and a deletion closes the
gap (`@view-edit-moves-its-place`, `@view-delete-touches-no-task`). `view
edit` takes the filters it holds one by one: a filter named takes the place of
the one held, named empty it goes, and one not named stays.

**It holds an objective by its number and a label by its name**, so a label
renamed is renamed in it, and an objective's tree is walked the day the view
is opened, reaching whatever sits under it then (`@view-follows-a-label-renamed`,
`@view-follows-the-tree`). A label a view names is kept as a label made on its
own, so it does not vanish the day nothing carries it.

**What it names may be gone.** A label deleted, an objective deleted and the
view kept, is a condition kept and marked gone: it narrows nothing, and every
answer says so, naming the commands that mend it (`@view-label-deleted-is-gone`).
Narrowing nothing rather than keeping nothing is deliberate: an empty board
reads as "no work", which is worse than a wider one said out loud. The one
exception is a view whose every condition is gone: narrowing nothing there
would show the whole board under a name that says otherwise, so it keeps
nothing and says why (`@view-every-condition-gone-keeps-nothing`). `label
delete` says how many views held the label; `objective delete` lists the views
naming it and asks what becomes of them ([objectives.md](objectives.md)).

**What lies outside is named apart, and never counted.** For the tasks a view
keeps, a task outside it that blocks one of them, waits on one of them or is
linked to one is named under `outside`, once, with every way it touches
them, after the rows and marked as outside
(`@view-outside-named-once-with-every-why`, `@view-outside-waits-on`). One
step only: two would bring the whole store back through the side door. It is
no row, so a listing stays flat and its counts and its pages are the view's
own; a listing without a view names nothing outside
(`@view-outside-not-without-a-view`).

**Changing a view changes no task**, so it writes no trace
(`@view-change-writes-no-trace`). And no figure about an objective moves with
a view: an objective's progress and time stay the store's, whatever view is
open. `view show` says so for each objective the view holds and still has,
one line each — `objective #1 Home automation  1/2 · 50% · 1h30m (the
store's)` — the same figures `objective show` gives, counted over its whole
subtree and never over what the view keeps; a machine reads them under
`objectives` (`@view-show-gives-the-objectives-figures-the-stores`). An
objective gone gives none: its condition already says it is gone.

**An agent reads the views and keeps none.** `view_list` is offered,
read-only, and `task_list` takes `view` and answers `outside`; `view new`,
`edit`, `show` and `delete` are the person's alone, beside `workspace add`
(`@view-mcp-only-reads-the-views`). An agent that needs a narrowing passes
the flags. No view opens on its own when a board opens, and an objective's
"show its tasks" is the same narrowing, a view without a name.

## A listing can keep the occurrences of a series

`task list --repeats` keeps only the tasks a series made, and `--series 3`
only those of series ↻3, archived ones too once `--with-archived` is said, as
every listing; a number no series carries is refused. With recurrence on, a
row says how often an occurrence comes back, `· ↻3 every weekday at 10:30`,
and an agent reads its series' number under `series`, 0 for any other task;
with recurrence off neither is said, and neither filter is offered
(`bdd/series.feature`, `@series-list-repeats`). Both boards offer the same as
a filter, `repeats` on `yes`, after the others but the focus, and only while there is
a series (`engine.BoardFilters`, `@gui-series-the-board-narrows-to-what-comes-back`,
`@tui-series-the-board-narrows-to-what-comes-back`). The narrowing is the engine's,
over what the store already handed back, like a view's search — see
[series.md](series.md).

## A listing can keep the tasks in focus, or the ones that are not

`task list --focus` keeps only the tasks in focus, `--no-focus` only the
ones that are not; both at once keep nothing and are refused, naming both
(`@ordering-list-in-focus-or-not`, `@ordering-list-in-focus-and-not-is-refused`).
The focus lives in its own table, so it narrows in SQL like the rest:
`number IN (SELECT task_number FROM focused)`, for tasks only — an objective
holds no focus, and a narrowing that names it reads every objective. A view
keeps it as `focus`, `in` or `out` (migration 25, `@view-new-holds-the-focus`),
and both boards offer it as their last filter, `focus` on `in focus` or `not
in focus` (`@gui-the-board-narrows-to-what-is-in-focus`,
`@tui-the-board-narrows-to-what-is-in-focus`). Without it, the one way to see
only what was being worked on was `task focus`, which no board offers.

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

- **Any of, rather than all of, for a narrowing** — `--state todo,current`.
  The one shape asked for was all of it; a set would need a second spelling
  and both commands would have to agree on it. What a search looks for can be
  any of it (`any`, above); what narrows it still all holds.
- **Narrowing what a trace answers** — `trace list` reads a period, which is
  already a narrowing of a different kind.
