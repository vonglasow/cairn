# The board in a window

```sh
cairn gui                        # the same board, in a window
cairn app install                # add cairn to the Applications or the desktop's menu
```

The same board as the terminal's, at a size a mouse can reach, on macOS and
Linux. The mouse comes first: no modes and no cursor to learn. The bar along
the bottom writes each key beside what it does, a key works exactly where the
bar offers it, and whatever the pointer rests on lights up.

It needs 720 by 560; below that it says so. The look follows the
[scheme](config.md#the-window) the configuration names.

## The board

Four columns — backlog, todo, current, done — the backlog full height, and
under the other three a band of three lanes: blocked, watching, paused. Each
column names itself and counts its cards; done is drawn at half strength.

A card says its number, its title, its priority as a stripe down its left
side, when it is due, its time spent against its estimate, its labels and the
objective it serves. Resting the pointer on a cut title, on `+n` labels, or on
`◇1` shows them whole. `⊳2` says two tasks wait on it.

The tasks in focus come first in `current`.

Pages, as tabs drawn like folders: **tasks**, **objectives**, **time**, and
**series** while recurrence is on. `tab` and `shift+tab` walk them; `g` then
`t`, `o` or `i` goes to one.

## The mouse

- **Click** chooses a card, **double-click** opens it.
- **Drag** a card to another column to change its state, or within its column
  to change its place; a line shows where it would land, and nothing else moves.
  Dropped into done with a step open, it asks first.
- **Double-click on an empty spot** of a column adds a task there.
- **Right-click** a card for its menu — edit, a new linked task, its labels,
  and the rest of what can be done to it — or a column to add a task to it.
- **`⌘`-click** gathers several cards, **`⇧`-click** takes every card between —
  `Ctrl` on Linux — and an action then applies to all of them.
- **Drag a gutter** between two columns or lanes to widen one, or the edge
  above the band to make it taller. A double-click on a gutter takes the size
  back; **Reset sizes** in the View menu takes them all back. To keep the sizes
  from one window to the next, see [configuration](config.md#the-window).
- The wheel scrolls what it began on.

## Keys

| key | what it does |
|---|---|
| `⌘N` | a new task |
| `⏎` | open the chosen task |
| `⌘F` | search |
| `⇧F` | the filters |
| `⌘K` | the command palette: any command, a number, a title |
| `m` | move the chosen task to a column |
| `o` | a new task linked to the chosen one |
| `x` | archive or unarchive the chosen task |
| `⇧A` | show what was archived |
| `⇧X` | archive the whole column — it asks first |
| `⌘⌫` | delete — it asks first |
| `⇧L` | the labels |
| `⌘B` `⇧B` | back the store up, list the backups |
| `⌘E` `⌘O` | export every task to a file, import one |
| `⌘1`…`⌘9` | open a view; `⌘0` the whole board |
| `⌘⇧O` | open another workspace |
| `⌘,` | the settings |
| `⌘Z` | in a form, bring back what was typed |
| `?` | help |
| `esc` | let go of the chosen card, close a panel |

The arrows walk the board. The menu bar offers the same actions.

## A task, open

The form opens over the board, and can be dragged by its ground. It **saves
itself** every couple of seconds; `⌘⏎` or `⌘S` saves at once.

- **Details** shows the fields; **Trace** (`⌘T`) what was done to it, as
  `trace show` prints it.
- The `labels` and `objectives` fields open lists to tick: typing narrows the
  list, and a label that does not exist yet is offered to create.
- Steps are tiles of the form: tick one with a click, drag one to reorder,
  choose one and press delete to remove it after asking.
- A deadline opens a calendar.
- At the foot, the line saying where the task's [notes](context.md) are.
- An occurrence of a series says its rule; saving it asks whether the change is
  for this one or the whole series.

## Searching, filtering, views

`⌘F` narrows the board in place as you type, every page alike. `⇧F`, or
**+ Filter**, adds a filter — column, priority, label, objective, a day due by,
series, focus — and what filters the board stands above it as pills. Each
value says how many tasks it would keep.

The filter menu lists the [views](listing.md#views) at its head; `⌘1`…`⌘9`
open them. **Save as a view…** keeps what filters the board under a name.

## Objectives

The Objectives page is a tree on the left — each objective with its
sub-objectives under it, a parent folded until opened — and the chosen one on
the right: how far along it is, its time, its sub-objectives as tiles, and its
tasks. `t` shows its tasks on the board, with a bar narrowing to each part.
Deleting one asks, row by row, what becomes of what it holds. See
[objectives](objectives.md).

## Time

The Time page shows a day, a week or a month — `1`, `2`, `3` — each day with
its total and a line per task. With `[time] weekends = "hidden"`, a week is
five days. See [the timesheet](trace.md#how-long-the-work-took).

## Labels, backups, settings

- `⇧L` lists every label with what carries it: `n` makes one, `r` renames, `c`
  chooses its colour, `⌘⌫` takes it off everything after asking.
- `⇧B` lists the backups; `⏎` restores one after asking, a backup of what the
  store holds made first.
- `⌘,` opens the settings of the workspace and the global configuration, every
  key with where its value comes from. A change is written to the file at once,
  keeping its comments.

## Workspaces

The window's title, and a pill at its top, name the workspace it shows, in the
workspace's colour. `⌘⇧O`, or a click on the pill, opens the switcher:
choosing another workspace opens it in a window of its own, or brings it
forward when it is already open. `⌘⇧` and a figure choose one directly.

## While it is open

What another terminal or an agent changes turns up on its own. When the window
opens, the board arrives by a short intro, which the configuration can
[change](config.md#the-window). A store that is behind this cairn is offered
its changes, a backup made first; one a newer cairn changed is refused.
