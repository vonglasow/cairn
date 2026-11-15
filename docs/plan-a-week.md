# Plan a week

Choose the week's tasks, put them in the order you will work them, and see on
Friday how long they took.

## Choose the week's tasks

Your week is the tasks in `todo`. Look at what waits in the backlog and move in
what you will do.

```sh
cairn task list --state backlog          # what waits
cairn task move 4 --state todo           # chosen for the week
```

On the board: drag the card into `todo` in the window, or press `m` then `t`
in the terminal.

## Put them in order

The order of a column is yours: cairn never changes it. Put first what you
start with.

```sh
cairn task move 7 --state todo --first   # the one to start with
cairn task move 9 --after 7              # right behind it
```

## Say when, and how long

A deadline is the day a task is wanted by. An estimate is how long it should
take. The card shows both, and the time spent beside the estimate.

```sh
cairn task edit 7 --deadline 2026-11-20 --estimate 3h
cairn task list --deadline-before 2026-11-20   # what is wanted by Friday
```

## Work, and let the time run

Focus a task when you start it. It moves to `current` and its time runs until
you unfocus it or put it `done`.

```sh
cairn task focus 7                       # the time starts
cairn task edit 7 --state done           # and stops
```

## Look back

The Time tab shows the week, day by day. At the command line:

```sh
cairn timesheet show --since 2026-11-16 --until 2026-11-20
```

> **Every Monday the same task?** Make it a [series](series.md). It comes back
> on its day, in the column you choose.
>
> ```sh
> cairn config set series.on true
> cairn task new "Plan the week" --every "monday 9:00"
> ```

Read next: [columns and order](ordering.md), [series](series.md),
[the trace and the time](trace.md), [follow an objective](follow-an-objective.md).
