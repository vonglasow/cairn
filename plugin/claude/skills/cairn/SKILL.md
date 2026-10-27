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

1. `task_focus` with the number. The hook ties the session to the task, and
   a task waiting in `backlog`, `todo` or `paused` moves into `current` with
   it. Do not move it yourself; a task in `blocked` or `watching` stays where
   it is: say so, and ask whether to move it.
2. `task_show` the task. Its `context` is the task's notes directory.
3. Read from that directory, nothing more:
   - `summary.md` — where the task stands;
   - the three most recent dated entries (files named `YYYY-MM-DD-…`, newest
     first), and only the names of the older ones;
   - `ticket.md`, when there is one.
4. Read its `steps` in the `task_show` answer: what is ticked and what is left.
5. Say in a few lines where the task stands and what comes next.

A missing file is not an error: a task new to this is a task with no summary
yet.

## A focus session that asks to be renewed

An answer may end with "this focus session has been up 2h: renew it by
focusing task N again, or end it by unfocusing task N". Still working on task
N: `task_focus` with N renews it. Not: `task_unfocus` with N. Answer it
before anything else, and do not ask the person unless you cannot tell.

## The ticket — only when asked

Read the external ticket only when the person asks for it ("charge le
ticket"). A task tied to Azure DevOps carries the label `ado`, and each work
item's id as a label of digits alone (`77594`). Read it through whatever
Azure DevOps access there is — its MCP server, or `az boards work-item show
--id <id>` — and rewrite `ticket.md` with its title, state and description.
When that fails, read the copy already there, and say it may be stale. No
label, no access: say so, and carry on without it.

## Choosing — "sur quoi je travaille ?"

1. `task_list` with `state: todo,current`.
2. Keep the first five that are not blocked (`blocked_by` empty), in the order
   the list gives — the board's order is the priority, and you rank nothing.
   A deadline already passed may be pointed out; it moves nothing.
3. Show each as `cairn#<n> — <title>`.
4. The person picks one: attach to it as above.

## Steps — splitting the work

A task is split into steps: a title each, done or not, named by its place in
the task, the first being 1. A step is not a task — it has no number on the
board, no focus, no notes — so never make a task for a piece of the one you
are on: add a step.

- `step_new` (`cairn step new <n> "<title>"`) when the work splits into pieces;
- `step_edit` with `done` (`cairn step edit <n> <place> --done`) as each piece
  is finished, `title` to rename one;
- `step_delete` for a piece that is no longer part of it.

Put a task in `done` only once its steps are ticked, or say which ones were
ticked with it: over MCP, finishing a task ticks its open steps.

## Done does not reopen

A task done whose work does not hold is not reopened: open a new task linked
to it. `task_new` with `link` (`cairn task new "…" --link <n>`), put the label
`bug` on it by hand, and the task done stays done. Moving a task out of `done`
is only for one put there by mistake.

## When cairn itself misbehaves

When the person says cairn did something wrong — a board drawn oddly, a key
that does nothing, an answer that cannot be right — do not guess at the
cause from here. Ask them to run what went wrong again with `CAIRN_DEBUG=1`
in front (`CAIRN_DEBUG=1 cairn gui`), to do the one thing that goes wrong,
and then `debug_report` (or `cairn debug report`) packs what was written
down into one archive; say where it is. Nothing a person wrote is in it.

## Detaching — the subject changes, or the person stops

1. Update `summary.md` in the task's notes: the decisions taken, where the
   task stands, the next step. Rewrite it whole; keep it short.
2. `task_unfocus` with the number. The hook closes the session's link.

To work on another task cairn, detach from this one first.
