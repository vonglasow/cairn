# The trace, and the time it took

Every change to a task is written down, and how long the work took is read
back from it.

## What happened to a task

```sh
cairn trace show 3               # everything ever done to task 3
cairn trace list                 # what was done today, across every task
cairn trace list --since 2026-08-20
cairn trace list --by sam        # what one author did
```

Each entry says what changed, from what to what, when, and who did it through
which surface — `(by sam, through gui)`. A trace reads as a sentence:

```
2026-08-28 17:50  column    backlog → blocked
2026-08-28 17:50  blockers  — → #1
2026-08-28 17:50  because   — → the client's answer
```

Written down: a task's creation, with every attribute it started with; every
attribute that changed — title, description, column, priority, labels,
blockers and their reason, links, deadline, estimate, steps, objectives,
archived — with its value before and after; a move; the focus taken, released
and renewed; time added by hand; the timer stopped or started; the deletion,
with everything the task held.

Looking is not an act: `list`, `show` and `search` write nothing. Nothing is
ever rewritten, so a task can be read back as it stood at any past moment.

**Who** is `CAIRN_AUTHOR` when it is set, the user logged in otherwise. An
agent over MCP is named by its client — `claude-code` — and the Claude Code
plugin's hook writes its own acts as `claude session <id>`. **Through** is
`cli`, `tui`, `gui`, `mcp`, `migration`, or `archive` for a task archived on
its own.

`task show` reads one date off the trace: since when the task has stood in
its column — `in todo since 2026-09-26 · 9 days`.

## A deleted task keeps its trace

Deleting a task does not delete its trace, and its time stays on the
timesheet, marked "(deleted)". Numbers are never handed out twice, so a
trace naming task 7 always means that one task. To have a deletion take its
time off the sheet, set `[time] deleted = "taken"`; see
[configuration](config.md#time).

There is no undo, but the deletion keeps everything the task held, so it can be
written again by hand.

`backup restore` puts back an older store, and its trace with it: what
happened after that backup is gone from both.

## How long the work took

```sh
cairn timesheet show                              # today, task by task
cairn timesheet show --since 2026-09-29 --until 2026-10-03
cairn timesheet add 3 1h --on 2026-10-03          # a meeting nobody focused
cairn timesheet add 3 --on 2026-10-03 -- -30m     # too much: take some back
cairn task edit 3 --spent +30m                    # time added by hand, today
```

**Focus is what is measured.** Focus a task when the work starts, unfocus it
when it stops; the time in between is counted. What was done without focus is
added by hand, and a focus left on too long is taken back the same way.

- **A focus stops at midnight.** One left open counts no night; the day it
  crossed says so. Work carried on past midnight is added by hand.
- **Two tasks in focus at once share the minutes they overlap**, so a day
  never comes to more than the clock.
- **Each task's time is to the minute**, a minute begun counted whole, and a
  day is the sum of its tasks. A day nothing was done on reads `0m`, so a week
  reads as a week.
- **No ceiling and no categories.** cairn says the real time, every hour of it.
  Bringing a day to what a sheet accepts is done when the sheet is filled in.

A task's time spent is the sum of its days: the time it was in focus and the
time added by hand.

## Turning the time off

`[pages] time` is `on`, `hidden` or `off`; see
[configuration](config.md#pages).

- **hidden** only stops showing it: time is still counted, and every minute is
  there when it is shown again.
- **off** counts nothing. A task can still be in focus, to say who is on what,
  without a timer. Turned back on, counting starts again from then; nothing is
  caught up, and nothing already counted is ever taken away.
