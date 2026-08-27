# cairn

A personal objective and task manager. CLI first, then MCP; the terminal and
desktop boards come later, once the engine stands on its own.

```sh
cairn task add "Write the spec" --desc "the one that unblocks the rest" --state todo --priority high --label spec --label team --deadline 2026-09-01 --estimate 90m
cairn task edit 1 --title "Write the other spec" --desc "" --state current --priority critical --deadline 2026-09-15 --estimate 2h
cairn task edit 1 --label +urgent --label -spec
cairn task add "Read the spec" --label spec,team
cairn task add "Write the tests" --parent 1
cairn task add "Write the docs" --objective 1
cairn task list      # sub-tasks are listed under the task they hang from
cairn task show 1
cairn task delete 1 --yes
cairn task move 3 --first        # to the head of the ones it sits with
cairn task move 3 --after 1      # directly behind another
cairn task move 3 --last
cairn task focus 3               # this is what I am working on
cairn task focus                 # which one is that again?
cairn task unfocus
```

| option | what it takes |
|---|---|
| `--title` | on `edit` only; `add` takes the title as its argument |
| `--desc` | any text |
| `--state` | `backlog` `todo` `current` `done` — a new task starts in `backlog` |
| `--priority` | `low` `normal` `high` `critical` — a new task carries `normal` |
| `--label` | one per flag or several at once: `--label spec --label team` and `--label spec,team` are the same thing. Lower-cased, deduplicated, sorted. Given plainly it replaces the whole set; on `edit`, `+urgent` adds one and `-spec` takes one away, leaving the rest alone. Each piece of a list counts on its own, so the two forms cannot be mixed there either |
| `--deadline` | a day: `2026-09-01` |
| `--estimate` | a duration: `90m`, `1h30m`, `2h` — read back as `1h30m`. Kept to the second, so anything finer is refused rather than rounded away |
| `--spent` | the time already spent. Given plainly it sets the tally, `+30m` adds to it, `-15m` takes off, and taking off more than was there leaves nothing. Kept to the second, like `--estimate` |
| `--parent` | the number of the task this one hangs under. One level only: a sub-task carries none of its own. Given empty, it sets the task free |
| `--objective` | the number of the objective this task serves. Given empty, it sets the task loose. A sub-task never carries one |

A title is one line: a line break or a tab in it is refused, since neither survives a listing or a card. Given empty, an option clears what was there. A task carrying sub-tasks is not
deleted while they are there. Add `--json` to any command to be answered as a
machine — including when it refuses.

## The order, and what you are working on

There is one list. The column gives the coarse order — `backlog`, then `todo`,
then `current`, then `done` — and inside a column the order is the one you set.
A new task takes the last place in its column, because arriving is not deciding,
and a task given another column lands at the end of its new section.

`task move` reorders inside a column and never changes it: changing column is an
edit, and one act with two meanings is an act nobody can undo confidently. So
moving a task behind one in another column is refused, and so is moving a
sub-task behind a root task — a sub-task takes its place among its sisters,
where its parent sits carrying the whole block.

`task focus` says which single task you are working on. One at a time, so putting
another there releases the first, and deleting it leaves nothing in focus. The
listing marks it. It is not the head of the list and does not move anything: a
day goes ten minutes on one task, an hour on another and ten minutes back on the
first, and none of that is a reordering.

The place itself is never shown. The order is what there is to see, and a second
number beside the one a task already carries would only mislead. See
[docs/ordering.md](docs/ordering.md) for why it was settled this way.

## Objectives

```sh
cairn objective add "Ship cairn 1.0" --priority high --deadline 2026-12-31
cairn objective list
cairn objective show 1           # what it carries, what is under it, how far along
cairn objective edit 1 --state reached --label +done
cairn objective delete 1 --yes
```

An objective takes the same options as a task, minus the ones that mean nothing
to it: no `--state` column, no `--parent`, no `--spent`. Its `--state` is one of
`active` `on-hold` `reached` `dropped`, and a new objective is `active`.

`objective show` works out how far along it is from what was put under it:

```
progress  1 of 2 tasks and 0 of 1 sub-tasks are done — the objective is 33% done
```

Tasks and sub-tasks are counted apart — three tasks and four sub-tasks is never
seven tasks — while the share is over all of them, since every one has to happen.
The share is never rounded to `100%` while something is left, nor to `0%` once
something is done. `task show` says the same thing about a task's sub-tasks.
Nothing under it, nothing said.

A task serves an objective, or hangs under another task, never both: a sub-task
belongs to the task above it, and is counted towards that task's objective
through it. An objective with tasks under it is not deleted.

## Backups and versions

```sh
cairn backup add                 # backup the store as it stands
cairn backup list                # every backup, with its number and date
cairn backup show 2              # when it was made, and what it holds
                                 # `holds 4 task(s) and 2 objective(s)`
cairn backup delete 2 --yes
cairn backup restore 2 --yes     # put one back; without a number, the most recent

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

## Answering an agent

```sh
cairn mcp serve                  # speak MCP over standard input and output
cairn mcp serve --access read    # look, and never write
cairn mcp serve --access full    # also what loses work
```

A client starts it as a child process of its own — no port is opened, ever. What
an agent may reach is settled when the server starts, not per call: `read` looks,
`write` also adds and changes and is the default, `full` also does what loses
work. An access it cannot make sense of is refused, naming the three it expected.

```json
{
  "cairn": {
    "command": "cairn",
    "args": ["mcp", "serve", "--access", "full"],
    "env": { "CAIRN_STORE": "/Users/…/.cairn/cairn.db" }
  }
}
```

**One tool per command, under the name the command has.** `task_add`,
`objective_show`, `migration_rollback` — the same grammar as the terminal, so
anybody who knows one knows the other. The list is never written down anywhere:
both surfaces are built from one description, so a command added tomorrow is a
tool the same day.

Each tool also says what calling it would cost, so a client can ask a person
first: the eight that only look are marked read-only, the five that can lose
work already done are marked as such, and every one of them says it reaches
nothing outside cairn.

**An agent and a script are told the same thing.** A tool answers twice over:
the card a person would have read at the terminal, and the very shape `--json`
prints. A refusal comes back as an answer too — an empty title, a state that is
not one, a store this cairn cannot use — with the reason in it, so an agent can
read it and correct itself. Only a tool the access level left out fails as a
call, because an absence ends the matter where a refusal invites another try.

## What cairn answers a script

| code | what it means |
|---|---|
| `0` | it worked |
| `1` | understood, and refused — a title that is empty, a state that is not one |
| `2` | not understood at all — a flag that does not exist, a number that is not one |
| `3` | the store could not be used — unreadable, out of date, or nothing may write to it |

Code 3 is the one worth branching on: nothing was wrong with the command, so
the same command may well work once the store is reachable again.

## Where your tasks live

More than one cairn can hold the store at once — a second terminal, a loop, a
server answering an agent. A command waits its turn rather than dropping the
work, and the server is only a command asked more often: it opens the store for
each call and closes it on the way out, like every other.

`~/.cairn/cairn.db`, on every system, with the backups beside it in
`~/.cairn/backups/`. Set `CAIRN_STORE` to put it elsewhere.

## Build

```sh
make build                      # ./bin/cairn
make run ARGS='task list'
```

## Trying it by hand

The store lives in a temporary folder, so nothing here touches your own.

```sh
make build
export CAIRN_STORE=$(mktemp -d)/cairn.db
./bin/cairn objective add "Sortir cairn 1.0" --desc "la première vraie" --priority high --label perso --label spec --deadline 2026-12-31 --estimate 40h
./bin/cairn objective list
./bin/cairn objective show 1
./bin/cairn task add "Écrire la spéc" --desc "la toute première" --state todo --priority high --label Spec --label team --deadline 2026-09-01 --estimate 90m
./bin/cairn task add "Relire" --parent 1
./bin/cairn task add "Publier" --parent 1
./bin/cairn task add "Autre chose"
./bin/cairn task list
./bin/cairn task show 1
./bin/cairn task show 2
./bin/cairn task show 1 --json
./bin/cairn task list --json
./bin/cairn task edit 1 --state current
./bin/cairn task edit 1 --label +urgent
./bin/cairn task edit 1 --label -spec
./bin/cairn task edit 1 --label "+draft,+wip"
./bin/cairn task add "Liste en un flag" --label a,b,c
./bin/cairn task edit 1 --spent 2h
./bin/cairn task edit 1 --spent +30m
./bin/cairn task edit 1 --spent -15m
./bin/cairn task edit 1 --desc ""
./bin/cairn task show 1
./bin/cairn task edit 1 --objective 1
./bin/cairn task edit 4 --objective 1
./bin/cairn objective show 1
./bin/cairn objective show 1 --json
./bin/cairn task edit 4 --state done
./bin/cairn objective show 1
./bin/cairn objective edit 1 --state on-hold
./bin/cairn objective edit 1 --label +urgent
./bin/cairn objective edit 1 --label -spec
./bin/cairn objective add "Apprendre Go" --state on-hold
./bin/cairn objective list
./bin/cairn objective show 2
./bin/cairn objective delete 2 --yes
./bin/cairn objective delete 1 --yes
./bin/cairn objective add ""
./bin/cairn objective show 99
./bin/cairn objective add "z" --state todo
./bin/cairn task edit 2 --objective 1
./bin/cairn task add "w" --parent 1 --objective 1
./bin/cairn objective show spec
./bin/cairn objective
./bin/cairn backup add
./bin/cairn backup list
./bin/cairn backup show 1
./bin/cairn backup list --json
./bin/cairn task delete 4 --yes
./bin/cairn task delete 1 --yes
./bin/cairn task list
./bin/cairn backup restore 1 --yes
./bin/cairn task list
./bin/cairn migration list
./bin/cairn migration show 1
./bin/cairn migration show 2
./bin/cairn migration rollback --yes
./bin/cairn task list
./bin/cairn migration apply
./bin/cairn task list
./bin/cairn task add ""
./bin/cairn task show 99
./bin/cairn task edit 1
./bin/cairn task add "y" --priority urgent
./bin/cairn task show spec
./bin/cairn task add "x" --colour red
./bin/cairn task add "y" --estimate 500ms
./bin/cairn task edit 1 --label "+urgent,spec"
./bin/cairn task add
./bin/cairn task addd "z"
```

## Working on it

[CLAUDE.md](CLAUDE.md) carries the rules, `docs/` the detail. `prek` runs the
checks at every commit — [docs/checks.md](docs/checks.md) says what they are
and how to install them.
