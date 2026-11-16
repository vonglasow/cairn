# Notes in Obsidian

How to open the notes of a workspace in Obsidian. cairn knows no note tool:
everything below is set up in the configuration and in Obsidian. What a task's
notes are is in [A task's notes](context.md).

## The notes inside the vault

A vault is a folder. Obsidian skips folders whose name starts with a dot, so
the default notes, `~/.cairn/notes` or `<project>/.cairn/notes`, are not
seen from a vault holding them. Name the workspace's notes inside the vault
instead, one of three ways.

**When the workspace is made.** In the window, the **New workspace** dialog
asks for its **Notes** under its folder: type the vault's folder, or
**Choose…** it. In the terminal, <kbd>^N</kbd> in the switcher asks for them after the
folder. On the command line:

```sh
cairn workspace add garden ~/.cairn/garden --notes ~/vaults/home/garden
```

Left empty, the notes stay in the workspace's folder.

**Later.** In the window's settings, **This workspace**, the **Notes folder**
row under **Files** takes another folder, typed or chosen; **Use the shared
value** puts them back in the workspace's folder. On the command line, from
inside the workspace:

```sh
cairn config set notes ~/vaults/home/garden
cairn config unset notes     # back in the workspace's folder
```

**By hand**, in `~/.config/cairn/config.toml`
([configuration](config.md#a-workspaces-notes)):

```toml
[workspaces.garden]
folder = "~/.cairn/garden"
notes = "~/vaults/home/garden"
```

Each way writes that table. A folder the board file is under is refused: the
store stays out of the vault.

The notes already written stay where they were until moved. The window says
so under the row, with **Move them** and **Leave them**; the next run of
cairn in a terminal asks, `[y/N]`; `cairn notes move` moves them. Nothing is
moved into a folder that already holds something, and the MCP server never
moves them ([configuration](config.md#notes-moved-by-hand)).
One workspace's notes are one folder of the vault: several workspaces may
share a vault, each in a folder of its own, so two boards never write into
the same task's folder. The store stays in the workspace's `folder`, outside
the vault: a database synced as a file by a vault's sync is a database two
machines may write at once.

**No symlink.** A link from the vault to `.cairn/notes` was weighed and set
aside: Obsidian supports links of the kind without recommending them
([forum](https://forum.obsidian.md/t/symbolic-links-symlinks-folder-links/1058)),
and a vault copied or synced elsewhere may carry the link and not the files.
Naming the notes where they are read leaves nothing to follow.

## What Obsidian shows

`<notes>/42/` is task 42: its `summary.md`, its dated entries, the session
entries the hook writes, and whatever else was put there, PDFs included.
Obsidian's own `.obsidian/` folder sits at the vault's root, outside every
task's folder, and cairn never looks at it.

A note kept elsewhere in the vault — the meetings of the week, filed by
date — serves a task by carrying `cairn: 42` in its front matter, which
Obsidian shows as a property. An agent asked to gather task 42's resources
looks for it; cairn does not.

## What to leave alone

The task's folder is named by its number and nothing else: renamed, it no
longer serves the task. Files inside it may be renamed and moved between
subfolders freely; moved out, they no longer serve it.
