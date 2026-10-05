# Checks

Read this when adding a tool, changing a threshold, or when a hook blocks a
commit.

## What runs, and when

`prek` runs on its own at every commit, and once more at every push. Nothing
here has to be remembered.

| check | what it answers |
|---|---|
| `trailing-whitespace`, `end-of-file-fixer`, `check-yaml`, `check-added-large-files` | is the file well formed? |
| `semgrep` | is there a known bad pattern? |
| `govulncheck` | does anything we call have a known vulnerability? |
| `go-fmt`, `go-imports` | is it formatted the way Go formats? |
| `golangci-lint` | does it hold up to static analysis, comments included? |
| `no-go-testing` | does non-test code import `testing`? |
| `tested-check` | does every package have a test file at all? |
| `harness-check` | is the binary free of the test harness? |
| `cover-check` | does it work — and did every line at least run, in total and in each package? |
| `mutation` | would a test have noticed if a changed line were wrong? |
| `commitlint` | is the message conventional? |
| `mutation-all` | at push only: would a test have noticed if *any* line were wrong? |

There is no `go-unit-tests` hook. `cover-check` builds its profile from `go
test -race ./...`, which fails the moment a test does, so a second hook running
the suite was the same answer paid for twice. The `bdd/` scenarios are Go
tests, so that one run covers them too.

**A hook is triggered by every file its answer depends on**, not only by Go.
`cover-check` also runs when a scenario, a pinned frame under `testdata/`, or
`go.mod` changes; `govulncheck` and `harness-check` when `go.mod` or `go.sum`
does; `golangci-lint` when `.golangci.yml` does. Each says so with `files` in
`prek.toml`. Filtered on Go alone they ran nothing at all for a commit that
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

## Getting them

Without these, the hooks cannot run and nothing can be committed:

```sh
brew install prek golangci-lint goimports govulncheck jq
go install github.com/go-gremlins/gremlins/cmd/gremlins@latest
prek install -t pre-commit -t commit-msg -t pre-push
```

`gremlins` lands in `$(go env GOPATH)/bin`, which is not always on `PATH` — the
Makefile looks there on its own, so only a direct call needs the full path.

`jq` is read by `scripts/mutation.sh` and by both hooks under `.claude/hooks`.

Two are not installed by hand: prek builds `semgrep` and `commitlint` in
environments of its own, from the revisions pinned in `prek.toml`. commitlint
needs Node on `PATH` for that.

Nothing here is pinned to a version. Latest is deliberate: it keeps semgrep's
rules and govulncheck's advisories current, and the cost is that a commit which
passed yesterday can fail today without the code having moved. That is the
right way round — the finding was already true yesterday.

### When Go itself moves

`govulncheck`, `golangci-lint` and `gremlins` are Go programs, and each one
parses the sources it is pointed at. A tool built against an older Go cannot
read a module written for a newer one, and it says so in its own words rather
than in anything about this repository:

| what it says | what it means |
|---|---|
| `Loading packages failed, possibly due to a mismatch between the Go version used to build govulncheck and the Go version on PATH` | govulncheck is behind |
| `panic: file requires newer Go version go1.27 (application built with go1.26)` | golangci-lint is behind |
| `go.mod requires go >= 1.26.6 (running go 1.26.5; GOTOOLCHAIN=local)` | Go itself is behind, and `GOTOOLCHAIN=local` forbids it from fetching the one the module asks for |

So a Go upgrade is not one command but four: the toolchain, then the three
tools above, by the same lines that installed them. None of it is a reason to
skip a hook — every one of these is a tool that cannot read the code, not a
check that found nothing.

Measured, on the day Go went from 1.26.5 to 1.27.1: five advisories appeared
against the standard library, and the two brew-installed tools then refused to
run at all. The hour that cost is the reason this section exists.

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

A scenario is what makes this easy to get wrong, and it is worth saying why:
coverage is measured per package, from that package's own test binary. A
scenario in `internal/cli` that reaches the engine, the store and a notion
proves the behaviour end to end and credits **only** `internal/cli`. The code
it runs elsewhere is exercised without being guarded — and once the scenario is
green there is no red left to make anybody write the unit tests. Green reads as
done.

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
report: any mutant LIVED or NOT COVERED fails the run. A package with a timeout
is judged again with three times the budget — a mutant that was only slow is
judged then, and one that still runs out is a hang: a loop whose counter the
mutant turned around, which the test waiting on it did notice. Measured, the
full sweep has five, all of that kind. And the baseline test run is never a
cached one (`-count=1`): read from the cache it takes milliseconds, the budget
it sets with it, and every mutant times out.

**A threshold is never lowered to get a commit through.** If a mutant survives,
look at the mutant. When it is genuinely equivalent — a mutation that cannot
change any observable behaviour, so no test can ever kill it — restructure the
code so the mutant does not exist, rather than arguing for it. The commit
truncation in `internal/cli` is there for exactly that reason: slicing on a
`min` has no branch to mutate, where `if len > n` had an untestable boundary.

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

**And a trap worth naming.** `go tool cover -func` lists **every membrane
method at 0.0%** — `func (c *Counter) Due(time.Time) {}` and eighty others.
They hold **zero statements between them**: they cannot be covered and they
change no percentage. A notion tells one fact per method and most membranes
ignore most facts, so the list is long, alarming, and empty. Read the statement
counts, not the function list.

**What does get closed** is the one class that is a real gap: code a package
does not test because somebody else calls it. Coverage is per package, so a
scenario in `internal/cli` that reaches the engine, the store and a notion
credits only `internal/cli`. That is how a package once sat at 0% with sixty
functions in it while every suite was green.

## Adding a check

**Every check belongs in `prek.toml`.** Adding a tool that verifies anything
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
make mutation-changed  # the lines changed since HEAD: what a commit runs
make mutation          # every package: what a push runs
make check             # every hook, commit's and push's, on every file
```

`prek run -a` is not everything: it runs the commit's hooks, and the commit's
mutation judges only what changed since `HEAD` — on a clean tree, nothing.
`make check` is the one command that leaves nothing unchecked.

Mutation is the slow one. Every package, every commit, took fourteen minutes
and grew with the suite — long enough to be worked around. So it is split, and
the split is fixed in `prek.toml`, not left to whoever runs it:

- **at commit**, only the mutants on lines changed since `HEAD`, in the
  packages holding them. Measured: about a minute instead of fourteen. A
  surviving mutant found here costs one test.
- **at push**, every package — about twenty minutes. A test weakened without touching the code it
  guarded changes no production line, so only this one sees it — and nothing
  leaves the machine before it has run.

Gremlins' own `--diff` never matched a single mutant here: it names a file from
the repository root and a mutant from its package. The script hands it a diff
relative to the package, with no context lines — check a run's summary reads
`KILLED`, not only `SKIPPED`, after touching that script.

Every mutant is a build, and builds land in the Go cache — one full sweep left
50 to 150 GB in it. Gremlins gets a cache of its own, `$(go env
GOCACHE)-gremlins`, wiped once it passes 10 GB.
