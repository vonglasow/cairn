# Checks

Read this when adding a tool, changing a threshold, or when a hook blocks a
commit.

## What runs, and when

`prek` runs on its own at every commit. Nothing here has to be remembered.

| check | what it answers |
|---|---|
| `trailing-whitespace`, `end-of-file-fixer`, `check-yaml`, `check-added-large-files` | is the file well formed? |
| `semgrep` | is there a known bad pattern? |
| `govulncheck` | does anything we call have a known vulnerability? |
| `go-fmt`, `go-imports` | is it formatted the way Go formats? |
| `golangci-lint` | does it hold up to static analysis, comments included? |
| `no-go-testing` | does non-test code import `testing`? |
| `go-unit-tests` | does it work? |
| `tested-check` | does every package have a test file at all? |
| `cover-check` | did every line at least run — in total, and in each package? |
| `mutation` | would a test have noticed if a line were wrong? |
| `commitlint` | is the message conventional? |

The `bdd/` scenarios are Go tests, so `go-unit-tests` already runs them. They
have no hook of their own — that would run them twice.

## Getting them

Without these, the hooks cannot run and nothing can be committed:

```sh
brew install prek golangci-lint goimports govulncheck
go install github.com/go-gremlins/gremlins/cmd/gremlins@latest
prek install -t pre-commit -t commit-msg
```

`gremlins` lands in `$(go env GOPATH)/bin`, which is not always on `PATH` — the
Makefile looks there on its own, so only a direct call needs the full path.

Two are not installed by hand: prek builds `semgrep` and `commitlint` in
environments of its own, from the revisions pinned in `prek.toml`. commitlint
needs Node on `PATH` for that.

Nothing here is pinned to a version. Latest is deliberate: it keeps semgrep's
rules and govulncheck's advisories current, and the cost is that a commit which
passed yesterday can fail today without the code having moved. That is the
right way round — the finding was already true yesterday.

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

- Statement coverage, in total: at least **90%** (`COVERAGE_MIN`).
- Statement coverage, per package: at least **85%** (`PACKAGE_COVERAGE_MIN`).
  Lower than the total on purpose: a package may sit honestly under the
  average — measured, the store and the harness both do — and a per-package
  floor set at the total would fail the repository as it stands rather than the
  code being added to it. `cmd/` is left out, because its only function is
  `main`, which no test can call without ending the process that runs it; what
  it does is exercised through the scenarios of `internal/cli`.
- Every package has a test file, with no threshold to argue about: it has one
  or it does not.
- Mutation efficacy: **100%** — of the mutants gremlins judged, all died.
- Mutant coverage: **100%** — it judged every mutant it made. Without this
  second one, code no test reaches produces mutants counted in neither column,
  and a run reports success having looked at nothing.

Both live in `.gremlins.yaml`, because gremlins ignores them when they are
passed as flags — measured, not assumed. Both read `99.99`, because the tool
fails on *less than or equal*, so a `100` would refuse a perfect run.

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
make mutation       # fail below 100% efficacy
```

`make mutation` is the slow one — tens of seconds per package, and it grows
with the suite. There is no flag to narrow it: one command, every package,
every time. A gate you can scope is a gate you scope out of habit, and the
package you stopped running is the one that rots.
