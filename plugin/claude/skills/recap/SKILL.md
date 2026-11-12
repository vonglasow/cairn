---
name: recap
description: Read back what was done on the cairn board over the last day worked or over a week. Use when the person asks "qu'est-ce que j'ai fait hier ?", "ma journée", "ma semaine", "what did I do yesterday?", "recap my week", or prepares a stand-up. Reads the trace, the timesheet and the objectives, finds the Claude Code sessions no task names, and proposes a task for each.
---

# Reading back a day or a week

cairn filters the facts; you read them and say what they mean. cairn keeps no
report and ranks nothing, and neither do you: say what happened.

## The period

- **The day** is the last day worked: the last day before today with at least
  one act in the trace. Ask for the two weeks before today — `trace_list` with
  `since` fourteen days ago and `until` yesterday — and take the day of the
  last act. Asked on a Monday, that is Friday; a weekend or a day off falls out
  by itself.
- **The week** runs Monday to Sunday, the previous one when asked on a Monday.
- A day or a week the person names is taken as named.

A day is written `YYYY-MM-DD`; `until` takes the whole of its day.

## The facts

For the period, from its first day `since` to its last `until`:

1. `trace_list` — every act: columns changed, focus taken and let go, steps
   ticked, tasks added and deleted. `by` says who; `claude session <id>` is a
   Claude Code session. For example `trace_list {"since": "2026-11-09", "until": "2026-11-09"}`.
2. `timesheet_show` with the same `since` and `until` — the time per task.
3. For a week: `objective_list {"state": "active"}`, then `objective_report`
   of each with `since` and `until` — the objectives that moved. One that did
   not move is left out.
4. `task_show` of a task the acts name without a title you know.

## What to say

Per task: its title, the column it ended the period in, what moved — put in
`done`, steps ticked, blocked, deleted — and the time per task. Then, for a
week, the objectives that moved. Then the sessions no task names (below).
Short: a table once there are more than a few tasks. What is not in the facts
is not said.

## The Claude Code sessions no task names

Claude Code keeps each session in
`~/.claude/projects/<folder>/<session id>.jsonl`, `<folder>` being the
directory it ran in with every `/` written `-`. cairn never reads them; you do.

1. Take the sessions whose file was written during the period, in every
   folder.
2. Leave out each one a task names: the trace carries `claude session <id>` in
   `by`, or a task's notes hold an entry `*-session-<id>.md`.
3. Leave out each one `sessions-ignored` lists (below).
4. Read enough of each one left to say what it was about: its first messages,
   the files it changed. Propose the task it served — the one whose title,
   notes or folder fits it — or a new task, and say why.
5. Show them in one table: when, where (the folder), what it was about, the
   task proposed, why.

Attach nothing until the person says, line by line. A session attached to a
task gets:

- an entry in the task's notes, `YYYY-MM-DD-session-<id>.md` for the day the
  session ran: `cd <its folder>`, `claude --resume <id>`, a line on what was
  done in it, and last `released <when it ended, UTC>` — without that line the
  hook takes the session for one still running;
- summary.md rewritten, when the session changes where the task stands;
- its time, `timesheet_add` with the task, `on` the day it ran, and `time` from
  its first message to its last, less the silences of more than half an hour.

A session the person says not to propose again — or every session of a
folder — goes in `sessions-ignored`, beside the store (in the folder of
`store.at` that `config_show {}` answers): one line each, the session id or the
folder, a tab, the reason. You keep that file and the person may edit it;
cairn never reads it.
