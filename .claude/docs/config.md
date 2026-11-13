# Where cairn finds its own places

How it is used is [docs/config.md](../../docs/config.md), which the site publishes;
this page is why it is that way, for whoever changes it.

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
3  the current directory  under a workspace's folder, or the project its .cairn sits in
4  the default workspace  [workspaces] default, when it is said
5  main                   ~/.cairn/cairn.db   ~/.cairn/notes
                          the first rung that says something answers,
                          and every rung under it is passed over
```

Five rungs, read from the top, and the first one that names a store answers.
**A workspace always answers**: the three lower rungs each name one, and the
places follow from it — its store, its notes, its settings. The directory and
the default are choices made and kept, and `main`, in `~/.cairn`, is what a
choice looks like when nobody made one — the same on every machine, so "where
are my tasks" never depends on which one you are sitting at
(`@workspace-main-when-none`, `@workspace-no-default-answers-main`).
Until v0.12.0 the file named a `store` and `notes` of its own at its top, as a
fourth rung; a file still written that way is turned into workspaces the first
time cairn reads it (see below), and the rung `the file` is only ever said of
a store no workspace's folder could hold.

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
**The current directory names the workspace of the work in front of you.**
A command run from inside a workspace's folder, or from inside the project a
`.cairn` folder sits in, answers from that workspace
(`@workspace-the-directory-names-it`, `@workspace-the-directory-inside-a-folder`);
the nearest folder wins when two hold the directory
(`@workspace-the-nearest-folder-wins`). It comes under the environment and
over the default (`@workspace-the-directory-under-the-environment`,
`@workspace-the-directory-above-the-default`), and `workspace show` says `the
current directory — over the default, home`
(`@workspace-show-says-the-directory`). It is what makes the shell, the
plugin's MCP server and its hook agree without a variable: each starts in the
project, so each finds the same workspace. A `.cairn` folder sitting in the
home itself is no project's — the home holds everything, and its workspace
would answer from everywhere under it, the default never again — so it
answers only from inside that folder (`@workspace-the-home-is-no-project`).
A folder gone names nothing, and the rungs below answer.

**The default workspace is a choice kept.** Once `[workspaces]` names a
default, the store and the notes are its own
(`@workspace-the-default-answers`, `@workspace-its-notes-are-in-its-folder`).
Without one, `main` answers. The environment still replaces it
(`@workspace-the-environment-replaces-the-default`).

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
otherwise the notes of `main` — those the configuration names for it, or
`~/.cairn/notes` — never the default workspace's
(`@workspace-a-store-named-keeps-its-own-notes`). `CAIRN_STORE` and `--store`
replace the store of the workspace chosen however it was chosen, the directory
included (`@workspace-cairn-store-overrides-the-chosen`).
Taking the default's notes beside a store from elsewhere would write one
board's notes into another's folder.


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
- `cairn workspace forget <name>` takes the workspace out of the file — its
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
`workspace use` and `workspace forget` are not (`@workspace-an-agent-cannot-add-or-switch`,
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

### The name a store gives itself

```toml
# ~/scratch/settings.toml
[workspace]
name = "scratch"
```

A store no workspace names is shown by `name` under `[workspace]` in its own
settings, or else by its folder's own name — never by a whole path, which a
pill, a title or a line of the switcher could only cut until it read nothing
(#285, `engine.WorkspaceShown.Name`, `RecentFolder.Name`). A workspace's own
name wins over it. `cairn fixtures` writes `name = "fixtures"`. The window's
settings offer it beside the colour; the whole path is said under the pointer.

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
  `[focus]`, `[approach]`, `[time]`, `[window]`, `[series]` — and one table of its own,
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

## A workspace's notes, where it says

```toml
[workspaces]
default = "cairn"
work = "~/cairn/work"

[workspaces.cairn]
folder = "~/code/cairn/.cairn"
notes = "~/vaults/cairn"
```

A workspace written the short way keeps its notes in its folder,
`<folder>/notes` (`@workspace-short-form-keeps-its-notes`). Written as a table
of its own, it names its `folder` and the `notes` it keeps elsewhere — a vault
a notes tool opens, one per workspace (`@workspace-names-its-notes`; for
Obsidian, [notes-obsidian.md](../../docs/notes-obsidian.md)). The two
forms sit in one file. A table naming no folder, or a key other than those
two, is refused naming it. `store` and `notes` are not keys of the file any
more: they are what the workspace answering gives, shown by `config show` as
comments and by the window as read when cairn opens, and `config set` refuses
them, naming `[workspaces]` and the variables.

## From v0.11.0, on its own

A configuration of v0.11.0 names its `store` and `notes` at its top. The first
command of v0.12.0, whatever the surface — the command line, a board, the MCP
server, the plugin's hook — rewrites it, without asking: the two places
become the workspace `main`, the default, or the workspace already at that
folder (`@workspace-old-configuration-upgraded`,
`@workspace-old-configuration-joins-a-workspace-at-the-same-place`,
`@workspace-old-configuration-upgraded-by-the-server`). **Nothing moves on
disk**: the workspace is written with the notes in use, even those nobody
named (`@workspace-old-configuration-keeps-the-notes-in-use`), and a place
can be changed afterwards by hand. A default already named stays the default,
and `main` is written beside it, since the top-level places were passed over.
The file as it was is kept beside it, `config.toml.v0.11.0`; what changed is
said once on standard error, and nothing is said after
(`@workspace-old-configuration-upgraded-once`). A store not called `cairn.db`
cannot be any workspace's: it is left at the top of the file, read as before,
and said. Asking first was weighed and set aside: the answer could only be
yes, and a no would have left every surface asking again.

## Notes moved by hand

Changing a workspace's `notes` in the file leaves the notes where they were.
cairn remembers, beside the store (`cairn.db.notes.json`), where it last found
them, so it knows both places: run in a terminal, it offers to move them,
`[y/N]` (`@workspace-notes-changed-offered-to-move`); on no nothing moves and
the next run asks again (`@workspace-notes-changed-left-on-no`). With nobody to
answer — the MCP server, the hook, a pipe — nothing moves and standard error
says a run in a terminal will offer it (`@workspace-notes-changed-without-a-person`).
A folder that already holds something is never moved into
(`@workspace-notes-not-moved-onto-others`). Notes the environment names are
nobody's to move.

## Written for you

`cairn config init` writes the file where it is read, whole: every key,
explained, set to what cairn uses when nothing is said
(`@config-init-writes-every-key`). **It names one workspace, `main`, on
`~/.cairn`, and makes it the default**, its folder made as `workspace add`
makes one (`@config-init-names-one-workspace-the-default`): the store and the
notes are where they always were, `~/.cairn/cairn.db` and `~/.cairn/notes`,
and the next `workspace add` joins a table already there. It is not called
`default`, as the study proposed, because `default` is the key naming the
default and so a name no workspace can take. A file of v0.11.0, with no
`[workspaces]`, keeps the places it used
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
here first. It is the reference's example to the byte — one constant,
`config.Example`, that the documentation site prints whole
(`@site-configuration`), and that a test holds to every key the reader
accepts — so neither can drift. The README no longer shows it: it links to
the site's configuration page (`@site-readme-points-to-the-site`).
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
(`bdd/time.feature`, `@time-hidden-records-what-it-does-not-show`). No time
spent is said of a task, an objective or a series either, in text or in JSON,
to an agent or on a board: the engine takes it out of every answer before a
surface sees it, so a surface has nothing to forget; an export still carries
it, being the store's data and not a picture of it
(`@time-not-shown-says-no-time-spent`). `"off"` is
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

With every page off a window says so, and offers the settings — a button,
`⌘,`, and Configuration… alone in its menu bar — since the settings are the
one way back to a page (#278, `@gui-every-page-off-is-an-empty-window`,
`@gui-every-page-off-leaves-what-belongs-to-none`). What else belongs to no
page — backups, migrations, the trace, the plugin — stays on the command line
and for an agent. A
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

## Tasks that come back, turned on per store

```toml
[series]
on = false   # true lets a task come back as a series
```

Off unless it says `true`, in the file or in a store's `settings.toml`, so one
workspace has series and another none. Off, every surface offers none of it:
`--every`, `--series` and `task edit --series` are no flags, `series` and
`task skip` no commands, and their tools are not offered to an agent — the
engine refuses them anyway, `recurrence-off`, to whatever reaches it another
way. Nothing is made, the series stay dormant as they were, and an occurrence
open is an ordinary task (`bdd/series.feature`, `@series-off-offers-nothing`).
Turned back on, every series is there again and makes nothing it missed; an MCP
server running sees it at its next call, its tools' shapes included. Anything
but `true` or `false` is refused naming `series.on`
(`@series-config-recurrence-is-on-or-off`), and `config show` says it
(`@series-config-recurrence-is-said`). What a series is: [series.md](series.md).

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
(`@config-weekends-hidden-or-shown`). The window's settings set it as a choice
of two, `shown` or `hidden`, written into the file like any key
([gui.md](gui.md#settings-changed-in-place)), and the week redraws at once.

## A file opened shows every key it can hold

`cairn config fill` writes into the settings of the store in use — or, with
`--global`, into the global configuration — every key that file can hold and
does not say yet, as a comment carrying the value in force and where it comes
from:

```toml
[window]
# theme = "nord"   # from the global configuration
```

Uncommenting a line sets it. A line or a comment already there stays as it
is, and a key already written, or already commented, is left alone
(`config.WithCommented`, `TestConfigFill_writesEveryKeyTheFileCanHoldAsACommentWithItsValueAndWhereItComesFrom`).
The window does it before it hands a file to the system's editor
(`TestOpeningAFileOfTheConfigurationFillsItFirst`, #291): the settings of a
workspace used to open nearly empty, only what it set itself, and the other
keys could be neither guessed nor changed by hand. Like `config set`, it is
for a person: no agent is offered it.

## One configuration, several ways in to it

```sh
cairn config set pages.time hidden           # the settings of the store in use
cairn config set focus.session 45m --global   # the global configuration
cairn config unset pages.time                # its line taken out: the value under it answers
```

**The configuration stays one, in one place.** The global `config.toml` and
the `settings.toml` beside each store are still the two files, and still the
truth. The window, the terminal and the command line are several ways in to
them, each a different picture of the same thing: the window draws a settings
screen ([gui.md](gui.md#settings-changed-in-place)), the terminal takes `:set`
and `:unset` on its `:` line ([tui.md](tui.md#help-and-the-command-line)), and
the command line `config set` and `config unset`. None of them writes the file
itself: a surface sends the key, the value and the scope it was asked, and the
engine checks and writes (`engine.ConfigSet`, `engine.ConfigUnset`), so one
change asked three ways is one change. No decision is reversed by this: the
file is not encoded again from a form, and a person who prefers an editor
loses nothing.

**Every key is in one catalogue** (`config.Keys`): its dotted name, its kind —
`bool`, `words`, `duration`, `day`, `number`, `text`, `path`, `colour`,
`theme` — the words it takes, whether it belongs to the machine (`store`,
`notes`, `plugin`) or to a store alone (`workspace.colour`), and the section
and words a surface names it by. The window's rows, the terminal's completion
and the writer read it; a test holds it to `config.Example`, key for key
(`TestTheCatalogueHoldsEveryKeyTheExampleWrites`). `[workspaces]` is not in
it: `workspace add`, `use` and `remove` write that table.

**Which file is written.** `--global` writes the global configuration.
Without it, the settings of the store in use, when a workspace names that
store or settings are already beside it; for a store no workspace names and
with no settings beside it, the global configuration, so a loose store does not
grow a file nobody asked for (`@config-set-writes-the-settings-of-the-workspace`,
`@config-set-a-loose-store-goes-in-the-global-file`). A key of the machine
always goes in the global file (`@config-set-a-key-of-the-machine-goes-in-the-global-file`),
and the colour a workspace wears never does
(`@config-set-a-key-of-a-store-is-not-global`). The engine decides this, not
the surface: the window offers the store's scope exactly when the engine would
write it.

**Edited as text, line by line**, as `config.WithWorkspace` does
(`config.WithKey`, `config.WithoutKey`). A value changed keeps its line and the
comment after it; a key added joins the end of its table, or a table of its
own at the end of the file; a key `config init` wrote commented
(`# session = "1h"`) is taken out of its comment where it stands, so its
explanation stays above it (`@config-set-global-takes-a-key-out-of-its-comment`);
a key taken out loses its line and nothing else — the comment above it stays,
since it may say why, and a person removes it (`@config-unset-takes-the-line-out`).
With no file, the file written holds that table and key alone, as `workspace
add` writes one, not the whole example
(`@config-set-without-a-file-writes-the-table-alone`). A key the file does not
say is not taken out of it (`@config-unset-refuses-a-key-not-said-there`).

**Nothing is written that does not read back.** Every write reads the file
first, edits that, then decodes the result twice: as TOML, every key but the
one changed compared with what it was, and with the whole reader
(`config.Readable`), the settings' rules included. A shape the editor does not
know — an inline table, a string over several lines — is left for the hand, as
`workspace add` leaves one. A value the key does not take is refused in words
naming the key — `"hiden" is not a word pages.time takes (on, hidden, off)`,
`focus.session: "1 hour" is not a duration — write 1h, 45m, 1h30m` — and
nothing is written (`@config-set-refuses-what-the-key-does-not-take`). A file
that cannot be read is not written into: writing a key into a file that does
not read would be guessing, so it is refused naming the file
(`@config-set-refuses-a-file-that-does-not-read`). The file is written beside
itself and renamed over, so a reader never sees half of it.

**Read again without reopening cairn.** The command line and the MCP server
read the configuration at every use. A board, which stays open while the file
changes, reads it again on the beat it already has — the window's
([gui.md](gui.md#settings-changed-in-place)) and the terminal's
([tui.md](tui.md#staying-true-while-the-store-changes)) — and a file that
cannot be read meanwhile leaves the last good one in force, said.

**A person's, not an agent's.** `config set` and `config unset` are
`engine.ForAPersonOnly`: turning a page off, or the time, is the person's to
decide, like `workspace use` (`@config-an-agent-cannot-set-a-key`).

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

## What is dragged

```toml
[window]
ghost = 65   # per cent, 20 to 90
```

What the pointer carries — a card, an objective, a step — is drawn
see-through, so the line saying where it would land stays in sight under it.
`ghost` says how much of it shows: 65 per cent unless said. Under 20 it is
lost against the board, over 90 it hides the line again; anything outside,
or not a whole number, is refused, naming `window.ghost`. The settings offer
it under Appearance ([gui.md](gui.md)).

## How the board arrives

```toml
[window]
intro = "ripple"   # ripple, deck, spring, accordion, sweep, flap, pull,
                   # particles, dolly, blueprint or stones
```

When the window opens, its board arrives by an intro, a breath long, which
`intro` names: `ripple` — the board born from the card in focus — unless it
names another (`@config-the-intro-is-ripple-unless-said`,
`@config-the-intro-is-said`). What each one does is
[gui.md](gui.md#the-board-arrives-by-an-intro)'s. A name that is none of
them is refused, naming `window.intro`
(`@config-an-unknown-intro-is-refused`); a system that asks for less motion
gets none, whatever the file says. It belongs to the store, like every key
of `[window]`, and the window's settings offer it as a choice played on the
board behind before it is kept.

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

**The terminal board does not wear it.** It draws in the sixteen colours
the terminal's own scheme sets, and so follows that scheme instead
([tui.md](tui.md#the-terminals-own-colours)).

**What it does not do yet.** A window does not follow the system between light and dark: `theme` is one
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
who wants one writes it, or asks for it: `config init` writes it whole,
`workspace add` and `workspace use` write `[workspaces]` into it, and `config
set`, from any surface, writes one key, leaving the rest as it was. `workspace add` also writes the settings of the workspace's
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
