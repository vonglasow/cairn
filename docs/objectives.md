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
cairn task edit 1 --objective +3     # one more, without naming the others
cairn task edit 1 --objective -1     # one fewer, without naming the others
cairn task edit 1 --objective 2      # the set becomes exactly this
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
The boards' objective cards did just that until cairn#120 — "1/5" over two
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
71h40m" in a description — because nothing said it (cairn#54). `objective show`
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
  another, so nothing narrows or attaches that way either.
- **Deleting an objective that has tasks.** Still refused, so nothing here
  releases links to an objective that is gone: there can never be one. A
  function for it was written and taken back out, because unreachable code is
  code nothing can vouch for.
