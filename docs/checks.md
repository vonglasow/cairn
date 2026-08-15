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
| `cover-check` | did every line at least run? |
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
mutants to judge, so mutation would pass it in silence. Coverage sees it drop.

## Thresholds

- Statement coverage: at least **90%** (`COVERAGE_MIN` in the Makefile).
- Mutation efficacy: **100%**.

**A threshold is never lowered to get a commit through.** If a mutant survives,
look at the mutant. When it is genuinely equivalent — a mutation that cannot
change any observable behaviour, so no test can ever kill it — restructure the
code so the mutant does not exist, rather than arguing for it. The commit
truncation in `internal/cli` is there for exactly that reason: slicing on a
`min` has no branch to mutate, where `if len > n` had an untestable boundary.

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
make cover-check    # fail below COVERAGE_MIN
make mutation       # fail below 100% efficacy
```

`make mutation` is the slow one — tens of seconds per package, and it grows
with the suite. There is no flag to narrow it: one command, every package,
every time. A gate you can scope is a gate you scope out of habit, and the
package you stopped running is the one that rots.
