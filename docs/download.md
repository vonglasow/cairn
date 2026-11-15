# Download cairn

One binary for macOS, Linux and Windows, nothing else to install.

| system | how |
|---|---|
| macOS | `brew install vonglasow/tap/cairn`, or the archive for Apple silicon (`darwin-arm64`) or Intel (`darwin-amd64`) |
| Linux | `brew install vonglasow/tap/cairn`, or the archive for x86-64 (`linux-amd64`) or ARM64 (`linux-arm64`) |
| Windows | the archive holding `cairn.exe` (`windows-amd64`): put it in a folder on your `PATH` |

Every version is on GitHub, in the
[latest release](https://github.com/vonglasow/cairn/releases/latest), with
what changed. Each archive comes with `SHA256SUMS`.

## From an archive

Unpack it and put `cairn` on your `PATH`:

```sh
install -m 755 cairn ~/.local/bin/cairn     # or /usr/local/bin, with sudo
cairn --help
```

## Upgrade

With Homebrew, `brew upgrade cairn`. From an archive, `cairn upgrade` fetches
the newest release for your system, checks it against `SHA256SUMS`, and puts
it in place of the one you ran; a cairn in `/usr/local/bin` is upgraded with
`sudo cairn upgrade`. Your work stays where it is: a version that changes how
it is kept says so, and `cairn migration apply` moves it on, saving a backup
first.

## Open it like an application

`cairn app install` once: on a Mac it writes `cairn.app` in
`~/Applications`; on Linux it adds cairn to the desktop's menu, under
`~/.local/share`, and refreshes it with `xdg-desktop-menu` when xdg-utils is
installed. Windows shows the logo on its own.

The window shows its logo without it: on Linux, `cairn gui` writes the same
entry and icons the first time it opens; macOS shows the logo in the Dock, and
`cairn.exe` carries it.

## Where your work is kept

In `~/.cairn`, unless you say otherwise: one file for the board, one folder
per task for its notes. `cairn config show` says where.

Then: [the first five minutes](usage.md), and
[plug in your agent](plug-in-an-agent.md).
