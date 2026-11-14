# cairn

A personal objective and task manager: a command line, an MCP server, a board
in the terminal and the same board in a window — four surfaces over one engine.

## Code

- **English** everywhere: code, commits, scenarios, docs.
- **No comments**, but a one-line doc comment on each exported symbol. The
  reasons live in the commits and the decision register.
- **Every change of state goes through `internal/engine`.** A surface parses,
  calls `Ask`, and prints what the engine answers. It decides nothing.
- **Nothing exists twice.** Look for it before writing it.
- **The least code.** Write only what the task asks. Add nothing unasked.
- No dependency without a written reason.

## Tests

- **A change of behaviour starts as a failing test**, failing for the right
  reason. Then the least code that turns it green.
- **A refactoring changes no test's verdict**: the suite is green before and
  after.
- **Behaviour is tested once**, in table tests in the engine. Each surface has
  one contract test. A scenario in `bdd/` says what a person sees; it is never
  reworded to suit the code.
- **No test waits on time.** Clocks and processes are injected.
- Coverage stays at or above its floors. No threshold is lowered to pass.

## Docs

- **`docs/` is for the person using cairn**: guides by task, a glossary, and
  a reference generated from the engine. Never why cairn is built so.
- **The developer reads `ARCHITECTURE.md` and the decision register**, one line
  per decision. Until cairn#318 lands, that is `.claude/docs/`.
- A behaviour that changes changes both.

## Process

- **A task cairn worked on is held.** The board is
  `CAIRN_STORE=$PWD/.cairn/cairn.db`, `CAIRN_NOTES=$PWD/.cairn/notes`. Put the
  task `current` and in focus before working on it. What was done or decided
  goes in `.cairn/notes/<n>/`: a dated entry `YYYY-MM-DD-<subject>.md` and
  `summary.md` rewritten whole — never in its description. When committed,
  unfocus it and put it `done`. The cairn skill has the rest.
- **One commit per task**, once it is done. Its message says why.
- **`prek` runs at every commit**; what it reports is fixed. Never
  `git commit --no-verify` or `-n`, never uninstall the hooks.
- **Merging into `main`, tagging and pushing are the user's.**

## Done

A box that cannot be ticked is said, with why, and the user is asked before
the commit.

- [ ] **Agreed** — what is built was agreed, and looked for first.
- [ ] **Tested** — a test was seen failing first, or, for a refactoring, the
      suite was green before and after.
- [ ] **Green** — `prek` and every suite pass; coverage is at its floors.
- [ ] **Exercised** — run against the built binary, refusals included.
- [ ] **Documented** — `docs/` says what cairn does, the developer docs why.
