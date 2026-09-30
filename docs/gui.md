# The desktop board

Read this before touching `internal/gui`, or anything a person sees in a
window.

The same board as the terminal, in a window: four columns, a stalled band, two
tabs, the same work underneath. Prettier, and driven by a mouse. Nothing here
is a second idea of what cairn is — where this file and
[tui.md](tui.md) disagree about what a board *means*, tui.md is right and this
is the port.

This says what it is and why, including what was turned down — because the
reason is the part the code cannot carry.

## What it is for

Three things, in the order they matter: **range, search, open**. The engine
answers one command at a time, and that work wants everything visible at once.

The mouse comes first. No modes, no cursor to move, no `esc`. A few flat
shortcuts stay — `⌘N` for a new task, `⌘F` to search, `?` for help — and
nothing that has to be learnt before the board can be used.

**macOS and Linux.** That one sentence closed more than it looks: SwiftUI was
out from the moment Linux was in.

## Why Gio

Three ways to draw the same window, measured rather than remembered:

| | UI written in | under prek, gremlins, the coverage floor | drag and drop |
|---|---|---|---|
| **Gio** `v0.10.2` | Go | **yes, all of it** | to write |
| **Wails** `v2.15.0` | HTML/CSS/TS | **no — the UI escapes every gate** | given |
| **Fyne** `v2.8.1` | Go | yes | partly given |

Wails has the highest ceiling on looks and was turned down for what it costs:
the part being judged on its appearance would sit outside every check this
repository has — no mutation, no coverage, no pinned frame — and it would add
an npm build to a repository that has no JavaScript at all. WebKitGTK to
install on Linux, and a v3 still in beta, were the smaller reasons.

Fyne would be quickest to something clickable, and its look is its theme's
look. That ceiling is the thing this port was asked to raise.

Gio keeps one language, one static binary, and every line under the gates. It
is also the most work: scrolling, dragging and text entry are built by hand.
That is the trade, stated plainly.

**One dependency, and this is its written reason** — `CLAUDE.md` asks for one:
`gioui.org`. It is the only way to a vector-drawn window that is identical on
macOS and Linux, in a single binary, with no system webview and no second
language.

## Where it lives

`internal/gui`, a fourth surface beside `cli`, `mcp` and `board`. `depguard`
denies it `store`, `database/sql`, `cobra`, `cli`, `mcp` — and `board`: two
surfaces do not reach each other.

**What it does not write again**, because `internal/reading` already carries
it: `Look` and the two listings turned into cards and goals, `Deal` into
columns, `Stopped`, `Deadline`, `Sorted`, `TaskFields`/`GoalFields` with each
field's `Arg`, `Asked` turning a saved form into what an operation takes, and
the `Work` interface — `Ask` and `Version`, two methods.

What is its own: geometry in pixels, drawing, and the mouse.

`cairn gui` is a command in `internal/cli`, built like `cairn tui`, with the
same seam (`var openGui = func(…)`) so a test never needs a window.

## The window

The terminal's shape, larger. Four columns — backlog, todo, current, done —
the backlog full height, the stalled band across the last three. Two tabs:
tasks, objectives.

**Cards keep a fixed height.** In the terminal that is what makes moving the
cursor arithmetic rather than measurement; with a mouse that reason is gone,
and it was kept anyway for what it still buys: a column can say how many cards
fit, where the n-th one is, and which card is under the pointer, without
drawing anything. The price is the same as the terminal's — a long title is
cut.

Variable height was the alternative and would have cut nothing, at the cost of
a column's geometry depending on its contents: scrolling and drop targets
computed instead of multiplied, and two columns no longer sharing a grid, which
shows the moment a card is dragged from one to the other.

**A card is three lines**: its number, its title, and the labels it carries.
The third fills room the fixed height already paid for and left empty, and it
answers the question a board is asked most — which sprint, which chantier, which
ADO ticket — without opening anything. What does not fit is counted rather than
dropped in silence: `ado:77594  pipelines  release  +1`. `Tagged` cuts the line
and the scenarios read it, so what is asserted is the decision rather than the
pixels.

It measures with `perSmallGlyph` and not `perGlyph`. Labels are written in the
small face while the title uses the wide one, and measuring with the title's
gabarit left a fifth of every label line empty — on a card carrying three, the
third was dropped for room that was there.

**Every size is a value worked out from the window**, redone on every resize,
exactly as `LayoutFor` does in the terminal. Nothing carries its own size. That
is what makes the geometry testable with no screen: a pure function from
(width, height) to rectangles, compared against pinned numbers.

In **dp** rather than pixels, so a Retina screen and a 96-dpi Linux screen give
the same physical thing.

**Too small**: below a threshold the window says what it needs instead of
drawing something unreadable — the terminal's answer, for the terminal's
reason. Imposing a minimum window size was the alternative and hides the
question rather than answering it; showing fewer columns was turned down as a
different board.

**1024 by 640**, which is four columns of 285dp and the height two rows of the
band and three cards above them come to. Written in rather than worked out,
because what makes a card readable is not something a sum knows — and measured
after the card was drawn rather than guessed before it.

**The beat stays at two seconds**: `Version()` is one integer, and the board
re-reads only when it moves. Waiting for the window to take focus was turned
down for a reason worth keeping: a board on a second screen, never focused, has
to be true.

## The mouse

A click chooses a card, or changes tab when it lands on one. A double-click
opens a card's detail. The wheel scrolls the column under the pointer.

**A search does not take the window.** It narrows the board in place, so the
board keeps answering the mouse: a card is chosen by a click and opened by a
double click while the box is still open and still being typed into. It used to
swallow both, and reaching what a search had just found meant closing the box
that found it.

A click over the box itself belongs to what is being typed, and lets go of
nothing. That is the whole of the exception — everywhere else on the board a
click does what it always did.

The guard that swallowed them was right about panels and wrong about this:
a detail, a form or the help is drawn *over* the board, and a click reaching
through one is how an open detail came to show something nobody asked it for.
A search is not drawn over anything. Reading "not the board" as "a panel over
the board" is what made one rule cover two unlike things.

**A keystroke can arrive twice**, and what a window does about that is not the
same on every system. Gio sends a `key.Event` for a key with a name and a
`key.EditEvent` for the text an input method made of it; X11 sends both for an
ordinary letter where macOS sends one. Listening to both meant `e` opened the
form and then wrote an `e` in the title, and `/` opened the search and then
filled it with a slash — which finds nothing, so every search from the keyboard
came back empty.

Text is now taken only where there was already somewhere to put it: `heard`
remembers which screen the window was on when the frame began, and an edit
event is delivered only if that screen was the form or the search. Judging by
where the window *was* rather than where it *is* is the whole of it — by the
time the text arrives, the shortcut has already moved the window somewhere that
would accept it.

Measured, and reachable without a second machine: the scenarios drive both
halves of one keystroke through Gio's own router, which is what made a
Linux-only fault reproducible on a Mac.

A double-click is worked out here rather than asked of the system: Gio reports
presses and says nothing about pairs of them. Two presses within 400ms and
within the same few pixels are one gesture.

**A click that moves becomes a drag.** Past a few pixels the card lifts and
follows the pointer; below that it is a click. Without the threshold, a shaky
click would move a task.

**A drop carries the column and the place.** That is `task edit --state`, then
`task move --after`. Two acts in the trace for one gesture, and no transaction
between them: if the second fails, the board refreshes and shows the truth —
column changed, place not. Visible rather than silent.

Carrying only the column was the alternative: one operation, one line of
history, faithful to an engine that refuses to change a column and a place in
one act. It was turned down because dropping a card precisely somewhere and
watching it land elsewhere is a small betrayal of the hand.

**Sub-tasks are not drawn on the board**, so every card on it hangs under
nothing and any card may take any place. There was a rule here that drew no
landing line where a sub-task could not go, and a mark on the card saying it
hung under another; both went when sub-tasks left the board, since nothing
draws one there to refuse or to mark.

That rule is what made this change happen. It was right — a sub-task could only
land among its sisters — and it read as arbitrary, because the sisters were
scattered through the columns with nothing grouping them. Fixing the refusal
showed that the hierarchy was invisible, and the answer turned out not to be a
better refusal.

**The stalled band is three lanes**, named `blocked`, `watching` and `paused`.
Dropping into a lane *is* the answer to which kind of stopped this is. The
terminal asks with a popup because a terminal band has no room for three named
lanes; a window does. It also means a stopped card says which kind it is by
where it sits, which the terminal needs a glyph for.

## The form

A modal dialog over the board, with *Save* and *Cancel*; `esc` cancels. One act
in the trace per save.

**The fields are not listed here.** They come from `reading.TaskFields(card)`
and `reading.GoalFields(goal)`: the word a person reads, the closed set when
there is one, the field that wants more than one line, and `Arg` — what the
engine calls the same thing. An attribute added to the engine and to the
terminal's form arrives here with no line written.

A field with `Choices` is a dropdown, a `Long` field a multi-line box, the rest
a line to type in. The field the keyboard is on is lit, and it is the only one:
that is the whole of what says where typing would go.

**The objectives a task serves are picked from a list, not typed.** `enter` on
that field opens the objectives over the form — over, and not instead of: a
choice made with the rest of the form out of sight is a choice made from
memory. The arrows walk it and so do `j`/`k`; **space** ticks and unticks;
`esc` closes it and keeps what was ticked. Nothing reaches the store until the
form is saved, the way nothing else in a form does.

`enter` saves everywhere else in a form and here it does not, so the bar says
`Choose` instead of `Save` while the keyboard is on this field, and `Save`
comes back the moment it leaves. The rule is the one the rest of the window
keeps — a key does what the bar offers — and a bar still offering *Save* while
`enter` opened a list would be a bar that lies, which is the one thing it
exists not to do.

**The field carries numbers and is drawn as names.** `1 2` is what the engine
takes, and a surface inventing its own spelling of an argument would be a
surface deciding something; `#1 Ship the desktop board, #2 Keep it honest` is
what a person reads, because nobody knows an objective by its number and the
one place that said what a number meant was the other tab.

**Every field is drawn in a box**, whether or not the keyboard is on it, and
the lit frame is the only difference between the two. A value floating on the
panel's own ground looked like something written rather than somewhere to
write.

**Nothing written runs out of its box.** Text wraps — at a space when there is
one, inside a word when the word is longer than the room — and the box is
clipped, so even a line of unusually wide characters cannot escape it. Cutting
a value down to one line was the alternative and is what this used to do; it
said nothing about what was lost, and a description is the field it lost most
of.

**The caret is a shape, not a character.** Written into the text it took a
letter's place and pushed everything after it along, so typing looked like the
text was being rewritten under the hand. It is drawn between two letters, at a
measured x, on the line it actually fell on — placed from the height of a
whole wrapped label it lands on the last line whatever was typed.

**What the arrows reach, they bring into sight.** Walking used to move the
choice and leave the column where it was, so a card could be chosen below the
bottom of its column — and a card chosen out of sight is a card whose title,
whose state and whose place are all unknown to whoever chose it. Walking down a
column and walking across into one somebody had scrolled both follow.

**The objectives are a grid**, filled across and then down, each one the width
of a card. One to a row and the width of the whole board, an objective was a
card's height stretched over four columns of nothing — and a dozen of them ran
off the bottom of the window, so the ones that mattered least were the only
ones in sight. Across, a dozen fit without scrolling at the smallest window
this board draws in, and the grid reflows with the window because it is laid
out in the board's own columns rather than on a sum of its own.

They scroll on a number of their own, being one grid rather than four columns.
They had none at all: the wheel did nothing over them and the arrows walked
past the bottom of the window. The arrows walk the grid — across to the next,
down to the one a row below.

**The tabs are folders, not two words one of which is coloured.** A band
closed by a line runs the width of the window, and the tab somebody is on is a
block carrying the page's own ground whose bottom breaks that line. It does not
stand out by its colour, it stands out by opening onto the page — which is what
says *you are on this page* rather than *this word is selected*. Each says how
much its page holds, counted off what is showing, so a search narrows what the
tabs claim as well. `tab` walks to the next page and `shift+tab` to the one
before, round the end rather than stopping at it; inside a form `tab` still
walks the fields, as it does in the terminal.

The columns stand a gap clear of that band. Against it, the tabs read as the
board's first row.

**A card carries two marks and no more.** The eleven the terminal board draws
were left out of this window on purpose; these two earn their place because
neither can be worked out by looking. A card gives no sign that somebody — or
some agent session — is on it, and a sub-task looks exactly like a task while
being the one thing that cannot be dragged wherever you like: it moves among
its sisters and nowhere else. So the number line carries `↳` and its parent's
number, and `●` when it is being worked on. `f` takes the chosen task up or
lets it go, under the same key the terminal board uses.

**A key does what the bar offers, where the bar offers it.** The bar is what
somebody reads to find out what a key does, so `Key` looks the key up among
`Offers()` rather than carrying a list of its own — an offer added to a screen
cannot arrive without its key, and a key cannot work on a screen whose bar says
nothing about it. Two defects came from the older shape: the bar offered *Edit*
on an open task while `e` did nothing there, and the key that adds worked over
an open form and threw away what had been typed into it.

**Archive** is one of those offers, under `x`, beside *Edit* and *Delete* on a
chosen card. On a card already off the board the bar reads *Restore* instead:
one key both ways, and the bar says which way it goes rather than leaving it to
be worked out. Unlike *Delete* it asks nothing first — nothing is lost.

**Archived** is another, under `⇧A`. It puts back on the board what
was taken off it, and takes it away again — the same key the terminal board
uses, because the two boards should not have to be learned twice. A card shown
that way is written in the faint ink and loses its priority stripe: it is on
screen because somebody asked, not because it is work to do. Which ink a card
is written in is `Skin.Writing`, so what a scenario asserts is the decision the
drawing makes rather than the pixels it made.

**Shift arrives apart from the key, but only for a letter.** Gio names a letter
in upper case whatever shift does and reports the modifier separately, so
`Typed` puts the two back together into the `shift+a` this board's vocabulary
uses. Punctuation is left alone, because Gio has already applied shift to it:
the `/` key with shift is named `?` before it gets here, which is how the help
is reached. Joining it would have taken the help away.

`⇧A` did nothing in a real window until this was written, and every scenario
passed anyway, because the step that presses a key drove `Key` directly. It now
goes through `Typed`, so a key no window could send can no longer make a
scenario green.

What a key is *called* and how the bar *writes* it are two functions.
`Shortcut("New")` is `cmd+n`, because that is how a key arrives; `Pretty` turns
it into `⌘N`, because a bar reading out `cmd+n` is reading out a name meant for
a program.

**The description scrolls, and says so.** On the detail it is read rather than
written, so nothing follows a caret and the four rows its box holds were the
whole of what a long description could ever show — the rest was clipped, with
nothing on screen saying it was there. It now takes the wheel and carries a bar
beside it, reusing the history's `Thumb` and window rather than growing a
second scrolling mechanism. The bar is drawn whether or not there is anything
to scroll, by the same rule the history goes by.

The form is left alone: there the caret drags the window along with it, so what
does not fit is already reachable.

**A wheel keeps what did not come to a line.** `by / lineHigh` threw away every
turn smaller than a row, which is every turn a trackpad sends — one pixel at a
time, each rounding to nothing, so a description could not be scrolled with a
mouse at all and a history answered only a notched wheel. `turning` adds the
leftover to the next turn instead. A history still moves by whole lines, which
is what its own test asks; six nudges of three pixels are now one of them.

The bar beside a description **can be pulled**, like the history's. Without it
a press on the bar fell through to the board behind the panel and closed the
task — and the scenario passed anyway, because the end of a description nobody
is showing is zero, which is where an untouched window already sits.

**A number somebody would have to look up is chosen from a list.** Two fields
are answered that way — the objective a task serves, and the task it hangs
under — and one list answers both. It carries a line that narrows it as it is
typed into, by the board's own matching rule rather than one of its own: a list
that matched its own way would answer a different set from the search box two
keystrokes away.

The two differ in one thing, and the list reads it off the field: a task hangs
under one parent, so `⏎` takes what the keyboard is on and closes; it serves as
many objectives as it serves, so `⏎` ticks and the list stays open until `esc`.
The field then reads `#1 Ship it` rather than `1` — the number is what the
engine takes and what a person has to go and look up, which is the whole of why
the list exists.

The row the keyboard is on is lit, and the row is what is lit. It used to be a
chip in the form's *name* column while the title was drawn in the *value*
column, a hundred and fifty pixels away — a list borrowing a form's two columns
when a list wants one.

**An open task is headed by its whole title.** The panel's heading said `task
243` and left the rest of the line empty; it now reads `task 243 · <title>`.
That line holds 99 characters at the smallest window this board draws in, 166
at the largest, where the title *field* below it cuts at thirty — so a title
worth reading had nowhere to be read. Not while the task is being edited: the
title is in the field being typed into, and a heading repeating it would go
stale between two keystrokes.

Measured before it was written, against three window sizes. Widening the field
or splitting it over two rows were the other two ways, and both moved every
field below them to buy a line that was already there.

**A card says what it holds and does not show.** `≡` that there is a
description to read, `▪ 3/8` how far the parts under it have got — beside the
number, where an objective's count already sits.

Those two and no more. The terminal compresses ten facts into marks because a
column is 24 characters wide; a window says most of them another way, and
saying them twice would be noise: where a card sits gives blocked, watching and
paused, its edge gives priority, the dot on its number gives the focus, and the
labels are written out rather than flagged.

The note this replaces held that only 2 of the 10 shapes rendered in the font
Gio ships, and parked the whole idea on the weight of a font that would have to
be embedded. Drawn and looked at, nine of the ten render; only `⚑` comes out an
empty box, and a window has no use for it. Measurements go stale — this one had
been true enough to stop the work.

## The detail

The panel a task opens into, with `s`, `enter` or a double click. Two columns:
what the task carries on the left, in the same boxes the form writes into, and
its history on the right.

**Two columns and not one above the other.** Stacked, twelve framed fields take
nearly the whole panel at the smallest window this board draws in, and the
history gets what is left — which is nothing. Beside them it has the panel's
full height at every size.

The history **scrolls**, by the wheel over it, by `↑↓`, or by pulling the bar
beside it. It is counted in lines and not in acts, because an act wraps onto as
many lines as it needs and a history that jumped an act at a time would skip
past what it wrapped. Acts are told one after another with a blank line
between two of them: without it, a wrapped act and the next one beginning
straight after read as one act nobody could tell the end of.

The bar is drawn whether or not there is anything to scroll. A bar that comes
and goes is a bar somebody reaches for and does not find. Its grip says two
things at once — how tall it is, how much of the history is showing; where it
sits, which part — and it is never drawn smaller than can be taken hold of.

**What is under a thing is drawn in one box**, under the fields of whatever is
open: the tasks that serve an objective, or the tasks hanging under a parent.
One box and one question — what is under this — with the heading saying which
was asked, `WHAT SERVES IT` or `WHAT HANGS UNDER IT`. Calling both by one word
would make the box say less than it knows.

An objective nothing serves says so; a task nothing hangs under says nothing.
Most tasks have no sub-tasks, and a line saying so on every one of them stops
being read by the third time.

A task's twelve fields leave that box far less room than an objective's seven:
120 pixels against 280 at the smallest window this board draws in, which is one
card against two. It grows with the window, and the remedy if it proves too
tight is to move the list rather than to shrink what it holds.

**An objective opens on the work that serves it**, in that box, headed
`WHAT SERVES IT`. They are drawn as **cards** — the same cards the board
draws, at the same height, scrolled a card at a time by the wheel over them or
by `↑↓`, with a bar beside them. A run of text lines was what it did first, and
it made the one place that lists tasks the one place a task does not look like
a task.

The right column stays the history's. Taking it for an objective was the other
way round, and it cost every task the thing that column was built for; an
objective carries one line there saying it has no history of its own, because
the trace is kept of tasks. An objective nothing serves yet says so in the box
in as many words — an empty frame reads as a box that failed to load rather
than as a question nobody has answered yet.

It opened on a history before, and on the wrong one: the panel asked for the
trace of the chosen number without asking which kind of thing carried it, so
opening objective 1 showed whatever had happened to *task* 1 — a stranger's
history, presented as the objective's. Measured, not reasoned about.

**An objective's card says how much it carries and how much is done** —
`7/12 · 58%`, on the number line it otherwise leaves half empty. The share is
rounded down, so an objective one task short of finished does not read 100%.
An objective nothing serves carries no count at all: `0/0` and a share of
nothing are two ways of dividing by zero, and `0%` on an objective nobody has
started reads as a failure rather than as one waiting for its first task.

Editing in place on the card was the alternative, and is what a mouse invites.
It was turned down for the trace: with no single moment of saving, every field
touched is its own `task edit`, and twelve fields touched is twelve lines of
history for one sitting.

## How it is tested

Measured: a Gio window with no screen draws and hands its pixels back, in
0.14 s. So the gates can reach this code. But **pixels are not what gets
pinned** — the terminal pins characters, which are the same everywhere, while a
pixel depends on font rasterisation, the Gio version and the graphics driver. A
frame pinned in pixels would go red on another machine for something that is
not a defect.

Three layers, each with its own gate:

| what | how | what it catches |
|---|---|---|
| the **geometry** — where each card is, which one is under the pointer, what is off screen | a pure function `(width, height, cards) → rectangles`, pinned as **numbers** in testdata | every layout sum, under mutation |
| the **decisions** — what a click chooses, what a drop sets off, which operations go to the engine | state kept in one struct of ours, changed by a function a test can call without drawing, and read back off the `*fake` | dragging, the two operations, every refusal |
| the **drawing** — that there is ink where there should be, in the right colour | a window with no screen, and assertions on one colour at a time: this rectangle holds some of the colour a chosen card is outlined in, that one holds none | having forgotten to draw something, and having drawn it in the wrong state |

**Nothing in the drawing works anything out**, and that is not tidiness — it is
what the mutation gate forced. Arithmetic inside a drawing is arithmetic no
assertion can judge: a title moved six pixels still looks like a title, and the
only test that would notice is the pinned pixel frame ruled out above. So every
rectangle, down to the four edges of an outline and the row a field's name sits
on, is worked out in `parts.go` where a test pins it as a number, and `draw.go`
is left with nothing it can get wrong but a colour.

Two things measured while building it, both worth keeping: counting "pixels
that are not the background" says nothing about a card, because a card fills
its own box — what discriminates is counting one colour. And the first drawing
test that passed was passing for the wrong reason.

The third does not judge whether it looks good and does not pretend to. Nobody
can judge that but a person looking at it, which is why the definition of done
asks for the built binary rather than an argument about it.

**The middle row is the one the library fights**, and saying so now is cheaper
than discovering it. Gio is immediate mode: there is no `Update` handing back a
new model the way bubbletea does, and a widget that keeps its own state is the
path of least resistance. So the state a decision turns on is kept in one
struct of ours, changed by our own function, and only then drawn — the widgets
hold what they must and decide nothing. Letting that slip is how this surface
would become the one the gates cannot reach.

## Deliberately not there yet

- **A permanent detail panel**, which the extra room would allow and which
  would stop this being the same board.
