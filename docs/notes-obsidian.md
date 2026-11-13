# A task's notes in Obsidian

Read this to open the notes of a workspace in Obsidian, or as the shape a page
for another note tool follows. cairn knows no tool: everything below is set up
in the configuration and in Obsidian, and nothing in cairn changes for it.
Why cairn stays loosely tied to its notes is in [context.md](context.md).

## The notes inside the vault

A vault is a folder. Obsidian skips folders whose name starts with a dot, so
the default notes, `~/.cairn/notes` or `<project>/.cairn/notes`, are not
seen from a vault holding them. Name the workspace's notes inside the vault
instead, in `~/.config/cairn/config.toml`
([config.md](config.md#a-workspaces-notes-where-it-says)):

```toml
[workspaces.cairn]
folder = "~/code/cairn/.cairn"
notes = "~/vaults/work/cairn"
```

The next run of cairn in a terminal offers to move the notes already there,
`[y/N]`; nothing is moved without that yes, and nothing is moved into a
folder that already holds something ([config.md](config.md#notes-moved-by-hand)).
One workspace's notes are one folder of the vault: several workspaces may
share a vault, each in a folder of its own, so two boards never write into
the same task's folder. The store stays in the workspace's `folder`, outside
the vault: a database synced as a file by a vault's sync is a database two
machines may write at once ([sharing.md](sharing.md)).

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
