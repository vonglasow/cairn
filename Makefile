# cairn — build/run helpers and the checks prek runs at commit time.

BIN     ?= bin/cairn
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse HEAD 2>/dev/null || echo none)
BUILT   ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
PKG      = cairn/internal/build
# go install puts tools in GOPATH/bin, which is not always on PATH.
GREMLINS ?= $(shell command -v gremlins || echo "$$(go env GOPATH)/bin/gremlins")
# Minimum statement coverage. Raise it deliberately; never lower it to get a
# commit through.
COVERAGE_MIN ?= 90
LDFLAGS  = -X $(PKG).version=$(VERSION) -X $(PKG).commit=$(COMMIT) -X $(PKG).built=$(BUILT)

.PHONY: build run cover cover-check harness-check mutation clean

# Make keeps a target whose recipe failed, and a half-written profile is newer
# than its sources — so the next run would grade it as though it were whole.
.DELETE_ON_ERROR:

## build: build the binary into ./bin, stamped with version, commit and time
build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/cairn

## run: run cairn with ARGS, e.g. `make run ARGS="task add 'write the spec'"`
run:
	go run -ldflags "$(LDFLAGS)" ./cmd/cairn $(ARGS)

# The profile is a real artifact, so both `cover` and `cover-check` read it
# instead of each running the suite themselves. It is rebuilt whenever a Go
# file or a scenario changes — a scenario is executed by the suite, so editing
# one changes what ran just as much as editing code does.
#
# The suite's output is not swallowed: this recipe is what runs during a
# red-green loop, and a failure that names no test is a failure you debug
# twice. `-race` costs about a second here and reports two goroutines touching
# the same variable unsynchronised, which `-covermode=atomic` does not — atomic
# only keeps the coverage counters themselves honest under concurrency.
coverage.out: $(shell find . -name '*.go' -not -path './bin/*') $(wildcard bdd/*.feature)
	@go test -race ./... -covermode=atomic -coverprofile=$@

## cover: print coverage per function, then the total
cover: coverage.out
	go tool cover -func=coverage.out

## cover-check: fail if statement coverage falls below COVERAGE_MIN
## Coverage only says a line ran — see `mutation` for whether a test noticed.
cover-check: coverage.out
	@total=$$(go tool cover -func=coverage.out | awk '/^total:/ {gsub("%","",$$NF); print $$NF}'); \
	awk -v t="$$total" -v m="$(COVERAGE_MIN)" 'BEGIN { \
		if (t + 0 < m + 0) { printf "✗ coverage %s%% is below the %s%% floor\n", t, m; exit 1 } \
		printf "✓ coverage %s%% ≥ %s%%\n", t, m }'

## harness-check: fail if the binary reaches the test harness
## `cairn/test` imports `testing` outside a _test.go file on purpose, because Go
## compiles a package's test files only for that package and two of them need
## the same world. So the guard is not "nothing imports testing" — it is the
## narrower thing that worry was always about: nothing the binary is made of
## reaches the harness. Finding it is the failure, which is why the grep is
## inverted.
harness-check:
	@if go list -deps ./cmd/cairn | grep -q '^cairn/test'; then \
		echo "✗ the binary reaches the test harness"; \
		go list -deps ./cmd/cairn | grep '^cairn/test'; \
		exit 1; \
	fi
	@echo "✓ the binary does not reach the test harness"

## mutation: break each line on purpose and check a test turns red
## gremlins takes one package at a time, so we loop. A surviving mutant is a
## test that would not have noticed the bug; the run fails below 100%.
##
## The thresholds live in .gremlins.yaml, not here: gremlins ignores them when
## they are passed as flags. That is measured, not assumed — see the comment
## there before moving them back.
mutation:
	@command -v $(GREMLINS) >/dev/null || { \
		echo "gremlins not found — go install github.com/go-gremlins/gremlins/cmd/gremlins@latest"; \
		exit 1; \
	}
	@for pkg in $$(go list ./... | sed "s|^cairn|.|"); do \
		ls $$pkg/*_test.go >/dev/null 2>&1 || continue; \
		echo "→ $$pkg"; \
		$(GREMLINS) unleash $$pkg || exit 1; \
	done

## clean: remove build and coverage artifacts
clean:
	rm -rf bin coverage.out
