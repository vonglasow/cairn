---
name: cairn
description: Work on a cairn task — a "task cairn", written cairn#123. Use when the person names one ("on travaille sur cairn#123", "work on cairn#42"), asks what to work on ("sur quoi je travaille ?", "what should I work on?"), asks to load a task's ticket, or stops working on it. Ties the session to the task, loads its notes as context, and writes back where it stands.
---

# A session tied to a cairn task

## Words

- **"task" alone** is your own to-do list, never a cairn task.
- **A cairn task** is said "task cairn" and written `cairn#123`. Write it that
  way in every answer.
- External tickets keep their own notation: `AB#123` (Azure Boards),
  `PROJ-123` (Jira), `owner/repo#123` (GitHub).

The plugin's hook does the mechanical part: it writes the session's entry in
the task's notes when you take the focus, holds the focus while the person
works, and lets it go when they stop. Do not write a session block in the
task's description, and do not manage the focus by hand beyond what follows.

## Attaching — "on travaille sur cairn#123"

1. `task_focus` with the number. The hook ties the session to the task.
2. `task_show` the task. Its `context` is the task's notes directory.
3. Read from that directory, nothing more:
   - `summary.md` — where the task stands;
   - the three most recent dated entries (files named `YYYY-MM-DD-…`, newest
     first), and only the names of the older ones;
   - `ticket.md`, when there is one.
4. A sub-task (its `parent` is not 0): read the parent's `summary.md` too, and
   nothing else of the parent.
5. Say in a few lines where the task stands and what comes next.

A missing file is not an error: a task new to this is a task with no summary
yet.

## The ticket — only when asked

Read the external ticket only when the person asks for it ("charge le
ticket"). Its id is the task's `ado:<id>` label. Read it through whatever
Azure DevOps access there is — its MCP server, or `az boards work-item show
--id <id>` — and rewrite `ticket.md` with its title, state and description.
When that fails, read the copy already there, and say it may be stale. No
label, no access: say so, and carry on without it.

## Choosing — "sur quoi je travaille ?"

1. `task_list` with `state: todo,current`.
2. Keep the first five that are not blocked (`blocked_by` empty), in the order
   the list gives — the board's order is the priority, and you rank nothing.
   A deadline already passed may be pointed out; it moves nothing.
3. Show each as `cairn#<n> — <title>`, with its parent's title for a sub-task.
4. The person picks one: attach to it as above.

## Detaching — the subject changes, or the person stops

1. Update `summary.md` in the task's notes: the decisions taken, where the
   task stands, the next step. Rewrite it whole; keep it short.
2. `task_unfocus` with the number. The hook closes the session's link.

To work on another task cairn, detach from this one first.
