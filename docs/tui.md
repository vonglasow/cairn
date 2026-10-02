# The board

Read this before touching the terminal board — its shape, its keys, its modes —
or before adding anything a person sees on screen.

`cairn tui` opens it. What follows is what it does and why it does it that
way, decision by decision — written down because thirty decisions taken in one
conversation are thirty decisions nobody can find afterwards.

The board is a surface, like the command line and the MCP server: it parses what
somebody pressed, calls `internal/engine`, and draws what comes back. It decides
nothing. It breaks one assumption the other two never did — **it stays open**,
and the store can change underneath it. That is dealt with at the end.

## Two tabs

**tasks** — the kanban, where the work is. **objectives** — the list of what the
work is for. `⇥` switches. Nothing else is a tab.

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

 hjkl ←↓↑→ columns   i insert   a add   / search   ⇥ tab   ? help   : command
```

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

At sixty it is tight but true: a title has twenty characters across its two
lines, labels drop off the facts line, and the three stalled states are glyphs
rather than words — which is why they are glyphs everywhere and not only here.

## A card is six lines

Always six: a border, four lines of content, a border.

| line | what is on it |
|---|---|
| 1 | the number on the left, the marks on the right |
| 2–3 | the title, cut at a word, `…` if it overruns |
| 4 | the facts, in this order: deadline, sub-tasks, labels — cut when the line is full, with `+n` saying how many were left out |

The height is a budget, not a maximum. A title longer than two lines is cut.

A fixed height was chosen over a snug one, and it bought something that was not
free before: **navigation becomes arithmetic.** A card is six rows, so `j` and
`k` know where they land, and how far a column has to scroll for the one under
the cursor to be wholly on screen is a division rather than a measurement. What
it costs is two empty lines on a card carrying nothing.

**Everything that can be longer than its window scrolls**: a column, the band,
the objectives, a task's history. Each is drawn behind a viewport that follows
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
board first, then the description mark, and the state glyph last. The state anchors the far right so the eye finds it
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
| `low` | blue |
| `normal` | grey |
| `high` | amber |
| `critical` | red |

Colouring the whole border was turned down: three `high` cards in a row turn a
column into an amber wall and nothing stands out any more.

**The task in focus has a double border**, and it is the only one on screen. A
dot beside the number was drawn and removed — the frame already says it, better,
and the dot was eating the most contested line on the card.

**A deadline is bare** — `2d`, `−3d` in red when it has passed — and it only
appears when it is close or gone. A date three months out is not information on a
board. It carries no clock glyph: the glyph rendered badly and said nothing the
`d` did not.

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
| `watching` — handed over, needs checking | `nf-oct-eye` | `◉` | blue |
| `paused` — waiting on your own decision | `nf-md-pause` | `‖` | grey |
| `active` objective | `nf-oct-play` | `▶` | green |
| `on-hold` objective | `nf-md-pause` | `‖` | grey |
| `reached` objective | `nf-oct-check` | `✓` | green |
| `dropped` objective | `nf-oct-x` | `✕` | dim |
| a description exists | `nf-oct-note` | `≡` | |
| sub-tasks done / total | `nf-oct-checklist` | `▪` | |
| a label | `nf-oct-tag` | `⚑` | |
| what is in the way | `nf-oct-blocked` | `↳` | |
| off the board | `nf-oct-archive` | `▭` | |

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

Every key below says which mode it works in. A key absent from a mode is absent
from that mode's bar, and pressing it there does nothing.

| | normal | insert | objectives list | |
|---|:---:|:---:|:---:|---|
| `hjkl` `←↓↑→` | ✓ | ✓ | ✓ | navigate — columns, tasks, objectives |
| `i` | ✓ | | ✓ | enter insert · on objectives, open the objective |
| `⇧hjkl` `⇧←↓↑→` | | ✓ | | move the task anywhere; its state follows where it lands |
| `a` | ✓ | | ✓ | add — the column under the cursor gives the state |
| `x` | | ✓ | | take the task off the board, or put it back |
| `⇧A` | ✓ | ✓ | | show what was archived, or hide it again |
| `⇧X` | ✓ | | | take every task in the column off the board — it asks first |
| `e` | | ✓ | ✓ | the form |
| `d` | | ✓ | ✓ | delete — it asks first, `⏎` agrees and anything else backs out |
| `f` | | ✓ | | take or let go of the focus — an objective has none |
| `⏎` | | ✓ | ✓ | the detail · on objectives, open the objective |
| `/` | ✓ | | ✓ | search |
| `?` | ✓ | ✓ | ✓ | help — the one key that works in every mode |
| `:` | ✓ | | ✓ | command |
| `⇥` | ✓ | | ✓ | next tab |
| `gg` `G` | | ✓ | ✓ | the first / the last of what you are in |
| `esc` | ✓ | ✓ | | back up one level |

On the board itself `esc` widens a search that is still in force, and does
nothing once there is none: there is nothing above it to back up to, and leaving
is `:q`. A search accepted with `⏎` outlives the prompt that typed it, and
nothing else on the board says so — so `esc` is the way back to the whole board. Inside an objective's own board there is — it goes back to
the list of objectives, with the cursor still on the one you were in.

**Taking a whole column off the board** is `⇧X`, in normal mode, on the column
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
- `⇧K` out of the band returns it to **the column the cursor came from**:
  todo → stalled → todo. That memory lives on the cursor, not on the task, so it
  dies with the session.
- `⇧H` from the leftmost card in the band sends it to the **backlog**, even
  though the band does not run under the backlog.

Reordering inside a column — `task move` with `first`, `after`, `last` — is not
here. `⇧J` and `⇧K` are spent on the band, and reordering was not wanted yet.

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
 │    because        —                                                        │
 │    deadline       2026-09-02                                               │
 │    estimate       3h                                                       │
 │    time spent     1h30                                                     │
 │    parent         —                                                        │
 │    objectives     3                                                        │
 │                                                                            │
 │                                             [ save ]          [ cancel ]   │
 └────────────────────────────────────────────────────────────────────────────┘

 hjkl ←↓↑→ navigate   ⏎ change   esc cancel
```

**Nothing is written until `save`.** Writing each field as it was left was tried
on paper and dropped: `cancel` would have meant nothing, and the same form could
not have served to create something that does not exist yet.

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
have squeezed the board and a popup would never have shown the whole history.

```
 ‹ TASKS ›  ▸ 14                                                          ── DETAIL ──

 ┃ 14   Write the spec                                       ≡    todo    high
 ┃ ⚑ spec  infra        due 02/09 · in 2d               3h estimated · 1h30 spent

 ┌─ DESCRIPTION ───────────────────────────────┬─ HISTORY ────────────────────────────┐
 │ The client wants two columns, not three.    │ 12/08  created    backlog · normal   │
 │ The third was in the mockup by accident     │ 18/08  priority   normal → high      │
 │ and nobody caught it before the review.     │ 21/08  column     backlog → todo     │
 │                                             │ 27/08  deadline   — → 02/09          │
 ├─ SUB-TASKS ────────────────────── 3/5 ──────┤ 29/08  focus      taken              │
 │ ✓ 15   Draft the columns          done      │ 29/08  focus      released           │
 │ ✓ 16   Get the client's nod       done      │ 30/08  spent      — → 1h30           │
 │ ✓ 17   Write the scenarios        done      │                                      │
 │   19   Client's answer          ⊘ blocked   │                                      │
 │   20   Ship it                    todo      │                                      │
 ├─ OBJECTIVES SERVED ─────────────────────────┤                                      │
 │ ▶ 3    Ship the desktop board     7/12      │                                      │
 └─────────────────────────────────────────────┴──────────────────────────────────────┘

 hjkl ←↓↑→ navigate   e edit   f focus   ⏎ open the sub-task   esc back
```

The band carries the short fields; the panels carry the lists. **A field that is
empty is not shown** — the opposite of a card, where an absence keeps its line to
hold the height. A full screen has no such constraint.

`⏎` on a sub-task opens its detail, so the tree is walked without going back
through the board, and `esc` climbs back one step at a time.

This is the only place the trace is laid out beside the task it belongs to.
`:trace show 14` prints the same history, the way the command line prints it, and
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

A flat list of results was turned down for losing the column each match is in,
which is most of what a board is for. Highlighting without hiding was turned down
because with two hundred tasks what is highlighted is usually off screen.

## Help, and the command line

```
 ‹ TASKS ›  help                                                            ── HELP ──

 ┌─ KEYS ──────────────────────────────┬─ LEGEND ──────────────────────────────────┐
 │                                     │                                           │
 │ NORMAL — cursor on a column         │ PRIORITY — the left border                │
 │   hjkl ←↓↑→  change column          │   ▏ blue   low       ▏ amber   high       │
 │   i          insert mode            │   ▏ grey   normal    ▏ red     critical   │
 │   a          add a task             │                                           │
 │   /          search                 │ STALLED — top right of the card           │
 │   ⇥          next tab               │   ⊘  blocked    waiting on somebody else  │
 │   :          command                │   ◉  watching   handed over, needs a look │
 │                                     │   ‖  paused     waiting on your decision  │
 │ INSERT — cursor on a task           │                                           │
 │   hjkl ←↓↑→  change task            │ OBJECTIVES                                │
 │   ⇧hjkl      move the task          │   ▶  active     ✓  reached                │
 │   ⏎          the detail             │   ‖  on-hold    ✕  dropped                │
 │   e          the form               │                                           │
 │   f          take / let go focus    │ ON THE CARD                               │
 │   d          delete                 │   ≡    a description exists               │
 │   esc        back to normal         │   ▪    sub-tasks done / total             │
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
 │   backup add             take a backup now                                 │
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
| `:backup` | `add` `list` `show <n>` `restore <n>` `delete <n>` |
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

## Objectives

```
 tasks  ‹ OBJECTIVES ›                                                    ── NORMAL ──

 OBJECTIVES                                                                             9
 ───────────────────────────────────────────────────────────────────────────────────────
 ┃───────────────────────────────────────────────────────────────────┐
 ┃ 3                                                       ≡    ▶    │
 ┃ Ship the desktop board                                            │
 ┃                                                                   │
 ┃ ████████████████░░░░░░░░░░░░  7/12   45d   ⚑ product              │
 ┃───────────────────────────────────────────────────────────────────┘
```

An objective carries eight fields and one thing a task does not: **an
advancement**, counted from the columns of the tasks that serve it. That is why
the bar is the headline — it is the only information an objective has that
nothing else does.

Same card, same six lines, same rules. The state is a glyph on the card rather
than a section heading, so there are no sections at all: the order comes from the
state, active first, and an objective that is reached simply ends up at the
bottom.

It is not a kanban. An objective does not circulate — nobody pushes one from
`active` to `reached` in the morning, it gets there when its tasks are done. And
there are ten of them, not two hundred, so four narrow columns would spend the
width on a notion that does not need it.

**`i` or `⏎` opens the objective**, and what opens is the board, filtered to the
tasks that serve it, with the objective's card flattened into a band on top.
`esc` comes back out.

```
 tasks  ‹ OBJECTIVES ›  ▸ 3                                               ── NORMAL ──

 ┃ 3   Ship the desktop board                       ≡   ▶   45d   3w   ⚑ product
 ┃ ████████████████░░░░░░░░░░░░  7/12

 BACKLOG                2 TODO                   3 CURRENT                1 DONE                   7
 ──────────────────────── ──────────────────────── ──────────────────────── ────────────────────────
 ...

 hjkl ←↓↑→ columns   i insert   a add to this objective   / search   esc back   ? help
```

Same columns, same band, same keys, same `⇧hjkl`. Nothing new to learn. Two
things it gives that a read-only list would not: `a` here creates a task that
already serves this objective — the one place where creating a task answers a
question, "what is missing to finish this" — and the whole thing is the board you
already know with less in it.

**Sub-tasks do not appear.** A sub-task never serves an objective in its own
right; it counts through its parent. So the filtered board shows the tasks that
serve it directly, and `7/12` counts those.

An objective's state changes through the form, not by pushing it anywhere. With
no sections there is nowhere to push it to, and `⇧hjkl` therefore does nothing
on the objectives list — so it is not on the bar there. `j` and `k` walk the
list, and it scrolls to whichever the cursor is on; `h` and `l` do nothing,
because the list is one column and a key that pointed sideways and moved
something would be a key that looked broken.

## Staying true while the store changes

The board stays open while a `cairn task add` in another terminal, or an agent
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

**What the board reads lives in `internal/reading`, not here.** The two
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
somewhere else — and `reading.Asked` then builds the whole request by walking the
form and reading each argument's `Kind` off the operation, so a field added to
the form, or an argument added to the engine, arrives with no line written for
it.

What the board still decides for itself: a number nobody can read is **not
named** rather than sent for the engine to refuse, because one unreadable field
would otherwise lose the whole form.

## Deliberately not there yet

- **`task move`** — `first`, `after`, `last`. Reordering inside a column has no
  gesture, and was not asked for.
- **`mcp serve`** — nobody starts a stdio server from a board. It stays a
  command-line command, and it is the only operation with no place here.
- **`:` for tasks and objectives.** `e` already does it, and one way in is
  enough.
- **What a card shows of its objectives and its parent.** Both were left off:
  an objective is a link to elsewhere and there is a tab for it, and a parent
  belongs to how a column draws its tree rather than repeated on every card.
- **Moving between the three stopped states.** Once a task is in the band, going
  from blocked to watching is the form's job. The popup only asks on the way in.
