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
- **A third level.** An objective holds sub-objectives and a sub-objective
  holds none; see [two levels](#two-levels).

## Two levels

Decided on 2026-11-09 in #222, after a study of how other tools organise work
and a mockup drawn over the rule; #226 builds it. The model, its migration,
the rules, the command line and the MCP server are built; what is left to
build is [below](#decided-not-built-yet).

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

**Two levels, and the names kept.** An objective may hold sub-objectives —
`objective new "Enclosure" --parent 4`, `objective edit 5 --parent 6`, or
`--parent none` to stand on its own again (`@objective-sub-made-under-a-parent`,
`@objective-sub-moves-to-level-one`). A sub-objective holds none: putting
anything under it, or putting an objective holding sub-objectives under
another, is refused — `an objective goes two levels deep at most`
(`@objective-sub-third-level-refused`, `@objective-sub-parent-stays-level-one`).
Two cover every case asked and draw on a board; the personal tools nest
freely, GitHub stops at eight, and a third need is a task with steps. Two can
grow into more without breaking anything, the day they prove short. The names
stay Objective › Sub-objective: Area › Project read oddly for a probation
objective, which ends, and Goal › Key result for an enclosure, which is no
result. Keeping the word keeps every command, every tool and the trace as they
are. `objective show` says the parent of each, `—` for one standing on its
own; a machine reads `parent`, null at level one, and `children`
(`@objective-sub-json-parent-and-children`).

**What a task serves, and how it counts.** A task serves objectives or
sub-objectives, as many as it serves. On a sub-objective it counts in the
parent too, and that link is worked out, never stored. A parent's time spent,
progress and what is left are read over the union of its own tasks and its
sub-objectives', each task once: adding the sub-objectives up would count a
task serving two of them twice (`@objective-sub-parent-counts-each-task-once`).
Across two trees the rule above holds — it counts in each. `objective show`
on a parent names its sub-objectives with their counts and time, then the
tasks on it directly apart from those in its sub-objectives
(`@objective-sub-show-splits-its-tasks`). `objective list` draws the tree, two
lines an objective — its number and title, then where it stands, its counts
over its subtree and its time — each sub-objective under its parent
(`@objective-sub-list-as-a-tree`); one listed without its parent says which it
sits under. `objective report` on a parent splits each month's time by
sub-objective and the parent itself, `2026-09  2h30m  · #2 2h  · itself 30m`,
a task serving two sub-objectives counting in both columns and once in the
total (`@objective-sub-report-splits-the-time`).

**Attached twice is refused, and only refused.** A task attached to a parent
and to one of its sub-objectives at once is refused —
`#44 is already counted in #4 Home automation through #5 Enclosure and #6 Firmware; nothing was changed`
— whichever is ticked second, and nothing changes
(`@objective-sub-attached-to-a-parent-and-its-sub-objectives-refused`,
`@objective-sub-attached-to-a-parent-then-a-sub-objective-refused`). The link
would add nothing the parent does not already count, and moving the task up in
the same act would write a set nobody asked for, as [one that does not
exist](#one-that-does-not-exist-turns-down-the-whole-set) does not drop the
rest. Two sub-objectives of one parent, or two trees, stay allowed
(`@objective-sub-two-sub-objectives-of-one-parent`). The rule is
`objective.CountedOnce`, asked by every act that changes what a task serves.

**A sub-objective moves, unless it would be attached twice.** Changing its
parent moves it with its tasks, and both trees count again
(`@objective-sub-moves-to-another-parent`). It is refused when one of its tasks
is already attached to the new parent: the move would make the double
attachment the rule refuses (`@objective-sub-move-refused-when-counted-twice`).

**What is listed under a parent.** `task list --objective 4` reaches its
sub-objectives too, and `--itself` keeps to the parent alone
(`@objective-sub-task-list-reaches-the-sub-objectives`); an agent passes
`itself` to `task_list`. A parent's tasks are its subtree's, as its numbers
count them; a listing that said fewer would disagree with the count. The
boards' filter still reaches the objective named alone, until the boards are
built for two levels.

**Reached is proposed, never put.** A parent is proposed as reached when every
sub-objective is: the act that reached the last one says so —
`every sub-objective of #1 Home automation is reached: it may be put reached too`
— and so does `objective show` on the parent; a machine reads
`proposed_reached` (`@objective-sub-reached-proposed-by-the-last-one`). It is
put there by the person. Its standing is theirs — a parent can hold work of
its own beyond its sub-objectives — and no average of percentages decides it.

**Deleting asks.** Deleting an objective — a parent or not — that holds
sub-objectives or carries tasks is refused, the refusal listing every
sub-objective with its tasks and every task it carries, and the flags that say
what goes with them (`@objective-sub-delete-asks`). For the sub-objectives,
`--sub-objectives keep|delete`, all or none, and `--keep 5,7 --delete 6` for
some, which win over the first; for the tasks attached to whatever is deleted,
`--tasks detach|delete` (`@objective-sub-delete-keeps-and-detaches`,
`@objective-sub-delete-takes-everything`, `@objective-sub-delete-one-by-one`).
A kept sub-objective stands on its own with its tasks. A task detached loses
the objectives deleted and keeps any other, its trace saying so; a task deleted
goes the way `task delete` takes it. Naming a number that is no sub-objective
of what is deleted is refused (`@objective-sub-delete-a-sub-objective-it-does-not-hold`).
An agent passes the same as `sub-objectives`, `keep`, `delete` and `tasks`.
Which work goes with an objective is a deliberate choice, and nothing is
decided for the person — which is why an objective with tasks and no
sub-objective, refused outright before, now asks the same way
(`@objective-sub-delete-level-one-detaching-its-tasks`).

**What the migration keeps.** One migration, #18 `an objective can sit under
another`, adds a parent to every objective, empty for every one already there
(`@objective-sub-migration-every-objective-level-one`). Every objective already
in a store becomes level one with its number, title, standing, labels, tasks,
time and trace as they were — checked on a copy of this repository's store on
2026-11-09: before, after and after the rollback, every objective's tasks,
progress, time and every one of its tasks' trace were identical. Rolling back
loses which parent, and says so before it is applied
(`@objective-sub-migration-says-what-rolling-back-loses`): a sub-objective
becomes level one with its own tasks, and a parent counts only the tasks
attached to it directly (`@objective-sub-migration-rollback-sets-every-one-loose`).

**What kind of thing it is, as a label.** An area, a project, a key result: a
type is a label the objective carries, as objectives already carry labels. A
label has a scope — tasks, objectives or both — so what an objective is
given is chosen among the labels meant for objectives (`label list --for
objectives`), and a label for objectives only is refused on a task. `area`
and `project` are the type labels suggested; nothing makes them for the
person. A fixed set of types would be a second list beside the labels, with
its own colours, rename and delete; nothing exists twice. A label nobody
scoped is for both, so every label already in a store is offered where it
was. Migration #19, `a label has a scope`; see
[labels.md](labels.md#a-label-is-for-tasks-for-objectives-or-for-both).

**A dependency warns, never blocks.** An objective or a sub-objective may say
it waits on another — `objective edit 7 --waits-on 6`, Features waits on
Firmware — as a link stored between the two, with the grammar of `--blocker`:
`+6` or `6` adds one, `-6` takes one away, `=6` says what the set becomes
(`@objective-waits-on-taken-off`). Read off the tasks' blockers, it could not
be seen at a glance; Linear puts the same link between projects and warns when
it is not kept. `objective show` says it on both: `waits on  #6 Firmware`, and
`#7 Features waits on it` under the other (`@objective-waits-on-is-said-on-both`);
a machine reads `waits_on` and `waited_on_by` (`@objective-waits-on-json`).
Levels do not matter: a level-one objective waits on a sub-objective of another
tree as well (`@objective-waits-on-any-level`).

It never shuts a column: an objective is not work but what work is for (see
[blocking.md](blocking.md#deliberately-not-there-yet)), and blocking is added
the day it proves needed. A task put in `current` — moved, edited, focused or
made there — while an objective it counts in waits on one not reached is put
there, and the answer warns, `#7 Features waits on #6 Firmware, not reached`
(`@objective-waits-on-warns-when-work-starts`,
`@objective-waits-on-warns-on-edit-and-focus`); a machine reads `warnings`.
Once the other is reached nothing is said
(`@objective-waits-on-silent-once-reached`). An objective waiting on itself,
or on one that already waits on it however far through others, is refused —
`#3 Features already waits on #1 Enclosure, through #2 Firmware` — since
neither could ever be reached first (`@objective-waits-on-itself-refused`,
`@objective-waits-on-in-a-circle-refused`); one that is not there is refused
too (`@objective-waits-on-nothing-there`). Deleting an objective takes its
dependencies with it, both ways (`@objective-waits-on-goes-with-the-deleted`).
Migration #20, `an objective can wait on another`, adds a table of its own,
`objective_waits`; rolling it back loses which waits on which and nothing
else (`@objective-waits-on-migration-rollback`).

**A task file carries the tree.** `task export` writes the objectives its
tasks serve, the parents of those and what they wait on, each with its
`parent` and its `waits_on`; `task import` brings them in under the same
parents, waiting on the same objectives, at their new numbers
(`@exchange-objectives-come-with-their-tree-and-what-they-wait-on`). A file
whose objective sits under or waits on one it does not hold is refused before
anything is added (`@exchange-a-parent-outside-the-file`,
`@exchange-waiting-outside-the-file`). The file went to version 5: an older
cairn would have read it and brought every objective in level one, waiting
on nothing, without a word. `cairn fixtures` holds such a tree, so a board
always has one to draw.

**What is not traced.** The trace is what was done to a task, and an
objective has none of its own: setting or changing a parent writes nothing
there, as no change to an objective ever has. What a deletion does to a task —
detached, deleted — is traced on the task, as any change to it is.

### Decided, not built yet

**On screen.** The window's Objectives tab becomes a tree on the left — each
level-one objective with its sub-objectives indented under it — and the chosen
objective on the right, its numbers over the union; the terminal's tab, rows in
a tree rather than cards. A grid cannot show what sits under what. A card names
the objective it is directly attached to, by name — `◇ Firmware` — whether a
level-one objective or a sub-objective; hovering gives the whole path, which is
too long for a card. "Show its tasks" narrows the board to an objective and its
sub-objectives with the filter that exists. The mockup is the reference: every
screen is checked against it at the end of #226. Until then the boards count
over the union like the command line, and draw the objectives flat.
