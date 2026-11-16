# In the terminal

```sh
cairn tui                        # open it
cairn tui --icons unicode        # without a Nerd Font installed
```

The same board as in the window, on a screen that stays open, in the
colours of the scheme the window wears. It reads like vim: <kbd>hjkl</kbd> walks the
cards, <kbd>:</kbd> takes any command, and <kbd>?</kbd> answers everywhere. `:q` leaves.

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

Tabs, walked with <kbd>⇥</kbd>: **tasks**, **objectives**, **time** — a week, day by
day, task by task — and **series** while recurrence is on. Each counts what it
holds. A page turned off in the [configuration](config.md#every-key) has no tab.

## The cursor

The cursor is always on a column, and on a card of it when there is one: the
board opens on the backlog's first card. There are no modes.

## Keys

<kbd>hjkl</kbd> and the arrows only ever navigate.

| key | what it does |
|---|---|
| <kbd>hjkl</kbd> <kbd>←↓↑→</kbd> | walk the cards: sideways over an empty column to the next card, down from a column's last card into the lane under it, up out of a lane |
| <kbd>^H</kbd> <kbd>^L</kbd> | the column left or right, at once, even an empty one |
| <kbd>^J</kbd> <kbd>^K</kbd> | into the band, out of it |
| <kbd>⇧hjkl</kbd> | move the task; sideways changes its column, up and down its place |
| <kbd>a</kbd> | a new task, in the column under the cursor |
| <kbd>o</kbd> | a new task linked to the one held, in `todo` |
| <kbd>e</kbd> | the form |
| <kbd>⏎</kbd> | the detail · on objectives, open one |
| <kbd>D</kbd> | delete — it asks first |
| <kbd>x</kbd> | archive or unarchive the task |
| <kbd>⇧A</kbd> | show what was archived, or hide it |
| <kbd>⇧X</kbd> | archive the whole column — it asks first |
| <kbd>m</kbd> then a letter | move to a column: <kbd>b</kbd> backlog, <kbd>t</kbd> todo, <kbd>c</kbd> current, <kbd>d</kbd> done, <kbd>w</kbd> watching, <kbd>p</kbd> paused |
| <kbd>p</kbd> then a letter | priority: <kbd>l</kbd> low, <kbd>n</kbd> normal, <kbd>h</kbd> high, <kbd>c</kbd> critical |
| <kbd>f</kbd> | focus or unfocus |
| <kbd>r</kbd> then a letter | on an occurrence: <kbd>s</kbd> skip it, <kbd>x</kbd> stop the series, <kbd>r</kbd> resume it, <kbd>e</kbd> open the series |
| <kbd>/</kbd> | search |
| <kbd>⇧F</kbd> | filters |
| <kbd>v</kbd> | the views |
| <kbd>gg</kbd> <kbd>G</kbd> | the first, the last |
| <kbd>b</kbd> <kbd>⇧B</kbd> | back the store up, list the backups |
| <kbd>^O</kbd> | open another workspace in place of this one; there, <kbd>^N</kbd> makes one, asking its name, its folder, its notes and its colour |
| <kbd>?</kbd> | help |
| <kbd>:</kbd> | a command |
| <kbd>⇥</kbd> | next tab |
| <kbd>esc</kbd> | back one level |

A terminal that sends <kbd>^H</kbd> or <kbd>^J</kbd> as another key keeps them for it; <kbd>hjkl</kbd>
reach the same places, but for an empty column.

## Objectives

The objectives tab is a tree, two lines an objective, each sub-objective under
its parent. <kbd>⏎</kbd> opens an objective's own board, holding only the tasks
that serve it; <kbd>esc</kbd> goes back to the list. <kbd>t</kbd> shows an objective's tasks on
the tasks board, <kbd>N</kbd> makes a sub-objective, <kbd>z</kbd> folds a parent away, and <kbd>[</kbd>
<kbd>]</kbd> narrow an objective's board to one sub-objective, the objective itself, or
all of it. See [objectives](objectives.md).

## Moving tasks

<kbd>⇧h</kbd> and <kbd>⇧l</kbd> change a task's column, and its state follows. <kbd>⇧j</kbd> and <kbd>⇧k</kbd>
change its place among the tasks of its column, never its column. From the
last task of a column, <kbd>⇧j</kbd> pushes it into the band and asks which of the three
it becomes; inside the band, <kbd>⇧j</kbd> and <kbd>⇧k</kbd> reorder among its kind, and <kbd>⇧k</kbd>
from the first takes it back to the column it came from.

Plain <kbd>j</kbd> past the bottom of a column carries the cursor into the band. Walking
sideways keeps the level.

Pushing a task into done while one of its steps is open asks first.

## The form

One form, for adding and for editing. Nothing is written until **save**, or
<kbd>^S</kbd> from anywhere. Navigate to a field and press <kbd>⏎</kbd>: a closed set opens a
list, the description opens in `$EDITOR`. The steps are written in the
form, one line each — `[x] Draft the columns`. Under `blockers` and `links`,
the tasks already given are named; type a word beside the numbers and the
tasks it reaches are named, the newest first, five at most — then type the
number of the one you want.

Leaving a form you typed into asks first; <kbd>esc</kbd> again gives it back whole.

## The detail

<kbd>⏎</kbd> on a task opens it full screen: its description in a box that scrolls
(<kbd>^D</kbd> <kbd>^U</kbd>), its steps — those left first — its objectives, its trace, and a
`CONTEXT` panel saying where its [notes](context.md) are. <kbd>jk</kbd> walks the steps,
and <kbd>⌫</kbd> deletes the one chosen after asking. An occurrence of a series says its
rule and what the series did so far.

## Searching, filtering, views

- <kbd>/</kbd> narrows the board in place as you type; the columns stay, and the
  headers count what matched. <kbd>esc</kbd> gives the whole board back. A search
  reaches what `task search` reaches.
- <kbd>⇧F</kbd> turns the foot into filters — column, priority, label, objective, a day
  due by, unblocked, series, focus — <kbd>←→</kbd> choosing one and <kbd>↑↓</kbd> walking its values, each
  with how many tasks it keeps. <kbd>⌫</kbd> clears them all.
- <kbd>v</kbd> lists the [views](listing.md#views): a figure opens one, <kbd>0</kbd> the whole
  board. `:view new Urgent` keeps what is filtered under a name.

A task outside a view that blocks, waits on or is linked to one inside it is
drawn faintly at the foot of its column, saying why.

## The command line inside it

<kbd>:</kbd> takes every command the command line takes — `:task new "…"`,
`:timesheet show` — and shows what it answers on a screen of its own. `:set`
and `:unset` change the configuration. `:q` leaves.

## While it is open

What another terminal or an agent changes turns up on its own, and the cursor
stays on the task it was holding. A store that a newer cairn changed, or one
behind this cairn, is asked about before anything else.
