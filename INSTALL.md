# Installing cairn from its sources

What a version of cairn is handed over as: its sources, built where they are.
The same archive builds on macOS and on Linux.

## What it needs

- **Go 1.26 or later.** `go version` says which one is there. A Go a little
  older fetches the toolchain `go.mod` names on its own, the first time.
- **make.**
- **On Linux, the system's graphics headers**, which the window is built
  against. On Debian or Ubuntu:

  ```sh
  sudo apt install gcc pkg-config libwayland-dev libx11-dev libx11-xcb-dev \
    libxkbcommon-x11-dev libgles2-mesa-dev libegl1-mesa-dev libffi-dev \
    libxcursor-dev libvulkan-dev
  ```

  On another distribution, the same libraries under its own names — Gio, the
  toolkit the window is drawn with, lists them at
  <https://gioui.org/doc/install/linux>.
- **On macOS, the command line tools** (`xcode-select --install`), for the C
  compiler the window needs.

## Building it

```sh
tar -xzf cairn-v1.2.0-src.tar.gz
cd cairn-v1.2.0
make build
bin/cairn --version
```

`bin/cairn --version` names the version the archive was made as, and the
commit it was made from. The first build fetches the modules cairn is built
from, so it needs the network once.

## Installing it

Put the binary somewhere on your `PATH`:

```sh
install -m 755 bin/cairn ~/.local/bin/cairn     # or /usr/local/bin, with sudo
cairn --help
```

cairn keeps its work in `~/.cairn` unless told otherwise — `cairn config
show` says where. Installing a newer version leaves it where it is; a version
that changes how it is kept says so, and `cairn migration apply` moves it on,
setting a backup aside first.

## Using it

- `cairn task add "Write the spec"`, `cairn task list` — the command line.
- `cairn tui` — the board, in the terminal.
- `cairn gui` — the same board, in a window.

`THIRD_PARTY_NOTICES`, beside this file, holds the licences of what cairn is
built from; it goes wherever the binary goes.

## With Claude Code

The `cairn` plugin ties a Claude Code session to the cairn task it serves: it
loads the task's notes when you say `on travaille sur cairn#12`, lists what to
work on when you ask, holds the task in focus while you work and lets it go
when you stop. [docs/claude-code.md](docs/claude-code.md) says how.

It needs `cairn` and `jq` on your `PATH`. In Claude Code:

```text
/plugin marketplace add /path/to/cairn
/plugin install cairn@cairn
```

The marketplace is this repository — a local copy, or its git URL. Restart
Claude Code afterwards.

The plugin declares cairn's MCP server itself. If you already declared one,
declare it with the same command, or Claude Code loads both and every cairn
tool appears twice:

```sh
claude mcp remove cairn
claude mcp add --scope user cairn -- cairn mcp serve --access full
```
