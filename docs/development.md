# Working on cairn

How to try cairn by hand without touching your own tasks, and where the rules
for changing it are. Building it is in [INSTALL.md](../INSTALL.md); `make help`
lists every target.

## Trying it by hand

The store lives in a temporary folder, so nothing here touches your own.

To look at a board full of believable work rather than build one by hand,
`make fixtures` runs `cairn fixtures --new --gui --debug`: a store made afresh
from the file the binary carries (`internal/engine/fixtures/fixtures.json`, days
counted from today), in `cairn-fixtures` under the temporary folder, with the
window opened on it in debug mode. `cairn fixtures` alone, with no repository
and no Go, keeps a folder it made as it was left — a store worked on stays for
the next look — and `--new` makes it again; `--tui` opens the terminal board,
`--from` reads another export, and a folder it did not make is left alone. It
replaces `scripts/fixtures.sh`, thirty commands that needed the repository.

The file holds every case a board draws, each task titled for what it shows
("Steps: twenty, half done", "Deadline tomorrow: …"): every column and lane, a
backlog long enough to scroll, the four priorities, long titles, accents and
emoji, long and empty descriptions, steps none, all, half and twenty, labels
none, one, too many, chosen colours, colours read off the name and a label
nothing carries, deadlines past, today, tomorrow, in a week and far off,
estimates overrun and kept, time over several weeks, blockers by a task and by
a reason and one freed, links, two tasks in focus, archived ones, labels for
objectives only, and objectives in every standing, late, served by nothing,
and in trees two levels deep, each parent labelled `area` and each
sub-objective `project` — Ship cairn 1.0 › Make the store dependable,
Packaging and Documentation; Home automation › Enclosure, Firmware and
Features, Features waiting on Firmware; Health › Physiotherapy and Run a 10k,
both reached, so Health is proposed as reached — and labels for tasks only
(`chore`), for objectives only, and for both carried by both. Every task
serves an objective but five, kept apart on purpose: an inbox of ideas not
sorted yet, and two that show a card alone. Four series: a stand-up with its
occurrence open and two days missed, a report stopped, plants watered every
10 days after done, and a physiotherapy ended by its count, one session
skipped, its occurrences archived; and three views, one naming the Home
automation objective, one its sub-objective Firmware. Its store's settings
turn recurrence on, so both boards draw them (`@fixtures-series`), and give
it a colour of its own, peach (`@fixtures-a-colour-of-its-own`). One store
holds one value of a setting: the time stays on, the only way the time on a
card, an estimate overrun and the timesheet are drawn, and the window wears
the default scheme, the one the mockups draw.
`TestFixtures_holdEveryCaseABoardDraws` fails when a case goes missing; a case
added to a board is added there.

When cairn misbehaves on one system only, debug mode records it there: run
it with `CAIRN_DEBUG=1`, do the one thing that goes wrong, and send what
`cairn debug report` packs. See [debug.md](debug.md).

```sh
make build
export CAIRN_STORE=$(mktemp -d)/cairn.db
./bin/cairn objective new "Ship cairn 1.0" --desc "the first real one" --priority high --label personal --label spec --deadline 2026-12-31 --estimate 40h
./bin/cairn objective list
./bin/cairn objective show 1
./bin/cairn task new "Write the spec" --desc "the very first" --state todo --priority high --label Spec --label team --deadline 2026-09-01 --estimate 90m
./bin/cairn step new 1 "Read it over"
./bin/cairn step new 1 "Publish it"
./bin/cairn step edit 1 1 --done
./bin/cairn step list 1
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
./bin/cairn step edit 1 9 --done
./bin/cairn objective show spec
./bin/cairn objective
./bin/cairn backup new
./bin/cairn backup list
./bin/cairn backup show 1
./bin/cairn backup list --json
./bin/cairn task delete 2 --yes
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
./bin/cairn task edit 1 --label "+urgent,=spec"
./bin/cairn task new
./bin/cairn task addd "z"
```

## Working on it

[CLAUDE.md](../CLAUDE.md) carries the rules, `docs/` the detail. `prek` runs the
checks at every commit — [docs/checks.md](checks.md) says what they are
and how to install them.
