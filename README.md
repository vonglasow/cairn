# cairn

A personal objective and task manager. CLI first, then MCP; the terminal and
desktop boards come later, once the engine stands on its own.

```sh
cairn task add "Write the spec" --desc "the one that unblocks the rest" --state todo --priority high --label spec --label team --deadline 2026-09-01 --estimate 90m
cairn task edit 1 --title "Write the other spec" --desc "" --state current --priority critical --deadline 2026-09-15 --estimate 2h
cairn task edit 1 --label +urgent --label -spec
cairn task list
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

Given empty, an option clears what was there. Add `--json` to any command to be
answered as a machine — including when it refuses.

## Where your tasks live

`~/.cairn/cairn.db`, on every system. Set `CAIRN_STORE` to put it elsewhere.

## Build

```sh
make build                      # ./bin/cairn
make run ARGS='task list'
```

## Working on it

[CLAUDE.md](CLAUDE.md) carries the rules, `docs/` the detail. `prek` runs the
checks at every commit — [docs/checks.md](docs/checks.md) says what they are
and how to install them.
