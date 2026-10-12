# Installing cairn from its sources

What a version of cairn is handed over as: its sources, built where they are.
The same archive builds on macOS, on Linux and on Windows.

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

  On another distribution, the same libraries under its own names: Gio, the
  toolkit the window is drawn with, lists them at
  <https://gioui.org/doc/install/linux>.
- **On macOS, the command line tools** (`xcode-select --install`), for the C
  compiler the window needs.
- **On Windows, nothing more than Go.** The window and the store need no C
  compiler there. `make` and the scripts are written for a POSIX shell, which
  Git for Windows provides as Git Bash; without one, see "On Windows" below.

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

### On Windows

In Git Bash, `make build` works as above and gives `bin/cairn`. Without a
POSIX shell, in PowerShell:

```powershell
go build -o cairn.exe ./cmd/cairn
.\cairn.exe --version
```

A build made this way does not know its version: it says `dev`, because the
version is stamped by `make`. Put `cairn.exe` in a folder on your `PATH`.

## Installing it

Put the binary somewhere on your `PATH`:

```sh
install -m 755 bin/cairn ~/.local/bin/cairn     # or /usr/local/bin, with sudo
cairn --help
```

cairn keeps its work in `~/.cairn` unless told otherwise; `cairn config
show` says where. Installing a newer version leaves it where it is; a version
that changes how it is kept says so, and `cairn migration apply` moves it on,
setting a backup aside first.

## Using it

- `cairn task add "Write the spec"`, `cairn task list`: the command line.
- `cairn tui`: the board, in the terminal.
- `cairn gui`: the same board, in a window.

`THIRD_PARTY_NOTICES`, beside this file, holds the licences of what cairn is
built from; it goes wherever the binary goes.

## With Claude Code

The `cairn` plugin ties a Claude Code session to the cairn task it serves: it
loads the task's notes when you say `on travaille sur cairn#12`, lists what to
work on when you ask, holds the task in focus while you work and lets it go
when you stop. [docs/claude-code.md](docs/claude-code.md) says how.

It needs `cairn` and `jq` on your `PATH`, and its hook is a shell script: on
Windows, Claude Code runs it through Git Bash, and `jq` comes with
`winget install jqlang.jq`.

The binary carries the plugin it was built with, so the sources are not needed
for it. Write it out, beside the store unless `--dir` names a folder:

```sh
cairn plugin write
```

It says where it wrote it. Then, in Claude Code:

```text
/plugin marketplace add <the folder cairn plugin write named>
/plugin install cairn@cairn
```

Restart Claude Code afterwards. After installing a newer cairn, run `cairn
plugin write` again, then `/plugin marketplace update cairn`, so the plugin
matches the cairn it calls. From a copy of this repository, the repository
itself is a marketplace too: `/plugin marketplace add /path/to/cairn`.

The plugin declares cairn's MCP server itself. If you already declared one,
declare it with the same command, or Claude Code loads both and every cairn
tool appears twice:

```sh
claude mcp remove cairn
claude mcp add --scope user cairn -- cairn mcp serve --access full
```
