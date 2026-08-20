# cairn

A personal objective and task manager. CLI first, then MCP; the terminal and
desktop boards come later, once the engine stands on its own.

```sh
cairn task add "Write the spec" --desc "the one that unblocks the rest" --state todo --priority high --label spec --label team --deadline 2026-09-01 --estimate 90m
cairn task edit 1 --title "Write the other spec" --desc "" --state current --priority critical --deadline 2026-09-15 --estimate 2h
cairn task edit 1 --label +urgent --label -spec
cairn task add "Write the tests" --parent 1
cairn task list      # sub-tasks are listed under the task they hang from
cairn task show 1
cairn task delete 1 --yes
```

| option | what it takes |
|---|---|
| `--title` | on `edit` only; `add` takes the title as its argument |
| `--desc` | any text |
| `--state` | `backlog` `todo` `current` `done` — a new task starts in `backlog` |
| `--priority` | `low` `normal` `high` `critical` — a new task carries `normal` |
| `--label` | one per flag: `--label spec --label team`. Lower-cased, deduplicated, sorted. Given plainly it replaces the whole set; on `edit`, `+urgent` adds one and `-spec` takes one away, leaving the rest alone. The two forms cannot be mixed |
| `--deadline` | a day: `2026-09-01` |
| `--estimate` | a duration: `90m`, `1h30m`, `2h` — read back as `1h30m` |
| `--parent` | the number of the task this one hangs under. One level only: a sub-task carries none of its own. Given empty, it sets the task free |

Given empty, an option clears what was there. A task carrying sub-tasks is not
deleted while they are there. Add `--json` to any command to be answered as a
machine — including when it refuses.

## Backups and versions

```sh
cairn backup add                 # backup the store as it stands
cairn backup list                # every backup, with its number and date
cairn backup show 2              # when it was made, and what it holds
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

## Where your tasks live

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
./bin/cairn task edit 1 --desc ""
./bin/cairn task show 1
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
./bin/cairn task add
./bin/cairn task addd "z"
```

## Working on it

[CLAUDE.md](CLAUDE.md) carries the rules, `docs/` the detail. `prek` runs the
checks at every commit — [docs/checks.md](docs/checks.md) says what they are
and how to install them.
