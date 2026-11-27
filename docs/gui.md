# In a window

```sh
cairn gui                        # the same board, in a window
cairn app install                # add cairn to the Applications or the desktop's menu
```

The same board as the terminal's, at a size a mouse can reach, on macOS and
Linux. The mouse comes first: no modes and no cursor to learn. The bar along
the bottom writes each key beside what it does, a key works exactly where the
bar offers it, and whatever the pointer rests on lights up.

It needs 720 by 560; below that it says so. The look follows the
[scheme](config.md#every-key) the configuration names.

## The board

Four columns — backlog, todo, current, done — the backlog full height, and
under the other three a band of three lanes: blocked, watching, paused. Each
column names itself and counts its cards; done is drawn at half strength.

A card says its number, its title, its priority as a stripe down its left
side, what it is waiting on while it is stopped, when it is due, its time
spent against its estimate, its steps, its labels and the objective it serves, under its parent and cut from its
start when short of room. Resting the pointer on a cut title, on `+n` labels,
or on what it serves shows them whole, just above. `⊳2` says two tasks wait on it.

The tasks in focus come first in `current`.

Pages, as tabs drawn like folders: **tasks**, **objectives**, **time**,
**series** while recurrence is on, and **agents**. <kbd>⇥</kbd> and <kbd>⇧⇥</kbd> walk them; <kbd>g</kbd> then
<kbd>t</kbd>, <kbd>o</kbd>, <kbd>i</kbd> or <kbd>a</kbd> goes to one.

A card a Claude session holds is ringed: amber while the session waits for
you, blue while it works.

## The mouse

- **Click** chooses a card, **double-click** opens it.
- **Drag** a card to another column to change its state, or within its column
  to change its place; a line shows where it would land, and nothing else moves.
  Dropped into done with a step open, it asks first.
- **Double-click on an empty spot** of a column adds a task there.
- **Right-click** a card for its menu — edit, a new linked task, its labels,
  and the rest of what can be done to it — or a column to add a task to it.
- **<kbd>⌘</kbd>-click** gathers several cards, **<kbd>⇧</kbd>-click** takes every card between —
  `Ctrl` on Linux — and an action then applies to all of them. A right-click on
  a gathered card keeps them gathered: its menu moves, prioritises, labels,
  focuses, archives or deletes all of them. Under **Labels…**, `✓` is a label
  all of them carry and `–` one some of them carry; a click puts it on all of
  them, or takes it off all of them.
- **Drag a gutter** between two columns or lanes to widen one, or the edge
  above the band to make it taller. A double-click on a gutter takes the size
  back; **Reset sizes** in the View menu takes them all back. To keep the sizes
  from one window to the next, see [configuration](config.md#every-key).
- The wheel scrolls what it began on.
- A column, a lane, the objectives or the Agents page that holds more than
  it shows has a bar at its right. Press the bar to go there, and drag it to
  scroll.

## Keys

| key | what it does |
|---|---|
| <kbd>⌘N</kbd> | a new task |
| <kbd>⏎</kbd> | open the chosen task |
| <kbd>⌘F</kbd> | search |
| <kbd>⇧F</kbd> | the filters |
| <kbd>⌘K</kbd> | the command palette: any command, a number, a title |
| <kbd>m</kbd> | move the chosen task to a column |
| <kbd>⇧</kbd> held, the arrows | carry the chosen task, and the cards gathered with it, as a drag does; letting <kbd>⇧</kbd> go drops it, <kbd>esc</kbd> puts it back |
| <kbd>o</kbd> | a new task linked to the chosen one |
| <kbd>x</kbd> | archive or unarchive the chosen task |
| <kbd>⇧A</kbd> | show what was archived |
| <kbd>⇧X</kbd> | archive the whole column — it asks first |
| <kbd>⌘⌫</kbd> | delete — it asks first |
| <kbd>⇧L</kbd> | the labels |
| <kbd>⌘B</kbd> <kbd>⇧B</kbd> | back the store up, list the backups |
| <kbd>⌘E</kbd> <kbd>⌘O</kbd> | export every task to a file, import one |
| <kbd>⌘1</kbd>…<kbd>⌘9</kbd> | open a view; <kbd>⌘0</kbd> the whole board |
| <kbd>⌘⇧O</kbd> | open another workspace |
| <kbd>⌘,</kbd> | the settings |
| <kbd>⌘Z</kbd> | in a form, bring back what was typed |
| <kbd>?</kbd> | help |
| <kbd>esc</kbd> | let go of the chosen card, close a panel |

On the Agents page, the arrows or <kbd>h</kbd> <kbd>j</kbd> <kbd>k</kbd>
<kbd>l</kbd> choose a session:

| key | what it does |
|---|---|
| <kbd>⏎</kbd> | go to the session, or copy the command that resumes it |
| <kbd>o</kbd> | open its task |
| <kbd>a</kbd> <kbd>d</kbd> | allow, deny what it asks; <kbd>a</kbd> opens its questions |
| <kbd>c</kbd> | copy `claude --resume` |
| <kbd>esc</kbd> | let the session go |

The arrows walk the board. The menu bar offers the same actions; on Windows,
About cairn closes its Help menu. On Linux,
where a window has no menu bar of its own, the button at the top right holds
it: Settings…, Upgrade, then Help and About cairn, then Quit.

## A task, open

The form opens over the board, and can be dragged by its ground. It **saves
itself** every couple of seconds; <kbd>⌘⏎</kbd> or <kbd>⌘S</kbd> saves at once.

- A task in focus says **In focus** beside its title, with the glowing dot its
  card wears.
- **Details** shows the fields; **Trace** (<kbd>⌘T</kbd>) what was done to it, as
  `trace show` prints it.
- <kbd>⇥</kbd> walks the fields in the order they are drawn, and from the last back to
  the title; <kbd>⇧⇥</kbd> walks back.
- The `labels` and `objectives` fields open lists to tick: typing narrows the
  list, and a label that does not exist yet is offered to create. A click,
  or <kbd>space</kbd>, ticks or unticks a line and the list stays open; with
  a word typed, <kbd>⏎</kbd> takes the line found. **Done** — <kbd>⏎</kbd>
  with nothing typed, or a click outside — keeps what is ticked; <kbd>esc</kbd>
  puts it back.
- The `blockers` and `links` fields open the other tasks to tick, the newest
  first; typing keeps the ones a search would find.
- Steps are tiles of the form: tick one with a click, drag one to reorder,
  choose one and press delete to remove it after asking.
- A deadline opens a calendar.
- At the foot, the line saying where the task's [notes](context.md) are.
- An occurrence of a series says its rule; saving it asks whether the change is
  for this one or the whole series.

## Searching, filtering, views

<kbd>⌘F</kbd> narrows the board in place as you type, every page alike. <kbd>⇧F</kbd>, or
**+ Filter**, adds a filter — column, priority, label, objective, a day due by,
unblocked, series, focus — and what filters the board stands above it as pills. Each
value says how many tasks it would keep.

The filter menu lists the [views](listing.md#views) at its head; <kbd>⌘1</kbd>…<kbd>⌘9</kbd>
open them. **Save as a view…** keeps what filters the board under a name.

## Objectives

The Objectives page is a tree on the left — each objective with its
sub-objectives under it, a parent folded until opened — and the chosen one on
the right: how far along it is, its time, its sub-objectives as tiles, and its
tasks. <kbd>t</kbd> shows its tasks on the board, with a bar narrowing to each part.
Deleting one asks, row by row, what becomes of what it holds. See
[objectives](objectives.md).

## Time

The Time page shows a day, a week or a month — <kbd>1</kbd>, <kbd>2</kbd>, <kbd>3</kbd> — each day with
its total and a line per task. With `[time] weekends = "hidden"`, a week is
five days. See [the timesheet](trace.md#how-long-the-work-took).

## Agents

The Agents page shows a card for each Claude session running in the workspace:
whether it works, waits for you or is idle, and since when; what it does or
what it waits for; the task it holds and its open steps; its branch, model and
context; its folder and terminal. A session holding no task shows too, under
the name Claude gave it, and **Maybe #n …? Tie** when a task in focus or in current is held
by no session — the one its branch names when several are. Two sessions
holding one task each say so. Resting the pointer on a card's title, what it
does, its branch, its model or where it runs shows them whole. **State**, **Task** and **Project** at
the top group the cards.

- **Go to session** brings its terminal to the front when cairn reaches it,
  **Open task** opens its task; a session holding none offers **New task** and
  **Tie…**. The mark after its id copies the command that resumes it. A
  session in a worktree Claude made is said under its repository's name.
- A session working that has written nothing for `agents.silent` (10m unless
  set) says **silent for 12m**, in amber. Its context turns amber at 70%, and
  red with **compact soon** at 85%.
  What it does is said in a few words: "Runs go", "Edits card.go".
- **Click** a card to open its task and bring its terminal to the front:
  iTerm2, Terminal and Ghostty on macOS, kitty on Linux. In any other terminal, or in
  an editor, the command that resumes the session is copied instead. On
  Windows the card names what the session runs in — Windows Terminal, VS
  Code, WezTerm or Claude Desktop — and a click copies the command, for
  PowerShell.
- **Right-click** a card to open its task, go to the session, copy the
  command that resumes it, make a task for it — named as Claude named the
  session, put in current — or tie it to a task of the board.
- **Right-click** a task card, then **Start a Claude session…**: choose the
  terminal and the folder; claude starts there, asked to work on the task.
- The red count beside the tab says how many sessions wait for you, and a
  desktop notification says when one starts waiting. `agents.count` and
  `agents.notify` turn each off.
- With `agents.permissions = true`, a session asking leave to use a tool asks
  on its card while the window is open: **Allow**, **Allow always…**,
  in the project's settings or in yours, or **Deny**. A session asking
  questions says how many; **Answer…** opens them, one option or several
  each, and **Send** answers once every question has one. With the window closed, or the key
  off, the terminal asks as it always does.

`pages.agents = "hidden"` takes the page away. See
[Claude Code](claude-code.md#the-sessions-running).

## Labels, backups, settings

- <kbd>⇧L</kbd> lists every label with what carries it: <kbd>n</kbd> makes one, <kbd>r</kbd> renames, <kbd>c</kbd>
  chooses its colour, <kbd>⌘⌫</kbd> takes it off everything after asking.
- <kbd>⇧B</kbd> lists the backups; <kbd>⏎</kbd> restores one after asking, a backup of what the
  store holds made first.
- <kbd>⌘,</kbd> opens the settings, for **This workspace** or for **All workspaces**.
  Each row says what it does; resting the pointer on it shows its key. Beside
  it, where its value comes from: **Default**, **Shared** by all workspaces,
  **Only here**, or **Changed**; **Use the shared value** and **Back to
  default** take a change back. A change is written to the file at once,
  keeping its comments; **Open the file** opens it. A field typed in is saved by <kbd>⏎</kbd> or a click elsewhere,
  and <kbd>esc</kbd> puts it back; a day is picked on the calendar, and the workspace's
  colour on its swatches, by a click or by the arrows and <kbd>⏎</kbd>; a number of days
  steps with **−** and **+**; the dragged card's opacity slides, by a drag
  or by <kbd>←</kbd> and <kbd>→</kbd> five at a time. For a workspace, **Files** shows where its
  board is, and its **Notes folder**, typed or picked with **Choose…**; the notes left in
  the folder before are offered **Move them** or **Leave them**.

## Workspaces

The window's title, and a pill at its top, name the workspace it shows, in the
workspace's colour. <kbd>⌘⇧O</kbd>, or a click on the pill, opens the switcher:
choosing another workspace opens it in a window of its own, or brings it
forward when it is already open. <kbd>⌘⇧</kbd> and a figure choose one directly.

**New workspace…**, at the foot of the switcher, opens a dialog:

1. Type its **name**: lowercase letters, digits and dashes.
2. Its **folder** follows the name, `~/.cairn/<name>`. Click the field to type
   another, or **Choose…** to pick it in the system's dialog. A folder that
   holds a board already is opened as it is.
3. Its **notes** stay in its folder unless another is typed or chosen — a
   vault's, for one: see [Notes in Obsidian](notes-obsidian.md).
4. Click a **colour**. Without one, it wears a colour read off its name.
5. Untick **Switch to it now** to make it without opening it.
6. **Create**, or <kbd>⏎</kbd>. <kbd>⇥</kbd> goes from field to field; <kbd>esc</kbd> or
   **Cancel** makes nothing. A name cairn does not take is said in the dialog.

## While it is open

What another terminal or an agent changes turns up on its own. When the window
opens, the board arrives by a short intro, which the configuration can
[change](config.md#every-key). A store that is behind this cairn is offered
its changes, a backup made first; one a newer cairn changed is refused.
