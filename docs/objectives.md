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
cairn task add "Write the spec" --objective 1 --objective 2
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

## The sub-task rule is untouched

A sub-task still serves no objective, and a set does not soften that. It belongs
to the task above it and is counted towards that task's objectives through it —
an objective it served alone would be one its parent might have nothing to do
with.

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
