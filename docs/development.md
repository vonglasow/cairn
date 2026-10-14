# Working on cairn

How to try cairn by hand without touching your own tasks, and where the rules
for changing it are. Building it is in [INSTALL.md](../INSTALL.md); `make help`
lists every target.

## Trying it by hand

The store lives in a temporary folder, so nothing here touches your own.

To look at a board full of believable work rather than build one by hand,
`make fixtures` (or `scripts/fixtures.sh [folder]`) makes a store under
`/tmp/cairn-fixtures` (or the folder named) and opens the window on it. It only
ever deletes a folder it made itself.

When the window misbehaves on one system only, `make trace` opens the same
board with every pointer and keyboard event written to a file under `traces/`,
headed by what the machine is — its system, display session, input method,
keyboard and language. Do the one thing that goes wrong, close the window, and
send the file it names.

```sh
make build
export CAIRN_STORE=$(mktemp -d)/cairn.db
./bin/cairn objective new "Ship cairn 1.0" --desc "the first real one" --priority high --label personal --label spec --deadline 2026-12-31 --estimate 40h
./bin/cairn objective list
./bin/cairn objective show 1
./bin/cairn task new "Write the spec" --desc "the very first" --state todo --priority high --label Spec --label team --deadline 2026-09-01 --estimate 90m
./bin/cairn task new "Read it over" --parent 1
./bin/cairn task new "Publish it" --parent 1
./bin/cairn task new "Something else"
./bin/cairn task list
./bin/cairn task show 1
./bin/cairn task show 2
./bin/cairn task show 1 --json
./bin/cairn task list --json
./bin/cairn task edit 1 --state current
./bin/cairn task edit 1 --label +urgent
./bin/cairn task edit 1 --label -spec
./bin/cairn task edit 1 --label "+draft,+wip"
./bin/cairn task new "A list in one flag" --label a,b,c
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
./bin/cairn objective new "Learn Go" --state on-hold
./bin/cairn objective list
./bin/cairn objective show 2
./bin/cairn objective delete 2 --yes
./bin/cairn objective delete 1 --yes
./bin/cairn objective new ""
./bin/cairn objective show 99
./bin/cairn objective new "z" --state todo
./bin/cairn task edit 2 --objective 1
./bin/cairn task new "w" --parent 1 --objective 1
./bin/cairn objective show spec
./bin/cairn objective
./bin/cairn backup new
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
./bin/cairn task new ""
./bin/cairn task show 99
./bin/cairn task edit 1
./bin/cairn task new "y" --priority urgent
./bin/cairn task show spec
./bin/cairn task new "x" --colour red
./bin/cairn task new "y" --estimate 500ms
./bin/cairn task edit 1 --label "+urgent,spec"
./bin/cairn task new
./bin/cairn task addd "z"
```

## Working on it

[CLAUDE.md](../CLAUDE.md) carries the rules, `docs/` the detail. `prek` runs the
checks at every commit — [docs/checks.md](checks.md) says what they are
and how to install them.
