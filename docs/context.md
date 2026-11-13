# A task's notes

A task carries more than cairn holds: meeting notes, a contract, what a working
session decided. Those live as ordinary files, in a folder per task, and cairn
says where that folder is.

## Where they go

`task show` names it:

```
context  ~/.cairn/notes/14
```

It is a directory named for the task's number under the notes of the workspace
in use — `~/.cairn/notes` unless the configuration names another place; see
[configuration](config.md#where-everything-is). `--json` and the MCP tool
`task_show` give it as `"context"`. The terminal board shows it in a `CONTEXT`
panel of the open task, and the window at the foot of the open task.

cairn never makes that directory, never reads it and never looks to see
whether it is there. It says where, and a person or an agent writes there.
Nothing is stored to link a task to its notes: the folder's name is the link.

## What goes in it

Anything, written by anybody, with any tool: Markdown, a PDF, a plain text
file. Everything in `<notes>/<n>/` serves task `n`, whatever it is called.
Moving a file into the folder attaches it to the task; moving it out detaches
it. Renaming the folder itself detaches everything in it.

Material written before its task exists can wait in a folder of any name;
renaming the folder to the task's number attaches it.

A note that has to live elsewhere — a meeting's notes filed by date — can name
its task in a header at its top:

```markdown
---
cairn: 42
---
```

An agent gathering a task's resources looks for it; cairn does not.

## How an agent keeps its account there

The description says what a task is; what was done or decided goes in its
notes. An agent working on a task — and the cairn skill of the
[Claude Code plugin](claude-code.md) — writes:

- one dated file per piece of work, `YYYY-MM-DD-<subject>.md`, added beside the
  others and never rewritten;
- `summary.md`, rewritten whole, saying where the task stands.

An agent coming back to the task reads `summary.md` and the latest entries, and
only the names of the rest. The plugin's hook adds an entry for each session
that works on a task.

## In a note tool

The folders are ordinary files, so any note tool opens them.
[Obsidian](notes-obsidian.md) says how to point a vault at them.
