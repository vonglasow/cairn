# Checks

Read this when adding a tool, changing a threshold, or when a hook blocks a
commit.

## What runs, and when

`prek` runs on its own at every commit, and once more at every push. Nothing
here has to be remembered.

| check | what it answers |
|---|---|
| `trailing-whitespace`, `end-of-file-fixer`, `check-yaml`, `check-added-large-files` | is the file well formed? |
| `go-fmt`, `go-imports` | is it formatted the way Go formats? |
| `shellcheck` | does a shell script hold up to static analysis? |
| `golangci-lint` | does it hold up to static analysis, comments included — and does anything but a test import `testing`? |
| `semgrep` | is there a known bad pattern, or a secret written into a file? |
| `tested-check` | does every package have a test file at all? |
| `harness-check` | is the binary free of the test harness? |
| `windows-check` | does cairn, tests included, still build for Windows? |
| `quiet-check` | does building cairn, and the window's tests, print nothing at all? |
| `govulncheck` | does anything we call have a known vulnerability? |
| `cover-check` | does it work — and did every line at least run, in total and in each package? |
| `mutation` | would a test have noticed if a changed line were wrong? |
| `commitlint` | is the message conventional? |
| `mutation-all` | at push only: would a test have noticed if *any* line were wrong? |

**They run in that order, and the first to fail stops the rest** (`fail_fast`
in `.pre-commit-config.yaml`). Cheapest first: a missing import is reported in a second,
rather than after semgrep, the suite and the mutation gate have run for
nothing.

There is no `go-unit-tests` hook. `cover-check` builds its profile from `go
test -race ./...`, which fails the moment a test does, so a second hook running
the suite was the same answer paid for twice. The `bdd/` scenarios are Go
tests, so that one run covers them too.

**A hook is triggered by every file its answer depends on**, not only by Go.
`cover-check` also runs when a scenario, a pinned frame under `testdata/`, a
shell script, the Claude Code plugin, or `go.mod` changes — the suite runs the
scripts and the plugin's hook; `govulncheck`, `harness-check` and `windows-check` when `go.mod` or `go.sum`
does; `golangci-lint` when `.golangci.yml` does. Each says so with `files` in
`.pre-commit-config.yaml`. Filtered on Go alone they ran nothing at all for a commit that
changed only a scenario — which the suite executes — or only a dependency.

## Before the commit: a test seen red

Two hooks under `.claude/hooks` hold "scenario, then test, then code" while an
agent works, before prek ever sees a commit.

- `red-first.sh` refuses any write to a production `.go` file under
  `internal/` — through Write, Edit, a redirection, an in-place editor, a copy
  or move into place, a checkout or a patch — unless that package holds a
  receipt younger than thirty minutes. It needs `jq`, and refuses everything
  when `jq` is missing rather than letting everything through.
- `make red PKG=./internal/<package> [RUN=TestName]` is the only thing that
  mints a receipt, and it mints one only when a test failed, or when a test
  does not build yet — its error located in a `_test.go` file, the honest
  first red of a function that does not exist. A package that does not exist,
  an invalid `RUN` pattern, a suite that passes, or production code that does
  not compile all mint nothing.

`make red PKG=… MOVING="what moves, and where to"` is the one door that runs no
test: code moving out of a package asks for no new behaviour, so it has nothing
to fail. The sentence is kept in the receipt, which says `moving` rather than
naming a failing test, so a write let through this way is visible to anyone
reading `.claude/red`. What it owes is that the suite still passes once the
move is done.

**On a fresh clone there are no receipts**: `.claude/red` is ignored by git,
and a receipt says a failure was seen on this machine, which no clone can
inherit. The first piece of work starts like every other — a failing test,
then `make red PKG=…` — and nothing has to be set up first.

## Getting them

Without these, the hooks cannot run and nothing can be committed:

```sh
brew install prek jq
prek install -t pre-commit -t commit-msg -t pre-push
```

Go is the rest. golangci-lint, govulncheck, gremlins, goimports, svu and
go-winres are named, each at its version, in `tools/go.mod`, and run from it
with `go tool -modfile=tools/go.mod <tool>`: built the first time a hook asks,
from the cache after that. Nothing is installed by hand, so nothing installed
by hand drifts from what the repository says. A module of its own, so what the
tools are built from never enters cairn's graph and govulncheck does not judge
cairn by it.

`jq` is read by `scripts/mutation.sh` and by both hooks under `.claude/hooks`.

Two are not built from `tools/go.mod`: prek builds `semgrep`, `shellcheck` and
`commitlint` in environments of its own, from the revisions pinned in
`.pre-commit-config.yaml`. commitlint needs Node on `PATH` for that.

golangci-lint advises against being built this way, since a module it shares
its dependencies with can change what it finds. It shares them with nothing
here: `tools/go.mod` holds the tools alone. One place for every version was
worth more than its advice.

**What the tools find stays current; what the tools are is pinned.** Two
different things used to be said as one:

- semgrep's rules and govulncheck's advisories are fetched as they stand on
  the day. A commit which passed yesterday can fail today without the code
  having moved, and that is the right way round — the finding was already true
  yesterday. The rules are the registry's Go and secrets packs, named rather
  than chosen by `--config auto`, which reported home to choose them.
- the tools themselves are not. Each hook repository in
  `.pre-commit-config.yaml` is named by the commit it was at, never by a tag a
  maintainer can move; commitlint's configuration carries its version; the Go
  tools are the versions `tools/go.mod` names. A tool that changed its own
  behaviour found nothing that was true yesterday — what `scripts/mutation.sh`
  works around was read in one version's source, and another can quietly
  undo it.

### Who moves them

**Dependabot proposes, weekly, and the pull request is judged like a commit**
(`.github/dependabot.yml`). It reads four places: `go.mod` for what cairn is
built from, `tools/go.mod` for the tools — only the tools, by name, not what
each is built from — the workflows' actions, and the hooks of
`.pre-commit-config.yaml`, which is why that file is YAML: Dependabot reads no
other. An update waits three days after its release, so a version its makers
withdraw is seldom proposed; a security update does not wait. Minor and patch
come as one pull request per place, a major one alone. Nothing merges on its
own.

**`check.yml` runs every hook on every file of a pull request**, on Linux, but
the mutation of what changed: a pull request has no `HEAD` to change lines
since, and the full sweep is the push's. A change Dependabot made was made on
no machine, so no hook judged it before this.

Left to a person: Go itself, the `toolchain` line of both `go.mod`, which
Dependabot does not move; prek and jq; commitlint's own dependencies; the
runners' images.

### When Go itself moves

A Go program built against an older Go cannot read a module written for a
newer one. The tools no longer meet that: `go tool` builds them with the Go
that runs it, the one `go.mod` names, so a Go upgrade moves them with it. What
remains is Go itself:

| what it says | what it means |
|---|---|
| `go.mod requires go >= 1.26.6 (running go 1.26.5; GOTOOLCHAIN=local)` | Go itself is behind, and `GOTOOLCHAIN=local` forbids it from fetching the one the module asks for |

Measured, on the day Go went from 1.26.5 to 1.27.1: five advisories appeared
against the standard library, and the two tools then installed by brew refused
to run at all. The hour that cost is why the tools are now built by Go.

## The three that are not the same question

The suite says it works. Coverage says a line ran. **Mutation says a test would
have noticed if that line were wrong.**

A suite can sit at 90% coverage and assert almost nothing. Mutation breaks each
line on purpose and checks a test turns red, so a **surviving mutant is a test
that proves nothing**.

Coverage still earns its place beside it: a package nobody tested at all has no
mutants to judge, so mutation would pass it in silence.

That last sentence used to end "Coverage sees it drop", and it was too kind to
both gates. Measured, on a package added with sixty functions and no test file:

- **mutation did not see it at all.** The loop skips a package with no test
  file — `ls $pkg/*_test.go || continue` — so the newest code is the code the
  strictest gate says least about.
- **the coverage total absorbed it.** One package at 0% cost eight points of an
  average of eleven, and until the floor gave way nothing said a word. It took
  five rounds of work to surface, and it surfaced because somebody looked, not
  because a gate spoke.

So there are two more gates, and they are cheap. `tested-check` fails when a
package has no test file, which is the case mutation is blind to. And `cover-
check` now carries a floor **per package** beside the total, so a package at
zero fails on its own rather than being diluted by ten that are not.

Coverage counts every test that runs a line, scenarios included: the profile is
built with `-coverpkg=./...`, so a scenario in `internal/cli` that reaches the
engine, the store and a notion credits all of them. Until 2026-10-14 it was
measured per package from that package's own tests, which kept thousands of
lines of unit tests asserting again what a scenario already asserted through
the real surface; they went once the scenarios counted. Mutation is what still
says whether a covered line is guarded.

## Thresholds

- Statement coverage, in total: at least `COVERAGE_MIN` in the Makefile.
- Statement coverage, per package: at least `PACKAGE_COVERAGE_MIN`, also in the
  Makefile — the numbers live there and nowhere else. Lower than the total on
  purpose: a package may sit honestly under the
  average — measured, the store and the harness both do — and a per-package
  floor set at the total would fail the repository as it stands rather than the
  code being added to it. `cmd/` is left out — see what cannot be reached,
  below; what it does is exercised through the scenarios of `internal/cli`.
- Every package has a test file, with no threshold to argue about: it has one
  or it does not.
- Mutation efficacy: **100%** — of the mutants gremlins judged, all died.
- Mutant coverage: **100%** — it judged every mutant it made. Without this
  second one, code no test reaches produces mutants counted in neither column,
  and a run reports success having looked at nothing.

Both live in `.gremlins.yaml`, because gremlins ignores them when they are
passed as flags — measured, not assumed. Both read `99.99`, because the tool
fails on *less than or equal*, so a `100` would refuse a perfect run.

Neither sees a mutant that **timed out**: gremlins counts it in no column.
Measured, under load, 219 timeouts hid 21 survivors behind a reported 100%. So
the verdict is not gremlins' exit code but `scripts/mutation.sh`, reading its
report: any mutant LIVED or NOT COVERED fails the run. The mutants that timed
out are judged again with three times the budget — those alone, on the lines
they sit on, handed to gremlins as a diff naming those lines and nothing else;
a whole package judged a second time for one slow mutant is what made a large
commit's gate outlast two hours. A mutant that was only slow is judged then,
and one that still runs out is a hang: a loop whose counter the mutant turned
around, which the test waiting on it did notice. Each hang caught is written
to `.git/mutation-hangs.log` with the day and the package, because a hang and
a test so slow it looks like one read the same to the gate, and only a person
can tell them apart; the run ends by counting them, and a package with one gets
no receipt, so the next push looks again. A retry that judged nothing is said
as a gate that did not run, never read as a pass. And the baseline test run is never a cached one
(`-count=1`): read from the cache it takes milliseconds, the budget it sets
with it, and every mutant times out.

**Two mutators are on for the core only.** `invert-logical` (`&&` for `||`) and
`invert-loopctrl` (`continue` for `break`) are off in `.gremlins.yaml` and
turned on by `scripts/mutation.sh` for the packages where the rules are —
domain, engine, task, objective, store, trace, migration, config, build. What
blocks, orders and narrows is written in `&&` and `||`, so a 100% without them
said least exactly there. The surfaces keep them off: their conditions are
about what is on screen, and measured, they gave more mutants no test could
tell from the original than ones pointing at a weak test.

**A threshold is never lowered to get a commit through.** If a mutant survives,
look at the mutant. When it is genuinely equivalent — a mutation that cannot
change any observable behaviour, so no test can ever kill it — restructure the
code so the mutant does not exist, rather than arguing for it. The commit
truncation in `internal/cli` is there for exactly that reason: slicing on a
`min` has no branch to mutate, where `if len > n` had an untestable boundary.

## The report a sweep leaves

A full sweep runs for a long while, and its verdict used to live only in the
terminal it ran in: scrolled away, or gone with the window, it had to be run
again to be read. `make mutation` now leaves it in `mutation-report/`, which is
ignored by git and rewritten by every full sweep:

- `summary.md` — package by package, every mutant that lived or that no test
  reached, at its file, line and column, with the line of code it sits on;
  every hang caught; and a package judged clean, or clean already and
  unchanged since, as one line. It is what fixing a survivor starts from.
- `json/` — what gremlins said about each package, in full.

The commit's run, on the lines it changed, leaves the report alone. A release
does not run the sweep — see [release.md](release.md) — and says instead when
the last one ran and whether it left anything to fix.

## What the last few per cent are, and why they stay there

Coverage sits in the mid-nineties and will not reach a hundred. That is not a
backlog, so what the remainder is made of is written down here — otherwise the
question comes back, and the honest answer takes an afternoon to rediscover.

Counted on the statements no test ever runs:

| | | |
|---|---|---|
| ~three quarters | the inside of an `if err != nil` — a `return fmt.Errorf(…)` after a call that does not fail in a test | **left there** |
| a handful | structurally unreachable | **cannot** |
| the rest | code called from another package and never from its own | **closed as it is found** |

**The error interiors stay.** Reaching them means making SQLite fail in the
middle of a statement, which needs a fault-injecting driver or a fake store —
and the store is tested against real SQLite on purpose. Nor would it buy
anything: **mutation already guards every one of those conditions.** Negating
`err != nil` returns an error where there is none, the happy path breaks, and
the mutant dies. The branch is unexercised and the condition is proven, which
is the distinction coverage cannot draw and mutation can.

**What cannot be reached.** `main` cannot be called from a test without ending
the process that runs it, which is why `cmd/` sits outside the per-package
floor. `test/World.Register` and its scenario hook are reachable only through a
real godog run, and the harness cannot drive one — see [mcp.md](mcp.md).
Building the MCP server inside its command's body would block on standard
input. A `json.Marshal` of a struct that cannot fail has an error branch that
cannot be taken.

**What does get closed** is the one class that is a real gap: code a package
does not test because somebody else calls it. Coverage is per package, so a
scenario in `internal/cli` that reaches the engine, the store and a notion
credits only `internal/cli`. That is how a package once sat at 0% with sixty
functions in it while every suite was green.

## Adding a check

**Every check belongs in `.pre-commit-config.yaml`.** Adding a tool that verifies anything
means wiring it there in the same commit. A check that has to be remembered is
a check that will not run.

Where the command itself lives:

- If it is a single command, prek runs it directly.
- If it needs any logic — a loop, a threshold, a fallback — it becomes a
  Makefile target and prek's entry calls it. `mutation` and `cover-check` work
  that way: gremlins takes one package at a time, so the loop has to live
  somewhere, and one definition beats one in TOML and one in a habit.

There is no `test` and no `lint` target: prek owns both. The one thing that
looks like a duplicate is not one — `coverage.out` runs the suite a second
time, on purpose, because a profile has to be produced by a run somebody
controls. That run carries `-race` and prints what failed, so it is also the
one to reach for by hand during a red-green loop.

## Running them by hand

```sh
make cover          # coverage per function, then the total
make cover-check    # fail below the total floor, or below the per-package one
make tested-check   # fail if a package has no test file
make harness-check  # fail if the binary reaches the test harness
make windows-check  # fail if cairn or its tests no longer build for Windows
make quiet-check    # fail if building cairn or the window's tests prints anything
make mutation-changed  # the lines changed since HEAD: what a commit runs
make mutation          # every package: what a push runs, and the report it leaves
make check             # every hook, commit's and push's, on every file
make release-check     # the same, but the mutation sweep: what a release runs
make linux             # build and run the suite on Linux, in a container
make help              # every target, and what it is for
```

**Linux is checked by `make linux`**, which builds cairn and runs the suite —
with `-race`, as the profile here does — in a container, as a user who is not
root and has the uid of whoever ran it, with the graphics headers the window
needs there. It needs Docker, so it is not a hook; it is run before a version
is made (see [release.md](release.md)). The first run found what nothing on a
Mac could: frames pinned in one time zone, and tests that make a store
read-only failing for root. Drawing off screen needs a GPU the container does
not have, so the machine says so with `CAIRN_WINDOWLESS=1` and the tests that
draw are skipped; anywhere that is not set, a test that cannot draw fails
rather than skipping itself.

`prek run -a` is not everything: it runs the commit's hooks, and the commit's
mutation judges only what changed since `HEAD` — on a clean tree, nothing.
`make check` is the one command that leaves nothing unchecked.

Mutation is the slow one. Every package, every commit, took fourteen minutes
and grew with the suite — long enough to be worked around. So it is split, and
the split is fixed in `.pre-commit-config.yaml`, not left to whoever runs it:

- **at commit**, only the mutants on lines changed since `HEAD`, in the
  packages holding them. Measured: about a minute instead of fourteen. A
  surviving mutant found here costs one test.
- **at push**, every package — about twenty minutes. A test weakened without touching the code it
  guarded changes no production line, so only this one sees it — and nothing
  leaves the machine before it has run. A package the sweep found clean is
  not judged again while nothing its verdict rests on has changed: its own
  files, the code its tests reach in other packages, the scenarios and pinned
  files, `go.sum`, this script, the mutators and the version of gremlins. The
  receipt is the hash of all of it, kept in `.git/mutation-receipts`. It was
  first the package's own files and nothing else — and a scenario weakened
  elsewhere, which is exactly what the push is there to catch, went through
  as judged.

Inside a package the tests run one after another, and `t.Parallel` is
nearly absent on purpose: most of the suite names its store through
`CAIRN_STORE`, which `t.Setenv` sets and `t.Parallel` refuses, or swaps a
package-level seam — the window that opens, the terminal program — that two
tests at once would fight over. `go test` already runs the packages side by
side.

The packages are judged one after another, and that is not a lost chance to
go faster: gremlins already runs one worker per core inside a package, and two
packages at once only made each mutant wait for a core — which is how timeouts
are made.

Gremlins' own `--diff` never matched a single mutant here: it names a file from
the repository root and a mutant from its package. The script hands it a diff
relative to the package, with no context lines — check a run's summary reads
`KILLED`, not only `SKIPPED`, after touching that script.

Every mutant is a build, and builds land in the Go cache — one full sweep left
50 to 150 GB in it. Gremlins gets a cache of its own, `$(go env
GOCACHE)-gremlins`, wiped once it passes 10 GB.

**Two mutants run at a time** (`workers: 2` in `.gremlins.yaml`), not one per
core. Each one runs its package's whole suite, and the window's took 2.5 to
3.5 GB: measured on 2026-10-12, eight at once beside a container VM took a
16 GB machine to 40, starved WindowServer, and macOS restarted the machine —
twice. A sweep is slower for it; a machine that restarts mid-sweep finishes
nothing.
