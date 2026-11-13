<p align="center"><img src="internal/logo/cairn-128.png" alt="cairn" width="96"></p>

<h1 align="center">cairn</h1>

<p align="center">Your tasks, the objectives they serve and the time they took, in one local store.</p>

<p align="center">
<a href="https://github.com/vonglasow/cairn/actions/workflows/check.yml"><img src="https://github.com/vonglasow/cairn/actions/workflows/check.yml/badge.svg" alt="check"></a>
<a href="https://github.com/vonglasow/cairn/releases/latest"><img src="https://img.shields.io/github/v/release/vonglasow/cairn" alt="latest release"></a>
</p>

cairn is a personal task manager for people who work at a terminal and with AI
agents. One binary reaches one board four ways: the command line, a board in
the terminal, a desktop window, and an MCP server for agents. Each surface
calls the same engine, so each one behaves the same way.

- **Tasks on a board**: backlog, todo, current, blocked, watching, paused,
  done. Each task has a priority, labels, steps, links, a deadline and an
  estimate.
- **Objectives**: a task serves objectives, and an objective adds up the
  progress and time of its tasks. An objective may hold sub-objectives, two
  levels deep, and counts over its whole tree.
- **Views**: the filters of a listing kept under a name, opened again in one
  act, with the tasks outside them that touch what they keep named apart.
- **Series**: a task that comes back — every weekday at 10:30, every other
  Thursday, three days after it was last done — one occurrence at a time,
  each a task of its own, with a tab of its own on both boards and carried
  by an export. Off until a store's settings turn it on.
- **Time from focus**: the task in focus counts its time. `timesheet` reads it
  back by day, and you can add time spent away from the board.
- **Built for agents**: an MCP server and a Claude Code plugin. An agent keeps
  its work in the task's notes, never in the description.
- **Local and yours**: one SQLite file and a folder of notes. There is no
  account, no server and no telemetry. A full trace records every change.

## Contents

- [Install](#install)
- [Quick start](#quick-start)
- [With Claude Code](#with-claude-code)
- [Upgrading](#upgrading)
- [Configuration](#configuration)
- [Reporting a bug](#reporting-a-bug)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

## Install

**macOS and Linux, with Homebrew**

```sh
brew install vonglasow/tap/cairn
```

**Windows, or without Homebrew**: download the archive for your system from
the [latest release](https://github.com/vonglasow/cairn/releases/latest) and
put `cairn` on your `PATH`. [INSTALL.md](INSTALL.md) covers each system,
including how to check the archive and Gatekeeper on macOS.

**From source** (needs Go and a C compiler, see [INSTALL.md](INSTALL.md)):

```sh
git clone https://github.com/vonglasow/cairn.git && cd cairn
make build          # writes bin/cairn
```

**As an application**: `cairn app install` adds cairn to the Mac's
Applications folder or to a Linux desktop's menu.

## Quick start

```sh
cairn task new "Write the spec"      # task 1, in the backlog
cairn task edit 1 --state todo       # move it
cairn task focus 1                   # work on it: its time is counted
cairn task list                      # the board, column by column
cairn timesheet show                 # how long today's work took
```

Then open the same board wherever you prefer:

| Command | Opens |
|---|---|
| `cairn tui` | the board in the terminal |
| `cairn gui` | the board in a window |
| `cairn mcp serve` | the board for an agent, over MCP |

`cairn --help` lists every command, and `cairn <thing> <verb> --help` shows
examples.

## With Claude Code

```sh
cairn plugin show
```

Then, in Claude Code, run the commands it prints, which install the plugin of
your cairn's version from its tag on GitHub:
`/plugin marketplace add vonglasow/cairn@vX.Y.Z` and
`/plugin install cairn@cairn`. The plugin brings the MCP server, the skills —
working on a task, reading back a day or a week, where things stand — and a
hook that ties a session to the task in focus. `cairn upgrade` prints the
commands for the new version; without GitHub, `cairn plugin write` writes the
plugin the binary carries. See [docs/claude-code.md](docs/claude-code.md).

## Upgrading

How you upgrade depends on how cairn was installed:

| Installed with | Upgrade with |
|---|---|
| Homebrew | `brew upgrade cairn` |
| a release archive | `cairn upgrade` (checks `SHA256SUMS`, replaces the binary in place) |
| a build from source | `git pull && make build` |

`cairn upgrade` refuses the other two cases on purpose. Replacing a Homebrew
install behind Homebrew's back would be undone at the next `brew upgrade`.
Replacing a build from source with a release could lose local work.

A new version may change the shape of the store. cairn then refuses to open it
until you run `cairn migration apply`, which saves a backup first.

## Configuration

Nothing needs configuring: cairn keeps everything under `~/.cairn`.
`cairn config init` writes `~/.config/cairn/config.toml` with every key and
its current value, and one workspace, `main`, on `~/.cairn`, the default — so
nothing moves — and never overwrites an existing file. `cairn config show`
prints the configuration in use and where each value came from.
`CAIRN_STORE` and `CAIRN_NOTES` in the environment override `store` and
`notes`.

A workspace is a named folder holding a store and its notes. `cairn workspace
add work ~/cairn/work` names one, `cairn workspace use work` makes it the
default, `cairn workspace forget work` forgets it and leaves its folder alone,
`cairn workspace list` and `cairn workspace show` say which there are and which
one answers, and why — the rung that named it and those it passed over — each
in its colour, with its time when it is not on. `--workspace <name>` or `--store <path>` on any
command picks another for that command alone, whatever the environment says.
`CAIRN_WORKSPACE=<name>` in the environment names one for whatever is started
from it — a repository's `.claude/settings.json` sets it for the plugin's MCP
server and its hook — under `CAIRN_STORE` and `CAIRN_NOTES`, which name places.
An MCP server keeps the workspace it started on — `cairn mcp serve --workspace
<name>` names it — and `workspace use` does not move it.
Beside each store, `settings.toml` says what belongs to that store —
`[pages]`, `[archive]`, `[focus]`, `[approach]`, `[time]`, `[window]`, and the
`[workspace]` colour it wears — and wins over the global file key by key;
`cairn workspace add` writes one that says nothing yet. A store a board opens
by path is remembered among the recent folders, in
`~/.local/state/cairn/recent.json`, never in the configuration. [docs/config.md](docs/config.md) explains why it works this
way.

<details>
<summary>The whole file</summary>

```toml
# Where cairn writes its Claude Code plugin, and keeps it at its own version:
# a folder. Beside the store when it is not said. Absolute, or under ~.
plugin = "~/.cairn/claude-plugin"

# The pages every surface offers: the boards' tabs, the commands, the MCP
# tools. Each is on unless it says false. The time is "on", recorded and
# shown; "hidden", recorded and not shown, there again once it is on; or
# "off", neither: a task in focus counts no time until it is on again.
[pages]
tasks = true
objectives = true
time = "on"

# Archiving on its own: days are counted from start, in periods of that many
# days, and when a board opens or the tasks are listed, a task that went done
# before the period under way began is archived. Off unless start is said.
[archive]
# start = 2026-10-21
days = 14

# How long a focus session lasts before cairn asks, in every answer the MCP
# server gives, to renew it by focusing the task again or to end it.
# Nothing is asked unless it is said.
[focus]
# session = "1h"

# The label that asks an agent to agree on the approach with you before any
# code: it writes the solution and what it set aside in the task's notes, and
# waits for your word. Nothing of the kind happens unless a label is said.
[approach]
# label = "discuss"

# What a deleted task does with its time: "kept" on the sheet, said to be
# deleted, or "taken" with it, as though it had never been spent. And whether a
# week of the time page shows its Saturday and Sunday: "shown", or "hidden" for
# a week of five working days.
[time]
deleted = "kept"
weekends = "shown"

# Whether the window opens with the columns, the lanes and the band as wide
# and as tall as it was left with, rather than as they start. Kept in a file
# beside the store. And the colour scheme it wears: one cairn carries —
# default, solarized-dark, solarized-light, gruvbox-dark, gruvbox-light,
# catppuccin-mocha, catppuccin-latte, dracula, nord — or a base16 file of
# your own, ~/.config/cairn/themes/<name>.yaml. A store's settings.toml can
# name its own, so each workspace wears one. Last, how the board arrives
# when the window opens: ripple, deck, spring, accordion, sweep, flap, pull,
# particles, dolly, blueprint or stones. None when the system asks for less
# motion. And how much of what is dragged shows, in per cent, 20 to 90.
[window]
keep_sizes = false
theme = "default"
intro = "ripple"
ghost = 65

# Whether a task may come back as a series: made with a rule, every weekday
# 10:30, and made again when it is done or skipped. Off unless it says true;
# turned off, no series makes anything and every task one made stays.
[series]
on = false

# Workspaces: each a name, and the folder holding its own store and notes
# (<folder>/cairn.db and <folder>/notes), or a table of its own naming its
# folder and the notes it keeps elsewhere, a vault say. Absolute, or under ~.
# One always answers: --workspace, CAIRN_WORKSPACE, the one whose folder the
# current directory sits under — or the project its .cairn folder sits in —
# then default, then main in ~/.cairn. CAIRN_STORE, CAIRN_NOTES and --store
# name a store or notes over it. cairn workspace add, use and remove write
# here for you. The colour a workspace wears is its own settings.toml's to
# say, under [workspace].
[workspaces]
default = "main"
main = "~/.cairn"
# work = "~/cairn/work"
# home = "~/cairn/home"
#
# [workspaces.cairn]
# folder = "~/code/cairn/.cairn"
# notes = "~/vaults/cairn"
```

</details>

## Reporting a bug

Run cairn again in debug mode, do the one thing that goes wrong, then pack what
was recorded:

```sh
cairn gui --debug            # or tui, or any command; CAIRN_DEBUG=1 does the same
cairn debug report           # writes cairn-debug-….zip in the current folder
```

Attach the archive to the
[issue](https://github.com/vonglasow/cairn/issues). It describes the machine,
the store and what cairn did, but none of the text you typed: every title and
description is replaced by a fingerprint and a length. See
[docs/debug.md](docs/debug.md).

## Documentation

Published with each version at
[vonglasow.github.io/cairn](https://vonglasow.github.io/cairn), the same pages
as these.

- [Using cairn](docs/usage.md): tasks, objectives, focus and time, labels,
  backups
- [Installing in detail](INSTALL.md): each system, and what a build needs
- [The MCP server](docs/mcp.md) and [the Claude Code plugin](docs/claude-code.md)
- [Configuration](docs/config.md)
- [Decisions](docs/decisions.md): what cairn leaves out on purpose, and why
- [Making a release](docs/release.md)

## Contributing

cairn is written scenario first, then test, then code, and every commit passes
the checks. Start with [docs/development.md](docs/development.md),
[docs/workflow.md](docs/workflow.md) and [docs/checks.md](docs/checks.md).

## License

No license has been chosen yet. The licenses of what cairn is built from ship
in every release archive, as `THIRD_PARTY_NOTICES`; `make notices` writes them
from a clone.
