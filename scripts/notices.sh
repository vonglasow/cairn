#!/usr/bin/env bash
# notices.sh — the licences of everything the binary is built from, in one file.
#
#   scripts/notices.sh > bin/THIRD_PARTY_NOTICES
#
# A binary handed to somebody carries code under other people's licences, and
# most of them ask for their text to travel with it. Read from the module
# cache, for the modules the binary actually links — not every module go.mod
# names, which includes what only the tests use — and Go's own, since the
# runtime and the standard library are linked too. Every licence and notice
# file a module carries is kept, not the first found: a module under two
# licences, or with a NOTICE beside its LICENSE, asks for both. A module with
# none fails the script: a notice with a hole in it is not one to ship.
set -euo pipefail

# Go answers in the system's own spelling, which on Windows is C:\... with its
# backslashes; Git Bash's find takes it for no path at all, and every licence
# looked for went missing. Spelt as the shell spells it there, it is found.
unix() {
  if command -v cygpath >/dev/null; then cygpath -u "$1"; else printf '%s' "$1"; fi
}
cache=$(unix "$(go env GOMODCACHE)")
missing=0

# escaped is a module path as the cache spells it: an upper-case letter as !
# and its lower case. Written out rather than left to sed, whose \L is GNU's.
escaped() {
  local path=$1 out="" c i
  for ((i = 0; i < ${#path}; i++)); do
    c=${path:i:1}
    if [[ $c == [[:upper:]] ]]; then
      out+="!$(printf '%s' "$c" | tr '[:upper:]' '[:lower:]')"
    else
      out+=$c
    fi
  done
  printf '%s' "$out"
}
echo "================================================================"
echo "Go $(go env GOVERSION) — the runtime and the standard library"
echo "================================================================"
# Where Go keeps its licence depends on who installed it: beside the sources,
# or one level up when a package manager moved it.
goroot=$(unix "$(go env GOROOT)")
golicence=$(find "$goroot" "$goroot/.." -maxdepth 1 -name LICENSE 2>/dev/null | head -1)
if [ -z "$golicence" ]; then
  echo "no licence file found for Go under $goroot" >&2
  missing=1
else
  cat "$golicence"
fi
echo

# What the tree itself carries from somebody else — a font cut from DejaVu
# Sans among it — keeps its licence beside it, and that travels too.
while read -r carried; do
  [ -n "$carried" ] || continue
  echo "================================================================"
  echo "$(dirname "$carried") — carried in the tree"
  echo "================================================================"
  cat "$carried"
  echo
done < <(git ls-files '*/LICENSE')

while read -r module; do
  [ -n "$module" ] || continue
  dir="$cache/$(escaped "$module")"
  echo "================================================================"
  echo "$module"
  echo "================================================================"
  found=$(find "$dir" -maxdepth 1 -type f \( -iname 'licen[cs]e*' -o -iname 'copying*' -o -iname 'notice*' -o -iname 'unlicense*' \) | sort)
  if [ -z "$found" ]; then
    echo "no licence file found in $dir" >&2
    missing=1
    continue
  fi
  while read -r file; do
    cat "$file"
    echo
  done <<<"$found"
done < <(go list -deps -f '{{with .Module}}{{if not .Main}}{{.Path}}@{{.Version}}{{end}}{{end}}' ./cmd/cairn | sort -u)

exit "$missing"
