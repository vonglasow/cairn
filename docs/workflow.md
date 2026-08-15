# Workflow

Read this before starting any piece of work.

## Order of work: BDD, then TDD, then code

1. **BDD first.** A scenario in `bdd/` says what the tool does, in the language
   of someone using it. It is written before anything else, and it is frozen:
   if a step cannot check what the scenario says, the step gets fixed, never the
   scenario.
2. **TDD next.** Each unit needed to satisfy the scenario starts as a failing
   test — failing for the right reason, a behaviour that is absent rather than a
   test that is miswired.
3. **Code last**, and only enough to turn the test green.

Never the reverse. Code written first decides what the test will say, and a
test written after the code documents what the code does rather than what was
wanted.

## Implementing a scenario

Every scenario added to `bdd/` is implemented through `/contract-first`: the
scenario is the contract, and the work is sliced against it rather than
improvised around it.

## Committing

Commits go through the `/commit` skill.

A commit is atomic, its message explains **why** rather than what, and it
leaves the suite green on its own — not green only once the next commit lands.

## Writing scenarios

Scenarios live in `bdd/*.feature`, in English, and are executed by godog. One
scenario is one Go test, selected by tag, so a failure names the scenario it
belongs to:

```go
func TestAcceptance_Version_DegradesCleanly(t *testing.T) {
    runAcceptance(t, "version.feature", initVersion(t), "@version-degraded")
}
```

Conventions:

- **The suite is defined once per package.** `runAcceptance` in
  `internal/cli/cli_test.go` takes the feature file, the initializer and the
  tag. A new feature adds a `*_bdd_test.go` holding its world, its steps and
  its tests — never a second copy of the suite. It is an unexported test
  helper, so a second surface cannot call it; lifting it out is the day that
  surface exists, not before (see `docs/architecture.md`).
- **A scenario cleans up after itself, not after the test.** Anything a step
  changes outside its own world — a package-level variable, a file, an
  environment — is put back by a `sc.After` hook. `t.Cleanup` looks like the
  place and is not: it fires when the Go test ends, so the next scenario in the
  same run starts on the previous one's leftovers. That is invisible while each
  tag holds one scenario, and wrong from the first `Scenario Outline`.
- **Steps live in the package they observe.** A single `.feature` shared across
  two packages fails with undefined steps under `Strict: true`, since each
  package only implements its own.
- **Steps drive the real surface**, not the layer below it. The CLI steps go
  through `run(t, …)`, which calls the same `cli.Main` the binary calls — fang's
  error rendering included — so what a step observes is what a user observes.
- **The two streams stay apart.** `run` returns stdout, stderr and the exit code
  separately, because a scenario that says a reason goes to standard error can
  only be checked if standard error is a thing on its own.
- **A step returns its error**; it never calls `t.Fatalf`. A Goexit from inside
  godog wrecks its subtest bookkeeping and surfaces as an unreadable
  "test executed panic(nil) or runtime.Goexit".
- **A precondition never goes through the code under test.** If it does, the
  precondition and the assertion cancel out, and the scenario stays green even
  when the behaviour is removed — proving nothing.
