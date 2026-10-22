# Links

Two tasks can be linked. A link has no kind and no direction: it is the same
link read from either task, and `task show` says it on both.

```sh
cairn task new "It does not hold" --link 105 --label bug
cairn task edit 120 --link +107 --link -105
```

## A task done is not reopened

A task done whose work turns out not to hold stays done. A new task is opened
for the defect, linked to it. Reopening one took it out of the period its
objective counted it finished in, mixed the time of the fix with the time of
the feature, gave the defect the feature's priority, and rewrote a story its
notes already told.

Moving a task out of `done` stays possible, with no reason asked: it is for a
task put there by mistake, its work never committed.

## Opening one from a task

Every surface opens a task already linked:

- the command line and the MCP: `task new "…" --link 105`, `task_new` with
  `link` (`@mcp-a-task-opened-with-a-link`);
- the board in the terminal: `o` on a card, in insert mode, opens the form of a
  new task in `todo` with `links` already holding the card
  (`@tui-o-opens-a-linked-task`);
- the window: `o` on a chosen card, or New linked task in its right-click menu,
  right after Edit (`@gui-o-opens-a-linked-task`,
  `@gui-a-cards-menu-opens-a-linked-task`).

It opens in `todo` whatever the card's column: what comes of a task done is
work to pick up, not a task done. Both forms carry a `links` field, under the
blockers in the terminal and first in the LINKS section of the window, so a
link is put or taken away there as any field is (`@tui-the-form-links-a-task`,
`@gui-the-form-links-a-task`). The cairn skill tells an agent the same
(`@plugin-the-skill-says-open-a-linked-task`).

## No kind, no direction

A link does not say "found in", "duplicates" or "follows". The task's labels
already say which of the two is the defect, and a kind would say it twice — or,
worse, disagree. Nothing is carried over when a linked task is opened: labels
and objectives are put by hand, because what was wanted automatic is the link,
not the rest.

## What it is held as

`--link` follows the grammar of `--blocker`: a plain number says what the set
becomes, `+3` adds one, `-3` takes one away, and the two are not mixed. The set
named plainly replaces the task's links on the other tasks too
(`@task-link-replaced`).

- One row per pair, the lower number first, so a pair is held once whichever
  side put it (`links` in `internal/migration/known.go`). Linking two tasks
  already linked changes nothing and writes nothing (`@task-link-twice`).
- A task cannot be linked to itself, nor to a task that is not there
  (`@task-link-itself`, `@task-link-nothing`). Nothing is refused for its
  column: a task done or archived is linked like any other
  (`@task-link-any-column`).
- Deleting a task takes its links away from the others
  (`@task-link-gone-with-the-task`).
- A link is written in the trace of both tasks — the one act cairn writes on a
  task nobody named. See [trace.md](trace.md#the-act-is-traced-not-its-consequences).
- An export carries `linked` on both tasks; an import adds them once every task
  of the file exists, at its new number, and refuses a file whose task is
  linked to one it does not hold (`@exchange-a-link-comes-with-its-tasks`,
  `@exchange-a-link-outside-the-file`). The file's version went to 3: an older
  cairn would have read it and dropped its links without a word.

## Deliberately not there yet

- A search does not reach links, and no listing is narrowed by them.
- The boards show a task's links in its form only: neither detail shows them,
  as neither shows its blockers, and no line is drawn between two cards.
- An objective is not linked.
