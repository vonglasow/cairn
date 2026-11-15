# Installing cairn

## With Homebrew

```sh
brew install vonglasow/tap/cairn          # the latest
brew install vonglasow/tap/cairn@0.4      # or a minor kept at its own name
```

The formula installs the binary made for your system and builds nothing.
`brew upgrade cairn` takes a newer version.

## From a release

Every version on GitHub comes with a binary for each system:
`cairn-vX.Y.Z-darwin-arm64.tar.gz`, `-darwin-amd64`, `-linux-amd64`,
`-linux-arm64`, and `-windows-amd64.zip`, with `SHA256SUMS` beside them. Take
the one for your system from the
[latest release](https://github.com/vonglasow/cairn/releases/latest), unpack
it, and put `cairn` somewhere on your `PATH`:

```sh
install -m 755 cairn ~/.local/bin/cairn     # or /usr/local/bin, with sudo
cairn --help
```

On Windows, put `cairn.exe` in a folder on your `PATH`.

`cairn upgrade` takes the newest version: it downloads the archive made for
your system from the latest release, checks it against `SHA256SUMS`, and puts
the new binary in the place of the one you ran. That place has to be one you
can write; a cairn in `/usr/local/bin` is upgraded with `sudo cairn upgrade`.
A cairn installed by Homebrew is left to `brew upgrade cairn`.

## As an application

The window shows its logo on its own: on Linux, `cairn gui` writes the desktop
entry and the icons a desktop looks for, under your own data folder
(`~/.local/share`), the first time it opens. macOS shows the logo in the Dock
on its own, and on Windows `cairn.exe` carries it.

To open the window from the desktop rather than a terminal, run `cairn app
install` once: on Linux it writes the same entry and icons without opening the
window, and refreshes the applications menu with `xdg-desktop-menu` when
xdg-utils is installed; on a Mac it writes `cairn.app` in `~/Applications`.

## Where your work is kept

cairn keeps its work in `~/.cairn` unless told otherwise; `cairn config
show` says where. Installing a newer version leaves it where it is; a version
that changes how it is kept says so, and `cairn migration apply` moves it on,
setting a backup aside first.

## Then

- [Using cairn](usage.md): the first five minutes.
- [With Claude Code](claude-code.md): the plugin, and the commands that install
  it.
