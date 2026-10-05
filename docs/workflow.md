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

Scenarios live in `bdd/*.feature`, in English, and are executed by godog. A
feature runs whole, every scenario in it, so one written there is one that runs
— a scenario that needs a test of its own to be listed is a scenario that can be
written and never run, and half of `config.feature` once was. godog names each
scenario as a subtest, so a failure still names the scenario it belongs to:

```go
func TestAcceptance_Config(t *testing.T) {
    runAcceptance(t, "config.feature", initConfig(t), "~@later")
}
```

A scenario written before what it describes is built carries `@later` in the
feature, beside its own tag, where whoever reads it sees it, and the run leaves
it out with `~@later`. It is the only way a scenario goes unrun.

Conventions:

- **The suite is defined once.** `test.Suite` in `test/harness.go` takes the
  feature file, the initializer and the tags; each package that holds scenarios
  has a one-line `runAcceptance` over it, so a failure is that package's own. A
  new feature adds a `*_bdd_test.go` holding its world and its steps — never a
  second copy of the suite.
- **A scenario cleans up after itself, not after the test.** Anything a step
  changes outside its own world — a package-level variable, a file, an
  environment — is put back by a `sc.After` hook. `t.Cleanup` looks like the
  place and is not: it fires when the Go test ends, so the next scenario in the
  same run starts on the previous one's leftovers.
- **Steps live in the package they observe** — unless more than one package
  claims them, and then they live once in `test`: `test.World` for what a
  command line runs, `test.Givens` for what both boards say is in the store. A
  step written twice is a step that will come to mean two things.
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
