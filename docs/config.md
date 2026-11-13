# Configuration

Nothing needs configuring: cairn keeps its tasks in `~/.cairn/cairn.db` and
their notes in `~/.cairn/notes` until told otherwise. `cairn config init`
writes the whole file with every key explained, set to what cairn uses when
nothing is said.

There are two files:

- **the global configuration**, `~/.config/cairn/config.toml` on every system:
  where things are, and the defaults;
- **a store's settings**, `settings.toml` beside each store: how that board
  behaves. A key said there wins over the global file, key by key.

## Setting a key

```sh
cairn config show                          # every key in force, and where it comes from
cairn config set pages.time hidden         # in the settings of the store in use
cairn config set series.on true --global   # in the global configuration
cairn config unset pages.time              # its line taken out; the value under it answers
cairn config fill                          # every key the file can hold, commented, with its value
```

A key set keeps every comment of the file, and is refused unless the file reads
back as meant. The window's settings (`⌘,`) and the terminal's `:set` and
`:unset` write the same way. You may also edit the files by hand: they are
TOML.

A key cairn does not know is passed over and said on standard error, so a file
written by a newer cairn still works with an older one. A value cairn cannot
read is refused, naming the key.

`config show` is the one command that reports a configuration it cannot make
sense of instead of refusing: it says what is wrong, and warns that everything
else will refuse until it is fixed.

## Where everything is

Every command finds its store and its notes by going down this list, and the
first line that says something answers:

```
1  a flag                 --workspace <name>, or --store <path>
2  the environment        CAIRN_STORE, CAIRN_NOTES, then CAIRN_WORKSPACE
3  the current directory  under a workspace's folder, or the project its .cairn sits in
4  the default workspace  [workspaces] default
5  main                   ~/.cairn/cairn.db and ~/.cairn/notes
```

`cairn workspace show` says which line answered, and which lines under it were
passed over:

```
why  CAIRN_STORE — over the default, home
```

- The environment and the flags are passing choices: they change nothing on
  disk, and the next command without them finds things as they were.
- A store named by `--store` or `CAIRN_STORE` keeps its own notes: those of its
  workspace when it is one, `CAIRN_NOTES` when it is set — never another
  board's.
- Every place is absolute or starts with `~`. A relative place is refused.
- The store belongs on a local disk. A network share or a synced folder (iCloud,
  Dropbox) can show two machines half a write each; move a board between
  machines with a backup.

## Workspaces

A workspace keeps a store and its notes together in one folder, under a name:
one board for work, another for home.

```sh
cairn workspace add work ~/cairn/work   # names it in the file, makes the folder
cairn workspace use work                # the one used when nothing else says
cairn workspace list                    # each one, the default marked
cairn workspace show                    # which one answers, and why
cairn workspace forget play             # forgets it; its folder is left alone
cairn task list --workspace home        # another, for this command alone
```

In the global file:

```toml
[workspaces]
default = "work"
work = "~/cairn/work"
home = "~/cairn/home"
```

Its store is `<folder>/cairn.db`, its notes `<folder>/notes`, and its settings
`<folder>/settings.toml`. A name is lowercase letters, digits and dashes. A
workspace the file does not name, or whose folder is gone, is refused.

A command run from inside a workspace's folder, or from a project holding a
`.cairn` folder, answers from that workspace. That is how the shell, the MCP
server and the Claude Code plugin agree on a board without a variable.

`CAIRN_WORKSPACE=work` names a workspace for a whole session, the way
`--workspace` does for one command.

On the boards, the terminal names the workspace at the head of its header and
`^O` opens another in its place; the window names it in its title and opens
another in a window of its own. A store opened by path from a board is
remembered among the recent folders, ten at most, which `workspace list` names.

An agent may see which workspace it answers from, never change it: adding,
using and forgetting a workspace are yours.

### A workspace's notes

A workspace can keep its notes elsewhere — in a vault a note tool opens:

```toml
[workspaces.cairn]
folder = "~/code/cairn/.cairn"
notes = "~/vaults/cairn"
```

See [Obsidian](notes-obsidian.md).

### Notes moved by hand

Changing a workspace's `notes` leaves the notes where they were. The next run
of cairn in a terminal offers to move them, `[y/N]`; nothing moves without that
yes, and nothing is moved into a folder that already holds something. With
nobody to answer — the MCP server, a pipe — nothing moves.

### Its colour and its name

```toml
# ~/cairn/home/settings.toml
[workspace]
colour = "green"     # one of the fourteen colours a label wears
name = "home"        # for a store no workspace names
```

A workspace wears a colour so two boards are told apart at a glance; not said,
one is read off its name. A store no workspace names is shown by its `name`,
or by its folder's name.

## What a store's settings may say

Every table but where things are: `[pages]`, `[archive]`, `[focus]`,
`[approach]`, `[time]`, `[window]`, `[series]`, and `[workspace]`. `store`,
`notes`, `plugin` and `[workspaces]` belong to the global file, and are refused
in a store's settings.

`config show` says beside each key whether it is set in the store's settings,
from the global configuration, or by default — and, for a key set here, what
the global file would give.

## Pages

```toml
[pages]
objectives = false  # tasks and objectives are on unless they say false
time = "hidden"     # "on", "hidden" or "off"
```

A page turned off is gone from every surface: no tab on the boards, none of its
commands, none of its tools for an agent.

The time takes one of three words: `on`, the default; `hidden`, counted and not
shown; `off`, not counted at all — the focus stays, without a timer. See
[time](trace.md#turning-the-time-off).

With every page off, the window offers its settings, the one way back.

## Archiving on its own

```toml
[archive]
start = 2026-10-21   # the day the count starts
days = 14            # how long each period lasts
```

Days are counted from `start` in periods of `days` calendar days, and a task
that went done before the period under way began is archived — when a board
opens, and when tasks are listed. Without `start`, nothing is archived on its
own. Such a task's trace says it was archived `through archive`.

## Tasks that come back

```toml
[series]
on = true
```

Off unless it says `true`. See [series](series.md).

## Focus

```toml
[focus]
session = "1h"   # how long a focus session lasts before it asks
```

A focus left open counts time nobody may be spending. Once a task has been in
focus for this long, cairn asks whether to renew the session or end it:
focusing the task again renews it, and the time runs on. Nothing said, nothing
is asked; nothing is asked while the time is off. Over MCP, the question comes
after every answer.

## An approach agreed first

```toml
[approach]
label = "discuss"
```

A task carrying this label is not coded before you agree on how: an agent
working on it writes `approach.md` in its notes and waits for your agreement.
cairn enforces nothing; the [Claude Code plugin](claude-code.md) follows it.
There is no default. In a store's settings, `label = ""` turns it off for that
store.

## Time

```toml
[time]
deleted = "taken"    # or "kept", the default
weekends = "hidden"  # or "shown", the default
```

- `deleted`: kept, a deleted task's time stays on the timesheet, marked
  deleted; taken, it leaves with the task.
- `weekends`: hidden, the time page shows weeks of five days. Time spent on a
  Saturday still counts in every total.

## The window

```toml
[window]
theme = "gruvbox-dark"  # a scheme cairn carries, or one of your own
keep_sizes = true       # open as wide and as tall as it was left
ghost = 65              # how much of a dragged card shows, 20 to 90 per cent
intro = "ripple"        # how the board arrives when the window opens
```

**theme** — nine schemes are carried: `default`, `solarized-dark`,
`solarized-light`, `gruvbox-dark`, `gruvbox-light`, `catppuccin-mocha`,
`catppuccin-latte`, `dracula` and `nord`. A scheme of your own is a base16 file
in `~/.config/cairn/themes/<name>.yaml`. A scheme that cannot be worn leaves
the window in `default`, and `cairn gui` says why. `config show` lists every
scheme it sees, and says when one of your own is hard to read. The terminal
board does not wear it: it uses the terminal's own colours.

**keep_sizes** — the widths of the columns and the height of the band, kept
beside the store. A double-click on a gutter, or Reset sizes, takes them back.

**intro** — `ripple`, `deck`, `spring`, `accordion`, `sweep`, `flap`, `pull`,
`particles`, `dolly`, `blueprint` or `stones`.

Each of them belongs to the store, so two windows side by side can wear two
schemes.

## The plugin's folder

```toml
plugin = "~/.cairn/claude-plugin"
```

Where `cairn plugin write` writes the Claude Code plugin, and where the MCP
server keeps it at its own version. Not said, it is `claude-plugin` beside the
store. See [Claude Code](claude-code.md).

## From v0.11.0

A configuration of v0.11.0 named `store` and `notes` at its top. The first
command of a newer cairn turns them into the workspace `main` on its own, moves
nothing on disk, and keeps the old file beside it as `config.toml.v0.11.0`.
