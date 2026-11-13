# Listing, searching, and views

Two commands answer with several tasks. `task list` narrows by facts from a
closed set — the column, the priority, a label. `task search` looks for words
in everything a task says. Both narrow the same way, and every surface uses the
same search.

## Narrowing a listing

```sh
cairn task list                                # the board, column by column
cairn task list --state todo --priority high   # only what carries all of it
cairn task list --label team --label urgent    # both labels, not either
cairn task list --objective 1                  # what serves 1 or a sub-objective of it
cairn task list --objective 1 --itself         # 1 alone
cairn task list --deadline-before 2026-10-01   # due on or before that day
cairn task list --blocker 3                    # what task 3 is in the way of
cairn task list --focus                        # in focus; --no-focus, the others
cairn task list --number 25 --number 18        # those tasks, archived or not
cairn task list --series                       # occurrences of a series; --series 3 for one
```

Everything named has to hold. A listing is flat and in the board's order:
column first, then the place you set.

- A column or a priority that does not exist is refused, naming those that do.
- An objective nobody made answers nothing.
- A task with no deadline is not "due before" anything.
- A number no task carries is said under `unknown` in JSON.

## What was archived

Archiving takes a task off the board without deleting it: it keeps its column,
its trace and everything it carries.

```sh
cairn task archive 1
cairn task archive --state done --yes    # a whole column at once
cairn task label sprint-12 --state done  # mark a set before sweeping it
cairn task list --archived               # only what was archived
cairn task list --with-archived          # both
cairn task unarchive 1
```

Every listing leaves the archived out until asked. `task show` answers an
archived task either way. On both boards, `x` archives or unarchives the card
under the cursor, and `⇧A` shows what was archived, written faintly.

## Searching

```sh
cairn task search spec                    # every word has to be found
cairn task search spec --state todo       # narrowed like a listing
cairn task search T010 T013 --any         # any of the words
```

A search looks in a task's title, description, steps, number, column,
priority, labels, deadline, estimate and time spent, ignoring case. It does not
look in what a task carries about other tasks — what it serves, what blocks it.
A number is looked for as text, so `228` also finds `1228`.

With `--any`, the JSON says under `matched` which tasks each word found,
including the words that found none.

Both boards search the same way, and filter by the same narrowings: column,
priority, label, objective, a day due by, series, focus. Each value of a filter
says how many tasks it would keep.

## What a row carries

A row is the number, the title, and what the task carries:

```
#12  Write the spec  · current  · high  · 2026-10-09  · backend sprint-172  · ⊳2
```

With `--json`, a row carries every fact of the task, and `described` rather
than the description, which `task show` gives. It also carries facts worked
out as it answers:

| field | what it says |
|---|---|
| `blocks` | how many open tasks name it as a blocker |
| `overdue` | how many days past its deadline |
| `due_in` | how many days left before it, 0 on the day |
| `age` | how many days since it last changed column |
| `objective_late` | an objective it serves is past its deadline and not reached |

## Long boards

A listing can be cut down, which matters most to an agent:

```sh
cairn task list --json --fields number --fields title   # only those fields
cairn task list --limit 50 --offset 100                 # a page; total says how many in all
cairn task list --count state                           # only how many per column
cairn task list --count label                           # only how many per label
cairn task search spec --limit 50                       # a search pages the same way
```

`--fields` shapes the JSON only; the line a person reads stays the same. A
field a row does not carry is refused, naming those it does. A page says what
it holds — `tasks 101 to 150 of 312`.

## Views

A view keeps the filters of a listing under a name — the flags `task list`
takes, and a search — and opens them again in one act:

```sh
cairn view new Home --objective 1                 # in place 1, the window's ⌘1
cairn view new Shopping --label shopping --state todo
cairn view new Spec --search spec
cairn view list                                   # in their places, with what each keeps
cairn view show Home                              # and the tasks outside it touching one it keeps
cairn task list --view Home --state current       # a flag beside it narrows further
cairn view edit Shopping --name Groceries --place 1 --state ""
cairn view delete Groceries                       # the view alone; nothing else changes
```

- A view holds at least one filter or a search, and its name is one no other
  view has.
- Its place is the key that opens it on a board: `⌘1`…`⌘9` in the window,
  `v` and a figure in the terminal.
- It is kept in the store, so a backup and an export carry it.
- A label it names follows a rename. Deleting an objective a view names asks
  what becomes of the view.
- Keeping, changing or deleting a view changes no task and writes no trace.
