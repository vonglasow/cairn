# Objectives

An objective is what work is for. Tasks serve it, and it counts how far along
they are, how long they took, and what is left.

```sh
cairn objective new "Ship cairn 1.0" --priority high --deadline 2026-12-31
cairn objective list                     # the tree, each with how far along it is
cairn objective list --state on-hold --label urgent
cairn objective show 1                   # what it carries, what is under it, how far along
cairn objective edit 1 --state reached --label +done
cairn objective delete 1 --yes
```

An objective takes the options a task takes, minus those that mean nothing to
it: no column, no steps, no time of its own. Its `--state` is `active`,
`on-hold`, `reached` or `dropped`; a new one is `active`.

## What a task serves

A task serves as many objectives as it actually serves, with the grammar of
labels:

```sh
cairn task new "Write the spec" --objective 1 --objective 2
cairn task edit 1 --objective +3     # one more, without naming the others
cairn task edit 1 --objective -1     # one fewer
cairn task edit 1 --objective =2     # the set becomes exactly this
cairn task edit 1 --objective ""     # the task is set loose
```

A task serving two objectives counts whole in both. Naming one that does not
exist refuses the whole change: the task keeps what it served.

## How far along

`objective show` works it out from what serves it:

```
progress  1 of 2 tasks and 0 of 1 steps are done — the objective is 33% done
time spent  4h20m
left      2 tasks, the next due 2026-11-05 (#3)
```

- Tasks and their steps are counted apart, and the share is over all of them.
  The share never reads `100%` while something is left, nor `0%` once something
  is done.
- **Archived tasks still count**: a task archived done is still done, so
  sweeping a column leaves every objective where it was.
- **Time spent** adds the time of every task serving it, archived ones
  included.
- **Left** counts the tasks neither done nor archived, with the soonest deadline
  among them.

`objective list` says how far along each one is.

## A review over a period

```sh
cairn objective report 1 --since 2026-09-01 --until 2026-11-30
```

It says how far along the objective is, the time spent on it in the period and
in all — month by month — and the tasks put in done in the period, each with its
day. The first day is required; the last is today when left out.

## Two levels

An objective may hold sub-objectives, two levels deep and no more:

```sh
cairn objective new "Enclosure" --parent 4      # a sub-objective of 4
cairn objective edit 5 --parent 6               # moved under another
cairn objective edit 5 --parent none            # standing on its own again
cairn objective move 5 --first                  # its place among its siblings
```

- A parent counts its own tasks and its sub-objectives', **each task once**.
- `objective list` draws the tree; `objective show` on a parent names its
  sub-objectives with their counts, then the tasks on it directly.
- `objective report` on a parent splits each month's time by sub-objective.
- `task list --objective 4` includes the sub-objectives' tasks; `--itself`
  keeps to the parent alone.
- A task attached both to a parent and to one of its sub-objectives is refused,
  and nothing changes. Two sub-objectives of one parent are fine.
- A sub-objective cannot move under a parent one of its tasks is already
  attached to.
- When every sub-objective is reached, cairn says the parent may be put reached.
  It never does so itself.
- Objectives keep the order you give them, a parent moving with its
  sub-objectives.

What kind of thing an objective is — an area, a project, a key result — is a
label scoped to objectives:

```sh
cairn label new area --scope objectives
cairn objective edit 4 --label +area
```

## One objective waiting on another

An objective can say it waits on another — Features waits on Firmware:

```sh
cairn objective edit 7 --waits-on 6     # +6, -6 and =6 as for labels
```

`objective show` says it on both sides. It never blocks anything: putting a
task in `current` while an objective it counts in waits on one not reached
goes through, with a warning — `#7 Features waits on #6 Firmware, not reached`.
Once the other is reached, nothing is said. Waiting on itself, or in a circle,
is refused.

## Deleting an objective

Deleting an objective that holds sub-objectives or carries tasks is refused
until you say what goes with it; the refusal lists them.

```sh
cairn objective delete 4 --sub-objectives keep --tasks detach --yes
cairn objective delete 4 --keep 5,7 --delete 6 --tasks delete --yes
cairn objective delete 4 --views off --yes      # take it off the views naming it
```

A kept sub-objective stands on its own with its tasks. A task detached keeps
any other objective it serves.

## On the boards

Both boards have an Objectives tab: a tree of objectives beside the chosen
one's detail — its numbers, its sub-objectives, its tasks. A card on the task
board names the objective it serves, `◇ Firmware`. From an objective, `t` shows
its tasks on the board. See [the board](tui.md#objectives) and
[the window](gui.md#objectives).

An occurrence of a [series](series.md) serves the objectives its template
serves. An export carries every objective with its parent and what it waits
on.
