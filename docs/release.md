# Making a version

Read this before handing a binary to anybody, including yourself on another
machine. Nothing here runs on its own: a version is made rarely, by hand, and
this is the list that keeps it from being made from memory.

## What a version is

A tag on a commit of `main`, and the binary built from it. There are no tags
yet: every build so far says the commit it came from (`git describe --always`),
which is what the Makefile stamps when there is no tag to name.

The number reads `vMAJOR.MINOR.PATCH`. What moves it:

| | moves when |
|---|---|
| major | a store written by the new version cannot be read by the old one *without a migration* the old one does not know, or an answer an agent or a script reads changes shape — see [mcp.md](mcp.md), on what changes in an answer |
| minor | a command, an argument, a field or a migration is added |
| patch | nothing anybody reads changes |

## The steps

1. **Everything checked, on both systems.**

   ```sh
   make check     # every hook, the commit's and the push's, on every file
   make linux     # the suite on Linux, in a container — needs Docker
   ```

   `make check` is what a push runs and more; `make linux` is the half of
   "macOS and Linux" no hook can check from a Mac. A version is not made past
   a red one of either.

2. **The tag.** On `main`, with nothing uncommitted:

   ```sh
   git tag -a v1.2.0 -m "v1.2.0"
   ```

   The repository has no remote today; the day it has one, the tag is pushed
   with the commit it names.

3. **The binary, and what travels with it.**

   ```sh
   make build     # bin/cairn, stamped with the tag, the commit and the time
   make notices   # bin/THIRD_PARTY_NOTICES: the licences of what it is built from
   bin/cairn --version
   ```

   `--version` has to name the tag, not a commit: a binary built from a dirty
   tree says `-dirty`, and is not a version. The notices go wherever the binary
   goes — most of what it is built from asks for its licence to travel with it,
   and `scripts/notices.sh` reads them from the module cache, for the modules the
   binary actually links.

4. **The other system.** A binary for Linux is built on Linux: the window needs
   cgo and the system's graphics headers there, which a cross-compile from a Mac
   cannot reach. `scripts/linux.sh` is where those headers are named.

## What a version is not

- **A changelog written afterwards.** The commits say why each change was made;
  a version is read as the commits since the last tag. A breaking change says
  so in its own commit, under `BREAKING CHANGE`, which is where somebody
  reading the log between two tags will look.
- **A migration run for somebody.** A store behind the new binary is refused
  until `cairn migration apply` — which sets a backup aside first. A version
  that changed the store's shape on the way past would be changing somebody's
  data as a side effect of upgrading. See [architecture.md](architecture.md).
