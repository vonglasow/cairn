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
  within its column it changes the place — the engine's two operations.
- **Steps are not on the board**; they are tiles of the form of the task they
  belong to, and deleting a task takes them with it.
- **Every size is worked out from the window** and pinned as a number in
  `parts*.go`, never in a drawing, where no test could judge it.
- **What a board reads and asks is shared** with the terminal in
  `internal/engine`; the window only draws and points.
- **It is tested without a screen**, by gestures on a `Window` value; the few
  tests that draw need a GPU, and a machine without one says so with
  `CAIRN_WINDOWLESS=1`.

## Contents

- [What it is for](#what-it-is-for)
- [Why Gio](#why-gio)
- [Where it lives](#where-it-lives)
- [The look](#the-look)
- [The design system](#the-design-system)
- [The window](#the-window)
- [The mouse](#the-mouse)
- [The form](#the-form)
- [An open task](#an-open-task)
- [Series](#series)
- [Several at once](#several-at-once)
- [The board arrives by an intro](#the-board-arrives-by-an-intro)
- [Settings changed in place](#settings-changed-in-place)
- [Workspaces, a window each](#workspaces-a-window-each)
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

**What it does not write again**, because `internal/engine` already carries
it: `Look` and the two listings turned into cards and goals, `Deal` into
columns, `Stopped`, `Deadline`, `Sorted`, `TaskFields`/`GoalFields` with each
field's `Arg`, `Asked` turning a saved form into what an operation takes, and
the `Work` interface — `Ask` and `Version`, two methods.

What is its own: geometry in pixels, drawing, and the mouse.

**What a panel keeps lives with the panel.** `Window` is one value, and every
panel added over the board used to lay its fields flat into it, so seventy
fields read as one list and a panel's own state was found by grepping. The
calendar's day, the menu's lines, the list of backups, the system's file
dialog and the time page's week are each a small struct in the panel's own
file (`calendarState`, `menuState`, `backupsState`, `filesState`,
`timeState`), embedded in `Window`. What is not split yet: each panel still
answers through the same four switches — `Choose`, `Offers`, `Click`,
`panel` — so a new panel touches all four.

`cairn gui` is a command in `internal/cli`, built like `cairn tui`, with the
same seam (`var openGui = func(…)`) so a test never needs a window.

## Opened as an application

`cairn app install` makes the window an application like any other, on a Mac
and on a Linux desktop. `engine.AppInstall` chooses which; elsewhere it is
refused (`no-application-here`). A cairn run by `go run` is refused on both
(`passing-build`): its binary is thrown away when it ends, and the application
would open nothing. Either way the application opens the installed cairn's
window rather than holding a copy of it, so it follows the cairn on the
machine: `make install`, a release, and the next opening is the new one — the
lag the MCP server has after an update (#227) is not repeated here. Opened
from the desktop it has no shell environment: no `CAIRN_STORE`, so it takes
the store the configuration file names, or the default place. The answer says
where it was written, which cairn it opens, and where to open it from.

**On a Mac** it writes `cairn.app` in `~/Applications` (`internal/macapp`,
`@app-installed`): an `Info.plist`, the icon made of the logo's PNGs at every
size macOS reads, and a launcher of two lines that runs the installed cairn's
window. It is signed for this machine alone (`codesign --sign -`), so macOS
opens it without a developer's certificate; handing it to somebody else would
need Apple's, and is not done. Installing again replaces the application cairn
wrote, and an application of the same name cairn did not write is left alone
(`app-not-ours`).

**On a Linux desktop** it writes the same entry and logo the window writes
when it opens (`internal/linuxapp`, `@app-installed-on-linux`):
`cairn.desktop` under `$XDG_DATA_HOME/applications`, or
`~/.local/share/applications`, following the freedesktop Desktop Entry
specification, and the logo at 16 to 512 pixels under `icons/hicolor` beside
it. Until then the menu knew cairn only once `cairn gui` had been run from a
terminal; now it can be the first way in. Then it runs `xdg-desktop-menu
forceupdate` when xdg-utils is there, so a menu that does not watch the folder
is rebuilt; a desktop without it is not refused
(`@app-on-linux-without-a-menu-tool`): GNOME and KDE watch the folder on their
own, and the entry is what matters. Installing again rewrites only what changed
(`@app-installed-on-linux-again`). An entry of the same name somebody else
wrote is left alone (`entry-not-ours`,
`@app-on-linux-an-entry-of-somebody-elses-is-kept`) — the command cannot ask,
as the window does, whether to write over it. The entry starts cairn by the
full path, as the window's does (`Starter`, below).

## The look

The window follows one design, settled on 2026-10-31 for v0.10.0: calm,
told apart by tone rather than by lines. Its tokens are `Dark()`'s, grouped
the way a design system groups them (`skin.go`): `Colour` (surface, line,
ink, accent, state, and the maps a name picks from), `Space`, `Radius`,
`Size`, `Type`, and the `Shadow` a raised thing casts.

| token | is | value |
|---|---|---|
| `Surface.Back` | the window | `#0c0d0f` |
| `Surface.Ground` | a column's or a lane's ground | `#17181b` |
| `Surface.Card` / `Hover` / `Tint` | a card, under the pointer, in focus | `#1f2024` / `#25262b` / `#1e243c` |
| `Surface.Raised` | anything raised: a panel, a menu, the command bar | `#26282d` |
| `Surface.Field` / `Control` / `ControlLit` | an input, a label or a button, a button lit | `#131417` / `#303238` / `#42444a` |
| `Line.Rule` / `Line.Raised` | a rule or a frame on the board, a rule on something raised | `#28292e` / `#34363c` |
| `Line.Edge` | the edge of a control | `#60626a` |
| `Ink.Text` / `Faint` / `Low` | text, secondary text, numbers and counts | `#e7e7ea` / `#9a9ba2` / `#696a71` |
| `Accent.Base` / `Line.Selected` | the accent (focus, primary button) / the keyboard's ring | `#7b93ff` / `#8d8f9c` |

A colour that used to be worked out where it was drawn — a shade lighter
under the pointer, a shade darker for a key's chip, the accent at a fraction
of its opacity for a field's halo or a card's glow, half the veil behind the
command bar — is a token of its own (`ControlLit`, `ChipOnControl`,
`Accent.Halo`, `Accent.Glow`, `VeilLight`…), so a second skin is a second
value and not a search through the drawing. Only a label's pill is still
mixed, a fifth of its colour over a card (`Skin.LabelPill`): there are
fourteen. **No drawing names a colour or a corner of its own**:
`TestNoDrawingNamesAColourOrARadiusOfItsOwn` reads every file of the
package but `skin.go` and fails on a `color.NRGBA{…}`, a number in a radius
or a stroke's width, or a constant named for a radius. What a layout lays
out — a gap, a height, a text's size — stays the layout's constant, since
the layout measures with it.

**A rule on a panel can be seen.** The line over a panel's offers, under
the filters' head and foot, under the command bar's line and over the
switcher's acts was drawn in `Line.Rule` on `Surface.Raised`, a contrast of
1.02: it was there and nobody could see it. It is drawn in `Line.Raised`,
held at 1.2 against a panel as a column's edge is against its ground
(`@gui-a-panel-rule-can-be-seen`).

**The window is drawn of components**, each drawn one way: see
[the design system](#the-design-system). Where two screens drew one
element two ways, the component draws it the one way the design settled
on (#230's sheet), and the window changed with it:

- a popover — a menu, a bubble, the swatches, a field's tip, the switcher —
  rounds at 10, a panel's corners at 12; the bubbles were at 8 and the
  switcher at 12 (`@gui-the-switcher-is-a-popover`,
  `@gui-the-labels-bubble-is-a-popover`);
- a priority's stripe is square along the edge it is attached to and
  rounded at 3 elsewhere, a free-standing one at 3 all round: a month's
  line gains its rounding, a list's and the form's go from 2 to 3;
- a label taken in a list is one chip with a stroked cross in its own ink,
  no longer a glyph `×` on a circle of its own;
- the count of the labels a card leaves out is a pill like any other, its
  ring gone, and every pill, an objective's standing too, is centred by the
  ink of `Hl` (`@gui-the-count-of-labels-is-a-plain-pill`);
- a key always stands on a chip: on a control's chip (`ChipOnControl`) on
  a button or on a picked line — the command bar's, the menu's, the
  switcher's — on a control's ground elsewhere, and the command bar's
  shortcuts gain theirs (`@gui-every-key-of-the-command-bar-is-on-a-chip`,
  `TestAKeyOnAPickedRowStandsOutFromTheRow`);
- a colour beside its name is a dot of 8 everywhere (`dotSide`); a label's
  dot in the labels and in the label filter was 12
  (`@gui-a-dot-beside-a-name-is-one-size`);
- an empty slot's dashed line rounds at a card's 8, drawn by the same
  `Dashed` as the dashed pill (`@gui-an-empty-slot-is-rounded`).

The controls, the rows and the small parts followed. Where the screens
disagreed, the sheet of #230 chose again:

- one segmented control (`Segmented`): its track rounds at a control's 7
  and a hair (`Space.Hair`, 2), its chosen segment inset a hair all round
  — the week's ranges were inset at the top and foot only, the details'
  tabs and the time's setting rounded their track at 7; a segment may count
  what it holds;
- a row writes its words in the reading ink in every list, the command
  bar's too, which wrote a line not picked faint; what is secondary — a
  time, a count — is faint at its right. A key or a pill on a row is
  grounded by the row (`Row.Chip`, `Row.Pill`), never by a flag set at the
  call (`@gui-every-line-of-the-command-bar-reads-alike`);
- one field (`Field`): a field's ground, a rule, rounded at 7, and lit with
  its halo while written into — the labels' search and their rename gain
  the halo, the switcher's line too (`@gui-a-label-search-is-lit-with-a-halo`);
  the square frames of a list's search, the steps, the history and the
  tasks serving an objective become a field at rest
  (`@gui-the-steps-stand-in-a-field`);
- a box before a step or a value (`Check`), rounded at 3: empty, a line of
  an edge; done, filled faint; taken, filled with the accent. A filter's
  values and every list chosen from — the labels, a column, a priority, an
  objective — gain the box where a bare `✓` stood, the row's words no
  longer carrying it (`@gui-a-taken-value-is-a-filled-box`,
  `TestEveryChoosingListShowsABoxWhereItWroteAGlyph`);
- one track for what is done (`Progress`), 3 high: an objective's form
  drew it at 4, and sized the box it stands in a point higher than its
  track too;
- a scrollbar with every line in sight is not drawn at all, as a day of
  the week already left it out: the description and the steps of a form
  showed a thumb as high as their track;
- two rings (`Ring`), always at the element's own radius, never square:
  what the keyboard is on in `Line.Selected` — a card, a step, a serving
  task, a month's line — and a target, written into or where a drop lands,
  in the accent. A step and a serving task the keyboard is on were edged in
  the accent, square;
- today is a ring in the accent wherever days are laid out (`TodayMark`),
  a fill meaning a day picked: the week's head filled today's date, the
  calendar ringed it in `Line.Selected` (`@gui-today-is-a-ring`);
- the pill of the workspace is ringed in the accent while its switcher is
  open (`@gui-the-workspace-pill-is-ringed-while-open`);
- one head (`Head`): 12, at 500, faint, in sentence case, its count in mono
  and low at its right — the help's heads were at 600, the command bar's
  groups in mono capitals and low, the switcher's and the filters' too,
  and `Details`, `Serving it` and `Progress` were in capitals at 400;
- a step's line is a tile rounded as a row is (`Tile`), as a serving
  task's was; a box of facts (`InfoBox`) is a card's ground at a card's 8,
  where the progress of an objective's form rounded at 10.

The form's section heads, `STATE`, `TIME`, are not a `Head` and stay as
they were: a scenario names them in capitals.

The design's own mock-up was darker still; `skin_test.go` holds every
shape to a contrast it can be told apart at (a panel 1.3 against the
board, an edge 3), so the tones were raised to meet it rather than the
thresholds lowered. Shapes are rounded (`rounded`, `Radius`): a card 8, a
column 10, a panel 12. A raised panel casts a shadow (`raised`, `Shadow`):
Gio has no blur, so the shadow is three rounded boxes under it, wider and
fainter as they go.

**A panel is raised, and its offers are buttons at its foot.** Every panel
— a form, the help, a question, the backups, the labels, the
configuration — is a rounded (12), shadowed ground over the veil, with no
edge and no rule under its heading. Its heading is its name — Backups,
Labels, Configuration — and what its keys do is said by its buttons and the
help: headings like "Labels — r renames one, d takes it off everything"
said it a third time, longer than the panel was wide (`HeadsOn`). A panel that can be closed carries a
cross at its top right that does what its `esc` does (`PanelClose`,
`@gui-every-panel-is-closed-by-its-cross`); a question has none, it is
answered. Its offers stand at its foot, right-aligned under a faint line,
the panel's verb last, at the right: filled with the accent when it is what
the panel is for, red when it destroys — Delete, Throw away, the yes of a
question that deletes — and plain otherwise (`OfferLook`). An open task's
context is written at the left of that foot, up to its first button. Every
button is one shape (`Button`): a rounded (7) ground and its words in
the middle, lighter under the pointer. A panel's button says its word, then
its key on a chip a shade darker than the button, `Cancel esc`, `Save ⌘⏎`,
as the command bar's hints show a key (`Button{Keyed}`, `ButtonParts`,
`TestAPanelButtonWritesItsWordThenItsKeyOnAChip`); it wrote the
key first, as plain text, where nothing told the key from the word. A key
mark Go Mono has no glyph for (`⌘`, `⏎`, `⌫`) is given an em of room on its
chip (`KeyChipWide`): reckoned as Go Mono, `⌘⌫` was cut to `⌘…`. A key is
centred on its chip by its ink, which the shaper measures, not by its line
(`sayOnChip`, `TestAKeyIsCentredOnItsChipWhateverItsGlyphs`): the face that
draws `⏎` sets it high, and `⌘⏎` stood two points above the middle of its
chip.

Words are written in Figtree, at 400, 500 and 600 (`TextFace`, `Faces`);
Go's faces follow it for what it lacks, then the symbols. Figtree's strokes
are thinner than Go's: no pixel of a word is the ink's exact colour, so a
test looking for words counts the pixels near it (`lettered`), and one
looking for a shape still counts them exactly (`painted`). A line is laid
out a little taller than its box and clipped to it (`textSlack`), Figtree
standing a point taller than the line Go was given.

## The design system

The window is drawn of a small design system (#230): tokens a skin fills,
components drawn from them, and screens that only place components. A new
screen assembles what is here; it re-draws nothing.

**Tokens** are `Skin`'s fields, grouped as a design system groups them
(`skin.go`, the table in [the look](#the-look)): `Colour` — `Surface`,
`Line`, `Ink`, `Accent`, `State`, and the maps a priority, a standing or a
label is looked up in — then `Space`, `Radius`, `Size`, `Type` and the
`Shadow` a raised thing casts. `Dark()` fills them; a component reads
`k.skin` and nothing else for a colour or a corner.

**Components** are values carrying their look, each drawn by
`Draw(k Kit, at Box) Box` in a `component_*.go` file, as Gio's
`material.ButtonStyle` is:

| component | is |
|---|---|
| `Backdrop` | the window's own back, under everything |
| `Ground` | what a column, a lane or a day stands on, raised for a day |
| `Panel`, `Heading` | a raised box over the board, and its title, with the number of what it shows beside it when it has one — a series' behind its loop (`Loop`) |
| `Popover` | a small raised box dropped over the rest: a menu, a bubble, a tip, the swatches, the switcher |
| `Veil` | what dims the board under a question or the command bar |
| `Rule` | a line between what something shows and what it offers |
| `Card`, `Stripe` | a task's card in each of its states — a reference saying why it is there (`Why`) among them — a sub-objective's in tile layout (`GoalTile`), and a priority's edge |
| `ListItem`, `Connector` | an objective or a series of the list beside a detail, two lines, moved right and joined to its parent below level one, a series' state on a pill and its rule and day on the second line; the line that joins, also in a list's row (`Row.Connector`) |
| `Facts`, `Figures`, `Banner` | names and their values; a row of numbers at the figures' size; a short message boxed in a panel, refused or warned |
| `LabelChip`, `Pill`, `KeyChip`, `Dot` | a label, a word on a rounded ground, a key on its chip, a colour beside its name, or the late ink's |
| `Button`, `Tabs`, `Segmented` | a word to click, the window's tabs, one choice among a few |
| `Row`, `Tile`, `Head` | a line of a list, a line inside a field, the line naming what follows |
| `Field`, `Check`, `Swatch` | what is typed into, the box before a value, a colour to pick |
| `Toggle`, `Picker`, `Strip`, `SchemeTile` | a setting turned on or off; a value chosen from a list that opens elsewhere, a dot or a scheme's strip before it; a scheme told at a glance by six of its colours; a scheme offered in the theme chooser, a small board in its own colours |
| `Progress`, `Spent`, `Scrollbar` | how much is done, how long a day took, which lines are in sight |
| `Ring`, `TodayMark`, `DayCell` | what the keyboard is on or a target, today, a day of a month |
| `Bar`, `Hint` | the command bar and one of its offers |
| `WorkspacePill`, `FilterPill`, `InfoBox` | the workspace, a filter or a view in force — a view changed on top marked by a late dot — a box of facts |
| `Empty`, `Dashed`, `Struck`, `Landing` | an empty slot, a dashed line, the line through what is done, where a drop lands |
| `Cross`, `Chevron`, `CalendarMark`, `Grip`, `FocusDot`, `CopyMark`, `MenuMark`, `Stone`, `StandingMark`, `ServesMark`, `LoopMark`, `SeriesNumber` | the marks; where an objective stands, what a task serves and what comes back are drawn, not written, since no face carries `▶ ‖ ◇ ↻`; a series' number is the loop and its figures |

**The `Kit` is what every drawing is handed**: Gio's context, the theme,
the skin, the layer it draws on, the clip it draws within, and the record
of what was drawn. The layers are planes, each over the one before:
`LayerBack` (the window's back), `LayerBoard`, `LayerBar`, `LayerDrag`,
`LayerPanel`, `LayerPopover`. A component records its box (`Kit.put`) in
the `Drawn` record, cut to the clip it is drawn within (`Kit.within`), so
a card scrolled under its column's edge is recorded as far as it is seen.
What is drawn on something is recorded inside it (`Kit.inside`): a stripe
in a card is `card/stripe`, a card on a column's ground `ground/card`, a
button on a panel `panel/button`, a panel on the veil `veil/panel`.

**Two guards keep it a system.** `TestNoDrawingNamesAColourOrARadiusOfItsOwn`
fails on a colour, a radius or a stroke's width written anywhere but
`skin.go`. `TestNoDrawingOutsideAComponentSpellsAShapeOut` reads every
function named `draw…` outside `component_*.go` and fails on a call to
`fill`, `rounded`, `outline`, `framed`, `lining`, `raised`, `disc`,
`ringed`, or anything of Gio's `clip` and `paint`; what clips, fades or
lifts a part is the `Kit`'s (`within`, `faded`, `shifted`), and so is the
window's measure (`measured`).

**Nothing is drawn over its neighbour.** `Drawn.Overlaps` is every pair of
parts on one layer, inside one parent, whose boxes cross. A part is never
compared with its parent, nor with anything on another layer — a popover
over a card is not a fault — and a part drawn over the whole of its parent
(a ring, a dashed line) is its frame, compared with nothing. Since a part
is never compared with its parent, a button laid past its panel's edge
crossed nothing and passed — the intro chooser's offers did, at 340 wide;
so `Drawn.Strays` is every part drawn on a panel that runs out of it, and
the settings' check fails on one too.
`TestTheFixturesDrawNoPartOverItsNeighbour` draws every screen on the
fixtures — the board, a chosen card, a task's form and its trace, an
objective's form, the help, a question, a menu, a calendar, a list, the
labels, the backups, the filters, the configuration, the command bar, the
switcher, the objectives, the week, the week before, a day, a month, its
months and its years, an objective holding sub-objectives, its tasks shown
and narrowed, its deletion asked, a sub-objective's form, the parents
offered and what a task serves refused, a view in force, one changed on
top, its popover, one being kept, the views, one gone and one renamed, a
sub-objective added under its parent, a part walked to, the views walked to
and a view narrowed to a part, then, recurrence on, the series, a stopped
one, one walked to, its form, its deletion asked and asked again, a new one,
the rules offered, a rule built by its parts, a rule read back, the board
marking them, an occurrence chosen and its form,
saving one asked, the help with the series, and the command bar answering a
series with recurrence off — the
fixtures holding three views, one of them gone, and two series — at the two sizes the layout is pinned at, 720 by
560 and 1200 by 800, and fails on any overlap. The settings are drawn the same
way, on a configuration of their own: a workspace's, scrolled, Global and its
places, a row walked to, a value refused, a field written, the theme chooser,
the board told of a change and the board and the panel with a file that
cannot be read (`TestTheSettingsDrawNoPartOverItsNeighbourAndEveryTextWhereItFits`). A card records where it
writes its number and its facts, so what it writes is judged too. What it
found when it was written, and what was done:

- on a narrow card — the week's at 720, a blocked card of the band — the
  facts ran over the number (`#10d≡`): they now give their last pieces up
  (what it serves, then the time, then the deadline) until they fit beside
  the number (`CardRowOf`,
  `@gui-a-narrow-week-card-keeps-its-number-and-facts-apart`);
- the form's grip lay over its last button by a few points: a form's
  buttons now end where the grip begins
  (`@gui-the-forms-grip-stays-clear-of-its-buttons`);
- the calendar and a list chosen from are dropped over the form: they are
  drawn on the popover layer, as what is dropped over the rest is; the
  command bar's panel and a question stand on their veil.

A card's title is broken between words only: a word the first line has
no room for — beside the glow of a card in focus — moves to the second,
and a word wider than any line is cut with an ellipsis. A card of the
week stands across its day as a card of the board across its column; it
was inset by the day's padding on top of its own, and at 720 broke its
title inside a word, `Wri` over `te th…`
(`@gui-a-title-never-breaks-inside-a-word`).

The text check found one more on the new screens: at 720 a day of the week
had no room for its total beside its date, and now writes it on the line
under it (`WeekTotalAt`,
`@gui-a-narrow-day-writes-its-total-under-its-date`).

**A new screen is assembled, not drawn.** Its sizes are worked out in a
`parts*.go` function a test pins; its drawing function takes the `Kit`, puts
itself on its layer (`k.on`), draws a container — a `Panel`, a `Popover`,
a `Ground` — and places components inside it (`k.inside("panel")`), each
at the box its parts gave. If it needs something no component draws, the
component is added first, in a `component_*.go` file, with its test, and
every screen that drew it by hand takes it.

**A theme sets tokens.** The window wears a colour scheme (#153): `Skin`'s
colours are a `theme.Colours`, and `Skinned` is `Dark()` wearing the colours
of a scheme, every other token — space, radius, size, type, shadow — as
`Dark()` has it. No component and no screen changes, since none of them
names a colour of its own. Which scheme a window wears, and where a person's
own is kept, is [config.md](config.md#the-colours-the-window-wears)'s;
the window is handed its skin when it opens (`gui.Run`), and from then on
the skin is the window's state (`Window.Skin`): the theme chooser re-skins it
while a scheme is previewed, and a configuration read again re-skins it when
its theme changed. A window on another workspace is another process, which
reads its own.

A scheme is a **base16 file as tinted-theming writes it** — `system`,
`name`, `author`, `variant`, and `palette` from `base00` to `base0F`; the
eight more slots of base24 are read past — so any of the several hundred
schemes shared in that format is dropped in unchanged. It may carry one key
more, which other tools pass over: `cairn`, with `tokens` (a token by its
name, below, to a colour) and `labels` (one of the fourteen label colours
by its name, to a colour). A colour is `#rrggbb`, or `#rrggbbaa` for a
token drawn see-through. `internal/theme` reads it (`theme.Parse`); a slot
left out, a colour that is not one, a name that is no token, or a file
that is not YAML is said, naming the file, the key and the value.

**The mapping.** `mix(a, b, t)` moves `a` a fraction `t` toward `b` in sRGB;
`bg` is `base00`, `fg` is `base05`, and `ramp(t)` is `mix(bg, fg, t)`. Every
neutral is a step up the ramp, so a light scheme works as a dark one does.
A token's name is its group and its field, lower-cased with `_` between
words (`ink.on_tint`, `surface.chip_on_control_lit`), then
`priority.<name>`, `standing.<state>.back|ink`, `stones.<0-2>.light|dark`,
and `label.<colour>` (set under `cairn.labels`, never `cairn.tokens`).

| token | from | why |
|---|---|---|
| `surface.back` | `bg` | "default background" in every scheme |
| `surface.ground` / `card` / `hover` / `field` | `ramp(.04)` / `ramp(.08)` / `ramp(.11)` / `ramp(.02)` | a column, a card, a step above it, an input sunk below it |
| `surface.raised` / `control` | `ramp(.16)` / `ramp(.19)` | a panel, a tab or a button |
| `surface.control_lit` | `ramp(.27)` | a button lit sits above it as `Dark()`'s does |
| `surface.chip_on_control` / `_lit` | `ramp(.06)` / `ramp(.14)` | a key's chip sinks below its control by the same step, lit or not |
| `surface.tint` / `ghost` | `mix(bg, base0D, .15)` / `.28` | the accent whispered, and a stronger whisper |
| `surface.veil` / `veil_light` | `mix(bg, black, .5)` at 158 / 79 | dark in both variants; the light veil is half the other |
| `line.rule` / `raised` / `edge` / `selected` | `ramp(.09)` / `ramp(.25)` / `ramp(.45)` / `ramp(.55)` | a column's rim, a rule on a panel, a control's edge, the keyboard's neutral ring |
| `ink.text` / `on_accent` | `fg` / `bg` | the ink, and the window's ground through a primary button |
| `ink.faint` / `low` | `ramp(.78)` / `ramp(.58)`, light `ramp(.86)` / `ramp(.70)` | secondary text, numbers and counts |
| `ink.on_tint` | `mix(fg, base0D, .40)` | text on the tint, pulled toward the accent |
| `accent.base` / `lit` | `base0D` / `mix(base0D, fg, .15)` | the scheme's main colour, and a step toward the ink |
| `accent.halo` / `glow` | `base0D` at 38 / 76 | a field written into, a card in focus |
| `accent.chip` / `lit_chip` | `mix(accent, bg, .15)`, lit or not | a chip on the accent sinks toward the ground |
| `state.late` / `done` | `base09` / `base0B` | orange — yellow falls under 3 on every light ground — and green |
| `state.wrong` | `mix(base08, fg, .5)` | red brought halfway to the ink, so an error reads as text and not as a critical stripe |
| `state.danger` / `danger_lit` | `mix(bg, base08, .18)` / `mix(danger, fg, .08)` | red whispered, and lit |
| `state.danger_chip` / `danger_lit_chip` | halfway from each back to `bg` | a chip on a dangerous button |
| `priority.critical` / `high` / `normal` / `low` | `base08` / `base0A` / `base0C` / `ramp(.5)` | stripes, which need no text contrast |
| `standing.active` / `on-hold` / `reached` | ink `base0B` / `base0A` / `base0D`, back `mix(bg, ink, .18)` | an objective's state |
| `standing.dropped` | back `control`, ink `faint` | a state with nothing left to do |
| `stones.0-2` | light `ramp(.30 / .38 / .46)`, dark `mix(light, bg, .4)` | the cairn mark, three neutrals |
| labels | `red` `base08`, `peach` `09`, `yellow` `0A`, `green` `0B`, `teal` `0C`, `blue` `0D`, `mauve` `0E`, `flamingo` `0F`; `maroon` `mix(08, 09, .4)`, `pink` `mix(08, 0E, .5)`, `rosewater` `mix(08, fg, .55)`, `sky` `mix(0C, 0D, .5)`, `sapphire` `mix(0C, 0D, .75)`, `lavender` `mix(mix(0D, 0E, .4), fg, .3)` | a label keeps its name in every scheme; on a light one its colour is pulled 45% toward the ink, else yellow on cream is unreadable |

The mapping is the mockup's (#153's, validated on 2026-11-09) with two
moves and what it did not draw. `state.wrong` was `base08`, the colour of a
critical stripe, which the skin's own test refuses; it is now red brought
halfway to the ink, as `Dark()`'s pale red is to its red. A label's pill is
a fifth of its colour over a card in every scheme — the mockup drew a derived
scheme's at an eighth — and the contrast checks below decide whether it
reads. `control_lit`, the chips, `line.raised`, `accent.lit`, the halo, the
glow, the danger's lit and chip shades and the stones were not in the
mockup: each is worked out from its neighbour the way `Dark()` relates
them, as the table says.

**The nine shipped**, carried in the binary (`internal/theme/schemes`):
`default`, `solarized-dark`, `solarized-light`, `gruvbox-dark`,
`gruvbox-light` (medium contrast), `catppuccin-mocha`, `catppuccin-latte`,
`dracula`, `nord`, each palette read from its makers' own source, named at
the head of its file. **`default` is `Dark()` pinned token by token**: its
sixteen slots let other tools use it, and every token and label is written
under `cairn`, so not one pixel moves; `TestTheDefaultSchemeIsDarkTokenByToken`
holds the two equal. Catppuccin's two give the fourteen labels their own
colours.

**Every shipped scheme meets every contrast** (`theme.Shortfalls`, a test
over the nine): text 4.5 — `ink.text` on the window, a column and a panel;
`ink.faint` on those and on a card; `ink.on_tint` on the tint;
`accent.base` and `state.wrong` and `state.late` on the window, `late` on a
panel — then 3 for `ink.low`, `accent.base` and `state.late` on a card, an
edge on the window and a label's name on its pill, 2 for an edge on a
panel, and the surfaces as the skin held them: a panel 1.3 against the
window, a column 1.08, a panel 1.15 against a column, a column's rim 1.2,
a rule on a panel 1.2. The skin's test adds that no priority is the
accent's colour, or the edge's, or another's, and an error is not a
critical task's. **A scheme that falls short gets overrides, never a lower
bar**: its file carries the tokens and labels moved under `cairn`, with a
line saying why — a colour too close to its ground moved the least it takes
toward black or white, a priority or an error too like its neighbour moved
along the ramp or toward the next slot. All eight besides the default
needed some, Solarized the most — its monotones sit close together — and
Latte moves eight of its own labels darker. A person's own scheme that
falls short is worn all the same, and `config show` names each pair and
its ratio.

**The default moved once, on purpose.** `Ink.Low` was `#5f6067`, 2.60
against a card where numbers and counts are held to 3; it is `#696a71`,
3.02, ten points lighter on each channel, the least that passes. Every
number, count and key hint on a card is that much brighter.

## The window

The terminal's shape, larger. Four columns — backlog, todo, current, done —
the backlog full height, the stalled band across the last three. Three tabs:
tasks, objectives, and the time spent — and a fourth, the series, while
recurrence is on ([series](#series)).

**Each column and the band stand on a ground of their own.** A region reads as
a group when it has a ground or an edge of its own (common region, Nielsen
Norman Group), and that beats proximity alone. In a dark theme a lighter tonal
surface does what a shadow does elsewhere, so every column, its header
included, and every lane of the band stand on a ground of their own, lighter
than the window's and rounded (10), with no edge: the tone says it, and the
cards stay the lightest layer. The gap between columns stays: it is the
gutter. There are no vertical rules and no alternating tints, which would
imply a meaning nobody gave them. The ground is a skin token, checked for
contrast with the others (`@gui-columns-stand-apart`).
Each head names its column in words, 'Backlog' rather than 'BACKLOG', at its
left, ten pixels in and seven under its top, and counts its cards at its
right in figures (`ColumnHeads`, `@gui-a-columns-head-names-and-counts-it`);
a centred name, which #142 set, gave the count no place of its own. The
band's lanes follow the same rule through the same function (`NameIn`,
`@gui-a-lanes-name-sits-as-a-columns-does`). A column or a lane holding
nothing says 'Nothing here' in a dashed box where its first card would be,
and a board holding no task says how to make one, in todo
(`EmptySays`, `@gui-an-empty-column-says-so`,
`@gui-an-empty-board-says-how-to-start`). A card in done, or archived, is
drawn at half strength: its work is behind it (`Dimmed`, `@gui-done-is-dimmed`).

**A column, or a lane, is widened by dragging the gutter beside it**. The gutter is the gap between two columns, from the top of the
head to the foot of the column on its left — right of the backlog it runs
beside the band too, since the band begins where the todo does — and the gap
between two lanes of the band. Every gutter, and the edge above the band, is
8 points: the gap of a board such as Trello's or GitHub's, enough to tell two
columns apart and to be taken by the pointer (`@gui-every-gutter-is-as-narrow`).
Dragged, it gives the one on its left what it
takes from the one on its right, and nothing else moves
(`@gui-a-column-widens-when-its-gutter-is-dragged`,
`@gui-a-lane-widens-when-its-gutter-is-dragged`). Over a gutter the pointer
takes the shape for widening sideways, and keeps it wherever it goes while the
gutter is dragged (`Window.PointerAt`, `@gui-a-gutter-shows-it-can-be-dragged`,
`@gui-a-lane-gutter-shows-it-can-be-dragged`). Neither goes under
`NarrowestColumn`, room for a card's title
(`@gui-a-column-dragged-past-its-neighbour-keeps-a-card-in-sight`). What is
kept is each one's share, not its width: a window made smaller makes every
column smaller, nothing holds it back, and a column no longer readable is the
person's to widen again (`@gui-columns-keep-their-shares-when-the-window-shrinks`).
The widths are the window's, not the store's: a window opened again has its
columns as wide as each other (`@gui-columns-open-as-wide-as-each-other`),
unless the configuration says to keep them (below).

**The band is made taller by dragging the edge above it.** The edge is the gap
between todo, current and done and the band, as wide as the band. Dragged, it
gives the band what it takes from the three columns above it, and the backlog,
which runs the window's height, stays as it was
(`@gui-the-band-grows-when-its-edge-is-dragged`). The band never goes under its
lanes' names and a card's room, nor the columns under a card's room
(`LowestBand`, `@gui-the-band-dragged-down-keeps-a-card-in-sight`,
`@gui-the-columns-dragged-up-keep-a-card-in-sight`). What is kept is the
band's share of the height, the way a column keeps its share of the width
(`@gui-the-band-keeps-its-share-when-the-window-shrinks`). Its height had been
fixed, two cards, which is right for a person who keeps little stopped and
wrong for one who keeps a great deal. Over the edge the pointer takes the shape
for resizing up and down, the one the drag of the edge keeps
(`@gui-the-bands-edge-shows-it-can-be-dragged`).

**A size given is taken back by a double-click, or all at once by the menu.**
A double-click on a gutter between columns makes every column as wide as the
others, on one between lanes every lane, and on the edge above the band gives
the band the height it opens with — what that gutter shares and nothing else
(`@gui-a-double-click-on-a-gutter-evens-the-columns`,
`@gui-a-double-click-on-a-lane-gutter-evens-the-lanes`,
`@gui-a-double-click-on-the-band-edge-gives-it-back-its-height`). `Reset
sizes`, in the View menu, takes back every one of them
(`@gui-reset-sizes-takes-every-size-back`). A size dragged by hand is not
easily dragged back exactly, and a way back that has to be found again by
hand is no way back.

**The sizes are kept from one window to the next when the configuration says
so** — `[window] keep_sizes = true`, see [config.md](config.md). They are then
written beside the store, `cairn.db.sizes.json` (`Places.Sizes`), every time a
gutter is let go of or a size is taken back, and read when the window opens
(`@gui-the-sizes-are-kept-when-the-configuration-says-so`). A reset is kept
like any other size (`@gui-sizes-reset-are-not-kept`). Off by default, because
some want the board as it starts each time and some the board as they left
it, and the first was the behaviour already there
(`@gui-the-band-opens-as-tall-as-ever`). Not in the store: the store may be
shared, and how wide one person likes a column is nobody else's. Not in the
configuration file either, which holds choices, not what a hand last dragged. On
the tasks page only.

**A form is moved by dragging it where it holds no text**. A press
inside the form on nothing written — no field, no heading, no offer, no text
of its trace — takes the form by its ground, and the pointer carries it
(`@gui-the-form-moves-when-dragged-by-its-ground`); a press on a text still
holds the text (`@gui-the-form-stays-when-its-text-is-dragged`). The form is
held wholly inside the window (`@gui-the-form-stays-inside-the-window`). A form
keeps its place until it is closed; the next one opens where a form opens,
whatever the last was moved to (`@gui-a-form-reopens-where-it-belongs`),
and the place is written nowhere. A press that
does not move past the grip is a click, as it was. A text is where its words are,
not the line it sits on: past the heading's words, or past a field's name on
the line that names it, the form is taken (`@gui-the-form-moves-when-dragged-beside-its-heading`,
`@gui-the-form-moves-when-dragged-beside-a-fields-name`); a field's input, and
the trace beside, keep their whole box. Everything a form draws or
reaches is placed from its panel, so moving the panel moves it all.

**Cards keep a fixed height.** In the terminal that is what makes moving the
cursor arithmetic rather than measurement; with a mouse that reason is gone,
and it was kept anyway for what it still buys: a column can say how many cards
fit, where the n-th one is, and which card is under the pointer, without
drawing anything. The price is the same as the terminal's — a long title is
cut.

A cut title is still readable: while the pointer rests on the title of a card
whose title was cut, a small bubble drawn above everything else says it whole, wrapped
over as many lines as it needs and kept inside the window. A card whose title
fits gets none. The cards keep their height, and the title is not lost to it.
It stands just above the title, its text beginning where the title's does, so
the eye goes from the cut words to the whole ones without a jump; placed above
the card, it stood a line of number away from what it completes.
Where there is no room above, it goes just under the title. A form's field
whose value is cut opens the same bubble over its value, and a field's name
opens the bubble that explains it, drawn the same way: they were a box with
an edge, and the explanation wrapped at a fixed width, so a sentence could
end on a word alone on its last line; every bubble now wraps at the
narrowest width that keeps its number of lines (`evenly`, `cutValueAt`,
`TestAFieldsValueCutShort`, `TestAFieldsExplanationIsABubbleWrappedEvenly`).
Scenarios:
`@gui-a-cut-title-is-read-whole-under-the-pointer`,
`@gui-a-title-that-fits-has-no-bubble`,
`@gui-the-bubble-stands-just-above-the-title`. The labels a card leaves out
are read the same way: resting on its `+n` opens a bubble under it holding
every label it carries as pills, wrapped to 180 points
(`LabelsBubbleBox`, `@gui-the-count-of-labels-opens-them-all`). Both bubbles
are popovers, rounded and shadowed, rather than boxes with an edge. The
title's bubble used to open anywhere on the card, and hid the row a person
was reaching for; it opens over the title alone now. The objective mark `◇1`
opens the same bubble, a line to each objective it serves, `#1  Ship the
desktop board`: a number alone named nothing (`servedAt`,
`@gui-the-objective-mark-names-what-it-serves`). Go Mono has no glyph for
`◇`, so the facts are measured with the face that stands in for it
(`mixedWide`); measured as Go Mono, they were cut to `1…`. The
row of a card is laid out once (`CardRowOf`) for the drawing and for the
pointer, so what is drawn and what answers it cannot drift.

Variable height was the alternative and would have cut nothing, at the cost of
a column's geometry depending on its contents: scrolling and drop targets
computed instead of multiplied, and two columns no longer sharing a grid, which
shows the moment a card is dragged from one to the other.

**A card's priority is a stripe five points wide** down its left side
(`@gui-a-card-carries-its-priority-in-a-broad-stripe`). Kanban boards keep an
accent stripe at three or four pixels and anything under reads as a border; five
is clearly an accent at this window's scale, and still well inside the padding
before the text. The card's outline is drawn over the stripe's first column.

**A card is its title, then a row, then its steps.** The title comes first,
in the reading ink and a medium weight, on two lines at most, and is cut with
an ellipsis at the end of the second (`TitleLines`,
`@gui-a-long-title-is-cut`); the whole of it is a bubble away. Under it, one
row: the number in figures, faint (`#152`), the description's mark after it,
then the labels as pills, two at most and then `+n` in an outlined pill for
the rest (`Pills`), and the facts at the row's right end. A label is shown
only when the count of those left out still fits beside it, so what is left
out is always counted rather than dropped in silence. At the foot, only when
the task has steps, a thin bar filling with the steps done and `k/n` beside it;
a task with no step has no bar and no empty line. The card stands on a rounded
ground (8), striped at its left by its priority (5 points, its right corners
rounded), lit under the pointer, ringed in a neutral grey when the keyboard is
on it, and tinted when it is in focus.

The right-hand end of that third line carries **what is due, what the work has
cost against what it was reckoned to cost, and what it serves** — `2d  9h/2h
◇4` — in the faint ink, or in `State.Late` when the day has passed or the
estimate is spent. The window used to know nothing at all about time: no
deadline, no notion of late, where the terminal has written it on the card all
along, and deciding what to do without knowing what is due is not deciding. A
day more than a week out is left off, for the terminal's reason: it will still
be there tomorrow and in ninety days, and saying so on every card every day
buys nothing. So is the day of a task that is done, which is wanted on no day
any more (`@gui-a-finished-card-says-nothing-of-its-day`). A blocked card waiting on nothing any more says that instead, and
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

**720 by 560**, which is four columns of 168dp, a card's room above the band
and the band's under it, and a task's whole form: under 640 high the panel a
form is drawn in takes the room the window has rather than five sixths of it
(`Over`, `overLeast`), so the form's last field, its steps, still ends inside
it (`TestEveryFieldOfAFormFitsAtTheLeastSize`). Written in rather than worked
out, because what makes a card readable is not something a sum knows — and
checked by drawing every screen at that size. The window cannot be dragged
under it (`app.MinSize`); the message saying it is too small stays for a
system that ignores the hint. It was 1024 by 640, four columns of 285dp,
before the redesign made a card's row of pills and facts fit in less.

**The beat stays at two seconds**: `Version()` is one integer, and the board
re-reads only when it moves. Waiting for the window to take focus was turned
down for a reason worth keeping: a board on a second screen, never focused, has
to be true. The one exception is a task in focus: its time spent grows
with no write to the store, so while one is, the window reads the store again
once a minute (`FocusShown`), the least the time is written in. It used to
show the time read when the board was last loaded and never move. The
terminal board does the same on its own beat. The same beat reads the configuration again
when one of its files changed ([settings changed in place](#settings-changed-in-place)).

## The mouse

A click chooses a card, or changes tab when it lands on one. A double-click
opens it to be edited. The space between two cards belongs to neither: a click
there chooses nothing (`@gui-a-click-between-two-cards-chooses-neither`), where
it used to choose the card above, which left no way to point between them. The wheel scrolls the column, or the lane, under the
pointer.

**Only the primary button presses and drags.** A right click, a middle click or
a side button used to run the same press and release as the left one, so they
chose cards and pressed offers. A right press now opens the menu of what is
under it (see *A right click says what can be done to what is under it*); the
others do nothing.

**What the pointer is resting on is lit.** The window answered nothing before
it was pressed, which made "Delete" read exactly like a label: the one
affordance that says a word can be clicked was missing from the whole surface.
Gio is asked for `Move`, `Enter` and `Leave`, the window keeps which offer the
pointer is over, and the drawing decides what that looks like — a value, like
everything else here, so a scenario can ask.

**Every offer is a button, and one kind of button.** The bar's offers, a
panel's `Save` and `Cancel`, the calendar's month, year and arrows and the time
strip's chips were words in one place, words lit on hover in another, chips in
a third. They are now drawn by one function (`Button`, `ButtonIn`): a box as
tall as the room given, a gap short of the next, edged, its words in its
middle. Under the pointer its edge takes the chosen ink; the one in force (a
range, an open grid) is filled; the board's offers behind an open panel are
written faint (`@gui-every-offer-is-a-button`,
`@gui-a-lit-offer-holds-its-words-in-its-middle`).

**A turn of the wheel belongs to what it began on.** A trackpad goes on
sending after the fingers leave it, and a task put away with `esc` while it was
still sending handed the rest of the turn to the board: the column under the
pointer scrolled though nobody had turned anything over it. A part of a turn
that comes within 150 ms of the last one goes on the same gesture, and is
dropped if what the gesture began on has gone; a part after a pause is a turn
of its own. Gio reports no phases for a gesture, so the pause is what says one
has ended.

**A panel keeps the wheel.** Over the calendar, the right-click menu or the
list of backups, the wheel scrolled the column hidden behind them. Nothing
under a panel moves now; the list of backups scrolls itself, a line a notch,
and the calendar walks its months, a notch a month
(`@gui-the-wheel-stays-inside-a-panel`).

**Every bar can be pulled.** The trace's, the description's, the steps of a
task, and the list's: all four have a thumb sized never to be too small to
grab, and two of them used to ignore a drag entirely.

**The board is filtered by what a task carries**, and what filters it stands as
pills at the right of the tabs. Each filter in force is a pill on the accent's
tint, its name faint and its value bright, `priority high`, its
cross taking it away; a search kept comes first, `/ spec`; then `Clear`,
which takes every filter and the search away; and last, against the menu,
`+ Filter`, a rounded pill drawn dashed (`Pills`,
`@gui-the-filters-stand-as-pills-at-the-right-of-the-tabs`,
`@gui-a-kept-search-stands-first-among-the-pills`). The row ends against the
menu, where the search's pill and the filters' text always stood: started
right after the last tab, as it was for a day, it sat where nothing had been,
and away from where the eye looked for it. A page added takes room from its
left; when the pills no longer fit between the last tab and the menu, they
fold into one, `Filters · 3`, standing where `+ Filter` stood and opening the
menu (`@gui-the-pills-fold-when-the-room-runs-out`).

**The pill that adds a filter never moves**
(`@gui-the-pill-adding-a-filter-stays-where-it-stands`): the row grows to its
left. With it last and how many tasks were kept, `2 of 3 tasks`, at the end of
the row, every filter added or taken off widened the count and pushed the pill
from under the pointer that was going back to it. That count is the tab's now:
a page narrowed by the search or the filters counts what it shows of how many,
`2/3`, where it counted what it showed (`TabCount`), Time aside, whose count is
the tasks worked in the period.
The filters had a panel of their own across the middle of the window, each
value walked with `‹ any ›`, and what was in force was a line of text beside
the tabs; Linear and GitHub show filters in force as pills.

`⇧F`, `+ Filter`, or the menu bar's Filter drops the menu of filters under the
pill, `Filter by` at its head: column, priority, label, objective, and a day
the work is due by — today, in a week, in a month — each with what it is on
after its name and, at its right, the letter that opens it. It is a menu, not
a panel: no heading of a panel, no cross, no buttons, its keys said in a line
at its foot (`FilterFoot`), as the mockup drew it. A line is chosen by the
arrows, and `⏎`, a click or its letter opens its values
(`@gui-a-filter-opens-by-its-letter`), a filter's pill opening them directly
(`FilterList`, `@gui-a-filter-pill-opens-its-values`,
`@gui-a-filter-opens-the-list-of-its-values`). The values are headed by the
filter's name and how they are taken, a label by its colour's dot, and each
says how many tasks it would keep under the other filters as they stand
(`FilterCounts`, counted by the engine); `any` is not one of them, the filter
taken off by its pill's cross or by choosing again the value it is on
(`@gui-a-value-chosen-again-takes-its-filter-off`). **Nothing narrows until a
value is chosen**: the arrows only move, over the filters as over their values
(`@gui-moving-over-the-filters-narrows-nothing`). `←` `→` walked a filter's
value in place, the board changing under the menu at every press, which is
what the panel did and the mockup did not. A value is taken by `⏎` or a click;
labels are ticked, as many as wanted, a task kept when it carries every one
(`@gui-the-labels-are-filtered-together`), the others take one. `esc` goes
back where the values were opened from — the menu, or the board for a pill
(`@gui-a-filters-list-is-left-untouched`). On a line of the menu, `⌫` takes
that filter off and that one alone (`@gui-a-filter-is-taken-off-by-its-key`);
`⌥⌫`, or No filter, puts every one back on `any`
(`@gui-every-filter-is-cleared-at-once`). On the board, with no menu open, `⌫`
takes the last pill off — the last filter in force, then the search — the way
a field of tokens loses its last one (`@gui-backspace-takes-the-last-pill-off-the-board`),
and `⌥⌫` takes them all (`@gui-alt-backspace-clears-every-filter-from-the-board`).
`⇧⌫` was the first thought and could not be: a backspace typed with shift held
is a backspace, so typing in capitals still deletes (`Typed`). The board narrows as
they are chosen (`@gui-a-filter-narrows-the-board`), two together keep what both
reach (`@gui-filters-narrow-together`), and the search narrows what they kept
(`@gui-a-filter-and-a-search-narrow-together`). The filters are the engine's
(`engine.Filters`): what each keeps is what `task list` keeps with the same flag.

**The same filters narrow every page.** One set, shared by the pages, so the
pills in force do not change with the tab: the question asked — the work on
`bug` — is about the subject, not the page. The objectives' page keeps the
objectives by what an objective is — its labels, its priority, which one it
is — the rule `objective list` follows, and counts `n of m objectives`; a
filter that means nothing to an objective, a column or a day due by, stays as
a pill drawn faint (`Pill.Faint`, `@gui-the-filters-narrow-the-objectives`).
What an objective's tasks carry is a view of its own, cairn#226. The time
page keeps the time of the tasks kept, its totals with them, as a search
already did (`TimeShown`, `@gui-the-filters-narrow-the-time`).

**A search does not take the window.** It narrows the board in place, so the
board keeps answering the mouse: a card is chosen by a click and opened by a
double click while the box is still open and still being typed into. It used to
swallow both, and reaching what a search had just found meant closing the box
that found it.

**A word found only in a step raises the task it belongs to**
(`@gui-searching-raises-the-task-a-step-belongs-to`): a step's title is part
of what its task says, and a step has no card of its own to raise.

A click over the box itself belongs to what is being typed, and lets go of
nothing. That is the whole of the exception — everywhere else on the board a
click does what it always did.

The guard that swallowed them was right about panels and wrong about this:
a form, a question or the help is drawn *over* the board, and a click reaching
through one is how an open task came to show something nobody asked it for.
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

**A letter made of two keys is the system's to make, and it has to be told
the text it is making it in.** On macOS a dead key — `^` then `e` for `ê` —
first writes the accent as a composition, then sends the letter with the run
the accent took, to be put in its place. Gio keeps its own idea of the text
and the caret for that, filled from what the window tells it each frame. The
window used to tell it nothing, and appended every edit at the caret:

- after the accent, Gio's idea of the text no longer matched what it was told,
  so it dropped the composition;
- the `e` then arrived on its own, and the field read `^e`.

`heard` now ends every frame by telling Gio the field's text and caret, after
the frame's events and not before. Told before, it lags one frame behind and
the composition is dropped all the same. An edit replaces the run it names.
Linux and Windows compose before the text reaches the window, which is why the
fault was macOS's alone. The scenario plays what macOS does, dropping the
composition exactly when the window would have made it.

A double-click is worked out here rather than asked of the system: Gio reports
presses and says nothing about pairs of them. Two presses within 400ms and
within the same few pixels are one gesture, and a third one more is a triple
click — unless the double opened something, in which case the third press
lands on what it opened and is a press there.

**A click that moves becomes a drag.** Past a few pixels the card lifts and
follows the pointer; below that it is a click. Without the threshold, a shaky
click would move a task.

**The board makes room where the card would land.** While a card is dragged
it leaves its place, and the column under the pointer opens a card's room
where it would go, the cards after it moved down and the room outlined — the
way a list a card is dragged through behaves in Trello and the like
(`TestADragOpensRoomWhereTheCardWouldLand`). Which room is the half of a card
the pointer is over: the upper half of a card puts it before, the lower half
after. A thin line across the top of the slot under the pointer said it before,
and since the slot was a card's whole height, the last card of a column
scrolled to its end covered its foot: there was no point anywhere that meant
"after it", and a card could not be put last
(`@gui-a-drag-reaches-the-foot-of-a-long-column`). The card in hand is never
counted where it is going, so it is never put behind itself. That holds until the
drop itself: the place is worked out while the card is still in hand, and a
card let go over the lower half of its own slot used to be sent behind itself,
and refused (`TestACardDroppedOverItsOwnLowerHalfStaysBehindTheOneBefore`).

**A drop carries the column and the place.** That is one `task move --state
todo --after 3`: the column and the place in one act, so a place refused
leaves the column as it was. It used to be an edit and then a move, with
nothing between them, and a refused move left the task in a column it was
never dropped in.

**Every card on the board is a task, and any card may take any place.** A
step is no card (`@gui-a-step-is-not-on-the-board`): it has no column and no
place among the tasks, so nothing on the board refuses a landing or marks a
card as belonging to another. There was such a rule while sub-tasks existed —
a sub-task could only land among its sisters, which were scattered through the
columns with nothing grouping them — and it read as arbitrary. That is part of
why sub-tasks gave way to steps; see [steps.md](steps.md).

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

**A lane takes a place the way a column does.** A card dragged over a lane
opens a card's room where it would land, and the drop carries the lane and the
place in one `task move` (`@gui-a-drag-in-a-lane-carries-the-place`,
`TestADragOpensRoomInALaneWhereTheCardWouldLand`). A drop used to carry the
lane alone: the card went last, and one dropped in its own lane did not move,
so the order of what was stopped could not be set from the window at all
(`@gui-a-drag-within-a-lane-only-moves`).

## The form

**A panel's heading stands out from what it holds.** Every panel, the form
included, writes its heading in bold, at 18 where a field is written at 16 or
less (`HeadingSize`, `HeadingWeight`), and the drawing and the measure of the
heading's words read those two alone
(`@gui-a-forms-heading-stands-out`).

**A step is reached in the form of the task it belongs to, and nowhere
else.** The form carries a `steps` field, which is where one is ticked, renamed,
added, taken away and put in its order. Saving the form writes the steps with
the rest of the task, in the one act that writes the task: the field says what
the steps are, and what it no longer says is gone. There is no deletion of a
step apart from that act.

A step carries a title and whether it is done, and nothing else; its place in
the task is the only name it has, and no task's number, so its tile shows its
title alone. Measured on a store of 219 tasks, the 49 sub-tasks that came
before steps were done 47 times and todo twice, with no deadline, no label, no
priority and no blocker among them — which is what a step keeps.
**Deleting a task takes its steps with it** (`@task-delete-takes-its-steps`).

A modal dialog over the board, with *Save* and *Cancel*; `esc` cancels. One act
in the trace per save.
The form's heading ends in a cross, at the top right, where a dialog's close
sits on Windows and in Material (a Mac sheet has none); it is a 24-point target
and does what *Cancel* does, asking first when something was typed.

**The fields are not listed here.** They come from `engine.TaskFields(card)`
and `engine.GoalFields(goal)`: the word a person reads, the closed set when
there is one, the field that wants more than one line, and `Arg` — what the
engine calls the same thing. An attribute added to the engine and to the
terminal's form arrives here with no line written.

**A field's name says what the field is.** The pointer resting on it opens an
explanation under it, and leaving it takes the explanation away
(`@gui-a-field-says-what-it-is`). The words are the engine's — what the command
line's help and an agent's schema say of the same argument (`Field.Means`) — so
the three cannot disagree. Only the name explains: over what is being typed, an
explanation would be in the way. The title has no name drawn, and none.

**An explanation is the size of what it says, and shows what to type.**
Amended 2026-10-16. It was a box of three lines and 360 points for
words that mostly fill one — "the day it is due" — so it hid the field under it
for nothing, and it said what a field is without saying how it is written. It
is now as wide as its widest line and as high as its lines, and a field typed
into ends it on an example, in the faint ink: `e.g. 2026-12-25`, `e.g. 1h30m`,
`e.g. +30m` (`@gui-a-field-shows-what-to-type`,
`@gui-an-explanation-is-the-size-of-what-it-says`). The examples are the
engine's, beside the fields of the form (`Field.Example`), so a form drawn
anywhere else gives the same ones; a field chosen from a list gives none, the
list being its own example.

A field with `Choices` is chosen from the same list as the objectives, a `Long` field
a multi-line box, the rest a line to type in. The field the keyboard is on is lit, and it is the only one:
that is the whole of what says where typing would go.

**The objectives a task serves are picked from a list, not typed.** `enter` on
that field opens the objectives over the form — over, and not instead of: a
choice made with the rest of the form out of sight is a choice made from
memory. The arrows walk it and so do `j`/`k`; **enter** takes the row the
keyboard is on and closes the list, **⌘⏎** takes it and keeps the list open for
the next; `esc` closes it. Nothing reaches the store until the
form is saved, the way nothing else in a form does.

**The list is a tree** (#226): each objective of level one, its
sub-objectives under it, joined to it and moved right as in the objectives
tab (`Snapshot.Picks`, `@gui-objective-the-task-form-draws-the-tree`). A
parent of what is ticked is drawn faint, saying `counted through 2 below` at
its right. Ticking it anyway is refused where it is ticked, in a `Banner` at
the list's foot — `#44 is already counted in #4 Home automation through #5
Enclosure and #6 Firmware` — and nothing changes; saving would be refused the
same way, the rule being the engine's (`Snapshot.CountedOnce`,
`@gui-objective-the-task-form-refuses-a-double-count`). Its second line says
what to untick: `Untick Enclosure and Firmware to put it on Home automation
itself. Nothing was changed.`, or, a sub-objective ticked under a parent
already ticked, `Untick Home automation to put it on Enclosure.`
(`ChoosingThen`, `@gui-objective-the-refusal-says-the-parent-to-untick`). Two sub-objectives of
one parent, or two trees, are ticked as any two objectives.

`enter` never saves a form: it opens the list a field answers with, and it
breaks a line in a field with room for more than one. **⌘⏎** (or ⌘S) saves,
wherever the keyboard is. So the bar offers *Save* on every field, and adds
*Choose* beside it on a field a list answers — it used to give way to *Choose*
there, and with the keyboard left on that field after the list closed there
was no way to save at all. A key does what the bar offers; the bar says both.

**The field carries numbers and is drawn as names**, an objective's parent
and what it waits on as well. `1 2` is what the engine
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
`BLOCKED`, `LINKS`, by what each field asks the engine for; then the steps.
Under a rule at the foot, and nowhere else, is what the form offers. The panel
keeps one size whatever the task holds, the panel every other screen uses, and
its description keeps its share: cut to what it held, it grew and shrank from
one task to the next and moved the description with it
(`TestTheFormKeepsItsSizeWhateverItHolds`). Its corner, at the bottom right,
resizes it when dragged, its top left staying where it is, down to
`FormLeast` and up to the window; the size given is kept with the board's
other sizes and the next form opens at it
(`FormGrip`, `Sizes.Form`, `TestTheFormIsResizedByItsCornerAndKeepsTheSize`). `FormPlace`
says where every field of a form goes, and everything that draws, clicks or
measures a form asks it.

It was one column, and the steps — a second long field — ran over *Save* and
*Cancel* in a window under eight hundred points. Then two columns, which left a
hole under eleven fields and two frames as tall as the panel with three lines
in them. When the window is short, the steps are given their room before the
description grows past four lines: a step out of sight is a step nobody ticks.

**The steps read as a list to tick** — a tile each, ticked or not, from what
the field says now, unsaved lines included; how the tiles are worked is said
further down, under the last thing a form holds.

**A new task goes in with its steps.** The form that adds a task carries the
same steps field as the one that edits it, empty: a task whose steps were
written down beforehand is added with them, in the act that adds it
(`@gui-a-new-task-goes-in-with-its-steps`). The terminal's form for a new task
does not carry it yet.

**A new task goes where the pointer last pointed.** A click on a column — on a
card or on nothing — says which column the next task goes in: New, `⌘N`, or New
from a right click adds it there, unless a card is chosen, whose column wins
(`@gui-a-task-is-added-to-the-column-clicked`,
`@gui-a-right-click-on-a-column-adds-there`). A double click where a column
holds no card adds one to it at once, the way a calendar adds an event where it
is double-clicked (`@gui-a-double-click-on-a-column-adds-there`). It went to
the backlog whatever was clicked, so a task wanted in `todo` had to be added
and then moved. A double click in the space between two cards adds it there,
behind the one above, in the act that makes it
(`@gui-a-double-click-between-two-cards-adds-there`); a task moved to another
column in the form before it is saved goes where that column puts it.

**Every box in a row ends where the others do.** The title, the description,
the last field of each section and the steps end on the same line, and a long
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

**On Windows the window opens with no console behind it.** cairn is one
binary, a console program, which the command line needs; started from the
Explorer, Windows gave the window a console of its own, empty, beside it. The
window lets go of a console nobody else is attached to as it starts
(`leaveConsole`), so that one goes, after a flash; a console shared with the
shell `cairn gui` was typed in stays, and what the window says there — where
debug mode wrote down, among it — still reaches whoever typed it.

**Every character the window writes has a face cairn carries.** Go's own
faces lack the keys of a Mac keyboard, a tick and a cross; a character no face
has sends the shaper to the system's fonts, and every process that drew one
read them off the disk and kept them — the window's tests reached 3.5 GB each,
and eight at once in a mutation sweep restarted the machine. Those dozen
symbols are cut from DejaVu Sans into a font of 12 KB carried in the binary
(`symbols.go`, its licence beside it and in the notices), and a test reads
every string the window writes for a character none of its faces has
(`TestEveryCharacterTheWindowWritesHasAFace`): measured, a run of the
window's tests now peaks at 0.6 GB. What a person types still reaches the
system's fonts, an emoji in a title among it. The button standing in for a
menu bar is three strokes for the same reason.

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

**The objectives are a tree beside the one chosen** (#226). A grid could not
show what sits under what, so the tab is split (`SplitOf`): a list on the
left, no wider than 420 and two fifths of the board, and the detail of the
objective chosen on the right — of the first while none is. The list draws an
item of two lines for each objective (`ListItem`): its standing's mark, its
number, its title and its first label on the first, its progress over its
subtree on the second; a sub-objective stands under its parent, moved right
by 24 and joined to it by a line (`Connector`,
`@gui-objective-sub-objectives-sit-under-their-parent`). An item has no
ground at rest, the hover's under the pointer, a card's and a ring when
chosen. The list scrolls on a number of its own, and the arrows walk it,
bringing what they reach into sight. The hover is the pointer's on an item, a
tile and an offer of the detail alike
(`@gui-objective-the-tree-is-lit-under-the-pointer`).

**A parent starts folded** (#272, the mockup validated on 2026-11-10). Drawn
unfolded, seventeen objectives of three trees took the whole list and the
level-one ones were lost among their parts. A parent shows a chevron in a
gutter of 18 left of its item, outside its ground and its ring, and the
number of its sub-objectives after its title, `· 3`, folded or not — the
chevron says which; one with none keeps the gutter empty, so every title
lines up (`FoldAt`, `@gui-objective-a-parent-starts-folded`). Unfolded, the
line joining a sub-objective hangs from the chevron's middle and elbows into
it. Its second line is the roll-up of its whole subtree either way. A click
on the chevron folds or unfolds and chooses nothing; a click on the title
chooses and leaves the fold as it was
(`@gui-objective-the-chevron-folds-and-unfolds`,
`@gui-objective-the-title-chooses-without-unfolding`). The keys are a macOS
outline's: `→` unfolds, and on a parent already open goes to its first
sub-objective; `←` folds, and on a sub-objective climbs to its parent; `z`
toggles as the terminal's does, and on a sub-objective folds its parent and
climbs to it; `⌥→` and `⌥←` do every parent; `↑` and `↓` walk the rows
drawn (`@gui-objective-the-arrows-unfold-and-fold`,
`@gui-objective-z-and-option-arrows-fold`,
`@gui-objective-up-and-down-walk-the-rows-drawn`). The folds are the
window's, for the session: a second window has its own, and nothing is
written to the store or the settings — a key for a view state can come the
day somebody asks (`@gui-objective-folds-are-kept-for-the-window`). A
parent holding what is chosen is drawn open, whatever chose it.

**The order is moved by hand** (#277). `⇧↑` and `⇧↓` move the chosen
objective among its siblings, a parent with its sub-objectives, and stop at
either end (`@gui-objective-shift-arrows-reorder`). An objective dragged is
let go over a row in its top quarter (before it), its middle half (under it,
last) or its bottom quarter (after it); a parent goes by the level-one row,
and nothing goes under a sub-objective or a parent under another — the drop
moves nothing then (`engine.Dropped`,
`@gui-objective-dragged-reorders-and-changes-parent`,
`@gui-objective-a-parent-dragged-onto-another-is-refused`). A refusal from
the store, a task counted twice, is said above the bar. A folded parent
held under the pointer for 600 ms unfolds, and folds back when the drop
lands elsewhere (`HoverOpens`, `@gui-objective-a-folded-parent-opens-under-a-drag`);
under a parent's last sub-objective, the pointer left of the sub-objectives
lands at level one after the parent
(`@gui-objective-x-chooses-the-level-under-a-last-sub-objective`); a tag
beside the landing says where it goes — `under #1`, `after #6`, `first`.

**A search unfolds the parent of what it reaches.** The list narrows in
place; a parent that does not match but holds a match stays, faint, saying
`holds a match`, its sub-objectives that match under it and the match marked
in the title; the head counts `1 of 17`, and the detail shows the first
match. Cleared, every fold is what it was before — but a parent under which
the person chose a sub-objective stays open, so what was chosen stays in
sight (`engine.FoundBranches`, `@gui-objective-a-search-unfolds-the-parent-of-a-match`,
`@gui-objective-a-match-chosen-keeps-its-parent-open`).

**A folded parent says what it hides that wants seeing**: a sub-objective
late — its deadline past, not reached — or waiting on one not reached, the
late first, one named and the rest counted, after a dot in the late ink,
`#13 Documentation late +1`; at 720 by number alone, `#13 late`, and the
counts give their time up first (`engine.Snapshot.Hidden`, `CountsIn`,
`@gui-objective-a-folded-parent-says-what-it-hides`,
`@gui-objective-a-narrow-window-shortens-what-is-hidden`). The pointer on
it says the whole of it. On hold and reached stay inside: they need no
action. Unfolded, the parent drops it, since each row says its own, `31 Oct,
late` or `waits on Firmware`, and a tile in the detail says `late · its
deadline was Fri 6 Nov`.

The detail heads the objective with its number, offers *Show its tasks* `t`,
on one of level one *+ Sub-objective* `⇧N` — the form of a new objective,
its parent set to this one (`@gui-objective-the-detail-adds-a-sub-objective`);
at 720, where the heading would lose its room, the button gives way and `⇧N`
stays — and *Edit* `e` beside it, and writes its standing, its labels, and where it
sits on the line under (`Detail`,
`@gui-objective-the-detail-shows-the-objective-chosen`). Its progress, the
time spent and what is left are counted over its subtree, each task once —
adding the sub-objectives up would count a task serving two of them twice
(`@gui-objective-the-detail-counts-each-task-once`). A parent shows its
sub-objectives as tiles (`Card` with a `GoalTile`), one row of them, each
saying what else its tasks serve and warning, in the late ink, of an
objective it waits on not reached (`@gui-objective-a-tile-warns-of-what-it-waits-on`);
then its tasks on it itself, then those in its sub-objectives, each after a
drawn `◇` and the sub-objectives it serves, the open first and the done
faded; what has no room is counted, `2 more · t shows them all on the board`.
What it waits on, and what waits on it, are said as facts
(`@gui-objective-the-detail-says-what-waits-on-what`). A parent every
sub-objective of which is reached is proposed as reached in a banner, and
left where it stands: its standing is the person's
(`@gui-objective-a-parent-is-proposed-as-reached`).

**`t` shows its tasks** on the tasks board, the objective filter put on it,
which reaches its sub-objectives (`@gui-objective-t-shows-its-tasks`). Beside
the pill, which then names the objective after a drawn `◇` and says `with its
sub-objectives` faint after it where there is room
(`@gui-objective-the-pill-says-with-its-sub-objectives`) — under the pointer
where there is none, as at 1200 beside the workspace and four segments — a
bar of segments counts all of it, each sub-objective, and the objective
itself, and narrows to the one clicked, or walked to with `[` and `]`
(`Narrowing`, `@gui-objective-the-narrowing-bar-counts-each-part`); the pill
gives up its name and the pills clearing and adding filters give up their
room for it. Where even then the whole bar has none, at 720, it goes compact:
the part in force alone, counted, between two chevrons walking the others
(`Segmented` with `Back` and `Next`,
`@gui-objective-a-narrow-window-narrows-too`), and the pills keep what room
is left. A view naming one objective holding sub-objectives draws the same
bar; a part chosen is a filter on top of the view, and *all* takes it off
again (`@gui-view-naming-an-objective-narrows-by-its-parts`). A task outside the objective that blocks one in it stays in
sight as a reference, dashed and faint, saying `outside this view · blocks
#2` in the late ink, at the foot of its column, and never counted
(`References`, `@gui-objective-a-blocker-outside-is-a-reference`). Its why
takes the place of its facts, cut short with `…` where the column is narrow,
its labels in what room is left (`ReferenceRowOf`); in a column holding
nothing else, it stands under *Nothing here*, never over it. A card
names what it serves by name after a drawn `◇` — `◇ Firmware`, `◇ Enclosure
+1` — where it wrote `◇6` (`@gui-objective-a-card-names-what-it-serves`).

**A view is the board's filters kept under a name** (#226,
[listing.md](listing.md#a-view-keeps-the-filters-under-a-name)), opened
again in one act. Whatever narrows the board — the filters and the search —
is kept from the foot of the filter menu, `v`, from the menu bar's View ›
*Keep as a view…*, or from the command bar, `view new Shopping`
(`@gui-view-kept-from-the-filters`, `@gui-view-kept-from-the-command-bar`).
`v` and not the mockup's `k`: in the filter menu `k` walks up, as `j` walks
down. The popover that keeps it drops under the pill, as a filter's values do
(`OnKeeping`, `KeepingBox`): a field holding the name proposed — what each
filter is on, an objective by its title, joined by a middle dot
(`Filters.Proposed`) — what it would keep of the board, and the key it will
get. `⏎` keeps it, and the pills fold into its own. What `t` shows of an
objective is the same narrowing, a view without a name: kept, it is proposed
under the objective's name (`@gui-view-shown-tasks-kept-under-the-objective`).

`⌘1`…`⌘9` open the view in that place, `⌘0` leaves it for the whole board;
`⌘⇧` and a figure still choose a workspace, so a figure under `⌘` is a view
and under `⌘⇧` a workspace (`@gui-view-opened-by-its-key`). No key used them
before. The filter menu lists the views at its head, under *Views*, each with
how many tasks it keeps and its key, the one in force taken, and *Edit the
views…* under them; a filter the view holds says *in the view* in the faint
ink (`ViewRows`, `@gui-view-listed-at-the-head-of-the-filters`). The arrows
walk up from the first filter into the views and down out of the last, and
`⏎` opens the one walked to (`ViewWalked`,
`@gui-view-the-keyboard-walks-the-views`). A tenth view has no key and is
opened from the menu. No view opens with the window.

**The view in force is one pill**, `view Home automation`, standing first,
where the filters it holds would stand — a view of four filters takes the
room of one (`PillView`). Its cross leaves it. A filter or a search added on
top is an ordinary pill after it, and marks the view's pill changed with a
dot in the late ink (`FilterPill{Changed}`, a `Dot{Late}`,
`@gui-view-a-filter-on-top-marks-it-changed`). Nothing changed is saved until
asked: a click on the pill opens the view's popover, saying what it holds,
what it shows of the board and how much is done, and for each objective it
holds that objective's own figures, *the store's* — a view narrows what is
shown, never what an objective counts. *Save to the view* `⏎` puts what is on
top into it (a label or an objective joins those it holds, any other filter
takes the place of its own: `Filters.Saving`), *Keep as a new view…* `v`
keeps both under another name. Leaving the view drops what was on top
without asking, since it never touched the view, and opening it again gives
the board it was kept with (`@gui-view-left-drops-the-change`).

**What lies outside is a reference, never counted.** A task outside the view
that blocks one inside it, waits on one, or is linked to one stands at the
foot of its column as a reference, once, every why joined — `outside this
view · waits on #3 · linked to #3` — and adds nothing to a column's head or a
tab's count (`engine.Touching`, `@gui-view-outside-is-a-reference`). One step
only: a reference's own blockers are not drawn. The board narrowed to an
objective by `t` draws its references the same way, so it says every why
too, where it named only blockers.

**The views have a panel of their own** (`OnViews`), opened from the menu's
last view line, the menu bar's View › *Views…* while there is a view, or the
command bar. A row is a view: its name, what it holds in the faint ink — the
late ink when something it holds is gone — how many tasks it keeps, and its
key. `r` renames it in place, `e` opens it on the board, `⇧↑` `⇧↓` move it,
which changes its key, and `⌘⌫` deletes it without asking, since no task, no
time and no note goes with it
(`@gui-view-the-panel-renames-moves-and-deletes`). A view whose objective or
label is gone says so in a warning banner under the rows, with what to do
(`@gui-view-gone-is-said`).

**Deleting an objective asks, row by row** (`Fates`,
`@gui-objective-deleting-asks-row-by-row`): each sub-objective deleted or kept
as level one, and the tasks it touches deleted or detached, each on a row
with a small `Segmented` at its right. Nothing is chosen for the person: the
button that deletes is faint, and does nothing, until every row is answered,
and then names what it deletes, `Delete #1 and #3`. `esc` keeps everything
(`@gui-objective-a-deletion-asked-is-kept-on-escape`). An objective holding
nothing is asked about as a task is.

**Work started early is told** above the bar, where the window says what it
did: putting a task in current — moved, focused, edited — while an objective
it counts in waits on one not reached says so, as the engine warns
(`@gui-objective-work-started-early-is-warned-about`).

**The tab somebody is on stands on a pill.** The tabs were folders, a band
closed by a line that the open tab broke; the design has no band and no line,
and the page's tab stands on a rounded pill with its words in the ink, the
others faint on the window's ground. Each writes its name in words and how
much its page holds, counted off what is showing, so a search narrows what the
tabs claim as well. `tab` walks to the next page and `shift+tab` to the one
before, round the end rather than stopping at it; inside a form `tab` still
walks the fields, as it does in the terminal.

The columns stand a gap clear of that band. Against it, the tabs read as the
board's first row.

**The card in focus carries a glowing dot at the right of its first line.** The terminal board's eleven
marks were left out of this window on purpose; this one earns its place
because it cannot be worked out by looking: a card gives no sign that somebody
— or some agent session — is on it. A glowing dot in the accent — a dot of
3.4 in a halo of 6 of the same ink at three tenths, drawn in fourteen and
scaled to its box (`GlowOf`, `Accent.Glow`) — stands at the right end of the
title's first line, where the eye finds it in the same place on every card;
the first line makes room for it and the second keeps the card's whole width
(`FocusMark`, `TitleRoom`, `@gui-the-card-in-focus-carries-a-glowing-dot`).
It was a small cairn of three stones; the dot replaced it, chosen among nine
proposals, in the same box, so nothing around it moved. It followed the last word of the title, and so
moved from card to card. The card
stands on the tint: the eye after its number was a pictogram nobody read as being
worked on, and was too small to find from across the board. `f` focuses or
unfocuses the chosen task, under the same key the terminal board uses.

**The bar at the window's foot is the command bar.** It is a rounded panel
across the foot: at its left the key that opens it, `⌘K`, on a chip, and a
line saying the task chosen — `#152 · Choose the month…` — or that a command,
a number or a title can be typed (`CommandSays`,
`@gui-the-command-bar-names-what-is-chosen`); `esc` on the board lets the
chosen card go, after what was taken together and a kept search, which it
clears first (`@gui-esc-lets-the-chosen-card-go`); at its right, the offers of the
moment as hints, each its key on a chip and its word, lit under the pointer,
and done by a click as by the key (`barAlong`, `drawHint`,
`@gui-a-lit-offer-holds-its-words-in-its-middle`). Under an open panel the
board's hints stay, faint. Until the palette lands, every offer the bar had
is still a hint, so nothing became reachable by key alone.

**The bar's line keeps its room.** When the hints would leave it less than
`barLineLeast`, they give their words up and show their keys alone, and if even
the keys do not fit, the first of them go — the generic ones, so the actions on
the task chosen stay (`barLaid`, `barWidths`, `@gui-a-narrow-bar-keeps-its-line`).
At the narrowest window, with a card chosen, eight worded hints used to start
past the window's left edge, over the `⌘K` chip, and the line was gone. What a
hint no longer says is still its key, and still in the palette.

**Every text is drawn where it fits.** `TestTheFixturesDrawEveryTextWhereItFits`
draws every screen the overlap check draws, on the fixtures at 720 by 560 and
1200 by 800, through the one function every text goes through (`sayWith`, its
`noteSaid` hook), and fails on a text drawn in an empty box or one taller than
its box. A one-line text wider than its box is not a fault: it is cut with "…".
Overlaps between neighbours are the design system's to find
([the design system](#the-design-system)).

**`⌘K` opens the command bar into a palette.** A click on the bar's line does
the same. It drops into place over the top of the window, under the tabs and
in its middle, half the window wide, what is typed at its head, while the bar
at the foot keeps its hints (`PaletteBox`, `@gui-the-palette-opens-at-the-top`):
it used to rise from the bar, where no palette elsewhere opens — VS Code,
GitHub, Linear and Superhuman all open theirs at the top. Every action says
the key that does the same on the board, at its right
(`PaletteRow.Key`), so the palette teaches the keys it stands for. It offers,
in this order: the
actions on the task chosen (`Move #152 to…`, `Priority of #152…`,
`Label #152 with…`, `Edit`, `Focus`, `Archive`, `Delete`) and the board's own
(`New task`, `Search`, `Filters`, `Labels`, `Help`…); then, once something is
typed, the tasks it reaches by the engine's own rule (`engine.Reaches`), five
at most, `#152` putting that task first; then what was done from it lately
(`PaletteRows`), each saying how long ago it was done, `2 min ago`
(`PaletteRow.Ago`). `↑↓` walk the lines, the one walked to carrying `⏎` on a
chip at its right (`PaletteKeyAt`), `⏎` does it, `esc` — said on a chip at
the head, `esc close` — or `⌫` on nothing typed — shuts it. An action that needs a value asks for it in the same place:
the columns, the four priorities, the board's labels and one to create, each
written through `task edit` as the right-click menu writes it; an action is
otherwise the key it stands for, so the palette does nothing the keys do not
(`palette.go`, `@gui-the-palette-moves-the-chosen-task`,
`@gui-the-palette-opens-a-task-found-by-title`, `@gui-esc-shuts-the-palette`). The
same action can be typed whole, value and all: `move to done`, `move 152 to done` or
`priority of #3 high` offers the line `Move #152 to done` that the two steps
would have reached, on the task the number names or else on the one chosen
(`commandRows`, `@gui-the-palette-takes-a-command-typed-whole`); typed whole,
it used to reach nothing, though it said exactly what choosing would do.

**The search is the command bar after a slash.** `/` on the board, or a slash
typed first in the palette, turns the bar into the board's search: a `/` chip
at its left, what is typed after it, and `⏎ keep` and `esc` at its right; the
board narrows at each letter as it always did. `⏎` keeps the search and gives
the keys back to the board, the search standing as a pill at the right of
the tabs whose cross clears it (`SearchPill`, `@gui-a-slash-in-the-palette-searches-the-board`,
`@gui-a-search-kept-is-a-pill`). **`esc` and `⌫` on nothing left go back to
where the search was started**: the board for `/` typed on it, the palette for
a slash typed into it (`leavingSearch`,
`@gui-esc-on-a-search-goes-back-where-it-came-from`,
`@gui-backspace-on-an-empty-search-goes-back-where-it-came-from`). A search
started on the board used to end in the palette, a place it had never been —
the W3C's modal dialog gives focus back to what opened it. A task's number,
`152` or `#152`, typed in the palette offers `Open #152 · its title` first,
and `⏎` opens it (`@gui-a-number-opens-its-task`); it used to only choose the
card, and a second `⏎` opened it.

**One letter, one verb, and no letter destroys.** On a chosen task `m` opens
its columns under its card, a column's first letter moving it there; `p`
its priorities; `l` its labels; `t` its deadline's calendar — the values the
right-click menu offers, the same writes (`settingUnderTheCard`,
`@gui-m-moves-the-chosen-task`, `@gui-p-sets-the-priority`,
`@gui-l-opens-the-labels-and-t-the-calendar`). Deleting is `⌘⌫` and asks
first; `d` deleted, one slip of a finger from losing a task, and now does
nothing on the board (`@gui-d-deletes-nothing`). `g` then `t`, `o` or `i`,
within a second, goes to a page; on TIME, `1`, `2` and `3` show a day, a week
or a month, where `d` and `m` did (`@gui-g-goes-to-a-page`). Arrows walk the
board; `h j k l` were left out, `l` being the labels.

**Labels are picked under the card, and written as they are taken.** `l` on
a chosen card, or Labels… in its right-click menu, opens the labels under the
card — above it when the window has no room below — with no form behind
(`Picking`, `@gui-l-picks-labels-under-the-card`): it used to open the whole
form for one field and write nothing until the form was saved. Each label
taken or let go is written at once, through `task edit` with `+label` or
`-label`, and `esc` closes. The labels the task carries stand as pills at the
head of the list, a cross on each, and `⌫` on nothing typed takes the last
off (`LabelsTaken`, `@gui-the-labels-taken-stand-as-pills`). Each row says
how many tasks carry its label (`@gui-a-label-offered-says-what-carries-it`).
What is typed narrows the list, and a label no row holds is offered to be
made after the ones found, so `⏎` takes the first label found rather than
making a new one from a beginning; in a form the same list answers its labels
field, written when the form is saved.

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
changed underneath, and made two places to click for one thing. The same
holds between panels: a form under a list or the calendar draws none of its
offers, which were theirs read a second time at its foot. Each offer is as wide
as what it says, never narrower than the others: at one width for all,
`backspace  No day` ran into the `esc  Cancel` after it. The bar under the
board lays its offers out the same way (`BarOffers`): spelled for Windows,
`Shift+A  Archived` ran into the `?  Help` after it.

**Archive** is one of those offers, under `x`, beside *Edit* and *Delete* on a
chosen card. On a card already archived the bar reads *Unarchive* instead:
one key both ways, and the bar says which way it goes rather than leaving it to
be worked out. Unlike *Delete* it asks nothing first — nothing is lost.

**A whole column is archived on `⇧X`**, the column the chosen card
stands in. It asks first and names the column: nothing is lost, but a key that
moves twelve cards is not a key to press and find out about. It is not on the
bar — the bar has room for what one card does, and this does what a column
does — so it lives in the help, under the terminal board's own key. The engine
has swept by a narrowing since the beginning; only the command line ever asked
it to.

**One screen asks the questions, and there are three.** Deleting a task and
throwing away a form somebody has typed into both lose work; archiving a column
does not, and is asked about for its size. The question is the panel's head,
the lines under it say what follows, softer, and the two offers are the
question's own verb and *Cancel*: `Delete ⏎`, `Archive ⏎`, `Restore ⏎`,
`Finish ⏎`, `Replace ⏎` (`agreeingVerb`, `@gui-a-question-answers-to-its-verb`).
They were *Agree* and *Cancel*, with a line saying "Enter to agree, or
Cancel." under every question: the buttons said it again, and *Agree* never
said what was agreed to. `⏎` does the verb; under the question, the board's
bar keeps the board's keys (`BarKey`). Deleting a task names what goes with
it: its steps, its labels, and its time spent when there is any
(`@gui-a-question-on-a-task-says-what-goes-with-it`). A form's question
offers *Throw away* and *Keep editing*. The targets are as tall as the bar's — they
used to be the smallest in the window, for the most destructive act in it.
A card dropped in done while one of its steps is open is asked about too —
"Put task 7 in done, ticking its steps?", with "Its steps not yet done are
ticked with it." under it — and agreed to, its steps are ticked in the act that
moves it (`@gui-a-drop-in-done-asks-about-open-steps`); see
[steps.md](steps.md#done-with-a-step-still-open).
A question is asked in a box as big as its sentences and its offers, not in
the panel a form is drawn in: two lines and two words used to sit in a box
nine tenths empty.
**A question stands out from what is behind it**: everything under it is
dimmed by a veil (`Surface.Veil`), and its box is outlined twice over in the
colour of what is chosen
(`@gui-a-question-stands-out-from-what-is-behind-it`). It was drawn like a
form, the same ground and the same edge, and over a form it blended into it.

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

**The description scrolls, and says so.** Read rather than written, nothing
follows a caret, and the rows its box holds were the whole of what a long
description could ever show — the rest was clipped, with nothing on screen
saying it was there. It now takes the wheel and carries a bar
beside it, reusing the trace's `Thumb` and window rather than growing a
second scrolling mechanism. The bar is drawn whether or not there is anything
to scroll, by the same rule the trace goes by.

The form is left alone: there the caret drags the window along with it, so what
does not fit is already reachable.

**The trace is written in the reading ink**, with only the timestamp left
faint. It was thirty lines of the palest grey in the window — a thing somebody
opened a task to read, written in the ink kept for the things beside what is
read. A line a long act wrapped onto carries no timestamp and is written whole,
because splitting one sentence across two inks makes it read as two.

**The logo is a stencil card: a card bitten into a c, cut in three, its last
piece lit** (#255, mark 21 of three rounds of proposals). The card is the
board; the two cuts, running from the centre of the bite, are its columns, and
the piece lit in the accent is where a task goes round to: done. It keeps a
card's solid silhouette, which is what holds up as an icon, and still reads in
one colour, the cuts carrying the three columns. As the icon it stands on the
window's own dark ground, `#0C0D0F`, a squircle with its corners left clear,
the card in the window's ink and the lit piece in its accent. It replaced a
card in focus built of stones in a dial that filled, which said too much to
be read at 32 pixels. Its cuts are diagonal and land on no pixel, so at 16
pixels they rely on antialiasing and are faint; a hand-drawn 16-pixel master
is what would make them crisp, and none is drawn yet.
`internal/logo/cairn.svg` is the one source; the PNGs
beside it, 16 to 1024 pixels, and the Windows resource are drawn from it by
`scripts/logo.sh`, so a build needs no drawing tool.

Each system takes it its own way, and the window is known to all three as
`cairn` (`logo.go`). Windows reads it from the `.exe`, where the resource beside
`cmd/cairn` puts it at every size; Gio gives the window the icon the program
carries. macOS shows a bare binary with no icon, so the window hands AppKit the
logo when it starts (`logo_darwin.go`), once the application has finished
launching: Gio makes it a Dock application then, which takes back an icon
handed before. Which came first was a race, lost under `make dev`, whose
`go run` starts a binary macOS has not seen, so the Dock showed the blank
icon there and the logo under `bin/cairn`. On Linux the window does both things a
desktop may read it from (`logo_linux.go`, `desktop.go`, the entry itself in `internal/linuxapp`). On X11 it hands the
window manager the logo itself, at 16 to 128 pixels, as `_NET_WM_ICON`
(`@gui-x11-is-handed-the-logo`) — Gio sets the window's class and nothing more.
On Wayland there is no such thing outside a protocol few compositors speak
(xdg-toplevel-icon), and a desktop takes the logo from the entry named after
the app_id; so the window writes that entry, and the logo at every size, under
the user's data folder each time it opens, rewriting only what changed
(`@gui-a-linux-desktop-is-given-an-entry`). It used to be `make desktop`'s
job, which nobody installing a binary ran, and the window had no logo. An
entry somebody else wrote there is theirs and is left as it is
(`@gui-an-entry-of-somebody-elses-is-kept`); cairn marks its own. But an entry
of theirs that shows another icon hides the logo, which is how one written by
hand on a KDE desktop, `Icon=view-list-details`, kept the logo out of sight
with nothing saying why. So when the window opens and finds one,
it asks "Show cairn's logo on this desktop?" (`Settling`, `Hiding`): agreed,
cairn writes its own entry over it
(`@gui-an-entry-hiding-the-logo-is-replaced-once-agreed`); refused, the entry
stays and a sum of it is kept under `$XDG_DATA_HOME/cairn/entry-kept`, so the
question comes back only once the entry has changed
(`@gui-an-entry-kept-is-asked-about-again-only-once-it-changes`). An entry of
theirs that shows the logo is left without a word
(`@gui-an-entry-showing-the-logo-asks-nothing`). The entry
starts cairn by a full path, never by its bare name (`Starter`): a desktop
drops an entry whose program its own `PATH` does not reach, and with it the
logo, and Homebrew's `bin` is often on the shell's `PATH` alone. When the
`PATH` leads to the cairn running, the entry names the path it leads by, such
as Homebrew's link, which an upgrade keeps
(`@gui-an-entry-starts-cairn-by-the-path-it-is-found-at`); otherwise it names
the cairn running (`@gui-an-entry-starts-a-cairn-off-the-path-where-it-is`).

**One intro builds the board of stones** (`stones`, one of the eleven
[intros](#the-board-arrives-by-an-intro); the only one until #255). The
content is what is built: each card arrives
as stones — courses of them, each stone cut irregular and fitted to its
neighbours like an old dry-stone wall, never bricks — falling from above the
window into its place, one card after another, then giving way to the card
(`StonesOf`, `Arriving`; `cairn.go`). The same card is built of the same
stones on every run, from a generator of its own seeded by the card's number.
A board with no task opens on nothing being built
(`@gui-no-task-no-cairn`). It lasts 1.8 s, starts on
the first frame a screen draws (`Born`), which no scenario draws, and the
first key or press ends it and still does what it does — it is never in the
way. Drawn rather than shaded: Gio runs no shader of a program's own, so a
stone is its outline filled with a gradient, lighter at its top as a stone
lit from above, and a grain painted over it — one image of mottle and pits,
made once from a fixed seed (`StoneGrain`). The joints are tight, a pixel or
two between neighbours, as a wall of dressed stone is laid; once landed,
the stones give way to the redesigned card.

**The time spent is seen a week or a month at a time, every task of it.** How
long the work took was a table the command line printed. The third page, TIME,
lays a week out (`week.go`): a column a day across the board's width, Monday
first, each day's total, and each task worked on drawn as the card the board
draws — its priority, its cairn, its labels, what it has cost — with its time that
day under it and a bar as long as its share of the longest stretch of the week:
a glance says which day went where. A task's lines used to be its own, a second
way of showing the same thing (`@gui-the-time-page-shows-cards`). A day of more tasks than its column holds scrolls under the wheel,
a task a notch, with a bar beside it: it used to cut them off at its foot
(`@gui-a-day-scrolls-to-its-last-task`). The time spent over the period
shown follows the period at the left of the strip, `26 Oct – 1 Nov 2026 ·
34h10m`, and a month's after its year (`TimeSays`, `PeriodAt`); it stood
left of the ranges, `week 34h10m` in Go Mono, where Today hid it. A day of the week says its date at
its head, today's ringed in the accent, and its total at the right end of the same
line, in Go Mono (`WeekTotalAt`): the total followed the date, set off by
spaces, and the spaces of a proportional face never matched today's pill, so
the two were written over each other; put on a line of its own under the
date, it read as a stray line. A week alone said nothing of a month,
so `m`, or Month in the strip along the top, lays out the month's six weeks,
each day its total and as many of its tasks as fit, then how many more; a day
clicked shows its week (`@gui-the-time-page-shows-a-month`). A month shown
names itself with the date picker's two buttons, its month and its year, and
they open the same grids in the room of its days: the twelve months, or twelve
years, the `‹ ›` paging a year or twelve, a pick going back down a level
(`gridCells`, `gridPicked`, shared with the deadline's calendar,
`@gui-the-time-month-is-chosen-from-the-twelve`). The pointer alone opens them
there, `m` being the Month range already. The strip's `‹ ›`
and `←→` walk weeks or months, Today and `t` come back to today's, and the
range shown and today are outlined. A task is chosen with a click and opened
with a double-click, on a week and in a day of a month alike, as on the board
(`@gui-a-card-of-the-week-opens`, `@gui-a-card-of-the-month-opens`); a click
beside the tasks of a month's day still shows its week. What has not happened
yet is not asked for:
the engine refuses a period running past today, which left the week in progress
unreadable, so a period is read up to today. A day is seen alone too: Day in the strip,
`d`, or a click on a day's head in a week, lays that day across the page, each
task the card the board draws, a third of the page wide so its title reads,
with its time and its bar beside it (`@gui-the-time-page-shows-a-day`,
`@gui-a-day-of-the-week-is-shown-alone`). The arrows then walk days; Today
comes back to today in whichever range is shown, as a calendar's Today does,
and Day, Week and Month say which range that is. They are buttons, drawn as
such — filled, edged, their name in their middle — and the range shown is the
one filled with the accent: drawn as plain words they were not seen as
something to press. The search narrows this page as it narrows the board: a
day keeps the tasks the search reaches, and its total counts only them
(`@gui-the-time-page-is-searched`). A card too narrow for all it says gives up
what has no room rather than writing it over the rest: the labels stop short of
the facts written hard right, and the title leaves room for the cairn of a
task in focus. The tab says how many tasks the
time shown was spent on, from the moment the window opens: the week is read
with every look at the store and not only once the page is opened, or the tab
said 0 until it was clicked (`@gui-the-time-tab-counts-before-it-is-opened`). The window is told the
day before it first reads the store, and reads no week before it knows the
day (`readingWeek`, `TestAWindowWithNoClockYetReadsNoWeek`): reading first,
it asked for the week of year one, the engine answered with today alone, and
the tab said 24 until it was opened and 64 after.
A company's week is five days, and Google Calendar lets
a week and a month leave out Saturday and Sunday for it (Settings, View
options, Show weekends, on by default; the day view is untouched), so the
configuration does the same: `[time] weekends = "hidden"` cuts the week to five
columns, Monday to Friday, each wider, and the month to five columns a week, and
the strip says the week ends on its Friday (`@gui-a-week-without-its-weekend`).
The days a week shows are the engine's to say (`Pages.WeekDays`), the one
question both boards ask, so the terminal's week is not a second answer. Only
what is drawn is left out: the week is still read whole, so time spent on a
Saturday is not lost from any total that counts it, and the hidden days' cards
are simply not drawn. The key is set from the settings, `Weekends`, a choice
of two, and the week redraws at once (see
[config.md](config.md#the-weekends-of-a-week)).
The page lies over the columns and draws none of them,
so nothing on it chooses a card behind it. With three pages, `tab` and `shift+tab` walk them round,
which with two was a difference nothing could see.

**Every label is seen, renamed and deleted from the window.** `⇧L`, or Labels
in the File menu, lists them with what carries each, as `cairn label list`
does (`labels.go`). The list is the engine's answer, not the board's cards, so
archived tasks count. A row is a dot in the label's colour, the label as the
board draws it, and what carries it at its right; the row chosen also carries
its actions as buttons, Rename `r`, Colour `c` and Delete `⌘⌫`
(`LabelRows`, `@gui-the-chosen-label-carries-its-actions`), so the panel's
foot only makes a label (`n`) or closes. They used to be five buttons at the
foot, acting on a row they stood far from. `r` types a new name in the row,
`⏎` takes it, `esc` gives it up; the dot, or `c`, opens the fourteen colours
under the row (`@gui-a-label-is-coloured-from-its-dot`); `⌘⌫` asks, then
takes the label off every task and objective
(`@gui-a-label-is-deleted-everywhere`).

**The labels are sought from the top of their panel.** Its first line is a
search: `/`, or a click on it, types into it, and the list keeps the labels
whose name holds what is typed (`LabelsShown`, `@gui-the-labels-are-sought`);
the panel keeps its size, so the rows do not jump as letters are typed. `esc`
leaves the search and lets every label back. A search no label holds offers
to make it, `Create “Release Train”`, and `⏎` makes it, written the way a
label is (`LabelCreating`, `@gui-a-label-sought-and-missing-is-made`) — the
way GitHub's labels page and Trello's picker make one from what was looked for.

**A new backup is made, and a backup restored, from the window.** Both were
the command line's, so a person at the window went to a terminal for either.
`⌘B` makes a new backup and says where it lies; `⇧B` lists the backups, and
`⏎` restores one once agreed to, the engine making a new backup of what the
store holds first, as it always does (`backups.go`). What it tells is said above the bar in
the ink of what is read, and goes as a refusal does. A line of the list says the backup's
number, when it was made and how big it is, `backup 3   2026-11-01 09:12
48 KB` (`domain.FormatSize`), and the panel offers *Restore* (`⏎`) and
*Close* (`esc`) rather than a head spelling out the key. Why a backup was
made is not said: nothing records it, and a name guessed from its moment
would be wrong as often as right.

**A store behind the window is offered the changes it waits for.** A newer
cairn — the MCP server Claude Code restarted, a terminal — may apply a change
to the store while an older window stays open, and the next window opened on
an older store used to be refused at the door, or worse, kept drawing a board
the store no longer had the tables for. Now `cairn gui` opens on such a store
(`@store-behind-opens-the-window`) and the window asks "Apply what this store
is waiting for?", naming each change, a backup set aside first. "Apply" runs
`migration apply` and the board comes back
(`@gui-a-store-behind-is-offered-its-changes`); "Quit" (`⌘Q`) closes the
window and leaves the store as it was
(`@gui-a-store-behind-is-left-as-it-was-when-quit`). Nothing else is reached
while the question stands (`@gui-a-store-behind-holds-the-board-back`), and
it is never aged away the way a refusal is: it is not an error, it is a
question. Which refusal is a question, and which act agrees to it, is the
engine's to say (`engine.AgreeingTo`), as for a task finished with a step
still open; the engine also refuses, on every ask, what needs a current store
while changes wait. The terminal board asks the same, see
[tui.md](tui.md).

The same holds while the window is open: a store another cairn rolls back
under it is noticed at the next beat, and asked about
(`@gui-a-store-put-behind-while-open-is-noticed`). A store a newer cairn
changed — under the window (`@gui-a-store-from-a-newer-cairn-stops-the-window`)
or before it opened (`@gui-a-store-from-a-newer-cairn-at-open`) — cannot be
read or written by this one: the window says "This store comes from a newer
cairn", offers only "Quit", and nothing dragged or typed reaches the store,
which the engine refuses whole (`Hold` keeps such a store without opening
it). An apply that fails is said, and the question is asked again; a question
that comes while a task is being added keeps what was typed as its draft, as
a window closing does.

**Tasks are exported, and imported, as a file chosen in the system's
dialog.** `⌘E` and `⌘O`, on the bar of a board with nothing chosen and on a
right click on the board, open the system's own file dialog (AppKit's panels
on macOS, `gioui.org/x/explorer` elsewhere, see
[dependencies.md](dependencies.md)): the engine writes every task out, with
every objective (`task export`), and the file is saved where the
dialog says; or a file chosen there is copied aside and the engine adds its
tasks as new ones (`task import`). Nothing on the board is replaced, so
nothing is asked first. An earlier version carried the whole store and put
one in its place, which is a backup's work, not an export's. The dialog blocks
until it is answered, and a window that is a value never waits: the window
says what it wants (`FileAsked`), the loop opens the dialog on the side
(`Filing`) and hands the answer back (`Filed`) — the one piece no test
reaches is the call to the system, and the scenarios play the dialog the way
they play the clipboard (`files.go`, `@gui-tasks-are-imported`).

**A page turned off is not drawn.** The pages the configuration turns off have
no tab, and `tab` and a click walk only the others; the window opens on the
first page left (`@gui-a-page-off-has-no-tab`, `@gui-the-first-page-shown-opens`).
With all three off the window is its ground and nothing else, and takes no key
and no click (`@gui-every-page-off-is-an-empty-window`). See
[config.md](config.md#the-pages-turned-off-one-by-one).

**Every act is in the window's menu bar.** The window had none of its own, and
a person at a window looks for what it does in its menus. File, Edit, View and
Help hold every act the window has (`MenuBar`), each line an offer the bar, the
help or a key already makes, done the same way (`Picked`) — from the board
only: over a form it does nothing, as a key the form does not take does
nothing. Gio builds the macOS menu itself and offers no way to add to it, so
cairn adds its menus to Gio's from a few lines of Objective-C, the ⌘ chords as
the menu's own keys (`appkit_darwin.go`); on Windows the menu is a Win32 one set on
the window Gio made, its commands caught in front of Gio's window procedure and
the keys written beside each line (`menu_windows.go`). It is set from the
window's own thread (`onTheWindowsThread`, through Gio's `Window.Run`): set
from the loop while Gio's thread waited for the loop to take the event that
named the window, `SetMenu` waited on that thread in turn, and the window
froze once drawn, on Windows 10, with nothing written anywhere. Windows names
a key by its place on an American keyboard, so `?` typed on an AZERTY one
arrives as `shift+,` and `/` as `shift+:`; the mark itself comes only as text.
On a screen that takes no text, a lone mark the system says was typed is taken
as its key, unless the press before it named it already, as macOS does
(`symbolTyped`). Neither call to the
system is driven by a test; what each line does is. Where the system gives a
window no menu bar — Linux — a `☰` button at the end of the tabs opens those
lines as one menu (`MenuButton`). Help ends on About cairn, which says which
build this is — its version, the commit and its date, what built it for which
system — the lines `cairn --version` prints, read from the same stamps
(`internal/build`), and where cairn is kept, its GitHub page (`build.Home`), in a box with only Close
(`@gui-the-version-is-shown-from-the-menu`).

**A line of the menu bar is offered when it can be done now.** The bar was
one fixed list, every line there whatever was in sight, so Focus stood over
the objectives and Edit with nothing chosen, and picking them did nothing. Each line now names the acts it does (`actsOf`), which the pages
offer or not (`engine.Pages`), and `Doable` says whether it does
something now: on the board, with what it acts on in sight — a card chosen to
edit or delete, a task to focus or archive, another page to go to
(`@gui-the-menu-offers-what-can-be-done-now`,
`@gui-the-menu-follows-the-page-in-sight`,
`@gui-the-menu-waits-while-a-panel-is-open`). The `☰` menu leaves out the
rest, and picking it from anywhere does nothing
(`@gui-a-line-that-cannot-be-done-does-nothing`); the native bars keep their
lines in place and grey them, reading what the window told them at its last
frame (`tellingTheBars`) as a menu opens — AppKit through `validateMenuItem:`,
Win32 on `WM_INITMENUPOPUP` — so a line does not move under the pointer as
the board changes.

**A right click says what can be done to what is under it.** What a card can
have done to it was on the bar once the card was chosen, so a hand on the
mouse went to the card and then down to the foot of the window. A right press
chooses what is under it and opens a menu beside the pointer (`menu.go`):
Open (`⏎`, as the design puts it first: reading a task is what a click on it
most often leads to), Edit, New linked task, Focus or Unfocus, Move to…,
Priority…, Labels…, Archive or Unarchive, Delete on a task; Edit and Delete on an objective; the board's own offers
where there is neither. Every line is an act the bar, a key or a drag already
does, done the same way. Move to… and Priority… turn the menu into the values
the task may take, but the one it holds, and the one clicked is set at once;
Labels… opens the list of its labels, as the form does
(`@gui-a-card-is-moved-from-its-menu`). New linked task is `o` on a chosen card:
the form of a new task, in `todo`, its `links` already holding the card's
number (`@gui-a-cards-menu-opens-a-linked-task`, `@gui-o-opens-a-linked-task`);
an unsaved new task held is asked about as for any other, and the link stays
whichever the answer. See [links.md](links.md). The arrows and
`j k` walk it, `⏎` takes a line, `esc` or a click off it closes it and does
nothing else.

**A deadline is chosen on a calendar.** It was typed as a date, which meant
knowing what date next Friday falls on. A click on the field, `⏎` on it, or
Choose, opens a month at a time hung under the field, the way a list is — it
first opened in the middle of the window, far from the field it fills in, and
only by a key (`calendar.go`, `@gui-a-click-on-a-deadline-opens-the-calendar`).
Its days share its width, a week at least seven of them at their narrowest. It
has no heading: the field it hangs from and its tooltip already say what it is
for. Its top row holds the month and the year left, as two chips of fixed
widths, and the `‹ ›` arrows right, so none moves from one month to the next —
the order of a date picker such as Material's. The row is 32 high and its words
14 and semibold, centred: the month and the year used to be written at 16 in
boxes 22 high, and were cut (`CalendarHead`,
`@gui-the-calendar-is-headed-by-its-month`). Those arrows and the
wheel walk the months for a hand on the mouse
(`@gui-the-pointer-walks-the-months`). Six weeks from the Monday on or
before the first, so no row moves from one month to the next; the arrows walk
days and weeks, `⇧←→` months — the last day of a longer month landing on the
last of a shorter one, never in the month after — `t` today, `⏎` or a click
takes a day, `⌫` leaves no day and `esc` leaves the field as it was. It opens
on the day the field holds, or today. Today is what the window was last told
it is, each frame (`At`), so a scenario names its own. The field still takes a
date typed. Tasks and objectives open the same one.

**The month and the year are words to click.** Walking to a date months away
by the arrows is a long way, and a calendar on a phone or in a browser lets the
name be tapped. The month's name and the year, at the left of the calendar's
top row, are chips, filled while their grid is open: the month's opens the twelve months, three across and four
down, the year's the twelve years from four before the one shown, and a pick
goes back down one level — a year to its months, a month to its days
(`@gui-a-calendar-month-is-chosen-from-the-twelve`,
`@gui-a-calendar-year-is-chosen-from-a-grid`). In a grid the arrows page a year, or twelve years, and the keyboard does as in the days:
the arrows move, `⏎` picks, `m` and `y` open the grids, `esc` goes back to the
days and does not close the calendar. The days have more room
to be hit and read: 32 high and 6 apart, where they were 28 and 4, and the
panel still sits whole in the smallest window the layouts are drawn at.

**The calendar is drawn with the window's one button.** Its months and years
are buttons, the one marked filled; a day is its number in the middle of its
place, the day held a filled button and today one ringed in the accent, as today is
in every grid — the marks a date picker usually gives them. Today was a black square, and the twelve cells were
words in the corner of a frame (`@gui-the-calendar-draws-its-days-and-cells-as-buttons`).

**A form saves itself.** `⌘S` was the only way a change reached the store,
and every other place a person writes in saves on its own. On each beat of the
window (`Beat`, two seconds) a form that was changed, and holds what it held at
the beat before, is saved where it stands and opened again over what it saved,
read back from the store: the keyboard and the caret where they were, a step
just added in its place, the steps taken still taken
(`@gui-steps-taken-outlast-a-save`), and the trace beside it read again, since it
otherwise lacked the very act just saved
(`@gui-a-form-that-saved-itself-shows-it`). Waiting for a beat with no typing
keeps a save from landing in the middle of a word
(`@gui-typing-holds-the-save-off`). A task being added is not: it is made when
it is saved, by `⌘⏎`, `⌘S` or Save, and not before
(`@gui-a-new-task-is-made-when-it-is-saved`). Amended 2026-10-16: it used to be made at the first beat with a title, so its trace
filled with the acts of its making while it was still being thought about, and
Cancel came too late, the task being already there. Now its trace stays empty
until it is saved (`@gui-a-new-task-has-an-empty-trace`), and Cancel or `esc`
makes nothing and keeps nothing to pick up, without a question — cancelling a
task not yet made loses nothing anybody decided to keep
(`@gui-cancelling-a-new-task-makes-nothing`,
`@gui-escape-on-a-new-task-makes-nothing`); with no title there is nothing to
make either.

What is typed into a new task is not lost for all that: at each pause it is
kept as a **draft**, in a file beside the store (`Places.Draft`,
`cairn.db.draft.json`), which is not a task and is in no trace. Saved, the task
is made and the draft goes; cancelled, it goes too; left — by a click off the
form once the question has been agreed to
(`@gui-clicking-off-a-new-task-saves-nothing`), which is not a decision to make
a task, or by closing the window, which keeps the last keys too — it stays,
and the next new task asks whether to pick it up or start afresh (`@gui-a-new-task-is-kept-as-a-draft`, `@gui-a-draft-can-be-left`,
`@gui-a-saved-task-leaves-no-draft`). Amended 2026-10-16: the first
answer to "Cancel came too late" gave up keeping the typing at all, and left
`⌘Z` on a new form to bring back what was thrown away — the wrong way round on
both counts. An edit's form keeps `⌘Z`; a new one has its draft instead. A form that saved itself has
nothing left to lose, so `esc` closes it without the question. `⌘S` still
saves at once, and closes. What the store refuses is said above the bar, as
any refusal is.

**A text is corrected the way every text field corrects one.** A field knew a
character at a time to either side and nothing more, so a word mistyped at the
start of a long title cost a walk through every letter after it. Now `home`
and `end`, or `⌘←→`, go to either end of the line; `⌥←→` walk a word at a
time; shift holds as either goes; `⇧↑↓` hold whole lines of a description; and
`⌦`, `⌘⌫` and `⌥⌫` take out what is after the caret, back to the start of the
line, or back a word, in one stroke. The search reads the same keys
(`editing.go`). For them to arrive, the window asks for ⌥ too, and a key that
moves the caret keeps every modifier it came with. A run held across the lines
of a description, with the pointer or with the keys, goes in one stroke
(`@gui-a-run-held-across-lines-goes-in-one-stroke`). Shift holds nothing
when deleting. `⇧⌫` and `⇧⌦` delete as `⌫` and `⌦` do, the way they do in
every other field, because a hand still on shift from a capital is common.
Named `shift+backspace`, the key used to reach no binding and did nothing
(`@gui-shift-held-still-deletes`).

**A drag on Windows is not yet understood.** On Windows the card in hand was
drawn at the top left of the window rather than under the pointer. Read
through, nothing between the system and the drawing differs by system: Gio
turns a move with a button held into a drag on every one, the pointer is read
in dp with the ratio the drawing is scaled by, and the card in hand is drawn
where the last drag put it. Rather than a fix guessed at from a Mac, debug
mode (`CAIRN_DEBUG=1 cairn gui`, [debug.md](debug.md)) writes down every
press, drag, release and scroll — what the system sent in pixels, the ratio,
the point in dp, and where the card in hand stands after it (`PointerTrace`);
never the pointer merely passing over (`PointerNoted`). A drag recorded on the
machine that misbehaves says which of those steps went wrong.

The same file holds what the keyboard sends (`KeyTrace`): every key with its
name, every piece of text — as its length, never as itself — with the range
it replaces, and what an input method says of the text and caret it holds;
and, each time it changes, what the window shows (`Seen`). On Linux an accented letter took two
presses; Gio hands keys to IBus on X11 and to the compositor's input method on
Wayland, and whether the letter came late, came as a range the window does
not read, or did not come is what one press recorded there says.

**On Linux and FreeBSD the window keeps IBus out** (`ibusKeptOut`). One press
recorded on KDE under X11 with IBus running said it: no key and no text at all.
Gio hands every key to IBus and takes the letter IBus answers with only once it
has noted a composition — which it notes after IBus has answered, and IBus
sends the letter first. So the first é was dropped and the second written, and
again after every change of focus, which resets the note. Pointed at an address
nothing listens at, Gio goes without IBus, and the keyboard's own layout writes
the letter, dead keys and compose included. `CAIRN_IBUS=1` lets IBus back in,
for an input method a person cannot do without; an `IBUS_ADDRESS` already set
is left as it is.

**Without IBus, a dead key is Gio's to compose, on a press only.** Gio v0.10.3
fed the key's release to xkb's compose state as well as its press, so a dead
key was spent by its own release: `^` then `o` wrote `o`, and `ô` came only by
holding `^` down while pressing `o`, leaving one more `ô` for the next `o`.
Gio fixed it upstream (fd8b8788, ticket 712) after its last release, and cairn
is held at that commit until a release carries it (`dependencies.md`). The fix
lives in Gio's internal xkb, which no test here reaches; it was seen in a Linux
container, where a French keymap typed `oo` before and `ôo` after.

**A key is written the way the keyboard in front of the person writes it.**
The window was written on a Mac and said ⌘ everywhere, so on Windows the help
told somebody to press a key their keyboard does not have. What arrives was
never the problem — `Typed` reads ⌘ and Ctrl alike — so the help and the bar
are still written once, the Mac's way, and `Keyboard.Spelled` writes `Ctrl+`
and `Shift+` for ⌘ and ⇧ away from one
(`@gui-the-help-names-the-keys-of-the-system`). Scenarios run as on a Mac
unless they say otherwise, so what they read does not depend on the machine.

**The help is a grid of keys, and fits the window it is opened in.** It
groups the keys under Board, Find, Forms, Store and Pointer, each key on a
chip and what it does in two or three words beside it, in two columns split
between two groups, what each key does starting where it does for every key
of its column (`HelpLines`, `@gui-the-help-is-a-grid-of-keys`). Thirty-five
lines fit in the least window, so nothing scrolls and no bar is drawn
(`@gui-the-help-fits-the-least-window`); the panel is as tall as its longer
column (`@gui-the-help-fits-what-it-says`). It was fifty lines of sentences
in one column, wrapped, with a bar to scroll them — the old look the
redesign had left behind. What a sentence said and a key cannot, the
buttons and the docs say: the help names the key, the window shows the
rest. See `help.go`. The terminal's help already read on with `↓↑`.

**A refusal gets two lines and the reading size.** The most instructive thing
cairn says — "it is still blocked, by #3 and by #4. take them away first: task
3, task 4" — was the smallest text in the window, on one line, cut just before
the part that says what to do. The two lines sit just above the bar, on a
ground of their own, so the foot of the window stays where it is.

**What went wrong never hides the bar, and does not stay.** It used to take
the bar's place until the next good read, and every offer went with it. Now
the bar keeps offering under it (`@gui-what-went-wrong-leaves-the-bar`), it
goes on its own after `Lingers` beats — four, eight seconds, time to read two
lines (`@gui-what-went-wrong-fades`) — and it belongs to the page it was said
on: changing tab lets go of it (`@gui-what-went-wrong-stays-on-its-page`). The
terminal does the same with the one line it has: the next key gives the bar
back.

**A wheel keeps what did not come to a line.** `by / lineHigh` threw away every
turn smaller than a row, which is every turn a trackpad sends — one pixel at a
time, each rounding to nothing, so a description could not be scrolled with a
mouse at all and a trace answered only a notched wheel. `turning` adds the
leftover to the next turn instead. A trace still moves by whole lines, which
is what its own test asks; six nudges of three pixels are now one of them.

The bar beside a description **can be pulled**, like the trace's. Without it
a press on the bar fell through to the board behind the panel and closed the
task — and the scenario passed anyway, because the end of a description nobody
is showing is zero, which is where an untouched window already sits.

**A number somebody would have to look up is chosen from a list.** The
objectives a task serves are answered that way. The list has no heading: it
hangs from the field that says what it is for, and the heading repeated that
in larger words than anything in the list; its rows begin at its top
(`@gui-a-list-has-no-heading`). The objectives are picked, not searched: what
is typed into their list narrows nothing
(`@gui-objectives-are-chosen-without-a-search`). The list of labels keeps a
line that narrows it as it is typed into, by the board's own matching rule
rather than one of its own: a list that matched its own way would answer a
different set from the search box two keystrokes away.

**A click off a list closes it, as Cancel does.** A list, like a native select,
is left by a click anywhere outside it; a click on the form under it leaves
the form open (`@gui-a-click-off-a-list-closes-it`).

A task serves as many objectives as it serves, so `⏎` ticks and the list
stays open until `esc`. The field then reads `#1 Ship it` rather than `1` — the number is what the
engine takes and what a person has to go and look up, which is the whole of why
the list exists.

**Labels are chosen from the same list**, by name: every label some task
carries, the task's own ticked, a row that creates what was typed under the
name the rule will keep it by, and `⏎` that ticks and stays open, as for
objectives. [labels.md](labels.md) says how a label is written and what the
list offers.

**A closed set is chosen from the same list.** A column, a priority and an
objective's standing open it too — on a click, or on `⏎` — offering the set in
the domain's order, the value already held ticked and under the keyboard, with
no line to search by: a closed list is a handful of fixed values, and a search
to type into only weighs on the field
(`@gui-a-closed-list-is-chosen-without-a-search`), where the list of labels
grows with use, is written into as often as picked, and keeps its search
whatever its length today (`@gui-an-open-list-keeps-its-search`). One answer, so `⏎` takes and closes; but taking the one already held leaves it held, because a task stands
in a column whatever is done to a list. The rows say the names alone: the
number a row is listed under is the list's, not one anybody has seen. `←`/`→`
on the field still step through the set without opening anything.

**The field itself takes no text.** Typed into, a priority once read
`normal2h`: a value the store refuses, and one no row of the list matches, so
the list opened on its first row and `⏎` took `low`. What is typed or pasted on
such a field opens its list, whole, and narrows nothing
(`@gui-a-priority-is-not-found-by-typing`); a key that takes text away does
nothing there.

**A list is the size of what it holds, up to what the window has room for.**
It used to take three quarters of the form whatever it offered: four
priorities sat in a panel made for fifty tasks, most of the form hidden behind
its empty rows. It now hangs under the field it answers, the way a field's
list opens anywhere else — above it when there is no room below — as wide as
its widest row and its offers, and as tall as its rows
(`ChoosingBox`, `@gui-a-list-is-the-size-of-what-it-holds`). It is sized on
what it offers before anything narrows it, so typing does not move every row
from under the pointer. The list of backups is sized the same way, across the
middle of the window (`BackupsBox`), and a form open on a task ends a bar's
height under its fields, as a form adding one does (`OpenOver`).

**Past that room, the list scrolls.** A board of fifty tasks drew fifty rows,
straight past the panel, past the window, and over everything behind both. It
is cut to the run the window has room for and scrolled to for the rest — by the wheel, by a bar beside it, and by the keyboard, which brings
the row it is on into sight the way a column brings a card. The same `Thumb`
and the same window of lines the trace and the description use.

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
list. The list's buttons say *Choose* (`⏎`, which ticks the row or lets it go
again), *Done* (`⌘⏎`) and *Cancel*; a list that takes one answer — the
column, the priority — has no *Done*, its *Choose* closes it. They said
*Take* and *Keep*, two words nobody could tell apart, and the form under the
list offered a *Choose* of its own that opened the list a person was already
in: the form offers no such button now, `⏎` on a field chosen from a list
still opens it (`@gui-enter-on-a-list-field-opens-its-list`). The labels of
a list of labels are pills after their tick, as on a card (`LabelPillAt`).
A double-click is choose-and-done in one gesture, which is what a hand
already does to a list.
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
description to read, `▪ 3/8` how far its steps have got — beside the
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

## An open task

A task opens to be edited, with `enter` or a double click: its form, and
beside the form what was done to it. There used to be a panel to read a task
in and a form to change it, one gesture apart; a task is opened to be worked
on, and the second gesture was a gesture spent for nothing. A form nobody typed
into is put away by `esc` or a click off it, the way the reading panel was. A
form typed into is not put away by a click off it without a word: it asks what
Cancel, `esc` and the cross ask, "Throw away what you typed?", and "Back" gives
the form back with everything in it (`clickedOff`,
`@gui-a-click-off-a-filled-form-asks-first`,
`@gui-clicking-off-a-form-typed-into-throws-nothing-away`). A click off was the
one way out that decided for the person — it used to save an edit, and keep a
new task as a draft, without asking — and a click is the easiest gesture to
make by mistake, so every way out of a filled form now asks the one question.
A new task thrown away that way is still kept as a draft, as closing the window
keeps it. A click
off the calendar or a list closes it and gives the form back; off the form as
well, both go (`@gui-a-click-off-the-form-closes-the-calendar-too`).

**What a task carries stands beside it, and its history behind a tab.** The
left column holds the title, the description and the steps, the description
given three parts of the height left and the steps two; the short fields —
column, priority, labels, the times, what blocks it, its links — stand in
the column beside, under a tab, *Details*, and the history under the other,
*History* (*Work* for an objective), `⌘T` or a click on a tab going from one
to the other (`BesideTabs`, `Tracing`,
`@gui-the-details-stand-beside-and-the-trace-behind-a-tab`). Details shows
first: what is changed most is seen without a click. The keyboard going to a
short field shows Details, and Details stays shown after it leaves: nothing
beside the form changes but by a tab. A narrow column drops the sections'
heads and keeps the fields' names. A field chosen from a list carries a
chevron at its right, a deadline a small calendar (`FieldMark`), and labels
too many for their field show two and how many more, as a card does. The
short fields used to sit between the description and the steps, and the
history beside them took room nobody was reading most of the time.

**An objective's form lists its work under its description.** In place of a
task's steps, under the description's three parts, *Serving it* lists the
tasks serving the objective, a line each: number, title, column as a pill,
time spent; what is current first, what is done last and struck
(`WorkTiles`, `@gui-an-objectives-work-is-current-first`). A click on one
opens it, and so does the keyboard: `tab` from the description reaches the
list, the arrows walk it, outlined and kept in sight, `⏎` opens the line it
is on, and `tab` goes on to the details (`WorkOn`,
`@gui-an-objectives-work-is-walked-with-the-keyboard`); its last line, *+ new task serving it*, opens a new task's form
serving the objective already (`engine.ServingTaskFields`,
`@gui-an-objective-makes-a-task-that-serves-it`). The column beside has no
tabs, an objective having no history: its details — its parent, chosen among
the objectives of level one or none (`Snapshot.Parents`,
`@gui-objective-the-form-offers-level-one-parents`), what it waits on, its
labels among those meant for objectives
(`@gui-objective-the-form-offers-labels-for-objectives`; a task's form offers
those meant for tasks) — then *Progress*, an `InfoBox`: the share done, its
bar, and as `Facts` what is done, the time spent against the estimate, what is
left, for a sub-objective the parent it is counted in too
(`GoalProgress`, `@gui-objective-the-form-says-where-it-is-counted`), and the
days to its deadline, `2026-11-19, in 10 days`
(`@gui-objective-the-form-says-the-days-to-the-deadline`). The box stops
above the form's foot: in a window too short for every fact, the last ones
are left out whole, never written over the rule (`Facts.Within`). The work stood in the column beside, behind a
tab, and the room under the description was empty.

**Two columns and not one above the other.** The fields take the left five
eighths — in the narrowest window this board draws in, room for two short
fields to a line — and the column beside them has the panel's full height at
every size. Stacked, the trace got what twelve framed fields left, which is
nothing. A task being *added* has the same two columns, its trace there and
empty from the start: the form saves itself, and a trace that appeared beside
it at the first beat moved every field out from under the person typing
(`@gui-a-new-task-shows-its-trace-from-the-start`). Only an objective being
added, which has no work under it yet, takes the whole width, cut to what it
holds.

The trace **scrolls**, by the wheel over it or by pulling the bar beside
it. It is counted in lines and not in acts, because an act wraps onto as many
lines as it needs and a trace that jumped an act at a time would skip past
what it wrapped. Acts are told **the last done first**, one straight after
another: each begins with when it was done, which is what says where it
starts. They used to be told oldest first with a blank line between two of
them, which put what was just done at the bottom of a box opened at its top,
and read as an empty line between every line of a trace whose acts each fit
on one.

The bar is drawn whether or not there is anything to scroll. A bar that comes
and goes is a bar somebody reaches for and does not find. Its grip says two
things at once — how tall it is, how much of the trace is showing; where it
sits, which part — and it is never drawn smaller than can be taken hold of.

**The description takes the wheel wherever the keyboard is.** It is what a
task is opened to read, and a description reached only once the keyboard had
been put on it was one somebody reading could not scroll.

**Where what was written about it elsewhere lives is said at its foot**, beside
the offers: `context  ~/.cairn/notes/14`, faint, and held and copied like any
other text. Not among the fields: the place is not something the task carries.
See [context.md](context.md).

**Every piece of text beside the form can be held and copied**, the way what is
in its fields can — the trace's head, the trace, and the place of the notes.
A drag across it holds a run, a double click holds a word, a triple click a
line, `⌘A` holds the whole of the text last pressed in, and `⌘C` copies. The
trace is one text, its lines joined by a line break, so a run dragged over
several lines holds every one of them and copies them as they read
(`@gui-a-run-of-the-trace-spans-its-lines`); each line used to be a text of
its own, and a run could not leave the line it began on. `Readables` lists
every piece of it, once, laid out the way it is drawn, and what a pointer lands
on and what is painted held are both worked out from it.

A piece of text answers the pointer inside the box it is drawn in, and nowhere
else. It used to answer over every line from the first in sight, drawn or not,
so a long description reached down to the foot of the panel: the place of the
notes could not be held, and the offers took a press as the start of a run
rather than as a click.

**A task's steps are the last thing its form holds**, as a list of
tiles (`@gui-the-steps-read-as-a-list-to-tick`). A step carries a box before its title — empty, or filled and ticked —
and what is left is written in the reading ink while what is done is faint and
struck through; what is left comes first, what is done after
(`@gui-the-steps-left-come-first`). A box is read by its shape, without
comparing two greys. A tile shows the title alone: a step's place is no task's
number, and writing it there would read as one.

With the keyboard on the list, the arrows walk the tiles
(`@gui-the-arrows-walk-the-steps`), space ticks the one it is on, and enter
renames it in its tile (`@gui-a-step-is-renamed-in-its-tile`), as a double
click does; in its tile, backspace takes a letter and nothing else
(`@gui-a-step-in-edit-loses-a-letter-not-itself`). A click selects a step, and
backspace or the forward delete on it asks "Delete the step “Draft”?", over
the form, which stays drawn behind the question as it does behind "Throw away
what you typed?" (`@gui-the-form-stays-behind-the-question-that-deletes-a-step`): agreed,
it goes at once, through a `step delete` of its own, and the steps behind it
move up a place; cancelled, nothing moves
(`@gui-a-step-is-deleted-once-agreed`,
`@gui-a-step-is-kept-when-its-delete-is-cancelled`). The rest of the form,
saved or not, is left as it was
(`@gui-a-step-deleted-leaves-the-rest-of-the-form-as-it-was`), and a step the
task never had is taken out of the list alone
(`@gui-a-step-never-saved-is-taken-out-of-the-list`). Backspace used to strike
a step out until the form was saved, which read as a delete that did not
happen. Several steps are taken together the way cards
are, and **what is taken is not what the keyboard is on**: a step taken is
filled (`Surface.Ghost`), the one the keyboard is on is ringed, and a step can
be either or both. A click takes that step alone and puts the keyboard on it;
`⌘`-click adds one or lets it go, down to none, the keyboard following the
click (`@gui-a-step-taken-is-let-go-alone`, `@gui-the-last-step-taken-can-be-let-go`);
`⇧`-click takes every step from the one the keyboard is on to the one clicked
(`@gui-steps-stretched-to-are-taken-together`); `⌘A` takes them all
(`@gui-every-step-is-taken-at-once`); the arrows move the keyboard and take
the step they reach, so a keyboard alone deletes where it is. A delete key
acts on what is taken and on nothing else: one step asks "Delete the step
“Draft”?", several "Delete 3 steps?", and agreed they go in one act that
writes the task's steps without them
(`@gui-steps-taken-together-are-deleted-at-once`); space ticks every step
taken, or unticks them all when every one is done already, so a keystroke has
one outcome whatever mix it meets (`@gui-space-ticks-every-step-taken`); it
used to tick the one the keyboard was on and leave the others as they were;
dragged, steps taken go
together, in their order, where the one in hand lands
(`@gui-steps-taken-together-move-together`). One outline used to say both
things, so the last step let go stayed outlined and a delete key still took
it. A tile is dragged to another
place the way a card is: the step in hand follows the pointer, and a line says
where it would land — over the tile it takes the place of when carried up,
under it when carried down, drawn over the step in hand so the thing aimed
never hides the aim (`@gui-a-step-in-hand-shows-where-it-lands`,
`@gui-a-step-is-dragged-to-its-place`). A click on its box ticks it. The last
line, "+ new step — a line of its own", adds one: enter, a title, enter, and
the next is already waiting (`@gui-steps-are-added-one-after-another`). A list
longer than its frame is scrolled the way a column is: by the wheel over it,
wherever the keyboard is, with a bar beside it saying how much is in sight
(`@gui-the-wheel-scrolls-the-steps`); the arrows move it only as far as keeps
the line they reach in sight.

**A tick is done at once**, with a `step edit` of its own, and asks nothing of
the form (`@gui-a-tick-is-kept-whatever-becomes-of-the-form`): a box ticked on
a step says it is done, and a step done only once its task's form was saved
was a box that lied until then. A tick the engine refuses leaves the box as it
was. A step deleted goes at once the same way. Everything else — a title, a step added, the order — waits
for the form to be saved and goes with a form thrown away. Saving writes the
steps with the task, in the one act that writes it, in the order the list
says (`@gui-steps-done-and-left-in-turn-are-saved`).

The list writes the field's own lines underneath — `#1 [x] Relire l'ADR`, a
new one without a number — and those lines are what is saved, the way they
were when the list was a text to edit; the `#n` is the step's place, read as
nothing more than a line. A form being added carries the same list, so a task
goes in with the steps already written down for it.

The heading says what was opened: `task 7 · Ship it`, the title as it was
saved and not as it is being typed, which the field below says.

**An open task follows what is done to it elsewhere** (`freshened`). Each new
reading of the store — the beat's, when somebody else wrote — rebuilds the
form from the task as it now stands, and reads its trace again: a step an agent
added appears while the task is open (`@gui-an-open-task-shows-a-step-added-elsewhere`).
A field the person has changed and not saved keeps what they typed; every other
field takes what the store says, and what it was opened with moves along, so
the form counts as changed only what the person changed
(`@gui-an-open-task-keeps-what-was-typed-when-it-changes-elsewhere`). It used
to be read once, when it opened, and had to be closed and opened again to see
a change.

**An objective opens on the work that serves it**, beside its form, headed
`WHAT SERVES IT`: an objective has no trace of its own — the trace is kept
of tasks — and what serves it is the question it is opened to answer. Each is
a tile of one line, scrolled by the wheel over the column or by the bar beside
it. An objective nothing serves yet says so in as many words — an empty frame
reads as a box that failed to load rather than as a question nobody has
answered yet.

It once opened on a trace, and on the wrong one: the panel asked for the
trace of the chosen number without asking which kind of thing carried it, so
opening objective 1 showed whatever had happened to *task* 1 — a stranger's
trace, presented as the objective's. Measured, not reasoned about.

**An objective's card says how much it carries and how much is done** —
`7/12 · 58%`, on the number line it otherwise leaves half empty. The share is
rounded down, so an objective one task short of finished does not read 100%.
An objective nothing serves carries no count at all: `0/0` and a share of
nothing are two ways of dividing by zero, and `0%` on an objective nobody has
started reads as a failure rather than as one waiting for its first task. When
time was spent on the tasks serving it, the count says it too, `7/12 · 58% ·
71h40m` — the sum `objective show` gives (`@gui-an-objective-card-says-the-time-spent`).

Editing in place on the card was the alternative, and is what a mouse invites.
It was turned down for the trace: with no single moment of saving, every field
touched is its own `task edit`, and twelve fields touched is twelve lines of
trace for one sitting.

## Series

```
 Tasks 38   Objectives 5   Time 3   Series 4
┌ Active                       3 ┐ ┌ Team stand-up ↻3                      active ┐
│ ↻3 Team stand-up        active │ │ every weekday at 10:30, on the calendar, …    │
│    every weekday 10:30   today │ │ Next: today · Tue 10 Nov · Wed 11 Nov         │
│ ↻1 Weekly review        active │ │ ┌Template──┐ ┌So far──────┐ ┌On the board now┐ │
│ Stopped                      1 │ │ └──────────┘ └────────────┘ └────────────────┘ │
│ ↻2 Monthly report      stopped │ │ Occurrences · 52     archived ones included   │
│    every 4 weeks … since 12 Oct │ │ today  #312  Team stand-up   todo             │
└────────────────────────────────┘ │ ─────────────────────────────────────────────  │
                                   │ 52 occurrences  Edit e  Skip s  Stop x  Delete ⌫│
                                   └────────────────────────────────────────────────┘
```

A series ([series.md](series.md)) is the window's fourth tab, after the time,
as the terminal has it (`@gui-series-the-tab-lists-them`) — `⇥` keeps the
order it had. It is there only while `[recurrence] on` is true; a window not
told otherwise has it off, as the configuration does, so `New` opens with
recurrence off and `Paged` gives it the store's. **The tab is a split, the
objectives' own** (`SplitOf`): the list on the left, the active series under
one head and the stopped under another, each an item of two lines — its
number `↻3`, its title and its state on a pill, then its rule as a row writes
it — `every weekday · 10:30`, `every 4 weeks · monday` (`engine.RuleListed`)
— and its next day, `today 10:30` on its day (`engine.NextDaySaid`), or the
day it was stopped; on the right the series chosen — the first
while none is (`SeriesShown`). The split keeps the clicks that land on it and
no others: the tabs above it and the bar under it answer as on every other tab
(`@gui-series-every-tab-is-reached-by-a-click-from-every-other`), where the
page once kept every click but the bar's, and no tab could be reached from it. Its heading, its state, its rule read back as a
sentence with its next days — or the days resuming it today would make — then
three boxes: the template the next one is made from, what it did so far, the
missed in the late ink, and the task open now; then every occurrence, archived
ones included, a line each, a missed day in the late ink with no task, the
open one *on the board*; what it did so far says how long each took on
average, `14m each` (`engine.EachSaid`); the head of the occurrences says
*all archived, all kept* once none is left on the board. A click
on an occurrence opens its task's form, an archived one too, which nothing
else on the board reaches (`@gui-series-the-tab-reaches-an-archived-occurrence`).
Where the detail is too narrow for three boxes — at 720 — the template and
what it did so far are a line each; where its offers do not fit, *Edit the
series* and *Skip the next* say *Edit* and *Skip*, and the buttons give up the
least room they can.

**The tab narrows to a state** by the pill at the right of the tabs,
`state: all`, `active` or `stopped`, each click walking to the next and its
cross listing them all again, the tab then counting `1/2`
(`engine.InState`, `@gui-series-the-tab-narrows-to-a-state`); *+ Filter* on
this tab walks the same pill. The foot says, beside the offers, the series'
number after the loop, the day it was made and how many occurrences it made
— `↻3 · made 2 Sep · 52 occurrences` — or who stopped it and when, `stopped
by me · 12 Oct` (`@gui-series-the-tab-says-it-as-drawn`).

**Its foot offers what the command line does**: *Edit the series* (`e`),
*Skip the next* (`s`) while one is open, *Stop* (`x`) or *Resume* (`r`), and
*Delete…* (`⌫`), the keys working from the tab as the buttons do
(`@gui-series-the-tab-skips-stops-and-resumes`); `↑↓` walk the list, `⏎` opens
the occurrence open. The bar offers *Open the occurrence* (`⏎`), *Edit*
(`e`), *Skip* (`s`), *Stop* (`x`) or *Resume* (`r`), *New series* (`n`), the
search — which narrows the list by title, rule or number, as `series list
--search` does — and the help. `n` opens a task's form in todo with the keyboard already on
its `repeats` field (`@gui-series-n-makes-a-new-one`); *Edit the series* opens
the series' own form — title, priority, labels, estimate, objectives, steps,
its rule and the day it counts from, with no trace beside it — and saving it
changes what comes next, never one made (`@gui-series-the-form-changes-the-series`).

**Deleting a series asks what becomes of its tasks**, three rows checked one
at a time, the first — keep every task as an ordinary task — chosen
(`OnSeriesFates`, `@gui-series-deleting-keeps-the-tasks-unless-told`):
keep them, delete the one open and keep the rest
(`@gui-series-deleting-the-open-one`), or delete them all, which asks once
more, naming how many (`@gui-series-deleting-every-task-asks-once-more`). It
is the question deleting an objective asks, row by row, for the same reason:
what is lost is said before it goes.

**A card that comes back is told at a glance**, marker A of the study: the
loop before its title, its title on one line and its rule on the line under
it in the faint ink, and on its row the series' number after its own, in the
accent (`@gui-series-a-card-carries-the-loop-and-the-rule`). A click on that
number opens the series in its tab (`@gui-series-the-chip-opens-its-series`).
At the right of the row is the day it was made for, as the engine says it
(`OccurrenceDay`): the time on its day, late once passed, its weekday within
a week, counted down once its day is gone — its deadline left unsaid, since
the day is the same. No face the window carries has `↻`, so the loop is drawn
(`LoopMark`), as the standing marks are.

**On a card chosen, `r` is a prefix**, as in the terminal: `r s` skips it,
`r x` stops its series, `r r` resumes it and `r e` opens its series
(`@gui-series-r-s-skips-the-chosen-occurrence`,
`@gui-series-r-x-stops-and-r-r-resumes`, `@gui-series-r-e-opens-its-series`);
the bar offers *Open its series*, *Skip it* and *Stop it* — or *Resume it* —
then (`@gui-series-the-bar-offers-the-r-keys-as-drawn`): not *skip*, *stop*
and *resume* alone, which the bar already answers with `s`, `x` and `r` on the
series' tab, and the help says the keys
while recurrence is on (`@gui-series-the-help-says-the-r-keys`).

**The form of an occurrence says its series.** Its heading is the loop, its
title and its number, then, faint at the right, the series it is one of and
the day it is for — `one of the series every weekday at 10:30 · this one is
Mon 9 Nov` (`OccurrenceHeading`,
`@gui-series-the-form-of-an-occurrence-heads-with-its-loop`). Under its
details a box headed *Series*, with *Edit the series…* (`r e`) at its right,
which opens the series in its tab, says the rule read back, then what it is
anchored on, the next day and its time, the day it counts from and when it
ends, what it did so far, and the days it missed, in the late ink, "nobody
was there; not made up"
(`@gui-series-the-form-of-an-occurrence-names-the-days-missed`); its foot offers *Skip this one* and
*Stop the series* — or *Resume the series* — beside *Save*
(`@gui-series-the-form-of-an-occurrence-says-its-series`). A task that is no
occurrence carries a `repeats` field, in a section of its own: a rule typed
there makes it the first occurrence of a new series
(`@gui-series-a-task-given-a-rule-becomes-the-first`). The field is typed into, and
**reads the rule back under it as it is typed**: how often, what it counts
from and when it ends, after the loop, then the next three days it comes on —
or, counted after done, its first and how long after — or why it is no rule,
in the wrong ink (`engine.RuleReadBack`,
`@gui-series-the-repeats-field-reads-the-rule-back`). A misread rule shows
before anything is saved. **A click on the field opens the picker** the
study drew, hung under what it reads back (`OnRule`, `RulePicked`): every so
many — a stepper, the chevrons walking 1 to 99 — and days, weeks or months;
the seven days of the week, each ticked or not, which counts in weeks; the
time, any or a half hour, a time typed kept among them; the end — never,
after so many times, ten at first, or on a day, four weeks off at first, and
how many or which day beside it; and whether the next comes on the calendar
or after the last is done. Each part chosen writes the field at once, in
the words it reads (`domain.RuleParts`), so the read-back follows
(`@gui-series-the-repeats-picker-builds-the-rule-by-its-parts`,
`@gui-series-the-repeats-picker-ends-and-counts-after-done`,
`@gui-series-the-repeats-picker-writes-the-words`). `↑↓` walk its lines and
`←→` change the one the keyboard is on, ringed, its caption in the accent;
on the days `←→` walk them and `space` ticks one
(`@gui-series-the-repeats-picker-walked-by-the-keyboard`). *Choose* (`⏎`)
keeps the words; *Cancel* (`esc`) or a click off it puts back what the field
read when it opened (`@gui-series-the-repeats-picker-cancelled-leaves-the-words`).
A day of the week counted after done is a rule cairn refuses, so ticking a
day counts on the calendar, and choosing after done lets the days go; an end
is a number of times or a day, never both, and the picker offers one at a
time. The picker sets **the day a series counts from** too, a stepper a day
at a time from today, written in the form's `from` field, beside `repeats`
in the *Series* section and chosen on a calendar as a deadline is
(`engine.FromField`, `@gui-series-the-picker-sets-the-day-it-counts-from`).
The picker hangs under the field where the window has room — at 1200 by 800
it does, standing past the form's foot — and over it where it has not, as at
720 by 560. **Saving a change to what
the next one is made from asks once** — title, priority, labels, estimate,
steps or objectives (`engine.ChangesTheSeries`): *This one only*, `⏎` and the
default, or *The whole series*, `s`
(`@gui-series-saving-an-occurrence-asks-this-one-or-all`,
`@gui-series-saving-this-one-only-leaves-the-series`).

**With recurrence off, nothing of a series shows**: no tab, no loop, no rule,
no series number, no `repeats` field, no `r` key and no offer
(`@gui-series-off-shows-nothing`) — the engine's `Pages.Recurring` takes them
off the snapshot the window reads, the one the terminal reads too. The one
place left is the command bar: a series asked for there is answered by a
banner saying recurrence is off and how to turn it on, the engine's own
refusal (`@gui-series-off-says-why-in-the-command-bar`), rather than by
nothing, which reads as a command that does not exist.

**An occurrence's history says where it came from**: `made from series ↻1`,
the act the engine writes on a task its series made, beside its first
attributes (`@gui-series-the-history-of-an-occurrence-says-its-series`), and
then, under *in the series, before this one*, what its series did before it
made it — made, skipped, missed — read from the series' own acts, which stay
in their own table: only the history joins them
(`engine.BeforeIn`, `@gui-series-the-history-of-an-occurrence-goes-on-with-its-series`).

**The board narrows to what comes back** by a filter of its own, `repeats`,
offered while there is a series, as `task list --repeats` narrows a listing
(`@gui-series-the-board-narrows-to-what-comes-back`).

## Several at once

**`⌘`-click gathers cards, `⇧`-click takes every card between** — `⌘` on
macOS, `Ctrl` on Linux: the shortcut key either way. `⌘` and a
click add a card to what is taken, or take it back out; `⇧` and a click take
every card from the one chosen to the one clicked, in the column, the lane or
the list of objectives they share. A plain click, `esc`, or another tab starts
again. What is taken is drawn the way what is chosen is. The keyboard gathers
nothing: `⇧` and the arrows already move a card, and the pointer is the hand
that picks several things out of a board.

**What is taken is dragged together.** Dragging one of the cards taken
brings the others along: they leave their places with it, and land behind it,
in the order they were taken, in one act — `task move` with the others under
`also` (`@gui-cards-taken-together-are-dragged-together`,
`@ordering-several-moved-together`). What is taken is let go of once it lands.

**What is taken goes in one act**, deleted or archived, all of it or none: the
engine is asked once, the first by its number and the rest under `also`, and
an act is one transaction. Three deletions asked one after another left, when
the third was refused, two gone and somebody to work out which.

**It is asked about first, and the question names what would cost more** —
the steps of a task, a task waiting on it, the task in focus — and **what
cannot go at all** — an objective work still serves, a task already off the
board. Read off what the board holds, the facts the engine goes by, so the
question says before what the engine would only refuse after. It lists what
is taken, a line each, `#1  Ship it — its 1 step goes with it`, six at most
and then how many more (`severalSays`); it wrote only what was flagged, its
number on a chip read as a key. The answer is one of three: the verb,
*Delete* or *Archive* (`⏎`), leaves out only what cannot go — it was *All
possible*, which named no act; *Skip flagged* (`s`) leaves out everything
named; *Cancel* (`esc`) takes nothing.
*Skip flagged* is offered only when something named could go all the same —
otherwise it is the first answer again.

## The board arrives by an intro

When the window opens, its board arrives by one of eleven intros, which
`window.intro` names ([config.md](config.md#how-the-board-arrives)); ten
were proposed and all ten kept (#255), and the stones that came before
them stay as the eleventh. Each ends on the exact still board, a key or a
press ends it at once and still does what it does, it starts on the first
frame drawn (`Born`), and a system asking for less motion gets none
(`@gui-less-motion-no-intro`) — macOS's Reduce motion, GNOME's animations
turned off, Windows' client-area animations off (`motion_*.go`).

| name | called | what it does |
|---|---|---|
| `ripple` | Ripple from the focus — the default | the card in focus pops on a snappy spring, its dot lights, a ring in the accent spreads from it, and every other card grows outward from it as the ring reaches it, nearest first |
| `deck` | Dealt from the deck | the cards wait as a deck above the bar and are dealt round by round, a card to each column, each flying on an arc with a deeper shadow and its rotation unwinding |
| `spring` | Spring cascade | cards rise on a soft spring in a diagonal wave from the top left |
| `accordion` | Accordion columns | each column unfolds from its left edge, dark while folded, and its cards slide in from the fold |
| `sweep` | Light sweep | a tilted band of light crosses the window; each card it passes is revealed over-exposed, then cools to its colours |
| `flap` | Split-flap board | each card drops on its top hinge, overshoots and settles, row by row, as a station's board |
| `pull` | Focus pull | the board starts soft and a little too close; the card in focus sharpens first, then the rest by distance |
| `particles` | Particles assemble | a dust in the cards' own colours streams in and condenses into each card |
| `dolly` | Camera dolly | the camera opens close on the card in focus and cranes back to the whole board |
| `blueprint` | Blueprint, then ink | outlines first, then the fills, the stripes drawn down and each title typed out |
| `stones` | Falling stones | each card built of dry stones falling into place |

**Every frame is a function of the time since the first one**
(`Staging`, `intro.go`): it takes the intro, the time, and the board as it
stands (`Stage` — its grounds, its cards in the order they are drawn, the
card in focus or the first in progress to start from), and says how far each
thing has come (`Pose`: seen, moved, grown, stretched, turned about a pivot,
lit, shaded, lifted, blurred, drafted) and what is laid over the board
(`Wave`, `Sweep`, `Grains`, `Pulse` — the focus dot's beat). So a test draws
any instant of any intro, and its last one is the still board, pixel for
pixel. A busier board plays faster rather than longer: an intro lasts as long
as its last card takes to land, never more than a fifth over the length the
page proposed it at (`IntroLasts`), and always under two seconds. A board
whose staggers would run past that has its whole timeline played faster to
fit: every delay and every motion shortened by the same share, so the order
and the curves stay and the cards come quicker one after another. The `Kit` lays a pose over what it draws
(`posed`), and the marks are components (`component_intro.go`), so the two
guards hold.

**Gio is not a browser, so some of the page is faked.** A blur is the card
drawn four times a little apart, faint, under itself; the light sweep's screen
blend is a band of the accent's lit ink at low alpha; a split flap is a scale
on the card's height about its top, with no perspective; the blueprint types
each title but does not flash each new letter in the accent. The terminal
board has no intro: its cells redraw whole, and none was trivial.

**The settings offer it as a choice played on the board behind.** Its row
is a picker; it opens a narrow chooser against the right of the window, the
board left undimmed beside it. Walking the list, or clicking an intro, plays
it on the board; `r` or Replay plays it again; `⏎`, Keep, or a second click
writes it; `esc` goes back with nothing written
(`@gui-settings-the-intro-is-played-then-kept`).

## Settings changed in place

**The configuration stays one, in one place; the window is one way in to
it.** `⌘,` — Settings on a Mac — or Configuration… in the File menu opens
the settings (`configuration.go`, `draw_settings.go`), drawn as the mockup of
#260 draws them, at both pinned sizes. Every key a person sets is changed
there, in the workspace's `settings.toml` or in the global `config.toml`, and
the window writes nothing itself: it sends the key, the value and the scope to
the engine (`engine.ConfigSet`, `engine.ConfigUnset`), which checks, edits one
line of the file and keeps every comment ([config.md](config.md#one-configuration-several-ways-in-to-it)).
It reverses no decision: the panel was read-only because a form that encodes
the TOML again loses the comments a person wrote, and nothing here encodes it
again; the file is still opened from the foot, *Open settings.toml* or *Open
config.toml*, in the system's editor (`open -t`, `xdg-open`, `cmd /c start`).

**The scope is chosen at the top.** A segmented choice, the workspace's name
with its dot or *Global*, picks the file written, and the path read is said
beside it: `settings.toml over config.toml`, or `config.toml · under every
workspace's own settings`. It opens on the workspace's when the window shows
a workspace, or a store with settings beside it; on Global for a loose store
(`@gui-settings-a-toggle-is-written-at-once`,
`@gui-settings-a-choice-is-written-in-the-global-configuration`).

**Sections at the left, a row a key.** Appearance, Pages, Time, Archive,
Agents and Places list down a column of their own at 1200, each with its
count, and the one in sight lit; a click scrolls to it. At 720 the panel takes
the room the window has, the sections become heads down one scrolling column,
and where a value comes from goes under its line. A row is the key's name, the
key in Go Mono, its control, and at the right where its value comes from — the
catalogue (`config.Keys`) says every one of these, so no key is listed twice:

- a **switch** (`Toggle`) for a key that is true or false, written when it is
  clicked; a **choice of a few** (`Segmented`) for words, written when one is
  clicked; a **field** for a duration, a day, a number, a label or a folder,
  written on `⏎` — `esc` gives it up — a value the reader refuses said under
  it in `state.wrong`, in the reader's own words, and nothing written
  (`@gui-settings-a-field-is-written-on-enter`); a **picker** for the theme,
  which opens the chooser, and for the colour a workspace wears;
- **where it comes from**, in the workspace's scope: `set here`, with what the
  global configuration would give, `← global: on`, and *Reset to global*,
  which takes the line out (`@gui-settings-reset-takes-the-line-out`); `from
  global`; or `default`. Under the pointer a row that is not set here offers
  *Set here*, which pins today's value
  (`@gui-settings-set-here-pins-the-global-value`). In Global the marks read
  one level down — `in config.toml` and *Reset to default*, or `default` —
  and a key the workspace overrides says so, `· home sets its own: off`;
- the **places** belong to the machine and live in Global only: the store and
  the notes are shown with their copy mark and never written from the window,
  `read when cairn opens`, since the window keeps the store it opened on
  (`@gui-settings-places-are-read-when-cairn-opens`); the plugin's folder is a
  field. The colour a workspace wears is never in Global.

Nothing waits for a Save button, and the foot says it: *Saved as it is changed
· the file is read again when it changes*. The keyboard walks the rows with
`↑` `↓`, acts on one with `⏎` or space, cycles a choice with `←` `→`, takes a
key out with `⌫`, and switches the scope with `tab`. With no global file, the
foot offers *Write the whole file*, which is `config init`
(`@gui-a-missing-configuration-is-written-from-the-window`).

**The theme is chosen with a live preview.** The Theme row opens the chooser in
place of the rows, *Theme of home*: the nine shipped schemes, then a person's
own from `~/.config/cairn/themes/`, each a tile showing a small board in its
own colours (`SchemeTile`). The keyboard or a click on a tile re-skins the
whole window at once — board, panel and chooser — since every component reads
`k.skin` and nothing else; `⏎`, a second click or *Wear* wears it and writes
`window.theme` in the scope chosen; `esc` goes back to the scheme in force
(`@gui-settings-the-theme-is-previewed-then-worn`). A scheme of one's own that
falls short of a contrast is offered all the same, marked `falls short ×n`,
the first shortfall said at the foot; one that cannot be read is shown with
why, and not offered.

**Read again on the beat the window already has.** Every two seconds, beside
the store's `PRAGMA data_version`, the window stats the global file, the
store's settings and the scheme of one's own worn (`engine.StampOf`); a size
or a time that moved is checked against the content's hash, so a save with
nothing changed does nothing. When one moved, it reads `config show` again
through the engine, applies all of it — the pages and the tabs, the time, the
weekends, the series, the theme, the workspace's colour — and says above the
bar, for a few beats, which keys changed and from what: `settings.toml read
again — pages.time hidden → off · window.theme default → nord`, with
*Settings* to open the panel (`engine.ConfigChanges`,
`@gui-settings-an-edit-by-hand-is-read-again`). No undo: undoing a hand edit
would write the person's file behind them. A write the window made itself
says nothing — it stamps the files again as it writes. A file watcher was
turned down: a new dependency, and editors save by writing a new file and
renaming it over the old one. Reading on focus was turned down for the
store already: a board on a second screen has to be true.

**A file that cannot be read keeps the last good configuration.** The window
does not fall back to defaults, which would turn pages on and change the
theme under the person's eyes. A banner stays above the bar until the file
reads — the file, the line, the reader's own words, and when the configuration
in force was read (`engine.TroubleIn`) — with *Open settings.toml*; in the
panel, that file's rows are faded and not offered, and the other scope stays
editable (`@gui-settings-a-file-that-cannot-be-read-keeps-the-last-good`). At
open the window behaves as before: a file that cannot be read stops it, naming
the file. The command line and the MCP server read the file at every use and
refuse a file they cannot read, as they always did.

**What applies at once, and what does not.** At once: the pages, the time,
recurrence, the weekends, the theme, the workspace's colour, the approach
label, the archive and the focus session — the last three the engine reads at
every use anyway. Not: the store, the notes and the workspaces, which the
window keeps as it opened, as it keeps its name after `workspace use`. Not
yet: the native menu bar is built when the window opens, so a page turned off
keeps its menu lines until the window opens again, and keeping the sizes
takes effect at the next window, which is the one that reads them.

## Workspaces, a window each

**The window says which workspace it shows, always.** Its title is `cairn —
work`, so the Dock, `⌘\``, Mission Control and a window manager tell two
windows apart; a store no workspace names is titled by its folder
(`Window.Title`, from `engine.WorkspaceShown.Name`, the rule the terminal's
header follows too). The top bar opens with a pill before the tabs: a dot in
the workspace's colour, its name and a caret; a loose store shows a ring and
its folder in Go Mono, cut at its start (`PlaceShown`)
(`@gui-workspace-the-title-and-the-pill-name-it`,
`@gui-workspace-a-loose-store-shows-its-folder`). The place is read once, when
the window first looks at its store (`engine.WorkspaceShow`): a window keeps
its store, so `workspace use` afterwards does not rename it. The pill is as
wide as its words, at most 128 points; at the narrowest window it gives way
to the tabs and the filters' pills and keeps only its dot and caret, the
title still naming it. Only a window that knows its place draws a pill, so
`LayoutFor` and the frames pinned from it are as they were; `Window.Layout`
moves the tabs right of the pill.

**The switcher drops under the pill.** A click on the pill, `⌘⇧O`, or *Open
a workspace…* in the File menu (`@gui-workspace-the-menu-offers-the-switcher`)
opens it: a line to type in, which narrows it
(`@gui-workspace-the-switcher-narrows`); Workspaces, in the order `workspace
list` gives them, each with its folder and what choosing it does — `this
window`, `open · brings it forward`, `opens a new window`, `not found`, and
`default` — and `⌘⇧1`…`⌘⇧9` by its place in that list; Recent folders, each
with the day it was opened or `not found now`; and at its foot *Open a
folder…*, *New workspace…* and *Settings of work…* (`⌘,`)
(`@gui-workspace-the-switcher-lists-them`,
`@gui-workspace-the-switcher-opens-the-settings`). `⌘O` was taken by Import,
so the switcher is `⌘⇧O`. The places are the list's, the same in every
window, where the terminal skips the one shown: there the digits only work
inside its list; here `⌘⇧2` works from the board, and a key that meant
another workspace in each window could not be learnt. A US Mac keyboard
names `⌘⇧2` by the `@` shift makes, so the shifted figures are read back as
their figure (`keyOf`). *Open a folder…* and *New workspace…* type into the
same line, through the engine's `Switch` rules and `workspace add`; a folder
that is not there is refused with the engine's reason and the line kept
(`@gui-workspace-a-folder-opens-in-a-window-of-its-own`,
`@gui-workspace-a-new-one-is-made-and-opened`). With the keyboard on a recent
folder that is still there, the foot offers *Keep as a workspace…* too: it
asks for a name, gives it and the folder to `workspace add` — the path the
terminal's `^K` and the command line take — and lists the switcher again,
the folder now among the workspaces and gone from the recent ones; nothing
is opened (`@gui-workspace-a-recent-folder-is-kept-as-a-workspace`). It is a
line of the foot rather than a key: `⌘K` is the command palette.

**Choosing opens it in a window of its own; this one stays.** A terminal has
one screen and switches in place; a desktop has windows, and two boards side
by side is the point. One window per store: the window asks for another
(`ApartAsked`, as it asks for a file dialog), the engine says what that would
hold and how to name it (`engine.OpenApart`, refusing what `Switch` refuses),
and the loop either brings forward the window already showing that store, or
starts `cairn gui --workspace home` — `--store <path>` for a folder — as a
process of its own (`OpeningApart`), and says which it did
(`@gui-workspace-another-opens-in-a-window-of-its-own`,
`@gui-workspace-the-one-open-is-brought-forward`).

**How a window knows another shows a store.** Each window, as it opens,
listens on a Unix socket named after its store —
`cairn-<16 hex of the store's path>.window` in `$XDG_RUNTIME_DIR` when it is
absolute, the system's temporary folder otherwise (`Windows`, `windows.go`).
Another cairn asking for that store connects, writes `raise`, and the window
answers `ok` and asks the system to bring it forward (Gio's
`system.ActionRaise`). The same happens when `cairn gui` is run on a store a
window already shows: it brings that window forward, says so, and opens
nothing (`@gui-workspace-a-second-window-on-a-store-raises-the-first`). A
socket left by a window that died answers nothing and is claimed over. A
socket was chosen over a lock or a pid file: it says a window is there and
reaches it in one act, and a pid outlives its process. Its limits, said
plainly: on Wayland a window cannot bring itself forward, so it is not; on
Windows the system may only flash the window's button; two windows opened
on one store in the same instant can both claim it; a socket cannot be made
where the folder refuses one, and the window then opens unclaimed. The
scenarios cannot see a second window of the system, so they observe what was
asked: the arguments a cairn would be started with, and the window that took
`raise`.

**The settings of a workspace say where each key comes from.** `⌘,`, on a
window showing a workspace, heads the configuration panel *Settings of
home*, then the `settings.toml` read and `over` the global file. Each key
says `set here`, `from the global configuration` or `default` — the
engine's `written_here` and `written` — and the time is drawn as a choice of
three, `on`, `hidden`, `off`, the one in force filled
(`@gui-workspace-its-settings-say-where-each-comes-from`). The approach label
and the colour the workspace wears have a row each, the colour after a swatch
of it (`@gui-workspace-its-settings-say-the-label-and-the-colour`). Beside a
key set here the panel says what the global configuration would give,
`← global: on` — `config show`'s `global` — so the way back is read before
the line is taken out of the file
(`@gui-workspace-a-key-set-here-says-the-global-value`). The study drew it
`↺`; none of the faces the window carries has that glyph, and a character
with no face is refused by a test, so it is the arrow they do have. Every key
is changed from there ([settings changed in place](#settings-changed-in-place)),
and *Open settings.toml* hands the file to the system's editor.

**A time hidden or off leaves the cards without time.** What a card cost
against its estimate is shown only while the time is
(`engine.Pages.Cost`); the cairn of a card in focus stays, since the focus
is kept under `off` (`@gui-workspace-time-off-leaves-the-card-without-time`).

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

**A drawing test that needs only a board draws the fixtures**
(`aFixturesBoard`, the store `cairn fixtures` makes), so every screen is
drawn from one believable board and a case nobody thought of is drawn
anyway; a test whose case needs a board of its own — an empty column, a
card in one lane, an objective nothing serves — still seeds it.

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

- **A task open beside the board for good**, which the extra room would allow
  and which would stop this being the same board.
- **`⌘Z` bringing a view deleted back.** The window has no undo for a
  deletion — not a task's, a label's or a backup's; `⌘Z` only recovers what a
  form threw away. A view's would be the first, and an undo for one kind of
  deletion is a rule nobody can guess; it waits for an undo the window has
  for every deletion.
