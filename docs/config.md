# Where cairn finds its own places

Read this before touching how the store or the notes are addressed, before
adding a key to the configuration, or before changing what happens when nothing
says where something is.

Cairn has two places to find before it can do anything: where the tasks are
kept, and where the notes taken elsewhere are looked for. Both are found the
same way, so there is one rule to know rather than two. The scenarios are in
`bdd/config.feature`; what the notes are for is in [context.md](context.md).

## The order, and why the environment is not part of it

```
1  a flag                 --workspace <name>, or --store <path>
2  the environment        CAIRN_STORE, CAIRN_NOTES, then CAIRN_WORKSPACE
3  the default workspace  [workspaces] default, when it is said
4  the file               ~/.config/cairn/config.toml: store, notes
5  the default            ~/.cairn/cairn.db   ~/.cairn/notes
                          the first rung that says something answers,
                          and every rung under it is passed over
```

Five rungs, read from the top, and the first one that names a store answers.
The three lower ones are the configuration: the default workspace and the
file's two keys are choices made and kept, and the default is what a choice
looks like when nobody made one — the same on every machine, so "where are my
tasks" never depends on which one you are sitting at.

The two upper ones are passing choices, and that is why the environment is
not one more rung of the configuration. Another store for one command, another
vault for one session: it does not take a turn between two kept choices, as a
fallback would, it replaces all of them at once — and it changes nothing on
disk, so the next command without it finds things exactly as they were. A
flag is the same, for one command, and it replaces the environment too.

Saying it as a chain where the environment waits its turn after the file was
tried and read wrong: it makes the environment look like a fallback, which is
the one thing it is not. So the diagram puts it above everything it replaces,
and `cairn workspace show` names, on its `why` line, the rung that answered and
every rung under it that named a store too and was passed over:
`CAIRN_STORE — over the default, home`, or `--workspace — over CAIRN_STORE
(work) and the default (home)`. When the configuration answers, the line says
which rungs above it said nothing — `the default (no --workspace, no
CAIRN_STORE)` — and a server says it as it was when it started
(`@workspace-show-why-names-the-rungs-passed-over`, `@workspace-show-why-as-json`).
A `store` the file still writes under a default workspace is named among the
rungs passed over, since nothing else would say it is no longer read.

**The default workspace is a choice kept, made over the file's two keys.**
Once `[workspaces]` names a default, the store and the notes are in its folder,
and `store` and `notes` are not read for the places
(`@workspace-the-default-answers`, `@workspace-its-notes-are-in-its-folder`).
Without a default, a table of workspaces changes nothing
(`@workspace-no-default-changes-nothing`), and a machine with no table
answers exactly as it did before workspaces existed. The environment still
replaces it, as it replaces the file (`@workspace-the-environment-replaces-the-default`).

**`CAIRN_WORKSPACE` names a workspace, under the places the environment
names.** `CAIRN_WORKSPACE=<name>` gives both places of that workspace, as
`--workspace` does for one command, and as the default does when nothing
else says (`@workspace-the-environment-names-a-workspace`). It is set where a
whole session is started from — a repository's `.claude/settings.json`, for
the plugin's MCP server and its hook ([claude-code.md](claude-code.md#a-repository-that-names-its-workspace)).
Inside the environment the variable naming a place wins over the one naming a
workspace, for that place, because it is the more precise of the two:
`CAIRN_STORE` replaces the workspace whole, and its store keeps its own notes
as below — never the notes of the workspace it replaced
(`@workspace-cairn-store-wins-over-cairn-workspace-for-the-store`);
`CAIRN_NOTES` replaces the notes alone, the store staying the workspace's
(`@workspace-cairn-notes-wins-over-cairn-workspace-for-the-notes`). A flag
replaces it like the rest of the environment
(`@workspace-the-flag-replaces-cairn-workspace`). A name no workspace has is
refused naming the variable, the name and the ones there are, and a folder gone
naming the folder, as `--workspace` is
(`@workspace-an-unknown-workspace-in-the-environment-is-refused`,
`@workspace-a-folder-gone-in-the-environment-is-refused`). `workspace show`
says `CAIRN_WORKSPACE — over the default, work`, and the variable among the
rungs passed over when a flag or `CAIRN_STORE` answered
(`@workspace-show-says-cairn-workspace`); the rungs said to be silent stay
`--workspace` and `CAIRN_STORE`, as `--store` is not listed there either. A
workspace it names is a workspace, never a recent folder
(`@workspace-a-workspace-the-environment-names-is-not-a-recent-folder`).

**A flag replaces even the environment.** `--workspace <name>` or `--store
<path>`, on any command, is the most passing choice of all: one command, typed
by whoever runs it, and it answers instead of everything below it, the
environment included (`@workspace-the-flag-picks-another`,
`@workspace-the-flag-replaces-the-environment`,
`@workspace-the-store-flag-names-a-store`). It is a person's own hand on this
one command, and a variable left in a shell would otherwise silently win over
what was just typed. The two flags name the same thing twice, so they are
not understood together (`@workspace-one-flag-or-the-other`).

**The notes never come from another workspace than the store.**
`--workspace` and the default workspace name both places at once, and
`CAIRN_NOTES` still replaces the notes alone. A store named by `--store` or
`CAIRN_STORE` keeps its own notes: `CAIRN_NOTES` when it is said; otherwise,
when that store is a workspace's `<folder>/cairn.db`, that workspace's notes;
otherwise the file's `notes` or the default, as before workspaces existed —
never the default workspace's (`@workspace-a-store-named-keeps-its-own-notes`).
Taking the default's notes beside a store from elsewhere would write one
board's notes into another's folder.

**This changes something that already worked.** Before, the store was the
environment or the default, with nothing in between. The file is a new rung, and
a machine with a file now answers differently than it did.

## Workspaces, each a folder holding a store and its notes

```toml
[workspaces]
default = "work"
work = "~/cairn/work"
home = "~/cairn/home"
```

A workspace is a name and a folder: its store is `<folder>/cairn.db`, its
notes `<folder>/notes`. One store per piece of life — work and home, a client
and another — each with its own board and its own notes, and switching is a
word rather than two variables. `default` names the one used when nothing else
says; every other key is a workspace. A folder is absolute or under `~`, like
every place, and one that is not is refused naming its key
(`@workspace-a-folder-that-is-not-absolute-is-refused`). It is not called
"context", which is already the word for a task's notes folder
([context.md](context.md)).

A workspace named that the file does not name is refused, naming it and the
ones there are (`@workspace-an-unknown-name-is-refused`), and so is a default
that names none (`@workspace-a-default-nobody-named-is-refused`). A workspace
whose folder is not there is refused naming the folder
(`@workspace-a-folder-that-is-not-there-is-refused`): a typo in a folder would
otherwise start a new, empty board somewhere nobody looks.

- `cairn workspace list` names each workspace and its folder, the default
  marked (`@workspace-list-names-each-and-the-default`, `--json` too:
  `@workspace-list-as-json`), one whose folder is gone marked `not found`
  (`@workspace-a-folder-gone-is-listed-not-found`), and the colour each
  wears; under them, the recent folders (see below).
- `cairn workspace add <name> <folder>` writes the workspace into the file
  and makes its folder, with an empty `notes` in it
  (`@workspace-add-writes-it-and-makes-its-folder`) and settings that say
  nothing yet (see below). The rest of the file is
  left byte for byte as it was — the entry joins the end of `[workspaces]`,
  or a table of its own after everything else
  (`@workspace-add-joins-the-table-there`) — and with no file it writes one
  holding that table alone, not the whole example
  (`@workspace-add-without-a-file-writes-only-the-table`). A name is lowercase
  letters, digits and dashes, and `default` is taken
  (`@workspace-add-refuses-a-name-that-is-not-plain`); a name already there is
  refused (`@workspace-add-refuses-a-name-known`), and so is a folder that is
  not absolute (`@workspace-add-refuses-a-folder-that-is-not-absolute`). The
  folder is written as it was typed, so `~` stays `~`.
- `cairn workspace use <name>` sets `default`, and says which one it was
  (`@workspace-use-sets-the-default`, `@workspace-use-refuses-an-unknown-name`).
  A workspace whose folder is gone is refused, naming the folder, and the file
  is left as it was (`@workspace-use-refuses-a-folder-gone`): made the
  default, it would have every later command refused in its turn.
- `cairn workspace remove <name>` takes the workspace out of the file — its
  line, and nothing else of the file — and touches no other file: its
  folder, its store and its notes stay where they are
  (`@workspace-remove-drops-only-the-entry`). A name the file does not give
  is refused (`@workspace-remove-refuses-an-unknown-name`), and so is the
  default, naming `cairn workspace use` to choose another first
  (`@workspace-remove-refuses-the-default`): removed, it would leave a default
  that names nobody, and every command refused for it. A folder moved by hand
  is named again with `cairn workspace add`.
- `cairn workspace show` says which workspace answers this command and why —
  the rung that named its store, and the rungs under it it passed over, as
  [the order](#the-order-and-why-the-environment-is-not-part-of-it) says —
  and the store and notes it uses (`@workspace-show-says-the-default-answers`,
  `@workspace-show-says-the-flag-answers`). In JSON, `why` carries the rung,
  `over` each rung passed over and what it named, `unsaid` the rungs above
  that said nothing, and `started` for a server's. A store that no workspace names,
  given by `--store`, the environment or the file, is said as such, with the
  `cairn workspace add` that would name it
  (`@workspace-show-says-no-workspace-names-the-store`). It says whether time
  is kept, `on`, `hidden` or `off`, and where that is said — set in the
  store's settings, from the global configuration, by default
  (`@workspace-show-says-the-time`), and which pages are on, said the same way
  (`@workspace-show-says-the-pages`).
- `cairn workspace list` marks a workspace whose time is not on — `time off`,
  `time hidden` — as its own settings, over the global configuration, say; a
  workspace whose time is on is not marked, nor is one whose settings cannot be
  read, which every command on it says already
  (`@workspace-list-marks-the-time`). The JSON carries it as `time`.
- `config show` lists `[workspaces]` too (`@workspace-config-show-lists-them`),
  and names a place a workspace gave as named by it.
- The terminal board names the workspace it shows at the head of its header,
  and `^O` opens another in its place without writing the file — see
  [tui.md](tui.md#another-workspace-in-place-of-this-one). The window names
  it in its title and its top bar, and opens another in a window of its
  own — see [gui.md](gui.md#workspaces-a-window-each).

`workspace list` and `workspace show` are tools of the MCP server too:
`workspace_show` says which workspace a server answers from — the one it
started on, which it keeps whatever the default becomes ([mcp.md](mcp.md)). `workspace add`,
`workspace use` and `workspace remove` are not (`@workspace-an-agent-cannot-add-or-switch`,
`@workspace-an-agent-cannot-remove-one`):
they change where every later command reads and writes, which is the
person's to decide, like `cairn upgrade` ([mcp.md](mcp.md)). They are
`engine.ForAPersonOnly`, which the command line offers and the server does not.

The four commands read the file themselves rather than through the store they
would find (`Needs: AConfiguration`), so a default whose folder is gone does not
stop `workspace use` from choosing another. Writing the file is a change of
state, so it is an operation of the engine like `config init`; what the file
becomes is `config.WithWorkspace` and `config.WithDefault`, which edit its text
rather than encode it again — encoding it again would drop every comment a
person wrote. A file that says its workspaces some other way, an inline table
for one, is refused rather than broken, and left to be changed by hand.

### The colour a workspace wears

```toml
# ~/cairn/home/settings.toml
[workspace]
colour = "green"
```

Each workspace wears a colour, so two boards side by side are told apart at a
glance: the dot before its name at the head of the terminal board, and in the
`^O` list ([tui.md](tui.md#another-workspace-in-place-of-this-one)). It is one
of the fourteen a label wears, said by `colour` under `[workspace]` in the
store's own settings (`@workspace-colour-said-in-its-settings`). Not said, it
is read off the workspace's name the way a label's is, so it is the same each
time and two workspaces seldom share one
(`@workspace-colour-read-off-its-name`). `workspace list` and `workspace show`
name it, `--json` too (`@workspace-colour-shown-and-as-json`); a store no
workspace names wears the colour its settings say, or none.

A colour that is none of the fourteen is refused like any value the settings
get wrong, naming it and `workspace.colour`
(`@workspace-a-colour-not-in-the-palette-is-refused`). **It belongs to the
store and to nothing else**: written in the global file it would colour every
workspace alike, so it is passed over there, and said
(`@workspace-colour-in-the-global-file-is-passed-over`). Listing reads each
workspace's settings on their own, so the colour of the store in use is never
lent to another (`config.ColourIn`).

### Recent folders, remembered and never configured

A store a board opens by path — `--store`, `CAIRN_STORE`, or a folder opened
from the terminal's `^O` list — that no workspace names is remembered among the
recent folders: ten at most, the newest first, one opened again moving back to
the head (`@workspace-a-board-opened-by-path-is-recent`,
`@workspace-a-board-opened-by-the-environment-is-recent`,
`@workspace-ten-recent-folders-are-kept`). `workspace list` names them under
the workspaces, each with the day it was last opened, and one whose store is
gone since as `not found now` (`@workspace-a-recent-folder-gone-is-not-found`,
`--json`: `@workspace-recent-folders-as-json`). A recent folder a workspace has
named since is no longer listed among them
(`@workspace-a-recent-folder-kept-as-a-workspace-leaves-the-list`).

**A board remembers, a command does not.** `cairn task list --store …` leaves
the list as it was (`@workspace-a-command-leaves-the-recent-folders-alone`):
a script that runs a command a thousand times, each with its own store, would
otherwise push out every folder a person opened by hand, and a list a script
churns is no list of what was opened. A board is a person at a screen; so is
the window, which opens through the same `engine.Hold`. A board opened on a
workspace, on the default or on the file's `store` leaves it alone too
(`@workspace-a-workspace-opened-is-not-a-recent-folder`): those are named
already.

**They are kept in the state directory, never in the configuration**
(`@workspace-recent-folders-are-not-the-configuration`):
`$XDG_STATE_HOME/cairn/recent.json` when that variable is absolute, otherwise
`~/.local/state/cairn/recent.json`
(`@workspace-recent-folders-follow-xdg-state-home`). What was opened last is a
fact cairn noticed, not a choice somebody made, and the configuration holds
choices — it is the file a person reads and keeps. The same place on every
system, macOS included, as the configuration is `~/.config/cairn` on every
system: `~/Library/Application Support` would split what cairn keeps between
two conventions on one machine. The window's sizes stay beside their store,
because they are that store's; the recent folders are about every store at
once, so they cannot sit beside any one of them. A file that cannot be read
is read as none, and written over the next time a board opens by path; one
that cannot be written leaves the board opening without it.

## Settings that belong to the store

```toml
# ~/cairn/home/settings.toml, beside ~/cairn/home/cairn.db
[pages]
objectives = false
```

A workspace's folder holds its store, its notes, and `settings.toml`. **What
belongs to the store is said there and wins over the global configuration, key
by key; what belongs to the machine stays in the global file.** One board is
work with objectives and a weekly archive, another home without either, and
switching workspace switches them with the store (`bdd/settings.feature`,
`@settings-two-workspaces-answer-differently`).

- **Which file**: `settings.toml` beside the store in use, in the same folder
  as its `cairn.db` (`config.SettingsBeside`). It is not looked up through the
  workspaces: a store named by `--store`, `CAIRN_STORE` or the file's `store`
  reads the settings beside it as well (`@settings-beside-a-store-named`), so
  a store moved by hand takes its settings with it.
- **What it may say**: every table but the machine's — `[pages]`, `[archive]`,
  `[focus]`, `[approach]`, `[time]`, `[window]` — and one table of its own,
  `[workspace]`, the colour it wears (above). A key it says wins
  (`@settings-the-store-wins`); a key it leaves out is the global file's, or
  the default (`@settings-a-key-not-said-comes-from-the-global-file`); each
  page is said on its own, so `objectives = false` here leaves
  `time = "hidden"` in the global file standing. `store`, `notes`, `plugin`
  and `[workspaces]` say where things are, which is the machine's to know
  before any store is open: written here, they are refused naming the global
  file they go in (`@settings-a-machine-key-is-refused`).
- **An empty string**: `label = ""` under `[approach]` means no approach label
  for this store, whatever the global file says
  (`@settings-an-empty-approach-label-is-none-here`). Any other key written
  empty here is refused naming it (`@settings-another-empty-string-is-refused`):
  in a file whose whole point is to replace a value, an empty one says nothing
  and would hide what the global file says.
- **What it does not know** is passed over and said, as the global file's is,
  the settings named (`@settings-an-unknown-key-is-passed-over`); a value it
  cannot read, or a file that is not TOML, is refused like the global file's.
- **No file is the ordinary case**: everything comes from the global file, as
  before there was one. `cairn workspace add` writes one that says nothing but
  what it is for — everything comes from the global configuration until it says
  otherwise — and leaves one already in the folder as it was
  (`@settings-workspace-add-writes-empty-settings`,
  `@settings-workspace-add-leaves-settings-there`).

Both files are read by the same reader: `config.Places.WithSettings` decodes
the settings exactly as the global file is decoded and lays each key it writes
over the one the global file gave. The engine reads them wherever it reads the
configuration — the places, the pages, the archive, the focus session, a
deleted task's time, the window's sizes — for the store it found, so a
command, a board and the MCP server all answer from the settings of their own
store (`@settings-the-server-reads-the-settings-of-its-store`); the server
reads them again at every call, as it reads the pages. The command line reads
`--workspace` and `--store` before it builds its commands, so a page another
workspace's settings turn off is not offered for that command.

`config show` names the settings read, `# the store's settings: <path>`, and
then says after each key of the store where its value comes from: `# set in
the store's settings`, `# from the global configuration`, or `# by default`
(`@settings-config-show-says-where-each-comes-from`). With no settings beside
the store it says nothing of them, as before. Its JSON carries `settings`, the
file read, and `written_here`, every key it writes, beside `written`, every key
the global file writes (`@settings-config-show-as-json`); the window's panel
says `set here` beside such a key, and `from the global configuration`
beside one only the global file writes
([gui.md](gui.md#workspaces-a-window-each)).

**Beside a key set here, what the global configuration would give.** A key
the settings write hides the global one, and taking it out is the way back;
so `config show` says where it would land: `# set in the store's settings;
global: "1h"` — the global file's value, or the default when it says none
(`@settings-config-show-says-the-global-value`). Its JSON carries `global`, a
value for each key in `written_here`. `workspace.colour` has none: no global
file says a store's colour. The window's panel says the same beside the key,
`← global: 1h`.

**The colour, with where it comes from.** `config show` writes the colour the
workspace wears under `[workspace]`, `# set in the store's settings` or `# read
off the workspace's name`, and its JSON carries it as `workspace.colour`
(`@settings-config-show-says-the-colour`). A store no workspace names and whose
settings say no colour has none, and the table is left out.

## Two places, and their names

```toml
# where cairn keeps your tasks
store = "~/.cairn/cairn.db"

# where cairn looks for notes taken elsewhere
notes = "~/.cairn/notes"
```

`store` and `notes`. Both name the place itself and not a folder to compose
something into: `store` is the file, `notes` is the directory task directories
sit under.

They fall under one folder by default, `~/.cairn`, so what cairn keeps stays
together and one directory is enough to move or back up by hand.

## Written for you

`cairn config init` writes the file where it is read, whole: every key,
explained, set to what cairn uses when nothing is said
(`@config-init-writes-every-key`). **It names one workspace, `main`, on
`~/.cairn`, and makes it the default**, its folder made as `workspace add`
makes one (`@config-init-names-one-workspace-the-default`): the store and the
notes are where they always were, `~/.cairn/cairn.db` and `~/.cairn/notes`, so a
machine that never heard of workspaces reads as before, and the next
`workspace add` joins a table already there. It is not called `default`, as
the study proposed, because `default` is the key naming the default and so a
name no workspace can take. `store` and `notes` are written commented, since
the default workspace would pass them over; a file written before, with no
`[workspaces]`, is read exactly as it was
(`@config-a-file-without-workspaces-reads-as-before`). **Every table is written out of its
comment**, and only a key that has no value of its own — the day archiving
counts from, how long a focus session lasts — is commented, under its table.
The file was first written with every table after `[pages]` commented whole:
taking the comment off `keep_sizes = true` alone put it under `[pages]`, and
every command then refused the file. Now any key taken out of its comment is
read where it stands (`@config-init-a-key-uncommented-is-read-under-its-table`,
`@config-init-a-value-changed-is-read`), and a test takes each commented key of
the example out of its comment in turn. A key cairn knows written under the
wrong table is refused with the table it belongs under
(`@config-a-key-under-the-wrong-table-names-its-own`). Writing one by hand meant finding its keys
here and in the README first. It is the README's example to the
byte — one constant, `config.Example`, that a test holds the README to, and
that another holds to every key the reader accepts — so neither can drift.
It never writes over a file already there, which may hold choices nobody
would want reset (`@config-init-leaves-a-file-alone`); and it writes what is
used when nothing is said, not what the environment says right now, because
a passing choice does not settle into the file.

## The plugin's folder

```toml
plugin = "~/.cairn/claude-plugin"   # where cairn writes its Claude Code plugin
```

Where `cairn plugin write` writes the plugin, and where `cairn mcp serve` keeps
it at its own version ([claude-code.md](claude-code.md)). Not said, it is
`claude-plugin` beside the store, wherever the store is. It is a key of the
file and nothing else — no flag, no variable — because two commands have to
agree on it: `plugin write --dir` named a folder the server never looked at,
and was taken out for it. Absolute, or under `~`, like the others
(`@config-names-the-plugin-folder`,
`@config-a-plugin-folder-that-is-not-absolute-is-refused`); `config show` says
where it is and whether the file named it.

## The pages, turned off one by one

```toml
[pages]
objectives = false  # tasks and objectives are on unless they say false
time = "hidden"     # "on", "hidden" or "off"
```

A page turned off is gone from every surface, and the engine is what says so
(`engine.Pages`): the boards draw no tab for it and `tab` walks past it, the
command line offers none of its commands — `task` and `step` with the tasks,
`objective`, `timesheet`, and `label` once neither tasks nor objectives are
left to label — and the MCP server lists none of its tools (`bdd/pages.feature`,
`@pages-off-takes-the-steps-with-the-tasks`,
`@pages-labels-go-with-the-last-page-they-label`, `@gui-a-page-off-has-no-tab`,
`@tui-a-page-off-has-no-tab`, `@mcp-a-page-off-offers-none-of-its-tools`).
A page turned on or off while an MCP server runs changes its tools from the
next call, and the client is told the list changed — no reconnect (see
[mcp.md](mcp.md#a-page-turned-on-or-off-while-it-runs-changes-its-tools-at-the-next-call)).

**One rule, asked by everything that offers an act.** It is `engine.Pages`:
`Offers` for an operation, `Called` for a thing and a verb typed. The window's
menu bar asked nothing at first — one fixed list, `MenuBar`, read by the native
bars of macOS and Windows and by the button standing in for them — so a page
turned off kept its lines there, and the terminal's `:` line still ran its
commands. Now each line of the menu bar names the acts it does,
and is offered while one of them is (`@gui-a-page-off-takes-its-menus`,
`@gui-the-tasks-off-take-their-menus`); a menu left with no line goes, and the
`:` line runs only what is offered (`@tui-a-page-off-takes-its-commands`).

**The time says one of three words.** `"on"`, the default, is recorded and
shown. `"hidden"` is recorded and not shown: every surface treats the page as
off — no tab, no time on a card, no timesheet command or tool — and the trace
keeps counting, so turned back on, every minute spent meanwhile is there
(`bdd/time.feature`, `@time-hidden-records-what-it-does-not-show`). `"off"` is
neither: shown as hidden is, and a task in focus counts no time at all — the
focus stays, without a timer, and turning the time back on starts the timer
then, nothing before caught up (`@time-off-records-nothing`,
`@time-on-from-off-starts-the-timer-then`); no focus session asks to be
renewed while it is off (`@time-off-asks-no-renewal`). Nothing already
recorded is taken away in any of the three (`@time-off-keeps-what-was-recorded`);
how that holds is in [trace.md](trace.md#time-turned-off-a-focus-without-a-timer).
`true` reads as `"on"` and `false` as `"hidden"`, so a file written before means
what it meant; `config show` writes the word (`@time-config-show-writes-the-word`).
Anything else is refused naming `pages.time` and the three words
(`@time-a-word-it-does-not-know-is-refused`). A store's settings may say it as
any page (`@time-off-in-the-store-settings`).

With every page off a board opens with nothing in it, its menu bar with
nothing either (`@gui-every-page-off-leaves-what-belongs-to-none`), and only
closing it does anything. What belongs to no page — the configuration,
backups, migrations, the trace, the plugin — stays on the command line and
for an agent. A
page cairn does not have is refused like any other unknown name, and `config
show` says which pages are off. A configuration that cannot be read turns no
page off: every command that needs it refuses anyway, and says why.



A place that depends on where the command was run from answers differently
depending on who asks, and a configuration is read from everywhere. So `notes`,
`./notes` and `../vault` are refused, naming what was wrong.

`~/vault` is absolute and is accepted — it is also what anyone writes by hand,
and it is the one form that means the same thing on macOS and on Linux, which
`/Users/me/vault` does not. A leading `~` is expanded first and the result is
what is judged: `filepath.IsAbs("~/vault")` is false, so the two steps cannot be
folded into one.

**A key written as the empty string is refused, and a variable set to nothing is
not.** The asymmetry is deliberate and reads as a bug if it is not said out
loud. `store = ""` is a choice somebody wrote down and that does nothing — the
same silence this file condemns two sections below, and it falls out of the
absoluteness rule for free, since an empty string is not an absolute path.
`CAIRN_STORE=` is a passing choice, and a passing choice that says nothing is
simply nothing: it leaves the answer as it was worked out. The file is written;
the environment is not.

## Archiving on its own

```toml
[archive]
start = 2026-10-21   # the day the count starts
days = 14            # how long each period lasts; 14 when it is not said
```

A done column fills up: 55 tasks sat there next to 129 archived before anybody
swept it by hand. With `[archive]` said, days are counted from
`start` in periods of `days` calendar days, and a task that went done before
the period under way began is archived. The periods are arbitrary: nothing in
cairn gives them a meaning beyond the boundary they draw. The moment a task
went done is its last move into `done` in the trace. This runs when a board
opens (`engine.Hold`) and when the tasks are listed, through the command line
or the MCP server (`@archive-what-went-done-before-the-period`,
`@archive-keeps-what-went-done-in-the-period`), through the same act as
`task archive`, written down as done through `archive`
(`@archive-on-its-own-is-said-in-the-trace`). Nothing said, nothing is
archived on its own (`@archive-nothing-on-its-own-unless-told`) — `days` alone
is read, and archives nothing — nor before
the day the count starts (`@archive-nothing-before-the-count-starts`). A start
that is not a day, or periods of no day, are refused like any key the file
gets wrong (`@config-archiving-in-periods-of-no-day-is-refused`,
`@config-archiving-counts-from-a-day`), and `config show` says it
(`@config-show-says-what-is-archived-on-its-own`).

Calendar days, not working days: counting working days would ask which days
are holidays, which cairn does not know. The day the period under way began is
not kept anywhere: it is computed from `start` each time, so changing `start`
or `days` takes effect the next time it runs.

Today is the day the clock says, unless `CAIRN_TODAY` names another, written
2026-11-04: a script, or a scenario, stands on that day to see what would be
archived then. A scenario cannot move the trace back — it asks the engine,
which only writes the present, and never opens the store — so it moves today
forward instead. A day it cannot read is taken as no day named.

## A focus session that asks to be renewed

```toml
[focus]
session = "1h"   # how long a focus session lasts before it asks
```

A focus left open counts time nobody spent, but no clock can tell that the work
stopped: cutting it at a time of day, or after a while unused, was built and
taken back the same day. The question goes to whoever holds the
focus instead. Each task in focus has a session, up since the focus was taken
or last renewed, read from the trace; the focus answer says it for every task
in focus (`sessions`, with `since` and `up`,
`@focus-session-says-how-long-it-has-been-up`). Once a session has been up for
`session`, it asks to be renewed or ended (`ask`,
`@focus-session-asks-to-be-renewed-past-its-length`), and every answer the MCP
server gives carries that question after its own text, whatever tool was
called (`@mcp-any-answer-asks-to-renew-a-focus-session`): an agent rarely calls
`task_focus`. Focusing the task again renews the session, written `focus
renewed` in the trace (`@focus-session-renewed`); within its length it writes
nothing. The time runs on through a renewal: only a release stops it. Nothing
said, nothing is asked; nor while the time is off, since nothing is counted
(`@time-off-asks-no-renewal`). A length that is not a duration is refused, naming
`focus.session` (`@config-a-focus-session-that-cannot-be-read-is-refused`), and
`config show` says it (`@config-show-says-how-long-a-focus-session-lasts`).

The Claude hook takes the focus at each prompt, so a prompt past the length
renews the session: a person writing is a person at work. The question is left
for an agent working alone for longer than a session.

Now is the clock, unless `CAIRN_NOW` names another moment, written in RFC 3339:
a scenario stands later to see a session grow, and the acts written meanwhile
are written at that moment.

## A label that asks for the approach to be agreed first

```toml
[approach]
label = "discuss"   # a task carrying it is not coded before you agree on how
```

Some tasks leave real choices open, and code written before they are made is
code that decides them. The label marks those tasks: an agent working on one
writes `approach.md` in its notes — the solution, what was set aside, the
scenarios, what is out of scope, the open questions, a mockup when the screen
changes — presents it, and writes no code before the person agrees; on
agreement a dated entry records it and the label comes off. You put the label
on; an agent may propose it, never put it on itself.

There is no default. Without the key, or with it empty, nothing of this exists:
the skill says nothing of it and an agent works as before
(`@config-an-empty-approach-label-is-no-label`). Any word will do — `discuss`
is only the one the file suggests, commented. cairn enforces nothing: it keeps
the key and shows it in `config show` and over MCP
(`@config-show-says-the-approach-label`); the skill gives the label its meaning
(see [claude-code.md](claude-code.md)). An older cairn passes the key over, and
says so.

## What a deleted task does with its time

```toml
[time]
deleted = "taken"   # or "kept", what is done when it is not said
```

Kept, a deleted task's time stays on the sheet, said to be deleted; taken, it
leaves with the task. See [trace](trace.md#the-trace-outlives-the-task). Any
other value is refused naming `time.deleted`
(`@config-what-a-deleted-task-does-with-its-time-is-one-of-two`), and
`config show` says which it is
(`@config-show-says-a-deleted-task-takes-its-time`).

## The weekends of a week

```toml
[time]
weekends = "hidden"   # or "shown", what is done when it is not said
```

Google Calendar has an option for the company calendar, Show weekends, on by
default: off, the week and month views lose Saturday and Sunday and the day
view does not. cairn's time page does the same, in the window and in the
terminal's time tab alike: hidden, a week is five days, Monday to Friday, and a
month five columns a week. Whether the weekends are shown is answered once by
the engine, which hands the boards how many days a week shows
(`Pages.WeekDays`, read from the configuration beside which pages are on), and
both lay out only those. Hiding is drawing less, not counting less: the week is
still read whole, so time spent on a Saturday stays in any total that counts the
week, and the hidden days' cards are simply not drawn. Any other value is
refused naming `time.weekends`
(`@config-weekends-is-one-of-two`), and `config show` says which it is
(`@config-weekends-hidden-or-shown`). The window has no toggle for it: the file
stays the one place a key is set, and `⌘,` shows the configuration in use and
opens the file from there ([gui.md](gui.md#the-window)).

## The environment names the same two places, or a workspace

`CAIRN_STORE` and `CAIRN_NOTES`, one per key, and `CAIRN_WORKSPACE`, which
names a workspace rather than a place (above). They replace the
answer rather than taking a turn before it — see the order above — and what they
name goes through the same expansion and the same absoluteness rule as what the
file names, and so does `--store`. A place is read the same way wherever it was
named. Only a flag replaces them in turn.

## The store wants a disk of its own

The store keeps its latest writes in a log beside it (`-wal`, `-shm`) until it
is closed — the journal that lets a board read while a command writes. That
log is shared through the machine's own memory, which a network share does not
offer and a synchronised folder does not respect: put `store` on NFS, SMB, or
in a folder iCloud or Dropbox keeps in step, and two machines can each see
half a write. The store goes on a local disk; what travels between machines is
a backup, which is one file written whole.

## The window's sizes, kept or not

```toml
[window]
keep_sizes = true   # the window opens as wide and as tall as it was left
```

Off unless it says true. On, the window keeps how wide each column and each
lane was dragged and how tall the band, in a file beside the store
(`cairn.db.sizes.json`), and the next window opens with them; off, every window
opens as it starts. Either way, a double-click on a gutter or `Reset sizes`
takes them back ([gui.md](gui.md)). Anything but `true` or `false` is refused,
naming `window.keep_sizes` (`@config-keep-sizes-is-yes-or-no`), and `config
show` says whether it is on (`@config-show-says-the-window-keeps-its-sizes`).

## The colours the window wears

```toml
[window]
theme = "gruvbox-dark"   # a scheme cairn carries, or one of your own
```

The window wears a colour scheme, named by `theme` under `[window]`:
`default` — the look it always had — when nothing names one. Nine are
carried in the binary: `default`, `solarized-dark`, `solarized-light`,
`gruvbox-dark`, `gruvbox-light`, `catppuccin-mocha`, `catppuccin-latte`,
`dracula` and `nord`. **One of your own** is a base16 file, as
tinted-theming writes it, in `~/.config/cairn/themes/<name>.yaml`, beside
this file; named as one cairn carries, yours is the one worn
(`@theme-a-scheme-of-my-own-wins-over-a-shipped-one`). What a scheme file
holds and how it becomes the window's colours is [gui.md](gui.md#the-design-system)'s.

**It belongs to the store**, like every key of `[window]`: a workspace's
`settings.toml` names its own, so two windows side by side are told apart
at a glance (`@theme-a-shipped-scheme-colours-its-workspace`). The window
reads it when it opens; a window opened on another workspace is a cairn of
its own, which reads that workspace's.

**A scheme that cannot be worn never keeps the window shut.** A name cairn
sees no scheme by, a file it cannot read, a colour that is not one: the
window opens in `default`, and `cairn gui` says why on standard error,
once, as a key passed over is said — `theme: "neon" is not a scheme cairn
sees; the window wears default`
(`@theme-an-unknown-scheme-falls-back-and-says-so`,
`@theme-a-bad-colour-is-said`). A name that is not one — capitals, a `/`,
a leading `.` — is a scheme nobody has, so a name can never reach outside
the themes folder. `theme = ""` in the global file is `default`; in a
store's settings it is refused like any key written empty.

`config show` writes `theme` with where it comes from, says why the scheme
named cannot be worn, `# theme: …`, and lists **every scheme it sees**, one
a line, `shipped` or the file of your own it reads, the one in force marked
(`@theme-config-show-lists-the-schemes`,
`@theme-config-show-says-what-cannot-be-used`); its JSON carries
`window.theme`, `window.unworn` and `themes`. A scheme of your own that
falls short of a contrast the window holds its colours to is worn all the
same, and `config show` says so, a line for each pair, the two tokens named
and the ratio they read at: `# dim falls short: ink.text on surface.back
reads at 1.63, and needs 4.5` (`@theme-a-scheme-hard-to-read-is-said`).
Every scheme cairn carries meets them all
(`@theme-every-shipped-scheme-passes-contrast`). The window's configuration
panel says `window.theme` beside the other keys.

**What it does not do yet.** The terminal board keeps its own colours: moving
it to the sixteen colours a terminal's own scheme sets is a task of its own.
A window does not follow the system between light and dark: `theme` is one
name, and can become a table later without breaking a file that names one.

## TOML, because a person opens this file

This file is for someone to read and change, so it has to be able to carry a
line above each place saying what that place is. That requirement decided the form on
its own: JSON has no comments and turns paths into escapes, and a hand-rolled
`key = value` reader means inventing rules for quoting, whitespace and repeats
that a specification already answers.

The dependency is real and this is its written reason.

## A name cairn does not know is passed over, and said

It was refused, once: a setting somebody believes is in force and that does
nothing is worse than no setting at all. The price came due when a newer cairn
wrote `time.weekends` and the plugin's server, an older cairn Claude Code had
kept, refused every call for it — the command line worked, the agent did not,
for a key it had no use for. Two cairn of different versions read one file
more often than it seems: the application left behind, a build from the
repository then the released one, another machine sharing the file.

So a key this cairn does not know is passed over, the rest of the file is read
and used, and it is said every time, never quietly: a line on standard error
for every command, `config: "time.weekends" is not a key this cairn knows —
…; it is ignored` (`@config-a-key-from-a-newer-cairn`), and a `# ignored:` line
in `config show` (`ignored` in its JSON), so a typo stays in sight
(`@config-an-unknown-name-is-passed-over`). A key written under the wrong table
still says the table it belongs under
(`@config-a-key-under-the-wrong-table-names-its-own`). The MCP server, whose
standard error nobody reads, answers all the same
(`@mcp-a-key-it-does-not-know-does-not-stop-it`).

What is still refused is a value cairn knows the key of and cannot read — a
duration that is not one, a place that is not absolute — and a file that is
not TOML: those are settings that would do something else than what was meant.

## Not posed yet

Cairn does not write this file on its own. Nothing has to exist for it to work —
no file is the ordinary case — but nothing leaves one behind either: a person
who wants one writes it, or asks for it: `config init` writes it whole, and
`workspace add` and `workspace use` write `[workspaces]` into it, leaving the
rest as it was. `workspace add` also writes the settings of the workspace's
store, which say nothing yet.

Posing it at the first run is specified in `bdd/config.feature` — once, never
completed, never holding what the environment said, never by a server told to
only read — and every one of those scenarios carries `@later`. They are not run,
and this page does not describe them as what cairn does, until they are built.

## Reading is not the same as not being there

A file that is there and is not TOML stops the command, naming the file. No file
at all is the ordinary case and stops nothing. A file that is there and cannot be
read stops the command too, but today as a store that cannot be used (exit 3);
answering it as a refusal that names the file is specified, and carries `@later`.

Three answers rather than one, because a permission and a typo are not an
absence, and a caller that cannot tell them apart looks for the wrong problem.

## `config show` reports where every other command refuses

`cairn config show` says the whole configuration cairn uses, written as the
file is: every key under its table with the value in force — what the file
says, or what is used when it says nothing — and, beside each place, what
named it: the file, the environment, or nothing
(`@config-show-says-every-key`). A key with no value is said as a comment that
says what that means. It used to say the places and, in words of its own, only
what the file changed: what cairn really used could not be read off it. What it
says is a configuration cairn reads as the same
(`@config-show-reads-as-a-configuration`), so it can be copied into the file.
Answered as JSON, it also lists `written`: every key the file writes, dotted
under its table, so a surface can say of any key whether the file or the
default set it (`config.Places.Written`); the window shows it beside each
value. It is the only command that does not
refuse a configuration it cannot make sense of — and that exception is the
point, not an oversight. A person reaches for it *because* something is not
where they expected it, so a diagnostic that answers with the same refusal the
last command gave is a diagnostic that is useless exactly when it is called.

It does not lie in the other direction either. On a file it cannot make sense
of it prints nothing about places — showing `store → somewhere` would say the
opposite of what is true, since that same file turns every other command away —
and it says so: the file, the problem, and the warning that the rest will refuse
until it is fixed. A scenario measures that last sentence by running another
command on the same file, rather than taking cairn's word for it.

Two consequences worth knowing rather than discovering. It carries
`Needs: AConfiguration`, which is what keeps the shared path from resolving a
store before it runs — without that it would be unreachable precisely when it
matters. And over MCP it answers normally rather than as a refusal, because a
refusal carries nothing for a machine, and the file's path is the one thing an
agent needs when everything else has started saying no.

## Deliberately not there yet

- **A command to write the file.** It is posed, and after that it is edited the
  way any file is. A `config set` would be a second way of doing one thing.
- **Saying what the file named when the environment replaced it.** `config show`
  says what won and who named it, never what lost. It was sketched and left out:
  no scenario observes it, and a capability nothing observes is either specified
  or dropped.
- **Anything but places — and the pages.** Nothing about how cairn behaves
  belongs here. The pages were that decision, taken on its own on 2026-10-16: which of them somebody uses is a fact about the person, read the
  same way by every surface, so it lives where every surface already looks.
  Any other preference is still a decision of its own.
- **A file per project.** One global file, at the place configuration lives,
  and beside each store the settings that belong to it — never a file looked
  for in the folder a command runs from. A store somewhere else is what a
  workspace, the environment or `--store` is for.
