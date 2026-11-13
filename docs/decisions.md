# Decisions

What cairn leaves out on purpose, and what it settled one way rather than
another — one line per decision, with the reason, and a link to where the
reasoning is written out in full. A decision is not a wall: it is what was true
when it was made, and a new reason reopens it.

## What a task is, and what it carries

| decided | why, in one line | where |
|---|---|---|
| no urgency field | it says what the priority and the column already say | [ordering.md](../.claude/docs/ordering.md#weighed-and-left-out) |
| a task that comes back is an occurrence of a series, made one at a time; still no work that never finishes | left out until #247, when the same work every day copied by hand was the mechanism nobody wanted; each occurrence reaches `done` like any task, and a meeting with no work in it still belongs to the calendar | [series.md](../.claude/docs/series.md) |
| a series has its own numbers, `↻3`, and every occurrence a new task number, notes and time of its own | a series is named even once its tasks are archived; a task shared by its occurrences would be a task that never reaches `done` | [series.md](../.claude/docs/series.md#a-notion-of-its-own) |
| one occurrence open at a time; an act that leaves none open makes the next at once, and cairn on its own makes one only on its day, never ahead | made ahead, occurrences flood the board, the way Taskwarrior's do; a day passed meanwhile is written missed, never caught up | [series.md](../.claude/docs/series.md#one-open-at-a-time-and-when-the-next-is-made) |
| stopping a series or turning recurrence off never removes a task it made; deleting one asks what becomes of its tasks | which work goes is a deliberate choice, as when an objective is deleted | [series.md](../.claude/docs/series.md#a-series-never-takes-back-what-it-made) |
| recurrence is off unless a store's settings or the file turn it on | a mechanism nobody asked for in a board stays out of its way; dormant series wake up as they were | [config.md](../.claude/docs/config.md#tasks-that-come-back-turned-on-per-store) |
| a series' trace is a table of its own, `series_acts` | a missed day or a stop names no task, and every reader of the tasks' trace reads it by task number | [trace.md](../.claude/docs/trace.md#a-series-keeps-a-trace-of-its-own) |
| an agent handles a series at its access level: lists and shows reading, edits, stops, resumes and skips writing, deletes at `full` | a series is work, and an agent may already make one with `task_new`; a view, by contrast, is how the person looks | [mcp.md](../.claude/docs/mcp.md#what-an-agent-may-do-is-settled-when-the-server-starts) |
| no precedence between two tasks | measured: it constrains without ordering | [ordering.md](../.claude/docs/ordering.md#weighed-and-left-out) |
| how long a task has sat in its column is read off the trace, never stored | the one date nothing else implies, and the trace already holds it | [ordering.md](../.claude/docs/ordering.md#what-the-store-gains) |
| no `unblock` verb | `--blocker -3` does it, and returning to `todo` belongs to the edit | [blocking.md](../.claude/docs/blocking.md#deliberately-not-there-yet) |
| no "what makes a task doable now" | two properties were found and the list plainly does not close | [ordering.md](../.claude/docs/ordering.md#deliberately-not-here-yet) |
| no task hangs under another; a task is split into steps, a title and a tick each | a sub-task was a whole task for what was most often one line to tick, and its parent and sisters ran through every rule and every surface | [steps.md](../.claude/docs/steps.md#why-steps-replaced-sub-tasks) |
| a link between two tasks has no kind and no direction; a task done is never reopened, a linked task is opened | a kind would say again what a label already says, and reopening a task rewrites the period, the time and the story it was finished in | [links.md](../.claude/docs/links.md) |
| no score and no rank computed by cairn; it gives raw facts on each task — what it blocks, its deadline, its age, its objective | a score's top row is a next task, and its inputs are a priority cairn cannot date and an estimate that fails worse than it is unreliable; the facts are what was missing when the backlog was ordered by hand (#253, built in #297) | [ordering.md](../.claude/docs/ordering.md#cairn-gives-the-facts-choosing-is-asked-of-an-agent) |
| choosing what to work on is the cairn skill's, when asked; the day's preference is said, never stored | cairn proposes nothing; an agent may, and every move it proposes is accepted one by one; the board's order stays the person's | [ordering.md](../.claude/docs/ordering.md#cairn-gives-the-facts-choosing-is-asked-of-an-agent) |
| no `task next` is already asked plainly — `--deadline-before`, `--state todo`, `task focus` — and choosing among them is the person's order, which cairn does not override | this page |
| no follow-up day and no person awaited on a task watched | its `because` says who, its deadline says when, and both boards already mark a deadline passed | [blocking.md](../.claude/docs/blocking.md#three-words-for-stopped-not-one) |
| no ticket field; a ticket is named in the title or the description | the search reaches both, and a link is between two tasks, never to a place outside cairn | [links.md](../.claude/docs/links.md) |

## What work is for, and how it is organised

Decided on 2026-11-09, built in #226: until then cairn does none of it.

| decided | why, in one line | where |
|---|---|---|
| objectives on two levels, a third refused | two cover every case asked — home automation in parts, a probation objective in key results — and draw on a board; a project beside the objective was three notions and a task counted twice | [objectives.md](../.claude/docs/objectives.md#two-levels) |
| the names stay Objective › Sub-objective; nothing is renamed | Area › Project reads oddly for an objective that ends, Goal › Key result for an enclosure; the word kept keeps every command, tool and the trace | [objectives.md](../.claude/docs/objectives.md#two-levels) |
| a type is a free label, and a label has a scope — tasks, objectives or both | one label system — colours, rename, delete — rather than a set of types beside it; the scope keeps an objective's picker to labels meant for it | [labels.md](../.claude/docs/labels.md#a-label-is-for-tasks-for-objectives-or-for-both) |
| a task on a sub-objective counts in its parent, worked out, never stored; a parent counts over the union of its subtree, each task once | adding the sub-objectives up counts a task serving two of them twice | [objectives.md](../.claude/docs/objectives.md#two-levels) |
| a task attached to a parent and one of its sub-objectives is refused, and the refusal only refuses | the link adds nothing the parent does not count; moving the task up in the same act writes a set nobody asked for | [objectives.md](../.claude/docs/objectives.md#two-levels) |
| a dependency between objectives is stored, warns and never blocks | read off the tasks' blockers it is not seen at a glance; an objective is not work, and blocking comes the day it proves needed | [objectives.md](../.claude/docs/objectives.md#two-levels) |
| a sub-objective may move to another parent, unless one of its tasks is already attached to the new parent | the move would make the double attachment the rule refuses | [objectives.md](../.claude/docs/objectives.md#two-levels) |
| deleting a parent asks, per sub-objective, delete or keep as level one, and for its tasks, delete or only detach | which work goes with an objective is a deliberate choice; nothing is decided for the person | [objectives.md](../.claude/docs/objectives.md#two-levels) |
| a parent is proposed as reached when every sub-objective is, never put there on its own | its standing is the person's: a parent can hold work of its own beyond its sub-objectives | [objectives.md](../.claude/docs/objectives.md#two-levels) |
| the Objectives tab is a tree beside the chosen objective's detail | a grid of cards cannot show what sits under what | [objectives.md](../.claude/docs/objectives.md#on-screen) |
| a card names, by name, the objective it is directly attached to | a number says nothing to a reader, and the whole path is too long for a card | [objectives.md](../.claude/docs/objectives.md#on-screen) |
| `--objective N` includes the sub-objectives, `--itself` the parent alone | a parent's tasks are its subtree's, as its numbers count them; a listing saying fewer would disagree | [objectives.md](../.claude/docs/objectives.md#two-levels) |

## What is remembered, and what is not

| decided | why, in one line | where |
|---|---|---|
| a deletion keeps the task's history | otherwise the past changes every time something is deleted | [trace.md](../.claude/docs/trace.md#the-trace-outlives-the-task) |
| no undo, and no putting a deleted task back | reading and undoing are two different asks; the history makes it possible to work out, nothing does it | [trace.md](../.claude/docs/trace.md#deliberately-not-here-yet) |
| a deletion keeps what the task held, whole, in its last act, and nothing purges it | the history of a task is the reason the trace exists, and a deletion that kept nothing made even putting it back by hand impossible; a purge — `trace purge`, a task named, `--yes` — is built the day something written by mistake has to go, not before. Until then, what must not be kept is not written | [trace.md](../.claude/docs/trace.md#a-deletion-writes-down-what-went-with-it) |
| time turned off or back on is written as an act on each task in focus, `timer stopped` or `timer started`, never worked out from the mode at read time | a mode read at read time would rewrite past hours each time the key changed; the act fixes what was counted when it was | [trace.md](../.claude/docs/trace.md#time-turned-off-a-focus-without-a-timer) |
| the timesheet gives the hours raw, by day and task | what a good day looks like is a judgement; grouping by objective or label was not asked for | [trace.md](../.claude/docs/trace.md#deliberately-not-here-yet) |
| no stand-up or weekly report in cairn: cairn filters, the plugin's `recap` skill reads | what a day or a week needs is the facts filtered — the trace (`trace list --since --until`), the timesheet, the objectives' reports — which cairn does, and read with judgement, which a skill does; the two are one feature, and a report in cairn would be the judgement in the wrong place | [trace.md](../.claude/docs/trace.md#deliberately-not-here-yet), [claude-code.md](../.claude/docs/claude-code.md#the-skills) |
| the Claude Code sessions are read by the skill, never by cairn | their format is Claude Code's to change; the skill proposes the task of a session no task names, attaches it on the person's word, and keeps `sessions-ignored` beside the store, a file cairn never reads | [claude-code.md](../.claude/docs/claude-code.md#the-skills) |
| a task abandoned is deleted, its reasons in its notes | `done` says the work was delivered and ticks the open steps; archived, it stayed on no board and in every count; the notes and the trace keep what it was | [claude-code.md](../.claude/docs/claude-code.md#the-skills) |
| cairn says where a task's notes go and never looks there | a cairn that read them would compete with the harness it plugs into; counting the notes is looking | [context.md](../.claude/docs/context.md#it-never-looks-so-there-is-nothing-here-that-can-fail) |

## The store

| decided | why, in one line | where |
|---|---|---|
| a store and its notes are chosen together as a workspace, a named folder; a flag on the command replaces even the environment | switching a piece of life was two variables to keep in step, and a variable left in a shell would otherwise win over what was just typed | [config.md](../.claude/docs/config.md#workspaces-each-a-folder-holding-a-store-and-its-notes) |
| a workspace always answers — the flag, the environment, the current directory, the default, then `main` in `~/.cairn` — and its store, notes and settings follow from it; the file names no top-level `store` or `notes` | the shell, the plugin's MCP server and its hook answered from different boards when only a shell set the variables; each starts in the project, so the directory makes them agree | [config.md](../.claude/docs/config.md#the-order-and-why-the-environment-is-not-part-of-it) |
| a configuration of v0.11.0 is turned into workspaces on its own, nothing moved on disk, the old file kept beside it | asking could only be answered yes, and a no would have every surface asking again | [config.md](../.claude/docs/config.md#from-v0110-on-its-own) |
| notes moved by hand in the file are offered to follow, in a terminal only | cairn remembers where it last found them, so it knows both places; with nobody to answer, moving a person's notes is not cairn's to do | [config.md](../.claude/docs/config.md#notes-moved-by-hand) |
| what belongs to a store is said in `settings.toml` beside it and wins key by key; where things are stays in the global file | two boards are run differently — pages, archive, focus — and a setting kept apart from its store does not follow it; the places have to be known before any store is | [config.md](../.claude/docs/config.md#settings-that-belong-to-the-store) |

## The surfaces

| decided | why, in one line | where |
|---|---|---|
| the configuration is set from the window, the terminal and the command line, and stays one file in one place | a key is written as one line of text, every comment kept, and only by the engine; the surfaces are several ways in to the same file, so no earlier decision is reversed: the file is never encoded again from a form, and the editor still works | [config.md](../.claude/docs/config.md#one-configuration-several-ways-in-to-it) |
| a refusal quotes what a person typed exactly as typed, with what a terminal would act on written out | what somebody typed is theirs and is never changed — not even its case, which fang's style used to change — and it reaches a screen shown, never obeyed | the command line's `saidToAPerson` |
| no `--dry-run` on what loses work | without `--yes` it is refused before anything is touched, and the refusal names what would have gone — that is the dry run | `Operation.AsksFirst` in the engine |
| no limit on a listing | a row carries facts and never what a task says, which is what made a listing too long for a client; a limit comes the day a board that large is measured | [listing.md](../.claude/docs/listing.md#a-listing-hands-over-the-facts-and-leaves-what-a-task-says-to-be-shown) |
| the contract with an agent is versioned by its commits until the first tag | a break says `BREAKING CHANGE` in the commit that makes it, and the schemas are pinned so it cannot be made without one; a version number with nothing tagged would number nothing | [mcp.md](../.claude/docs/mcp.md#what-changes-in-a-tools-answer-and-what-does-not) |
| the MCP server looks only, unless told otherwise | an agent can be talked into a write by what it reads | [mcp.md](../.claude/docs/mcp.md#what-an-agent-may-do-is-settled-when-the-server-starts) |
| a tool above the access level is absent, not refused | a refusal invites another try; an absence ends it | [mcp.md](../.claude/docs/mcp.md#what-an-agent-may-do-is-settled-when-the-server-starts) |
| a refusal's kind travels under `_meta`, not `structuredContent` | the client measures structured content against the answer's shape even on a refusal | [mcp.md](../.claude/docs/mcp.md) |

## Sharing

| decided | why, in one line | where |
|---|---|---|
| no decentralised store (IPFS), no CRDT library, no synced store file | public by default and undeletable; every CRDT engine fitting SQLite needs C; a file synced mid-write is a corrupt store | [sharing.md](../.claude/docs/sharing.md#study-1--a-decentralised-store) |
| sharing, when it is needed, is the trace merged over a relay of the same binary | the trace is already a log, and a relay is where a deletion reaches everybody | [sharing.md](../.claude/docs/sharing.md#study-1--a-decentralised-store) |
| no blockchain | colleagues of one employer trust each other, and a chain cannot forget what the GDPR says must be forgettable | [sharing.md](../.claude/docs/sharing.md#study-2--a-blockchain) |
