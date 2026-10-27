# cairn

A personal task and objective manager: a command line, a terminal board, a
desktop window, and an MCP server for agents — one binary, one local store.

## Install

### macOS and Linux, with Homebrew

```sh
brew install vonglasow/tap/cairn
```

### From a release

Download the archive for your system from the
[latest release](https://github.com/vonglasow/cairn/releases/latest), check it
against `SHA256SUMS`, and put `cairn` on your `PATH`.

| System | Archive |
|---|---|
| macOS, Apple silicon | `cairn-vX.Y.Z-darwin-arm64.tar.gz` |
| macOS, Intel | `cairn-vX.Y.Z-darwin-amd64.tar.gz` |
| Linux, x86-64 | `cairn-vX.Y.Z-linux-amd64.tar.gz` |
| Linux, ARM | `cairn-vX.Y.Z-linux-arm64.tar.gz` |
| Windows | `cairn-vX.Y.Z-windows-amd64.zip` |

**macOS and Linux**

```sh
tar -xzf cairn-vX.Y.Z-*.tar.gz cairn
install -m 755 cairn ~/.local/bin/cairn
```

On macOS, a binary downloaded with a browser is held by Gatekeeper; release
it once with `xattr -d com.apple.quarantine ~/.local/bin/cairn`. Homebrew does
not need this.

**Windows** — unzip it and put `cairn.exe` in a folder on your `PATH`.

### From source

```sh
git clone https://github.com/vonglasow/cairn.git
cd cairn
make build          # bin/cairn — needs Go and a C compiler, see INSTALL.md
```

## Quick start

```sh
cairn task new "Write the spec"      # task 1, in the backlog
cairn task edit 1 --state todo       # move it
cairn task focus 1                   # what you work on now — its time is counted
cairn task list                      # the board, column by column
cairn timesheet show                 # how long today's work took
```

Then open the same board where you prefer:

| | |
|---|---|
| `cairn tui` | in the terminal |
| `cairn gui` | in a window |
| `cairn mcp serve` | for an agent, over MCP |
| `cairn plugin write` | the Claude Code plugin |

`cairn <thing> <verb> --help` says the rest, with examples.

## Configuration

Nothing needs configuring: cairn keeps everything under `~/.cairn`. To choose
otherwise, `cairn config init` writes `~/.config/cairn/config.toml` with every
key, set to what is used now, to change from there — never over a file
already there. Every key is optional, and a key cairn does not know is refused
rather than ignored. The whole file:

```toml
# Where cairn keeps your tasks: a file. Absolute, or under ~.
store = "~/.cairn/cairn.db"

# Where cairn looks for the notes of each task: a folder, one per task inside
# it (~/.cairn/notes/12 for task 12). Absolute, or under ~.
notes = "~/.cairn/notes"

# Where cairn writes its Claude Code plugin, and keeps it at its own version:
# a folder. Beside the store when it is not said. Absolute, or under ~.
plugin = "~/.cairn/claude-plugin"

# The pages every surface offers: the boards' tabs, the commands, the MCP
# tools. Each is on unless it says false.
[pages]
tasks = true
objectives = true
time = true

# Archiving on its own: days are counted from start, in periods of that many
# days, and when a board opens or the tasks are listed, a task that went done
# before the period under way began is archived. Off unless start is said;
# a period lasts 14 days unless days says otherwise.
# [archive]
# start = 2026-10-21
# days = 14

# How long a focus session lasts before cairn asks, in every answer the MCP
# server gives, to renew it by focusing the task again or to end it.
# Nothing is asked unless it is said.
# [focus]
# session = "1h"

# What a deleted task does with its time: "kept" on the sheet, said to be
# deleted, or "taken" with it, as though it had never been spent.
# [time]
# deleted = "kept"

# Whether the window opens with the columns, the lanes and the band as wide
# and as tall as it was left with, rather than as they start. Kept in a file
# beside the store. Off unless it says true.
# [window]
# keep_sizes = true
```

`CAIRN_STORE` and `CAIRN_NOTES`, in the environment, replace `store` and
`notes` for as long as they are set. `CAIRN_AUTHOR` names who the acts are
written down as in the trace, the one logged in otherwise. Beside the store, cairn also keeps its
backups and the draft of a task being added in the window. `cairn config show` says which
places are in use and where each was found;
[docs/config.md](docs/config.md) says why it works this way.

## Reporting a bug

When cairn does something wrong, run it again in debug mode, do the one thing
that goes wrong, then pack what was written down:

```sh
CAIRN_DEBUG=1 cairn gui      # or tui, or any command
cairn debug report           # writes cairn-debug-….zip in the current folder
```

Send the archive with the report. It says what the machine and the store are
and what cairn did, and nothing anybody wrote: every title, description and
text typed is replaced by a fingerprint and a length. Debug mode keeps its
files in `debug/` beside the store, the last twenty, and stops with the
command it was turned on for. [docs/debug.md](docs/debug.md) says more.

## Documentation

- [Using cairn](docs/usage.md) — tasks, objectives, focus and time, labels, backups
- [Installing in detail](INSTALL.md) — what a build needs on each system
- [The MCP server](docs/mcp.md) and [the Claude Code plugin](docs/claude-code.md)
- [Making a release](docs/release.md)
- [Debug mode](docs/debug.md) — what it writes down, and what it never does

## Contributing

cairn is written scenario first, then test, then code; every commit passes
the checks. Start with [docs/development.md](docs/development.md),
[docs/workflow.md](docs/workflow.md) and [docs/checks.md](docs/checks.md).
