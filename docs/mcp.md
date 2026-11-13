# The MCP server

`cairn mcp serve` lets an agent work on your board over MCP. A client — Claude
Code, Claude Desktop, any MCP client — starts it as a child process and speaks
to it over standard input and output. No port is ever opened.

```json
{
  "cairn": {
    "command": "cairn",
    "args": ["mcp", "serve", "--access", "full"],
    "env": { "CAIRN_WORKSPACE": "work" }
  }
}
```

With Claude Code, the [plugin](claude-code.md) declares it for you.

## What an agent may do

What an agent may reach is settled when the server starts, by one flag:

| `--access` | what it reaches |
|---|---|
| `read` (the default) | listing and showing, nothing else |
| `write` | also adding and changing, exports and imports, a backup made, a migration applied |
| `full` | also what can lose work: deleting a task, an objective, a backup or a series, restoring a backup, rolling a migration back |

A tool above the level is not offered at all, rather than offered and refused.
Anything past looking is a level somebody typed: an agent can be talked into a
write by what it reads.

Each tool also says what calling it costs, so a client can ask you first: what
only looks is marked read-only; what can lose work or replace what was written
is marked destructive; what a second call would not change again is marked
idempotent; and every tool says it reaches nothing outside cairn.

## One tool per command

Every command is a tool, under the name the command has: `task_new`,
`objective_show`, `migration_rollback`. Somebody who knows the command line
knows the tools, and a command added to cairn is a tool the same day. The
reference on this site lists every tool with its arguments.

A tool answers twice over: the text a person would have read at the terminal,
and the very shape `--json` prints. A refusal is an answer too — an empty
title, a state that is not one — with its reason, so an agent can correct
itself; its kind and a reason that does not change when the sentence does are
under `_meta`.

## Keeping answers short

An agent pays for every line it is handed, so a listing can be cut down:

- `fields` — only the fields named on each row: `["number","state","title"]`;
- `number` — only the tasks named, with `unknown` for numbers no task carries;
- `limit` and `offset` — a page, with `total` for how many there are;
- `count` — `"state"` or `"label"`, only how many in each;
- `any` on `task_search` — tasks carrying any of the words, with `matched`
  saying which word found which task.

A listing never carries descriptions, only whether a task has one; `task_show`
gives it. `task_edit` and `objective_edit` take `replace` and `by` to change
one piece of a description in one call. See [listings](listing.md).

## Which board it answers from

A server answers from one workspace for as long as it runs: the one
`--workspace` or `--store` names when it starts, or else the one
`CAIRN_WORKSPACE`, `CAIRN_STORE` or the configuration named then. Changing the
default workspace while it runs does not move it, so an agent never ends up
writing on another board without being told. `workspace_show` says which
workspace it answers from, and why. Another workspace is another server.

## While it runs

- **An upgrade takes over.** When the `cairn` on the `PATH` is replaced, the
  next call is answered by the new one, its tools included, with no reconnect.
- **The configuration is read at every call.** A page turned on or off, or
  recurrence switched, changes the tools offered at the next call.
- **The store is opened for each call** and closed again, so the server
  shares it with the boards and the command line.
- **The greeting tells an agent where its work goes**: what was done or decided
  on a task goes in its notes, never in its description. See
  [A task's notes](context.md).
- **A focus session past its length** adds, after every answer, the question
  whether to renew it or end it; see [configuration](config.md#focus).

What an agent does is written in each task's trace under the client's name —
`claude-code` — unless `CAIRN_AUTHOR` is set for the server.

## The shape of an answer

An agent relies on the shape of an answer as a script relies on a column:
adding a field breaks nothing, while renaming or removing one is a breaking
change, said as such in the changelog. A field that has
nothing to say is there, empty. A list of task or objective numbers takes them
written as numbers or as strings.
