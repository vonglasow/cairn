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

On Linux, `make desktop` gives the window its icon in the desktop's launcher.

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

## Documentation

- [Using cairn](docs/usage.md) — tasks, objectives, focus and time, labels, backups
- [Installing in detail](INSTALL.md) — what a build needs on each system
- [The MCP server](docs/mcp.md) and [the Claude Code plugin](docs/claude-code.md)
- [Making a release](docs/release.md)

## Contributing

cairn is written scenario first, then test, then code; every commit passes
the checks. Start with [docs/development.md](docs/development.md),
[docs/workflow.md](docs/workflow.md) and [docs/checks.md](docs/checks.md).
