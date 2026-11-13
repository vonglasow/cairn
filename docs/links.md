# Links

Two tasks can be linked. A link has no kind and no direction: it is the same
link read from either task, and `task show` names it on both.

```sh
cairn task new "It does not hold" --link 105 --label bug
cairn task edit 120 --link +107 --link -105
cairn task edit 120 --link =107          # the links become exactly these
```

A plain number or `+3` adds a link, `-3` takes one away, and `=3` says what the
set becomes; an `=` is not mixed with a sign. Linking two tasks already linked
changes nothing. A task cannot be linked to itself or to a task that does not
exist. Any column will do, `done` and archived included.

## A task done is not reopened

When the work of a task done turns out not to hold, the task stays done: open a
new task for the defect, linked to it, and put the `bug` label on it. Reopening
it would take it out of the period its objective counted it finished in, mix
the time of the fix with the time of the feature, and rewrite what its notes
already say.

Moving a task out of `done` is still possible, for a task put there by mistake.

## Opening a linked task

Every surface opens a new task already linked to the one you are on, in `todo`
whatever that one's column:

- the command line: `cairn task new "…" --link 105`;
- the terminal board: `o` on a card, in insert mode, opens the form with the
  card already in `links`;
- the window: `o` on a chosen card, or **New linked task** in its right-click
  menu;
- an agent: `task_new` with `link`.

Nothing else is carried over: labels and objectives are put by hand. Both
forms have a `links` field, so a link is put or taken away there like any
other field.

## What a link does not do

- It has no kind — "duplicates", "found in", "follows". The labels already say
  which task is the defect.
- A search does not reach links, and no listing narrows by them.
- The boards show a task's links in its form only, and draw no line between
  two cards.
- An objective has no links.

Deleting a task takes its links away from the other tasks. A link is written in
the trace of both tasks. An export carries the links on both sides, and an
import links the tasks again at their new numbers.
