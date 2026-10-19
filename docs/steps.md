# Steps

Read this before touching a task's steps — how one is added, ticked or taken
away, how they cross to a machine, or what a task is split into.

A step is a smaller piece of a task: a title, and whether it is done, in its
place among the task's other steps. It is **not a task**. It has no number on
the board, no column, no focus, no time, no notes, no labels, no objectives and
no blockers (`@task-step-has-no-number`); all of that stays with the task it
belongs to. What it is for is splitting a task into pieces each done or not,
added, read, renamed and taken away where the task is edited.

## Why steps replaced sub-tasks

A task used to be able to hang under another, as a sub-task. A sub-task was a
whole task — a number, a column, a focus, time, notes, a trace — for what was
most often one line to tick. Measured on a store of 219 tasks, its 49
sub-tasks were done 47 times and todo twice, with no deadline, no label, no
priority and no blocker among them.

What that whole task cost was paid everywhere. A parent and an order among
sisters ran through every rule and every surface: a position counted under
each parent as well as in each column, a move refused behind a task that was
not a sister, a sub-task that could not serve an objective and had to be
refused one, a sweep that counted sub-tasks while another did not, a search
that answered the sub-task on the command line and its parent on the boards.
And the boards never showed sub-tasks as cards anyway: they were lines of their
parent's form already. Decided with the owner on 2026-10-16 (cairn#91): what
the usage had settled on became the model, and the rest went.

What was given up, said plainly: a piece of a task can no longer be focused,
timed, blocked, labelled or given notes of its own. A piece that needs any of
that is a task, and is made one.

## A step is named by its place

A step has no number, so it is named by its place in its task, the first being
1. `step new` puts one behind those already there (`@step-new`,
`@step-new-goes-last`), done if asked (`@step-new-done`); `step edit` ticks it,
unticks it or renames it (`@step-edit-done`, `@step-edit-not-done`,
`@step-edit-title`); `step delete` takes it away, and the ones behind move up a
place (`@step-delete`). A place a task has no step at is refused, "the task has
no such step" (`@step-unknown`, refusal `no-such-step`); a title with nothing in
it is refused, "a step needs a title" (`@step-new-needs-a-title`); `--done` and
`--not-done` together are refused, "a step is done or not done, not both"
(`@step-edit-both-ways`).

A place moves when a step before it goes, which is what a place is. Naming a
step by a number of its own was weighed and left out: it would be a second run
of numbers beside the tasks', for things nobody refers to from elsewhere.

Steps keep the order they were written in (`@ordering-steps-keep-their-order`),
and a task moved carries them with it (`@ordering-a-task-carries-its-steps`).
They are kept with their task: they outlive the command
(`@task-steps-survive`) and are deleted with it
(`@task-delete-takes-its-steps`).

## Written with the task

`task new` and `task edit` take `--steps`, one line a step — `[x] its title`
done, `[ ] its title` left — and a leading `#n` is accepted and read as
nothing more than a line (`@task-new-with-its-steps`). On an edit they
**replace** the steps the task had (`@task-edit-with-its-steps`), and given
empty the task keeps none (`@task-edit-with-no-steps`). That is the boards'
forms too: a form shows the whole list, so saving it says the whole of it, in
the act that saves the task. See [ordering.md](ordering.md#a-tasks-steps-are-written-with-it).
Two acts on a board do not wait for the form: a tick, and a delete. A step
chosen and deleted with a delete key goes at once, once agreed, through a
`step delete` of its own, the way the command line deletes one — in the
window's form, and on the terminal's open task (cairn#105).

## Done with a step still open

A task put or moved into done while one of its steps is open is refused unless
it is confirmed — "a step of it is not done", refusal `step-still-open`, naming
the places of the steps in the way (`@task-done-refused-while-a-step-is-open`,
`@task-done-with-open-steps-says-how-to-confirm`). Confirmed with `--yes`, its
steps are ticked with it in the same act (`task.Task.Ticked`;
`@task-done-confirmed-ticks-its-steps`,
`@task-moved-to-done-confirmed-ticks-its-steps`). A task with no step open
finishes as before (`@task-done-allowed-once-every-step-is-done`,
`@task-done-with-no-steps`).

The boards ask rather than refuse: the terminal "put task N in done, ticking
its steps?" (`@tui-done-asks-about-open-steps`), the window "Put task N in
done, ticking its steps?" with "Its steps not yet done are ticked with it."
(`@gui-a-drop-in-done-asks-about-open-steps`). The MCP server confirms by
itself, as it does every act. Why the rule is a refusal and not a cascade is
in [ordering.md](ordering.md#a-task-is-not-done-while-one-of-its-steps-is-open).

## How they cross

- **The command line**: the `step` group — `step new <task> "<title>"
  [--done]`, `step list <task>`, `step edit <task> <place> [--title …] [--done |
  --not-done]`, `step delete <task> <place>`. `task show` writes a `steps`
  block, each line the `(done)` mark or a blank, the place and the title
  (`@task-show-names-its-steps`); `task list` leaves steps out
  (`@task-list-leaves-steps-out`).
- **JSON**: the `step` commands answer `{"task": n, "steps": [{"place",
  "title", "done"}]}` (`@step-list-json`). A task's shape carries `"steps":
  [{"title", "done"}]` (`@task-show-json-carries-its-steps`), and its progress
  is `progress.steps {total, done}`, read as "1 of 2 steps are done".
- **MCP**: `step_new`, `step_list`, `step_edit` and `step_delete`, generated
  from the engine's operations like every other tool. See [mcp.md](mcp.md).
- **Search**: a step's title is part of what its task says (`task.Said`), so a
  word found only in a step reaches the task (`@task-search-reaches-a-step`).
  See [listing.md](listing.md).
- **Objectives**: a task's steps count towards what it serves, apart from the
  tasks (`@objective-progress-counts-steps`). See
  [objectives.md](objectives.md).
- **Export and import**: the export file is version 2, each task carrying its
  `steps` and none a `parent`. Import still reads a version 1 file, and its
  sub-tasks come in as steps of their parent, those left first, then those done
  (`@exchange-import-reads-sub-tasks-as-steps`).
- **The boards**: neither draws a step as a card
  (`@tui-a-step-is-not-on-the-board`, `@gui-a-step-is-not-on-the-board`). See
  [tui.md](tui.md) and [gui.md](gui.md).

## The trace

A step has no trace of its own. A change of steps is a *changed* act on its
task, with the field `steps`, the values written as the whole list,
`[x] a; [ ] b` (`@step-traced`). See [trace.md](trace.md).

## What the migration did

The thirteenth change to a store's shape, "a task is split into steps, and no
task hangs under another", turned every sub-task into a step of its parent:
done when it was done, those left first and then those done, each in its old
order (`@migration-sub-tasks-become-steps`). A sub-task that held the focus let
go of it, with a *focus released* act written; each sub-task's trace gained a
*deleted* act, "became a step of task N"; the sub-tasks then went, with their
labels, blockers, objectives and focus, and the `parent` column was dropped
(`@migration-sub-tasks-leave-the-board`). Their past acts, and the time they
carry, stay in the trace under the old number, like those of any deleted task.

**Rolling it back loses every step.** It puts the `parent` column back and drops
the `steps` table, and it says so: the sub-tasks the steps were made from are
not brought back. Bringing them back would mean inventing numbers, columns and
places for things that had become lines, and a backup is set aside before
either direction anyway.
