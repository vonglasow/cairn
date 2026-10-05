# cairn — build/run helpers and the checks prek runs at commit time.

BIN     ?= bin/cairn
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse HEAD 2>/dev/null || echo none)
BUILT   ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
PKG      = cairn/internal/build
# go install puts tools in GOPATH/bin, which is not always on PATH.
GREMLINS ?= $(shell command -v gremlins || echo "$$(go env GOPATH)/bin/gremlins")
# The version the mutation gate is judged with. Named, never @latest: what
# scripts/mutation.sh works around was read in this version's source, and a
# gate that changed under a commit nobody made would pass or fail for a reason
# that is not in this repository.
GREMLINS_VERSION ?= v0.6.0
# The linter and the vulnerability scanner, named for the same reason. What
# govulncheck finds stays current — its database is read on the day — and what
# it is does not move under a commit nobody made (docs/checks.md).
GOLANGCI_VERSION ?= 2.13.2
GOVULNCHECK_VERSION ?= v1.8.0
# The two coverage floors, in total and per package. Why each is where it is,
# and why the second sits below the first, is said once, in docs/checks.md.
# Raise them deliberately; never lower them to get a commit through.
COVERAGE_MIN ?= 90
PACKAGE_COVERAGE_MIN ?= 85
LDFLAGS  = -X $(PKG).version=$(VERSION) -X $(PKG).commit=$(COMMIT) -X $(PKG).built=$(BUILT)

.PHONY: release tools-present help build run red cover cover-check tested-check harness-check gremlins-present mutation mutation-changed check linux notices clean

## help: say what every target here does, read off the comment above it
## The first line of each `## target:` comment, so a target added with one is
## listed without anything else to keep true. The first target, so a bare
## `make` answers this rather than building.
help:
	@awk 'NR == FNR { if (match($$0, /^[a-z-]+:/)) target[substr($$0, 1, RLENGTH - 1)] = 1; next } \
		/^## [a-z-]+:/ { sub(/^## /, ""); split($$0, said, ": "); \
			if (said[1] in target) printf "  %-17s %s\n", said[1], substr($$0, length(said[1]) + 3) }' \
		$(firstword $(MAKEFILE_LIST)) $(firstword $(MAKEFILE_LIST))

# Make keeps a target whose recipe failed, and a half-written profile is newer
# than its sources — so the next run would grade it as though it were whole.
.DELETE_ON_ERROR:

## build: build the binary into ./bin, stamped with version, commit and time
build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/cairn

## run: run cairn with ARGS, e.g. `make run ARGS="task add 'write the spec'"`
run:
	go run -ldflags "$(LDFLAGS)" ./cmd/cairn $(ARGS)

## red: run PKG's tests, and record the failure so production code may be written
## Nothing else mints the receipt that .claude/hooks/red-first.sh looks for, and
## nothing mints one from a suite that passed. What it attests is not that a
## test exists — coverage already says that — but that its failure was seen.
##
## MOVING is the one way past a suite that passes, and it is for code leaving a
## package: there is no absent behaviour to write a test for, so there is no
## failure to see. It demands a sentence, keeps it in the receipt, and marks the
## receipt as a move — so a write let through this way stays legible afterwards.
##
##   make red PKG=./internal/tui MOVING="askedFrom moves to internal/reading"
red:
	@.claude/hooks/red.sh "$(PKG)" "$(RUN)" "$(MOVING)"

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
coverage.out: $(shell find . -name '*.go' -not -path './bin/*') $(wildcard bdd/*.feature) $(wildcard internal/*/testdata/*) go.mod go.sum
	@go test -race ./... -covermode=atomic -coverprofile=$@

## cover: print coverage per function, then the total
cover: coverage.out
	go tool cover -func=coverage.out

## cover-check: fail if coverage falls below the floor, in total and per package
## Coverage only says a line ran — see `mutation` for whether a test noticed.
##
## Two floors and not one, because a total dilutes: ten packages high absorb one
## at zero until they cannot, and by then the untested package has been there a
## while. Measured — a package added at 0% cost eight points of the total and
## only then broke the floor.
##
## `cmd/` is left out of the per-package floor — docs/checks.md says why.
cover-check: coverage.out
	@failed=0; \
	total=$$(go tool cover -func=coverage.out | awk '/^total:/ {gsub("%","",$$NF); print $$NF}'); \
	awk -v t="$$total" -v m="$(COVERAGE_MIN)" 'BEGIN { \
		if (t + 0 < m + 0) { printf "\xe2\x9c\x97 coverage %s%% is below the %s%% floor\n", t, m; exit 1 } \
		printf "\xe2\x9c\x93 coverage %s%% \xe2\x89\xa5 %s%%\n", t, m }' || failed=1; \
	awk -v m="$(PACKAGE_COVERAGE_MIN)" ' \
		NR > 1 { \
			split($$1, where, ":"); pkg = where[1]; sub("/[^/]*$$", "", pkg); \
			statements[pkg] += $$2; if ($$3 > 0) reached[pkg] += $$2 } \
		END { \
			for (pkg in statements) { \
				if (index(pkg, "cairn/cmd/") == 1) continue; \
				share = 100 * reached[pkg] / statements[pkg]; \
				if (share + 0 < m + 0) { \
					printf "\xe2\x9c\x97 %s is at %.1f%%, below the %s%% each package owes\n", pkg, share, m; \
					bad = 1 } } \
			if (bad) exit 1; \
			printf "\xe2\x9c\x93 every package is at %s%% or better\n", m }' coverage.out || failed=1; \
	if [ $$failed -eq 1 ]; then exit 1; fi

## tested-check: fail if a package has no test file at all
## The mutation gate does the opposite: it skips a package with no test file, so
## the newest code is the code it says least about. Measured — a package added
## with sixty functions and no test file passed `make mutation` untouched, while
## the coverage total absorbed it as an average. This is the gate that was
## missing, and it is the cheapest of the three: a package either has a test
## file or it does not.
tested-check:
	@untested=$$(scripts/packages.sh --untested); \
	if [ -n "$$untested" ]; then \
		for pkg in $$untested; do echo "✗ $$pkg has no test file"; done; \
		exit 1; \
	fi; \
	echo "✓ every package has a test file"

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

## gremlins-present: both mutation targets need it, so they ask once here.
gremlins-present:
	@command -v $(GREMLINS) >/dev/null || { \
		echo "gremlins not found — go install github.com/go-gremlins/gremlins/cmd/gremlins@$(GREMLINS_VERSION)"; \
		exit 1; \
	}
	@go version -m $$(command -v $(GREMLINS)) | grep -q "gremlins[[:space:]]*$(GREMLINS_VERSION)" || { \
		echo "gremlins is not $(GREMLINS_VERSION) — go install github.com/go-gremlins/gremlins/cmd/gremlins@$(GREMLINS_VERSION)"; \
		exit 1; \
	}

## tools-present: refuse to judge with a linter or a scanner other than the one named here
## brew upgrades both on its own, and a lint rule that appeared overnight would
## fail — or pass — a commit for a reason that is not in this repository.
tools-present:
	@golangci-lint --version 2>/dev/null | grep -q "version $(GOLANGCI_VERSION) " || { \
		echo "golangci-lint is not $(GOLANGCI_VERSION) — see docs/checks.md, Getting them"; exit 1; }
	@govulncheck -version 2>/dev/null | grep -q "govulncheck@$(GOVULNCHECK_VERSION)" || { \
		echo "govulncheck is not $(GOVULNCHECK_VERSION) — go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)"; exit 1; }

## mutation: break each line on purpose and check a test turns red
## gremlins takes one package at a time, so we loop. A surviving mutant is a
## test that would not have noticed the bug; the run fails below 100%.
##
## The thresholds live in .gremlins.yaml, not here: gremlins ignores them when
## they are passed as flags. That is measured, not assumed — see the comment
## there before moving them back.
mutation: gremlins-present
	@GREMLINS=$(GREMLINS) scripts/mutation.sh

## mutation-changed: the same judgement on the lines changed since HEAD only, in
## the packages they are in. It is what a commit runs; the full sweep above is
## what a push runs, so nothing leaves the machine without every package judged.
mutation-changed: gremlins-present
	@GREMLINS=$(GREMLINS) scripts/mutation.sh --changed

## check: every hook prek has, on every file — the commit's, then the push's.
## `prek run -a` alone runs only the commit's, and the commit's mutation judges
## the lines changed since HEAD: on a clean tree, none. This is the one command
## after which nothing is left unchecked.
check:
	@prek run -a
	@prek run -a --stage pre-push

## linux: build cairn and run its suite on Linux, in a container
## macOS is where cairn is written and Linux is the other half of what it says
## it runs on; this is what checks that half. It needs Docker. See the script.
linux:
	@scripts/linux.sh

## release: check that this is a version, check everything, then build it and what travels with it
## A version is a tag on a clean main, and nothing else; this refuses to build
## one from anything else, then runs what docs/release.md lists, in its order,
## and writes the binary's checksum beside it.
release:
	@[ "$$(git rev-parse --abbrev-ref HEAD)" = main ] || { echo "a version is made on main"; exit 1; }
	@[ -z "$$(git status --porcelain)" ] || { echo "a version is made from a clean tree"; exit 1; }
	@git describe --tags --exact-match >/dev/null 2>&1 || { echo "a version is a tag on this commit — see docs/release.md"; exit 1; }
	@$(MAKE) check
	@$(MAKE) linux
	@$(MAKE) build
	@$(MAKE) notices
	@cd bin && shasum -a 256 cairn > cairn.sha256 && cat cairn.sha256

## notices: write the licences of what the binary is built from, beside it
## What travels with a binary handed to somebody — see docs/release.md.
notices:
	@mkdir -p bin && scripts/notices.sh > bin/THIRD_PARTY_NOTICES

## clean: remove build and coverage artifacts
clean:
	rm -rf bin coverage.out
