---
name: cairn
description: Work on a cairn task — a "task cairn", written cairn#123. Use when the person names one ("on travaille sur cairn#123", "work on cairn#42"), asks what to work on ("sur quoi je travaille ?", "what should I work on?"), asks to load a task's ticket, or stops working on it. Ties the session to the task, loads its notes as context, and writes back where it stands.
---

# A session tied to a cairn task

## Words

- **"task" alone** is your own to-do list, never a cairn task.
- **A cairn task** is said "task cairn" and written `cairn#123`. `#123` alone
  will do while no other tracker is in the conversation; as soon as one is,
  write `cairn#123` every time.
- External tickets keep their own notation: `AB#123` (Azure Boards),
  `PROJ-123` (Jira), `owner/repo#123` (GitHub).

The plugin's hook does the mechanical part: it writes the session's entry in
the task's notes when you take the focus, holds the focus while the person
works, and lets it go when they stop. Do not write a session block in the
task's description, and do not manage the focus by hand beyond what follows.

The cairn server answers from one workspace, the one it started on; the hook
names it at the session's start, and `workspace_show {}` says it again.
Switching is the person's: another workspace is another server, which they
start.

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

## Gathering a task's resources — only when asked

Everything in the task's notes directory serves the task, whatever the file
is — a PDF, an export, a note written in another tool. A file moved out of
it no longer does. The names above (`summary.md`, dated entries,
`ticket.md`) are conventions of this skill, not rules of cairn.

When the person asks for everything tied to a task ("tout ce qui concerne
cairn#42"), list its directory, then look in the folders they point you to
for notes carrying, at their top:

```markdown
---
cairn: 42
---
```

cairn reads no file and finds none of these: you do. Say which ones you
found and where.

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

cairn gives the facts and ranks nothing; the choosing is yours, only when
asked, and nothing you propose is written without the person's yes.

1. List each column on its own, `todo` then `current` — a listing takes one
   column, and `todo,current` is refused — asking only for what you weigh:
   - `task_list {"state": "todo", "fields": ["number", "title", "priority", "blocked_by", "blocks", "deadline", "overdue", "due_in", "age", "objective_late", "estimate", "time_spent"]}`
   - the same with `"state": "current"`.
   `blocks` is how many open tasks it is in the way of (`task_list
   {"blocker": N}` names them), `overdue`/`due_in` how far from its deadline,
   `age` the days in its column, `objective_late` an objective it serves past
   its deadline unreached.
2. Ask the day's preference, in one question: the time there is, what is
   known from outside, the wish of the day — and above all the return, the
   least work for the most result. Do not store the answer anywhere.
3. Set aside what is blocked (`blocked_by` not empty). Weigh the value — the
   priority, what it frees (`blocks`), an objective late, a deadline near or
   passed — against the cost — the estimate, and the time already spent.
   The board's order is the durable preference: between two close ones, the
   one higher in its column comes first.
4. Propose at most five, each as `cairn#<n> — <title>` and one line of why,
   naming the facts it rests on ("frees 3 tasks, 1h estimated").
5. A move — a task up its column, another into `todo` — may be proposed, one
   at a time, each done only on the person's yes (`task_move`). Never reorder
   the board on your own.
6. The person picks one: attach to it as above.

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

## Several tasks in one session

cairn holds several tasks in focus at once, and a session may work through a
batch of them.

- Focus each task when its work starts, and keep it while its work runs —
  several at once when they run together.
- A subagent does not hold the focus: the session that sends it does, and
  focuses the task before sending it.
- Close each task (see below) before moving to the next.
- At the end, close or detach from every task still in focus.

## Closing — a task put in done

When the work is delivered, write before moving the task:

1. a dated entry naming what was delivered — the commit, the branch, pushed
   or not, what was checked — and what was left out;
2. summary.md rewritten to say it is finished, and where the work is;
3. `task_edit` with `state: done`, then `task_unfocus`.

A task closed because it is obsolete, or merged into another, gets the same
entry, saying which: the task that carries the work now, or why none does.

## Abandoning — a task given up

A task abandoned is deleted, not put in `done`: `done` says the work was
delivered, and would tick its open steps. Before deleting it:

1. a dated entry saying it is abandoned and why, and what to start from if the
   need comes back;
2. summary.md rewritten to say it was abandoned;
3. `task_unfocus` if it is in focus, then `task_delete`.

The notes stay where they are: cairn deletes the task, never its notes, and
the trace keeps what the task held.

## A decision that changes the task

A task that asked a question ("decide between A, B and C") and got its
answer is now the work decided:

- the title becomes the work decided ("Implement B");
- the description is rewritten to say what the task is now;
- the question asked and the decision go in a dated entry, never in the
  description;
- labels and blockers follow the decision: a label meaning "waiting for a
  decision" comes off, a task the decision put in the way becomes a blocker.

A description citing something that moved — a commit hash after a history
rewrite, a branch renamed — is fixed with `task_edit` `replace` and `by`, not
written again whole.

## Agreeing on the approach — only when a label asks for it

`config_show` (`cairn config show`) may name an `[approach] label`. Without
one, say nothing of it and work as usual.

With one — say `discuss` — a task carrying that label is not coded before
the person agrees on how:

1. write `approach.md` in its notes: the solution, the alternatives set aside
   and why, the scenarios, what is out of scope, the open questions, and a
   mockup when anything on screen changes;
2. present it, and write no code before the person agrees;
3. on agreement, a dated entry records it and the label comes off.

The person puts the label on. You may propose it when a task leaves real
choices open; never put it on yourself.

## The command line

When the MCP server answers an error, or for a long batch, drive `cairn`
from the shell instead, and say so. What it does that an agent would not
expect:

- On `task edit`, `--label x` adds one; `--label -x` takes one off;
  `--label =x` replaces them all — the same for `--blocker`, `--link` and
  `--objective`. The answer says what was added and taken off: read it.
- In zsh, quote an `=`: `--label '=x'`.
- `--replace "old" --by "new"` changes one piece of a description, which must
  be there exactly once; `--desc` replaces it whole.
- `task move N --after M` takes N into M's column when it is elsewhere.
- `--json` gives what the MCP server would have answered.

## Done does not reopen

A task done whose work does not hold is not reopened: open a new task linked
to it. `task_new` with `link` (`cairn task new "…" --link <n>`), put the label
`bug` on it by hand, and the task done stays done. Moving a task out of `done`
is only for one put there by mistake.

## When cairn itself misbehaves

When the person says cairn did something wrong — a board drawn oddly, a key
that does nothing, an answer that cannot be right — do not guess at the
cause from here. Ask them to run what went wrong again with `--debug`
(`cairn gui --debug`, or `CAIRN_DEBUG=1` in front), to do the one thing that goes wrong,
and then `debug_report` (or `cairn debug report`) packs what was written
down into one archive; say where it is. Nothing a person wrote is in it.

## Detaching — the subject changes, or the person stops

1. Update `summary.md` in the task's notes: the decisions taken, where the
   task stands, the next step. Rewrite it whole; keep it short.
2. `task_unfocus` with the number. The hook closes the session's link.
