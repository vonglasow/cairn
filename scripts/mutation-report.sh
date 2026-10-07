#!/usr/bin/env bash
# mutation-report.sh — write what one package's mutation said, for somebody to
# fix it later.
#
#   mutation-report.sh ./internal/gui ROOT < gremlins-report.json
#
# A full sweep runs for a long while and its verdict used to live only in the
# terminal it ran in: scrolled away, or gone with the window, it had to be run
# again to be read. This writes each package's part of the report kept in
# mutation-report/summary.md — every mutant that lived or that no test reached,
# where it is, and the line it sits on, which is what fixing one starts from;
# and the hangs caught, which somebody has to look at. A clean package is one
# line. ROOT is the repository the package's files are read from.
set -euo pipefail

pkg=${1:?usage: mutation-report.sh PACKAGE ROOT < report.json}
root=${2:?usage: mutation-report.sh PACKAGE ROOT < report.json}
dir=${pkg#./}
said=$(cat)

listed() {
  jq -r --args '.files[]? as $f | $f.mutations[]
    | select(.status | IN($ARGS.positional[]))
    | [.status, .type, $f.file_name, (.line | tostring), (.column | tostring)] | @tsv' "$@" <<<"$said"
}

bad=$(listed "LIVED" "NOT COVERED")
hangs=$(listed "TIMED OUT")
killed=$(jq '[.files[]?.mutations[] | select(.status == "KILLED")] | length' <<<"$said")

if [ -z "$bad" ] && [ -z "$hangs" ]; then
  echo "- \`$pkg\` — clean: $killed killed"
  exit 0
fi

# The line a mutant sits on, as it reads, without its indentation.
line() { sed -n "${2}p" "$root/$dir/$1" | sed 's/^[[:space:]]*//'; }

fix=$(printf '%s' "$bad" | grep -c . || true)
if [ "$fix" -gt 0 ]; then
  echo "## \`$pkg\` — $fix to fix"
else
  echo "## \`$pkg\` — nothing to fix, hangs to look at"
fi
echo
while IFS=$'\t' read -r status type file at column; do
  [ -n "$status" ] || continue
  echo "- $status $type at \`$dir/$file:$at:$column\`"
  echo "  \`$(line "$file" "$at")\`"
done <<<"$bad"
while IFS=$'\t' read -r _ type file at column; do
  [ -n "$type" ] || continue
  echo "- hang caught: $type at \`$dir/$file:$at:$column\` — a loop a test noticed, or a test too slow to tell from one"
  echo "  \`$(line "$file" "$at")\`"
done <<<"$hangs"
echo
