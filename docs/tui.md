# The board in the terminal

```sh
cairn tui                        # open it
cairn tui --icons unicode        # without a Nerd Font installed
```

The same board as every other surface, on a screen that stays open, in the
colours of the scheme the window wears. It reads like vim: `hjkl` walks the
cards, `:` takes any command, and `?` answers everywhere. `:q` leaves.

It needs a terminal of 60 columns by 26 rows; below that it says so rather than
draw something nobody can read.

## What it shows

```
 TASKS 11   objectives 2   time 0

 BACKLOG                3│TODO                   2│CURRENT                1│DONE                   1
 ────────────────────────│────────────────────────│────────────────────────│────────────────────────
 ┏━━━━━━━━━━━━━━━━━━━━━━┓│┏━━━━━━━━━━━━━━━━━━━━━━┓│╔══════════════════════╗│┏━━━━━━━━━━━━━━━━━━━━━━┓
 ┃ 1                    ┃│┃ 4                  ≡ ┃│║ 6                  ≡ ║│┃ 7                    ┃
 ┃ Buy a lamp           ┃│┃ Write the spec       ┃│║ Spec review          ║│┃ Migrate the store    ┃
 ┃                      ┃│┃                      ┃│║                      ║│┃                      ┃
 ┃                      ┃│┃ 2d  ▪1/2  ⚑ spec     ┃│║ 1d  ▪2/4  ⚑ spec     ║│┃                      ┃
 ┗━━━━━━━━━━━━━━━━━━━━━━┛│┗━━━━━━━━━━━━━━━━━━━━━━┛│╚══════════════════════╝│┗━━━━━━━━━━━━━━━━━━━━━━┛
                         │STALLED                4│                        │
                         │blocked                2│watching               1│paused                 1
                         │────────────────────────│────────────────────────│────────────────────────
                         │┏━━━━━━━━━━━━━━━━━━━━━━┓│┏━━━━━━━━━━━━━━━━━━━━━━┓│┏━━━━━━━━━━━━━━━━━━━━━━┓
                         │┃ 8                ≡ ⊘ ┃│┃ 9                ≡ ◉ ┃│┃ 10                 ‖ ┃
                         │┃ Client's answer      ┃│┃ Handed to Marie      ┃│┃ Refactor migrate     ┃
 hjkl ←↓↑→ tasks   ⇧hjkl move   ⏎ open   e edit   f focus   D delete   x archive   ? help   a new
```

- **Four columns** — backlog, todo, current, done — and under the last three a
  band, **STALLED**, in three lanes: `blocked` under todo, `watching` under
  current, `paused` under done, as in the window. An empty column stays on
  screen.
- **The header** names the workspace shown, with its colour, then the tabs.
- **A card** says what a card says in the window: what it is waiting on, how
  far its deadline is, its time spent against its estimate, its steps, its
  labels and the objectives it serves. The left border carries its priority.
  `⊳2` says two tasks wait on it.
- **The tasks in focus** come first in `current`.
- **The bar at the bottom** shows only the keys that work where you are. A key
  that is not on the bar does nothing there.

Tabs, walked with `⇥`: **tasks**, **objectives**, **time** — a week, day by
day, task by task — and **series** while recurrence is on. Each counts what it
holds. A page turned off in the [configuration](config.md#every-key) has no tab.

## The cursor

The cursor is always on a column, and on a card of it when there is one: the
board opens on the backlog's first card. There are no modes.

## Keys

`hjkl` and the arrows only ever navigate.

| key | what it does |
|---|---|
| `hjkl` `←↓↑→` | walk the cards: sideways over an empty column to the next card, down from a column's last card into the lane under it, up out of a lane |
| `^H` `^L` | the column left or right, at once, even an empty one |
| `^J` `^K` | into the band, out of it |
| `⇧hjkl` | move the task; sideways changes its column, up and down its place |
| `a` | a new task, in the column under the cursor |
| `o` | a new task linked to the one held, in `todo` |
| `e` | the form |
| `⏎` | the detail · on objectives, open one |
| `D` | delete — it asks first |
| `x` | archive or unarchive the task |
| `⇧A` | show what was archived, or hide it |
| `⇧X` | archive the whole column — it asks first |
| `m` then a letter | move to a column: `b` backlog, `t` todo, `c` current, `d` done, `w` watching, `p` paused |
| `p` then a letter | priority: `l` low, `n` normal, `h` high, `c` critical |
| `f` | focus or unfocus |
| `r` then a letter | on an occurrence: `s` skip it, `x` stop the series, `r` resume it, `e` open the series |
| `/` | search |
| `⇧F` | filters |
| `v` | the views |
| `gg` `G` | the first, the last |
| `b` `⇧B` | back the store up, list the backups |
| `^O` | open another workspace in place of this one |
| `?` | help |
| `:` | a command |
| `⇥` | next tab |
| `esc` | back one level |

A terminal that sends `^H` or `^J` as another key keeps them for it; `hjkl`
reach the same places, but for an empty column.

## Objectives

The objectives tab is a tree, two lines an objective, each sub-objective under
its parent. `⏎` opens an objective's own board, holding only the tasks
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
form, one line each — `[x] Draft the columns`. Under `blockers` and `links`,
the tasks already given are named; type a word beside the numbers and the
tasks it reaches are named, the newest first, five at most — then type the
number of the one you want.

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
  due by, unblocked, series, focus — `←→` choosing one and `↑↓` walking its values, each
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
