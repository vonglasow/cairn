# Ordering

Read this before touching how tasks are ordered, or before adding anything that
decides what to do next.

`bdd/ordering.feature` says what happens; this file keeps the reasons, the way
[mcp.md](mcp.md) does for `bdd/mcp.feature`.

## The question this answers

cairn already knows a great deal about a task, and none of it says which one to
do. The question is not what a task carries — it is how a list becomes an
order.

Two things had to be true of the answer before anything else: it holds in the
store, so it survives without a model in the room; and it is yours, so nothing
rewrites it behind your back.

## Three orders, and they are not the same thing

Conflating them is what made the question look hard.

| | |
|---|---|
| **arrival** | the number. Strictly increasing and never reused — measured: delete `#2` and the next task added takes `#4`, and sub-tasks share the same run. Free, already there, and it says nothing about what to do |
| **the criteria** | priority, deadline, labels, the objective served, how far the sub-tasks have got. They **propose**. They never decide |
| **the order to do** | yours, stored. This is the one that was missing |

The middle row is the one that misleads. An estimate is a guess, and it fails
worse than it is unreliable: four hours estimated against eleven spent leaves
minus seven, which reads as *already finished*. And a priority is a judgement
cairn cannot date, so a `critical` set in March is indistinguishable from one
set this morning.

Neither of those breaks an order made by hand. They can only make a poor
suggestion, which is seen and corrected. That is the whole reason the criteria
were given no power to write.

## There is one list, and the column is its coarse order

A task sits in one of seven columns, and the columns are read in this order:

| | |
|---|---|
| `backlog` | first, because this is the part not yet decided — the work of prioritising |
| `todo` | next, because this is the order in which to do them |
| `current` | next, because this is what is already under way |
| `blocked` | next: under way without you, and waiting on somebody who needs chasing |
| `watching` | next: handed over, and needing checking rather than chasing |
| `paused` | next, because it waits on nothing but your own decision to pick it up |
| `done` | last, because it is done |

The first three are what you can act on and the last is history; the three in
between all mean "not moving", and they are three words rather than one because
what gets them moving again differs. `blocked` needs a nudge to somebody else,
`watching` needs a look, `paused` needs you to choose. Collapsed into one
column, the list would be shorter and "what do I do about this?" would have no
answer in it.

That is a reading order, not a doing order, and the distinction matters: the
list is read top to bottom as *what needs deciding, then what is decided, then
what is in flight, then what is history*.

So a list is never four lists.

Asking for one column on its own was weighed and left out here, then added
later: `task list --state todo`. The reasoning that left it out was that
`task search` already answered close to it, searching every field including the
column. That was measured and true — `search todo` returns a task sitting in
the backlog whose title happens to contain the word — and it was the wrong
conclusion, because it made an inexact answer the only way to ask. Looking gave
up the column instead, so there is still one way to ask. See
[listing.md](listing.md). Nothing in this design rests on either.

## A position is counted within its column

Each task carries an integer position, and the order is the column followed by
the position. The positions are counted **within a column**, not across the
store, and that is not an optimisation — it is what makes the acted design
literal. Landing at the end of a section is one more than the largest position
in that column, and the invariant is local: within a column and under one
parent, positions run 1..N with no hole.

Holding that invariant costs one thing that is easy to leave out, and was:
**every departure closes the gap behind it.** A task deleted, a task given
another column, a task given another parent — each leaves a place empty, and
nothing looks wrong when it does, because the order still reads correctly with
the places running 2, 5, 9.

What goes wrong comes later, and there is exactly one place it shows: a move to
the last place aims one past the tasks in the group and finds that place
already taken, so the task does not go last. That is what the scenarios stand
on — one per departure, each built so the expected order comes out wrong while
a hole is there. Measured on the way in: with the gap left open, three
scenarios out of thirty-two fail and the other twenty-nine do not, which is
what says they are about this and nothing else.

A test walks every group after every act that can touch a place as well, and
fails unless each runs 1..N. The invariant is worth asserting directly and not
only through the one act that exposes it.

Counting across the whole store would let a `done` task hold position 2 while a
`backlog` task holds 50. The order still reads correctly, but the numbers stop
meaning anything on their own, and the invariant can only be checked after
sorting everything.

Inserting shifts what follows, and that is the truth rather than a cost: put a
task in front and the others come after. Measured: shifting every position
across 2000 tasks takes under 10 ms. So there are no sparse integers to
renumber on a path that almost never runs, and no linked list whose broken link
would truncate the list in silence.

## What recalculates a place, and what must not

A place is worked out again whenever a task joins or leaves the tasks it sits
with, and at no other time. Both halves matter, and each has something holding
it — written down here because two of these guards are not obvious from the
scenario that carries them, and a reworded scenario would take one away without
anybody noticing.

| the change | the place | what holds it |
|---|---|---|
| a task is added | last of its group | `@ordering-added-last`, `@ordering-last-means-last-in-its-column` |
| a task is deleted | the group closes up behind it | `@ordering-a-deletion-leaves-no-hole` |
| the column changes | last of the new group, old group closes | `@ordering-a-column-change-lands-it-last`, `@ordering-a-column-left-leaves-no-hole` |
| the parent changes | last of the new sisters, old sisters close | `@ordering-a-sub-task-given-another-parent`, `@ordering-a-parent-left-leaves-no-hole` |
| a task is moved | the places shift inside the group | `@ordering-moved-first`, `@ordering-moved-after`, `@ordering-moved-last` |
| anything else a task carries | **untouched** | `@ordering-an-ordinary-edit-moves-nothing`, `@ordering-a-title-change-moves-nothing` |

Two of those guards were measured rather than assumed, by breaking the code on
purpose and counting who noticed. Neither is visible from the scenario's own
words, so both are written here:

**That a group starts at one** is held by `@ordering-moved-first` and
`@ordering-sub-task-moved-among-its-sisters`. Turning `COALESCE(MAX, 0) + 1`
into `COALESCE(MAX, -1) + 1` makes every group start at zero and those two turn
red. It is caught as a side effect: neither scenario says anything about where
a group starts, and a scenario saying it would be the same scenario written
twice, since the only behaviour that tells zero from one is the one they
already describe. So it is recorded here instead.

**That a departure closes the gap** is held by the three `@ordering-*-leaves-
no-hole` scenarios. Leaving the gap open turns exactly three of the thirty-odd
ordering scenarios red and leaves the rest green, which is what says they are
about this and nothing else.

Neither of these lives inside gremlins' reach: both are arithmetic inside a SQL
string, which it does not read. The mutation gate says nothing about them, and
that is precisely why they are measured by hand and written down.

## A sub-task is done or it is not

A sub-task is a part of the work its parent names, split off to be sure the
whole is finished. The only thing worth knowing about a part is whether it is
done: where the work stands is the parent's to say, because a part being
"blocked" is the parent being blocked and a part being "current" is the parent
being worked on.

It carried all seven columns and used two. Measured on a real store of 219
tasks: its 49 sub-tasks were `done` forty-eight times and `todo` once, never
`current`, never stopped, never carrying a deadline or a blocker of their own.
The model was not narrowed here so much as caught up with what the usage had
already settled on.

A task hung under another is **settled** on the way in: `done` stays `done`,
every other column becomes `todo`. What it was doing on its own stops being
true the moment it becomes a part of something — except being finished, which
stays finished. The ninth change to the store's shape does the same to what was
already there, and does not undo: the column each sub-task stood in is gone by
then, and putting them all back in the backlog would be inventing an answer
rather than restoring one.

## A sub-task takes its place among its sisters

A sub-task carries a column of its own — measured: one can sit in `current`
while its parent sits in the backlog — and `task list` already shows sub-tasks
nested under their parent. Those two facts cannot both place a sub-task in the
list: it would have to be under its parent in one section and in another
section at the same time.

So nesting is the fine order one level down. Where the parent sits carries the
whole block, sub-tasks order themselves among their sisters, and a sub-task's
own column says how far along it is rather than where it goes.

That is not only a resolution, it is what keeps moving from growing a second
meaning: a sub-task is ranged among its sisters exactly as a root task is
ranged among the tasks of its column.

The refusal generalises with it — a sub-task cannot be moved behind a root task
— but the reason it gives does not, and that was worth its own wording. Between
two columns it is the column that differs; between a sub-task and a root task
it is the level, and saying "not in that column" of two tasks that may well
share one would be false. So this refusal names whose sub-task it is instead,
which is also the useful thing to be told: it says where the task *can* go.

The cost is real and small: a sub-task under way does not show up in a
`current` section. That is the question `focus` answers, and it answers it
better — what is being worked on is read from a mark, not from a section.

## A task's parts are written with it

A board writes a task's parts in its form, one a line — `#4 [x] its title` for
a part there already, `[ ] its title` for a new one — and `task new` and `task
edit` take those lines as `--parts`. They are settled in the act that saves the
task: added, renamed, ticked and put in the order the lines give, column by
column, every one of them or none. Saved one act a part, as they were, a part
refused halfway left half of them written.

A part the lines no longer name is left where it is. Deleting a task is an act
of its own, at its own level of access, and saving a list is not a way around
it: a board deletes those parts one by one, once the task is saved.

## A parent is not done while something under it is open

A parent is done *because* what hangs under it is done. If it can be finished
with a sub-task still open, then either that sub-task was never part of it — in
which case it is another task — or the parent was closed over work nobody did.

Refused rather than cascaded. Marking the children done on the parent's behalf
would finish work nobody finished, and silently; the refusal names which ones
are in the way, which is the useful thing to be told. It says where the work
*is*, the same way the refusal about sisters says where a task *can* go.

Asked of `done` alone, and only when that is where the task lands and is not
where it was. `blocked`, `watching` and `paused` say something about the parent
itself; `done` is the only state that claims something about what hangs under
it. And a parent already finished stays editable for anything else — asking the
rule of every edit would make it fail for a reason nobody caused today, since
the sub-task was either open when it was finished or reopened since, and
neither is this edit's doing.

The rule guards the transition and not the stored state. Rows already in that
shape are left alone: reopening work somebody marked finished, on their behalf
and without asking, is worse than an inconsistency that nothing depends on.
Measured on a real store of 235 tasks — none at first, one a week later, which
is the rule finding its own case.

## `focus` is what is being worked on

A day is not a queue. Ten minutes on `#1`, an hour on `#2`, ten minutes on a
`#3` that has just arrived, ten minutes back on `#2`, twenty back on `#1`
because feedback came in. It **goes backwards**, and the head of a list does
not.

So this is a thing of its own, and the closest thing to it is git's `HEAD`: a
reference that moves constantly and remembers nothing about where it has been.

**It held one task, and now holds as many as are under way.** Amended
2026-09-14. One at a time is how a person works, and it was stored to match:
one row, pinned to one row by a constraint, so the single holder was true by
construction rather than by a rule anybody had to keep. But an agent session is
started per task, and several run at once — a focus that held one made every
other one invisible to whoever asked what was being worked on. It is a row per
task now, keyed by the task so nothing can be held twice, and the order the
rows were made is the order they were focused in. Holding none is an ordinary
state, not a missing row.

What that costs is worth naming: the single holder no longer comes free from
the shape of the table, and "focusing what is already in focus changes nothing" is
now a rule the store keeps rather than one the schema makes impossible to
break. It is kept in one statement — an insert that ignores a conflict — which
is the smallest place it could live.

**A task in `done` holds no focus.** Putting a task in `done` — by an edit, a
move, a board — lets go of its focus in the same act, and the trace says it was
released then; putting a task already done in focus is refused
(`@ordering-focus-let-go-when-done`, `@ordering-focus-refused-on-done`). A
focus is the work under way, and a done task has none left: a focus left on it
went on counting time nobody spent.

A task it names may be deleted, and it is unfocused when that happens —
by hand, in the same statement's company as the labels. The schema declares the
reference `ON DELETE SET NULL` and that clause never fires: SQLite leaves
foreign keys off unless a connection turns them on, and cairn's does not. The
`ON DELETE CASCADE` on the labels has never fired either; the labels disappear
because `Delete` clears them itself.

That is worth writing down because it was believed the other way round first,
on a measurement that looked conclusive and was not: deleting a task through
`sqlite3` with foreign keys off leaves its labels behind as orphans, while
deleting it through cairn does not — which proves only that cairn cleans up,
not that the database does. The scenario is what caught it.

The name was measured too. `current` was taken — it is a column, and three
tasks can sit in it at once. `head` invites the one misreading this design had
to undo, since focus is precisely *not* the head of the list. `doing`, `now`
and `active` already appear 10, 14 and 27 times as ordinary English in the code
and the docs, so as a term each would blur into prose. `focus` appeared
nowhere. It carried a singularity that no longer holds, and the word was kept
anyway: renaming a term that is in the store, three surfaces, the MCP contract
and every scenario would cost more than the slight stretch of "several things
in focus" costs a reader.

## Moving reorders; changing column is an edit

Moving a task places it inside the column it is already in — to the front,
after a named task, or at the end. It never changes the column.

The cost is deliberate and was chosen with it in view: an urgent task that has
just been created takes two acts, one to bring it out of the backlog and one to
put it at the front. What is bought is that a move has one meaning. It also
makes one refusal fall out rather than be invented: naming a target in another
column is refused, because a move cannot reach there.

Changing the column is an edit, and it lands the task at the end of the new
section.

**Both, in one act.** `task move --state todo --after 3` changes the column and
then places the task, in the same transaction: a place refused leaves the
column as it was. A card dropped on a board somewhere is one gesture, and it
was two acts — a refused place left the column changed. The move still means
one thing; the column is only where it moves first.

## The criteria answer a question, and never write

They are asked, they answer, they touch nothing. You give the moment's context
— the time there is, whether the mind is fresh — and the answer names a task
and says why. The context is never stored, because it is true of the moment and
not of the task.

Two other roles were considered and rejected: proposing a place when a task
arrives, and proposing a reordering to accept or ignore. Both help *build* the
order. The one kept helps *depart* from it, which is what the question was, and
it is the only one that cannot damage the list because it writes nothing at
all.

## What the store gains

| | |
|---|---|
| a position on a task | 1..N with no hole, within its column and under one parent |
| `focused` | a row per task being worked on, in the order they were focused |

One column and one table, then, and not the two columns this design first
carried. **When a task last changed column is not a column here.** It was
settled as an element and it stays settled — it is the one date no other field
implies — but it is not stored beside the task: it is derived from the trace,
which records every act performed on one. See [trace.md](trace.md).

`task new` places a task at the end of the backlog, `task list` and `task
search` order by column then position, and `task edit` moves a task only when
it leaves the tasks it sat with — a different column, or a different parent —
landing it at the end of its new group and closing the gap behind it. Editing a
title or a priority moves nothing. Three operations join the ones already
there: moving a task, putting one in focus or asking which one is, and letting
go of it.

## Weighed and left out

| | why |
|---|---|
| **urgency as a field** | it says what `priority` and the column already say |
| **`estimate` minus `time spent`** | it goes negative the moment the estimate is passed, and reads as finished |
| **a precedence between two tasks** | measured: `4 before 2` is satisfied by 12 of the 24 orders of four tasks. It constrains; it does not order |
| **recurrence** | a real problem was named — sorting email never reaches `done` — and turning it into a mechanism was not asked for |
| **the date a task was created, the date a priority was set, the date anything last changed** | four dates were offered and four were taken, which is a menu artefact. Creation nearly doubles the number, and a priority's date is a special case of the last change. One carries a signal nothing else does: how long a task has sat where it is |
| **the criteria ordering the list themselves** | they are a force of proposal, not gospel — and nothing that writes can be that |

## Deliberately not here yet

- **When a task last changed column, kept anywhere.** It is not a column, and
  it is shown all the same: `task show` reads it off the trace — the last act
  that named the column, or the creation — and says `in todo since 2026-09-26 ·
  9 days`, `in_column_since` to a machine. See [trace.md](trace.md).
- **What makes a task doable right now.** Two properties were found — being
  stoppable at any moment without leaving anything half-done, and being a short
  action followed by waiting — and the list plainly does not close. Each
  example adds one.
- **Work that never finishes**, such as sorting email, and **work that comes
  back**, such as a report. cairn only models work that reaches `done`.
- **Meetings.** They are not candidates for what to do next; they consume the
  time, so they belong to the moment's context.
- **The trace of where focus has been**, with durations. In git's terms this is
  the reflog to focus's `HEAD`, and git keeps them apart for the same reason:
  `HEAD` does not know where it came from, and that is what keeps it simple.
  Designed now, and in its own document: taking and releasing the focus are two
  of the acts the trace records, and the two notions of time stay apart there —
  the total, for a timesheet, typed by hand; the shape of a day, derived. See
  [trace.md](trace.md).
