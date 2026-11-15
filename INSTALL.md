# Installing cairn

Building cairn from a clone, on macOS, on Linux or on Windows. To install a
version, with Homebrew or from a release, read
[docs/install.md](docs/install.md).

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
git clone git@github.com:vonglasow/cairn.git
cd cairn
make build
bin/cairn --version
```

`bin/cairn --version` names the version the clone is at, and the commit it
was built from. The first build fetches the modules cairn is built
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

Put `bin/cairn` on your `PATH`, as [docs/install.md](docs/install.md) says. A
cairn built from source is no version: `cairn upgrade` leaves it to being
built again, and `cairn plugin write` gives Claude Code the plugin it carries
([docs/claude-code.md](docs/claude-code.md)).

`THIRD_PARTY_NOTICES`, beside this file, holds the licences of what cairn is
built from; it goes wherever the binary goes.
