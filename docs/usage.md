# Using cairn

Everything cairn does, by what you do with it. `cairn <thing> <verb> --help`
says the rest of any command, with examples, and the reference on this site
lists every command and every flag.

## The first five minutes

```sh
cairn task new "Write the spec"                # a task, numbered 1, in the backlog
cairn task edit 1 --state todo --priority high # what it carries changes by name
cairn task focus 1                             # what you are working on now
cairn task list                                # the board, column by column
cairn tui                                      # the same board, in the terminal
cairn gui                                      # the same board, in a window
```

Three ideas carry the rest. **A task stands in a column** — `backlog`, `todo`,
`current`, `done`, and three for work that is stopped — and inside a column
the order is the one you set. **An objective is what work is for**: a task
serves it, and it counts how far along its tasks are. **Nothing done is
forgotten**: every change is written down, so `cairn trace show 1` says what
happened to a task and `cairn timesheet show` how long the day's work took.

## The guides

| to | read |
|---|---|
| set the order, and say what you are working on | [The order](ordering.md) |
| split a task into steps | [Steps](steps.md) |
| say what stops a task, and start it again | [Blocking](blocking.md) |
| link a task to another, open one for a defect | [Links](links.md) |
| put labels on tasks, colour them | [Labels](labels.md) |
| narrow a listing, search, keep a view | [Listing and views](listing.md) |
| say what work is for, and how far along it is | [Objectives](objectives.md) |
| have a task come back every week | [Series](series.md) |
| read what happened, and how long it took | [The trace and the time](trace.md) |
| keep notes beside a task | [A task's notes](context.md) |
| work on the board in a terminal | [The board in the terminal](tui.md) |
| work on the board in a window | [The board in a window](gui.md) |
| let an agent work on your board | [The MCP server](mcp.md) |
| work with Claude Code | [With Claude Code](claude-code.md) |
| keep several boards, change a setting | [Configuration](config.md) |

## Everything a task carries

```sh
cairn task new "Write the spec" --desc "the one that unblocks the rest" --state todo --priority high --label spec --label team --deadline 2026-09-01 --estimate 90m
cairn task edit 1 --title "Write the other spec" --desc "" --state current --priority critical --deadline 2026-09-15 --estimate 2h
cairn task edit 1 --label +urgent --label -spec
cairn task edit 5 --state blocked --blocker 3 --because "the client's answer"
cairn task edit 5 --blocker -3   # the last thing out of its way puts it in todo
cairn task new "It does not hold" --link 4 --label bug   # task 4 says it too
cairn task edit 6 --link +2 --link -4
cairn task new "Read the spec" --label spec,team
cairn task new "Ship it" --steps - < steps.txt   # one line a step: [ ] to do, [x] done
cairn task new "Write the docs" --objective 1
cairn task new "Write the spec" --objective 1 --objective 2   # it serves both
cairn task edit 1 --objective +3   # one more, without naming the others
cairn task list      # each row says its column, priority, deadline and labels
                     # a task's steps are not listed: task show names them
cairn task list --state todo --priority high   # only what carries all of it
cairn task list --label team --label urgent    # both labels, not either
cairn task list --objective 1                  # only what serves it
cairn task list --view Home                    # what a view keeps
cairn task search spec    # looks in the title, description, steps, deadline,
                          # estimate, time spent — not in what --state and friends narrow by
cairn task show 1
cairn task delete 1 --yes
cairn task delete 1 --also 4 --yes   # several in one act: all of them, or none
cairn task move 3 --first        # to the head of the ones it sits with
cairn task move 3 --after 1      # directly behind another
cairn task move 3 --last
cairn task focus 3               # this is what I am working on
cairn task focus                 # which ones are those again?
cairn task unfocus 3             # done with that one
cairn task unfocus               # done with all of them
cairn task archive 1             # archived, kept whole
cairn task archive --state done --yes  # the whole column in one go
cairn task label sprint-12 --state done  # mark a sprint before sweeping it
cairn task list --archived       # only what was taken off
cairn task unarchive 1           # back on the board
```

| option | what it takes |
|---|---|
| `--title` | on `edit` only; `new` takes the title as its argument |
| `--desc` | any text |
| `--state` | `backlog` `todo` `current` `blocked` `watching` `paused` `done` — a new task starts in `backlog` |
| `--blocker` | the number of a task in its way; `+3` adds one, `-3` takes one away |
| `--because` | what is in its way and is not a task cairn holds |
| `--link` | the number of a task it is linked to, shown on both; `+3` adds one, `-3` takes one away. A link has no kind and no direction, and any two tasks can be linked but a task and itself |
| `--priority` | `low` `normal` `high` `critical` — a new task carries `normal` |
| `--label` | one per flag or several at once: `--label spec --label team` and `--label spec,team` are the same thing. On `edit`, a label given plainly or as `+urgent` joins the others, `-spec` takes one away, and `=spec,team` says what the set becomes. See [labels](labels.md) for how a label is written |
| `--deadline` | a day: `2026-09-01` |
| `--estimate` | a duration: `90m`, `1h30m`, `2h` — read back as `1h30m`. Kept to the second, so anything finer is refused rather than rounded away |
| `--spent` | the time spent, which is the time the task was in focus and the time added by hand. `+30m` adds time by hand today, `-15m` takes some back, and given plainly it sets the whole, the difference added today. It never falls below nothing, and a minute begun counts whole |
| `--steps` | the steps the task is split into, one a line: `[ ] its title` for one to do, `[x] its title` for one done, in their order; `-` reads them from standard input. On `edit` they replace the steps it had, and given empty it keeps none |
| `--objective` | the number of an objective this task serves; `+4` adds one, `-1` takes one away. Given empty, it sets the task loose |

A title is one line: a line break or a tab in it is refused, since neither survives a listing or a card. A description is several lines if it likes, but nothing a terminal would read as an instruction — an escape, a bell — gets in, and a label is one line like a title. A description too long to type is read from standard input, named as `-`: `cairn task new "Write the spec" --desc - < notes.md`. Given empty, an option clears what was there. `task show` says since when a task has stood in its column — `in todo since 2026-09-26 · 9 days`, `in_column_since` to a machine — read off what was done to it. Deleting a task
deletes its steps with it. Add `--json` to any command to be answered as a
machine — including when it refuses.

## Backups, exports and versions

```sh
cairn backup new                 # a new backup of the store as it stands
cairn backup list                # every backup, with its number and date
cairn backup show 2              # when it was made, and what it holds
                                 # `holds 4 task(s) and 2 objective(s)`
cairn backup delete 2 --yes
cairn backup restore 2 --yes     # restore one; without a number, the most recent

cairn task export ~/tasks.json   # every task and objective, never over a file
cairn task import ~/tasks.json   # add the tasks of such a file, as new tasks

cairn migration list             # every change, applied or waiting
cairn migration show 2           # what it does, and what rolling it back does
cairn migration apply            # apply what is waiting; `apply 2` for one
cairn migration rollback --yes   # unmake the last change applied
```

When cairn knows a change a store on disk does not, it says so and stops —
`the store is behind this cairn`, then `cairn migration apply` on a line of its
own, so it is copied whole; a machine reading the refusal as JSON gets it in
one sentence. The boards ask instead: opened on
such a store, they offer to apply what is waiting, or to quit. Nothing about the
shape of your store changes unless you asked for it. `apply` and `rollback` set
a backup aside first, and so does `backup restore` before replacing anything, so
picking the wrong one costs a command rather than the work it overwrote.

A backup is read in the shape it was written, never today's: one taken before
objectives existed says it holds none, rather than refusing to be read at all.

An export is a file of version 7. Beside each task and the objectives it
serves, it carries the time spent day by day (`"spent": {"2026-10-01": "2h"}`),
whether the task is in focus (`"focused": true`), every label with its
colour and, when it is not for both, its scope (`"scope": "objectives"`),
labels nothing carries included, and each objective's `parent` and
`waits_on`, the objectives they name written with it, and every series, its
rule, template, state and trace (see [series](series.md#carried-in-a-file)), and every
view, its name and its filters (see [views](listing.md#views)). All of it is optional, so a file
written by hand — fixtures, a board to show — says only what it needs, and may
count its days from the day it is read: `"deadline": "+3d"`, `"-1d"`, `"+2w"`,
the same in the keys of `spent`. An
export always writes the days as they are. A day that cannot be read refuses
the whole file before anything is added.

The time comes in as time added by hand on its day, so the Time page shows it
where it was spent. A file of version 3 carries the total alone, and it is
added on the day the file is read. A task in
focus comes back in focus, which puts it under way as focusing does. A label
already in the store keeps its colour and its scope, and a view already named
here keeps its filters: nothing here is replaced.

## What cairn answers a script

| code | what it means |
|---|---|
| `0` | it worked |
| `1` | understood, and refused — a title that is empty, a state that is not one — or something nobody foresaw went wrong |
| `2` | not understood at all — a flag that does not exist, a number that is not one |
| `3` | the store could not be used — unreadable, out of date, or nothing may write to it |

Code 3 is the one worth branching on: nothing was wrong with the command, so
the same command may well work once the store is reachable again.

Asked for `--json`, a refusal says the same as a document a machine reads:
`{"error": "task 4: there is no such task", "kind": "refused", "reason":
"no-such-task"}`. The kind is `refused`, `usage`, `store` or `failed`, and the
reason is a word that stays the same when the sentence is reworded — the thing
to branch on. A deletion refused for want of `--yes` names what it would have
taken.

## Upgrading

`brew upgrade cairn`, or `cairn upgrade` for a binary installed from a release
archive. A store a new version reshapes waits for `cairn migration apply`,
which saves a backup first.

## Reporting a bug

Run cairn again with `--debug` (or `CAIRN_DEBUG=1`), do the one thing that goes
wrong, then `cairn debug report`, and attach the archive it writes to an
[issue](https://github.com/vonglasow/cairn/issues). None of the text you typed
is in it.
