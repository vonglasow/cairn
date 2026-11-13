# Decisions

Read this before proposing something cairn does not do. Most of what it leaves
out was weighed and left out on purpose, and the reason is written beside the
rule it belongs to — which is the right place for it and the wrong place to
find it from. This is the index: one line per decision, and where the reason
lives.

A decision here is not a wall. It is what was true when it was made, with the
reason that made it; a new reason reopens it. What it saves is somebody
arriving at the same idea, building half of it, and finding the paragraph that
already said why not.

## What a task is, and what it carries

| decided | why, in one line | where |
|---|---|---|
| no urgency field | it says what the priority and the column already say | [ordering.md](ordering.md#weighed-and-left-out) |
| a task that comes back is an occurrence of a series, made one at a time; still no work that never finishes | left out until #247, when the same work every day copied by hand was the mechanism nobody wanted; each occurrence reaches `done` like any task, and a meeting with no work in it still belongs to the calendar | [series.md](series.md) |
| a series has its own numbers, `↻3`, and every occurrence a new task number, notes and time of its own | a series is named even once its tasks are archived; a task shared by its occurrences would be a task that never reaches `done` | [series.md](series.md#a-notion-of-its-own) |
| one occurrence open at a time; an act that leaves none open makes the next at once, and cairn on its own makes one only on its day, never ahead | made ahead, occurrences flood the board, the way Taskwarrior's do; a day passed meanwhile is written missed, never caught up | [series.md](series.md#one-open-at-a-time-and-when-the-next-is-made) |
| stopping a series or turning recurrence off never removes a task it made; deleting one asks what becomes of its tasks | which work goes is a deliberate choice, as when an objective is deleted | [series.md](series.md#a-series-never-takes-back-what-it-made) |
| recurrence is off unless a store's settings or the file turn it on | a mechanism nobody asked for in a board stays out of its way; dormant series wake up as they were | [config.md](config.md#tasks-that-come-back-turned-on-per-store) |
| the rule grammar is written by hand, rrule-go is not added | the words typed are not RRULE, the subset is small, the library has not moved since 2023 | [dependencies.md](dependencies.md#weighed-and-not-added) |
| a series' trace is a table of its own, `series_acts` | a missed day or a stop names no task, and every reader of the tasks' trace reads it by task number | [trace.md](trace.md#a-series-keeps-a-trace-of-its-own) |
| an agent handles a series at its access level: lists and shows reading, edits, stops, resumes and skips writing, deletes at `full` | a series is work, and an agent may already make one with `task_new`; a view, by contrast, is how the person looks | [mcp.md](mcp.md#what-an-agent-may-do-is-settled-when-the-server-starts) |
| no precedence between two tasks | measured: it constrains without ordering | [ordering.md](ordering.md#weighed-and-left-out) |
| how long a task has sat in its column is read off the trace, never stored | the one date nothing else implies, and the trace already holds it | [ordering.md](ordering.md#what-the-store-gains) |
| no `unblock` verb | `--blocker -3` does it, and returning to `todo` belongs to the edit | [blocking.md](blocking.md#deliberately-not-there-yet) |
| no "what makes a task doable now" | two properties were found and the list plainly does not close | [ordering.md](ordering.md#deliberately-not-here-yet) |
| no task hangs under another; a task is split into steps, a title and a tick each | a sub-task was a whole task for what was most often one line to tick, and its parent and sisters ran through every rule and every surface | [steps.md](steps.md#why-steps-replaced-sub-tasks) |
| a link between two tasks has no kind and no direction; a task done is never reopened, a linked task is opened | a kind would say again what a label already says, and reopening a task rewrites the period, the time and the story it was finished in | [links.md](links.md) |
| no score and no rank computed by cairn; it gives raw facts on each task — what it blocks, its deadline, its age, its objective | a score's top row is a next task, and its inputs are a priority cairn cannot date and an estimate that fails worse than it is unreliable; the facts are what was missing when the backlog was ordered by hand (#253, built in #297) | [ordering.md](ordering.md#cairn-gives-the-facts-choosing-is-asked-of-an-agent) |
| choosing what to work on is the cairn skill's, when asked; the day's preference is said, never stored | cairn proposes nothing; an agent may, and every move it proposes is accepted one by one; the board's order stays the person's | [ordering.md](ordering.md#cairn-gives-the-facts-choosing-is-asked-of-an-agent) |
| a study and what it decides to build land in the same version | #222 and #226 landed a version apart, and a version shipped a decision with nothing of it built | [workflow.md](workflow.md#a-study-and-its-build) |
| no `task next` is already asked plainly — `--deadline-before`, `--state todo`, `task focus` — and choosing among them is the person's order, which cairn does not override | this page |
| no follow-up day and no person awaited on a task watched | its `because` says who, its deadline says when, and both boards already mark a deadline passed | [blocking.md](blocking.md#three-words-for-stopped-not-one) |
| no ticket field; a ticket is named in the title or the description | the search reaches both, and a link is between two tasks, never to a place outside cairn | [links.md](links.md) |

## What work is for, and how it is organised

Decided on 2026-11-09, built in #226: until then cairn does none of it.

| decided | why, in one line | where |
|---|---|---|
| objectives on two levels, a third refused | two cover every case asked — home automation in parts, a probation objective in key results — and draw on a board; a project beside the objective was three notions and a task counted twice | [objectives.md](objectives.md#two-levels) |
| the names stay Objective › Sub-objective; nothing is renamed | Area › Project reads oddly for an objective that ends, Goal › Key result for an enclosure; the word kept keeps every command, tool and the trace | [objectives.md](objectives.md#two-levels) |
| a type is a free label, and a label has a scope — tasks, objectives or both | one label system — colours, rename, delete — rather than a set of types beside it; the scope keeps an objective's picker to labels meant for it | [labels.md](labels.md#a-label-is-for-tasks-for-objectives-or-for-both) |
| a task on a sub-objective counts in its parent, worked out, never stored; a parent counts over the union of its subtree, each task once | adding the sub-objectives up counts a task serving two of them twice | [objectives.md](objectives.md#two-levels) |
| a task attached to a parent and one of its sub-objectives is refused, and the refusal only refuses | the link adds nothing the parent does not count; moving the task up in the same act writes a set nobody asked for | [objectives.md](objectives.md#two-levels) |
| a dependency between objectives is stored, warns and never blocks | read off the tasks' blockers it is not seen at a glance; an objective is not work, and blocking comes the day it proves needed | [objectives.md](objectives.md#two-levels) |
| a sub-objective may move to another parent, unless one of its tasks is already attached to the new parent | the move would make the double attachment the rule refuses | [objectives.md](objectives.md#two-levels) |
| deleting a parent asks, per sub-objective, delete or keep as level one, and for its tasks, delete or only detach | which work goes with an objective is a deliberate choice; nothing is decided for the person | [objectives.md](objectives.md#two-levels) |
| a parent is proposed as reached when every sub-objective is, never put there on its own | its standing is the person's: a parent can hold work of its own beyond its sub-objectives | [objectives.md](objectives.md#two-levels) |
| the Objectives tab is a tree beside the chosen objective's detail | a grid of cards cannot show what sits under what | [objectives.md](objectives.md#on-screen) |
| a card names, by name, the objective it is directly attached to | a number says nothing to a reader, and the whole path is too long for a card | [objectives.md](objectives.md#on-screen) |
| `--objective N` includes the sub-objectives, `--itself` the parent alone | a parent's tasks are its subtree's, as its numbers count them; a listing saying fewer would disagree | [objectives.md](objectives.md#two-levels) |

## What is remembered, and what is not

| decided | why, in one line | where |
|---|---|---|
| a deletion keeps the task's history | otherwise the past changes every time something is deleted | [trace.md](trace.md#the-trace-outlives-the-task) |
| no undo, and no putting a deleted task back | reading and undoing are two different asks; the history makes it possible to work out, nothing does it | [trace.md](trace.md#deliberately-not-here-yet) |
| a deletion keeps what the task held, whole, in its last act, and nothing purges it | the history of a task is the reason the trace exists, and a deletion that kept nothing made even putting it back by hand impossible; a purge — `trace purge`, a task named, `--yes` — is built the day something written by mistake has to go, not before. Until then, what must not be kept is not written | [trace.md](trace.md#a-deletion-writes-down-what-went-with-it) |
| time turned off or back on is written as an act on each task in focus, `timer stopped` or `timer started`, never worked out from the mode at read time | a mode read at read time would rewrite past hours each time the key changed; the act fixes what was counted when it was | [trace.md](trace.md#time-turned-off-a-focus-without-a-timer) |
| the timesheet gives the hours raw, by day and task | what a good day looks like is a judgement; grouping by objective or label was not asked for | [trace.md](trace.md#deliberately-not-here-yet) |
| no stand-up or weekly report in cairn: cairn filters, the plugin's `recap` skill reads | what a day or a week needs is the facts filtered — the trace (`trace list --since --until`), the timesheet, the objectives' reports — which cairn does, and read with judgement, which a skill does; the two are one feature, and a report in cairn would be the judgement in the wrong place | [trace.md](trace.md#deliberately-not-here-yet), [claude-code.md](claude-code.md#the-skills) |
| the Claude Code sessions are read by the skill, never by cairn | their format is Claude Code's to change; the skill proposes the task of a session no task names, attaches it on the person's word, and keeps `sessions-ignored` beside the store, a file cairn never reads | [claude-code.md](claude-code.md#the-skills) |
| a task abandoned is deleted, its reasons in its notes | `done` says the work was delivered and ticks the open steps; archived, it stayed on no board and in every count; the notes and the trace keep what it was | [claude-code.md](claude-code.md#the-skills) |
| the plugin is installed from GitHub at the binary's tag; cairn says the commands, never runs them | the plugin then is the one the version was made with, and Claude Code versions it by its manifest, which the release writes; Claude Code's own commands are the person's to type | [claude-code.md](claude-code.md#where-it-lives-and-how-it-is-installed), [release.md](release.md) |
| cairn says where a task's notes go and never looks there | a cairn that read them would compete with the harness it plugs into; counting the notes is looking | [context.md](context.md#it-never-looks-so-there-is-nothing-here-that-can-fail) |

## The store

| decided | why, in one line | where |
|---|---|---|
| SQLite, in Go, one file | one static binary, no cgo for the store | [dependencies.md](dependencies.md) |
| every write is one transaction, every migration one per change | an act, or a change of shape, half kept is worse than none | [architecture.md](architecture.md) |
| integrity between tables is kept by the code, not by foreign keys | SQLite leaves them off per connection; the declared `ON DELETE` clauses of the first shape are inert, and what refers to a task is removed by hand on a deletion, inside its transaction | `internal/store/store.go`, `Delete` |
| closed sets — a column, a priority — are refused in `domain`, not by a `CHECK` | one place says what a column is, and a `CHECK` would be a second list to keep true | `internal/domain/attributes.go` |
| no index on the tasks table | a store holds hundreds of tasks; an index is added the day a listing is measured slow | this page |
| a change that alters data and not the shape is forgotten, not refused, when rolled back | it has nothing to undo, and refusing it would wall off every older shape behind it; what it leaves as it was is said by its `loses`, and applying it again changes nothing | `internal/store/migrate.go`, `RollBack` |
| a backup is refused as a restore when SQLite cannot read it whole | a damaged backup put in place trades a store that opens for one that does not | `internal/store/backup.go` |
| an engine held open follows a store put back under it, rather than a restore being refused while a board is open | measured: a board's connection stayed on the file the restore replaced, did not see what came back, and wrote where nobody reads, without an error; the engine now sees its file was replaced before every act and every beat, and reopens. A lock held against boards would be a lock a forgotten window holds for ever | `internal/engine/engine.go`, `following` |
| a store and its notes are chosen together as a workspace, a named folder; a flag on the command replaces even the environment | switching a piece of life was two variables to keep in step, and a variable left in a shell would otherwise win over what was just typed | [config.md](config.md#workspaces-each-a-folder-holding-a-store-and-its-notes) |
| a workspace always answers — the flag, the environment, the current directory, the default, then `main` in `~/.cairn` — and its store, notes and settings follow from it; the file names no top-level `store` or `notes` | the shell, the plugin's MCP server and its hook answered from different boards when only a shell set the variables; each starts in the project, so the directory makes them agree | [config.md](config.md#the-order-and-why-the-environment-is-not-part-of-it) |
| a configuration of v0.11.0 is turned into workspaces on its own, nothing moved on disk, the old file kept beside it | asking could only be answered yes, and a no would have every surface asking again | [config.md](config.md#from-v0110-on-its-own) |
| notes moved by hand in the file are offered to follow, in a terminal only | cairn remembers where it last found them, so it knows both places; with nobody to answer, moving a person's notes is not cairn's to do | [config.md](config.md#notes-moved-by-hand) |
| what belongs to a store is said in `settings.toml` beside it and wins key by key; where things are stays in the global file | two boards are run differently — pages, archive, focus — and a setting kept apart from its store does not follow it; the places have to be known before any store is | [config.md](config.md#settings-that-belong-to-the-store) |
| looking at a backup, or at the store before backing it up, opens the file to be read and nothing else | opened the shared way, a store that was not there was made readable by anybody, and a backup looked at was rewritten and its date moved | `internal/store/backup.go`, `heldIn` |

## The surfaces

| decided | why, in one line | where |
|---|---|---|
| every change goes through the engine; a surface decides nothing | the rules live once | [architecture.md](architecture.md) |
| the configuration is set from the window, the terminal and the command line, and stays one file in one place | a key is written as one line of text, every comment kept, and only by the engine; the surfaces are several ways in to the same file, so no earlier decision is reversed: the file is never encoded again from a form, and the editor still works | [config.md](config.md#one-configuration-several-ways-in-to-it) |
| what kind of answer a failure is — refused, not understood, store, unforeseen — is the engine's | a script's exit code and an agent's refusal are read off the same list | `internal/engine/refusal.go` |
| a refusal quotes what a person typed exactly as typed, with what a terminal would act on written out | what somebody typed is theirs and is never changed — not even its case, which fang's style used to change — and it reaches a screen shown, never obeyed | the command line's `saidToAPerson` |
| changes to the shape of a store are made in order, and a rollback reads the ledger again inside its transaction | each change was written on the shape the ones before it leave; two rollbacks read before either ran both said they undid the same change | `internal/store/migrate.go` |
| a database holding tables, and none of a store's, is refused before anything is written into it | a path named by mistake can lead to somebody else's database, and opening it the usual way already writes into it | `internal/store/open.go`, `refuseAnotherDatabase` |
| several tasks or objectives are deleted or archived in one act, under `--also`, all of them or none | three deletions asked one after another left, when the third was refused, two gone and somebody to work out which; one act is one transaction, and a new optional argument breaks no caller | `internal/engine/together.go` |
| a version's number is read off the commits by svu, and its tag is made by `make release` once every check has passed | a number typed by hand is a number somebody misread the history for, and a tag made before the checks says a failed build was a version | [release.md](release.md) |
| a release runs every check but the mutation sweep, which is run on its own and leaves a report | the sweep took the better part of an hour, and a release that waited on it was a release nobody made; every commit still mutates the lines it changed, and the report says what the last sweep found | [checks.md](checks.md#the-report-a-sweep-leaves) |
| a version is handed over as binaries, each built on the system it runs on, and no source archive | the window needs cgo and the system's graphics headers, which a cross-compile cannot reach, so each binary is built where it runs; whoever builds cairn clones it | [release.md](release.md) |
| `--json` is a flag of every command, including those that never answer in JSON | one flag at the root reads the same everywhere; on `completion` or `gui` it changes nothing and costs nothing | `internal/cli/cli.go` |
| no `--dry-run` on what loses work | without `--yes` it is refused before anything is touched, and the refusal names what would have gone — that is the dry run | `Operation.AsksFirst` in the engine |
| no comments in the code but one line per exported symbol, and no prose in the scenarios | a quarter of the code was comments, much of it the history of a line rather than its reason; the reasons live in `docs/` and in the commits | [design-principles.md](design-principles.md#comments) |
| which acts ask first is the engine's, and every surface asks the same ones | the boards asked before sweeping a column and the command line did not: one act, two rules. The engine refuses an act that loses work, or sweeps up what a narrowing reaches, unless it comes confirmed; `--yes`, a board's question and an agent's access are only how each surface says it asked | `Operation.AsksFirst`, `Input.Confirmed` |
| what a terminal would obey, or what formats text rather than shows it — every Unicode Cf character, emoji joiners included — is refused when written and replaced when shown | the category is the rule, since a list of the characters somebody thought of let the rest through, tag characters an agent reads included; the refusal keeps it out from now on; showing it replaced covers what was kept before the refusal existed, or comes back with an older backup | `internal/domain/title.go`, `Prose`, `Shown` |
| no limit on a listing | a row carries facts and never what a task says, which is what made a listing too long for a client; a limit comes the day a board that large is measured | [listing.md](listing.md#a-listing-hands-over-the-facts-and-leaves-what-a-task-says-to-be-shown) |
| the contract with an agent is versioned by its commits until the first tag | a break says `BREAKING CHANGE` in the commit that makes it, and the schemas are pinned so it cannot be made without one; a version number with nothing tagged would number nothing | [mcp.md](mcp.md#what-changes-in-a-tools-answer-and-what-does-not) |
| the MCP server looks only, unless told otherwise | an agent can be talked into a write by what it reads | [mcp.md](mcp.md#what-an-agent-may-do-is-settled-when-the-server-starts) |
| a tool above the access level is absent, not refused | a refusal invites another try; an absence ends it | [mcp.md](mcp.md#what-an-agent-may-do-is-settled-when-the-server-starts) |
| a refusal's kind travels under `_meta`, not `structuredContent` | the client measures structured content against the answer's shape even on a refusal | [mcp.md](mcp.md) |
| a tool whose answer shape cannot be derived is offered without one, rather than stopping the server | a test proves every tool derives one, so this can only happen to code nobody ran | `internal/mcp/tools.go` |
| the Claude sessions holding a task are the hook's, not the engine's | the engine would write and read the session entries in a task's notes, which cairn never looks at; it keeps only what every surface needs, letting go of a focus and asking for a session to be renewed | [claude-code.md](claude-code.md#where-it-lives-and-how-it-is-installed) |
| the two boards share where a card goes, not how the cursor walks | they draw the stalled band differently — a grid of three in the terminal, lanes under the columns in the window — so the same key is a different journey | [architecture.md](architecture.md) |
| what a board reads lives in the engine, not in a package between it and the boards | a surface is an entry point; the package in between grew decisions of its own, and its search drifted from the engine's | [architecture.md](architecture.md) |
| the engine is one package, split into files by notion | one door (`Ask`) and one place for the rules; sub-packages wait until a file cannot be read on its own | [architecture.md](architecture.md) |

## Sharing

| decided | why, in one line | where |
|---|---|---|
| no decentralised store (IPFS), no CRDT library, no synced store file | public by default and undeletable; every CRDT engine fitting SQLite needs C; a file synced mid-write is a corrupt store | [sharing.md](sharing.md#study-1--a-decentralised-store) |
| sharing, when it is needed, is the trace merged over a relay of the same binary | the trace is already a log, and a relay is where a deletion reaches everybody | [sharing.md](sharing.md#study-1--a-decentralised-store) |
| no blockchain | colleagues of one employer trust each other, and a chain cannot forget what the GDPR says must be forgettable | [sharing.md](sharing.md#study-2--a-blockchain) |

## The checks

| decided | why, in one line | where |
|---|---|---|
| what the tools find stays current; what the tools are is pinned | a new rule found something already true; a tool changing its own behaviour found nothing | [checks.md](checks.md#getting-them) |
| no threshold is lowered | a gate that bends is not a gate | [checks.md](checks.md#thresholds) |
| `&&`/`||` and `continue`/`break` mutated in the core only | that is where the rules are; the surfaces gave mostly mutants no test can tell apart | [checks.md](checks.md#thresholds) |
| the surfaces are never mutated; the core is mutated in CI, and no sweep is run by hand before a push or a version | the scenarios judge what is on screen, and a sweep of every package was estimated at forty to fifty hours, the window's for most of it | [checks.md](checks.md#thresholds) |
| packages are mutated one after another | gremlins already uses every core inside a package; two at once only made timeouts | [checks.md](checks.md#running-them-by-hand) |
| Linux is checked by hand, in a container, before a version | it needs Docker, which no hook may assume | [release.md](release.md) |
| the container's image is named by tag and Debian release, not digest; commitlint's own dependencies are not locked | the first is a check run by hand, not a gate, and a digest would be moved at every Go release for nothing anyone notices; the second runs only on a message, locally, and locking npm's tree would mean a lockfile in a Go repository for one hook | `scripts/linux.sh`, `.pre-commit-config.yaml` |
| a form's commands are waited for in the terminal board's tests, all but the cursor's blink, passed over at once | the blink is a timer of 530 ms nobody can wait out, and its function names it before it runs; racing every command against five milliseconds, as before, dropped one a machine busy with the rest of the suite was slow to schedule (cairn#298), and raising that deadline made the suite three times slower | `internal/tui/fake_test.go`, `quickly` |

## Known and not yet answered

What a review found and nobody has decided yet. It is written here so it is
not found a second time as news, and it leaves this list when it is fixed or
decided.

Nothing, at the moment.

## One word for each notion, on every surface

The window, the terminal, the command line, the MCP server and these pages
used to name one notion several ways: a task that comes back was a series, a
recurrence and something that repeats; a day due was a deadline and a due;
the settings were a configuration. Each notion now has one word, the one the
person reads (#292, decided 2026-11-11):

- **series** — the setting reads Series, its key `series.on` (the old
  `[recurrence] on` is still read), the filter `series` on `all`,
  `task list --series` alone keeps every occurrence and `--series 3` those of
  ↻3; the refusal says *series are off*;
- **rule** — the field a series comes back by; `--every` stays the flag;
- **occurrence** — what a series made, opened as one;
- **settings** — every screen; `cairn config` and the file names stay;
- **tab** — what was a page, in the settings' section *Tabs*;
- **trace** — what was the history of a task;
- **column** — the filter pill reads `column: todo`; every filter pill reads
  `name: value`, the view's and the search's as before;
- **state** — an objective's, in its form;
- **deadline** — `--deadline-before`, the filter `deadline`, `deadline` where
  a day was `due`;
- **notes** — what a form called its context;
- **left** — `3 left · 5 done`, never `to do` beside a column TODO;
- **blockers**, **parent** — the form's sections;
- **save** — to keep something written, *Save as a view…*, *Save search*;
  *keep* only when a deletion is refused;
- **new, list, show, edit, delete** — the words the commands' groups say;
- **report** — `objective report`, never a review;
- **details**, **help** — the headings.

- **forget** — `workspace forget`, which was `workspace remove`;
- **description** — `--description` is `--desc` spelled whole;
- **theme** — never a scheme, in what an agent reads too;
- **Series** joins the View menu; outside the Series tab the series' acts
  read *Open the series*, *Skip this one*, *Stop the series*, *Resume the
  series*; in the tab, where the series is said already, *Edit the series*,
  *Skip the next*, *Stop*, *Delete…*;
- **`…`** ends a line only when a choice follows it, and a line opening with
  `+` is written in sentence case: *+ New task serving it*.
