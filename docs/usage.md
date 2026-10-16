# Using cairn

Everything cairn does, by what you do with it. The [README](../README.md) says
how to install it and start; `cairn <thing> <verb> --help` says the rest of any
command, with examples.

## The first five minutes

```sh
make build && export PATH="$PWD/bin:$PATH"   # needs Go 1.26 — see "Build" below
cairn task new "Write the spec"                # a task, numbered 1, in the backlog
cairn task edit 1 --state todo --priority high # what it carries changes by name
cairn task focus 1                             # what you are working on now
cairn task list                                # the board, column by column
cairn tui                                      # the same board, in the terminal
```

Three ideas carry the rest. **A task stands in a column** — `backlog`, `todo`,
`current`, `done`, and three for work that is stopped — and inside a column
the order is the one you set. **An objective is what work is for**: a task
serves it, and it counts how far along its tasks are. **Nothing done is
forgotten**: every change is written down, so `cairn trace show 1` says what
happened to a task and `cairn timesheet show` how long the day's work took.
`cairn <thing> <verb> --help` says the rest, with examples.

## Everything a task carries

```sh
cairn task new "Write the spec" --desc "the one that unblocks the rest" --state todo --priority high --label spec --label team --deadline 2026-09-01 --estimate 90m
cairn task edit 1 --title "Write the other spec" --desc "" --state current --priority critical --deadline 2026-09-15 --estimate 2h
cairn task edit 1 --label +urgent --label -spec
cairn task edit 5 --state blocked --blocker 3 --because "the client's answer"
cairn task edit 5 --blocker -3   # the last thing out of its way puts it in todo
cairn task new "Read the spec" --label spec,team
cairn task new "Ship it" --steps - < steps.txt   # one line a step: [ ] to do, [x] done
cairn task new "Write the docs" --objective 1
cairn task new "Write the spec" --objective 1 --objective 2   # it serves both
cairn task edit 1 --objective +3   # one more, without naming the others
cairn task list      # each row says its column, priority, deadline and labels
                     # a task's steps are not listed: task show names them
cairn task list --state todo --priority high   # only what carries all of it
cairn task list --label team --label urgent    # both labels, not either
cairn task list --objective 1                  # only what serves it
cairn task search spec    # looks in the title, description, steps, deadline,
                          # estimate, time spent — not in what --state and friends narrow by
cairn task show 1
cairn task delete 1 --yes
cairn task delete 1 --also 4 --yes   # several in one act: all of them, or none
cairn task move 3 --first        # to the head of the ones it sits with
cairn task move 3 --after 1      # directly behind another
cairn task move 3 --last
cairn task focus 3               # this is what I am working on
cairn task focus                 # which ones are those again?
cairn task unfocus 3             # done with that one
cairn task unfocus               # done with all of them
cairn task archive 1             # archived, kept whole
cairn task archive --state done --yes  # the whole column in one go
cairn task label sprint-12 --state done  # mark a sprint before sweeping it
cairn task list --archived       # only what was taken off
cairn task unarchive 1           # back on the board
```

| option | what it takes |
|---|---|
| `--title` | on `edit` only; `new` takes the title as its argument |
| `--desc` | any text |
| `--state` | `backlog` `todo` `current` `blocked` `watching` `paused` `done` — a new task starts in `backlog` |
| `--blocker` | the number of a task in its way; `+3` adds one, `-3` takes one away |
| `--because` | what is in its way and is not a task cairn holds |
| `--priority` | `low` `normal` `high` `critical` — a new task carries `normal` |
| `--label` | one per flag or several at once: `--label spec --label team` and `--label spec,team` are the same thing. Lower-cased, deduplicated, sorted. Given plainly it replaces the whole set; on `edit`, `+urgent` adds one and `-spec` takes one away, leaving the rest alone. Each piece of a list counts on its own, so the two forms cannot be mixed there either |
| `--deadline` | a day: `2026-09-01` |
| `--estimate` | a duration: `90m`, `1h30m`, `2h` — read back as `1h30m`. Kept to the second, so anything finer is refused rather than rounded away |
| `--spent` | the time spent, which is the time the task was in focus and the time added by hand. `+30m` adds time by hand today, `-15m` takes some back, and given plainly it sets the whole, the difference added today. It never falls below nothing, and a minute begun counts whole |
| `--steps` | the steps the task is split into, one a line: `[ ] its title` for one to do, `[x] its title` for one done, in their order; `-` reads them from standard input. On `edit` they replace the steps it had, and given empty it keeps none |
| `--objective` | the number of an objective this task serves; `+4` adds one, `-1` takes one away. Given empty, it sets the task loose |

A title is one line: a line break or a tab in it is refused, since neither survives a listing or a card. A description is several lines if it likes, but nothing a terminal would read as an instruction — an escape, a bell — gets in, and a label is one line like a title. A description too long to type is read from standard input, named as `-`: `cairn task new "Write the spec" --desc - < notes.md`. Given empty, an option clears what was there. `task show` says since when a task has stood in its column — `in todo since 2026-09-26 · 9 days`, `in_column_since` to a machine — read off what was done to it. Deleting a task
deletes its steps with it. Add `--json` to any command to be answered as a
machine — including when it refuses.

## Steps

A task is split into steps: a title each, done or not, in their order. A step
is not a task — no number on the board, no column, no focus, no time, no notes —
and it is named by its place in the task, the first being 1:

```sh
cairn step new 1 "Write the spec"           # behind the steps already there
cairn step new 1 "Plan it" --done           # already done
cairn step list 1                           # each with its place, done or not
cairn step edit 1 2 --done                  # ticked; --not-done unticks it
cairn step edit 1 2 --title "Plan the week" # renamed
cairn step delete 1 2                       # the ones behind move up a place
```

`task show` names a task's steps and says how many are done. Putting a task in
done while one of its steps is open is refused, naming the steps; `--yes` on
`task edit` or `task move` puts it there and ticks them in the same act. See
[docs/steps.md](steps.md).

## The order, and what you are working on

There is one list. The column gives the coarse order — `backlog`, then `todo`,
then `current`, then `done` — and inside a column the order is the one you set.
A new task takes the last place in its column, because arriving is not deciding,
and a task given another column lands at the end of its new section.

`task move` reorders inside a column and never changes it: changing column is an
edit, and one act with two meanings is an act nobody can undo confidently. So
moving a task behind one in another column is refused. A task's steps move
with it, in their own order.

`task archive` archives a task without deleting it: the column it
finished in, its trace and everything it carries stay. It takes a number or a
narrowing, so a finished column goes in one command, and `task label` marks a
set first when a sprint is what you want to sweep. A listing leaves the archived
out until `--archived` or `--with-archived` asks for them; `task show` answers
one either way. On both boards `x` archives or unarchives the task under the
cursor, and `⇧A` shows what was archived, written faintly.

`task focus` says what you are working on. Several at once is fine — one agent
session per task, and they run together — so focusing another keeps the first,
`task unfocus 3` lets go of one, and bare `task unfocus` lets go of them all.
Deleting a task takes it out. The listing marks every one of them. It is not the
head of the list and does not move anything: a day goes ten minutes on one task,
an hour on another and ten minutes back on the first, and none of that is a
reordering.

The place itself is never shown. The order is what there is to see, and a second
number beside the one a task already carries would only mislead. See
[docs/ordering.md](ordering.md) for why it was settled this way.

## Objectives

```sh
cairn objective new "Ship cairn 1.0" --priority high --deadline 2026-12-31
cairn objective list
cairn objective list --state on-hold --label urgent   # only what carries all of it
cairn objective show 1           # what it carries, what is under it, how far along
cairn objective edit 1 --state reached --label +done
cairn objective delete 1 --yes
```

Three of the seven columns mean the work is stopped: `blocked` waits on somebody
else, `watching` on somebody you handed it to, `paused` on your own decision. A
blocked task can say what is in its way — another task, or words for something
cairn does not hold — and nothing leaves that column on its own: taking the last
thing away is what puts it back in `todo`. See
[docs/blocking.md](blocking.md).

A listing narrows by what comes from a closed set — the column, the priority, a
label, the objective served — and a search looks in what you typed yourself: the
title, the description, the deadline, the estimate, the time spent. Neither
reaches what the other does, so each question has one way of being asked. See
[docs/listing.md](listing.md) for what that boundary cost and bought.

An objective takes the same options as a task, minus the ones that mean nothing
to it: no `--state` column, no `--steps`, no `--spent`. Its `--state` is one of
`active` `on-hold` `reached` `dropped`, and a new objective is `active`.

`objective show` works out how far along it is from what was put under it:

```
progress  1 of 2 tasks and 0 of 1 steps are done — the objective is 33% done
```

Tasks and their steps are counted apart — three tasks and four steps is never
seven tasks — while the share is over all of them, since every one has to happen.
The share is never rounded to `100%` while something is left, nor to `0%` once
something is done. `task show` says the same thing about a task's steps.
Nothing under it, nothing said.

A task serves as many objectives as it actually serves, and its steps are
counted towards them through it. A task serving two counts in both,
which is the point — made to serve one, work that advanced two had to pick, and
the objective it did not name then read low. An objective with tasks under it is
not deleted. See [docs/objectives.md](objectives.md) for what serving
several cost and bought.

## What happened, and how the day went

```sh
cairn trace show 3               # everything ever done to a task
cairn trace list                 # what was done today, across every task
cairn trace list --since 2026-08-20
```

Every act on a task is written down: its creation with what it started as, every
attribute that changed with what it went from and to, the focus taken and let
go, a move with where it went, and its deletion. Nothing is ever rewritten, so a
task can be read back as it stood at any past moment.

Looking at a task is not an act. `list`, `show` and `search` add nothing — a
trace records what changed, never what was read.

Every label is seen at once, and one is renamed or taken off everything at once:

```sh
cairn label list                                  # each with how many tasks and objectives carry it
cairn label rename sp26-20 sprint-20              # on every task and objective; onto one there, the two become one
cairn label delete sp26-20 --yes                  # off every task and objective
```

How long the work took is the timesheet's to say:

```sh
cairn timesheet show                              # today, task by task
cairn timesheet show --since 2026-09-29 --until 2026-10-03
cairn timesheet add 3 1h --on 2026-10-03          # a meeting nobody focused
cairn timesheet add 3 --on 2026-10-03 -- -30m     # too much: take some back
```

Focus is what is measured: take it up when the work starts, let it go when it
stops — it moves when it is done, never after the fact. Each day says each task's time, to the minute begun, and the sum
of those — the real time, every hour of it, with no ceiling and no kind of work:
sorting a day out and bringing it down to what a sheet accepts is done when the
sheet is filled in. Two tasks in focus at once share the minutes they overlap,
and a focus left open stops at midnight and says so. See
[docs/trace.md](trace.md).

A trace outlives the task it describes — deleting a task would otherwise
change your past — and an act done by an agent over MCP says so, where one done
at the terminal says nothing. See [docs/trace.md](trace.md) for why it was
settled this way.

## Backups, exports and versions

```sh
cairn backup new                 # a new backup of the store as it stands
cairn backup list                # every backup, with its number and date
cairn backup show 2              # when it was made, and what it holds
                                 # `holds 4 task(s) and 2 objective(s)`
cairn backup delete 2 --yes
cairn backup restore 2 --yes     # restore one; without a number, the most recent

cairn task export ~/tasks.json   # every task and what it serves, never over a file
cairn task import ~/tasks.json   # add the tasks of such a file, as new tasks

cairn migration list             # every change, applied or waiting
cairn migration show 2           # what it does, and what rolling it back does
cairn migration apply            # apply what is waiting; `apply 2` for one
cairn migration rollback --yes   # unmake the last change applied
```

When cairn knows a change a store on disk does not, it says so and stops —
`the store is behind this cairn — run cairn migration apply`. Nothing about the
shape of your store changes unless you asked for it. `apply` and `rollback` set
a backup aside first, and so does `backup restore` before replacing anything, so
picking the wrong one costs a command rather than the work it overwrote.

A backup is read in the shape it was written, never today's: one taken before
objectives existed says it holds none, rather than refusing to be read at all.

## The board

```sh
cairn tui                        # open it
cairn tui --icons unicode        # without a Nerd Font installed
```

The same cairn on a screen that stays open. Four columns — backlog, todo,
current, done — with a band underneath for the three states that mean the work
is not moving: blocked, watching, paused. A second tab for the objectives, and
each of them opens the same board holding only what serves it.

It reads like vim. `hjkl` and the arrows only ever navigate; `i` takes hold of a
task and `⇧hjkl` moves it, its state following wherever it lands. `a` adds, `e`
opens the form, `⏎` opens the detail, `/` narrows the board in place, `:` takes
every operation the command line takes, and `?` answers in every mode. `:q`
leaves.

Two things lose work and both are asked about first: `d`, which deletes, and
leaving a form you have typed into. `⇧X` archives a whole column and
asks as well — nothing is lost, but twelve cards moving on one key is not
something to find out about afterwards.

It stays true while it is open: something added at another terminal, or by an
agent over MCP, turns up on its own — and your cursor stays on the task it was
holding rather than on the place that task used to be.

It needs sixty columns. Below that it says so, rather than drawing something
nobody can read. [docs/tui.md](tui.md) says why every part of it is the
way it is.

## The window

```sh
cairn gui                        # the same board, in a window
```

The same four columns and the same band, at a size a mouse can reach. Click to
choose, double-click to open, drag a card into a column to change its state —
where it lands is its new place as well, so the order is set by putting things
where they go rather than by a command. The wheel scrolls what is under it.

A card says when it is wanted, what the work has cost against what it was
reckoned to cost, and what it serves — on the line that carries its labels, so
none of it needs opening anything.

The keys say what they do: the bar along the bottom writes each one beside what
it does, a key works exactly where the bar offers it, and whatever the pointer
is resting on lights up. `tab` and `shift+tab`
walk the two pages, which are drawn as folders — the one you are on opens onto
the page rather than being merely a different colour.

It needs 1024 by 640. Below that it says so and says what it needs, rather than
drawing something nobody can read. It follows the store the way the terminal
board does: what another terminal or an agent changes turns up on its own.

macOS and Linux. [docs/gui.md](gui.md) says why every part of it is the way
it is, including what was turned down and for what reason.

## Answering an agent

```sh
cairn mcp serve                  # speak MCP over standard input and output
cairn mcp serve --access read    # look, and never write
cairn mcp serve --access full    # also what loses work
```

A client starts it as a child process of its own — no port is opened, ever. What
an agent may reach is settled when the server starts, not per call: `read` looks
and is the default, `write` also adds and changes, `full` also does what loses
work. Anything past looking is a level somebody typed. An access it cannot make
sense of is refused, naming the three it expected.

```json
{
  "cairn": {
    "command": "cairn",
    "args": ["mcp", "serve", "--access", "full"],
    "env": { "CAIRN_STORE": "/Users/…/.cairn/cairn.db" }
  }
}
```

**One tool per command, under the name the command has.** `task_new`,
`objective_show`, `migration_rollback` — the same grammar as the terminal, so
anybody who knows one knows the other. The list is never written down anywhere:
both surfaces are built from one description, so a command added tomorrow is a
tool the same day.

Each tool also says what calling it would cost, so a client can ask a person
first: what only looks is marked read-only; what can lose work, or replace what
was written, is marked destructive; what a second call would not change again is
marked idempotent; and every one says it reaches nothing outside cairn.

**An agent and a script are told the same thing.** A tool answers twice over:
the card a person would have read at the terminal, and the very shape `--json`
prints. A refusal comes back as an answer too — an empty title, a state that is
not one, a store this cairn cannot use — with the reason in it, so an agent can
read it and correct itself, and with its kind and a reason that does not change
when the sentence does, under `_meta`. Only a tool the access level left out
fails as a call, because an absence ends the matter where a refusal invites
another try.

## What cairn answers a script

| code | what it means |
|---|---|
| `0` | it worked |
| `1` | understood, and refused — a title that is empty, a state that is not one — or something nobody foresaw went wrong |
| `2` | not understood at all — a flag that does not exist, a number that is not one |
| `3` | the store could not be used — unreadable, out of date, or nothing may write to it |

Code 3 is the one worth branching on: nothing was wrong with the command, so
the same command may well work once the store is reachable again.

Asked for `--json`, a refusal says the same as a document a machine reads:
`{"error": "task 4: there is no such task", "kind": "refused", "reason":
"no-such-task"}`. The kind is `refused`, `usage`, `store` or `failed`, and the
reason is a word that stays the same when the sentence is reworded — the thing
to branch on. A deletion refused for want of `--yes` names what it would have
taken.

## Where your tasks live

More than one cairn can hold the store at once — a second terminal, a loop, a
server answering an agent. A command waits its turn rather than dropping the
work, and the server is only a command asked more often: it opens the store for
each call and closes it on the way out, like every other.

`~/.cairn/cairn.db`, on every system, with the backups beside it in
`~/.cairn/backups/`. Set `CAIRN_STORE` to put it elsewhere.

A file says it too, if you would rather not carry a variable around:
`~/.config/cairn/config.toml`, the same path on every system, with two keys —
`store` for the tasks and `notes` for what you write beside them. A name cairn
does not know is refused rather than passed over, and a place that is not
absolute is refused naming itself. The environment still wins over the file: it
replaces the answer rather than queueing behind it, and writes nothing down.

```sh
cairn config show                # which places cairn will use, and what named each
```

`task show` says where a task's notes go: `context  ~/.cairn/notes/14`, a
directory named for its number under `notes`. Cairn never makes it, reads it or
looks into it — it says where, and a person or an agent writes there. An agent
keeps its account of the work there, one dated file per entry, rather than in
the description, which an edit rewrites whole.

It answers with the file it read and, for each place, whether the file, the
environment or nothing named it. It is the one command that reports a
configuration it cannot make sense of instead of refusing — you reach for it
precisely when something is not where you expected, so it tells you what is
wrong and warns that everything else will refuse until you fix it.
