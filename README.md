# cairn

A personal objective and task manager. CLI first, then MCP; the terminal and
desktop boards come later, once the engine stands on its own.

Nothing works yet beyond `cairn --version`: the project is still its scaffold.

## Build

```sh
make build                      # ./bin/cairn
make run ARGS="--version"
```

## Working on it

[CLAUDE.md](CLAUDE.md) carries the rules, `docs/` the detail. `prek` runs the
checks at every commit — [docs/checks.md](docs/checks.md) says what they are
and how to install them.
