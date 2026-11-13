# The board

How it is used is [docs/tui.md](../../docs/tui.md), which the site publishes;
this page is why it is that way, for whoever changes it.

Read this before touching the terminal board — its shape, its keys, its modes —
or before adding anything a person sees on screen.

`cairn tui` opens it. What follows is what it does and why it does it that
way, decision by decision — written down because thirty decisions taken in one
conversation are thirty decisions nobody can find afterwards.

The board is a surface, like the command line and the MCP server: it parses what
somebody pressed, calls `internal/engine`, and draws what comes back. It decides
nothing. It breaks one assumption the other two never did — **it stays open**,
and the store can change underneath it. That is dealt with at the end.

## Read this first

The whole file in eight lines, for whoever has a change to make and not an
afternoon to spend:

- **Three tabs**, tasks, objectives as a tree of two-line rows, and the time spent; four columns and a band of three stopped
  states under the last three, and the backlog full height. A fourth, the
  series, while recurrence is on.
- **Modes**: normal stands on a column, insert holds a task, edit is the form.
  The bar shows only what works where it is.
- **The cursor holds a task's number**, never a place on screen — a redraw can
  move it, a keystroke cannot land on the wrong task.
- **`hjkl` moves the cursor; `⇧hjkl` moves the task.** Sideways is a change of
  column, which is an edit; up and down is a place within its column.
- **It decides nothing**: every change goes through the engine, and what a
  board reads and asks is shared with the window in `internal/engine`.
- **It stays true while the store changes**: a beat reads `data_version` and
  redraws when another connection wrote.
- Frames are pinned in `internal/tui/testdata`; a change that moves one is
  written again with `UPDATE_FRAMES=1`, on purpose.

## Contents

- [Three tabs](#three-tabs)
- [The board's shape](#the-boards-shape)
- [Width](#width)
- [A card is six lines](#a-card-is-six-lines)
- [Colour carries what a glyph would cost width for](#colour-carries-what-a-glyph-would-cost-width-for)
- [The terminal's own colours](#the-terminals-own-colours)
- [Glyphs](#glyphs)
- [The modes](#the-modes)
- [Every key](#every-key)
- [The form](#the-form)
- [The detail](#the-detail)
- [Searching](#searching)
- [Help, and the command line](#help-and-the-command-line)
- [Objectives](#objectives)
- [Series](#series)
- [Another workspace, in place of this one](#another-workspace-in-place-of-this-one)
- [Staying true while the store changes](#staying-true-while-the-store-changes)
- [What it is built from](#what-it-is-built-from)
- [How the form is laid out](#how-the-form-is-laid-out)
- [How it reaches the engine](#how-it-reaches-the-engine)
- [Deliberately not there yet](#deliberately-not-there-yet)

## Three tabs

**tasks** — the kanban, where the work is. **objectives** — the list of what the
work is for. **time** — how long the work took, a week at a time, as the
window's TIME page lays it out: each day down the screen, its total, and a line
for each task worked on with its time, read whole rather than cut to a seventh
of a terminal's width; `←→` (or `h l`) walk the weeks, `t` comes back to
today's, `j k` scroll. `⇥` walks the three round. Nothing else is a tab.

The week is laid out as the window lays it out, and learned once: with
`[time] weekends = "hidden"` it is five days, Monday to Friday, the days the
engine says a week shows (`Pages.WeekDays`), as Google Calendar's "Show weekends"
off does for a company's week (`@tui-a-week-without-its-weekend`). The week is
still read whole: time spent on a Saturday is not lost from a total, its lines
are only not drawn (see [config.md](config.md#the-weekends-of-a-week)).

**series** — the tasks that come back, a fourth tab after the time, there only
while `[series] on` is true; see [below](#series).

**Each tab counts what it holds**, as the window's tabs do: the tasks the board
shows, the objectives listed, the tasks the week shown was worked on
(`Timesheet.Worked`, the window's count too), read again each time the board
is, so the count is there before the tab is ever opened
(`@tui-the-time-tab-counts-before-it-is-opened`), and the series listed —
`shown/all` while a search, a filter or a view narrows a list
(`@tui-series-the-tabs-count-what-they-hold`).

A page the configuration turns off has no tab, and `⇥` walks the others; the
board opens on the first left. With all three off the screen is empty and only
`ctrl+c` does anything (`@tui-a-page-off-has-no-tab`,
`@tui-every-page-off-is-an-empty-screen`; see
[config.md](config.md#the-pages-turned-off-one-by-one)).

**The backups, and what a command answers.** `b` backs the store up and `⇧B`
lists the backups, as the window's `⌘B` and `⇧B` do. A command typed at `:`
used to run and say nothing of what it answered, so `:timesheet show` or
`:backup list` showed nothing at all; what it answers is now said on a screen
of its own, scrolled with `j k`, until `esc` (`@tui-a-command-says-its-answer`).
A gesture on the board says nothing: what it did is on the board.

## The board's shape

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
 ┌──────────────────────┐ ┌──────────────────────┐                          ┌──────────────────────┐
 │ 18                   │ │ 33                   │                          │ 11                   │
 │ Read the ADR         │ │ Fix the hook         │                          │ Answer over MCP      │
 │                      │ │                      │                          │                      │
 │                      │ │                      │                          │                      │
 └──────────────────────┘ └──────────────────────┘                          └──────────────────────┘
 ┃──────────────────────┐ STALLED                                                                  6
 ┃ 21                   │ ──────────────────────────────────────────────────────────────────────────
 ┃ Ask legal about the  │ ┃──────────────────────┐ ┌──────────────────────┐ ┌──────────────────────┐
 ┃ retention clause     │ ┃ 19                 ⊘ │ │ 24               ≡ ◉ │ │ 30                 ‖ │
 ┃ −3d                  │ ┃ Client's answer      │ │ Handed to Marie      │ │ Refactor migrate     │
 ┃──────────────────────┘ ┃                      │ │                      │ │                      │
                          ┃ ↳ 12                 │ │ ↳ she confirms first │ │ ↳ your call          │
                          ┃──────────────────────┘ └──────────────────────┘ └──────────────────────┘

 hjkl ←↓↑→ columns   i insert   a new   / search   ⇥ tab   ? help   : command
```

**The header opens with the workspace the board shows**, `● work`, before the
tabs (`@tui-workspace-the-header-names-it`). A store no workspace names — given by
`--store`, `CAIRN_STORE` or the file — shows its folder instead, faint, after
`◌`, cut at its start when it is long so that its end is kept
(`@tui-workspace-a-loose-store-shows-its-folder`). Which workspace it is comes
from `workspace show`, asked once when the board opens and once each time
another is opened; nothing works it out a second time. The dot wears the
workspace's colour, the one its settings say or one read off its name
(`@tui-workspace-the-header-wears-its-colour`,
[config.md](config.md#the-colour-a-workspace-wears)); its name is in the
accent.

Four columns side by side, left to right: **backlog, todo, current, done**. The
backlog runs the full height, because it is the heap everything lands in and it
holds more than the rest put together.

A horizontal band underneath the other three holds the three states that mean
"not moving" — `blocked`, `watching`, `paused`. It takes three cards across,
exactly the width of the columns above it, and scrolls past that.

The band is called **STALLED** and not "waiting", because a state inside it is
called `watching` and two seven-letter words that contain one another get read
for each other. That mistake was already paid for once in this repository, on
`versions/` and `_archive/`, in the room where both had just been named.

**An empty column stays on screen.** A `current` with nothing in it takes a
quarter of the width to say that nothing is being worked on, and that is worth a
quarter of the width.

**A dim rule stands between each two columns**, in the cell the gap already
was, so no column lost a cell to it, and it stops where the band starts rather
than cutting it (`@tui-columns-stand-apart`). A box round each column would take
two cells a column, a wider gap a cell more, and a terminal has no subtle tint
to set one column off from the next; a single line in the faint colour is the
cheapest edge there is. The objective's own board is drawn by the same code and
has the same rules.

## Width

Everything is worked out from the width the terminal reports, and reported again
on every resize. Nothing is written in.

```
column        = (W − 1 − 3) / 4   one left margin, three gaps between columns
card interior = column − 4        two borders, two paddings
```

| terminal | card interior |
|---|---|
| 120 | 25 |
| 100 | 20 |
| 80 | 15 |
| 60 | 10 |

**Sixty columns by twenty-six rows is the floor**, and they are the two numbers
written in rather than worked out. Below either, the board draws two lines and
nothing else:

```
 cairn needs a bigger window.
 60 by 26 at least. This one is 47 by 20.
```

**The band always has a share of the height.** What is not moving is the thing
most worth seeing, and a band pushed off the bottom of the screen might as well
not be there — so the height is split before anything is drawn: **two rows of
cards** to the band when the window can spare them, half of what is left when it
cannot, and never none. Two rather than one, because three cards across is a row
and one row makes the band a strip that has to be scrolled to be read at all.

**Every screen fits its window.** A screen one row too tall pushes the tabs off
the top and the bar off the bottom, and neither comes back — so the help, the
form and the objectives are each held to what the window has, and each scrolls
inside it.

It is checked on the fixtures the binary carries, every case a board draws, at
80 by 26 and 120 by 40 (`@tui-the-fixtures-fit`): the board, a task's detail and
form, the help, the objectives and an objective's detail, and the time — no line
wider than the terminal, no screen taller. 80 by 24, the classic terminal, is
under the 26 rows the board needs, so it is the "bigger window" message there,
not a board.

At sixty it is tight but true: a title has twenty characters across its two
lines, labels drop off the facts line, and the three stalled states are glyphs
rather than words — which is why they are glyphs everywhere and not only here.

## A card is six lines

Always six: a border, four lines of content, a border.

| line | what is on it |
|---|---|
| 1 | the number on the left, the marks on the right |
| 2–3 | the title, cut at a word, `…` if it overruns; on an occurrence of a series, line 3, when the title leaves it free, holds its rule in faint ink, `every weekday 10:30` |
| 4 | the facts, in this order: deadline — on an occurrence, its day instead — steps, labels, the objectives it serves by name (`◇ Firmware`) — cut when the line is full, with `+n` saying how many were left out |

The height is a budget, not a maximum. A title longer than two lines is cut.

A terminal has no reliable pointer to hover with, so a cut title is read another way: in insert mode, when the card under the cursor cuts its title, the board writes the whole title on the line just above the bar, in the row that is otherwise blank, so the board does not grow. A title the card holds whole adds nothing. A title wider than the window is cut at its width with `…`; the line never wraps. Tags: `@tui-a-cut-title-is-read-whole-in-insert`, `@tui-a-title-that-fits-is-not-said-twice`.

A fixed height was chosen over a snug one, and it bought something that was not
free before: **navigation becomes arithmetic.** A card is six rows, so `j` and
`k` know where they land, and how far a column has to scroll for the one under
the cursor to be wholly on screen is a division rather than a measurement. What
it costs is two empty lines on a card carrying nothing.

**Everything that can be longer than its window scrolls**: a column, the band,
the objectives, a task's trace. Each is drawn behind a viewport that follows
the cursor, by whole cards — a column that scrolled by lines would leave half a
card at the top, and half a card says its number and nothing else.

A full line used to drop whatever did not fit and say nothing, so a task
carrying four labels read as a task carrying one. It now drops one more fact to
make room for `+3`: a card that says less than it holds is better than a card
that hides how much it is hiding.

A stalled card's fourth line says what is in the way — or, once every task it
waits on is done, that **nothing blocks it any more**, in the bright ink and in
place of the list. That is the one thing about blocking that has to be acted on,
and it reached the command line's listing and nowhere else.

The marks on line 1 are right-aligned in a fixed order: `▭` for a card off the
board first, then `↻` for an occurrence of a series, then the description mark, and the state glyph last. The state anchors the far right so the eye finds it
in the same place on every card in the band. Both fit alongside a four-digit
number in ten characters, which is the real ceiling — numbers only ever go up.

## Colour carries what a glyph would cost width for

**A card's border is thick.** A hairline round four sides of the thing the eye
lands on disappears against everything else on screen. The one in focus is
double, which reads as heavier still without being the same weight twice.

**Priority is the colour of the left border, and nothing else.** It costs no
character, it reads as a stripe down a column, and it does not fight the focus,
which changes the border's *style* rather than its colour.

| | |
|---|---|
| `low` | cyan |
| `normal` | the terminal's own foreground |
| `high` | yellow |
| `critical` | red |

Colouring the whole border was turned down: three `high` cards in a row turn a
column into a yellow wall and nothing stands out any more.

**The task in focus has a double border**, and it is the only one on screen. A
dot beside the number was drawn and removed — the frame already says it, better,
and the dot was eating the most contested line on the card.

**A deadline is bare** — `2d`, `−3d` in red when it has passed — and it only
appears when it is close or gone. A date three months out is not information on a
board. It carries no clock glyph: the glyph rendered badly and said nothing the
`d` did not.

**A task that is done counts no day down.** It is wanted on no day any more,
and a `−3d` in red on it would call for an act nobody has left to do. The open
task still says the day it was due, without the count. The window reads the
same rule, `engine.CardDeadline`.

## The terminal's own colours

**The board draws in the terminal's sixteen colours and nothing else**, so it
wears whatever theme the terminal wears
(`@tui-colours-follow-the-terminal`). The 256-colour indexes it once used
(39, 203, 214, 245…) are fixed in most terminals: a person who themed theirs
Gruvbox got cairn's blues and pinks on a Gruvbox ground. The sixteen are the
ones every terminal scheme sets — base16's shell templates write `base08`
into red, `base0D` into blue — so the board follows the theme for nothing.

| what | drawn in |
|---|---|
| the accent — the tab open, the crumb, a key, a title, the chosen card | blue (4) |
| `low`, `watching` | cyan (6) |
| `high`, the STALLED band | yellow (3) |
| `critical`, `blocked`, a deadline passed, what is refused | red (1) |
| a `reached` objective | green (2) |
| `normal`, `paused`, `on-hold`, `dropped`, what is said quietly, the rules | faint, on the terminal's own foreground |
| a title, a heading, `done` | the terminal's own foreground |
| the tab open, the mode, the choice picked, the button in focus | reversed: the colour becomes the ground, the terminal's ground the ink |
| the task in focus | a double border — a shape, not a colour |

**Faint is an attribute, never "bright black"**
(`@tui-colours-faint-is-the-terminals-own-foreground`). In Solarized, bright
black is `base03`, the background itself; text drawn in it disappears. Faint
dims the terminal's own foreground, whatever it is, so it reads on a light
ground as on a dark one. Black, white and their bright twins are not used at
all: themes spend them on grounds as often as on inks.

**A border carries no faint.** A terminal cannot dim a border lipgloss draws,
so a card's frame and a panel's are in the terminal's own foreground, and a
`normal` card's left edge with them — the colours of `low`, `high` and
`critical` stand out against it, which is what the edge is for
(`@tui-colours-a-priority-is-one-of-the-terminals`). The rules between the
columns are text, and are faint.

**A label's colour is the nearest the terminal has.** The fourteen label
colours are Catppuccin's, made for a window that paints its own ground; the
terminal has six accents and their bright twins. The workspace's dot wears
the nearest (`@tui-colours-a-workspace-wears-the-nearest-of-the-terminals`),
and two colours close in Catppuccin may meet in one:

| label colour | the terminal's |
|---|---|
| `red` | red (1) |
| `flamingo`, `maroon` | bright red (9) |
| `green` | green (2) |
| `yellow` | yellow (3) |
| `peach` | bright yellow (11) |
| `blue` | blue (4) |
| `sapphire`, `lavender` | bright blue (12) |
| `pink` | magenta (5) |
| `rosewater`, `mauve` | bright magenta (13) |
| `teal` | cyan (6) |
| `sky` | bright cyan (14) |

In a base16 terminal the bright twins are often the plain six again, and a
pair of the table then wears one colour: the label's name is still beside
it.

**The window's theme does not reach the terminal.** The scheme a window
wears ([gui.md](gui.md), [config.md](config.md#the-colours-the-window-wears))
is the window's; a terminal app that repainted its ground in another scheme
would fight the terminal's own.

**Nothing else can come back.** A test reads every file of `internal/tui`
and refuses a colour written in hex, `lipgloss.Color`, `ANSIColor`, a colour
made by hand, and black, white and their bright twins; another reads every
pinned frame and the fixtures' every screen and refuses a colour of 256 or of
true colour, and bright black. What the form borrows from huh keeps huh's
white prompt and line numbers on the fields not in focus: white is one of
the sixteen, and huh does not hand them over.

## Glyphs

Nerd Font by default, plain Unicode when the terminal cannot. Which one is said
rather than guessed at, because detecting a Nerd Font is not reliable enough:

```
cairn tui --icons unicode
```

Two answers and no others; anything else is refused by name, and named at the
board rather than at the command line — which sets exist is the board's to know,
and a second list at the surface would be a second list to keep true.

A resting place for this in a configuration file is a separate matter, and it
waits on the configuration file itself.

Every code point was read off `glyphnames.json`, the Nerd Fonts project's own
table, rather than recalled — and they all sit in the octicon and material
ranges, which are one cell wide. A two-cell glyph would slide every card's right
edge out of line.

| meaning | Nerd Font | Unicode | colour |
|---|---|---|---|
| `blocked` — waiting on somebody else | `nf-oct-blocked` | `⊘` | red |
| `watching` — handed over, needs checking | `nf-oct-eye` | `◉` | cyan |
| `paused` — waiting on your own decision | `nf-md-pause` | `‖` | faint |
| `active` objective | `nf-oct-play` | `▶` | blue |
| `on-hold` objective | `nf-md-pause` | `‖` | faint |
| `reached` objective | `nf-oct-check` | `✓` | green |
| `dropped` objective | `nf-oct-x` | `✕` | faint |
| a description exists | `nf-oct-note` | `≡` | |
| steps done / total | `nf-oct-checklist` | `▪` | |
| a label | `nf-oct-tag` | `⚑` | |
| the objective a task serves | `◇` | `◇` | |
| what is in the way | `nf-oct-blocked` | `↳` | |
| archived | `nf-oct-archive` | `▭` | |
| it comes back, an occurrence of a series | `nf-md-repeat` | `↻` | |

`‖` does double duty — `paused` on a task, `on-hold` on an objective. It is the
same idea in two tabs, so it kept one symbol rather than inventing a second for
one meaning.

## The modes

| mode | the cursor is on | where |
|---|---|---|
| **normal** | a column | tasks, an objective's board |
| **insert** | a task | tasks, an objective's board |
| **edit** | a field | the form |
| **normal** | an objective | the objectives list, which has this mode only |

The objectives list has no insert mode: normal already moves between objectives,
so it plays the part insert plays on the board. Two modes on one tab, one on the
other, and that is a consequence of the tabs holding different shapes rather than
an inconsistency to file down.

**The bar at the bottom only shows what works in the view it is in.** If `⇧hjkl`
is on the bar, it does something here. If it is not, it does nothing here. The
bar is cut to the width of the screen, dropping whole entries from the end
rather than wrapping — a wrapped bar pushes the board up out of the window.

**In normal, the column under the cursor says so**: a bar in front of its name
and a heavier rule beneath it. Nothing in the column is picked out, because
nothing in it is held.

## Every key

**`hjkl` and the arrows only ever navigate.** Everywhere, in every mode, without
exception. That rule is why no field has a letter of its own — see the form.
On an open task they navigate inside it, and never out of it: `jk` `↓↑` walk
its steps, `hl` `←→` do nothing, and `esc` is the way back to the board.

Every key below says which mode it works in. A key absent from a mode is absent
from that mode's bar, and pressing it there does nothing.

| | normal | insert | objectives list | |
|---|:---:|:---:|:---:|---|
| `hjkl` `←↓↑→` | ✓ | ✓ | ✓ | navigate — columns, tasks, objectives |
| `i` | ✓ | | ✓ | enter insert · on objectives, open the objective |
| `⇧hjkl` `⇧←↓↑→` | | ✓ | | move the task anywhere; its state follows where it lands |
| `a` | ✓ | | ✓ | add — the column under the cursor gives the state |
| `o` | | ✓ | | a new task linked to this one, in `todo` — last in the bar, the first left out when the bar is narrow; see [links.md](links.md) |
| `x` | | ✓ | | archive or unarchive the task |
| `⇧A` | ✓ | ✓ | | show what was archived, or hide it again |
| `⇧X` | ✓ | | | archive every task in the column — it asks first |
| `e` | | ✓ | ✓ | the form |
| `D` | | ✓ | ✓ | delete — it asks first, `⏎` agrees and anything else backs out. It was `d`, one slip from losing a task; the window deletes with `⌘⌫`, which a terminal is not sent, so here it is the shifted letter. An objective holding sub-objectives or tasks asks what goes with it — [below](#deleting-an-objective) |
| `t` | | | ✓ | its tasks, on the tasks board |
| `N` | | | ✓ | a new sub-objective, under the objective the cursor is on, or beside the sub-objective it is on |
| `z` | | | ✓ | fold a parent's sub-objectives away, or show them again |
| `v` | ✓ | | | the views — a figure opens one, `0` the whole board, `esc` closes them |
| `[` `]` | ✓ | ✓ | | on an objective's board, or on the tasks an objective shows, narrow to one sub-objective, the objective itself, or all of it |
| `m` then a letter | | ✓ | | move the task to the column that letter begins — `b` backlog, `t` todo, `c` current, `d` done, `w` watching, `p` paused — as the window's `m` does |
| `p` then a letter | | ✓ | | its priority — `l` low, `n` normal, `h` high, `c` critical |
| `r` then a letter | | ✓ | | on an occurrence of a series, and in its detail: `s` skip it, `x` stop the series, `r` resume it, `e` open the series in its tab — [below](#series) |
| `f` | | ✓ | | focus or unfocus — an objective has none |
| `⏎` | | ✓ | ✓ | the detail · on objectives, open the objective |
| `/` | ✓ | | ✓ | search |
| `?` | ✓ | ✓ | ✓ | help — the one key that works in every mode |
| `:` | ✓ | | ✓ | command |
| `⇥` | ✓ | | ✓ | next tab |
| `gg` `G` | | ✓ | ✓ | the first / the last of what you are in |
| `^O` | ✓ | ✓ | ✓ | another workspace, in place of this one — see [below](#another-workspace-in-place-of-this-one) |
| `esc` | ✓ | ✓ | | back one level |

On the board itself `esc` widens a search that is still in force, and does
nothing once there is none: there is nothing above it to back up to, and leaving
is `:q`. A search accepted with `⏎` outlives the prompt that typed it, and
nothing else on the board says so — so `esc` is the way back to the whole board. Inside an objective's own board there is — it goes back to
the list of objectives, with the cursor still on the one you were in.

**Archiving a whole column** is `⇧X`, in normal mode, on the column
the cursor is on. It asks first and names the column, because a key that moves
twelve cards is not a key to press and find out about. The engine has swept by a
narrowing since the beginning; only the command line ever asked it to.

`?` is the one key with no mode. Help that cannot be reached from the mode you
are stuck in is help you have to already understand to find, so it answers
everywhere — in insert, in the form, in the detail, and while searching. Asking
it again closes it and puts back whatever it was asked over, form included: help
must not be a way of losing work by asking a question.

Inside an objective's board the columns are the tasks tab's columns, so the two
first rows apply there unchanged — with `esc` added, which goes back to the
objectives list.

### Moving a task

`⇧hjkl` moves the task itself, in insert mode, where a task is under the cursor.
One gesture, four directions, across columns and in and out of the band.

`⇧h` and `⇧l` change the column, and the state follows. **`⇧j` and `⇧k` change
the task's place among the ones it sits with**, and never its column — one act
with two meanings is an act nobody can undo confidently, which is why the engine
keeps a move and an edit apart and this respects it.

The one exception is the last task in a column: there is nowhere lower in the
column, and what is below the column is the band. That is the only time `⇧j`
asks which kind of stopped the task has become.

Plain `j` off the bottom of a column carries the *cursor* on into the band, and
plain `k` off the top of the band brings it back to the column it came down
from. A column that stopped at its last card would leave what is stalled
unreachable whenever a column happened to be short.

**The band is walked the way it is drawn.** Three cards to a row, so `h` and `l`
go along a row and `j` and `k` go a row at a time. A key that pointed down and
moved the cursor sideways would be a key nobody could trust.

**Walking sideways keeps the level.** The third task in one column and the third
in the next are the same journey, not two: a cursor that jumped back to the top
of every column it entered would make reading across a board impossible. A
column with fewer in it stops at its last.

**`gg` and `G`** reach the first and the last of whatever the cursor is standing
in — a column, the band, the objectives. They are vim's, and anybody who has
used vim will try them before reading this. A `g` that leads nowhere is
forgotten on the next key rather than lying in wait, and in a form or at a
prompt a `g` is a letter.

- Pushing a task **into** the band opens a popup asking which of the three states
  it lands in. Merely navigating into the band asks nothing — you must be able to
  look at what is stalled without answering a question.
- **Inside the band, `⇧J` and `⇧K` change a task's place among the tasks of its
  own kind**, and never its kind (`@tui-moving-a-task-down-among-its-kind-in-the-band`,
  `@tui-moving-a-task-up-among-its-kind-in-the-band`). The band reads its three
  kinds one after another, and a place is counted within a state, so a step
  past a task of another kind would land nowhere the engine keeps. The last of
  its kind goes no lower (`@tui-the-last-of-its-kind-in-the-band-stays`): below
  it is another kind, and changing the kind is the popup's answer, not a push.
- `⇧K` from the first of its kind takes it out of the band, to **the column the cursor came from**:
  todo → stalled → todo. That memory lives on the cursor, not on the task, so it
  dies with the session.
- `⇧H` from the leftmost card in the band sends it to the **backlog**, even
  though the band does not run under the backlog.

Reordering inside a column is `⇧J` and `⇧K` — `task move`, behind the card
below or at the head of the column — with the band reached only from the last
card of a column, as above, and the same keys reordering inside the band among
a task's own kind. Where a pushed card goes is worked out once, in
the engine, for this board and the window alike.

## The form

One form, for adding and for editing, behaving the same way both times.

```
 ‹ TASKS ›  ▸ 14   edit                                                    ── EDIT ──

 ┌────────────────────────────────────────────────────────────────────────────┐
 │  ▸ title          Write the spec                                           │
 │    description    ≡ 3 lines                            ⏎ opens $EDITOR     │
 │    column         todo                                                     │
 │    priority       high                                                     │
 │    labels         spec  infra                                              │
 │    blockers       —                                                        │
 │    links          —                                                        │
 │    because        —                                                        │
 │    deadline       2026-09-02                                               │
 │    estimate       3h                                                       │
 │    time spent     1h30                                                     │
 │    objectives     3                                                        │
 │    steps          [x] Draft the columns                                    │
 │                                                                            │
 │                                             [ save ]          [ cancel ]   │
 └────────────────────────────────────────────────────────────────────────────┘

 hjkl ←↓↑→ navigate   ⏎ change   esc cancel
```

**Nothing is written until `save`.** Writing each field as it was left was tried
on paper and dropped: `cancel` would have meant nothing, and the same form could
not have served to create something that does not exist yet.

**A form is filled in the window's way.** `⏎` writes a line in the
description and moves on everywhere else, and never saves; `↑↓` walk the
description's lines and `⇥` leaves it; `^S` saves from anywhere. huh had enter
submit the description and down leave it, so a description could neither be
written in paragraphs nor read through.

The description carries its line numbers down the side and its length beside
its name — `description · 25 lines` — which is the half of a bar's job a
terminal box can do: huh draws the box and offers no bar.

**What was typed is what is saved.** The form handed the engine the values it
had opened with and only the title as it had been typed, so every other change
made on the board was dropped on save without a word, since the board first had
a form.

**The description on an open task is a box that scrolls**, with a bar down its
right — `┃` where the window is, `│` the rest of the track — and `^D` `^U` walk
it half a box at a time. It used to be drawn whole, and a long one ran off the
bottom of the terminal with everything under it.

**On an open task the steps left come first**, each with `[ ]` before it and
its place, and the steps done after, `[x]`, faint and struck through; the
panel's name counts them, `STEPS   1 to do · 1 done`
(`@tui-the-steps-left-come-first`). The window draws the same thing with a
box; the marks are the ones the form's field writes.

**A step of an open task is chosen with `jk` `↓↑` and deleted with a delete
key.** `›` marks the one chosen, in the order the steps are shown; `⌫` or the
forward delete asks at the foot, "delete step 1 of task 1, “Draft”?", `⏎`
deletes it at once through a `step delete` of its own, and `esc` keeps it
(`@tui-a-step-is-deleted-once-agreed`,
`@tui-a-step-is-kept-when-its-delete-is-cancelled`). Up and down used to walk
to the task above and below from an open task, which was pointless inside a
task.

**An open task follows what is done to it elsewhere.** The open task is drawn
from each reading the beat brings (`@tui-an-open-task-shows-a-step-added-elsewhere`),
and so is its form: a reading closes the form only when it is the one its own
save asked for. Any other reading rebuilds it, keeping a field the person
changed as they typed it and taking every other one from the store, with the
keyboard left on the field it was on
(`@tui-a-form-keeps-what-was-typed-when-its-task-changes-elsewhere`). A reading
used to close the form whatever brought it, so a change made by an agent threw
away what was being typed.

**The steps of a task are written in its form**, in the `steps` field, one line
each — `[x] Relire l'ADR`. Saving writes them
with the task, in the same act, and they replace the steps it had: a line the
field no longer carries is a step that has gone. The window does the same
thing in the same words. A step is no card and has no column, so nothing on
the board stands for it (`@tui-a-step-is-not-on-the-board`).

A step carries a title and whether it is done, and nothing else; see
[steps.md](steps.md). Deleting a task takes its steps with it.

**Pushing a task into done while a step is open asks first** — "put task 1 in
done, ticking its steps?" at the foot — and `⏎` puts it there with its steps
ticked in the same act (`@tui-done-asks-about-open-steps`).

**A store the board does not match is asked about, as the window does.**
`cairn tui` opens on a store behind (`@store-behind-opens-the-board`), and a
store rolled back under an open board is noticed at its next look
(`@tui-a-store-put-behind-while-open-is-noticed`): the foot asks "apply what
this store is waiting for?", `⏎` applies it, a backup set aside first
(`@tui-a-store-behind-is-offered-its-changes`), and `esc` leaves the board
with the store as it was (`@tui-a-store-behind-is-left-as-it-was-when-quit`).
An apply that fails is said, and asked again. A store a newer cairn changed
says "this store comes from a newer cairn", and only `esc` does anything
(`@tui-a-store-from-a-newer-cairn-stops-the-board`). No other key is reached
while either stands. The changes waiting are not named on the foot: naming
them would read the store inside the board's update; `cairn migration list`
names them. Why: [gui.md](gui.md).

**A form somebody typed into is asked about before it goes.** `esc` on an
untouched form closes it; `esc` on one with a change in it puts the question at
the foot, and `esc` again gives the form back with everything still in it. The
board guards a deletion of two words, and used to throw ten minutes of typing
away on one key.

**No field has a key of its own.** A key per field was designed and thrown away —
`hjkl` had already taken four letters and `d` was delete, so `deadline` and
`labels` ended up on `t` and `#`, letters chosen by constraint rather than by
sense. Letters like that are not remembered. Navigating to a field is better, and
it keeps the rule that `hjkl` only navigates.

`⏎` on a field enters it. A field with a closed set of values opens a list, and
`jk` moves in that list — the same gesture again. `h` and `l` earn their keep on
the last row, where `save` and `cancel` sit side by side.

**The description opens `$EDITOR`.** Nobody writes twelve lines of prose into a
terminal field. The form says it exists and how many lines it runs to.

Adding opens the same form empty, with the column already filled in from where
the cursor was, and the title the only thing `save` insists on.

## The detail

`⏎` on a task takes the whole screen. There is enough to show that a panel would
have squeezed the board and a popup would never have shown the whole trace.

```
 ‹ TASKS ›  ▸ 14                                                          ── DETAIL ──

 ┃ 14   Write the spec                                       ≡    todo    high
 ┃ ⚑ spec  infra        due 02/09 · in 2d               3h estimated · 1h30 spent

 ┌─ DESCRIPTION ───────────────────────────────┬─ TRACE ──────────────────────────────┐
 │ The client wants two columns, not three.    │ 12/08  created    backlog · normal   │
 │ The third was in the mockup by accident     │ 18/08  priority   normal → high      │
 │ and nobody caught it before the review.     │ 21/08  column     backlog → todo     │
 │                                             │ 27/08  deadline   — → 02/09          │
 ├─ STEPS   2 to do · 3 done ──────────────────┤ 29/08  focus      taken              │
 │ › [ ] 4  Client's answer                    │ 29/08  focus      released           │
 │   [ ] 5  Ship it                            │ 30/08  spent      — → 1h30           │
 │   [x] 1  Draft the columns                  │                                      │
 │   [x] 2  Get the client's nod               │                                      │
 │   [x] 3  Write the scenarios                │                                      │
 ├─ OBJECTIVES SERVED ─────────────────────────┤                                      │
 │ ▶ 3    Ship the desktop board     7/12      │                                      │
 └─────────────────────────────────────────────┴──────────────────────────────────────┘

 jk ↓↑ steps   e edit   f focus   ⌫ delete the step   ^D ^U description   ? help   esc back
```

An occurrence of a series says so on a line of the band of its own:
`↻1 every weekday at 10:30, on the calendar · 46 done · 3 skipped · 2 missed ·
next Tue 10 Nov` — how often, what it counts from, what the series did so
far, and the day after its own (`@tui-series-the-detail-says-the-rule`). The
bar then offers the `r` keys.

Under the lists, a `CONTEXT` panel says where what was written about the task
elsewhere lives — the directory a harness writes its dated notes into, said
whether or not anything is there yet. See [context.md](context.md).

The band carries the short fields; the panels carry the lists. **A field that is
empty is not shown** — the opposite of a card, where an absence keeps its line to
hold the height. A full screen has no such constraint.

This is the only place the trace is laid out beside the task it belongs to.
`:trace show 14` prints the same trace, the way the command line prints it, and
neither is the other's shortcut — one is read while looking at the task, the
other while looking for something. The trace had been in the engine a while with
nowhere to be looked at.

## Searching

```
 ‹ TASKS ›  objectives                                                   ── SEARCH ──

 BACKLOG                0 TODO                   1 CURRENT                1 DONE                   0
 ──────────────────────── ──────────────────────── ──────────────────────── ────────────────────────
                          ┃──────────────────────┐ ╔══════════════════════╗
                          ┃ 14                 ≡ │ ║ 31                 ≡ ║
                          ┃ Write the spec       │ ║ Spec review          ║
                          ┃                      │ ║                      ║
                          ┃ 2d  ▪3/5  ⚑ spec     │ ║ 1d  ▪2/4  ⚑ spec     ║
                          ┃──────────────────────┘ ╚══════════════════════╝
                          STALLED                                                                  0
                          ──────────────────────────────────────────────────────────────────────────

 /spec ▏                                                                  2 of 51 tasks
```

`/` **filters the board in place.** The columns stay — empty ones included, for
the same reason they stay on the whole board — and show only what matches, so you
see at once where the matches sit in the flow. `esc` gives the whole board back,
whether the search is still being typed or has already been accepted with `⏎`.

**What a search reaches is what the engine reaches.** Every fact a task carries,
its labels, its column and its priority included, and every word typed has to be
carried — not one of them. The boards used to look in six fields with one word
where the engine looked everywhere with as many as were given, so one word,
"search", meant two things depending on where it was typed, and searching by
label was impossible on either board.

The counts in the headers are the counts of what matched, and the foot says how
much of the store that is. Typing narrows as you go.

**`⇧F` filters the board by what a task carries**: its column, its
priority, a label, an objective, a day it is due by — today, in a week, in a
month. The foot becomes the five filters, `←→` choosing one and `↑↓` walking its
values, each opening on `any`; the board narrows as they move, the way the
search narrows as it is typed, and `⌫` puts every one back on `any` at once
(`@tui-a-filter-narrows-the-board`, `@tui-every-filter-is-cleared-at-once`).
`esc` closes them and leaves them in force, said above the foot — `filtered by
priority high · label backend` (`@tui-the-board-says-what-filters-it`). A search
narrows what they kept. What a filter keeps is what `task list` keeps with the
same flag: the engine's `Filters` judges a card by the narrowing the listing is
read with, so the two cannot disagree.

**`v` turns the foot into the views** (#226,
[listing.md](listing.md#a-view-keeps-the-filters-under-a-name)), the way `⇧F`
turns it into the filters: each view after the figure of its place, with how
many tasks it keeps, then `0` for the whole board and `esc` (`VIEWS`,
`@tui-view-v-lists-the-views`). A figure opens the view in that place, `0`
leaves it, and `esc` closes the list and leaves what is in force
(`@tui-view-a-figure-opens-it`, `@tui-view-0-goes-back-to-the-whole-board`).
Bare figures on the board would be typed too easily, so they work inside the
list only, as the workspaces' do; `v` was free in every mode, and works in
normal on the tasks board. Opening a view clears the filters and the search,
so a view always opens on the board it was kept with.

The line above the foot names it — `view Sprint · 1 of 3 tasks` — and marks it
`*` when a filter or a search stands on top of it, followed by what filters it
(`@tui-view-save-puts-the-change-in`). Keeping is a command: `:view new
Urgent` keeps the filters and the search in force, and `:view new` alone
proposes a name from them, an objective by its title, so `t` then `:view new`
keeps an objective's tasks under its name (`@tui-view-new-keeps-the-filters`,
`@tui-view-new-proposes-the-objective`). `:view save` puts what stands on top
into the view in force, and `:view` and a name opens one, whatever its case
(`@tui-view-opened-by-its-name`). `:view list`, `show`, `edit` and `delete`
are the command line's, through the command line as every other command.

**What lies outside a view is a reference**: a task outside it that blocks,
waits on or is linked to one inside it is drawn once, at the foot of its
column, in `┄` and `┆`, its last line saying why in the late colour —
`blocks 3 · linked to 3` — and never counted
(`@tui-view-outside-is-a-reference`). The why takes the line the facts had,
on an objective's board too; it is words, not the mockup's `⊘ ↔ ⧗`, so no
glyph had to be checked for its width.

A flat list of results was turned down for losing the column each match is in,
which is most of what a board is for. Highlighting without hiding was turned down
because with two hundred tasks what is highlighted is usually off screen.

## Help, and the command line

```
 ‹ TASKS ›  help                                                            ── HELP ──

 ┌─ KEYS ──────────────────────────────┬─ LEGEND ──────────────────────────────────┐
 │                                     │                                           │
 │ NORMAL — cursor on a column         │ PRIORITY — the left border                │
 │   hjkl ←↓↑→  change column          │   ▏ cyan   low       ▏ yellow  high       │
 │   i          insert mode            │   ▏ own    normal    ▏ red     critical   │
 │   a          a new task             │                                           │
 │   /          search                 │ STALLED — top right of the card           │
 │   ⇥          next tab               │   ⊘  blocked    waiting on somebody else  │
 │   :          command                │   ◉  watching   handed over, needs a look │
 │                                     │   ‖  paused     waiting on your decision  │
 │ INSERT — cursor on a task           │                                           │
 │   hjkl ←↓↑→  change task            │ OBJECTIVES                                │
 │   ⇧hjkl      move the task          │   ▶  active     ✓  reached                │
 │   ⏎          the detail             │   ‖  on-hold    ✕  dropped                │
 │   e          the form               │                                           │
 │   f          focus or unfocus       │ ON THE CARD                               │
 │   d          delete                 │   ≡    a description exists               │
 │   esc        back to normal         │   ▪    steps done / total                 │
 │                                     │   ⚑    a label                            │
 │ EDIT — inside the form              │   ↳    what is in the way                 │
 │   hjkl ←↓↑→  navigate               │   2d   deadline close                     │
 │   ⏎          change                 │   −3d  deadline passed, in red            │
 │   esc        cancel                 │   ╔═╗  the task in focus                  │
 │                                     │                                           │
 └─────────────────────────────────────┴───────────────────────────────────────────┘

 esc close
```

`?` takes the whole screen: the keys on the left, grouped by mode, and the legend
on the right. It is eight lines of legend that get read three times and never
again, so it lives behind a key rather than on the board.

**The help is contextual too**, like the bar it explains. The screen above is the
tasks tab's; on the objectives tab the left panel carries that tab's one mode and
its keys, and nothing about insert. A help screen listing keys that do nothing
here is a help screen that has to be second-guessed.

```
 BACKLOG               12 TODO                   5 CURRENT                2 DONE      31
 ──────────────────────── ──────────────────────── ──────────────────────── ────────────
 ┌──────────────────────┐ ┌──────────────────────┐ ┌──────────────────────┐ ┌──────────┐
 │ 26                   │ │ 14                 ≡ │ │ 31                 ≡ │ │ 08       │
 ...                                     ( dimmed )

 ┌────────────────────────────────────────────────────────────────────────────┐
 │   backup new             take a backup now                                 │
 │   backup list            the backups taken                                 │
 │ ▸ backup show      <n>   what a backup holds                               │
 │   backup restore   <n>   put the store back in that state                  │
 │   backup delete    <n>   forget a backup                                   │
 └────────────────────────────────────────────────────────────────────────────┘
 :backup ▏

 jk choose   ⏎ run   esc cancel
```

`:` opens a prompt at the foot of the screen with completion above it, the board
dimmed behind. **The words are the command line's words**, so nothing here has to
be learnt twice.

| | |
|---|---|
| `:backup` | `new` `list` `show <n>` `restore <n>` `delete <n>` |
| `:migration` | `apply` `list` `show <n>` `rollback <n>` |
| `:trace` | `list`, `show <n>` |
| `:q` | quit |

The table above is what the prompt was planned to take. It now takes **every
operation the engine has**, read off `engine.All()` rather than written out
again: a thing, a verb, then the positional arguments in the order the operation
lists them and `--name value` for the rest. `:task edit 1 --state current` works
here exactly as it works at the command line, and an operation added to cairn is
understood here the day it is added. A prompt that knew two words out of thirty
made a bar offering `:` an offer of almost nothing.

Everything `task` and `objective` do also has a key. The prompt is the way in
for the arguments no key can carry, not a second way of doing what a key does.

```
 :set pages.ti▏
    pages.time           hidden         from global
    pages.tasks          true           default
```

**`:set` and `:unset` set a key of the configuration** without leaving the
board: `:set pages.time off`, `:set --global focus.session 45m`, `:unset
series.on`. They are not a second editor: they call `engine.ConfigSet` and
`engine.ConfigUnset`, the operations `cairn config set` and `unset` and the
window's settings call, so the same file is written the same way — one line,
every comment kept, and nothing written that does not read back
([config.md](config.md#one-configuration-several-ways-in-to-it)). Without
`--global`, the engine writes the settings of the store the board holds
when a workspace names it or settings are beside it. The board wears what was
written at once — a tab goes or comes, the week loses its weekend — and the
message line says only what the engine wrote, `pages.time = "off", written in
…/settings.toml`; the beat that follows says nothing more, since the board
stamped the file after its own write (`@tui-set-a-key-from-the-command-line`,
`@tui-set-a-key-in-the-global-configuration`, `@tui-unset-a-key`). A value the
key does not take is said in the reader's words and nothing is written
(`@tui-set-refuses-what-the-key-does-not-take`). They are typed rather than
read off `engine.All()`: they are a person's (`engine.ForAPersonOnly`), and
the prompt otherwise offers what an agent may do too.

**The keys are completed from the catalogue** (`config.Keys`): after `:set ` or
`:unset `, the lines above the prompt list the keys starting with what is
typed, five at most, each with its value in force and where it comes from —
`set here ← global: on`, `from global`, or `default`, as the window's rows say
it — and `⇥` completes the first of them (`@tui-set-completes-the-keys`). There
is no settings screen: it would draw a second editor of the same keys for a
board that has the window beside it, and the terminal does not wear the
window's theme, which it leaves to the terminal's own colours. That is to be
asked again if somebody uses the terminal with no window.

## Objectives

```
 tasks  ‹ OBJECTIVES ›  time                                              ── NORMAL ──

 OBJECTIVES                                                                         9
 ──────────────────────────────────────────────────────────────────────────────────────
 ┃ ▶  4  Home automation                                                       ⚑ area
 ┃    ███████████████░░░░░░░░░░░░░  5/15  14h50m  01/03/2027
   ├─ ▶  5  Enclosure                                                       ⚑ project
   │     ██████████████░░░░░░░░░░░░░░  3/6  8h25m
   └─ ‖  7  Features                                                        ⚑ project
         ░░░░░░░░░░░░░░░░░░░░░░░░░░░░  0/3  waits on Firmware
   ▶  1  Tidy the workshop
      ███████████░░░░░░░░░░░░░░░░░  2/5  1h50m
```

**An objective is two lines, in a tree** (`@tui-objective-sub-objectives-sit-under-their-parent`).
Six-line cards held three objectives to a screen, and a tree of a dozen needs
density more than room: the first line says its standing, its number, its
title and its type labels at the far right; the second its bar, its count, the
time spent, its deadline and, in the yellow of what is stalled, what it waits
on (`@tui-objective-a-row-says-what-it-waits-on`). A sub-objective sits under
its parent behind `├─` or `└─`, and the one under the cursor carries `┃` down
its left on both lines.

The bar, the count and the time are the **union of the subtree**, read from
the engine as `objective list` reads them: a task serving two sub-objectives
counts once in their parent (`@tui-objective-a-parent-counts-each-task-once`).
Beside the count, faint, the time spent on the tasks serving it, archived ones
included, as `objective show` does
(`@tui-an-objective-card-says-the-time-spent`).

The order comes from the state, active first, and each sub-objective follows its
parent; one whose parent is left out by a search stands at level one. `j` and
`k` walk the rows in that order, and the list scrolls two lines at a time.

**`z` folds a parent away**: its sub-objectives leave the list, and its first
line says how many it hides, `+3`; `z` again shows them
(`@tui-objective-z-folds-a-parent`). On a sub-objective it folds its parent,
and the cursor climbs to it, since it cannot stay on a row no longer drawn
(`@tui-objective-z-on-a-sub-objective-folds-its-parent`). A tree of a dozen
objectives is read one parent at a time; nothing is kept once the board
closes. **`N` adds a sub-objective**: the objective's form, its parent
already the one under the cursor — or, on a sub-objective, that one's parent,
since a third level is refused (`@tui-objective-shift-n-adds-a-sub-objective`,
`@tui-objective-shift-n-on-a-sub-objective-adds-beside-it`). `a` still adds
one of level one. The bar there walks with `jk ↓↑`, since `h` and `l` do
nothing on one column, and the room it gives back holds `N` and `z`.

It is not a kanban. An objective does not circulate — nobody pushes one from
`active` to `reached` in the morning, it gets there when its tasks are done.

**`i` or `⏎` opens the objective**, and what opens is the board, filtered to the
tasks that serve it or any of its sub-objectives, with the objective flattened
into a band on top. `esc` comes back out.

```
 tasks  ‹ OBJECTIVES ›  ▸ 4                                               ── NORMAL ──

 4   Home automation                         waits on #2 Garden   ▶ active
 ███████████████░░░░░░░░░░░░░  5/15   ‹ all 15 ›  Enclosure 6  Firmware 5  itself 2

 BACKLOG                2 TODO                   3 CURRENT                1 DONE                   7
 ...

 hjkl ←↓↑→ columns   i insert   [ ] sub-objective   a new here   / search   esc objectives
```

**The band carries the sub-objectives as chips**, each counted over what the
board holds, and `[` and `]` walk them: all of it, each sub-objective, then the
objective itself, round again
(`@tui-objective-the-band-carries-the-sub-objectives`,
`@tui-objective-brackets-walk-the-sub-objectives`). It is the terminal's
version of the window's segmented bar, and the counts come from the engine's
`Segments`. What the objective waits on is said on the band
(`@tui-objective-its-board-says-what-it-waits-on`).

**A card names what it serves**, `◇ Firmware`, on its facts line
(`@tui-objective-a-card-names-what-it-serves`) — the objective it is attached to
directly, whether of level one or a sub-objective. Inside an objective's board
that objective goes without saying.

**A task outside that stands in the way of one inside is kept in sight**, in its
column, as a reference: a dashed frame, `┆`, every ink faint, never counted
in a column's header and never reached by the cursor
(`@tui-objective-a-blocker-outside-is-a-reference`). What stops a task is half
of what it says, and narrowing the board used to hide it.

**`t` shows its tasks on the tasks board** (`@tui-objective-t-shows-its-tasks`):
the objective filter `⇧F` already had, put on the objective, and widened by the
engine to its sub-objectives. For a parent the filter offers both
`#4 Home automation, with its sub-objectives` and `#4 Home automation, itself`,
which leaves its sub-objectives out
(`@tui-objective-itself-leaves-its-sub-objectives-out`). The window's filter
reads the same values. On the tasks so shown, `[` and `]` walk the same
chips as an objective's board — all of it, each sub-objective, then the
objective itself — by putting the filter on each in turn, and the chips stand
after `filtered by …`, counted over the whole board
(`@tui-objective-brackets-walk-its-tasks-shown`). Each chip is the engine's
`Segment`, which names the filter's value it stands for, so the window's bar
and this one walk one list.

**The objective's form says where it sits.** `parent` takes a number, or
nothing to stand at level one, and says beneath it which are offered — those of
level one, itself left out (`@tui-objective-the-form-puts-it-under-a-parent`);
`waits on` takes numbers (`@tui-objective-the-form-says-what-it-waits-on`); and
`labels` says which labels are meant for objectives
(`@tui-objective-the-form-offers-labels-for-objectives`). A terminal field takes
what is typed, so a list is said under it rather than picked in it.

**A task's `objectives` field draws the trees it touches**: each objective of
the tree ticked or not, and a parent left unticked said to be counted through
the sub-objectives ticked below it
(`@tui-objective-the-task-form-draws-the-tree`). A task put on a parent and on
one of its sub-objectives is refused by the engine, and the refusal is said at
the foot of the form, which stays open
(`@tui-objective-the-task-form-refuses-a-double-count`).

**Work started early is said.** When a task goes current — focus, a push, a
move — and an objective it counts in waits on one not reached, the engine's
warning is said above the bar until the next key
(`@tui-objective-work-started-early-is-warned-about`). It never refuses.

An objective's state changes through the form, not by pushing it anywhere, and
`⇧hjkl` therefore does nothing on the objectives list — so it is not on the bar
there. `h` and `l` do nothing either, because the list is one column.

### Deleting an objective

```
 ┌─ DELETE #4 HOME AUTOMATION ──────────────────────────────────────────┐
 │                                                                      │
 │ › sub-objective #5 Enclosure      ‹ keep it, on its own ›            │
 │   sub-objective #6 Firmware       delete it                          │
 │   tasks #44 #47                   choose                             │
 │                                                                      │
 └──────────────────────────────────────────────────────────────────────┘
 what goes with objective 4?   jk choose   hl change   ⏎ delete   esc keep it
```

`D` on an objective that holds sub-objectives, or that tasks serve, opens a
box over the board: a row for each sub-objective — kept, standing on its own,
or deleted — and a row for the tasks touched, detached or deleted, which
appears once some are: those on the objective itself, and those of a
sub-objective deleted with it. **Every row opens on `choose`**, and `⏎` deletes
only once each is chosen; before, the foot says `choose for each first`
(`@tui-objective-deleting-a-parent-asks-what-goes-with-it`). `esc` keeps it
(`@tui-objective-a-delete-asked-about-is-kept-on-escape`). Which work goes with
an objective is never decided for the person; the engine refuses a deletion
that leaves something undecided, and the box only asks it the way the command
line's `--keep`, `--delete` and `--tasks` do. An objective holding nothing is
asked about at the foot, as a task is.

## Series

```
 tasks 38   objectives 5   time 3   SERIES 4                                 NORMAL

 Active                                2 │↻1  Team stand-up                         active
 ┃ ↻1  Team stand-up                today│every weekday at 10:30, on the calendar, from…
 ┃     every weekday 10:30 · 46 done · 3…│NEXT  Tue 10 Nov · Wed 11 Nov · Thu 12 Nov
   ↻2  Retrospective              Thu 12 │
       every 2 weeks on thu 14:00 · 31 d…│TEMPLATE — what the next one is made from
 Stopped                               1 │  Team stand-up · normal · work · 3 steps · 15m
   ↻3  Monthly exp… stopped since 12 Oct │SO FAR
       every 4 weeks on mon · 10 done ·… │  46 done · 3 skipped · 2 missed · 11h20m
                                         │ON THE BOARD NOW  #312
                                         │
                                         │OCCURRENCES · 52  archived ones included
                                         │› Tue 10 Nov  #312   todo
                                         │  Fri 6 Nov   #309   done              16m
                                         │  Thu 5 Nov   #305   skipped · archived
                                         │  Wed 4 Nov    —     missed

 jk ↓↑ series   ⏎ occurrences   e edit   s skip   x stop   ⌫ delete   n new series
```

A series is a notion of its own ([series.md](series.md)), and while
`[series] on` is true it has a tab of its own, the fourth, after the time
— `⇥` already reached the three others in that order, and a tab put between
them would have changed where every `⇥` lands. **The list is on the left, the
series chosen on the right**, as the objectives' tab and the window lay it
out (`@tui-series-the-tab-lists-them`): the active ones, then the stopped
ones, each in two rows — its number `↻1`, its title and its next day, or the
day it was stopped; then its rule as a row writes it, `every weekday ·
10:30`, and the counts that are not nothing
(`@tui-series-the-tab-says-it-as-drawn`). On the right, its
rule in words, its next days, the template the next occurrence is made from,
what it did so far, the task open now, and every occurrence it made or missed,
archived ones included.

`jk` walk the series; `⏎` goes into its occurrences, `jk` walk them, and `⏎`
opens the task of one in its detail — an archived one too, which no board
shows otherwise (`@tui-series-the-tab-reaches-an-archived-occurrence`). `esc`
comes back out. `s` skips the occurrence open, `x` stops the series and `r`
resumes it (`@tui-series-the-tab-skips-stops-and-resumes`); `e` opens its form
— title, priority, labels, estimate, objectives, steps, its rule as `repeats`
and the day it counts from — and saving changes what the next ones are made
from, never one already made (`@tui-series-the-form-changes-the-series`). `n`
opens a task's form in todo with the cursor on its `rule` field
(`@tui-series-n-lands-on-the-rule`), and saving it makes the task the first
occurrence of a new series (`@tui-series-n-makes-a-new-one`). `/` narrows the list by its title, its rule
or its number, as `series list --search` does.

**Deleting asks what becomes of its tasks, as the command line does.** `⌫` or
`D` opens a box over the tab with the three choices of `series delete`, the
first chosen: keep every task as an ordinary task, delete the one open and keep
the rest, or delete them all (`@tui-series-deleting-keeps-the-tasks-unless-told`,
`@tui-series-deleting-the-open-one`). Deleting them all asks once more, naming
how many (`@tui-series-deleting-every-task-asks-once-more`). Stopping a series,
or turning recurrence off, never takes a task away.

**On the tasks board, `r` is a prefix**, as `m` and `p` are: on an occurrence,
in insert or in its detail, `r s` skips it, `r x` stops its series, `r r`
resumes it, and `r e` opens the series in its tab (`@tui-series-r-s-skips`,
`@tui-series-r-x-stops-and-r-r-resumes`, `@tui-series-r-e-opens-its-series`).
The bar offers them only while the card under the cursor is an occurrence
(`@tui-series-r-is-offered-only-on-a-recurring-card`). `r` was free on both
boards.

**A card that comes back is told at a glance** (marker A of the study):
`↻` among its marks and its rule on the line its title leaves free
(`@tui-series-a-card-carries-the-loop-and-the-rule`), and on its facts line the
day it was made for — the time on its day, in the yellow of what is stalled
once it has passed; its weekday and date within a week; counted down in red once
passed; nothing further off, as the study's answers chose. The engine says
which, `engine.OccurrenceDay`, so the window says the same. Its detail's
trace says where it came from: `made from series  ↻1`, the act the engine
writes on a task its series made (`@tui-series-the-trace-says-made-from-series`),
after what its series did before it, faint under *in the series, before this
one* (`@tui-series-the-trace-goes-on-with-the-series-before-it`). The filters
end with `repeats`, while there is a series
(`@tui-series-the-board-narrows-to-what-comes-back`).

**A task becomes a series in its form.** A task that is no occurrence carries
a `rule` field: a rule typed there makes it the first occurrence of a new
series (`@tui-series-a-task-given-a-rule-becomes-the-first`). An occurrence has
no such field, its series' rule is changed in the series' form. **Saving a
change to an occurrence's title, priority, labels, estimate, steps or
objectives asks once** — `change the series too?` — `⏎` for this one only, the
default, and `s` for the whole series, the ones to come
(`@tui-series-saving-an-occurrence-asks-this-one-or-all`,
`@tui-series-saving-this-one-only-leaves-the-series`). A change of column, time
or deadline asks nothing: they are this one's alone.

**With recurrence off there is none of it**: no tab, no `↻`, no rule line, no
`rule` field and no `r` key; the cards are ordinary tasks
(`@tui-series-off-shows-nothing`). The series tab is held to the window as
every screen is, on the fixtures (`@tui-series-the-fixtures-fit`).

## Another workspace, in place of this one

```
 ● work  TASKS  objectives  time                                     ── WORKSPACES ──

 BACKLOG ┌─ OPEN A WORKSPACE ─────────────────────────────────────────────────────┐
 ┃ 160 R │ › ▏                                                                    │
 ┃ 158 R │                                                                        │
         │ WORKSPACES                                                             │
         │ › ● client-a  /Users/me/cairn/client-a                               1 │
         │   ● home      /Users/me/cairn/home                       not found · 2 │
         │   ● work      /Users/me/cairn/work                      here · default │
         │                                                                        │
         │ RECENT FOLDERS                                                         │
         │   ◌ /Users/me/scratch                                 opened yesterday │
         │   ◌ /Volumes/usb/trip-2026                               not found now │
         │                                                                        │
         │ ^O open a folder   ^N new workspace                                    │
         └────────────────────────────────────────────────────────────────────────┘

 ↑↓ choose   ⏎ open here   1-9 by place   esc back
```

`^O` opens the switcher over the board: every workspace the configuration
names, in the order `workspace list` gives them, each with its folder; the one
shown is marked `here`, the default `default`, one whose folder is not there
`not found`, and every other one carries its place, `1` to `9`
(`@tui-workspace-ctrl-o-lists-them`). The line at its head narrows the list as
a name is typed (`@tui-workspace-the-list-narrows-as-one-types`). `↑↓` choose,
`⏎` opens the one chosen, a digit opens the one at that place
(`@tui-workspace-chosen-by-its-place`), and `esc` leaves the board as it was
(`@tui-workspace-esc-leaves-the-board-as-it-was`).

**Every letter is typed.** The switcher has a line to type into, so `j` and
`k` are letters there, as they are at `/` and `:`; `↑↓` choose. A digit
chooses by place rather than being typed: a name is narrowed by its letters.

**Choosing replaces the board in place.** A terminal is one process and one
screen, so the store held is closed and the other opened in the same process,
through `engine.Switch`, and the board starts again on it — the cursor in the
backlog, no search, no filter. The header says the new name, and the line above
the bar says what happened until the next key: `opened home in place of work`
(`@tui-workspace-choosing-replaces-the-board`). A focus held in the one left is
not released by switching — the store keeps it — and the same line says so,
`work's focus on #152 is still held`
(`@tui-workspace-the-focus-left-is-still-held`). The pages are read again from
the settings of the store opened, so a workspace whose time is off has no time
tab once it is shown (`@tui-workspace-time-off-there-takes-the-time-tab`).

**The recent folders come under the workspaces**: the stores a board opened by
path that no workspace names, the newest first, each with the day it was last
opened, and one gone since said `not found now`
(`@tui-workspace-recent-folders-listed`, and what they are in
[config.md](config.md#recent-folders-remembered-and-never-configured)). The
same `↑↓` walks on from the workspaces into them, the line at the head narrows
them too, and `⏎` opens one in place (`@tui-workspace-a-recent-folder-opens-in-place`).
No digit opens one: the places `1` to `9` are the workspaces'.

**Three keys at the foot of the list, each a control key.** Every letter is
typed into the line at its head, so a letter cannot also be an act: `o` would
make `home` impossible to type. So:

- `^O`, pressed again inside the list, asks for a folder on that line and
  opens it in place, remembered among the recent folders
  (`@tui-workspace-a-folder-opened-in-place`). It is absolute or under `~`,
  like every place cairn reads, and one that is not is refused with the
  engine's reason, the line kept as typed
  (`@tui-workspace-a-relative-folder-is-refused`); a folder that is not
  there is refused too (`engine.ErrNotThere`), rather than starting an empty
  board where nobody looks. A folder holding no store yet opens on a new one.
- `^N` asks a name, then a folder, and makes the workspace through the same
  operation as `cairn workspace add` — the same rules, the same refusals — and
  opens it in place (`@tui-workspace-a-new-one-made-and-opened`,
  `@tui-workspace-a-new-one-refuses-a-name-known`). The default stays what it was.
- `^K`, on a recent folder, asks a name and keeps the folder as a workspace,
  again through `workspace add`; the list is read again, the folder now among
  the workspaces and gone from the recent ones
  (`@tui-workspace-a-recent-folder-kept`). A store that is not its folder's
  `cairn.db` cannot be kept, and `^K` is not offered on it.

`esc` on one of those lines goes back to the list, and a second leaves it.

**A workspace whose folder is gone is refused when chosen**, with the
engine's reason, and the switcher stays open on the board as it was
(`@tui-workspace-a-folder-gone-is-refused`): opening it would make an empty
store where nobody looks.

**The switcher changes what this board shows, and nothing else.** It does not
write the configuration: the default stays what it was, and `workspace use` —
which changes the default for every command — is not what `^O` does. `cairn tui
--workspace home` starts the board on `home`.

The settings of a workspace are not changed from here; the window's panel is
where they will be.

## Staying true while the store changes

The board stays open while a `cairn task new` in another terminal, or an agent
over MCP, changes the store underneath it.

- A slow beat — two seconds — reads **`PRAGMA data_version`**. That is one
  integer SQLite bumps when another connection writes, not the store.
- If it moved, the board re-reads and redraws. There is no key to press.
- **The cursor holds a task's number, not a position on screen.** A redraw can
  make the cursor jump; it cannot make a keystroke land on the wrong task.

That last point is what made a live redraw safe. It was nearly rejected on the
grounds that a card moving under the cursor would make you move the wrong task —
which was wrong twice over: SQLite serialises the writes, so there is no conflict
in the data, and a gesture names a number rather than a row.

| what happens | what the cursor does |
|---|---|
| the task under it is deleted elsewhere | the next task in the same column |
| there is no next one | the previous one |
| the column empties | back to normal mode, on that column |

**The configuration is read again on the same beat.** The board keeps the
configuration it read (`engine.ConfigShow`) and a stamp of the files it came
from — the global `config.toml`, the `settings.toml` beside the store, and a
scheme of one's own worn (`engine.StampOf`). At each beat it looks at their size
and time, and only when one moved, at what it holds: a save with nothing
changed does nothing. When a file says something else, the board reads the
configuration again through the engine, wears it — the pages, the weekends, the
series — and says on its message line which keys changed and from what,
`settings.toml read again: pages.time on → off`, until the next key
(`engine.ConfigChanges`, `@tui-config-read-again-says-what-changed`). A theme
changed reaches it as nothing: the terminal draws in its own sixteen colours.

**A file that cannot be read leaves the board as it was.** The board does not
fall back to the defaults, which would bring a page back or take one away under
the person's eyes: it keeps the last configuration that read, and says where
and why on its message line — the file, the line the key at fault is written
on, and the reader's own words, `settings.toml:2 pages.time "hiden": …` — with,
in the bar's place, `the board keeps what it read at 10:42 until the file
reads again` (`engine.TroubleIn`,
`@tui-config-a-file-that-cannot-be-read-keeps-what-it-had`). Both stay until
the file reads; then the change is worn and said as any other. At open, a
configuration that cannot be read stops the board as it always did: there is
nothing good to keep yet.

`PRAGMA data_version` was measured on `modernc.org/sqlite` rather than assumed:
it moves when a second connection commits and stays put when nothing happens.
The scenarios that hold it are in `internal/store/version_test.go`. It says
nothing about *what* changed, and it is not a count — two writes may move it by
more than one. Equal means nothing happened; different means the only honest
answer is to look again.

## What it is built from

| | why it is there |
|---|---|
| **Bubble Tea** | there is no TUI without an event loop and screen state |
| **Lip Gloss** | already pulled in indirectly by Fang; goes direct for borders and joins |
| **Huh** | the form: fields, focus, and a text area that wraps |
| **Bubbles** | a viewport for what is longer than its box, and key bindings |

Every dependency's reason is in [dependencies.md](dependencies.md).

## How the form is laid out

**The fields are dealt into two columns**, under a head that draws the thing
being changed the way a card draws it: its number, its title, where it stands and
what priority it carries, in the colours it carries them in. A dozen small facts
in a single column is a scroll rather than a shape, and a form headed "form"
tells you nothing you did not already know.

**The arrows walk the fields**, and a field with a closed set of values is
changed in place with left and right — `← todo →` rather than a list that
unrolls. Seven columns and four priorities are a value to cycle, not a page to
read.

Arrows and not `hjkl`, and this is the one rule the board gives up: inside a
field being typed into, a letter is a letter. `hjkl` navigate everywhere else; in
a title they spell.

The last thing the form asks is whether to save, answered yes unless somebody
says otherwise — a form is opened to change something. `esc` is the other way
out, and it writes nothing. Neither of those two keys ever reaches the form
itself: help answers over it, and escape closes it, because the library binds
neither and a bar that promised escape while escape did nothing was worse than
no bar.

The board draws in English, like the command line and the MCP server. Nothing a
person reads on screen is in another language.

## How it reaches the engine

The board asks for the operations `internal/engine` describes, through one door:

```go
type Work interface {
	Ask(engine.Operation, engine.Input) (engine.Answer, error)
	Version() (int64, error)
}
```

It named sixteen engine methods to begin with, and that was wrong for one
reason: the command line and the MCP server name none. Both are *built* from
`engine.All()`, so a verb added to the engine reaches them without a line
written in either — and the board was the one surface it would not reach.
`*engine.Engine` satisfies the two methods above as it stands, so nothing
adapts anything.

`Ask` is a second door beside `Do`, not a replacement. `Do` opens the store,
runs one operation and closes it, which is right for a surface that lives for
one invocation. A board on screen cannot pay an open and a close on every
keystroke and does not have to: it is holding the engine already.

**`Version` is the one thing that is not an operation**, and it stays a method
of its own. It reads `PRAGMA data_version` — the store noticing that another
connection wrote — and nobody asks cairn to do that, so it has no verb and
should not get one. `cairn version` already means the version of the binary.

**What the board reads lives in `internal/engine`, not here.** The two
listings turned into cards and goals, the deal into columns, which fields a
form asks about — none of that is about a terminal, so none of it is in this
package. What stayed is everything that measures in cells and everything that
draws. See [architecture.md](architecture.md).

**A field says which argument it fills.** `Field` carries both `Name`, what a
person reads beside the box, and `Arg`, what the engine calls the same thing:

```go
{Name: "labels", Arg: "label", ...}
{Name: "time spent", Arg: "spent", ...}
```

Neither is wrong. A repeatable flag reads singular on a command line
(`--label api --label urgent`); a box holding all of them reads plural on a
screen. Saying it on the field is what stops a translation table existing
somewhere else — and `engine.Asked` then builds the whole request by walking the
form and reading each argument's `Kind` off the operation, so a field added to
the form, or an argument added to the engine, arrives with no line written for
it.

What the board still decides for itself: a number nobody can read is **not
named** rather than sent for the engine to refuse, because one unreadable field
would otherwise lose the whole form.

## Deliberately not there yet

- **`mcp serve`** — nobody starts a stdio server from a board. It stays a
  command-line command, and it is the only operation with no place here.
- **The window's scheme in the terminal, `[board] colours = "theme"`.** It
  would paint the window's tokens in true colour, for whoever wants one look
  in both; it is not built. The sixteen colours stay the default either way,
  since a board that repaints the ground fights the terminal's own.
- **Moving between the three stopped states.** Once a task is in the band, going
  from blocked to watching is the form's job. The popup only asks on the way in.
