# Making a version

Read this before handing cairn to anybody, including yourself on another
machine. A version is made when somebody says the work is ready: the tag is
made by hand, and pushing it is what hands the version over — GitHub builds it
on every system and publishes it (`.github/workflows/release.yml`).

## What a version is

A tag on a commit of `main`, and what is made from it: a binary for each
system, and the Homebrew formula that installs it. Whoever wants to build cairn
clones it; the release hands over what is already built. A source archive was
made and handed over as the version until binaries were: it is gone, and
`make dist` with it. A build
with no tag to name says the commit it came from (`git describe --always`).

The number reads `vMAJOR.MINOR.PATCH`. What moves it:

| | moves when |
|---|---|
| major | a store written by the new version cannot be read by the old one *without a migration* the old one does not know, or an answer an agent or a script reads changes shape — see [mcp.md](mcp.md), on what changes in an answer |
| minor | a command, an argument, a field or a migration is added |
| patch | nothing anybody reads changes |

## The steps

A version is made with one command, on `main`, with nothing uncommitted:

```sh
make next-version   # what it would be tagged — says it, and does nothing
make release        # makes it
```

`make release` runs `scripts/release.sh`, and it is the one way a version is
made: there is no tag typed by hand anywhere in this. It does, in this order:

1. **Refuses what is not a version.** A branch other than `main`, a tree with
   something uncommitted.

2. **Reads the number off the commits.** [svu](https://github.com/caarlos0/svu)
   reads the commits since the last version and answers the next one: a `fix`
   moves the patch, a `feat` the minor, a `BREAKING CHANGE` the major — the
   table above, kept by the commit messages commitlint already holds to. While
   cairn is under 1.0 a break moves the minor (`--v0`): 1.0 is a decision, not
   something a commit message makes by accident. The first version is
   `v0.1.0`. A history since the last version with no `feat`, no `fix` and no
   break is refused: there is nothing to give a number to. svu is named at the
   version the Makefile pins and run with `go run`, so there is nothing to
   install first.

3. **Checks everything, on both systems.**

   ```sh
   make release-check  # every hook, the commit's and the push's, on every file — but the mutation sweep
   make linux          # the suite on Linux, in a container — needs Docker
   make build          # and it builds here
   ```

   `make linux` is the half of "macOS and Linux" no hook can check from a Mac.
   **The mutation sweep is not run**: it is by far the longest check, and a
   release that waited on it took the better part of an hour. It is run on its
   own, now and then, with `make mutation`, and what it last found is kept in
   `mutation-report/summary.md` — the release says the date of that sweep and
   whether it left anything to fix, so a version is not made in ignorance of
   it. Every commit's own mutation, on the lines it changed, still runs.

4. **Tags it**, an annotated tag named by the number, and only now: a version
   whose checks failed was never a version, and a tag made before them would
   say it was. Run again on a commit tagged already, the release checks it
   again, under that tag.

   `INSTALL.md` says how to install a version and how to build cairn from a
   clone. It is written in ASCII alone (`@version-install-reads-anywhere`): it
   is read before anything is installed, in whatever the machine has, and on
   Windows a tool that takes a file with no mark of its encoding for the
   system's code page wrote every dash as `â€”`.

5. **Is pushed, by hand.** `git push origin main` and `git push origin vX.Y.Z`:
   the person pushing decides the version goes out, and nothing goes out
   before. gaia, whose release runs on every push to `main`, showed what the
   other way costs — a version for every `feat`, many of them for very little.

## What pushing the tag does

`.github/workflows/release.yml` runs on a pushed tag `vX.Y.Z`, in three jobs:

1. **A binary on each system.** The window needs cgo — and the graphics
   headers on Linux — so no binary is built for a system from another: one job
   per system builds it where it runs, with `make package`, which
   `scripts/package.sh` is. macOS on Apple silicon and on Intel, Linux on
   amd64 and arm64, Windows. Each archive, `cairn-vX.Y.Z-<os>-<arch>.tar.gz`
   or `.zip` for Windows, holds the binary stamped with the version,
   `README.md`, `INSTALL.md` and `THIRD_PARTY_NOTICES` (`@version-binary-archive`).
   GoReleaser, which gaia uses, cannot do this without its paid edition: its
   free one builds every system from one machine, which cgo rules out.
2. **The release.** It holds the binaries, `SHA256SUMS` with every one's sum
   (`scripts/sums.sh`), and `CHANGELOG.md`; the tag's message — the version's
   part of the changelog — is its notes. No source archive: GitHub adds its
   own, and building is for whoever clones.
3. **The tap.** The formula, written by `scripts/formula.sh` from the sums,
   fetches the archive made for the system it is installed on and installs the
   binary, building nothing (`@version-brew-formula`); with a system's archive
   missing, none is written. It goes to `vonglasow/homebrew-tap`, the tap gaia's
   is in: `Formula/cairn.rb`, the latest, and `Formula/cairn@X.Y.rb`, kept for
   every minor (`@version-brew-formula-for-its-minor`), so
   `brew install vonglasow/tap/cairn` takes the latest and
   `brew install vonglasow/tap/cairn@0.4` an older one.

**The Claude Code plugin is the binary's, and the binary keeps it current.**
The plugin is embedded in the binary, so an archive needs nothing beside it,
and the formula installs the binary alone and says, in its caveats, to run
`cairn plugin write` once (`@version-brew-formula-says-how-to-add-the-plugin`).
After that, `cairn mcp serve` — which Claude Code starts with every session —
writes the plugin beside the store again whenever another version wrote it
(`@plugin-kept-current`), and Claude Code reads a marketplace on a local folder
in place, at the start of each session
([plugin loading](https://code.claude.com/docs/en/plugins/loading)). One way,
whatever installed cairn: Homebrew, an archive on Linux or Windows, a build
(cairn#90). Shipping the plugin in the archive for Homebrew to share under
`opt/cairn/share` was tried first and left: it covered brew alone, and gave a
second place for the same plugin.

It needs one secret, `GH_TOKEN`: a token that may push to
`vonglasow/homebrew-tap`, as gaia's. The release itself is made with the
workflow's own token.

**While the repository is private**, the workflow runs and the release is made,
but the formula cannot fetch a binary from it without a token, so
`brew install` fails for anybody: the repository goes public once everything
else is wired.

## What a version is not

- **A changelog written afterwards.** The commits say why each change was made;
  a version is read as the commits since the last tag. A breaking change says
  so in its own commit, under `BREAKING CHANGE` or with a `!`, which is where
  somebody reading the log between two tags will look.

## The changelog

`scripts/changelog.sh` reads it off the history every time, and no file in the
repository keeps it: every version is a tag, and what it brought is the
conventional commits between it and the one before, which commitlint already
holds to their shape. Newest first, each version says what breaks, what is new
and what was fixed — a `feat`, a `fix`, and either with a `!` or a
`BREAKING CHANGE`. The rest, docs and tests and chores, is in the history for
whoever wants it, and not in what somebody upgrading reads.

It goes two places: beside the binaries in the release, and into the
annotated tag, which carries its own version's part — `git show v1.2.0` says
what that version brought. `make changelog` prints it with what the next
version would bring, and makes nothing.

Kept as a file it would be one more thing to write in every commit, or a
release that commits on its own; read off the tags it cannot drift from them.
- **A migration run for somebody.** A store behind the new binary is refused
  until `cairn migration apply` — which sets a backup aside first. A version
  that changed the store's shape on the way past would be changing somebody's
  data as a side effect of upgrading. See [architecture.md](architecture.md).
