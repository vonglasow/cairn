# The board in the terminal

```sh
cairn tui                        # open it
cairn tui --icons unicode        # without a Nerd Font installed
```

The same board as every other surface, on a screen that stays open. It reads
like vim: `hjkl` moves, `i` takes hold of a task, `:` takes any command, and `?`
answers in every mode. `:q` leaves.

It needs a terminal of 60 columns by 26 rows; below that it says so rather than
draw something nobody can read.

## What it shows

```
 ‹ TASKS ›  objectives                                                    ── NORMAL ──
 BACKLOG               12 TODO                   5 CURRENT                2 DONE                  31
 ──────────────────────── ──────────────────────── ──────────────────────── ────────────────────────
 ┌──────────────────────┐ ┃──────────────────────┐ ╔══════════════════════╗ ┌──────────────────────┐
 │ 26                   │ ┃ 14                 ≡ │ ║ 31                 ≡ ║ │ 08                   │
 │ Buy a lamp           │ ┃ Write the spec       │ ║ Spec review          ║ │ Migrate the store    │
 │                      │ ┃                      │ ║                      ║ │                      │
 │                      │ ┃ 2d  ▪3/5  ⚑ spec     │ ║ 1d  ▪2/4  ⚑ spec     ║ │                      │
 └──────────────────────┘ ┃──────────────────────┘ ╚══════════════════════╝ └──────────────────────┘
                          STALLED                                                                  6
                          ──────────────────────────────────────────────────────────────────────────
                          ┃ 19                 ⊘ │ │ 24               ≡ ◉ │ │ 30                 ‖ │
                          ┃ Client's answer      │ │ Handed to Marie      │ │ Refactor migrate     │
                          ┃ ↳ 12                 │ │ ↳ she confirms first │ │ ↳ your call          │
 hjkl ←↓↑→ columns   i insert   a new   / search   ⇥ tab   ? help   : command
```

- **Four columns** — backlog, todo, current, done — and under the last three a
  band, **STALLED**, holding `blocked`, `watching` and `paused`, three cards
  across. An empty column stays on screen.
- **The header** names the workspace shown, with its colour, then the tabs.
- **A card** says its number, its title, how far its deadline is, its steps,
  its labels, and what it is waiting on. The left border carries its priority.
  `⊳2` says two tasks wait on it.
- **The tasks in focus** come first in `current`.
- **The bar at the bottom** shows only the keys that work where you are. A key
  that is not on the bar does nothing there.

Tabs, walked with `⇥`: **tasks**, **objectives**, **time** — a week, day by
day, task by task — and **series** while recurrence is on. Each counts what it
holds. A page turned off in the [configuration](config.md#pages) has no tab.

## Modes

| mode | the cursor is on |
|---|---|
| **normal** | a column |
| **insert** | a task |
| **edit** | a field of the form |

The objectives list has only normal mode: moving between objectives is already
holding one.

## Keys

`hjkl` and the arrows only ever navigate, in every mode.

| key | where | what it does |
|---|---|---|
| `hjkl` `←↓↑→` | everywhere | navigate |
| `i` | normal | take hold of a task (insert) · on objectives, open one |
| `⇧hjkl` | insert | move the task; sideways changes its column, up and down its place |
| `a` | normal | a new task, in the column under the cursor |
| `o` | insert | a new task linked to this one, in `todo` |
| `e` | insert | the form |
| `⏎` | insert | the detail |
| `D` | insert | delete — it asks first |
| `x` | insert | archive or unarchive the task |
| `⇧A` | normal, insert | show what was archived, or hide it |
| `⇧X` | normal | archive the whole column — it asks first |
| `m` then a letter | insert | move to a column: `b` backlog, `t` todo, `c` current, `d` done, `w` watching, `p` paused |
| `p` then a letter | insert | priority: `l` low, `n` normal, `h` high, `c` critical |
| `f` | insert | focus or unfocus |
| `r` then a letter | insert, on an occurrence | `s` skip it, `x` stop the series, `r` resume it, `e` open the series |
| `/` | normal | search |
| `⇧F` | normal | filters |
| `v` | normal | the views |
| `gg` `G` | insert | the first, the last |
| `b` `⇧B` | normal | back the store up, list the backups |
| `^O` | everywhere | open another workspace in place of this one |
| `?` | everywhere | help |
| `:` | normal | a command |
| `⇥` | normal | next tab |
| `esc` | | back one level |


## Objectives

The objectives tab is a tree, two lines an objective, each sub-objective under
its parent. `i` or `⏎` opens an objective's own board, holding only the tasks
that serve it; `esc` goes back to the list. `t` shows an objective's tasks on
the tasks board, `N` makes a sub-objective, `z` folds a parent away, and `[`
`]` narrow an objective's board to one sub-objective, the objective itself, or
all of it. See [objectives](objectives.md).

## Moving tasks

`⇧h` and `⇧l` change a task's column, and its state follows. `⇧j` and `⇧k`
change its place among the tasks of its column, never its column. From the
last task of a column, `⇧j` pushes it into the band and asks which of the three
it becomes; inside the band, `⇧j` and `⇧k` reorder among its kind, and `⇧k`
from the first takes it back to the column it came from.

Plain `j` past the bottom of a column carries the cursor into the band. Walking
sideways keeps the level.

Pushing a task into done while one of its steps is open asks first.

## The form

One form, for adding and for editing. Nothing is written until **save**, or
`^S` from anywhere. Navigate to a field and press `⏎`: a closed set opens a
list, the description opens in `$EDITOR`. The steps are written in the
form, one line each — `[x] Draft the columns`.

Leaving a form you typed into asks first; `esc` again gives it back whole.

## The detail

`⏎` on a task opens it full screen: its description in a box that scrolls
(`^D` `^U`), its steps — those left first — its objectives, its trace, and a
`CONTEXT` panel saying where its [notes](context.md) are. `jk` walks the steps,
and `⌫` deletes the one chosen after asking. An occurrence of a series says its
rule and what the series did so far.

## Searching, filtering, views

- `/` narrows the board in place as you type; the columns stay, and the
  headers count what matched. `esc` gives the whole board back. A search
  reaches what `task search` reaches.
- `⇧F` turns the foot into filters — column, priority, label, objective, a day
  due by, series, focus — `←→` choosing one and `↑↓` walking its values, each
  with how many tasks it keeps. `⌫` clears them all.
- `v` lists the [views](listing.md#views): a figure opens one, `0` the whole
  board. `:view new Urgent` keeps what is filtered under a name.

A task outside a view that blocks, waits on or is linked to one inside it is
drawn faintly at the foot of its column, saying why.

## The command line inside it

`:` takes every command the command line takes — `:task new "…"`,
`:timesheet show` — and shows what it answers on a screen of its own. `:set`
and `:unset` change the configuration. `:q` leaves.

## While it is open

What another terminal or an agent changes turns up on its own, and the cursor
stays on the task it was holding. A store that a newer cairn changed, or one
behind this cairn, is asked about before anything else.
