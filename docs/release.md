# Making a version

Read this before handing cairn to anybody, including yourself on another
machine. Nothing here runs on its own: a version is made rarely, by hand, and
this is the list that keeps it from being made from memory.

## What a version is

A tag on a commit of `main`, and the source archive made from it. There are no tags
yet: every build so far says the commit it came from (`git describe --always`),
which is what the Makefile stamps when there is no tag to name.

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
   say it was. The repository has no remote today; the day it has one, the tag
   is pushed with the commit it names. Run again on a commit tagged already,
   the release checks it again and makes its archive again, under that tag.

5. **Makes the archive.**

   ```sh
   make dist      # dist/cairn-v1.2.0-src.tar.gz, dist/CHANGELOG.md, dist/SHA256SUMS
   ```

   A version is handed over as **its sources**, built where they are with
   `make build` — the same archive on macOS and on Linux, with nothing hosted
   anywhere: the archive and its checksum go wherever colleagues fetch things
   from. `SHA256SUMS` keeps a line for every archive made in the folder: the
   one just made is written again, the others are left as they were — written
   whole, it lost the sum of every earlier version the moment a later one was
   made. It holds the tree the tag names and five files git would not give:

   - `VERSION`, `COMMIT` and `COMMIT_DATE`, what `git describe`, the commit
     and its date would have said. An archive has no history to ask, so the
     Makefile reads them first — without them a build from it called itself
     `dev`, made from `none`, on a date `unknown`.
     `make stamps` says what a build would be stamped with.
   - `CHANGELOG.md`, what changed, version by version — below.
   - `THIRD_PARTY_NOTICES`, the licences of what cairn is built from, read by
     `scripts/notices.sh` from the module cache for the modules the binary
     actually links, with Go's own; every licence and notice file of a module
     is kept, and a module with none fails it.

   `INSTALL.md`, at the root of the tree and so of the archive, says what a
   build needs — Go, make, a C compiler, and on Linux the graphics headers
   `scripts/linux.sh` installs — and how to put the binary on a `PATH`.

   `Formula/cairn.rb`, beside the archive, is the Homebrew formula that
   builds it: its version, its checksum, and `make build` with Go as a build
   dependency (`@version-brew-formula`). It fetches the archive from where it
   was made, a `file://` address on that machine, unless `FORMULA_URL` names
   the place archives are hosted (`@version-brew-formula-hosted`). Homebrew
   installs only from a tap, so `INSTALL.md` says how to make a local one.
   Made with the archive, it can never name another archive or another sum.

   It is written in ASCII alone (`@version-install-reads-anywhere`). It is read
   before anything is installed, in whatever the machine has, and on Windows
   a tool that takes a file with no mark of its encoding for the system's code
   page wrote every dash as `â€”`.

A binary built for another system is not part of a version: the window needs
cgo, and the system's graphics headers on Linux, which a cross-compile from a
Mac cannot reach. Whoever needs one builds it from the archive, where they are.

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

It goes three places: into the archive and beside it in `dist/`, and into the
annotated tag, which carries its own version's part — `git show v1.2.0` says
what that version brought. `make changelog` prints it with what the next
version would bring, and makes nothing.

Kept as a file it would be one more thing to write in every commit, or a
release that commits on its own; read off the tags it cannot drift from them.
- **A migration run for somebody.** A store behind the new binary is refused
  until `cairn migration apply` — which sets a backup aside first. A version
  that changed the store's shape on the way past would be changing somebody's
  data as a side effect of upgrading. See [architecture.md](architecture.md).
