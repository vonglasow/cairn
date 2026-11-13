# Steps

A task is split into steps: a title each, done or not, in their order. A step
is not a task. It has no number on the board, no column, no focus, no time, no
notes, no labels and no objectives — all of that stays with its task. A piece
of work that needs any of it is a task of its own.

## Adding, ticking, taking away

A step is named by its place in its task, the first being 1:

```sh
cairn step new 1 "Write the spec"           # behind the steps already there
cairn step new 1 "Plan it" --done           # already done
cairn step list 1                           # each with its place, done or not
cairn step edit 1 2 --done                  # ticked; --not-done unticks it
cairn step edit 1 2 --title "Plan the week" # renamed
cairn step delete 1 2                       # the ones behind move up a place
```

A place moves when a step before it goes. A place the task has no step at, a
step with no title, and `--done` with `--not-done` are refused.

## Written with the task

`task new` and `task edit` take every step at once with `--steps`, one line a
step — `[ ] its title` to do, `[x] its title` done — and `-` reads them from
standard input:

```sh
cairn task new "Ship it" --steps - < steps.txt
cairn task edit 1 --steps ""                 # none left
```

On an edit, the list given **replaces** the steps the task had. The boards'
forms work the same way: a form shows the whole list, and saving it says the
whole of it. Ticking a step, or deleting one, happens at once without saving
the form.

## A task is not done while a step is open

Putting a task in `done` while one of its steps is open is refused, naming the
steps in the way. `--yes` on `task edit` or `task move` puts it there and ticks
them in the same act. Both boards ask instead of refusing — "put task 4 in
done, ticking its steps?" — and agreeing does the same thing.

## Where steps show

- `task show` names the steps, each with its place, and says how many are done.
  `task list` leaves them out.
- A word found only in a step finds its task in a search.
- A task's steps count towards the objectives it serves, apart from the tasks:
  "1 of 2 tasks and 3 of 4 steps are done".
- Neither board draws a step as a card; they are lines of their task's form and
  detail.
- A task moved, exported or deleted takes its steps with it.
- Over MCP: `step_new`, `step_list`, `step_edit` and `step_delete`.

A change of steps is written in the task's trace as a change of the field
`steps`, with the whole list before and after.
