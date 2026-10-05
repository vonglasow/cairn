#!/usr/bin/env bash
# mutation.sh — run gremlins over the packages, and judge what it reports.
#
#   mutation.sh            every package with a test file: the full sweep, at push
#   mutation.sh --changed  only the lines changed since HEAD, in the packages they
#                          are in: at commit
#
# The judgement is made here and not by gremlins' own thresholds, for two
# reasons read in its source (v0.6.0, internal/report/report.go):
#
#   - a mutant that TIMED OUT is counted in neither efficacy nor mutant
#     coverage, and no threshold sees it: measured on this repository, 219
#     timeouts under load hid 21 survivors behind a reported 100%. Yet some
#     timeouts are real — a mutant that turns a loop's counter around never
#     ends, and the test that waits on it did notice. So the mutants that
#     timed out are run again with three times the budget — those alone, on
#     the lines they sit on, not the whole package a second time: a mutant that
#     was only slow is judged that time, and one that still runs out is a
#     hang, caught, and written down for somebody to look at.
#   - in --diff mode the mutants off the changed lines are kept as SKIPPED, so a
#     package whose diff reaches no mutable line reports 0 killed out of 0 and
#     fails its own threshold for having had nothing to judge.
#
# So: a run fails on any mutant LIVED or NOT COVERED, and on nothing else —
# which is the 100% the thresholds asked for, with the hole closed.
set -euo pipefail

GREMLINS=${GREMLINS:-gremlins}
# What "changed" is measured against. HEAD at commit; anything else is for
# measuring the gate on a commit already made.
since=${MUTATION_SINCE:-HEAD}
changed=false
[ "${1:-}" = "--changed" ] && changed=true

# Gremlins builds every mutant, and every build lands in the Go cache: one full
# sweep left 50 to 150 GB in it. It gets a cache of its own, so the one the rest
# of the work uses stays warm and small, and its own is wiped once it passes a
# ceiling rather than after every run — wiped every run, every run starts cold.
cache=${GREMLINS_CACHE:-$(go env GOCACHE)-gremlins}
ceiling_kb=$((10 * 1024 * 1024))
mkdir -p "$cache"
if [ "$(du -sk "$cache" | cut -f1)" -gt "$ceiling_kb" ]; then
  echo "gremlins' cache is past 10 GB: wiping it"
  rm -rf "$cache" && mkdir -p "$cache"
fi
export GOCACHE=$cache
# Gremlins times each mutant against 30 times its baseline test run. With a warm
# cache that baseline is a cached result, read in milliseconds, and every mutant
# then times out: measured, 66 timeouts on lines no mutant can make loop. The
# baseline is run for real, so the budget it sets is one a test can meet.
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-count=1"

packages() {
  if $changed; then
    # The directories holding a production file that changed. A test file has
    # no mutants of its own; a test weakened in it is what the full sweep, at
    # push, is there to catch.
    git diff --name-only "$since" -- '*.go' ':!*_test.go' | xargs -n1 dirname 2>/dev/null | sort -u |
      while read -r dir; do [ -d "$dir" ] && echo "./$dir"; done
  else
    scripts/packages.sh
  fi
}

# Gremlins' --diff names a file by its path from the repository root, and a
# mutant by its path from the package it was asked for: the two never meet, and
# every mutant is skipped. It also reads a hunk's added lines as if they were
# one run, which context lines break. So git diff, as gremlins calls it, is
# asked for paths relative to the package and for no context — the diff it
# gets then names what it mutates.
shim=$(mktemp -d)
report=$shim/report.json
trap 'rm -rf "$shim"' EXIT
git=$(command -v git)
# A second run names its own lines: RETRYING points at a diff written below,
# naming only the lines whose mutants timed out, and gremlins mutates those.
cat >"$shim/git" <<EOF
#!/bin/sh
[ "\$1" = diff ] && [ -n "\$RETRYING" ] && exec cat "\$RETRYING"
[ "\$1" = diff ] && shift && exec "$git" diff --relative="\$MUTATING" -U0 "\$@"
exec "$git" "\$@"
EOF
chmod +x "$shim/git"
# unleash runs gremlins on one package. Its exits 10 and 11 are its own
# thresholds, judged here from what it reports; anything else is gremlins
# failing to run, which is never a pass.
unleash() {
  local pkg=$1 status=0
  shift
  rm -f "$report"
  MUTATING=${pkg#./} PATH="$shim:$PATH" "$GREMLINS" "$@" "$pkg" || status=$?
  if [ "$status" -ne 0 ] && [ "$status" -ne 10 ] && [ "$status" -ne 11 ]; then
    echo "gremlins could not run on $pkg (exit $status)"
    exit 1
  fi
}

# judged lists the mutants of the last report in any of the statuses given.
judged() {
  jq -r --args '.files[]? as $f | $f.mutations[]
    | select(.status | IN($ARGS.positional[]))
    | "  \(.status) \(.type) at \($f.file_name):\(.line):\(.column)"' "$@" <"$report"
}

# timedOut writes a diff naming each line a mutant timed out on, one hunk a
# line: the shape gremlins reads a changed line from.
timedOut() {
  jq -r '.files[]? as $f | $f.mutations[] | select(.status == "TIMED OUT") | "\($f.file_name) \(.line)"' <"$report" |
    sort -u | while read -r file line; do
      printf 'diff --git a/%s b/%s\n--- a/%s\n+++ b/%s\n@@ -%s,0 +%s,1 @@\n+\n' \
        "$file" "$file" "$file" "$file" "$line" "$line"
    done
}

# What the full sweep found clean, by the content of the package it judged:
# a package whose files have not changed since is not judged again at the
# next push. Kept in the repository's own git directory, out of the tree.
receipts=$(git rev-parse --git-dir)/mutation-receipts
mkdir -p "$receipts"
contentOf() { cat "$1"/*.go go.sum | git hash-object --stdin; }
receiptOf() { echo "$receipts/$(echo "${1#./}" | tr / _)"; }

# Where a hang caught is written down: each one is either a loop a test did
# notice, or a test so slow it looks like one, and only a person can tell.
hanged=$(git rev-parse --git-dir)/mutation-hangs.log

# The packages judged with && and || swapped and continue turned into break as
# well: the core, where the rules are. Those two are off in .gremlins.yaml for
# the surfaces, whose conditions are about what is on screen and give more
# mutants no test can tell from the original than ones pointing at a weak
# test. Every rule of what blocks, orders and narrows is written in && and ||,
# so the core is where a 100% without them meant least.
logical=" internal/domain internal/engine internal/task internal/objective internal/store internal/trace internal/migration internal/config internal/build "

failed=0
ran=0
# A package with no test file is tested-check's to refuse, not this gate's to
# judge: it has no test that could kill a mutant. Nothing changed is nothing to
# judge — not every package, which is what the lister answers when handed none.
candidates=$(packages)
judged_packages=
if [ -n "$candidates" ]; then
  # shellcheck disable=SC2086
  judged_packages=$(scripts/packages.sh --tested $candidates)
fi
for pkg in $judged_packages; do
  if ! $changed && [ "$(cat "$(receiptOf "$pkg")" 2>/dev/null)" = "$(contentOf "$pkg")" ]; then
    echo "→ $pkg: judged clean already, and unchanged since"
    continue
  fi
  ran=$((ran + 1))
  echo "→ $pkg"
  args=(unleash -o "$report")
  $changed && args+=(--diff "$since")
  [[ $logical == *" ${pkg#./} "* ]] && args+=(--invert-logical --invert-loopctrl)
  unleash "$pkg" "${args[@]}"
  # A package with nothing to mutate gets no report at all.
  [ -f "$report" ] || continue
  bad=$(judged "LIVED" "NOT COVERED")
  jq -r '"  " + ([.files[]?.mutations[].status] | group_by(.) | map("\(.[0]): \(length)") | join(", "))' "$report"
  if [ -n "$(judged "TIMED OUT")" ]; then
    echo "  timeouts: judging those again, with three times the budget"
    retry=$shim/retry.diff
    timedOut >"$retry"
    RETRYING=$retry unleash "$pkg" "${args[@]}" --diff "$since" --timeout-coefficient 90
    bad=$(printf '%s\n%s' "$bad" "$(judged "LIVED" "NOT COVERED")" | sed '/^$/d')
    hangs=$(judged "TIMED OUT")
    if [ -n "$hangs" ]; then
      echo "${hangs//TIMED OUT/HANGS (caught)}"
      echo "${hangs//TIMED OUT/$(date +%F) $pkg}" >>"$hanged"
      echo "  written down in $hanged"
    fi
    jq -r '"  again: " + ([.files[]?.mutations[] | select(.status != "SKIPPED") | .status] | group_by(.) | map("\(.[0]): \(length)") | join(", "))' "$report"
  fi
  if [ -n "$bad" ]; then
    echo "$bad"
    failed=1
    continue
  fi
  $changed || contentOf "$pkg" >"$(receiptOf "$pkg")"
done

if $changed && [ "$ran" -eq 0 ]; then
  echo "no production code changed: nothing to mutate"
fi
exit $failed
