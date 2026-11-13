# The order, and what you are working on

## One list, in seven columns

A task sits in one of seven columns, read in this order:

| column | what it holds |
|---|---|
| `backlog` | what is not decided yet |
| `todo` | what is decided, in the order to do it |
| `current` | what is under way |
| `blocked` | waiting on somebody else, or on another task — see [blocking](blocking.md) |
| `watching` | handed to somebody, and needing a look |
| `paused` | waiting on your own decision to pick it up |
| `done` | done |

A new task starts in `backlog`. Inside a column, the order is the one you set.
A task's number says when it arrived and nothing else; numbers are never
handed out twice.

## Setting the order

```sh
cairn task move 3 --first        # to the head of its column
cairn task move 3 --after 1      # directly behind another
cairn task move 3 --last
cairn task move 3 --state todo --after 5   # into todo, behind 5, in one act
```

- A new task takes the last place in its column.
- A task given another column lands at the end of it.
- Going behind a task in another column takes it into that column, and the
  answer says so: `task 3 moved from todo to current`.
- Naming a column and a task that is in another one is refused.
- Going into `done` with a step still open asks for `--yes`, as an edit does;
  see [steps](steps.md#a-task-is-not-done-while-a-step-is-open).
- Changing anything else a task carries never moves it.

On both boards, a card moved or dropped somewhere is a move: where it lands is
its new place.

A place is never shown as a number. The order is what there is to see.

## What you are working on

`task focus` says what you are working on now. Several tasks can be in focus
at once — one agent session per task, say:

```sh
cairn task focus 3               # this is what I am working on
cairn task focus                 # which ones are those?
cairn task unfocus 3             # done with that one
cairn task unfocus               # done with all of them
```

- Focusing a task in `backlog`, `todo` or `paused` moves it to the head of
  `current`. One in `blocked` or `watching` stays where it is.
- Putting a task in `done` lets go of its focus; focusing a task already done
  is refused.
- Moving a task into `blocked` lets go of its focus.
- Both boards draw the tasks in focus first in `current`. That is drawn only:
  the order kept does not change.
- A focus is what the [timesheet](trace.md#how-long-the-work-took) counts.

A focus that has run for a long time asks to be renewed; see
[configuration](config.md#focus).

## Choosing what to do next

cairn ranks nothing and proposes nothing. It gives on each task the facts that
help choose:

| fact | what it says |
|---|---|
| `blocks` | how many open tasks wait on this one, `⊳3` on every surface |
| `overdue`, `due_in` | how far past or ahead of its deadline |
| `age` | how many days it has sat in its column |
| `objective_late` | whether an objective it serves is behind |

They are fields of every row of `task list` and `task search`; see
[listings](listing.md#what-a-row-carries). Nothing is sorted or moved for them.

Asked "what do I work on?", the cairn skill of the
[Claude Code plugin](claude-code.md) reads those facts and the board's order,
asks how the day looks, and proposes a few tasks with a reason each. Any move it
proposes is yours to accept.
