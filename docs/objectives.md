# What a task serves

A task serves as many objectives as it actually serves. It used to serve one,
and this is what that cost and what changing it cost back.

## Why one was not enough

An objective is what work is for, and one piece of work can be for two things.
Made to serve one, a task that advanced two had to pick — and the objective it
did not name then counted it as nothing, so how far along that objective was
read low. The fix, before this, was to remember what cairn had not been told.

That is the whole reason. It is not a modelling preference: a wrong number on a
card is worse than no number, because a wrong one gets believed.

## How it is said

The grammar of the labels and the blockers, so there is nothing new to learn:

```sh
cairn task new "Write the spec" --objective 1 --objective 2
cairn task edit 1 --objective 3      # one more, without naming the others
cairn task edit 1 --objective +3     # the same, signed
cairn task edit 1 --objective -1     # one fewer, without naming the others
cairn task edit 1 --objective =2     # the set becomes exactly this
cairn task edit 1 --objective ""     # the task is set loose
```

`domain.SplitSign` is the one place either sign means anything, and
`changedNumbers` in `internal/engine/set.go` is the one place a set of numbers
is read — blockers and objectives both go through it, and the words each refuses
with are handed in so a reader is told about the thing they were naming.

## What counts twice, and why that is right

A task serving two objectives counts in both. So work done on it advances both,
and an objective's share is over the tasks that serve it whether or not they
serve anything else.

That is the point rather than a side effect. The alternative — splitting a task
across the objectives it serves, so it counts as half in each — would make an
objective's share depend on work that has nothing to do with it, and there is
no honest fraction to pick.

The same objective named twice is served once. Without that, a task could weigh
twice in one objective's share.

## One that does not exist turns down the whole set

`--objective +99` on a task already serving objective 1 fails, and the task
still serves objective 1 alone. Attaching the ones that exist and dropping the
rest would write a set nobody asked for, and say nothing about the one that went
missing.

A **filter** is the opposite: `task list --objective 9` answers nothing. A filter
says what is wanted, where an attachment writes a link, and a link to nothing is
a fact that was never true. See [listing.md](listing.md).

## A task's steps count towards what it serves

A step serves nothing of its own — it carries no objectives, being no task —
and it counts towards every objective its task serves. `objective show` counts
the steps apart from the tasks and both in the share: a task left open with its
one step ticked reads "0 of 1 tasks and 1 of 1 steps are done", half of the
objective (`@objective-progress-counts-steps`). Counting the task alone would
leave an objective at nothing while the work under it was half done; counting
its steps as tasks would say there were more tasks than anybody wrote.
The boards' objective cards used to do just that — "1/5" over two
tasks and three steps, beside a list of two — and now write them apart the
same way, shorter: "0/1 · 1/2 steps", the share after it counting both
(`@gui-an-objective-card-counts-steps-apart`,
`@tui-an-objective-card-counts-steps-apart`).

It replaces the rule that a sub-task served no objective: a sub-task was a task
that could have been given one, and had to be refused it; a step cannot carry
one at all. See [steps.md](steps.md).

## What was archived still counts

Archiving a task keeps it whole, and what it did for an objective is part of
it: a task archived done is still a task done towards what it served, and one
archived open is still open. Sweeping the done column at the end of a sprint
leaves every objective exactly as far along as it was; `objective show` still
lists them. Counting only what is on the board would make an objective go
backwards every time somebody tidied up.

## The time spent on it, and what is left of it

An objective was written up by hand to be reviewed — "Temps passé sur #1 :
71h40m" in a description — because nothing said it. `objective show`
now says both, from its tasks:

- **time spent** adds the time of every task serving it, archived ones
  included, the way progress counts them: the time was spent whatever became of
  the card. It is the same time `task show` says of each, read from the trace
  (`@objective-says-the-time-spent-on-it`).
- **left** counts the tasks neither done nor archived, and names the soonest
  deadline among them with its task — `2 tasks, the next due 2026-11-05 (#3)`
  (`@objective-says-what-is-left`); an objective whose every task is done has
  `nothing` left (`@objective-with-nothing-left-says-so`). An archived task that
  was open is not counted there: archiving it took it out of the work.

A script reads them as `time_spent` and `left`, the numbers of those tasks
(`@objective-json-says-the-time-and-what-is-left`). `objective list` says how far
along each objective is, `· 50% done`, and carries its `progress` in JSON; an
objective with nothing to count says nothing of it
(`@objective-list-says-how-far-along`).

## A review over a period

`objective report N --since DAY [--until DAY]` writes what a review of an
objective needs, over the days asked — the probation objectives are reviewed
on a date, and the time was being added up by hand for it:

- how far along it is, its deadline, and what is left, as `show` says them;
- the time spent on its tasks in the period and in all, then **month by
  month** — the timesheet's days, those of its tasks only, folded into months
  (`@objective-report-says-the-time-month-by-month`);
- the tasks **put in done in the period** and still done, each with the day it
  was, read from the trace's column changes
  (`@objective-report-names-what-was-done-in-the-period`,
  `@objective-report-before-the-work-names-nothing-done`).

The first day is asked for, never guessed (`@objective-report-needs-the-first-day`);
the last is today when left out. A script reads it as `objective`, `since`,
`until`, `time_spent`, `months` and `done`
(`@objective-report-reads-by-a-machine`).

## Where it lives, and what the migration does

In a table of its own, `task_objectives`, the same shape as the labels and the
blockers. The column on `tasks` is gone.

The migration carries the data rather than only changing the shape: every
`tasks.objective` that named something becomes a row in the new table before the
column is dropped. Without that, applying it would silently detach every task in
a store that had any.

**Rolling it back keeps the smallest number** and says so: `loses: every
objective a task served beyond the first`. Losing them all would break the tasks
that were fine before — most of them serve one — and picking "the first named"
is not a thing the store knows, since a set has no order but its numbers.

## Deliberately not there yet

- **Weighting.** "This task is mostly for objective 1" — there is no honest
  number to put on that, and nothing has asked for one.
- **An objective serving an objective.** Nothing puts an objective under
  another, so nothing narrows or attaches that way either. Decided, not built:
  see [two levels](#two-levels--decided-built-in-226).
- **Deleting an objective that has tasks.** Still refused, so nothing here
  releases links to an objective that is gone: there can never be one. A
  function for it was written and taken back out, because unreachable code is
  code nothing can vouch for. Deleting a parent that asks what goes with it is
  decided, not built: see [two levels](#two-levels--decided-built-in-226).

## Two levels — decided, built in #226

Decided on 2026-11-09 in #222, after a study of how other tools organise work
and a mockup drawn over the rule; #226 builds it. **None of it is built yet.**
Today an objective sits under nothing, and what follows is what cairn will do,
not what it does: this page does not describe it as what cairn does until it
is built, and each part moves into the sections above as it is.

**Why a second level.** Home automation is an enclosure, a firmware and some
features that wait on one another and have to be seen together; a probation
objective is reviewed through its key results. Flat objectives held neither:
the parts were one objective that hid them, or three that did not know they
were one. A project beside the objective, as Linear and Asana have, was
weighed and left out: three notions to learn, and a task serving an objective
and a project counted twice. So was a project per store: the parts would no
longer be seen together, wait on each other or add up. A store is for what
must never mix — a workspace, see [config.md](config.md) — and what may ever
be linked shares one.

**Two levels, and the names kept.** An objective may hold sub-objectives; a
sub-objective holds none, and a third level is refused. Two cover every case
asked and draw on a board; the personal tools nest freely, GitHub stops at
eight, and a third need is a task with steps. Two can grow into more without
breaking anything, the day they prove short. The names stay Objective ›
Sub-objective: Area › Project read oddly for a probation objective, which
ends, and Goal › Key result for an enclosure, which is no result. Keeping the
word keeps every command, every tool and the trace as they are.

**What kind of thing it is, as a label.** An area, a project, a key result: a
type is a label the objective carries, as objectives already carry labels. A
label gains a scope — tasks, objectives or both — so an objective's form offers
only the labels meant for objectives. A fixed set of types would be a second
list beside the labels, with its own colours, rename and delete; nothing exists
twice. A label nobody scoped is for both, so every label already in a store is
offered where it was.

**What a task serves, and how it counts.** A task serves objectives or
sub-objectives, as many as it serves. On a sub-objective it counts in the
parent too, and that link is worked out, never stored. A parent's time spent,
progress and what is left are read over the union of its own tasks and its
sub-objectives', each task once: adding the sub-objectives up would count a
task serving two of them twice. Across two trees the rule above holds — it
counts in each.

**Attached twice is refused, and only refused.** A task attached to a parent
and to one of its sub-objectives at once is refused —
`#44 is already counted in #4 Home automation through #5 Enclosure` — whichever
is ticked second, and nothing changes. The link would add nothing the parent
does not already count, and moving the task up in the same act would write a
set nobody asked for, as [one that does not
exist](#one-that-does-not-exist-turns-down-the-whole-set) does not drop the
rest. Two sub-objectives of one parent, or two trees, stay allowed.

**A sub-objective moves, unless it would be attached twice.** Changing its
parent moves it with its tasks, and both trees count again. It is refused when
one of its tasks is already attached to the new parent: the move would make the
double attachment the rule refuses. A parent is not given a parent of its own:
that would be a third level.

**A dependency warns, never blocks.** An objective or a sub-objective may say
it waits on another — "Firmware waits on Enclosure" — as a link stored between
the two. Read off the tasks' blockers, it could not be seen at a glance; Linear
puts the same link between projects and warns when it is not kept. It never
shuts a column: an objective is not work but what work is for (see
[blocking.md](blocking.md#deliberately-not-there-yet)), and blocking is added
the day it proves needed.

**Reached is proposed, never put.** A parent is proposed as reached when every
sub-objective is; it is put there by the person. Its standing is theirs — a
parent can hold work of its own beyond its sub-objectives — and no average of
percentages decides it.

**Deleting a parent asks.** It lists the tasks it touches, then asks, for each
sub-objective, whether to delete it or keep it as a level-one objective — all,
none or some — and, for the tasks, whether to delete them or only detach them.
Which work goes with an objective is a deliberate choice, and nothing is decided
for the person.

**On screen.** The window's Objectives tab becomes a tree on the left — each
level-one objective with its sub-objectives indented under it — and the chosen
objective on the right, its numbers over the union; the terminal's tab, rows in
a tree rather than cards. A grid cannot show what sits under what. A card names
the objective it is directly attached to, by name — `◇ Firmware` — whether a
level-one objective or a sub-objective; hovering gives the whole path, which is
too long for a card. "Show its tasks" narrows the board to an objective and its
sub-objectives with the filter that exists. The mockup is the reference: every
screen is checked against it at the end of #226.

**What is listed under a parent.** `task list --objective 4` reaches its
sub-objectives too, and `--itself` keeps to the parent alone. A parent's tasks
are its subtree's, as its numbers count them; a listing that said fewer would
disagree with the count.

**What the migration keeps.** One migration per change: a parent on
objectives, empty for every one already there; a scope on labels, none meaning
both; the links between objectives, none. Every objective already in a store
becomes level one with its number, title, standing, labels, tasks, time and
trace as they were — checked on a copy of this repository's store and of the
fixtures on 2026-11-09. Rolling back loses which parent, which scope and which
dependency: a sub-objective becomes level one with its own tasks, and a parent
counts only the tasks attached to it directly.
