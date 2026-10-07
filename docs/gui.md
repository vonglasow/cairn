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

## Read this first

The whole file in eight lines, for whoever has a change to make and not an
afternoon to spend:

- **The same board as the terminal's**, in a window: where the two disagree
  about what a board means, [tui.md](tui.md) is right.
- **The mouse comes first**: no modes, no cursor to learn. A key does what the
  bar offers, where the bar offers it, and a few flat shortcuts stay.
- **A drag is an edit**: dropped in another column it changes the state,
  among its sisters it changes the place — the engine's two operations.
- **Sub-tasks are not on the board**; they are lines of the form of the task
  they belong to, and deleting a task takes them with it.
- **Every size is worked out from the window** and pinned as a number in
  `parts*.go`, never in a drawing, where no test could judge it.
- **What a board reads and asks is shared** with the terminal in
  `internal/reading`; the window only draws and points.
- **It is tested without a screen**, by gestures on a `Window` value; the few
  tests that draw need a GPU, and a machine without one says so with
  `CAIRN_WINDOWLESS=1`.

## Contents

- [What it is for](#what-it-is-for)
- [Why Gio](#why-gio)
- [Where it lives](#where-it-lives)
- [The window](#the-window)
- [The mouse](#the-mouse)
- [The form](#the-form)
- [The detail](#the-detail)
- [How it is tested](#how-it-is-tested)
- [Deliberately not there yet](#deliberately-not-there-yet)

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

**Two dependencies, and their written reasons are in
[dependencies.md](dependencies.md)** — `CLAUDE.md` asks for one: `gioui.org`,
the only way to a vector-drawn window that is identical on macOS and Linux, in
a single binary, with no system webview and no second language; and
`golang.org/x/image`, for the fixed-point numbers Gio measures text in.

## Where it lives

`internal/gui`, a fourth surface beside `cli`, `mcp` and `tui`. `depguard`
denies it `store`, `database/sql`, `cobra`, `cli`, `mcp` — and `tui`: two
surfaces do not reach each other, and `tui` and `mcp` are denied it in return.

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

The right-hand end of that third line carries **what is due, what the work has
cost against what it was reckoned to cost, and what it serves** — `2d  9h/2h
◇4` — in the faint ink, or in `Skin.Late` when the day has passed or the
estimate is spent. The window used to know nothing at all about time: no
deadline, no notion of late, where the terminal has written it on the card all
along, and deciding what to do without knowing what is due is not deciding. A
day more than a week out is left off, for the terminal's reason: it will still
be there tomorrow and in ninety days, and saying so on every card every day
buys nothing. A blocked card waiting on nothing any more says that instead, and
crowds the rest out — it is the one thing about it that has to be acted on.

**An objective's card says where it stands**: `▶ active`, `‖ on-hold`,
`✓ reached`, `✕ dropped`, written where a task writes its labels. Objectives
were drawn exactly like tasks, so a list mixed what is running with what was
abandoned and nothing on either said which was which.

It measures with `perSmallGlyph` and not `perGlyph`. Labels are written in the
small face while the title uses the wide one, and measuring with the title's
gabarit left a fifth of every label line empty — on a card carrying three, the
third was dropped for room that was there.

**The spacing is one scale, in steps of four**: 4 inside a frame, 12 between
two things, 32 from one line of anything to the next. A panel's rows and a
form's fields used to advance at 26 and 32 — the same text at two rhythms, so
the help panel and the form beside it read as two pieces of software.

**Every size is a value worked out from the window**, redone on every resize,
exactly as `LayoutFor` does in the terminal. Nothing carries its own size. That
is what makes the geometry testable with no screen: a pure function from
(width, height) to rectangles, compared against pinned numbers.

In **dp** rather than pixels, so a Retina screen and a 96-dpi Linux screen give
the same physical thing.

**Too small**: below a threshold the window says what it needs instead of
drawing something unreadable — the terminal's answer, for the terminal's
reason. What was open is remembered rather than closed: a window dragged narrow
for a moment used to come back as an empty board, taking with it a form
somebody had been filling in. Imposing a minimum window size was the alternative and hides the
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
opens a card's detail. The wheel scrolls the column, or the lane, under the
pointer.

**Only the primary button acts.** A right click, a middle click or a side
button used to run the same press and release as the left one, so they chose
cards and pressed offers — and no context menu exists to have meant it.

**What the pointer is resting on is lit.** The window answered nothing before
it was pressed, which made "Delete" read exactly like a label: the one
affordance that says a word can be clicked was missing from the whole surface.
Gio is asked for `Move`, `Enter` and `Leave`, the window keeps which offer the
pointer is over, and the drawing decides what that looks like — a value, like
everything else here, so a scenario can ask.

**A turn of the wheel belongs to what it began on.** A trackpad goes on
sending after the fingers leave it, and a task put away with `esc` while it was
still sending handed the rest of the turn to the board: the column under the
pointer scrolled though nobody had turned anything over it. A part of a turn
that comes within 150 ms of the last one goes on the same gesture, and is
dropped if what the gesture began on has gone; a part after a pause is a turn
of its own. Gio reports no phases for a gesture, so the pause is what says one
has ended.

**Every bar can be pulled.** The history's, the description's, the work under a
task, and the list's: all four have a thumb sized never to be too small to
grab, and two of them used to ignore a drag entirely.

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
within the same few pixels are one gesture, and a third one more is a triple
click — unless the double opened something, in which case the third press
lands on what it opened and is a press there.

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

**A lane scrolls, on a number of its own.** It was left still on purpose — two
rows of stopped work, and a third read as a sign to go and unblock something
rather than to scroll. What a lane held past its second card was then out of
reach: the wheel did nothing over it and the arrows chose cards nobody could
see. It now scrolls the way a column does, between the same two ends.

## The form

**A part of the work is reached in the form of the task it belongs to, and
nowhere else.** The form carries a field holding one line per part — `#2 [x]
Relire l'ADR` — which is where one is ticked, renamed, added and taken away.
Saving settles them one at a time against what the task already carried: a line
with a number is that part, a line without is one nobody had written down yet,
and a number the field no longer carries is a part that has gone. Each is an
add, an edit or a delete through the engine; the engine has no argument for a
set of parts and should not grow one.

The board forced it. Parts left the board when a board of 219 tasks drew 219
things of which 170 were the work and the rest its pieces — and once off it, a
task carrying one could neither be changed nor deleted, because the deletion was
refused while a part was there and nothing was left that could remove the part.
**Deleting a task now takes its parts with it**, which is the rule archiving
already went by: a part goes off the board with the work it belongs to.

A part carries a title and whether it is done, and nothing else. Measured on a
store of 219 tasks, its 49 parts were done 47 times and todo twice, with no
deadline, no label, no priority and no blocker among them.

A modal dialog over the board, with *Save* and *Cancel*; `esc` cancels. One act
in the trace per save.

**The fields are not listed here.** They come from `reading.TaskFields(card)`
and `reading.GoalFields(goal)`: the word a person reads, the closed set when
there is one, the field that wants more than one line, and `Arg` — what the
engine calls the same thing. An attribute added to the engine and to the
terminal's form arrives here with no line written.

A field with `Choices` is chosen from the same list as a parent, a `Long` field
a multi-line box, the rest a line to type in. The field the keyboard is on is lit, and it is the only one:
that is the whole of what says where typing would go.

**The objectives a task serves are picked from a list, not typed.** `enter` on
that field opens the objectives over the form — over, and not instead of: a
choice made with the rest of the form out of sight is a choice made from
memory. The arrows walk it and so do `j`/`k`; **enter** takes the row the
keyboard is on and closes the list, **⌘⏎** takes it and keeps the list open for
the next; `esc` closes it. Nothing reaches the store until the
form is saved, the way nothing else in a form does.

`enter` never saves a form: it opens the list a field answers with, and it
breaks a line in a field with room for more than one. **⌘⏎** (or ⌘S) saves,
wherever the keyboard is. So the bar offers *Save* on every field, and adds
*Choose* beside it on a field a list answers — it used to give way to *Choose*
there, and with the keyboard left on that field after the list closed there
was no way to save at all. A key does what the bar offers; the bar says both.

**The field carries numbers and is drawn as names.** `1 2` is what the engine
takes, and a surface inventing its own spelling of an argument would be a
surface deciding something; `#1 Ship the desktop board, #2 Keep it honest` is
what a person reads, because nobody knows an objective by its number and the
one place that said what a number meant was the other tab.

**Every field is drawn in a box**, whether or not the keyboard is on it, and
the lit frame is the only difference between the two. A value floating on the
panel's own ground looked like something written rather than somewhere to
write.

**A form reads from the top, and its offers keep a line of their own.** The
title first, across the whole width and as large as a panel's heading; the
description under it, across the width, a line for each it says — four at the
least, ten at the most; the short fields grouped a section to a row, the head
of the section first and each field's name before its box — `STATE`, `TIME`,
`BLOCKED`, `LINKS`, by what each field asks the engine for; then the parts.
Under a rule at the foot, and nowhere else, is what the form offers. The panel
is as tall as what it holds, up to the panel every other screen uses: it
stops under the parts rather than under a floor of empty ground. `FormPlace`
says where every field of a form goes, and everything that draws, clicks or
measures a form asks it.

It was one column, and the parts — a second long field — ran over *Save* and
*Cancel* in a window under eight hundred points. Then two columns, which left a
hole under eleven fields and two frames as tall as the panel with three lines
in them. When the window is short, the parts are given their room before the
description grows past four lines: a part out of sight is a part nobody ticks.

**The parts read as a list to tick while the keyboard is elsewhere** — a tile
each, ticked or not, from what the field says now, unsaved lines included, and
a line saying another is added by typing one. With the keyboard on them they
are the lines being typed, `#2 [x] Relire l'ADR`, which is where a part is
ticked, renamed, added and taken away.

**A new task goes in with its parts.** The form that adds a task carries the
same parts field as the one that edits it, empty: a task whose parts were
written down beforehand is added with them, each line a part under the task
just added. The terminal's form does not carry it yet. Parts are never settled
under number 0, which is no task — every task with no parent would read as a
part the field no longer names.

**Every box in a row ends where the others do.** The title, the description,
the last field of each section and the parts end on the same line, and a long
field's bar sits beside its box, in a margin every field keeps.

**Nothing written runs out of its box.** Text wraps — at a space when there is
one, inside a word when the word is longer than the room — and the box is
clipped, so even a line of unusually wide characters cannot escape it. Cutting
a value down to one line was the alternative and is what this used to do; it
said nothing about what was lost, and a description is the field it lost most
of.

**A line ends where the face runs out of room**, measured in the face it is
drawn in, and not after a number of characters reckoned for the widest of
them: counted, a description stopped two thirds of the way across its box.
`Window.wrapping` is the one answer for drawing a text, scrolling it, putting a
caret in it and holding a run of it, so the four agree on where a line ends.
How many rows a description asks for is still reckoned — the form is laid out
with no face — which can leave it a line or two longer than it needs.

**The caret is a shape, not a character.** Written into the text it took a
letter's place and pushed everything after it along, so typing looked like the
text was being rewritten under the hand. It is drawn between two letters, at a
measured x, on the line it actually fell on — placed from the height of a
whole wrapped label it lands on the last line whatever was typed.

**What the arrows reach, they bring into sight.** Walking used to move the
choice and leave the column where it was, so a card could be chosen below the
bottom of its column — and a card chosen out of sight is a card whose title,
whose state and whose place are all unknown to whoever chose it. Walking down a
column and walking across into one somebody had scrolled both follow, and so
does walking along a lane of the band.

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

**A panel offers what can be done at its own foot; the bar under the board
stays the board's.** Under a panel it is drawn faint and answers nothing — it
says what the board will offer again once the panel is put away. It used to
change to repeat the panel's offers, which read as the board itself having
changed underneath, and made two places to click for one thing.

**Archive** is one of those offers, under `x`, beside *Edit* and *Delete* on a
chosen card. On a card already off the board the bar reads *Restore* instead:
one key both ways, and the bar says which way it goes rather than leaving it to
be worked out. Unlike *Delete* it asks nothing first — nothing is lost.

**A whole column goes off the board on `⇧X`**, the column the chosen card
stands in. It asks first and names the column: nothing is lost, but a key that
moves twelve cards is not a key to press and find out about. It is not on the
bar — the bar has room for what one card does, and this does what a column
does — so it lives in the help, under the terminal board's own key. The engine
has swept by a narrowing since the beginning; only the command line ever asked
it to.

**One screen asks the questions, and there are three.** Deleting a task and
throwing away a form somebody has typed into both lose work; taking a column
off the board does not, and is asked about for its size. *Agree* and *Cancel*
are the two offers, `⏎` agrees, and the targets are as tall as the bar's — they
used to be the smallest in the window, for the most destructive act in it.
A question is asked in a box as big as its sentences and its offers, not in
the panel a form is drawn in: two lines and two words used to sit in a box
nine tenths empty.

The question about a form is answered the way a hand leaving a form answers.
`esc` again throws it away and `⏎` goes back to it (*Throw away* and *Back*):
a second `esc` that sent somebody back into the form they were leaving was a
guard they had to fight. Nothing is lost for all that — what was typed is set
aside, and the form of the same task, opened untouched, offers *Recover*
under `⌘Z`. It is kept while the window is open, one form at a time. A form
nobody typed into closes without a question, because there is nothing to lose
and a question with no stake is noise.

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

**The history is written in the reading ink**, with only the timestamp left
faint. It was thirty lines of the palest grey in the window — a thing somebody
opened a task to read, written in the ink kept for the things beside what is
read. A line a long act wrapped onto carries no timestamp and is written whole,
because splitting one sentence across two inks makes it read as two.

**A refusal gets two lines and the reading size.** The most instructive thing
cairn says — "it is still blocked, by #3 and by #4. take them away first: task
3, task 4" — was the smallest text in the window, on one line, cut just before
the part that says what to do. The line grows upward, so the foot of the window
stays where it is.

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

**A closed set is chosen from the same list.** A column, a priority and an
objective's standing open it too — on a click, or on `⏎` — offering the set in
the domain's order, the value already held ticked and under the keyboard, and
narrowed the same way by typing. One answer, so `⏎` takes and closes, as for a
parent; but taking the one already held leaves it held, because a task stands
in a column whatever is done to a list. The rows say the names alone: the
number a row is listed under is the list's, not one anybody has seen. `←`/`→`
on the field still step through the set without opening anything.

**The list is as long as the panel, not as long as what it offers.** A board of
fifty tasks drew fifty rows, straight past the panel, past the window, and over
everything behind both. It is cut to the run the panel has and scrolled to for
the rest — by the wheel, by a bar beside it, and by the keyboard, which brings
the row it is on into sight the way a column brings a card. The same `Thumb`
and the same window of lines the history and the description use.

Counted in rows and not in pixels. The walk in pixels had to divide by a row's
height at the end, and a run rarely divides into whole rows: the remainder was
arithmetic no assertion could judge, and it survived mutation twice before it
was written out as rows.

**Text is held, carried, and put down again.** A run is selected with
`⇧←→`, with `⌘A`, or by dragging across it, and `⌘C` `⌘X` `⌘V` carry it.
The anchor lives beside the caret in `Written`, counted **from** the caret
rather than from the start: nought is nothing held, so every literal written
before any of this existed still holds nothing. Stored the other way round, a
`Written{Value, At}` meant "held from the beginning" and three tests fell at
once.

The run is measured with the font, the way the caret is. Measured at so many
pixels a glyph — the width a card's title uses — it sat a word away from its
own caret at the far end of a line.

**What a pointer lands on is measured with the same face.** The window is
given a `Measure` — the edges between a line's characters, shaped by the
theme's own shaper — and a press lands on the edge nearest to it. Reckoned at
an average width, the run painted under the pointer and the run held ran apart
as the pointer moved along a line, and holding text felt like dragging it
through glue. A test window with no face reckons, so no test depends on a font.

**A run grows by what began it.** A press grows it a character at a time, a
double click a word, a triple click a paragraph — and once the pointer is down
it goes on growing wherever the pointer goes, out of the field included: a
hand overshoots, and a run that stopped at the frame stopped short.

The clipboard goes through the loop. A window is a value and cannot execute
anything, so it says what it wants: `Copying()` is read and done, `Pasting()`
asks, and the answer comes back as an event like any other. What the system
hands back is read in `pasting`, outside the loop, so a scenario can reach it —
the loop decides nothing.

**One rule for a form, on both boards.** `⏎` writes a line in the
description and moves on to the next field everywhere else; it never saves.
Inside the description `↑↓` walk its lines and `⇥` is what leaves it. `⌘⏎`
saves from anywhere in the form, and so does `⌘S` — because the terminal board
cannot be sent `⌘` and saves on `^S`, and the same fingers should work in both.

Enter used to save, while the bar offered Save under `⌘⏎`: a key doing what the
bar does not say. And `↓` used to leave the description for the next field, so
the only way through a long text was a character at a time — which is why it
read as cut off.

**The description is the same box on an open task and in the form**: its
frame, its bar inside the right edge, the wheel, and a bar that can be pulled.
`sayingOn` answers for both screens, and the lines it counts are laid out on
the width they are drawn on. They were not: the end of the scroll was worked out
from the text laid out a few points wider than it was drawn, so what was drawn
ran longer than what was counted, and the last lines of a long description
could never be brought into the box — bar or no bar. What is in sight follows
the caret when the caret moves, and is left where the wheel put it otherwise.

**Save is always offered**, under `⌘⏎`. It used to give way to *Choose* on a
field a list answers, and since the keyboard stays on that field after the list
closes, there was then no way to save at all: the one act a form exists for,
unreachable for having chosen something.

**A list offers what is reached for first.** The board reads from what has not
been started to what is finished; work is hung under what is being worked on,
so the list reads the other way — current, then todo, then what is stopped,
then the backlog, then what is over. Opened in the board's own order it put two
dozen backlog tasks in front of the three anybody would pick.

Each row says which column its task sits in, hard right and faint. Fifty rows
that read alike are scanned by something other than the words.

**A click says which row, `⏎` takes it, `⌘⏎` keeps the list.** One click that took made reading a row impossible without choosing it, and one
key doing both taking and keeping left no word for being done with the whole
list. The bar says all three — *Take*, *Keep*, *Cancel* — and a double-click is
take-and-keep in one gesture, which is what a hand already does to a list.
`esc` backs out of the list and not out of the form under it: the form holds
what somebody typed.

Not space, which is the obvious key and the wrong one here. A list with a line
to type into cannot spend it: `Ship it` has a space in it, so the filter needs
the character and the shortcut cannot have the key. Taking a row and being done
with the list are two acts, and the second takes the modifier.

The line that narrows a list is corrected like any other text here, backspace
included. Without it a mistyped letter meant closing the list and opening it
again.

**A long field takes a line break under `⇧⏎`**, since `⏎` saves the form. And
`Wrapped` breaks where the text breaks, which it never did: a description
written in paragraphs came back as one run of words with the breaks still in
it. Both halves were needed — there was no way to write a paragraph, and no way
to draw one.

Clicking the field opens it. Neither that nor the mouse in the list was true
when the list first went in: `Click` had no branch for the list, so a click
meant for a row fell through to the form drawn behind the panel and moved the
keyboard to whatever field was there. The list answered the keyboard and
nothing else, which reads as a list that does not work at all.

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

The panel a task opens into, with `enter` or a double click. Two columns:
what the task carries on the left, in the same boxes the form writes into, and
its history on the right.

**Two columns and not one above the other.** Stacked, twelve framed fields take
nearly the whole panel at the smallest window this board draws in, and the
history gets what is left — which is nothing. Beside them it has the panel's
full height at every size.

The history **scrolls**, by the wheel over it, by `↑↓`, or by pulling the bar
beside it. It is counted in lines and not in acts, because an act wraps onto as
many lines as it needs and a history that jumped an act at a time would skip
past what it wrapped. Acts are told **the last done first**, one straight
after another: each begins with when it was done, which is what says where it
starts. They used to be told oldest first with a blank line between two of
them, which put what was just done at the bottom of a box opened at its top,
and read as an empty line between every line of a history whose acts each fit
on one.

The bar is drawn whether or not there is anything to scroll. A bar that comes
and goes is a bar somebody reaches for and does not find. Its grip says two
things at once — how tall it is, how much of the history is showing; where it
sits, which part — and it is never drawn smaller than can be taken hold of.

**What is under a thing is drawn in one box**, under the fields of whatever is
open: the tasks that serve an objective, or the tasks hanging under a parent.
One box and one question — what is under this — with the heading saying which
was asked, `WHAT SERVES IT` or `WHAT HANGS UNDER IT`. Calling both by one word
would make the box say less than it knows.

**Where what was written about it elsewhere lives is said at its foot**, beside
the offers: `context  ~/.cairn/notes/14`, faint, and held and copied like any
other text. Not among the fields, which is where it was first put: the place is
not something the task carries, and a thirteenth field took a row from the box
of its parts — measured at 1200 by 800, three tiles down to two. The foot had
the room and lost nothing to it. See [context.md](context.md).

**Every piece of text an open task writes can be held and copied** — its
heading, the name and the value of every field, the description, the box under
it and each part in it, the history and each of its lines. A drag across it
holds a run, a double click holds a word, `⌘A` holds the whole of the text last
pressed in, and `⌘C` copies. A part is the exception for the drag, which still
moves it among its sisters; a double click holds a word of it.

A window draws text as shapes, and nothing said where a shape's letters were.
`Readables` now lists every piece of text an open task writes, once, laid out
the way it is drawn, and what a pointer lands on and what is painted held are
both worked out from it. A field of the form holds a word on a double click too.

A piece of text answers the pointer inside the box it is drawn in, and nowhere
else. It used to answer over every line from the first in sight, drawn or not,
so a long description reached down to the foot of the panel: the place of the
notes could not be held, and `Edit` and `Close` took a press as the start of a
run rather than as a click.

**What is left reads first, and reads as left.** A part carries a box before
its name — empty, or filled and ticked — and what is left is written in the
reading ink while what is done is faint and struck through; what is left comes
first, what is done after, each in the order it was put in. The heading counts
both: `WHAT HANGS UNDER IT   2 to do · 1 done`. The ink used to be the only
difference, and the other way round — done bright, to do faint — so a part to
do read as one finished with. A box is read by its shape, without comparing two
greys. The terminal says the same with `[ ]` and `[x]`, the marks the form's
own field writes.

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
