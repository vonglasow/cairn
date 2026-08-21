# Definition of done

The steps a piece of work goes through, in order. Ticked before the
commit that finishes it — see [workflow.md](workflow.md) for what each
step means.

A box that cannot be ticked is said out loud and confirmed before the
commit, never ticked anyway. The list describes the work, not the
mistakes: a step is added here when the work gains one, not when
something goes wrong.

## Agree
- [ ] What is being built was agreed, not assumed
- [ ] It was looked for first: nothing here exists twice

## Contract
- [ ] A scenario in bdd/ describes it, agreed before any code
- [ ] It reads in the language of somebody using cairn

## Test
- [ ] Each unit started as a test failing for the right reason

## Build
- [ ] The code written is only what turns those tests green
- [ ] Every change of state goes through internal/engine

## Verify
- [ ] Every suite is green and prek passed on its own
- [ ] Coverage is at or above the floor, mutation at 100%
- [ ] It was exercised against the built binary, not reasoned about
- [ ] Edge cases were probed, including what should be refused

## Document
- [ ] README and docs/ say what cairn now does

## Deliver
- [ ] The commit is atomic and leaves the suite green alone
- [ ] Its message follows the rule in workflow.md
