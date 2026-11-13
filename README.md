<p align="center"><img src="internal/logo/cairn-128.png" alt="cairn" width="96"></p>

<h1 align="center">cairn</h1>

<p align="center">Your tasks, the objectives they serve and the time they took, in one local store.</p>

<p align="center">
<a href="https://github.com/vonglasow/cairn/actions/workflows/check.yml"><img src="https://github.com/vonglasow/cairn/actions/workflows/check.yml/badge.svg" alt="check"></a>
<a href="https://github.com/vonglasow/cairn/releases/latest"><img src="https://img.shields.io/github/v/release/vonglasow/cairn" alt="latest release"></a>
<a href="https://vonglasow.github.io/cairn/"><img src="https://img.shields.io/badge/docs-vonglasow.github.io%2Fcairn-7b93ff" alt="documentation"></a>
<a href="go.mod"><img src="https://img.shields.io/github/go-mod/go-version/vonglasow/cairn" alt="Go version"></a>
<img src="https://img.shields.io/badge/runs_on-macOS_%7C_Linux_%7C_Windows-555" alt="macOS, Linux and Windows">
<a href="https://vonglasow.github.io/cairn/tools.html"><img src="https://img.shields.io/badge/MCP-server-7b93ff" alt="MCP server"></a>
</p>

cairn is a personal task manager for people who work at a terminal and with AI
agents. One binary reaches one board four ways: the command line, a board in
the terminal, a desktop window, and an MCP server for agents. Each surface
calls the same engine, so each one behaves the same way.

**[Read the documentation](https://vonglasow.github.io/cairn/)** — the guide,
every command, every MCP tool and every configuration key.

- **Tasks on a board**, with priorities, labels, steps, links, deadlines and
  estimates.
- **Objectives** that add up the progress and time of the tasks serving them.
- **Views**: the filters of a listing kept under a name.
- **Series**: a task that comes back, every weekday or three days after it
  was done.
- **Time from focus**: the task in focus counts its time; `timesheet` reads it
  back.
- **Built for agents**: an MCP server and a Claude Code plugin.
- **Local and yours**: one SQLite file and a folder of notes; no account, no
  server, no telemetry.

## Install

```sh
brew install vonglasow/tap/cairn
```

On Windows, or without Homebrew, take the archive for your system from the
[latest release](https://github.com/vonglasow/cairn/releases/latest) and put
`cairn` on your `PATH`. [INSTALL.md](INSTALL.md) covers each system and
building from source; `cairn app install` adds cairn to the Mac's
Applications or a Linux desktop's menu.

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

## Documentation

- [Using cairn](https://vonglasow.github.io/cairn/usage.html): tasks,
  objectives, focus and time, labels, backups
- [Commands](https://vonglasow.github.io/cairn/commands.html) and
  [MCP tools](https://vonglasow.github.io/cairn/tools.html), read off cairn
  itself
- [Configuration](https://vonglasow.github.io/cairn/config.html): nothing
  needs configuring; `cairn config init` writes every key, and workspaces keep
  several boards apart
- [With Claude Code](https://vonglasow.github.io/cairn/claude-code.html):
  `cairn plugin show` prints the two commands that install the plugin
- Upgrading: `brew upgrade cairn`, or `cairn upgrade` for a release archive;
  a store a new version reshapes waits for `cairn migration apply`, which
  saves a backup first
- [Decisions](https://vonglasow.github.io/cairn/decisions.html): what cairn
  leaves out on purpose, and why

## Reporting a bug

Run cairn again with `--debug` (or `CAIRN_DEBUG=1`), do the one thing that
goes wrong, then `cairn debug report`, and attach the archive it writes to an
[issue](https://github.com/vonglasow/cairn/issues). None of the text you typed
is in it.

## Contributing

cairn is written scenario first, then test, then code, and every commit passes
the checks. What whoever develops cairn reads is in `.claude/docs/`: start with
[development](.claude/docs/development.md), [workflow](.claude/docs/workflow.md)
and [checks](.claude/docs/checks.md); [release](.claude/docs/release.md) says
how a version is made.

## License

No license has been chosen yet. The licenses of what cairn is built from ship
in every release archive, as `THIRD_PARTY_NOTICES`; `make notices` writes them
from a clone.
