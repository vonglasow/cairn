---
name: status
description: Say where things stand on the cairn board, at any moment. Use when the person asks "où j'en suis ?", "status", "where do things stand?", "what's going on?". The tasks current and in focus with their open steps, what was done since the last day worked, what is left in the version, what waits on the person.
---

# Where things stand

Facts read from cairn, said in a few lines; you rank nothing and propose
nothing unless asked.

1. **In hand** — the tasks in focus and the tasks in `current`, with their open
   steps:
   - `task_list {"focus": true, "fields": ["number", "title", "state"]}`
   - `task_list {"state": "current", "fields": ["number", "title", "blocked_by"]}`
   - `step_list` of each, keeping the steps not done.
2. **Done since** — `trace_list` from the last day worked to today: the last
   day before today with at least one act in the trace, as the recap skill
   finds it. Say the tasks put in `done` and the steps ticked, nothing more.
3. **What is left in the version** — the version is the label `vX.Y.Z` of the
   next one, the smallest such label still on a task not done
   (`label_list {}`). `task_list` with that label and `"count": "state"` says
   how many sit in each column; the ones not in `done` are what is left.
4. **What waits on the person** — `task_list {"state": "blocked", "fields": ["number", "title", "because"]}`,
   whatever its `because` says the person must do; `task_list {"state": "watching", "fields": ["number", "title"]}`;
   a focus session asking to be renewed; an `approach.md` not yet agreed, when
   the configuration names an `[approach] label`.

Say each part in a line or a short list, and leave out a part with nothing in
it.
